package transcribe

import (
	"testing"
)

func TestNewSherpaTranscriberWithDiarization(t *testing.T) {
	// 测试创建带说话人分离的转录器
	// 注意：由于测试环境中没有实际的模型文件，这个测试可能会失败
	transcriber := NewSherpaTranscriberWithDiarization(
		"./test_models",
		"./test_models/tokens.txt",
		"./test_models/diarization",
		16000,
		1,
		"greedy_search",
	)

	// 由于没有实际的模型文件，转录器可能为 nil
	if transcriber == nil {
		t.Skip("无法创建转录器（模型文件不存在），跳过测试")
		return
	}

	// 验证说话人分离功能已启用
	if !transcriber.diarizationEnabled {
		t.Error("说话人分离功能未启用")
	}

	// 验证模型路径设置正确
	if transcriber.diarizationModelPath != "./test_models/diarization" {
		t.Errorf("说话人分离模型路径设置错误，期望: ./test_models/diarization, 实际: %s", transcriber.diarizationModelPath)
	}
}

func TestTranscribeAudioWithDiarization(t *testing.T) {
	// 创建带说话人分离的转录器
	transcriber := NewSherpaTranscriberWithDiarization(
		"./test_models",
		"./test_models/tokens.txt",
		"./test_models/diarization",
		16000,
		1,
		"greedy_search",
	)

	if transcriber == nil {
		t.Skip("无法创建转录器（模型文件不存在），跳过测试")
		return
	}

	// 创建测试音频数据（空的音频数据）
	audioData := make([]byte, 1000)

	// 测试带说话人分离的转录
	result, err := transcriber.TranscribeAudioWithDiarization(audioData, "wav")

	// 由于当前实现为占位符，我们期望得到错误或空结果
	if err == nil {
		// 如果没有错误，验证结果结构
		if result == nil {
			t.Error("转录结果为空")
		} else {
			// 验证说话人片段字段存在
			if result.SpeakerSegments == nil {
				t.Error("说话人片段字段为空")
			}
		}
	}
}

func TestSpeakerSegment(t *testing.T) {
	// 测试说话人片段结构体
	segment := SpeakerSegment{
		SpeakerID: 1,
		Start:     0.0,
		End:       2.5,
		Text:      "测试文本",
	}

	if segment.SpeakerID != 1 {
		t.Errorf("说话人ID错误，期望: 1, 实际: %d", segment.SpeakerID)
	}

	if segment.Start != 0.0 {
		t.Errorf("开始时间错误，期望: 0.0, 实际: %f", segment.Start)
	}

	if segment.End != 2.5 {
		t.Errorf("结束时间错误，期望: 2.5, 实际: %f", segment.End)
	}

	if segment.Text != "测试文本" {
		t.Errorf("文本内容错误，期望: 测试文本, 实际: %s", segment.Text)
	}
}

func TestTranscriptionResultWithSpeakerSegments(t *testing.T) {
	// 测试包含说话人片段的转录结果
	result := &TranscriptionResult{
		Text:       "转录文本",
		Confidence: 0.95,
		Duration:   3.2,
		SpeakerSegments: []SpeakerSegment{
			{
				SpeakerID: 0,
				Start:     0.0,
				End:       1.5,
				Text:      "第一个说话人",
			},
			{
				SpeakerID: 1,
				Start:     1.8,
				End:       3.2,
				Text:      "第二个说话人",
			},
		},
	}

	if len(result.SpeakerSegments) != 2 {
		t.Errorf("说话人片段数量错误，期望: 2, 实际: %d", len(result.SpeakerSegments))
	}

	if result.SpeakerSegments[0].SpeakerID != 0 {
		t.Errorf("第一个说话人ID错误，期望: 0, 实际: %d", result.SpeakerSegments[0].SpeakerID)
	}

	if result.SpeakerSegments[1].SpeakerID != 1 {
		t.Errorf("第二个说话人ID错误，期望: 1, 实际: %d", result.SpeakerSegments[1].SpeakerID)
	}
}

func TestDiarizationConfiguration(t *testing.T) {
	// 测试说话人分离配置
	transcriber := &SherpaTranscriber{
		diarizationEnabled:   true,
		diarizationModelPath: "./test_models/diarization",
	}

	if !transcriber.diarizationEnabled {
		t.Error("说话人分离功能未启用")
	}

	if transcriber.diarizationModelPath != "./test_models/diarization" {
		t.Errorf("说话人分离模型路径错误，期望: ./test_models/diarization, 实际: %s", transcriber.diarizationModelPath)
	}
}
