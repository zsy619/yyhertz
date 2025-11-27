# 🗄️ 数据库集成

YYHertz提供了灵活的数据库集成方案，支持多种数据库和ORM框架。

## 支持的数据库

### 关系型数据库

- **MySQL** - 最流行的开源关系数据库
- **PostgreSQL** - 功能强大的开源数据库
- **SQLite** - 轻量级文件数据库
- **SQL Server** - Microsoft企业级数据库
- **Oracle** - 企业级数据库解决方案

### NoSQL数据库

- **MongoDB** - 文档型数据库
- **Redis** - 内存数据库
- **Elasticsearch** - 搜索引擎

## GORM集成

### 基础配置

```go
package config

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type DatabaseConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    Database string
    Charset  string
}

func NewDatabase(config *DatabaseConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
        config.Username,
        config.Password,
        config.Host,
        config.Port,
        config.Database,
        config.Charset,
    )
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### 模型定义

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// 基础模型
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 用户模型
type User struct {
    BaseModel
    Username string    `gorm:"uniqueIndex;size:50" json:"username"`
    Email    string    `gorm:"uniqueIndex;size:100" json:"email"`
    Password string    `gorm:"size:255" json:"-"`
    Avatar   string    `gorm:"size:255" json:"avatar"`
    Status   int       `gorm:"default:1" json:"status"`
    Profile  *Profile  `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Posts    []Post    `gorm:"foreignKey:UserID" json:"posts,omitempty"`
}

// 用户资料模型
type Profile struct {
    BaseModel
    UserID   uint   `gorm:"uniqueIndex" json:"user_id"`
    Nickname string `gorm:"size:50" json:"nickname"`
    Bio      string `gorm:"type:text" json:"bio"`
    Phone    string `gorm:"size:20" json:"phone"`
    Address  string `gorm:"size:255" json:"address"`
}

// 文章模型
type Post struct {
    BaseModel
    UserID   uint      `gorm:"index" json:"user_id"`
    Title    string    `gorm:"size:200" json:"title"`
    Content  string    `gorm:"type:longtext" json:"content"`
    Status   int       `gorm:"default:1;index" json:"status"`
    Views    int       `gorm:"default:0" json:"views"`
    User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Tags     []Tag     `gorm:"many2many:post_tags;" json:"tags,omitempty"`
}

