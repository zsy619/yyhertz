package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// GenericHandler 泛型通用处理器接口
type GenericHandler[T any, R any] interface {
	// Handle 核心处理方法
	Handle(ctx context.Context, input T) (*GenericResult[R], error)
	
	// Validate 验证输入参数
	Validate(input T) error
	
	// PreProcess 预处理
	PreProcess(ctx context.Context, input T) (T, error)
	
	// PostProcess 后处理
	PostProcess(ctx context.Context, result R) (R, error)
}

// GenericResult 通用结果结构
type GenericResult[T any] struct {
	Data      T             `json:"data"`
	Success   bool          `json:"success"`
	Message   string        `json:"message"`
	Code      int           `json:"code"`
	RequestID string        `json:"request_id,omitempty"`
	Timestamp int64         `json:"timestamp"`
	Meta      map[string]any `json:"meta,omitempty"`
}

// NewGenericResult 创建通用结果
func NewGenericResult[T any](data T, success bool, message string, code int) *GenericResult[T] {
	return &GenericResult[T]{
		Data:      data,
		Success:   success,
		Message:   message,
		Code:      code,
		Timestamp: time.Now().Unix(),
		Meta:      make(map[string]any),
	}
}

// NewSuccessResult 创建成功结果
func NewSuccessResult[T any](data T, message string) *GenericResult[T] {
	return NewGenericResult(data, true, message, 200)
}

// NewErrorResult 创建错误结果
func NewErrorResult[T any](message string, code int) *GenericResult[T] {
	var zero T
	return NewGenericResult(zero, false, message, code)
}

// BaseGenericHandler 基础泛型处理器实现
type BaseGenericHandler[T any, R any] struct {
	Name        string
	Description string
	validator   func(T) error
	preProcessor func(context.Context, T) (T, error)
	processor   func(context.Context, T) (R, error)
	postProcessor func(context.Context, R) (R, error)
}

// NewBaseGenericHandler 创建基础泛型处理器
func NewBaseGenericHandler[T any, R any](name string) *BaseGenericHandler[T, R] {
	return &BaseGenericHandler[T, R]{
		Name:        name,
		Description: fmt.Sprintf("Generic handler for %s", name),
	}
}

// WithValidator 设置验证器
func (h *BaseGenericHandler[T, R]) WithValidator(validator func(T) error) *BaseGenericHandler[T, R] {
	h.validator = validator
	return h
}

// WithPreProcessor 设置预处理器
func (h *BaseGenericHandler[T, R]) WithPreProcessor(preProcessor func(context.Context, T) (T, error)) *BaseGenericHandler[T, R] {
	h.preProcessor = preProcessor
	return h
}

// WithProcessor 设置核心处理器
func (h *BaseGenericHandler[T, R]) WithProcessor(processor func(context.Context, T) (R, error)) *BaseGenericHandler[T, R] {
	h.processor = processor
	return h
}

// WithPostProcessor 设置后处理器
func (h *BaseGenericHandler[T, R]) WithPostProcessor(postProcessor func(context.Context, R) (R, error)) *BaseGenericHandler[T, R] {
	h.postProcessor = postProcessor
	return h
}

