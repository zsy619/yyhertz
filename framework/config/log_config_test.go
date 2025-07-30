package config

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLogConfig(t *testing.T) {
	config := DefaultLogConfig()

	assert.NotNil(t, config)
	assert.Equal(t, LogLevelInfo, config.Level)
	assert.Equal(t, LogFormatJSON, config.Format)
	assert.True(t, config.EnableConsole)
	assert.True(t, config.EnableFile)
	assert.Equal(t, "./logs/app.log", config.FilePath)
	assert.True(t, config.ShowCaller)
	assert.True(t, config.ShowTimestamp)
}

func TestCreateLogger(t *testing.T) {
	config := DefaultLogConfig()
	logger := config.CreateLogger()

	assert.NotNil(t, logger)
	assert.NotNil(t, logger.Logger())
}

func TestLogLevels(t *testing.T) {
	testCases := []struct {
		configLevel   LogLevel
		expectedLevel logrus.Level
	}{
		{LogLevelDebug, logrus.DebugLevel},
		{LogLevelInfo, logrus.InfoLevel},
		{LogLevelWarn, logrus.WarnLevel},
		{LogLevelError, logrus.ErrorLevel},
		{LogLevelFatal, logrus.FatalLevel},
		{LogLevelPanic, logrus.PanicLevel},
	}

	for _, tc := range testCases {
		t.Run(string(tc.configLevel), func(t *testing.T) {
			config := &LogConfig{
				Level:           tc.configLevel,
				Format:          LogFormatText,
				EnableConsole:   true,
				EnableFile:      false,
				ShowCaller:      false,
				ShowTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
				Fields:          make(map[string]any),
			}

			logger := config.CreateLogger()
			assert.Equal(t, tc.expectedLevel, logger.Logger().Level)
		})
	}
}

func TestLogFormats(t *testing.T) {
	t.Run("JSON格式", func(t *testing.T) {
		config := &LogConfig{
			Level:           LogLevelInfo,
			Format:          LogFormatJSON,
			EnableConsole:   true,
			EnableFile:      false,
			ShowCaller:      false,
			ShowTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			Fields:          make(map[string]any),
		}

		logger := config.CreateLogger()
		_, ok := logger.Logger().Formatter.(*logrus.JSONFormatter)
		assert.True(t, ok)
	})

	t.Run("文本格式", func(t *testing.T) {
		config := &LogConfig{
			Level:           LogLevelInfo,
			Format:          LogFormatText,
			EnableConsole:   true,
			EnableFile:      false,
			ShowCaller:      false,
			ShowTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			Fields:          make(map[string]any),
		}

		logger := config.CreateLogger()
		_, ok := logger.Logger().Formatter.(*logrus.TextFormatter)
		assert.True(t, ok)
	})
}

func TestFileLogging(t *testing.T) {
	// 创建临时日志文件路径
	tmpDir := t.TempDir()
	logFile := tmpDir + "/test.log"

	config := &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatJSON,
		EnableConsole:   false,
		EnableFile:      true,
		FilePath:        logFile,
		ShowCaller:      false,
		ShowTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		Fields:          make(map[string]any),
	}

	logger := config.CreateLogger()
	assert.NotNil(t, logger)

	// 验证日志文件是否被创建
	_, err := os.Stat(logFile)
	assert.NoError(t, err)
}

func TestLoggerWithFields(t *testing.T) {
	config := DefaultLogConfig()
	config.Fields = map[string]any{
		"service": "test-service",
		"version": "1.0.0",
	}

	logger := config.CreateLogger()
	assert.NotNil(t, logger)

	// 注意：这里不能直接测试字段是否被设置，因为logrus的WithField会返回新的Entry
	// 但我们可以确保CreateLogger不会panic
}

func TestHelperFunctions(t *testing.T) {
	logger := logrus.New()

	t.Run("LoggerWithRequestID", func(t *testing.T) {
		entry := LoggerWithRequestID(logger, "test-request-id")
		assert.NotNil(t, entry)
		assert.Equal(t, "test-request-id", entry.Data["request_id"])
	})

	t.Run("LoggerWithUserID", func(t *testing.T) {
		entry := LoggerWithUserID(logger, "test-user-id")
		assert.NotNil(t, entry)
		assert.Equal(t, "test-user-id", entry.Data["user_id"])
	})

	t.Run("LoggerWithFields", func(t *testing.T) {
		fields := map[string]any{
			"key1": "value1",
			"key2": 123,
		}
		entry := LoggerWithFields(logger, fields)
		assert.NotNil(t, entry)
		assert.Equal(t, "value1", entry.Data["key1"])
		assert.Equal(t, 123, entry.Data["key2"])
	})
}
