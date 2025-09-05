# YYHertz API接口设计指南

<div align="center">

🔌 **构建优雅的RESTful API** | 从设计到实现的完整指南

</div>

---

## 📋 目录

- [API设计原则](#api设计原则)
- [RESTful架构风格](#restful架构风格)
- [请求响应规范](#请求响应规范)
- [版本控制策略](#版本控制策略)
- [认证授权机制](#认证授权机制)
- [错误处理规范](#错误处理规范)
- [API文档生成](#api文档生成)
- [测试与调试](#测试与调试)

---

## 🎯 API设计原则

### 核心设计理念

- **🎯 一致性**: 统一的命名规范和响应格式
- **📝 可读性**: 清晰的URL结构和语义化的资源命名
- **🔄 无状态**: 每个请求独立，不依赖服务器状态
- **🏗️ 分层**: 清晰的业务逻辑分层
- **🛡️ 安全性**: 完善的认证授权机制
- **📊 可监控**: 完整的日志记录和性能监控

### API设计最佳实践

```go
// 优雅的API控制器设计
package api

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/util/validator"
)

// API基础控制器
type APIController struct {
    mvc.BaseController
}

// 统一响应格式
type APIResponse struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    Meta      *Meta       `json:"meta,omitempty"`
    Timestamp int64       `json:"timestamp"`
}

type Meta struct {
    Page       int   `json:"page,omitempty"`
    PageSize   int   `json:"page_size,omitempty"`
    Total      int64 `json:"total,omitempty"`
    TotalPages int   `json:"total_pages,omitempty"`
}

// 成功响应
func (c *APIController) Success(data interface{}) {
    c.JSON(APIResponse{
        Code:      200,
        Message:   "success",
        Data:      data,
        Timestamp: time.Now().Unix(),
    })
}

// 分页响应
func (c *APIController) SuccessWithPagination(data interface{}, meta *Meta) {
    c.JSON(APIResponse{
        Code:      200,
        Message:   "success",
        Data:      data,
        Meta:      meta,
        Timestamp: time.Now().Unix(),
    })
}

// 错误响应
func (c *APIController) Error(code int, message string) {
    c.JSON(APIResponse{
        Code:      code,
        Message:   message,
        Timestamp: time.Now().Unix(),
    })
}

// 验证错误响应
func (c *APIController) ValidationError(errors map[string]string) {
    c.JSON(APIResponse{
        Code:      400,
        Message:   "Validation failed",
        Data:      errors,
        Timestamp: time.Now().Unix(),
    })
}
```

---

## 🌐 RESTful架构风格

### 1. 资源命名规范

```go
// 用户资源API
type UserAPIController struct {
    APIController
}

// GET /api/users - 获取用户列表
func (c *UserAPIController) GetIndex() {
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    keyword := c.GetString("keyword")
    
    // 参数验证
    if pageSize > 100 {
        pageSize = 100
    }
    
    filters := &UserFilters{
        Keyword: keyword,
    }
    
    users, total, err := c.userService.GetUserList(filters, page, pageSize)
    if err != nil {
        c.Error(500, "Internal server error")
        return
    }
    
    // 计算分页信息
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
    
    meta := &Meta{
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: totalPages,
    }
    
    c.SuccessWithPagination(users, meta)
}

// GET /api/users/:id - 获取单个用户
func (c *UserAPIController) GetShow() {
    id, err := c.GetInt("id")
    if err != nil {
        c.Error(400, "Invalid user ID")
        return
    }
    
    user, err := c.userService.GetUserByID(uint(id))
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.Error(404, "User not found")
        } else {
            c.Error(500, "Internal server error")
        }
        return
    }
    
    c.Success(user.ToAPIResponse())
}

// POST /api/users - 创建用户
func (c *UserAPIController) PostStore() {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "Invalid request body")
        return
    }
    
    // 参数验证
    if errors := validator.Validate(req); len(errors) > 0 {
        c.ValidationError(errors)
        return
    }
    
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        if errors.Is(err, ErrUserExists) {
            c.Error(409, "User already exists")
        } else {
            c.Error(500, "Failed to create user")
        }
        return
    }
    
    c.Success(user.ToAPIResponse())
}

// PUT /api/users/:id - 完整更新用户
func (c *UserAPIController) PutUpdate() {
    id, err := c.GetInt("id")
    if err != nil {
        c.Error(400, "Invalid user ID")
        return
    }
    
    var req UpdateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "Invalid request body")
        return
    }
    
    if errors := validator.Validate(req); len(errors) > 0 {
        c.ValidationError(errors)
        return
    }
    
    user, err := c.userService.UpdateUser(uint(id), &req)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.Error(404, "User not found")
        } else {
            c.Error(500, "Failed to update user")
        }
        return
    }
    
    c.Success(user.ToAPIResponse())
}

// PATCH /api/users/:id - 部分更新用户
func (c *UserAPIController) PatchUpdate() {
    id, err := c.GetInt("id")
    if err != nil {
        c.Error(400, "Invalid user ID")
        return
    }
    
    var req PatchUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "Invalid request body")
        return
    }
    
    user, err := c.userService.PatchUser(uint(id), &req)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.Error(404, "User not found")
        } else {
            c.Error(500, "Failed to update user")
        }
        return
    }
    
    c.Success(user.ToAPIResponse())
}

// DELETE /api/users/:id - 删除用户
func (c *UserAPIController) DeleteDestroy() {
    id, err := c.GetInt("id")
    if err != nil {
        c.Error(400, "Invalid user ID")
        return
    }
    
    err = c.userService.DeleteUser(uint(id))
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.Error(404, "User not found")
        } else {
            c.Error(500, "Failed to delete user")
        }
        return
    }
    
    c.Success(nil)
}
```

### 2. 嵌套资源处理

```go
// 用户订单资源 - 嵌套资源示例
type UserOrderAPIController struct {
    APIController
}

// GET /api/users/:userId/orders - 获取用户的订单列表
func (c *UserOrderAPIController) GetIndex() {
    userID, err := c.GetInt("userId")
    if err != nil {
        c.Error(400, "Invalid user ID")
        return
    }
    
    // 验证用户是否存在
    if _, err := c.userService.GetUserByID(uint(userID)); err != nil {
        if errors.Is(err, ErrUserNotFound) {
            c.Error(404, "User not found")
        } else {
            c.Error(500, "Internal server error")
        }
        return
    }
    
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    status := c.GetString("status")
    
    filters := &OrderFilters{
        UserID: uint(userID),
        Status: status,
    }
    
    orders, total, err := c.orderService.GetOrderList(filters, page, pageSize)
    if err != nil {
        c.Error(500, "Failed to get orders")
        return
    }
    
    meta := &Meta{
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
    }
    
    c.SuccessWithPagination(orders, meta)
}

// GET /api/users/:userId/orders/:id - 获取用户的特定订单
func (c *UserOrderAPIController) GetShow() {
    userID, err := c.GetInt("userId")
    if err != nil {
        c.Error(400, "Invalid user ID")
        return
    }
    
    orderID, err := c.GetInt("id")
    if err != nil {
        c.Error(400, "Invalid order ID")
        return
    }
    
    order, err := c.orderService.GetUserOrder(uint(userID), uint(orderID))
    if err != nil {
        if errors.Is(err, ErrOrderNotFound) {
            c.Error(404, "Order not found")
        } else if errors.Is(err, ErrUnauthorized) {
            c.Error(403, "Access denied")
        } else {
            c.Error(500, "Internal server error")
        }
        return
    }
    
    c.Success(order.ToAPIResponse())
}
```

### 3. 批量操作API

```go
// 批量操作API
func (c *UserAPIController) PostBatch() {
    var req BatchOperationRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "Invalid request body")
        return
    }
    
    if len(req.IDs) == 0 {
        c.Error(400, "No IDs provided")
        return
    }
    
    if len(req.IDs) > 100 {
        c.Error(400, "Too many IDs, maximum 100 allowed")
        return
    }
    
    var result *BatchOperationResult
    var err error
    
    switch req.Operation {
    case "activate":
        result, err = c.userService.BatchActivateUsers(req.IDs)
    case "deactivate":
        result, err = c.userService.BatchDeactivateUsers(req.IDs)
    case "delete":
        result, err = c.userService.BatchDeleteUsers(req.IDs)
    default:
        c.Error(400, "Invalid operation")
        return
    }
    
    if err != nil {
        c.Error(500, "Batch operation failed")
        return
    }
    
    c.Success(result)
}

// 批量操作请求结构
type BatchOperationRequest struct {
    Operation string `json:"operation" validate:"required,oneof=activate deactivate delete"`
    IDs       []uint `json:"ids" validate:"required,min=1,max=100"`
}

// 批量操作结果
type BatchOperationResult struct {
    TotalRequested int      `json:"total_requested"`
    Successful     int      `json:"successful"`
    Failed         int      `json:"failed"`
    Errors         []string `json:"errors,omitempty"`
}
```

---

## 📄 请求响应规范

### 1. 请求格式规范

```go
// 通用请求结构
type BaseRequest struct {
    // 请求ID，用于追踪
    RequestID string `json:"request_id,omitempty" header:"X-Request-ID"`
    
    // 客户端信息
    ClientVersion string `json:"client_version,omitempty" header:"X-Client-Version"`
    UserAgent     string `json:"user_agent,omitempty" header:"User-Agent"`
    
    // 时间戳（防重放攻击）
    Timestamp int64 `json:"timestamp,omitempty"`
}

// 创建用户请求
type CreateUserRequest struct {
    BaseRequest
    Username  string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=8"`
    FirstName string `json:"first_name" validate:"max=50"`
    LastName  string `json:"last_name" validate:"max=50"`
    Phone     string `json:"phone,omitempty" validate:"omitempty,len=11,numeric"`
}

// 更新用户请求（PATCH）
type PatchUserRequest struct {
    BaseRequest
    FirstName *string `json:"first_name,omitempty" validate:"omitempty,max=50"`
    LastName  *string `json:"last_name,omitempty" validate:"omitempty,max=50"`
    Phone     *string `json:"phone,omitempty" validate:"omitempty,len=11,numeric"`
    Status    *int    `json:"status,omitempty" validate:"omitempty,oneof=0 1"`
}

// 查询用户请求
type QueryUsersRequest struct {
    BaseRequest
    Page     int    `form:"page" validate:"min=1"`
    PageSize int    `form:"page_size" validate:"min=1,max=100"`
    Keyword  string `form:"keyword" validate:"max=100"`
    Status   int    `form:"status" validate:"oneof=0 1"`
    Role     string `form:"role" validate:"max=50"`
    SortBy   string `form:"sort_by" validate:"oneof=id username email created_at"`
    SortDir  string `form:"sort_dir" validate:"oneof=asc desc"`
}
```

### 2. 响应格式规范

```go
// 用户API响应格式
type UserAPIResponse struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    FirstName string    `json:"first_name"`
    LastName  string    `json:"last_name"`
    Phone     string    `json:"phone,omitempty"`
    Avatar    string    `json:"avatar"`
    Status    int       `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    
    // 关联数据（可选）
    Profile *UserProfileResponse `json:"profile,omitempty"`
    Roles   []RoleResponse       `json:"roles,omitempty"`
}

// 扩展响应（包含统计信息）
type UserDetailResponse struct {
    UserAPIResponse
    Statistics *UserStatistics `json:"statistics,omitempty"`
}

type UserStatistics struct {
    OrderCount    int     `json:"order_count"`
    TotalSpent    float64 `json:"total_spent"`
    LastOrderDate *time.Time `json:"last_order_date"`
    LoginCount    int     `json:"login_count"`
    LastLoginIP   string  `json:"last_login_ip"`
}

// 模型转换方法
func (u *User) ToAPIResponse() *UserAPIResponse {
    return &UserAPIResponse{
        ID:        u.ID,
        Username:  u.Username,
        Email:     u.Email,
        FirstName: u.FirstName,
        LastName:  u.LastName,
        Phone:     u.Phone,
        Avatar:    u.Avatar,
        Status:    u.Status,
        CreatedAt: u.CreatedAt,
        UpdatedAt: u.UpdatedAt,
    }
}

func (u *User) ToDetailResponse(stats *UserStatistics) *UserDetailResponse {
    return &UserDetailResponse{
        UserAPIResponse: *u.ToAPIResponse(),
        Statistics:      stats,
    }
}
```

### 3. 错误响应规范

```go
// 错误代码定义
const (
    // 客户端错误 4xx
    ErrCodeBadRequest          = 400001 // 请求参数错误
    ErrCodeUnauthorized        = 401001 // 未认证
    ErrCodeForbidden          = 403001 // 权限不足
    ErrCodeNotFound           = 404001 // 资源不存在
    ErrCodeConflict           = 409001 // 资源冲突
    ErrCodeValidationFailed   = 422001 // 验证失败
    ErrCodeTooManyRequests    = 429001 // 请求过于频繁
    
    // 服务器错误 5xx
    ErrCodeInternalServer     = 500001 // 内部服务器错误
    ErrCodeServiceUnavailable = 503001 // 服务不可用
)

// 错误响应结构
type ErrorResponse struct {
    Code      int                    `json:"code"`
    Message   string                 `json:"message"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Timestamp int64                  `json:"timestamp"`
    RequestID string                 `json:"request_id,omitempty"`
}

// 详细错误处理
func (c *APIController) HandleError(err error, requestID string) {
    var errorResp ErrorResponse
    errorResp.Timestamp = time.Now().Unix()
    errorResp.RequestID = requestID
    
    switch {
    case errors.Is(err, ErrUserNotFound):
        errorResp.Code = ErrCodeNotFound
        errorResp.Message = "User not found"
        c.ctx.JSON(404, errorResp)
        
    case errors.Is(err, ErrUserExists):
        errorResp.Code = ErrCodeConflict
        errorResp.Message = "User already exists"
        errorResp.Details = map[string]interface{}{
            "field": "username",
            "value": "already_taken",
        }
        c.ctx.JSON(409, errorResp)
        
    case errors.Is(err, ErrValidationFailed):
        errorResp.Code = ErrCodeValidationFailed
        errorResp.Message = "Validation failed"
        if validationErr, ok := err.(*ValidationError); ok {
            errorResp.Details = map[string]interface{}{
                "fields": validationErr.Fields,
            }
        }
        c.ctx.JSON(422, errorResp)
        
    default:
        // 记录未知错误到日志
        c.Logger.Error("Unhandled error", "error", err.Error(), "request_id", requestID)
        
        errorResp.Code = ErrCodeInternalServer
        errorResp.Message = "Internal server error"
        c.ctx.JSON(500, errorResp)
    }
}
```

---

## 📈 版本控制策略

### 1. URL路径版本控制

```go
// 版本控制路由设置
func setupAPIRoutes(app *mvc.Application) {
    // API v1
    v1 := app.Group("/api/v1")
    {
        // 用户相关接口
        users := v1.Group("/users")
        users.Use(middleware.AuthRequired())
        {
            users.GET("", userV1Controller.GetIndex)
            users.GET("/:id", userV1Controller.GetShow)
            users.POST("", userV1Controller.PostStore)
            users.PUT("/:id", userV1Controller.PutUpdate)
            users.DELETE("/:id", userV1Controller.DeleteDestroy)
        }
    }
    
    // API v2
    v2 := app.Group("/api/v2")
    {
        // 用户相关接口 - v2版本
        users := v2.Group("/users")
        users.Use(middleware.JWTAuth())
        {
            users.GET("", userV2Controller.GetIndex)
            users.GET("/:id", userV2Controller.GetShow)
            users.POST("", userV2Controller.PostStore)
            users.PUT("/:id", userV2Controller.PutUpdate)
            users.PATCH("/:id", userV2Controller.PatchUpdate) // v2新增PATCH支持
            users.DELETE("/:id", userV2Controller.DeleteDestroy)
            
            // v2新增批量操作
            users.POST("/batch", userV2Controller.PostBatch)
        }
    }
    
    // 默认版本（最新版本）
    api := app.Group("/api")
    {
        api.GET("/version", func(c *mvc.Context) {
            c.JSON(map[string]interface{}{
                "current_version": "v2",
                "supported_versions": []string{"v1", "v2"},
                "deprecated_versions": []string{},
            })
        })
        
        // 重定向到最新版本
        api.Use(func(c *mvc.Context) {
            if c.Request.URL.Path == "/api" || c.Request.URL.Path == "/api/" {
                c.Redirect("/api/v2"+c.Request.URL.Path[4:], 301)
                return
            }
            c.Next()
        })
    }
}
```

### 2. Header版本控制

```go
// 基于Header的版本控制中间件
func APIVersionMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        version := c.GetHeader("API-Version")
        if version == "" {
            version = c.Query("version") // 支持查询参数fallback
        }
        if version == "" {
            version = "v2" // 默认最新版本
        }
        
        // 验证版本号
        supportedVersions := []string{"v1", "v2"}
        isSupported := false
        for _, v := range supportedVersions {
            if v == version {
                isSupported = true
                break
            }
        }
        
        if !isSupported {
            c.JSON(400, map[string]interface{}{
                "error": "Unsupported API version",
                "supported_versions": supportedVersions,
            })
            c.Abort()
            return
        }
        
        c.Set("api_version", version)
        c.Next()
    }
}

// 版本化的控制器
func (c *UserAPIController) GetIndex() {
    version := c.GetString("api_version")
    
    switch version {
    case "v1":
        c.getUsersV1()
    case "v2":
        c.getUsersV2()
    default:
        c.Error(400, "Unsupported API version")
    }
}

func (c *UserAPIController) getUsersV1() {
    // v1版本的实现
    users, err := c.userService.GetUsersV1()
    if err != nil {
        c.Error(500, "Internal server error")
        return
    }
    
    c.Success(users)
}

func (c *UserAPIController) getUsersV2() {
    // v2版本的实现（包含更多字段和功能）
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    
    users, total, err := c.userService.GetUsersV2(page, pageSize)
    if err != nil {
        c.Error(500, "Internal server error")
        return
    }
    
    meta := &Meta{
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: int(math.Ceil(float64(total) / float64(pageSize))),
    }
    
    c.SuccessWithPagination(users, meta)
}
```

### 3. 版本废弃处理

```go
// 版本废弃警告中间件
func DeprecationWarningMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        version := c.GetString("api_version")
        
        deprecatedVersions := map[string]string{
            "v1": "2024-12-31", // v1将在2024年12月31日废弃
        }
        
        if deprecateDate, ok := deprecatedVersions[version]; ok {
            c.Header("Warning", fmt.Sprintf("299 - \"API version %s is deprecated. Support will end on %s. Please upgrade to v2.\"", version, deprecateDate))
            c.Header("Sunset", deprecateDate)
            c.Header("Link", "</api/v2>; rel=\"successor-version\"")
        }
        
        c.Next()
    }
}
```

---

## 🔐 认证授权机制

### 1. JWT认证实现

```go
// JWT认证中间件
func JWTAuthMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        token := extractToken(c)
        if token == "" {
            c.JSON(401, ErrorResponse{
                Code:    ErrCodeUnauthorized,
                Message: "Missing access token",
            })
            c.Abort()
            return
        }
        
        claims, err := validateJWTToken(token)
        if err != nil {
            c.JSON(401, ErrorResponse{
                Code:    ErrCodeUnauthorized,
                Message: "Invalid or expired token",
            })
            c.Abort()
            return
        }
        
        // 设置用户信息到上下文
        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)
        c.Set("token_expires", claims.ExpiresAt)
        
        c.Next()
    }
}

