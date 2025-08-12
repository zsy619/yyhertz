# AddFuncMap 模板函数添加功能

YYHertz MVC 框架提供了强大的 `AddFuncMap` 功能，允许开发者添加自定义模板函数，扩展模板引擎的功能。

## 🎯 功能特点

- **全局函数注册**: 一次注册，所有模板可用
- **线程安全**: 支持并发环境下的函数管理
- **灵活的API**: 支持任意类型的函数添加
- **自动集成**: 与框架模板系统无缝集成
- **易于使用**: 简单的API设计

## 📚 API 参考

### 静态方法（推荐使用）

```go
// 添加全局模板函数
mvc.AddFuncMap(name string, fn any)

// 获取全局模板函数映射
funcMap := mvc.GetGlobalFuncMap() 

// 移除全局模板函数
mvc.RemoveFuncMap(name string)

// 列出所有已注册的模板函数名称
funcNames := mvc.ListFuncMap()
```

### 应用实例方法

```go
app := mvc.HertzApp

// 添加模板函数
app.AddFuncMap(name string, fn any)

// 获取函数映射
funcMap := app.GetGlobalFuncMap()

// 移除函数
app.RemoveFuncMap(name string)

// 列出函数名称
funcNames := app.ListFuncMap()
```

## 🚀 快速开始

### 1. 基本使用

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/util"
    "strings"
)

func main() {
    // 添加内置工具函数
    mvc.AddFuncMap("containString", util.ContainString)
    
    // 添加自定义字符串函数
    mvc.AddFuncMap("upper", strings.ToUpper)
    mvc.AddFuncMap("lower", strings.ToLower)
    
    // 添加格式化函数
    mvc.AddFuncMap("formatPrice", func(price float64) string {
        return "¥" + util.FmtFloat2(price)
    })
    
    // 启动应用
    app := mvc.HertzApp
    app.AutoRouters(&YourController{})
    app.Run(":8080")
}
```

### 2. 在模板中使用

```html
<!DOCTYPE html>
<html>
<head>
    <title>模板函数示例</title>
</head>
<body>
    <h1>{{upper "hello world"}}</h1>
    
    <!-- 条件判断 -->
    {{if containString .Tags "important"}}
        <div class="important">重要内容</div>
    {{end}}
    
    <!-- 价格格式化 -->
    <p>价格: {{formatPrice .Price}}</p>
    
    <!-- 字符串处理 -->
    <p>用户名: {{upper .Username}}</p>
    <p>小写: {{lower .Description}}</p>
</body>
</html>
```

## 📖 常用函数示例

### 字符串处理函数

```go
// 字符串转换
mvc.AddFuncMap("upper", strings.ToUpper)
mvc.AddFuncMap("lower", strings.ToLower)
mvc.AddFuncMap("title", strings.Title)

// 字符串操作
mvc.AddFuncMap("trim", strings.TrimSpace)
mvc.AddFuncMap("reverse", func(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
})

// 字符串检查
mvc.AddFuncMap("isEmpty", func(s string) bool {
    return strings.TrimSpace(s) == ""
})

mvc.AddFuncMap("contains", func(str, substr string) bool {
    return strings.Contains(str, substr)
})
```

### 数字格式化函数

```go
// 使用框架内置工具
mvc.AddFuncMap("formatFloat2", util.FmtFloat2)
mvc.AddFuncMap("formatByte", util.FmtByte)

// 自定义格式化
mvc.AddFuncMap("formatPrice", func(price float64) string {
    return "¥" + util.FmtFloat2(price)
})

mvc.AddFuncMap("formatPercent", func(value float64) string {
    return util.FmtFloat2(value*100) + "%"
})
```

### 数组/切片处理函数

```go
// 数组连接
mvc.AddFuncMap("join", func(items []string, separator string) string {
    return strings.Join(items, separator)
})

// 数组包含检查
mvc.AddFuncMap("contains", func(items []string, item string) bool {
    for _, v := range items {
        if v == item {
            return true
        }
    }
    return false
})

// 数组长度
mvc.AddFuncMap("length", func(items any) int {
    switch v := items.(type) {
    case []string:
        return len(v)
    case []int:
        return len(v)
    case string:
        return len(v)
    default:
        return 0
    }
})
```

### 时间处理函数

```go
import "time"

// 使用框架内置函数
mvc.AddFuncMap("getTime", util.GetTime)
mvc.AddFuncMap("formatTime", util.FormatTime)

// 自定义时间函数
mvc.AddFuncMap("now", time.Now)
mvc.AddFuncMap("formatDate", func(t time.Time) string {
    return t.Format("2006-01-02")
})

mvc.AddFuncMap("timeAgo", func(t time.Time) string {
    duration := time.Since(t)
    if duration < time.Hour {
        return fmt.Sprintf("%.0f分钟前", duration.Minutes())
    } else if duration < 24*time.Hour {
        return fmt.Sprintf("%.0f小时前", duration.Hours())
    } else {
        return fmt.Sprintf("%.0f天前", duration.Hours()/24)
    }
})
```

## 🛠️ 高级用法

### 1. 函数管理

```go
// 批量添加函数
funcs := map[string]any{
    "upper":   strings.ToUpper,
    "lower":   strings.ToLower,
    "trim":    strings.TrimSpace,
    "reverse": reverseString,
}

