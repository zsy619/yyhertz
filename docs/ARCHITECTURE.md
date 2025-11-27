# YYHertz 架构设计文档

<div align="center">

🏛️ **深入理解YYHertz框架架构** | 设计理念与实现原理

</div>

---

## 🎯 设计理念

### 核心原则
- **🏗️ 分层架构**: 清晰的MVC分层，职责分离
- **📦 模块化**: 功能模块独立，低耦合高内聚
- **⚡ 高性能**: 基于CloudWeGo-Hertz的高性能引擎
- **🔄 可扩展**: 插件化设计，易于扩展和定制
- **🛡️ 安全第一**: 内置多层安全防护机制

---

## 🏛️ 整体架构

### 分层架构图
```
┌─────────────────── 表示层 (Presentation Layer) ───────────────────┐
│  Web界面  │  RESTful API  │  WebSocket  │  静态资源服务        │
├─────────────────── 控制层 (Controller Layer) ─────────────────────┤
│  路由分发  │  请求处理  │  参数验证  │  响应格式化  │  中间件管道  │
├─────────────────── 业务层 (Service Layer) ──────────────────────────┤
│  业务逻辑  │  数据转换  │  事务管理  │  缓存策略  │  权限控制    │
├─────────────────── 数据层 (Data Layer) ────────────────────────────┤
│  ORM抽象  │  数据库操作  │  缓存操作  │  消息队列  │  文件存储   │
└─────────────────── 基础设施层 (Infrastructure) ──────────────────────┘
```

### 核心组件交互
```
                    ┌─────────────┐
                    │   Client    │
                    └──────┬──────┘
                           │ HTTP Request
                    ┌──────▼──────┐
                    │  HTTP服务器  │ (CloudWeGo-Hertz)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ 中间件管道   │ (Recovery, Logger, Auth...)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  路由分发   │ (AutoRouter + Manual)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │   控制器   │ (BaseController)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │  业务服务   │ (Service Layer)
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ 数据访问层  │ (Repository Pattern)
                    └──────┬──────┘
                           │
            ┌──────────────▼──────────────┐
            │        数据存储层           │
            │  Database │ Cache │ Queue  │
            └─────────────────────────────┘
```

---

## 🔧 核心模块详解

### 1. MVC核心模块
```go
// MVC架构的核心抽象
type Application struct {
    *hertz.Hertz                    // HTTP引擎
    controllers []ControllerInterface // 控制器注册表
    middlewares []HandlerFunc        // 中间件栈
    config      *Config             // 配置管理
    logger      *Logger             // 日志组件
}

// 控制器接口定义
type ControllerInterface interface {
    Prepare()  // 前置处理
    Finish()   // 后置处理
}

// 基础控制器实现
type BaseController struct {
    Ctx    *Context    // 请求上下文
    Data   M           // 模板数据
    Logger *Logger     // 日志组件
}
```

### 2. 路由系统设计
```go
// 路由管理器
type Router struct {
    engine      *hertz.Hertz        // HTTP引擎
    routes      map[string]*Route   // 路由映射表
    middlewares []HandlerFunc       // 全局中间件
    groups      []*RouterGroup      // 路由组
}

// 自动路由注册
func (r *Router) AutoRouter(controller ControllerInterface) {
    // 通过反射分析控制器方法
    // HTTP方法 + 控制器名 + 方法名 = 路由路径
    // 例: GET + User + Index = GET /user/index
}

// 命名空间路由
type Namespace struct {
    prefix      string              // 路径前缀
    middlewares []HandlerFunc       // 命名空间中间件
    controllers []ControllerInterface
}
```

### 3. 中间件系统
```go
// 中间件函数签名
type HandlerFunc func(*Context)

// 中间件管道实现
type MiddlewarePipeline struct {
    handlers []HandlerFunc
    index    int
}

func (m *MiddlewarePipeline) Next(c *Context) {
    for m.index < len(m.handlers) {
        handler := m.handlers[m.index]
        m.index++
        handler(c)
        
        if c.IsAborted() {
            return
        }
    }
}

// 4层中间件架构
// Layer 1: 请求层 - Recovery, Logger, CORS
// Layer 2: 安全层 - Auth, CSRF, Rate Limit
// Layer 3: 业务层 - Validation, Cache
// Layer 4: 响应层 - Compression, Headers
```

