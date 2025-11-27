# 缓存系统

YYHertz 框架提供了高性能、多层级的缓存系统，支持内存缓存、Redis 缓存、分布式缓存等多种缓存策略，帮助提升应用程序性能。

## 概述

缓存是提高 Web 应用性能的关键技术。YYHertz 的缓存系统设计简洁、功能强大，支持：

- 多种缓存后端（内存、Redis、Memcached）
- 分层缓存策略
- 缓存预热和失效
- 分布式缓存一致性
- 缓存统计和监控

## 基本使用

### 初始化缓存

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/cache"
)

func main() {
    app := mvc.HertzApp
    
    // 使用内存缓存
    memCache := cache.NewMemoryCache(cache.MemoryConfig{
        MaxSize: 1000,
        TTL: 300, // 5 minutes
    })
    
    // 使用 Redis 缓存
    redisCache := cache.NewRedisCache(cache.RedisConfig{
        Addr: "localhost:6379",
        Password: "",
        DB: 0,
        PoolSize: 10,
    })
    
    // 注册缓存实例
    cache.RegisterCache("memory", memCache)
    cache.RegisterCache("redis", redisCache)
    
    app.Run()
}
```

### 在控制器中使用缓存

```go
package controllers

import (
    "fmt"
    "time"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/cache"
)

type ProductController struct {
    mvc.Controller
}

func (c *ProductController) GetProduct() {
    productID := c.GetString("id")
    cacheKey := fmt.Sprintf("product:%s", productID)
    
    // 获取缓存实例
    cache := cache.GetCache("redis")
    
    // 尝试从缓存获取
    var product Product
    if found := cache.Get(cacheKey, &product); found {
        c.JSON(200, product)
        return
    }
    
    // 缓存未命中，从数据库查询
    product, err := c.productService.GetByID(productID)
    if err != nil {
        c.JSON(500, map[string]string{"error": "Product not found"})
        return
    }
    
    // 存入缓存
    cache.Set(cacheKey, product, 10*time.Minute)
    
    c.JSON(200, product)
}

func (c *ProductController) UpdateProduct() {
    productID := c.GetString("id")
    
    // 更新产品数据
    if err := c.productService.Update(productID, updateData); err != nil {
        c.JSON(500, map[string]string{"error": "Update failed"})
        return
    }
    
    // 使缓存失效
    cacheKey := fmt.Sprintf("product:%s", productID)
    cache.GetCache("redis").Delete(cacheKey)
    
    // 也可以删除相关的缓存
    cache.GetCache("redis").DeletePattern("product:*")
    cache.GetCache("redis").DeletePattern("category:*")
    
    c.JSON(200, map[string]string{"status": "updated"})
}
```

## 缓存配置

### 内存缓存配置

```go
type MemoryConfig struct {
    // 最大缓存项数量
    MaxSize int `json:"max_size"`
    
    // 默认 TTL（秒）
    TTL int `json:"ttl"`
    
    // 清理间隔（秒）
    CleanupInterval int `json:"cleanup_interval"`
    
    // LRU 策略
    LRU bool `json:"lru"`
    
    // 是否启用统计
    EnableStats bool `json:"enable_stats"`
}

// 使用示例
memCache := cache.NewMemoryCache(cache.MemoryConfig{
    MaxSize: 10000,
    TTL: 600,
    CleanupInterval: 60,
    LRU: true,
    EnableStats: true,
})
```

### Redis 缓存配置

```go
type RedisConfig struct {
    // Redis 地址
    Addr string `json:"addr"`
    
    // 密码
    Password string `json:"password"`
    
    // 数据库编号
    DB int `json:"db"`
    
    // 连接池大小
    PoolSize int `json:"pool_size"`
    
    // 连接超时
    DialTimeout time.Duration `json:"dial_timeout"`
    
    // 读取超时
    ReadTimeout time.Duration `json:"read_timeout"`
    
    // 写入超时
    WriteTimeout time.Duration `json:"write_timeout"`
    
    // 键前缀
    KeyPrefix string `json:"key_prefix"`
    
    // 序列化方式 ("json", "gob", "msgpack")
    Serializer string `json:"serializer"`
}

