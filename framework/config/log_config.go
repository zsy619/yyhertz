package config

import (
	"fmt"
	"time"
)

// DefaultLogConfig 返回默认日志配置
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatGin,
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
		Format:          LogFormatGin,
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
			"service":    "yyhertz",
			"version":    "1.0.0",
			"deployment": "cloud",
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

// ============= 框架专用日志配置 =============

// GinLogConfig Gin框架专用日志配置
func GinLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatGin,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/gin.log",
		MaxSize:         100,
		MaxAge:          7,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      false, // Gin通常不需要显示调用位置
		ShowTimestamp:   true,
		TimestampFormat: "2006/01/02 - 15:04:05",
		Fields: map[string]any{
			"framework": "gin",
			"service":   "yyhertz",
		},
		Outputs:      []string{"console", "file"},
		OutputConfig: make(map[string]OutputConfig),
	}
}

// MicroserviceLogConfig 微服务日志配置（适用于go-zero等）
func MicroserviceLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatGoZero,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/microservice.log",
		MaxSize:         100,
		MaxAge:          30,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"service":     "yyhertz",
			"environment": "production",
			"version":     "1.0.0",
		},
		Outputs:      []string{"console", "file"},
		OutputConfig: make(map[string]OutputConfig),
	}
}

// DatabaseLogConfig 数据库ORM日志配置（适用于Ent等）
func DatabaseLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelDebug, // 数据库通常需要调试级别
		Format:          LogFormatEnt,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/database.log",
		MaxSize:         50,
		MaxAge:          7,
		MaxBackups:      5,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05",
		Fields: map[string]any{
			"component": "database",
			"orm":       "ent",
		},
		Outputs:      []string{"console", "file"},
		OutputConfig: make(map[string]OutputConfig),
	}
}

// HighPerformanceFrameworkLogConfig 高性能框架日志配置（适用于Fiber等）
func HighPerformanceFrameworkLogConfig() *LogConfig {
	return &LogConfig{
		Level:           LogLevelWarn, // 高性能场景减少日志级别
		Format:          LogFormatFiber,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/fiber.log",
		MaxSize:         100,
		MaxAge:          3,
		MaxBackups:      3,
		Compress:        true,
		ShowCaller:      false, // 高性能场景不显示调用位置
		ShowTimestamp:   true,
		TimestampFormat: "15:04:05",
		Fields: map[string]any{
			"framework":   "fiber",
			"performance": "optimized",
		},
		Outputs:      []string{"console", "file"},
		OutputConfig: make(map[string]OutputConfig),
	}
}

// WebFrameworkLogConfig 通用Web框架日志配置（适用于Echo、Iris等）
func WebFrameworkLogConfig(framework string) *LogConfig {
	format := LogFormatEcho
	timestampFormat := time.RFC3339

	switch framework {
	case "iris":
		format = LogFormatIris
		timestampFormat = "2006/01/02 15:04:05"
	case "echo":
		format = LogFormatEcho
	case "revel":
		format = LogFormatRevel
		timestampFormat = "2006/01/02 15:04:05"
	case "buffalo":
		format = LogFormatBuffalo
		timestampFormat = "2006/01/02 15:04:05"
	}

	return &LogConfig{
		Level:           LogLevelInfo,
		Format:          format,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        fmt.Sprintf("./logs/%s.log", framework),
		MaxSize:         100,
		MaxAge:          7,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: timestampFormat,
		Fields: map[string]any{
			"framework": framework,
			"service":   "yyhertz",
		},
		Outputs:      []string{"console", "file"},
		OutputConfig: make(map[string]OutputConfig),
	}
}
