# YYHertz 最佳实践指南

<div align="center">

🧪 **生产环境部署建议** | 企业级应用开发规范

</div>

---

## 📋 目录

- [项目架构最佳实践](#项目架构最佳实践)
- [代码规范与质量](#代码规范与质量)  
- [性能优化实践](#性能优化实践)
- [安全防护实践](#安全防护实践)
- [数据库设计规范](#数据库设计规范)
- [部署运维实践](#部署运维实践)
- [监控告警体系](#监控告警体系)
- [团队协作规范](#团队协作规范)

---

## 🏗️ 项目架构最佳实践

### 1. 标准目录结构

```
myapp/                              # 项目根目录
├── cmd/                           # 应用程序入口
│   └── server/
│       └── main.go               # 主程序入口
├── internal/                     # 私有代码，不可被外部import
│   ├── controllers/              # 控制器层
│   │   ├── admin/               # 管理员控制器
│   │   ├── api/                 # API控制器  
│   │   └── web/                 # Web页面控制器
│   ├── services/                # 业务逻辑层
│   │   ├── user.go
│   │   ├── order.go
│   │   └── notification.go
│   ├── repositories/            # 数据访问层
│   │   ├── interfaces/          # 仓储接口定义
│   │   └── mysql/              # MySQL实现
│   ├── models/                  # 数据模型
│   │   ├── entities/           # 业务实体
│   │   ├── dto/                # 数据传输对象
│   │   └── vo/                 # 视图对象
│   ├── middleware/              # 自定义中间件
│   ├── config/                  # 配置管理
│   └── pkg/                     # 内部工具包
├── web/                         # Web资源
│   ├── static/                  # 静态文件
│   │   ├── css/
│   │   ├── js/
│   │   └── images/
│   └── views/                   # 模板文件
│       ├── layouts/
│       ├── partials/
│       └── pages/
├── configs/                     # 配置文件
│   ├── app.yaml                # 应用配置
│   ├── database.yaml           # 数据库配置
│   └── redis.yaml              # Redis配置
├── scripts/                     # 构建和部署脚本
│   ├── build.sh
│   ├── deploy.sh
│   └── migrate.sh
├── deployments/                 # 部署文件
│   ├── docker/
│   ├── kubernetes/
│   └── docker-compose.yml
├── tests/                       # 测试文件
│   ├── integration/            # 集成测试
│   ├── e2e/                    # 端到端测试
│   └── fixtures/               # 测试数据
├── docs/                        # 项目文档
├── go.mod                       # Go模块文件
├── go.sum                       # 依赖校验文件
├── Makefile                     # 构建任务
├── Dockerfile                   # Docker镜像构建
├── .gitignore                   # Git忽略文件
├── .env.example                 # 环境变量示例
└── README.md                    # 项目说明
```

### 2. 分层架构设计

```go
// 控制器层 - 处理HTTP请求
type UserController struct {
    userService services.UserServiceInterface
}

func (c *UserController) GetProfile(ctx *mvc.Context) {
    userID := ctx.GetInt("user_id")
    
    user, err := c.userService.GetByID(userID)
    if err != nil {
        ctx.ErrorJSON(500, "获取用户信息失败")
        return
    }
    
    ctx.JSON(user)
}

// 业务逻辑层 - 处理业务规则
type UserService struct {
    userRepo repositories.UserRepositoryInterface
    logger   *logrus.Logger
}

func (s *UserService) GetByID(id int) (*models.User, error) {
    // 业务逻辑：参数验证
    if id <= 0 {
        return nil, errors.New("invalid user id")
    }
    
    // 调用数据访问层
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        s.logger.Error("failed to find user", "id", id, "error", err)
        return nil, err
    }
    
    // 业务逻辑：数据处理
    user.SensitiveDataMask()
    
    return user, nil
}

// 数据访问层 - 处理数据存储
type UserRepository struct {
    db *gorm.DB
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    return &user, err
}
```

### 3. 依赖注入设计

```go
// 使用接口实现依赖倒置
type UserServiceInterface interface {
    GetByID(id int) (*models.User, error)
    Create(user *models.User) error
    Update(user *models.User) error
    Delete(id int) error
}

type UserRepositoryInterface interface {
    FindByID(id int) (*models.User, error)
    Create(user *models.User) error
    Update(user *models.User) error
    Delete(id int) error
}

// 依赖注入容器
type Container struct {
    userService UserServiceInterface
    userRepo    UserRepositoryInterface
}

func NewContainer() *Container {
    // 初始化依赖
    db := orm.GetDB()
    userRepo := repositories.NewUserRepository(db)
    userService := services.NewUserService(userRepo)
    
    return &Container{
        userService: userService,
        userRepo:    userRepo,
    }
}

// 控制器工厂
func (c *Container) NewUserController() *UserController {
    return &UserController{
        userService: c.userService,
    }
}
```

---

## 📝 代码规范与质量

### 1. Go代码规范

```go
// ✅ 好的实践：清晰的函数命名和注释
// CreateUser 创建新用户，包含数据验证和权限检查
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // 参数验证
    if err := s.validateCreateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // 业务逻辑
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Status:   UserStatusActive,
    }
    
    // 密码哈希
    hashedPassword, err := s.hashPassword(req.Password)
    if err != nil {
        return nil, fmt.Errorf("failed to hash password: %w", err)
    }
    user.PasswordHash = hashedPassword
    
    // 数据存储
    if err := s.userRepo.Create(user); err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}

// ❌ 避免的做法：函数过长、无注释、错误处理不当
func (s *UserService) BadCreateUser(username, email, password string) *User {
    user := &User{}
    user.Username = username
    user.Email = email
    // 直接存储明文密码 - 危险！
    user.Password = password
    s.userRepo.Create(user) // 忽略错误 - 危险！
    return user
}
```

### 2. 错误处理规范

```go
// 自定义错误类型
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Cause   error  `json:"-"`
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

// 错误包装和分类
var (
    ErrUserNotFound     = &AppError{Code: 4041, Message: "用户不存在"}
    ErrInvalidPassword  = &AppError{Code: 4001, Message: "密码错误"}
    ErrUserExists       = &AppError{Code: 4091, Message: "用户已存在"}
    ErrDatabaseError    = &AppError{Code: 5001, Message: "数据库错误"}
)

// 统一错误处理中间件
func ErrorHandlerMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        c.Next()
        
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            var appErr *AppError
            if errors.As(err, &appErr) {
                c.JSON(appErr.Code/100, appErr)
            } else {
                c.JSON(500, &AppError{
                    Code:    5000,
                    Message: "内部服务器错误",
                })
            }
        }
    }
}
```

### 3. 日志记录规范

```go
// 结构化日志配置
func setupLogger() *logrus.Logger {
    logger := logrus.New()
    
    // JSON格式便于日志分析
    logger.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: time.RFC3339,
    })
    
    // 设置日志级别
    if config.GetBool("app.debug") {
        logger.SetLevel(logrus.DebugLevel)
    } else {
        logger.SetLevel(logrus.InfoLevel)
    }
    
    // 日志文件轮转
    writer := &lumberjack.Logger{
        Filename:   "logs/app.log",
        MaxSize:    100, // MB
        MaxBackups: 5,
        MaxAge:     30, // days
        Compress:   true,
    }
    logger.SetOutput(writer)
    
    return logger
}

// 业务日志记录实践
func (s *UserService) LoginUser(username, password string) (*User, error) {
    // 记录尝试登录
    s.logger.WithFields(logrus.Fields{
        "action":   "user_login",
        "username": username,
        "ip":       s.getClientIP(),
    }).Info("User login attempt")
    
    user, err := s.userRepo.FindByUsername(username)
    if err != nil {
        // 记录失败原因
        s.logger.WithFields(logrus.Fields{
            "action":   "user_login",
            "username": username,
            "error":    "user_not_found",
        }).Warn("Login failed: user not found")
        
        return nil, ErrUserNotFound
    }
    
    if !s.checkPassword(password, user.PasswordHash) {
        // 记录密码错误
        s.logger.WithFields(logrus.Fields{
            "action":   "user_login", 
            "username": username,
            "error":    "invalid_password",
        }).Warn("Login failed: invalid password")
        
        return nil, ErrInvalidPassword
    }
    
    // 记录成功登录
    s.logger.WithFields(logrus.Fields{
        "action":  "user_login",
        "user_id": user.ID,
        "username": username,
    }).Info("User login successful")
    
    return user, nil
}
```

---

## ⚡ 性能优化实践

### 1. 数据库优化

```go
// 连接池优化配置
func setupDatabase() *gorm.DB {
    dsn := config.GetString("database.dsn")
    
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        // 预编译语句缓存
        PrepareStmt: true,
        
        // 批量插入大小
        CreateBatchSize: 1000,
        
        // 禁用自动事务
        SkipDefaultTransaction: true,
        
        // 自定义日志
        Logger: logger.New(
            log.New(os.Stdout, "\r\n", log.LstdFlags),
            logger.Config{
                SlowThreshold: time.Second,   // 慢查询阈值
                LogLevel:      logger.Silent, // 生产环境静默
            },
        ),
    })
    
    if err != nil {
        panic("failed to connect database")
    }
    
    // 连接池配置
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)                    // 最大空闲连接
    sqlDB.SetMaxOpenConns(100)                   // 最大打开连接
    sqlDB.SetConnMaxLifetime(time.Hour)          // 连接最大生存时间
    sqlDB.SetConnMaxIdleTime(time.Minute * 10)   // 连接最大空闲时间
    
    return db
}

// 查询优化实践
func (r *UserRepository) FindUsersWithPagination(page, size int, filters *UserFilters) ([]User, int64, error) {
    var users []User
    var total int64
    
    // 构建基础查询
    query := r.db.Model(&User{})
    
    // 应用过滤条件
    if filters.Status != nil {
        query = query.Where("status = ?", *filters.Status)
    }
    if filters.CreatedAfter != nil {
        query = query.Where("created_at > ?", *filters.CreatedAfter)
    }
    
    // 先获取总数（使用count查询，不加载全部数据）
    if err := query.Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // 分页查询（只查询需要的字段）
    offset := (page - 1) * size
    err := query.Select("id, username, email, status, created_at").
        Offset(offset).
        Limit(size).
        Order("created_at DESC").
        Find(&users).Error
    
    return users, total, err
}

// 批量操作优化
func (r *UserRepository) CreateUsersInBatch(users []User) error {
    // 批量插入，提高性能
    return r.db.CreateInBatches(users, 100).Error
}

// 预加载优化，避免N+1问题
func (r *UserRepository) FindUsersWithProfiles(page, size int) ([]User, error) {
    var users []User
    
    err := r.db.Preload("Profile").
        Preload("Roles").
        Offset((page-1)*size).
        Limit(size).
        Find(&users).Error
    
    return users, err
}
```

### 2. 缓存策略

```go
// Redis缓存配置
func setupRedis() *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:            config.GetString("redis.addr"),
        Password:        config.GetString("redis.password"),
        DB:              config.GetInt("redis.db"),
        PoolSize:        20,                    // 连接池大小
        MinIdleConns:    5,                     // 最小空闲连接
        MaxRetries:      3,                     // 重试次数
        DialTimeout:     5 * time.Second,       // 连接超时
        ReadTimeout:     3 * time.Second,       // 读取超时
        WriteTimeout:    3 * time.Second,       // 写入超时
        PoolTimeout:     4 * time.Second,       // 池获取连接超时
        IdleTimeout:     5 * time.Minute,       // 空闲连接超时
    })
    
    return rdb
}

// 缓存服务封装
type CacheService struct {
    redis  *redis.Client
    logger *logrus.Logger
}

func (c *CacheService) Get(key string) (string, error) {
    val, err := c.redis.Get(context.Background(), key).Result()
    if err == redis.Nil {
        return "", nil // 缓存未命中
    }
    return val, err
}

func (c *CacheService) Set(key, value string, expiration time.Duration) error {
    return c.redis.Set(context.Background(), key, value, expiration).Err()
}

func (c *CacheService) Delete(key string) error {
    return c.redis.Del(context.Background(), key).Err()
}

// 业务缓存应用
func (s *UserService) GetUserByID(id int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // 尝试从缓存获取
    cached, err := s.cache.Get(cacheKey)
    if err == nil && cached != "" {
        var user User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            s.logger.Debug("cache hit", "key", cacheKey)
            return &user, nil
        }
    }
    
    // 缓存未命中，从数据库获取
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存
    if data, err := json.Marshal(user); err == nil {
        s.cache.Set(cacheKey, string(data), time.Hour)
        s.logger.Debug("cache set", "key", cacheKey)
    }
    
    return user, nil
}

// 缓存穿透防护
func (s *UserService) GetUserByIDSafe(id int) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", id)
    
    // 使用singleflight防止缓存击穿
    result, err, shared := s.singleFlight.Do(cacheKey, func() (interface{}, error) {
        return s.getUserFromDBWithCache(id)
    })
    
    if shared {
        s.logger.Debug("singleflight hit", "key", cacheKey)
    }
    
    if err != nil {
        return nil, err
    }
    
    return result.(*User), nil
}
```

---

## 🔒 安全防护实践

### 1. 输入验证和过滤

```go
import "github.com/go-playground/validator/v10"

// 请求验证结构体
type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%^&*"`
    Phone    string `json:"phone" validate:"omitempty,len=11,numeric"`
}

// 自定义验证器
func setupValidator() *validator.Validate {
    validate := validator.New()
    
    // 自定义验证规则
    validate.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
        password := fl.Field().String()
        return hasUppercase(password) && 
               hasLowercase(password) && 
               hasDigit(password) && 
               hasSpecialChar(password)
    })
    
    return validate
}

// 控制器中的验证
func (c *UserController) CreateUser(ctx *mvc.Context) {
    var req CreateUserRequest
    
    // 绑定并验证请求
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.ErrorJSON(400, "Invalid JSON format")
        return
    }
    
    if err := c.validator.Struct(&req); err != nil {
        errors := make(map[string]string)
        for _, err := range err.(validator.ValidationErrors) {
            errors[err.Field()] = getValidationErrorMsg(err)
        }
        
        ctx.JSON(422, map[string]interface{}{
            "error": "Validation failed",
            "details": errors,
        })
        return
    }
    
    // 额外的业务验证
    if c.userService.IsUsernameExists(req.Username) {
        ctx.ErrorJSON(409, "Username already exists")
        return
    }
    
    // 创建用户
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        ctx.ErrorJSON(500, "Failed to create user")
        return
    }
    
    ctx.JSON(201, user)
}
```

### 2. 认证和授权

```go
// JWT服务实现
type JWTService struct {
    secretKey []byte
    issuer    string
    expiry    time.Duration
}

func NewJWTService(secret, issuer string, expiry time.Duration) *JWTService {
    return &JWTService{
        secretKey: []byte(secret),
        issuer:    issuer,
        expiry:    expiry,
    }
}

func (j *JWTService) GenerateToken(userID int, username string, roles []string) (string, error) {
    claims := jwt.MapClaims{
        "user_id":  userID,
        "username": username,
        "roles":    roles,
        "iss":      j.issuer,
        "exp":      time.Now().Add(j.expiry).Unix(),
        "iat":      time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(j.secretKey)
}

func (j *JWTService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return j.secretKey, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return &claims, nil
    }
    
    return nil, errors.New("invalid token")
}

// RBAC权限检查
type AuthorizationService struct {
    userService *UserService
}

func (a *AuthorizationService) HasPermission(userID int, resource, action string) bool {
    user, err := a.userService.GetUserByID(userID)
    if err != nil {
        return false
    }
    
    for _, role := range user.Roles {
        for _, permission := range role.Permissions {
            if permission.Resource == resource && permission.Action == action {
                return true
            }
        }
    }
    
    return false
}

// 权限检查中间件
func RequirePermission(resource, action string) mvc.HandlerFunc {
    return func(c *mvc.Context) {
        userID := c.GetInt("user_id")
        
        authSvc := c.MustGet("auth_service").(*AuthorizationService)
        if !authSvc.HasPermission(userID, resource, action) {
            c.AbortWithJSON(403, map[string]string{
                "error": "Insufficient permissions",
            })
            return
        }
        
        c.Next()
    }
}

// 使用权限中间件
userRoutes := app.Group("/api/users", middleware.JWTAuth())
{
    userRoutes.GET("", RequirePermission("user", "read"), getUserList)
    userRoutes.POST("", RequirePermission("user", "create"), createUser)
    userRoutes.DELETE("/:id", RequirePermission("user", "delete"), deleteUser)
}
```

---

## 🗄️ 数据库设计规范

### 1. 表设计规范

```sql
-- 用户表设计示例
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID',
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    email VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱地址',
    phone VARCHAR(20) DEFAULT NULL COMMENT '手机号码',
    password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希',
    status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '状态：0=禁用，1=正常',
    avatar VARCHAR(255) DEFAULT NULL COMMENT '头像URL',
    last_login_at TIMESTAMP NULL DEFAULT NULL COMMENT '最后登录时间',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at TIMESTAMP NULL DEFAULT NULL COMMENT '删除时间（软删除）',
    
    -- 索引设计
    INDEX idx_users_email (email),
    INDEX idx_users_phone (phone),
    INDEX idx_users_status (status),
    INDEX idx_users_created_at (created_at),
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 角色表
CREATE TABLE roles (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE COMMENT '角色名称',
    description TEXT COMMENT '角色描述',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- 用户角色关联表
CREATE TABLE user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id INT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';
```

### 2. 数据迁移规范

```go
// 迁移文件结构
type Migration struct {
    Version     string
    Description string
    Up          func(*gorm.DB) error
    Down        func(*gorm.DB) error
}

// 创建用户表迁移
func CreateUsersTable() *Migration {
    return &Migration{
        Version:     "20240905_001",
        Description: "Create users table",
        Up: func(db *gorm.DB) error {
            return db.Exec(`
                CREATE TABLE users (
                    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
                    username VARCHAR(50) NOT NULL UNIQUE,
                    email VARCHAR(100) NOT NULL UNIQUE,
                    password_hash VARCHAR(255) NOT NULL,
                    status TINYINT UNSIGNED NOT NULL DEFAULT 1,
                    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
                    deleted_at TIMESTAMP NULL DEFAULT NULL,
                    
                    INDEX idx_users_email (email),
                    INDEX idx_users_status (status)
                ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
            `).Error
        },
        Down: func(db *gorm.DB) error {
            return db.Exec("DROP TABLE IF EXISTS users").Error
        },
    }
}

// 迁移管理器
type MigrationManager struct {
    db         *gorm.DB
    migrations []*Migration
}

func (m *MigrationManager) RunMigrations() error {
    // 创建迁移记录表
    if err := m.createMigrationTable(); err != nil {
        return err
    }
    
    for _, migration := range m.migrations {
        if !m.isMigrationExecuted(migration.Version) {
            log.Printf("Running migration: %s - %s", migration.Version, migration.Description)
            
            if err := migration.Up(m.db); err != nil {
                return fmt.Errorf("migration %s failed: %w", migration.Version, err)
            }
            
            if err := m.recordMigration(migration); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

---

## 🚀 部署运维实践

### 1. Docker化部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

# 安装必要工具
RUN apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum，利用Docker缓存层
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 安装CA证书和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 创建应用用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/web ./web

# 更改所有者
RUN chown -R appuser:appgroup /root

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8888

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8888/health || exit 1

# 启动应用
CMD ["./main"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8888:8888"
    environment:
      - GO_ENV=production
      - DB_HOST=mysql
      - REDIS_HOST=redis
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - app-network

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: myapp
      MYSQL_USER: appuser
      MYSQL_PASSWORD: apppassword
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      timeout: 10s
      retries: 5
    restart: unless-stopped
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      timeout: 10s
      retries: 5
    restart: unless-stopped
    networks:
      - app-network

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf
      - ./nginx/ssl:/etc/nginx/ssl
      - ./web/static:/usr/share/nginx/html/static
    depends_on:
      - app
    restart: unless-stopped
    networks:
      - app-network

volumes:
  mysql_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

### 2. Kubernetes部署

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  labels:
    app: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp
        image: myapp:v2.0.0
        ports:
        - containerPort: 8888
        env:
        - name: GO_ENV
          value: "production"
        - name: DB_HOST
          valueFrom:
            configMapKeyRef:
              name: myapp-config
              key: db_host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: myapp-secret
              key: db_password
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        readinessProbe:
          httpGet:
            path: /health
            port: 8888
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8888
          initialDelaySeconds: 30
          periodSeconds: 10

---
apiVersion: v1
kind: Service
metadata:
  name: myapp-service
spec:
  selector:
    app: myapp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8888
  type: LoadBalancer

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: myapp-config
data:
  db_host: "mysql-service"
  redis_host: "redis-service"
  log_level: "info"

---
apiVersion: v1
kind: Secret
metadata:
  name: myapp-secret
type: Opaque
data:
  db_password: <base64-encoded-password>
  jwt_secret: <base64-encoded-jwt-secret>
```

---

## 📊 监控告警体系

### 1. 应用监控

```go
// 监控指标收集
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP请求总数
    httpRequests = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    // HTTP请求持续时间
    httpDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    
    // 数据库连接数
    dbConnections = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "db_connections",
            Help: "Current database connections",
        },
        []string{"status"}, // active, idle
    )
)

// 监控中间件
func PrometheusMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        
        httpRequests.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
            strconv.Itoa(c.Writer.Status()),
        ).Inc()
        
        httpDuration.WithLabelValues(
            c.Request.Method,
            c.FullPath(),
        ).Observe(duration)
    }
}

// 业务监控指标
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // 业务指标：用户注册数
    userRegistrations.Inc()
    
    start := time.Now()
    user, err := s.userRepo.Create(req)
    
    // 数据库操作耗时
    dbOperationDuration.WithLabelValues("users", "create").Observe(time.Since(start).Seconds())
    
    if err != nil {
        // 错误计数
        dbErrors.WithLabelValues("users", "create").Inc()
        return nil, err
    }
    
    return user, nil
}
```

### 2. 健康检查

```go
// 健康检查端点
type HealthController struct {
    db    *gorm.DB
    redis *redis.Client
}

func (h *HealthController) GetHealth(c *mvc.Context) {
    status := "healthy"
    checks := make(map[string]interface{})
    
    // 检查数据库连接
    if err := h.checkDatabase(); err != nil {
        status = "unhealthy"
        checks["database"] = map[string]interface{}{
            "status": "error",
            "error":  err.Error(),
        }
    } else {
        checks["database"] = map[string]interface{}{
            "status": "ok",
        }
    }
    
    // 检查Redis连接
    if err := h.checkRedis(); err != nil {
        status = "unhealthy"
        checks["redis"] = map[string]interface{}{
            "status": "error",
            "error":  err.Error(),
        }
    } else {
        checks["redis"] = map[string]interface{}{
            "status": "ok",
        }
    }
    
    response := map[string]interface{}{
        "status":    status,
        "timestamp": time.Now().Unix(),
        "checks":    checks,
        "version":   config.GetString("app.version"),
    }
    
    if status == "healthy" {
        c.JSON(200, response)
    } else {
        c.JSON(503, response)
    }
}

func (h *HealthController) checkDatabase() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    sqlDB, err := h.db.DB()
    if err != nil {
        return err
    }
    
    return sqlDB.PingContext(ctx)
}

