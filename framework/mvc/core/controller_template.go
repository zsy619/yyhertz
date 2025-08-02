package core

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	templatemanager "github.com/zsy619/yyhertz/framework/template"
)

// ============= 模板渲染方法 =============

// Render 渲染模板（Beego ControllerInterface兼容）
func (c *BaseController) Render() error {
	// 如果模板名称为空，自动推导
	if c.TplName == "" {
		controllerName := strings.ToLower(c.ControllerName)
		actionName := strings.ToLower(c.ActionName)
		c.TplName = fmt.Sprintf("%s/%s", controllerName, actionName)
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

// RenderHTML 直接渲染HTML模板（兼容旧版本）
func (c *BaseController) RenderHTML(viewName string, data ...map[string]any) {
	if len(data) > 0 {
		for k, v := range data[0] {
			c.Data[k] = v
		}
	}

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
func (c *BaseController) GetTemplateManager() *templatemanager.TemplateManager {
	return templatemanager.GetTemplateManager()
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
	if !c.EnableRender {
		return fmt.Errorf("template rendering is disabled")
	}

	// 确定模板文件名
	tplName := c.TplName
	if tplName == "" {
		return fmt.Errorf("template name is empty")
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
		if c.Layout != "" {
			// 使用布局渲染
			if content, err := c.templateEngine.RenderWithLayout(tplName, c.Layout, c.Data); err != nil {
				return err
			} else {
				c.Ctx.RequestContext.Header("Content-Type", "text/html; charset=utf-8")
				c.Ctx.RequestContext.Write([]byte(content))
			}
		} else {
			// 直接渲染模板
			if content, err := c.templateEngine.Render(tplName, c.Data); err != nil {
				return err
			} else {
				c.Ctx.RequestContext.Header("Content-Type", "text/html; charset=utf-8")
				c.Ctx.RequestContext.Write([]byte(content))
			}
		}
		return nil
	}

	// 降级到基础模板渲染
	return c.renderBasicTemplate(tplName)
}

// renderBasicTemplate 基础模板渲染（降级方案）
func (c *BaseController) renderBasicTemplate(tplName string) error {
	viewPath := filepath.Join(c.ViewPath, tplName)

	// 检查文件是否存在
	if _, err := os.Stat(viewPath); os.IsNotExist(err) {
		return fmt.Errorf("template file not found: %s", viewPath)
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

			c.Ctx.RequestContext.Header("Content-Type", "text/html; charset=utf-8")
			return tmpl.ExecuteTemplate(c.Ctx.RequestContext, "layout", c.Data)
		}
	}

	// 只解析视图文件
	tmpl, err = tmpl.ParseFiles(viewPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	c.Ctx.RequestContext.Header("Content-Type", "text/html; charset=utf-8")
	return tmpl.Execute(c.Ctx.RequestContext, c.Data)
}