// 提取token
func extractToken(c *mvc.Context) string {
    // 1. 从Header中获取
    authHeader := c.GetHeader("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
        return strings.TrimPrefix(authHeader, "Bearer ")
    }
    
    // 2. 从Query参数获取
    if token := c.Query("access_token"); token != "" {
        return token
    }
    
    // 3. 从Cookie获取
    if token, err := c.Cookie("access_token"); err == nil {
        return token
    }
    
    return ""
}

// JWT Claims结构
type JWTClaims struct {
    UserID    uint   `json:"user_id"`
    Username  string `json:"username"`
    Role      string `json:"role"`
    ExpiresAt int64  `json:"exp"`
    IssuedAt  int64  `json:"iat"`
}

// 验证JWT Token
func validateJWTToken(tokenString string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method")
        }
        return []byte(config.GetString("jwt.secret")), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, fmt.Errorf("invalid token")
}
```

### 2. API Key认证

```go
// API Key认证中间件
func APIKeyAuthMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            apiKey = c.Query("api_key")
        }
        
        if apiKey == "" {
            c.JSON(401, ErrorResponse{
                Code:    ErrCodeUnauthorized,
                Message: "Missing API key",
            })
            c.Abort()
            return
        }
        
        // 验证API Key
        client, err := validateAPIKey(apiKey)
        if err != nil {
            c.JSON(401, ErrorResponse{
                Code:    ErrCodeUnauthorized,
                Message: "Invalid API key",
            })
            c.Abort()
            return
        }
        
        // 检查API Key权限
        if !client.IsActive {
            c.JSON(403, ErrorResponse{
                Code:    ErrCodeForbidden,
                Message: "API key is disabled",
            })
            c.Abort()
            return
        }
        
        // 检查速率限制
        if err := checkRateLimit(client.ID); err != nil {
            c.JSON(429, ErrorResponse{
                Code:    ErrCodeTooManyRequests,
                Message: "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        c.Set("client_id", client.ID)
        c.Set("client_name", client.Name)
        c.Next()
    }
}

// API Client结构
type APIClient struct {
    ID          uint      `json:"id"`
    Name        string    `json:"name"`
    APIKey      string    `json:"api_key"`
    SecretKey   string    `json:"secret_key"`
    IsActive    bool      `json:"is_active"`
    RateLimit   int       `json:"rate_limit"`   // 每小时请求限制
    Permissions []string  `json:"permissions"` // 权限列表
    CreatedAt   time.Time `json:"created_at"`
    ExpiresAt   *time.Time `json:"expires_at"`
}

// 验证API Key
func validateAPIKey(apiKey string) (*APIClient, error) {
    // 从缓存或数据库查找API Key
    client, err := apiKeyService.GetClientByAPIKey(apiKey)
    if err != nil {
        return nil, err
    }
    
    // 检查过期时间
    if client.ExpiresAt != nil && time.Now().After(*client.ExpiresAt) {
        return nil, errors.New("API key expired")
    }
    
    return client, nil
}
```

### 3. OAuth2实现

```go
// OAuth2认证控制器
type OAuthController struct {
    APIController
}

// 授权码获取
func (c *OAuthController) PostAuthorize() {
    var req AuthorizeRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "Invalid request")
        return
    }
    
    // 验证客户端
    client, err := c.oauthService.ValidateClient(req.ClientID, req.ClientSecret)
    if err != nil {
        c.Error(401, "Invalid client credentials")
        return
    }
    
    // 验证重定向URI
    if !client.IsValidRedirectURI(req.RedirectURI) {
        c.Error(400, "Invalid redirect URI")
        return
    }
    
    // 生成授权码
    authCode, err := c.oauthService.GenerateAuthorizationCode(
        client.ID,
        req.UserID,
        req.Scope,
        req.RedirectURI,
    )
    if err != nil {
        c.Error(500, "Failed to generate authorization code")
        return
    }
    
    c.Success(map[string]interface{}{
        "code":         authCode.Code,
        "expires_in":   authCode.ExpiresIn,
        "redirect_uri": req.RedirectURI,
    })
}

