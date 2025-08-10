# MyBatis基础集成

YYHertz框架内置MyBatis集成支持，提供Go语言化的SQL映射和数据库操作功能。

## 🚀 快速开始

### 1. 框架自动配置

YYHertz框架会自动根据 `conf/database.yaml` 配置初始化MyBatis：

```yaml
# conf/database.yaml
primary:
  driver: "mysql"
  host: "localhost"
  port: 3306
  database: "yyhertz"
  username: "root"
  password: ""
  charset: "utf8mb4"

# MyBatis配置（可选，有默认值）
mybatis:
  enable: true                    # 启用MyBatis集成
  mapper_locations: "./mappers/*.xml"
  cache_enabled: true
  map_underscore_map: true        # 下划线到驼峰映射
```

### 2. 在控制器中使用

```go
package controllers

import (
    "context"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mybatis"
)

type UserController struct {
    mvc.BaseController
    session mybatis.SimpleSession  // 注入SimpleSession
}

// 构造函数，框架会自动注入
func NewUserController(session mybatis.SimpleSession) *UserController {
    return &UserController{session: session}
}

func (c *UserController) GetIndex() {
    ctx := context.Background()
    
    // 查询用户列表
    users, err := c.session.SelectList(ctx, 
        "SELECT * FROM users WHERE status = ? ORDER BY id DESC LIMIT 10", 
        "active")
    if err != nil {
        c.Error(500, "查询失败: "+err.Error())
        return
    }
    
    c.JSON(mvc.Result{
        Success: true,
        Data:    users,
    })
}
```

## 📊 核心功能

### SimpleSession 接口

YYHertz的MyBatis提供了简洁的SimpleSession接口，专为Go语言优化：

```go
type SimpleSession interface {
    // 查询操作
    SelectOne(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
    SelectList(ctx context.Context, sql string, args ...interface{}) ([]interface{}, error)
    SelectPage(ctx context.Context, sql string, page PageRequest, args ...interface{}) (*PageResult, error)
    
    // 数据操作  
    Insert(ctx context.Context, sql string, args ...interface{}) (int64, error)
    Update(ctx context.Context, sql string, args ...interface{}) (int64, error)
    Delete(ctx context.Context, sql string, args ...interface{}) (int64, error)
}
```

### 1. 基础CRUD操作

#### 查询单条记录

```go
func (c *UserController) GetShow() {
    id := c.GetParamInt64("id")
    ctx := context.Background()
    
    user, err := c.session.SelectOne(ctx, 
        "SELECT * FROM users WHERE id = ? AND deleted_at IS NULL", id)
    if err != nil {
        c.Error(500, "查询用户失败")
        return
    }
    
    if user == nil {
        c.Error(404, "用户不存在")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: user})
}
```

#### 查询列表

```go
func (c *UserController) GetList() {
    ctx := context.Background()
    
    users, err := c.session.SelectList(ctx,
        "SELECT id, name, email, status, created_at FROM users WHERE deleted_at IS NULL ORDER BY id DESC")
    if err != nil {
        c.Error(500, "查询失败")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: users})
}
```

#### 插入数据

```go
func (c *UserController) PostCreate() {
    ctx := context.Background()
    
    // 获取表单数据
    name := c.GetForm("name")
    email := c.GetForm("email")
    
    // 插入用户
    id, err := c.session.Insert(ctx,
        "INSERT INTO users (name, email, status, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())",
        name, email, "active")
    if err != nil {
        c.Error(500, "创建用户失败")
        return
    }
    
    c.JSON(mvc.Result{
        Success: true,
        Data:    map[string]interface{}{"id": id},
        Message: "用户创建成功",
    })
}
```

#### 更新数据

```go
func (c *UserController) PutUpdate() {
    id := c.GetParamInt64("id")
    ctx := context.Background()
    
    name := c.GetForm("name")
    email := c.GetForm("email")
    
    affected, err := c.session.Update(ctx,
        "UPDATE users SET name = ?, email = ?, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL",
        name, email, id)
    if err != nil {
        c.Error(500, "更新用户失败")
        return
    }
    
    if affected == 0 {
        c.Error(404, "用户不存在")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Message: "更新成功"})
}
```

#### 软删除

```go
func (c *UserController) DeleteUser() {
    id := c.GetParamInt64("id")
    ctx := context.Background()
    
    affected, err := c.session.Update(ctx,
        "UPDATE users SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
    if err != nil {
        c.Error(500, "删除用户失败")
        return
    }
    
    if affected == 0 {
        c.Error(404, "用户不存在")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Message: "删除成功"})
}
```

### 2. 智能分页查询

YYHertz的MyBatis提供了智能分页功能，自动处理COUNT查询和LIMIT/OFFSET：

```go
func (c *UserController) GetPage() {
    ctx := context.Background()
    
    // 获取分页参数
    page := c.GetQueryInt("page", 1)
    size := c.GetQueryInt("size", 10)
    status := c.GetQuery("status", "active")
    
    // 自动分页查询
    pageResult, err := c.session.SelectPage(ctx,
        "SELECT * FROM users WHERE status = ? ORDER BY id DESC",
        mybatis.PageRequest{Page: page, Size: size},
        status)
    if err != nil {
        c.Error(500, "查询失败")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: pageResult})
}
```

