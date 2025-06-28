package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/layzdonw/transerver/transcribe"
)

func TestHealthCheck(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试服务器
	transcriber := transcribe.NewSherpaTranscriber("", "", 16000, 1, "greedy_search")
	srv := NewServer(transcriber)

	// 创建测试请求
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)

	// 执行请求
	srv.router.ServeHTTP(w, req)

	// 检查响应
	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 %d，得到 %d", http.StatusOK, w.Code)
	}

	// 检查响应内容
	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("无法解析响应 JSON: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("期望状态 'ok'，得到 '%v'", response["status"])
	}
}

func TestTranscribeHandlerInvalidJSON(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试服务器
	transcriber := transcribe.NewSherpaTranscriber("", "", 16000, 1, "greedy_search")
	srv := NewServer(transcriber)

	// 创建无效的 JSON 请求
	w := httptest.NewRecorder()
	reqBody := `{"invalid": "json", "audio_data": "not_base64"}`
	req, _ := http.NewRequest("POST", "/transcribe", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	// 执行请求
	srv.router.ServeHTTP(w, req)

	// 检查响应
	if w.Code != http.StatusBadRequest {
		t.Errorf("期望状态码 %d，得到 %d", http.StatusBadRequest, w.Code)
	}

	// 检查响应内容
	var response TranscribeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("无法解析响应 JSON: %v", err)
	}

	if response.Success {
		t.Error("期望请求失败，但得到了成功响应")
	}
}
