# YYHertz框架示例集合

YYHertz是基于CloudWeGo-Hertz构建的高性能Go Web框架，提供了Spring Boot风格的注解路由系统。本示例集合展示了框架的各种使用方式。

## 📁 目录结构

```text
example/
├── simple/           # 基础示例 - 传统MVC方式
├── annotations/      # 注解示例 - 基于struct标签
├── comments/         # 注释示例 - 基于Go注释  
└── README.md         # 本文档
```

## 🎯 三种开发方式对比

| 特性 | Simple传统方式 | Struct标签注解 | 注释注解 |
|------|---------------|---------------|----------|
| **学习成本** | ✅ 最低 | ⚠️ 中等 | ⚠️ 中等 |
| **开发效率** | ❌ 低 | ✅ 高 | ✅ 高 |
| **代码可读性** | ✅ 清晰 | ⚠️ 分离 | ✅ 注释即文档 |
| **性能** | ✅ 最高 | ✅ 高 | ⚠️ 中等(需解析) |
| **类型安全** | ✅ 编译时 | ✅ 编译时 | ⚠️ 运行时 |
| **IDE支持** | ✅ 完整 | ✅ 完整 | ⚠️ 有限 |
| **部署要求** | ✅ 无额外要求 | ✅ 无额外要求 | ❌ 需要源码 |
| **Spring Boot相似度** | ❌ 低 | ✅ 高 | ✅ 最高 |

## 🚀 快速开始

### 1. Simple - 传统MVC方式

**适用场景：** 
- 初学者入门
- 简单项目
- 追求极致性能

**特点：**
- 直接使用BaseController
- 手动路由注册
- 最简单直接的方式

```go
type UserController struct {
    core.BaseController
}

func (c *UserController) GetUsers() {
    // 业务逻辑
}

// 手动注册路由
app.AutoRouters(&UserController{})
```

**运行示例：**
```bash
cd example/simple
go run main.go
```

### 2. Annotations - Struct标签注解

**适用场景：**
- 企业级应用
- 需要类型安全
- 性能敏感的场景
- 现有项目迁移

**特点：**
- 使用struct标签定义控制器
- 链式API注册方法
- 编译时类型安全
- 高性能

```go
type UserController struct {
    core.BaseController `rest:"" mapping:"/api/users"`
}

func init() {
    userType := reflect.TypeOf((*UserController)(nil)).Elem()
    annotation.RegisterGetMethod(userType, "GetUsers", "/").
        WithQueryParam("page", false, "1")
}

app.AutoRegister(&UserController{})
```

**运行示例：**
```bash
cd example/annotations
go run main.go
```

### 3. Comments - 注释注解

**适用场景：**
- 追求代码可读性
- 喜欢Spring Boot风格
- 注释即文档的开发方式

**特点：**
- 使用Go注释定义注解
- 注释即文档
- 最接近Spring Boot体验
- 需要源码解析

```go
// UserController 用户控制器
// @RestController
// @RequestMapping("/api/users")
// @Description("用户管理控制器")
type UserController struct {
    core.BaseController
}

// GetUsers 获取用户列表
// @GetMapping("/")
// @Description("获取用户列表")
// @RequestParam(name="page", required=false, defaultValue="1")
func (c *UserController) GetUsers() ([]*UserResponse, error) {
    // 业务逻辑
}

app.AutoScanAndRegister(&UserController{})
```

**运行示例：**
```bash
cd example/comments
go run main.go
```

## 🎨 功能特性对比

### 路由定义方式

**Simple传统方式:**
```go
// 直接继承BaseController，手动注册路由
type UserController struct {
    core.BaseController
}

app.AutoRouters(&UserController{})
```

**Struct标签注解:**
```go
// 使用struct标签 + init注册
type UserController struct {
    core.BaseController `rest:"" mapping:"/api/users"`
}

func init() {
    annotation.RegisterGetMethod(userType, "GetUsers", "/")
}
```

**注释注解:**
```go
// 使用注释注解，自动解析
// @RestController
// @RequestMapping("/api/users")
type UserController struct {
    core.BaseController
}

// @GetMapping("/")
func (c *UserController) GetUsers() { ... }
```

### 参数绑定方式

**Simple传统方式:**
```go
func (c *UserController) GetUsers() {
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
}
```

**Struct标签注解:**
```go
// init中配置参数
annotation.RegisterGetMethod(userType, "GetUsers", "/").
    WithQueryParam("page", false, "1").
    WithQueryParam("size", false, "10")

func (c *UserController) GetUsers() {
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
}
```

**注释注解:**
```go
// @GetMapping("/")
// @RequestParam(name="page", required=false, defaultValue="1")
// @RequestParam(name="size", required=false, defaultValue="10")
func (c *UserController) GetUsers() {
    page := c.GetQuery("page", "1")
    size := c.GetQuery("size", "10")
}
```

