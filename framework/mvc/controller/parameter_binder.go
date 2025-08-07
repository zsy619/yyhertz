package controller

import (
	"fmt"
	"reflect"

	"github.com/zsy619/yyhertz/framework/mvc/binding"
	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// NewParameterBinder 创建参数绑定器（适配器）
func NewParameterBinder(methodType reflect.Type) (*ParameterBinder, error) {
	bindingBinder, err := binding.NewParameterBinder(methodType)
	if err != nil {
		return nil, err
	}

	return &ParameterBinder{
		binder: bindingBinder,
	}, nil
}

// ParameterBinder 参数绑定器适配器
type ParameterBinder struct {
	binder *binding.ParameterBinder
}

// BindParameters 绑定参数
func (pb *ParameterBinder) BindParameters(ctx *context.Context) ([]interface{}, error) {
	return pb.binder.BindParameters(ctx)
}

// NewMethodValidator 创建方法验证器
func NewMethodValidator(methodType reflect.Type) *MethodValidator {
	return &MethodValidator{
		validator: binding.NewParameterValidator(),
		methodType: methodType,
	}
}

// MethodValidator 方法验证器
type MethodValidator struct {
	validator  *binding.ParameterValidator
	methodType reflect.Type
}

// ValidateParameters 验证参数
func (mv *MethodValidator) ValidateParameters(params []interface{}) error {
	for _, param := range params {
		if err := mv.validator.ValidateStruct(param); err != nil {
			return fmt.Errorf("parameter validation failed: %w", err)
		}
	}
	return nil
}