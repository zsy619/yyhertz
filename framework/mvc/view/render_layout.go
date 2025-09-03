package view

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
)

// GetTemplateWithLayout 获取带布局的模板（支持布局继承）
func (e *TemplateEngine) GetTemplateWithLayout(templateName, layoutName string) (*template.Template, error) {
	// 【关键修复】标准化布局名称，去除可能的.html扩展名
	// 确保布局名称与注册时的名称一致
	normalizedLayoutName := layoutName
	if strings.HasSuffix(normalizedLayoutName, e.extension) {
		normalizedLayoutName = strings.TrimSuffix(normalizedLayoutName, e.extension)
		config.Debugf("Normalized layout name from '%s' to '%s'", layoutName, normalizedLayoutName)
	}

	normalizedTemplateName := templateName
	if strings.HasSuffix(normalizedTemplateName, e.extension) {
		normalizedTemplateName = strings.TrimSuffix(normalizedTemplateName, e.extension)
		config.Debugf("Normalized template name from '%s' to '%s'", templateName, normalizedTemplateName)
	}

	// 【关键修复】：基于Beego机制生成正确的缓存键
	// 先获取模板的基础名称，然后生成匹配的缓存键
	contentPath, err := e.FindTemplateFile(normalizedTemplateName)
	if err != nil {
		return nil, fmt.Errorf("content template '%s' not found: %w", normalizedLayoutName, err)
	}

	templateBaseName := e.getTemplateBaseName(contentPath)
	// templateBaseName := e.getTemplateBaseName(normalizedTemplateName)
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
	if len(e.layouts) > 0 {
		prefix := ""
		if parts := strings.Split(e.layoutPath, "/"); len(parts) > 0 {
			prefix = parts[len(parts)-1]
		}
		if prefix != "" && !strings.HasPrefix(normalizedLayoutName, prefix) {
			normalizedLayoutName = prefix + "/" + normalizedLayoutName
		}
		if _, exists := e.layouts[normalizedLayoutName]; !exists {
			// 获取可用布局列表（避免调用GetLayoutList导致死锁）
			availableLayouts := make([]string, 0, len(e.layouts))
			for name := range e.layouts {
				availableLayouts = append(availableLayouts, name)
			}
			return nil, fmt.Errorf("layout '%s' not found, available layouts: %v", normalizedLayoutName, availableLayouts)
		}
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
	if Len(e.layouts) > 0 {
		if _, exists := e.layouts[layoutName]; !exists {
			return nil, fmt.Errorf("layout '%s' not found", layoutName)
		}
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
	// 获取当前工作目录用于调试
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// 移除路径和扩展名，获取文件名
	baseName := strings.TrimPrefix(templatePath, wd)
	ext := filepath.Ext(baseName)
	if ext != "" {
		baseName = strings.TrimSuffix(baseName, ext)
	}
	baseName = strings.TrimPrefix(baseName, string(os.PathSeparator))
	baseName = strings.ReplaceAll(baseName, "\\", "_")
	baseName = strings.ReplaceAll(baseName, "/", "_")
	return baseName
}

// getLayoutContent 获取布局模板的内容
func (e *TemplateEngine) getLayoutContent(layoutName string) (string, error) {
	layoutFileName := layoutName
	// 确保模板名有扩展名
	layoutFileName = strings.TrimSuffix(layoutFileName, e.extension)

	// 获取当前工作目录用于调试
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	config.Debugf("  [getLayoutContent]当前工作目录: %s", wd)

	// 尝试在布局路径中搜索文件
	layoutDir := e.layoutPath
	if parts := strings.Split(e.layoutPath, "/"); len(parts) > 0 {
		prefix := parts[len(parts)-1]
		if prefix != "" {
			layoutFileName = strings.TrimPrefix(layoutName, prefix)
		}
	}
	if layoutDir != "" {
		possiblePath := ""
		baseLayoutPath := filepath.Join(layoutDir, layoutFileName)
		isAbs := filepath.IsAbs(baseLayoutPath)
		if !isAbs {
			possiblePath = filepath.Join(wd, baseLayoutPath+e.extension)
		} else {
			possiblePath = baseLayoutPath + e.extension
		}
		if _, err := os.Stat(possiblePath); err == nil {
			contentBytes, err := os.ReadFile(possiblePath)
			if err != nil {
				return "", fmt.Errorf("failed to read layout file '%s': %w", baseLayoutPath, err)
			}
			return e.processLayoutContent(string(contentBytes)), nil
		}

		// 尝试其他扩展名
		for _, ext := range []string{".tpl", ".gohtml", ".tmpl"} {
			if !isAbs {
				possiblePath = filepath.Join(wd, baseLayoutPath+ext)
			} else {
				possiblePath = baseLayoutPath + ext
			}
			if _, err := os.Stat(possiblePath); err == nil {
				contentBytes, err := os.ReadFile(possiblePath)
				if err != nil {
					return "", fmt.Errorf("failed to read layout file '%s': %w", possiblePath, err)
				}
				return e.processLayoutContent(string(contentBytes)), nil
			}
		}
	}

	layoutsDirs := []string{
		filepath.Join("layout"),
		filepath.Join("layouts"),
	}
	// 如果layoutPath为空或未找到，尝试在所有视图路径中搜索
	for _, viewDir := range e.viewPaths {
		// 检查 layouts 子目录
		for _, dir := range layoutsDirs {
			possiblePath := ""
			layoutsDir := filepath.Join(viewDir, dir)
			isAbs := filepath.IsAbs(layoutsDir)
			if !isAbs {
				possiblePath = filepath.Join(wd, layoutsDir, layoutFileName+e.extension)
			} else {
				possiblePath = filepath.Join(layoutsDir, layoutFileName+e.extension)
			}
			if _, err := os.Stat(possiblePath); err == nil {
				contentBytes, err := os.ReadFile(possiblePath)
				if err != nil {
					return "", fmt.Errorf("failed to read layout file '%s': %w", possiblePath, err)
				}
				return e.processLayoutContent(string(contentBytes)), nil
			}

			// 尝试其他扩展名
			for _, ext := range []string{".tpl", ".gohtml", ".tmpl"} {
				if !isAbs {
					possiblePath = filepath.Join(wd, layoutsDir, layoutFileName+ext)
				} else {
					possiblePath = filepath.Join(layoutsDir, layoutFileName+ext)
				}
				if _, err := os.Stat(possiblePath); err == nil {
					contentBytes, err := os.ReadFile(possiblePath)
					if err != nil {
						return "", fmt.Errorf("failed to read layout file '%s': %w", possiblePath, err)
					}
					return e.processLayoutContent(string(contentBytes)), nil
				}
			}
		}
	}

	return "", fmt.Errorf("layout file for '%s' not found in layout path '%s' or any view path", layoutName, e.layoutPath)
}

// processLayoutContent 处理布局内容，支持 {{.LayoutContent}} 占位符
func (e *TemplateEngine) processLayoutContent(layoutContent string) string {
	// 支持多种布局内容占位符格式，兼容不同的模板引擎风格
	contentPlaceholders := []string{
		fmt.Sprintf("%s.LayoutContent%s", e.delimLeft, e.delimRight),           // 主要格式
		fmt.Sprintf("%s .LayoutContent %s ", e.delimLeft, e.delimRight),        // 带空格格式
		fmt.Sprintf("%s.Content%s", e.delimLeft, e.delimRight),                 // 通用格式
		fmt.Sprintf("%s .Content %s ", e.delimLeft, e.delimRight),              // 带空格通用格式
		fmt.Sprintf("%stemplate \"content\" .%s", e.delimLeft, e.delimRight),   // Beego 风格
		fmt.Sprintf("%s template \"content\" .%s ", e.delimLeft, e.delimRight), // Beego 风格
		fmt.Sprintf("%syield%s", e.delimLeft, e.delimRight),                    // Rails 风格
		fmt.Sprintf("%s yield%s ", e.delimLeft, e.delimRight),                  // Rails 风格
		fmt.Sprintf("%sblock \"content\" .%s", e.delimLeft, e.delimRight),      // Go 标准模板 block 格式
		fmt.Sprintf("%s block \"content\" .%s ", e.delimLeft, e.delimRight),    // Go 标准模板 block 格式
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
		fmt.Sprintf("%s.LayoutContent%s", e.delimLeft, e.delimRight),           // 主要格式
		fmt.Sprintf("%s .LayoutContent %s ", e.delimLeft, e.delimRight),        // 带空格格式
		fmt.Sprintf("%s.Content%s", e.delimLeft, e.delimRight),                 // 通用格式
		fmt.Sprintf("%s .Content %s ", e.delimLeft, e.delimRight),              // 带空格通用格式
		fmt.Sprintf("%stemplate \"content\" .%s", e.delimLeft, e.delimRight),   // Beego 风格
		fmt.Sprintf("%s template \"content\" .%s ", e.delimLeft, e.delimRight), // Beego 风格
		fmt.Sprintf("%syield%s", e.delimLeft, e.delimRight),                    // Rails 风格
		fmt.Sprintf("%s yield%s ", e.delimLeft, e.delimRight),                  // Rails 风格
		fmt.Sprintf("%sblock \"content\" .%s", e.delimLeft, e.delimRight),      // Go 标准模板 block 格式
		fmt.Sprintf("%s block \"content\" .%s ", e.delimLeft, e.delimRight),    // Go 标准模板 block 格式
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
	if len(layoutSections) == 0 {
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
			fmt.Sprintf("%s.%s%s", e.delimLeft, sectionName, e.delimRight),                                // {{.HtmlHead}}
			fmt.Sprintf("%s .%s %s", e.delimLeft, sectionName, e.delimRight),                              // {{ .HtmlHead }}
			fmt.Sprintf("%s.%s | safehtml%s", e.delimLeft, sectionName, e.delimRight),                     // {{.HtmlHead | safehtml}}
			fmt.Sprintf("%stemplate \"%s\" .%s", e.delimLeft, strings.ToLower(sectionName), e.delimRight), // {{template "htmlhead" .}}
			fmt.Sprintf("%sblock \"%s\" .%s", e.delimLeft, strings.ToLower(sectionName), e.delimRight),    // {{block "htmlhead" .}}{{end}}
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
			fmt.Sprintf("%s.%s%s", e.delimLeft, sectionName, e.delimRight),
			fmt.Sprintf("%s .%s %s", e.delimLeft, sectionName, e.delimRight),
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
