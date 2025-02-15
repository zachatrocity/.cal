package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sugar *zap.SugaredLogger

func init() {
	// Configure zap logger
	config := zap.NewDevelopmentConfig()

	// Set log level based on DEV_MODE
	if os.Getenv("DEV_MODE") == "true" {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	// Configure output format
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create logger
	logger, _ := config.Build()
	sugar = logger.Sugar()
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	sugar.Debugf(format, v...)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	sugar.Errorf(format, v...)
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	sugar.Infof(format, v...)
}