// 标签模型
type Tag struct {
    BaseModel
    Name  string `gorm:"uniqueIndex;size:50" json:"name"`
    Color string `gorm:"size:7;default:#007bff" json:"color"`
    Posts []Post `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}
```

### 数据库迁移

```go
package migrations

import (
    "gorm.io/gorm"
    "yourapp/models"
)

func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.User{},
        &models.Profile{},
        &models.Post{},
        &models.Tag{},
    )
}

// 创建初始数据
func CreateSeeds(db *gorm.DB) error {
    // 创建管理员用户
    admin := &models.User{
        Username: "admin",
        Email:    "admin@example.com",
        Password: hashPassword("admin123"),
        Status:   1,
    }
    
    if err := db.FirstOrCreate(admin, models.User{Username: "admin"}).Error; err != nil {
        return err
    }
    
    // 创建默认标签
    tags := []models.Tag{
        {Name: "技术", Color: "#007bff"},
        {Name: "生活", Color: "#28a745"},
        {Name: "随笔", Color: "#ffc107"},
    }
    
    for _, tag := range tags {
        db.FirstOrCreate(&tag, models.Tag{Name: tag.Name})
    }
    
    return nil
}
```

## Repository模式

### 基础Repository

```go
package repositories

import (
    "gorm.io/gorm"
    "yourapp/models"
)

type BaseRepository struct {
    DB *gorm.DB
}

func NewBaseRepository(db *gorm.DB) *BaseRepository {
    return &BaseRepository{DB: db}
}

// 通用CRUD操作
func (r *BaseRepository) Create(model interface{}) error {
    return r.DB.Create(model).Error
}

func (r *BaseRepository) GetByID(model interface{}, id uint) error {
    return r.DB.First(model, id).Error
}

func (r *BaseRepository) Update(model interface{}) error {
    return r.DB.Save(model).Error
}

func (r *BaseRepository) Delete(model interface{}, id uint) error {
    return r.DB.Delete(model, id).Error
}

func (r *BaseRepository) List(models interface{}, conditions ...interface{}) error {
    query := r.DB
    for _, condition := range conditions {
        query = query.Where(condition)
    }
    return query.Find(models).Error
}
```

### 用户Repository

```go
package repositories

import (
    "gorm.io/gorm"
    "yourapp/models"
)

type UserRepository struct {
    *BaseRepository
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        BaseRepository: NewBaseRepository(db),
    }
}

func (r *UserRepository) GetByUsername(username string) (*models.User, error) {
    var user models.User
    err := r.DB.Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.DB.Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetWithProfile(id uint) (*models.User, error) {
    var user models.User
    err := r.DB.Preload("Profile").First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *UserRepository) GetActiveUsers() ([]models.User, error) {
    var users []models.User
    err := r.DB.Where("status = ?", 1).Find(&users).Error
    return users, err
}

func (r *UserRepository) SearchUsers(keyword string, page, size int) ([]models.User, int64, error) {
    var users []models.User
    var total int64
    
    query := r.DB.Model(&models.User{})
    if keyword != "" {
        query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
    }
    
    // 获取总数
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // 分页查询
    offset := (page - 1) * size
    err := query.Offset(offset).Limit(size).Find(&users).Error
    
    return users, total, err
}
```

## 服务层集成

### 用户服务

```go
package services

import (
    "errors"
    "golang.org/x/crypto/bcrypt"
    "yourapp/models"
    "yourapp/repositories"
)

type UserService struct {
    userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
    return &UserService{
        userRepo: userRepo,
    }
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*models.User, error) {
    // 检查用户名是否存在
    if _, err := s.userRepo.GetByUsername(req.Username); err == nil {
        return nil, errors.New("用户名已存在")
    }
    
    // 检查邮箱是否存在
    if _, err := s.userRepo.GetByEmail(req.Email); err == nil {
        return nil, errors.New("邮箱已被使用")
    }
    
    // 加密密码
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }
    
    user := &models.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
        Status:   1,
    }
    
    if err := s.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    return user, nil
}

func (s *UserService) AuthenticateUser(username, password string) (*models.User, error) {
    user, err := s.userRepo.GetByUsername(username)
    if err != nil {
        return nil, errors.New("用户不存在")
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return nil, errors.New("密码错误")
    }
    
    return user, nil
}

func (s *UserService) GetUserProfile(id uint) (*models.User, error) {
    return s.userRepo.GetWithProfile(id)
}

func (s *UserService) UpdateUserProfile(id uint, req *UpdateProfileRequest) error {
    user, err := s.userRepo.GetByID(&models.User{}, id)
    if err != nil {
        return err
    }
    
    // 更新用户信息
    if req.Avatar != "" {
        user.(*models.User).Avatar = req.Avatar
    }
    
    return s.userRepo.Update(user)
}
```

## 控制器中使用数据库

### 用户控制器

```go
package controllers

import (
    "strconv"
    "yourapp/services"
    "github.com/zsy619/yyhertz/framework/mvc"
)

type UserController struct {
    mvc.BaseController
    userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
    return &UserController{
        userService: userService,
    }
}

func (c *UserController) PostRegister() {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "参数格式错误")
        return
    }
    
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        c.Error(400, err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "message": "注册成功",
        "user": map[string]interface{}{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
        },
    })
}

func (c *UserController) PostLogin() {
    username := c.GetForm("username")
    password := c.GetForm("password")
    
    if username == "" || password == "" {
        c.Error(400, "用户名和密码不能为空")
        return
    }
    
    user, err := c.userService.AuthenticateUser(username, password)
    if err != nil {
        c.Error(401, err.Error())
        return
    }
    
    // 生成token
    token, err := generateJWT(user.ID)
    if err != nil {
        c.Error(500, "生成token失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "message": "登录成功",
        "token":   token,
        "user": map[string]interface{}{
            "id":       user.ID,
            "username": user.Username,
            "email":    user.Email,
        },
    })
}

func (c *UserController) GetProfile() {
    idStr := c.GetParam("id")
    id, err := strconv.ParseUint(idStr, 10, 32)
    if err != nil {
        c.Error(400, "无效的用户ID")
        return
    }
    
    user, err := c.userService.GetUserProfile(uint(id))
    if err != nil {
        c.Error(404, "用户不存在")
        return
    }
    
    // 设置模板数据
    c.SetData("Title", "用户资料")
    c.SetData("User", user)
    c.RenderHTML("user/profile.html")
}
```

## 数据库事务

### 事务处理

```go
package services

import "gorm.io/gorm"

func (s *UserService) CreateUserWithProfile(req *CreateUserRequest) error {
    return s.userRepo.DB.Transaction(func(tx *gorm.DB) error {
        // 创建用户
        user := &models.User{
            Username: req.Username,
            Email:    req.Email,
            Password: hashPassword(req.Password),
        }
        
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        // 创建用户资料
        profile := &models.Profile{
            UserID:   user.ID,
            Nickname: req.Nickname,
            Bio:      req.Bio,
        }
        
        if err := tx.Create(profile).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

## 连接池配置

### 数据库配置优化

```go
func NewDatabaseWithPool(config *DatabaseConfig) (*gorm.DB, error) {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return nil, err
    }
    
    // 获取底层sql.DB
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 设置连接池参数
    sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
    sqlDB.SetMaxOpenConns(100)          // 最大连接数
    sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
    
    return db, nil
}
```

## 数据库监控

### 查询日志

```go
import (
    "log"
    "time"
    "gorm.io/gorm/logger"
)

