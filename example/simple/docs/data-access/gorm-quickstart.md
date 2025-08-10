# 🚀 GORM快速开始

GORM作为YYHertz统一ORM解决方案的高性能引擎，专注处理简单高效的CRUD操作，为开发者提供优雅简洁的数据访问接口。

## ✨ 核心优势

### 🎯 性能表现
- **创建操作**: 16,278 ops/sec (67.3μs/op)
- **查询操作**: 44,816 ops/sec (26.3μs/op)
- **更新操作**: 37,648 ops/sec (31.5μs/op)
- **批量操作**: 高效的批处理支持

### 🛠️ 主要特性
- **简洁API** - 直观的链式调用语法
- **自动映射** - 结构体到数据库表的智能映射
- **关联查询** - 强大的预加载和关联功能
- **事务支持** - 完整的ACID事务管理
- **自动迁移** - 数据库表结构自动同步

## 🚀 快速开始

### 1. 基础配置

在控制器中获取GORM实例：

```go
type UserController struct {
    mvc.BaseController
}

// 获取GORM数据库连接
func (c *UserController) getDB() *gorm.DB {
    return orm.GetDefault().GORM() // 通过统一ORM获取GORM引擎
}
```

### 2. 模型定义

```go
// models/user.go
type User struct {
    ID        uint      `gorm:"primarykey" json:"id"`
    Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email     string    `gorm:"uniqueIndex;size:100;not null" json:"email"`  
    Status    int       `gorm:"default:1" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // 关联关系
    Profile *UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Posts   []Post       `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
}

type UserProfile struct {
    ID       uint   `gorm:"primarykey" json:"id"`
    UserID   uint   `gorm:"uniqueIndex;not null" json:"user_id"`
    RealName string `gorm:"size:50" json:"real_name"`
    Phone    string `gorm:"size:20" json:"phone"`
    Bio      string `gorm:"type:text" json:"bio"`
    
    User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

### 3. CRUD操作

#### 创建记录

```go
// 创建单个用户
func (c *UserController) PostCreate() {
    user := &User{
        Username: "john_doe",
        Email:    "john@example.com",
        Status:   1,
    }
    
    db := c.getDB()
    if err := db.Create(user).Error; err != nil {
        c.Error(500, "创建用户失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    user, // 包含自动生成的ID
    })
}

// 批量创建
func (c *UserController) PostBatchCreate() {
    users := []User{
        {Username: "user1", Email: "user1@example.com"},
        {Username: "user2", Email: "user2@example.com"},
        {Username: "user3", Email: "user3@example.com"},
    }
    
    db := c.getDB()
    // 批量插入，高效处理大量数据
    if err := db.CreateInBatches(users, 100).Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "count":   len(users),
    })
}
```

#### 查询记录

```go
// 查询用户列表
func (c *UserController) GetList() {
    var users []User
    db := c.getDB()
    
    // 基础查询
    if err := db.Where("status = ?", 1).Find(&users).Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "users":   users,
    })
}

// 查询用户详情（含关联数据）
func (c *UserController) GetDetail() {
    id := c.GetParam("id")
    var user User
    db := c.getDB()
    
    // 预加载关联数据
    err := db.Preload("Profile").          // 预加载用户资料
            Preload("Posts", func(db *gorm.DB) *gorm.DB {
                return db.Order("created_at DESC").Limit(5) // 最新5篇文章
            }).
            First(&user, id).Error
            
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.Error(404, "用户不存在")
        } else {
            c.Error(500, err.Error())
        }
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    user,
    })
}

