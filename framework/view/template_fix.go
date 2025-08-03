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
func NewTemplateIncludeEngine(cfg *TemplateConfig) (*TemplateIncludeEngine, error) {
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
	e.masterTemplate = template.New("master").
		Delims(e.delimLeft, e.delimRight).
		Funcs(e.funcMap)
	
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
		// 尝试其他可能的名称
		alternativeKeys := e.generateAlternativeKeys(templateName)
		for _, key := range alternativeKeys {
			if tmpl = e.masterTemplate.Lookup(key); tmpl != nil {
				break
			}
		}
		
		if tmpl == nil {
			return "", fmt.Errorf("template '%s' not found (tried: %v)", templateName, append([]string{templateKey}, alternativeKeys...))
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
func (e *TemplateIncludeEngine) normalizeTemplateKey(templateName string) string {
	// 移除扩展名
	key := strings.TrimSuffix(templateName, e.extension)
	
	// 标准化路径分隔符
	key = strings.ReplaceAll(key, "\\", "/")
	
	return key
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
	
	// 添加只有文件名的版本
	baseName := strings.TrimSuffix(filepath.Base(templateName), e.extension)
	baseNameWithExt := filepath.Base(templateName)
	
	if baseName != templateName && baseName != nameWithoutExt {
		keys = append(keys, baseName)
	}
	if baseNameWithExt != templateName {
		keys = append(keys, baseNameWithExt)
	}
	
	// 添加相对路径版本（移除views/前缀）
	for _, viewPath := range e.viewPaths {
		if strings.HasPrefix(templateName, viewPath+"/") {
			relPath := strings.TrimPrefix(templateName, viewPath+"/")
			if relPath != templateName {
				keys = append(keys, relPath)
				keys = append(keys, strings.TrimSuffix(relPath, e.extension))
			}
		}
	}
	
	return keys
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