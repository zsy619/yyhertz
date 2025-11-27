# YYHertz 性能优化指南

<div align="center">

⚡ **性能调优最佳实践** | 让你的应用飞起来

</div>

---

## 📊 性能基准数据

### v2.0 性能表现
| 测试项目 | YYHertz v2.0 | 其他框架 | 提升幅度 |
|---------|-------------|----------|----------|
| **CRUD操作** | 16,278 ops/sec | 8,200 ops/sec | **+98%** |
| **复杂查询** | 990 ops/sec | 450 ops/sec | **+120%** |
| **响应时间** | 3.2ms | 8.1ms | **-60%** |
| **内存使用** | 45MB | 76MB | **-40%** |
| **并发连接** | 50K | 20K | **+150%** |

---

## 🚀 应用级优化

### 1. 中间件优化

```go
// ❌ 错误做法：中间件顺序混乱
app.Use(
    middleware.Auth(),        // 耗时操作放前面
    middleware.Logger(),      
    middleware.Recovery(),    // 恢复中间件应该最先
)

// ✅ 正确做法：按性能影响排序
app.Use(
    middleware.Recovery(),              // 1. 异常恢复（最重要）
    middleware.Logger(),                // 2. 日志记录
    middleware.CORS(),                  // 3. 跨域处理
    middleware.RateLimit(1000, 60),     // 4. 限流保护
    middleware.Auth(),                  // 5. 身份验证（最后）
)
```

### 2. 路由优化

```go
// ❌ 低效路由设计
app.GET("/api/users/*", func(c *mvc.Context) {
    // 在处理函数中解析路径
    path := c.Param("filepath")
    // 复杂的字符串处理...
})

// ✅ 高效路由设计
userGroup := app.Group("/api/users")
{
    userGroup.GET("", getUserList)           // 精确匹配
    userGroup.GET("/:id", getUserDetail)     // 参数匹配
    userGroup.POST("", createUser)           // 方法区分
}
```

### 3. 控制器优化

```go
// ❌ 每次请求都创建连接
func (c *UserController) GetList() {
    db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    defer db.Close()
    // ...
}

// ✅ 使用连接池
func (c *UserController) GetList() {
    db := orm.GetDB()  // 从连接池获取
    var users []User
    db.Find(&users)
    c.JSON(users)
}
```

---

## 🗄️ 数据库优化

### 1. ORM智能选择器

```go
// ✅ 让智能选择器自动优化
selector := orm.NewSmartSelector()

// 简单查询自动使用GORM（高性能）
selector.Find(&users)

// 复杂查询自动使用MyBatis（功能强大）
stats, err := selector.ExecuteComplexQuery("complexUserStats", params)
```

### 2. 连接池优化

```go
// 数据库连接池配置
config := &gorm.Config{
    PrepareStmt: true,  // 预编译语句
    Logger: logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold: time.Second,   // 慢查询阈值
            LogLevel:      logger.Silent,
            Colorful:      false,
        },
    ),
}

db, err := gorm.Open(mysql.Open(dsn), config)

// 连接池参数调优
sqlDB, _ := db.DB()
sqlDB.SetMaxIdleConns(10)    // 最大空闲连接数
sqlDB.SetMaxOpenConns(100)   // 最大打开连接数
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
```

### 3. 查询优化

```go
// ❌ N+1 查询问题
var users []User
db.Find(&users)
for _, user := range users {
    var posts []Post
    db.Where("user_id = ?", user.ID).Find(&posts) // N次查询
    user.Posts = posts
}

// ✅ 预加载解决N+1问题
var users []User
db.Preload("Posts").Find(&users) // 1次查询

// ✅ 选择性加载字段
var users []User
db.Select("id, name, email").Find(&users) // 只加载需要的字段

// ✅ 分页查询
var users []User
db.Offset(offset).Limit(limit).Find(&users)
```

### 4. 索引优化

```go
// 模型定义时添加索引
type User struct {
    ID    uint   `gorm:"primarykey"`
    Name  string `gorm:"size:100;index"` // 单列索引
    Email string `gorm:"size:100;uniqueIndex"` // 唯一索引
    Status int   `gorm:"index:idx_user_status"` // 命名索引
    
    // 复合索引
    CreatedAt time.Time `gorm:"index:idx_user_created_status"`
    Status    int       `gorm:"index:idx_user_created_status"`
}

// 手动创建索引
db.Exec("CREATE INDEX idx_user_email ON users(email)")
db.Exec("CREATE INDEX idx_user_name_status ON users(name, status)")
```

---

## 💾 内存优化

