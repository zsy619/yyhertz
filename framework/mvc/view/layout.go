package view

import (
	"fmt"
	"html/template"
	"strings"
	"sync"

	globalConfig "github.com/zsy619/yyhertz/framework/config"
)

// LayoutManager 布局管理器
type LayoutManager struct {
	engine        *TemplateEngine
	layouts       map[string]*LayoutConfig
	defaultLayout string
}

// LayoutConfig 布局配置
type LayoutConfig struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Sections    []string          `json:"sections"`
	Variables   map[string]string `json:"variables"`
	Parent      string            `json:"parent"`
	Components  []string          `json:"components"`
	Description string            `json:"description"`
	Enabled     bool              `json:"enabled"`
}

// NewLayoutManager 创建布局管理器
func NewLayoutManager(engine *TemplateEngine) *LayoutManager {
	return &LayoutManager{
		engine:  engine,
		layouts: make(map[string]*LayoutConfig),
	}
}

// RegisterLayout 注册布局
func (lm *LayoutManager) RegisterLayout(name string, config *LayoutConfig) error {
	if lm.layouts == nil {
		lm.layouts = make(map[string]*LayoutConfig)
	}

	config.Name = name
	if config.Sections == nil {
		config.Sections = []string{"content"}
	}
	if config.Variables == nil {
		config.Variables = make(map[string]string)
	}

	lm.layouts[name] = config

	globalConfig.Infof("Registered layout: %s", name)
	return nil
}

// SetDefaultLayout 设置默认布局
func (lm *LayoutManager) SetDefaultLayout(layoutName string) error {
	if _, exists := lm.layouts[layoutName]; !exists {
		return fmt.Errorf("layout '%s' not found", layoutName)
	}

	lm.defaultLayout = layoutName
	globalConfig.Infof("Set default layout: %s", layoutName)
	return nil
}

// GetDefaultLayout 获取默认布局
func (lm *LayoutManager) GetDefaultLayout() string {
	return lm.defaultLayout
}

// GetLayout 获取布局配置
func (lm *LayoutManager) GetLayout(name string) (*LayoutConfig, bool) {
	layout, exists := lm.layouts[name]
	return layout, exists
}

// GetAllLayouts 获取所有布局
func (lm *LayoutManager) GetAllLayouts() map[string]*LayoutConfig {
	return lm.layouts
}

// IsLayoutEnabled 检查布局是否启用
func (lm *LayoutManager) IsLayoutEnabled(name string) bool {
	if layout, exists := lm.layouts[name]; exists {
		return layout.Enabled
	}
	return false
}

// BuildLayoutChain 构建布局继承链
func (lm *LayoutManager) BuildLayoutChain(layoutName string) ([]*LayoutConfig, error) {
	chain := make([]*LayoutConfig, 0)
	visited := make(map[string]bool)

	current := layoutName
	for current != "" {
		if visited[current] {
			return nil, fmt.Errorf("circular dependency detected in layout chain: %s", current)
		}

		layout, exists := lm.layouts[current]
		if !exists {
			return nil, fmt.Errorf("layout '%s' not found", current)
		}

		if !layout.Enabled {
			return nil, fmt.Errorf("layout '%s' is disabled", current)
		}

		chain = append(chain, layout)
		visited[current] = true
		current = layout.Parent
	}

	// 反转链，使父布局在前
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain, nil
}

// ============= 布局渲染器 =============

// LayoutRenderer 布局渲染器
type LayoutRenderer struct {
	manager     *LayoutManager
	engine      *TemplateEngine
	currentData *RenderData
}

// NewLayoutRenderer 创建布局渲染器
func NewLayoutRenderer(manager *LayoutManager, engine *TemplateEngine) *LayoutRenderer {
	return &LayoutRenderer{
		manager: manager,
		engine:  engine,
	}
}

