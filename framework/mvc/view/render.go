package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// CSRFTokenProvider CSRF token提供者接口
//
// 此接口用于解决循环导入问题，允许view包获取CSRF token
// 而不需要直接依赖mvc包
type CSRFTokenProvider interface {
	// GenerateSimpleToken 生成简单的CSRF token
	GenerateSimpleToken() string
}

// 全局CSRF token提供者
var (
	globalCSRFProvider CSRFTokenProvider
	csrfProviderMutex  sync.RWMutex
)

// SetCSRFTokenProvider 设置CSRF token提供者
//
// 此方法由mvc包调用，用于注册CSRF token提供者
func SetCSRFTokenProvider(provider CSRFTokenProvider) {
	csrfProviderMutex.Lock()
	defer csrfProviderMutex.Unlock()
	globalCSRFProvider = provider
}

// GetCSRFTokenProvider 获取CSRF token提供者
func GetCSRFTokenProvider() CSRFTokenProvider {
	csrfProviderMutex.RLock()
	defer csrfProviderMutex.RUnlock()
	return globalCSRFProvider
}

// RenderData 渲染数据结构
type RenderData struct {
	Data           any               `json:"data"`
	Meta           *MetaData         `json:"meta,omitempty"`
	Flash          *FlashData        `json:"flash,omitempty"`
	CSRF           string            `json:"csrf,omitempty"`
	CsrfToken      string            `json:"csrf_token,omitempty"` // 添加驼峰命名字段，用于模板中的{{.CsrfToken}}访问
	Theme          string            `json:"theme,omitempty"`
	User           any               `json:"user,omitempty"`
	Request        *RequestData      `json:"request,omitempty"`
	LayoutSections map[string]string `json:"layout_sections,omitempty"` // 布局区块内容
}

// ============= 缓存键管理器 =============

// CacheKeyManager 缓存键管理器
type CacheKeyManager struct {
	delimiter string // 分隔符
}

// DefaultCacheKeyManager 默认缓存键管理器
var DefaultCacheKeyManager = &CacheKeyManager{
	delimiter: "@",
}

// GenerateLayoutKey 生成布局模板缓存键
func (ckm *CacheKeyManager) GenerateLayoutKey(templateName, layoutName string) string {
	if layoutName == "" {
		return templateName
	}
	return fmt.Sprintf("%s%s%s", templateName, ckm.delimiter, layoutName)
}

// GenerateTemplateKey 生成模板缓存键
func (ckm *CacheKeyManager) GenerateTemplateKey(templateName string) string {
	// 🔧 关键修复：缓存键应该使用文件basename（与LoadTemplate中的模板名称保持一致）
	// 因为LoadTemplate中使用filepath.Base(templatePath)作为模板名，缓存键也应该对应

	// 先获取文件的basename
	baseName := filepath.Base(templateName)

	// 再去掉扩展名
	idx := strings.LastIndex(baseName, ".")
	if idx > 0 {
		baseName = baseName[:idx]
	}

	config.Debugf("🔧 CacheKey generation: original=%s, basename=%s, final_key=%s", templateName, filepath.Base(templateName), baseName)
	return baseName
}

// ParseLayoutKey 解析布局缓存键
func (ckm *CacheKeyManager) ParseLayoutKey(key string) (templateName, layoutName string) {
	parts := strings.Split(key, ckm.delimiter)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return key, ""
}

// IsLayoutKey 检查是否是布局缓存键
func (ckm *CacheKeyManager) IsLayoutKey(key string) bool {
	return strings.Contains(key, ckm.delimiter)
}

// ValidateKey 验证缓存键的有效性
func (ckm *CacheKeyManager) ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("cache key cannot be empty")
	}
	if strings.Contains(key, " ") {
		return fmt.Errorf("cache key cannot contain spaces: %s", key)
	}
	return nil
}

// ============= 更新的模板渲染和缓存方法 =============

// GetCSRFToken 获取CSRF令牌（别名方法）
func (r *RenderData) GetCSRFToken() string {
	return r.CSRF
}

// Csrf_token 获取CSRF令牌（下划线命名）
func (r *RenderData) Csrf_token() string {
	return r.CSRF
}

// MetaData 页面元数据
type MetaData struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    string            `json:"keywords"`
	Author      string            `json:"author"`
	Canonical   string            `json:"canonical"`
	Image       string            `json:"image"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// FlashData 闪存消息
type FlashData struct {
	Success []string `json:"success,omitempty"`
	Error   []string `json:"error,omitempty"`
	Warning []string `json:"warning,omitempty"`
	Info    []string `json:"info,omitempty"`
}

// RequestData 请求信息
type RequestData struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Query     string `json:"query"`
	UserAgent string `json:"user_agent"`
	IP        string `json:"ip"`
	Timestamp int64  `json:"timestamp"`
}

// Render 渲染模板
func (e *TemplateEngine) Render(templateName string, data any) (string, error) {
	return e.RenderWithLayout(templateName, "", data)
}

// RenderWithLayout 使用布局渲染模板
func (e *TemplateEngine) RenderWithLayout(templateName, layoutName string, data any) (string, error) {
	// 准备渲染数据
	renderData := e.prepareRenderData(data)

	var tmpl *template.Template
	var err error

	if layoutName != "" {
		// 使用布局渲染
		tmpl, err = e.GetTemplateWithLayout(templateName, layoutName)
	} else {
		// 直接渲染模板
		tmpl, err = e.GetTemplate(templateName)
	}

	if err != nil {
		return "", fmt.Errorf("template loading error: %w", err)
	}

	// 渲染模板 - 增强调试信息
	var buf bytes.Buffer
	config.Debugf("🎯 开始执行模板: %s (布局: %s)", templateName, layoutName)
	config.Debugf("📦 渲染数据类型: %T, 数据内容预览: %+v", renderData.Data, renderData.Data)

	if err := tmpl.Execute(&buf, renderData.Data); err != nil {
		// 提供更详细的错误信息
		errorDetails := fmt.Sprintf(`
❌ 模板执行失败详情:
  - 模板名称: %s
  - 布局名称: %s  
  - 模板路径: %v
  - 错误类型: %T
  - 错误内容: %v
  - 数据类型: %T
  - 可用函数数量: %d`,
			templateName, layoutName, e.viewPaths, err, err, renderData.Data, len(e.funcMap))

		return "", fmt.Errorf("template execution error: %s", errorDetails)
	}

	result := buf.String()

	// 如果启用压缩，移除多余空白
	if e.enableCompress {
		result = e.compressHTML(result)
	}

	return result, nil
}

// RenderWithLayoutSections 使用布局和区块渲染模板
func (e *TemplateEngine) RenderWithLayoutSections(templateName, layoutName string, data any, layoutSections map[string]string) (string, error) {
	// 准备渲染数据
	renderData := e.prepareRenderData(data)

	// 添加布局区块到渲染数据中
	if renderData.LayoutSections == nil {
		renderData.LayoutSections = make(map[string]string)
	}

	// 合并传入的布局区块
	for key, value := range layoutSections {
		renderData.LayoutSections[key] = value
	}

	var tmpl *template.Template
	var err error

	if layoutName != "" {
		// 使用布局渲染（支持布局区块）
		tmpl, err = e.GetTemplateWithLayoutSections(templateName, layoutName, layoutSections)
	} else {
		// 直接渲染模板
		tmpl, err = e.GetTemplate(templateName)
	}

	if err != nil {
		return "", fmt.Errorf("template loading error: %w", err)
	}

	// 渲染模板 - 增强调试信息 (布局区块版本)
	var buf bytes.Buffer
	config.Debugf("🎯 开始执行布局区块模板: %s (布局: %s, 区块数: %d)", templateName, layoutName, len(layoutSections))
	config.Debugf("📦 渲染数据类型: %T, 数据内容预览: %+v", renderData.Data, renderData.Data)

	if err := tmpl.Execute(&buf, renderData.Data); err != nil {
		// 提供更详细的错误信息
		sectionNames := make([]string, 0, len(layoutSections))
		for name := range layoutSections {
			sectionNames = append(sectionNames, name)
		}

		errorDetails := fmt.Sprintf(`
