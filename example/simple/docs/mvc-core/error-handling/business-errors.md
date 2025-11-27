# 💼 YYHertz 业务错误码管理

本文档详细介绍如何在YYHertz框架中设计和管理业务错误码，建立统一的错误码体系。

## 🏗️ 业务错误码架构

### 错误码分类体系

YYHertz支持多层次的错误码分类：

```go
// 错误码结构设计
type BusinessError struct {
    Code        int                 `json:"code"`         // 业务错误码
    Message     string              `json:"message"`      // 错误信息
    Description string              `json:"description"`  // 详细描述
    Category    ErrorCategory       `json:"category"`     // 错误分类
    Severity    ErrorSeverity       `json:"severity"`     // 严重级别
    Solutions   []string            `json:"solutions"`    // 解决方案
    Metadata    map[string]any      `json:"metadata"`     // 附加数据
}

// 错误码分段规则
const (
    // 系统级错误 (10000-19999)
    SystemErrorBase = 10000
    
    // 业务级错误 (20000-29999)
    BusinessErrorBase = 20000
    
    // 用户相关错误 (20000-20999)
    UserErrorBase = 20000
    
    // 订单相关错误 (21000-21999)
    OrderErrorBase = 21000
    
    // 支付相关错误 (22000-22999)
    PaymentErrorBase = 22000
    
    // 商品相关错误 (23000-23999)
    ProductErrorBase = 23000
)
```

## 📋 错误码定义规范

### 1. 错误码命名规范

```go
// 用户相关错误码定义
const (
    // 用户不存在
    ErrUserNotFound = 20001
    // 用户已存在
    ErrUserAlreadyExists = 20002
    // 密码错误
    ErrInvalidPassword = 20003
    // 用户被禁用
    ErrUserDisabled = 20004
    // 用户权限不足
    ErrInsufficientPermission = 20005
)

// 订单相关错误码定义
const (
    // 订单不存在
    ErrOrderNotFound = 21001
    // 订单状态无效
    ErrInvalidOrderStatus = 21002
    // 库存不足
    ErrInsufficientStock = 21003
    // 订单已支付
    ErrOrderAlreadyPaid = 21004
    // 订单已取消
    ErrOrderCancelled = 21005
)

// 支付相关错误码定义
const (
    // 支付失败
    ErrPaymentFailed = 22001
    // 支付金额不正确
    ErrInvalidPaymentAmount = 22002
    // 支付方式不支持
    ErrUnsupportedPaymentMethod = 22003
    // 余额不足
    ErrInsufficientBalance = 22004
    // 支付超时
    ErrPaymentTimeout = 22005
)
```

### 2. 错误信息管理

```go
// 错误信息映射表
var BusinessErrorMessages = map[int]BusinessError{
    ErrUserNotFound: {
        Code:        ErrUserNotFound,
        Message:     "用户不存在",
        Description: "系统中未找到指定的用户信息",
        Category:    CategoryValidation,
        Severity:    SeverityMedium,
        Solutions: []string{
            "检查用户ID是否正确",
            "确认用户是否已注册",
            "联系管理员确认用户状态",
        },
    },
    ErrInvalidPassword: {
        Code:        ErrInvalidPassword,
        Message:     "密码错误",
        Description: "提供的密码与用户账户密码不匹配",
        Category:    CategoryAuthentication,
        Severity:    SeverityLow,
        Solutions: []string{
            "检查密码是否输入正确",
            "使用忘记密码功能重置",
            "检查大小写和特殊字符",
        },
    },
    ErrInsufficientStock: {
        Code:        ErrInsufficientStock,
        Message:     "库存不足",
        Description: "商品库存数量不够，无法完成订单",
        Category:    CategoryBusiness,
        Severity:    SeverityHigh,
        Solutions: []string{
            "减少购买数量",
            "等待商品补货",
            "选择其他类似商品",
        },
    },
}
```

## 🛠️ 业务错误处理实现

### 1. 自定义业务错误类型

```go
// CustomBusinessError 自定义业务错误
type CustomBusinessError struct {
    Code        int                 `json:"code"`
    Message     string              `json:"message"`
    Details     map[string]any      `json:"details,omitempty"`
    Timestamp   time.Time           `json:"timestamp"`
    RequestID   string              `json:"request_id,omitempty"`
    UserID      string              `json:"user_id,omitempty"`
}

func (e *CustomBusinessError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewBusinessError 创建业务错误
func NewBusinessError(code int, details ...map[string]any) *CustomBusinessError {
    errorInfo, exists := BusinessErrorMessages[code]
    if !exists {
        return &CustomBusinessError{
            Code:      code,
            Message:   "未知业务错误",
            Timestamp: time.Now(),
        }
    }
    
    err := &CustomBusinessError{
        Code:      errorInfo.Code,
        Message:   errorInfo.Message,
        Timestamp: time.Now(),
    }
    
    if len(details) > 0 {
        err.Details = details[0]
    }
    
    return err
}

// WithContext 添加上下文信息
func (e *CustomBusinessError) WithContext(requestID, userID string) *CustomBusinessError {
    e.RequestID = requestID
    e.UserID = userID
    return e
}

// WithDetails 添加详细信息
func (e *CustomBusinessError) WithDetails(details map[string]any) *CustomBusinessError {
    if e.Details == nil {
        e.Details = make(map[string]any)
    }
    for k, v := range details {
        e.Details[k] = v
    }
    return e
}
```

