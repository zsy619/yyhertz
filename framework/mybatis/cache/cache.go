// Package cache 提供MyBatis缓存实现
//
// 支持多种缓存策略：LRU、FIFO、弱引用、软引用等
// 提供缓存装饰器：阻塞、日志、序列化、同步、事务等
package cache

import (
	"sync"
	"time"
)

// Cache 缓存接口
type Cache interface {
	// GetId 获取缓存ID
	GetId() string

	// Put 存储对象
	Put(key string, value any)

	// Get 获取对象
	Get(key string) (any, bool)

	// Remove 移除对象
	Remove(key string) any

	// Clear 清空缓存
	Clear()

	// GetSize 获取缓存大小
	GetSize() int

	// 兼容旧接口
	PutObject(key any, value any)
	GetObject(key any) any
	RemoveObject(key any) any
}

// CacheStats 缓存统计信息
type CacheStats struct {
	HitCount     int64         // 命中次数
	MissCount    int64         // 未命中次数
	PutCount     int64         // 存储次数
	EvictCount   int64         // 淘汰次数
	LoadTime     time.Duration // 总加载时间
	RequestCount int64         // 请求总数
}

// PerpetualCache 永久缓存实现
type PerpetualCache struct {
	id       string
	cache    map[string]any
	objCache map[any]any // 兼容旧接口
	mutex    sync.RWMutex
}

// NewPerpetualCache 创建永久缓存
func NewPerpetualCache(id string) *PerpetualCache {
	return &PerpetualCache{
		id:       id,
		cache:    make(map[string]any),
		objCache: make(map[any]any),
	}
}

// GetId 获取缓存ID
func (c *PerpetualCache) GetId() string {
	return c.id
}

// Put 存储对象（新接口）
func (c *PerpetualCache) Put(key string, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache[key] = value
}

// Get 获取对象（新接口）
func (c *PerpetualCache) Get(key string) (any, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, exists := c.cache[key]
	return value, exists
}

// Remove 移除对象（新接口）
func (c *PerpetualCache) Remove(key string) any {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	value := c.cache[key]
	delete(c.cache, key)
	return value
}

// PutObject 存储对象（兼容旧接口）
func (c *PerpetualCache) PutObject(key any, value any) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.objCache[key] = value
}

// GetObject 获取对象（兼容旧接口）
func (c *PerpetualCache) GetObject(key any) any {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.objCache[key]
}

// RemoveObject 移除对象（兼容旧接口）
func (c *PerpetualCache) RemoveObject(key any) any {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	value := c.objCache[key]
	delete(c.objCache, key)
	return value
}

// Clear 清空缓存
func (c *PerpetualCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]any)
	c.objCache = make(map[any]any)
}

// GetSize 获取缓存大小
func (c *PerpetualCache) GetSize() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.cache) + len(c.objCache)
}

// LruCache LRU缓存
type LruCache struct {
	delegate Cache
	keyMap   map[string]*lruNode
	head     *lruNode
	tail     *lruNode
	capacity int
	mutex    sync.RWMutex
}

// lruNode LRU节点
type lruNode struct {
	key  string
	time time.Time
	prev *lruNode
	next *lruNode
}

// NewLruCache 创建LRU缓存
func NewLruCache(delegate Cache, capacity int) *LruCache {
	cache := &LruCache{
		delegate: delegate,
		keyMap:   make(map[string]*lruNode),
		capacity: capacity,
	}

	// 创建哨兵节点
	cache.head = &lruNode{}
	cache.tail = &lruNode{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head

	return cache
}

// GetId 获取缓存ID
func (cache *LruCache) GetId() string {
	return cache.delegate.GetId()
}

// Put 存储缓存项
func (cache *LruCache) Put(key string, value any) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, exists := cache.keyMap[key]; exists {
		// 更新现有节点
		cache.moveToHead(node)
	} else {
		// 添加新节点
		node := &lruNode{key: key, time: time.Now()}
		cache.keyMap[key] = node
		cache.addToHead(node)

		// 检查容量
		if len(cache.keyMap) > cache.capacity {
			tail := cache.removeTail()
			delete(cache.keyMap, tail.key)
			cache.delegate.Remove(tail.key)
		}
	}

	cache.delegate.Put(key, value)
}

