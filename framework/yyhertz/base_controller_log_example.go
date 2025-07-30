package yyhertz

import (
	"fmt"
	"time"
)

// ExampleController 展示BaseController日志功能的示例控制器
type ExampleController struct {
	BaseController
}

// ShowLogExamples 展示各种日志使用方法
func (c *ExampleController) ShowLogExamples() {
	// ============= 1. 基础日志方法 =============

	c.LogInfo("用户请求处理开始")
	c.LogDebug("调试信息: 参数验证通过")
	c.LogWarn("警告: 缓存未命中，使用数据库查询")
	c.LogError("错误: 连接超时")

	// ============= 2. 带字段的结构化日志 =============

	c.LogWithFields("info", "用户操作记录", map[string]any{
		"user_id": "user123",
		"action":  "view_profile",
		"source":  "web",
	})

	// ============= 3. 请求/响应日志 =============

	// 记录请求开始
	c.LogRequest()

	// 模拟处理逻辑
	c.LogInfo("处理业务逻辑...")

	// 记录响应
	c.LogResponse(200, "处理成功")

	// ============= 4. 带请求ID的日志 =============

	c.LogWithRequestID("info", "开始处理用户数据")
	c.LogWithRequestID("debug", "验证用户权限")
	c.LogWithRequestID("info", "数据处理完成")

	// ============= 5. 带用户ID的日志 =============

	c.LogWithUserID("info", "用户登录成功", "user456")
	c.LogWithUserID("warn", "用户尝试访问受限资源", "user456")

	// ============= 6. 控制器动作日志 =============

	c.LogControllerAction("GetUserProfile", map[string]any{
		"user_id": "user123",
		"fields":  []string{"name", "email", "avatar"},
	})

	// ============= 7. 验证错误日志 =============

	validationErrors := []string{"邮箱格式不正确", "密码长度不足"}
	inputData := map[string]any{
		"email":    "invalid-email",
		"password": "123",
		"secret":   "should-be-hidden", // 敏感信息会被过滤
	}
	c.LogValidationError(validationErrors, inputData)

	// ============= 8. 数据库操作日志 =============

	// 模拟数据库操作
	start := time.Now()
	time.Sleep(50 * time.Millisecond) // 模拟数据库延迟
	duration := time.Since(start)

	// 成功操作
	c.LogDatabaseOperation("SELECT", "users", duration, nil)

	// 失败操作
	//c.LogDatabaseOperation("UPDATE", "users", duration, errors.New("connection timeout"))

	// ============= 9. 业务逻辑日志 =============

	c.LogBusinessLogic("user_profile_updated", map[string]any{
		"user_id":        "user123",
		"updated_fields": []string{"name", "phone"},
		"source":         "admin_panel",
		"operator":       "admin_user",
	})

	c.LogBusinessLogic("payment_processed", map[string]any{
		"order_id":       "order456",
		"amount":         99.99,
		"currency":       "USD",
		"payment_method": "credit_card",
		"gateway":        "stripe",
	})
}

// SimulateUserRegistration 模拟用户注册流程的日志记录
func (c *ExampleController) SimulateUserRegistration() {
	c.LogControllerAction("UserRegistration")

	// 记录请求开始
	c.LogRequest()

	// 1. 参数验证
	c.LogWithRequestID("info", "开始用户注册参数验证")

	// 模拟验证失败
	if c.GetString("email") == "" {
		c.LogValidationError([]string{"邮箱不能为空"}, map[string]any{
			"email": "",
			"name":  c.GetString("name"),
		})
		c.JSONBadRequest("参数验证失败")
		return
	}

	// 2. 检查用户是否存在
	c.LogWithRequestID("debug", "检查用户邮箱是否已存在")
	start := time.Now()
	time.Sleep(30 * time.Millisecond) // 模拟数据库查询
	c.LogDatabaseOperation("SELECT", "users", time.Since(start), nil)

	// 3. 创建用户
	c.LogBusinessLogic("user_registration_start", map[string]any{
		"email": c.GetString("email"),
		"name":  c.GetString("name"),
		"ip":    c.GetClientIP(),
	})

	start = time.Now()
	time.Sleep(100 * time.Millisecond) // 模拟用户创建
	c.LogDatabaseOperation("INSERT", "users", time.Since(start), nil)

	// 4. 发送欢迎邮件
	c.LogBusinessLogic("welcome_email_sent", map[string]any{
		"user_id":  "new_user_123",
		"email":    c.GetString("email"),
		"template": "welcome_email_v2",
	})

	// 5. 记录响应
	c.LogResponse(201, "用户注册成功")

	c.JSONOK("注册成功", map[string]any{
		"user_id": "new_user_123",
		"message": "欢迎邮件已发送",
	})
}