### 2. 业务错误处理器

```go
// BusinessErrorHandler 业务错误处理器
type BusinessErrorHandler struct {
    logger    *log.Logger
    i18n      *I18nManager
    notifier  NotificationService
}

func NewBusinessErrorHandler() *BusinessErrorHandler {
    return &BusinessErrorHandler{
        logger:   log.New(os.Stdout, "[BusinessError] ", log.LstdFlags),
        i18n:     GetI18nManager(),
        notifier: GetNotificationService(),
    }
}

func (h *BusinessErrorHandler) Handle(ctx *Context, statusCode int, err error) error {
    // 检查是否为业务错误
    businessErr, ok := err.(*CustomBusinessError)
    if !ok {
        return fmt.Errorf("not a business error")
    }
    
    // 记录业务错误日志
    h.logBusinessError(ctx, businessErr)
    
    // 发送通知（如果需要）
    if h.shouldNotify(businessErr) {
        go h.notifyError(businessErr)
    }
    
    // 构建响应
    response := h.buildResponse(ctx, businessErr)
    
    // 返回JSON响应
    ctx.JSON(statusCode, response)
    return nil
}

func (h *BusinessErrorHandler) CanHandle(statusCode int, err error) bool {
    _, ok := err.(*CustomBusinessError)
    return ok
}

func (h *BusinessErrorHandler) Priority() int {
    return 10 // 高优先级
}

func (h *BusinessErrorHandler) logBusinessError(ctx *Context, err *CustomBusinessError) {
    logData := map[string]any{
        "code":       err.Code,
        "message":    err.Message,
        "details":    err.Details,
        "request_id": err.RequestID,
        "user_id":    err.UserID,
        "path":       string(ctx.Path()),
        "method":     string(ctx.Method()),
        "timestamp":  err.Timestamp,
    }
    
    h.logger.Printf("Business Error: %+v", logData)
}

func (h *BusinessErrorHandler) shouldNotify(err *CustomBusinessError) bool {
    // 根据错误码决定是否需要通知
    errorInfo, exists := BusinessErrorMessages[err.Code]
    if !exists {
        return false
    }
    
    // 高严重级别的错误需要通知
    return errorInfo.Severity == SeverityHigh || errorInfo.Severity == SeverityCritical
}

func (h *BusinessErrorHandler) notifyError(err *CustomBusinessError) {
    notification := &ErrorNotification{
        Title:     "业务错误告警",
        Message:   fmt.Sprintf("错误码: %d, 消息: %s", err.Code, err.Message),
        Severity:  "business",
        Timestamp: time.Now(),
        Details:   err.Details,
    }
    
    h.notifier.Send(notification)
}

func (h *BusinessErrorHandler) buildResponse(ctx *Context, err *CustomBusinessError) map[string]any {
    response := map[string]any{
        "success":   false,
        "code":      err.Code,
        "message":   h.getLocalizedMessage(ctx, err),
        "timestamp": err.Timestamp.Unix(),
    }
    
    // 添加请求ID用于追踪
    if err.RequestID != "" {
        response["request_id"] = err.RequestID
    }
    
    // 开发环境显示详细信息
    if isDevelopmentMode() {
        response["details"] = err.Details
        if errorInfo, exists := BusinessErrorMessages[err.Code]; exists {
            response["solutions"] = errorInfo.Solutions
            response["description"] = errorInfo.Description
        }
    }
    
    return response
}

func (h *BusinessErrorHandler) getLocalizedMessage(ctx *Context, err *CustomBusinessError) string {
    // 获取用户语言偏好
    lang := ctx.GetHeader("Accept-Language")
    if lang == "" {
        lang = "zh-CN"
    }
    
    // 返回本地化消息
    return h.i18n.GetMessage(lang, fmt.Sprintf("error.%d", err.Code), err.Message)
}
```

## 🎯 业务场景应用

### 1. 用户管理场景