// RenderWithInheritance 使用继承渲染布局
func (lr *LayoutRenderer) RenderWithInheritance(templateName, layoutName string, data any) (string, error) {
	lr.currentData = lr.engine.prepareRenderData(data)

	// 如果没有指定布局，使用默认布局
	if layoutName == "" {
		layoutName = lr.manager.GetDefaultLayout()
	}

	if layoutName == "" {
		// 没有布局，直接渲染模板
		return lr.engine.Render(templateName, lr.currentData)
	}

	// 构建布局继承链
	layoutChain, err := lr.manager.BuildLayoutChain(layoutName)
	if err != nil {
		return "", fmt.Errorf("failed to build layout chain: %w", err)
	}

	// 渲染内容模板
	content, err := lr.engine.Render(templateName, lr.currentData)
	if err != nil {
		return "", fmt.Errorf("failed to render content template: %w", err)
	}

	// 为布局数据添加内容
	layoutData := lr.prepareLayoutData(content, templateName)

	// 从最内层布局开始渲染
	result := content
	for i := len(layoutChain) - 1; i >= 0; i-- {
		layout := layoutChain[i]

		// 更新布局数据
		layoutData["content"] = template.HTML(result)
		layoutData["layout"] = layout

		// 渲染当前布局
		result, err = lr.renderSingleLayout(layout, layoutData)
		if err != nil {
			return "", fmt.Errorf("failed to render layout '%s': %w", layout.Name, err)
		}
	}

	return result, nil
}

// renderSingleLayout 渲染单个布局
func (lr *LayoutRenderer) renderSingleLayout(layout *LayoutConfig, data map[string]any) (string, error) {
	// 获取布局模板
	layoutTemplate, exists := lr.engine.layouts[layout.Name]
	if !exists {
		return "", fmt.Errorf("layout template '%s' not found", layout.Name)
	}

	// 添加布局变量
	for key, value := range layout.Variables {
		data[key] = value
	}

	// 添加布局特定的函数
	data["section"] = lr.createSectionFunc(layout)
	data["yield"] = lr.createYieldFunc()
	data["block"] = lr.createBlockFunc()

	// 执行模板
	var buf strings.Builder
	if err := layoutTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("layout template execution error: %w", err)
	}

	return buf.String(), nil
}

// prepareLayoutData 准备布局数据
func (lr *LayoutRenderer) prepareLayoutData(content, templateName string) map[string]any {
	data := map[string]any{
		"content":      template.HTML(content),
		"templateName": templateName,
		"theme":        lr.engine.currentTheme,
		"engine":       lr.engine,
	}

	// 添加原始数据
	if lr.currentData != nil {
		data["data"] = lr.currentData.Data
		data["meta"] = lr.currentData.Meta
		data["flash"] = lr.currentData.Flash
		data["csrf"] = lr.currentData.CSRF
		data["user"] = lr.currentData.User
		data["request"] = lr.currentData.Request
	}

	return data
}

// createSectionFunc 创建section函数
func (lr *LayoutRenderer) createSectionFunc(layout *LayoutConfig) func(string) template.HTML {
	return func(sectionName string) template.HTML {
		// 检查是否是有效的section
		for _, section := range layout.Sections {
			if section == sectionName {
				return template.HTML(fmt.Sprintf("<!-- Section: %s -->", sectionName))
			}
		}
		return template.HTML("")
	}
}

// createYieldFunc 创建yield函数
func (lr *LayoutRenderer) createYieldFunc() func(...string) template.HTML {
	return func(sectionName ...string) template.HTML {
		if len(sectionName) == 0 {
			return template.HTML("{{.content}}")
		}
		return template.HTML(fmt.Sprintf("<!-- Yield: %s -->", sectionName[0]))
	}
}

// createBlockFunc 创建block函数
func (lr *LayoutRenderer) createBlockFunc() func(string, ...any) template.HTML {
	return func(blockName string, args ...any) template.HTML {
		return template.HTML(fmt.Sprintf("<!-- Block: %s -->", blockName))
	}
}

// ============= 预定义布局 =============

