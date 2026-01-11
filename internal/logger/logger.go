package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// Logger 全局日志实例
	Logger *zap.Logger
	// Sugar 全局 Sugar 日志实例（更简单的 API）
	Sugar *zap.SugaredLogger
)

// Config 日志配置
type Config struct {
	Level      string `toml:"level"`       // 日志级别: debug, info, warn, error
	FilePath   string `toml:"filepath"`    // 日志文件路径
	MaxSize    int    `toml:"max_size"`    // 单个日志文件最大大小（MB），默认 100MB
	MaxBackups int    `toml:"max_backups"` // 保留的旧日志文件数量，默认 5
	MaxAge     int    `toml:"max_age"`     // 日志文件保留天数，默认 7 天
	Compress   bool   `toml:"compress"`    // 是否压缩旧日志文件，默认 true
	LocalTime  bool   `toml:"local_time"`  // 是否使用本地时间，默认 true
}

// Init 初始化日志
func Init(cfg *Config) error {
	// 设置默认值
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 // 默认 100MB
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 5 // 默认保留 5 个文件
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 7 // 默认保留 7 天
	}

	// 解析日志级别
	level := parseLevel(cfg.Level)

	// 创建编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// 创建核心写入器
	var cores []zapcore.Core

	// 控制台输出（开发环境）
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	cores = append(cores, consoleCore)

	// 文件输出（生产环境）
	if cfg.FilePath != "" {
		// 使用 lumberjack 实现日志轮转
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  cfg.LocalTime,
		}

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(fileWriter), level)
		cores = append(cores, fileCore)
	}

	// 合并所有核心
	core := zapcore.NewTee(cores...)

	// 创建 Logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Sugar = Logger.Sugar()

	return nil
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal 致命错误日志（会退出程序）
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	}
}

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Debugf(template, args...)
	}
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Infof(template, args...)
	}
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Warnf(template, args...)
	}
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Errorf(template, args...)
	}
}

// Fatalf 格式化致命错误日志（会退出程序）
func Fatalf(template string, args ...interface{}) {
	if Sugar != nil {
		Sugar.Fatalf(template, args...)
	}
}

// With 添加字段
func With(fields ...zap.Field) *zap.Logger {
	if Logger != nil {
		return Logger.With(fields...)
	}
	return zap.NewNop()
}

// WithFields 添加字段（Sugar 风格）
func WithFields(keysAndValues ...interface{}) *zap.SugaredLogger {
	if Sugar != nil {
		return Sugar.With(keysAndValues...)
	}
	return zap.NewNop().Sugar()
}
