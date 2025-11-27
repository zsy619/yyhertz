package view

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
)

// TemplateIncludeEngine 支持template include的模板引擎
type TemplateIncludeEngine struct {
	*TemplateEngine
	
	// 所有模板的统一实例
	masterTemplate *template.Template
}

// NewTemplateIncludeEngine 创建支持include的模板引擎
func NewTemplateIncludeEngine(cfg *config.TemplateConfig) (*TemplateIncludeEngine, error) {
	baseEngine, err := NewTemplateEngine(cfg)
	if err != nil {
		return nil, err
	}
	
	engine := &TemplateIncludeEngine{
		TemplateEngine: baseEngine,
	}
	
	// 重新加载模板以支持include
	if err := engine.loadAllTemplatesForInclude(); err != nil {
		return nil, fmt.Errorf("failed to load templates for include: %w", err)
	}
	
	return engine, nil
}

// loadAllTemplatesForInclude 加载所有模板到一个主模板中
func (e *TemplateIncludeEngine) loadAllTemplatesForInclude() error {
	// 创建主模板
	// 动态获取最新的合并函数（包含用户通过mvc.AddFuncMap注册的函数）
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)
	
	e.masterTemplate = template.New("master").
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs)
	
	// 收集所有模板文件
	templateFiles := make([]string, 0)
	
	// 扫描所有目录
	dirs := append(e.viewPaths, e.layoutPath, e.componentPath)
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			
			if !d.IsDir() && strings.HasSuffix(path, e.extension) {
				templateFiles = append(templateFiles, path)
			}
			
			return nil
		})
		
		if err != nil {
			config.Warnf("Error walking directory %s: %v", dir, err)
		}
	}
	
	// 解析所有模板文件到主模板
	if len(templateFiles) > 0 {
		if _, err := e.masterTemplate.ParseFiles(templateFiles...); err != nil {
			return fmt.Errorf("failed to parse template files: %w", err)
		}
	}
	
	config.Infof("Loaded %d template files for include support", len(templateFiles))
	
	return nil
}

// RenderTemplate 渲染模板（支持include）
func (e *TemplateIncludeEngine) RenderTemplate(templateName string, data any) (string, error) {
	if e.masterTemplate == nil {
		return "", fmt.Errorf("master template not initialized")
	}
	
	// 标准化模板名称
	templateKey := e.normalizeTemplateKey(templateName)
	
	// 查找模板
	tmpl := e.masterTemplate.Lookup(templateKey)
	if tmpl == nil {
		// 🔧 增强调试：收集所有尝试的键名
		alternativeKeys := e.generateAlternativeKeys(templateName)
		
		// 尝试其他可能的名称
		for _, key := range alternativeKeys {
			if tmpl = e.masterTemplate.Lookup(key); tmpl != nil {
				// 找到模板，记录成功匹配的键名（简化版本，不依赖config包）
				break
			}
		}
		
		if tmpl == nil {
			// 🔧 增强错误诊断：提供更详细的调试信息
			availableTemplates := e.ListAvailableTemplates()
			return "", fmt.Errorf("❌ 模板查找失败:\n  - 请求模板: %s\n  - 标准化键名: %s\n  - 尝试的键名: %v\n  - 可用模板 (前10个): %v\n  - 总可用模板数: %d", 
				templateName, templateKey, append([]string{templateKey}, alternativeKeys...), 
				func() []string {
					if len(availableTemplates) > 10 {
						return availableTemplates[:10]
					}
					return availableTemplates
				}(),
				len(availableTemplates))
		}
	}
	
	// 直接使用原始数据，不进行包装
	renderData := data
	
	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderData); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}
	
	return buf.String(), nil
}

// normalizeTemplateKey 标准化模板键名 
// 🔧 关键修复：优先使用文件名作为模板键，因为ParseFiles创建的模板名是文件basename
func (e *TemplateIncludeEngine) normalizeTemplateKey(templateName string) string {
	// 标准化路径分隔符
	normalized := strings.ReplaceAll(templateName, "\\", "/")
	
	// 🔧 优先返回文件basename（含扩展名），因为这与ParseFiles的行为一致
	// ParseFiles会使用文件的basename作为模板名，如 "index.html"
	if strings.Contains(normalized, "/") {
		// 如果是路径形式（如 "template/index.html"），返回文件basename
		result := filepath.Base(normalized)
		fmt.Printf("🔧 normalizeTemplateKey: %s -> %s (using basename)\n", templateName, result)
		return result
	}
	
	// 如果已经是简单文件名，直接返回
	fmt.Printf("🔧 normalizeTemplateKey: %s -> %s (simple name)\n", templateName, normalized)
	return normalized
}

// generateAlternativeKeys 生成可能的模板键名
func (e *TemplateIncludeEngine) generateAlternativeKeys(templateName string) []string {
	keys := make([]string, 0)
	
	// 添加原始名称
	keys = append(keys, templateName)
	
	// 添加不含扩展名的版本
	nameWithoutExt := strings.TrimSuffix(templateName, e.extension)
	if nameWithoutExt != templateName {
		keys = append(keys, nameWithoutExt)
	}
	
	// 🔧 关键修复：处理路径形式的模板名，如 "template/index.html"
	// 添加只有文件名的版本
	baseName := strings.TrimSuffix(filepath.Base(templateName), e.extension)
	baseNameWithExt := filepath.Base(templateName)
	
	if baseName != templateName && baseName != nameWithoutExt {
		keys = append(keys, baseName)
	}
	if baseNameWithExt != templateName {
		keys = append(keys, baseNameWithExt)
	}
	
	// 🔧 新增：特殊处理路径形式的模板名
	// 当请求 "template/index.html" 时，尝试查找 "index.html"
	if strings.Contains(templateName, "/") {
		// 提取最后的文件名部分
		fileName := filepath.Base(templateName)
		fileNameWithoutExt := strings.TrimSuffix(fileName, e.extension)
		
		// 🔧 重要：由于template.ParseFiles()创建的模板名通常是文件的basename
		// 我们需要优先尝试basename形式，将其添加到keys开头
		if !contains(keys, fileNameWithoutExt) && fileNameWithoutExt != templateName {
			// 将basename形式放在前面，提高匹配优先级
			keys = append([]string{fileNameWithoutExt}, keys...)
		}
		
		// 添加带扩展名的文件名
		if !contains(keys, fileName) && fileName != templateName {
			keys = append(keys, fileName)
		}
	}
	
	// 添加相对路径版本（移除views/前缀）
	for _, viewPath := range e.viewPaths {
		if strings.HasPrefix(templateName, viewPath+"/") {
			relPath := strings.TrimPrefix(templateName, viewPath+"/")
			if relPath != templateName && !contains(keys, relPath) {
				keys = append(keys, relPath)
			}
			relPathWithoutExt := strings.TrimSuffix(relPath, e.extension)
			if !contains(keys, relPathWithoutExt) {
				keys = append(keys, relPathWithoutExt)
			}
		}
	}
	
	return keys
}

// contains 检查字符串切片是否包含特定值
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetMasterTemplate 获取主模板实例
func (e *TemplateIncludeEngine) GetMasterTemplate() *template.Template {
	return e.masterTemplate
}

// ListAvailableTemplates 列出所有可用的模板
func (e *TemplateIncludeEngine) ListAvailableTemplates() []string {
	if e.masterTemplate == nil {
		return []string{}
	}
	
	templates := make([]string, 0)
	for _, tmpl := range e.masterTemplate.Templates() {
		if tmpl.Name() != "" && tmpl.Name() != "master" {
			templates = append(templates, tmpl.Name())
		}
	}
	
	return templates
}