// 访问令牌获取
func (c *OAuthController) PostToken() {
    var req TokenRequest
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "Invalid request")
        return
    }
    
    switch req.GrantType {
    case "authorization_code":
        c.handleAuthorizationCodeGrant(&req)
    case "refresh_token":
        c.handleRefreshTokenGrant(&req)
    case "client_credentials":
        c.handleClientCredentialsGrant(&req)
    default:
        c.Error(400, "Unsupported grant type")
    }
}

func (c *OAuthController) handleAuthorizationCodeGrant(req *TokenRequest) {
    // 验证授权码
    authCode, err := c.oauthService.ValidateAuthorizationCode(req.Code)
    if err != nil {
        c.Error(400, "Invalid authorization code")
        return
    }
    
    // 验证客户端
    if authCode.ClientID != req.ClientID {
        c.Error(400, "Client ID mismatch")
        return
    }
    
    // 生成访问令牌
    accessToken, refreshToken, err := c.oauthService.GenerateTokens(
        authCode.ClientID,
        authCode.UserID,
        authCode.Scope,
    )
    if err != nil {
        c.Error(500, "Failed to generate tokens")
        return
    }
    
    c.Success(map[string]interface{}{
        "access_token":  accessToken.Token,
        "refresh_token": refreshToken.Token,
        "token_type":    "Bearer",
        "expires_in":    accessToken.ExpiresIn,
        "scope":         authCode.Scope,
    })
}
```

---

## 📝 API文档生成

### 1. Swagger/OpenAPI注解

```go
// 用户API文档注解
type UserAPIController struct {
    APIController
}