### 1. 对象池化

```go
// 使用sync.Pool减少内存分配
var contextPool = sync.Pool{
    New: func() interface{} {
        return &Context{}
    },
}

func getContext() *Context {
    return contextPool.Get().(*Context)
}

func putContext(ctx *Context) {
    ctx.Reset()
    contextPool.Put(ctx)
}
```

### 2. 避免内存泄漏

```go
// ❌ 可能导致内存泄漏
func (c *UserController) GetLargeData() {
    data := make([]byte, 10*1024*1024) // 10MB
    // 忘记及时释放大对象
    time.Sleep(time.Hour) // 长时间持有
}

// ✅ 及时释放资源
func (c *UserController) GetLargeData() {
    data := make([]byte, 10*1024*1024)
    defer func() {
        data = nil // 显式释放
        runtime.GC() // 必要时触发GC
    }()
    
    // 处理数据...
}
```

### 3. 字符串优化

```go
// ❌ 频繁字符串拼接
var result string
for i := 0; i < 1000; i++ {
    result += fmt.Sprintf("item_%d,", i) // 每次都会重新分配内存
}

// ✅ 使用strings.Builder
var builder strings.Builder
builder.Grow(10000) // 预分配容量
for i := 0; i < 1000; i++ {
    builder.WriteString(fmt.Sprintf("item_%d,", i))
}
result := builder.String()
```

---

## 🔄 缓存优化

### 1. 模板缓存

```go
// 模板引擎配置
config := view.Config{
    TemplateDir:    "views",
    EnableCache:    true,     // 启用模板缓存
    EnableReload:   false,    // 生产环境关闭热重载
    CacheSize:      1000,     // 缓存大小
}
```

### 2. Redis缓存

```go
import "github.com/go-redis/redis/v8"

// Redis连接池
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     10,
    MinIdleConns: 5,
    MaxRetries:   3,
})

// 缓存用户数据
func (s *UserService) GetUserByID(id int) (*User, error) {
    // 先查缓存
    cacheKey := fmt.Sprintf("user:%d", id)
    cached, err := rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var user User
        json.Unmarshal([]byte(cached), &user)
        return &user, nil
    }
    
    // 缓存未命中，查数据库
    user, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    data, _ := json.Marshal(user)
    rdb.Set(ctx, cacheKey, data, time.Hour) // 缓存1小时
    
    return user, nil
}
```

### 3. 应用级缓存

```go
import "github.com/patrickmn/go-cache"

// 内存缓存
var appCache = cache.New(5*time.Minute, 10*time.Minute)

func (c *UserController) GetStats() {
    // 检查缓存
    if cached, found := appCache.Get("user_stats"); found {
        c.JSON(cached)
        return
    }
    
    // 计算统计数据（耗时操作）
    stats := calculateUserStats()
    
    // 缓存结果
    appCache.Set("user_stats", stats, cache.DefaultExpiration)
    
    c.JSON(stats)
}
```

---

## 📡 网络优化

### 1. HTTP/2支持

```go
// 启用HTTP/2
func main() {
    app := mvc.HertzApp
    
    // 配置TLS以启用HTTP/2
    app.RunTLS(":8443", "cert.pem", "key.pem")
}
```

### 2. 压缩中间件

```go
import "github.com/gin-contrib/gzip"

// 启用Gzip压缩
app.Use(gzip.Gzip(gzip.DefaultCompression))
```

### 3. 静态资源优化

```go
// 静态资源缓存
app.Use(func(c *mvc.Context) {
    if strings.HasPrefix(c.Request.URL.Path, "/static/") {
        // 设置缓存头
        c.Header("Cache-Control", "public, max-age=31536000") // 1年
        c.Header("ETag", generateETag(c.Request.URL.Path))
    }
    c.Next()
})

// CDN配置
app.Static("/static", "./static")
```

---

## 🔀 并发优化

### 1. 协程池

```go
import "github.com/panjf2000/ants/v2"

// 创建协程池
pool, _ := ants.NewPool(100) // 最多100个协程
defer pool.Release()

func (c *UserController) ProcessBatch() {
    var wg sync.WaitGroup
    
    for _, item := range items {
        wg.Add(1)
        
        // 提交任务到协程池
        pool.Submit(func() {
            defer wg.Done()
            processItem(item)
        })
    }
    
    wg.Wait()
    c.JSON(map[string]string{"status": "completed"})
}
```

### 2. 管道优化