// RegisterDefaultLayouts 注册默认布局
func RegisterDefaultLayouts(manager *LayoutManager) error {
	// 基础布局
	baseLayout := &LayoutConfig{
		Name:     "base",
		Path:     "layouts/base.html",
		Sections: []string{"content", "title", "meta", "styles", "scripts"},
		Variables: map[string]string{
			"charset": "UTF-8",
			"lang":    "zh-CN",
		},
		Components:  []string{"header", "footer"},
		Description: "基础HTML5布局",
		Enabled:     true,
	}

	// 应用布局
	appLayout := &LayoutConfig{
		Name:        "app",
		Path:        "layouts/app.html",
		Sections:    []string{"content", "sidebar", "navbar"},
		Parent:      "base",
		Components:  []string{"navbar", "sidebar", "breadcrumb"},
		Description: "应用主布局",
		Enabled:     true,
	}

	// 管理后台布局
	adminLayout := &LayoutConfig{
		Name:     "admin",
		Path:     "layouts/admin.html",
		Sections: []string{"content", "sidebar", "header", "footer"},
		Parent:   "base",
		Variables: map[string]string{
			"theme": "admin",
			"brand": "管理后台",
		},
		Components:  []string{"admin-header", "admin-sidebar", "admin-footer"},
		Description: "管理后台布局",
		Enabled:     true,
	}

	// 简单布局
	simpleLayout := &LayoutConfig{
		Name:        "simple",
		Path:        "layouts/simple.html",
		Sections:    []string{"content"},
		Description: "简单布局",
		Enabled:     true,
	}

	// 邮件布局
	emailLayout := &LayoutConfig{
		Name:     "email",
		Path:     "layouts/email.html",
		Sections: []string{"content", "header", "footer"},
		Variables: map[string]string{
			"width": "600px",
			"align": "center",
		},
		Description: "邮件模板布局",
		Enabled:     true,
	}

	// 注册所有布局
	layouts := []*LayoutConfig{
		baseLayout, appLayout, adminLayout, simpleLayout, emailLayout,
	}

	for _, layout := range layouts {
		if err := manager.RegisterLayout(layout.Name, layout); err != nil {
			return fmt.Errorf("failed to register layout %s: %w", layout.Name, err)
		}
	}

	// 设置默认布局
	if err := manager.SetDefaultLayout("app"); err != nil {
		return fmt.Errorf("failed to set default layout: %w", err)
	}

	globalConfig.Info("Default layouts registered successfully")
	return nil
}

// ============= 模板区块系统 =============

// BlockManager 区块管理器
type BlockManager struct {
	blocks map[string]map[string]template.HTML // theme -> block_name -> content
	mutex  sync.RWMutex
}

// NewBlockManager 创建区块管理器
func NewBlockManager() *BlockManager {
	return &BlockManager{
		blocks: make(map[string]map[string]template.HTML),
	}
}

// DefineBlock 定义区块
func (bm *BlockManager) DefineBlock(theme, blockName string, content template.HTML) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if bm.blocks[theme] == nil {
		bm.blocks[theme] = make(map[string]template.HTML)
	}

	bm.blocks[theme][blockName] = content
}

// GetBlock 获取区块内容
func (bm *BlockManager) GetBlock(theme, blockName string) (template.HTML, bool) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	if themeBlocks, exists := bm.blocks[theme]; exists {
		if content, exists := themeBlocks[blockName]; exists {
			return content, true
		}
	}

	return "", false
}

// RenderBlock 渲染区块
func (bm *BlockManager) RenderBlock(theme, blockName string, defaultContent ...template.HTML) template.HTML {
	if content, exists := bm.GetBlock(theme, blockName); exists {
		return content
	}

	if len(defaultContent) > 0 {
		return defaultContent[0]
	}

	return template.HTML(fmt.Sprintf("<!-- Block not found: %s -->", blockName))
}

// GetBlockList 获取区块列表
func (bm *BlockManager) GetBlockList(theme string) []string {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	blocks := make([]string, 0)
	if themeBlocks, exists := bm.blocks[theme]; exists {
		for blockName := range themeBlocks {
			blocks = append(blocks, blockName)
		}
	}

	return blocks
}

// ClearBlocks 清除区块
func (bm *BlockManager) ClearBlocks(theme string) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if theme == "" {
		// 清除所有主题的区块
		bm.blocks = make(map[string]map[string]template.HTML)
	} else {
		// 清除指定主题的区块
		delete(bm.blocks, theme)
	}
}

// ============= 便捷函数 =============

// RenderWithLayout 使用布局渲染（全局函数）
func RenderWithLayoutInheritance(templateName, layoutName string, data any) (string, error) {
	engine := GetDefaultEngine()
	manager := NewLayoutManager(engine)

	// 注册默认布局
	if err := RegisterDefaultLayouts(manager); err != nil {
		return "", fmt.Errorf("failed to register default layouts: %w", err)
	}

	renderer := NewLayoutRenderer(manager, engine)
	return renderer.RenderWithInheritance(templateName, layoutName, data)
}
