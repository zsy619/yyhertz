package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLogConfig_DefaultConfig(t *testing.T) {
	config := DefaultLogConfig()

	assert.NotNil(t, config)
	assert.Equal(t, LogLevelInfo, config.Level)
	assert.Equal(t, LogFormatBeego, config.Format)
	assert.True(t, config.EnableConsole)
	assert.True(t, config.EnableFile)
	assert.Equal(t, "./logs/app.log", config.FilePath)
	assert.True(t, config.ShowCaller)
	assert.True(t, config.ShowTimestamp)
	assert.Equal(t, []string{"console", "file"}, config.Outputs)
	assert.NotNil(t, config.OutputConfig)
}

func TestLogConfig_PresetConfigs(t *testing.T) {
	t.Run("DevelopmentConfig", func(t *testing.T) {
		config := DevelopmentLogConfig()
		assert.Equal(t, LogLevelDebug, config.Level)
		assert.Equal(t, LogFormatBeego, config.Format)
		assert.True(t, config.EnableConsole)
		assert.True(t, config.EnableFile)
		assert.Contains(t, config.Fields, "env")
		assert.Equal(t, "development", config.Fields["env"])
	})

	t.Run("ProductionConfig", func(t *testing.T) {
		config := ProductionLogConfig()
		assert.Equal(t, LogLevelInfo, config.Level)
		assert.Equal(t, LogFormatLogstash, config.Format)
		assert.False(t, config.EnableConsole)
		assert.True(t, config.EnableFile)
		assert.Contains(t, config.Outputs, "fluentd")
		assert.Contains(t, config.OutputConfig, "fluentd")
	})

	t.Run("TestConfig", func(t *testing.T) {
		config := TestLogConfig()
		assert.Equal(t, LogLevelWarn, config.Level)
		assert.True(t, config.EnableConsole)
		assert.False(t, config.EnableFile)
		assert.Equal(t, []string{"console"}, config.Outputs)
	})

	t.Run("HighPerformanceConfig", func(t *testing.T) {
		config := HighPerformanceLogConfig()
		assert.Equal(t, LogLevelError, config.Level)
		assert.Equal(t, LogFormatJSON, config.Format)
		assert.False(t, config.EnableConsole)
		assert.True(t, config.EnableFile)
	})

	t.Run("CloudConfig", func(t *testing.T) {
		config := CloudLogConfig()
		assert.Equal(t, LogLevelInfo, config.Level)
		assert.Equal(t, LogFormatCloudWatch, config.Format)
		assert.Contains(t, config.Outputs, "cloudwatch")
		assert.Contains(t, config.Outputs, "azure_insights")
		assert.Contains(t, config.OutputConfig, "cloudwatch")
		assert.Contains(t, config.OutputConfig, "azure_insights")
	})
}

func TestLogConfig_ConfigMethods(t *testing.T) {
	config := DefaultLogConfig()

	t.Run("UpdateConfigLevel", func(t *testing.T) {
		newConfig := config.UpdateConfigLevel(LogLevelDebug)
		assert.Equal(t, LogLevelDebug, newConfig.Level)
		assert.Equal(t, LogLevelInfo, config.Level) // 原配置不变
	})

	t.Run("UpdateConfigFormat", func(t *testing.T) {
		newConfig := config.UpdateConfigFormat(LogFormatJSON)
		assert.Equal(t, LogFormatJSON, newConfig.Format)
		assert.Equal(t, LogFormatBeego, config.Format) // 原配置不变
	})

	t.Run("AddConfigFields", func(t *testing.T) {
		fields := map[string]any{
			"test_key": "test_value",
			"number":   123,
		}
		newConfig := config.AddConfigFields(fields)
		assert.Contains(t, newConfig.Fields, "test_key")
		assert.Equal(t, "test_value", newConfig.Fields["test_key"])
		assert.Contains(t, newConfig.Fields, "number")
		assert.Equal(t, 123, newConfig.Fields["number"])
	})

	t.Run("AddOutput", func(t *testing.T) {
		syslogConfig := SyslogConfig{
			Network:  "udp",
			Address:  "localhost:514",
			Priority: 16,
			Tag:      "test",
		}
		newConfig := config.AddOutput("syslog", syslogConfig)
		
		assert.Contains(t, newConfig.Outputs, "syslog")
		assert.Contains(t, newConfig.OutputConfig, "syslog")
		
		retrievedConfig, exists := newConfig.GetOutputConfig("syslog")
		assert.True(t, exists)
		assert.Equal(t, syslogConfig, retrievedConfig)
	})

	t.Run("RemoveOutput", func(t *testing.T) {
		// 先添加一个输出
		newConfig := config.AddOutput("syslog", nil)
		assert.Contains(t, newConfig.Outputs, "syslog")
		
		// 然后移除
		finalConfig := newConfig.RemoveOutput("syslog")
		assert.NotContains(t, finalConfig.Outputs, "syslog")
		assert.NotContains(t, finalConfig.OutputConfig, "syslog")
	})

	t.Run("HasOutput", func(t *testing.T) {
		assert.True(t, config.HasOutput("console"))
		assert.True(t, config.HasOutput("file"))
		assert.False(t, config.HasOutput("syslog"))
	})
}

