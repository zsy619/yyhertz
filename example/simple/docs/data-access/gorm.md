# 🗄️ GORM集成

GORM是Go语言最受欢迎的ORM库，YYHertz MVC框架深度集成了GORM，提供了强大的数据库操作能力。

## 🌟 核心特性

### ✨ 框架集成优势
- **🔌 无缝集成** - 与MVC控制器完美结合
- **🏗️ 模型绑定** - 自动模型映射和验证
- **🔄 事务管理** - 声明式事务支持
- **📊 连接池** - 智能数据库连接池管理
- **🔍 查询构建** - 链式查询构建器

### 🎯 GORM功能支持
- **完整的CRUD操作**
- **关联关系映射**
- **数据库迁移**
- **钩子函数**
- **软删除**
- **批量操作**

## 🚀 快速开始

### 1. 安装依赖

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql
go get -u gorm.io/driver/postgres
go get -u gorm.io/driver/sqlite
```

### 2. 数据库配置

```yaml
# config/database.yaml
database:
  driver: "mysql"
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: "1h"
  log_level: "info"
  
  # 多数据库配置
  databases:
    primary:
      driver: "mysql"
      dsn: "user:password@tcp(localhost:3306)/primary?charset=utf8mb4"
    readonly:
      driver: "mysql" 
      dsn: "user:password@tcp(localhost:3307)/primary?charset=utf8mb4"
```

### 3. 初始化数据库连接

```go
// database/connection.go
package database

import (
    "log"
    "time"
    
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "github.com/zsy619/yyhertz/framework/config"
)

var (
    DB         *gorm.DB
    ReadOnlyDB *gorm.DB
)

type Config struct {
    Driver          string        `yaml:"driver"`
    DSN             string        `yaml:"dsn"`
    MaxIdleConns    int           `yaml:"max_idle_conns"`
    MaxOpenConns    int           `yaml:"max_open_conns"`
    ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
    LogLevel        string        `yaml:"log_level"`
}

func InitDB() error {
    config := config.GetDatabaseConfig()
    
    // 初始化主数据库
    var err error
    DB, err = connectDB(config.Primary)
    if err != nil {
        return fmt.Errorf("failed to connect primary database: %w", err)
    }
    
    // 初始化只读数据库（可选）
    if config.ReadOnly.DSN != "" {
        ReadOnlyDB, err = connectDB(config.ReadOnly)
        if err != nil {
            log.Printf("Warning: failed to connect readonly database: %v", err)
            ReadOnlyDB = DB // 降级到主数据库
        }
    } else {
        ReadOnlyDB = DB
    }
    
    log.Println("Database connections initialized successfully")
    return nil
}

func connectDB(config Config) (*gorm.DB, error) {
    // 配置GORM日志级别
    var logLevel logger.LogLevel
    switch config.LogLevel {
    case "silent":
        logLevel = logger.Silent
    case "error":
        logLevel = logger.Error
    case "warn":
        logLevel = logger.Warn
    case "info":
        logLevel = logger.Info
    default:
        logLevel = logger.Warn
    }
    
    // GORM配置
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "t_",    // 表名前缀
            SingularTable: true,    // 使用单数表名
        },
    }
    
    // 根据驱动类型连接数据库
    var dialector gorm.Dialector
    switch config.Driver {
    case "mysql":
        dialector = mysql.Open(config.DSN)
    case "postgres":
        dialector = postgres.Open(config.DSN)
    case "sqlite":
        dialector = sqlite.Open(config.DSN)
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
    }
    
    db, err := gorm.Open(dialector, gormConfig)
    if err != nil {
        return nil, err
    }
    
    // 配置连接池
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    sqlDB.SetMaxIdleConns(config.MaxIdleConns)
    sqlDB.SetMaxOpenConns(config.MaxOpenConns)
    sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
    
    return db, nil
}

// GetDB 获取主数据库连接
func GetDB() *gorm.DB {
    return DB
}