---

## 🗄️ 数据层架构

### 智能ORM双引擎
```go
// 智能选择器接口
type SmartSelector interface {
    Create(value interface{}) error
    Find(dest interface{}, conditions ...interface{}) error
    Update(value interface{}) error
    Delete(value interface{}) error
    ExecuteComplexQuery(queryID string, params map[string]interface{}) ([]map[string]interface{}, error)
}

// 引擎选择策略
type EngineSelector struct {
    gormEngine    *GORMEngine     // GORM引擎 (简单CRUD)
    mybatisEngine *MyBatisEngine  // MyBatis引擎 (复杂查询)
    analyzer      *QueryAnalyzer  // 查询分析器
}

func (s *EngineSelector) SelectEngine(operation Operation) Engine {
    complexity := s.analyzer.AnalyzeComplexity(operation)
    
    if complexity.Score < SIMPLE_THRESHOLD {
        return s.gormEngine  // 简单操作使用GORM
    }
    
    return s.mybatisEngine   // 复杂操作使用MyBatis
}
```

### 缓存架构
```go
// 多级缓存系统
type CacheManager struct {
    l1Cache MemoryCache    // L1: 进程内缓存
    l2Cache RedisCache     // L2: 分布式缓存
    l3Cache DatabaseCache  // L3: 数据库查询缓存
}

// 缓存策略接口
type CacheStrategy interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl time.Duration) error
    Delete(key string) error
    Clear() error
}
```

---

## 🎨 模板系统架构

### 模板引擎设计
```go
// 模板引擎接口
type TemplateEngine interface {
    LoadTemplates(pattern string) error
    Render(name string, data interface{}) ([]byte, error)
    AddFunc(name string, fn interface{})
    SetDelims(left, right string)
}

// 模板功能增强
type EnhancedTemplate struct {
    *template.Template
    funcMap     template.FuncMap    // 150+自定义函数
    cache       *TemplateCache      // 模板缓存
    components  map[string]*Component // 组件系统
    layouts     map[string]*Layout   // 布局系统
}

// 组件化模板
type Component struct {
    Name     string
    Template *template.Template
    Props    map[string]interface{}
    Slots    map[string]string
}
```

### 函数冲突解决机制
```go
// 安全函数映射表
var SafeFunctionMap = template.FuncMap{
    // Go内置冲突函数的安全别名
    "equal":        Eq,          // 替代 'eq'
    "notEqual":     Ne,          // 替代 'ne'
    "lessThan":     Lt,          // 替代 'lt'
    "length":       Len,         // 替代 'len'
    "templatefunc": TemplateInclude, // 替代 'template'
    
    // 保留原函数以维持向后兼容
    "eq": Eq,    // 保留但标记为deprecated
    "ne": Ne,    // 保留但标记为deprecated
}

// 冲突检测器
type ConflictDetector struct {
    goBuiltins []string
    userFuncs  map[string]interface{}
}

func (c *ConflictDetector) DetectConflicts() []Conflict {
    // 检测并报告命名冲突
    // 提供自动修复建议
}
```

---

## 🔌 插件系统架构

### 插件接口设计
```go
// 插件生命周期接口
type Plugin interface {
    Name() string
    Version() string
    Init(app *Application) error
    Start() error
    Stop() error
}

// 插件管理器
type PluginManager struct {
    plugins    map[string]Plugin
    hooks      map[string][]HookFunc
    config     *PluginConfig
}

// 钩子系统
type HookFunc func(ctx *Context, args ...interface{}) error

func (pm *PluginManager) RegisterHook(event string, hook HookFunc) {
    pm.hooks[event] = append(pm.hooks[event], hook)
}

// 内置钩子事件
const (
    HookBeforeRequest  = "before_request"
    HookAfterRequest   = "after_request"
    HookBeforeResponse = "before_response"
    HookAfterResponse  = "after_response"
)
```

---

## 🛡️ 安全架构

