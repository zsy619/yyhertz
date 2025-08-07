package controller

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	mvcContext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// LifecycleManager 控制器生命周期管理器
type LifecycleManager struct {
	pools       sync.Map                    // 控制器池映射
	config      *CompilerConfig            // 配置
	hooks       map[LifecycleHook][]HookFunc // 生命周期钩子
	metrics     *LifecycleMetrics          // 生命周期指标
	mu          sync.RWMutex               // 读写锁
}

// LifecycleHook 生命周期钩子类型
type LifecycleHook int

const (
	HookBeforeCreate LifecycleHook = iota // 创建前
	HookAfterCreate                       // 创建后
	HookBeforeInit                        // 初始化前
	HookAfterInit                         // 初始化后
	HookBeforeDestroy                     // 销毁前
	HookAfterDestroy                      // 销毁后
)

// HookFunc 钩子函数类型
type HookFunc func(controller interface{}, ctx *mvcContext.Context) error

// LifecycleMetrics 生命周期指标
type LifecycleMetrics struct {
	CreatedCount   int64         // 创建数量
	DestroyedCount int64         // 销毁数量
	ActiveCount    int64         // 活跃数量
	PoolHitRate    float64       // 池命中率
	AverageLifetime time.Duration // 平均生命周期
	mu             sync.RWMutex  // 指标锁
}

// ControllerPool 控制器池
type ControllerPool struct {
	factory    ControllerFactory           // 控制器工厂
	pool       chan interface{}            // 控制器池
	controllerType reflect.Type            // 控制器类型
	maxSize    int                        // 最大池大小
	created    int64                      // 已创建数量
	borrowed   int64                      // 已借出数量
	returned   int64                      // 已归还数量
	mu         sync.RWMutex               // 池锁
}

// ControllerFactory 控制器工厂接口
type ControllerFactory interface {
	CreateController() (interface{}, error)
	InitController(controller interface{}, ctx *mvcContext.Context) error
	DestroyController(controller interface{}) error
}

// DefaultControllerFactory 默认控制器工厂
type DefaultControllerFactory struct {
	controllerType reflect.Type
	lifecycle      *LifecycleManager
}

// ControllerInstance 控制器实例
type ControllerInstance struct {
	Controller interface{}   // 控制器对象
	CreatedAt  time.Time     // 创建时间
	LastUsed   time.Time     // 最后使用时间
	UsageCount int64         // 使用次数
	Pooled     bool          // 是否来自池
}

// NewLifecycleManager 创建生命周期管理器
func NewLifecycleManager(config *CompilerConfig) *LifecycleManager {
	manager := &LifecycleManager{
		config:  config,
		hooks:   make(map[LifecycleHook][]HookFunc),
		metrics: &LifecycleMetrics{},
	}

	// 启动清理协程
	go manager.startCleanupRoutine()

	return manager
}

// NewControllerPool 创建控制器池
func NewControllerPool(controllerType reflect.Type, maxSize int) *ControllerPool {
	return &ControllerPool{
		pool:           make(chan interface{}, maxSize),
		controllerType: controllerType,
		maxSize:        maxSize,
	}
}

// NewDefaultControllerFactory 创建默认控制器工厂
func NewDefaultControllerFactory(controllerType reflect.Type, lifecycle *LifecycleManager) *DefaultControllerFactory {
	return &DefaultControllerFactory{
		controllerType: controllerType,
		lifecycle:      lifecycle,
	}
}

// CreateController 创建控制器实例
func (lm *LifecycleManager) CreateController(controllerType reflect.Type, ctx *mvcContext.Context) (*ControllerInstance, error) {
	// 尝试从池中获取
	if pool, exists := lm.getPool(controllerType); exists {
		if controller := pool.Get(); controller != nil {
			instance := &ControllerInstance{
				Controller: controller,
				LastUsed:   time.Now(),
				Pooled:     true,
			}
			
			// 初始化控制器
			if err := lm.initController(controller, ctx); err != nil {
				pool.Put(controller) // 归还到池
				return nil, fmt.Errorf("failed to initialize controller: %w", err)
			}
			
			lm.metrics.updateActive(1)
			return instance, nil
		}
	}

	// 创建新实例
	controller, err := lm.createNewController(controllerType, ctx)
	if err != nil {
		return nil, err
	}

	instance := &ControllerInstance{
		Controller: controller,
		CreatedAt:  time.Now(),
		LastUsed:   time.Now(),
		UsageCount: 0,
		Pooled:     false,
	}

	lm.metrics.updateCreated(1)
	lm.metrics.updateActive(1)

	return instance, nil
}

