package core

import (
	"strings"

	"haedu.gov.cn/tools/xgeneric"
	"haedu.gov.cn/tools/xstring"
)

func (c *BaseController) GetStringTrim(key string, def ...string) string {
	outStr := c.GetString(key, def...)
	outStr = strings.TrimSpace(outStr)
	return outStr
}

// GetSafeString 获取安全字符串
func (c *BaseController) GetSafeString(key string, def ...string) string {
	data := c.GetString(key, def...)
	if data == "" {
		return ""
	}
	return xstring.GetSafeString(data)
}

func (c *BaseController) GetSafeStringTuple2(key1, key2 string) (cnt xgeneric.Tuple2[string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	return
}

func (c *BaseController) GetSafeStringTuple3(key1, key2, key3 string) (cnt xgeneric.Tuple3[string, string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	cnt.C = c.GetSafeString(key3)
	return
}

func (c *BaseController) GetSafeStringTuple4(key1, key2, key3, key4 string) (cnt xgeneric.Tuple4[string, string, string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	cnt.C = c.GetSafeString(key3)
	cnt.D = c.GetSafeString(key4)
	return
}

func (c *BaseController) GetSafeStringTuple5(key1, key2, key3, key4, key5 string) (cnt xgeneric.Tuple5[string, string, string, string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	cnt.C = c.GetSafeString(key3)
	cnt.D = c.GetSafeString(key4)
	cnt.E = c.GetSafeString(key5)
	return
}

func (c *BaseController) GetSafeStringTuple6(key1, key2, key3, key4, key5, key6 string) (cnt xgeneric.Tuple6[string, string, string, string, string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	cnt.C = c.GetSafeString(key3)
	cnt.D = c.GetSafeString(key4)
	cnt.E = c.GetSafeString(key5)
	cnt.F = c.GetSafeString(key6)
	return
}

func (c *BaseController) GetSafeStringTuple7(key1, key2, key3, key4, key5, key6, key7 string) (cnt xgeneric.Tuple7[string, string, string, string, string, string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	cnt.C = c.GetSafeString(key3)
	cnt.D = c.GetSafeString(key4)
	cnt.E = c.GetSafeString(key5)
	cnt.F = c.GetSafeString(key6)
	cnt.G = c.GetSafeString(key7)
	return
}

func (c *BaseController) GetSafeStringTuple8(key1, key2, key3, key4, key5, key6, key7, key8 string) (cnt xgeneric.Tuple8[string, string, string, string, string, string, string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	cnt.C = c.GetSafeString(key3)
	cnt.D = c.GetSafeString(key4)
	cnt.E = c.GetSafeString(key5)
	cnt.F = c.GetSafeString(key6)
	cnt.G = c.GetSafeString(key7)
	cnt.H = c.GetSafeString(key8)
	return
}

func (c *BaseController) GetSafeStringTuple9(key1, key2, key3, key4, key5, key6, key7, key8, key9 string) (cnt xgeneric.Tuple9[string, string, string, string, string, string, string, string, string]) {
	cnt.A = c.GetSafeString(key1)
	cnt.B = c.GetSafeString(key2)
	cnt.C = c.GetSafeString(key3)
	cnt.D = c.GetSafeString(key4)
	cnt.E = c.GetSafeString(key5)
	cnt.F = c.GetSafeString(key6)
	cnt.G = c.GetSafeString(key7)
	cnt.H = c.GetSafeString(key8)
	cnt.I = c.GetSafeString(key9)
	return
}