// GetIndex 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，支持关键字搜索和状态筛选
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param keyword query string false "搜索关键字"
// @Param status query int false "用户状态" Enums(0, 1)
// @Success 200 {object} APIResponse{data=[]UserAPIResponse,meta=Meta}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v2/users [get]
func (c *UserAPIController) GetIndex() {
    // 实现代码...
}

// GetShow 获取单个用户
// @Summary 获取用户详情
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} APIResponse{data=UserDetailResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v2/users/{id} [get]
func (c *UserAPIController) GetShow() {
    // 实现代码...
}

// PostStore 创建用户
// @Summary 创建新用户
// @Description 创建一个新的用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "用户信息"
// @Success 201 {object} APIResponse{data=UserAPIResponse}
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security ApiKeyAuth
// @Router /api/v2/users [post]
func (c *UserAPIController) PostStore() {
    // 实现代码...
}
```

### 2. 自动化文档生成

```go
// 文档生成配置
func setupSwagger(app *mvc.Application) {
    // Swagger配置
    swaggerConfig := swagger.Config{
        URL:         "http://localhost:8888/swagger/doc.json",
        DocExpansion: "list",
        Title:       "YYHertz API Documentation",
        Version:     "v2.0",
        Description: "YYHertz框架API接口文档",
        Contact: swagger.Contact{
            Name:  "API Support",
            Email: "api-support@example.com",
            URL:   "https://support.example.com",
        },
        License: swagger.License{
            Name: "MIT",
            URL:  "https://opensource.org/licenses/MIT",
        },
    }
    
    // 注册Swagger中间件
    app.GET("/swagger/*", swagger.WrapHandler(swaggerFiles.Handler, swaggerConfig))
    
    // API文档JSON端点
    app.GET("/swagger/doc.json", func(c *mvc.Context) {
        c.JSON(getSwaggerDoc())
    })
}

// 生成API文档
func generateAPIDocs() {
    // 使用swag工具生成文档
    cmd := exec.Command("swag", "init", 
        "--generalInfo", "main.go",
        "--dir", ".",
        "--output", "./docs",
        "--parseDependency", "true",
        "--parseInternal", "true",
    )
    
    if err := cmd.Run(); err != nil {
        log.Printf("Failed to generate API docs: %v", err)
        return
    }
    
    log.Println("API documentation generated successfully")
}
```

### 3. 交互式API文档

```html
<!-- 自定义Swagger UI模板 -->
<!DOCTYPE html>
<html>
<head>
    <title>YYHertz API Documentation</title>
    <link rel="stylesheet" type="text/css" href="/static/swagger-ui/swagger-ui-bundle.css" />
    <style>
        .swagger-ui .topbar { display: none; }
        .swagger-ui .info .title { color: #3b82f6; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    
    <script src="/static/swagger-ui/swagger-ui-bundle.js"></script>
    <script src="/static/swagger-ui/swagger-ui-standalone-preset.js"></script>
    <script>
    window.onload = function() {
        const ui = SwaggerUIBundle({
            url: '/swagger/doc.json',
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout",
            requestInterceptor: function(request) {
                // 自动添加API Key
                const apiKey = localStorage.getItem('api_key');
                if (apiKey) {
                    request.headers['X-API-Key'] = apiKey;
                }
                return request;
            },
            responseInterceptor: function(response) {
                // 处理响应
                return response;
            }
        });
    };
    </script>
</body>
</html>
```

---

## 🧪 测试与调试

### 1. API单元测试

```go
// API测试示例
func TestUserAPI(t *testing.T) {
    // 设置测试应用
    app := mvc.NewTestApp()
    app.AutoRouters(&UserAPIController{})
    
    // 创建测试数据
    testUser := &User{
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    t.Run("创建用户", func(t *testing.T) {
        body := `{
            "username": "testuser",
            "email": "test@example.com",
            "password": "password123"
        }`
        
        req := httptest.NewRequest("POST", "/api/v2/users", strings.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "test-api-key")
        
        w := httptest.NewRecorder()
        app.ServeHTTP(w, req)
        
        assert.Equal(t, 201, w.Code)
        
        var response APIResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, 200, response.Code)
        assert.NotNil(t, response.Data)
    })
    
    t.Run("获取用户列表", func(t *testing.T) {
        req := httptest.NewRequest("GET", "/api/v2/users?page=1&page_size=10", nil)
        req.Header.Set("X-API-Key", "test-api-key")
        
        w := httptest.NewRecorder()
        app.ServeHTTP(w, req)
        
        assert.Equal(t, 200, w.Code)
        
        var response APIResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, 200, response.Code)
        assert.NotNil(t, response.Meta)
    })
    
    t.Run("参数验证测试", func(t *testing.T) {
        body := `{
            "username": "ab",
            "email": "invalid-email",
            "password": "123"
        }`
        
        req := httptest.NewRequest("POST", "/api/v2/users", strings.NewReader(body))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("X-API-Key", "test-api-key")
        
        w := httptest.NewRecorder()
        app.ServeHTTP(w, req)
        
        assert.Equal(t, 422, w.Code)
        
        var response ErrorResponse
        err := json.Unmarshal(w.Body.Bytes(), &response)
        assert.NoError(t, err)
        assert.Equal(t, ErrCodeValidationFailed, response.Code)
        assert.NotNil(t, response.Details)
    })
}
```

### 2. 性能测试

```go
// API性能测试
func BenchmarkUserAPI(b *testing.B) {
    app := mvc.NewTestApp()
    app.AutoRouters(&UserAPIController{})
    
    req := httptest.NewRequest("GET", "/api/v2/users", nil)
    req.Header.Set("X-API-Key", "test-api-key")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        w := httptest.NewRecorder()
        app.ServeHTTP(w, req)
        
        if w.Code != 200 {
            b.Errorf("Expected status 200, got %d", w.Code)
        }
    }
}

