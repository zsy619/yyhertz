package define

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"

	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// 类型别名定义
type RequestContext = app.RequestContext

// HandlerFunc 定义处理函数类型 (基础处理器)
type HandlerFunc = func(context.Context, *RequestContext)

// type HandlerFunc = app.HandlerFunc

// SimpleHandlerFunc 简单处理器 - 只需要context，适用于简单逻辑
type SimpleHandlerFunc = func(context.Context)

// DirectHandlerFunc 直接处理器 - 直接访问增强Context，最简洁的处理器类型
type DirectHandlerFunc = mvcContext.HandlerFunc

// LightHandlerFunc 轻量级处理器 - 无参数，适用于静态响应和健康检查
type LightHandlerFunc = func()

// ResponseHandlerFunc 响应处理器 - 返回响应数据，适用于REST API
type ResponseHandlerFunc = func(*mvcContext.Context) any

// AsyncHandlerFunc 异步处理器 - 支持异步处理，适用于耗时操作
type AsyncHandlerFunc = func(*mvcContext.Context) <-chan any

// StreamHandlerFunc 流式处理器 - 支持流式响应，适用于大数据传输
type StreamHandlerFunc = func(*mvcContext.Context, chan<- []byte) error

// FilterFunc 过滤器函数类型
type FilterFunc = mvcContext.HandlerFunc