// Get 获取缓存项
func (cache *LruCache) Get(key string) (any, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, exists := cache.keyMap[key]; exists {
		cache.moveToHead(node)
		return cache.delegate.Get(key)
	}

	return nil, false
}

// Remove 删除缓存项
func (cache *LruCache) Remove(key string) any {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if node, exists := cache.keyMap[key]; exists {
		cache.removeNode(node)
		delete(cache.keyMap, key)
	}

	return cache.delegate.Remove(key)
}

// Clear 清空缓存
func (cache *LruCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.keyMap = make(map[string]*lruNode)
	cache.head.next = cache.tail
	cache.tail.prev = cache.head
	cache.delegate.Clear()
}

// GetSize 获取缓存大小
func (cache *LruCache) GetSize() int {
	return cache.delegate.GetSize()
}

// PutObject 兼容旧接口
func (cache *LruCache) PutObject(key any, value any) {
	cache.delegate.PutObject(key, value)
}

// GetObject 兼容旧接口
func (cache *LruCache) GetObject(key any) any {
	return cache.delegate.GetObject(key)
}

// RemoveObject 兼容旧接口
func (cache *LruCache) RemoveObject(key any) any {
	return cache.delegate.RemoveObject(key)
}

// LRU节点操作方法
func (cache *LruCache) addToHead(node *lruNode) {
	node.prev = cache.head
	node.next = cache.head.next
	cache.head.next.prev = node
	cache.head.next = node
}

func (cache *LruCache) removeNode(node *lruNode) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (cache *LruCache) moveToHead(node *lruNode) {
	cache.removeNode(node)
	cache.addToHead(node)
}

func (cache *LruCache) removeTail() *lruNode {
	lastNode := cache.tail.prev
	cache.removeNode(lastNode)
	return lastNode
}

// FifoCache FIFO缓存
type FifoCache struct {
	delegate Cache
	keyQueue []string
	capacity int
	mutex    sync.RWMutex
}

// NewFifoCache 创建FIFO缓存
func NewFifoCache(delegate Cache, capacity int) *FifoCache {
	return &FifoCache{
		delegate: delegate,
		keyQueue: make([]string, 0),
		capacity: capacity,
	}
}

// FIFO缓存方法实现
func (cache *FifoCache) GetId() string { return cache.delegate.GetId() }

func (cache *FifoCache) Put(key string, value any) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	// 检查容量
	if len(cache.keyQueue) >= cache.capacity {
		// 移除最旧的
		oldestKey := cache.keyQueue[0]
		cache.keyQueue = cache.keyQueue[1:]
		cache.delegate.Remove(oldestKey)
	}

	cache.keyQueue = append(cache.keyQueue, key)
	cache.delegate.Put(key, value)
}

func (cache *FifoCache) Get(key string) (any, bool) { return cache.delegate.Get(key) }
func (cache *FifoCache) Remove(key string) any      { return cache.delegate.Remove(key) }
func (cache *FifoCache) Clear() {
	cache.delegate.Clear()
	cache.keyQueue = cache.keyQueue[:0]
}
func (cache *FifoCache) GetSize() int                 { return cache.delegate.GetSize() }
func (cache *FifoCache) PutObject(key any, value any) { cache.delegate.PutObject(key, value) }
func (cache *FifoCache) GetObject(key any) any        { return cache.delegate.GetObject(key) }
func (cache *FifoCache) RemoveObject(key any) any     { return cache.delegate.RemoveObject(key) }

// BlockingCache 阻塞缓存
type BlockingCache struct {
	delegate Cache
	locks    map[string]*sync.Mutex
	mutex    sync.RWMutex
}

// NewBlockingCache 创建阻塞缓存
func NewBlockingCache(delegate Cache) *BlockingCache {
	return &BlockingCache{
		delegate: delegate,
		locks:    make(map[string]*sync.Mutex),
	}
}

// BlockingCache方法实现
func (cache *BlockingCache) GetId() string { return cache.delegate.GetId() }

func (cache *BlockingCache) Put(key string, value any) {
	cache.acquireLock(key)
	defer cache.releaseLock(key)
	cache.delegate.Put(key, value)
}

