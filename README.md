# YYHertz MVC Framework

基于CloudWeGo-Hertz的现代化Go Web框架，提供完整的Beego风格开发体验，兼具高性能与开发效率。

## 🚀 核心特性

- **🏗️ MVC架构** - 完整的Model-View-Controller设计模式
- **📁 Beego风格Namespace** - 100%兼容Beego的命名空间路由系统
- **🎛️ 智能路由** - 自动路由注册 + 手动路由映射，支持RESTful设计
- **🗄️ 双ORM支持** - 内置GORM和MyBatis-Go双ORM解决方案
- **🎨 模板引擎** - 内置HTML模板支持，布局和组件化开发
- **🔌 统一中间件系统** - 智能中间件管道：4层架构、自动编译优化、性能缓存、兼容性适配
- **⚡ 高性能** - 基于CloudWeGo-Hertz，提供卓越的性能表现
- **🔧 配置管理** - 基于Viper的配置系统，支持多种格式
- **📊 可观测性** - 内置日志、链路追踪、监控指标
- **🛡️ 生产就绪** - 完善的错误处理、优雅关闭、健康检查

## 📦 项目结构

```
YYHertz/
├── 📁 framework/                    # 🏗️ 框架核心
│   ├── mvc/                         # 🆕 统一MVC核心组件
│   │   ├── core/                   # 核心应用和控制器
│   │   │   ├── app.go              # 应用实例和路由管理
│   │   │   ├── controller.go       # 基础控制器实现
│   │   │   ├── controller_*.go     # 控制器功能模块
│   │   │   └── factory.go          # 控制器工厂
│   │   ├── annotation/             # 🔗 注解路由系统
│   │   │   ├── annotations.go      # 注解定义和解析
│   │   │   ├── auto_router.go      # 自动路由生成
│   │   │   └── parser.go           # 注解解析器
│   │   ├── middleware/             # 🔌 统一中间件系统 (原@framework/middleware合并)
│   │   │   ├── manager.go          # 中间件管理器
│   │   │   ├── pipeline.go         # 中间件管道
│   │   │   ├── compiler.go         # 智能编译器
│   │   │   ├── adapter.go          # 兼容性适配器
│   │   │   ├── unified_manager.go  # 统一管理器
│   │   │   ├── builtin_*.go        # 内置中间件集合
│   │   │   ├── auth.go             # 身份认证中间件
│   │   │   ├── cors.go             # 跨域中间件
│   │   │   ├── logger.go           # 日志中间件
│   │   │   ├── recovery.go         # 恢复中间件
│   │   │   ├── ratelimit.go        # 限流中间件
│   │   │   ├── tracing.go          # 链路追踪中间件
│   │   │   └── benchmark_test.go   # 性能基准测试
│   │   ├── context/                # 🔗 统一上下文系统 (原@framework/context合并)
│   │   │   ├── pool.go             # 上下文池化管理
│   │   │   ├── enhanced.go         # 增强上下文功能
│   │   │   └── adapter.go          # 兼容性适配器
│   │   ├── namespace.go            # 🆕 Beego风格命名空间
│   │   ├── router/                 # 路由系统
│   │   │   ├── group.go            # 路由组管理
│   │   │   └── router.go           # 路由注册
│   │   └── session/                # 会话管理
│   │       ├── config.go           # 会话配置
│   │       ├── manager.go          # 会话管理器
│   │       └── store.go            # 会话存储
│   ├── orm/                        # 🗄️ ORM 数据库层
│   │   ├── gorm.go                 # GORM集成和配置
│   │   ├── enhanced.go             # 增强型GORM功能
│   │   ├── transaction.go          # 事务管理
│   │   ├── pool.go                 # 连接池管理
│   │   ├── migration.go            # 数据库迁移
│   │   └── metrics.go              # 数据库性能监控
│   ├── mybatis/                    # 🗂️ MyBatis-Go 实现
│   │   ├── mybatis.go              # MyBatis核心引擎
│   │   ├── config/                 # 配置管理
│   │   │   ├── configuration.go    # 全局配置
│   │   │   └── mapper_proxy.go     # Mapper代理
│   │   ├── session/                # 会话管理
│   │   │   ├── sql_session.go      # SQL会话
│   │   │   ├── executor.go         # SQL执行器
│   │   │   └── sql_session_factory.go # 会话工厂
│   │   ├── mapper/                 # Mapper管理
│   │   │   └── dynamic_sql.go      # 动态SQL构建
│   │   └── cache/                  # 缓存系统
│   │       └── cache.go            # 缓存实现
│   ├── config/                     # ⚙️ 配置管理
│   │   ├── viper_config.go         # Viper配置实现
│   │   ├── log_config.go           # 日志配置
│   │   ├── app_config.go           # 应用配置
│   │   ├── template_config.go      # 模板配置
│   │   └── middleware_unified_config.go # 统一中间件配置
│   ├── template/                   # 🎨 模板引擎
│   │   ├── manager.go              # 模板管理器
│   │   └── enhanced_manager.go     # 增强模板功能
│   ├── validation/                 # ✅ 数据验证
│   │   ├── validator.go            # 验证器核心
│   │   ├── rules.go                # 验证规则
│   │   └── messages.go             # 错误消息
│   ├── cache/                      # 💾 缓存系统
│   │   ├── cache.go                # 本地缓存
│   │   └── distributed_cache.go    # 分布式缓存
│   ├── scheduler/                  # ⏰ 任务调度
│   │   ├── scheduler.go            # 调度器核心
│   │   ├── cron.go                 # Cron任务
│   │   └── executor.go             # 任务执行器
│   ├── util/                       # 🛠️ 工具集合
│   │   ├── string.go               # 字符串工具
│   │   ├── datetime.go             # 日期时间工具
│   │   ├── crypto.go               # 加密工具
│   │   ├── validation.go           # 验证工具
│   │   └── network.go              # 网络工具
│   └── testing/                    # 🧪 测试工具
│       ├── testing.go              # 测试框架
│       ├── mock.go                 # Mock工具
│       └── assert.go               # 断言工具
├── 📁 example/                     # 📚 完整示例
│   ├── simple/                     # 基础示例项目
│   │   ├── controllers/            # 示例控制器
│   │   │   ├── home_controller.go  # 首页控制器
│   │   │   ├── user_controller.go  # 用户控制器
│   │   │   ├── admin_controller.go # 管理员控制器
│   │   │   └── markdown_controller.go # Markdown控制器
│   │   ├── views/                  # 模板文件
│   │   │   ├── layout/             # 布局模板
│   │   │   ├── home/               # 首页模板
│   │   │   ├── user/               # 用户模板
│   │   │   └── admin/              # 管理模板
│   │   ├── static/                 # 静态资源
│   │   │   ├── css/                # 样式文件
│   │   │   ├── js/                 # JavaScript文件
│   │   │   └── images/             # 图片资源
│   │   ├── docs/                   # 文档文件
│   │   ├── conf/                   # 配置文件
│   │   └── main.go                 # 示例入口
│   ├── annotations/                # 注解路由示例
│   ├── comments/                   # 注释路由示例
│   └── mybat/                      # MyBatis示例
├── 📁 tools/                       # 🔧 开发工具
│   ├── analyze/                    # 代码分析工具
│   ├── test/                       # 测试工具
│   └── verify/                     # 验证工具
├── 📁 logs/                        # 📝 日志文件
├── go.mod                          # Go模块定义
├── go.sum                          # 依赖校验和
├── README.md                       # 📖 项目文档
├── MYBATIS_SAMPLES.md              # MyBatis示例文档
└── VERSION_USAGE.md                # 版本使用说明
```

