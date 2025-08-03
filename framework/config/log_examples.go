package config

import (
	"fmt"
	"time"
)

// SingletonLoggerUsageExample 展示单例日志管理器的使用方法
func SingletonLoggerUsageExample() {
	fmt.Println("=== 单例日志管理器使用示例 ===")

	// ============= 1. 初始化全局日志 =============

	// 方式1: 使用默认配置
	InitGlobalLogger(DefaultLogConfig())

	// 方式2: 使用自定义配置
	customConfig := &LogConfig{
		Level:           LogLevelDebug,
		Format:          LogFormatBeego,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/singleton.log",
		MaxSize:         50,
		MaxAge:          7,
		MaxBackups:      5,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: "2006/01/02 15:04:05.000",
		Fields: map[string]any{
			"service": "hertz-singleton-demo",
			"version": "2.0.0",
		},
		Outputs: []string{"console", "file"},
	}
	InitGlobalLogger(customConfig)

	// ============= 2. 基础日志使用 =============

	// 直接使用全局函数（推荐方式）
	Debug("这是调试信息")
	Debugf("用户ID: %s 执行了操作", "user123")

	Info("应用启动成功")
	Infof("服务监听在端口: %d", 8080)

	Warn("内存使用率较高")
	Warnf("连接池使用率: %.2f%%", 85.5)

	Error("数据库连接失败")
	Errorf("Redis连接超时: %v", "5s")

	// ============= 3. 结构化日志使用 =============

	// 使用WithFields添加结构化字段
	WithFields(map[string]any{
		"user_id":    "12345",
		"request_id": "req-789",
		"action":     "login",
	}).Info("用户登录成功")

	// 使用WithField添加单个字段
	WithField("order_id", "order-456").
		WithField("amount", 99.99).
		Info("订单创建成功")

	// 使用WithError添加错误信息
	err := fmt.Errorf("database connection timeout")
	WithError(err).Error("数据库操作失败")

	// ============= 4. 特定用途的日志 =============

	// HTTP请求日志
	LogHTTPRequest("GET", "/api/users", "192.168.1.100", 200, 120.5)

	// API调用日志
	LogAPICall("getUserInfo", true, 50.2, "")
	LogAPICall("sendEmail", false, 1500.0, "SMTP服务器不可用")

	// 数据库查询日志
	LogDBQuery("SELECT * FROM users WHERE id = ?", 15.3, 1, nil)

	// 性能日志
	LogPerformance("userAuthentication", 89.2, "认证操作完成")

	// 安全事件日志
	LogSecurityEvent("login_attempt", "user123", "192.168.1.100", map[string]any{
		"success": true,
		"method":  "password",
	})

	// 业务事件日志
	LogBusinessEvent("order_created", "order-789", "order", map[string]any{
		"amount":      199.99,
		"customer_id": "cust-456",
		"items":       3,
	})

	// ============= 5. 动态配置更新 =============

	// 动态更新日志级别
	UpdateGlobalLogLevel(LogLevelWarn)
	Info("这条消息不会显示，因为级别已设为WARN")
	Warn("这条警告消息会显示")

	// 动态更新日志格式
	UpdateGlobalLogFormat(LogFormatJSON)
	Info("现在使用JSON格式输出")

	// 重置为原始级别和格式
	UpdateGlobalLogLevel(LogLevelInfo)
	UpdateGlobalLogFormat(LogFormatBeego)
	Info("日志配置已重置")
}

