package view

import (
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 模板解析和加载相关方法 =============

// loadAllTemplates 加载所有模板
func (e *TemplateEngine) loadAllTemplates() error {
	// 加载布局
	if err := e.loadLayouts(); err != nil {
		return fmt.Errorf("failed to load layouts: %w", err)
	}

	// 加载组件
	if err := e.loadComponents(); err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}

	// 加载视图模板
	if err := e.loadViewTemplates(); err != nil {
		return fmt.Errorf("failed to load view templates: %w", err)
	}

	config.Infof("Loaded %d templates, %d layouts, %d components",
		len(e.templates), len(e.layouts), len(e.components))

	return nil
}

// loadLayouts 加载布局模板
func (e *TemplateEngine) loadLayouts() error {
	if e.layoutPath == "" {
		return nil
	}

	return filepath.WalkDir(e.layoutPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续
		}

		if !d.IsDir() && strings.HasSuffix(path, e.extension) {
			layoutName := e.getTemplateName(path)

			// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
			manager := GetGlobalFunctionManager()
			mergedFuncs := manager.GetMergedFunctions(e.funcMap)

			tmpl := template.New(layoutName).
				Delims(e.delimLeft, e.delimRight).
				Funcs(mergedFuncs)

			if _, err := tmpl.ParseFiles(path); err != nil {
				config.Errorf("Failed to parse layout %s: %v", path, err)
				return nil
			}

			e.layouts[layoutName] = tmpl
			config.Debugf("Loaded layout: %s", layoutName)
		}

		return nil
	})
}

// loadComponents 加载组件模板
func (e *TemplateEngine) loadComponents() error {
	if e.componentPath == "" {
		return nil
	}

	return filepath.WalkDir(e.componentPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 忽略错误，继续
		}

		if !d.IsDir() && strings.HasSuffix(path, e.extension) {
			componentName := e.getTemplateName(path)

			// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
			manager := GetGlobalFunctionManager()
			mergedFuncs := manager.GetMergedFunctions(e.funcMap)

			tmpl := template.New(componentName).
				Delims(e.delimLeft, e.delimRight).
				Funcs(mergedFuncs)

			if _, err := tmpl.ParseFiles(path); err != nil {
				config.Errorf("Failed to parse component %s: %v", path, err)
				return nil
			}

			e.components[componentName] = tmpl
			config.Debugf("Loaded component: %s", componentName)
		}

		return nil
	})
}

// loadViewTemplates 加载视图模板
func (e *TemplateEngine) loadViewTemplates() error {
	for _, viewPath := range e.viewPaths {
		err := filepath.WalkDir(viewPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // 忽略错误，继续
			}

			if !d.IsDir() && strings.HasSuffix(path, e.extension) {
				// 跳过布局和组件目录
				if strings.Contains(path, e.layoutPath) || strings.Contains(path, e.componentPath) {
					return nil
				}

				templateName := e.getTemplateName(path)

				// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
				manager := GetGlobalFunctionManager()
				mergedFuncs := manager.GetMergedFunctions(e.funcMap)

				tmpl := template.New(templateName).
					Delims(e.delimLeft, e.delimRight).
					Funcs(mergedFuncs)

				parsedTmpl, err := tmpl.ParseFiles(path)
				if err != nil {
					config.Errorf("Failed to parse template %s: %v", path, err)
					return nil
				}

				// 查找实际的模板（类似loadTemplate中的逻辑）
				templates := parsedTmpl.Templates()
				var actualTemplate *template.Template

				for _, t := range templates {
					if t.Tree != nil && t.Tree.Root != nil {
						actualTemplate = t
						break
					}
				}

				if actualTemplate != nil {
					e.templates[templateName] = actualTemplate
					config.Debugf("Loaded template: %s -> %s", templateName, actualTemplate.Name())
				} else {
					config.Warnf("Template %s is empty or invalid", templateName)
				}
			}

			return nil
		})

		if err != nil {
			config.Warnf("Error walking view path %s: %v", viewPath, err)
		}
	}

	return nil
}

// getTemplateName 从文件路径获取模板名称
func (e *TemplateEngine) getTemplateName(filePath string) string {
	// 移除扩展名
	name := strings.TrimSuffix(filepath.Base(filePath), e.extension)

	// 如果包含目录，保留相对路径
	for _, viewPath := range e.viewPaths {
		if strings.HasPrefix(filePath, viewPath) {
			relPath, _ := filepath.Rel(viewPath, filePath)
			name = strings.TrimSuffix(relPath, e.extension)
			break
		}
	}

	return strings.ReplaceAll(name, "\\", "/") // 标准化路径分隔符
}

