package core

import (
	"reflect"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
)

// ============= 基于Beego Init机制的真正解决方案 =============

// 注意：InitializableController接口已经合并到IController中
// 现在所有Controller都使用统一的IController接口

// BeegoStyleBaseController Beego风格的基础控制器
type BeegoStyleBaseController struct {
	// 私有字段（模仿Beego）
	controllerName string
	actionName     string

	// HTTP上下文
	Ctx *app.RequestContext

	// 公有字段
	Data          map[string]any // 模板数据
	ViewPath      string         // 视图路径
	LayoutPath    string         // 布局路径
	Layout        string         // 布局文件
	TplName       string         // 模板名称
	EnableRender  bool           // 是否启用渲染
	AppController any            // 应用控制器引用

	// 初始化状态
	initialized bool
	initOnce    sync.Once
}

// ============= 核心Init方法（完全模仿Beego）=============

// Init 生命周期方法（IController接口要求）
func (c *BeegoStyleBaseController) Init() {
	// 初始化Data
	if c.Data == nil {
		c.Data = make(map[string]any)
	}

	// 设置默认值
	if c.ViewPath == "" {
		c.ViewPath = "views"
	}
	if c.LayoutPath == "" {
		c.LayoutPath = "views/layout"
	}
	if c.Layout == "" {
		c.Layout = "layout.html"
	}
	c.EnableRender = true
}

// Prepare 生命周期方法（IController接口要求）
func (c *BeegoStyleBaseController) Prepare() {
	// 可以在这里添加预处理逻辑
}

// Finish 生命周期方法（IController接口要求）
func (c *BeegoStyleBaseController) Finish() {
	// 可以在这里添加清理逻辑
}

// InitWithContext Beego风格的初始化方法（IController接口要求）
func (c *BeegoStyleBaseController) InitWithContext(ctx *app.RequestContext, controllerName, actionName string, app any) {
	c.Ctx = ctx
	c.controllerName = controllerName // 关键：由框架注入！
	c.actionName = actionName         // 关键：由框架注入！
	c.AppController = app
	c.initialized = true

	// 调用基础初始化
	c.Init()
}

// GetControllerName 获取控制器名称（Beego兼容）
func (c *BeegoStyleBaseController) GetControllerName() string {
	return c.controllerName
}

// GetActionName 获取动作名称（Beego兼容）
func (c *BeegoStyleBaseController) GetActionName() string {
	return c.actionName
}

// ControllerName 属性访问器（用户友好）
func (c *BeegoStyleBaseController) ControllerName() string {
	// 如果没有初始化，尝试自动初始化
	if !c.initialized {
		c.autoInit()
	}
	return c.controllerName
}

// ActionName 属性访问器（用户友好）
func (c *BeegoStyleBaseController) ActionName() string {
	if !c.initialized {
		c.autoInit()
	}
	return c.actionName
}

// autoInit 自动初始化（用户直接实例化时的救援机制）
func (c *BeegoStyleBaseController) autoInit() {
	c.initOnce.Do(func() {
		// 通过反射获取Controller类型
		controllerName := c.detectControllerNameFromType()

		// 调用InitWithContext方法进行标准初始化
		c.InitWithContext(nil, controllerName, "index", c)
	})
}

// detectControllerNameFromType 从类型检测控制器名称
func (c *BeegoStyleBaseController) detectControllerNameFromType() string {
	// 这里仍然面临Go语言的限制，但我们可以尝试一些技巧
	actualType := reflect.TypeOf(c)
	if actualType.Kind() == reflect.Ptr {
		actualType = actualType.Elem()
	}

	typeName := actualType.Name()
	if typeName != "BeegoStyleBaseController" && strings.HasSuffix(typeName, "Controller") {
		return typeName[:len(typeName)-10]
	}

	// 如果检测失败，返回默认值
	return "Unknown"
}

// ============= 手动初始化方法（用户可调用）=============

// InitWithName 手动初始化（用户友好版本）
func (c *BeegoStyleBaseController) InitWithName(controllerName string) {
	c.InitWithContext(nil, controllerName, "index", c)
}

// QuickInit 快速初始化（自动推断名称）
func (c *BeegoStyleBaseController) QuickInit() {
	controllerName := c.detectControllerNameFromType()
	c.InitWithContext(nil, controllerName, "index", c)
}

// ============= 辅助方法 =============

// SetData 设置模板数据
func (c *BeegoStyleBaseController) SetData(key string, value any) {
	if c.Data == nil {
		c.Data = make(map[string]any)
	}
	c.Data[key] = value
}

// GetData 获取模板数据
func (c *BeegoStyleBaseController) GetData(key string) any {
	if c.Data == nil {
		return nil
	}
	return c.Data[key]
}

// DelData 删除模板数据
func (c *BeegoStyleBaseController) DelData(key string) {
	if c.Data != nil {
		delete(c.Data, key)
	}
}

// ============= 注意：ControllerManager现在统一在factory.go中实现 =============

// 这里不再重复定义ControllerManager，使用factory.go中的统一实现

// 保留这些便捷函数的实现将在factory.go中完成

// ============= 向后兼容 =============

// 注意：不再需要 BaseController 类型别名，因为已经直接在 controller.go 中实现
// 所有的 Beego 风格功能现在都整合在统一的 BaseController 中

// NewBeegoStyleBaseController 创建Beego风格基础控制器
func NewBeegoStyleBaseController() *BeegoStyleBaseController {
	return &BeegoStyleBaseController{
		Data:         make(map[string]any),
		EnableRender: true,
		ViewPath:     "views",
		LayoutPath:   "views/layout",
		Layout:       "layout.html",
	}
}
