package routing

import (
	"reflect"
	"strings"
	
	"github.com/zsy619/yyhertz/framework/context"
)

// CombinePath 组合路径（从annotation包提取）
func CombinePath(basePath, methodPath string) string {
	basePath = NormalizePath(basePath)
	methodPath = NormalizePath(methodPath)
	
	if basePath == "" {
		return methodPath
	}
	
	if methodPath == "" || methodPath == "/" {
		return basePath
	}
	
	return basePath + methodPath
}

// NormalizePath 规范化路径（从annotation包提取）
func NormalizePath(path string) string {
	if path == "" {
		return ""
	}
	
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	if strings.HasSuffix(path, "/") && path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	
	return path
}

// GetPackageName 从包路径获取包名（从comment包提取）
func GetPackageName(pkgPath string) string {
	if pkgPath == "" {
		return ""
	}
	
	parts := strings.Split(pkgPath, "/")
	return parts[len(parts)-1]
}

// IsContextType 检查是否为Context类型（从comment包提取）
func IsContextType(t reflect.Type) bool {
	return t.String() == "*context.Context" || 
		   strings.Contains(t.String(), "Context")
}

// IsStructType 检查是否为结构体类型（从comment包提取）
func IsStructType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

// IsStringType 检查是否为字符串类型（从comment包提取）
func IsStringType(t reflect.Type) bool {
	return t.Kind() == reflect.String
}

// ConvertValue 转换值类型（从comment包提取并增强）
func ConvertValue(value string, targetType reflect.Type) (reflect.Value, error) {
	switch targetType.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 简化实现，实际应该进行类型转换
		return reflect.ValueOf(0), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return reflect.ValueOf(uint(0)), nil
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(0.0), nil
	case reflect.Bool:
		return reflect.ValueOf(false), nil
	default:
		return reflect.Zero(targetType), nil
	}
}

// GetControllerName 获取控制器名称
func GetControllerName(controllerType reflect.Type) string {
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}
	return controllerType.Name()
}

// GetMethodType 获取方法类型
func GetMethodType(controllerType reflect.Type, methodName string) (reflect.Type, bool) {
	// 确保是指针类型
	if controllerType.Kind() != reflect.Ptr {
		controllerType = reflect.PtrTo(controllerType)
	}
	
	method, exists := controllerType.MethodByName(methodName)
	if !exists {
		return nil, false
	}
	
	return method.Type, true
}

// ValidateControllerType 验证控制器类型是否有效
func ValidateControllerType(controllerType reflect.Type) error {
	if controllerType == nil {
		return &RouteError{
			Type:    ErrorTypeInvalidController,
			Message: "controller type is nil",
		}
	}
	
	// 确保是结构体类型
	elemType := controllerType
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	
	if elemType.Kind() != reflect.Struct {
		return &RouteError{
			Type:    ErrorTypeInvalidController,
			Message: "controller must be a struct type",
		}
	}
	
	return nil
}

// ValidateMethodName 验证方法名是否有效
func ValidateMethodName(controllerType reflect.Type, methodName string) error {
	if methodName == "" {
		return &RouteError{
			Type:    ErrorTypeInvalidMethod,
			Message: "method name is empty",
		}
	}
	
	// 确保方法存在
	_, exists := GetMethodType(controllerType, methodName)
	if !exists {
		return &RouteError{
			Type:    ErrorTypeInvalidMethod,
			Message: "method '" + methodName + "' not found in controller",
		}
	}
	
	return nil
}

// ValidateHTTPMethod 验证HTTP方法是否有效
func ValidateHTTPMethod(method string) error {
	method = strings.ToUpper(method)
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "ANY"}
	
	for _, valid := range validMethods {
		if method == valid {
			return nil
		}
	}
	
	return &RouteError{
		Type:    ErrorTypeInvalidHTTPMethod,
		Message: "invalid HTTP method: " + method,
	}
}

// ValidatePath 验证路径是否有效
func ValidatePath(path string) error {
	if path == "" {
		return &RouteError{
			Type:    ErrorTypeInvalidPath,
			Message: "path is empty",
		}
	}
	
	// 简单的路径验证
	if !strings.HasPrefix(path, "/") && path != "*" {
		return &RouteError{
			Type:    ErrorTypeInvalidPath,
			Message: "path must start with '/' or be '*'",
		}
	}
	
	return nil
}

// CreateEnhancedContext 创建增强的上下文（从comment包提取）
func CreateEnhancedContext(c interface{}) *context.Context {
	// 这里需要根据实际的RequestContext类型进行适配
	if ctx, ok := c.(*context.Context); ok {
		return ctx
	}
	
	// 如果是其他类型，需要进行转换
	// 这里简化处理，实际使用时需要具体实现
	return &context.Context{}
}

// RouteError 路由错误类型
type RouteError struct {
	Type    ErrorType
	Message string
	Cause   error
}

func (e *RouteError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// ErrorType 错误类型枚举
type ErrorType string

const (
	ErrorTypeInvalidController  ErrorType = "invalid_controller"
	ErrorTypeInvalidMethod      ErrorType = "invalid_method"
	ErrorTypeInvalidHTTPMethod  ErrorType = "invalid_http_method"
	ErrorTypeInvalidPath        ErrorType = "invalid_path"
	ErrorTypeInvalidParam       ErrorType = "invalid_param"
	ErrorTypeRegistrationError  ErrorType = "registration_error"
	ErrorTypeParsingError       ErrorType = "parsing_error"
)

// RouteConflictError 路由冲突错误
type RouteConflictError struct {
	ExistingRoute *RouteInfo
	NewRoute      *RouteInfo
}

func (e *RouteConflictError) Error() string {
	return "route conflict: " + e.NewRoute.HTTPMethod + " " + e.NewRoute.Path + 
		   " already registered by " + e.ExistingRoute.TypeName + "." + e.ExistingRoute.MethodName
}

// Helper functions for creating common parameter info

// NewPathParam 创建路径参数
func NewPathParam(name string, required bool) *ParamInfo {
	return &ParamInfo{
		Name:     name,
		Source:   ParamSourcePath,
		Required: required,
		Type:     "string",
	}
}

// NewQueryParam 创建查询参数
func NewQueryParam(name, defaultValue string, required bool) *ParamInfo {
	return &ParamInfo{
		Name:         name,
		Source:       ParamSourceQuery,
		Required:     required,
		DefaultValue: defaultValue,
		Type:         "string",
	}
}

// NewBodyParam 创建请求体参数
func NewBodyParam(required bool) *ParamInfo {
	return &ParamInfo{
		Name:     "body",
		Source:   ParamSourceBody,
		Required: required,
		Type:     "object",
	}
}

// NewHeaderParam 创建请求头参数
func NewHeaderParam(name, defaultValue string, required bool) *ParamInfo {
	return &ParamInfo{
		Name:         name,
		Source:       ParamSourceHeader,
		Required:     required,
		DefaultValue: defaultValue,
		Type:         "string",
	}
}

// NewCookieParam 创建Cookie参数
func NewCookieParam(name, defaultValue string, required bool) *ParamInfo {
	return &ParamInfo{
		Name:         name,
		Source:       ParamSourceCookie,
		Required:     required,
		DefaultValue: defaultValue,
		Type:         "string",
	}
}