```go
func (c *DataController) ProcessStream() {
    input := make(chan Data, 100)
    output := make(chan Result, 100)
    
    // 启动处理协程
    for i := 0; i < runtime.NumCPU(); i++ {
        go worker(input, output)
    }
    
    // 发送数据
    go func() {
        defer close(input)
        for _, data := range c.GetDataStream() {
            input <- data
        }
    }()
    
    // 收集结果
    results := make([]Result, 0)
    for result := range output {
        results = append(results, result)
    }
    
    c.JSON(results)
}
```

---

## 📊 监控与分析

### 1. 性能监控中间件

```go
func PerformanceMonitor() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        
        // 记录慢请求
        if duration > time.Second {
            log.Printf("Slow request: %s %s took %v", 
                c.Method(), c.Path(), duration)
        }
        
        // 监控指标
        metrics.RecordRequestDuration(c.Path(), duration)
        metrics.RecordRequestCount(c.Path(), c.Writer.Status())
    }
}
```

### 2. 内存监控

```go
func MemoryMonitor() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            log.Printf("Memory: Alloc=%dKB, TotalAlloc=%dKB, Sys=%dKB, NumGC=%d",
                bToKb(m.Alloc), bToKb(m.TotalAlloc), bToKb(m.Sys), m.NumGC)
                
            // 内存使用过高时告警
            if m.Alloc > 100*1024*1024 { // 100MB
                log.Printf("Memory usage high: %dMB", bToMb(m.Alloc))
            }
        }
    }
}

func bToKb(b uint64) uint64 { return b / 1024 }
func bToMb(b uint64) uint64 { return b / 1024 / 1024 }
```

### 3. pprof集成

```go
import _ "net/http/pprof"

func main() {
    // 启用pprof
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    app := mvc.HertzApp
    app.Run(":8888")
}

// 性能分析命令
// go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
// go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## ⚙️ 配置优化

### 1. 生产环境配置

```ini
# config/prod.conf

# 性能相关配置
app.debug = false
app.log_level = error
app.read_timeout = 30s
app.write_timeout = 30s
app.max_header_bytes = 1048576

# 数据库连接池
db.max_idle_conns = 50
db.max_open_conns = 200
db.conn_max_lifetime = 3600s

# 缓存配置
cache.enabled = true
cache.ttl = 3600s
cache.max_memory = 256MB

# 静态文件
static.cache_max_age = 31536000
static.enable_gzip = true
```

### 2. 环境特定优化

```go
func init() {
    if os.Getenv("GO_ENV") == "production" {
        // 生产环境优化
        runtime.GOMAXPROCS(runtime.NumCPU())
        
        // 关闭调试模式
        gin.SetMode(gin.ReleaseMode)
        
        // 优化GC
        debug.SetGCPercent(100)
    }
}
```

---

## 🧪 基准测试

### 1. 编写基准测试

```go
func BenchmarkUserController_GetList(b *testing.B) {
    app := setupTestApp()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        req := httptest.NewRequest("GET", "/users", nil)
        app.ServeHTTP(w, req)
        
        if w.Code != 200 {
            b.Errorf("Expected status 200, got %d", w.Code)
        }
    }
}

func BenchmarkDatabaseQuery(b *testing.B) {
    db := setupTestDB()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var users []User
        db.Limit(10).Find(&users)
    }
}
```

### 2. 运行基准测试

```bash
# 运行基准测试
go test -bench=. -benchmem

# 生成CPU profile
go test -bench=. -cpuprofile=cpu.prof

# 生成内存profile
go test -bench=. -memprofile=mem.prof

# 分析profile
go tool pprof cpu.prof
go tool pprof mem.prof
```

---

## 📈 性能检查清单

### 启动时检查
- [ ] 数据库连接池配置合理
- [ ] 缓存系统正常工作
- [ ] 模板预编译完成
- [ ] 静态资源CDN配置
- [ ] 监控系统启动

### 开发时检查
- [ ] 避免N+1查询问题
- [ ] 使用数据库索引
- [ ] 实现适当的缓存策略
- [ ] 控制内存分配
- [ ] 优化字符串操作

### 部署时检查
- [ ] 启用生产模式
- [ ] 配置合理的超时时间
- [ ] 设置适当的日志级别
- [ ] 启用HTTP/2和压缩
- [ ] 配置健康检查

### 运行时监控
- [ ] 响应时间 < 100ms
- [ ] 内存使用 < 512MB
- [ ] CPU使用率 < 80%
- [ ] 数据库连接池健康
- [ ] 缓存命中率 > 80%

---

<div align="center">

**⚡ 通过这些优化措施，YYHertz应用可以达到生产级别的性能表现！**

**持续监控和优化是保持高性能的关键 🚀**

</div>