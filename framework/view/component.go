package view

import (
	"fmt"
	"html/template"
	"strings"
	"sync"

	globalConfig "github.com/zsy619/yyhertz/framework/config"
)

// ComponentManager 组件管理器
type ComponentManager struct {
	engine     *TemplateEngine
	components map[string]*ComponentConfig
	mutex      sync.RWMutex
}

// ComponentConfig 组件配置
type ComponentConfig struct {
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	Props        map[string]any `json:"props"`
	Slots        []string       `json:"slots"`
	Cache        bool           `json:"cache"`
	CacheKey     string         `json:"cache_key"`
	Theme        string         `json:"theme"`
	Description  string         `json:"description"`
	Enabled      bool           `json:"enabled"`
	Dependencies []string       `json:"dependencies"`
}

// ComponentInstance 组件实例
type ComponentInstance struct {
	Config *ComponentConfig
	Props  map[string]any
	Slots  map[string]template.HTML
	Theme  string
}

// NewComponentManager 创建组件管理器
func NewComponentManager(engine *TemplateEngine) *ComponentManager {
	return &ComponentManager{
		engine:     engine,
		components: make(map[string]*ComponentConfig),
	}
}

// RegisterComponent 注册组件
func (cm *ComponentManager) RegisterComponent(name string, config *ComponentConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	config.Name = name
	if config.Props == nil {
		config.Props = make(map[string]any)
	}
	if config.Slots == nil {
		config.Slots = []string{"default"}
	}

	// 检查依赖组件
	for _, dep := range config.Dependencies {
		if _, exists := cm.components[dep]; !exists {
			return fmt.Errorf("dependency component '%s' not found", dep)
		}
	}

	cm.components[name] = config

	globalConfig.Infof("Registered component: %s", name)
	return nil
}

// GetComponent 获取组件配置
func (cm *ComponentManager) GetComponent(name string) (*ComponentConfig, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	component, exists := cm.components[name]
	return component, exists
}

// GetAllComponents 获取所有组件
func (cm *ComponentManager) GetAllComponents() map[string]*ComponentConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	components := make(map[string]*ComponentConfig)
	for name, config := range cm.components {
		components[name] = config
	}
	return components
}

// IsComponentEnabled 检查组件是否启用
func (cm *ComponentManager) IsComponentEnabled(name string) bool {
	if component, exists := cm.GetComponent(name); exists {
		return component.Enabled
	}
	return false
}

// RenderComponent 渲染组件
func (cm *ComponentManager) RenderComponent(name string, props map[string]any, slots map[string]template.HTML) (template.HTML, error) {
	config, exists := cm.GetComponent(name)
	if !exists {
		return "", fmt.Errorf("component '%s' not found", name)
	}

	if !config.Enabled {
		return "", fmt.Errorf("component '%s' is disabled", name)
	}

	// 创建组件实例
	instance := &ComponentInstance{
		Config: config,
		Props:  cm.mergeProps(config.Props, props),
		Slots:  cm.mergeSlots(config.Slots, slots),
		Theme:  cm.engine.currentTheme,
	}

	// 渲染组件
	return cm.renderComponentInstance(instance)
}

// renderComponentInstance 渲染组件实例
func (cm *ComponentManager) renderComponentInstance(instance *ComponentInstance) (template.HTML, error) {
	// 获取组件模板
	componentTemplate, exists := cm.engine.components[instance.Config.Name]
	if !exists {
		return "", fmt.Errorf("component template '%s' not found", instance.Config.Name)
	}

	// 准备渲染数据
	data := map[string]any{
		"props":     instance.Props,
		"slots":     instance.Slots,
		"theme":     instance.Theme,
		"component": instance.Config,
	}

	// 添加组件特定函数
	data["slot"] = cm.createSlotFunc(instance)
	data["prop"] = cm.createPropFunc(instance)
	data["emit"] = cm.createEmitFunc(instance)

	// 执行模板
	var buf strings.Builder
	if err := componentTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("component template execution error: %w", err)
	}

	return template.HTML(buf.String()), nil
}

// mergeProps 合并属性
func (cm *ComponentManager) mergeProps(defaultProps, userProps map[string]any) map[string]any {
	merged := make(map[string]any)

	// 复制默认属性
	for key, value := range defaultProps {
		merged[key] = value
	}

	// 覆盖用户属性
	for key, value := range userProps {
		merged[key] = value
	}

	return merged
}

// mergeSlots 合并插槽
func (cm *ComponentManager) mergeSlots(slotNames []string, userSlots map[string]template.HTML) map[string]template.HTML {
	merged := make(map[string]template.HTML)

	// 初始化所有插槽
	for _, slotName := range slotNames {
		merged[slotName] = ""
	}

	// 设置用户插槽内容
	for slotName, content := range userSlots {
		merged[slotName] = content
	}

	return merged
}

