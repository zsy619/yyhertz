package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
)

// EnhancedTemplateEngine 增强的模板引擎，支持Beego风格的template include
type EnhancedTemplateEngine struct {
	*TemplateEngine
	
	// 模板定义缓存
	templateDefs    map[string]*template.Template // 已定义的模板
	templateDefsMux sync.RWMutex
	
	// 全局模板实例（包含所有定义）
	globalTemplate *template.Template
	globalMux      sync.RWMutex
}

// NewEnhancedTemplateEngine 创建增强模板引擎
func NewEnhancedTemplateEngine(cfg *TemplateConfig) (*EnhancedTemplateEngine, error) {
	baseEngine, err := NewTemplateEngine(cfg)
	if err != nil {
		return nil, err
	}
	
	enhanced := &EnhancedTemplateEngine{
		TemplateEngine: baseEngine,
		templateDefs:   make(map[string]*template.Template),
	}
	
	// 注册增强的模板函数
	enhanced.registerEnhancedFunctions()
	
	// 重新加载所有模板以支持include
	if err := enhanced.loadAllTemplatesWithIncludes(); err != nil {
		return nil, fmt.Errorf("failed to load templates with includes: %w", err)
	}
	
	return enhanced, nil
}

// registerEnhancedFunctions 注册增强的模板函数
func (e *EnhancedTemplateEngine) registerEnhancedFunctions() {
	// 合并现有函数
	for name, fn := range e.funcMap {
		BeegoTemplateFuncs[name] = fn
	}
	
	// 添加模板include相关函数
	BeegoTemplateFuncs["include"] = e.includeTemplateFunc
	BeegoTemplateFuncs["template"] = e.templateFunc
	BeegoTemplateFuncs["partial"] = e.partialFunc
	BeegoTemplateFuncs["component"] = e.componentFunc
	BeegoTemplateFuncs["render"] = e.renderFunc
	
	// 更新函数映射
	e.funcMap = BeegoTemplateFuncs
}

// loadAllTemplatesWithIncludes 加载所有模板并支持includes
func (e *EnhancedTemplateEngine) loadAllTemplatesWithIncludes() error {
	e.globalMux.Lock()
	defer e.globalMux.Unlock()
	
	// 创建全局模板实例
	e.globalTemplate = template.New("global").
		Delims(e.delimLeft, e.delimRight).
		Funcs(e.funcMap)
	
	// 第一步：扫描并收集所有模板文件
	templateFiles := make(map[string]string) // name -> path
	
	// 扫描所有视图路径
	for _, viewPath := range e.viewPaths {
		if err := e.scanTemplateFiles(viewPath, templateFiles); err != nil {
			config.Warnf("Error scanning view path %s: %v", viewPath, err)
		}
	}
	
	// 扫描布局目录
	if e.layoutPath != "" {
		if err := e.scanTemplateFiles(e.layoutPath, templateFiles); err != nil {
			config.Warnf("Error scanning layout path %s: %v", e.layoutPath, err)
		}
	}
	
	// 扫描组件目录
	if e.componentPath != "" {
		if err := e.scanTemplateFiles(e.componentPath, templateFiles); err != nil {
			config.Warnf("Error scanning component path %s: %v", e.componentPath, err)
		}
	}
	
	// 第二步：解析所有模板文件到全局模板
	for _, filePath := range templateFiles {
		if _, err := e.globalTemplate.ParseFiles(filePath); err != nil {
			config.Errorf("Failed to parse template file %s: %v", filePath, err)
			continue
		}
	}
	
	// 第三步：为每个模板创建单独的实例
	e.templateDefsMux.Lock()
	defer e.templateDefsMux.Unlock()
	
	e.templateDefs = make(map[string]*template.Template)
	
	for name := range templateFiles {
		// 克隆全局模板为每个模板创建实例
		if tmpl := e.globalTemplate.Lookup(name); tmpl != nil {
			cloned, err := e.globalTemplate.Clone()
			if err != nil {
				config.Errorf("Failed to clone template for %s: %v", name, err)
				continue
			}
			e.templateDefs[name] = cloned
		}
	}
	
	config.Infof("Loaded %d template files with include support", len(templateFiles))
	return nil
}

// scanTemplateFiles 扫描模板文件
func (e *EnhancedTemplateEngine) scanTemplateFiles(dir string, files map[string]string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续
		}
		
		if !d.IsDir() && strings.HasSuffix(path, e.extension) {
			// 计算模板名称
			name := e.getTemplateName(path)
			files[name] = path
			
			// 同时使用文件名（不含扩展名）作为key
			baseName := strings.TrimSuffix(filepath.Base(path), e.extension)
			if baseName != name {
				files[baseName] = path
			}
		}
		
		return nil
	})
}

// RenderWithIncludes 支持include的渲染方法
func (e *EnhancedTemplateEngine) RenderWithIncludes(templateName string, data any) (string, error) {
	e.templateDefsMux.RLock()
	defer e.templateDefsMux.RUnlock()
	
	// 确保模板名有正确的格式
	templateKey := e.normalizeTemplateName(templateName)
	
	// 查找模板
	tmpl := e.findTemplate(templateKey)
	if tmpl == nil {
		return "", fmt.Errorf("template '%s' not found", templateName)
	}
	
	// 准备渲染数据
	renderData := e.prepareRenderData(data)
	
	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateKey, renderData); err != nil {
		// 如果指定名称失败，尝试使用第一个可用的模板
		if err := tmpl.Execute(&buf, renderData); err != nil {
			return "", fmt.Errorf("template execution error: %w", err)
		}
	}
	
	return buf.String(), nil
}