// 使用示例
redisCache := cache.NewRedisCache(cache.RedisConfig{
    Addr: "localhost:6379",
    Password: "",
    DB: 0,
    PoolSize: 20,
    DialTimeout: 5 * time.Second,
    ReadTimeout: 3 * time.Second,
    WriteTimeout: 3 * time.Second,
    KeyPrefix: "app:",
    Serializer: "json",
})
```

## 缓存接口

### 基本操作

```go
// 缓存接口定义
type Cache interface {
    // 设置缓存
    Set(key string, value interface{}, ttl time.Duration) error
    
    // 获取缓存
    Get(key string, dest interface{}) bool
    
    // 删除缓存
    Delete(key string) error
    
    // 检查是否存在
    Exists(key string) bool
    
    // 设置过期时间
    Expire(key string, ttl time.Duration) error
    
    // 批量操作
    MSet(items map[string]interface{}, ttl time.Duration) error
    MGet(keys []string) map[string]interface{}
    MDelete(keys []string) error
    
    // 模式匹配删除
    DeletePattern(pattern string) error
    
    // 清空缓存
    Clear() error
    
    // 获取统计信息
    Stats() CacheStats
}
```

### 高级操作

```go
cache := cache.GetCache("redis")

// 原子递增
count, err := cache.Increment("visit_count", 1)

// 原子递减  
count, err := cache.Decrement("stock:123", 1)

// 列表操作
cache.ListPush("queue", "item1")
cache.ListPush("queue", "item2")
item, err := cache.ListPop("queue")

// 集合操作
cache.SetAdd("tags", "go", "web", "framework")
members := cache.SetMembers("tags")
exists := cache.SetExists("tags", "go")

// 哈希操作
cache.HashSet("user:123", "name", "John")
cache.HashSet("user:123", "age", 30)
name := cache.HashGet("user:123", "name")
user := cache.HashGetAll("user:123")
```

## 分层缓存

### 多级缓存策略

```go
package cache

// 分层缓存配置
type TieredCacheConfig struct {
    L1 Cache // 一级缓存（通常是内存）
    L2 Cache // 二级缓存（通常是 Redis）
    L3 Cache // 三级缓存（可选，如数据库缓存）
    
    // 各级缓存的 TTL
    L1TTL time.Duration
    L2TTL time.Duration
    L3TTL time.Duration
    
    // 缓存策略
    WriteThrough bool // 写穿透
    WriteBack    bool // 写回
}

// 创建分层缓存
func NewTieredCache(config TieredCacheConfig) Cache {
    return &TieredCache{
        config: config,
        stats:  NewCacheStats(),
    }
}

// 使用示例
memCache := cache.NewMemoryCache(cache.MemoryConfig{MaxSize: 1000})
redisCache := cache.NewRedisCache(cache.RedisConfig{Addr: "localhost:6379"})

tieredCache := cache.NewTieredCache(cache.TieredCacheConfig{
    L1: memCache,
    L2: redisCache,
    L1TTL: 5 * time.Minute,
    L2TTL: 30 * time.Minute,
    WriteThrough: true,
})

cache.RegisterCache("tiered", tieredCache)
```

### 缓存预热

```go
// 缓存预热器
type CacheWarmer struct {
    cache Cache
    dataSource DataSource
}

func (w *CacheWarmer) WarmUp() error {
    // 预热热门产品
    hotProducts, err := w.dataSource.GetHotProducts(100)
    if err != nil {
        return err
    }
    
    for _, product := range hotProducts {
        key := fmt.Sprintf("product:%d", product.ID)
        w.cache.Set(key, product, 1*time.Hour)
    }
    
    // 预热分类数据
    categories, err := w.dataSource.GetAllCategories()
    if err != nil {
        return err
    }
    
    w.cache.Set("categories:all", categories, 24*time.Hour)
    
    return nil
}

