package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/layzdonw/transerver/config"
	"github.com/layzdonw/transerver/transcribe"
	"github.com/sirupsen/logrus"
)

type Server struct {
	transcriber *transcribe.SherpaTranscriber
	router      *gin.Engine
	upgrader    websocket.Upgrader
	logger      *logrus.Logger
}

type TranscribeRequest struct {
	AudioData []byte `json:"audio_data"`
	Format    string `json:"format"`
}

type TranscribeResponse struct {
	Success bool                            `json:"success"`
	Result  *transcribe.TranscriptionResult `json:"result,omitempty"`
	Error   string                          `json:"error,omitempty"`
}

// 实时转录会话
type RealtimeSession struct {
	conn       *websocket.Conn
	recognizer *sherpa_onnx.OnlineRecognizer
	stream     *sherpa_onnx.OnlineStream
	sampleRate int
	logger     *logrus.Logger
	mu         sync.Mutex
	isActive   bool
}

func NewServer(transcriber *transcribe.SherpaTranscriber) *Server {
	server := &Server{
		transcriber: transcriber,
		router:      gin.Default(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源，生产环境中应该更严格
			},
		},
		logger: logrus.New(),
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// 健康检查端点
	s.router.GET("/health", s.healthCheck)

	// 转录端点
	s.router.POST("/transcribe", s.transcribeHandler)

	// WebSocket 端点用于实时转录
	s.router.GET("/ws/realtime", s.realtimeTranscribeHandler)

	// 静态文件服务（可选）
	s.router.Static("/static", "./static")
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "转录服务器运行正常",
	})
}

func (s *Server) transcribeHandler(c *gin.Context) {
	var req TranscribeRequest

	// 处理 multipart/form-data
	if c.ContentType() == "multipart/form-data" {
		file, err := c.FormFile("audio")
		if err != nil {
			c.JSON(http.StatusBadRequest, TranscribeResponse{
				Success: false,
				Error:   "无法获取音频文件: " + err.Error(),
			})
			return
		}

		f, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, TranscribeResponse{
				Success: false,
				Error:   "无法打开音频文件: " + err.Error(),
			})
			return
		}
		defer f.Close()

		var buf bytes.Buffer
		_, err = buf.ReadFrom(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, TranscribeResponse{
				Success: false,
				Error:   "无法读取音频文件: " + err.Error(),
			})
			return
		}

		req.AudioData = buf.Bytes()
		req.Format = "wav" // 默认格式
	} else {
		// 处理 JSON 请求
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, TranscribeResponse{
				Success: false,
				Error:   "无效的请求格式: " + err.Error(),
			})
			return
		}
	}

	// 执行转录
	result, err := s.transcriber.TranscribeAudio(req.AudioData, req.Format)
	if err != nil {
		s.logger.Errorf("转录失败: %v", err)
		c.JSON(http.StatusInternalServerError, TranscribeResponse{
			Success: false,
			Error:   "转录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TranscribeResponse{
		Success: true,
		Result:  result,
	})
}

func (s *Server) realtimeTranscribeHandler(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		s.logger.Errorf("WebSocket 升级失败: %v", err)
		return
	}

	s.logger.Info("实时转录 WebSocket 连接已建立")

	// 创建实时转录会话
	session := &RealtimeSession{
		conn:       conn,
		recognizer: s.transcriber.GetRecognizer(),
		sampleRate: s.transcriber.GetSampleRate(),
		logger:     s.logger,
		isActive:   true,
	}

	// 创建新的音频流
	stream := sherpa_onnx.NewOnlineStream(session.recognizer)
	if stream == nil {
		session.logger.Error("无法创建音频流")
		conn.WriteJSON(TranscribeResponse{
			Success: false,
			Error:   "无法创建音频流",
		})
		conn.Close()
		return
	}
	session.stream = stream

	// 发送连接成功消息
	conn.WriteJSON(TranscribeResponse{
		Success: true,
		Result: &transcribe.TranscriptionResult{
			Text: "连接已建立，开始实时转录...",
		},
	})

	// 启动实时转录处理
	go session.handleRealtimeTranscription()

	// 等待连接关闭
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
	session.cleanup()
}