// createNewController 创建新的控制器实例
func (lm *LifecycleManager) createNewController(controllerType reflect.Type, ctx *mvcContext.Context) (interface{}, error) {
	// 执行创建前钩子
	if err := lm.executeHooks(HookBeforeCreate, nil, ctx); err != nil {
		return nil, fmt.Errorf("before create hook failed: %w", err)
	}

	// 创建控制器实例
	controllerValue := reflect.New(controllerType)
	controller := controllerValue.Interface()

	// 执行创建后钩子
	if err := lm.executeHooks(HookAfterCreate, controller, ctx); err != nil {
		return nil, fmt.Errorf("after create hook failed: %w", err)
	}

	// 初始化控制器
	if err := lm.initController(controller, ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize controller: %w", err)
	}

	return controller, nil
}

// initController 初始化控制器
func (lm *LifecycleManager) initController(controller interface{}, ctx *mvcContext.Context) error {
	// 执行初始化前钩子
	if err := lm.executeHooks(HookBeforeInit, controller, ctx); err != nil {
		return fmt.Errorf("before init hook failed: %w", err)
	}

	// 执行控制器的初始化方法
	if initializer, ok := controller.(ControllerInitializer); ok {
		if err := initializer.Init(ctx); err != nil {
			return fmt.Errorf("controller init failed: %w", err)
		}
	}

	// 执行初始化后钩子
	if err := lm.executeHooks(HookAfterInit, controller, ctx); err != nil {
		return fmt.Errorf("after init hook failed: %w", err)
	}

	return nil
}

// ReturnController 归还控制器到池
func (lm *LifecycleManager) ReturnController(instance *ControllerInstance) error {
	if !instance.Pooled {
		// 非池化实例直接销毁
		return lm.destroyController(instance.Controller)
	}

	// 获取控制器类型
	controllerType := reflect.TypeOf(instance.Controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	// 归还到池
	if pool, exists := lm.getPool(controllerType); exists {
		// 重置控制器状态
		if resettable, ok := instance.Controller.(ControllerResettable); ok {
			resettable.Reset()
		}

		pool.Put(instance.Controller)
		lm.metrics.updateActive(-1)
		return nil
	}

	// 池不存在，直接销毁
	return lm.destroyController(instance.Controller)
}

// destroyController 销毁控制器
func (lm *LifecycleManager) destroyController(controller interface{}) error {
	// 执行销毁前钩子
	if err := lm.executeHooks(HookBeforeDestroy, controller, nil); err != nil {
		return fmt.Errorf("before destroy hook failed: %w", err)
	}

	// 执行控制器的销毁方法
	if destroyer, ok := controller.(ControllerDestroyer); ok {
		if err := destroyer.Destroy(); err != nil {
			return fmt.Errorf("controller destroy failed: %w", err)
		}
	}

	// 执行销毁后钩子
	if err := lm.executeHooks(HookAfterDestroy, controller, nil); err != nil {
		return fmt.Errorf("after destroy hook failed: %w", err)
	}

	lm.metrics.updateDestroyed(1)
	lm.metrics.updateActive(-1)

	return nil
}

// RegisterHook 注册生命周期钩子
func (lm *LifecycleManager) RegisterHook(hook LifecycleHook, fn HookFunc) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.hooks[hook] == nil {
		lm.hooks[hook] = make([]HookFunc, 0)
	}
	lm.hooks[hook] = append(lm.hooks[hook], fn)
}

// executeHooks 执行生命周期钩子
func (lm *LifecycleManager) executeHooks(hook LifecycleHook, controller interface{}, ctx *mvcContext.Context) error {
	lm.mu.RLock()
	hooks := lm.hooks[hook]
	lm.mu.RUnlock()

	for _, hookFunc := range hooks {
		if err := hookFunc(controller, ctx); err != nil {
			return err
		}
	}

	return nil
}

// getPool 获取控制器池
func (lm *LifecycleManager) getPool(controllerType reflect.Type) (*ControllerPool, bool) {
	if value, exists := lm.pools.Load(controllerType.Name()); exists {
		return value.(*ControllerPool), true
	}
	return nil, false
}