// MultiFormatLogExample 展示多种日志格式的使用
func MultiFormatLogExample() {
	fmt.Println("\n=== 多种日志格式示例 ===")

	// Beego格式
	beegoConfig := DefaultLogConfig()
	beegoConfig.Format = LogFormatBeego
	beegoLogger := beegoConfig.CreateLogger()
	beegoLogger.Logger().Info("这是Beego风格的日志")

	// Log4Go格式
	log4goConfig := DefaultLogConfig()
	log4goConfig.Format = LogFormatLog4Go
	log4goLogger := log4goConfig.CreateLogger()
	log4goLogger.Logger().Info("这是Log4Go风格的日志")

	// Logstash格式
	logstashConfig := DefaultLogConfig()
	logstashConfig.Format = LogFormatLogstash
	logstashLogger := logstashConfig.CreateLogger()
	logstashLogger.Logger().WithField("custom_field", "custom_value").Info("这是Logstash风格的日志")

	// Syslog格式
	syslogConfig := DefaultLogConfig()
	syslogConfig.Format = LogFormatSyslog
	syslogLogger := syslogConfig.CreateLogger()
	syslogLogger.Logger().Info("这是Syslog风格的日志")

	// Fluentd格式
	fluentdConfig := DefaultLogConfig()
	fluentdConfig.Format = LogFormatFluentd
	fluentdLogger := fluentdConfig.CreateLogger()
	fluentdLogger.Logger().WithField("environment", "production").Info("这是Fluentd风格的日志")

	// CloudWatch格式
	cloudwatchConfig := DefaultLogConfig()
	cloudwatchConfig.Format = LogFormatCloudWatch
	cloudwatchLogger := cloudwatchConfig.CreateLogger()
	cloudwatchLogger.Logger().Info("这是CloudWatch风格的日志")

	// Azure Insights格式
	azureConfig := DefaultLogConfig()
	azureConfig.Format = LogFormatApplicationInsights
	azureLogger := azureConfig.CreateLogger()
	azureLogger.Logger().Info("这是Azure Application Insights风格的日志")
}

// MultiOutputLogExample 展示多种输出目标的使用
func MultiOutputLogExample() {
	fmt.Println("\n=== 多种输出目标示例 ===")

	// 创建包含多种输出的配置
	config := &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatLogstash,
		EnableConsole:   true,
		EnableFile:      true,
		FilePath:        "./logs/multi-output.log",
		MaxSize:         100,
		MaxAge:          7,
		MaxBackups:      10,
		Compress:        true,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"service":     "multi-output-demo",
			"version":     "1.0.0",
			"environment": "production",
		},
		Outputs: []string{"console", "file", "fluentd", "syslog"},
		OutputConfig: map[string]OutputConfig{
			"fluentd": FluentdConfig{
				Host:    "localhost",
				Port:    24224,
				Tag:     "yyhertz.demo",
				Timeout: 3 * time.Second,
				Extra: map[string]string{
					"environment": "production",
					"datacenter":  "us-east-1",
				},
			},
			"syslog": SyslogConfig{
				Network:  "udp",
				Address:  "localhost:514",
				Priority: 16, // local0.info
				Tag:      "yyhertz-demo",
			},
		},
	}

	// 创建日志器
	logger := config.CreateLogger()
	logrusLogger := logger.Logger()

	// 输出日志到多个目标
	logrusLogger.Info("这条日志会同时输出到控制台、文件、Fluentd和Syslog")
	logrusLogger.WithFields(map[string]interface{}{
		"user_id":    "user123",
		"session_id": "sess456",
		"action":     "multi_output_test",
	}).Info("带字段的多输出日志")

	fmt.Println("日志已发送到多个输出目标")
}

// CloudLogExample 展示云端日志服务的使用
func CloudLogExample() {
	fmt.Println("\n=== 云端日志服务示例 ===")

	// AWS CloudWatch配置
	cloudwatchConfig := &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatCloudWatch,
		EnableConsole:   true,
		EnableFile:      false,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"service":     "yyhertz-cloud",
			"version":     "1.0.0",
			"environment": "production",
		},
		Outputs: []string{"console", "cloudwatch"},
		OutputConfig: map[string]OutputConfig{
			"cloudwatch": CloudWatchConfig{
				Region:        "us-east-1",
				LogGroupName:  "/aws/yyhertz/application",
				LogStreamName: "yyhertz-instance-001",
				// 在实际使用中，这些应该从环境变量获取
				// AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
				// SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
			},
		},
	}

	cloudwatchLogger := cloudwatchConfig.CreateLogger()
	cloudwatchLogger.Logger().Info("这是发送到AWS CloudWatch的日志")

	// Azure Application Insights配置
	azureConfig := &LogConfig{
		Level:           LogLevelInfo,
		Format:          LogFormatApplicationInsights,
		EnableConsole:   true,
		EnableFile:      false,
		ShowCaller:      true,
		ShowTimestamp:   true,
		TimestampFormat: time.RFC3339,
		Fields: map[string]any{
			"service":     "yyhertz-azure",
			"version":     "1.0.0",
			"environment": "production",
		},
		Outputs: []string{"console", "azure_insights"},
		OutputConfig: map[string]OutputConfig{
			"azure_insights": AzureInsightsConfig{
				InstrumentationKey: "your-instrumentation-key-here",
				Endpoint:           "https://dc.services.visualstudio.com/v2/track",
				Properties: map[string]string{
					"application": "yyhertz",
					"environment": "production",
				},
			},
		},
	}

	azureLogger := azureConfig.CreateLogger()
	azureLogger.Logger().Info("这是发送到Azure Application Insights的日志")

	fmt.Println("云端日志服务配置完成")
}