// Handle 实现核心处理逻辑
func (h *BaseGenericHandler[T, R]) Handle(ctx context.Context, input T) (*GenericResult[R], error) {
	start := time.Now()
	
	// 记录处理开始
	config.WithFields(map[string]any{
		"handler": h.Name,
		"input_type": reflect.TypeOf(input).String(),
		"start_time": start.Format(time.RFC3339),
	}).Debug("Generic handler processing started")

	// 1. 验证输入
	if err := h.Validate(input); err != nil {
		config.WithFields(map[string]any{
			"handler": h.Name,
			"error": err.Error(),
			"duration": time.Since(start).String(),
		}).Warn("Generic handler validation failed")
		return NewErrorResult[R]("Validation failed: " + err.Error(), 400), err
	}

	// 2. 预处理
	processedInput, err := h.PreProcess(ctx, input)
	if err != nil {
		config.WithFields(map[string]any{
			"handler": h.Name,
			"error": err.Error(),
			"duration": time.Since(start).String(),
		}).Error("Generic handler pre-processing failed")
		return NewErrorResult[R]("Pre-processing failed: " + err.Error(), 500), err
	}

	// 3. 核心处理
	if h.processor == nil {
		err := errors.New("processor not configured")
		config.WithFields(map[string]any{
			"handler": h.Name,
			"error": err.Error(),
		}).Error("Generic handler processor not configured")
		return NewErrorResult[R]("Processor not configured", 500), err
	}

	result, err := h.processor(ctx, processedInput)
	if err != nil {
		config.WithFields(map[string]any{
			"handler": h.Name,
			"error": err.Error(),
			"duration": time.Since(start).String(),
		}).Error("Generic handler processing failed")
		return NewErrorResult[R]("Processing failed: " + err.Error(), 500), err
	}

	// 4. 后处理
	finalResult, err := h.PostProcess(ctx, result)
	if err != nil {
		config.WithFields(map[string]any{
			"handler": h.Name,
			"error": err.Error(),
			"duration": time.Since(start).String(),
		}).Error("Generic handler post-processing failed")
		return NewErrorResult[R]("Post-processing failed: " + err.Error(), 500), err
	}

	duration := time.Since(start)
	
	// 记录处理成功
	config.WithFields(map[string]any{
		"handler": h.Name,
		"duration": duration.String(),
		"duration_ms": duration.Milliseconds(),
		"success": true,
	}).Info("Generic handler processing completed successfully")

	successResult := NewSuccessResult(finalResult, "Processing completed successfully")
	successResult.Meta["duration"] = duration.String()
	successResult.Meta["handler"] = h.Name
	
	return successResult, nil
}

// Validate 实现验证逻辑
func (h *BaseGenericHandler[T, R]) Validate(input T) error {
	if h.validator != nil {
		return h.validator(input)
	}
	return nil
}

// PreProcess 实现预处理逻辑
func (h *BaseGenericHandler[T, R]) PreProcess(ctx context.Context, input T) (T, error) {
	if h.preProcessor != nil {
		return h.preProcessor(ctx, input)
	}
	return input, nil
}

// PostProcess 实现后处理逻辑
func (h *BaseGenericHandler[T, R]) PostProcess(ctx context.Context, result R) (R, error) {
	if h.postProcessor != nil {
		return h.postProcessor(ctx, result)
	}
	return result, nil
}

// GenericChain 泛型处理链
type GenericChain[T any] struct {
	handlers []func(context.Context, T) (T, error)
	name     string
}

// NewGenericChain 创建泛型处理链
func NewGenericChain[T any](name string) *GenericChain[T] {
	return &GenericChain[T]{
		handlers: make([]func(context.Context, T) (T, error), 0),
		name:     name,
	}
}

// Add 添加处理器到链中
func (c *GenericChain[T]) Add(handler func(context.Context, T) (T, error)) *GenericChain[T] {
	c.handlers = append(c.handlers, handler)
	return c
}

// Execute 执行处理链
func (c *GenericChain[T]) Execute(ctx context.Context, input T) (T, error) {
	start := time.Now()
	current := input
	
	config.WithFields(map[string]any{
		"chain": c.name,
		"handlers_count": len(c.handlers),
	}).Debug("Generic chain execution started")

	for i, handler := range c.handlers {
		stepStart := time.Now()
		result, err := handler(ctx, current)
		stepDuration := time.Since(stepStart)
		
		if err != nil {
			config.WithFields(map[string]any{
				"chain": c.name,
				"step": i + 1,
				"error": err.Error(),
				"step_duration": stepDuration.String(),
				"total_duration": time.Since(start).String(),
			}).Error("Generic chain step failed")
			return current, fmt.Errorf("chain step %d failed: %w", i+1, err)
		}
		
		config.WithFields(map[string]any{
			"chain": c.name,
			"step": i + 1,
			"step_duration": stepDuration.String(),
		}).Debug("Generic chain step completed")
		
		current = result
	}

	totalDuration := time.Since(start)
	config.WithFields(map[string]any{
		"chain": c.name,
		"total_duration": totalDuration.String(),
		"steps_completed": len(c.handlers),
	}).Info("Generic chain execution completed")

	return current, nil
}

