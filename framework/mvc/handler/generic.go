package handler

import (
	"context"
	"fmt"
)

// GenericHandler 简化的泛型处理器接口
type GenericHandler[T any, R any] interface {
	// Handle 处理请求的核心方法
	Handle(ctx context.Context, input T) (R, error)
	
	// Validate 验证输入数据
	Validate(input T) error
	
	// GetName 获取处理器名称
	GetName() string
}

// GenericResult 泛型处理结果
type GenericResult[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewGenericResult 创建成功结果
func NewGenericResult[T any](data T) *GenericResult[T] {
	return &GenericResult[T]{
		Success: true,
		Data:    data,
		Code:    0,
		Message: "操作成功",
	}
}

// NewGenericError 创建错误结果
func NewGenericError[T any](err error) *GenericResult[T] {
	return &GenericResult[T]{
		Success: false,
		Error:   err.Error(),
		Code:    -1,
		Message: "操作失败",
	}
}

// BaseGenericHandler 基础泛型处理器实现
type BaseGenericHandler[T any, R any] struct {
	name      string
	validator func(T) error
	processor func(context.Context, T) (R, error)
}

// NewBaseGenericHandler 创建基础泛型处理器
func NewBaseGenericHandler[T any, R any](name string) *BaseGenericHandler[T, R] {
	return &BaseGenericHandler[T, R]{
		name: name,
	}
}

// GetName 获取处理器名称
func (h *BaseGenericHandler[T, R]) GetName() string {
	return h.name
}

// WithValidator 设置验证器
func (h *BaseGenericHandler[T, R]) WithValidator(validator func(T) error) *BaseGenericHandler[T, R] {
	h.validator = validator
	return h
}

// WithProcessor 设置处理器
func (h *BaseGenericHandler[T, R]) WithProcessor(processor func(context.Context, T) (R, error)) *BaseGenericHandler[T, R] {
	h.processor = processor
	return h
}

// Validate 验证输入数据
func (h *BaseGenericHandler[T, R]) Validate(input T) error {
	if h.validator != nil {
		return h.validator(input)
	}
	return nil
}

// Handle 处理请求
func (h *BaseGenericHandler[T, R]) Handle(ctx context.Context, input T) (R, error) {
	var zero R
	
	// 验证输入
	if err := h.Validate(input); err != nil {
		return zero, fmt.Errorf("验证失败: %w", err)
	}
	
	// 执行处理逻辑
	if h.processor != nil {
		return h.processor(ctx, input)
	}
	
	return zero, fmt.Errorf("处理器 %s 未设置处理逻辑", h.name)
}

// ProcessWithResult 处理请求并返回GenericResult
func (h *BaseGenericHandler[T, R]) ProcessWithResult(ctx context.Context, input T) *GenericResult[R] {
	result, err := h.Handle(ctx, input)
	if err != nil {
		return NewGenericError[R](err)
	}
	return NewGenericResult(result)
}

// SimpleHandler 简单的函数式处理器
type SimpleHandler[T any, R any] struct {
	name      string
	processor func(context.Context, T) (R, error)
}

// NewSimpleHandler 创建简单处理器
func NewSimpleHandler[T any, R any](name string, processor func(context.Context, T) (R, error)) *SimpleHandler[T, R] {
	return &SimpleHandler[T, R]{
		name:      name,
		processor: processor,
	}
}

// GetName 获取处理器名称
func (h *SimpleHandler[T, R]) GetName() string {
	return h.name
}

// Validate 简单处理器默认不验证
func (h *SimpleHandler[T, R]) Validate(input T) error {
	return nil
}

// Handle 处理请求
func (h *SimpleHandler[T, R]) Handle(ctx context.Context, input T) (R, error) {
	if h.processor != nil {
		return h.processor(ctx, input)
	}
	var zero R
	return zero, fmt.Errorf("处理器 %s 未设置处理逻辑", h.name)
}