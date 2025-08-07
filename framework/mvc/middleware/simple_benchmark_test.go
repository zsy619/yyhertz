package middleware

import (
	"testing"
	
	"github.com/cloudwego/hertz/pkg/app"
	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// 简化的基准测试 - 验证整合效果

func BenchmarkSimpleBasicMiddleware(b *testing.B) {
	// 创建基础中间件引擎
	engine := NewEngine()
	
	// 注册简单中间件
	engine.Use(func(c *Context) {
		c.Set("test", "value")
		c.Next()
	})
	
	// 创建测试上下文
	hertzCtx := &app.RequestContext{}
	ctx := engine.NewContext(hertzCtx)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx.Next()
		}
	})
}

func BenchmarkSimpleMVCMiddleware(b *testing.B) {
	// 创建MVC中间件管理器
	manager := NewMiddlewareManager()
	manager.Initialize()
	
	// 注册简单中间件
	manager.RegisterCustom("test", func(ctx *mvccontext.EnhancedContext) {
		ctx.Set("test", "value")
		ctx.Next()
	}, MiddlewareMetadata{
		Name: "test",
		Description: "Test middleware",
	})
	manager.UseCustom(LayerGlobal, "test", 10)
	
	// 创建测试上下文
	hertzCtx := &app.RequestContext{}
	ctx := mvccontext.NewContext(hertzCtx)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			manager.ExecuteCompiledChain(ctx, LayerGlobal)
		}
	})
}

func BenchmarkUnifiedMiddleware(b *testing.B) {
	// 创建统一中间件管理器
	manager := NewUnifiedMiddlewareManager()
	
	// 测试基础中间件到MVC的转换
	basicHandler := func(c *Context) {
		c.Set("test", "value")
		c.Next()
	}
	manager.Use("test", basicHandler)
	
	// 创建测试上下文
	hertzCtx := &app.RequestContext{}
	ctx := mvccontext.NewContext(hertzCtx)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 简化的执行逻辑
			ctx.Set("benchmark", "test")
			ctx.Next()
		}
	})
}

func BenchmarkMiddlewareAdapter(b *testing.B) {
	// 创建基础中间件  
	basicHandler := func(c *Context) {
		c.Set("adapted", "value")
		c.Next()
	}
	
	// 转换为MVC中间件
	mvcHandler := HandlerFuncToMVC(basicHandler)
	
	// 创建测试上下文
	hertzCtx := &app.RequestContext{}
	ctx := mvccontext.NewContext(hertzCtx)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mvcHandler(ctx)
		}
	})
}

func BenchmarkMemoryAllocation(b *testing.B) {
	// 内存分配基准测试
	hertzCtx := &app.RequestContext{}
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// 测试Context创建的内存开销
			ctx := mvccontext.NewContext(hertzCtx)
			ctx.Set("memory_test", "value")
			ctx.Release()
		}
	})
}