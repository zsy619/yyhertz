// Package xhttp 提供HTTP相关的常用工具函数
//
// 本包提供了丰富的HTTP操作功能，包括：
// - Cookie处理和管理
// - 验证码生成和验证
// - HTTP消息格式化
// - 请求响应处理工具
//
// 使用示例：
//   import "framework/pkg/xhttp"
//
//   // Cookie操作
//   cookie := xhttp.SetCookie("name", "value", 3600)
//
//   // 验证码生成
//   captcha := xhttp.GenerateCaptcha(5)
//
//   // HTTP消息格式化
//   msg := xhttp.FormatResponse(200, "success", data)
//
package xhttp