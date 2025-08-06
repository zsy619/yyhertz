// Package plugin 插件管理器实现
//
// 提供插件的注册、配置、启用/禁用等管理功能
package plugin

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/zsy619/yyhertz/framework/mybatis/config"
)

// PluginManager 插件管理器
type PluginManager struct {
	registry       *PluginRegistry
	plugins        map[string]Plugin
	enabledPlugins []Plugin
	configuration  *config.Configuration
	mutex          sync.RWMutex
}

// NewPluginManager 创建插件管理器
func NewPluginManager(configuration *config.Configuration) *PluginManager {
	manager := &PluginManager{
		registry:       NewPluginRegistry(),
		plugins:        make(map[string]Plugin),
		enabledPlugins: make([]Plugin, 0),
		configuration:  configuration,
	}

	// 注册默认插件
	manager.registerDefaultPlugins()

	return manager
}

// registerDefaultPlugins 注册默认插件
func (manager *PluginManager) registerDefaultPlugins() {
	// 分页插件
	paginationPlugin := NewPaginationPlugin()
	manager.RegisterPlugin(paginationPlugin)

	// 性能监控插件
	performancePlugin := NewPerformancePlugin()
	manager.RegisterPlugin(performancePlugin)

	// SQL日志插件
	sqlLogPlugin := NewSqlLogPlugin()
	manager.RegisterPlugin(sqlLogPlugin)

	// 缓存增强插件
	cachePlugin := NewCacheEnhancerPlugin()
	manager.RegisterPlugin(cachePlugin)

	// 参数验证插件
	validatorPlugin := NewValidatorPlugin()
	manager.RegisterPlugin(validatorPlugin)

	// 结果转换插件
	transformerPlugin := NewResultTransformerPlugin()
	manager.RegisterPlugin(transformerPlugin)
}

// RegisterPlugin 注册插件
func (manager *PluginManager) RegisterPlugin(plugin Plugin) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	name := plugin.GetName()
	if _, exists := manager.plugins[name]; exists {
		return fmt.Errorf("插件 %s 已经注册", name)
	}

	manager.plugins[name] = plugin
	manager.registry.RegisterPlugin(plugin)

	// 如果插件默认启用，添加到启用列表
	manager.refreshEnabledPlugins()

	return nil
}

// UnregisterPlugin 注销插件
func (manager *PluginManager) UnregisterPlugin(name string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	if _, exists := manager.plugins[name]; !exists {
		return fmt.Errorf("插件 %s 未注册", name)
	}

	delete(manager.plugins, name)
	manager.refreshEnabledPlugins()

	return nil
}

// EnablePlugin 启用插件
func (manager *PluginManager) EnablePlugin(name string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	plugin, exists := manager.plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 未注册", name)
	}

	// 检查是否已启用
	for _, enabledPlugin := range manager.enabledPlugins {
		if enabledPlugin.GetName() == name {
			return nil // 已经启用
		}
	}

	manager.enabledPlugins = append(manager.enabledPlugins, plugin)
	manager.sortPluginsByOrder()

	return nil
}

