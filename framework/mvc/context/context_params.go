package context

// ============= 兼容性访问器方法 =============
// 这些方法提供向后兼容性，允许现有代码无缝迁移

// Params 获取路由参数
func (ctx *Context) Params() Params {
	return ctx.params
}

// SetParams 设置路由参数
func (ctx *Context) SetParams(params Params) {
	ctx.params = params
}

// ParamKeys 获取所有路由参数的键名
func (ctx *Context) ParamKeys() []string {
	if len(ctx.params) == 0 {
		return nil
	}

	keys := make([]string, len(ctx.params))
	for i, param := range ctx.params {
		keys[i] = param.Key
	}
	return keys
}

// ParamMap 将路由参数转换为map[string]string
func (ctx *Context) ParamMap() map[string]string {
	if len(ctx.params) == 0 {
		return make(map[string]string)
	}

	paramMap := make(map[string]string, len(ctx.params))
	for _, param := range ctx.params {
		paramMap[param.Key] = param.Value
	}
	return paramMap
}

// ParamValues 获取所有路由参数的值
func (ctx *Context) ParamValues() []string {
	if len(ctx.params) == 0 {
		return nil
	}

	values := make([]string, len(ctx.params))
	for i, param := range ctx.params {
		values[i] = param.Value
	}
	return values
}

// FullPath 获取完整路径
func (ctx *Context) FullPath() string {
	return ctx.fullPath
}

// SetFullPath 设置完整路径
func (ctx *Context) SetFullPath(path string) {
	ctx.fullPath = path
}