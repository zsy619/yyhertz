// Package plugin 提供MyBatis插件系统
//
// 插件系统允许在SQL执行的各个阶段进行拦截和处理，
// 类似于Java MyBatis的Interceptor机制
package plugin

import (
	"context"
	"reflect"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis/config"
)

// Plugin 插件接口
type Plugin interface {
	// Intercept 拦截方法调用
	Intercept(invocation *Invocation) (any, error)

	// Plugin 包装目标对象
	Plugin(target any) any

	// SetProperties 设置插件属性
	SetProperties(properties map[string]any)

	// GetName 获取插件名称
	GetName() string

	// GetOrder 获取插件执行顺序（数字越小越先执行）
	GetOrder() int
}

// Invocation 方法调用信息
type Invocation struct {
	Target     any             // 目标对象
	Method     reflect.Method  // 调用的方法
	Args       []any           // 方法参数
	Context    context.Context // 上下文
	StartTime  time.Time       // 开始时间
	Properties map[string]any  // 附加属性
}

// InterceptorChain 拦截器链
type InterceptorChain struct {
	interceptors []Plugin
}

// PluginRegistry 插件注册表
type PluginRegistry struct {
	plugins map[string]Plugin
	chain   *InterceptorChain
}

// Signature 方法签名注解（用于指定要拦截的方法）
type Signature struct {
	Type   reflect.Type   // 目标类型
	Method string         // 方法名
	Args   []reflect.Type // 参数类型
}

// Intercepts 拦截器注解
type Intercepts struct {
	Signatures []Signature
}

// NewInvocation 创建方法调用信息
func NewInvocation(target any, method reflect.Method, args []any) *Invocation {
	return &Invocation{
		Target:     target,
		Method:     method,
		Args:       args,
		Context:    context.Background(),
		StartTime:  time.Now(),
		Properties: make(map[string]any),
	}
}

// Proceed 继续执行方法调用
func (inv *Invocation) Proceed() (any, error) {
	// 使用反射调用原始方法
	values := make([]reflect.Value, len(inv.Args))
	for i, arg := range inv.Args {
		values[i] = reflect.ValueOf(arg)
	}

	targetValue := reflect.ValueOf(inv.Target)
	results := targetValue.MethodByName(inv.Method.Name).Call(values)

	// 处理返回值
	if len(results) == 0 {
		return nil, nil
	}

	if len(results) == 1 {
		if results[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if results[0].IsNil() {
				return nil, nil
			}
			return nil, results[0].Interface().(error)
		}
		return results[0].Interface(), nil
	}

	// 多个返回值，最后一个通常是error
	lastResult := results[len(results)-1]
	if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !lastResult.IsNil() {
			return nil, lastResult.Interface().(error)
		}

		if len(results) == 2 {
			return results[0].Interface(), nil
		}

		// 多个返回值，返回除error外的所有值
		returnValues := make([]any, len(results)-1)
		for i := 0; i < len(results)-1; i++ {
			returnValues[i] = results[i].Interface()
		}
		return returnValues, nil
	}

	// 所有返回值都不是error
	returnValues := make([]any, len(results))
	for i, result := range results {
		returnValues[i] = result.Interface()
	}
	return returnValues, nil
}

// NewInterceptorChain 创建拦截器链
func NewInterceptorChain() *InterceptorChain {
	return &InterceptorChain{
		interceptors: make([]Plugin, 0),
	}
}

// AddInterceptor 添加拦截器
func (chain *InterceptorChain) AddInterceptor(plugin Plugin) {
	chain.interceptors = append(chain.interceptors, plugin)
}

// PluginAll 为目标对象应用所有插件
func (chain *InterceptorChain) PluginAll(target any) any {
	for _, plugin := range chain.interceptors {
		target = plugin.Plugin(target)
	}
	return target
}

// GetInterceptors 获取所有拦截器
func (chain *InterceptorChain) GetInterceptors() []Plugin {
	return chain.interceptors
}

// NewPluginRegistry 创建插件注册表
func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
		chain:   NewInterceptorChain(),
	}
}

