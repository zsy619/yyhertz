# YYHertz 数据库操作基础

<div align="center">

🗄️ **数据库操作完全指南** | 从GORM到MyBatis的双引擎使用

</div>

---

## 📋 目录

- [数据库初始化](#数据库初始化)
- [模型定义](#模型定义)
- [GORM基础操作](#gorm基础操作)
- [MyBatis复杂查询](#mybatis复杂查询)
- [智能ORM选择器](#智能orm选择器)
- [事务管理](#事务管理)
- [数据库迁移](#数据库迁移)
- [性能优化](#性能优化)

---

## 🚀 数据库初始化

### 1. 基础数据库连接

```go
package main

import (
    "github.com/zsy619/yyhertz/framework/orm"
    "github.com/zsy619/yyhertz/framework/config"
    "gorm.io/gorm"
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/driver/postgres"
)

func initDatabase() (*gorm.DB, error) {
    // 方式1: 使用YYHertz的ORM初始化
    db, err := orm.InitDB("mysql", "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local")
    if err != nil {
        return nil, err
    }
    
    return db, nil
}

// 方式2: 手动初始化不同数据库
func initMySQLDatabase() (*gorm.DB, error) {
    dsn := "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        // 预编译语句，提高性能
        PrepareStmt: true,
        // 批量插入大小
        CreateBatchSize: 1000,
    })
    if err != nil {
        return nil, err
    }
    
    return db, nil
}

func initSQLiteDatabase() (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    return db, nil
}

func initPostgreSQLDatabase() (*gorm.DB, error) {
    dsn := "host=localhost user=username password=password dbname=mydb port=5432 sslmode=disable"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### 2. 连接池配置

```go
import (
    "time"
    "database/sql"
)

func setupConnectionPool(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }
    
    // 设置最大空闲连接数
    sqlDB.SetMaxIdleConns(10)
    
    // 设置最大打开连接数
    sqlDB.SetMaxOpenConns(100)
    
    // 设置连接的最大生存时间
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    // 设置连接的最大空闲时间
    sqlDB.SetConnMaxIdleTime(time.Minute * 10)
    
    return nil
}

// 完整的数据库初始化
func InitDatabaseWithConfig() (*gorm.DB, error) {
    // 从配置文件读取数据库配置
    driver := config.GetString("db.driver")
    host := config.GetString("db.host")
    port := config.GetInt("db.port")
    database := config.GetString("db.name")
    username := config.GetString("db.user")
    password := config.GetString("db.password")
    
    var dsn string
    var dialector gorm.Dialector
    
    switch driver {
    case "mysql":
        dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            username, password, host, port, database)
        dialector = mysql.Open(dsn)
    case "postgres":
        dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
            host, username, password, database, port)
        dialector = postgres.Open(dsn)
    case "sqlite":
        dialector = sqlite.Open(database)
    default:
        return nil, fmt.Errorf("unsupported database driver: %s", driver)
    }
    
    db, err := gorm.Open(dialector, &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: true, // 使用单数表名
        },
    })
    if err != nil {
        return nil, err
    }
    
    // 配置连接池
    if err := setupConnectionPool(db); err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### 3. 在YYHertz应用中使用

```go
func main() {
    // 初始化数据库
    db, err := InitDatabaseWithConfig()
    if err != nil {
        panic("failed to connect database: " + err.Error())
    }
    
    // 自动迁移
    db.AutoMigrate(&User{}, &Product{}, &Order{})
    
    app := mvc.HertzApp
    
    // 将数据库连接注入到应用中
    app.Use(func(c *mvc.Context) {
        c.Set("db", db)
        c.Next()
    })
    
    app.Run(":8888")
}
```

---

## 📊 模型定义

### 1. 基础模型结构

```go
import (
    "time"
    "gorm.io/gorm"
)

// 基础模型，包含通用字段
type BaseModel struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 用户模型
type User struct {
    BaseModel
    Username     string         `gorm:"size:50;not null;uniqueIndex" json:"username" validate:"required,min=3,max=50"`
    Email        string         `gorm:"size:100;not null;uniqueIndex" json:"email" validate:"required,email"`
    PasswordHash string         `gorm:"size:255;not null" json:"-"`
    FirstName    string         `gorm:"size:50" json:"first_name" validate:"max=50"`
    LastName     string         `gorm:"size:50" json:"last_name" validate:"max=50"`
    Phone        string         `gorm:"size:20" json:"phone" validate:"omitempty,len=11"`
    Avatar       string         `gorm:"size:255" json:"avatar"`
    Status       UserStatus     `gorm:"default:1" json:"status"`
    LastLoginAt  *time.Time     `json:"last_login_at"`
    
    // 关联关系
    Profile *UserProfile `gorm:"foreignKey:UserID" json:"profile,omitempty"`
    Orders  []Order      `gorm:"foreignKey:UserID" json:"orders,omitempty"`
    Roles   []Role       `gorm:"many2many:user_roles" json:"roles,omitempty"`
}

// 用户状态枚举
type UserStatus int

const (
    UserStatusInactive UserStatus = 0
    UserStatusActive   UserStatus = 1
    UserStatusSuspended UserStatus = 2
)

// 表名指定
func (User) TableName() string {
    return "users"
}

// 模型钩子 - 创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 自动生成默认头像
    if u.Avatar == "" {
        u.Avatar = generateDefaultAvatar(u.Email)
    }
    return nil
}

// 模型钩子 - 更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    // 更新时间戳
    u.UpdatedAt = time.Now()
    return nil
}

// 业务方法
func (u *User) IsActive() bool {
    return u.Status == UserStatusActive
}

func (u *User) GetFullName() string {
    return strings.TrimSpace(u.FirstName + " " + u.LastName)
}

func (u *User) HasRole(roleName string) bool {
    for _, role := range u.Roles {
        if role.Name == roleName {
            return true
        }
    }
    return false
}
```

### 2. 关联模型

```go
// 用户资料模型
type UserProfile struct {
    ID        uint       `gorm:"primarykey" json:"id"`
    UserID    uint       `gorm:"not null;index" json:"user_id"`
    Bio       string     `gorm:"type:text" json:"bio"`
    Website   string     `gorm:"size:255" json:"website"`
    Location  string     `gorm:"size:100" json:"location"`
    Birthday  *time.Time `json:"birthday"`
    Gender    int        `gorm:"default:0" json:"gender"` // 0:未知 1:男 2:女
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
    
    // 关联关系
    User User `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`
}

// 产品模型
type Product struct {
    BaseModel
    Name        string          `gorm:"size:100;not null" json:"name" validate:"required,max=100"`
    Description string          `gorm:"type:text" json:"description"`
    Price       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
    Stock       int             `gorm:"default:0" json:"stock"`
    Status      ProductStatus   `gorm:"default:1" json:"status"`
    CategoryID  uint            `gorm:"not null;index" json:"category_id"`
    
    // 关联关系
    Category Category `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"category,omitempty"`
    Images   []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
    OrderItems []OrderItem  `gorm:"foreignKey:ProductID" json:"-"`
}

type ProductStatus int

const (
    ProductStatusInactive ProductStatus = 0
    ProductStatusActive   ProductStatus = 1
    ProductStatusOutOfStock ProductStatus = 2
)

// 分类模型
type Category struct {
    BaseModel
    Name     string `gorm:"size:50;not null" json:"name"`
    ParentID *uint  `gorm:"index" json:"parent_id"`
    Sort     int    `gorm:"default:0" json:"sort"`
    
    // 自关联
    Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
    Children []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
    Products []Product  `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// 订单模型
type Order struct {
    BaseModel
    OrderNo     string        `gorm:"size:50;not null;uniqueIndex" json:"order_no"`
    UserID      uint          `gorm:"not null;index" json:"user_id"`
    TotalAmount decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_amount"`
    Status      OrderStatus   `gorm:"default:1" json:"status"`
    PaymentAt   *time.Time    `json:"payment_at"`
    ShippedAt   *time.Time    `json:"shipped_at"`
    CompletedAt *time.Time    `json:"completed_at"`
    
    // 关联关系
    User       User        `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"user,omitempty"`
    OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
}

type OrderStatus int

const (
    OrderStatusPending   OrderStatus = 1
    OrderStatusPaid      OrderStatus = 2
    OrderStatusShipped   OrderStatus = 3
    OrderStatusCompleted OrderStatus = 4
    OrderStatusCancelled OrderStatus = 5
)

// 订单项模型
type OrderItem struct {
    ID        uint            `gorm:"primarykey" json:"id"`
    OrderID   uint            `gorm:"not null;index" json:"order_id"`
    ProductID uint            `gorm:"not null;index" json:"product_id"`
    Quantity  int             `gorm:"not null" json:"quantity"`
    Price     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price"`
    
    // 关联关系
    Order   Order   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"order,omitempty"`
    Product Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"product,omitempty"`
}
```

---

## 💾 GORM基础操作

### 1. 创建操作

```go
// 控制器中使用数据库
type UserController struct {
    mvc.BaseController
}

func (c *UserController) PostCreate() {
    db := c.MustGet("db").(*gorm.DB)
    
    // 单条记录创建
    user := User{
        Username: "john_doe",
        Email:    "john@example.com",
        Status:   UserStatusActive,
    }
    
    // 创建记录
    result := db.Create(&user)
    if result.Error != nil {
        c.ErrorJSON(500, "创建用户失败: "+result.Error.Error())
        return
    }
    
    c.JSON(201, user)
}

// 批量创建
func (c *UserController) PostBatchCreate() {
    db := c.MustGet("db").(*gorm.DB)
    
    users := []User{
        {Username: "user1", Email: "user1@example.com"},
        {Username: "user2", Email: "user2@example.com"},
        {Username: "user3", Email: "user3@example.com"},
    }
    
    // 批量创建
    result := db.Create(&users)
    if result.Error != nil {
        c.ErrorJSON(500, "批量创建失败")
        return
    }
    
    c.JSON(201, map[string]any{
        "created_count": result.RowsAffected,
        "users":         users,
    })
}

// 忽略冲突创建
func (c *UserController) PostCreateIgnoreConflict() {
    db := c.MustGet("db").(*gorm.DB)
    
    var req struct {
        Username string `json:"username"`
        Email    string `json:"email"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    user := User{
        Username: req.Username,
        Email:    req.Email,
        Status:   UserStatusActive,
    }
    
    // 忽略冲突创建（如果记录已存在则忽略）
    result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&user)
    
    if result.RowsAffected == 0 {
        c.JSON(200, map[string]any{
            "message": "用户已存在，忽略创建",
        })
    } else {
        c.JSON(201, user)
    }
}
```

### 2. 查询操作

```go
// 基础查询
func (c *UserController) GetList() {
    db := c.MustGet("db").(*gorm.DB)
    
    var users []User
    
    // 查询所有用户
    db.Find(&users)
    
    c.JSON(users)
}

// 条件查询
func (c *UserController) GetActiveUsers() {
    db := c.MustGet("db").(*gorm.DB)
    
    var users []User
    
    // WHERE条件查询
    db.Where("status = ?", UserStatusActive).Find(&users)
    
    c.JSON(users)
}

// 复杂查询
func (c *UserController) GetUsersWithFilters() {
    db := c.MustGet("db").(*gorm.DB)
    
    // 获取查询参数
    keyword := c.Query("keyword")
    status := c.QueryInt("status")
    page := c.QueryInt("page")
    if page <= 0 {
        page = 1
    }
    pageSize := 20
    
    // 构建查询
    query := db.Model(&User{})
    
    // 动态条件
    if keyword != "" {
        query = query.Where("username LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
    }
    
    if status >= 0 {
        query = query.Where("status = ?", status)
    }
    
    // 获取总数
    var total int64
    query.Count(&total)
    
    // 分页查询
    var users []User
    offset := (page - 1) * pageSize
    query.Offset(offset).Limit(pageSize).
        Order("created_at DESC").
        Preload("Profile").  // 预加载关联
        Find(&users)
    
    c.JSON(map[string]any{
        "users": users,
        "pagination": map[string]any{
            "current_page": page,
            "page_size":    pageSize,
            "total":        total,
            "total_pages":  int(math.Ceil(float64(total) / float64(pageSize))),
        },
    })
}

// 单条记录查询
func (c *UserController) GetDetail() {
    db := c.MustGet("db").(*gorm.DB)
    
    id, err := c.ParamInt("id")
    if err != nil {
        c.ErrorJSON(400, "无效的用户ID")
        return
    }
    
    var user User
    
    // 根据主键查询
    result := db.Preload("Profile").Preload("Roles").First(&user, id)
    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        c.ErrorJSON(404, "用户不存在")
        return
    }
    
    if result.Error != nil {
        c.ErrorJSON(500, "查询失败")
        return
    }
    
    c.JSON(user)
}

// 聚合查询
func (c *UserController) GetStats() {
    db := c.MustGet("db").(*gorm.DB)
    
    var stats struct {
        TotalUsers   int64 `json:"total_users"`
        ActiveUsers  int64 `json:"active_users"`
        InactiveUsers int64 `json:"inactive_users"`
    }
    
    // 总用户数
    db.Model(&User{}).Count(&stats.TotalUsers)
    
    // 活跃用户数
    db.Model(&User{}).Where("status = ?", UserStatusActive).Count(&stats.ActiveUsers)
    
    // 非活跃用户数
    db.Model(&User{}).Where("status = ?", UserStatusInactive).Count(&stats.InactiveUsers)
    
    c.JSON(stats)
}
```

### 3. 更新操作

```go
// 更新单条记录
func (c *UserController) PutUpdate() {
    db := c.MustGet("db").(*gorm.DB)
    
    id, err := c.ParamInt("id")
    if err != nil {
        c.ErrorJSON(400, "无效的用户ID")
        return
    }
    
    var req struct {
        FirstName *string     `json:"first_name"`
        LastName  *string     `json:"last_name"`
        Status    *UserStatus `json:"status"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    // 检查用户是否存在
    var user User
    if err := db.First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.ErrorJSON(404, "用户不存在")
        } else {
            c.ErrorJSON(500, "查询失败")
        }
        return
    }
    
    // 构建更新数据
    updateData := make(map[string]interface{})
    if req.FirstName != nil {
        updateData["first_name"] = *req.FirstName
    }
    if req.LastName != nil {
        updateData["last_name"] = *req.LastName
    }
    if req.Status != nil {
        updateData["status"] = *req.Status
    }
    
    if len(updateData) == 0 {
        c.ErrorJSON(400, "没有要更新的字段")
        return
    }
    
    // 执行更新
    result := db.Model(&user).Updates(updateData)
    if result.Error != nil {
        c.ErrorJSON(500, "更新失败")
        return
    }
    
    // 重新查询更新后的数据
    db.Preload("Profile").First(&user, id)
    
    c.JSON(user)
}

// 批量更新
func (c *UserController) PutBatchUpdate() {
    db := c.MustGet("db").(*gorm.DB)
    
    var req struct {
        UserIDs []uint      `json:"user_ids"`
        Status  UserStatus  `json:"status"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    if len(req.UserIDs) == 0 {
        c.ErrorJSON(400, "用户ID列表不能为空")
        return
    }
    
    // 批量更新
    result := db.Model(&User{}).
        Where("id IN ?", req.UserIDs).
        Update("status", req.Status)
    
    if result.Error != nil {
        c.ErrorJSON(500, "批量更新失败")
        return
    }
    
    c.JSON(map[string]any{
        "message":        "批量更新成功",
        "affected_count": result.RowsAffected,
    })
}
```

### 4. 删除操作

```go
// 软删除
func (c *UserController) DeleteSoft() {
    db := c.MustGet("db").(*gorm.DB)
    
    id, err := c.ParamInt("id")
    if err != nil {
        c.ErrorJSON(400, "无效的用户ID")
        return
    }
    
    // 软删除（设置deleted_at字段）
    result := db.Delete(&User{}, id)
    
    if result.Error != nil {
        c.ErrorJSON(500, "删除失败")
        return
    }
    
    if result.RowsAffected == 0 {
        c.ErrorJSON(404, "用户不存在")
        return
    }
    
    c.JSON(map[string]any{
        "message": "删除成功",
    })
}

// 硬删除
func (c *UserController) DeleteHard() {
    db := c.MustGet("db").(*gorm.DB)
    
    id, err := c.ParamInt("id")
    if err != nil {
        c.ErrorJSON(400, "无效的用户ID")
        return
    }
    
    // 硬删除（真正删除记录）
    result := db.Unscoped().Delete(&User{}, id)
    
    if result.Error != nil {
        c.ErrorJSON(500, "删除失败")
        return
    }
    
    if result.RowsAffected == 0 {
        c.ErrorJSON(404, "用户不存在")
        return
    }
    
    c.JSON(map[string]any{
        "message": "永久删除成功",
    })
}

// 批量删除
func (c *UserController) DeleteBatch() {
    db := c.MustGet("db").(*gorm.DB)
    
    var req struct {
        UserIDs []uint `json:"user_ids"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    if len(req.UserIDs) == 0 {
        c.ErrorJSON(400, "用户ID列表不能为空")
        return
    }
    
    // 批量软删除
    result := db.Delete(&User{}, req.UserIDs)
    
    if result.Error != nil {
        c.ErrorJSON(500, "批量删除失败")
        return
    }
    
    c.JSON(map[string]any{
        "message":        "批量删除成功",
        "affected_count": result.RowsAffected,
    })
}
```

---

## 🧠 智能ORM选择器

### 1. 使用智能选择器

```go
import "github.com/zsy619/yyhertz/framework/orm"

func (c *UserController) GetUsersWithSmartSelector() {
    // 创建智能选择器
    selector := orm.NewSmartSelector()
    
    // 简单查询会自动选择GORM引擎
    var users []User
    err := selector.Find(&users, "status = ?", UserStatusActive)
    if err != nil {
        c.ErrorJSON(500, "查询失败")
        return
    }
    
    c.JSON(users)
}

// 复杂统计查询会自动选择MyBatis引擎
func (c *ReportController) GetUserStats() {
    selector := orm.NewSmartSelector()
    
    // 复杂查询参数
    params := map[string]interface{}{
        "startDate": c.Query("start_date"),
        "endDate":   c.Query("end_date"),
        "status":    UserStatusActive,
    }
    
    // 执行复杂查询
    result, err := selector.ExecuteComplexQuery("getUserStatsReport", params)
    if err != nil {
        c.ErrorJSON(500, "统计查询失败")
        return
    }
    
    c.JSON(result)
}
```

### 2. 手动选择引擎

```go
func (c *UserController) GetUsersWithGORM() {
    db := orm.GetDB()
    
    var users []User
    db.Where("status = ?", UserStatusActive).
        Order("created_at DESC").
        Limit(10).
        Find(&users)
    
    c.JSON(users)
}

func (c *ReportController) GetComplexReport() {
    mybatisSession := orm.GetMyBatisSession()
    
    params := map[string]interface{}{
        "year":   c.QueryInt("year"),
        "month":  c.QueryInt("month"),
        "status": "active",
    }
    
    result, err := mybatisSession.SelectList("getMonthlyUserReport", params)
    if err != nil {
        c.ErrorJSON(500, "报表查询失败")
        return
    }
    
    c.JSON(result)
}
```

---

## 🔄 事务管理

### 1. GORM事务

```go
func (c *OrderController) PostCreate() {
    db := c.MustGet("db").(*gorm.DB)
    
    var req struct {
        UserID      uint                 `json:"user_id"`
        OrderItems  []CreateOrderItem    `json:"order_items"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    // 开启事务
    err := db.Transaction(func(tx *gorm.DB) error {
        // 1. 创建订单
        order := Order{
            OrderNo: generateOrderNo(),
            UserID:  req.UserID,
            Status:  OrderStatusPending,
        }
        
        if err := tx.Create(&order).Error; err != nil {
            return err
        }
        
        var totalAmount decimal.Decimal
        
        // 2. 创建订单项并扣减库存
        for _, item := range req.OrderItems {
            // 检查库存
            var product Product
            if err := tx.First(&product, item.ProductID).Error; err != nil {
                return fmt.Errorf("商品不存在: %d", item.ProductID)
            }
            
            if product.Stock < item.Quantity {
                return fmt.Errorf("商品库存不足: %s", product.Name)
            }
            
            // 扣减库存
            if err := tx.Model(&product).Update("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
                return err
            }
            
            // 创建订单项
            orderItem := OrderItem{
                OrderID:   order.ID,
                ProductID: item.ProductID,
                Quantity:  item.Quantity,
                Price:     product.Price,
            }
            
            if err := tx.Create(&orderItem).Error; err != nil {
                return err
            }
            
            totalAmount = totalAmount.Add(product.Price.Mul(decimal.NewFromInt(int64(item.Quantity))))
        }
        
        // 3. 更新订单总金额
        if err := tx.Model(&order).Update("total_amount", totalAmount).Error; err != nil {
            return err
        }
        
        return nil
    })
    
    if err != nil {
        c.ErrorJSON(500, "创建订单失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]any{
        "message": "订单创建成功",
    })
}
```

### 2. 手动事务控制

```go
func (c *UserController) PostTransferPoints() {
    db := c.MustGet("db").(*gorm.DB)
    
    var req struct {
        FromUserID uint `json:"from_user_id"`
        ToUserID   uint `json:"to_user_id"`
        Points     int  `json:"points"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    // 开始事务
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    if tx.Error != nil {
        c.ErrorJSON(500, "开启事务失败")
        return
    }
    
    // 扣减转出用户积分
    result := tx.Model(&User{}).
        Where("id = ? AND points >= ?", req.FromUserID, req.Points).
        Update("points", gorm.Expr("points - ?", req.Points))
    
    if result.Error != nil {
        tx.Rollback()
        c.ErrorJSON(500, "扣减积分失败")
        return
    }
    
    if result.RowsAffected == 0 {
        tx.Rollback()
        c.ErrorJSON(400, "用户不存在或积分不足")
        return
    }
    
    // 增加转入用户积分
    if err := tx.Model(&User{}).
        Where("id = ?", req.ToUserID).
        Update("points", gorm.Expr("points + ?", req.Points)).Error; err != nil {
        tx.Rollback()
        c.ErrorJSON(500, "增加积分失败")
        return
    }
    
    // 提交事务
    if err := tx.Commit().Error; err != nil {
        c.ErrorJSON(500, "提交事务失败")
        return
    }
    
    c.JSON(map[string]any{
        "message": "积分转账成功",
    })
}
```

---

## 🎯 数据库最佳实践

### 1. 查询优化

```go
// ✅ 好的实践
func (c *UserController) GetUsersOptimized() {
    db := c.MustGet("db").(*gorm.DB)
    
    var users []User
    
    // 1. 只查询需要的字段
    db.Select("id, username, email, status, created_at").
        Where("status = ?", UserStatusActive).
        Order("created_at DESC").
        Limit(20).
        Find(&users)
    
    // 2. 避免N+1查询，使用预加载
    db.Preload("Profile").
        Preload("Roles", "status = ?", 1). // 条件预加载
        Find(&users)
    
    c.JSON(users)
}

// ❌ 避免的做法
func (c *UserController) GetUsersUnoptimized() {
    db := c.MustGet("db").(*gorm.DB)
    
    var users []User
    
    // 查询所有字段（包括不需要的）
    db.Find(&users)
    
    // N+1查询问题
    for i := range users {
        db.Where("user_id = ?", users[i].ID).Find(&users[i].Orders)
    }
    
    c.JSON(users)
}
```

### 2. 索引使用

```go
// 模型定义时添加合适的索引
type User struct {
    BaseModel
    Username string `gorm:"size:50;not null;uniqueIndex:idx_username" json:"username"`
    Email    string `gorm:"size:100;not null;uniqueIndex:idx_email" json:"email"`
    Status   int    `gorm:"default:1;index:idx_status" json:"status"`
    
    // 复合索引
    CreatedAt time.Time `gorm:"index:idx_created_status"`
    Status    int       `gorm:"index:idx_created_status"`
}

// 查询时利用索引
func (c *UserController) GetUsersByStatusAndDate() {
    db := c.MustGet("db").(*gorm.DB)
    
    var users []User
    
    // 利用复合索引 idx_created_status
    db.Where("status = ? AND created_at >= ?", 
        UserStatusActive, 
        time.Now().AddDate(0, -1, 0)).
        Find(&users)
    
    c.JSON(users)
}
```

### 3. 连接池监控

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    dbConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "db_connections",
            Help: "Current database connections",
        },
        []string{"status"},
    )
)

func monitorDBConnections(db *gorm.DB) {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for range ticker.C {
            sqlDB, _ := db.DB()
            stats := sqlDB.Stats()
            
            dbConnections.WithLabelValues("open").Set(float64(stats.OpenConnections))
            dbConnections.WithLabelValues("idle").Set(float64(stats.Idle))
            dbConnections.WithLabelValues("in_use").Set(float64(stats.InUse))
            
            // 记录连接池状态
            if stats.OpenConnections > 80 { // 假设最大连接数为100
                log.Printf("Warning: High database connection usage: %d/100", stats.OpenConnections)
            }
        }
    }()
}
```

---

<div align="center">

**🗄️ 掌握YYHertz数据库操作，让数据处理更高效！**

**合理使用双引擎架构，简单用GORM，复杂用MyBatis 🚀**

</div>