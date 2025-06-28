package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type TranscribeResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Text string `json:"text"`
	} `json:"result,omitempty"`
	Error string `json:"error,omitempty"`
}

func main() {
	serverURL := "http://localhost:8080"

	// 示例1: 健康检查
	fmt.Println("=== 健康检查 ===")
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		fmt.Printf("健康检查失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("健康检查响应: %s\n\n", string(body))

	// 示例2: 文件上传转录
	fmt.Println("=== 文件上传转录 ===")
	if len(os.Args) > 1 {
		audioFile := os.Args[1]
		result, err := uploadAndTranscribe(serverURL, audioFile)
		if err != nil {
			fmt.Printf("转录失败: %v\n", err)
		} else {
			fmt.Printf("转录结果: %s\n", result)
		}
	} else {
		fmt.Println("请提供音频文件路径作为参数")
	}
}

func uploadAndTranscribe(serverURL, audioFile string) (string, error) {
	// 打开音频文件
	file, err := os.Open(audioFile)
	if err != nil {
		return "", fmt.Errorf("无法打开音频文件: %v", err)
	}
	defer file.Close()

	// 创建 multipart 请求
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加音频文件
	part, err := writer.CreateFormFile("audio", "audio.wav")
	if err != nil {
		return "", fmt.Errorf("创建表单文件失败: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return "", fmt.Errorf("复制文件内容失败: %v", err)
	}

	writer.Close()

	// 发送请求
	resp, err := http.Post(serverURL+"/transcribe", writer.FormDataContentType(), &buf)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var transcribeResp TranscribeResponse
	if err := json.NewDecoder(resp.Body).Decode(&transcribeResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if !transcribeResp.Success {
		return "", fmt.Errorf("转录失败: %s", transcribeResp.Error)
	}

	return transcribeResp.Result.Text, nil
}
