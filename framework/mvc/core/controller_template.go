package core

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// ============= 模板渲染方法 =============

// Render 渲染模板（Beego ControllerInterface兼容）
func (c *BaseController) Render() error {
	// 添加详细调试信息
	c.LogInfof("=== Render() 开始 ===")
	c.LogInfof("初始状态 - ControllerName: %s, ActionName: %s, TplName: %s, ViewPath: %s",
		c.ControllerName, c.ActionName, c.TplName, c.ViewPath)

	// 如果模板名称为空，自动推导
	if c.TplName == "" {
		controllerName := strings.ToLower(c.ControllerName)
		actionName := strings.ToLower(c.ActionName)
		c.TplName = fmt.Sprintf("%s/%s", controllerName, actionName)
		c.LogInfof("自动推导模板名称: %s", c.TplName)
	}

	// 调用现有的renderTemplate方法
	return c.renderTemplate()
}

// RenderWithViewName 渲染指定模板（向后兼容版本）
func (c *BaseController) RenderWithViewName(viewName ...string) error {
	if len(viewName) > 0 {
		c.TplName = viewName[0]
	}

	// 如果模板名称为空，自动推导
	if c.TplName == "" {
		controllerName := strings.ToLower(c.ControllerName)
		actionName := strings.ToLower(c.ActionName)
		c.TplName = fmt.Sprintf("%s/%s", controllerName, actionName)
	}

	// 调用现有的renderTemplate方法
	return c.renderTemplate()
}

// RenderWithLayout 使用布局渲染模板（兼容旧版本）
func (c *BaseController) RenderWithLayout(viewName, layoutName string) {
	c.TplName = viewName
	if layoutName != "" {
		c.Layout = layoutName
	}
	c.Render()
}

// RenderBytes 渲染模板并返回字节数组（Beego兼容）
func (c *BaseController) RenderBytes() ([]byte, error) {
	if c.TplName == "" {
		return nil, fmt.Errorf("template name is empty")
	}

	if c.templateEngine != nil {
		if c.Layout != "" {
			if content, err := c.templateEngine.RenderWithLayout(c.TplName, c.Layout, c.Data); err != nil {
				return nil, err
			} else {
				return []byte(content), nil
			}
		} else {
			if content, err := c.templateEngine.Render(c.TplName, c.Data); err != nil {
				return nil, err
			} else {
				return []byte(content), nil
			}
		}
	}

	// 降级方案：直接使用Go模板
	var buf bytes.Buffer
	tplName := c.TplName
	if !strings.HasSuffix(tplName, c.TplExt) {
		tplName += c.TplExt
	}

	viewPath := filepath.Join(c.ViewPath, tplName)
	tmpl := template.New(filepath.Base(tplName))
	if len(c.TplFuncs) > 0 {
		tmpl = tmpl.Funcs(c.TplFuncs)
	}

	if c.Layout != "" {
		layoutPath := filepath.Join(c.LayoutPath, c.Layout)
		if _, err := os.Stat(layoutPath); err == nil {
			tmpl, err = tmpl.ParseFiles(layoutPath, viewPath)
			if err != nil {
				return nil, err
			}
			err = tmpl.ExecuteTemplate(&buf, filepath.Base(c.Layout), c.Data)
			return buf.Bytes(), err
		}
	}

	tmpl, err := tmpl.ParseFiles(viewPath)
	if err != nil {
		return nil, err
	}

	err = tmpl.Execute(&buf, c.Data)
	return buf.Bytes(), err
}

