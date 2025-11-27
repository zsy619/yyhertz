package context

import "sync"

// ============= 核心数据访问方法（优化版本） =============

// Set 设置键值对 - 使用sync.Map优化并发性能
func (ctx *Context) Set(key string, value any) {
	ctx.keys.Store(key, value)
}

// Get 获取值 - 使用sync.Map优化并发性能
func (ctx *Context) Get(key string) (any, bool) {
	return ctx.keys.Load(key)
}

// MustGet 必须获取值
func (ctx *Context) MustGet(key string) any {
	if value, exists := ctx.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// ============= 增强Keys操作（批量操作） =============

// SetMultiple 批量设置键值对
func (ctx *Context) SetMultiple(pairs map[string]any) {
	for key, value := range pairs {
		ctx.keys.Store(key, value)
	}
}

// GetMultiple 批量获取值
func (ctx *Context) GetMultiple(keys []string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, exists := ctx.keys.Load(key); exists {
			result[key] = value
		}
	}
	return result
}

// DeleteMultiple 批量删除键
func (ctx *Context) DeleteMultiple(keys []string) {
	for _, key := range keys {
		ctx.keys.Delete(key)
	}
}

// ============= 增强Keys操作（获取内部结构） =============

func (ctx *Context) GetKeys() *sync.Map {
	return &ctx.keys
}

// ============= 增强Keys操作（条件操作） =============

// SetIfNotExists 仅当键不存在时设置值
func (ctx *Context) SetIfNotExists(key string, value any) bool {
	_, loaded := ctx.keys.LoadOrStore(key, value)
	return !loaded // 返回true表示成功设置，false表示键已存在
}

// GetOrSet 获取值，如果不存在则设置默认值
func (ctx *Context) GetOrSet(key string, defaultValue any) any {
	actual, _ := ctx.keys.LoadOrStore(key, defaultValue)
	return actual
}

// CompareAndSwap 比较并交换值（原子操作）
func (ctx *Context) CompareAndSwap(key string, oldValue, newValue any) bool {
	return ctx.keys.CompareAndSwap(key, oldValue, newValue)
}

// Delete 删除键
func (ctx *Context) Delete(key string) {
	ctx.keys.Delete(key)
}

// Exists 检查键是否存在
func (ctx *Context) Exists(key string) bool {
	_, exists := ctx.keys.Load(key)
	return exists
}

// ============= 增强Keys操作（遍历和过滤） =============

// ForEach 遍历所有键值对
func (ctx *Context) ForEach(fn func(key string, value any) bool) {
	ctx.keys.Range(func(k, v any) bool {
		key, ok := k.(string)
		if !ok {
			return true // 跳过非字符串键
		}
		return fn(key, v)
	})
}

// Filter 过滤键值对，返回满足条件的键列表
func (ctx *Context) Filter(predicate func(key string, value any) bool) []string {
	var result []string
	ctx.keys.Range(func(k, v any) bool {
		key, ok := k.(string)
		if !ok {
			return true
		}
		if predicate(key, v) {
			result = append(result, key)
		}
		return true
	})
	return result
}

// MapKeys 对所有键值对执行映射操作
func (ctx *Context) MapKeys(mapper func(key string, value any) any) map[string]any {
	result := make(map[string]any)
	ctx.keys.Range(func(k, v any) bool {
		key, ok := k.(string)
		if !ok {
			return true
		}
		result[key] = mapper(key, v)
		return true
	})
	return result
}

// Clear 清空所有键值对
func (ctx *Context) Clear() {
	ctx.keys.Range(func(key, value any) bool {
		ctx.keys.Delete(key)
		return true
	})
}
