// Package xstring 提供字符串处理的常用工具函数
//
// 本包提供了丰富的字符串操作功能，包括：
// - 字符串长度和子串操作
// - 字符串查找和替换
// - 字符串转换（大小写、首字母大写等）
// - 字符串修剪和填充
// - 字符串分割和连接
// - 正则表达式操作
// - HTML转义处理
//
// 使用示例：
//   import "framework/pkg/xstring"
//
//   // 字符串长度
//   length := xstring.MbStrlen("你好世界")
//
//   // 字符串截取
//   sub := xstring.Substr("hello world", 0, 5)
//
//   // 首字母大写
//   title := xstring.Ucfirst("hello")
//
//   // HTML转义
//   escaped := xstring.Htmlspecialchars("<script>alert('xss')</script>")
//
package xstring