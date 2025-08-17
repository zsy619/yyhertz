// Package gin 提供Gin风格的Web框架API
//
// 本包基于CloudWeGo Hertz引擎，提供完全兼容Gin框架的API接口，
// 解决了原生Gin与Hertz的context类型冲突问题，让开发者可以
// 无缝迁移现有Gin项目，同时享受Hertz的高性能优势。
//
// 主要特性：
//   - 完全兼容Gin API
//   - 基于高性能Hertz引擎
//   - 统一的Context类型
//   - 支持中间件链式调用
//   - 支持路由组和嵌套路由
//   - 多种数据绑定和渲染方式
//   - 内置常用中间件
//
// 使用示例：
//
//	r := gin.Default()
//	r.GET("/ping", func(c *gin.Context) {
//		c.JSON(200, gin.H{"message": "pong"})
//	})
//	r.Run(":8080")
package gin

const defaultMultipartMemory = 32 << 20 // 32 MB
const escapedColon = "\\:"
const colon = ":"
const backslash = "\\"
