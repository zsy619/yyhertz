# 高性能对象映射器 (Object Mapper)

## 概述

**YYHertz 高性能对象映射器**是一个现代化的 Go 语言对象映射库，提供多种映射策略和极致性能优化。支持结构体到结构体的字段映射，具有智能策略选择、类型缓存、性能监控等高级特性。

### ✨ 核心特性

- 🚀 **多策略映射**：自动选择、反射映射、JSON转换、代码生成四种策略
- ⚡ **极致性能**：智能缓存、零反射开销的代码生成策略
- 🔧 **灵活配置**：字段标签映射、忽略字段、自定义转换器
- 📊 **性能监控**：详细的映射统计和性能指标
- 💾 **智能缓存**：类型信息缓存、LRU淘汰策略、后台清理
- 🛡️ **类型安全**：使用 Go 1.18+ 的 `any` 类型，更好的类型安全

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/YYHertz/framework/util/mapper"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Age  int    `json:"age"`
}

type UserDTO struct {
    ID       int    `json:"id"`
    FullName string `json:"name"`
    Age      int    `json:"age"`
}

func main() {
    // 创建映射器
    m := mapper.NewMapper()
    
    // 源对象
    user := &User{
        ID:   1,
        Name: "张三",
        Age:  25,
    }
    
    // 目标对象
    var userDTO UserDTO
    
    // 执行映射
    if err := m.Map(user, &userDTO); err != nil {
        fmt.Printf("映射失败: %v\n", err)
        return
    }
    
    fmt.Printf("映射结果: %+v\n", userDTO)
    // 输出: 映射结果: {ID:1 FullName:张三 Age:25}
}
```

### 高级配置

```go
// 使用自定义配置
m := mapper.NewMapper(
    mapper.WithStrategy(mapper.StrategyJSON),     // 指定JSON策略
    mapper.WithTagName("json"),                   // 使用json标签
    mapper.WithIgnoreFields("Password"),          // 忽略Password字段
    mapper.WithCaseSensitive(false),             // 大小写不敏感
)

// 添加自定义类型转换器
m := mapper.NewMapper(
    mapper.WithConverter(reflect.TypeOf(time.Time{}), func(src any) (any, error) {
        if t, ok := src.(time.Time); ok {
            return t.Format("2006-01-02 15:04:05"), nil
        }
        return src, nil
    }),
)

// 批量映射切片
var users []User
var userDTOs []UserDTO

if err := m.MapSlice(users, &userDTOs); err != nil {
    fmt.Printf("切片映射失败: %v\n", err)
}
```

## 映射策略

### 1. 自动策略 (StrategyAuto) - **推荐**

根据数据结构复杂度自动选择最优策略：
- 简单结构体 → JSON策略 (高性能)
- 复杂结构体 → 反射策略 (功能完整)
- 已有代码生成 → 代码生成策略 (零开销)

```go
m := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyAuto))
```

### 2. JSON策略 (StrategyJSON) - **高性能**

通过JSON序列化/反序列化实现映射，在简单结构体映射时性能最佳：

**性能特点**：
- ✅ 简单结构体: **最快**
- ⚠️  复杂嵌套结构体: 性能下降
- ❌ 不支持某些复杂类型 (channel, func等)

```go
m := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyJSON))
```

### 3. 反射策略 (StrategyReflection) - **功能完整**

基于反射实现的全功能映射器，支持所有Go类型：

**功能特点**：
- ✅ 支持所有Go类型
- ✅ 复杂嵌套结构体支持
- ✅ 自定义类型转换
- ⚠️  性能相对较低

```go
m := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyReflection))
```

### 4. 代码生成策略 (StrategyCodegen) - **零开销**

编译期生成映射代码，运行时零反射开销：

**特点**：
- ✅ **零反射开销**，性能最高
- ✅ 编译期类型检查
- ⚠️  需要预先注册类型对
- ❌ 动态类型支持有限

```go
m := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyCodegen))

