// Package xdate 提供日期时间处理的常用工具函数
//
// 本包提供了丰富的日期时间操作功能，包括：
// - Unix时间戳操作
// - 日期格式化输出
// - 时区处理（本地时间、UTC时间）
// - 日期计算和转换
// - PHP风格的日期函数
//
// 使用示例：
//   import "framework/pkg/xdate"
//
//   // 获取当前时间戳
//   now := xdate.Time()
//
//   // 格式化日期
//   dateStr := xdate.Date("2006-01-02 15:04:05")
//
//   // 创建时间戳
//   timestamp := xdate.Mktime(14, 30, 0, 12, 25, 2023)
//
//   // PHP风格日期格式化
//   phpDate := xdate.PhpDate("Y-m-d H:i:s")
//
package xdate