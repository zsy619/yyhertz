package main

import (
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/util"
	"strings"
)

// ExampleController 示例控制器
type ExampleController struct {
	mvc.BaseController
}

// GetIndex 首页
func (c *ExampleController) GetIndex() {
	c.SetData("Title", "AddFuncMap 功能演示")
	c.SetData("Tags", "important,priority,urgent")
	c.SetData("Username", "张三")
	c.SetData("Price", 299.99)
	c.RenderHTML("index.html")
}

// GetTest 测试页面
func (c *ExampleController) GetTest() {
	c.SetData("Message", "hello world")
	c.SetData("Items", []string{"apple", "banana", "orange"})
	c.RenderHTML("test.html")
}

func main() {
	// ============= 添加自定义模板函数 =============
	
	// 1. 使用框架内置的工具函数
	mvc.AddFuncMap("containString", util.ContainString)
	mvc.AddFuncMap("formatFloat2", util.FmtFloat2)
	mvc.AddFuncMap("formatByte", util.FmtByte)
	mvc.AddFuncMap("getTime", util.GetTime)
	
	// 2. 添加自定义字符串处理函数
	mvc.AddFuncMap("upper", func(s string) string {
		return strings.ToUpper(s)
	})
	
	mvc.AddFuncMap("lower", func(s string) string {
		return strings.ToLower(s)
	})
	
	mvc.AddFuncMap("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})
	
	// 3. 添加格式化函数
	mvc.AddFuncMap("formatPrice", func(price float64) string {
		return "¥" + util.FmtFloat2(price)
	})
	
	mvc.AddFuncMap("formatUser", func(username string) string {
		return "用户: " + username
	})
	
	// 4. 添加数组处理函数
	mvc.AddFuncMap("join", func(items []string, separator string) string {
		return strings.Join(items, separator)
	})
	
	mvc.AddFuncMap("contains", func(items []string, item string) bool {
		for _, v := range items {
			if v == item {
				return true
			}
		}
		return false
	})
	
	// 5. 添加条件判断函数
	mvc.AddFuncMap("isEmpty", func(s string) bool {
		return strings.TrimSpace(s) == ""
	})
	
	mvc.AddFuncMap("isNotEmpty", func(s string) bool {
		return strings.TrimSpace(s) != ""
	})
	
	// ============= 设置静态文件路径 =============
	mvc.SetStaticPath("static", "/static")
	
	// ============= 注册控制器 =============
	app := mvc.HertzApp
	app.AutoRouters(&ExampleController{})
	
	// ============= 显示已注册的模板函数 =============
	funcList := mvc.ListFuncMap()
	app.LogInfof("已注册的模板函数: %v", funcList)
	
	// ============= 启动应用 =============
	app.Run(":8080")
}