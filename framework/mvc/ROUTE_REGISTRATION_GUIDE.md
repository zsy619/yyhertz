# 路由注册方法优化指南

## 📋 新的路由注册方法

### 🔄 自动路由注册（推荐）

```go
// 1. 注册多个控制器（自动推导路由）
app.RegisterControllers(homeController, userController, adminController)
// HomeController.GetIndex -> GET /
// UserController.GetInfo  -> GET /user/info
// AdminController.PostSettings -> POST /admin/settings

// 2. 注册多个控制器（带路径前缀）
app.RegisterControllersWithPrefix("/api/v1", userController, orderController)
// UserController.GetInfo  -> GET /api/v1/user/info
// OrderController.PostCreate -> POST /api/v1/order/create

// 3. 注册单个控制器（自动推导）
app.RegisterController(userController)
// UserController.GetIndex -> GET /user/index
```

### 🎯 手动路由映射

```go
// 1. 手动映射控制器路由
app.MapRoutes(userController,
    "GetInfo", "GET:/user/profile",      // 自定义路径
    "PostUpdate", "POST:/user/update",   // 明确指定HTTP方法
    "DeleteUser", "DELETE:/user/:id",    // 带参数的路径
)

// 2. 手动映射控制器路由（带前缀）
app.MapRoutesWithPrefix("/api", userController,
    "GetInfo", "GET:/info",              // 实际路径: /api/info  
    "PostCreate", "POST:/create",        // 实际路径: /api/create
)
```

### 🔀 混合注册（智能模式）

```go
// 无routes参数 -> 自动注册
app.RegisterController(homeController)

// 有routes参数 -> 手动映射
app.RegisterController(userController,
    "GetInfo", "GET:/user/profile",
    "PostUpdate", "POST:/user/settings",
)

// 带前缀的混合注册
app.RegisterControllerWithPrefix("/admin", adminController,
    "GetDashboard", "GET:/",             // /admin/
    "GetUsers", "GET:/users",            // /admin/users
    "PostSettings", "POST:/config",      // /admin/config
)
```

## 🔄 迁移指南

### 旧方法 → 新方法映射

```go
// 旧方法（已废弃）
app.Include("", homeController, userController)
app.Router("/api", userController, "GetInfo", "GET:/info")

// 新方法（推荐）
app.RegisterControllers(homeController, userController)
app.MapRoutesWithPrefix("/api", userController, "GetInfo", "GET:/info")
```

## 📚 方法名称对比

| 使用场景 | 旧方法名 | 新方法名 | 优势 |
|---------|---------|---------|------|
| 自动注册多个控制器 | `Include` | `RegisterControllers` | 语义更清晰 |
| 自动注册（带前缀） | `Include` | `RegisterControllersWithPrefix` | 明确表达前缀功能 |
| 手动映射路由 | `Router` | `MapRoutes` | 避免与Router类型混淆 |
| 手动映射（带前缀） | `Router` | `MapRoutesWithPrefix` | 明确表达前缀功能 |
| 智能注册 | `IncludeRoutes` | `RegisterController` | 职责更单一 |

## ✅ 最佳实践

1. **优先使用自动注册**：
   ```go
   app.RegisterControllers(homeController, userController, adminController)
   ```

2. **API路由使用前缀**：
   ```go
   app.RegisterControllersWithPrefix("/api/v1", apiControllers...)
   ```

3. **特殊路由使用手动映射**：
   ```go
   app.MapRoutes(specialController, 
       "CustomAction", "GET:/special/path/:id",
   )
   ```

4. **管理后台使用前缀和混合注册**：
   ```go
   app.RegisterControllerWithPrefix("/admin", adminController,
       "GetDashboard", "GET:/",
       "GetStats", "GET:/statistics", 
   )
   ```

## 🔧 向后兼容性

旧的 `Include` 和 `Router` 方法仍然可用，但已标记为废弃：

```go
// 这些方法仍然可用，但建议迁移到新方法
app.Include("", controllers...)        // 已废弃
app.Router("/api", controller, routes...) // 已废弃
```

编译器会显示废弃警告，建议逐步迁移到新的方法名。