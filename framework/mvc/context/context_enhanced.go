package context

import (
	"sync"
	"time"
)

// ============= 增强型键值操作扩展 =============

// SetWithExpiry 设置带过期时间的键值对
func (ctx *Context) SetWithExpiry(key string, value any, ttl time.Duration) {
	expiryData := &ExpiryValue{
		Value:      value,
		ExpiryTime: time.Now().Add(ttl),
	}
	ctx.keys.Store(key, expiryData)
}

// GetWithExpiry 获取可能过期的值
func (ctx *Context) GetWithExpiry(key string) (any, bool) {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return nil, false
	}
	
	if expiryValue, ok := value.(*ExpiryValue); ok {
		if time.Now().After(expiryValue.ExpiryTime) {
			ctx.keys.Delete(key) // 删除过期值
			return nil, false
		}
		return expiryValue.Value, true
	}
	
	// 普通值，直接返回
	return value, true
}

// ExpiryValue 带过期时间的值
type ExpiryValue struct {
	Value      any       `json:"value"`
	ExpiryTime time.Time `json:"expiry_time"`
}

// ============= 并发安全的计数器 =============

// IncrementCounter 增加计数器
func (ctx *Context) IncrementCounter(key string) int64 {
	for {
		value, loaded := ctx.keys.LoadOrStore(key, int64(1))
		if !loaded {
			return 1
		}
		
		if counter, ok := value.(int64); ok {
			newValue := counter + 1
			if ctx.keys.CompareAndSwap(key, counter, newValue) {
				return newValue
			}
		} else {
			// 类型不匹配，重置为1
			ctx.keys.Store(key, int64(1))
			return 1
		}
	}
}

// DecrementCounter 减少计数器
func (ctx *Context) DecrementCounter(key string) int64 {
	for {
		value, loaded := ctx.keys.LoadOrStore(key, int64(0))
		if !loaded {
			return 0
		}
		
		if counter, ok := value.(int64); ok {
			newValue := counter - 1
			if newValue < 0 {
				newValue = 0
			}
			if ctx.keys.CompareAndSwap(key, counter, newValue) {
				return newValue
			}
		} else {
			// 类型不匹配，重置为0
			ctx.keys.Store(key, int64(0))
			return 0
		}
	}
}

// GetCounter 获取计数器值
func (ctx *Context) GetCounter(key string) int64 {
	value, exists := ctx.keys.Load(key)
	if !exists {
		return 0
	}
	
	if counter, ok := value.(int64); ok {
		return counter
	}
	return 0
}

// ============= 链式操作支持 =============

// Chain 链式操作构建器
type Chain struct {
	ctx     *Context
	success bool
	lastErr error
}

// NewChain 创建新的链式操作
func (ctx *Context) NewChain() *Chain {
	return &Chain{
		ctx:     ctx,
		success: true,
	}
}

// Set 链式设置值
func (c *Chain) Set(key string, value any) *Chain {
	if c.success {
		c.ctx.Set(key, value)
	}
	return c
}

// SetIf 条件设置值
func (c *Chain) SetIf(condition bool, key string, value any) *Chain {
	if c.success && condition {
		c.ctx.Set(key, value)
	}
	return c
}

// SetMultiple 链式批量设置
func (c *Chain) SetMultiple(pairs map[string]any) *Chain {
	if c.success {
		c.ctx.SetMultiple(pairs)
	}
	return c
}

// Validate 验证操作
func (c *Chain) Validate(validator func(*Context) error) *Chain {
	if c.success {
		if err := validator(c.ctx); err != nil {
			c.success = false
			c.lastErr = err
		}
	}
	return c
}

// Execute 执行操作
func (c *Chain) Execute(operation func(*Context) error) *Chain {
	if c.success {
		if err := operation(c.ctx); err != nil {
			c.success = false
			c.lastErr = err
		}
	}
	return c
}

// Result 获取链式操作结果
func (c *Chain) Result() (bool, error) {
	return c.success, c.lastErr
}

// ============= 事务性操作支持 =============

// Transaction 事务性操作
type Transaction struct {
	ctx     *Context
	backup  map[string]any
	mutex   sync.RWMutex
	active  bool
}

// BeginTransaction 开始事务
func (ctx *Context) BeginTransaction() *Transaction {
	tx := &Transaction{
		ctx:    ctx,
		backup: make(map[string]any),
		active: true,
	}
	
	// 备份当前状态
	ctx.keys.Range(func(key, value any) bool {
		if keyStr, ok := key.(string); ok {
			tx.backup[keyStr] = value
		}
		return true
	})
	
	return tx
}

// Set 事务性设置值
func (tx *Transaction) Set(key string, value any) *Transaction {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()
	
	if tx.active {
		tx.ctx.Set(key, value)
	}
	return tx
}

// Delete 事务性删除值
func (tx *Transaction) Delete(key string) *Transaction {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()
	
	if tx.active {
		tx.ctx.Delete(key)
	}
	return tx
}

// Commit 提交事务
func (tx *Transaction) Commit() {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()
	
	tx.active = false
	tx.backup = nil // 清理备份
}

// Rollback 回滚事务
func (tx *Transaction) Rollback() {
	tx.mutex.Lock()
	defer tx.mutex.Unlock()
	
	if tx.active {
		// 清空当前状态
		tx.ctx.Clear()
		
		// 恢复备份状态
		for key, value := range tx.backup {
			tx.ctx.Set(key, value)
		}
		
		tx.active = false
		tx.backup = nil
	}
}

// ============= 观察者模式支持 =============

// Observer 观察者接口
type Observer interface {
	OnKeyChanged(ctx *Context, key string, oldValue, newValue any)
	OnKeyDeleted(ctx *Context, key string, oldValue any)
}

