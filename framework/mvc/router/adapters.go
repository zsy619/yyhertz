package router

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/mvc/define"
	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ===== 处理器适配器函数 =====
// 将各种处理器类型转换为标准 HandlerFunc

// adaptSimpleHandler 适配简单处理器
// SimpleHandlerFunc: func(context.Context) -> HandlerFunc: func(context.Context, *RequestContext)
func adaptSimpleHandler(handler define.SimpleHandlerFunc) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		handler(ctx)
	}
}

// adaptDirectHandler 适配直接处理器
// DirectHandlerFunc: func(*mvcContext.Context) -> HandlerFunc: func(context.Context, *RequestContext)
func adaptDirectHandler(handler define.DirectHandlerFunc) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		// 创建增强Context
		enhancedCtx := mvcContext.NewContextWithContext(c, ctx)
		defer enhancedCtx.Release()
		
		handler(enhancedCtx)
	}
}

// adaptLightHandler 适配轻量级处理器
// LightHandlerFunc: func() -> HandlerFunc: func(context.Context, *RequestContext)
func adaptLightHandler(handler define.LightHandlerFunc) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		handler()
	}
}

// adaptResponseHandler 适配响应处理器
// ResponseHandlerFunc: func(*mvcContext.Context) any -> HandlerFunc: func(context.Context, *RequestContext)
func adaptResponseHandler(handler define.ResponseHandlerFunc) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				// 错误恢复，返回500错误
				c.SetStatusCode(consts.StatusInternalServerError)
				c.SetContentType("application/json")
				errorResponse := map[string]any{
					"success": false,
					"error":   fmt.Sprintf("Handler panic: %v", r),
					"code":    500,
				}
				if data, err := json.Marshal(errorResponse); err == nil {
					c.Write(data)
				}
			}
		}()

		// 创建增强Context
		enhancedCtx := mvcContext.NewContextWithContext(c, ctx)
		defer enhancedCtx.Release()
		
		// 执行处理器获取响应数据
		result := handler(enhancedCtx)

		// 如果处理器已经设置了响应，则不再处理
		if c.Response.Header.ContentLength() > 0 {
			return
		}

		// 根据返回值类型设置响应
		if result == nil {
			// 返回空响应
			c.SetStatusCode(consts.StatusNoContent)
			return
		}

		// 设置JSON响应
		c.SetContentType("application/json")
		c.SetStatusCode(consts.StatusOK)

		// 序列化响应数据
		if data, err := json.Marshal(result); err != nil {
			// 序列化失败，返回错误
			c.SetStatusCode(consts.StatusInternalServerError)
			errorResponse := map[string]any{
				"success": false,
				"error":   "Failed to serialize response: " + err.Error(),
				"code":    500,
			}
			if errorData, jsonErr := json.Marshal(errorResponse); jsonErr == nil {
				c.Write(errorData)
			}
		} else {
			c.Write(data)
		}
	}
}

// adaptAsyncHandler 适配异步处理器
// AsyncHandlerFunc: func(*mvcContext.Context) <-chan any -> HandlerFunc: func(context.Context, *RequestContext)
func adaptAsyncHandler(handler define.AsyncHandlerFunc) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				// 错误恢复，返回500错误
				c.SetStatusCode(consts.StatusInternalServerError)
				c.SetContentType("application/json")
				errorResponse := map[string]any{
					"success": false,
					"error":   fmt.Sprintf("Async handler panic: %v", r),
					"code":    500,
				}
				if data, err := json.Marshal(errorResponse); err == nil {
					c.Write(data)
				}
			}
		}()

		// 创建增强Context
		enhancedCtx := mvcContext.NewContextWithContext(c, ctx)
		defer enhancedCtx.Release()
		
		// 执行异步处理器
		resultChan := handler(enhancedCtx)

		// 设置超时控制
		timeout := 30 * time.Second // 默认30秒超时
		if deadline, ok := ctx.Deadline(); ok {
			timeout = time.Until(deadline)
		}

		select {
		case result := <-resultChan:
			// 成功获取结果
			if result == nil {
				c.SetStatusCode(consts.StatusNoContent)
				return
			}

			// 设置JSON响应
			c.SetContentType("application/json")
			c.SetStatusCode(consts.StatusOK)

			if data, err := json.Marshal(result); err != nil {
				c.SetStatusCode(consts.StatusInternalServerError)
				errorResponse := map[string]any{
					"success": false,
					"error":   "Failed to serialize async response: " + err.Error(),
					"code":    500,
				}
				if errorData, jsonErr := json.Marshal(errorResponse); jsonErr == nil {
					c.Write(errorData)
				}
			} else {
				c.Write(data)
			}

		case <-time.After(timeout):
			// 超时处理
			c.SetStatusCode(consts.StatusRequestTimeout)
			c.SetContentType("application/json")
			errorResponse := map[string]any{
				"success": false,
				"error":   "Async operation timeout",
				"code":    408,
			}
			if data, err := json.Marshal(errorResponse); err == nil {
				c.Write(data)
			}

		case <-ctx.Done():
			// 上下文取消
			c.SetStatusCode(consts.StatusRequestTimeout)
			c.SetContentType("application/json")
			errorResponse := map[string]any{
				"success": false,
				"error":   "Request cancelled",
				"code":    499,
			}
			if data, err := json.Marshal(errorResponse); err == nil {
				c.Write(data)
			}
		}
	}
}

