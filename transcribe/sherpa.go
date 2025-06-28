package transcribe

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
	"github.com/sirupsen/logrus"
)

// 由于 sherpa-onnx-go 可能不在标准库中，我们先使用一个简化的实现
// 在实际使用中，您需要安装 sherpa-onnx-go 库

type SherpaTranscriber struct {
	recognizer *sherpa_onnx.OnlineRecognizer
	logger     *logrus.Logger
	config     *sherpa_onnx.OnlineRecognizerConfig
	// 添加说话人分离相关字段
	diarizationEnabled   bool
	diarizationModelPath string
}

// 说话人分离结果结构体
type SpeakerSegment struct {
	SpeakerID int     `json:"speaker_id"`
	Start     float64 `json:"start"`
	End       float64 `json:"end"`
	Text      string  `json:"text"`
}

type TranscriptionResult struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence,omitempty"`
	Duration   float64 `json:"duration,omitempty"`
	// 添加说话人分离结果
	SpeakerSegments []SpeakerSegment `json:"speaker_segments,omitempty"`
}

type SherpaRequest struct {
	AudioData []byte `json:"audio_data"`
	Format    string `json:"format"`
}

// 新增：支持说话人分离的构造函数
func NewSherpaTranscriberWithDiarization(modelPath, tokensPath, diarizationModelPath string, sampleRate, numThreads int, decodingMethod string) *SherpaTranscriber {
	logger := logrus.New()

	// 创建 sherpa-onnx 配置
	config := &sherpa_onnx.OnlineRecognizerConfig{}

	// 设置特征配置
	config.FeatConfig.SampleRate = sampleRate
	config.FeatConfig.FeatureDim = 80

	// 设置模型配置
	config.ModelConfig.Transducer.Encoder = filepath.Join(modelPath, "encoder.onnx")
	config.ModelConfig.Transducer.Decoder = filepath.Join(modelPath, "decoder.onnx")
	config.ModelConfig.Transducer.Joiner = filepath.Join(modelPath, "joiner.onnx")
	config.ModelConfig.Tokens = tokensPath
	config.ModelConfig.NumThreads = numThreads
	config.ModelConfig.Provider = "cpu"

	// 设置识别器配置
	config.DecodingMethod = decodingMethod
	config.EnableEndpoint = 1
	config.Rule1MinTrailingSilence = 2.4
	config.Rule2MinTrailingSilence = 1.2
	config.Rule3MinUtteranceLength = 300

	// 创建识别器
	recognizer := sherpa_onnx.NewOnlineRecognizer(config)
	if recognizer == nil {
		logger.Errorf("创建识别器失败")
		return nil
	}

	return &SherpaTranscriber{
		recognizer:           recognizer,
		logger:               logger,
		config:               config,
		diarizationEnabled:   true,
		diarizationModelPath: diarizationModelPath,
	}
}

func NewSherpaTranscriber(modelPath, tokensPath string, sampleRate, numThreads int, decodingMethod string) *SherpaTranscriber {
	logger := logrus.New()

	// 创建 sherpa-onnx 配置
	config := &sherpa_onnx.OnlineRecognizerConfig{}

	// 设置特征配置
	config.FeatConfig.SampleRate = sampleRate
	config.FeatConfig.FeatureDim = 80

	// 设置模型配置
	config.ModelConfig.Transducer.Encoder = filepath.Join(modelPath, "encoder.onnx")
	config.ModelConfig.Transducer.Decoder = filepath.Join(modelPath, "decoder.onnx")
	config.ModelConfig.Transducer.Joiner = filepath.Join(modelPath, "joiner.onnx")
	config.ModelConfig.Tokens = tokensPath
	config.ModelConfig.NumThreads = numThreads
	config.ModelConfig.Provider = "cpu"

	// 设置识别器配置
	config.DecodingMethod = decodingMethod
	config.EnableEndpoint = 1
	config.Rule1MinTrailingSilence = 2.4
	config.Rule2MinTrailingSilence = 1.2
	config.Rule3MinUtteranceLength = 300

	// 创建识别器
	recognizer := sherpa_onnx.NewOnlineRecognizer(config)
	if recognizer == nil {
		logger.Errorf("创建识别器失败")
		return nil
	}

	return &SherpaTranscriber{
		recognizer:           recognizer,
		logger:               logger,
		config:               config,
		diarizationEnabled:   false,
		diarizationModelPath: "",
	}
}

// 新增：带说话人分离的转录方法
func (st *SherpaTranscriber) TranscribeAudioWithDiarization(audioData []byte, format string) (*TranscriptionResult, error) {
	if !st.diarizationEnabled {
		return nil, fmt.Errorf("说话人分离功能未启用")
	}

	// 处理音频数据
	audioSamples, err := st.processAudioData(audioData, format)
	if err != nil {
		return nil, fmt.Errorf("处理音频数据失败: %v", err)
	}

	// 创建说话人分离器（简化实现）
	// 注意：这里需要根据实际的 sherpa-onnx-go API 来调整
	// 当前实现为占位符，实际使用时需要参考官方示例
	diarizer, err := st.createSpeakerDiarizer()
	if err != nil {
		return nil, fmt.Errorf("创建说话人分离器失败: %v", err)
	}
	defer st.cleanupSpeakerDiarizer(diarizer)

	// 执行说话人分离
	speakerSegments, err := st.performDiarization(diarizer, audioSamples)
	if err != nil {
		return nil, fmt.Errorf("说话人分离计算失败: %v", err)
	}

	// 创建流进行语音识别
	stream := sherpa_onnx.NewOnlineStream(st.recognizer)
	if stream == nil {
		return nil, fmt.Errorf("创建音频流失败")
	}
	defer sherpa_onnx.DeleteOnlineStream(stream)

	// 将音频数据输入到流中
	stream.AcceptWaveform(st.config.FeatConfig.SampleRate, audioSamples)
	stream.InputFinished()

	// 获取识别结果
	recognitionResult := st.recognizer.GetResult(stream)

	return &TranscriptionResult{
		Text:            strings.TrimSpace(recognitionResult.Text),
		Confidence:      0.95,
		Duration:        float64(len(audioSamples)) / float64(st.config.FeatConfig.SampleRate),
		SpeakerSegments: speakerSegments,
	}, nil
}