func (h *HealthController) checkRedis() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    return h.redis.Ping(ctx).Err()
}
```

---

## 👥 团队协作规范

### 1. Git工作流

```bash
# 功能开发流程
git checkout -b feature/user-management    # 创建功能分支
# ... 开发代码 ...
git add .
git commit -m "feat: add user CRUD operations"  # 遵循conventional commits规范
git push origin feature/user-management    # 推送到远程
# 创建Pull Request

# 提交信息规范
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式化
refactor: 代码重构
test: 测试相关
chore: 构建过程或辅助工具的变动
```

### 2. 代码审查清单

```markdown
## 代码审查清单

### 功能性
- [ ] 代码实现了预期功能
- [ ] 边界条件处理正确
- [ ] 错误处理完善
- [ ] 测试覆盖充分

### 代码质量
- [ ] 代码结构清晰
- [ ] 命名规范统一
- [ ] 注释适当详细
- [ ] 无重复代码

### 性能
- [ ] 无明显性能问题
- [ ] 数据库查询优化
- [ ] 内存使用合理
- [ ] 缓存使用恰当

### 安全
- [ ] 输入验证充分
- [ ] 权限检查正确
- [ ] 无敏感信息泄露
- [ ] SQL注入防护

### 可维护性
- [ ] 遵循项目架构
- [ ] 依赖关系清晰
- [ ] 配置外部化
- [ ] 日志记录完整
```

### 3. 自动化CI/CD

```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: testpassword
          MYSQL_DATABASE: testdb
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3
      
      redis:
        image: redis:7
        ports:
          - 6379:6379
        options: >-
          --health-cmd="redis-cli ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3

    steps:
    - uses: actions/checkout@v2
    
    - name: Set up Go
      uses: actions/setup-go@v2
      with:
        go-version: 1.21
    
    - name: Cache dependencies
      uses: actions/cache@v2
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage reports
      uses: codecov/codecov-action@v1
      with:
        file: ./coverage.out
    
    - name: Run security scan
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec ./...
    
    - name: Build
      run: go build -v ./cmd/server

  deploy:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v2
    
    - name: Build and push Docker image
      env:
        DOCKER_REGISTRY: ${{ secrets.DOCKER_REGISTRY }}
        IMAGE_TAG: ${{ github.sha }}
      run: |
        docker build -t $DOCKER_REGISTRY/myapp:$IMAGE_TAG .
        docker push $DOCKER_REGISTRY/myapp:$IMAGE_TAG
    
    - name: Deploy to production
      run: |
        # 部署脚本
        echo "Deploying to production..."
```

---

<div align="center">

**🧪 遵循这些最佳实践，构建高质量、可维护、安全的YYHertz应用！**

**好的实践是成功项目的基础 🏗️**

</div>