package view

import (
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// BeegoLayoutManager Beego风格的布局管理器
type BeegoLayoutManager struct {
	layouts         map[string]*BeegoLayout  // 布局缓存
	layoutMutex     sync.RWMutex             // 布局锁
	defaultLayout   string                   // 默认布局
	enableInherit   bool                     // 是否启用继承
}

// BeegoLayout Beego风格的布局定义
type BeegoLayout struct {
	Name            string                   // 布局名称
	Template        *template.Template       // 布局模板
	Parent          string                   // 父布局
	Sections        map[string]string        // 区块定义
	Blocks          map[string]*BeegoBlock   // 块定义
	ContentBlocks   []string                 // 内容块列表
	LayoutFile      string                   // 布局文件路径
}

// BeegoBlock Beego风格的模板块
type BeegoBlock struct {
	Name            string    // 块名称
	Content         string    // 块内容
	IsOverridden    bool      // 是否被重写
	Parent          string    // 父块名称
}

// NewBeegoLayoutManager 创建Beego布局管理器
func NewBeegoLayoutManager() *BeegoLayoutManager {
	return &BeegoLayoutManager{
		layouts:       make(map[string]*BeegoLayout),
		defaultLayout: "main",
		enableInherit: true,
	}
}

// RegisterLayout 注册布局
func (m *BeegoLayoutManager) RegisterLayout(name string, templateContent string, parentLayout ...string) error {
	m.layoutMutex.Lock()
	defer m.layoutMutex.Unlock()
	
	layout := &BeegoLayout{
		Name:          name,
		Sections:      make(map[string]string),
		Blocks:        make(map[string]*BeegoBlock),
		ContentBlocks: make([]string, 0),
		LayoutFile:    fmt.Sprintf("%s.html", name),
	}
	
	// 设置父布局
	if len(parentLayout) > 0 {
		layout.Parent = parentLayout[0]
	}
	
	// 解析模板内容中的区块和块定义
	if err := m.parseLayoutContent(layout, templateContent); err != nil {
		return fmt.Errorf("failed to parse layout content: %w", err)
	}
	
	// 创建模板实例
	tmpl := template.New(name)
	if _, err := tmpl.Parse(templateContent); err != nil {
		return fmt.Errorf("failed to parse layout template: %w", err)
	}
	
	layout.Template = tmpl
	m.layouts[name] = layout
	
	config.Infof("Registered Beego layout: %s", name)
	return nil
}

// parseLayoutContent 解析布局内容
func (m *BeegoLayoutManager) parseLayoutContent(layout *BeegoLayout, content string) error {
	lines := strings.Split(content, "\n")
	
	var currentSection string
	var currentBlock *BeegoBlock
	var blockContent strings.Builder
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		
		// 解析 @section 定义
		if strings.HasPrefix(trimmed, "@section") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				currentSection = strings.Trim(parts[1], "()'\"")
				layout.ContentBlocks = append(layout.ContentBlocks, currentSection)
				config.Debugf("Found section: %s at line %d", currentSection, i+1)
			}
		}
		
		// 解析 @yield 指令
		if strings.Contains(trimmed, "@yield") {
			parts := strings.Fields(trimmed)
			for _, part := range parts {
				if strings.HasPrefix(part, "@yield(") {
					yieldName := strings.Trim(part, "@yield()'\"")
					if yieldName != "" {
						layout.Sections[yieldName] = fmt.Sprintf("{{.%s}}", strings.Title(yieldName))
						config.Debugf("Found yield: %s at line %d", yieldName, i+1)
					}
				}
			}
		}
		
		// 解析 {{block}} 定义
		if strings.Contains(trimmed, "{{block") {
			blockName := m.extractBlockName(trimmed)
			if blockName != "" {
				currentBlock = &BeegoBlock{
					Name:         blockName,
					IsOverridden: false,
				}
				blockContent.Reset()
				config.Debugf("Started block: %s at line %d", blockName, i+1)
			}
		}
		
		// 收集块内容
		if currentBlock != nil {
			blockContent.WriteString(line + "\n")
		}
		
		// 结束块定义
		if currentBlock != nil && strings.Contains(trimmed, "{{end}}") {
			currentBlock.Content = blockContent.String()
			layout.Blocks[currentBlock.Name] = currentBlock
			config.Debugf("Ended block: %s at line %d", currentBlock.Name, i+1)
			currentBlock = nil
		}
	}
	
	return nil
}

// extractBlockName 提取块名称
func (m *BeegoLayoutManager) extractBlockName(line string) string {
	// 查找 {{block "name" 模式
	if idx := strings.Index(line, "{{block"); idx >= 0 {
		remaining := line[idx+7:] // 跳过 "{{block"
		if idx := strings.Index(remaining, "\""); idx >= 0 {
			nameStart := idx + 1
			if endIdx := strings.Index(remaining[nameStart:], "\""); endIdx >= 0 {
				return remaining[nameStart : nameStart+endIdx]
			}
		}
	}
	return ""
}

// GetLayout 获取布局
func (m *BeegoLayoutManager) GetLayout(name string) (*BeegoLayout, error) {
	m.layoutMutex.RLock()
	defer m.layoutMutex.RUnlock()
	
	layout, exists := m.layouts[name]
	if !exists {
		return nil, fmt.Errorf("layout '%s' not found", name)
	}
	
	return layout, nil
}

// BuildInheritanceChain 构建继承链
func (m *BeegoLayoutManager) BuildInheritanceChain(layoutName string) ([]*BeegoLayout, error) {
	var chain []*BeegoLayout
	current := layoutName
	visited := make(map[string]bool)
	
	for current != "" {
		if visited[current] {
			return nil, fmt.Errorf("circular inheritance detected in layout: %s", current)
		}
		visited[current] = true
		
		layout, err := m.GetLayout(current)
		if err != nil {
			return nil, fmt.Errorf("layout '%s' not found in inheritance chain", current)
		}
		
		chain = append([]*BeegoLayout{layout}, chain...) // 在前面插入
		current = layout.Parent
	}
	
	return chain, nil
}