func (cache *BlockingCache) Get(key string) (any, bool) {
	cache.acquireLock(key)
	defer cache.releaseLock(key)
	return cache.delegate.Get(key)
}

func (cache *BlockingCache) Remove(key string) any {
	cache.acquireLock(key)
	defer cache.releaseLock(key)
	return cache.delegate.Remove(key)
}

func (cache *BlockingCache) Clear()                       { cache.delegate.Clear() }
func (cache *BlockingCache) GetSize() int                 { return cache.delegate.GetSize() }
func (cache *BlockingCache) PutObject(key any, value any) { cache.delegate.PutObject(key, value) }
func (cache *BlockingCache) GetObject(key any) any        { return cache.delegate.GetObject(key) }
func (cache *BlockingCache) RemoveObject(key any) any     { return cache.delegate.RemoveObject(key) }

// acquireLock 获取锁
func (cache *BlockingCache) acquireLock(key string) {
	cache.mutex.Lock()
	lock, exists := cache.locks[key]
	if !exists {
		lock = &sync.Mutex{}
		cache.locks[key] = lock
	}
	cache.mutex.Unlock()

	lock.Lock()
}

// releaseLock 释放锁
func (cache *BlockingCache) releaseLock(key string) {
	cache.mutex.RLock()
	lock, exists := cache.locks[key]
	cache.mutex.RUnlock()

	if exists {
		lock.Unlock()

		// 清理锁
		cache.mutex.Lock()
		delete(cache.locks, key)
		cache.mutex.Unlock()
	}
}

// SynchronizedCache 同步缓存
type SynchronizedCache struct {
	delegate Cache
	mutex    sync.RWMutex
}

// NewSynchronizedCache 创建同步缓存
func NewSynchronizedCache(delegate Cache) *SynchronizedCache {
	return &SynchronizedCache{
		delegate: delegate,
	}
}

// SynchronizedCache方法实现
func (cache *SynchronizedCache) GetId() string {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.delegate.GetId()
}

func (cache *SynchronizedCache) Put(key string, value any) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.delegate.Put(key, value)
}

func (cache *SynchronizedCache) Get(key string) (any, bool) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.delegate.Get(key)
}

func (cache *SynchronizedCache) Remove(key string) any {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	return cache.delegate.Remove(key)
}

func (cache *SynchronizedCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.delegate.Clear()
}

func (cache *SynchronizedCache) GetSize() int {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.delegate.GetSize()
}

func (cache *SynchronizedCache) PutObject(key any, value any) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.delegate.PutObject(key, value)
}

func (cache *SynchronizedCache) GetObject(key any) any {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	return cache.delegate.GetObject(key)
}

func (cache *SynchronizedCache) RemoveObject(key any) any {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	return cache.delegate.RemoveObject(key)
}

// 其他缓存类型的简化实现

// WeakCache 弱引用缓存
type WeakCache struct {
	delegate       Cache
	expirationTime time.Duration
	expirationMap  map[string]time.Time
	mutex          sync.RWMutex
}

// SoftCache 软引用缓存
type SoftCache struct {
	delegate      Cache
	maxMemory     int64
	currentMemory int64
	mutex         sync.RWMutex
}

// LoggingCache 日志缓存
type LoggingCache struct {
	delegate Cache
	stats    *CacheStats
	logger   Logger
	mutex    sync.RWMutex
}

// SerializedCache 序列化缓存
type SerializedCache struct {
	delegate   Cache
	serializer Serializer
}

// TransactionalCache 事务缓存
type TransactionalCache struct {
	delegate             Cache
	entriesToAddOnCommit map[string]any
	entriesMissedInCache map[string]bool
	clearOnCommit        bool
	mutex                sync.RWMutex
}

// Logger 日志接口
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Serializer 序列化接口
type Serializer interface {
	Serialize(obj any) ([]byte, error)
	Deserialize(data []byte) (any, error)
}

// CacheBuilder 缓存构建器
type CacheBuilder struct {
	id             string
	implementation string
	decorators     []string
	properties     map[string]any
	size           int
	clearInterval  time.Duration
	blocking       bool
	stats          bool
	logger         Logger
}