// createSlotFunc 创建插槽函数
func (cm *ComponentManager) createSlotFunc(instance *ComponentInstance) func(string, ...template.HTML) template.HTML {
	return func(slotName string, defaultContent ...template.HTML) template.HTML {
		if content, exists := instance.Slots[slotName]; exists && content != "" {
			return content
		}

		if len(defaultContent) > 0 {
			return defaultContent[0]
		}

		return template.HTML("")
	}
}

// createPropFunc 创建属性函数
func (cm *ComponentManager) createPropFunc(instance *ComponentInstance) func(string, ...any) any {
	return func(propName string, defaultValue ...any) any {
		if value, exists := instance.Props[propName]; exists {
			return value
		}

		if len(defaultValue) > 0 {
			return defaultValue[0]
		}

		return nil
	}
}

// createEmitFunc 创建事件发射函数
func (cm *ComponentManager) createEmitFunc(instance *ComponentInstance) func(string, ...any) string {
	return func(eventName string, args ...any) string {
		// 这里可以实现事件系统
		return fmt.Sprintf("<!-- Event: %s -->", eventName)
	}
}

// ============= 预定义组件 =============

// RegisterDefaultComponents 注册默认组件
func RegisterDefaultComponents(manager *ComponentManager) error {
	// 页面头部组件
	headerComponent := &ComponentConfig{
		Name: "header",
		Path: "components/header.html",
		Props: map[string]any{
			"title": "默认标题",
			"logo":  "/static/images/logo.png",
		},
		Slots:       []string{"nav", "actions"},
		Description: "页面头部组件",
		Enabled:     true,
	}

	// 页面底部组件
	footerComponent := &ComponentConfig{
		Name: "footer",
		Path: "components/footer.html",
		Props: map[string]any{
			"copyright": "© 2024 All Rights Reserved",
			"links":     []string{},
		},
		Slots:       []string{"links", "social"},
		Description: "页面底部组件",
		Enabled:     true,
	}

	// 导航栏组件
	navbarComponent := &ComponentConfig{
		Name: "navbar",
		Path: "components/navbar.html",
		Props: map[string]any{
			"brand":    "Brand",
			"items":    []string{},
			"position": "top",
			"theme":    "light",
		},
		Slots:       []string{"brand", "items", "actions"},
		Description: "导航栏组件",
		Enabled:     true,
	}

	// 侧边栏组件
	sidebarComponent := &ComponentConfig{
		Name: "sidebar",
		Path: "components/sidebar.html",
		Props: map[string]any{
			"width":     "250px",
			"collapsed": false,
			"items":     []string{},
		},
		Slots:       []string{"header", "menu", "footer"},
		Description: "侧边栏组件",
		Enabled:     true,
	}

	// 面包屑组件
	breadcrumbComponent := &ComponentConfig{
		Name: "breadcrumb",
		Path: "components/breadcrumb.html",
		Props: map[string]any{
			"items":     []string{},
			"separator": "/",
		},
		Description: "面包屑导航组件",
		Enabled:     true,
	}

	// 卡片组件
	cardComponent := &ComponentConfig{
		Name: "card",
		Path: "components/card.html",
		Props: map[string]any{
			"title":  "",
			"shadow": true,
			"border": true,
		},
		Slots:       []string{"header", "body", "footer"},
		Description: "卡片组件",
		Enabled:     true,
	}

	// 模态框组件
	modalComponent := &ComponentConfig{
		Name: "modal",
		Path: "components/modal.html",
		Props: map[string]any{
			"title":    "Modal",
			"size":     "medium",
			"closable": true,
			"backdrop": true,
		},
		Slots:       []string{"header", "body", "footer"},
		Description: "模态框组件",
		Enabled:     true,
	}

	// 表格组件
	tableComponent := &ComponentConfig{
		Name: "table",
		Path: "components/table.html",
		Props: map[string]any{
			"columns":    []string{},
			"data":       []string{},
			"striped":    true,
			"bordered":   true,
			"hover":      true,
			"pagination": false,
		},
		Slots:       []string{"header", "body", "footer"},
		Description: "数据表格组件",
		Enabled:     true,
	}

	// 分页组件
	paginationComponent := &ComponentConfig{
		Name: "pagination",
		Path: "components/pagination.html",
		Props: map[string]any{
			"current":    1,
			"total":      0,
			"pageSize":   10,
			"showInfo":   true,
			"showJumper": true,
		},
		Description: "分页组件",
		Enabled:     true,
	}

	// 表单组件
	formComponent := &ComponentConfig{
		Name: "form",
		Path: "components/form.html",
		Props: map[string]any{
			"method": "POST",
			"action": "",
			"inline": false,
		},
		Slots:       []string{"fields", "actions"},
		Description: "表单组件",
		Enabled:     true,
	}

	// 按钮组件
	buttonComponent := &ComponentConfig{
		Name: "button",
		Path: "components/button.html",
		Props: map[string]any{
			"type":     "button",
			"variant":  "primary",
			"size":     "medium",
			"disabled": false,
			"loading":  false,
		},
		Slots:       []string{"default", "icon"},
		Description: "按钮组件",
		Enabled:     true,
	}

	// 输入框组件
	inputComponent := &ComponentConfig{
		Name: "input",
		Path: "components/input.html",
		Props: map[string]any{
			"type":        "text",
			"placeholder": "",
			"required":    false,
			"disabled":    false,
			"readonly":    false,
		},
		Description: "输入框组件",
		Enabled:     true,
	}

	// 注册所有组件
	components := []*ComponentConfig{
		headerComponent, footerComponent, navbarComponent, sidebarComponent,
		breadcrumbComponent, cardComponent, modalComponent, tableComponent,
		paginationComponent, formComponent, buttonComponent, inputComponent,
	}

	for _, component := range components {
		if err := manager.RegisterComponent(component.Name, component); err != nil {
			return fmt.Errorf("failed to register component %s: %w", component.Name, err)
		}
	}

	globalConfig.Info("Default components registered successfully")
	return nil
}

