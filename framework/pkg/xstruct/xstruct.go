// Package xstruct 提供结构体工具函数，包括变量类型检测、键值对操作和 JSON 处理等功能。
//
// 主要功能：
//   - 变量类型检测和验证 (Isset, Empty, IsNull, IsArray 等)
//   - 数据类型转换 (Strval, Boolval 等)
//   - 键值对接口和实现
//   - JSON 编码解码及相关操作
//   - 数据调试和输出 (VarDump, PrintR 等)
//
// 使用示例：
//
//	import "your-project/pkg/xstruct"
//
//	// 变量检测
//	if xstruct.Isset(someVar) {
//	    fmt.Println("变量已设置")
//	}
//
//	// JSON 操作
//	jsonStr := xstruct.JsonEncode(data)
//	result := xstruct.JsonDecode(jsonStr, true)
//
//	// 键值对操作
//	kv := &xstruct.SimpleKV{Key: "name", Value: "张三"}
//	kvs := xstruct.NewKVs(kv)
//
// 注意事项：
//   - 该包中的函数大多支持 any 类型，提供了很好的通用性
//   - JSON 相关函数提供了丰富的选项和错误处理机制
//   - 键值对接口设计支持链式调用，使用便捷
package xstruct