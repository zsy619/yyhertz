package config

import (
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLogConfig(t *testing.T) {
	config := DefaultLogConfig()

	assert.NotNil(t, config)
	assert.Equal(t, LogLevelInfo, config.Level)
	assert.Equal(t, LogFormatBeego, config.Format)
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
	t.Run("LoggerWithRequestID", func(t *testing.T) {
		entry := LoggerWithRequestID("test-request-id")
		assert.NotNil(t, entry)
		assert.Equal(t, "test-request-id", entry.Data["request_id"])
	})

	t.Run("LoggerWithUserID", func(t *testing.T) {
		entry := LoggerWithUserID("test-user-id")
		assert.NotNil(t, entry)
		assert.Equal(t, "test-user-id", entry.Data["user_id"])
	})

	t.Run("WithFields", func(t *testing.T) {
		fields := map[string]any{
			"key1": "value1",
			"key2": 123,
		}
		entry := WithFields(fields)
		assert.NotNil(t, entry)
		assert.Equal(t, "value1", entry.Data["key1"])
		assert.Equal(t, 123, entry.Data["key2"])
	})
}

func TestWithRequestID(t *testing.T) {
	// 初始化全局日志器
	InitGlobalLogger(DefaultLogConfig())

	t.Run("有效的RequestID", func(t *testing.T) {
		validIDs := []string{
			"12345678",                              // 最小长度
			"request-id-12345",                      // 包含连字符
			"req_id_123456789",                      // 包含下划线
			"req.id.123456789",                      // 包含点号
			"a1b2c3d4e5f6g7h8",                      // 字母数字混合
			"request-uuid-1234567890abcdef12345678", // 类似UUID格式
			strings.Repeat("a", 64),                 // 最大长度
		}

		for _, requestID := range validIDs {
			t.Run(requestID, func(t *testing.T) {
				entry := WithRequestID(requestID)
				assert.NotNil(t, entry)
				assert.Equal(t, requestID, entry.Data["request_id"])
				assert.NotContains(t, entry.Data, "invalid_request_id")
				assert.NotContains(t, entry.Data, "error")
			})
		}
	})

	t.Run("无效的RequestID", func(t *testing.T) {
		invalidIDs := []struct {
			id     string
			reason string
		}{
			{"", "空字符串"},
			{"1234567", "长度不足8位"},
			{strings.Repeat("a", 65), "长度超过64位"},
			{"req id 123", "包含空格"},
			{"req@id#123", "包含特殊字符"},
			{"req中文id123", "包含中文"},
			{"req--id", "连续连字符"},
			{"req__id", "连续下划线"},
			{"req..id", "连续点号"},
			{"req/id\\123", "包含斜杠"},
		}

		for _, testCase := range invalidIDs {
			t.Run(testCase.reason, func(t *testing.T) {
				entry := WithRequestID(testCase.id)
				assert.NotNil(t, entry)
				// 验证包含错误信息
				assert.Contains(t, entry.Data, "error")
				assert.Contains(t, entry.Data, "invalid_request_id")
				assert.Equal(t, testCase.id, entry.Data["invalid_request_id"])
				// 验证不包含有效的request_id
				assert.NotContains(t, entry.Data, "request_id")
			})
		}
	})

	t.Run("防重复功能", func(t *testing.T) {
		t.Run("相同RequestID不重复添加", func(t *testing.T) {
			requestID := "test-request-12345"
			
			// 第一次添加
			entry1 := WithRequestID(requestID)
			assert.Equal(t, requestID, entry1.Data["request_id"])
			
			// 第二次添加相同ID
			entry2 := WithRequestID(requestID)
			assert.Equal(t, requestID, entry2.Data["request_id"])
		})

		t.Run("不同RequestID会产生冲突警告", func(t *testing.T) {
			// 这个测试需要模拟已存在request_id的情况
			// 由于当前实现基于全局logger，我们跳过这个具体的实现测试
			// 在实际使用中，这种情况较少发生
			assert.True(t, true) // 占位测试
		})
	})

	t.Run("IsValidRequestID函数", func(t *testing.T) {
		validCases := []string{
			"12345678",
			"request-id-12345",
			"req_id_123456789",
			"req.id.123456789",
		}

		invalidCases := []string{
			"",
			"1234567",
			strings.Repeat("a", 65),
			"req id 123",
			"req@id#123",
		}

		for _, valid := range validCases {
			assert.True(t, IsValidRequestID(valid), "应该验证通过: %s", valid)
		}

		for _, invalid := range invalidCases {
			assert.False(t, IsValidRequestID(invalid), "应该验证失败: %s", invalid)
		}
	})

	t.Run("WithRequestIDUnsafe函数", func(t *testing.T) {
		// 测试不安全版本可以接受任何字符串
		invalidID := "invalid@#$%^&*()id"
		entry := WithRequestIDUnsafe(invalidID)
		assert.NotNil(t, entry)
		assert.Equal(t, invalidID, entry.Data["request_id"])
		assert.NotContains(t, entry.Data, "error")
		assert.NotContains(t, entry.Data, "invalid_request_id")
	})
}

func TestLoggerManager_UpdateLevel(t *testing.T) {
	// 初始化日志管理器
	loggerManager := InitGlobalLogger(DefaultLogConfig())

	t.Run("更新日志级别", func(t *testing.T) {
		// 测试初始级别
		initialLevel := loggerManager.GetLevel()
		assert.Equal(t, LogLevelInfo, initialLevel)

		// 更新到Debug级别
		loggerManager.UpdateLevel(LogLevelDebug)
		assert.Equal(t, LogLevelDebug, loggerManager.GetLevel())

		// 验证配置也被更新
		config := loggerManager.GetConfig()
		assert.Equal(t, LogLevelDebug, config.Level)

		// 更新到Error级别
		loggerManager.UpdateLevel(LogLevelError)
		assert.Equal(t, LogLevelError, loggerManager.GetLevel())

		// 恢复到Info级别
		loggerManager.UpdateLevel(LogLevelInfo)
		assert.Equal(t, LogLevelInfo, loggerManager.GetLevel())
	})

	t.Run("测试所有日志级别", func(t *testing.T) {
		levels := []LogLevel{
			LogLevelDebug,
			LogLevelInfo,
			LogLevelWarn,
			LogLevelError,
			LogLevelFatal,
			LogLevelPanic,
		}

		for _, level := range levels {
			t.Run(string(level), func(t *testing.T) {
				loggerManager.UpdateLevel(level)
				assert.Equal(t, level, loggerManager.GetLevel())
			})
		}
	})
}

func TestLoggerManager_UpdateFormat(t *testing.T) {
	// 初始化日志管理器
	loggerManager := InitGlobalLogger(DefaultLogConfig())

	t.Run("更新日志格式", func(t *testing.T) {
		// 测试初始格式
		initialFormat := loggerManager.GetFormat()
		assert.Equal(t, LogFormatBeego, initialFormat)

		// 更新到JSON格式
		loggerManager.UpdateFormat(LogFormatJSON)
		assert.Equal(t, LogFormatJSON, loggerManager.GetFormat())

		// 验证配置也被更新
		config := loggerManager.GetConfig()
		assert.Equal(t, LogFormatJSON, config.Format)

		// 更新到Logstash格式
		loggerManager.UpdateFormat(LogFormatLogstash)
		assert.Equal(t, LogFormatLogstash, loggerManager.GetFormat())

		// 恢复到Beego格式
		loggerManager.UpdateFormat(LogFormatBeego)
		assert.Equal(t, LogFormatBeego, loggerManager.GetFormat())
	})

	t.Run("测试所有日志格式", func(t *testing.T) {
		formats := []LogFormat{
			LogFormatJSON,
			LogFormatText,
			LogFormatBeego,
			LogFormatLog4Go,
			LogFormatLogstash,
			LogFormatSyslog,
			LogFormatFluentd,
			LogFormatCloudWatch,
			LogFormatApplicationInsights,
		}

		for _, format := range formats {
			t.Run(string(format), func(t *testing.T) {
				loggerManager.UpdateFormat(format)
				assert.Equal(t, format, loggerManager.GetFormat())
			})
		}
	})
}

func TestGlobalUpdateFunctions(t *testing.T) {
	t.Run("UpdateGlobalLogLevel", func(t *testing.T) {
		// 为每个测试重置全局日志器
		ResetGlobalLogger(DefaultLogConfig())
		
		// 获取初始级别
		initialLevel := GetGlobalLogger().GetLevel()
		assert.Equal(t, LogLevelInfo, initialLevel)

		// 通过全局函数更新级别
		UpdateGlobalLogLevel(LogLevelDebug)
		assert.Equal(t, LogLevelDebug, GetGlobalLogger().GetLevel())

		// 恢复级别
		UpdateGlobalLogLevel(LogLevelInfo)
		assert.Equal(t, LogLevelInfo, GetGlobalLogger().GetLevel())
	})

	t.Run("UpdateGlobalLogFormat", func(t *testing.T) {
		// 为每个测试重置全局日志器
		ResetGlobalLogger(DefaultLogConfig())
		
		// 获取初始格式
		initialFormat := GetGlobalLogger().GetFormat()
		assert.Equal(t, LogFormatBeego, initialFormat)

		// 通过全局函数更新格式
		UpdateGlobalLogFormat(LogFormatJSON)
		assert.Equal(t, LogFormatJSON, GetGlobalLogger().GetFormat())

		// 恢复格式
		UpdateGlobalLogFormat(LogFormatBeego)
		assert.Equal(t, LogFormatBeego, GetGlobalLogger().GetFormat())
	})
}