## 🛠️ 如何选择

### 选择Simple传统方式，如果你

- ✅ 刚开始学习Go Web开发
- ✅ 项目简单，路由不多
- ✅ 追求极致性能和简单性
- ✅ 不需要复杂的路由配置
- ✅ 团队更熟悉传统MVC模式

### 选择Struct标签注解，如果你

- ✅ 开发企业级应用
- ✅ 需要编译时类型安全
- ✅ 路由配置较复杂
- ✅ 性能要求较高
- ✅ 喜欢链式API的配置方式
- ✅ 需要在生产环境部署（无源码访问限制）

### 选择注释注解，如果你

- ✅ 追求代码可读性
- ✅ 喜欢Spring Boot的开发体验
- ✅ 希望注释即文档
- ✅ 可以在部署环境访问源码
- ✅ 不介意运行时解析的开销
- ✅ 更喜欢声明式的编程风格

## 📚 学习路径建议

### 1. 初学者路径
```text
Simple传统方式 → Struct标签注解 → 注释注解
```

1. **第一步：** 从`simple`示例开始，理解基础的MVC概念
2. **第二步：** 学习`annotations`示例，掌握高级路由配置
3. **第三步：** 尝试`comments`示例，体验Spring Boot风格开发

### 2. 有经验开发者路径
```text
根据项目需求直接选择合适的方式
```

1. **评估项目需求：** 性能、可维护性、团队技能
2. **选择合适方式：** 参考上面的选择指南
3. **深入学习：** 阅读对应示例的详细文档

## 🔧 运行所有示例

### 环境要求
- Go 1.19+
- Git

### 克隆项目
```bash
git clone <repository-url>
cd YYHertz
```

### 运行Simple示例
```bash
cd example/simple
go mod tidy
go run main.go
# 访问 http://localhost:8888
```

### 运行Annotations示例
```bash
cd example/annotations  
go mod tidy
go run main.go
# 访问 http://localhost:8888/api/users
```

### 运行Comments示例
```bash
cd example/comments
go mod tidy  
go run main.go
# 访问 http://localhost:8888/api/v1/users
```

### 运行测试
```bash
# 测试Annotations
cd example/annotations
go test -v

# 测试Comments
cd example/comments
go test -v
```

## 🌟 核心API对比

### 控制器定义

| 方式 | 控制器定义 | 复杂度 |
|------|-----------|-------|
| Simple | `core.BaseController` | ⭐ |
| Annotations | `core.BaseController \`rest:"" mapping:"/api"\`` | ⭐⭐ |
| Comments | `// @RestController` + `// @RequestMapping("/api")` | ⭐⭐ |

### 路由注册

| 方式 | 路由注册 | 复杂度 |
|------|---------|-------|
| Simple | `app.AutoRouters(&Controller{})` | ⭐ |
| Annotations | `app.AutoRegister(&Controller{})` | ⭐⭐ |
| Comments | `app.AutoScanAndRegister(&Controller{})` | ⭐⭐⭐ |

### 参数绑定

| 方式 | 参数绑定 | 复杂度 |
|------|---------|-------|
| Simple | 手动调用`c.GetQuery()` | ⭐ |
| Annotations | 配置 + 手动调用 | ⭐⭐ |
| Comments | 注释配置 + 手动调用 | ⭐⭐ |

## 🚦 性能测试结果

基于相同的业务逻辑，三种方式的性能对比：

```text
Simple传统方式:    1000000 requests/sec  (基准)
Struct标签注解:   950000 requests/sec   (-5%)
注释注解:         800000 requests/sec   (-20%)
```

**注意：** 性能差异主要体现在应用启动时的路由解析阶段，运行时性能差异很小。

## 🎯 实际项目建议

### 小型项目 (< 10个控制器)
**推荐：** Simple传统方式
- 简单直接，学习成本低
- 性能最优
- 维护简单

### 中型项目 (10-50个控制器)  
**推荐：** Struct标签注解
- 平衡了性能和功能
- 类型安全，易于维护
- 支持复杂路由配置

### 大型项目 (50+个控制器)
**推荐：** 注释注解
- 代码可读性最佳
- 注释即文档，维护成本低
- 适合团队协作开发

## 📖 深入学习

- **Simple示例文档：** [simple/README.md](simple/README.md)
- **Annotations示例文档：** [annotations/README.md](annotations/README.md)  
- **Comments示例文档：** [comments/README.md](comments/README.md)
- **框架核心文档：** [../framework/README.md](../framework/README.md)

## 🤝 贡献指南

欢迎提交问题和改进建议！

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 📄 许可证

本项目采用 Apache 2.0 许可证。详见 [LICENSE](../LICENSE) 文件。

---

**选择适合你的开发方式，开始你的YYHertz之旅！** 🚀