**分页结果结构：**
```go
type PageResult struct {
    Data       []interface{} `json:"data"`        // 数据列表
    Total      int64         `json:"total"`       // 总记录数
    Page       int           `json:"page"`        // 当前页码
    Size       int           `json:"size"`        // 每页大小
    TotalPages int           `json:"total_pages"` // 总页数
    HasNext    bool          `json:"has_next"`    // 是否有下一页
    HasPrev    bool          `json:"has_prev"`    // 是否有上一页
}
```

## 🔧 实体映射

### 定义Go结构体

```go
package models

import "time"

type User struct {
    ID        int64     `json:"id" db:"id"`
    Name      string    `json:"name" db:"name" validate:"required,min=2,max=50"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    Status    string    `json:"status" db:"status"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// 表名映射
func (User) TableName() string {
    return "users"
}
```

### 结果映射处理

```go
// 将map结果转换为结构体
func mapToUser(result interface{}) (*User, error) {
    if result == nil {
        return nil, nil
    }
    
    resultMap, ok := result.(map[string]interface{})
    if !ok {
        return nil, fmt.Errorf("invalid result type")
    }
    
    user := &User{}
    if id, ok := resultMap["id"]; ok {
        user.ID = id.(int64)
    }
    if name, ok := resultMap["name"]; ok {
        user.Name = name.(string)
    }
    // ... 其他字段映射
    
    return user, nil
}
```

## 🛡️ 安全最佳实践

### 1. 参数化查询

**✅ 推荐做法：**
```go
// 使用参数化查询防止SQL注入
users, err := session.SelectList(ctx, 
    "SELECT * FROM users WHERE name LIKE ? AND status = ?", 
    "%"+keyword+"%", "active")
```

**❌ 错误做法：**
```go
// 直接拼接SQL容易造成注入
sql := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", userInput)
users, err := session.SelectList(ctx, sql)  // 危险！
```

### 2. 输入验证

```go
func (c *UserController) PostCreate() {
    // 验证输入
    name := strings.TrimSpace(c.GetForm("name"))
    if len(name) < 2 || len(name) > 50 {
        c.Error(400, "用户名长度必须在2-50个字符之间")
        return
    }
    
    email := strings.TrimSpace(c.GetForm("email"))
    if !isValidEmail(email) {
        c.Error(400, "邮箱格式不正确")
        return
    }
    
    // ... 执行插入
}
```

## ⚡ 性能优化

### 1. 连接池配置

```yaml
# conf/database.yaml
primary:
  max_open_conns: 50      # 最大打开连接数
  max_idle_conns: 25      # 最大空闲连接数  
  conn_max_lifetime: "1h" # 连接最大生存时间
  conn_max_idle_time: "30m" # 连接最大空闲时间
```

### 2. 查询优化

```go
// 使用LIMIT避免大结果集
users, err := session.SelectList(ctx,
    "SELECT id, name, email FROM users ORDER BY id DESC LIMIT 100")

// 使用索引字段查询
user, err := session.SelectOne(ctx,
    "SELECT * FROM users WHERE email = ?", email) // email有索引
```

### 3. 批量操作

```go
// 批量插入（在高级特性中详细介绍）
func (c *UserController) PostBatchCreate() {
    ctx := context.Background()
    users := c.GetJSONArray("users") // 获取用户数组
    
    // 构建批量插入SQL
    values := make([]string, len(users))
    args := make([]interface{}, 0, len(users)*3)
    
    for i, user := range users {
        values[i] = "(?, ?, ?)"
        args = append(args, user.Name, user.Email, "active")
    }
    
    sql := fmt.Sprintf("INSERT INTO users (name, email, status) VALUES %s", 
        strings.Join(values, ", "))
    
    affected, err := c.session.Insert(ctx, sql, args...)
    // ... 处理结果
}
```

## 🔍 错误处理

### 标准错误处理模式

```go
func (c *UserController) GetUser() {
    id := c.GetParamInt64("id")
    ctx := context.Background()
    
    user, err := c.session.SelectOne(ctx, 
        "SELECT * FROM users WHERE id = ?", id)
    if err != nil {
        // 记录详细错误日志
        c.Logger.Error("查询用户失败", "id", id, "error", err)
        
        // 返回用户友好的错误信息
        c.Error(500, "系统繁忙，请稍后重试")
        return
    }
    
    if user == nil {
        c.Error(404, "用户不存在")
        return
    }
    
    c.JSON(mvc.Result{Success: true, Data: user})
}
```

## 📚 下一步

学习了MyBatis基础用法后，您可以继续学习：

- **[MyBatis高级特性](./mybatis-advanced)** - XML映射器、动态SQL、钩子系统
- **[MyBatis性能优化](./mybatis-performance)** - 性能测试、监控、生产环境最佳实践
- **[事务管理](./transaction)** - 数据库事务的完整处理方案

## 🔗 相关资源

- [完整示例项目](../../gobatis/) - 包含性能测试和压力测试的完整示例
- [数据库配置](./database-config) - database.yaml的完整配置说明
- [GORM快速开始](./gorm-quickstart) - 可与MyBatis配合使用的高性能CRUD引擎