// GenericPipeline 泛型管道处理器
type GenericPipeline[T any, R any] struct {
	stages []func(context.Context, any) (any, error)
	name   string
}

// NewGenericPipeline 创建泛型管道
func NewGenericPipeline[T any, R any](name string) *GenericPipeline[T, R] {
	return &GenericPipeline[T, R]{
		stages: make([]func(context.Context, any) (any, error), 0),
		name:   name,
	}
}

// AddStage 添加管道阶段
func (p *GenericPipeline[T, R]) AddStage(stage func(context.Context, any) (any, error)) *GenericPipeline[T, R] {
	p.stages = append(p.stages, stage)
	return p
}

// Execute 执行管道
func (p *GenericPipeline[T, R]) Execute(ctx context.Context, input T) (R, error) {
	start := time.Now()
	var current any = input
	
	config.WithFields(map[string]any{
		"pipeline": p.name,
		"stages_count": len(p.stages),
		"input_type": reflect.TypeOf(input).String(),
	}).Debug("Generic pipeline execution started")

	for i, stage := range p.stages {
		stepStart := time.Now()
		result, err := stage(ctx, current)
		stepDuration := time.Since(stepStart)
		
		if err != nil {
			config.WithFields(map[string]any{
				"pipeline": p.name,
				"stage": i + 1,
				"error": err.Error(),
				"step_duration": stepDuration.String(),
				"total_duration": time.Since(start).String(),
			}).Error("Generic pipeline stage failed")
			var zero R
			return zero, fmt.Errorf("pipeline stage %d failed: %w", i+1, err)
		}
		
		config.WithFields(map[string]any{
			"pipeline": p.name,
			"stage": i + 1,
			"step_duration": stepDuration.String(),
		}).Debug("Generic pipeline stage completed")
		
		current = result
	}

	totalDuration := time.Since(start)
	
	// 类型断言到结果类型
	if result, ok := current.(R); ok {
		config.WithFields(map[string]any{
			"pipeline": p.name,
			"total_duration": totalDuration.String(),
			"stages_completed": len(p.stages),
			"output_type": reflect.TypeOf(result).String(),
		}).Info("Generic pipeline execution completed successfully")
		return result, nil
	}
	
	// 如果类型断言失败，尝试JSON序列化/反序列化转换
	resultBytes, err := json.Marshal(current)
	if err != nil {
		config.WithFields(map[string]any{
			"pipeline": p.name,
			"error": "type conversion failed: " + err.Error(),
		}).Error("Generic pipeline type conversion failed")
		var zero R
		return zero, fmt.Errorf("type conversion failed: %w", err)
	}
	
	var result R
	if err := json.Unmarshal(resultBytes, &result); err != nil {
		config.WithFields(map[string]any{
			"pipeline": p.name,
			"error": "JSON conversion failed: " + err.Error(),
		}).Error("Generic pipeline JSON conversion failed")
		var zero R
		return zero, fmt.Errorf("JSON conversion failed: %w", err)
	}
	
	config.WithFields(map[string]any{
		"pipeline": p.name,
		"total_duration": totalDuration.String(),
		"stages_completed": len(p.stages),
		"output_type": reflect.TypeOf(result).String(),
		"conversion_method": "json",
	}).Info("Generic pipeline execution completed with type conversion")
	
	return result, nil
}