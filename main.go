package main

import (
	"flag"

	"github.com/layzdonw/transerver/config"
	"github.com/layzdonw/transerver/server"
	"github.com/layzdonw/transerver/transcribe"
	"github.com/sirupsen/logrus"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	if err := config.LoadConfig(*configPath); err != nil {
		logrus.Fatalf("加载配置失败: %v", err)
	}

	// 创建转录器
	cfg := config.AppConfig.Sherpa
	var transcriber *transcribe.SherpaTranscriber

	if cfg.EnableDiarization {
		logrus.Info("启用说话人分离功能")
		transcriber = transcribe.NewSherpaTranscriberWithDiarization(
			cfg.ModelPath,
			cfg.TokensPath,
			cfg.DiarizationModelPath,
			cfg.SampleRate,
			cfg.NumThreads,
			cfg.DecodingMethod,
		)
	} else {
		logrus.Info("使用标准转录功能")
		transcriber = transcribe.NewSherpaTranscriber(
			cfg.ModelPath,
			cfg.TokensPath,
			cfg.SampleRate,
			cfg.NumThreads,
			cfg.DecodingMethod,
		)
	}

	if transcriber == nil {
		logrus.Fatalf("创建转录器失败")
	}

	// 创建服务器
	srv := server.NewServer(transcriber)

	// 启动服务器
	logrus.Info("启动转录服务器...")
	if err := srv.Start(); err != nil {
		logrus.Fatalf("服务器启动失败: %v", err)
	}
}
