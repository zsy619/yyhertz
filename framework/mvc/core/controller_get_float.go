package core

import "haedu.gov.cn/tools/xgeneric"

// GetFloatTuple2 从请求参数中获取两个浮点数值，并返回一个包含这两个值的 Tuple2 结构体。
// 参数 key1 和 key2 分别对应需要获取的浮点数值的键名。
// 返回值 cnt 是一个 Tuple2 结构体，其中 A 和 B 字段分别存储 key1 和 key2 对应的浮点数值。
func (c *BaseController) GetFloat32Tuple2(key1, key2 string) (cnt xgeneric.Tuple2[float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	return
}

func (c *BaseController) GetFloat32Tuple3(key1, key2, key3 string) (cnt xgeneric.Tuple3[float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	return
}

func (c *BaseController) GetFloat32Tuple4(key1, key2, key3, key4 string) (cnt xgeneric.Tuple4[float32, float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	cnt.D = c.GetFloat32(key4)
	return
}

func (c *BaseController) GetFloat32Tuple5(key1, key2, key3, key4, key5 string) (cnt xgeneric.Tuple5[float32, float32, float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	cnt.D = c.GetFloat32(key4)
	cnt.E = c.GetFloat32(key5)
	return
}

func (c *BaseController) GetFloat32Tuple6(key1, key2, key3, key4, key5, key6 string) (cnt xgeneric.Tuple6[float32, float32, float32, float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	cnt.D = c.GetFloat32(key4)
	cnt.E = c.GetFloat32(key5)
	cnt.F = c.GetFloat32(key6)
	return
}

func (c *BaseController) GetFloat32Tuple7(key1, key2, key3, key4, key5, key6, key7 string) (cnt xgeneric.Tuple7[float32, float32, float32, float32, float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	cnt.D = c.GetFloat32(key4)
	cnt.E = c.GetFloat32(key5)
	cnt.F = c.GetFloat32(key6)
	cnt.G = c.GetFloat32(key7)
	return
}