func NewDatabaseWithLogging(config *DatabaseConfig) (*gorm.DB, error) {
    newLogger := logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             time.Second,   // 慢查询阈值
            LogLevel:                  logger.Info,   // 日志级别
            IgnoreRecordNotFoundError: true,          // 忽略记录未找到错误
            Colorful:                  true,          // 彩色输出
        },
    )
    
    return gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: newLogger,
    })
}
```

### 性能监控

```go
package middleware

import (
    "time"
    "gorm.io/gorm"
)

func DatabaseMetrics() gorm.Plugin {
    return &metricsPlugin{}
}

type metricsPlugin struct{}

func (p *metricsPlugin) Name() string {
    return "metrics"
}

func (p *metricsPlugin) Initialize(db *gorm.DB) error {
    db.Callback().Query().Before("gorm:query").Register("metrics:before", beforeQuery)
    db.Callback().Query().After("gorm:query").Register("metrics:after", afterQuery)
    return nil
}

func beforeQuery(db *gorm.DB) {
    db.Set("start_time", time.Now())
}

func afterQuery(db *gorm.DB) {
    if startTime, ok := db.Get("start_time"); ok {
        duration := time.Since(startTime.(time.Time))
        log.Printf("Query executed in %v: %s", duration, db.Statement.SQL.String())
    }
}
```

## 最佳实践

### 1. 数据库设计原则

- 合理的索引设计
- 适当的数据类型选择
- 规范的命名约定
- 合理的表关系设计

### 2. 查询优化

```go
// 预加载关联数据
users := []models.User{}
db.Preload("Profile").Preload("Posts").Find(&users)

// 选择特定字段
db.Select("id", "username", "email").Find(&users)

// 批量操作
db.CreateInBatches(users, 100)
```

### 3. 错误处理

```go
func (r *UserRepository) GetByID(id uint) (*models.User, error) {
    var user models.User
    err := r.DB.First(&user, id).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    
    return &user, nil
}
```

### 4. 数据迁移管理

```go
// 版本化迁移
type Migration struct {
    Version string
    Up      func(*gorm.DB) error
    Down    func(*gorm.DB) error
}

var migrations = []Migration{
    {
        Version: "20250101_create_users_table",
        Up: func(db *gorm.DB) error {
            return db.AutoMigrate(&models.User{})
        },
        Down: func(db *gorm.DB) error {
            return db.Migrator().DropTable(&models.User{})
        },
    },
}
```

---

数据库是应用的核心，合理的数据库设计和使用方式能够确保应用的高性能和可维护性！