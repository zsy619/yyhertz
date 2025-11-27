# 🎮 YYHertz 控制器概览

YYHertz的BaseController是一个功能强大、高度模块化的控制器基类，提供了Web开发所需的全部功能。本文档将深入解析其架构设计和核心特性。

## 🏗️ 控制器架构概览

### 核心结构设计

BaseController采用模块化设计，将功能划分为13个专门模块：

```go
type BaseController struct {
    // 🔧 核心组件
    Ctx            *context.Context     // 统一上下文
    ControllerName string               // 控制器名称
    ActionName     string               // 当前动作名称
    
    // 🎨 模板系统
    ViewPath       string               // 视图路径
    Layout         string               // 布局文件
    Data           map[string]any       // 模板数据
    TplFuncs       template.FuncMap     // 自定义模板函数
    
    // 🛣️ 路由管理
    RoutePattern   string                      // 路由模式
    RouteParams    map[string]string           // 路由参数
    URLGenerator   func(string, ...any) string // URL生成函数
    
    // 🛠️ 辅助工具
    cookieHelper   *cookie.Helper              // Cookie助手
    sessionHelper  *session.Manager            // Session管理
    templateEngine *view.TemplateEngine        // 模板引擎
    
    // ⚡ 优化特性
    optimizationEnabled bool               // 优化特性开关
    middlewareList      []string           // 中间件列表
}
```

### 功能模块分布

| 模块 | 文件 | 主要功能 | 方法数量 |
|------|------|----------|----------|
| 🔄 **生命周期** | controller.go | 初始化、资源管理 | 8个 |
| 📥 **请求处理** | controller_request.go | 参数获取、类型转换 | 15个 |
| 📤 **响应输出** | controller_response.go | JSON、HTML、文件输出 | 16个 |
| 🎨 **模板引擎** | controller_template.go | 模板渲染、布局管理 | 18个 |
| 🍪 **Session/Cookie** | controller_session.go<br/>controller_cookie.go | 会话和Cookie管理 | 10个 |
| 🛣️ **路由管理** | controller_routing.go | URL映射、路由参数 | 14个 |
| 📊 **数据管理** | controller_data.go | 模板数据存储 | 3个 |
| 🔒 **安全功能** | controller_security.go | XSRF防护、安全Cookie | 6个 |
| 🚦 **流程控制** | controller_flow.go | 请求流程控制 | 3个 |
| 🛠️ **工具方法** | controller_utils.go | 反射工具、方法映射 | 6个 |

### 设计理念

1. **模块化分离** - 每个功能模块独立，职责清晰
2. **Beego兼容** - 保持与Beego Controller的API兼容性
3. **高性能优化** - 内置缓存和性能优化特性
4. **安全第一** - 默认启用各种安全防护措施
5. **开发友好** - 丰富的辅助方法，简化开发流程

## 🚀 快速开始

### 基础控制器示例

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

type HomeController struct {
    core.BaseController
}

// GET /
func (c *HomeController) GetIndex() {
    c.SetData("Title", "欢迎使用YYHertz")
    c.SetData("Message", "高性能Go Web框架")
    c.Render("home/index.html")
}

// POST /api/data
func (c *HomeController) PostData() {
    // 获取请求参数
    name := c.GetString("name")
    age := c.GetInt("age", 0)
    
    // 参数验证
    if name == "" {
        c.JSONError("姓名不能为空")
        return
    }
    
    // 业务处理
    result := map[string]any{
        "id":   1,
        "name": name,
        "age":  age,
    }
    
    // 返回JSON响应
    c.JSONSuccess("创建成功", result)
}
```

## 🔧 核心特性

### 1. 自动路由映射

- **RESTful风格** - 支持GET、POST、PUT、DELETE等HTTP方法
- **方法映射** - 自动将HTTP方法映射到控制器方法
- **参数注入** - 自动注入路由参数到方法

### 2. 模板引擎集成

- **多模板支持** - 支持HTML模板和自定义模板引擎
- **布局管理** - 支持主布局和子模板
- **数据绑定** - 简化模板数据传递

### 3. 安全防护

- **XSRF保护** - 防止跨站请求伪造攻击
- **安全Cookie** - 加密Cookie数据传输
- **输入过滤** - 自动过滤恶意输入

### 4. 性能优化

- **方法缓存** - 缓存反射获取的方法信息
- **资源池** - 复用控制器实例减少内存分配
- **懒加载** - 按需加载辅助组件

## 📖 深入学习

- [生命周期管理](lifecycle.md) - 了解控制器的完整生命周期
- [请求处理](request-handling.md) - 掌握各种请求参数获取方法
- [响应输出](response-output.md) - 学习JSON、HTML等响应输出方式
- [模板系统](template-system.md) - 深入了解模板引擎和视图渲染
- [安全特性](security-features.md) - 掌握XSRF保护和安全最佳实践

## 🎯 最佳实践

1. **继承BaseController** - 所有控制器都应继承BaseController
2. **重写Prepare方法** - 在Prepare中进行通用的预处理逻辑
3. **使用XSRF保护** - 对于状态变更操作启用XSRF防护
4. **合理使用缓存** - 利用内置缓存机制提升性能
5. **错误处理** - 统一的错误处理和响应格式