// getOrCreatePool 获取或创建控制器池
func (lm *LifecycleManager) getOrCreatePool(controllerType reflect.Type) *ControllerPool {
	typeName := controllerType.Name()
	
	if value, exists := lm.pools.Load(typeName); exists {
		return value.(*ControllerPool)
	}

	pool := NewControllerPool(controllerType, lm.config.PoolSize)
	pool.factory = NewDefaultControllerFactory(controllerType, lm)
	
	lm.pools.Store(typeName, pool)
	return pool
}

// startCleanupRoutine 启动清理协程
func (lm *LifecycleManager) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			lm.cleanup()
		}
	}
}

// cleanup 清理过期的控制器
func (lm *LifecycleManager) cleanup() {
	lm.pools.Range(func(key, value interface{}) bool {
		pool := value.(*ControllerPool)
		pool.cleanup(lm.config.MaxIdleTime)
		return true
	})
}

// ControllerPool 方法实现

// Get 从池中获取控制器
func (cp *ControllerPool) Get() interface{} {
	select {
	case controller := <-cp.pool:
		cp.mu.Lock()
		cp.borrowed++
		cp.mu.Unlock()
		return controller
	default:
		// 池为空，返回nil让调用方创建新实例
		return nil
	}
}

// Put 将控制器放回池中
func (cp *ControllerPool) Put(controller interface{}) {
	select {
	case cp.pool <- controller:
		cp.mu.Lock()
		cp.returned++
		cp.mu.Unlock()
	default:
		// 池已满，丢弃控制器
	}
}

// cleanup 清理池中的过期控制器
func (cp *ControllerPool) cleanup(maxIdleTime time.Duration) {
	// 这里可以实现更复杂的清理逻辑
	// 例如检查控制器的最后使用时间等
}

// Stats 获取池统计信息
func (cp *ControllerPool) Stats() map[string]interface{} {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return map[string]interface{}{
		"pool_size":  len(cp.pool),
		"max_size":   cp.maxSize,
		"created":    cp.created,
		"borrowed":   cp.borrowed,
		"returned":   cp.returned,
	}
}

// 控制器接口定义

// ControllerInitializer 控制器初始化接口
type ControllerInitializer interface {
	Init(ctx *mvcContext.Context) error
}

// ControllerDestroyer 控制器销毁接口
type ControllerDestroyer interface {
	Destroy() error
}

// ControllerResettable 控制器重置接口
type ControllerResettable interface {
	Reset()
}

// LifecycleMetrics 方法实现

// updateCreated 更新创建数量
func (lm *LifecycleMetrics) updateCreated(delta int64) {
	lm.mu.Lock()
	lm.CreatedCount += delta
	lm.mu.Unlock()
}

// updateDestroyed 更新销毁数量
func (lm *LifecycleMetrics) updateDestroyed(delta int64) {
	lm.mu.Lock()
	lm.DestroyedCount += delta
	lm.mu.Unlock()
}

// updateActive 更新活跃数量
func (lm *LifecycleMetrics) updateActive(delta int64) {
	lm.mu.Lock()
	lm.ActiveCount += delta
	lm.mu.Unlock()
}

// GetMetrics 获取生命周期指标
func (lm *LifecycleManager) GetMetrics() *LifecycleMetrics {
	lm.metrics.mu.RLock()
	defer lm.metrics.mu.RUnlock()

	// 返回指标的副本
	return &LifecycleMetrics{
		CreatedCount:   lm.metrics.CreatedCount,
		DestroyedCount: lm.metrics.DestroyedCount,
		ActiveCount:    lm.metrics.ActiveCount,
		PoolHitRate:    lm.metrics.PoolHitRate,
		AverageLifetime: lm.metrics.AverageLifetime,
	}
}

// CreateController 默认工厂实现
func (f *DefaultControllerFactory) CreateController() (interface{}, error) {
	controllerValue := reflect.New(f.controllerType)
	return controllerValue.Interface(), nil
}

// InitController 初始化控制器
func (f *DefaultControllerFactory) InitController(controller interface{}, ctx *mvcContext.Context) error {
	if initializer, ok := controller.(ControllerInitializer); ok {
		return initializer.Init(ctx)
	}
	return nil
}

// DestroyController 销毁控制器
func (f *DefaultControllerFactory) DestroyController(controller interface{}) error {
	if destroyer, ok := controller.(ControllerDestroyer); ok {
		return destroyer.Destroy()
	}
	return nil
}