// 预注册类型对
m.PrecompileMappers([]mapper.TypePair{
    {reflect.TypeOf(User{}), reflect.TypeOf(UserDTO{})},
})
```

## 性能基准

基于 `BenchmarkMapper*` 测试结果：

| 策略 | 性能 | 内存分配 | 适用场景 |
|------|------|----------|----------|
| **JSON策略** | ~1000 ns/op | 低 | 简单结构体映射 |
| **反射策略** | ~2000 ns/op | 中等 | 复杂结构体映射 |
| **代码生成策略** | ~500 ns/op | 极低 | 高频映射场景 |

## 配置选项

### 基础配置

```go
config := &mapper.MapConfig{
    Strategy:      mapper.StrategyAuto,  // 映射策略
    TagName:       "json",               // 标签名称
    CaseSensitive: false,               // 大小写敏感
    DeepCopy:      true,                // 深拷贝模式
    ZeroFields:    true,                // 映射零值字段
    MaxDepth:      10,                  // 最大递归深度
    IgnoreFields:  []string{"Password"}, // 忽略字段
}

m := mapper.NewMapper(mapper.WithConfig(config))
```

### 链式配置

```go
config := mapper.DefaultMapConfig().
    WithStrategy(mapper.StrategyJSON).
    WithTagName("json").
    WithIgnoreFields("Password", "Secret").
    WithCaseSensitive(false).
    WithMaxDepth(5)

m := mapper.NewMapper(mapper.WithConfig(config))
```

## 性能监控

### 获取统计信息

```go
m := mapper.NewMapper()

// 执行一些映射操作...
for i := 0; i < 1000; i++ {
    m.Map(user, &userDTO)
}

// 获取统计信息
stats := m.GetStats()
fmt.Printf("总映射次数: %d\n", stats.TotalMaps)
fmt.Printf("成功映射次数: %d\n", stats.SuccessfulMaps)
fmt.Printf("失败映射次数: %d\n", stats.FailedMaps)
fmt.Printf("平均执行时间: %d ns\n", stats.AverageTime)
fmt.Printf("缓存命中率: %.2f%%\n", 
    float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses)*100)
```

### 缓存统计

```go
// 获取缓存统计（需要访问内部实现）
cacheStats := mapper.GetCacheStats()
fmt.Printf("缓存命中率: %.2f%%\n", cacheStats.HitRatio*100)
fmt.Printf("结构体信息缓存: %d 条\n", cacheStats.StructInfoCount)
fmt.Printf("字段映射缓存: %d 条\n", cacheStats.FieldMappingCount)
```

## 高级特性

### 自定义类型转换器

```go
import "time"

// 时间格式转换
timeConverter := func(src any) (any, error) {
    if t, ok := src.(time.Time); ok {
        return t.Format("2006-01-02 15:04:05"), nil
    }
    return src, nil
}

// 字符串到整数转换
stringToIntConverter := func(src any) (any, error) {
    if s, ok := src.(string); ok {
        return strconv.Atoi(s)
    }
    return src, nil
}

m := mapper.NewMapper(
    mapper.WithConverter(reflect.TypeOf(time.Time{}), timeConverter),
    mapper.WithConverter(reflect.TypeOf(""), stringToIntConverter),
)
```

### 标签映射

支持多种标签格式：

```go
type User struct {
    ID       int    `json:"id" xml:"user_id"`
    Name     string `json:"name" xml:"user_name"`
    Email    string `json:"email" xml:"email_addr"`
    Password string `json:"-"`              // 忽略字段
    Internal string `json:"internal,omitempty"` // 零值时忽略
}

// 使用json标签
m1 := mapper.NewMapper(mapper.WithTagName("json"))

// 使用xml标签
m2 := mapper.NewMapper(mapper.WithTagName("xml"))
```

### 切片和嵌套结构体映射

```go
type Department struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type Employee struct {
    ID         int        `json:"id"`
    Name       string     `json:"name"`
    Department Department `json:"department"`
    Skills     []string   `json:"skills"`
}

type EmployeeDTO struct {
    ID         int           `json:"id"`
    FullName   string        `json:"name"`
    Department DepartmentDTO `json:"department"`
    Skills     []string      `json:"skills"`
}

// 自动处理嵌套结构体和切片
var employee Employee
var employeeDTO EmployeeDTO

