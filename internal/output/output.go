package output

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	err    error
)

func InitLogger(
	ioFunc func() []string,
	EncodeFunc func() string,
	appModeFunc func() (zap.AtomicLevel, bool)) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,                   // 大写编码器
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"), // 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 编码器
	}

	// 设置日志级别
	level, dev := appModeFunc()
	config := zap.Config{
		Level:            level,                    // 日志级别
		Development:      dev,                      // 开发模式，堆栈跟踪
		Encoding:         EncodeFunc(),             // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,            // 编码器配置
		InitialFields:    map[string]interface{}{}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      ioFunc(),                 // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}

	// 构建日志
	logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}
	logger.Info("初始化成功")
}

func GetLogger() *zap.Logger {
	return logger
}
