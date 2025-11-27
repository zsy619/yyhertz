package config

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBeegoFormatter(t *testing.T) {
	formatter := &BeegoFormatter{
		TimestampFormat: "2006/01/02 15:04:05.000",
		ShowCaller:      true,
	}

	entry := &logrus.Entry{
		Time:    time.Date(2023, 12, 25, 10, 30, 45, 123456789, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data: logrus.Fields{
			"user_id": "123",
			"action":  "login",
		},
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	assert.Contains(t, output, "[I]")
	assert.Contains(t, output, "2023/12/25 10:30:45.123")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "user_id=123")
	assert.Contains(t, output, "action=login")
	assert.True(t, strings.HasSuffix(output, "\n"))
}

func TestBeegoFormatter_LevelChars(t *testing.T) {
	formatter := &BeegoFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
		ShowCaller:      false,
	}

	testCases := []struct {
		level        logrus.Level
		expectedChar string
	}{
		{logrus.DebugLevel, "D"},
		{logrus.InfoLevel, "I"},
		{logrus.WarnLevel, "W"},
		{logrus.ErrorLevel, "E"},
		{logrus.FatalLevel, "F"},
		{logrus.PanicLevel, "P"},
	}

	for _, tc := range testCases {
		t.Run(tc.level.String(), func(t *testing.T) {
			entry := &logrus.Entry{
				Time:    time.Now(),
				Level:   tc.level,
				Message: "test",
			}

			result, err := formatter.Format(entry)
			require.NoError(t, err)

			output := string(result)
			assert.Contains(t, output, "["+tc.expectedChar+"]")
		})
	}
}

func TestLog4GoFormatter(t *testing.T) {
	formatter := &Log4GoFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
		ShowCaller:      true,
	}

	entry := &logrus.Entry{
		Time:    time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data: logrus.Fields{
			"key": "value",
		},
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	assert.Contains(t, output, "[2023/12/25 10:30:45]")
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key=value")
}

func TestLogstashFormatter(t *testing.T) {
	formatter := &LogstashFormatter{
		TimestampFormat: time.RFC3339,
		ServiceName:     "test-service",
		Version:         "1.0.0",
	}

	entry := &logrus.Entry{
		Time:    time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data: logrus.Fields{
			"user_id": "123",
		},
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	
	// 验证JSON格式
	assert.Contains(t, output, `"@timestamp"`)
	assert.Contains(t, output, `"@version":"1"`)
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"message":"test message"`)
	assert.Contains(t, output, `"service":"test-service"`)
	assert.Contains(t, output, `"version":"1.0.0"`)
	assert.Contains(t, output, `"user_id":"123"`)
	assert.True(t, strings.HasSuffix(output, "\n"))
}

func TestSyslogFormatter(t *testing.T) {
	formatter := &SyslogFormatter{
		TimestampFormat: "Jan 2 15:04:05",
		Hostname:        "testhost",
		Tag:             "testapp",
		Facility:        16, // local0
	}

	entry := &logrus.Entry{
		Time:    time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data: logrus.Fields{
			"key": "value",
		},
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	
	// 验证Syslog格式
	assert.Contains(t, output, "<134>") // priority: 16*8 + 6 = 134
	assert.Contains(t, output, "Dec 25 10:30:45")
	assert.Contains(t, output, "testhost")
	assert.Contains(t, output, "testapp:")
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "key=value")
}

func TestSyslogFormatter_Priorities(t *testing.T) {
	formatter := &SyslogFormatter{
		Facility: 16, // local0
	}

	testCases := []struct {
		level            logrus.Level
		expectedSeverity int
	}{
		{logrus.PanicLevel, 0}, // Emergency
		{logrus.FatalLevel, 2}, // Critical
		{logrus.ErrorLevel, 3}, // Error
		{logrus.WarnLevel, 4},  // Warning
		{logrus.InfoLevel, 6},  // Info
		{logrus.DebugLevel, 7}, // Debug
	}

	for _, tc := range testCases {
		t.Run(tc.level.String(), func(t *testing.T) {
			entry := &logrus.Entry{
				Time:    time.Now(),
				Level:   tc.level,
				Message: "test",
			}

			result, err := formatter.Format(entry)
			require.NoError(t, err)

			expectedPriority := 16*8 + tc.expectedSeverity
			output := string(result)
			assert.Contains(t, output, "<"+string(rune(expectedPriority+48))+">") // 转换为字符
		})
	}
}

func TestFluentdFormatter(t *testing.T) {
	formatter := &FluentdFormatter{
		TimestampFormat: time.RFC3339,
		ServiceName:     "test-service",
		Environment:     "test",
	}

	entry := &logrus.Entry{
		Time:    time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data: logrus.Fields{
			"custom": "field",
		},
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	
	// 验证JSON格式
	assert.Contains(t, output, `"timestamp"`)
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"message":"test message"`)
	assert.Contains(t, output, `"service":"test-service"`)
	assert.Contains(t, output, `"environment":"test"`)
	assert.Contains(t, output, `"custom":"field"`)
}