m := mapper.NewMapper()
err := m.Map(&employee, &employeeDTO)
```

## 错误处理

### 常见错误类型

```go
err := m.Map(src, dst)
if err != nil {
    switch {
    case strings.Contains(err.Error(), "destination must be a pointer"):
        // 目标对象必须是指针
        fmt.Println("目标对象必须传入指针")
    
    case strings.Contains(err.Error(), "source cannot be nil"):
        // 源对象不能为nil
        fmt.Println("源对象不能为空")
    
    case strings.Contains(err.Error(), "type conversion failed"):
        // 类型转换失败
        fmt.Println("字段类型转换失败")
    
    case strings.Contains(err.Error(), "maximum mapping depth exceeded"):
        // 超过最大映射深度（防止循环引用）
        fmt.Println("映射深度超限，可能存在循环引用")
    
    default:
        fmt.Printf("映射失败: %v\n", err)
    }
}
```

### 错误预防

```go
// 1. 确保目标是指针
var dst TargetType
err := m.Map(src, &dst) // ✅ 正确：传入指针

// 2. 检查源对象
if src == nil {
    return errors.New("源对象不能为空")
}

// 3. 设置合理的最大深度
m := mapper.NewMapper(mapper.WithMaxDepth(20))

// 4. 使用类型断言进行预检查
if reflect.TypeOf(dst).Kind() != reflect.Ptr {
    return errors.New("目标必须是指针类型")
}
```

## 最佳实践

### 1. 策略选择指南

```go
// 简单DTO映射 - 使用JSON策略
type UserBasic struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}
m := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyJSON))

// 复杂业务对象映射 - 使用反射策略
type UserComplex struct {
    Profile   UserProfile
    Settings  map[string]interface{}
    CreatedAt time.Time
}
m := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyReflection))

// 高频映射场景 - 使用代码生成策略
m := mapper.NewMapper(mapper.WithStrategy(mapper.StrategyCodegen))
```

### 2. 性能优化

```go
// 复用映射器实例
var globalMapper = mapper.NewMapper(
    mapper.WithStrategy(mapper.StrategyAuto),
    mapper.WithTagName("json"),
)

// 预热缓存
func init() {
    // 执行一次映射来预热缓存
    var dummy1 SourceType
    var dummy2 TargetType
    globalMapper.Map(&dummy1, &dummy2)
}

// 批量操作使用MapSlice
func ConvertUsers(users []User) ([]UserDTO, error) {
    var userDTOs []UserDTO
    err := globalMapper.MapSlice(users, &userDTOs)
    return userDTOs, err
}
```

### 3. 内存管理

```go
// 对于大量小对象映射，考虑对象池
var userDTOPool = sync.Pool{
    New: func() interface{} {
        return &UserDTO{}
    },
}

func ConvertUser(user *User) *UserDTO {
    dto := userDTOPool.Get().(*UserDTO)
    defer userDTOPool.Put(dto)
    
    if err := globalMapper.Map(user, dto); err != nil {
        return nil
    }
    
    // 返回副本，避免池对象被修改
    result := *dto
    return &result
}
```

### 4. 错误处理最佳实践

```go
func SafeMap(m mapper.Mapper, src, dst any) error {
    // 参数验证
    if src == nil {
        return errors.New("源对象不能为空")
    }
    
    dstValue := reflect.ValueOf(dst)
    if dstValue.Kind() != reflect.Ptr {
        return errors.New("目标对象必须是指针")
    }
    
    if dstValue.IsNil() {
        return errors.New("目标对象指针不能为空")
    }
    
    // 执行映射
    if err := m.Map(src, dst); err != nil {
        return fmt.Errorf("对象映射失败: %w", err)
    }
    
    return nil
}
```

## 测试

### 运行基准测试

```bash
# 运行所有基准测试
go test -bench=. ./framework/util/mapper

# 运行内存分析
go test -bench=. -benchmem ./framework/util/mapper