// 创建说话人分离器（占位符实现）
func (st *SherpaTranscriber) createSpeakerDiarizer() (interface{}, error) {
	// 这里需要根据实际的 sherpa-onnx-go API 来实现
	// 参考：https://github.com/k2-fsa/sherpa-onnx/blob/master/go-api-examples/non-streaming-speaker-diarization/main.go
	st.logger.Info("创建说话人分离器（功能待实现）")
	return nil, nil
}

// 清理说话人分离器（占位符实现）
func (st *SherpaTranscriber) cleanupSpeakerDiarizer(diarizer interface{}) {
	// 这里需要根据实际的 sherpa-onnx-go API 来实现
	st.logger.Info("清理说话人分离器（功能待实现）")
}

// 执行说话人分离（占位符实现）
func (st *SherpaTranscriber) performDiarization(diarizer interface{}, audioSamples []float32) ([]SpeakerSegment, error) {
	// 这里需要根据实际的 sherpa-onnx-go API 来实现
	// 当前返回空的说话人片段
	st.logger.Info("执行说话人分离（功能待实现）")
	return []SpeakerSegment{}, nil
}

func (st *SherpaTranscriber) TranscribeAudio(audioData []byte, format string) (*TranscriptionResult, error) {
	if st.recognizer == nil {
		return nil, fmt.Errorf("识别器未初始化")
	}

	// 如果启用了说话人分离，使用带说话人分离的方法
	if st.diarizationEnabled {
		return st.TranscribeAudioWithDiarization(audioData, format)
	}

	// 创建流
	stream := sherpa_onnx.NewOnlineStream(st.recognizer)
	if stream == nil {
		return nil, fmt.Errorf("创建音频流失败")
	}
	defer sherpa_onnx.DeleteOnlineStream(stream)

	// 处理音频数据
	audioSamples, err := st.processAudioData(audioData, format)
	if err != nil {
		return nil, fmt.Errorf("处理音频数据失败: %v", err)
	}

	// 将音频数据输入到流中
	stream.AcceptWaveform(st.config.FeatConfig.SampleRate, audioSamples)

	// 标记输入结束
	stream.InputFinished()

	// 获取识别结果
	result := st.recognizer.GetResult(stream)

	return &TranscriptionResult{
		Text:       strings.TrimSpace(result.Text),
		Confidence: 0.95, // sherpa-onnx 可能不提供置信度
		Duration:   float64(len(audioSamples)) / float64(st.config.FeatConfig.SampleRate),
	}, nil
}

func (st *SherpaTranscriber) TranscribeStream(audioData []byte) (*TranscriptionResult, error) {
	// 流式转录实现
	return st.TranscribeAudio(audioData, "wav")
}

func (st *SherpaTranscriber) processAudioData(audioData []byte, format string) ([]float32, error) {
	switch strings.ToLower(format) {
	case "wav":
		return st.processWavData(audioData)
	case "pcm":
		return st.processPcmData(audioData)
	default:
		return nil, fmt.Errorf("不支持的音频格式: %s", format)
	}
}

func (st *SherpaTranscriber) processWavData(audioData []byte) ([]float32, error) {
	// 简单的 WAV 文件处理
	// 这里假设是 16-bit PCM WAV 文件
	if len(audioData) < 44 {
		return nil, fmt.Errorf("音频数据太短，不是有效的 WAV 文件")
	}

	// 跳过 WAV 头部（44字节）
	audioSamples := make([]float32, 0, (len(audioData)-44)/2)

	for i := 44; i < len(audioData)-1; i += 2 {
		// 将 16-bit 整数转换为 float32
		sample := int16(binary.LittleEndian.Uint16(audioData[i : i+2]))
		audioSamples = append(audioSamples, float32(sample)/32768.0)
	}

	return audioSamples, nil
}

func (st *SherpaTranscriber) processPcmData(audioData []byte) ([]float32, error) {
	// 处理原始 PCM 数据
	audioSamples := make([]float32, 0, len(audioData)/2)

	for i := 0; i < len(audioData)-1; i += 2 {
		sample := int16(binary.LittleEndian.Uint16(audioData[i : i+2]))
		audioSamples = append(audioSamples, float32(sample)/32768.0)
	}

	return audioSamples, nil
}

func (st *SherpaTranscriber) Close() error {
	if st.recognizer != nil {
		sherpa_onnx.DeleteOnlineRecognizer(st.recognizer)
	}
	return nil
}

// 添加获取识别器的方法
func (st *SherpaTranscriber) GetRecognizer() *sherpa_onnx.OnlineRecognizer {
	return st.recognizer
}

// 添加获取采样率的方法
func (st *SherpaTranscriber) GetSampleRate() int {
	return st.config.FeatConfig.SampleRate
}