```go
// UserService 用户服务
type UserService struct {
    userRepo UserRepository
}

func (s *UserService) GetUser(userID string) (*User, error) {
    if userID == "" {
        return nil, NewBusinessError(ErrUserNotFound).
            WithDetails(map[string]any{"reason": "用户ID为空"})
    }
    
    user, err := s.userRepo.FindByID(userID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, NewBusinessError(ErrUserNotFound).
                WithDetails(map[string]any{"user_id": userID})
        }
        return nil, err
    }
    
    if !user.IsActive {
        return nil, NewBusinessError(ErrUserDisabled).
            WithDetails(map[string]any{
                "user_id": userID,
                "disabled_at": user.DisabledAt,
                "reason": user.DisableReason,
            })
    }
    
    return user, nil
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
    // 检查用户是否已存在
    existingUser, _ := s.userRepo.FindByEmail(req.Email)
    if existingUser != nil {
        return nil, NewBusinessError(ErrUserAlreadyExists).
            WithDetails(map[string]any{
                "email": req.Email,
                "existing_user_id": existingUser.ID,
            })
    }
    
    // 验证密码强度
    if !isValidPassword(req.Password) {
        return nil, NewBusinessError(ErrInvalidPassword).
            WithDetails(map[string]any{
                "requirements": []string{
                    "至少8个字符",
                    "包含大小写字母",
                    "包含数字",
                    "包含特殊字符",
                },
            })
    }
    
    // 创建用户
    user, err := s.userRepo.Create(req)
    if err != nil {
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }
    
    return user, nil
}
```

### 2. 订单处理场景

```go
// OrderService 订单服务
type OrderService struct {
    orderRepo   OrderRepository
    productRepo ProductRepository
    stockService StockService
}

func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*Order, error) {
    // 验证商品库存
    for _, item := range req.Items {
        stock, err := s.stockService.GetStock(item.ProductID)
        if err != nil {
            return nil, err
        }
        
        if stock.Available < item.Quantity {
            return nil, NewBusinessError(ErrInsufficientStock).
                WithDetails(map[string]any{
                    "product_id": item.ProductID,
                    "requested": item.Quantity,
                    "available": stock.Available,
                })
        }
    }
    
    // 检查用户权限
    if !s.hasPermission(req.UserID, "create_order") {
        return nil, NewBusinessError(ErrInsufficientPermission).
            WithDetails(map[string]any{
                "user_id": req.UserID,
                "required_permission": "create_order",
            })
    }
    
    // 创建订单
    order, err := s.orderRepo.Create(req)
    if err != nil {
        return nil, fmt.Errorf("创建订单失败: %w", err)
    }
    
    return order, nil
}

func (s *OrderService) PayOrder(orderID, paymentMethod string, amount float64) error {
    order, err := s.orderRepo.FindByID(orderID)
    if err != nil {
        return NewBusinessError(ErrOrderNotFound).
            WithDetails(map[string]any{"order_id": orderID})
    }
    
    if order.Status == "paid" {
        return NewBusinessError(ErrOrderAlreadyPaid).
            WithDetails(map[string]any{
                "order_id": orderID,
                "paid_at": order.PaidAt,
            })
    }
    
    if order.Status == "cancelled" {
        return NewBusinessError(ErrOrderCancelled).
            WithDetails(map[string]any{
                "order_id": orderID,
                "cancelled_at": order.CancelledAt,
            })
    }
    
    if amount != order.TotalAmount {
        return NewBusinessError(ErrInvalidPaymentAmount).
            WithDetails(map[string]any{
                "expected": order.TotalAmount,
                "provided": amount,
                "order_id": orderID,
            })
    }
    
    // 处理支付...
    return nil
}
```

### 3. 控制器层使用

```go
// UserController 用户控制器
type UserController struct {
    mvc.BaseController
    userService *UserService
}

func (c *UserController) GetUser() {
    userID := c.GetString("id")
    requestID := c.GetString("request_id")
    
    user, err := c.userService.GetUser(userID)
    if err != nil {
        // 检查是否为业务错误
        if businessErr, ok := err.(*CustomBusinessError); ok {
            // 添加上下文信息
            businessErr.WithContext(requestID, c.GetCurrentUserID())
            errors.Handle(c.Ctx, 400, businessErr)
            return
        }
        
        // 系统错误
        errors.Handle(c.Ctx, 500, err)
        return
    }
    
    c.JSONSuccess("获取用户成功", user)
}

func (c *UserController) CreateUser() {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        businessErr := NewBusinessError(ErrInvalidPassword).
            WithDetails(map[string]any{"bind_error": err.Error()})
        errors.Handle(c.Ctx, 400, businessErr)
        return
    }
    
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        if businessErr, ok := err.(*CustomBusinessError); ok {
            errors.Handle(c.Ctx, 400, businessErr)
            return
        }
        
        errors.Handle(c.Ctx, 500, err)
        return
    }
    
    c.JSONSuccess("创建用户成功", user)
}
```

## 🌍 国际化支持

### 错误消息国际化