# 运行性能对比测试
go test -bench=BenchmarkMapper -benchtime=5s ./framework/util/mapper
```

### 示例输出

```
BenchmarkMapperJSON-8           1000000    1043 ns/op    312 B/op    8 allocs/op
BenchmarkMapperReflection-8      500000    2156 ns/op    456 B/op   12 allocs/op  
BenchmarkMapperCodegen-8        2000000     542 ns/op    128 B/op    3 allocs/op
BenchmarkMapperAuto-8            800000    1289 ns/op    278 B/op    7 allocs/op
```

## API 参考

### 核心接口

```go
// Mapper 映射器接口
type Mapper interface {
    Map(src, dst any) error
    MapWithConfig(src, dst any, config *MapConfig) error  
    MapSlice(src, dst any) error
    SetStrategy(strategy MappingStrategy)
    GetStats() *MapperStats
}

// 创建映射器
func NewMapper(options ...MapperOption) Mapper

// 全局映射函数
func Map(src, dst any) error
func MapWithConfig(src, dst any, config *MapConfig) error
func MapSlice(src, dst any) error
```

### 配置选项

```go
func WithStrategy(strategy MappingStrategy) MapperOption
func WithTagName(tagName string) MapperOption
func WithIgnoreFields(fields ...string) MapperOption
func WithCaseSensitive(sensitive bool) MapperOption
func WithDeepCopy(deep bool) MapperOption
func WithConverter(srcType reflect.Type, converter TypeConverter) MapperOption
```

### 配置构建器

```go
func DefaultMapConfig() *MapConfig
func (c *MapConfig) WithStrategy(strategy MappingStrategy) *MapConfig
func (c *MapConfig) WithIgnoreFields(fields ...string) *MapConfig
func (c *MapConfig) WithTagName(tag string) *MapConfig
func (c *MapConfig) WithCaseSensitive(sensitive bool) *MapConfig
func (c *MapConfig) WithMaxDepth(depth int) *MapConfig
```

## 常见问题 (FAQ)

### Q: 如何选择合适的映射策略？

**A**: 根据场景选择：
- **简单DTO映射**: `StrategyJSON` - 性能最佳
- **复杂业务对象**: `StrategyReflection` - 功能最全
- **高频调用场景**: `StrategyCodegen` - 零开销
- **不确定场景**: `StrategyAuto` - 自动选择

### Q: 为什么映射失败了？

**A**: 检查以下几点：
1. 目标对象是否传入了指针 `&dst`
2. 字段名称和标签是否匹配
3. 字段类型是否兼容
4. 是否存在循环引用（设置合理的MaxDepth）

### Q: 如何提升映射性能？

**A**: 性能优化建议：
1. 复用映射器实例，避免重复创建
2. 选择合适的映射策略
3. 使用`MapSlice`进行批量操作
4. 预热缓存（执行一次映射）
5. 考虑使用对象池管理大量小对象

### Q: 支持哪些字段类型？

**A**: 支持所有Go基础类型及复合类型：
- 基础类型: int, string, bool, float等
- 复合类型: struct, slice, array, map, pointer
- 特殊类型: time.Time, interface{}
- 自定义类型通过TypeConverter支持

### Q: 如何处理不同的字段名？

**A**: 使用结构体标签：
```go
type User struct {
    ID   int    `json:"user_id"`     // 映射到user_id
    Name string `json:"full_name"`   // 映射到full_name
}
```

### Q: 如何忽略某些字段？

**A**: 两种方式：
```go
// 1. 使用配置忽略
m := mapper.NewMapper(mapper.WithIgnoreFields("Password", "Internal"))

// 2. 使用标签忽略  
type User struct {
    Password string `json:"-"`  // 忽略此字段
}
```

## 许可证

本项目基于 MIT 许可证开源。详见 [LICENSE](LICENSE) 文件。

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支: `git checkout -b feature/your-feature`
3. 提交更改: `git commit -am 'Add some feature'`
4. 推送分支: `git push origin feature/your-feature`
5. 提交 Pull Request

## 支持

如果遇到问题或需要帮助：

- 📖 查看 [文档](./docs/)
- 🐛 提交 [Issue](https://github.com/YYHertz/framework/issues)
- 💬 参与 [讨论](https://github.com/YYHertz/framework/discussions)

---

**YYHertz 高性能对象映射器** - 让Go语言对象映射变得简单高效！ 🚀