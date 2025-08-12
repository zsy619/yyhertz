package core

import (
	"fmt"
	"path/filepath"

	templatemanager "github.com/zsy619/yyhertz/framework/template"
	"github.com/zsy619/yyhertz/framework/view"
)

// initializeEnhancedTemplateEngine 初始化增强的模板引擎
func (c *BaseController) initializeEnhancedTemplateEngine() {
	if c.templateEngine == nil {
		// 尝试使用增强的模板引擎
		cfg := view.DefaultTemplateConfig()
		if c.ViewPath != "" {
			// 扩展视图路径，包含基础路径、shared目录和其他常用目录
			cfg.ViewPaths = []string{
				c.ViewPath,                             // 主视图路径
				filepath.Join(c.ViewPath, "shared"),    // shared模板目录
				filepath.Join(c.ViewPath, "layouts"),   // 布局目录
				filepath.Join(c.ViewPath, "components"), // 组件目录
			}
		}
		if enhanced, err := view.NewTemplateIncludeEngine(cfg); err == nil {
			c.templateEngine = enhanced.TemplateEngine
			c.includeEngine = enhanced
			
			// 自动添加Beego风格的模板函数
			c.addBeegoTemplateFunctions()
			
			// 添加应用级别的全局模板函数
			c.addGlobalTemplateFunctions()
		} else {
			// 降级到标准模板引擎
			c.templateEngine = templatemanager.GetTemplateManager().GetEngine()
		}
	}
}

// addBeegoTemplateFunctions 添加Beego风格的模板函数（内部方法）
func (c *BaseController) addBeegoTemplateFunctions() {
	if c.templateEngine != nil {
		// 添加所有Beego风格的模板函数
		for name, fn := range view.BeegoTemplateFuncs {
			c.templateEngine.AddFunction(name, fn)
		}
		
		// 如果有include引擎，也添加到它那里
		if c.includeEngine != nil {
			for name, fn := range view.BeegoTemplateFuncs {
				c.includeEngine.AddFunction(name, fn)
			}
		}
	}
}

// RenderHTMLWithIncludes 使用支持include的模板引擎渲染
func (c *BaseController) RenderHTMLWithIncludes(viewName string, data ...map[string]any) error {
	if len(data) > 0 {
		for k, v := range data[0] {
			c.Data[k] = v
		}
	}
	
	// 初始化增强模板引擎
	c.initializeEnhancedTemplateEngine()
	
	// 如果有include引擎，使用它
	if c.includeEngine != nil {
		content, err := c.includeEngine.RenderTemplate(viewName, c.Data)
		if err != nil {
			return err
		}
		
		c.Ctx.RequestContext.Header("Content-Type", "text/html; charset=utf-8")
		c.Ctx.RequestContext.Write([]byte(content))
		return nil
	}
	
	// 降级到标准渲染
	return c.renderTemplate()
}

// SetTemplateIncludeEngine 设置template include引擎
func (c *BaseController) SetTemplateIncludeEngine(engine *view.TemplateIncludeEngine) {
	c.includeEngine = engine
	if engine != nil {
		c.templateEngine = engine.TemplateEngine
	}
}

// GetTemplateIncludeEngine 获取template include引擎
func (c *BaseController) GetTemplateIncludeEngine() *view.TemplateIncludeEngine {
	if c.includeEngine == nil {
		c.initializeEnhancedTemplateEngine()
	}
	return c.includeEngine
}

// CreateTemplateDefinition 创建模板定义
func (c *BaseController) CreateTemplateDefinition(name, content string) error {
	engine := c.GetTemplateIncludeEngine()
	if engine == nil {
		return fmt.Errorf("template include engine not available")
	}
	
	// 这里可以扩展支持动态模板定义
	return fmt.Errorf("dynamic template definition not implemented yet")
}

// ListAvailableTemplates 列出可用的模板
func (c *BaseController) ListAvailableTemplates() []string {
	engine := c.GetTemplateIncludeEngine()
	if engine == nil {
		return []string{}
	}
	
	return engine.ListAvailableTemplates()
}

// AddBeegoTemplateFunctions 添加Beego风格的模板函数
func (c *BaseController) AddBeegoTemplateFunctions() {
	for name, fn := range view.BeegoTemplateFuncs {
		c.AddTemplateFunction(name, fn)
	}
}

// addGlobalTemplateFunctions 添加应用级别的全局模板函数（内部方法）
func (c *BaseController) addGlobalTemplateFunctions() {
	if c.app != nil {
		// 从应用实例获取全局模板函数
		globalFuncs := c.app.GetGlobalFuncMap()
		
		if c.templateEngine != nil {
			// 添加到模板引擎
			for name, fn := range globalFuncs {
				c.templateEngine.AddFunction(name, fn)
			}
		}
		
		// 如果有include引擎，也添加到它那里
		if c.includeEngine != nil {
			for name, fn := range globalFuncs {
				c.includeEngine.AddFunction(name, fn)
			}
		}
	}
}