// PresetConfigExample 展示预设配置的使用
func PresetConfigExample() {
	fmt.Println("\n=== 预设配置示例 ===")

	// 开发环境配置
	fmt.Println("开发环境配置：")
	devConfig := DevelopmentLogConfig()
	devLogger := devConfig.CreateLogger()
	devLogger.Logger().Debug("开发环境调试日志")
	devLogger.Logger().Info("开发环境信息日志")

	// 生产环境配置
	fmt.Println("生产环境配置：")
	prodConfig := ProductionLogConfig()
	prodLogger := prodConfig.CreateLogger()
	prodLogger.Logger().Info("生产环境信息日志")
	prodLogger.Logger().Error("生产环境错误日志")

	// 测试环境配置
	fmt.Println("测试环境配置：")
	testConfig := TestLogConfig()
	testLogger := testConfig.CreateLogger()
	testLogger.Logger().Warn("测试环境警告日志")
	testLogger.Logger().Error("测试环境错误日志")

	// 高性能配置
	fmt.Println("高性能配置：")
	perfConfig := HighPerformanceLogConfig()
	perfLogger := perfConfig.CreateLogger()
	perfLogger.Logger().Error("高性能配置只记录错误日志")

	// 云端配置
	fmt.Println("云端配置：")
	cloudConfig := CloudLogConfig()
	cloudLogger := cloudConfig.CreateLogger()
	cloudLogger.Logger().Info("云端配置日志")

	fmt.Println("所有预设配置演示完成")
}

// DynamicConfigExample 展示动态配置管理
func DynamicConfigExample() {
	fmt.Println("\n=== 动态配置管理示例 ===")

	// 初始化默认配置
	InitGlobalLogger(DefaultLogConfig())
	Info("初始配置 - Beego格式，Info级别")

	// 动态添加Fluentd输出
	fluentdConfig := FluentdConfig{
		Host:    "localhost",
		Port:    24224,
		Tag:     "yyhertz.dynamic",
		Timeout: 5 * time.Second,
	}
	AddGlobalLogOutput("fluentd", fluentdConfig)
	Info("已添加Fluentd输出")

	// 动态更新级别为Debug
	UpdateGlobalLogLevel(LogLevelDebug)
	Debug("现在可以看到Debug日志了")

	// 动态更新格式为JSON
	UpdateGlobalLogFormat(LogFormatJSON)
	Info("现在使用JSON格式")

	// 动态移除Fluentd输出
	RemoveGlobalLogOutput("fluentd")
	Info("已移除Fluentd输出")

	// 恢复原始配置
	UpdateGlobalLogLevel(LogLevelInfo)
	UpdateGlobalLogFormat(LogFormatBeego)
	Info("配置已恢复")

	fmt.Println("动态配置管理演示完成")
}

// PerformanceLogExample 展示性能和监控日志
func PerformanceLogExample() {
	fmt.Println("\n=== 性能和监控日志示例 ===")

	// 模拟一些性能数据
	startTime := time.Now()

	// 模拟一些操作
	time.Sleep(100 * time.Millisecond)

	// 记录性能日志
	duration := float64(time.Since(startTime).Nanoseconds()) / 1e6
	LogPerformance("example_operation", duration, "操作完成")

	// 记录指标日志
	LogMetric("response_time", duration, map[string]string{
		"endpoint": "/api/users",
		"method":   "GET",
	}, "API响应时间")

	LogMetric("memory_usage", 85.5, map[string]string{
		"component": "cache",
		"node":      "node-001",
	}, "内存使用率")

	// 记录事件日志
	LogEvent("cache_miss", map[string]any{
		"key":       "user:12345",
		"cache_ttl": 300,
		"hit_rate":  0.95,
	}, "缓存未命中事件")

	// 记录系统事件
	LogSystemEvent("service_started", "user_service", map[string]any{
		"port":    8080,
		"version": "1.2.3",
		"node_id": "node-001",
	})

	// 记录启动日志
	LogStartup("yyhertz-demo", "1.0.0", 8080, map[string]any{
		"environment": "production",
		"region":      "us-east-1",
		"debug":       false,
	})

	fmt.Println("性能和监控日志演示完成")
}

