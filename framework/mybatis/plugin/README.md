# MyBatis 插件系统

MyBatis 插件系统为框架提供了强大的扩展能力，允许在 SQL 执行的各个阶段进行拦截和处理。

## 🚀 功能特性

### 核心插件

1. **分页插件 (PaginationPlugin)**
   - 自动处理分页查询
   - 支持多种数据库方言
   - 智能参数提取和验证
   - 分页结果包装

2. **性能监控插件 (PerformancePlugin)**
   - SQL 执行性能监控
   - 慢查询检测和记录
   - 并发统计和性能报告
   - 详细的性能指标收集

3. **SQL日志插件 (SqlLogPlugin)**
   - 详细的 SQL 执行日志记录
   - 可配置的日志级别和格式
   - 参数和结果记录
   - 多种日志格式化器

4. **缓存增强插件 (CacheEnhancerPlugin)**
   - 高级缓存功能
   - 缓存统计和监控
   - LRU缓存策略
   - 缓存命中率统计

5. **参数验证插件 (ValidatorPlugin)**
   - 输入参数自动验证
   - 多种内置验证器（必需、长度、范围、正则、邮箱、手机号等）
   - 自定义验证规则支持
   - 结构体和Map验证

6. **结果转换插件 (ResultTransformerPlugin)**
   - 查询结果自动转换
   - 多种数据格式支持（Map、JSON、字符串、数字、时间）
   - 命名风格转换（驼峰、下划线）
   - 自定义转换器支持

## 📁 文件结构

```
framework/mybatis/plugin/
├── plugin.go              # 插件系统核心接口和基础实现
├── manager.go              # 插件管理器
├── pagination.go           # 分页插件
├── performance.go          # 性能监控插件
├── sqllog.go              # SQL日志插件
├── cache_enhancer.go      # 缓存增强插件
├── validator.go           # 参数验证插件
├── result_transformer.go  # 结果转换插件
├── example.go             # 使用示例
└── README.md              # 文档说明
```

## 🔧 快速开始

### 基本使用

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mybatis/config"
    "github.com/zsy619/yyhertz/framework/mybatis/plugin"
)

func main() {
    // 创建配置
    configuration := config.NewConfiguration()
    
    // 创建插件管理器
    manager := plugin.NewPluginManager(configuration)
    
    // 配置分页插件
    manager.ConfigurePlugin("pagination", map[string]any{
        "defaultPageSize": 20,
        "maxPageSize": 1000,
        "dialect": "mysql",
    })
    
    // 配置性能监控插件
    manager.ConfigurePlugin("performance", map[string]any{
        "slowQueryThreshold": 1000, // 1秒
        "enableMetrics": true,
    })
    
    // 启用插件
    manager.EnablePlugin("pagination")
    manager.EnablePlugin("performance")
    manager.EnablePlugin("sqllog")
    
    // 应用插件到目标对象
    target := &MyMapper{}
    proxiedTarget := manager.ApplyPlugins(target)
    
    // 使用带插件的对象
    result, err := proxiedTarget.SelectUsers(pageRequest)
}
```

### 分页插件使用

```go
// 分页请求
pageRequest := &plugin.PageRequest{
    PageNum:  1,
    PageSize: 20,
    OrderBy:  "created_at DESC",
}

// 执行分页查询
result, err := sqlSession.SelectList("selectUsers", pageRequest)

// 获取分页结果
if pageResult, ok := result.(*plugin.PageResult); ok {
    fmt.Printf("总记录数: %d\n", pageResult.Total)
    fmt.Printf("总页数: %d\n", pageResult.Pages)
    fmt.Printf("当前页: %d\n", pageResult.PageNum)
    fmt.Printf("数据: %v\n", pageResult.List)
}
```

### 参数验证插件使用

```go
// 创建验证插件
validator := plugin.NewValidatorPlugin()

// 添加验证规则
validator.AddRule("insertUser", plugin.ValidationRule{
    Field:    "name",
    Type:     "required",
    Required: true,
    Message:  "用户名不能为空",
})

validator.AddRule("insertUser", plugin.ValidationRule{
    Field:   "email",
    Type:    "email",
    Message: "邮箱格式不正确",
})

validator.AddRule("insertUser", plugin.ValidationRule{
    Field: "age",
    Type:  "range",
    Params: map[string]any{
        "min": 18,
        "max": 100,
    },
    Message: "年龄必须在18-100之间",
})

// 注册插件
manager.RegisterPlugin(validator)
```

### 性能监控插件使用

```go
// 创建性能监控插件
performance := plugin.NewPerformancePlugin()

// 配置慢查询阈值
performance.SetProperties(map[string]any{
    "slowQueryThreshold": 500, // 500毫秒
    "enableMetrics": true,
})

// 获取性能报告
report := performance.GetPerformanceReport()
fmt.Printf("性能报告: %+v\n", report)

// 获取慢查询记录
slowQueries := performance.GetSlowQueries()
for _, query := range slowQueries {
    fmt.Printf("慢查询: %s, 执行时间: %v\n", query.SQL, query.ExecutionTime)
}
```

### 缓存增强插件使用

```go
// 创建缓存增强插件
cachePlugin := plugin.NewCacheEnhancerPlugin()

