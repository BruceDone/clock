package config

import (
	"os"

	"github.com/BurntSushi/toml"

	"clock/internal/logger"
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
	Backend         string `toml:"backend"` // sqlite3, mysql, postgres
	Conn            string `toml:"conn"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxLifetime int    `toml:"conn_max_lifetime"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `toml:"level"`       // 日志级别: debug, info, warn, error
	FilePath   string `toml:"filepath"`    // 日志文件路径
	MaxSize    int    `toml:"max_size"`    // 单个日志文件最大大小（MB），默认 100MB
	MaxBackups int    `toml:"max_backups"` // 保留的旧日志文件数量，默认 5
	MaxAge     int    `toml:"max_age"`     // 日志文件保留天数，默认 7 天
	Compress   bool   `toml:"compress"`    // 是否压缩旧日志文件，默认 true
	LocalTime  bool   `toml:"local_time"`  // 是否使用本地时间，默认 true
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

	// 环境变量覆盖
	if jwtSecret := os.Getenv("CLOCK_JWT_SECRET"); jwtSecret != "" {
		cfg.Auth.JWTSecret = jwtSecret
	}
	if user := os.Getenv("CLOCK_USER"); user != "" {
		cfg.Auth.User = user
	}
	if password := os.Getenv("CLOCK_PASSWORD"); password != "" {
		cfg.Auth.Password = password
	}

	// 初始化日志
	logCfg := &logger.Config{
		Level:      cfg.Log.Level,
		FilePath:   cfg.Log.FilePath,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
		LocalTime:  cfg.Log.LocalTime,
	}
	if err := logger.Init(logCfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
