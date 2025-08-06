package mvc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zsy619/yyhertz/framework/mvc/handler"
)

// TestGenericHandler 测试泛型处理器
func TestGenericHandler(t *testing.T) {
	// 创建简单处理器
	h := handler.NewBaseGenericHandler[string, string]("test_handler")
	assert.NotNil(t, h, "Handler should not be nil")
	assert.Equal(t, "test_handler", h.GetName(), "Handler name should match")
}

// TestGenericHandlerWithProcessor 测试带处理逻辑的泛型处理器
func TestGenericHandlerWithProcessor(t *testing.T) {
	h := handler.NewBaseGenericHandler[string, string]("echo_handler").
		WithProcessor(func(ctx context.Context, input string) (string, error) {
			return "echo: " + input, nil
		})

	result, err := h.Handle(context.Background(), "test")
	assert.NoError(t, err, "Handler should not return error")
	assert.Equal(t, "echo: test", result, "Handler should return processed result")
}

// TestGenericHandlerWithValidator 测试带验证的泛型处理器
func TestGenericHandlerWithValidator(t *testing.T) {
	h := handler.NewBaseGenericHandler[string, string]("validator_handler").
		WithValidator(func(input string) error {
			if input == "" {
				return assert.AnError
			}
			return nil
		}).
		WithProcessor(func(ctx context.Context, input string) (string, error) {
			return input, nil
		})

	// 测试有效输入
	result, err := h.Handle(context.Background(), "valid")
	assert.NoError(t, err, "Valid input should not return error")
	assert.Equal(t, "valid", result, "Handler should return input")

	// 测试无效输入
	_, err = h.Handle(context.Background(), "")
	assert.Error(t, err, "Invalid input should return error")
}

// TestSimpleHandler 测试简单处理器
func TestSimpleHandler(t *testing.T) {
	h := handler.NewSimpleHandler("double_handler", func(ctx context.Context, input int) (int, error) {
		return input * 2, nil
	})

	result, err := h.Handle(context.Background(), 5)
	assert.NoError(t, err, "Handler should not return error")
	assert.Equal(t, 10, result, "Handler should double the input")
}