// 条件查询和分页
func (c *UserController) GetSearch() {
    keyword := c.GetQuery("keyword", "")
    page, _ := strconv.Atoi(c.GetQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.GetQuery("page_size", "10"))
    
    var users []User
    var total int64
    
    db := c.getDB().Model(&User{})
    
    // 动态构建查询条件
    if keyword != "" {
        db = db.Where("username LIKE ? OR email LIKE ?", 
                     "%"+keyword+"%", "%"+keyword+"%")
    }
    
    // 获取总数
    db.Count(&total)
    
    // 分页查询
    offset := (page - 1) * pageSize
    if err := db.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "data": map[string]interface{}{
            "users":  users,
            "total":  total,
            "page":   page,
            "size":   pageSize,
        },
    })
}
```

#### 更新记录

```go
// 更新用户信息
func (c *UserController) PutUpdate() {
    id := c.GetParam("id")
    
    // 检查用户是否存在
    var user User
    db := c.getDB()
    if err := db.First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.Error(404, "用户不存在")
        } else {
            c.Error(500, err.Error())
        }
        return
    }
    
    // 绑定更新数据
    var updateData struct {
        Username string `json:"username"`
        Email    string `json:"email"`
        Status   int    `json:"status"`
    }
    
    if err := c.BindJSON(&updateData); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 选择性更新字段
    updates := make(map[string]interface{})
    if updateData.Username != "" {
        updates["username"] = updateData.Username
    }
    if updateData.Email != "" {
        updates["email"] = updateData.Email
    }
    if updateData.Status != 0 {
        updates["status"] = updateData.Status
    }
    
    if err := db.Model(&user).Updates(updates).Error; err != nil {
        c.Error(500, "更新失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    user,
    })
}

// 批量更新
func (c *UserController) PutBatchUpdate() {
    var req struct {
        IDs    []uint `json:"ids"`
        Status int    `json:"status"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, err.Error())
        return
    }
    
    db := c.getDB()
    result := db.Model(&User{}).Where("id IN ?", req.IDs).Update("status", req.Status)
    
    if result.Error != nil {
        c.Error(500, result.Error.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success":  true,
        "affected": result.RowsAffected,
    })
}
```

#### 删除记录

```go
// 软删除用户
func (c *UserController) DeleteUser() {
    id := c.GetParam("id")
    db := c.getDB()
    
    // 软删除（如果模型包含DeletedAt字段）
    if err := db.Delete(&User{}, id).Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "message": "用户删除成功",
    })
}

// 永久删除
func (c *UserController) DeletePermanent() {
    id := c.GetParam("id")
    db := c.getDB()
    
    // 硬删除
    if err := db.Unscoped().Delete(&User{}, id).Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "message": "用户永久删除成功",
    })
}
```

## 🔄 关联查询

### 一对一关联

```go
// 查询用户及其资料
func (c *UserController) GetUserWithProfile() {
    id := c.GetParam("id")
    var user User
    db := c.getDB()
    
    // 预加载一对一关联
    err := db.Preload("Profile").First(&user, id).Error
    if err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    user,
    })
}
```

### 一对多关联

```go
// 查询用户及其文章
func (c *UserController) GetUserWithPosts() {
    id := c.GetParam("id")
    var user User
    db := c.getDB()
    
    // 预加载一对多关联，按时间倒序
    err := db.Preload("Posts", func(db *gorm.DB) *gorm.DB {
        return db.Order("created_at DESC")
    }).First(&user, id).Error
    
    if err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    user,
    })
}
```

### 多对多关联

```go
// 用户角色多对多关联示例
type User struct {
    ID    uint   `gorm:"primarykey"`
    Name  string
    Roles []Role `gorm:"many2many:user_roles"`
}

type Role struct {
    ID    uint   `gorm:"primarykey"`
    Name  string
    Users []User `gorm:"many2many:user_roles"`
}

// 查询用户及其角色
func (c *UserController) GetUserWithRoles() {
    id := c.GetParam("id")
    var user User
    db := c.getDB()
    
    err := db.Preload("Roles").First(&user, id).Error
    if err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    user,
    })
}
```

## 🔁 事务操作

### 简单事务

```go
// 创建用户和资料（事务）
func (c *UserController) PostCreateWithProfile() {
    var req struct {
        User    User        `json:"user"`
        Profile UserProfile `json:"profile"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, err.Error())
        return
    }
    
    db := c.getDB()
    
    // 使用事务
    err := db.Transaction(func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(&req.User).Error; err != nil {
            return err
        }
        
        // 设置关联ID
        req.Profile.UserID = req.User.ID
        
        // 创建资料
        if err := tx.Create(&req.Profile).Error; err != nil {
            return err
        }
        
        return nil
    })
    
    if err != nil {
        c.Error(500, "创建失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    req.User,
        "profile": req.Profile,
    })
}
```

### 手动事务控制

```go
// 复杂业务事务
func (c *UserController) PostComplexTransaction() {
    db := c.getDB()
    
    // 开始事务
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    if err := tx.Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    // 操作1
    user := &User{Username: "test", Email: "test@example.com"}
    if err := tx.Create(user).Error; err != nil {
        tx.Rollback()
        c.Error(500, err.Error())
        return
    }
    
    // 操作2
    profile := &UserProfile{UserID: user.ID, RealName: "Test User"}
    if err := tx.Create(profile).Error; err != nil {
        tx.Rollback()
        c.Error(500, err.Error())
        return
    }
    
    // 提交事务
    if err := tx.Commit().Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "message": "事务执行成功",
    })
}
```

## ⚡ 性能优化

### 1. 查询优化

```go
// 选择特定字段
func (c *UserController) GetUsernames() {
    var users []User
    db := c.getDB()
    
    // 只查询需要的字段
    err := db.Select("id", "username", "email").Find(&users).Error
    if err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "users":   users,
    })
}

