package view

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

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

	// // 检查关键函数是否存在
	// criticalFuncs := []string{"formatFileSize", "dateformat", "now"}
	// for _, funcName := range criticalFuncs {
	// 	if _, exists := e.funcMap[funcName]; exists {
	// 		config.Debugf("✅ Critical function '%s' is available", funcName)
	// 	} else {
	// 		config.Warnf("⚠️ Critical function '%s' is missing", funcName)
	// 	}
	// }
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

	// 🔧 核心修复：收集同目录下的所有相关模板文件一起解析，支持{{template}}引用
	relatedFiles := e.collectRelatedTemplateFiles(templatePath, "")
	tmpl, err := template.New("").
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs).
		ParseFiles(relatedFiles...)

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

// FindTemplateFile 查找模板文件（增强版）
func (e *TemplateEngine) FindTemplateFile(templateName string) (string, error) {
	// 确保模板名有扩展名
	if !strings.HasSuffix(templateName, e.extension) {
		templateName += e.extension
	}

	config.Debugf("🔍 查找模板文件: %s，在路径: %v", templateName, e.viewPaths)

	// 获取当前工作目录用于调试
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	config.Debugf("  [FindTemplateFile]当前工作目录: %s", wd)

	// 在所有视图路径中搜索
	for _, viewPath := range e.viewPaths {
		// 先尝试绝对路径
		templatePath := viewPath
		if !filepath.IsAbs(viewPath) {
			// 如果是相对路径，尝试不同的基准目录
			possibleBases := []string{
				wd,                             // 当前工作目录
				filepath.Join(wd, "templates"), // templates子目录
				filepath.Dir(wd),               // 上级目录
				filepath.Join(filepath.Dir(wd), "templates"), // 上级目录下的templates
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
			filepath.Join(wd, "templates"),
			filepath.Dir(wd),
			filepath.Join(filepath.Dir(wd), "templates"),
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

// RecoverFromTemplateLoadError 从模板加载错误中恢复
func (e *TemplateEngine) RecoverFromTemplateLoadError(templateName string, originalErr error) (*template.Template, error) {
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

	// 🔧 修复：收集相关模板文件一起解析，支持{{template}}引用
	relatedFiles := e.collectRelatedTemplateFiles(templatePath, "")
	_, err = tmpl.ParseFiles(relatedFiles...)
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