func (rs *RealtimeSession) handleRealtimeTranscription() {
	defer rs.cleanup()

	for rs.isActive {
		// 读取音频数据
		_, message, err := rs.conn.ReadMessage()
		if err != nil {
			rs.logger.Errorf("读取 WebSocket 消息失败: %v", err)
			break
		}

		// 解析消息
		var req TranscribeRequest
		if err := json.Unmarshal(message, &req); err != nil {
			rs.sendError("无效的消息格式: " + err.Error())
			continue
		}

		// 处理音频数据
		if err := rs.processAudioChunk(req.AudioData, req.Format); err != nil {
			rs.sendError("处理音频数据失败: " + err.Error())
			continue
		}
	}
}

func (rs *RealtimeSession) processAudioChunk(audioData []byte, format string) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	// 处理音频数据
	audioSamples, err := rs.processAudioData(audioData, format)
	if err != nil {
		return fmt.Errorf("处理音频数据失败: %v", err)
	}

	// 将音频数据输入到流中
	rs.stream.AcceptWaveform(rs.sampleRate, audioSamples)

	// 获取识别结果
	result := rs.recognizer.GetResult(rs.stream)
	if result.Text != "" {
		// 检查是否是最终结果（这里简化处理）
		// 在实际应用中，您可能需要更复杂的逻辑来判断是否是最终结果
		isFinal := len(audioSamples) < rs.sampleRate // 如果音频块小于1秒，可能是最终结果
		rs.sendResult(result.Text, isFinal)
	}

	return nil
}

func (rs *RealtimeSession) processAudioData(audioData []byte, format string) ([]float32, error) {
	// 这里复用 transcribe 包中的音频处理逻辑
	// 简化处理，假设是原始 PCM 数据
	if format == "pcm" {
		audioSamples := make([]float32, 0, len(audioData)/2)
		for i := 0; i < len(audioData)-1; i += 2 {
			sample := int16(audioData[i]) | (int16(audioData[i+1]) << 8)
			audioSamples = append(audioSamples, float32(sample)/32768.0)
		}
		return audioSamples, nil
	}

	return nil, fmt.Errorf("不支持的音频格式: %s", format)
}

func (rs *RealtimeSession) sendResult(text string, isFinal bool) {
	response := TranscribeResponse{
		Success: true,
		Result: &transcribe.TranscriptionResult{
			Text: text,
		},
	}

	// 添加最终结果标识
	if isFinal {
		response.Result.Text += " [FINAL]"
	}

	rs.conn.WriteJSON(response)
}

func (rs *RealtimeSession) sendError(message string) {
	response := TranscribeResponse{
		Success: false,
		Error:   message,
	}
	rs.conn.WriteJSON(response)
}

func (rs *RealtimeSession) cleanup() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.isActive = false

	if rs.stream != nil {
		sherpa_onnx.DeleteOnlineStream(rs.stream)
		rs.stream = nil
	}

	if rs.conn != nil {
		rs.conn.Close()
	}

	rs.logger.Info("实时转录会话已清理")
}

func (s *Server) Start() error {
	cfg := config.AppConfig.Server

	if cfg.UseUnixSocket {
		// 删除已存在的 socket 文件
		os.Remove(cfg.UnixSocket)

		// 创建 Unix socket
		listener, err := net.Listen("unix", cfg.UnixSocket)
		if err != nil {
			return fmt.Errorf("无法创建 Unix socket: %v", err)
		}
		defer listener.Close()

		s.logger.Infof("服务器启动在 Unix socket: %s", cfg.UnixSocket)
		return http.Serve(listener, s.router)
	} else {
		// 使用 TCP 端口
		addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		s.logger.Infof("服务器启动在: %s", addr)
		return s.router.Run(addr)
	}
}