// ============= 组件缓存 =============

// ComponentCache 组件缓存
type ComponentCache struct {
	cache map[string]template.HTML
	mutex sync.RWMutex
}

// NewComponentCache 创建组件缓存
func NewComponentCache() *ComponentCache {
	return &ComponentCache{
		cache: make(map[string]template.HTML),
	}
}

// Get 获取缓存内容
func (cc *ComponentCache) Get(key string) (template.HTML, bool) {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	content, exists := cc.cache[key]
	return content, exists
}

// Set 设置缓存内容
func (cc *ComponentCache) Set(key string, content template.HTML) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.cache[key] = content
}

// Delete 删除缓存
func (cc *ComponentCache) Delete(key string) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	delete(cc.cache, key)
}

// Clear 清空缓存
func (cc *ComponentCache) Clear() {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.cache = make(map[string]template.HTML)
}

// Size 获取缓存大小
func (cc *ComponentCache) Size() int {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return len(cc.cache)
}

// ============= 组件构建器 =============

// ComponentBuilder 组件构建器
type ComponentBuilder struct {
	manager   *ComponentManager
	cache     *ComponentCache
	component string
	props     map[string]any
	slots     map[string]template.HTML
}

// NewComponentBuilder 创建组件构建器
func NewComponentBuilder(manager *ComponentManager) *ComponentBuilder {
	return &ComponentBuilder{
		manager: manager,
		cache:   NewComponentCache(),
		props:   make(map[string]any),
		slots:   make(map[string]template.HTML),
	}
}

// Component 设置组件名称
func (cb *ComponentBuilder) Component(name string) *ComponentBuilder {
	cb.component = name
	return cb
}

// Prop 设置属性
func (cb *ComponentBuilder) Prop(key string, value any) *ComponentBuilder {
	cb.props[key] = value
	return cb
}

// Props 设置多个属性
func (cb *ComponentBuilder) Props(props map[string]any) *ComponentBuilder {
	for key, value := range props {
		cb.props[key] = value
	}
	return cb
}

// Slot 设置插槽内容
func (cb *ComponentBuilder) Slot(name string, content template.HTML) *ComponentBuilder {
	cb.slots[name] = content
	return cb
}

// DefaultSlot 设置默认插槽内容
func (cb *ComponentBuilder) DefaultSlot(content template.HTML) *ComponentBuilder {
	cb.slots["default"] = content
	return cb
}

// Render 渲染组件
func (cb *ComponentBuilder) Render() (template.HTML, error) {
	if cb.component == "" {
		return "", fmt.Errorf("component name not set")
	}

	return cb.manager.RenderComponent(cb.component, cb.props, cb.slots)
}

// ============= 便捷函数 =============

// Component 创建组件构建器（全局函数）
func Component(name string) *ComponentBuilder {
	engine := GetDefaultEngine()
	manager := NewComponentManager(engine)

	// 注册默认组件
	if err := RegisterDefaultComponents(manager); err != nil {
		globalConfig.Errorf("Failed to register default components: %v", err)
	}

	return NewComponentBuilder(manager).Component(name)
}

// RenderComponentWithProps 渲染带属性的组件（全局函数）
func RenderComponentWithProps(name string, props map[string]any) (template.HTML, error) {
	return Component(name).Props(props).Render()
}