// 在应用启动时进行缓存预热
func main() {
    app := mvc.HertzApp
    
    // 初始化缓存
    cache := cache.GetCache("redis")
    
    // 缓存预热
    warmer := &CacheWarmer{
        cache: cache,
        dataSource: dataSource,
    }
    
    go func() {
        if err := warmer.WarmUp(); err != nil {
            log.Printf("Cache warm up failed: %v", err)
        }
    }()
    
    app.Run()
}
```

## 缓存装饰器

### 方法级缓存

```go
// 缓存装饰器
func Cacheable(key string, ttl time.Duration, cache Cache) func(interface{}) interface{} {
    return func(fn interface{}) interface{} {
        // 使用反射实现方法缓存
        // 这里简化展示概念
        return func(args ...interface{}) interface{} {
            // 生成缓存键
            cacheKey := generateCacheKey(key, args...)
            
            // 尝试获取缓存
            var result interface{}
            if cache.Get(cacheKey, &result) {
                return result
            }
            
            // 调用原方法
            result = callOriginalMethod(fn, args...)
            
            // 缓存结果
            cache.Set(cacheKey, result, ttl)
            
            return result
        }
    }
}

// 使用示例
type UserService struct {
    cache cache.Cache
}

// @Cacheable(key="user:profile:{0}", ttl="10m")
func (s *UserService) GetUserProfile(userID int) (*User, error) {
    key := fmt.Sprintf("user:profile:%d", userID)
    
    var user User
    if s.cache.Get(key, &user) {
        return &user, nil
    }
    
    // 从数据库查询
    user, err := s.getUserFromDB(userID)
    if err != nil {
        return nil, err
    }
    
    // 缓存结果
    s.cache.Set(key, user, 10*time.Minute)
    
    return &user, nil
}
```

### 自动失效装饰器

```go
// 缓存失效装饰器
func CacheEvict(patterns []string, cache Cache) func(interface{}) interface{} {
    return func(fn interface{}) interface{} {
        return func(args ...interface{}) interface{} {
            // 执行原方法
            result := callOriginalMethod(fn, args...)
            
            // 失效相关缓存
            for _, pattern := range patterns {
                cacheKey := generateCacheKey(pattern, args...)
                cache.DeletePattern(cacheKey)
            }
            
            return result
        }
    }
}

// @CacheEvict(patterns=["user:*", "user:profile:{0}"])
func (s *UserService) UpdateUser(userID int, data UpdateData) error {
    // 更新用户数据
    err := s.updateUserInDB(userID, data)
    if err != nil {
        return err
    }
    
    // 失效相关缓存
    patterns := []string{
        fmt.Sprintf("user:profile:%d", userID),
        "user:list:*",
        "user:count:*",
    }
    
    for _, pattern := range patterns {
        s.cache.DeletePattern(pattern)
    }
    
    return nil
}
```

## 缓存中间件

### HTTP 响应缓存

```go
// HTTP 缓存中间件
func HTTPCacheMiddleware(config HTTPCacheConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 只缓存 GET 请求
        if c.Request.Method != "GET" {
            c.Next()
            return
        }
        
        // 生成缓存键
        cacheKey := generateHTTPCacheKey(c.Request)
        
        // 尝试获取缓存
        var response CachedResponse
        if config.Cache.Get(cacheKey, &response) {
            // 设置响应头
            for k, v := range response.Headers {
                c.Header(k, v)
            }
            
            // 返回缓存的响应
            c.Data(response.StatusCode, response.ContentType, response.Body)
            c.Abort()
            return
        }
        
        // 包装响应写入器
        wrapper := &responseWrapper{
            ResponseWriter: c.Writer,
            body: &bytes.Buffer{},
        }
        c.Writer = wrapper
        
        // 执行请求
        c.Next()
        
        // 缓存响应
        if wrapper.status < 400 {
            response := CachedResponse{
                StatusCode: wrapper.status,
                Headers: wrapper.headers,
                Body: wrapper.body.Bytes(),
                ContentType: wrapper.Header().Get("Content-Type"),
            }
            
            config.Cache.Set(cacheKey, response, config.TTL)
        }
    }
}

