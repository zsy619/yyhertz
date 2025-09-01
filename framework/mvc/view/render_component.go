package view

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
)

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

// LoadComponent 加载组件
func (e *TemplateEngine) LoadComponent(componentName string) (*template.Template, error) {
	componentPath, err := e.FindComponentFile(componentName)
	if err != nil {
		config.Errorf("❌ Component file not found: %s", componentName)
		return nil, fmt.Errorf("component '%s' not found: %w", componentName, err)
	}

	config.Debugf("🔄 Loading component: %s from path: %s", componentName, componentPath)

	// 动态获取最新的合并函数
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(e.funcMap)

	// 创建组件模板
	tmpl, err := template.New(componentName).
		Delims(e.delimLeft, e.delimRight).
		Funcs(mergedFuncs).
		ParseFiles(componentPath)

	if err != nil {
		config.Errorf("❌ Failed to parse component %s at %s: %v", componentName, componentPath, err)
		return nil, fmt.Errorf("failed to parse component %s at %s: %w", componentName, componentPath, err)
	}

	// 查找实际的组件模板
	templates := tmpl.Templates()
	var actualComponent *template.Template

	for _, t := range templates {
		if t.Tree != nil && t.Tree.Root != nil {
			actualComponent = t
			break
		}
	}

	if actualComponent == nil {
		return nil, fmt.Errorf("component %s is empty or invalid", componentName)
	}

	config.Debugf("✅ Component loaded successfully: %s", componentName)
	return actualComponent, nil
}

// FindComponentFile 查找组件文件
func (e *TemplateEngine) FindComponentFile(componentName string) (string, error) {
	// 确保组件名有扩展名
	if !strings.HasSuffix(componentName, e.extension) {
		componentName += e.extension
	}

	// 首先在组件路径中查找
	if e.componentPath != "" {
		componentPath := filepath.Join(e.componentPath, componentName)
		if stat, err := os.Stat(componentPath); err == nil && !stat.IsDir() {
			return componentPath, nil
		}
	}

	// 然后在视图路径的components子目录中查找
	for _, viewPath := range e.viewPaths {
		componentsDir := filepath.Join(viewPath, "components")
		componentPath := filepath.Join(componentsDir, componentName)
		if stat, err := os.Stat(componentPath); err == nil && !stat.IsDir() {
			return componentPath, nil
		}
	}

	return "", fmt.Errorf("component file '%s' not found in component paths", componentName)
}

// RegisterComponent 注册组件
func (e *TemplateEngine) RegisterComponent(componentName string, component *template.Template) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	if e.components == nil {
		e.components = make(map[string]*template.Template)
	}

	e.components[componentName] = component
	config.Debugf("Component registered: %s", componentName)
}

// UnregisterComponent 注销组件
func (e *TemplateEngine) UnregisterComponent(componentName string) {
	e.templateMutex.Lock()
	defer e.templateMutex.Unlock()

	delete(e.components, componentName)
	config.Debugf("Component unregistered: %s", componentName)
}

// LoadAndRegisterComponent 加载并注册组件
func (e *TemplateEngine) LoadAndRegisterComponent(componentName string) error {
	component, err := e.LoadComponent(componentName)
	if err != nil {
		return fmt.Errorf("failed to load component %s: %w", componentName, err)
	}

	e.RegisterComponent(componentName, component)
	return nil
}

// LoadAllComponents 加载所有组件
func (e *TemplateEngine) LoadAllComponents() error {
	if e.componentPath == "" {
		return nil // 没有配置组件路径
	}

	return filepath.Walk(e.componentPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, e.extension) {
			return nil
		}

		// 获取相对路径作为组件名
		relPath, err := filepath.Rel(e.componentPath, path)
		if err != nil {
			return err
		}

		componentName := strings.TrimSuffix(relPath, e.extension)
		return e.LoadAndRegisterComponent(componentName)
	})
}

// GetComponentNames 获取已注册的组件名称列表
func (e *TemplateEngine) GetComponentNames() []string {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	names := make([]string, 0, len(e.components))
	for name := range e.components {
		names = append(names, name)
	}
	return names
}

// HasComponent 检查组件是否存在
func (e *TemplateEngine) HasComponent(componentName string) bool {
	e.templateMutex.RLock()
	defer e.templateMutex.RUnlock()

	_, exists := e.components[componentName]
	return exists
}