// RenderWithLayout 使用布局渲染内容
func (m *BeegoLayoutManager) RenderWithLayout(layoutName string, contentTemplate *template.Template, data any) (string, error) {
	layout, err := m.GetLayout(layoutName)
	if err != nil {
		return "", err
	}
	
	// 如果启用继承，构建继承链
	if m.enableInherit && layout.Parent != "" {
		return m.renderWithInheritance(layoutName, contentTemplate, data)
	}
	
	// 简单布局渲染
	return m.renderSimpleLayout(layout, contentTemplate, data)
}

// renderWithInheritance 使用继承渲染
func (m *BeegoLayoutManager) renderWithInheritance(layoutName string, contentTemplate *template.Template, data any) (string, error) {
	chain, err := m.BuildInheritanceChain(layoutName)
	if err != nil {
		return "", err
	}
	
	// 合并所有布局的模板和块定义
	mergedTemplate := template.New("merged")
	mergedBlocks := make(map[string]*BeegoBlock)
	
	// 从基础布局开始合并
	for _, layout := range chain {
		// 添加布局模板定义
		if layout.Template != nil {
			for _, tmpl := range layout.Template.Templates() {
				mergedTemplate = template.Must(mergedTemplate.AddParseTree(tmpl.Name(), tmpl.Tree))
			}
		}
		
		// 合并块定义（子布局的块覆盖父布局）
		for name, block := range layout.Blocks {
			mergedBlocks[name] = block
		}
	}
	
	// 准备渲染数据
	renderData := m.prepareLayoutData(data, mergedBlocks)
	
	// 执行渲染
	var result strings.Builder
	if err := mergedTemplate.Execute(&result, renderData); err != nil {
		return "", fmt.Errorf("failed to execute layout template: %w", err)
	}
	
	return result.String(), nil
}

// renderSimpleLayout 简单布局渲染
func (m *BeegoLayoutManager) renderSimpleLayout(layout *BeegoLayout, contentTemplate *template.Template, data any) (string, error) {
	// 准备渲染数据
	renderData := m.prepareLayoutData(data, layout.Blocks)
	
	// 执行布局模板
	var result strings.Builder
	if err := layout.Template.Execute(&result, renderData); err != nil {
		return "", fmt.Errorf("failed to execute simple layout: %w", err)
	}
	
	return result.String(), nil
}

// prepareLayoutData 准备布局数据
func (m *BeegoLayoutManager) prepareLayoutData(data any, blocks map[string]*BeegoBlock) map[string]any {
	renderData := make(map[string]any)
	
	// 添加原始数据
	if data != nil {
		switch v := data.(type) {
		case map[string]any:
			for k, val := range v {
				renderData[k] = val
			}
		default:
			renderData["Data"] = data
		}
	}
	
	// 添加块内容
	for name, block := range blocks {
		renderData[strings.Title(name)] = template.HTML(block.Content)
	}
	
	// 添加布局辅助函数的数据
	renderData["LayoutContent"] = renderData["Content"]
	if renderData["Content"] == nil {
		renderData["LayoutContent"] = ""
	}
	
	return renderData
}

// SetDefaultLayout 设置默认布局
func (m *BeegoLayoutManager) SetDefaultLayout(name string) {
	m.layoutMutex.Lock()
	defer m.layoutMutex.Unlock()
	
	m.defaultLayout = name
	config.Infof("Default layout set to: %s", name)
}

// GetDefaultLayout 获取默认布局
func (m *BeegoLayoutManager) GetDefaultLayout() string {
	m.layoutMutex.RLock()
	defer m.layoutMutex.RUnlock()
	
	return m.defaultLayout
}

// ListLayouts 列出所有布局
func (m *BeegoLayoutManager) ListLayouts() []string {
	m.layoutMutex.RLock()
	defer m.layoutMutex.RUnlock()
	
	layouts := make([]string, 0, len(m.layouts))
	for name := range m.layouts {
		layouts = append(layouts, name)
	}
	
	return layouts
}

// GetLayoutInfo 获取布局信息
func (m *BeegoLayoutManager) GetLayoutInfo(name string) (map[string]any, error) {
	layout, err := m.GetLayout(name)
	if err != nil {
		return nil, err
	}
	
	info := map[string]any{
		"name":           layout.Name,
		"parent":         layout.Parent,
		"sections":       layout.Sections,
		"blocks":         len(layout.Blocks),
		"content_blocks": layout.ContentBlocks,
		"layout_file":    layout.LayoutFile,
	}
	
	// 添加块详情
	blockInfo := make(map[string]map[string]any)
	for name, block := range layout.Blocks {
		blockInfo[name] = map[string]any{
			"name":         block.Name,
			"is_overridden": block.IsOverridden,
			"parent":       block.Parent,
			"content_length": len(block.Content),
		}
	}
	info["block_details"] = blockInfo
	
	return info, nil
}

// ClearLayouts 清空所有布局
func (m *BeegoLayoutManager) ClearLayouts() {
	m.layoutMutex.Lock()
	defer m.layoutMutex.Unlock()
	
	m.layouts = make(map[string]*BeegoLayout)
	config.Info("All layouts cleared")
}

// EnableInheritance 启用/禁用继承
func (m *BeegoLayoutManager) EnableInheritance(enable bool) {
	m.layoutMutex.Lock()
	defer m.layoutMutex.Unlock()
	
	m.enableInherit = enable
	config.Infof("Layout inheritance: %v", enable)
}