❌ 布局区块模板执行失败详情:
  - 模板名称: %s
  - 布局名称: %s
  - 布局区块: %v  
  - 模板路径: %v
  - 错误类型: %T
  - 错误内容: %v
  - 数据类型: %T
  - 可用函数数量: %d`,
			templateName, layoutName, sectionNames, e.viewPaths, err, err, renderData.Data, len(e.funcMap))

		return "", fmt.Errorf("template execution error: %s", errorDetails)
	}

	result := buf.String()

	// 如果启用压缩，移除多余空白
	if e.enableCompress {
		result = e.compressHTML(result)
	}

	return result, nil
}

// RenderComponent 渲染组件
func (e *TemplateEngine) RenderComponent(componentName string, data any) (string, error) {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	component, exists := e.components[componentName]
	if !exists {
		return "", fmt.Errorf("component '%s' not found", componentName)
	}

	renderData := e.prepareRenderData(data)

	var buf bytes.Buffer
	if err := component.Execute(&buf, renderData); err != nil {
		return "", fmt.Errorf("component execution error: %w", err)
	}

	return buf.String(), nil
}

// GetTemplate 获取模板（增强版，支持 Beego 热重载）
func (e *TemplateEngine) GetTemplate(templateName string) (*template.Template, error) {
	// 使用统一的缓存键管理器
	cacheKey := DefaultCacheKeyManager.GenerateTemplateKey(templateName)

	// 验证缓存键的有效性
	if err := DefaultCacheKeyManager.ValidateKey(cacheKey); err != nil {
		return nil, fmt.Errorf("invalid template cache key: %w", err)
	}

	// Beego风格：开发模式下检查热重载
	if e.RunMode == "dev" && e.enableReload {
		if e.shouldReloadTemplate(templateName) {
			// 重新加载模板
			if err := e.reloadTemplate(templateName); err != nil {
				config.Warnf("Failed to reload template %s: %v", templateName, err)
			}
		}
	}

	// 从缓存获取
	if e.enableCache {
		if tmpl, exists := e.templates[cacheKey]; exists {
			config.Debugf("✅ Template %s loaded from cache (key: %s)", templateName, cacheKey)
			config.Debugf("🔧 CACHED template name: %s", tmpl.Name())
			return tmpl, nil
		}
	}

	// 动态加载模板
	config.Debugf("🔄 Template %s not in cache, loading from disk (key: %s)", templateName, cacheKey)
	config.Debugf("📋 Available template functions: %d", len(e.funcMap))

	// 检查关键函数是否存在
	criticalFuncs := []string{"formatFileSize", "dateformat", "now"}
	for _, funcName := range criticalFuncs {
		if _, exists := e.funcMap[funcName]; exists {
			config.Debugf("✅ Critical function '%s' is available", funcName)
		} else {
			config.Warnf("⚠️ Critical function '%s' is missing", funcName)
		}
	}
	tmpl, err := e.LoadTemplate(templateName)
	if err != nil {
		config.Errorf("Failed to load template %s: %v", templateName, err)
		return nil, err
	}

	// 缓存模板 - 确保缓存的是正确的模板
	if e.enableCache {
		e.templates[cacheKey] = tmpl
		config.Debugf("Template %s cached successfully with key: %s, template name: %s", templateName, cacheKey, tmpl.Name())
	}

	config.Debugf("🔧 GetTemplate final result: requested=%s, returned_name=%s", templateName, tmpl.Name())
	return tmpl, nil
}

// GetTemplateWithLayout 获取带布局的模板（支持布局继承）
func (e *TemplateEngine) GetTemplateWithLayout(templateName, layoutName string) (*template.Template, error) {
	// 【关键修复】标准化布局名称，去除可能的.html扩展名
	// 确保布局名称与注册时的名称一致
	normalizedLayoutName := layoutName
	if strings.HasSuffix(layoutName, e.extension) {
		normalizedLayoutName = strings.TrimSuffix(layoutName, e.extension)
		config.Debugf("Normalized layout name from '%s' to '%s'", layoutName, normalizedLayoutName)
	}

	// 【关键修复】：基于Beego机制生成正确的缓存键
	// 先获取模板的基础名称，然后生成匹配的缓存键
	contentPath, err := e.FindTemplateFile(normalizedLayoutName)
	if err != nil {
		return nil, fmt.Errorf("content template '%s' not found: %w", normalizedLayoutName, err)
	}

	templateBaseName := e.getTemplateBaseName(contentPath)
	// 将布局名称中的斜杠替换为下划线，保持与实际模板名一致
	safeLayoutName := strings.ReplaceAll(normalizedLayoutName, "/", "_")
	combinedTemplateName := fmt.Sprintf("%s_with_%s", templateBaseName, safeLayoutName)
	cacheKey := combinedTemplateName // 使用实际模板名作为缓存键

	// 先用读锁检查缓存
	if e.enableCache {
		// e.templateMutex.RLock()
		if tmpl, exists := e.templates[cacheKey]; exists {
			// e.templateMutex.RUnlock()
			config.Debugf("Template with layout %s@%s loaded from cache with key: %s", templateName, normalizedLayoutName, cacheKey)
			return tmpl, nil
		}
		// e.templateMutex.RUnlock()
	}

	// 需要创建新模板，使用写锁
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 双重检查，防止在等待写锁期间其他协程已经创建了模板
	if e.enableCache {
		if tmpl, exists := e.templates[cacheKey]; exists {
			config.Debugf("Template with layout %s@%s loaded from cache (double-check) with key: %s", templateName, normalizedLayoutName, cacheKey)
			return tmpl, nil
		}
	}

	config.Debugf("Creating new template with layout %s@%s (cache key: %s)", templateName, normalizedLayoutName, cacheKey)

	// 检查布局是否存在
	if _, exists := e.layouts[normalizedLayoutName]; !exists {
		// 获取可用布局列表（避免调用GetLayoutList导致死锁）
		availableLayouts := make([]string, 0, len(e.layouts))
		for name := range e.layouts {
			availableLayouts = append(availableLayouts, name)
		}
		return nil, fmt.Errorf("layout '%s' not found, available layouts: %v", normalizedLayoutName, availableLayouts)
	}

	// 优先使用基于Beego机制的布局组合
	combinedTemplate, err := e.createOptimizedTemplateWithLayout(templateName, normalizedLayoutName)
	if err != nil {
		// 如果Beego机制失败，尝试布局继承系统作为备选方案
		config.Debugf("Beego-style layout failed for %s@%s: %v, trying layout inheritance", templateName, normalizedLayoutName, err)
		if tmpl, inheritErr := e.createTemplateWithLayoutInheritance(templateName, normalizedLayoutName); inheritErr == nil {
			// 为继承方案也使用相同的缓存键命名规则
			inheritanceKey := fmt.Sprintf("%s_inherit_%s", templateBaseName, normalizedLayoutName)
			if e.enableCache {
				e.templates[inheritanceKey] = tmpl
				config.Debugf("Cached template with layout inheritance %s@%s (key: %s)", templateName, normalizedLayoutName, inheritanceKey)
			}
			return tmpl, nil
		} else {
			config.Debugf("Layout inheritance also failed for %s@%s: %v", templateName, normalizedLayoutName, inheritErr)
		}
		return nil, fmt.Errorf("failed to create template with layout using both Beego-style and inheritance methods: %w", err)
	}

	// 缓存成功创建的模板，使用与模板名匹配的缓存键
	if e.enableCache && combinedTemplate != nil {
		e.templates[cacheKey] = combinedTemplate
		config.Debugf("Cached Beego-style template with layout %s@%s (key: %s, template name: %s)", templateName, normalizedLayoutName, cacheKey, combinedTemplate.Name())
	}

	return combinedTemplate, nil
}

// GetTemplateWithLayoutSections 获取包含布局区块的模板
func (e *TemplateEngine) GetTemplateWithLayoutSections(templateName, layoutName string, layoutSections map[string]string) (*template.Template, error) {
	// 【关键修复】标准化布局名称，去除可能的.html扩展名
	normalizedLayoutName := layoutName
	if strings.HasSuffix(layoutName, e.extension) {
		normalizedLayoutName = strings.TrimSuffix(layoutName, e.extension)
		config.Debugf("Normalized layout name from '%s' to '%s' in sections", layoutName, normalizedLayoutName)
	}

	// 基于现有的 GetTemplateWithLayout 但支持布局区块处理
	contentPath, err := e.FindTemplateFile(templateName)
	if err != nil {
		return nil, fmt.Errorf("content template '%s' not found: %w", templateName, err)
	}

	templateBaseName := e.getTemplateBaseName(contentPath)

	// 为布局区块生成特殊的缓存键，包含区块内容的哈希
	sectionsHash := e.GenerateSectionsHash(layoutSections)
	safeLayoutName := strings.ReplaceAll(normalizedLayoutName, "/", "_")
	combinedTemplateName := fmt.Sprintf("%s_with_%s_sections_%s", templateBaseName, safeLayoutName, sectionsHash)
	cacheKey := combinedTemplateName

	// 检查缓存
	if e.enableCache {
		if tmpl, exists := e.templates[cacheKey]; exists {
			config.Debugf("Template with layout sections %s@%s loaded from cache", templateName, normalizedLayoutName)
			return tmpl, nil
		}
	}

	// 需要创建新模板
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	// 双重检查缓存
	if e.enableCache {
		if tmpl, exists := e.templates[cacheKey]; exists {
			return tmpl, nil
		}
	}

	config.Debugf("Creating new template with layout sections %s@%s", templateName, normalizedLayoutName)

	// 检查布局是否存在
	if _, exists := e.layouts[normalizedLayoutName]; !exists {
		availableLayouts := make([]string, 0, len(e.layouts))
		for name := range e.layouts {
			availableLayouts = append(availableLayouts, name)
		}
		return nil, fmt.Errorf("layout '%s' not found, available layouts: %v", normalizedLayoutName, availableLayouts)
	}

	// 创建支持布局区块的模板
	combinedTemplate, err := e.createTemplateWithLayoutSections(templateName, normalizedLayoutName, layoutSections)
	if err != nil {
		return nil, fmt.Errorf("failed to create template with layout sections: %w", err)
	}

	// 缓存模板
	if e.enableCache && combinedTemplate != nil {
		e.templates[cacheKey] = combinedTemplate
		config.Debugf("Cached template with layout sections %s@%s", templateName, normalizedLayoutName)
	}

	return combinedTemplate, nil
}

// createTemplateWithLayoutInheritance 使用布局继承系统创建模板
func (e *TemplateEngine) createTemplateWithLayoutInheritance(templateName, layoutName string) (*template.Template, error) {
	// 创建临时的LayoutManager来处理布局继承
	layoutManager := NewLayoutManager(e)

	// 注册默认布局（如果还没有注册的话）
	if err := e.ensureDefaultLayoutsRegistered(layoutManager); err != nil {
		return nil, fmt.Errorf("failed to register default layouts: %w", err)
	}

	// 构建布局继承链
	layoutChain, err := layoutManager.BuildLayoutChain(layoutName)
	if err != nil {
		return nil, fmt.Errorf("failed to build layout chain: %w", err)
	}

	if len(layoutChain) == 0 {
		return nil, fmt.Errorf("empty layout chain for '%s'", layoutName)
	}

	// 使用最顶层的布局作为基础模板
	topLayout := layoutChain[len(layoutChain)-1]
	baseLayout, exists := e.layouts[topLayout.Name]
	if !exists {
		return nil, fmt.Errorf("top layout '%s' not found in cached layouts", topLayout.Name)
	}

	// 动态获取最新的合并函数
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 克隆基础布局模板
	tmpl, err := baseLayout.Clone()
	if err != nil {
		return nil, fmt.Errorf("failed to clone top layout '%s': %w", topLayout.Name, err)
	}

	// 重新设置分隔符和函数映射
	tmpl = tmpl.Delims(e.delimLeft, e.delimRight).Funcs(mergedFuncs)

	// 加载内容模板
	contentPath, err := e.FindTemplateFile(templateName)
	if err != nil {
		return nil, fmt.Errorf("failed to find content template '%s': %w", templateName, err)
	}

	// 解析内容模板
	if _, err := tmpl.ParseFiles(contentPath); err != nil {
		return nil, fmt.Errorf("failed to parse content template '%s': %w", templateName, err)
	}

	config.Debugf("Successfully created template with layout inheritance: %s@%s", templateName, layoutName)
	return tmpl, nil
}

// createOptimizedTemplateWithLayout 创建优化的模板+布局组合（基于Beego机制）
func (e *TemplateEngine) createOptimizedTemplateWithLayout(templateName, layoutName string) (*template.Template, error) {
	// 检查布局是否存在
	if _, exists := e.layouts[layoutName]; !exists {
		return nil, fmt.Errorf("layout '%s' not found", layoutName)
	}

	// 动态获取最新的合并函数
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 查找内容模板文件路径
	contentPath, err := e.FindTemplateFile(templateName)
	if err != nil {
		return nil, fmt.Errorf("content template '%s' not found: %w", templateName, err)
	}

	config.Debugf("Creating template with layout using Beego-style mechanism: %s@%s, content path: %s", templateName, layoutName, contentPath)

	// 【关键修复】：使用Beego风格的模板组合机制
	// 1. 创建新的模板实例，使用统一的命名规则
	templateBaseName := e.getTemplateBaseName(contentPath)
	// 将布局名称中的斜杠替换为下划线，避免模板名称中的路径问题
	safeLayoutName := strings.ReplaceAll(layoutName, "/", "_")
	combinedTemplateName := fmt.Sprintf("%s_with_%s", templateBaseName, safeLayoutName)

	// 2. 创建新模板并设置分隔符和函数
	finalTemplate := template.New(combinedTemplateName).Delims(e.delimLeft, e.delimRight).Funcs(mergedFuncs)

	// 3. 读取内容模板
	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read content template '%s' at %s: %w", templateName, contentPath, err)
	}
	contentTemplate := string(contentBytes)

	// 4. 获取布局模板内容
	layoutContent, err := e.getLayoutContent(layoutName)
	if err != nil {
		return nil, fmt.Errorf("failed to get layout content for '%s': %w", layoutName, err)
	}

	// 5. 实现 {{.LayoutContent}} 占位符替换机制
	finalTemplateContent, err := e.embedContentInLayout(layoutContent, contentTemplate, templateBaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to embed content in layout: %w", err)
	}

	// 6. 解析最终合并的模板内容
	if _, err := finalTemplate.Parse(finalTemplateContent); err != nil {
		return nil, fmt.Errorf("failed to parse combined template '%s@%s': %w", templateName, layoutName, err)
	}

	// 验证最终模板的有效性
	if finalTemplate == nil {
		return nil, fmt.Errorf("parsed template is nil for '%s' with layout '%s'", templateName, layoutName)
	}

	// 确保模板具有有效的执行树
	if err := e.validateTemplate(finalTemplate); err != nil {
		return nil, fmt.Errorf("template validation failed for '%s' with layout '%s': %w", templateName, layoutName, err)
	}

	config.Debugf("Successfully created Beego-style template with layout: %s@%s, final template name: %s", templateName, layoutName, finalTemplate.Name())
	return finalTemplate, nil
}

// getTemplateBaseName 获取模板的基础名称（用于缓存键）
func (e *TemplateEngine) getTemplateBaseName(templatePath string) string {
	// 移除路径和扩展名，获取纯文件名
	baseName := filepath.Base(templatePath)
	ext := filepath.Ext(baseName)
	if ext != "" {
		baseName = strings.TrimSuffix(baseName, ext)
	}
	return baseName
}

// getLayoutContent 获取布局模板的内容
func (e *TemplateEngine) getLayoutContent(layoutName string) (string, error) {
	// 处理布局名称，如果已经包含目录前缀，需要去掉基础目录部分
	layoutFileName := layoutName
	if strings.HasPrefix(layoutName, "layouts/") {
		layoutFileName = strings.TrimPrefix(layoutName, "layouts/")
	}

	// 尝试在布局路径中搜索文件
	layoutDir := e.layoutPath
	if layoutDir != "" {
		possiblePath := filepath.Join(layoutDir, layoutFileName+".html")
		if _, err := os.Stat(possiblePath); err == nil {
			contentBytes, err := os.ReadFile(possiblePath)
			if err != nil {
				return "", fmt.Errorf("failed to read layout file '%s': %w", possiblePath, err)
			}
			return e.processLayoutContent(string(contentBytes)), nil
		}

		// 尝试其他扩展名
		for _, ext := range []string{".tpl", ".gohtml", ".tmpl"} {
			possiblePath = filepath.Join(layoutDir, layoutFileName+ext)
			if _, err := os.Stat(possiblePath); err == nil {
				contentBytes, err := os.ReadFile(possiblePath)
				if err != nil {
					return "", fmt.Errorf("failed to read layout file '%s': %w", possiblePath, err)
				}
				return e.processLayoutContent(string(contentBytes)), nil
			}
		}
	}

	// 如果layoutPath为空或未找到，尝试在所有视图路径中搜索
	for _, viewDir := range e.viewPaths {
		// 检查 layouts 子目录
		layoutsDir := filepath.Join(viewDir, "layouts")
		possiblePath := filepath.Join(layoutsDir, layoutName+".html")
		if _, err := os.Stat(possiblePath); err == nil {
			contentBytes, err := os.ReadFile(possiblePath)
			if err != nil {
				return "", fmt.Errorf("failed to read layout file '%s': %w", possiblePath, err)
			}
			return e.processLayoutContent(string(contentBytes)), nil
		}

		// 尝试其他扩展名
		for _, ext := range []string{".tpl", ".gohtml", ".tmpl"} {
			possiblePath = filepath.Join(layoutsDir, layoutName+ext)
			if _, err := os.Stat(possiblePath); err == nil {
				contentBytes, err := os.ReadFile(possiblePath)
				if err != nil {
					return "", fmt.Errorf("failed to read layout file '%s': %w", possiblePath, err)
				}
				return e.processLayoutContent(string(contentBytes)), nil
			}
		}
	}

	return "", fmt.Errorf("layout file for '%s' not found in layout path '%s' or any view path", layoutName, e.layoutPath)
}

// processLayoutContent 处理布局内容，支持 {{.LayoutContent}} 占位符
func (e *TemplateEngine) processLayoutContent(layoutContent string) string {
	// 支持多种布局内容占位符格式，兼容不同的模板引擎风格
	contentPlaceholders := []string{
		"{{.LayoutContent}}",             // 主要格式
		"{{.Content}}",                   // 通用格式
		"{{ .LayoutContent }}",           // 带空格格式
		"{{ .Content }}",                 // 带空格通用格式
		"{{template \"content\" .}}",     // Beego 风格
		"{{yield}}",                      // Rails 风格
		"{{block \"content\" .}}{{end}}", // Go 标准模板 block 格式
	}

	// 为每种占位符格式添加标记，以便在渲染时识别和替换
	processedContent := layoutContent
	for _, placeholder := range contentPlaceholders {
		if strings.Contains(processedContent, placeholder) {
			config.Debugf("Found layout content placeholder: %s", placeholder)
			// 暂时保持占位符不变，在实际渲染时会被替换为具体的模板内容
			// 这里可以添加一些标记来帮助后续处理
			break
		}
	}

	return processedContent
}

// embedContentInLayout 将内容模板嵌入到布局模板中，支持 {{.LayoutContent}} 占位符
func (e *TemplateEngine) embedContentInLayout(layoutContent, contentTemplate, templateBaseName string) (string, error) {
	// 支持多种布局内容占位符格式
	contentPlaceholders := []string{
		"{{.LayoutContent}}",             // 主要格式
		"{{.Content}}",                   // 通用格式
		"{{ .LayoutContent }}",           // 带空格格式
		"{{ .Content }}",                 // 带空格通用格式
		"{{template \"content\" .}}",     // Beego 风格
		"{{yield}}",                      // Rails 风格
		"{{block \"content\" .}}{{end}}", // Go 标准模板 block 格式
	}

	finalContent := layoutContent
	contentEmbedded := false

	// 尝试替换每种占位符格式
	for _, placeholder := range contentPlaceholders {
		if strings.Contains(finalContent, placeholder) {
			config.Debugf("Replacing layout content placeholder '%s' with template content", placeholder)

			// 特殊处理不同的占位符格式
			switch {
			case strings.Contains(placeholder, "block"):
				// Go 标准模板 block 格式需要定义 block 内容
				var blockContent string
				if strings.Contains(contentTemplate, `{{define "content"}}`) {
					// 内容模板已经有定义，直接使用
					blockContent = contentTemplate
				} else {
					// 包装内容模板
					blockContent = fmt.Sprintf(`{{define "content"}}%s{{end}}`, contentTemplate)
				}
				finalContent = strings.ReplaceAll(finalContent, placeholder, blockContent)

			case strings.Contains(placeholder, "template"):
				// Beego 风格 template 调用需要定义对应的模板
				if strings.Contains(contentTemplate, `{{define "content"}}`) {
					// 内容模板已经有定义，直接放到前面
					finalContent = contentTemplate + "\n" + finalContent
				} else {
					// 包装内容模板
					templateDef := fmt.Sprintf(`{{define "content"}}%s{{end}}`, contentTemplate)
					finalContent = templateDef + "\n" + finalContent
				}

			default:
				// 直接替换占位符
				finalContent = strings.ReplaceAll(finalContent, placeholder, contentTemplate)
			}

			contentEmbedded = true
			config.Debugf("Successfully embedded content template into layout using placeholder: %s", placeholder)
			break
		}
	}

	// 如果没有找到占位符，使用默认的嵌入策略
	if !contentEmbedded {
		config.Debugf("No layout content placeholder found, using default embedding strategy")
		// 检查布局模板的结构，智能插入内容
		if strings.Contains(layoutContent, "<body>") && strings.Contains(layoutContent, "</body>") {
			// HTML 模板，在 body 结束前插入内容
			finalContent = strings.Replace(layoutContent, "</body>", contentTemplate+"\n</body>", 1)
			config.Debugf("Embedded content before </body> tag")
		} else if strings.Contains(layoutContent, "{{") {
			// 包含模板语法，在末尾添加内容
			finalContent = layoutContent + "\n" + contentTemplate
			config.Debugf("Appended content to layout template")
		} else {
			// 纯文本布局，直接连接
			finalContent = layoutContent + "\n" + contentTemplate
			config.Debugf("Concatenated content with layout")
		}
	}

	return finalContent, nil
}

// processLayoutSections 处理布局区块，支持 {{.HtmlHead}}、{{.SideBar}}、{{.Scripts}} 等
func (e *TemplateEngine) processLayoutSections(layoutContent string, layoutSections map[string]string) string {
	if layoutSections == nil || len(layoutSections) == 0 {
		return layoutContent
	}

	processedContent := layoutContent

	// 支持的布局区块格式
	sectionPlaceholders := []string{
		"HtmlHead", "SideBar", "Scripts", "Styles", "Meta", "Header", "Footer",
		"Navigation", "Breadcrumb", "Content", "MainContent", "Sidebar",
		"LeftSidebar", "RightSidebar", "TopBar", "BottomBar", "Modal", "Alert",
	}

	// 处理每个已定义的布局区块
	for sectionName, sectionContent := range layoutSections {
		// 支持多种占位符格式
		placeholderFormats := []string{
			fmt.Sprintf("{{.%s}}", sectionName),                                    // {{.HtmlHead}}
			fmt.Sprintf("{{ .%s }}", sectionName),                                  // {{ .HtmlHead }}
			fmt.Sprintf("{{.%s | safehtml}}", sectionName),                         // {{.HtmlHead | safehtml}}
			fmt.Sprintf("{{template \"%s\" .}}", strings.ToLower(sectionName)),     // {{template "htmlhead" .}}
			fmt.Sprintf("{{block \"%s\" .}}{{end}}", strings.ToLower(sectionName)), // {{block "htmlhead" .}}{{end}}
		}

		for _, placeholder := range placeholderFormats {
			if strings.Contains(processedContent, placeholder) {
				config.Debugf("Replacing layout section placeholder '%s' with content (%d chars)", placeholder, len(sectionContent))
				processedContent = strings.ReplaceAll(processedContent, placeholder, sectionContent)
			}
		}
	}

	// 同时处理预定义的常用布局区块占位符，即使没有在 LayoutSections 中定义
	for _, sectionName := range sectionPlaceholders {
		sectionContent, exists := layoutSections[sectionName]
		if !exists {
			// 如果没有定义，用空字符串替换占位符，避免显示未处理的模板语法
			sectionContent = ""
		}

		emptyPlaceholders := []string{
			fmt.Sprintf("{{.%s}}", sectionName),
			fmt.Sprintf("{{ .%s }}", sectionName),
		}

		for _, placeholder := range emptyPlaceholders {
			if strings.Contains(processedContent, placeholder) {
				processedContent = strings.ReplaceAll(processedContent, placeholder, sectionContent)
			}
		}
	}

	config.Debugf("Processed %d layout sections", len(layoutSections))
	return processedContent
}

// GenerateSectionsHash 为布局区块生成哈希值，用于缓存键
func (e *TemplateEngine) GenerateSectionsHash(layoutSections map[string]string) string {
	if len(layoutSections) == 0 {
		return "empty"
	}

	// 创建确定性的哈希
	keys := make([]string, 0, len(layoutSections))
	for key := range layoutSections {
		keys = append(keys, key)
	}

	// 排序确保一致性
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	hashContent := ""
	for _, key := range keys {
		hashContent += fmt.Sprintf("%s:%s;", key, layoutSections[key])
	}

	// 简单哈希（前8位）
	hash := 0
	for _, c := range hashContent {
		hash = hash*31 + int(c)
	}

	if hash < 0 {
		hash = -hash
	}

	return fmt.Sprintf("%x", hash)[:8]
}

// createTemplateWithLayoutSections 创建支持布局区块的模板
func (e *TemplateEngine) createTemplateWithLayoutSections(templateName, layoutName string, layoutSections map[string]string) (*template.Template, error) {
	// 获取动态合并函数
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 查找内容模板文件路径
	contentPath, err := e.FindTemplateFile(templateName)
	if err != nil {
		return nil, fmt.Errorf("content template '%s' not found: %w", templateName, err)
	}

	config.Debugf("Creating template with layout sections: %s@%s, content path: %s, sections: %d", templateName, layoutName, contentPath, len(layoutSections))

	// 创建新模板实例
	templateBaseName := e.getTemplateBaseName(contentPath)
	safeLayoutName := strings.ReplaceAll(layoutName, "/", "_")
	combinedTemplateName := fmt.Sprintf("%s_with_%s_sections", templateBaseName, safeLayoutName)

	finalTemplate := template.New(combinedTemplateName).Delims(e.delimLeft, e.delimRight).Funcs(mergedFuncs)

	// 读取内容模板
	contentBytes, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read content template '%s' at %s: %w", templateName, contentPath, err)
	}
	contentTemplate := string(contentBytes)

	// 获取布局模板内容
	layoutContent, err := e.getLayoutContent(layoutName)
	if err != nil {
		return nil, fmt.Errorf("failed to get layout content for '%s': %w", layoutName, err)
	}

	// 处理布局区块
	layoutWithSections := e.processLayoutSections(layoutContent, layoutSections)

	// 嵌入内容到布局中
	finalTemplateContent, err := e.embedContentInLayout(layoutWithSections, contentTemplate, templateBaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to embed content in layout: %w", err)
	}

	// 解析最终模板内容
	if _, err := finalTemplate.Parse(finalTemplateContent); err != nil {
		return nil, fmt.Errorf("failed to parse combined template '%s@%s': %w", templateName, layoutName, err)
	}

	// 验证模板
	if err := e.validateTemplate(finalTemplate); err != nil {
		return nil, fmt.Errorf("template validation failed for '%s' with layout '%s': %w", templateName, layoutName, err)
	}

	config.Debugf("Successfully created template with layout sections: %s@%s", templateName, layoutName)
	return finalTemplate, nil
}

// validateTemplate 验证模板的有效性
func (e *TemplateEngine) validateTemplate(tmpl *template.Template) error {
	if tmpl == nil {
		return fmt.Errorf("template is nil")
	}

	// 检查模板是否有可执行的内容
	templates := tmpl.Templates()
	hasValidTemplate := false

	for _, t := range templates {
		if t.Tree != nil && t.Tree.Root != nil {
			hasValidTemplate = true
			config.Debugf("Found valid template: %s", t.Name())
		}
	}

	if !hasValidTemplate {
		return fmt.Errorf("no valid executable templates found")
	}

	return nil
}

// ensureDefaultLayoutsRegistered 确保默认布局已注册
func (e *TemplateEngine) ensureDefaultLayoutsRegistered(layoutManager *LayoutManager) error {
	// 检查是否已经有布局注册
	if len(layoutManager.GetAllLayouts()) > 0 {
		return nil // 已经有布局了
	}

	// 尝试注册默认布局
	return RegisterDefaultLayouts(layoutManager)
}

// LoadTemplate 动态加载模板（增强错误处理）
func (e *TemplateEngine) LoadTemplate(templateName string) (*template.Template, error) {
	templatePath, err := e.FindTemplateFile(templateName)
	if err != nil {
		// 【增强错误处理】提供详细的查找失败信息
		config.Errorf("❌ Template file not found: %s", templateName)
		config.Errorf("   Searched paths: %v", e.viewPaths)
		config.Errorf("   Extension: %s", e.extension)
		return nil, fmt.Errorf("template '%s' not found: %w", templateName, err)
	}

	// 添加调试日志
	config.Debugf("🔄 Loading template: %s from path: %s", templateName, templatePath)

	// 🔧 关键修复：正确的Go模板加载机制
	// 方法1：直接使用ParseFiles创建模板，不预先创建空模板
	// 这样可以确保ParseFiles创建的模板就是主模板

	// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 🔧 核心修复：ParseFiles会自动创建以文件basename命名的主模板
	// 我们不需要预先创建模板，这样避免了空模板覆盖有内容的模板的问题
	tmpl, err := template.New("").
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs).
		ParseFiles(templatePath)

	if err != nil {
		config.Errorf("❌ Failed to parse template %s at %s: %v", templateName, templatePath, err)
		// 尝试提供更有用的错误信息
		if strings.Contains(err.Error(), "function") {
			config.Errorf("   💡 Hint: Check if all template functions are properly registered")
			config.Errorf("   📋 Available functions: %d", len(mergedFuncs))
			for name := range mergedFuncs {
				config.Debugf("     - %s", name)
			}
		}
		return nil, fmt.Errorf("failed to parse template %s at %s: %w", templateName, templatePath, err)
	}

	// 查找实际的模板文件名（通常是文件的basename）
	templates := tmpl.Templates()
	var actualTemplate *template.Template

	config.Debugf("📋 Available templates after parsing:")
	for _, t := range templates {
		hasValidTree := t.Tree != nil && t.Tree.Root != nil
		config.Debugf("  - Template name: %s, Tree: %v, Root: %v", t.Name(), t.Tree != nil, hasValidTree)
	}

	// 🔧 增强模板查找逻辑：查找由ParseFiles创建的实际模板
	baseFileName := filepath.Base(templatePath)
	config.Debugf("🎯 Looking for template with name: %s", baseFileName)

	for _, t := range templates {
		if t.Tree != nil && t.Tree.Root != nil {
			// 🔧 关键修复：优先选择文件basename匹配的模板，这是ParseFiles创建的实际模板
			if t.Name() == baseFileName {
				actualTemplate = t
				config.Debugf("✅ Found exact file basename match: %s", t.Name())
				break
			}
			// 如果没找到精确匹配，选择第一个有效模板作为备选
			if actualTemplate == nil {
				actualTemplate = t
				config.Debugf("📌 Fallback template selected: %s", t.Name())
			}
		}
	}

	if actualTemplate == nil {
		templateNames := getTemplateNames(templates)
		config.Errorf("❌ No executable template found in %s", templateName)
		config.Errorf("   📋 Parsed templates: %v", templateNames)
		config.Errorf("   🎯 Expected template name: %s", baseFileName)

		return nil, fmt.Errorf("template %s is empty or invalid - no executable template found in parsed templates: %v", templateName, templateNames)
	}

	config.Debugf("✅ Final selected template: %s (requested: %s)", actualTemplate.Name(), templateName)

	// 最后验证模板的可执行性
	if err := e.validateTemplateExecution(actualTemplate, templateName); err != nil {
		config.Warnf("⚠️ Template execution validation warning for %s: %v", templateName, err)
	}

	return actualTemplate, nil
}

// validateTemplateFunctions 验证模板函数的可用性
func (e *TemplateEngine) validateTemplateFunctions(funcMap template.FuncMap) error {
	// 检查关键函数是否存在
	criticalFunctions := []string{
		"dateformat", "now", "formatFileSize", "safeHTML",
		"csrf_token", "json", "dict", "slice",
	}

	var missingFunctions []string
	for _, funcName := range criticalFunctions {
		if _, exists := funcMap[funcName]; !exists {
			missingFunctions = append(missingFunctions, funcName)
		}
	}

	if len(missingFunctions) > 0 {
		return fmt.Errorf("missing critical template functions: %v", missingFunctions)
	}

	config.Debugf("✅ All critical template functions are available (%d total)", len(funcMap))
	return nil
}

// validateTemplateExecution 验证模板的可执行性
func (e *TemplateEngine) validateTemplateExecution(tmpl *template.Template, templateName string) error {
	if tmpl == nil {
		return fmt.Errorf("template is nil")
	}

	// 检查模板是否有执行树
	if tmpl.Tree == nil || tmpl.Tree.Root == nil {
		return fmt.Errorf("template has no execution tree")
	}

	// 尝试用空数据执行模板以验证语法
	testData := map[string]any{
		"now": time.Now(),
		"Stats": map[string]any{
			"total": 0,
		},
	}

	var testBuf strings.Builder
	if err := tmpl.Execute(&testBuf, testData); err != nil {
		// 如果是函数未定义错误，提供具体建议
		if strings.Contains(err.Error(), "function") && strings.Contains(err.Error(), "not defined") {
			return fmt.Errorf("template execution test failed - function not defined: %w\n"+
				"💡 Hint: Check if all template functions are properly registered before template parsing", err)
		}
		return fmt.Errorf("template execution test failed: %w", err)
	}

	config.Debugf("✅ Template execution validation passed for: %s", templateName)
	return nil
}

// recoverFromTemplateLoadError 从模板加载错误中恢复
func (e *TemplateEngine) recoverFromTemplateLoadError(templateName string, originalErr error) (*template.Template, error) {
	config.Warnf("🔧 Attempting to recover from template load error for: %s", templateName)

	// 1. 尝试清除相关缓存并重新加载
	e.templateMutex.Lock()
	delete(e.templates, templateName)
	e.templateMutex.Unlock()

	// 2. 重新验证函数注册
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)
	if err := e.validateTemplateFunctions(mergedFuncs); err != nil {
		config.Warnf("Function validation failed during recovery: %v", err)
		// 重新注册默认函数
		e.registerDefaultFunctions()
		mergedFuncs = manager.GetMergedFunctions(e.funcMap)
	}

	// 3. 尝试使用基础模板名称加载
	baseTemplateName := strings.TrimSuffix(templateName, e.extension)
	if baseTemplateName != templateName {
		config.Debugf("Trying base template name: %s", baseTemplateName)
		if tmpl, err := e.loadTemplateWithName(baseTemplateName, mergedFuncs); err == nil {
			return tmpl, nil
		}
	}

	// 4. 尝试在所有视图路径中查找相似名称的模板
	similarTemplates := e.findSimilarTemplates(templateName)
	if len(similarTemplates) > 0 {
		config.Warnf("Found similar templates: %v. Consider using one of these instead.", similarTemplates)
	}

	// 5. 返回原始错误，但增加恢复建议
	return nil, fmt.Errorf("template recovery failed for '%s': %w\n"+
		"💡 Recovery suggestions:\n"+
		"  - Check if template file exists and has correct extension\n"+
		"  - Verify template syntax is valid\n"+
		"  - Ensure all used functions are registered\n"+
		"  - Similar templates found: %v", templateName, originalErr, similarTemplates)
}

// loadTemplateWithName 使用指定名称加载模板
func (e *TemplateEngine) loadTemplateWithName(templateName string, funcMap template.FuncMap) (*template.Template, error) {
	templatePath, err := e.FindTemplateFile(templateName)
	if err != nil {
		return nil, err
	}

	// 使用传入的函数映射
	tmpl := template.New(filepath.Base(templatePath)).
		Delims(e.delimLeft, e.delimRight).
		Funcs(funcMap)

	_, err = tmpl.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

// findSimilarTemplates 查找相似的模板名称
func (e *TemplateEngine) findSimilarTemplates(templateName string) []string {
	var similar []string
	baseName := strings.ToLower(strings.TrimSuffix(templateName, e.extension))

	// 遍历所有视图路径查找相似模板
	for _, viewPath := range e.viewPaths {
		filepath.Walk(viewPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			if !strings.HasSuffix(path, e.extension) {
				return nil
			}

			fileName := strings.ToLower(filepath.Base(path))
			fileBaseName := strings.TrimSuffix(fileName, e.extension)

			// 简单的相似性检查
			if strings.Contains(fileBaseName, baseName) || strings.Contains(baseName, fileBaseName) {
				relPath, _ := filepath.Rel(viewPath, path)
				similar = append(similar, strings.TrimSuffix(relPath, e.extension))
			}

			return nil
		})
	}

	return similar
}

// getTemplateNames 获取模板名称列表（用于调试）
func getTemplateNames(templates []*template.Template) []string {
	names := make([]string, len(templates))
	for i, t := range templates {
		names[i] = t.Name()
	}
	return names
}

// FindTemplateFile 查找模板文件（增强版）
func (e *TemplateEngine) FindTemplateFile(templateName string) (string, error) {
	// 确保模板名有扩展名
	if !strings.HasSuffix(templateName, e.extension) {
		templateName += e.extension
	}

	config.Debugf("🔍 查找模板文件: %s，在路径: %v", templateName, e.viewPaths)

	// 获取当前工作目录用于调试
	wd, _ := os.Getwd()
	config.Debugf("  当前工作目录: %s", wd)

	// 在所有视图路径中搜索
	for _, viewPath := range e.viewPaths {
		// 先尝试绝对路径
		templatePath := viewPath
		if !filepath.IsAbs(viewPath) {
			// 如果是相对路径，尝试不同的基准目录
			possibleBases := []string{
				wd,                            // 当前工作目录
				filepath.Join(wd, "examples"), // examples子目录
				filepath.Dir(wd),              // 上级目录
				filepath.Join(filepath.Dir(wd), "examples"), // 上级目录下的examples
			}

			for _, base := range possibleBases {
				candidatePath := filepath.Join(base, viewPath, templateName)
				config.Debugf("  检查候选路径: %s", candidatePath)

				if stat, err := os.Stat(candidatePath); err == nil && !stat.IsDir() {
					config.Debugf("✅ 找到模板文件: %s", candidatePath)
					return candidatePath, nil
				}
			}
		} else {
			// 绝对路径直接检查
			templatePath = filepath.Join(viewPath, templateName)
			config.Debugf("  检查绝对路径: %s", templatePath)

			if stat, err := os.Stat(templatePath); err == nil && !stat.IsDir() {
				config.Debugf("✅ 找到模板文件: %s", templatePath)
				return templatePath, nil
			}
		}
	}

	// 如果在配置的路径中没找到，尝试递归搜索
	config.Debugf("🔍 在配置路径中未找到，开始递归搜索...")
	for _, viewPath := range e.viewPaths {
		// 对每个可能的基准目录进行递归搜索
		possibleBases := []string{
			wd,
			filepath.Join(wd, "examples"),
			filepath.Dir(wd),
			filepath.Join(filepath.Dir(wd), "examples"),
		}

		for _, base := range possibleBases {
			if !filepath.IsAbs(viewPath) {
				searchPath := filepath.Join(base, viewPath)
				if foundPath := e.recursiveSearchTemplate(searchPath, templateName); foundPath != "" {
					config.Debugf("✅ 递归搜索找到模板文件: %s", foundPath)
					return foundPath, nil
				}
			} else {
				if foundPath := e.recursiveSearchTemplate(viewPath, templateName); foundPath != "" {
					config.Debugf("✅ 递归搜索找到模板文件: %s", foundPath)
					return foundPath, nil
				}
			}
		}
	}

	// 提供更详细的错误信息
	searchedPaths := make([]string, 0)
	for _, viewPath := range e.viewPaths {
		if !filepath.IsAbs(viewPath) {
			searchedPaths = append(searchedPaths,
				filepath.Join(wd, viewPath),
				filepath.Join(wd, "examples", viewPath),
				filepath.Join(filepath.Dir(wd), viewPath),
				filepath.Join(filepath.Dir(wd), "examples", viewPath),
			)
		} else {
			searchedPaths = append(searchedPaths, viewPath)
		}
	}

	return "", fmt.Errorf("template file '%s' not found in searched paths: %v\nCurrent working directory: %s",
		templateName, searchedPaths, wd)
}

// recursiveSearchTemplate 递归搜索模板文件
func (e *TemplateEngine) recursiveSearchTemplate(basePath, templateName string) string {
	var foundPath string

	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略访问错误，继续搜索
		}

		if info.IsDir() {
			return nil // 继续遍历目录
		}

		// 检查文件名是否匹配
		if filepath.Base(path) == templateName {
			foundPath = path
			return fmt.Errorf("found") // 停止遍历
		}

		return nil
	})

	return foundPath
}

// prepareRenderData 准备渲染数据
func (e *TemplateEngine) prepareRenderData(data any) *RenderData {
	renderData := &RenderData{
		Data:  data,
		Theme: e.currentTheme,
	}

	// 如果数据已经是RenderData类型，直接使用
	if rd, ok := data.(*RenderData); ok {
		renderData = rd
		if renderData.Theme == "" {
			renderData.Theme = e.currentTheme
		}
		// 确保CSRF token已设置，如果为空则从模板引擎获取
		if renderData.CSRF == "" {
			renderData.CSRF = e.getCSRFToken()
		}
		// 同时设置CsrfToken字段，确保与CSRF字段同步
		renderData.CsrfToken = renderData.CSRF
	} else {
		// 为新的RenderData设置默认的CSRF token
		csrfToken := e.getCSRFToken()
		renderData.CSRF = csrfToken
		renderData.CsrfToken = csrfToken // 同时设置两个字段
	}

	return renderData
}

// compressHTML 压缩HTML
func (e *TemplateEngine) compressHTML(html string) string {
	// 简单的HTML压缩：移除多余的空白字符
	lines := strings.Split(html, "\n")
	compressed := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			compressed = append(compressed, trimmed)
		}
	}

	return strings.Join(compressed, "\n")
}

// ============= 模板函数实现 =============

// includeTemplate 包含其他模板
func (e *TemplateEngine) includeTemplate(templateName string, data any) template.HTML {
	content, err := e.Render(templateName, data)
	if err != nil {
		config.Errorf("Include template error: %v", err)
		return template.HTML(fmt.Sprintf("<!-- Include error: %s -->", err.Error()))
	}
	return template.HTML(content)
}

// renderComponent 渲染组件
func (e *TemplateEngine) renderComponent(componentName string, data any) template.HTML {
	content, err := e.RenderComponent(componentName, data)
	if err != nil {
		config.Errorf("Component render error: %v", err)
		return template.HTML(fmt.Sprintf("<!-- Component error: %s -->", err.Error()))
	}
	return template.HTML(content)
}

// getThemeVariable 获取主题变量
func (e *TemplateEngine) getThemeVariable(key string) string {
	if theme, exists := e.themes[e.currentTheme]; exists {
		if value, exists := theme.Variables[key]; exists {
			return value
		}
	}
	return ""
}

// getAssetURL 获取资源URL
func (e *TemplateEngine) getAssetURL(path string) string {
	if theme, exists := e.themes[e.currentTheme]; exists {
		return fmt.Sprintf("/%s/%s", theme.StaticPath, strings.TrimPrefix(path, "/"))
	}
	return fmt.Sprintf("/static/%s", strings.TrimPrefix(path, "/"))
}

// buildURL 构建URL
func (e *TemplateEngine) buildURL(path string, params ...any) string {
	url := strings.TrimPrefix(path, "/")

	// 如果有参数，构建查询字符串
	if len(params) > 0 {
		query := make([]string, 0, len(params)/2)
		for i := 0; i < len(params)-1; i += 2 {
			key := fmt.Sprintf("%v", params[i])
			value := fmt.Sprintf("%v", params[i+1])
			query = append(query, fmt.Sprintf("%s=%s", key, value))
		}
		if len(query) > 0 {
			url += "?" + strings.Join(query, "&")
		}
	}

	return "/" + url
}

// getCSRFToken 获取CSRF令牌
func (e *TemplateEngine) getCSRFToken() string {
	// 尝试通过CSRF提供者获取token
	provider := GetCSRFTokenProvider()
	if provider != nil {
		return provider.GenerateSimpleToken()
	}

	// 如果没有提供者，返回占位符
	return "csrf-token-placeholder"
}

// getFlashMessage 获取Flash消息
func (e *TemplateEngine) getFlashMessage(msgType string) []string {
	// 这里应该从会话中获取Flash消息
	// 暂时返回空切片
	return []string{}
}

// truncateString 截断字符串
func (e *TemplateEngine) truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length] + "..."
}

// renderMarkdown 渲染Markdown（简化版）
func (e *TemplateEngine) renderMarkdown(content string) template.HTML {
	// 这里应该使用真正的Markdown渲染器
	// 暂时进行简单的转换
	html := strings.ReplaceAll(content, "\n", "<br>")
	html = strings.ReplaceAll(html, "**", "<strong>")
	html = strings.ReplaceAll(html, "**", "</strong>")
	return template.HTML(html)
}

// toJSON 转换为JSON
func (e *TemplateEngine) toJSON(data any) template.JS {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return template.JS("{}")
	}
	return template.JS(jsonData)
}

// safeHTML 安全HTML
func (e *TemplateEngine) safeHTML(content string) template.HTML {
	return template.HTML(content)
}

// createDict 创建字典
func (e *TemplateEngine) createDict(values ...any) map[string]any {
	dict := make(map[string]any)
	for i := 0; i < len(values)-1; i += 2 {
		key := fmt.Sprintf("%v", values[i])
		dict[key] = values[i+1]
	}
	return dict
}

// createSlice 创建切片
func (e *TemplateEngine) createSlice(values ...any) []any {
	return values
}

// createRange 创建数字范围
func (e *TemplateEngine) createRange(start, end int) []int {
	if start > end {
		return []int{}
	}

	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

// formatDate 格式化日期
func (e *TemplateEngine) formatDate(date any, format string) string {
	var t time.Time

	switch v := date.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v != nil {
			t = *v
		} else {
			return ""
		}
	case int64:
		t = time.Unix(v, 0)
	case string:
		// 尝试解析字符串日期
		if parsed, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
			t = parsed
		} else {
			return v
		}
	default:
		return fmt.Sprintf("%v", date)
	}

	// 格式化映射
	switch format {
	case "date":
		return t.Format("2006-01-02")
	case "datetime":
		return t.Format("2006-01-02 15:04:05")
	case "time":
		return t.Format("15:04:05")
	case "iso":
		return t.Format(time.RFC3339)
	case "rfc":
		return t.Format(time.RFC822)
	default:
		return t.Format(format)
	}
}

// formatCurrency 格式化货币
func (e *TemplateEngine) formatCurrency(amount any, currency string) string {
	var value float64

	switch v := amount.(type) {
	case float64:
		value = v
	case float32:
		value = float64(v)
	case int:
		value = float64(v)
	case int64:
		value = float64(v)
	default:
		return fmt.Sprintf("%v", amount)
	}

	switch currency {
	case "CNY", "RMB", "¥":
		return fmt.Sprintf("¥%.2f", value)
	case "USD", "$":
		return fmt.Sprintf("$%.2f", value)
	case "EUR", "€":
		return fmt.Sprintf("€%.2f", value)
	default:
		return fmt.Sprintf("%.2f %s", value, currency)
	}
}

// formatFileSize 格式化文件大小
func (e *TemplateEngine) formatFileSize(size any) string {
	var bytes int64

	switch v := size.(type) {
	case int64:
		bytes = v
	case int:
		bytes = int64(v)
	case float64:
		bytes = int64(v)
	default:
		return fmt.Sprintf("%v", size)
	}

	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// ============= 增强的缓存统计和监控 =============

// CacheStats 详细缓存统计信息
type CacheStats struct {
	TemplateCount  int            `json:"template_count"`
	LayoutCount    int            `json:"layout_count"`
	ComponentCount int            `json:"component_count"`
	TotalCount     int            `json:"total_count"`
	MemoryEstimate int            `json:"memory_estimate_bytes"`
	CacheEnabled   bool           `json:"cache_enabled"`
	CacheHitRate   float64        `json:"cache_hit_rate"`
	CacheKeys      []string       `json:"cache_keys,omitempty"`
	LayoutKeys     []string       `json:"layout_keys,omitempty"`
	RecentAccess   []string       `json:"recent_access,omitempty"`
	KeyTypes       map[string]int `json:"key_types"`
}

// GetEnhancedCacheStats 获取增强的缓存统计信息
func (e *TemplateEngine) GetEnhancedCacheStats() *CacheStats {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	stats := &CacheStats{
		TemplateCount:  len(e.templates),
		LayoutCount:    len(e.layouts),
		ComponentCount: len(e.components),
		TotalCount:     len(e.templates) + len(e.layouts) + len(e.components),
		CacheEnabled:   e.enableCache,
		KeyTypes:       make(map[string]int),
		CacheKeys:      make([]string, 0, len(e.templates)),
		LayoutKeys:     make([]string, 0),
	}

	// 分析缓存键类型
	layoutKeyCount := 0
	for key := range e.templates {
		stats.CacheKeys = append(stats.CacheKeys, key)
		if DefaultCacheKeyManager.IsLayoutKey(key) {
			layoutKeyCount++
			stats.LayoutKeys = append(stats.LayoutKeys, key)
		}
	}

	stats.KeyTypes["template_keys"] = len(e.templates) - layoutKeyCount
	stats.KeyTypes["layout_keys"] = layoutKeyCount

	// 估算内存使用
	var totalMemory int
	for key, tmpl := range e.templates {
		if tmpl != nil {
			// 基础估算: 缓存键长度 + 模板对象开销
			totalMemory += len(key)*2 + 2048 // 每个模板约2KB开销
		}
	}
	for key := range e.layouts {
		totalMemory += len(key)*2 + 1024 // 每个布局约1KB开销
	}
	for key := range e.components {
		totalMemory += len(key)*2 + 512 // 每个组件约512B开销
	}

	stats.MemoryEstimate = totalMemory

	// 计算缓存命中率（简化版，实际应该维护独立的命中统计）
	if stats.TotalCount > 0 {
		// 假设有缓存的模板命中率较高
		stats.CacheHitRate = 0.85 // 示例值，实际实现中应该动态计算
	}

	return stats
}

// AnalyzeCachePerformance 分析缓存性能
func (e *TemplateEngine) AnalyzeCachePerformance() map[string]any {
	stats := e.GetEnhancedCacheStats()

	analysis := map[string]any{
		"basic_stats": stats,
		"performance_metrics": map[string]any{
			"cache_efficiency": "good", // 根据实际统计确定
			"memory_usage":     "acceptable",
			"recommendations":  []string{},
		},
	}

	// 性能分析建议
	recommendations := []string{}

	if stats.TotalCount > 1000 {
		recommendations = append(recommendations, "Consider implementing cache size limits")
	}

	if stats.MemoryEstimate > 50*1024*1024 { // > 50MB
		recommendations = append(recommendations, "High memory usage detected, consider cache cleanup")
	}

	if float64(stats.KeyTypes["layout_keys"])/float64(stats.TotalCount) > 0.7 {
		recommendations = append(recommendations, "High ratio of layout templates, consider layout optimization")
	}

	analysis["performance_metrics"].(map[string]any)["recommendations"] = recommendations

	return analysis
}

// ============= 模板预加载机制 =============

// TemplatePreloader 模板预加载器
type TemplatePreloader struct {
	engine      *TemplateEngine
	preloadList []string          // 预加载模板列表
	layoutPairs map[string]string // 模板-布局配对
	enabled     bool
	mu          sync.RWMutex
}

// NewTemplatePreloader 创建模板预加载器
func NewTemplatePreloader(engine *TemplateEngine) *TemplatePreloader {
	return &TemplatePreloader{
		engine:      engine,
		preloadList: []string{},
		layoutPairs: make(map[string]string),
		enabled:     true,
	}
}

// AddPreloadTemplate 添加预加载模板
func (tp *TemplatePreloader) AddPreloadTemplate(templateName string) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// 避免重复添加
	for _, existing := range tp.preloadList {
		if existing == templateName {
			return
		}
	}

	tp.preloadList = append(tp.preloadList, templateName)
	config.Debugf("Added template to preload list: %s", templateName)
}

// AddPreloadTemplateWithLayout 添加带布局的预加载模板
func (tp *TemplatePreloader) AddPreloadTemplateWithLayout(templateName, layoutName string) {
	tp.AddPreloadTemplate(templateName)

	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.layoutPairs[templateName] = layoutName
	config.Debugf("Added template-layout pair: %s@%s", templateName, layoutName)
}

// PreloadAll 预加载所有模板
func (tp *TemplatePreloader) PreloadAll() error {
	if !tp.enabled {
		config.Debugf("Template preloader is disabled")
		return nil
	}

	tp.mu.RLock()
	preloadList := make([]string, len(tp.preloadList))
	copy(preloadList, tp.preloadList)
	layoutPairs := make(map[string]string)
	for k, v := range tp.layoutPairs {
		layoutPairs[k] = v
	}
	tp.mu.RUnlock()

	if len(preloadList) == 0 {
		config.Debugf("No templates to preload")
		return nil
	}

	config.Infof("Starting template preloading: %d templates", len(preloadList))
	startTime := time.Now()

	successCount := 0
	errorCount := 0

	for _, templateName := range preloadList {
		// 检查是否需要预加载布局组合
		if layoutName, hasLayout := layoutPairs[templateName]; hasLayout {
			err := tp.preloadTemplateWithLayout(templateName, layoutName)
			if err != nil {
				config.Warnf("Failed to preload template with layout %s@%s: %v", templateName, layoutName, err)
				errorCount++
			} else {
				successCount++
			}
		} else {
			err := tp.preloadSingleTemplate(templateName)
			if err != nil {
				config.Warnf("Failed to preload template %s: %v", templateName, err)
				errorCount++
			} else {
				successCount++
			}
		}
	}

	elapsed := time.Since(startTime)
	config.Infof("Template preloading completed: %d success, %d failed, took %v",
		successCount, errorCount, elapsed)

	if errorCount > 0 {
		return fmt.Errorf("failed to preload %d out of %d templates", errorCount, len(preloadList))
	}

	return nil
}

// preloadSingleTemplate 预加载单个模板
func (tp *TemplatePreloader) preloadSingleTemplate(templateName string) error {
	// 检查是否已缓存
	cacheKey := DefaultCacheKeyManager.GenerateTemplateKey(templateName)
	if tp.engine.enableCache {
		tp.engine.templateMutex.RLock()
		if _, exists := tp.engine.templates[cacheKey]; exists {
			tp.engine.templateMutex.RUnlock()
			config.Debugf("Template %s already cached, skipping preload", templateName)
			return nil
		}
		tp.engine.templateMutex.RUnlock()
	}

	// 预加载模板
	_, err := tp.engine.GetTemplate(templateName)
	if err != nil {
		return fmt.Errorf("failed to preload template %s: %w", templateName, err)
	}

	config.Debugf("Successfully preloaded template: %s", templateName)
	return nil
}

// preloadTemplateWithLayout 预加载带布局的模板
func (tp *TemplatePreloader) preloadTemplateWithLayout(templateName, layoutName string) error {
	// 检查是否已缓存
	cacheKey := DefaultCacheKeyManager.GenerateLayoutKey(templateName, layoutName)
	if tp.engine.enableCache {
		tp.engine.templateMutex.RLock()
		if _, exists := tp.engine.templates[cacheKey]; exists {
			tp.engine.templateMutex.RUnlock()
			config.Debugf("Template with layout %s@%s already cached, skipping preload", templateName, layoutName)
			return nil
		}
		tp.engine.templateMutex.RUnlock()
	}

	// 预加载模板
	_, err := tp.engine.GetTemplateWithLayout(templateName, layoutName)
	if err != nil {
		return fmt.Errorf("failed to preload template with layout %s@%s: %w", templateName, layoutName, err)
	}

	config.Debugf("Successfully preloaded template with layout: %s@%s", templateName, layoutName)
	return nil
}

// GetPreloadStats 获取预加载统计
func (tp *TemplatePreloader) GetPreloadStats() map[string]any {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	return map[string]any{
		"enabled":            tp.enabled,
		"preload_list_count": len(tp.preloadList),
		"layout_pairs_count": len(tp.layoutPairs),
		"preload_list":       tp.preloadList,
		"layout_pairs":       tp.layoutPairs,
	}
}

// SetEnabled 设置预加载器启用状态
func (tp *TemplatePreloader) SetEnabled(enabled bool) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.enabled = enabled
	config.Debugf("Template preloader enabled: %v", enabled)
}

// Clear 清空预加载列表
func (tp *TemplatePreloader) Clear() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.preloadList = []string{}
	tp.layoutPairs = make(map[string]string)
	config.Debugf("Template preloader list cleared")
}

// AddCommonTemplates 添加常用模板到预加载列表
func (tp *TemplatePreloader) AddCommonTemplates() {
	commonTemplates := []string{
		"index.html",
		"error.html",
		"404.html",
		"500.html",
		"login.html",
	}

	commonLayoutPairs := map[string]string{
		"index.html": "app.html",
		"error.html": "app.html",
		"login.html": "auth.html",
	}

	for _, templateName := range commonTemplates {
		tp.AddPreloadTemplate(templateName)
	}

	tp.mu.Lock()
	for template, layout := range commonLayoutPairs {
		tp.layoutPairs[template] = layout
	}
	tp.mu.Unlock()

	config.Debugf("Added %d common templates to preload list", len(commonTemplates))
}

// 全局预加载器实例
var (
	globalPreloader *TemplatePreloader
	preloaderOnce   sync.Once
)

// GetGlobalTemplatePreloader 获取全局模板预加载器
func GetGlobalTemplatePreloader() *TemplatePreloader {
	preloaderOnce.Do(func() {
		defaultEngine := GetDefaultEngine()
		globalPreloader = NewTemplatePreloader(defaultEngine)
	})
	return globalPreloader
}

// PreloadCommonTemplates 预加载常用模板（便捷函数）
func PreloadCommonTemplates() error {
	preloader := GetGlobalTemplatePreloader()
	preloader.AddCommonTemplates()
	return preloader.PreloadAll()
}

// AddPreloadTemplate 添加预加载模板（便捷函数）
func AddPreloadTemplate(templateName string) {
	GetGlobalTemplatePreloader().AddPreloadTemplate(templateName)
}

// AddPreloadTemplateWithLayout 添加带布局的预加载模板（便捷函数）
func AddPreloadTemplateWithLayout(templateName, layoutName string) {
	GetGlobalTemplatePreloader().AddPreloadTemplateWithLayout(templateName, layoutName)
}

// GetTemplateList 获取模板列表
func (e *TemplateEngine) GetTemplateList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	templates := make([]string, 0, len(e.templates))
	for name := range e.templates {
		templates = append(templates, name)
	}
	return templates
}

// GetLayoutList 获取布局列表
func (e *TemplateEngine) GetLayoutList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	layouts := make([]string, 0, len(e.layouts))
	for name := range e.layouts {
		layouts = append(layouts, name)
	}
	return layouts
}

// GetComponentList 获取组件列表
func (e *TemplateEngine) GetComponentList() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	components := make([]string, 0, len(e.components))
	for name := range e.components {
		components = append(components, name)
	}
	return components
}

// ClearCache 清除模板缓存
func (e *TemplateEngine) ClearCache() {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	oldCount := len(e.templates) + len(e.layouts) + len(e.components)

	e.templates = make(map[string]*template.Template)
	e.layouts = make(map[string]*template.Template)
	e.components = make(map[string]*template.Template)

	config.Infof("Template cache cleared: removed %d cached items", oldCount)
}

// ClearTemplateCache 清除指定类型的模板缓存
func (e *TemplateEngine) ClearTemplateCache(cacheType string) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	var clearedCount int

	switch cacheType {
	case "templates":
		clearedCount = len(e.templates)
		e.templates = make(map[string]*template.Template)
	case "layouts":
		clearedCount = len(e.layouts)
		e.layouts = make(map[string]*template.Template)
	case "components":
		clearedCount = len(e.components)
		e.components = make(map[string]*template.Template)
	default:
		config.Warnf("Unknown cache type: %s", cacheType)
		return
	}

	config.Infof("%s cache cleared: removed %d items", cacheType, clearedCount)
}

// ClearTemplateByPattern 根据模式清除模板缓存
func (e *TemplateEngine) ClearTemplateByPattern(pattern string) int {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	clearedCount := 0

	// 清理匹配的模板缓存
	for key := range e.templates {
		if strings.Contains(key, pattern) {
			delete(e.templates, key)
			clearedCount++
		}
	}

	if clearedCount > 0 {
		config.Infof("Cleared %d template cache entries matching pattern: %s", clearedCount, pattern)
	}

	return clearedCount
}

// OptimizeCacheWithLimit 优化缓存，移除不再需要的模板（带限制）
func (e *TemplateEngine) OptimizeCacheWithLimit(maxEntries int) int {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	currentCount := len(e.templates)
	if currentCount <= maxEntries {
		return 0 // 不需要优化
	}

	// 简单的LRU策略：保留最近的maxEntries个条目
	// 这里使用一个简化的实现，实际中可以使用更复杂的LRU算法

	// 将所有key收集到slice中（Go的map遍历是随机的，这里模拟LRU）
	keys := make([]string, 0, len(e.templates))
	for key := range e.templates {
		keys = append(keys, key)
	}

	// 保留前maxEntries个，删除其余的
	removedCount := 0
	if len(keys) > maxEntries {
		toRemove := keys[maxEntries:]
		for _, key := range toRemove {
			delete(e.templates, key)
			removedCount++
		}
	}

	if removedCount > 0 {
		config.Infof("Cache optimized: removed %d entries, %d remaining", removedCount, len(e.templates))
	}

	return removedCount
}

// PrepareRenderData 公开的渲染数据准备方法（用于测试和外部调用）
func (e *TemplateEngine) PrepareRenderData(data any) *RenderData {
	return e.prepareRenderData(data)
}

// GetStats 获取模板引擎统计信息
func (e *TemplateEngine) GetStats() map[string]any {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	return map[string]any{
		"templates_count":  len(e.templates),
		"layouts_count":    len(e.layouts),
		"components_count": len(e.components),
		"current_theme":    e.currentTheme,
		"available_themes": e.GetAvailableThemes(),
		"cache_enabled":    e.enableCache,
		"reload_enabled":   e.enableReload,
		"compress_enabled": e.enableCompress,
		"view_paths":       e.viewPaths,
		"layout_path":      e.layoutPath,
		"component_path":   e.componentPath,
	}
}
