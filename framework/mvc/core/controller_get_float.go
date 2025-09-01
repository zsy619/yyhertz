package core

import "github.com/zsy619/tools/xgeneric"

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

func (c *BaseController) GetFloat32Tuple8(key1, key2, key3, key4, key5, key6, key7, key8 string) (cnt xgeneric.Tuple8[float32, float32, float32, float32, float32, float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	cnt.D = c.GetFloat32(key4)
	cnt.E = c.GetFloat32(key5)
	cnt.F = c.GetFloat32(key6)
	cnt.G = c.GetFloat32(key7)
	cnt.H = c.GetFloat32(key8)
	return
}

func (c *BaseController) GetFloat32Tuple9(key1, key2, key3, key4, key5, key6, key7, key8, key9 string) (cnt xgeneric.Tuple9[float32, float32, float32, float32, float32, float32, float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	cnt.D = c.GetFloat32(key4)
	cnt.E = c.GetFloat32(key5)
	cnt.F = c.GetFloat32(key6)
	cnt.G = c.GetFloat32(key7)
	cnt.H = c.GetFloat32(key8)
	cnt.I = c.GetFloat32(key9)
	return
}

func (c *BaseController) GetFloat32Tuple10(key1, key2, key3, key4, key5, key6, key7, key8, key9, key10 string) (cnt xgeneric.Tuple10[float32, float32, float32, float32, float32, float32, float32, float32, float32, float32]) {
	cnt.A = c.GetFloat32(key1)
	cnt.B = c.GetFloat32(key2)
	cnt.C = c.GetFloat32(key3)
	cnt.D = c.GetFloat32(key4)
	cnt.E = c.GetFloat32(key5)
	cnt.F = c.GetFloat32(key6)
	cnt.G = c.GetFloat32(key7)
	cnt.H = c.GetFloat32(key8)
	cnt.I = c.GetFloat32(key9)
	cnt.J = c.GetFloat32(key10)
	return
}

func (c *BaseController) GetFloat64Tuple2(key1, key2 string) (cnt xgeneric.Tuple2[float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	return
}

func (c *BaseController) GetFloat64Tuple3(key1, key2, key3 string) (cnt xgeneric.Tuple3[float64, float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	cnt.C = c.GetFloat64(key3)
	return
}

func (c *BaseController) GetFloat64Tuple4(key1, key2, key3, key4 string) (cnt xgeneric.Tuple4[float64, float64, float64, float64]) {
	tmp := c.GetFloat64Tuple3(key1, key2, key3)
	cnt.A = tmp.A
	cnt.B = tmp.B
	cnt.C = tmp.C
	cnt.D = c.GetFloat64(key4)
	return
}

func (c *BaseController) GetFloat64Tuple5(key1, key2, key3, key4, key5 string) (cnt xgeneric.Tuple5[float64, float64, float64, float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	cnt.C = c.GetFloat64(key3)
	cnt.D = c.GetFloat64(key4)
	cnt.E = c.GetFloat64(key5)
	return
}

func (c *BaseController) GetFloat64Tuple6(key1, key2, key3, key4, key5, key6 string) (cnt xgeneric.Tuple6[float64, float64, float64, float64, float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	cnt.C = c.GetFloat64(key3)
	cnt.D = c.GetFloat64(key4)
	cnt.E = c.GetFloat64(key5)
	cnt.F = c.GetFloat64(key6)
	return
}

func (c *BaseController) GetFloat64Tuple7(key1, key2, key3, key4, key5, key6, key7 string) (cnt xgeneric.Tuple7[float64, float64, float64, float64, float64, float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	cnt.C = c.GetFloat64(key3)
	cnt.D = c.GetFloat64(key4)
	cnt.E = c.GetFloat64(key5)
	cnt.F = c.GetFloat64(key6)
	cnt.G = c.GetFloat64(key7)
	return
}

func (c *BaseController) GetFloat64Tuple8(key1, key2, key3, key4, key5, key6, key7, key8 string) (cnt xgeneric.Tuple8[float64, float64, float64, float64, float64, float64, float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	cnt.C = c.GetFloat64(key3)
	cnt.D = c.GetFloat64(key4)
	cnt.E = c.GetFloat64(key5)
	cnt.F = c.GetFloat64(key6)
	cnt.G = c.GetFloat64(key7)
	cnt.H = c.GetFloat64(key8)
	return
}

func (c *BaseController) GetFloat64Tuple9(key1, key2, key3, key4, key5, key6, key7, key8, key9 string) (cnt xgeneric.Tuple9[float64, float64, float64, float64, float64, float64, float64, float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	cnt.C = c.GetFloat64(key3)
	cnt.D = c.GetFloat64(key4)
	cnt.E = c.GetFloat64(key5)
	cnt.F = c.GetFloat64(key6)
	cnt.G = c.GetFloat64(key7)
	cnt.H = c.GetFloat64(key8)
	cnt.I = c.GetFloat64(key9)
	return
}

func (c *BaseController) GetFloat64Tuple10(key1, key2, key3, key4, key5, key6, key7, key8, key9, key10 string) (cnt xgeneric.Tuple10[float64, float64, float64, float64, float64, float64, float64, float64, float64, float64]) {
	cnt.A = c.GetFloat64(key1)
	cnt.B = c.GetFloat64(key2)
	cnt.C = c.GetFloat64(key3)
	cnt.D = c.GetFloat64(key4)
	cnt.E = c.GetFloat64(key5)
	cnt.F = c.GetFloat64(key6)
	cnt.G = c.GetFloat64(key7)
	cnt.H = c.GetFloat64(key8)
	cnt.I = c.GetFloat64(key9)
	return
}
