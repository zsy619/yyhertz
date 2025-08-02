package annotation

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// MethodRegistry 方法注解注册器
type MethodRegistry struct {
	mutex           sync.RWMutex
	methodMappings  map[string]*MethodMapping // key: TypeName.MethodName
	controllerInfos map[reflect.Type]*ControllerInfo
}

// MethodMapping 方法映射信息
type MethodMapping struct {
	ControllerType reflect.Type
	MethodName     string
	Path           string
	HTTPMethod     string
	Params         []*ParamMapping
	Middlewares    []string
	Description    string
	Tags           map[string]string
}

// ParamMapping 参数映射信息
type ParamMapping struct {
	Name         string
	Source       ParamSource // path, query, body, header, cookie
	Required     bool
	DefaultValue string
	Validation   *ValidationInfo
}

// ParamSource 参数来源
type ParamSource string

const (
	ParamSourcePath   ParamSource = "path"
	ParamSourceQuery  ParamSource = "query"
	ParamSourceBody   ParamSource = "body"
	ParamSourceHeader ParamSource = "header"
	ParamSourceCookie ParamSource = "cookie"
	ParamSourceForm   ParamSource = "form"
)

var (
	globalRegistry = &MethodRegistry{
		methodMappings:  make(map[string]*MethodMapping),
		controllerInfos: make(map[reflect.Type]*ControllerInfo),
	}
)

// GetRegistry 获取全局注册器
func GetRegistry() *MethodRegistry {
	return globalRegistry
}

// RegisterGetMapping 注册GET方法映射
func (r *MethodRegistry) RegisterGetMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return r.registerMapping(controllerType, methodName, path, "GET")
}

// RegisterPostMapping 注册POST方法映射
func (r *MethodRegistry) RegisterPostMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return r.registerMapping(controllerType, methodName, path, "POST")
}

// RegisterPutMapping 注册PUT方法映射
func (r *MethodRegistry) RegisterPutMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return r.registerMapping(controllerType, methodName, path, "PUT")
}

// RegisterDeleteMapping 注册DELETE方法映射
func (r *MethodRegistry) RegisterDeleteMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return r.registerMapping(controllerType, methodName, path, "DELETE")
}

// RegisterHeadMapping 注册HEAD方法映射
func (r *MethodRegistry) RegisterHeadMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return r.registerMapping(controllerType, methodName, path, "HEAD")
}

// RegisterPatchMapping 注册PATCH方法映射
func (r *MethodRegistry) RegisterPatchMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return r.registerMapping(controllerType, methodName, path, "PATCH")
}

// RegisterAnyMapping 注册任意方法映射
func (r *MethodRegistry) RegisterAnyMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return r.registerMapping(controllerType, methodName, path, "ANY")
}

// registerMapping 注册方法映射
func (r *MethodRegistry) registerMapping(controllerType reflect.Type, methodName, path, httpMethod string) *MethodMappingBuilder {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	key := fmt.Sprintf("%s.%s", controllerType.Name(), methodName)
	mapping := &MethodMapping{
		ControllerType: controllerType,
		MethodName:     methodName,
		Path:           normalizePath(path),
		HTTPMethod:     httpMethod,
		Params:         make([]*ParamMapping, 0),
		Middlewares:    make([]string, 0),
		Tags:           make(map[string]string),
	}

	r.methodMappings[key] = mapping
	return &MethodMappingBuilder{mapping: mapping, registry: r}
}

// GetMethodMapping 获取方法映射
func (r *MethodRegistry) GetMethodMapping(controllerType reflect.Type, methodName string) *MethodMapping {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	key := fmt.Sprintf("%s.%s", controllerType.Name(), methodName)
	return r.methodMappings[key]
}

// GetControllerMappings 获取控制器的所有方法映射
func (r *MethodRegistry) GetControllerMappings(controllerType reflect.Type) []*MethodMapping {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var mappings []*MethodMapping
	prefix := controllerType.Name() + "."

	for key, mapping := range r.methodMappings {
		if strings.HasPrefix(key, prefix) {
			mappings = append(mappings, mapping)
		}
	}

	return mappings
}