// loadSingleTemplate 加载单个模板（内部方法）
func (e *TemplateEngine) loadSingleTemplate(templateName string) (*template.Template, error) {
	// 查找模板文件
	templateFile, err := e.FindTemplateFile(templateName)
	if err != nil {
		return nil, fmt.Errorf("template file not found: %s", templateName)
	}

	// 动态获取最新的合并函数
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 创建和解析模板
	tmpl := template.New(templateName).
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs)

	_, err = tmpl.ParseFiles(templateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
	}

	return tmpl, nil
}

// CreateInlineTemplate 创建内联模板（用于测试）
func (e *TemplateEngine) CreateInlineTemplate(name, content string) (*template.Template, error) {
	// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	tmpl := template.New(name).
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs)

	return tmpl.Parse(content)
}

// createInlineTemplate 创建内联模板（用于测试）- 内部方法兼容性
func (e *TemplateEngine) createInlineTemplate(name, content string) (*template.Template, error) {
	return e.CreateInlineTemplate(name, content)
}

// ============= Beego 兼容方法 (从 BeegoTemplateEngine 合并) =============

// initBeegoFunctions 初始化Beego风格的模板函数
func (e *TemplateEngine) initBeegoFunctions() {
	// 获取现有的Beego函数
	beegoFuncs := GetBeegoTemplateFuncs()

	// 复制到当前引擎 - 确保Beego函数不被覆盖
	for name, fn := range beegoFuncs {
		e.funcMap[name] = fn
		config.Debugf("Registered Beego template function: %s", name)
	}

	// 特别确保关键函数的注册
	if dateformatFunc, exists := beegoFuncs["dateformat"]; exists {
		e.funcMap["dateformat"] = dateformatFunc
		config.Infof("✅ Critical function 'dateformat' registered successfully")
	} else {
		config.Errorf("❌ Critical function 'dateformat' not found in Beego functions")
	}

	// 添加Beego特有的额外函数
	e.funcMap["yield"] = func() string {
		return "{{.LayoutContent}}"
	}

	e.funcMap["content"] = func() string {
		return "{{.Content}}"
	}

	e.funcMap["partial"] = func(name string, data ...any) string {
		return fmt.Sprintf("{{template \"%s\" %v}}", name, data)
	}

	e.funcMap["block"] = func(name string, data ...any) string {
		return fmt.Sprintf("{{block \"%s\" %v}}{{end}}", name, data)
	}

	e.funcMap["section"] = func(name string) string {
		return fmt.Sprintf("{{define \"%s\"}}", name)
	}

	e.funcMap["endsection"] = func() string {
		return "{{end}}"
	}

	// Beego布局相关函数
	e.funcMap["layout"] = func(layoutName string) string {
		return fmt.Sprintf("{{template \"%s\" .}}", layoutName)
	}

	// 静态资源函数
	e.funcMap["static"] = func(path string) string {
		return "/static/" + strings.TrimPrefix(path, "/")
	}

	e.funcMap["css"] = func(path string) string {
		return fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\">", e.funcMap["static"].(func(string) string)(path))
	}

	e.funcMap["js"] = func(path string) string {
		return fmt.Sprintf("<script src=\"%s\"></script>", e.funcMap["static"].(func(string) string)(path))
	}

	// Beego路由函数
	e.funcMap["urlfor"] = func(controller, action string, params ...string) string {
		url := "/" + strings.ToLower(controller) + "/" + strings.ToLower(action)
		if len(params) > 0 {
			url += "?" + strings.Join(params, "&")
		}
		return url
	}

	// 开发/生产模式检查
	e.funcMap["isDev"] = func() bool {
		return e.RunMode == "dev"
	}

	e.funcMap["isProd"] = func() bool {
		return e.RunMode == "prod"
	}

	config.Infof("Initialized %d Beego-style template functions", len(e.funcMap))
}

// BuildAllTemplates 方法已移至 engine_template.go 文件中