func TestLogConfig_OutputConfigs(t *testing.T) {
	t.Run("SyslogConfig", func(t *testing.T) {
		config := SyslogConfig{
			Network:  "tcp",
			Address:  "localhost:514",
			Priority: 16,
			Tag:      "test-app",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("FluentdConfig", func(t *testing.T) {
		config := FluentdConfig{
			Host:    "localhost",
			Port:    24224,
			Tag:     "test.logs",
			Timeout: 5 * time.Second,
			Extra: map[string]string{
				"env": "test",
			},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("CloudWatchConfig", func(t *testing.T) {
		config := CloudWatchConfig{
			Region:          "us-east-1",
			LogGroupName:    "/aws/test/application",
			LogStreamName:   "test-stream",
			AccessKeyID:     "test-key",
			SecretAccessKey: "test-secret",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("AzureInsightsConfig", func(t *testing.T) {
		config := AzureInsightsConfig{
			InstrumentationKey: "test-key",
			Endpoint:           "https://dc.services.visualstudio.com/v2/track",
			Properties: map[string]string{
				"app": "test",
			},
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("ElasticsearchConfig", func(t *testing.T) {
		config := ElasticsearchConfig{
			URLs:     []string{"http://localhost:9200"},
			Index:    "test-logs",
			Username: "elastic",
			Password: "password",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("KafkaConfig", func(t *testing.T) {
		config := KafkaConfig{
			Brokers:  []string{"localhost:9092"},
			Topic:    "test-logs",
			ClientID: "test-client",
		}
		err := config.Validate()
		assert.NoError(t, err)
	})
}

func TestLogConfig_ValidateConfig(t *testing.T) {
	t.Run("ValidConfig", func(t *testing.T) {
		config := DefaultLogConfig()
		err := config.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("ConfigWithOutputs", func(t *testing.T) {
		config := ProductionLogConfig()
		err := config.ValidateConfig()
		assert.NoError(t, err)
	})
}

func TestLogConfig_CreateLogger(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultLogConfig()
		logger := config.CreateLogger()
		
		assert.NotNil(t, logger)
		assert.NotNil(t, logger.Logger())
	})

	t.Run("DifferentFormats", func(t *testing.T) {
		formats := []LogFormat{
			LogFormatBeego,
			LogFormatLog4Go,
			LogFormatLogstash,
			LogFormatSyslog,
			LogFormatFluentd,
			LogFormatCloudWatch,
			LogFormatApplicationInsights,
			LogFormatJSON,
			LogFormatText,
		}

		for _, format := range formats {
			t.Run(string(format), func(t *testing.T) {
				config := DefaultLogConfig()
				config.Format = format
				logger := config.CreateLogger()
				
				assert.NotNil(t, logger)
				assert.NotNil(t, logger.Logger())
			})
		}
	})
}

func TestLogConfig_ConfigInterface(t *testing.T) {
	config := LogConfig{}

	t.Run("GetConfigName", func(t *testing.T) {
		name := config.GetConfigName()
		assert.Equal(t, LogConfigName, name)
	})

	t.Run("GenerateDefaultContent", func(t *testing.T) {
		content := config.GenerateDefaultContent()
		assert.NotEmpty(t, content)
		assert.Contains(t, content, "level:")
		assert.Contains(t, content, "format:")
		assert.Contains(t, content, "outputs:")
		assert.Contains(t, content, "beego:")
		assert.Contains(t, content, "log4go:")
		assert.Contains(t, content, "logstash:")
		assert.Contains(t, content, "syslog:")
		assert.Contains(t, content, "fluentd:")
		assert.Contains(t, content, "cloudwatch:")
		assert.Contains(t, content, "azure_insights:")
	})
}


func TestLogOutputs(t *testing.T) {
	outputs := []LogOutput{
		LogOutputConsole,
		LogOutputFile,
		LogOutputSyslog,
		LogOutputFluentd,
		LogOutputCloudWatch,
		LogOutputAzureInsights,
		LogOutputElasticsearch,
		LogOutputKafka,
	}

	for _, output := range outputs {
		t.Run(string(output), func(t *testing.T) {
			// 验证输出类型定义
			assert.NotEmpty(t, string(output))
		})
	}
}

// 基准测试
func BenchmarkDefaultLogConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		config := DefaultLogConfig()
		_ = config
	}
}

func BenchmarkCreateLogger(b *testing.B) {
	config := DefaultLogConfig()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		logger := config.CreateLogger()
		_ = logger
	}
}

func BenchmarkLogMessage(b *testing.B) {
	config := DefaultLogConfig()
	config.EnableFile = false // 只输出到控制台以提高性能
	logger := config.CreateLogger()
	logrusLogger := logger.Logger()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		logrusLogger.Info("benchmark log message")
	}
}

func BenchmarkLogWithFields(b *testing.B) {
	config := DefaultLogConfig()
	config.EnableFile = false // 只输出到控制台以提高性能
	logger := config.CreateLogger()
	logrusLogger := logger.Logger()
	
	fields := map[string]interface{}{
		"user_id":    "user123",
		"request_id": "req456",
		"action":     "benchmark",
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		logrusLogger.WithFields(fields).Info("benchmark log message with fields")
	}
}