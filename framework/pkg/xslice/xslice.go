// Package xslice 提供了数组和切片操作的工具函数
//
// 主要功能：
//   - 数组操作：查找、过滤、映射、去重、排序等
//   - 切片操作：追加、删除、插入、拼接等
//   - 类型转换：字符串数组转换、数字数组转换等
//   - 集合操作：交集、并集、差集、对称差集等
//
// 基础用法：
//
//	import "your-project/framework/pkg/xslice"
//
//	// 数组查找
//	arr := []string{"apple", "banana", "orange"}
//	index := xslice.InArray("banana", arr) // 返回索引位置
//	exists := xslice.ArrayKeyExists(1, arr) // 检查索引是否存在
//
//	// 数组操作
//	unique := xslice.ArrayUnique([]int{1, 2, 2, 3, 3, 4}) // 去重
//	filtered := xslice.ArrayFilter(arr, func(v string) bool {
//		return len(v) > 5
//	})
//
//	// 切片操作
//	slice := []int{1, 2, 3, 4, 5}
//	chunk := xslice.ArraySlice(slice, 1, 3) // 获取子切片
//	reversed := xslice.ArrayReverse(slice)   // 反转
//
// 高级用法：
//
//	// 数组映射
//	numbers := []int{1, 2, 3, 4, 5}
//	doubled := xslice.ArrayMap(numbers, func(v int) int {
//		return v * 2
//	})
//
//	// 数组聚合
//	sum := xslice.ArrayReduce(numbers, 0, func(acc, v int) int {
//		return acc + v
//	})
//
//	// 多维数组操作
//	nested := [][]int{{1, 2}, {3, 4}, {5, 6}}
//	flattened := xslice.ArrayFlatten(nested)
//
//	// 集合操作
//	set1 := []int{1, 2, 3, 4}
//	set2 := []int{3, 4, 5, 6}
//	intersection := xslice.ArrayIntersect(set1, set2) // 交集
//	union := xslice.ArrayMerge(set1, set2)           // 并集
//
package xslice