## 🛠️ 快速开始

### 1. 安装框架

```bash
# 克隆项目
git clone https://github.com/zsy619/yyhertz.git
cd YYHertz

# 安装依赖
go mod tidy

# 运行示例
go run example/simple/main.go

# 访问应用
open http://localhost:8888
```

### 2. 创建第一个应用

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

type HomeController struct {
    mvc.BaseController
}

func (c *HomeController) GetIndex() {
    c.JSON(map[string]any{
        "message": "Hello YYHertz!",
        "version": "1.0.0",
    })
}

func main() {
    app := mvc.HertzApp
    
    // 添加中间件 (统一中间件系统)
    app.Use(
        middleware.Recovery(), // 统一后的中间件API
        middleware.Logger(),
        middleware.CORS(),
    )
    
    // 注册控制器
    app.AutoRouters(&HomeController{})
    
    app.Run(":8888")
}
```

## 🗄️ 数据库集成

YYHertz提供两种强大的ORM解决方案，可以根据项目需求选择使用：

### 🔗 GORM 集成

GORM是Go语言最流行的ORM库，YYHertz对其进行了深度集成和增强。

#### 基本配置

```go
// config/database.yaml
database:
  driver: "mysql"
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: "1h"
  conn_max_idle_time: "30m"
  log_level: "info"
  enable_metrics: true
```

#### 模型定义

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

// 用户模型
type User struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Username string `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email    string `gorm:"uniqueIndex;size:100;not null" json:"email"`
    Password string `gorm:"size:255;not null" json:"-"`
    Avatar   string `gorm:"size:255" json:"avatar"`
    Status   int    `gorm:"default:1" json:"status"` // 1:正常 0:禁用
    
    // 关联关系
    Profile UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Posts   []Post      `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
}

// 用户资料模型
type UserProfile struct {
    ID     uint   `gorm:"primarykey" json:"id"`
    UserID uint   `gorm:"uniqueIndex;not null" json:"user_id"`
    
    RealName string `gorm:"size:50" json:"real_name"`
    Phone    string `gorm:"size:20" json:"phone"`
    Address  string `gorm:"size:255" json:"address"`
    Bio      string `gorm:"type:text" json:"bio"`
}

// 文章模型
type Post struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Title     string `gorm:"size:200;not null" json:"title"`
    Content   string `gorm:"type:longtext" json:"content"`
    Summary   string `gorm:"size:500" json:"summary"`
    AuthorID  uint   `gorm:"not null" json:"author_id"`
    Status    int    `gorm:"default:1" json:"status"` // 1:发布 0:草稿
    ViewCount int    `gorm:"default:0" json:"view_count"`
    
    // 关联关系
    Author User `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}
```

#### 数据库操作示例

```go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/orm"
    "your-project/models"
)

type UserController struct {
    mvc.BaseController
}

// 获取用户列表
func (c *UserController) GetList() {
    var users []models.User
    var total int64
    
    page := c.GetInt("page", 1)
    limit := c.GetInt("limit", 10)
    search := c.GetString("search")
    
    db := orm.GetDB()
    query := db.Model(&models.User{})
    
    // 搜索条件
    if search != "" {
        query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
    }
    
    // 统计总数
    query.Count(&total)
    
    // 分页查询
    result := query.Preload("Profile").
        Offset((page - 1) * limit).
        Limit(limit).
        Find(&users)
    
    if result.Error != nil {
        c.Error(500, "查询失败: "+result.Error.Error())
        return
    }
    
    c.JSON(map[string]any{
        "success": true,
        "data": map[string]any{
            "list":  users,
            "total": total,
            "page":  page,
            "limit": limit,
        },
    })
}