// 索引字段查询
func (c *UserController) GetByEmail() {
    email := c.GetQuery("email")
    var user User
    db := c.getDB()
    
    // 使用索引字段查询
    err := db.Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.Error(404, "用户不存在")
        } else {
            c.Error(500, err.Error())
        }
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "user":    user,
    })
}
```

### 2. 批量操作

```go
// 高效批量插入
func (c *UserController) PostBatchInsert() {
    var users []User
    if err := c.BindJSON(&users); err != nil {
        c.Error(400, err.Error())
        return
    }
    
    db := c.getDB()
    
    // 批量插入，每批100条
    if err := db.CreateInBatches(users, 100).Error; err != nil {
        c.Error(500, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "success": true,
        "count":   len(users),
    })
}
```

## 🛠️ 调试和监控

### SQL调试

```go
// 开发环境调试
func (c *UserController) DebugQuery() {
    db := c.getDB()
    
    // 开启SQL日志
    db = db.Debug()
    
    var users []User
    db.Where("status = ?", 1).Find(&users)
    
    c.JSON(map[string]interface{}{
        "success": true,
        "users":   users,
    })
}
```

### 连接池监控

```go
// 获取数据库连接池状态
func (c *UserController) GetDBStats() {
    db := c.getDB()
    sqlDB, err := db.DB()
    if err != nil {
        c.Error(500, err.Error())
        return
    }
    
    stats := sqlDB.Stats()
    
    c.JSON(map[string]interface{}{
        "success": true,
        "stats": map[string]interface{}{
            "open_connections": stats.OpenConnections,
            "in_use":          stats.InUse,
            "idle":            stats.Idle,
            "wait_count":      stats.WaitCount,
            "wait_duration":   stats.WaitDuration.String(),
        },
    })
}
```

## 💡 最佳实践

### 1. 模型设计
- 使用合适的字段类型和约束
- 定义清晰的关联关系
- 合理使用索引优化查询性能

### 2. 查询优化
- 避免N+1查询，合理使用Preload
- 选择必要的字段，避免查询大字段
- 使用分页避免一次查询大量数据

### 3. 事务管理
- 保持事务简短，避免长时间锁定
- 合理使用事务，避免不必要的事务开销
- 正确处理事务错误和回滚

### 4. 性能监控
- 在开发环境启用SQL调试
- 监控数据库连接池状态
- 定期检查慢查询日志

## 📚 相关文档

- **[统一ORM概览](./orm-unified)** - 了解GORM在统一ORM中的定位
- **[MyBatis基础](./mybatis-basic)** - 复杂查询的MyBatis解决方案
- **[数据库配置](./database-config)** - 连接池和多数据源配置
- **[事务管理](./transaction)** - 深入理解事务机制
- **[数据库调优](./database-tuning)** - 数据库性能优化指南

---

**🚀 GORM引擎 - 简单高效的数据访问首选！**

*在YYHertz统一ORM解决方案中，GORM专注于提供最高性能的CRUD操作体验*