// RegisterPlugin 注册插件
func (registry *PluginRegistry) RegisterPlugin(plugin Plugin) {
	registry.plugins[plugin.GetName()] = plugin
	registry.chain.AddInterceptor(plugin)
}

// GetPlugin 获取插件
func (registry *PluginRegistry) GetPlugin(name string) Plugin {
	return registry.plugins[name]
}

// GetAllPlugins 获取所有插件
func (registry *PluginRegistry) GetAllPlugins() map[string]Plugin {
	return registry.plugins
}

// GetInterceptorChain 获取拦截器链
func (registry *PluginRegistry) GetInterceptorChain() *InterceptorChain {
	return registry.chain
}

// PluginProxy 插件代理
type PluginProxy struct {
	target      any
	interceptor Plugin
}

// NewPluginProxy 创建插件代理
func NewPluginProxy(target any, interceptor Plugin) *PluginProxy {
	return &PluginProxy{
		target:      target,
		interceptor: interceptor,
	}
}

// Invoke 调用方法
func (proxy *PluginProxy) Invoke(method reflect.Method, args []any) (any, error) {
	invocation := NewInvocation(proxy.target, method, args)
	return proxy.interceptor.Intercept(invocation)
}

// BasePlugin 基础插件实现
type BasePlugin struct {
	name       string
	order      int
	properties map[string]any
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(name string, order int) *BasePlugin {
	return &BasePlugin{
		name:       name,
		order:      order,
		properties: make(map[string]any),
	}
}

// GetName 获取插件名称
func (plugin *BasePlugin) GetName() string {
	return plugin.name
}

// GetOrder 获取插件执行顺序
func (plugin *BasePlugin) GetOrder() int {
	return plugin.order
}

// SetProperties 设置插件属性
func (plugin *BasePlugin) SetProperties(properties map[string]any) {
	plugin.properties = properties
}

// GetProperty 获取插件属性
func (plugin *BasePlugin) GetProperty(key string) any {
	return plugin.properties[key]
}

// GetPropertyString 获取字符串属性
func (plugin *BasePlugin) GetPropertyString(key string, defaultValue string) string {
	if value, exists := plugin.properties[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// GetPropertyInt 获取整数属性
func (plugin *BasePlugin) GetPropertyInt(key string, defaultValue int) int {
	if value, exists := plugin.properties[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return defaultValue
}

// GetPropertyBool 获取布尔属性
func (plugin *BasePlugin) GetPropertyBool(key string, defaultValue bool) bool {
	if value, exists := plugin.properties[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// PluginConfiguration 插件配置
type PluginConfiguration struct {
	Enabled bool           `json:"enabled" yaml:"enabled"`
	Plugins []PluginConfig `json:"plugins" yaml:"plugins"`
}

// PluginConfig 单个插件配置
type PluginConfig struct {
	Name       string         `json:"name" yaml:"name"`
	Enabled    bool           `json:"enabled" yaml:"enabled"`
	Order      int            `json:"order" yaml:"order"`
	Properties map[string]any `json:"properties" yaml:"properties"`
}

// LoadPluginConfiguration 加载插件配置
func LoadPluginConfiguration(config *config.Configuration) *PluginConfiguration {
	// 从配置中加载插件设置
	// 这里简化实现，实际应该从配置文件读取
	return &PluginConfiguration{
		Enabled: true,
		Plugins: []PluginConfig{
			{
				Name:    "pagination",
				Enabled: true,
				Order:   1,
				Properties: map[string]any{
					"defaultPageSize": 20,
					"maxPageSize":     1000,
				},
			},
			{
				Name:    "performance",
				Enabled: true,
				Order:   2,
				Properties: map[string]any{
					"slowQueryThreshold": 1000, // 毫秒
					"enableMetrics":      true,
				},
			},
			{
				Name:    "sqllog",
				Enabled: true,
				Order:   3,
				Properties: map[string]any{
					"logLevel":     "INFO",
					"logSql":       true,
					"logResult":    false,
					"logParameter": true,
				},
			},
		},
	}
}
