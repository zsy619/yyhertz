package router

import (
	"context"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/stretchr/testify/assert"

	"github.com/zsy619/yyhertz/framework/mvc/define"
	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// TestAdapters 测试适配器函数
func TestAdapters(t *testing.T) {
	// 测试SimpleHandlerFunc适配器
	t.Run("SimpleHandler", func(t *testing.T) {
		called := false
		simpleHandler := func(ctx context.Context) {
			called = true
		}

		adaptedHandler := adaptSimpleHandler(simpleHandler)
		c := &app.RequestContext{}

		adaptedHandler(context.Background(), c)
		assert.True(t, called, "Simple handler should be called")
	})

	// 测试LightHandlerFunc适配器
	t.Run("LightHandler", func(t *testing.T) {
		called := false
		lightHandler := func() {
			called = true
		}

		adaptedHandler := adaptLightHandler(lightHandler)
		c := &app.RequestContext{}

		adaptedHandler(context.Background(), c)
		assert.True(t, called, "Light handler should be called")
	})

	// 测试DirectHandlerFunc适配器
	t.Run("DirectHandler", func(t *testing.T) {
		called := false
		directHandler := func(c *mvcContext.Context) {
			called = true
			// 验证增强Context的功能
			c.Set("test", "value")
			if val, exists := c.Get("test"); exists {
				assert.Equal(t, "value", val)
			}
		}

		adaptedHandler := adaptDirectHandler(directHandler)
		c := &app.RequestContext{}

		adaptedHandler(context.Background(), c)
		assert.True(t, called, "Direct handler should be called")
	})

	// 测试ResponseHandlerFunc适配器
	t.Run("ResponseHandler", func(t *testing.T) {
		responseHandler := func(c *mvcContext.Context) any {
			// 测试增强Context的方法
			c.Set("handled", true)
			return map[string]string{"message": "test"}
		}

		adaptedHandler := adaptResponseHandler(responseHandler)
		c := &app.RequestContext{}

		adaptedHandler(context.Background(), c)

		// 验证响应被设置
		assert.True(t, len(c.Response.Body()) > 0, "Response body should be set")
	})

	// 测试AsyncHandlerFunc适配器
	t.Run("AsyncHandler", func(t *testing.T) {
		asyncHandler := func(c *mvcContext.Context) <-chan any {
			resultChan := make(chan any, 1)
			go func() {
				defer close(resultChan)
				// 测试增强Context的并发操作
				c.Set("async_processed", true)
				resultChan <- "async result"
			}()
			return resultChan
		}

		adaptedHandler := adaptAsyncHandler(asyncHandler)
		c := &app.RequestContext{}

		adaptedHandler(context.Background(), c)

		// 验证响应被设置
		assert.True(t, len(c.Response.Body()) > 0, "Async response body should be set")
	})

	// 测试StreamHandlerFunc适配器
	t.Run("StreamHandler", func(t *testing.T) {
		streamHandler := func(c *mvcContext.Context, dataChan chan<- []byte) error {
			// 测试增强Context的方法
			c.Set("stream_start", true)
			dataChan <- []byte("stream data")
			return nil
		}

		adaptedHandler := adaptStreamHandler(streamHandler)
		c := &app.RequestContext{}

		adaptedHandler(context.Background(), c)

		// 验证流式响应
		assert.True(t, len(c.Response.Body()) > 0, "Stream response body should be set")
		assert.Contains(t, string(c.Response.Body()), "stream data")
	})
}

// TestHandlerTypeChecks 测试处理器类型检查函数
func TestHandlerTypeChecks(t *testing.T) {
	// 测试有效的处理器类型
	simpleHandler := define.SimpleHandlerFunc(func(ctx context.Context) {})
	assert.True(t, isHandlerFuncType(simpleHandler), "SimpleHandlerFunc should be valid")

	lightHandler := define.LightHandlerFunc(func() {})
	assert.True(t, isHandlerFuncType(lightHandler), "LightHandlerFunc should be valid")

	responseHandler := define.ResponseHandlerFunc(func(c *mvcContext.Context) any { return nil })
	assert.True(t, isHandlerFuncType(responseHandler), "ResponseHandlerFunc should be valid")

	// 测试无效的处理器类型
	invalidHandler := "not a handler"
	assert.False(t, isHandlerFuncType(invalidHandler), "String should not be valid handler type")
}

// TestHandlerTypeNames 测试处理器类型名称获取
func TestHandlerTypeNames(t *testing.T) {
	simpleHandler := define.SimpleHandlerFunc(func(ctx context.Context) {})
	assert.Equal(t, "SimpleHandlerFunc", getHandlerTypeName(simpleHandler))

	lightHandler := define.LightHandlerFunc(func() {})
	assert.Equal(t, "LightHandlerFunc", getHandlerTypeName(lightHandler))

	unknownHandler := "unknown"
	assert.Equal(t, "Unknown", getHandlerTypeName(unknownHandler))
}

// BenchmarkAdapters 性能基准测试
func BenchmarkAdapters(b *testing.B) {
	ctx := context.Background()
	c := &app.RequestContext{}

	b.Run("SimpleHandler", func(b *testing.B) {
		handler := adaptSimpleHandler(func(ctx context.Context) {})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handler(ctx, c)
		}
	})

	b.Run("LightHandler", func(b *testing.B) {
		handler := adaptLightHandler(func() {})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handler(ctx, c)
		}
	})

	b.Run("DirectHandler", func(b *testing.B) {
		handler := adaptDirectHandler(func(c *mvcContext.Context) {})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handler(ctx, c)
		}
	})

	b.Run("ResponseHandler", func(b *testing.B) {
		handler := adaptResponseHandler(func(c *mvcContext.Context) any {
			return "test"
		})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			handler(ctx, c)
		}
	})
}