// 并发性能测试
func TestAPIPerformance(t *testing.T) {
    app := mvc.NewTestApp()
    app.AutoRouters(&UserAPIController{})
    
    server := httptest.NewServer(app)
    defer server.Close()
    
    // 并发测试
    concurrency := 100
    requests := 1000
    
    var wg sync.WaitGroup
    results := make(chan int, requests)
    
    startTime := time.Now()
    
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            client := &http.Client{Timeout: 10 * time.Second}
            
            for j := 0; j < requests/concurrency; j++ {
                resp, err := client.Get(server.URL + "/api/v2/users")
                if err != nil {
                    results <- 0
                    continue
                }
                results <- resp.StatusCode
                resp.Body.Close()
            }
        }()
    }
    
    wg.Wait()
    close(results)
    
    duration := time.Since(startTime)
    
    successCount := 0
    totalCount := 0
    for statusCode := range results {
        totalCount++
        if statusCode == 200 {
            successCount++
        }
    }
    
    rps := float64(totalCount) / duration.Seconds()
    successRate := float64(successCount) / float64(totalCount) * 100
    
    t.Logf("总请求数: %d", totalCount)
    t.Logf("成功请求数: %d", successCount)
    t.Logf("成功率: %.2f%%", successRate)
    t.Logf("请求速率: %.2f RPS", rps)
    t.Logf("总耗时: %v", duration)
    
    assert.True(t, successRate > 95, "成功率应该大于95%")
    assert.True(t, rps > 1000, "RPS应该大于1000")
}
```

### 3. API调试工具

```go
// API调试中间件
func APIDebugMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        if !config.GetBool("debug.api") {
            c.Next()
            return
        }
        
        // 记录请求信息
        startTime := time.Now()
        
        // 读取请求体
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }
        
        // 创建响应写入器包装
        respWriter := &debugResponseWriter{
            ResponseWriter: c.Writer,
            body:          bytes.NewBuffer(nil),
        }
        c.Writer = respWriter
        
        c.Next()
        
        duration := time.Since(startTime)
        
        // 输出调试信息
        log.Printf(`
=== API Debug Info ===
Method: %s
Path: %s
Headers: %v
Query: %v
Request Body: %s
Response Status: %d
Response Body: %s
Duration: %v
======================`,
            c.Request.Method,
            c.Request.URL.Path,
            c.Request.Header,
            c.Request.URL.Query(),
            string(requestBody),
            respWriter.statusCode,
            respWriter.body.String(),
            duration,
        )
    }
}

type debugResponseWriter struct {
    http.ResponseWriter
    body       *bytes.Buffer
    statusCode int
}

func (w *debugResponseWriter) WriteHeader(code int) {
    w.statusCode = code
    w.ResponseWriter.WriteHeader(code)
}

func (w *debugResponseWriter) Write(data []byte) (int, error) {
    w.body.Write(data)
    return w.ResponseWriter.Write(data)
}
```

---

<div align="center">

**🔌 设计优雅的API是构建现代应用的基础！**

**遵循RESTful规范，提供清晰的文档和完善的测试 🚀**

</div>