package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

// Config 应用配置
type Config struct {
	Server  ServerConfig  `toml:"server"`
	Storage StorageConfig `toml:"storage"`
	Log     LogConfig     `toml:"log"`
	Auth    AuthConfig    `toml:"auth"`
	Message MessageConfig `toml:"message"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `toml:"host"`
}

// StorageConfig 存储配置
type StorageConfig struct {
	Backend string `toml:"backend"` // sqlite3, mysql, postgres
	Conn    string `toml:"conn"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `toml:"level"`
	FilePath string `toml:"filepath"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	User      string `toml:"user"`
	Password  string `toml:"password"`
	JWTSecret string `toml:"jwt_secret"`
}

// MessageConfig 消息配置
type MessageConfig struct {
	Size int `toml:"size"`
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Message.Size <= 0 {
		cfg.Message.Size = 1000
	}

	// 配置日志级别
	configureLogger(&cfg.Log)

	return &cfg, nil
}

// configureLogger 配置日志
func configureLogger(cfg *LogConfig) {
	switch cfg.Level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	if cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			logrus.SetOutput(file)
		}
	}
}