// findTemplate 查找模板
func (e *EnhancedTemplateEngine) findTemplate(templateName string) *template.Template {
	// 尝试直接查找
	if tmpl, exists := e.templateDefs[templateName]; exists {
		return tmpl
	}
	
	// 尝试从全局模板查找
	e.globalMux.RLock()
	defer e.globalMux.RUnlock()
	
	if e.globalTemplate != nil {
		if tmpl := e.globalTemplate.Lookup(templateName); tmpl != nil {
			// 克隆一份用于独立渲染
			if cloned, err := e.globalTemplate.Clone(); err == nil {
				return cloned
			}
		}
	}
	
	return nil
}

// normalizeTemplateName 标准化模板名称
func (e *EnhancedTemplateEngine) normalizeTemplateName(name string) string {
	// 移除扩展名
	if strings.HasSuffix(name, e.extension) {
		name = strings.TrimSuffix(name, e.extension)
	}
	
	// 标准化路径分隔符
	name = strings.ReplaceAll(name, "\\", "/")
	
	return name
}

// ============= 模板函数实现 =============

// includeTemplateFunc include模板函数
func (e *EnhancedTemplateEngine) includeTemplateFunc(templateName string, data ...any) template.HTML {
	var templateData any
	if len(data) > 0 {
		templateData = data[0]
	}
	
	content, err := e.RenderWithIncludes(templateName, templateData)
	if err != nil {
		config.Errorf("Include template error: %v", err)
		return template.HTML(fmt.Sprintf("<!-- Include error: %s -->", err.Error()))
	}
	
	return template.HTML(content)
}

// templateFunc template模板函数（与include相同）
func (e *EnhancedTemplateEngine) templateFunc(templateName string, data ...any) template.HTML {
	return e.includeTemplateFunc(templateName, data...)
}

// partialFunc partial模板函数
func (e *EnhancedTemplateEngine) partialFunc(templateName string, data ...any) template.HTML {
	// Partial通常用于组件，可能在特定目录下
	partialName := templateName
	if !strings.Contains(templateName, "/") {
		partialName = "partials/" + templateName
	}
	
	return e.includeTemplateFunc(partialName, data...)
}

// componentFunc 组件函数
func (e *EnhancedTemplateEngine) componentFunc(componentName string, data ...any) template.HTML {
	// 组件通常在components目录下
	if !strings.Contains(componentName, "/") {
		componentName = "components/" + componentName
	}
	
	return e.includeTemplateFunc(componentName, data...)
}

// renderFunc 渲染函数
func (e *EnhancedTemplateEngine) renderFunc(templateName string, data ...any) template.HTML {
	return e.includeTemplateFunc(templateName, data...)
}

// ============= 辅助方法 =============

// CreateTemplate 创建模板定义
func (e *EnhancedTemplateEngine) CreateTemplate(name, content string) error {
	e.globalMux.Lock()
	defer e.globalMux.Unlock()
	
	if e.globalTemplate == nil {
		e.globalTemplate = template.New("global").
			Delims(e.delimLeft, e.delimRight).
			Funcs(e.funcMap)
	}
	
	// 解析模板内容
	tmpl, err := e.globalTemplate.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	
	// 添加到定义缓存
	e.templateDefsMux.Lock()
	defer e.templateDefsMux.Unlock()
	
	e.templateDefs[name] = tmpl
	
	config.Debugf("Created template definition: %s", name)
	return nil
}

// GetTemplateDefinition 获取模板定义
func (e *EnhancedTemplateEngine) GetTemplateDefinition(name string) *template.Template {
	e.templateDefsMux.RLock()
	defer e.templateDefsMux.RUnlock()
	
	return e.templateDefs[name]
}

// ListTemplateDefinitions 列出所有模板定义
func (e *EnhancedTemplateEngine) ListTemplateDefinitions() []string {
	e.templateDefsMux.RLock()
	defer e.templateDefsMux.RUnlock()
	
	names := make([]string, 0, len(e.templateDefs))
	for name := range e.templateDefs {
		names = append(names, name)
	}
	
	return names
}

// ReloadTemplates 重新加载所有模板
func (e *EnhancedTemplateEngine) ReloadTemplates() error {
	config.Info("Reloading all templates with include support")
	return e.loadAllTemplatesWithIncludes()
}

// ============= 便捷函数 =============

// GetEnhancedEngine 获取增强模板引擎实例
func GetEnhancedEngine() *EnhancedTemplateEngine {
	// 创建新的增强引擎实例
	cfg := DefaultTemplateConfig()
	enhanced, err := NewEnhancedTemplateEngine(cfg)
	if err != nil {
		config.Errorf("Failed to create enhanced template engine: %v", err)
		return nil
	}
	
	return enhanced
}

// RenderWithIncludes 渲染支持include的模板（便捷函数）
func RenderWithIncludes(templateName string, data any) (string, error) {
	engine := GetEnhancedEngine()
	if engine == nil {
		return "", fmt.Errorf("enhanced template engine not available")
	}
	
	return engine.RenderWithIncludes(templateName, data)
}

// CreateGlobalTemplate 创建全局模板定义（便捷函数）
func CreateGlobalTemplate(name, content string) error {
	engine := GetEnhancedEngine()
	if engine == nil {
		return fmt.Errorf("enhanced template engine not available")
	}
	
	return engine.CreateTemplate(name, content)
}