// SimulateAPIError 模拟API错误处理的日志记录
func (c *ExampleController) SimulateAPIError() {
	c.LogControllerAction("SimulateError")
	c.LogRequest()

	// 模拟不同类型的错误
	errorType := c.GetString("type", "validation")

	switch errorType {
	case "validation":
		c.LogValidationError([]string{"用户ID格式错误"}, map[string]any{
			"user_id": c.GetString("user_id"),
		})
		c.LogResponse(400, "参数验证失败")
		c.JSONBadRequest("参数验证失败")

	case "database":
		start := time.Now()
		time.Sleep(200 * time.Millisecond)
		c.LogDatabaseOperation("SELECT", "users", time.Since(start), fmt.Errorf("connection timeout"))
		c.LogResponse(500, "数据库连接超时")
		c.JSONInternalError("服务器内部错误")

	case "business":
		c.LogBusinessLogic("insufficient_balance", map[string]any{
			"user_id":   c.GetString("user_id"),
			"required":  100.0,
			"available": 50.0,
		})
		c.LogResponse(400, "余额不足")
		c.JSONBadRequest("余额不足")

	default:
		c.LogWithRequestID("error", "未知错误类型: "+errorType)
		c.LogResponse(400, "未知错误类型")
		c.JSONBadRequest("未知错误类型")
	}
}

// ShowAdvancedLogging 展示高级日志功能
func (c *ExampleController) ShowAdvancedLogging() {
	// 组合多种日志方法
	userID := c.GetString("user_id", "anonymous")

	// 同时记录用户ID和请求ID
	c.LogWithUserID("info", "高级功能访问", userID)
	c.LogWithRequestID("debug", "执行高级业务逻辑")

	// 复杂的业务逻辑日志
	c.LogBusinessLogic("advanced_operation", map[string]any{
		"user_id":   userID,
		"operation": "data_export",
		"filters": map[string]any{
			"date_range": "2024-01-01 to 2024-12-31",
			"categories": []string{"finance", "user_data"},
			"format":     "csv",
		},
		"estimated_records": 10000,
		"priority":          "high",
	})

	// 性能监控日志
	operationStart := time.Now()

	// 模拟复杂操作
	for i := 0; i < 3; i++ {
		stepStart := time.Now()
		time.Sleep(100 * time.Millisecond)

		c.LogDatabaseOperation(
			fmt.Sprintf("COMPLEX_QUERY_STEP_%d", i+1),
			"analytics",
			time.Since(stepStart),
			nil,
		)
	}

	totalDuration := time.Since(operationStart)

	// 记录总体性能
	c.LogBusinessLogic("advanced_operation_completed", map[string]any{
		"user_id":         userID,
		"total_duration":  totalDuration.String(),
		"duration_ms":     totalDuration.Milliseconds(),
		"steps_completed": 3,
		"status":          "success",
	})

	c.LogResponse(200, "高级操作完成")
	c.JSONOK("操作完成", map[string]any{
		"duration": totalDuration.String(),
		"records":  10000,
	})
}