```go
// I18nManager 国际化管理器
type I18nManager struct {
    messages map[string]map[string]string
}

func NewI18nManager() *I18nManager {
    manager := &I18nManager{
        messages: make(map[string]map[string]string),
    }
    
    // 加载中文错误消息
    manager.messages["zh-CN"] = map[string]string{
        "error.20001": "用户不存在",
        "error.20002": "用户已存在",
        "error.20003": "密码错误",
        "error.21001": "订单不存在",
        "error.21003": "库存不足",
        "error.22001": "支付失败",
    }
    
    // 加载英文错误消息
    manager.messages["en-US"] = map[string]string{
        "error.20001": "User not found",
        "error.20002": "User already exists",
        "error.20003": "Invalid password",
        "error.21001": "Order not found",
        "error.21003": "Insufficient stock",
        "error.22001": "Payment failed",
    }
    
    return manager
}

func (m *I18nManager) GetMessage(lang, key, defaultMsg string) string {
    if messages, ok := m.messages[lang]; ok {
        if msg, exists := messages[key]; exists {
            return msg
        }
    }
    
    return defaultMsg
}
```

## 📊 错误码统计和分析

### 错误码监控

```go
// BusinessErrorMonitor 业务错误监控
type BusinessErrorMonitor struct {
    stats map[int]*ErrorStats
    mutex sync.RWMutex
}

type ErrorStats struct {
    Code        int       `json:"code"`
    Count       int64     `json:"count"`
    LastOccur   time.Time `json:"last_occur"`
    FirstOccur  time.Time `json:"first_occur"`
    Frequency   float64   `json:"frequency"` // 每小时发生次数
}

func (m *BusinessErrorMonitor) RecordError(err *CustomBusinessError) {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    stats, exists := m.stats[err.Code]
    if !exists {
        stats = &ErrorStats{
            Code:       err.Code,
            Count:      0,
            FirstOccur: time.Now(),
        }
        m.stats[err.Code] = stats
    }
    
    stats.Count++
    stats.LastOccur = time.Now()
    
    // 计算频率
    duration := stats.LastOccur.Sub(stats.FirstOccur)
    if duration > 0 {
        stats.Frequency = float64(stats.Count) / duration.Hours()
    }
}

func (m *BusinessErrorMonitor) GetTopErrors(limit int) []*ErrorStats {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    
    var stats []*ErrorStats
    for _, stat := range m.stats {
        stats = append(stats, stat)
    }
    
    // 按发生次数排序
    sort.Slice(stats, func(i, j int) bool {
        return stats[i].Count > stats[j].Count
    })
    
    if len(stats) > limit {
        stats = stats[:limit]
    }
    
    return stats
}
```

## 🚀 最佳实践

### 1. 错误码设计原则

```go
// ✅ 推荐的错误码设计
const (
    // 明确的分类和命名
    ErrUserNotFound          = 20001  // 用户不存在
    ErrUserAlreadyExists     = 20002  // 用户已存在
    ErrInvalidUserPassword   = 20003  // 用户密码无效
    
    // 避免过于通用的错误码
    // ❌ 不推荐
    ErrBadRequest = 40000  // 太通用
    ErrError      = 50000  // 无意义
)
```

### 2. 错误处理模式

```go
// ✅ 推荐的错误处理模式
func (s *UserService) UpdateUser(id string, updates *UserUpdates) error {
    // 参数验证
    if id == "" {
        return NewBusinessError(ErrInvalidParameter).
            WithDetails(map[string]any{"field": "id", "reason": "不能为空"})
    }
    
    // 业务逻辑验证
    user, err := s.GetUser(id)
    if err != nil {
        return err // 透传业务错误
    }
    
    // 权限检查
    if !updates.UserID == user.ID {
        return NewBusinessError(ErrInsufficientPermission).
            WithDetails(map[string]any{"required": "owner", "provided": "other"})
    }
    
    // 执行更新
    return s.userRepo.Update(id, updates)
}
```

### 3. 错误响应格式标准化

```go
// 标准化的错误响应格式
type StandardErrorResponse struct {
    Success   bool                   `json:"success"`
    Code      int                    `json:"code"`
    Message   string                 `json:"message"`
    Timestamp int64                  `json:"timestamp"`
    RequestID string                 `json:"request_id,omitempty"`
    Details   map[string]any         `json:"details,omitempty"`
    Solutions []string               `json:"solutions,omitempty"`
}
```

## 📚 相关文档

- **[快速开始](quick-start.md)** - 了解基础错误处理配置
- **[自定义处理器](custom-handlers.md)** - 实现业务错误处理器
- **[错误监控](monitoring.md)** - 监控业务错误指标
- **[最佳实践](best-practices.md)** - 错误处理最佳实践

---

> 💡 **提示**: 良好的业务错误码设计能显著提升系统的可维护性和用户体验。建议按业务模块划分错误码段，保持错误信息的准确性和一致性。