// GetReadOnlyDB 获取只读数据库连接
func GetReadOnlyDB() *gorm.DB {
    return ReadOnlyDB
}
```

### 4. 模型定义

```go
// models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User 用户模型
type User struct {
    BaseModel
    Username    string     `gorm:"uniqueIndex;size:50;not null" json:"username" validate:"required,min=3,max=50"`
    Email       string     `gorm:"uniqueIndex;size:100;not null" json:"email" validate:"required,email"`
    Password    string     `gorm:"size:255;not null" json:"-"`
    Nickname    string     `gorm:"size:50" json:"nickname"`
    Avatar      string     `gorm:"size:255" json:"avatar"`
    Status      UserStatus `gorm:"default:1" json:"status"`
    LastLoginAt *time.Time `json:"last_login_at"`
    
    // 关联关系
    Profile *UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Posts   []Post       `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
    Roles   []Role       `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

type UserStatus int

const (
    UserStatusInactive UserStatus = 0
    UserStatusActive   UserStatus = 1
    UserStatusBlocked  UserStatus = 2
)

// TableName 自定义表名
func (User) TableName() string {
    return "users"
}

// BeforeCreate GORM钩子 - 创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 密码加密
    if u.Password != "" {
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
        if err != nil {
            return err
        }
        u.Password = string(hashedPassword)
    }
    return nil
}

// AfterFind GORM钩子 - 查询后
func (u *User) AfterFind(tx *gorm.DB) error {
    // 可以在这里添加查询后的处理逻辑
    return nil
}

// UserProfile 用户资料模型
type UserProfile struct {
    BaseModel
    UserID    uint   `gorm:"uniqueIndex;not null" json:"user_id"`
    RealName  string `gorm:"size:50" json:"real_name"`
    Phone     string `gorm:"size:20" json:"phone"`
    Birthday  *time.Time `json:"birthday"`
    Gender    int    `gorm:"default:0" json:"gender"` // 0:未知 1:男 2:女
    Bio       string `gorm:"type:text" json:"bio"`
    Location  string `gorm:"size:100" json:"location"`
    
    // 关联关系
    User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// Post 文章模型
type Post struct {
    BaseModel
    Title     string    `gorm:"size:200;not null" json:"title" validate:"required,max=200"`
    Content   string    `gorm:"type:longtext" json:"content"`
    Summary   string    `gorm:"size:500" json:"summary"`
    AuthorID  uint      `gorm:"not null;index" json:"author_id"`
    Status    PostStatus `gorm:"default:1" json:"status"`
    ViewCount int       `gorm:"default:0" json:"view_count"`
    
    // 关联关系
    Author *User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
    Tags   []Tag `gorm:"many2many:post_tags" json:"tags,omitempty"`
}

type PostStatus int

const (
    PostStatusDraft     PostStatus = 0
    PostStatusPublished PostStatus = 1
    PostStatusArchived  PostStatus = 2
)

// Role 角色模型
type Role struct {
    BaseModel
    Name        string `gorm:"uniqueIndex;size:50;not null" json:"name"`
    Description string `gorm:"size:200" json:"description"`
    
    // 关联关系
    Users       []User       `gorm:"many2many:user_roles" json:"users,omitempty"`
    Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
}

// Permission 权限模型
type Permission struct {
    BaseModel
    Name        string `gorm:"uniqueIndex;size:100;not null" json:"name"`
    Action      string `gorm:"size:50;not null" json:"action"`
    Resource    string `gorm:"size:50;not null" json:"resource"`
    Description string `gorm:"size:200" json:"description"`
    
    // 关联关系
    Roles []Role `gorm:"many2many:role_permissions" json:"roles,omitempty"`
}
```

## 🏗️ 控制器集成

### 1. 基础控制器扩展

```go
// controllers/base_controller.go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/database"
    "gorm.io/gorm"
)

type BaseController struct {
    mvc.BaseController
}

// GetDB 获取数据库连接
func (c *BaseController) GetDB() *gorm.DB {
    return database.GetDB()
}

// GetReadOnlyDB 获取只读数据库连接
func (c *BaseController) GetReadOnlyDB() *gorm.DB {
    return database.GetReadOnlyDB()
}

// WithTransaction 在事务中执行操作
func (c *BaseController) WithTransaction(fn func(*gorm.DB) error) error {
    return c.GetDB().Transaction(fn)
}

