package core

import (
	"fmt"
	"net/http"

	contextenhanced "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/define"
	"github.com/zsy619/yyhertz/framework/mvc/errors"
)

// ============= 错误处理方法 =============

// SetErrorConfig 设置错误配置
func (app *App) SetErrorConfig(config ErrorConfig) {
	app.errorMutex.Lock()
	defer app.errorMutex.Unlock()
	app.errorConfig = config
}

// GetErrorConfig 获取错误配置
func (app *App) GetErrorConfig() ErrorConfig {
	app.errorMutex.RLock()
	defer app.errorMutex.RUnlock()
	return app.errorConfig
}

// SetErrorRegistry 设置错误注册器
func (app *App) SetErrorRegistry(registry *ErrorRegistry) {
	app.errorMutex.Lock()
	defer app.errorMutex.Unlock()
	app.errorRegistry = registry
}

// GetErrorRegistry 获取错误注册器
func (app *App) GetErrorRegistry() *ErrorRegistry {
	app.errorMutex.RLock()
	defer app.errorMutex.RUnlock()
	return app.errorRegistry
}

// ErrorHandler 注册函数错误处理器（类似Beego的ErrorHandler）
func (app *App) ErrorHandler(statusCode int, handler ErrorHandlerFunc) error {
	// 使用全局错误注册器
	registry := errors.GetGlobalErrorRegistry()
	if registry == nil {
		return fmt.Errorf("global error registry not available")
	}

	// 简化处理，直接使用全局注册器
	return errors.RegisterErrorHandlerFunc(statusCode, "app-handler", 100,
		func(statusCode int, err error) bool { return true },
		func(ctx *errors.Context, statusCode int, err error) error {
			// 上下文适配暂时简化处理
			return nil
		})
}

// CustomErrorHandler 注册自定义错误处理器
func (app *App) CustomErrorHandler(handler func(ctx *define.RequestContext, statusCode int, err error) error) error {
	// 使用全局错误注册器
	return errors.RegisterErrorHandlerFunc(500, "custom-handler", 100,
		func(statusCode int, err error) bool { return true },
		func(ctx *errors.Context, statusCode int, err error) error {
			// 需要进行上下文适配
			return nil // 简化处理
		})
}

// TriggerError 触发错误处理
func (app *App) TriggerError(ctx *define.RequestContext, statusCode int, err error) error {
	// 直接使用全局错误注册器处理错误
	registry := errors.GetGlobalErrorRegistry()
	if registry == nil {
		// 如果注册器不可用，使用基础错误处理
		return app.handleBasicError(ctx, statusCode, err)
	}

	// 创建增强的MVC Context
	mvcCtx := createEnhancedContext(ctx)

	// 委托给全局错误处理系统
	return registry.HandleError(mvcCtx, statusCode, err)
}

// createEnhancedContext 创建增强的MVC上下文
func createEnhancedContext(reqCtx *define.RequestContext) *errors.Context {
	// 通过mvc/context包创建增强上下文，然后转换为errors.Context
	mvcCtx := contextenhanced.NewContext(reqCtx)
	return (*errors.Context)(mvcCtx)
}

// handleBasicError 基础错误处理
func (app *App) handleBasicError(ctx *define.RequestContext, statusCode int, err error) error {
	// 简单的JSON错误响应
	ctx.JSON(statusCode, map[string]any{
		"code":    statusCode,
		"message": fmt.Sprintf("Error %d occurred", statusCode),
		"error":   fmt.Sprintf("%v", err),
	})
	return nil
}

// ============= Beego风格的自动错误处理 =============

// setupAutoErrorHandling 设置自动错误处理（类似Beego）
func (app *App) setupAutoErrorHandling() {
	// 设置NoRoute处理器 - 当路由不匹配时自动触发404错误处理
	app.NoRoute(func(ctx *contextenhanced.Context, c *define.RequestContext) {
		// 自动触发404错误处理
		app.TriggerError(c, http.StatusNotFound, fmt.Errorf("route not found: %s", string(c.Path())))
	})

	// 设置NoMethod处理器 - 当方法不被允许时自动触发405错误处理
	app.NoMethod(func(ctx *contextenhanced.Context, c *define.RequestContext) {
		// 自动触发405错误处理
		app.TriggerError(c, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed: %s", string(c.Method())))
	})
}

// EnableAutoErrorHandling 启用自动错误处理（类似Beego的ErrorController）
func (app *App) EnableAutoErrorHandling() *App {
	app.setupAutoErrorHandling()
	return app
}

// Abort 中止请求并触发错误处理（类似Beego的Abort方法）
func (app *App) Abort(ctx *define.RequestContext, statusCode int, message ...string) {
	var err error
	if len(message) > 0 {
		err = fmt.Errorf("%s", message[0])
	} else {
		err = fmt.Errorf("request aborted with status %d", statusCode)
	}

	// 触发错误处理
	app.TriggerError(ctx, statusCode, err)
}

// AbortWithError 中止请求并使用指定错误（类似Beego的Abort）
func (app *App) AbortWithError(ctx *define.RequestContext, statusCode int, err error) {
	app.TriggerError(ctx, statusCode, err)
}
