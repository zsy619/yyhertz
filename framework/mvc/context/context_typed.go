package context

// ============= 增强Keys操作（类型安全） =============

// SetString 设置字符串类型值（类型安全）
func (ctx *Context) SetString(key string, value string) {
	ctx.keys.Store(key, value)
}

// GetString 获取字符串类型值（类型安全）
func (ctx *Context) GetString(key string) (string, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// SetTypedInt 设置整数类型值（类型安全）
func (ctx *Context) SetInt(key string, value int) {
	ctx.keys.Store(key, value)
}

// GetTypedInt 获取整数类型值（类型安全）
func (ctx *Context) GetInt(key string) (int, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return 0, false
	}
	intVal, ok := value.(int)
	return intVal, ok
}

// SetBool 设置布尔类型值（类型安全）
func (ctx *Context) SetBool(key string, value bool) {
	ctx.keys.Store(key, value)
}

// GetBool 获取布尔类型值（类型安全）
func (ctx *Context) GetBool(key string) (bool, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return false, false
	}
	boolVal, ok := value.(bool)
	return boolVal, ok
}

// SetFloat64 设置浮点数类型值（类型安全）
func (ctx *Context) SetFloat64(key string, value float64) {
	ctx.keys.Store(key, value)
}

// GetFloat64 获取浮点数类型值（类型安全）
func (ctx *Context) GetFloat64(key string) (float64, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return 0, false
	}
	floatVal, ok := value.(float64)
	return floatVal, ok
}

func (ctx *Context) SetInt64(key string, value int64) {
	ctx.keys.Store(key, value)
}

func (ctx *Context) GetInt64(key string) (int64, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return 0, false
	}
	int64Val, ok := value.(int64)
	return int64Val, ok
}

// ============= 类型断言辅助方法 =============

// GetString 获取字符串类型值
func (ctx *Context) GetStringValue(key string) (string, bool) {
	value, exists := ctx.Get(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetInt 获取整数类型值
func (ctx *Context) GetIntValue(key string) (int, bool) {
	value, exists := ctx.Get(key)
	if !exists {
		return 0, false
	}
	intVal, ok := value.(int)
	return intVal, ok
}

// GetBool 获取布尔类型值
func (ctx *Context) GetBoolValue(key string) (bool, bool) {
	value, exists := ctx.Get(key)
	if !exists {
		return false, false
	}
	boolVal, ok := value.(bool)
	return boolVal, ok
}