for name, fn := range funcs {
    mvc.AddFuncMap(name, fn)
}

// 检查函数是否存在
funcNames := mvc.ListFuncMap()
for _, name := range funcNames {
    fmt.Printf("已注册函数: %s\n", name)
}

// 动态移除函数
if someCondition {
    mvc.RemoveFuncMap("temporaryFunc")
}
```

### 2. 条件注册

```go
// 根据环境注册不同的函数
if config.IsDebug() {
    mvc.AddFuncMap("debug", func(v any) string {
        return fmt.Sprintf("DEBUG: %+v", v)
    })
} else {
    mvc.AddFuncMap("debug", func(v any) string {
        return "" // 生产环境不输出调试信息
    })
}
```

### 3. 复杂函数示例

```go
// 多参数函数
mvc.AddFuncMap("substring", func(str string, start, length int) string {
    runes := []rune(str)
    if start >= len(runes) {
        return ""
    }
    end := start + length
    if end > len(runes) {
        end = len(runes)
    }
    return string(runes[start:end])
})

// 可变参数函数
mvc.AddFuncMap("concat", func(strs ...string) string {
    return strings.Join(strs, "")
})

// 返回多值的函数（在模板中只使用第一个返回值）
mvc.AddFuncMap("parseInt", func(s string) (int, error) {
    return strconv.Atoi(s)
})
```

## 📝 模板使用技巧

### 1. 管道操作

```html
<!-- 链式调用 -->
<h1>{{.Title | upper | trim}}</h1>

<!-- 复合操作 -->
<p>{{.Description | lower | substring 0 100}}</p>
```

### 2. 条件判断

```html
<!-- 字符串检查 -->
{{if isEmpty .Comment}}
    <p>暂无评论</p>
{{else}}
    <p>{{.Comment}}</p>
{{end}}

<!-- 数组检查 -->
{{if contains .Roles "admin"}}
    <div class="admin-panel">管理员面板</div>
{{end}}
```

### 3. 循环处理

```html
<!-- 数组处理 -->
<ul>
{{range .Items}}
    <li>{{upper .}}</li>
{{end}}
</ul>

<!-- 带索引的循环 -->
{{range $index, $item := .Items}}
    <p>{{$index}}: {{upper $item}}</p>
{{end}}
```

## ⚠️ 注意事项

### 1. 函数命名

- 使用清晰、描述性的函数名
- 避免与内置函数冲突
- 遵循驼峰命名规范

```go
// ✅ 推荐
mvc.AddFuncMap("formatPrice", formatPriceFunc)
mvc.AddFuncMap("isValidEmail", isValidEmailFunc)

// ❌ 不推荐
mvc.AddFuncMap("f1", someFunc)
mvc.AddFuncMap("helper", genericFunc)
```

### 2. 函数类型

- 函数参数和返回值应该是模板友好的类型
- 避免复杂的结构体作为参数
- 错误处理在函数内部完成

```go
// ✅ 推荐
mvc.AddFuncMap("formatDate", func(t time.Time) string {
    if t.IsZero() {
        return ""
    }
    return t.Format("2006-01-02")
})

// ❌ 不推荐
mvc.AddFuncMap("complexFunc", func(data ComplexStruct) (ComplexResult, error) {
    // 模板难以处理复杂类型和错误
    return result, err
})
```

### 3. 性能考虑

- 函数应该是轻量级的
- 避免在函数中执行耗时操作
- 考虑缓存计算结果

```go
// ✅ 推荐：简单快速的函数
mvc.AddFuncMap("upper", strings.ToUpper)

// ❌ 避免：耗时的操作
mvc.AddFuncMap("heavyComputation", func(data string) string {
    // 避免数据库查询、网络请求等耗时操作
    time.Sleep(time.Second) // 这会影响页面渲染速度
    return processedData
})
```

## 🧪 测试

```go
func TestAddFuncMap(t *testing.T) {
    // 添加测试函数
    mvc.AddFuncMap("testFunc", func(s string) string {
        return "test_" + s
    })
    
    // 验证函数是否注册成功
    funcs := mvc.GetGlobalFuncMap()
    if _, exists := funcs["testFunc"]; !exists {
        t.Error("testFunc was not registered")
    }
    
    // 清理测试函数
    mvc.RemoveFuncMap("testFunc")
}
```

## 🔧 故障排除

### 常见问题

1. **函数未生效**: 确保在注册控制器之前调用 `AddFuncMap`
2. **函数名冲突**: 使用 `ListFuncMap()` 检查已注册的函数
3. **模板错误**: 检查函数参数类型和返回值类型
4. **性能问题**: 避免在模板函数中执行耗时操作

### 调试技巧

```go
// 列出所有已注册的函数
funcList := mvc.ListFuncMap()
fmt.Printf("已注册的模板函数: %v\n", funcList)

// 检查特定函数是否存在
funcs := mvc.GetGlobalFuncMap()
if _, exists := funcs["myFunc"]; exists {
    fmt.Println("myFunc 已注册")
} else {
    fmt.Println("myFunc 未注册")
}
```

## 📚 更多资源

- [YYHertz MVC 框架文档](../../README.md)
- [Go 模板语法参考](https://golang.org/pkg/text/template/)
- [框架工具函数参考](../../../util/README.md)

---

通过 `AddFuncMap` 功能，您可以轻松扩展模板引擎的功能，让模板更加强大和灵活！