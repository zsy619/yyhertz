package errors

import (
	"fmt"
	"log"
	"os"
)

// =============================================================================
// 模块：配置管理和工厂方法
// 职责：提供配置结构、默认配置和便捷的控制器创建方法
// =============================================================================

// DefaultErrorControllerConfig 返回默认错误控制器配置
func DefaultErrorControllerConfig() ErrorControllerConfig {
	return ErrorControllerConfig{
		ShowDetailedError: true,
		Language:          DefaultLanguage,
		CustomTitle:       DefaultTitle,
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   true,
	}
}

// SetErrorControllerConfig 配置错误控制器行为
func SetErrorControllerConfig(controller *DefaultErrorController, config ErrorControllerConfig) {
	if controller == nil {
		return
	}

	controller.ShowDetailedError = config.ShowDetailedError
	controller.Language = config.Language
	controller.CustomTitle = config.CustomTitle
	controller.SupportEmail = config.SupportEmail
	controller.SupportPhone = config.SupportPhone
	controller.EnableDebugInfo = config.EnableDebugInfo
}

// EnableErrorDebugging 开启/关闭错误调试模式
func EnableErrorDebugging(controller *DefaultErrorController, enabled bool) {
	if controller == nil {
		return
	}

	controller.EnableDebugInfo = enabled
	controller.ShowDetailedError = enabled
}

// CreateDefaultErrorHandlers 创建一套默认错误处理器
func CreateDefaultErrorHandlers() map[int]ErrorHandler {
	handlers := make(map[int]ErrorHandler)

	// 创建默认控制器
	controller := NewDefaultErrorController()

	// 为常见状态码创建专用处理器
	handlers[401] = controller
	handlers[403] = controller
	handlers[404] = controller
	handlers[500] = controller

	return handlers
}

// RegisterDefaultHandlers 注册默认错误处理器到注册器
func RegisterDefaultHandlers(registry *ErrorRegistry) error {
	if registry == nil {
		return fmt.Errorf("registry cannot be nil")
	}

	handlers := CreateDefaultErrorHandlers()
	for statusCode, handler := range handlers {
		if err := registry.RegisterHandler(statusCode, handler); err != nil {
			return fmt.Errorf("failed to register handler for status %d: %w", statusCode, err)
		}
	}

	return nil
}

// QuickSetupDefaultHandlers 快速设置默认错误处理器
func QuickSetupDefaultHandlers(registry *ErrorRegistry, env string) error {
	if registry == nil {
		return fmt.Errorf("registry cannot be nil")
	}

	var controller *DefaultErrorController

	// 根据环境创建相应的控制器
	switch env {
	case "production", "prod":
		controller = NewProductionErrorController()
	case "development", "dev":
		controller = NewDevelopmentErrorController()
	default:
		controller = NewDefaultErrorController()
	}

	// 注册更多的错误状态码
	statusCodes := []int{400, 401, 402, 403, 404, 405, 406, 408, 409, 410, 413, 415, 418, 422, 429, 500, 501, 502, 503, 504, 505}
	for _, statusCode := range statusCodes {
		if err := registry.RegisterHandler(statusCode, controller); err != nil {
			return fmt.Errorf("failed to register handler for status %d: %w", statusCode, err)
		}
	}

	return nil
}

// NewProductionErrorController 创建生产环境错误控制器
func NewProductionErrorController() *DefaultErrorController {
	controller := &DefaultErrorController{
		ShowDetailedError: false, // 生产环境不显示详细错误
		Language:          DefaultLanguage,
		CustomTitle:       DefaultTitle,
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   false, // 生产环境不显示调试信息

		// 生产环境配置
		EnableErrorLogging: true,
		ErrorLogger:        log.New(os.Stderr, "[PROD-ERROR] ", log.LstdFlags),
		ErrorStatistics:    NewErrorStatistics(),
		CustomTemplates:    make(map[int]string),
		ErrorHooks:         make(map[string][]ErrorHook),
		RetryableErrors:    InitRetryableErrors(),
	}

	controller.initDefaultHooks()
	return controller
}

// NewDevelopmentErrorController 创建开发环境错误控制器
func NewDevelopmentErrorController() *DefaultErrorController {
	controller := &DefaultErrorController{
		ShowDetailedError: true, // 开发环境显示详细错误
		Language:          DefaultLanguage,
		CustomTitle:       DefaultTitle + " - 开发环境",
		SupportEmail:      "",
		SupportPhone:      "",
		EnableDebugInfo:   true, // 开发环境显示调试信息

		// 开发环境配置
		EnableErrorLogging: true,
		ErrorLogger:        log.New(os.Stdout, "[DEV-ERROR] ", log.LstdFlags|log.Lshortfile),
		ErrorStatistics:    NewErrorStatistics(),
		CustomTemplates:    make(map[int]string),
		ErrorHooks:         make(map[string][]ErrorHook),
		RetryableErrors:    InitRetryableErrors(),
	}

	controller.initDefaultHooks()
	return controller
}