// 使用示例
app.Use(HTTPCacheMiddleware(HTTPCacheConfig{
    Cache: cache.GetCache("redis"),
    TTL: 5 * time.Minute,
    SkipPaths: []string{"/api/", "/admin/"},
}))
```

## 缓存统计和监控

### 缓存统计

```go
type CacheStats struct {
    Hits        int64     `json:"hits"`
    Misses      int64     `json:"misses"`
    Sets        int64     `json:"sets"`
    Deletes     int64     `json:"deletes"`
    Errors      int64     `json:"errors"`
    HitRate     float64   `json:"hit_rate"`
    LastReset   time.Time `json:"last_reset"`
    
    // 性能指标
    AvgSetTime  time.Duration `json:"avg_set_time"`
    AvgGetTime  time.Duration `json:"avg_get_time"`
    
    // 内存使用（仅内存缓存）
    MemoryUsage int64 `json:"memory_usage"`
    ItemCount   int64 `json:"item_count"`
}

// 获取缓存统计
func (c *RedisCache) Stats() CacheStats {
    return c.stats.Get()
}

// 重置统计
func (c *RedisCache) ResetStats() {
    c.stats.Reset()
}
```

### 监控接口

```go
// 缓存监控控制器
type CacheController struct {
    mvc.Controller
}

func (c *CacheController) GetStats() {
    stats := make(map[string]CacheStats)
    
    // 获取所有缓存的统计信息
    for name, cache := range cache.GetAllCaches() {
        stats[name] = cache.Stats()
    }
    
    c.JSON(200, stats)
}

func (c *CacheController) ClearCache() {
    cacheName := c.GetString("name")
    
    if cacheName == "" {
        // 清空所有缓存
        for _, cache := range cache.GetAllCaches() {
            cache.Clear()
        }
    } else {
        // 清空指定缓存
        if cache := cache.GetCache(cacheName); cache != nil {
            cache.Clear()
        }
    }
    
    c.JSON(200, map[string]string{"status": "cleared"})
}

func (c *CacheController) GetKeys() {
    cacheName := c.GetString("name")
    pattern := c.GetString("pattern", "*")
    
    cache := cache.GetCache(cacheName)
    if cache == nil {
        c.JSON(404, map[string]string{"error": "Cache not found"})
        return
    }
    
    keys := cache.Keys(pattern)
    c.JSON(200, map[string]interface{}{
        "keys": keys,
        "count": len(keys),
    })
}
```

## 最佳实践

### 1. 缓存键设计

```go
// 好的缓存键设计
const (
    KeyUserProfile    = "user:profile:%d"
    KeyUserPosts      = "user:%d:posts:page:%d"
    KeyCategoryTree   = "category:tree"
    KeyProductByID    = "product:id:%d"
    KeyProductsByTag  = "products:tag:%s:page:%d"
)

// 使用函数生成缓存键
func UserProfileKey(userID int) string {
    return fmt.Sprintf(KeyUserProfile, userID)
}

func UserPostsKey(userID, page int) string {
    return fmt.Sprintf(KeyUserPosts, userID, page)
}
```

### 2. TTL 策略

```go
var (
    ShortTTL  = 5 * time.Minute   // 快速变化的数据
    MediumTTL = 1 * time.Hour     // 中等变化的数据
    LongTTL   = 24 * time.Hour    // 很少变化的数据
    PermanentTTL = 0              // 永不过期（需要手动失效）
)

