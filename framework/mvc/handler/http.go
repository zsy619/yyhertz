package handler

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
)

// HTTPGenericHandler HTTP泛型处理器接口
type HTTPGenericHandler[T any, R any] interface {
	GenericHandler[T, R]
	
	// ParseRequest 解析HTTP请求
	ParseRequest(ctx *app.RequestContext) (T, error)
	
	// WriteResponse 写入HTTP响应
	WriteResponse(ctx *app.RequestContext, result *GenericResult[R])
}

// HTTPGenericRequest 通用HTTP请求结构
type HTTPGenericRequest[T any] struct {
	Data      T      `json:"data"`
	RequestID string `json:"request_id,omitempty"`
}

// HTTPGenericResponse 通用HTTP响应结构
type HTTPGenericResponse[T any] struct {
	Success   bool   `json:"success"`
	Data      T      `json:"data,omitempty"`
	Error     string `json:"error,omitempty"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

// BaseHTTPGenericHandler 基础HTTP泛型处理器
type BaseHTTPGenericHandler[T any, R any] struct {
	*BaseGenericHandler[T, R]
	
	// HTTP特定配置
	ValidateJSON bool
	ContentType  string
}

// NewBaseHTTPGenericHandler 创建基础HTTP泛型处理器
func NewBaseHTTPGenericHandler[T any, R any](name string) *BaseHTTPGenericHandler[T, R] {
	return &BaseHTTPGenericHandler[T, R]{
		BaseGenericHandler: NewBaseGenericHandler[T, R](name),
		ValidateJSON:       true,
		ContentType:        "application/json",
	}
}

// ParseRequest 解析HTTP请求
func (h *BaseHTTPGenericHandler[T, R]) ParseRequest(ctx *app.RequestContext) (T, error) {
	var input T
	
	// 尝试解析JSON请求体
	body, err := ctx.Body()
	if err != nil {
		return input, fmt.Errorf("读取请求体失败: %w", err)
	}
	
	if len(body) == 0 {
		// 如果没有请求体，返回零值
		return input, nil
	}
	
	// 解析JSON
	if err := json.Unmarshal(body, &input); err != nil {
		return input, fmt.Errorf("JSON解析失败: %w", err)
	}
	
	return input, nil
}

// WriteResponse 写入HTTP响应
func (h *BaseHTTPGenericHandler[T, R]) WriteResponse(ctx *app.RequestContext, result *GenericResult[R]) {
	response := &HTTPGenericResponse[R]{
		Success: result.Success,
		Data:    result.Data,
		Error:   result.Error,
		Code:    result.Code,
		Message: result.Message,
	}
	
	// 设置响应头
	ctx.Header("Content-Type", h.ContentType)
	
	// 根据结果设置HTTP状态码
	statusCode := 200
	if !result.Success {
		statusCode = 400 // 业务错误返回400
	}
	
	ctx.JSON(statusCode, response)
}

// HandleHTTP 处理HTTP请求的完整流程
func (h *BaseHTTPGenericHandler[T, R]) HandleHTTP(ctx *app.RequestContext) {
	// 解析请求
	input, err := h.ParseRequest(ctx)
	if err != nil {
		h.WriteResponse(ctx, NewGenericError[R](err))
		return
	}
	
	// 处理请求
	httpCtx := context.Background()
	result := h.ProcessWithResult(httpCtx, input)
	
	// 写入响应
	h.WriteResponse(ctx, result)
}

// SimpleHTTPHandler 简单的HTTP处理器
type SimpleHTTPHandler[T any, R any] struct {
	name      string
	processor func(context.Context, T) (R, error)
}

// NewSimpleHTTPHandler 创建简单HTTP处理器
func NewSimpleHTTPHandler[T any, R any](name string, processor func(context.Context, T) (R, error)) *SimpleHTTPHandler[T, R] {
	return &SimpleHTTPHandler[T, R]{
		name:      name,
		processor: processor,
	}
}

// HandleHTTP 处理HTTP请求
func (h *SimpleHTTPHandler[T, R]) HandleHTTP(ctx *app.RequestContext) {
	// 解析请求
	var input T
	body, err := ctx.Body()
	if err == nil && len(body) > 0 {
		json.Unmarshal(body, &input)
	}
	
	// 处理请求
	httpCtx := context.Background()
	result, err := h.processor(httpCtx, input)
	
	// 构建响应
	response := &HTTPGenericResponse[R]{
		Success: err == nil,
		Message: "操作完成",
	}
	
	if err != nil {
		response.Error = err.Error()
		response.Code = -1
		response.Message = "操作失败"
		ctx.JSON(400, response)
	} else {
		response.Data = result
		response.Code = 0
		ctx.JSON(200, response)
	}
}