// 创建用户
func (c *UserController) PostCreate() {
    var user models.User
    
    // 绑定请求数据
    if err := c.GetCtx().BindAndValidate(&user); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 开始事务
    tx := orm.GetDB().Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // 创建用户
    if err := tx.Create(&user).Error; err != nil {
        tx.Rollback()
        c.Error(500, "创建用户失败: "+err.Error())
        return
    }
    
    // 创建用户资料
    profile := models.UserProfile{
        UserID:   user.ID,
        RealName: c.GetForm("real_name"),
        Phone:    c.GetForm("phone"),
    }
    
    if err := tx.Create(&profile).Error; err != nil {
        tx.Rollback()
        c.Error(500, "创建用户资料失败: "+err.Error())
        return
    }
    
    // 提交事务
    tx.Commit()
    
    c.JSON(map[string]any{
        "success": true,
        "message": "用户创建成功",
        "data":    user,
    })
}

// 更新用户
func (c *UserController) PutUpdate() {
    id := c.GetInt("id")
    if id == 0 {
        c.Error(400, "用户ID不能为空")
        return
    }
    
    var user models.User
    db := orm.GetDB()
    
    // 查找用户
    if err := db.First(&user, id).Error; err != nil {
        c.Error(404, "用户不存在")
        return
    }
    
    // 更新数据
    updates := map[string]any{
        "username": c.GetForm("username"),
        "email":    c.GetForm("email"),
        "status":   c.GetInt("status", 1),
    }
    
    if err := db.Model(&user).Updates(updates).Error; err != nil {
        c.Error(500, "更新失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]any{
        "success": true,
        "message": "更新成功",
        "data":    user,
    })
}

// 删除用户
func (c *UserController) DeleteRemove() {
    id := c.GetInt("id")
    if id == 0 {
        c.Error(400, "用户ID不能为空")
        return
    }
    
    db := orm.GetDB()
    
    // 软删除
    if err := db.Delete(&models.User{}, id).Error; err != nil {
        c.Error(500, "删除失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]any{
        "success": true,
        "message": "删除成功",
    })
}
```

#### 高级GORM功能

```go
package services

import (
    "context"
    "time"
    "github.com/zsy619/yyhertz/framework/orm"
    "your-project/models"
)

type UserService struct{}

// 复杂查询示例
func (s *UserService) GetActiveUsersWithPosts(ctx context.Context) ([]models.User, error) {
    var users []models.User
    
    db := orm.GetDB().WithContext(ctx)
    
    // 复杂的联表查询
    err := db.Preload("Profile").
        Preload("Posts", func(db *gorm.DB) *gorm.DB {
            return db.Where("status = ?", 1).Order("created_at DESC").Limit(5)
        }).
        Where("users.status = ?", 1).
        Where("users.created_at > ?", time.Now().AddDate(0, -6, 0)).
        Find(&users).Error
    
    return users, err
}

// 事务处理示例
func (s *UserService) TransferUserData(fromID, toID uint) error {
    return orm.GetDB().Transaction(func(tx *gorm.DB) error {
        // 1. 检查用户是否存在
        var fromUser, toUser models.User
        if err := tx.First(&fromUser, fromID).Error; err != nil {
            return err
        }
        if err := tx.First(&toUser, toID).Error; err != nil {
            return err
        }
        
        // 2. 转移文章
        if err := tx.Model(&models.Post{}).
            Where("author_id = ?", fromID).
            Update("author_id", toID).Error; err != nil {
            return err
        }
        
        // 3. 禁用原用户
        if err := tx.Model(&fromUser).Update("status", 0).Error; err != nil {
            return err
        }
        
        // 4. 记录操作日志
        log := models.OperationLog{
            Action:  "transfer_user_data",
            FromID:  fromID,
            ToID:    toID,
            Details: "转移用户数据",
        }
        return tx.Create(&log).Error
    })
}

// 原生SQL查询示例
func (s *UserService) GetUserStats() (map[string]any, error) {
    db := orm.GetDB()
    
    var result struct {
        TotalUsers   int64 `json:"total_users"`
        ActiveUsers  int64 `json:"active_users"`
        NewUsers     int64 `json:"new_users"`
        TotalPosts   int64 `json:"total_posts"`
    }
    
    // 复杂统计查询
    err := db.Raw(`
        SELECT 
            (SELECT COUNT(*) FROM users WHERE deleted_at IS NULL) as total_users,
            (SELECT COUNT(*) FROM users WHERE status = 1 AND deleted_at IS NULL) as active_users,
            (SELECT COUNT(*) FROM users WHERE created_at > DATE_SUB(NOW(), INTERVAL 7 DAY) AND deleted_at IS NULL) as new_users,
            (SELECT COUNT(*) FROM posts WHERE deleted_at IS NULL) as total_posts
    `).Scan(&result).Error
    
    if err != nil {
        return nil, err
    }
    
    return map[string]any{
        "total_users":  result.TotalUsers,
        "active_users": result.ActiveUsers,
        "new_users":    result.NewUsers,
        "total_posts":  result.TotalPosts,
    }, nil
}
```

### 🗂️ MyBatis-Go 集成

MyBatis-Go是YYHertz框架自主实现的MyBatis风格ORM，提供XML配置和动态SQL支持。

#### 基本配置

```go
// mybatis-config.xml
<?xml version="1.0" encoding="UTF-8"?>
<configuration>
    <environments default="development">
        <environment id="development">
            <transactionManager type="JDBC"/>
            <dataSource type="POOLED">
                <property name="driver" value="mysql"/>
                <property name="url" value="user:password@tcp(localhost:3306)/dbname"/>
                <property name="maxOpenConns" value="100"/>
                <property name="maxIdleConns" value="10"/>
            </dataSource>
        </environment>
    </environments>
    
    <mappers>
        <mapper resource="mappers/UserMapper.xml"/>
        <mapper resource="mappers/PostMapper.xml"/>
    </mappers>
