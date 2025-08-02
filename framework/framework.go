// Package framework 提供了基于CloudWeGo-Hertz的类Beego MVC框架
//
// 这个包是框架的统一入口，提供了便捷的导入和初始化功能。
//
// 主要模块包括：
//   - mvc: MVC控制器和路由管理
//   - orm: 数据库ORM集成
//   - middleware: 中间件支持
//   - cache: 缓存管理
//   - config: 配置管理
//   - session: 会话管理
//   - util: 实用工具集
//   - validation: 数据验证器
//   - view: 视图和模板支持
//   - testing: 单元测试工具集
//   - scheduler: 任务调度系统
//   - i18n: 国际化支持
//   - metrics: 监控和指标

package framework

import (
	_ "github.com/zsy619/yyhertz/framework/cache"      // 缓存管理
	_ "github.com/zsy619/yyhertz/framework/config"     // 配置管理
	_ "github.com/zsy619/yyhertz/framework/constant"   // 常量定义
	_ "github.com/zsy619/yyhertz/framework/mvc"        // MVC控制器（包含路由和处理器）
	_ "github.com/zsy619/yyhertz/framework/orm"        // ORM数据库集成
	_ "github.com/zsy619/yyhertz/framework/response"   // 响应管理
	_ "github.com/zsy619/yyhertz/framework/session"    // 会话管理
	_ "github.com/zsy619/yyhertz/framework/util"       // 通用类型和工具
	_ "github.com/zsy619/yyhertz/framework/validation" // 数据验证
	_ "github.com/zsy619/yyhertz/framework/view"       // 视图模板
	_ "github.com/zsy619/yyhertz/framework/testing"    // 单元测试工具
	_ "github.com/zsy619/yyhertz/framework/scheduler"  // 任务调度系统
	_ "github.com/zsy619/yyhertz/framework/i18n"       // 国际化支持
	_ "github.com/zsy619/yyhertz/framework/metrics"    // 监控和指标
	_ "github.com/zsy619/yyhertz/framework/middleware" // 增强中间件系统
	_ "github.com/zsy619/yyhertz/framework/binding"    // 参数绑定系统
	_ "github.com/zsy619/yyhertz/framework/render"     // 渲染系统
	_ "github.com/zsy619/yyhertz/framework/context"    // 增强上下文系统
	_ "github.com/zsy619/yyhertz/framework/gin"        // Gin风格API
)

const (
	// Version 框架版本
	Version = "1.4.0"

	// Name 框架名称
	Name = "YYHertz Framework"

	// Description 框架描述
	Description = "基于CloudWeGo-Hertz的现代化Go Web框架，融合Beego和Gin的最佳实践"
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