// DisablePlugin 禁用插件
func (manager *PluginManager) DisablePlugin(name string) error {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	for i, plugin := range manager.enabledPlugins {
		if plugin.GetName() == name {
			// 从启用列表中移除
			manager.enabledPlugins = append(manager.enabledPlugins[:i], manager.enabledPlugins[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("插件 %s 未启用", name)
}

// GetPlugin 获取插件
func (manager *PluginManager) GetPlugin(name string) (Plugin, error) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	plugin, exists := manager.plugins[name]
	if !exists {
		return nil, fmt.Errorf("插件 %s 未注册", name)
	}

	return plugin, nil
}

// GetAllPlugins 获取所有插件
func (manager *PluginManager) GetAllPlugins() map[string]Plugin {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	result := make(map[string]Plugin)
	for name, plugin := range manager.plugins {
		result[name] = plugin
	}

	return result
}

// GetEnabledPlugins 获取启用的插件
func (manager *PluginManager) GetEnabledPlugins() []Plugin {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	result := make([]Plugin, len(manager.enabledPlugins))
	copy(result, manager.enabledPlugins)

	return result
}

// IsPluginEnabled 检查插件是否启用
func (manager *PluginManager) IsPluginEnabled(name string) bool {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	for _, plugin := range manager.enabledPlugins {
		if plugin.GetName() == name {
			return true
		}
	}

	return false
}

// ConfigurePlugin 配置插件
func (manager *PluginManager) ConfigurePlugin(name string, properties map[string]any) error {
	plugin, err := manager.GetPlugin(name)
	if err != nil {
		return err
	}

	plugin.SetProperties(properties)
	return nil
}

// ApplyPlugins 应用插件到目标对象
func (manager *PluginManager) ApplyPlugins(target any) any {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	result := target
	for _, plugin := range manager.enabledPlugins {
		result = plugin.Plugin(result)
	}

	return result
}

// refreshEnabledPlugins 刷新启用的插件列表
func (manager *PluginManager) refreshEnabledPlugins() {
	manager.enabledPlugins = manager.enabledPlugins[:0]

	// 默认启用所有插件
	for _, plugin := range manager.plugins {
		manager.enabledPlugins = append(manager.enabledPlugins, plugin)
	}

	manager.sortPluginsByOrder()
}

// sortPluginsByOrder 按执行顺序排序插件
func (manager *PluginManager) sortPluginsByOrder() {
	sort.Slice(manager.enabledPlugins, func(i, j int) bool {
		return manager.enabledPlugins[i].GetOrder() < manager.enabledPlugins[j].GetOrder()
	})
}

// GetPluginStatus 获取插件状态
func (manager *PluginManager) GetPluginStatus() map[string]PluginStatus {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	status := make(map[string]PluginStatus)

	for name, plugin := range manager.plugins {
		enabled := manager.IsPluginEnabled(name)
		status[name] = PluginStatus{
			Name:    name,
			Enabled: enabled,
			Order:   plugin.GetOrder(),
			Plugin:  plugin,
		}
	}

	return status
}

// PluginStatus 插件状态
type PluginStatus struct {
	Name    string // 插件名称
	Enabled bool   // 是否启用
	Order   int    // 执行顺序
	Plugin  Plugin // 插件实例
}

// GetPluginReport 获取插件报告
func (manager *PluginManager) GetPluginReport() map[string]any {
	status := manager.GetPluginStatus()

	report := map[string]any{
		"总插件数":  len(manager.plugins),
		"启用插件数": len(manager.enabledPlugins),
		"插件状态":  make(map[string]any),
	}

	pluginDetails := make(map[string]any)
	for name, s := range status {
		pluginDetails[name] = map[string]any{
			"启用状态": s.Enabled,
			"执行顺序": s.Order,
		}

		// 添加特定插件的详细信息
		switch plugin := s.Plugin.(type) {
		case *PerformancePlugin:
			if s.Enabled {
				pluginDetails[name].(map[string]any)["性能报告"] = plugin.GetPerformanceReport()
			}
		case *CacheEnhancerPlugin:
			if s.Enabled {
				pluginDetails[name].(map[string]any)["缓存报告"] = plugin.GetCacheReport()
			}
		}
	}

	report["插件状态"] = pluginDetails
	return report
}

// LoadConfiguration 加载插件配置
func (manager *PluginManager) LoadConfiguration(pluginConfig *PluginConfiguration) error {
	if !pluginConfig.Enabled {
		// 禁用所有插件
		manager.mutex.Lock()
		manager.enabledPlugins = manager.enabledPlugins[:0]
		manager.mutex.Unlock()
		return nil
	}

	for _, config := range pluginConfig.Plugins {
		if !config.Enabled {
			manager.DisablePlugin(config.Name)
			continue
		}

		// 配置插件属性
		if err := manager.ConfigurePlugin(config.Name, config.Properties); err != nil {
			return fmt.Errorf("配置插件 %s 失败: %v", config.Name, err)
		}

		// 启用插件
		if err := manager.EnablePlugin(config.Name); err != nil {
			return fmt.Errorf("启用插件 %s 失败: %v", config.Name, err)
		}
	}

	return nil
}

// ExecuteWithPlugins 使用插件执行方法
func (manager *PluginManager) ExecuteWithPlugins(target any, methodName string, args []any) (any, error) {
	// 创建调用信息
	method, exists := manager.findMethod(target, methodName)
	if !exists {
		return nil, fmt.Errorf("方法 %s 不存在", methodName)
	}

	invocation := NewInvocation(target, method, args)

	// 获取启用的插件
	enabledPlugins := manager.GetEnabledPlugins()

	// 如果没有插件，直接执行原方法
	if len(enabledPlugins) == 0 {
		return invocation.Proceed()
	}

	// 递归执行插件链
	return manager.executePluginChain(invocation, enabledPlugins, 0)
}

// executePluginChain 递归执行插件链
func (manager *PluginManager) executePluginChain(invocation *Invocation, plugins []Plugin, index int) (any, error) {
	if index >= len(plugins) {
		// 所有插件都已执行，执行原方法
		return invocation.Proceed()
	}

	// 获取当前插件
	plugin := plugins[index]

	// 创建新的调用实例，重写Proceed方法
	chainInvocation := &Invocation{
		Target:     invocation.Target,
		Method:     invocation.Method,
		Args:       invocation.Args,
		Context:    invocation.Context,
		StartTime:  invocation.StartTime,
		Properties: invocation.Properties,
	}

	// 保存管理器和插件信息到Properties中
	if chainInvocation.Properties == nil {
		chainInvocation.Properties = make(map[string]any)
	}
	chainInvocation.Properties["_manager"] = manager
	chainInvocation.Properties["_plugins"] = plugins
	chainInvocation.Properties["_index"] = index
	chainInvocation.Properties["_originalInvocation"] = invocation

	return plugin.Intercept(chainInvocation)
}

// findMethod 查找方法
func (manager *PluginManager) findMethod(target any, methodName string) (method reflect.Method, exists bool) {
	targetType := reflect.TypeOf(target)
	return targetType.MethodByName(methodName)
}

// GetInterceptorChain 获取拦截器链
func (manager *PluginManager) GetInterceptorChain() *InterceptorChain {
	return manager.registry.GetInterceptorChain()
}

// CreatePluginProxy 创建插件代理
func (manager *PluginManager) CreatePluginProxy(target any) any {
	return manager.ApplyPlugins(target)
}