// Paginate 分页查询
func (c *BaseController) Paginate(page, pageSize int) func(*gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        if page == 0 {
            page = 1
        }
        
        switch {
        case pageSize > 100:
            pageSize = 100
        case pageSize <= 0:
            pageSize = 10
        }
        
        offset := (page - 1) * pageSize
        return db.Offset(offset).Limit(pageSize)
    }
}
```

### 2. 用户控制器实现

```go
// controllers/user_controller.go
package controllers

import (
    "net/http"
    "strconv"
    
    "github.com/zsy619/yyhertz/models"
    "gorm.io/gorm"
)

type UserController struct {
    BaseController
}

// GetIndex 获取用户列表
func (c *UserController) GetIndex() {
    // 获取查询参数
    page, _ := strconv.Atoi(c.GetQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.GetQuery("page_size", "10"))
    keyword := c.GetQuery("keyword", "")
    status := c.GetQuery("status", "")
    
    // 构建查询
    query := c.GetReadOnlyDB().Model(&models.User{})
    
    // 关键词搜索
    if keyword != "" {
        query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?", 
            "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
    }
    
    // 状态筛选
    if status != "" {
        query = query.Where("status = ?", status)
    }
    
    // 获取总数
    var total int64
    query.Count(&total)
    
    // 分页查询
    var users []models.User
    err := query.Preload("Profile").
        Scopes(c.Paginate(page, pageSize)).
        Find(&users).Error
        
    if err != nil {
        c.Error(500, "查询用户失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "success",
        "data": map[string]interface{}{
            "users": users,
            "pagination": map[string]interface{}{
                "page":      page,
                "page_size": pageSize,
                "total":     total,
                "pages":     (total + int64(pageSize) - 1) / int64(pageSize),
            },
        },
    })
}

// GetShow 获取用户详情
func (c *UserController) GetShow() {
    id, err := strconv.ParseUint(c.GetParam("id"), 10, 32)
    if err != nil {
        c.Error(400, "无效的用户ID")
        return
    }
    
    var user models.User
    err = c.GetReadOnlyDB().
        Preload("Profile").
        Preload("Roles").
        First(&user, id).Error
        
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            c.Error(404, "用户不存在")
        } else {
            c.Error(500, "查询用户失败: "+err.Error())
        }
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "success",
        "data": user,
    })
}

// PostCreate 创建用户
func (c *UserController) PostCreate() {
    var user models.User
    if err := c.BindJSON(&user); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 验证数据
    if err := c.Validate(&user); err != nil {
        c.Error(400, "数据验证失败: "+err.Error())
        return
    }
    
    // 检查用户名和邮箱是否已存在
    var count int64
    c.GetDB().Model(&models.User{}).
        Where("username = ? OR email = ?", user.Username, user.Email).
        Count(&count)
        
    if count > 0 {
        c.Error(400, "用户名或邮箱已存在")
        return
    }
    
    // 创建用户（密码加密在BeforeCreate钩子中处理）
    if err := c.GetDB().Create(&user).Error; err != nil {
        c.Error(500, "创建用户失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "用户创建成功",
        "data": user,
    })
}

// PutUpdate 更新用户
func (c *UserController) PutUpdate() {
    id, err := strconv.ParseUint(c.GetParam("id"), 10, 32)
    if err != nil {
        c.Error(400, "无效的用户ID")
        return
    }
    
    // 检查用户是否存在
    var user models.User
    if err := c.GetDB().First(&user, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.Error(404, "用户不存在")
        } else {
            c.Error(500, "查询用户失败: "+err.Error())
        }
        return
    }
    
    // 绑定更新数据
    var updateData models.User
    if err := c.BindJSON(&updateData); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 验证数据
    if err := c.Validate(&updateData); err != nil {
        c.Error(400, "数据验证失败: "+err.Error())
        return
    }
    
    // 检查用户名和邮箱唯一性（排除当前用户）
    if updateData.Username != "" || updateData.Email != "" {
        query := c.GetDB().Model(&models.User{}).Where("id != ?", id)
        if updateData.Username != "" {
            query = query.Where("username = ?", updateData.Username)
        }
        if updateData.Email != "" {
            query = query.Or("email = ?", updateData.Email)
        }
        
        var count int64
        query.Count(&count)
        if count > 0 {
            c.Error(400, "用户名或邮箱已存在")
            return
        }
    }
    
    // 更新用户（排除不允许更新的字段）
    updates := map[string]interface{}{}
    if updateData.Username != "" {
        updates["username"] = updateData.Username
    }
    if updateData.Email != "" {
        updates["email"] = updateData.Email
    }
    if updateData.Nickname != "" {
        updates["nickname"] = updateData.Nickname
    }
    if updateData.Avatar != "" {
        updates["avatar"] = updateData.Avatar
    }
    if updateData.Status != 0 {
        updates["status"] = updateData.Status
    }
    
    if err := c.GetDB().Model(&user).Updates(updates).Error; err != nil {
        c.Error(500, "更新用户失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "用户更新成功",
        "data": user,
    })
}

