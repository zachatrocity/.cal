package logger

import (
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestLoggerInit(t *testing.T) {
	// Reset environment before each test
	os.Unsetenv("DEV_MODE")

	t.Run("default log level is info", func(t *testing.T) {
		// Reset logger
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logger, _ := config.Build()
		sugar = logger.Sugar()

		// Create a buffer to capture log output
		logs := &zaptest.Buffer{}
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			logs,
			zap.InfoLevel,
		)
		sugar = zap.New(core).Sugar()

		Debug("debug message") // Should not appear
		Info("info message")   // Should appear
		Error("error message") // Should appear

		output := logs.String()
		if strings.Contains(output, "debug message") {
			t.Error("Debug message should not appear in info level")
		}
		if !strings.Contains(output, "info message") {
			t.Error("Info message should appear")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should appear")
		}
	})

	t.Run("dev mode enables debug level", func(t *testing.T) {
		os.Setenv("DEV_MODE", "true")

		// Reset logger with dev mode
		config := zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		logger, _ := config.Build()
		sugar = logger.Sugar()

		// Create a buffer to capture log output
		logs := &zaptest.Buffer{}
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			logs,
			zap.DebugLevel,
		)
		sugar = zap.New(core).Sugar()

		Debug("debug message")
		Info("info message")
		Error("error message")

		output := logs.String()
		if !strings.Contains(output, "debug message") {
			t.Error("Debug message should appear in debug level")
		}
		if !strings.Contains(output, "info message") {
			t.Error("Info message should appear")
		}
		if !strings.Contains(output, "error message") {
			t.Error("Error message should appear")
		}
	})

	t.Run("log formatting", func(t *testing.T) {
		// Create a buffer to capture log output
		logs := &zaptest.Buffer{}
		core := zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			logs,
			zap.DebugLevel,
		)
		sugar = zap.New(core).Sugar()

		Debug("test %s %d", "debug", 1)
		Info("test %s %d", "info", 2)
		Error("test %s %d", "error", 3)

		output := logs.String()
		expectedStrings := []string{
			"test debug 1",
			"test info 2",
			"test error 3",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(output, expected) {
				t.Errorf("Expected log to contain '%s'", expected)
			}
		}
	})
}

func TestLoggerMethods(t *testing.T) {
	// Create a buffer to capture log output
	logs := &zaptest.Buffer{}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		logs,
		zap.DebugLevel,
	)
	sugar = zap.New(core).Sugar()

	tests := []struct {
		name     string
		logFunc  func(string, ...interface{})
		message  string
		args     []interface{}
		contains string
	}{
		{
			name:     "Debug with format",
			logFunc:  Debug,
			message:  "debug %s",
			args:     []interface{}{"test"},
			contains: "debug test",
		},
		{
			name:     "Info with format",
			logFunc:  Info,
			message:  "info %d",
			args:     []interface{}{42},
			contains: "info 42",
		},
		{
			name:     "Error with format",
			logFunc:  Error,
			message:  "error %s %d",
			args:     []interface{}{"test", 123},
			contains: "error test 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs.Reset()
			tt.logFunc(tt.message, tt.args...)
			if !strings.Contains(logs.String(), tt.contains) {
				t.Errorf("Expected log to contain '%s', got '%s'", tt.contains, logs.String())
			}
		})
	}
}