// adaptStreamHandler 适配流式处理器
// StreamHandlerFunc: func(*mvcContext.Context, chan<- []byte) error -> HandlerFunc: func(context.Context, *RequestContext)
func adaptStreamHandler(handler define.StreamHandlerFunc) define.HandlerFunc {
	return func(ctx context.Context, c *define.RequestContext) {
		defer func() {
			if r := recover(); r != nil {
				// 错误恢复，返回500错误
				c.SetStatusCode(consts.StatusInternalServerError)
				c.SetContentType("application/json")
				errorResponse := map[string]any{
					"success": false,
					"error":   fmt.Sprintf("Stream handler panic: %v", r),
					"code":    500,
				}
				if data, err := json.Marshal(errorResponse); err == nil {
					c.Write(data)
				}
			}
		}()

		// 创建增强Context
		enhancedCtx := mvcContext.NewContextWithContext(c, ctx)
		defer enhancedCtx.Release()
		
		// 创建数据流通道
		dataChan := make(chan []byte, 100) // 缓冲100个数据块
		defer close(dataChan)

		// 设置流式响应头
		c.SetContentType("application/octet-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.SetStatusCode(consts.StatusOK)

		// 在 goroutine 中执行流式处理器
		errChan := make(chan error, 1)
		go func() {
			defer close(errChan)
			if err := handler(enhancedCtx, dataChan); err != nil {
				errChan <- err
			}
		}()

		// 处理流式数据和错误
		for {
			select {
			case data, ok := <-dataChan:
				if !ok {
					// 数据通道关闭，流式传输完成
					return
				}
				// 写入数据块
				if _, err := c.Write(data); err != nil {
					// 写入失败，可能客户端断开连接
					return
				}

			case err := <-errChan:
				if err != nil {
					// 流式处理出错
					errorMsg := fmt.Sprintf("\n[ERROR]: %v\n", err)
					c.Write([]byte(errorMsg))
				}
				return

			case <-ctx.Done():
				// 上下文取消，停止流式传输
				cancelMsg := "\n[CANCELLED]: Request cancelled\n"
				c.Write([]byte(cancelMsg))
				return
			}
		}
	}
}

// ===== 处理器类型检查辅助函数 =====

// isHandlerFuncType 检查给定的函数是否为有效的处理器类型
func isHandlerFuncType(handler any) bool {
	switch handler.(type) {
	case define.HandlerFunc, define.SimpleHandlerFunc,
		define.LightHandlerFunc, define.ResponseHandlerFunc, define.AsyncHandlerFunc,
		define.StreamHandlerFunc:
		return true
	default:
		return false
	}
}

// getHandlerTypeName 获取处理器类型名称（用于调试和日志）
// 注意：由于 DirectHandlerFunc 与 HandlerFunc 是相同的类型别名，
// 在运行时无法区分，都会被识别为 HandlerFunc
func getHandlerTypeName(handler any) string {
	switch handler.(type) {
	case define.HandlerFunc:
		return "HandlerFunc"
	case define.SimpleHandlerFunc:
		return "SimpleHandlerFunc"
	case define.DirectHandlerFunc:
		return "DirectHandlerFunc"
	case define.LightHandlerFunc:
		return "LightHandlerFunc"
	case define.ResponseHandlerFunc:
		return "ResponseHandlerFunc"
	case define.AsyncHandlerFunc:
		return "AsyncHandlerFunc"
	case define.StreamHandlerFunc:
		return "StreamHandlerFunc"
	default:
		return "Unknown"
	}
}