// WithRequestIDExample 展示改进的WithRequestID使用方法
func WithRequestIDExample() {
	fmt.Println("\n=== WithRequestID方法示例 ===")

	// 初始化全局日志器
	InitGlobalLogger(DevelopmentLogConfig())

	fmt.Println("1. 有效RequestID示例：")
	validIDs := []string{
		"req-001-12345678",                       // HTTP请求ID
		"batch-job-20250803101000",               // 批处理任务ID
		"user.session.a1b2c3d4e5f6g7h8",         // 用户会话ID
		"transaction-uuid-1234567890abcdef",      // 交易ID
	}

	for _, requestID := range validIDs {
		// 使用WithRequestID添加请求ID并记录日志
		logger := WithRequestID(requestID)
		logger.WithField("action", "process_request").Info("处理请求开始")
		logger.WithField("status", "success").Info("请求处理完成")
	}

	fmt.Println("\n2. 无效RequestID处理示例：")
	invalidIDs := []string{
		"",              // 空字符串
		"short",         // 长度不足
		"invalid req",   // 包含空格
		"bad@request#1", // 特殊字符
	}

	for _, requestID := range invalidIDs {
		// WithRequestID会自动处理无效ID并记录错误
		logger := WithRequestID(requestID)
		logger.Warn("尝试使用无效RequestID处理请求")
	}

	fmt.Println("\n3. HTTP请求处理示例：")
	simulateHTTPRequestProcessing("POST", "/api/users", "http-req-20250803-001")
	simulateHTTPRequestProcessing("GET", "/api/orders", "http-req-20250803-002")

	fmt.Println("\n4. RequestID校验示例：")
	testRequestIDs := []string{
		"valid-request-12345678",
		"invalid req",
		"good_request_id",
		"bad@request",
	}

	for _, id := range testRequestIDs {
		isValid := IsValidRequestID(id)
		if isValid {
			WithRequestID(id).Infof("RequestID '%s' 通过校验", id)
		} else {
			Info(fmt.Sprintf("RequestID '%s' 校验失败", id))
		}
	}

	fmt.Println("\n5. 不安全版本示例（特殊场景）：")
	// 在某些特殊情况下，可能需要跳过校验
	externalRequestID := "external@system#request$123"
	WithRequestIDUnsafe(externalRequestID).
		WithField("source", "external_system").
		Info("来自外部系统的请求（使用不安全版本）")

	fmt.Println("WithRequestID方法演示完成")
}

// simulateHTTPRequestProcessing 模拟HTTP请求处理流程
func simulateHTTPRequestProcessing(method, path, requestID string) {
	// 使用WithRequestID创建带有请求ID的日志器
	requestLogger := WithRequestID(requestID)

	// 记录请求开始
	requestLogger.WithField("method", method).
		WithField("path", path).
		WithField("client_ip", "192.168.1.100").
		Info("HTTP请求开始")

	// 模拟中间件处理
	requestLogger.WithField("middleware", "auth").Debug("身份验证中间件")
	requestLogger.WithField("middleware", "cors").Debug("CORS中间件")

	// 模拟业务逻辑
	requestLogger.WithField("handler", "user_handler").Info("执行业务逻辑")

	// 模拟数据库操作
	requestLogger.WithField("operation", "db_query").
		WithField("table", "users").
		WithField("duration_ms", 25.3).
		Debug("数据库查询")

	// 记录请求完成
	requestLogger.WithField("status_code", 200).
		WithField("response_time_ms", 89.5).
		Info("HTTP请求完成")
}

// RunAllLogExamples 运行所有日志示例
func RunAllLogExamples() {
	fmt.Println("=================== 日志系统完整示例 ===================")

	SingletonLoggerUsageExample()
	MultiFormatLogExample()
	MultiOutputLogExample()
	CloudLogExample()
	PresetConfigExample()
	DynamicConfigExample()
	PerformanceLogExample()
	WithRequestIDExample()

	fmt.Println("\n================= 所有日志示例执行完成 =================")
}