// DeleteDestroy 删除用户（软删除）
func (c *UserController) DeleteDestroy() {
    id, err := strconv.ParseUint(c.GetParam("id"), 10, 32)
    if err != nil {
        c.Error(400, "无效的用户ID")
        return
    }
    
    // 检查用户是否存在
    var user models.User
    if err := c.GetDB().First(&user, id).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.Error(404, "用户不存在")
        } else {
            c.Error(500, "查询用户失败: "+err.Error())
        }
        return
    }
    
    // 软删除用户
    if err := c.GetDB().Delete(&user).Error; err != nil {
        c.Error(500, "删除用户失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "用户删除成功",
    })
}

// PostBatch 批量操作
func (c *UserController) PostBatch() {
    var req struct {
        Action string   `json:"action" validate:"required,oneof=delete activate deactivate"`
        IDs    []uint   `json:"ids" validate:"required,min=1"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    if err := c.Validate(&req); err != nil {
        c.Error(400, "数据验证失败: "+err.Error())
        return
    }
    
    // 在事务中执行批量操作
    err := c.WithTransaction(func(tx *gorm.DB) error {
        switch req.Action {
        case "delete":
            return tx.Delete(&models.User{}, req.IDs).Error
        case "activate":
            return tx.Model(&models.User{}).
                Where("id IN ?", req.IDs).
                Update("status", models.UserStatusActive).Error
        case "deactivate":
            return tx.Model(&models.User{}).
                Where("id IN ?", req.IDs).
                Update("status", models.UserStatusInactive).Error
        default:
            return fmt.Errorf("不支持的操作: %s", req.Action)
        }
    })
    
    if err != nil {
        c.Error(500, "批量操作失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "message": "批量操作成功",
    })
}
```

## 🔄 高级功能

### 1. 关联查询

```go
// 预加载关联数据
func (c *UserController) GetWithRelations() {
    var users []models.User
    
    // 预加载单个关联
    c.GetDB().Preload("Profile").Find(&users)
    
    // 预加载多个关联
    c.GetDB().Preload("Profile").Preload("Roles").Find(&users)
    
    // 嵌套预加载
    c.GetDB().Preload("Posts.Tags").Find(&users)
    
    // 条件预加载
    c.GetDB().Preload("Posts", "status = ?", models.PostStatusPublished).Find(&users)
    
    // 自定义预加载
    c.GetDB().Preload("Posts", func(db *gorm.DB) *gorm.DB {
        return db.Order("created_at DESC").Limit(5)
    }).Find(&users)
}

// 联合查询
func (c *UserController) GetWithJoins() {
    var users []models.User
    
    // 内连接
    c.GetDB().Joins("Profile").Find(&users)
    
    // 左连接
    c.GetDB().Joins("LEFT JOIN user_profiles ON users.id = user_profiles.user_id").
        Find(&users)
        
    // 条件联合
    c.GetDB().Joins("Profile").
        Where("user_profiles.gender = ?", 1).
        Find(&users)
}
```

### 2. 复杂查询

```go
// 子查询
func (c *UserController) GetActiveUsers() {
    var users []models.User
    
    // 子查询示例：获取有文章的用户
    c.GetDB().Where("id IN (?)", 
        c.GetDB().Model(&models.Post{}).
            Select("DISTINCT author_id").
            Where("status = ?", models.PostStatusPublished),
    ).Find(&users)
    
    // 使用子查询计算字段
    c.GetDB().Select("*, (?) as post_count", 
        c.GetDB().Model(&models.Post{}).
            Select("COUNT(*)").
            Where("author_id = users.id"),
    ).Find(&users)
}

// 聚合查询
func (c *UserController) GetStatistics() {
    var result struct {
        TotalUsers   int64 `json:"total_users"`
        ActiveUsers  int64 `json:"active_users"`
        BlockedUsers int64 `json:"blocked_users"`
    }
    
    // 总用户数
    c.GetDB().Model(&models.User{}).Count(&result.TotalUsers)
    
    // 活跃用户数
    c.GetDB().Model(&models.User{}).
        Where("status = ?", models.UserStatusActive).
        Count(&result.ActiveUsers)
        
    // 被封用户数
    c.GetDB().Model(&models.User{}).
        Where("status = ?", models.UserStatusBlocked).
        Count(&result.BlockedUsers)
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "data": result,
    })
}

// 原生SQL查询
func (c *UserController) GetCustomQuery() {
    var results []map[string]interface{}
    
    c.GetDB().Raw(`
        SELECT 
            u.username,
            u.email,
            COUNT(p.id) as post_count,
            AVG(p.view_count) as avg_views
        FROM users u
        LEFT JOIN posts p ON u.id = p.author_id
        WHERE u.status = ?
        GROUP BY u.id
        HAVING post_count > 0
        ORDER BY post_count DESC
    `, models.UserStatusActive).Scan(&results)
    
    c.JSON(map[string]interface{}{
        "code": 0,
        "data": results,
    })
}
```

### 3. 数据库迁移

```go
// database/migrate.go
package database

import (
    "log"
    "github.com/zsy619/yyhertz/models"
)

// AutoMigrate 自动迁移
func AutoMigrate() error {
    log.Println("Starting database migration...")
    
    err := DB.AutoMigrate(
        &models.User{},
        &models.UserProfile{},
        &models.Post{},
        &models.Tag{},
        &models.Role{},
        &models.Permission{},
    )
    
    if err != nil {
        return fmt.Errorf("migration failed: %w", err)
    }
    
    log.Println("Database migration completed successfully")
    return nil
}

// CreateIndexes 创建索引
func CreateIndexes() error {
    log.Println("Creating database indexes...")
    
    // 创建复合索引
    if err := DB.Exec("CREATE INDEX idx_posts_author_status ON posts(author_id, status)").Error; err != nil {
        log.Printf("Failed to create index: %v", err)
    }
    
    // 创建全文索引（MySQL）
    if err := DB.Exec("CREATE FULLTEXT INDEX idx_posts_content ON posts(title, content)").Error; err != nil {
        log.Printf("Failed to create fulltext index: %v", err)
    }
    
    log.Println("Database indexes created successfully")
    return nil
}

// SeedData 种子数据
func SeedData() error {
    log.Println("Seeding database...")
    
    // 创建默认角色
    roles := []models.Role{
        {Name: "admin", Description: "系统管理员"},
        {Name: "user", Description: "普通用户"},
    }
    
    for _, role := range roles {
        var count int64
        DB.Model(&models.Role{}).Where("name = ?", role.Name).Count(&count)
        if count == 0 {
            if err := DB.Create(&role).Error; err != nil {
                return fmt.Errorf("failed to create role %s: %w", role.Name, err)
            }
        }
    }
    
    log.Println("Database seeding completed successfully")
    return nil
}
```

## 📊 性能优化

### 1. 查询优化

```go
// 批量查询优化
func (c *UserController) GetOptimized() {
    var users []models.User
    
    // 避免N+1查询问题
    c.GetDB().Preload("Profile").Preload("Roles").Find(&users)
    
    // 选择特定字段
    c.GetDB().Select("id", "username", "email").Find(&users)
    
    // 使用索引字段查询
    c.GetDB().Where("username = ?", "john").Find(&users)
    
    // 限制查询结果
    c.GetDB().Limit(100).Find(&users)
}

// 缓存查询结果
func (c *UserController) GetCached() {
    cacheKey := "users:active"
    
    // 尝试从缓存获取
    if cached := cache.Get(cacheKey); cached != nil {
        c.JSON(cached)
        return
    }
    
    // 查询数据库
    var users []models.User
    c.GetDB().Where("status = ?", models.UserStatusActive).Find(&users)
    
    // 设置缓存
    cache.Set(cacheKey, users, 5*time.Minute)
    
    c.JSON(users)
}
```

### 2. 连接池监控

```go
// database/monitor.go
package database

import (
    "log"
    "time"
)

type DBStats struct {
    OpenConnections int    `json:"open_connections"`
    InUse          int    `json:"in_use"`
    Idle           int    `json:"idle"`
    WaitCount      int64  `json:"wait_count"`
    WaitDuration   string `json:"wait_duration"`
}

func GetDBStats() *DBStats {
    sqlDB, err := DB.DB()
    if err != nil {
        return nil
    }
    
    stats := sqlDB.Stats()
    return &DBStats{
        OpenConnections: stats.OpenConnections,
        InUse:          stats.InUse,
        Idle:           stats.Idle,
        WaitCount:      stats.WaitCount,
        WaitDuration:   stats.WaitDuration.String(),
    }
}

// 定期监控数据库连接状态
func StartDBMonitoring() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        stats := GetDBStats()
        if stats != nil {
            log.Printf("DB Stats - Open: %d, InUse: %d, Idle: %d, Wait: %d",
                stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
        }
    }
}
```

## 🧪 测试支持

### 1. 单元测试

```go
// controllers/user_controller_test.go
package controllers_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/zsy619/yyhertz/database"
    "github.com/zsy619/yyhertz/models"
)

func TestUserController_GetShow(t *testing.T) {
    // 设置测试数据库
    setupTestDB()
    defer teardownTestDB()
    
    // 创建测试用户
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    database.GetDB().Create(user)
    
    // 测试获取用户详情
    controller := &UserController{}
    // ... 测试逻辑
    
    assert.NotNil(t, user.ID)
}

func setupTestDB() {
    // 初始化测试数据库
    database.InitTestDB()
}

func teardownTestDB() {
    // 清理测试数据
    database.CleanupTestDB()
}
```

### 2. 数据库事务测试

```go
func TestUserController_WithTransaction(t *testing.T) {
    setupTestDB()
    defer teardownTestDB()
    
    controller := &UserController{}
    
    // 测试事务成功
    err := controller.WithTransaction(func(tx *gorm.DB) error {
        user := &models.User{Username: "tx_user", Email: "tx@example.com"}
        return tx.Create(user).Error
    })
    
    assert.NoError(t, err)
    
    // 验证数据已提交
    var count int64
    database.GetDB().Model(&models.User{}).Where("username = ?", "tx_user").Count(&count)
    assert.Equal(t, int64(1), count)
    
    // 测试事务回滚
    err = controller.WithTransaction(func(tx *gorm.DB) error {
        user := &models.User{Username: "rollback_user", Email: "rollback@example.com"}
        tx.Create(user)
        return errors.New("intentional error")
    })
    
    assert.Error(t, err)
    
    // 验证数据已回滚
    database.GetDB().Model(&models.User{}).Where("username = ?", "rollback_user").Count(&count)
    assert.Equal(t, int64(0), count)
}
```

## 📚 最佳实践

### 1. 模型设计
- 使用合适的字段类型和约束
- 定义清晰的关联关系
- 利用GORM钩子函数
- 实现数据验证

### 2. 查询优化
- 避免N+1查询问题
- 合理使用预加载
- 选择必要的字段
- 使用数据库索引

### 3. 事务管理
- 在必要时使用事务
- 保持事务简短
- 正确处理事务错误
- 避免长时间锁定

### 4. 连接池配置
- 根据应用需求调整连接池大小
- 监控连接池状态
- 设置合理的连接超时
- 定期检查连接健康状态

## 🔗 相关资源

- [事务管理](./transaction.md)
- [数据库配置](./database-config.md)
- [MyBatis集成](./mybatis.md)
- [性能优化指南](../dev-tools/performance.md)

---

> 💡 **提示**: GORM是功能强大的ORM框架，合理使用可以大幅提升开发效率。建议深入学习GORM的高级特性和最佳实践。