// NewCacheBuilder 创建缓存构建器
func NewCacheBuilder(id string) *CacheBuilder {
	return &CacheBuilder{
		id:             id,
		implementation: "LRU",
		decorators:     make([]string, 0),
		properties:     make(map[string]any),
		size:           1024,
		blocking:       false,
		stats:          false,
	}
}

// Implementation 设置实现类型
func (builder *CacheBuilder) Implementation(impl string) *CacheBuilder {
	builder.implementation = impl
	return builder
}

// Size 设置缓存大小
func (builder *CacheBuilder) Size(size int) *CacheBuilder {
	builder.size = size
	return builder
}

// Decorator 添加装饰器
func (builder *CacheBuilder) Decorator(decorator string) *CacheBuilder {
	builder.decorators = append(builder.decorators, decorator)
	return builder
}

// Property 设置属性
func (builder *CacheBuilder) Property(key string, value any) *CacheBuilder {
	builder.properties[key] = value
	return builder
}

// ClearInterval 设置清理间隔
func (builder *CacheBuilder) ClearInterval(interval time.Duration) *CacheBuilder {
	builder.clearInterval = interval
	return builder
}

// Blocking 设置是否阻塞
func (builder *CacheBuilder) Blocking(blocking bool) *CacheBuilder {
	builder.blocking = blocking
	return builder
}

// Stats 设置是否统计
func (builder *CacheBuilder) Stats(stats bool) *CacheBuilder {
	builder.stats = stats
	return builder
}

// Logger 设置日志器
func (builder *CacheBuilder) Logger(logger Logger) *CacheBuilder {
	builder.logger = logger
	return builder
}

// Build 构建缓存
func (builder *CacheBuilder) Build() (Cache, error) {
	// 创建基础缓存
	var cache Cache

	switch builder.implementation {
	case "FIFO":
		cache = NewFifoCache(NewPerpetualCache(builder.id), builder.size)
	case "SOFT":
		cache = NewSoftCache(NewPerpetualCache(builder.id))
	case "WEAK":
		cache = NewWeakCache(NewPerpetualCache(builder.id))
	default: // LRU
		cache = NewLruCache(NewPerpetualCache(builder.id), builder.size)
	}

	// 应用装饰器
	for _, decorator := range builder.decorators {
		switch decorator {
		case "BLOCKING":
			cache = NewBlockingCache(cache)
		case "LOGGING":
			cache = NewLoggingCache(cache, builder.logger)
		case "SERIALIZED":
			cache = NewSerializedCache(cache, nil)
		case "SYNCHRONIZED":
			cache = NewSynchronizedCache(cache)
		case "TRANSACTIONAL":
			cache = NewTransactionalCache(cache)
		}
	}

	return cache, nil
}

// 其他缓存类型的构造函数和方法实现

func NewWeakCache(delegate Cache) *WeakCache {
	return &WeakCache{
		delegate:       delegate,
		expirationTime: 5 * time.Minute,
		expirationMap:  make(map[string]time.Time),
	}
}

func NewSoftCache(delegate Cache) *SoftCache {
	return &SoftCache{
		delegate:  delegate,
		maxMemory: 64 * 1024 * 1024, // 64MB
	}
}

func NewLoggingCache(delegate Cache, logger Logger) *LoggingCache {
	return &LoggingCache{
		delegate: delegate,
		stats:    &CacheStats{},
		logger:   logger,
	}
}

func NewSerializedCache(delegate Cache, serializer Serializer) *SerializedCache {
	return &SerializedCache{
		delegate:   delegate,
		serializer: serializer,
	}
}

func NewTransactionalCache(delegate Cache) *TransactionalCache {
	return &TransactionalCache{
		delegate:             delegate,
		entriesToAddOnCommit: make(map[string]any),
		entriesMissedInCache: make(map[string]bool),
	}
}

// 为其他缓存类型实现完整的Cache接口