// buildTemplatesInPath 构建指定路径下的所有模板
func (e *TemplateEngine) buildTemplatesInPath(viewPath string) error {
	return filepath.WalkDir(viewPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			config.Warnf("Template walk error at %s: %v", path, err)
			return nil // 忽略错误，继续处理
		}

		// 跳过目录
		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名
		if !e.isTemplateFile(path) {
			return nil
		}

		// 计算相对路径作为模板名
		templateName, err := filepath.Rel(viewPath, path)
		if err != nil {
			config.Warnf("Failed to get relative path for %s: %v", path, err)
			return nil
		}

		// 统一路径分隔符
		templateName = filepath.ToSlash(templateName)

		// 移除扩展名
		templateName = e.removeExtension(templateName)

		// 构建模板
		if err := e.buildSingleTemplate(templateName, path); err != nil {
			config.Errorf("Failed to build template %s at %s: %v", templateName, path, err)
			return nil // 不中断构建过程
		}

		config.Debugf("Built template: %s -> %s", templateName, path)
		return nil
	})
}

// buildSingleTemplate 构建单个模板
func (e *TemplateEngine) buildSingleTemplate(templateName, filePath string) error {
	// 获取文件修改时间
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file %s: %w", filePath, err)
	}

	// 检查是否需要重新构建
	if lastMod, exists := e.lastModTime[templateName]; exists {
		if !fileInfo.ModTime().After(lastMod) && e.RunMode == "prod" {
			return nil // 生产模式下不重建未修改的模板
		}
	}

	// 创建新模板
	tmpl := template.New(templateName)
	tmpl = tmpl.Delims(e.delimLeft, e.delimRight)
	tmpl = tmpl.Funcs(e.funcMap)

	// 解析模板文件
	tmpl, err = tmpl.ParseFiles(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse template file %s: %w", filePath, err)
	}

	// 处理模板依赖和包含
	if err := e.processTemplateIncludes(tmpl, filePath); err != nil {
		return fmt.Errorf("failed to process template includes for %s: %w", templateName, err)
	}

	// 缓存模板
	e.templates[templateName] = tmpl
	e.lastModTime[templateName] = fileInfo.ModTime()

	return nil
}

// processTemplateIncludes 处理模板包含和依赖 (Beego风格)
func (e *TemplateEngine) processTemplateIncludes(tmpl *template.Template, basePath string) error {
	baseDir := filepath.Dir(basePath)

	// 查找当前目录和父目录中的模板文件进行自动包含
	for _, ext := range e.TemplateExt {
		// 检查布局文件
		layoutPath := filepath.Join(baseDir, "layout"+ext)
		if _, err := os.Stat(layoutPath); err == nil {
			if _, err := tmpl.ParseFiles(layoutPath); err != nil {
				config.Warnf("Failed to include layout %s: %v", layoutPath, err)
			}
		}

		// 检查公共文件
		commonPath := filepath.Join(baseDir, "common"+ext)
		if _, err := os.Stat(commonPath); err == nil {
			if _, err := tmpl.ParseFiles(commonPath); err != nil {
				config.Warnf("Failed to include common %s: %v", commonPath, err)
			}
		}
	}

	return nil
}

// shouldReloadTemplate 检查是否应该重新加载模板
func (e *TemplateEngine) shouldReloadTemplate(templateName string) bool {
	lastMod, exists := e.lastModTime[templateName]
	if !exists {
		return true
	}

	// 查找对应的文件路径
	for _, viewPath := range e.viewPaths {
		for _, ext := range e.TemplateExt {
			fullPath := filepath.Join(viewPath, templateName+ext)
			if fileInfo, err := os.Stat(fullPath); err == nil {
				return fileInfo.ModTime().After(lastMod)
			}
		}
	}

	return false
}

// reloadTemplate 重新加载模板
func (e *TemplateEngine) reloadTemplate(templateName string) error {
	// 查找模板文件
	for _, viewPath := range e.viewPaths {
		// 先检查是否已经带扩展名
		if filepath.Ext(templateName) != "" {
			// 如果已经带扩展名，直接使用
			fullPath := filepath.Join(viewPath, templateName)
			if _, err := os.Stat(fullPath); err == nil {
				return e.buildSingleTemplate(templateName, fullPath)
			}
		} else {
			// 如果没有扩展名，尝试添加各种扩展名
			for _, ext := range e.TemplateExt {
				fullPath := filepath.Join(viewPath, templateName+ext)
				if _, err := os.Stat(fullPath); err == nil {
					return e.buildSingleTemplate(templateName, fullPath)
				}
			}
		}
	}

	return fmt.Errorf("template file not found for %s", templateName)
}

// isTemplateFile 检查是否为模板文件
func (e *TemplateEngine) isTemplateFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	for _, templateExt := range e.TemplateExt {
		if ext == templateExt {
			return true
		}
	}
	return false
}
