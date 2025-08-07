package context

// Params 路由参数
type Params []Param

// Param 单个参数
type Param struct {
	Key   string
	Value string
}

// ByName 根据名称获取参数值
func (ps Params) ByName(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}