// 配置缓存
cachePlugin.SetProperties(map[string]any{
    "enableStatistics": true,
    "enablePreload": false,
})

// 获取缓存统计
stats := cachePlugin.GetCacheStatistics()
fmt.Printf("缓存命中率: %.2f%%\n", stats.HitRate*100)
fmt.Printf("缓存未命中率: %.2f%%\n", stats.MissRate*100)

// 获取缓存报告
report := cachePlugin.GetCacheReport()
fmt.Printf("缓存报告: %+v\n", report)
```

### 结果转换插件使用

```go
// 创建结果转换插件
transformer := plugin.NewResultTransformerPlugin()

// 添加转换规则 - 将结果转换为Map
transformer.AddRule("selectUser", plugin.TransformRule{
    FromType: "struct",
    ToType:   "map",
    Method:   "map",
})

// 添加转换规则 - 将字段名转换为驼峰命名
transformer.AddRule("selectUsers", plugin.TransformRule{
    FromType: "map",
    ToType:   "map",
    Method:   "camelCase",
})

// 注册插件
manager.RegisterPlugin(transformer)
```

## 🛠️ 自定义插件开发

### 创建自定义插件

```go
// 自定义插件结构
type CustomPlugin struct {
    *plugin.BasePlugin
    customProperty string
}

// 创建自定义插件
func NewCustomPlugin() *CustomPlugin {
    return &CustomPlugin{
        BasePlugin: plugin.NewBasePlugin("custom", 10),
        customProperty: "default",
    }
}

// 实现拦截方法
func (p *CustomPlugin) Intercept(invocation *plugin.Invocation) (any, error) {
    // 前置处理
    fmt.Printf("执行前: %s\n", invocation.Method.Name)
    
    // 执行原方法
    result, err := invocation.Proceed()
    
    // 后置处理
    fmt.Printf("执行后: %s\n", invocation.Method.Name)
    
    return result, err
}

// 实现包装方法
func (p *CustomPlugin) Plugin(target any) any {
    return target
}

// 实现属性设置方法
func (p *CustomPlugin) SetProperties(properties map[string]any) {
    p.BasePlugin.SetProperties(properties)
    
    if prop, exists := properties["customProperty"]; exists {
        if str, ok := prop.(string); ok {
            p.customProperty = str
        }
    }
}
```

### 注册和使用自定义插件

```go
// 创建并注册自定义插件
customPlugin := NewCustomPlugin()
manager.RegisterPlugin(customPlugin)

// 配置插件属性
manager.ConfigurePlugin("custom", map[string]any{
    "customProperty": "custom_value",
})

// 启用插件
manager.EnablePlugin("custom")
```

## 📊 插件执行顺序

插件按照优先级顺序执行，优先级越小越先执行：

1. **参数验证插件** (优先级: 1) - 最先执行，验证输入参数
2. **性能监控插件** (优先级: 2) - 监控整个执行过程
3. **分页插件** (优先级: 3) - 处理分页逻辑
4. **缓存增强插件** (优先级: 4) - 缓存处理
5. **SQL日志插件** (优先级: 5) - 记录SQL执行日志
6. **结果转换插件** (优先级: 6) - 最后执行，转换结果格式

## 🔍 调试和监控

### 启用调试模式

```go
// 启用插件调试模式
manager.SetDebugMode(true)

// 查看插件执行链
chain := manager.GetExecutionChain()
for _, plugin := range chain {
    fmt.Printf("插件: %s, 优先级: %d, 状态: %s\n", 
        plugin.GetName(), plugin.GetOrder(), plugin.GetStatus())
}
```

### 性能监控

```go
// 获取所有插件的性能统计
stats := manager.GetPluginStatistics()
for name, stat := range stats {
    fmt.Printf("插件 %s: 执行次数=%d, 平均耗时=%v\n", 
        name, stat.ExecutionCount, stat.AvgExecutionTime)
}
```

## 📝 配置文件支持

支持通过配置文件配置插件：

```yaml
# mybatis-plugins.yml
plugins:
  pagination:
    enabled: true
    properties:
      defaultPageSize: 20
      maxPageSize: 1000
      dialect: mysql
      
  performance:
    enabled: true
    properties:
      slowQueryThreshold: 1000
      enableMetrics: true
      
  validator:
    enabled: true
    properties:
      enableValidation: true
      
  cache_enhancer:
    enabled: false
    properties:
      enableStatistics: true
      enablePreload: false
```

```go
// 从配置文件加载插件配置
manager.LoadConfigFromFile("mybatis-plugins.yml")
```

## 🚨 注意事项

1. **插件顺序**: 插件执行顺序很重要，确保按正确的优先级配置
2. **性能影响**: 过多的插件可能影响性能，建议只启用必要的插件
3. **异常处理**: 插件中的异常会中断执行链，需要妥善处理
4. **线程安全**: 插件需要考虑线程安全问题
5. **资源管理**: 及时释放插件占用的资源

## 🤝 贡献指南

欢迎贡献新的插件或改进现有插件：

1. Fork 项目
2. 创建功能分支
3. 实现插件功能
4. 添加测试用例
5. 提交 Pull Request

## 📄 许可证

本项目采用 MIT 许可证，详见 LICENSE 文件。