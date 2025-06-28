package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Sherpa SherpaConfig `mapstructure:"sherpa"`
}

type ServerConfig struct {
	Port          int    `mapstructure:"port"`
	UnixSocket    string `mapstructure:"unix_socket"`
	UseUnixSocket bool   `mapstructure:"use_unix_socket"`
	Host          string `mapstructure:"host"`
}

type SherpaConfig struct {
	ModelPath            string `mapstructure:"model_path"`
	TokensPath           string `mapstructure:"tokens_path"`
	SampleRate           int    `mapstructure:"sample_rate"`
	NumThreads           int    `mapstructure:"num_threads"`
	DecodingMethod       string `mapstructure:"decoding_method"`
	EnableDiarization    bool   `mapstructure:"enable_diarization"`
	DiarizationModelPath string `mapstructure:"diarization_model_path"`
}

var AppConfig Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.use_unix_socket", false)
	viper.SetDefault("server.unix_socket", "/tmp/transcribe.sock")
	viper.SetDefault("sherpa.sample_rate", 16000)
	viper.SetDefault("sherpa.num_threads", 1)
	viper.SetDefault("sherpa.decoding_method", "greedy_search")
	viper.SetDefault("sherpa.enable_diarization", false)
	viper.SetDefault("sherpa.diarization_model_path", "")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("无法读取配置文件: %v", err)
		return err
	}

	if err := viper.Unmarshal(&AppConfig); err != nil {
		log.Printf("无法解析配置文件: %v", err)
		return err
	}

	return nil
}