// 根据数据特性选择 TTL
cache.Set(UserProfileKey(userID), user, MediumTTL)
cache.Set("site:config", config, LongTTL)
cache.Set("trending:posts", posts, ShortTTL)
```

### 3. 错误处理

```go
func (s *UserService) GetUser(userID int) (*User, error) {
    key := UserProfileKey(userID)
    
    var user User
    if s.cache.Get(key, &user) {
        return &user, nil
    }
    
    // 从数据库获取
    user, err := s.userRepo.GetByID(userID)
    if err != nil {
        return nil, err
    }
    
    // 缓存时处理错误，但不影响主流程
    if err := s.cache.Set(key, user, MediumTTL); err != nil {
        // 记录错误但继续执行
        log.Printf("Failed to cache user %d: %v", userID, err)
    }
    
    return &user, nil
}
```

### 4. 缓存穿透防护

```go
// 使用空值缓存防止缓存穿透
func (s *UserService) GetUser(userID int) (*User, error) {
    key := UserProfileKey(userID)
    
    var cached CachedUser
    if s.cache.Get(key, &cached) {
        if cached.IsNull {
            return nil, ErrUserNotFound
        }
        return cached.User, nil
    }
    
    user, err := s.userRepo.GetByID(userID)
    if err == ErrUserNotFound {
        // 缓存空值，较短的 TTL
        nullCached := CachedUser{IsNull: true}
        s.cache.Set(key, nullCached, ShortTTL)
        return nil, err
    }
    
    if err != nil {
        return nil, err
    }
    
    // 缓存正常值
    cached = CachedUser{User: user, IsNull: false}
    s.cache.Set(key, cached, MediumTTL)
    
    return user, nil
}
```

## 示例：完整的产品缓存系统

```go
package services

type ProductService struct {
    cache    cache.Cache
    repo     ProductRepository
    logger   Logger
}

func NewProductService(cache cache.Cache, repo ProductRepository) *ProductService {
    return &ProductService{
        cache: cache,
        repo: repo,
        logger: logger.WithField("service", "product"),
    }
}

// 获取产品详情（带缓存）
func (s *ProductService) GetProduct(id int) (*Product, error) {
    key := fmt.Sprintf("product:detail:%d", id)
    
    var product Product
    if s.cache.Get(key, &product) {
        s.logger.Debug("Cache hit", "key", key)
        return &product, nil
    }
    
    s.logger.Debug("Cache miss", "key", key)
    
    // 从数据库获取
    product, err := s.repo.GetByID(id)
    if err != nil {
        return nil, err
    }
    
    // 异步缓存，不阻塞响应
    go func() {
        if err := s.cache.Set(key, product, 30*time.Minute); err != nil {
            s.logger.Error("Failed to cache product", "id", id, "error", err)
        }
    }()
    
    return &product, nil
}

// 获取产品列表（带分页缓存）
func (s *ProductService) GetProducts(categoryID int, page, size int) (*ProductList, error) {
    key := fmt.Sprintf("products:category:%d:page:%d:size:%d", categoryID, page, size)
    
    var list ProductList
    if s.cache.Get(key, &list) {
        return &list, nil
    }
    
    list, err := s.repo.GetByCategory(categoryID, page, size)
    if err != nil {
        return nil, err
    }
    
    // 缓存列表数据，较短 TTL
    s.cache.Set(key, list, 10*time.Minute)
    
    return &list, nil
}

// 更新产品（使缓存失效）
func (s *ProductService) UpdateProduct(id int, data UpdateData) error {
    // 更新数据库
    if err := s.repo.Update(id, data); err != nil {
        return err
    }
    
    // 失效相关缓存
    patterns := []string{
        fmt.Sprintf("product:detail:%d", id),
        fmt.Sprintf("products:category:%d:*", data.CategoryID),
        "products:search:*",
        "products:featured:*",
    }
    
    for _, pattern := range patterns {
        if err := s.cache.DeletePattern(pattern); err != nil {
            s.logger.Error("Failed to invalidate cache", "pattern", pattern, "error", err)
        }
    }
    
    return nil
}
```

YYHertz 的缓存系统提供了完整的缓存解决方案，从简单的键值缓存到复杂的分层缓存策略，能够有效提升应用程序的性能和用户体验。
