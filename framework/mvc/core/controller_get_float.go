package core

import "haedu.gov.cn/tools/xgeneric"

// GetFloatTuple2 从请求参数中获取两个浮点数值，并返回一个包含这两个值的 Tuple2 结构体。
// 参数 key1 和 key2 分别对应需要获取的浮点数值的键名。
// 返回值 cnt 是一个 Tuple2 结构体，其中 A 和 B 字段分别存储 key1 和 key2 对应的浮点数值。
func (c *BaseController) GetFloatTuple2(key1, key2 string) (cnt xgeneric.Tuple2[float64, float64]) {
	cnt.A = c.GetFloat(key1)
	cnt.B = c.GetFloat(key2)
	return
}