// ObserverManager 观察者管理器
type ObserverManager struct {
	observers map[string][]Observer
	mutex     sync.RWMutex
}

// AddObserver 添加观察者
func (ctx *Context) AddObserver(key string, observer Observer) {
	manager := ctx.getObserverManager()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	if manager.observers == nil {
		manager.observers = make(map[string][]Observer)
	}
	
	manager.observers[key] = append(manager.observers[key], observer)
}

// RemoveObserver 移除观察者
func (ctx *Context) RemoveObserver(key string, observer Observer) {
	manager := ctx.getObserverManager()
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	
	if observers, exists := manager.observers[key]; exists {
		for i, obs := range observers {
			if obs == observer {
				manager.observers[key] = append(observers[:i], observers[i+1:]...)
				break
			}
		}
	}
}

// SetWithNotify 设置值并通知观察者
func (ctx *Context) SetWithNotify(key string, value any) {
	oldValue, hasOldValue := ctx.Get(key)
	ctx.Set(key, value)
	
	manager := ctx.getObserverManager()
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	
	if observers, exists := manager.observers[key]; exists {
		var old any
		if hasOldValue {
			old = oldValue
		}
		
		for _, observer := range observers {
			observer.OnKeyChanged(ctx, key, old, value)
		}
	}
}

// DeleteWithNotify 删除值并通知观察者
func (ctx *Context) DeleteWithNotify(key string) {
	oldValue, exists := ctx.Get(key)
	if exists {
		ctx.Delete(key)
		
		manager := ctx.getObserverManager()
		manager.mutex.RLock()
		defer manager.mutex.RUnlock()
		
		if observers, observersExist := manager.observers[key]; observersExist {
			for _, observer := range observers {
				observer.OnKeyDeleted(ctx, key, oldValue)
			}
		}
	}
}

// getObserverManager 获取观察者管理器
func (ctx *Context) getObserverManager() *ObserverManager {
	const observerKey = "__observer_manager__"
	
	value, exists := ctx.keys.Load(observerKey)
	if !exists {
		manager := &ObserverManager{}
		actual, loaded := ctx.keys.LoadOrStore(observerKey, manager)
		if loaded {
			return actual.(*ObserverManager)
		}
		return manager
	}
	
	return value.(*ObserverManager)
}

// ============= 缓存策略支持 =============

// CacheStrategy 缓存策略接口
type CacheStrategy interface {
	ShouldCache(key string, value any) bool
	GetTTL(key string, value any) time.Duration
}

// SimpleCacheStrategy 简单缓存策略
type SimpleCacheStrategy struct {
	DefaultTTL time.Duration
	MaxSize    int
}

// ShouldCache 是否应该缓存
func (s *SimpleCacheStrategy) ShouldCache(key string, value any) bool {
	return true // 简单策略：总是缓存
}

// GetTTL 获取TTL
func (s *SimpleCacheStrategy) GetTTL(key string, value any) time.Duration {
	return s.DefaultTTL
}

// SetWithCache 使用缓存策略设置值
func (ctx *Context) SetWithCache(key string, value any, strategy CacheStrategy) {
	if strategy.ShouldCache(key, value) {
		ttl := strategy.GetTTL(key, value)
		if ttl > 0 {
			ctx.SetWithExpiry(key, value, ttl)
		} else {
			ctx.Set(key, value)
		}
	}
}

// ============= 性能优化的批量操作 =============

// BatchOperation 批量操作
type BatchOperation struct {
	ctx        *Context
	operations []func()
	mutex      sync.Mutex
}

// NewBatch 创建新的批量操作
func (ctx *Context) NewBatch() *BatchOperation {
	return &BatchOperation{
		ctx:        ctx,
		operations: make([]func(), 0),
	}
}

// AddSet 添加设置操作
func (b *BatchOperation) AddSet(key string, value any) *BatchOperation {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.operations = append(b.operations, func() {
		b.ctx.Set(key, value)
	})
	return b
}

// AddDelete 添加删除操作
func (b *BatchOperation) AddDelete(key string) *BatchOperation {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	b.operations = append(b.operations, func() {
		b.ctx.Delete(key)
	})
	return b
}

// Execute 执行批量操作
func (b *BatchOperation) Execute() {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	
	for _, operation := range b.operations {
		operation()
	}
	
	// 清空操作列表
	b.operations = b.operations[:0]
}

// ============= 调试和诊断增强 =============

// ContextSnapshot Context快照
type ContextSnapshot struct {
	Timestamp time.Time          `json:"timestamp"`
	Keys      map[string]any     `json:"keys"`
	Params    map[string]string  `json:"params"`
	Errors    []string           `json:"errors"`
	KeysCount int               `json:"keys_count"`
	Age       time.Duration     `json:"age"`
}

// TakeSnapshot 创建Context快照
func (ctx *Context) TakeSnapshot() *ContextSnapshot {
	snapshot := &ContextSnapshot{
		Timestamp: time.Now(),
		Keys:      make(map[string]any),
		Params:    ctx.ParamMap(),
		Errors:    make([]string, 0),
	}
	
	// 复制键值对
	ctx.keys.Range(func(key, value any) bool {
		if keyStr, ok := key.(string); ok {
			// 跳过内部管理键
			if keyStr != "__observer_manager__" {
				snapshot.Keys[keyStr] = value
			}
		}
		return true
	})
	
	// 复制错误
	for _, err := range ctx.GetErrors() {
		snapshot.Errors = append(snapshot.Errors, err.Error())
	}
	
	// 获取统计信息
	snapshot.KeysCount = ctx.KeysCount()
	snapshot.Age = time.Since(ctx.Acquired())
	
	return snapshot
}