### 多层安全防护
```go
// 安全管理器
type SecurityManager struct {
    csrfProtector    *CSRFProtector
    xssFilter        *XSSFilter
    sqlInjectionGuard *SQLInjectionGuard
    rateLimiter      *RateLimiter
    authManager      *AuthManager
}

// 权限控制系统
type RBAC struct {
    users       map[int]*User
    roles       map[int]*Role
    permissions map[int]*Permission
    policies    []Policy
}

// 安全策略接口
type SecurityPolicy interface {
    Evaluate(ctx *Context) (bool, error)
    GetViolationResponse() Response
}
```

---

## 📊 性能监控架构

### 性能指标收集
```go
// 性能监控系统
type PerformanceMonitor struct {
    metrics     map[string]*Metric
    collectors  []MetricCollector
    exporters   []MetricExporter
    alerter     *AlertManager
}

// 指标类型定义
type MetricType int

const (
    CounterMetric   MetricType = iota
    GaugeMetric
    HistogramMetric
    SummaryMetric
)

// 内置性能指标
var BuiltinMetrics = []string{
    "http_requests_total",
    "http_request_duration_seconds",
    "db_connections_active",
    "db_query_duration_seconds",
    "cache_hit_rate",
    "memory_usage_bytes",
    "goroutines_count",
}
```

---

## 🔧 配置管理架构

### 分层配置系统
```go
// 配置管理器
type ConfigManager struct {
    providers []ConfigProvider  // 配置提供者
    watchers  []ConfigWatcher   // 配置监听器
    cache     map[string]interface{}
    mutex     sync.RWMutex
}

// 配置提供者接口
type ConfigProvider interface {
    Load() (map[string]interface{}, error)
    Watch() <-chan ConfigChange
}

// 支持的配置源
type ConfigSources struct {
    File        *FileProvider     // 文件配置
    Environment *EnvProvider      // 环境变量
    Database    *DBProvider       // 数据库配置
    Remote      *RemoteProvider   // 远程配置中心
}

// 配置热更新
type ConfigWatcher interface {
    OnChange(key string, oldValue, newValue interface{})
}
```

---

## 🚀 扩展点设计

### 可扩展的组件接口

```go
// 1. 自定义控制器
type CustomController struct {
    BaseController
    customService *CustomService
}

// 2. 自定义中间件
func CustomMiddleware() HandlerFunc {
    return func(c *Context) {
        // 自定义逻辑
        c.Next()
    }
}

// 3. 自定义ORM引擎
type CustomORMEngine struct {
    // 实现Engine接口
}

// 4. 自定义模板函数
func RegisterCustomTemplateFuncs() template.FuncMap {
    return template.FuncMap{
        "customFunc": func(input string) string {
            // 自定义逻辑
            return processedInput
        },
    }
}

// 5. 自定义缓存策略
type CustomCacheStrategy struct {
    // 实现CacheStrategy接口
}
```

---

## 📈 性能优化设计

### 关键性能特性
- **🔄 连接池管理**: 数据库和Redis连接池优化
- **💾 智能缓存**: 多级缓存策略，提高响应速度
- **📦 对象池化**: 减少GC压力，提高并发性能
- **⚡ 预编译优化**: 模板和SQL语句预编译
- **🎯 批量操作**: 数据库批量插入、更新优化

### 性能监控指标
```go
// 关键性能指标
type PerformanceMetrics struct {
    RequestsPerSecond   float64  // QPS
    AverageResponseTime float64  // 平均响应时间
    P95ResponseTime     float64  // 95分位响应时间
    ErrorRate           float64  // 错误率
    ActiveConnections   int      // 活跃连接数
    MemoryUsage         int64    // 内存使用量
    CPUUsage            float64  // CPU使用率
}
```

---

## 🔮 未来发展方向

### 路线图
1. **微服务支持** - 服务发现、负载均衡、熔断器
2. **云原生特性** - Kubernetes集成、服务网格支持
3. **GraphQL支持** - 现代API查询语言集成
4. **AI集成** - 智能运维、自动优化建议
5. **多协议支持** - gRPC、WebSocket、Server-Sent Events

---

<div align="center">

**🏛️ YYHertz架构设计兼顾了性能、安全、可扩展性**

**深入理解架构，让你的应用开发更加得心应手！🚀**

</div>