// RegisterControllerInfo 注册控制器信息
func (r *MethodRegistry) RegisterControllerInfo(controllerType reflect.Type, info *ControllerInfo) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.controllerInfos[controllerType] = info
}

// GetControllerInfo 获取控制器信息
func (r *MethodRegistry) GetControllerInfo(controllerType reflect.Type) *ControllerInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.controllerInfos[controllerType]
}

// GetAllControllers 获取所有注册的控制器
func (r *MethodRegistry) GetAllControllers() []*ControllerInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var controllers []*ControllerInfo
	for _, info := range r.controllerInfos {
		controllers = append(controllers, info)
	}

	return controllers
}

// MethodMappingBuilder 方法映射构建器
type MethodMappingBuilder struct {
	mapping  *MethodMapping
	registry *MethodRegistry
}

// WithPathParam 添加路径参数
func (b *MethodMappingBuilder) WithPathParam(name string, required bool) *MethodMappingBuilder {
	param := &ParamMapping{
		Name:     name,
		Source:   ParamSourcePath,
		Required: required,
	}
	b.mapping.Params = append(b.mapping.Params, param)
	return b
}

// WithQueryParam 添加查询参数
func (b *MethodMappingBuilder) WithQueryParam(name string, required bool, defaultValue string) *MethodMappingBuilder {
	param := &ParamMapping{
		Name:         name,
		Source:       ParamSourceQuery,
		Required:     required,
		DefaultValue: defaultValue,
	}
	b.mapping.Params = append(b.mapping.Params, param)
	return b
}

// WithBodyParam 添加请求体参数
func (b *MethodMappingBuilder) WithBodyParam(required bool) *MethodMappingBuilder {
	param := &ParamMapping{
		Name:     "body",
		Source:   ParamSourceBody,
		Required: required,
	}
	b.mapping.Params = append(b.mapping.Params, param)
	return b
}

// WithHeaderParam 添加请求头参数
func (b *MethodMappingBuilder) WithHeaderParam(name string, required bool, defaultValue string) *MethodMappingBuilder {
	param := &ParamMapping{
		Name:         name,
		Source:       ParamSourceHeader,
		Required:     required,
		DefaultValue: defaultValue,
	}
	b.mapping.Params = append(b.mapping.Params, param)
	return b
}

// WithCookieParam 添加Cookie参数
func (b *MethodMappingBuilder) WithCookieParam(name string, required bool, defaultValue string) *MethodMappingBuilder {
	param := &ParamMapping{
		Name:         name,
		Source:       ParamSourceCookie,
		Required:     required,
		DefaultValue: defaultValue,
	}
	b.mapping.Params = append(b.mapping.Params, param)
	return b
}

// WithMiddleware 添加中间件
func (b *MethodMappingBuilder) WithMiddleware(middlewares ...string) *MethodMappingBuilder {
	b.mapping.Middlewares = append(b.mapping.Middlewares, middlewares...)
	return b
}

// WithDescription 设置描述
func (b *MethodMappingBuilder) WithDescription(desc string) *MethodMappingBuilder {
	b.mapping.Description = desc
	return b
}

// WithTag 添加标签
func (b *MethodMappingBuilder) WithTag(key, value string) *MethodMappingBuilder {
	b.mapping.Tags[key] = value
	return b
}

// Build 构建完成
func (b *MethodMappingBuilder) Build() *MethodMapping {
	return b.mapping
}

// 便捷的全局注册函数

// GetMapping 注册GET映射
func GetMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterGetMapping(controllerType, methodName, path)
}

// PostMapping 注册POST映射
func PostMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterPostMapping(controllerType, methodName, path)
}

// PutMapping 注册PUT映射
func PutMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterPutMapping(controllerType, methodName, path)
}

// DeleteMapping 注册DELETE映射
func DeleteMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterDeleteMapping(controllerType, methodName, path)
}

// PatchMapping 注册PATCH映射
func PatchMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterPatchMapping(controllerType, methodName, path)
}

// AnyMapping 注册任意方法映射
func AnyMapping(controllerType reflect.Type, methodName, path string) *MethodMappingBuilder {
	return GetRegistry().RegisterAnyMapping(controllerType, methodName, path)
}