func TestCloudWatchFormatter(t *testing.T) {
	formatter := &CloudWatchFormatter{
		TimestampFormat: time.RFC3339,
		ServiceName:     "test-service",
		LogGroupName:    "/aws/test/app",
		LogStreamName:   "test-stream",
	}

	entry := &logrus.Entry{
		Time:    time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "test message",
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	
	// 验证CloudWatch格式
	assert.Contains(t, output, `"timestamp"`)
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"message":"test message"`)
	assert.Contains(t, output, `"service":"test-service"`)
	assert.Contains(t, output, `"logGroup":"/aws/test/app"`)
	assert.Contains(t, output, `"logStream":"test-stream"`)
}

func TestAzureInsightsFormatter(t *testing.T) {
	formatter := &AzureInsightsFormatter{
		TimestampFormat:    time.RFC3339,
		ServiceName:        "test-service",
		InstrumentationKey: "test-key-12345",
	}

	entry := &logrus.Entry{
		Time:    time.Date(2023, 12, 25, 10, 30, 45, 0, time.UTC),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Data: logrus.Fields{
			"custom": "property",
		},
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	
	// 验证Azure Insights格式
	assert.Contains(t, output, `"time"`)
	assert.Contains(t, output, `"iKey":"test-key-12345"`)
	assert.Contains(t, output, `"name":"Microsoft.ApplicationInsights.Message"`)
	assert.Contains(t, output, `"baseType":"MessageData"`)
	assert.Contains(t, output, `"message":"test message"`)
	assert.Contains(t, output, `"severityLevel"`)
}

func TestAzureInsightsFormatter_SeverityLevels(t *testing.T) {
	formatter := &AzureInsightsFormatter{}

	testCases := []struct {
		level            logrus.Level
		expectedSeverity int
	}{
		{logrus.DebugLevel, 0}, // Verbose
		{logrus.InfoLevel, 1},  // Information
		{logrus.WarnLevel, 2},  // Warning
		{logrus.ErrorLevel, 3}, // Error
		{logrus.FatalLevel, 4}, // Critical
		{logrus.PanicLevel, 4}, // Critical
	}

	for _, tc := range testCases {
		t.Run(tc.level.String(), func(t *testing.T) {
			severity := formatter.getInsightsSeverity(tc.level)
			assert.Equal(t, tc.expectedSeverity, severity)
		})
	}
}

func TestGetFormatter(t *testing.T) {
	config := DefaultLogConfig()

	testCases := []struct {
		format         LogFormat
		expectedType   string
		expectNotNil   bool
	}{
		{LogFormatBeego, "*config.BeegoFormatter", true},
		{LogFormatLog4Go, "*config.Log4GoFormatter", true},
		{LogFormatLogstash, "*config.LogstashFormatter", true},
		{LogFormatSyslog, "*config.SyslogFormatter", true},
		{LogFormatFluentd, "*config.FluentdFormatter", true},
		{LogFormatCloudWatch, "*config.CloudWatchFormatter", true},
		{LogFormatApplicationInsights, "*config.AzureInsightsFormatter", true},
		{LogFormatJSON, "*logrus.JSONFormatter", true},
		{LogFormatText, "*logrus.TextFormatter", true},
		{"unknown", "*config.BeegoFormatter", true}, // 默认回退到Beego
	}

	for _, tc := range testCases {
		t.Run(string(tc.format), func(t *testing.T) {
			formatter := GetFormatter(tc.format, config)
			
			if tc.expectNotNil {
				assert.NotNil(t, formatter)
			} else {
				assert.Nil(t, formatter)
			}
		})
	}
}

func TestFormatterWithCaller(t *testing.T) {
	formatter := &BeegoFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
		ShowCaller:      true,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "test message",
		Caller: &runtime.Frame{
			File: "/path/to/test.go",
			Line: 42,
			Function: "TestFunction",
		},
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	assert.Contains(t, output, "[test.go:42]")
}

func TestFormatterWithoutCaller(t *testing.T) {
	formatter := &BeegoFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
		ShowCaller:      false,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "test message",
	}

	result, err := formatter.Format(entry)
	require.NoError(t, err)

	output := string(result)
	assert.NotContains(t, output, "[")
	assert.NotContains(t, output, ":")
	assert.NotContains(t, output, "]")
}

// 基准测试
func BenchmarkBeegoFormatter(b *testing.B) {
	formatter := &BeegoFormatter{
		TimestampFormat: "2006/01/02 15:04:05.000",
		ShowCaller:      true,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "benchmark test message",
		Data: logrus.Fields{
			"user_id": "123",
			"action":  "test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

func BenchmarkLog4GoFormatter(b *testing.B) {
	formatter := &Log4GoFormatter{
		TimestampFormat: "2006/01/02 15:04:05",
		ShowCaller:      true,
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "benchmark test message",
		Data: logrus.Fields{
			"user_id": "123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

func BenchmarkLogstashFormatter(b *testing.B) {
	formatter := &LogstashFormatter{
		TimestampFormat: time.RFC3339,
		ServiceName:     "test-service",
		Version:         "1.0.0",
	}

	entry := &logrus.Entry{
		Time:    time.Now(),
		Level:   logrus.InfoLevel,
		Message: "benchmark test message",
		Data: logrus.Fields{
			"user_id": "123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = formatter.Format(entry)
	}
}

func BenchmarkGetFormatter(b *testing.B) {
	config := DefaultLogConfig()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetFormatter(LogFormatBeego, config)
	}
}