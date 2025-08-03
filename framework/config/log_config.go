package config

import (
	"time"
)

// DefaultLogConfig 返回默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatBeego,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/app.log",
		MaxSize:         100,
		MaxAge:          7,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05.000",
		Fields:          make(map[string]any),
		Outputs:         []string{"console", "file"},
		OutputConfig:    make(map[string]OutputConfig),
	}
}

// DevelopmentLogConfig 开发环境日志配置
func DevelopmentLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelDebug,
		Format:          LogFormatBeego,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/dev.log",
		MaxSize:         50,
		MaxAge:          3,
		MaxBackups:      5,
		Compress:        false,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05.000",
		Fields: map[string]any{
			"env":     "development",
			"service": "yyhertz",
		},
		Outputs:      []string{"console", "file"},
		OutputConfig: make(map[string]OutputConfig),
	}
}

// ProductionLogConfig 生产环境日志配置
func ProductionLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatLogstash,
		EnableConsole:   false,
		EnableFile:      true,
		FilePath:        "./logs/prod.log",
		MaxSize:         100,
		MaxAge:          30,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      false,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"env":     "production",
			"service": "yyhertz",
			"version": "1.0.0",
		},
		Outputs: []string{"file", "fluentd"},
		OutputConfig: map[string]OutputConfig{
			"fluentd": FluentdConfig{
				Host:    "localhost",
				Port:    24224,
				Tag:     "yyhertz.prod",
				Timeout: 3 * time.Second,
				Extra: map[string]string{
					"environment": "production",
				},
			},
		},
	}
}

// TestLogConfig 测试环境日志配置
func TestLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelWarn,
		Format:          LogFormatBeego,
		EnableConsole:   true,
		EnableFile:      false,
		ShowCaller:      false,
		ShowTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05.000",
		Fields:          map[string]any{},
		Outputs:         []string{"console"},
		OutputConfig:    make(map[string]OutputConfig),
	}
}

// HighPerformanceLogConfig 高性能日志配置（最小日志）
func HighPerformanceLogConfig() *LogConfig {
	return &LogConfig{
		Level:         LogLevelError,
		Format:        LogFormatJSON,
		EnableConsole: false,
		EnableFile:    true,
		FilePath:      "./logs/error.log",
		MaxSize:       200,
		MaxAge:        7,
		MaxBackups:    3,
		Compress:      true,
		ShowCaller:    true,
		ShowTimestamp: true,
		Fields: map[string]any{
			"mode": "high-performance",
		},
		Outputs:      []string{"file"},
		OutputConfig: make(map[string]OutputConfig),
	}
}

// CloudLogConfig 云端日志配置（支持多种云服务）
func CloudLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatCloudWatch,
		EnableConsole:   false,
		EnableFile:      true,
		FilePath:        "./logs/cloud.log",
		MaxSize:         100,
		MaxAge:          30,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"service":     "yyhertz",
			"version":     "1.0.0",
			"deployment":  "cloud",
		},
		Outputs: []string{"file", "cloudwatch", "azure_insights"},
		OutputConfig: map[string]OutputConfig{
			"cloudwatch": CloudWatchConfig{
				Region:        "us-east-1",
				LogGroupName:  "/aws/yyhertz/application",
				LogStreamName: "yyhertz-instance-001",
			},
			"azure_insights": AzureInsightsConfig{
				Endpoint: "https://dc.services.visualstudio.com/v2/track",
				Properties: map[string]string{
					"application": "yyhertz",
					"environment": "cloud",
				},
			},
		},
	}
}

// UpdateConfigLevel 更新配置的日志级别
func (cfg *LogConfig) UpdateConfigLevel(level LogLevel) *LogConfig {
	newConfig := *cfg // 复制配置
	newConfig.Level = level
	return &newConfig
}

// UpdateConfigFormat 更新配置的日志格式
func (cfg *LogConfig) UpdateConfigFormat(format LogFormat) *LogConfig {
	newConfig := *cfg // 复制配置
	newConfig.Format = format
	return &newConfig
}

// AddConfigFields 向配置添加字段
func (cfg *LogConfig) AddConfigFields(fields map[string]any) *LogConfig {
	newConfig := *cfg // 复制配置
	if newConfig.Fields == nil {
		newConfig.Fields = make(map[string]any)
	}
	for k, v := range fields {
		newConfig.Fields[k] = v
	}
	return &newConfig
}

// AddOutput 添加输出目标
func (cfg *LogConfig) AddOutput(output string, config OutputConfig) *LogConfig {
	newConfig := *cfg // 复制配置
	
	// 添加输出类型
	found := false
	for _, o := range newConfig.Outputs {
		if o == output {
			found = true
			break
		}
	}
	if !found {
		newConfig.Outputs = append(newConfig.Outputs, output)
	}
	
	// 添加输出配置
	if newConfig.OutputConfig == nil {
		newConfig.OutputConfig = make(map[string]OutputConfig)
	}
	if config != nil {
		newConfig.OutputConfig[output] = config
	}
	
	return &newConfig
}

// RemoveOutput 移除输出目标
func (cfg *LogConfig) RemoveOutput(output string) *LogConfig {
	newConfig := *cfg // 复制配置
	
	// 移除输出类型
	newOutputs := make([]string, 0, len(newConfig.Outputs))
	for _, o := range newConfig.Outputs {
		if o != output {
			newOutputs = append(newOutputs, o)
		}
	}
	newConfig.Outputs = newOutputs
	
	// 移除输出配置
	if newConfig.OutputConfig != nil {
		delete(newConfig.OutputConfig, output)
	}
	
	return &newConfig
}

// HasOutput 检查是否包含指定输出
func (cfg *LogConfig) HasOutput(output string) bool {
	for _, o := range cfg.Outputs {
		if o == output {
			return true
		}
	}
	return false
}

// GetOutputConfig 获取指定输出的配置
func (cfg *LogConfig) GetOutputConfig(output string) (OutputConfig, bool) {
	if cfg.OutputConfig == nil {
		return nil, false
	}
	config, exists := cfg.OutputConfig[output]
	return config, exists
}

// ValidateConfig 验证配置的有效性
func (cfg *LogConfig) ValidateConfig() error {
	// 验证各输出配置
	for output, config := range cfg.OutputConfig {
		if config != nil {
			if err := config.Validate(); err != nil {
				return err
			}
		}
		_ = output // 可以添加更多验证逻辑
	}
	return nil
}