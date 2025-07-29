// Package framework 提供了基于CloudWeGo-Hertz的类Beego MVC框架
//
// 这个包是框架的统一入口，提供了便捷的导入和初始化功能。
//
// 主要模块包括：
//   - controller: 控制器和路由管理
//   - middleware: 中间件支持
//   - cache: 缓存管理
//   - captcha: 验证码生成
//   - config: 配置管理
//   - session: 会话管理
//   - types: 通用类型和常量
//   - util: 实用工具集
//   - view: 视图和模板支持
//
// 示例用法：
//
//	import "hertz-controller/framework"
//	import "hertz-controller/framework/controller"
//	import "hertz-controller/framework/middleware"
//
//	func main() {
//	    app := controller.NewApp()
//	    app.Use(middleware.LoggerMiddleware())
//	    app.RegisterController("/api", &MyController{})
//	    app.Spin()
//	}
package framework

import (
	_ "hertz-controller/framework/cache"     // 缓存管理
	_ "hertz-controller/framework/captcha"   // 验证码支持
	_ "hertz-controller/framework/config"    // 配置管理
	_ "hertz-controller/framework/session"   // 会话管理
	_ "hertz-controller/framework/types"     // 通用类型
	_ "hertz-controller/framework/util"      // 工具函数
	_ "hertz-controller/framework/view"      // 视图模板
)

const (
	// Version 框架版本
	Version = "1.0.0"
	
	// Name 框架名称
	Name = "Hertz MVC Framework"
	
	// Description 框架描述
	Description = "基于CloudWeGo-Hertz的类Beego MVC框架"
)

// GetVersion 获取框架版本信息
func GetVersion() string {
	return Version
}

// GetName 获取框架名称
func GetName() string {
	return Name
}

// GetDescription 获取框架描述
func GetDescription() string {
	return Description
}

// GetInfo 获取框架信息
func GetInfo() map[string]string {
	return map[string]string{
		"name":        Name,
		"version":     Version,
		"description": Description,
	}
}