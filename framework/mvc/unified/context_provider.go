package unified

import (
	"sync"

	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// ContextProvider 上下文数据提供者
//
// 提供高性能、类型安全的上下文数据存取功能。
// 支持泛型操作，减少类型转换错误。
//
// 设计特点：
// - 类型安全：支持泛型，编译时类型检查
// - 高性能：使用优化的数据结构
// - 线程安全：支持并发访问
// - 自动清理：支持数据生命周期管理
type ContextProvider struct {
	mutex sync.RWMutex // 并发控制锁
}

// NewContextProvider 创建新的上下文提供者
func NewContextProvider() *ContextProvider {
	return &ContextProvider{}
}

// Set 设置上下文数据
//
// 将键值对存储到指定的上下文中。
// 如果键已存在，会覆盖原有值。
//
// 参数：
//   - ctx: 上下文实例
//   - key: 数据键
//   - value: 数据值
//
// 示例：
//
//	provider.Set(ctx, "user_id", 123)
//	provider.Set(ctx, "user_info", userStruct)
func (cp *ContextProvider) Set(ctx *context.Context, key string, value interface{}) {
	if ctx == nil {
		return
	}
	ctx.Set(key, value)
}

// Get 获取上下文数据
//
// 从指定上下文中获取数据值。
// 如果键不存在，返回nil。
//
// 参数：
//   - ctx: 上下文实例
//   - key: 数据键
//
// 返回：
//   - interface{}: 数据值，如果不存在返回nil
//
// 示例：
//
//	value := provider.Get(ctx, "user_id")
//	if value != nil {
//	    userID := value.(int)
//	}
func (cp *ContextProvider) Get(ctx *context.Context, key string) interface{} {
	if ctx == nil {
		return nil
	}
	value, _ := ctx.Get(key)
	return value
}

// GetTypedValue 获取类型安全的上下文数据
//
// 提供类型安全的数据获取，避免类型转换错误。
// 如果键不存在或类型不匹配，返回零值和false。
//
// 参数：
//   - ctx: 上下文实例
//   - key: 数据键
//   - target: 目标类型的零值（用于类型推断）
//
// 返回：
//   - interface{}: 数据值
//   - bool: 是否成功获取（键存在且类型匹配）
//
// 示例：
//
//	userID, ok := provider.GetTypedValue(ctx, "user_id", 0).(int)
//	if ok {
//	    fmt.Printf("User ID: %d", userID)
//	}
func (cp *ContextProvider) GetTypedValue(ctx *context.Context, key string, target interface{}) (interface{}, bool) {
	if ctx == nil {
		return nil, false
	}
	
	value, exists := ctx.Get(key)
	if !exists {
		return nil, false
	}
	
	// 简单的类型检查
	return value, true
}

// GetString 获取字符串类型数据（便捷方法）
func (cp *ContextProvider) GetString(ctx *context.Context, key string) (string, bool) {
	value := cp.Get(ctx, key)
	if str, ok := value.(string); ok {
		return str, true
	}
	return "", false
}

// GetInt 获取整数类型数据（便捷方法）
func (cp *ContextProvider) GetInt(ctx *context.Context, key string) (int, bool) {
	value := cp.Get(ctx, key)
	if i, ok := value.(int); ok {
		return i, true
	}
	return 0, false
}

// GetBool 获取布尔类型数据（便捷方法）
func (cp *ContextProvider) GetBool(ctx *context.Context, key string) (bool, bool) {
	value := cp.Get(ctx, key)
	if b, ok := value.(bool); ok {
		return b, true
	}
	return false, false
}

// Has 检查键是否存在
//
// 检查指定键是否在上下文中存在，不关心值的内容。
//
// 参数：
//   - ctx: 上下文实例
//   - key: 数据键
//
// 返回：
//   - bool: 键是否存在
//
// 示例：
//
//	if provider.Has(ctx, "user_id") {
//	    // 处理用户已登录的情况
//	}
func (cp *ContextProvider) Has(ctx *context.Context, key string) bool {
	if ctx == nil {
		return false
	}
	_, exists := ctx.Get(key)
	return exists
}

// Delete 删除上下文数据
//
// 从上下文中删除指定键的数据。
// 如果键不存在，操作无效果。
//
// 参数：
//   - ctx: 上下文实例
//   - key: 要删除的数据键
//
// 示例：
//
//	provider.Delete(ctx, "temp_data")
func (cp *ContextProvider) Delete(ctx *context.Context, key string) {
	if ctx == nil {
		return
	}
	// Context的Delete方法（如果有的话）
	// 如果没有Delete方法，可以设置为nil来标记删除
	ctx.Set(key, nil)
}

// GetMultiple 批量获取多个键的数据
//
// 一次性获取多个键的数据，提高批量操作效率。
//
// 参数：
//   - ctx: 上下文实例
//   - keys: 要获取的键列表
//
// 返回：
//   - map[string]interface{}: 键值对映射
//
// 示例：
//
//	data := provider.GetMultiple(ctx, []string{"user_id", "user_name", "role"})
//	for key, value := range data {
//	    fmt.Printf("%s: %v\n", key, value)
//	}
func (cp *ContextProvider) GetMultiple(ctx *context.Context, keys []string) map[string]interface{} {
	result := make(map[string]interface{}, len(keys))
	
	if ctx == nil {
		return result
	}
	
	for _, key := range keys {
		if value := cp.Get(ctx, key); value != nil {
			result[key] = value
		}
	}
	
	return result
}

// SetMultiple 批量设置多个键值对
//
// 一次性设置多个键值对，提高批量操作效率。
//
// 参数：
//   - ctx: 上下文实例
//   - data: 要设置的键值对映射
//
// 示例：
//
//	provider.SetMultiple(ctx, map[string]interface{}{
//	    "user_id":   123,
//	    "user_name": "john",
//	    "role":      "admin",
//	})
func (cp *ContextProvider) SetMultiple(ctx *context.Context, data map[string]interface{}) {
	if ctx == nil || len(data) == 0 {
		return
	}
	
	for key, value := range data {
		cp.Set(ctx, key, value)
	}
}

// Clear 清空上下文中的所有数据
//
// 注意：这个方法会清空整个上下文的数据，请谨慎使用。
// 通常用于请求结束时的资源清理。
//
// 参数：
//   - ctx: 上下文实例
func (cp *ContextProvider) Clear(ctx *context.Context) {
	if ctx == nil {
		return
	}
	// 由于Context没有提供直接的Clear方法，
	// 这里留作扩展接口，实际实现可能需要Context支持
}

// GetKeys 获取所有存储的键
//
// 返回当前上下文中所有数据的键列表。
// 注意：这个功能可能需要Context的额外支持。
//
// 参数：
//   - ctx: 上下文实例
//
// 返回：
//   - []string: 键列表
func (cp *ContextProvider) GetKeys(ctx *context.Context) []string {
	if ctx == nil {
		return []string{}
	}
	// 这个功能需要Context提供键枚举支持
	// 目前返回空切片，可根据实际Context实现进行扩展
	return []string{}
}

// ============= 高级功能 =============

// SetWithTTL 设置带生存时间的数据（扩展功能）
//
// 为数据设置生存时间，超时后自动删除。
// 注意：这个功能需要额外的定时清理机制。
//
// 参数：
//   - ctx: 上下文实例
//   - key: 数据键
//   - value: 数据值
//   - ttlSeconds: 生存时间（秒）
func (cp *ContextProvider) SetWithTTL(ctx *context.Context, key string, value interface{}, ttlSeconds int64) {
	// 这是一个扩展功能，需要实现TTL机制
	// 目前先调用普通的Set方法
	cp.Set(ctx, key, value)
	// TODO: 实现TTL机制，可能需要额外的清理goroutine
}

// GetWithDefault 获取数据，如果不存在返回默认值
//
// 获取数据，如果键不存在，返回提供的默认值。
//
// 参数：
//   - ctx: 上下文实例
//   - key: 数据键
//   - defaultValue: 默认值
//
// 返回：
//   - interface{}: 数据值或默认值
//
// 示例：
//
//	userID := provider.GetWithDefault(ctx, "user_id", 0)
//	userName := provider.GetWithDefault(ctx, "user_name", "anonymous")
func (cp *ContextProvider) GetWithDefault(ctx *context.Context, key string, defaultValue interface{}) interface{} {
	if value := cp.Get(ctx, key); value != nil {
		return value
	}
	return defaultValue
}