// WeakCache 方法实现
func (cache *WeakCache) GetId() string                { return cache.delegate.GetId() }
func (cache *WeakCache) Put(key string, value any)    { cache.delegate.Put(key, value) }
func (cache *WeakCache) Get(key string) (any, bool)   { return cache.delegate.Get(key) }
func (cache *WeakCache) Remove(key string) any        { return cache.delegate.Remove(key) }
func (cache *WeakCache) Clear()                       { cache.delegate.Clear() }
func (cache *WeakCache) GetSize() int                 { return cache.delegate.GetSize() }
func (cache *WeakCache) PutObject(key any, value any) { cache.delegate.PutObject(key, value) }
func (cache *WeakCache) GetObject(key any) any        { return cache.delegate.GetObject(key) }
func (cache *WeakCache) RemoveObject(key any) any     { return cache.delegate.RemoveObject(key) }

// SoftCache 方法实现
func (cache *SoftCache) GetId() string                { return cache.delegate.GetId() }
func (cache *SoftCache) Put(key string, value any)    { cache.delegate.Put(key, value) }
func (cache *SoftCache) Get(key string) (any, bool)   { return cache.delegate.Get(key) }
func (cache *SoftCache) Remove(key string) any        { return cache.delegate.Remove(key) }
func (cache *SoftCache) Clear()                       { cache.delegate.Clear() }
func (cache *SoftCache) GetSize() int                 { return cache.delegate.GetSize() }
func (cache *SoftCache) PutObject(key any, value any) { cache.delegate.PutObject(key, value) }
func (cache *SoftCache) GetObject(key any) any        { return cache.delegate.GetObject(key) }
func (cache *SoftCache) RemoveObject(key any) any     { return cache.delegate.RemoveObject(key) }

// LoggingCache 方法实现
func (cache *LoggingCache) GetId() string                { return cache.delegate.GetId() }
func (cache *LoggingCache) Put(key string, value any)    { cache.delegate.Put(key, value) }
func (cache *LoggingCache) Get(key string) (any, bool)   { return cache.delegate.Get(key) }
func (cache *LoggingCache) Remove(key string) any        { return cache.delegate.Remove(key) }
func (cache *LoggingCache) Clear()                       { cache.delegate.Clear() }
func (cache *LoggingCache) GetSize() int                 { return cache.delegate.GetSize() }
func (cache *LoggingCache) PutObject(key any, value any) { cache.delegate.PutObject(key, value) }
func (cache *LoggingCache) GetObject(key any) any        { return cache.delegate.GetObject(key) }
func (cache *LoggingCache) RemoveObject(key any) any     { return cache.delegate.RemoveObject(key) }

// SerializedCache 方法实现
func (cache *SerializedCache) GetId() string                { return cache.delegate.GetId() }
func (cache *SerializedCache) Put(key string, value any)    { cache.delegate.Put(key, value) }
func (cache *SerializedCache) Get(key string) (any, bool)   { return cache.delegate.Get(key) }
func (cache *SerializedCache) Remove(key string) any        { return cache.delegate.Remove(key) }
func (cache *SerializedCache) Clear()                       { cache.delegate.Clear() }
func (cache *SerializedCache) GetSize() int                 { return cache.delegate.GetSize() }
func (cache *SerializedCache) PutObject(key any, value any) { cache.delegate.PutObject(key, value) }
func (cache *SerializedCache) GetObject(key any) any        { return cache.delegate.GetObject(key) }
func (cache *SerializedCache) RemoveObject(key any) any     { return cache.delegate.RemoveObject(key) }

// TransactionalCache 方法实现
func (cache *TransactionalCache) GetId() string                { return cache.delegate.GetId() }
func (cache *TransactionalCache) Put(key string, value any)    { cache.delegate.Put(key, value) }
func (cache *TransactionalCache) Get(key string) (any, bool)   { return cache.delegate.Get(key) }
func (cache *TransactionalCache) Remove(key string) any        { return cache.delegate.Remove(key) }
func (cache *TransactionalCache) Clear()                       { cache.delegate.Clear() }
func (cache *TransactionalCache) GetSize() int                 { return cache.delegate.GetSize() }
func (cache *TransactionalCache) PutObject(key any, value any) { cache.delegate.PutObject(key, value) }
func (cache *TransactionalCache) GetObject(key any) any        { return cache.delegate.GetObject(key) }
func (cache *TransactionalCache) RemoveObject(key any) any     { return cache.delegate.RemoveObject(key) }