// RenderString 渲染模板并返回字符串（Beego兼容）
func (c *BaseController) RenderString() (string, error) {
	bytes, err := c.RenderBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// RenderHTML 直接渲染HTML模板（兼容旧版本，现已增强支持template函数）
func (c *BaseController) RenderHTML(viewName string, data ...map[string]any) {
	if len(data) > 0 {
		for k, v := range data[0] {
			c.Data[k] = v
		}
	}

	// 优先使用增强的模板引擎来支持template include功能
	c.initializeEnhancedTemplateEngine()

	// 如果有include引擎，使用它来支持{{template}}函数
	if c.includeEngine != nil {
		content, err := c.includeEngine.RenderTemplate(viewName, c.Data)
		if err != nil {
			// 如果增强引擎失败，降级到基础渲染
			c.renderHTMLFallback(viewName)
			return
		}

		c.SetHeader("Content-Type", "text/html; charset=utf-8")
		_, err = c.Write([]byte(content))
		if err != nil {
			c.Errorf("写入响应失败: %v", err)
		}
		return
	}

	// 降级到原始方法
	c.renderHTMLFallback(viewName)
}

// renderHTMLFallback 降级的HTML渲染方法
func (c *BaseController) renderHTMLFallback(viewName string) {
	// 设置模板名称，不使用布局
	c.TplName = viewName
	originalLayout := c.Layout
	c.Layout = "" // 临时取消布局

	c.Render()

	// 恢复原始布局设置
	c.Layout = originalLayout
}

// RenderTemplate 直接使用模板管理器渲染模板
func (c *BaseController) RenderTemplate(templateName string, data any) (string, error) {
	return c.GetTemplateManager().Render(templateName, data)
}

// RenderTemplateWithLayout 使用模板管理器渲染带布局的模板
func (c *BaseController) RenderTemplateWithLayout(templateName, layoutName string, data any) (string, error) {
	return c.GetTemplateManager().RenderWithLayout(templateName, layoutName, data)
}

// RenderTemplateComponent 渲染模板组件
func (c *BaseController) RenderTemplateComponent(componentName string, data any) (string, error) {
	return c.GetTemplateManager().RenderComponent(componentName, data)
}

// ============= 模板配置方法 =============

// SetTplName 设置模板名称（Beego兼容）
func (c *BaseController) SetTplName(tplName string) {
	c.TplName = tplName
}

// GetTplName 获取模板名称
func (c *BaseController) GetTplName() string {
	return c.TplName
}

// SetLayout 设置布局文件（Beego兼容）
func (c *BaseController) SetLayout(layout string) {
	c.Layout = layout
}

// GetLayout 获取布局文件
func (c *BaseController) GetLayout() string {
	return c.Layout
}

// AddTplFunc 添加模板函数（Beego兼容）
func (c *BaseController) AddTplFunc(name string, fn any) {
	if c.TplFuncs == nil {
		c.TplFuncs = make(template.FuncMap)
	}
	c.TplFuncs[name] = fn
}

// GetTemplateManager 获取模板管理器
func (c *BaseController) GetTemplateManager() *view.TemplateManager {
	return view.GetTemplateManager()
}

// SetTemplatePath 设置模板路径（便捷方法）
func (c *BaseController) SetTemplatePath(viewPath, layoutPath string) {
	c.ViewPath = viewPath
	c.LayoutPath = layoutPath
	c.ViewsPath = viewPath // 保持兼容性
}

// SetTemplateTheme 设置模板主题
func (c *BaseController) SetTemplateTheme(themeName string) error {
	return c.GetTemplateManager().SetTheme(themeName)
}

// GetTemplateTheme 获取当前模板主题
func (c *BaseController) GetTemplateTheme() string {
	return c.GetTemplateManager().GetCurrentTheme()
}

// AddTemplateFunction 添加模板函数
func (c *BaseController) AddTemplateFunction(name string, fn any) {
	c.GetTemplateManager().AddFunction(name, fn)
}

// ============= 内部模板渲染方法 =============

// renderTemplate 内部模板渲染方法（使用模板管理器）
func (c *BaseController) renderTemplate() error {
	c.LogInfof("=== renderTemplate() 开始 ===")
	c.LogInfof("EnableRender: %v, templateEngine: %v", c.EnableRender, c.templateEngine != nil)

	if !c.EnableRender {
		c.LogErrorf("Template rendering is disabled")
		return c.Errorf("template rendering is disabled")
	}

	// 确定模板文件名
	tplName := c.TplName
	if tplName == "" {
		return c.Errorf("template name is empty")
	}

	// 构建完整的模板路径
	if !strings.HasSuffix(tplName, c.TplExt) {
		tplName += c.TplExt
	}

	if c.TplPrefix != "" {
		tplName = c.TplPrefix + tplName
	}

	// 使用增强的模板引擎渲染
	if c.templateEngine != nil {
		// viewPath := filepath.Join(c.ViewPath, tplName)
		viewPath := tplName
		if c.Layout != "" {
			// 使用布局渲染
			if content, err := c.templateEngine.RenderWithLayout(viewPath, c.Layout, c.Data); err != nil {
				return err
			} else {
				c.SetHeader("Content-Type", "text/html; charset=utf-8")
				if _, err := c.Write([]byte(content)); err != nil {
					return err
				}
			}
		} else {
			// 直接渲染模板
			if content, err := c.templateEngine.Render(viewPath, c.Data); err != nil {
				return err
			} else {
				c.SetHeader("Content-Type", "text/html; charset=utf-8")
				if _, err := c.Write([]byte(content)); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// 降级到基础模板渲染
	return c.renderBasicTemplate(tplName)
}

// renderBasicTemplate 基础模板渲染（降级方案）
func (c *BaseController) renderBasicTemplate(tplName string) error {
	c.LogInfof("=== renderBasicTemplate() 开始 ===")
	c.LogInfof("模板参数 - tplName: %s, ViewPath: %s", tplName, c.ViewPath)

	viewPath := filepath.Join(c.ViewPath, tplName)
	c.LogInfof("完整模板路径: %s", viewPath)

	// 检查文件是否存在
	if _, err := os.Stat(viewPath); os.IsNotExist(err) {
		// 添加调试信息
		c.LogErrorf("Template file not found: %s (ControllerName: %s, ActionName: %s, TplName: %s)",
			viewPath, c.ControllerName, c.ActionName, c.TplName)
		return c.Errorf("template file not found: %s", viewPath)
	}

	var tmpl *template.Template
	var err error

	// 创建模板并添加自定义函数
	tmpl = template.New(filepath.Base(tplName))
	if len(c.TplFuncs) > 0 {
		tmpl = tmpl.Funcs(c.TplFuncs)
	}

	// 如果有布局文件
	if c.Layout != "" {
		layoutPath := filepath.Join(c.LayoutPath, c.Layout)
		if _, err := os.Stat(layoutPath); err == nil {
			tmpl, err = tmpl.ParseFiles(layoutPath, viewPath)
			if err != nil {
				return fmt.Errorf("failed to parse template with layout: %v", err)
			}

			c.SetHeader("Content-Type", "text/html; charset=utf-8")
			var buf bytes.Buffer
			if err := tmpl.ExecuteTemplate(&buf, "layout", c.Data); err != nil {
				return err
			}
			_, err := c.Write(buf.Bytes())
			return err
		}
	}

	// 解析视图文件和相关子模板
	templateFiles := []string{viewPath}

	// 尝试找到同目录下的子模板文件
	dir := filepath.Dir(viewPath)
	if files, err := filepath.Glob(filepath.Join(dir, "_*.html")); err == nil {
		templateFiles = append(templateFiles, files...)
	}

	tmpl, err = tmpl.ParseFiles(templateFiles...)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c.Data); err != nil {
		return err
	}
	// 添加调试日志：显示模板内容长度
	c.LogInfof("Template rendered successfully: %d bytes (ControllerName: %s, ActionName: %s)",
		len(buf.Bytes()), c.ControllerName, c.ActionName)

	_, err = c.Write(buf.Bytes())
	if err != nil {
		c.LogErrorf("Failed to write response: %v", err)
	}
	return err
}