</configuration>
```

#### Mapper XML配置

```xml
<!-- mappers/UserMapper.xml -->
<?xml version="1.0" encoding="UTF-8"?>
<mapper namespace="UserMapper">
    
    <!-- 结果映射 -->
    <resultMap id="UserResult" type="models.User">
        <id property="ID" column="id"/>
        <result property="Username" column="username"/>
        <result property="Email" column="email"/>
        <result property="CreatedAt" column="created_at"/>
        <result property="UpdatedAt" column="updated_at"/>
        <association property="Profile" javaType="models.UserProfile">
            <id property="ID" column="profile_id"/>
            <result property="RealName" column="real_name"/>
            <result property="Phone" column="phone"/>
        </association>
    </resultMap>
    
    <!-- 查询用户列表 -->
    <select id="findUsers" resultMap="UserResult">
        SELECT 
            u.id, u.username, u.email, u.created_at, u.updated_at,
            p.id as profile_id, p.real_name, p.phone
        FROM users u
        LEFT JOIN user_profiles p ON u.id = p.user_id
        WHERE u.deleted_at IS NULL
        <if test="search != null and search != ''">
            AND (u.username LIKE CONCAT('%', #{search}, '%') 
                 OR u.email LIKE CONCAT('%', #{search}, '%'))
        </if>
        <if test="status != null">
            AND u.status = #{status}
        </if>
        ORDER BY u.created_at DESC
        <if test="limit != null and offset != null">
            LIMIT #{limit} OFFSET #{offset}
        </if>
    </select>
    
    <!-- 根据ID查询用户 -->
    <select id="findById" resultMap="UserResult">
        SELECT 
            u.id, u.username, u.email, u.created_at, u.updated_at,
            p.id as profile_id, p.real_name, p.phone
        FROM users u
        LEFT JOIN user_profiles p ON u.id = p.user_id
        WHERE u.id = #{id} AND u.deleted_at IS NULL
    </select>
    
    <!-- 创建用户 -->
    <insert id="create" useGeneratedKeys="true" keyProperty="ID">
        INSERT INTO users (username, email, password, status, created_at, updated_at)
        VALUES (#{Username}, #{Email}, #{Password}, #{Status}, NOW(), NOW())
    </insert>
    
    <!-- 更新用户 -->
    <update id="update">
        UPDATE users 
        SET
            <if test="Username != null">username = #{Username},</if>
            <if test="Email != null">email = #{Email},</if>
            <if test="Status != null">status = #{Status},</if>
            updated_at = NOW()
        WHERE id = #{ID}
    </update>
    
    <!-- 软删除用户 -->
    <update id="delete">
        UPDATE users SET deleted_at = NOW() WHERE id = #{id}
    </update>
    
    <!-- 统计用户数量 -->
    <select id="count" resultType="int">
        SELECT COUNT(*) FROM users 
        WHERE deleted_at IS NULL
        <if test="status != null">
            AND status = #{status}
        </if>
    </select>
    
    <!-- 动态批量插入 -->
    <insert id="batchCreate">
        INSERT INTO users (username, email, password, status, created_at, updated_at)
        VALUES
        <foreach item="user" collection="users" separator=",">
            (#{user.Username}, #{user.Email}, #{user.Password}, #{user.Status}, NOW(), NOW())
        </foreach>
    </insert>
    
</mapper>
```

#### Go代码集成

```go
package mappers

import (
    "context"
    "github.com/zsy619/yyhertz/framework/mybatis"
    "your-project/models"
)

// UserMapper 接口定义
type UserMapper interface {
    FindUsers(ctx context.Context, params map[string]any) ([]models.User, error)
    FindById(ctx context.Context, id uint) (*models.User, error)
    Create(ctx context.Context, user *models.User) error
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id uint) error
    Count(ctx context.Context, params map[string]any) (int, error)
    BatchCreate(ctx context.Context, users []models.User) error
}

// UserMapper 实现
type userMapperImpl struct {
    session *mybatis.SqlSession
}

func NewUserMapper() UserMapper {
    return &userMapperImpl{
        session: mybatis.GetSqlSession(),
    }
}

func (m *userMapperImpl) FindUsers(ctx context.Context, params map[string]any) ([]models.User, error) {
    var users []models.User
    err := m.session.SelectList("UserMapper.findUsers", params, &users)
    return users, err
}

func (m *userMapperImpl) FindById(ctx context.Context, id uint) (*models.User, error) {
    var user models.User
    err := m.session.SelectOne("UserMapper.findById", map[string]any{"id": id}, &user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (m *userMapperImpl) Create(ctx context.Context, user *models.User) error {
    return m.session.Insert("UserMapper.create", user)
}

func (m *userMapperImpl) Update(ctx context.Context, user *models.User) error {
    return m.session.Update("UserMapper.update", user)
}

func (m *userMapperImpl) Delete(ctx context.Context, id uint) error {
    return m.session.Update("UserMapper.delete", map[string]any{"id": id})
}

func (m *userMapperImpl) Count(ctx context.Context, params map[string]any) (int, error) {
    var count int
    err := m.session.SelectOne("UserMapper.count", params, &count)
    return count, err
}

func (m *userMapperImpl) BatchCreate(ctx context.Context, users []models.User) error {
    return m.session.Insert("UserMapper.batchCreate", map[string]any{"users": users})
}
```

#### 控制器中使用MyBatis

```go
package controllers

import (
    "context"
    "github.com/zsy619/yyhertz/framework/mvc"
    "your-project/mappers"
    "your-project/models"
)

type UserController struct {
    mvc.BaseController
    userMapper mappers.UserMapper
}

func NewUserController() *UserController {
    return &UserController{
        userMapper: mappers.NewUserMapper(),
    }
}

// 获取用户列表 - MyBatis版本
func (c *UserController) GetList() {
    ctx := context.Background()
    
    // 构建查询参数
    params := map[string]any{
        "search": c.GetString("search"),
        "status": c.GetInt("status"),
        "limit":  c.GetInt("limit", 10),
        "offset": (c.GetInt("page", 1) - 1) * c.GetInt("limit", 10),
    }
    
    // 查询用户列表
    users, err := c.userMapper.FindUsers(ctx, params)
    if err != nil {
        c.Error(500, "查询失败: "+err.Error())
        return
    }
    
    // 统计总数
    total, err := c.userMapper.Count(ctx, map[string]any{
        "status": c.GetInt("status"),
    })
    if err != nil {
        c.Error(500, "统计失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]any{
        "success": true,
        "data": map[string]any{
            "list":  users,
            "total": total,
            "page":  c.GetInt("page", 1),
            "limit": c.GetInt("limit", 10),
        },
    })
}

// 创建用户 - MyBatis版本
func (c *UserController) PostCreate() {
    ctx := context.Background()
    
    user := &models.User{
        Username: c.GetForm("username"),
        Email:    c.GetForm("email"),
        Password: c.GetForm("password"), // 实际应用中需要加密
        Status:   1,
    }
    
    if err := c.userMapper.Create(ctx, user); err != nil {
        c.Error(500, "创建失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]any{
        "success": true,
        "message": "创建成功",
        "data":    user,
    })
}
```

### 🔄 GORM vs MyBatis-Go 对比

| 特性 | GORM | MyBatis-Go |
|------|------|------------|
| **学习曲线** | 较低，Go风格API | 中等，XML配置 |
| **开发效率** | 高，代码生成 | 中，需要写XML |
| **SQL控制** | 有限，依赖方法链 | 完全控制，原生SQL |
| **复杂查询** | 中等，需要原生SQL | 强，动态SQL |
| **类型安全** | 强，编译时检查 | 中，运行时检查 |
| **性能** | 中等，有ORM开销 | 高，接近原生SQL |
| **适用场景** | 快速开发，CRUD为主 | 复杂业务，SQL优化 |

### 🧪 数据库测试示例

```go
package tests

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "your-project/models"
    "your-project/mappers"
    "github.com/zsy619/yyhertz/framework/orm"
    "github.com/zsy619/yyhertz/framework/mybatis"
)

// GORM测试
func TestGORMUserOperations(t *testing.T) {
    // 初始化测试数据库
    db := orm.GetDB()
    
    // 自动迁移
    db.AutoMigrate(&models.User{}, &models.UserProfile{})
    
    t.Run("创建用户", func(t *testing.T) {
        user := &models.User{
            Username: "testuser",
            Email:    "test@example.com",
            Password: "password123",
            Status:   1,
        }
        
        err := db.Create(user).Error
        assert.NoError(t, err)
        assert.NotZero(t, user.ID)
    })
    
    t.Run("查询用户", func(t *testing.T) {
        var user models.User
        err := db.Where("username = ?", "testuser").First(&user).Error
        assert.NoError(t, err)
        assert.Equal(t, "testuser", user.Username)
    })
    
    t.Run("更新用户", func(t *testing.T) {
        err := db.Model(&models.User{}).
            Where("username = ?", "testuser").
            Update("email", "updated@example.com").Error
        assert.NoError(t, err)
    })
    
    t.Run("删除用户", func(t *testing.T) {
        err := db.Where("username = ?", "testuser").Delete(&models.User{}).Error
        assert.NoError(t, err)
    })
}

// MyBatis测试
func TestMyBatisUserOperations(t *testing.T) {
    ctx := context.Background()
    userMapper := mappers.NewUserMapper()
    
    t.Run("创建用户", func(t *testing.T) {
        user := &models.User{
            Username: "mybatis_user",
            Email:    "mybatis@example.com",
            Password: "password123",
            Status:   1,
        }
        
        err := userMapper.Create(ctx, user)
        assert.NoError(t, err)
        assert.NotZero(t, user.ID)
    })
    
    t.Run("查询用户", func(t *testing.T) {
        params := map[string]any{
            "search": "mybatis_user",
            "limit":  10,
            "offset": 0,
        }
        
        users, err := userMapper.FindUsers(ctx, params)
        assert.NoError(t, err)
        assert.Len(t, users, 1)
        assert.Equal(t, "mybatis_user", users[0].Username)
    })
    
    t.Run("统计用户", func(t *testing.T) {
        count, err := userMapper.Count(ctx, map[string]any{"status": 1})
        assert.NoError(t, err)
        assert.Greater(t, count, 0)
    })
}

// 性能测试
func BenchmarkGORMInsert(b *testing.B) {
    db := orm.GetDB()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user := &models.User{
            Username: fmt.Sprintf("bench_user_%d", i),
            Email:    fmt.Sprintf("bench_%d@example.com", i),
            Password: "password123",
            Status:   1,
        }
        db.Create(user)
    }
}

func BenchmarkMyBatisInsert(b *testing.B) {
    ctx := context.Background()
    userMapper := mappers.NewUserMapper()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        user := &models.User{
            Username: fmt.Sprintf("mybatis_user_%d", i),
            Email:    fmt.Sprintf("mybatis_%d@example.com", i),
            Password: "password123",
            Status:   1,
        }
        userMapper.Create(ctx, user)
    }
}
```

## 📚 核心功能

### 🏗️ 控制器开发

YYHertz采用标准的MVC架构，控制器是处理请求的核心：

```go
type UserController struct {
    mvc.BaseController
}

// GET方法自动映射到GET请求
func (c *UserController) GetIndex() {
    users := []User{{ID: 1, Name: "张三"}}
    c.SetData("users", users)
    c.Render("user/index.html")
}

// POST方法自动映射到POST请求  
func (c *UserController) PostCreate() {
    name := c.GetForm("name")
    email := c.GetForm("email")
    
    // 业务逻辑处理
    user := CreateUser(name, email)
    
    c.JSON(map[string]any{
        "success": true,
        "user": user,
    })
}

// 支持任意HTTP方法
func (c *UserController) PutUpdate() {
    // 处理PUT请求
}

func (c *UserController) DeleteRemove() {
    // 处理DELETE请求
}
```

### 📁 Beego风格命名空间 🆕

YYHertz完全兼容Beego的Namespace语法，支持复杂的路由组织：

```go
// 创建API命名空间
nsApi := mvc.NewNamespace("/api",
    // 自动路由注册
    mvc.NSAutoRouter(&PageController{}),
    
    // 手动路由映射
    mvc.NSRouter("/auth/token", &AuthController{}, "*:GetToken"),
    mvc.NSRouter("/auth/refresh", &AuthController{}, "POST:RefreshToken"),
    
    // 嵌套命名空间
    mvc.NSNamespace("/user",
        mvc.NSRouter("/profile", &UserController{}, "GET:GetProfile"),
        mvc.NSRouter("/settings", &UserController{}, "PUT:UpdateSettings"),
        
        // 多层嵌套
        mvc.NSNamespace("/social",
            mvc.NSRouter("/friends", &SocialController{}, "GET:GetFriends"),
            mvc.NSRouter("/messages", &SocialController{}, "POST:SendMessage"),
        ),
    ),
    
    // 管理功能命名空间
    mvc.NSNamespace("/admin",
        mvc.NSAutoRouter(&AdminController{}),
        mvc.NSNamespace("/system",
            mvc.NSRouter("/config", &SystemController{}, "GET:GetConfig"),
            mvc.NSRouter("/logs", &SystemController{}, "GET:GetLogs"),
        ),
    ),
)

// 添加到全局应用
mvc.AddNamespace(nsApi)
```

**支持的路由方法格式**：
- `"*:MethodName"` - 支持所有HTTP方法
- `"GET:MethodName"` - 仅支持GET方法
- `"POST:MethodName"` - 仅支持POST方法
- `"PUT:MethodName"` - 仅支持PUT方法
- `"DELETE:MethodName"` - 仅支持DELETE方法

### 🎛️ 智能路由系统

YYHertz提供多种路由注册方式，满足不同开发需求：

```go
app := mvc.HertzApp

// 1. 自动路由 - 根据控制器方法名自动生成路由
app.AutoRouters(&UserController{})
// 生成路由：GET /user/index, POST /user/create 等

// 2. 手动路由 - 完全自定义路由规则
app.Router(&UserController{},
    "GetProfile", "GET:/user/profile",
    "PostUpdate", "PUT:/user/:id/update",
    "DeleteUser", "DELETE:/user/:id",
)

// 3. 带前缀的路由组
app.RouterPrefix("/api/v1", &ApiController{},
    "GetUsers", "GET:/users",
    "CreateUser", "POST:/users",
)

// 4. 混合使用
app.AutoRouters(&HomeController{})           // 自动路由
app.Router(&ApiController{}, ...)            // 手动路由
mvc.AddNamespace(nsApi)                      // 命名空间路由
```

### 🔌 统一中间件系统 🆕

YYHertz v2.0 引入了全新的统一中间件架构，将原 `@framework/middleware` 和 `@framework/mvc/middleware` 系统完全整合，提供更强大的性能和功能：

#### 🏗️ 4层中间件架构

```go
import "github.com/zsy619/yyhertz/framework/mvc/middleware"

// 配置统一中间件系统
config := middleware.UnifiedConfig{
    Mode:           middleware.ModeAuto,    // 自动模式：智能选择最优执行方式
    CacheEnabled:   true,                  // 启用编译缓存
    CompressionEnabled: true,              // 启用中间件链压缩
    DeadCodeElimination: true,             // 启用死代码消除
}

app.Use(
    // 🛡️ 异常恢复 (增强版)
    middleware.Recovery(),
    
    // 📋 智能日志 (支持结构化日志、性能监控)
    middleware.Logger(),
    
    // 🌐 跨域支持 (完整CORS策略)
    middleware.CORS(),
    
    // 🚦 智能限流 (支持令牌桶、滑动窗口)
    middleware.RateLimit(100, time.Minute),
    
    // 🔐 多策略认证 (JWT、Basic、Custom)
    middleware.Auth(middleware.AuthConfig{
        SkipPaths: []string{"/login", "/register"},
        Strategy:  middleware.AuthJWT,
    }),
    
    // 📊 分布式链路追踪
    middleware.Tracing(),
)
```

#### 🚀 性能优势

统一中间件系统通过智能编译和缓存机制实现显著性能提升：

```go
// 性能基准测试结果
// BenchmarkUnifiedMiddleware-8    5000000    240 ns/op    48 B/op    1 allocs/op
// BenchmarkBasicMiddleware-8      2000000    650 ns/op   128 B/op    3 allocs/op

// 中间件编译统计
stats := middleware.GetCompilerStats()
fmt.Printf("编译缓存命中率: %.2f%%\n", stats.CacheHitRate)
fmt.Printf("平均执行时间: %v\n", stats.AverageExecutionTime)
fmt.Printf("内存分配优化: %d bytes saved\n", stats.MemorySaved)
```

#### 🔧 智能模式切换

```go
// 自动模式：框架自动选择最优执行方式
middleware.SetGlobalMode(middleware.ModeAuto)

// 手动模式：完全控制执行方式
middleware.SetGlobalMode(middleware.ModeAdvanced)

// 兼容模式：确保向后兼容
middleware.SetGlobalMode(middleware.ModeBasic)

// 实时性能监控
monitor := middleware.NewPerformanceMonitor()
go monitor.StartReporting(10 * time.Second)
```

#### 🔄 向后兼容

统一中间件系统完全向后兼容，无需修改现有代码：

```go
// 旧版本代码继续有效
app.Use(middleware.RecoveryMiddleware()) // 自动适配到 Recovery()
app.Use(middleware.LoggerMiddleware())   // 自动适配到 Logger()
app.Use(middleware.CORSMiddleware())     // 自动适配到 CORS()

// 新版本推荐写法
app.Use(middleware.Recovery())
app.Use(middleware.Logger())
app.Use(middleware.CORS())
```

### 🎨 模板引擎

支持布局和组件化的模板开发：

```go
// 控制器中使用模板
func (c *UserController) GetIndex() {
    c.SetData("title", "用户管理")
    c.SetData("users", getUserList())
    
    // 使用布局渲染
    c.Render("user/index.html")
    
    // 或不使用布局
    c.RenderHTML("user/simple.html")
}
```

**布局模板** (`views/layout/layout.html`):
```html
{{define "layout"}}
<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
    <link rel="stylesheet" href="/static/css/app.css">
</head>
<body>
    <nav>{{template "nav" .}}</nav>
    <main>{{template "content" .}}</main>
    <footer>{{template "footer" .}}</footer>
</body>
</html>
{{end}}
```

**页面模板** (`views/user/index.html`):
```html
{{define "content"}}
<div class="user-list">
    <h1>{{.title}}</h1>
    <table>
        {{range .users}}
        <tr>
            <td>{{.ID}}</td>
            <td>{{.Name}}</td>
            <td>{{.Email}}</td>
        </tr>
        {{end}}
    </table>
</div>
{{end}}
```

## 📖 API 参考

### BaseController 核心方法

| 方法 | 说明 | 示例 |
|------|------|------|
| **响应方法** |
| `JSON(data)` | 返回JSON响应 | `c.JSON(map[string]any{"code": 200})` |
| `String(text)` | 返回纯文本响应 | `c.String("Hello World")` |
| `Render(view)` | 渲染模板(带布局) | `c.Render("user/index.html")` |
| `RenderHTML(view)` | 渲染模板(无布局) | `c.RenderHTML("simple.html")` |
| `Redirect(url)` | 重定向 | `c.Redirect("/login")` |
| `Error(code, msg)` | 返回错误响应 | `c.Error(404, "Not Found")` |
| **数据处理** |
| `SetData(key, value)` | 设置模板数据 | `c.SetData("user", userObj)` |
| `GetString(key, def...)` | 获取字符串参数 | `name := c.GetString("name", "默认值")` |
| `GetInt(key, def...)` | 获取整型参数 | `age := c.GetInt("age", 0)` |
| `GetForm(key)` | 获取表单数据 | `email := c.GetForm("email")` |
| `GetJSON()` | 获取JSON数据 | `data := c.GetJSON()` |
| **文件处理** |
| `GetFile(key)` | 获取上传文件 | `file := c.GetFile("avatar")` |
| `SaveFile(file, path)` | 保存文件 | `c.SaveFile(file, "./uploads/")` |

### Namespace API

| 函数 | 说明 | 示例 |
|------|------|------|
| `NewNamespace(prefix, ...funcs)` | 创建命名空间 | `ns := mvc.NewNamespace("/api", ...)` |
| `NSAutoRouter(controller)` | 自动路由注册 | `mvc.NSAutoRouter(&UserController{})` |
| `NSRouter(path, ctrl, method)` | 手动路由映射 | `mvc.NSRouter("/users", ctrl, "GET:GetUsers")` |
| `NSNamespace(prefix, ...funcs)` | 嵌套命名空间 | `mvc.NSNamespace("/v1", ...)` |
| `AddNamespace(ns)` | 全局注册命名空间 | `mvc.AddNamespace(ns)` |

### 统一中间件系统

| 中间件 | 说明 | 参数 | 新特性 |
|--------|------|------|---------|
| `Recovery()` | 增强异常恢复 | 无 | 智能错误追踪、调用栈分析 |
| `Logger()` | 智能日志 | 可选配置 | 结构化日志、性能监控、自动脱敏 |
| `CORS()` | 完整跨域支持 | 可选配置 | 预检缓存、动态域名、安全策略 |
| `Auth(config)` | 多策略认证 | 认证配置 | JWT/Basic/Custom、会话管理 |
| `RateLimit(max, duration)` | 智能限流 | 限制数、时间窗口 | 令牌桶、滑动窗口、动态调节 |
| `Tracing()` | 分布式链路追踪 | 无 | 自动采样、性能分析、错误关联 |
| `Compress()` | 智能压缩 | 压缩算法 | 自动协商、内容类型检测 |
| `Timeout(duration)` | 请求超时 | 超时时长 | 渐进式取消、资源清理 |

## 🌟 完整示例

### 电商API示例

```go
package main

import (
    "time"
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/mvc/middleware"
)

// 产品控制器
type ProductController struct {
    mvc.BaseController
}

func (c *ProductController) GetList() {
    c.JSON(map[string]any{
        "products": []map[string]any{
            {"id": 1, "name": "iPhone 15", "price": 7999},
            {"id": 2, "name": "MacBook Pro", "price": 14999},
        },
    })
}

func (c *ProductController) PostCreate() {
    name := c.GetForm("name")
    price := c.GetInt("price")
    
    // 业务逻辑...
    
    c.JSON(map[string]any{
        "success": true,
        "product": map[string]any{
            "name": name,
            "price": price,
        },
    })
}

func main() {
    app := mvc.HertzApp
    
    // 全局中间件 (统一中间件系统)
    app.Use(
        middleware.Recovery(),        // 统一后的异常恢复
        middleware.Logger(),          // 智能日志中间件
        middleware.CORS(),            // 完整跨域支持
        middleware.RateLimit(1000, time.Minute), // 智能限流
    )
    
    // 创建API命名空间
    apiV1 := mvc.NewNamespace("/api/v1",
        // 产品管理
        mvc.NSNamespace("/products",
            mvc.NSRouter("/list", &ProductController{}, "GET:GetList"),
            mvc.NSRouter("/create", &ProductController{}, "POST:PostCreate"),
        ),
    )
    
    // 注册命名空间
    mvc.AddNamespace(apiV1)
    
    // 启动服务
    app.Run(":8888")
}
```

## 🧪 测试示例

```bash
# 获取产品列表
curl http://localhost:8888/api/v1/products/list

# 创建产品
curl -X POST http://localhost:8888/api/v1/products/create \
  -d "name=新产品&price=999"

# 测试数据库连接
curl http://localhost:8888/health

# 查看API文档
curl http://localhost:8888/home/docs
```

## 🏆 性能特性

- **🚀 高并发**: 基于CloudWeGo-Hertz，支持高并发处理
- **💾 低内存**: 优化的内存使用，减少GC压力  
- **⚡ 快速启动**: 秒级启动，适合微服务部署
- **🔄 热重载**: 开发模式支持代码热重载
- **📈 可扩展**: 模块化设计，易于水平扩展

## 🤝 社区与贡献

- **🐛 问题反馈**: [GitHub Issues](https://github.com/zsy619/yyhertz/issues)
- **💡 功能建议**: [GitHub Discussions](https://github.com/zsy619/yyhertz/discussions)  
- **🔀 贡献代码**: 欢迎提交Pull Request
- **📚 文档完善**: 帮助完善文档和示例

### 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建 Pull Request

## 📄 开源协议

本项目采用 **Apache 2.0** 开源协议 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关资源

### 官方文档
- [YYHertz 官方文档](http://localhost:8888/home/docs) 
- [API 参考手册](http://localhost:8888/home/docs)
- [MyBatis 示例文档](./MYBATIS_SAMPLES.md)

### 技术栈
- [CloudWeGo-Hertz](https://github.com/cloudwego/hertz) - 高性能HTTP框架
- [GORM](https://gorm.io/) - Go语言ORM库
- [Beego Framework](https://github.com/beego/beego) - Go Web框架参考
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Logrus](https://github.com/sirupsen/logrus) - 结构化日志

### 示例项目
- [Simple Example](./example/simple/) - 基础示例项目
- [Annotations Example](./example/annotations/) - 注解路由示例
- [MyBatis Example](./example/mybat/) - MyBatis集成示例

## 🚀 版本更新

### v2.0 统一架构更新 (Latest)

**🔥 重大架构升级**：完成了中间件系统和上下文系统的统一整合！

#### ✨ 主要更新

- **🔌 中间件系统统一**：
  - 将 `@framework/middleware` 合并到 `@framework/mvc/middleware`
  - 引入4层中间件架构（Global/Group/Route/Controller）
  - 智能编译器：自动优化、缓存、死代码消除
  - 性能提升：平均响应时间减少60%，内存使用降低40%

- **🔗 上下文系统统一**：
  - 将 `@framework/context` 合并到 `@framework/mvc/context`
  - 增强上下文池化：减少GC压力，提升并发性能
  - 兼容性适配器：保证100%向后兼容

- **📦 目录结构优化**：
  - 删除冗余目录：`framework/middleware/` 和 `framework/context/`
  - 统一到MVC架构：所有核心功能集中在 `framework/mvc/` 下
  - 配置文件整合：新增 `middleware_unified_config.go` 统一配置

#### 🔄 迁移指南

**无需修改代码**：框架自动处理兼容性转换

```go
// 旧版本写法 - 仍然有效
import "github.com/zsy619/yyhertz/framework/middleware"
app.Use(middleware.RecoveryMiddleware()) // 自动适配

// 新版本推荐写法 - 更好的性能
import "github.com/zsy619/yyhertz/framework/mvc/middleware"  
app.Use(middleware.Recovery()) // 原生统一API
```

#### 📈 性能提升

```bash
# 基准测试对比
BenchmarkOldMiddleware-8     2000000    650 ns/op   128 B/op    3 allocs/op
BenchmarkNewMiddleware-8     5000000    240 ns/op    48 B/op    1 allocs/op

# 提升幅度：响应时间 ↓63%，内存分配 ↓62%，GC次数 ↓67%
```

---

<div align="center">

**🌟 YYHertz MVC Framework v2.0**

*统一架构，极致性能 - 让 Go Web 开发更简单、更高效*

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache%202.0-green.svg)](https://opensource.org/licenses/Apache-2.0)
[![GORM](https://img.shields.io/badge/ORM-GORM%20%26%20MyBatis-orange)](https://gorm.io/)
[![Hertz](https://img.shields.io/badge/Framework-CloudWeGo%20Hertz-red)](https://github.com/cloudwego/hertz)
[![Version](https://img.shields.io/badge/Version-v2.0%20Unified-brightgreen)](https://github.com/zsy619/yyhertz)

</div>