# YYHertz 微服务架构实战教程

<div align="center">

🏗️ **构建企业级微服务系统** | 从单体到微服务的完整实践

</div>

---

## 📋 目录

- [微服务架构概述](#微服务架构概述)
- [服务拆分策略](#服务拆分策略)
- [服务注册发现](#服务注册发现)
- [API网关设计](#api网关设计)
- [服务间通信](#服务间通信)
- [配置中心管理](#配置中心管理)
- [链路追踪监控](#链路追踪监控)
- [故障容错机制](#故障容错机制)
- [部署运维实践](#部署运维实践)
- [完整项目案例](#完整项目案例)

---

## 🎯 微服务架构概述

### 微服务 vs 单体架构

```
单体架构 (Monolith)                    微服务架构 (Microservices)
┌─────────────────────┐                ┌────────┐ ┌────────┐ ┌────────┐
│                     │                │ 用户服务 │ │ 订单服务 │ │ 商品服务 │
│    单体应用           │  ────────────▶ │  User  │ │ Order  │ │Product │
│                     │                │ Service│ │Service │ │Service │
│  ┌─────────────────┐│                └────────┘ └────────┘ └────────┘
│  │  用户模块        ││                     │        │        │
│  │  订单模块        ││                     ▼        ▼        ▼
│  │  商品模块        ││                ┌────────┐ ┌────────┐ ┌────────┐
│  │  支付模块        ││                │用户数据库 │ │订单数据库 │ │商品数据库 │
│  └─────────────────┘│                └────────┘ └────────┘ └────────┘
│                     │
│   ┌─────────────┐   │                ┌────────┐ ┌────────┐
│   │   数据库     │   │                │ 支付服务 │ │ 通知服务 │
│   └─────────────┘   │                │Payment │ │ Notice │
│                     │                │Service │ │Service │
└─────────────────────┘                └────────┘ └────────┘
```

### YYHertz微服务特性

| 特性 | 传统框架 | YYHertz优势 |
|------|----------|-------------|
| **服务发现** | 手动配置 | 自动注册发现，支持多种注册中心 |
| **负载均衡** | 外部LB | 客户端负载均衡，多种算法 |
| **熔断降级** | 第三方组件 | 内置熔断器，自适应降级 |
| **链路追踪** | 复杂集成 | 开箱即用，自动埋点 |
| **配置管理** | 静态配置 | 动态配置，热更新 |
| **API网关** | 独立部署 | 集成网关，统一入口 |

---

## 📦 服务拆分策略

### 1. 领域驱动拆分

```go
// 用户领域服务
package user

import (
    "github.com/zsy619/yyhertz/framework/microservice"
)

// UserService 用户服务
type UserService struct {
    microservice.BaseService
    repo *UserRepository
}

// ServiceInfo 服务信息
func (s *UserService) ServiceInfo() *microservice.ServiceInfo {
    return &microservice.ServiceInfo{
        Name:        "user-service",
        Version:     "v1.0.0",
        Description: "用户管理服务",
        Port:        8001,
        Health:      "/health",
        Endpoints: []microservice.Endpoint{
            {Path: "/users", Methods: []string{"GET", "POST"}},
            {Path: "/users/{id}", Methods: []string{"GET", "PUT", "DELETE"}},
            {Path: "/users/{id}/profile", Methods: []string{"GET", "PUT"}},
        },
    }
}

// RegisterRoutes 注册路由
func (s *UserService) RegisterRoutes(app *mvc.Application) {
    userGroup := app.Group("/users")
    userGroup.Use(middleware.Auth()) // 认证中间件
    {
        userGroup.GET("", s.GetUsers)
        userGroup.POST("", s.CreateUser)
        userGroup.GET("/:id", s.GetUser)
        userGroup.PUT("/:id", s.UpdateUser)
        userGroup.DELETE("/:id", s.DeleteUser)
        
        // 用户资料子资源
        userGroup.GET("/:id/profile", s.GetUserProfile)
        userGroup.PUT("/:id/profile", s.UpdateUserProfile)
    }
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(c *mvc.Context) {
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    
    users, total, err := s.repo.FindUsers(page, pageSize)
    if err != nil {
        c.ErrorJSON(500, "查询用户失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "users": users,
        "total": total,
        "page":  page,
        "page_size": pageSize,
    })
}

// CreateUser 创建用户
func (s *UserService) CreateUser(c *mvc.Context) {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数格式错误")
        return
    }
    
    // 参数验证
    if err := validator.Validate(req); err != nil {
        c.ValidationErrorJSON(err)
        return
    }
    
    // 创建用户
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Phone:    req.Phone,
    }
    
    if err := s.repo.CreateUser(user); err != nil {
        if errors.Is(err, ErrUserExists) {
            c.ErrorJSON(409, "用户已存在")
        } else {
            c.ErrorJSON(500, "创建用户失败")
        }
        return
    }
    
    // 发布用户创建事件
    s.PublishEvent(&UserCreatedEvent{
        UserID:   user.ID,
        Username: user.Username,
        Email:    user.Email,
    })
    
    c.JSON(user)
}
```

### 2. 订单服务实现

```go
// 订单领域服务
package order

import (
    "github.com/zsy619/yyhertz/framework/microservice"
    "github.com/zsy619/yyhertz/framework/microservice/client"
)

// OrderService 订单服务
type OrderService struct {
    microservice.BaseService
    repo        *OrderRepository
    userClient  client.ServiceClient  // 用户服务客户端
    productClient client.ServiceClient // 商品服务客户端
    paymentClient client.ServiceClient // 支付服务客户端
}

// NewOrderService 创建订单服务
func NewOrderService() *OrderService {
    service := &OrderService{
        repo: NewOrderRepository(),
    }
    
    // 初始化服务客户端
    service.userClient = client.NewServiceClient("user-service")
    service.productClient = client.NewServiceClient("product-service")
    service.paymentClient = client.NewServiceClient("payment-service")
    
    return service
}

// ServiceInfo 服务信息
func (s *OrderService) ServiceInfo() *microservice.ServiceInfo {
    return &microservice.ServiceInfo{
        Name:        "order-service",
        Version:     "v1.0.0", 
        Description: "订单管理服务",
        Port:        8002,
        Dependencies: []string{"user-service", "product-service", "payment-service"},
    }
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(c *mvc.Context) {
    var req CreateOrderRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数格式错误")
        return
    }
    
    // 1. 验证用户存在性
    user, err := s.validateUser(req.UserID)
    if err != nil {
        c.ErrorJSON(400, "用户验证失败: "+err.Error())
        return
    }
    
    // 2. 验证商品信息和库存
    orderItems, totalAmount, err := s.validateOrderItems(req.Items)
    if err != nil {
        c.ErrorJSON(400, "商品验证失败: "+err.Error())
        return
    }
    
    // 3. 创建订单
    order := &Order{
        UserID:      req.UserID,
        Items:       orderItems,
        TotalAmount: totalAmount,
        Status:      OrderStatusPending,
        CreatedAt:   time.Now(),
    }
    
    if err := s.repo.CreateOrder(order); err != nil {
        c.ErrorJSON(500, "创建订单失败")
        return
    }
    
    // 4. 发起支付
    go s.processPayment(order)
    
    // 5. 发布订单创建事件
    s.PublishEvent(&OrderCreatedEvent{
        OrderID:     order.ID,
        UserID:      order.UserID,
        TotalAmount: order.TotalAmount,
    })
    
    c.JSON(order)
}

// validateUser 验证用户
func (s *OrderService) validateUser(userID uint) (*User, error) {
    var user User
    
    // 调用用户服务验证用户
    resp, err := s.userClient.Get(fmt.Sprintf("/users/%d", userID))
    if err != nil {
        return nil, fmt.Errorf("调用用户服务失败: %w", err)
    }
    
    if resp.StatusCode == 404 {
        return nil, fmt.Errorf("用户不存在")
    }
    
    if err := resp.JSON(&user); err != nil {
        return nil, fmt.Errorf("解析用户信息失败: %w", err)
    }
    
    return &user, nil
}

// validateOrderItems 验证订单商品
func (s *OrderService) validateOrderItems(items []OrderItemRequest) ([]OrderItem, float64, error) {
    var orderItems []OrderItem
    var totalAmount float64
    
    for _, item := range items {
        // 调用商品服务验证商品
        var product Product
        resp, err := s.productClient.Get(fmt.Sprintf("/products/%d", item.ProductID))
        if err != nil {
            return nil, 0, fmt.Errorf("调用商品服务失败: %w", err)
        }
        
        if resp.StatusCode == 404 {
            return nil, 0, fmt.Errorf("商品 %d 不存在", item.ProductID)
        }
        
        if err := resp.JSON(&product); err != nil {
            return nil, 0, fmt.Errorf("解析商品信息失败: %w", err)
        }
        
        // 检查库存
        if product.Stock < item.Quantity {
            return nil, 0, fmt.Errorf("商品 %s 库存不足", product.Name)
        }
        
        // 计算金额
        itemAmount := product.Price * float64(item.Quantity)
        totalAmount += itemAmount
        
        orderItems = append(orderItems, OrderItem{
            ProductID:   item.ProductID,
            ProductName: product.Name,
            Quantity:    item.Quantity,
            Price:       product.Price,
            Amount:      itemAmount,
        })
    }
    
    return orderItems, totalAmount, nil
}

// processPayment 处理支付
func (s *OrderService) processPayment(order *Order) {
    // 调用支付服务创建支付
    paymentReq := map[string]interface{}{
        "order_id": order.ID,
        "amount":   order.TotalAmount,
        "currency": "CNY",
    }
    
    resp, err := s.paymentClient.Post("/payments", paymentReq)
    if err != nil {
        s.handlePaymentFailed(order, err)
        return
    }
    
    var payment Payment
    if err := resp.JSON(&payment); err != nil {
        s.handlePaymentFailed(order, err)
        return
    }
    
    // 更新订单状态
    order.PaymentID = payment.ID
    order.Status = OrderStatusPaid
    s.repo.UpdateOrder(order)
    
    // 发布支付成功事件
    s.PublishEvent(&OrderPaidEvent{
        OrderID:   order.ID,
        PaymentID: payment.ID,
    })
}

// handlePaymentFailed 处理支付失败
func (s *OrderService) handlePaymentFailed(order *Order, err error) {
    order.Status = OrderStatusFailed
    order.FailReason = err.Error()
    s.repo.UpdateOrder(order)
    
    // 发布支付失败事件
    s.PublishEvent(&OrderPaymentFailedEvent{
        OrderID: order.ID,
        Reason:  err.Error(),
    })
}
```

---

## 🔍 服务注册发现

### 1. 服务注册中心

```go
package registry

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/zsy619/yyhertz/framework/microservice"
)

// ServiceRegistry 服务注册接口
type ServiceRegistry interface {
    Register(info *microservice.ServiceInfo) error
    Deregister(serviceName string) error
    Discover(serviceName string) ([]*microservice.ServiceInfo, error)
    Watch(serviceName string) <-chan []*microservice.ServiceInfo
    HealthCheck(serviceName string) error
}

// RedisRegistry Redis服务注册中心
type RedisRegistry struct {
    client       redis.Cmdable
    keyPrefix    string
    ttl          time.Duration
    watchMap     map[string][]chan []*microservice.ServiceInfo
    watchMapLock sync.RWMutex
    ctx          context.Context
    cancel       context.CancelFunc
}

// NewRedisRegistry 创建Redis注册中心
func NewRedisRegistry(client redis.Cmdable) *RedisRegistry {
    ctx, cancel := context.WithCancel(context.Background())
    
    registry := &RedisRegistry{
        client:    client,
        keyPrefix: "microservice:registry:",
        ttl:       30 * time.Second, // 30秒TTL
        watchMap:  make(map[string][]chan []*microservice.ServiceInfo),
        ctx:       ctx,
        cancel:    cancel,
    }
    
    // 启动健康检查
    go registry.startHealthCheck()
    
    return registry
}

// Register 注册服务
func (r *RedisRegistry) Register(info *microservice.ServiceInfo) error {
    key := r.buildServiceKey(info.Name, info.InstanceID)
    
    data, err := json.Marshal(info)
    if err != nil {
        return fmt.Errorf("序列化服务信息失败: %w", err)
    }
    
    // 设置服务信息，带TTL
    if err := r.client.Set(r.ctx, key, data, r.ttl).Err(); err != nil {
        return fmt.Errorf("注册服务失败: %w", err)
    }
    
    // 添加到服务列表
    listKey := r.buildServiceListKey(info.Name)
    r.client.SAdd(r.ctx, listKey, info.InstanceID)
    r.client.Expire(r.ctx, listKey, r.ttl*2)
    
    // 通知观察者
    r.notifyWatchers(info.Name)
    
    return nil
}

// Deregister 注销服务
func (r *RedisRegistry) Deregister(serviceName, instanceID string) error {
    key := r.buildServiceKey(serviceName, instanceID)
    
    // 删除服务信息
    r.client.Del(r.ctx, key)
    
    // 从服务列表中移除
    listKey := r.buildServiceListKey(serviceName)
    r.client.SRem(r.ctx, listKey, instanceID)
    
    // 通知观察者
    r.notifyWatchers(serviceName)
    
    return nil
}

// Discover 发现服务
func (r *RedisRegistry) Discover(serviceName string) ([]*microservice.ServiceInfo, error) {
    listKey := r.buildServiceListKey(serviceName)
    
    // 获取所有实例ID
    instanceIDs, err := r.client.SMembers(r.ctx, listKey).Result()
    if err != nil {
        return nil, fmt.Errorf("获取服务实例列表失败: %w", err)
    }
    
    var services []*microservice.ServiceInfo
    
    for _, instanceID := range instanceIDs {
        key := r.buildServiceKey(serviceName, instanceID)
        data, err := r.client.Get(r.ctx, key).Result()
        if err != nil {
            continue // 忽略已过期的实例
        }
        
        var info microservice.ServiceInfo
        if err := json.Unmarshal([]byte(data), &info); err != nil {
            continue
        }
        
        services = append(services, &info)
    }
    
    return services, nil
}

// Watch 观察服务变化
func (r *RedisRegistry) Watch(serviceName string) <-chan []*microservice.ServiceInfo {
    r.watchMapLock.Lock()
    defer r.watchMapLock.Unlock()
    
    ch := make(chan []*microservice.ServiceInfo, 1)
    
    if r.watchMap[serviceName] == nil {
        r.watchMap[serviceName] = make([]chan []*microservice.ServiceInfo, 0)
    }
    r.watchMap[serviceName] = append(r.watchMap[serviceName], ch)
    
    // 立即发送当前服务列表
    go func() {
        if services, err := r.Discover(serviceName); err == nil {
            select {
            case ch <- services:
            case <-r.ctx.Done():
            }
        }
    }()
    
    return ch
}

// notifyWatchers 通知观察者
func (r *RedisRegistry) notifyWatchers(serviceName string) {
    r.watchMapLock.RLock()
    watchers := r.watchMap[serviceName]
    r.watchMapLock.RUnlock()
    
    if len(watchers) == 0 {
        return
    }
    
    services, err := r.Discover(serviceName)
    if err != nil {
        return
    }
    
    for _, ch := range watchers {
        select {
        case ch <- services:
        default:
            // 非阻塞发送
        }
    }
}

// startHealthCheck 启动健康检查
func (r *RedisRegistry) startHealthCheck() {
    ticker := time.NewTicker(r.ttl / 3) // 每10秒检查一次
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            r.performHealthCheck()
        case <-r.ctx.Done():
            return
        }
    }
}

// performHealthCheck 执行健康检查
func (r *RedisRegistry) performHealthCheck() {
    pattern := r.keyPrefix + "*"
    keys, err := r.client.Keys(r.ctx, pattern).Result()
    if err != nil {
        return
    }
    
    for _, key := range keys {
        // 检查key是否仍然存在
        exists, err := r.client.Exists(r.ctx, key).Result()
        if err != nil || exists == 0 {
            // 服务已下线，通知相关观察者
            serviceName := r.extractServiceNameFromKey(key)
            if serviceName != "" {
                r.notifyWatchers(serviceName)
            }
        }
    }
}

// 辅助方法
func (r *RedisRegistry) buildServiceKey(serviceName, instanceID string) string {
    return fmt.Sprintf("%s%s:%s", r.keyPrefix, serviceName, instanceID)
}

func (r *RedisRegistry) buildServiceListKey(serviceName string) string {
    return fmt.Sprintf("%slist:%s", r.keyPrefix, serviceName)
}

func (r *RedisRegistry) extractServiceNameFromKey(key string) string {
    // 从key中提取服务名
    // 实现略...
    return ""
}
```

### 2. 服务客户端实现

```go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "sync"
    "time"
    
    "github.com/zsy619/yyhertz/framework/microservice/registry"
    "github.com/zsy619/yyhertz/framework/microservice/loadbalancer"
)

// ServiceClient 服务客户端接口
type ServiceClient interface {
    Get(path string) (*Response, error)
    Post(path string, body interface{}) (*Response, error)
    Put(path string, body interface{}) (*Response, error)
    Delete(path string) (*Response, error)
    Request(method, path string, body interface{}) (*Response, error)
}

// HTTPServiceClient HTTP服务客户端
type HTTPServiceClient struct {
    serviceName   string
    registry      registry.ServiceRegistry
    loadBalancer  loadbalancer.LoadBalancer
    httpClient    *http.Client
    circuitBreaker *CircuitBreaker
    
    // 服务实例缓存
    instances     []*microservice.ServiceInfo
    instancesLock sync.RWMutex
}

// NewServiceClient 创建服务客户端
func NewServiceClient(serviceName string) ServiceClient {
    client := &HTTPServiceClient{
        serviceName: serviceName,
        registry:    registry.DefaultRegistry,
        loadBalancer: loadbalancer.NewRoundRobinBalancer(),
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     90 * time.Second,
            },
        },
        circuitBreaker: NewCircuitBreaker(CircuitBreakerConfig{
            MaxRequests:        10,
            Interval:           60 * time.Second,
            Timeout:           30 * time.Second,
            ReadyToTrip:       func(counts Counts) bool {
                return counts.Requests >= 3 && counts.TotalFailures >= 2
            },
        }),
    }
    
    // 启动服务发现
    go client.startServiceDiscovery()
    
    return client
}

// Get GET请求
func (c *HTTPServiceClient) Get(path string) (*Response, error) {
    return c.Request("GET", path, nil)
}

// Post POST请求
func (c *HTTPServiceClient) Post(path string, body interface{}) (*Response, error) {
    return c.Request("POST", path, body)
}

// Put PUT请求
func (c *HTTPServiceClient) Put(path string, body interface{}) (*Response, error) {
    return c.Request("PUT", path, body)
}

// Delete DELETE请求
func (c *HTTPServiceClient) Delete(path string) (*Response, error) {
    return c.Request("DELETE", path, nil)
}

// Request 通用请求方法
func (c *HTTPServiceClient) Request(method, path string, body interface{}) (*Response, error) {
    // 使用熔断器
    result, err := c.circuitBreaker.Execute(func() (interface{}, error) {
        return c.doRequest(method, path, body)
    })
    
    if err != nil {
        return nil, err
    }
    
    return result.(*Response), nil
}

// doRequest 执行实际请求
func (c *HTTPServiceClient) doRequest(method, path string, body interface{}) (*Response, error) {
    // 选择服务实例
    instance, err := c.selectInstance()
    if err != nil {
        return nil, fmt.Errorf("选择服务实例失败: %w", err)
    }
    
    // 构建请求URL
    url := fmt.Sprintf("http://%s:%d%s", instance.Host, instance.Port, path)
    
    // 准备请求体
    var reqBody io.Reader
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("序列化请求体失败: %w", err)
        }
        reqBody = bytes.NewBuffer(jsonData)
    }
    
    // 创建HTTP请求
    req, err := http.NewRequest(method, url, reqBody)
    if err != nil {
        return nil, fmt.Errorf("创建请求失败: %w", err)
    }
    
    // 设置请求头
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", "YYHertz-ServiceClient/1.0")
    
    // 发送请求
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("发送请求失败: %w", err)
    }
    
    return NewResponse(resp), nil
}

// selectInstance 选择服务实例
func (c *HTTPServiceClient) selectInstance() (*microservice.ServiceInfo, error) {
    c.instancesLock.RLock()
    instances := c.instances
    c.instancesLock.RUnlock()
    
    if len(instances) == 0 {
        return nil, fmt.Errorf("没有可用的服务实例")
    }
    
    // 使用负载均衡器选择实例
    instance := c.loadBalancer.Select(instances)
    return instance, nil
}

// startServiceDiscovery 启动服务发现
func (c *HTTPServiceClient) startServiceDiscovery() {
    // 初始发现
    c.updateInstances()
    
    // 监听服务变化
    ch := c.registry.Watch(c.serviceName)
    for instances := range ch {
        c.instancesLock.Lock()
        c.instances = instances
        c.instancesLock.Unlock()
    }
}

// updateInstances 更新服务实例
func (c *HTTPServiceClient) updateInstances() {
    instances, err := c.registry.Discover(c.serviceName)
    if err != nil {
        return
    }
    
    c.instancesLock.Lock()
    c.instances = instances
    c.instancesLock.Unlock()
}

// Response HTTP响应封装
type Response struct {
    *http.Response
    body []byte
}

// NewResponse 创建响应
func NewResponse(resp *http.Response) *Response {
    body, _ := io.ReadAll(resp.Body)
    resp.Body.Close()
    
    return &Response{
        Response: resp,
        body:     body,
    }
}

// JSON 解析JSON响应
func (r *Response) JSON(v interface{}) error {
    return json.Unmarshal(r.body, v)
}

// String 获取字符串响应
func (r *Response) String() string {
    return string(r.body)
}

// Bytes 获取字节响应
func (r *Response) Bytes() []byte {
    return r.body
}
```

---

## 🌐 API网关设计

### 1. 网关核心实现

```go
package gateway

import (
    "net/http"
    "net/http/httputil"
    "net/url"
    "strings"
    "sync"
    
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/framework/microservice/registry"
)

// Gateway API网关
type Gateway struct {
    registry      registry.ServiceRegistry
    routes        map[string]*Route
    middleware    []GatewayMiddleware
    routesLock    sync.RWMutex
}

// Route 路由配置
type Route struct {
    Pattern     string            `json:"pattern"`     // 路径模式
    Service     string            `json:"service"`     // 目标服务
    StripPrefix bool              `json:"strip_prefix"` // 是否去除前缀
    Middleware  []string          `json:"middleware"`   // 中间件
    Headers     map[string]string `json:"headers"`      // 添加的请求头
    Timeout     int               `json:"timeout"`      // 超时时间(秒)
    RateLimit   *RateLimit        `json:"rate_limit"`   // 限流配置
}

// RateLimit 限流配置
type RateLimit struct {
    Requests int `json:"requests"` // 请求数量
    Window   int `json:"window"`   // 时间窗口(秒)
}

// GatewayMiddleware 网关中间件
type GatewayMiddleware func(c *mvc.Context, next func()) error

// NewGateway 创建API网关
func NewGateway() *Gateway {
    gateway := &Gateway{
        registry: registry.DefaultRegistry,
        routes:   make(map[string]*Route),
    }
    
    // 注册默认中间件
    gateway.Use(
        LoggingMiddleware(),        // 日志记录
        CORSMiddleware(),           // CORS处理
        AuthMiddleware(),           // 认证验证
        RateLimitMiddleware(),      // 限流控制
        CircuitBreakerMiddleware(), // 熔断保护
    )
    
    return gateway
}

// Use 添加中间件
func (g *Gateway) Use(middleware ...GatewayMiddleware) {
    g.middleware = append(g.middleware, middleware...)
}

// AddRoute 添加路由
func (g *Gateway) AddRoute(route *Route) {
    g.routesLock.Lock()
    defer g.routesLock.Unlock()
    
    g.routes[route.Pattern] = route
}

// LoadRoutes 从配置加载路由
func (g *Gateway) LoadRoutes(configFile string) error {
    routes, err := loadRoutesFromConfig(configFile)
    if err != nil {
        return err
    }
    
    g.routesLock.Lock()
    defer g.routesLock.Unlock()
    
    for _, route := range routes {
        g.routes[route.Pattern] = route
    }
    
    return nil
}

// ServeHTTP 处理HTTP请求
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    c := mvc.NewContext(w, r)
    
    // 匹配路由
    route := g.matchRoute(r.URL.Path)
    if route == nil {
        c.JSON(404, map[string]string{"error": "Route not found"})
        return
    }
    
    c.Set("route", route)
    
    // 执行中间件链
    g.executeMiddleware(c, 0, func() {
        g.proxyRequest(c, route)
    })
}

// matchRoute 匹配路由
func (g *Gateway) matchRoute(path string) *Route {
    g.routesLock.RLock()
    defer g.routesLock.RUnlock()
    
    // 精确匹配
    if route, exists := g.routes[path]; exists {
        return route
    }
    
    // 前缀匹配
    for pattern, route := range g.routes {
        if strings.HasPrefix(path, pattern) {
            return route
        }
    }
    
    return nil
}

// executeMiddleware 执行中间件链
func (g *Gateway) executeMiddleware(c *mvc.Context, index int, final func()) {
    if index >= len(g.middleware) {
        final()
        return
    }
    
    err := g.middleware[index](c, func() {
        g.executeMiddleware(c, index+1, final)
    })
    
    if err != nil {
        c.JSON(500, map[string]string{"error": err.Error()})
    }
}

// proxyRequest 代理请求到后端服务
func (g *Gateway) proxyRequest(c *mvc.Context, route *Route) {
    // 发现后端服务实例
    services, err := g.registry.Discover(route.Service)
    if err != nil || len(services) == 0 {
        c.JSON(503, map[string]string{"error": "Service unavailable"})
        return
    }
    
    // 选择服务实例（简单轮询）
    service := services[0] // TODO: 实现负载均衡
    
    // 构建目标URL
    targetURL := &url.URL{
        Scheme: "http",
        Host:   fmt.Sprintf("%s:%d", service.Host, service.Port),
        Path:   c.Request.URL.Path,
    }
    
    // 去除前缀
    if route.StripPrefix {
        targetURL.Path = strings.TrimPrefix(targetURL.Path, route.Pattern)
    }
    
    // 创建反向代理
    proxy := httputil.NewSingleHostReverseProxy(targetURL)
    
    // 修改请求
    proxy.ModifyResponse = func(resp *http.Response) error {
        // 添加网关信息头
        resp.Header.Set("X-Gateway", "YYHertz-Gateway")
        return nil
    }
    
    // 错误处理
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        c.JSON(502, map[string]string{"error": "Bad gateway"})
    }
    
    // 添加自定义请求头
    for key, value := range route.Headers {
        c.Request.Header.Set(key, value)
    }
    
    // 执行代理
    proxy.ServeHTTP(c.Writer, c.Request)
}

// 中间件实现
func LoggingMiddleware() GatewayMiddleware {
    return func(c *mvc.Context, next func()) error {
        start := time.Now()
        
        next()
        
        duration := time.Since(start)
        log.Printf("[Gateway] %s %s - %d - %v",
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            duration,
        )
        
        return nil
    }
}

func CORSMiddleware() GatewayMiddleware {
    return func(c *mvc.Context, next func()) error {
        origin := c.GetHeader("Origin")
        if origin != "" {
            c.Header("Access-Control-Allow-Origin", origin)
            c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        }
        
        if c.Request.Method == "OPTIONS" {
            c.Status(204)
            return nil
        }
        
        next()
        return nil
    }
}

func AuthMiddleware() GatewayMiddleware {
    return func(c *mvc.Context, next func()) error {
        route := c.Get("route").(*Route)
        
        // 检查路由是否需要认证
        needAuth := false
        for _, middleware := range route.Middleware {
            if middleware == "auth" {
                needAuth = true
                break
            }
        }
        
        if needAuth {
            token := c.GetHeader("Authorization")
            if token == "" {
                c.JSON(401, map[string]string{"error": "Missing authorization token"})
                return nil
            }
            
            // 验证token（这里简化处理）
            if !isValidToken(token) {
                c.JSON(401, map[string]string{"error": "Invalid token"})
                return nil
            }
        }
        
        next()
        return nil
    }
}

func RateLimitMiddleware() GatewayMiddleware {
    limiter := NewRateLimiter()
    
    return func(c *mvc.Context, next func()) error {
        route := c.Get("route").(*Route)
        
        if route.RateLimit != nil {
            key := fmt.Sprintf("%s:%s", c.ClientIP(), route.Pattern)
            
            allowed, err := limiter.Allow(key, route.RateLimit.Requests, time.Duration(route.RateLimit.Window)*time.Second)
            if err != nil {
                return err
            }
            
            if !allowed {
                c.JSON(429, map[string]string{"error": "Rate limit exceeded"})
                return nil
            }
        }
        
        next()
        return nil
    }
}
```

### 2. 路由配置管理

```yaml
# gateway-routes.yaml
routes:
  - pattern: "/api/v1/users"
    service: "user-service"
    strip_prefix: true
    middleware: ["auth", "rate_limit"]
    timeout: 30
    rate_limit:
      requests: 100
      window: 60
    headers:
      X-Service-Version: "v1"
      
  - pattern: "/api/v1/orders"
    service: "order-service"
    strip_prefix: true
    middleware: ["auth"]
    timeout: 45
    
  - pattern: "/api/v1/products"
    service: "product-service"
    strip_prefix: true
    middleware: ["rate_limit"]
    rate_limit:
      requests: 200
      window: 60
      
  - pattern: "/api/v1/payments"
    service: "payment-service"
    strip_prefix: true
    middleware: ["auth", "encryption"]
    timeout: 60
```

---

## 💬 服务间通信

### 1. 事件驱动架构

```go
package events

import (
    "encoding/json"
    "fmt"
    "reflect"
    "sync"
    
    "github.com/streadway/amqp"
)

// EventBus 事件总线
type EventBus struct {
    conn       *amqp.Connection
    channel    *amqp.Channel
    handlers   map[string][]EventHandler
    handlersMu sync.RWMutex
}

// EventHandler 事件处理器
type EventHandler interface {
    Handle(event Event) error
}

// Event 事件接口
type Event interface {
    EventType() string
    EventData() interface{}
}

// BaseEvent 基础事件
type BaseEvent struct {
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp int64       `json:"timestamp"`
    Source    string      `json:"source"`
}

func (e *BaseEvent) EventType() string    { return e.Type }
func (e *BaseEvent) EventData() interface{} { return e.Data }

// NewEventBus 创建事件总线
func NewEventBus(amqpURL string) (*EventBus, error) {
    conn, err := amqp.Dial(amqpURL)
    if err != nil {
        return nil, err
    }
    
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    
    // 声明交换机
    err = ch.ExchangeDeclare(
        "microservice.events", // name
        "topic",               // type
        true,                  // durable
        false,                 // auto-deleted
        false,                 // internal
        false,                 // no-wait
        nil,                   // arguments
    )
    if err != nil {
        return nil, err
    }
    
    return &EventBus{
        conn:     conn,
        channel:  ch,
        handlers: make(map[string][]EventHandler),
    }, nil
}

// Publish 发布事件
func (eb *EventBus) Publish(event Event) error {
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    return eb.channel.Publish(
        "microservice.events", // exchange
        event.EventType(),     // routing key
        false,                 // mandatory
        false,                 // immediate
        amqp.Publishing{
            ContentType: "application/json",
            Body:        data,
        },
    )
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(eventType string, handler EventHandler) error {
    eb.handlersMu.Lock()
    defer eb.handlersMu.Unlock()
    
    if eb.handlers[eventType] == nil {
        eb.handlers[eventType] = make([]EventHandler, 0)
        
        // 创建队列
        queueName := fmt.Sprintf("events.%s", eventType)
        queue, err := eb.channel.QueueDeclare(
            queueName, // name
            true,      // durable
            false,     // delete when unused
            false,     // exclusive
            false,     // no-wait
            nil,       // arguments
        )
        if err != nil {
            return err
        }
        
        // 绑定队列到交换机
        err = eb.channel.QueueBind(
            queue.Name,            // queue name
            eventType,             // routing key
            "microservice.events", // exchange
            false,
            nil,
        )
        if err != nil {
            return err
        }
        
        // 启动消费者
        go eb.startConsumer(queue.Name, eventType)
    }
    
    eb.handlers[eventType] = append(eb.handlers[eventType], handler)
    return nil
}

// startConsumer 启动消费者
func (eb *EventBus) startConsumer(queueName, eventType string) {
    msgs, err := eb.channel.Consume(
        queueName, // queue
        "",        // consumer
        false,     // auto-ack
        false,     // exclusive
        false,     // no-local
        false,     // no-wait
        nil,       // args
    )
    if err != nil {
        return
    }
    
    for msg := range msgs {
        // 解析事件
        var event BaseEvent
        if err := json.Unmarshal(msg.Body, &event); err != nil {
            msg.Nack(false, true) // 重新排队
            continue
        }
        
        // 处理事件
        eb.handleEvent(&event)
        
        // 确认消息
        msg.Ack(false)
    }
}

// handleEvent 处理事件
func (eb *EventBus) handleEvent(event Event) {
    eb.handlersMu.RLock()
    handlers := eb.handlers[event.EventType()]
    eb.handlersMu.RUnlock()
    
    for _, handler := range handlers {
        go func(h EventHandler) {
            if err := h.Handle(event); err != nil {
                // 记录错误日志
                fmt.Printf("Event handler error: %v\n", err)
            }
        }(handler)
    }
}
```

### 2. 具体事件实现

```go
// 用户事件
type UserCreatedEvent struct {
    BaseEvent
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Email    string `json:"email"`
}

func NewUserCreatedEvent(userID uint, username, email string) *UserCreatedEvent {
    return &UserCreatedEvent{
        BaseEvent: BaseEvent{
            Type:      "user.created",
            Timestamp: time.Now().Unix(),
            Source:    "user-service",
        },
        UserID:   userID,
        Username: username,
        Email:    email,
    }
}

// 订单事件
type OrderCreatedEvent struct {
    BaseEvent
    OrderID     uint    `json:"order_id"`
    UserID      uint    `json:"user_id"`
    TotalAmount float64 `json:"total_amount"`
}

type OrderPaidEvent struct {
    BaseEvent
    OrderID   uint `json:"order_id"`
    PaymentID uint `json:"payment_id"`
}

// 事件处理器实现
type UserEventHandler struct {
    orderService *OrderService
}

func (h *UserEventHandler) Handle(event Event) error {
    switch e := event.(type) {
    case *UserCreatedEvent:
        return h.handleUserCreated(e)
    }
    return nil
}

func (h *UserEventHandler) handleUserCreated(event *UserCreatedEvent) error {
    // 为新用户创建购物车
    cart := &Cart{
        UserID: event.UserID,
        Items:  []CartItem{},
    }
    
    return h.orderService.CreateCart(cart)
}

// 库存事件处理器
type InventoryEventHandler struct {
    inventoryService *InventoryService
}

func (h *InventoryEventHandler) Handle(event Event) error {
    switch e := event.(type) {
    case *OrderPaidEvent:
        return h.handleOrderPaid(e)
    }
    return nil
}

func (h *InventoryEventHandler) handleOrderPaid(event *OrderPaidEvent) error {
    // 扣减库存
    order, err := h.orderService.GetOrder(event.OrderID)
    if err != nil {
        return err
    }
    
    for _, item := range order.Items {
        err := h.inventoryService.DeductStock(item.ProductID, item.Quantity)
        if err != nil {
            // 库存扣减失败，可能需要补偿操作
            return err
        }
    }
    
    return nil
}
```

---

## ⚙️ 配置中心管理

### 1. 配置中心实现

```go
package config

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"
    
    "github.com/go-redis/redis/v8"
)

// ConfigCenter 配置中心
type ConfigCenter struct {
    client      redis.Cmdable
    keyPrefix   string
    watchers    map[string][]ConfigWatcher
    watchersMu  sync.RWMutex
    ctx         context.Context
    cancel      context.CancelFunc
}

// ConfigWatcher 配置监听器
type ConfigWatcher interface {
    OnConfigChanged(key string, value interface{})
}

// ConfigEntry 配置项
type ConfigEntry struct {
    Key       string      `json:"key"`
    Value     interface{} `json:"value"`
    Version   int64       `json:"version"`
    UpdatedAt time.Time   `json:"updated_at"`
    UpdatedBy string      `json:"updated_by"`
}

// NewConfigCenter 创建配置中心
func NewConfigCenter(client redis.Cmdable) *ConfigCenter {
    ctx, cancel := context.WithCancel(context.Background())
    
    cc := &ConfigCenter{
        client:    client,
        keyPrefix: "config:",
        watchers:  make(map[string][]ConfigWatcher),
        ctx:       ctx,
        cancel:    cancel,
    }
    
    // 启动配置监听
    go cc.startWatching()
    
    return cc
}

// Set 设置配置
func (cc *ConfigCenter) Set(key string, value interface{}, updatedBy string) error {
    entry := &ConfigEntry{
        Key:       key,
        Value:     value,
        Version:   time.Now().Unix(),
        UpdatedAt: time.Now(),
        UpdatedBy: updatedBy,
    }
    
    data, err := json.Marshal(entry)
    if err != nil {
        return err
    }
    
    redisKey := cc.keyPrefix + key
    
    // 设置配置值
    err = cc.client.Set(cc.ctx, redisKey, data, 0).Err()
    if err != nil {
        return err
    }
    
    // 发布配置变更通知
    cc.client.Publish(cc.ctx, "config_changed", key)
    
    return nil
}

// Get 获取配置
func (cc *ConfigCenter) Get(key string) (*ConfigEntry, error) {
    redisKey := cc.keyPrefix + key
    
    data, err := cc.client.Get(cc.ctx, redisKey).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, fmt.Errorf("config key not found: %s", key)
        }
        return nil, err
    }
    
    var entry ConfigEntry
    err = json.Unmarshal([]byte(data), &entry)
    if err != nil {
        return nil, err
    }
    
    return &entry, nil
}

// GetString 获取字符串配置
func (cc *ConfigCenter) GetString(key, defaultValue string) string {
    entry, err := cc.Get(key)
    if err != nil {
        return defaultValue
    }
    
    if str, ok := entry.Value.(string); ok {
        return str
    }
    
    return defaultValue
}

// GetInt 获取整数配置
func (cc *ConfigCenter) GetInt(key string, defaultValue int) int {
    entry, err := cc.Get(key)
    if err != nil {
        return defaultValue
    }
    
    if num, ok := entry.Value.(float64); ok {
        return int(num)
    }
    
    return defaultValue
}

// GetBool 获取布尔配置
func (cc *ConfigCenter) GetBool(key string, defaultValue bool) bool {
    entry, err := cc.Get(key)
    if err != nil {
        return defaultValue
    }
    
    if b, ok := entry.Value.(bool); ok {
        return b
    }
    
    return defaultValue
}

// Watch 监听配置变化
func (cc *ConfigCenter) Watch(key string, watcher ConfigWatcher) {
    cc.watchersMu.Lock()
    defer cc.watchersMu.Unlock()
    
    if cc.watchers[key] == nil {
        cc.watchers[key] = make([]ConfigWatcher, 0)
    }
    cc.watchers[key] = append(cc.watchers[key], watcher)
}

// startWatching 启动配置监听
func (cc *ConfigCenter) startWatching() {
    pubsub := cc.client.Subscribe(cc.ctx, "config_changed")
    defer pubsub.Close()
    
    ch := pubsub.Channel()
    
    for {
        select {
        case msg := <-ch:
            cc.handleConfigChanged(msg.Payload)
        case <-cc.ctx.Done():
            return
        }
    }
}

// handleConfigChanged 处理配置变更
func (cc *ConfigCenter) handleConfigChanged(key string) {
    cc.watchersMu.RLock()
    watchers := cc.watchers[key]
    cc.watchersMu.RUnlock()
    
    if len(watchers) == 0 {
        return
    }
    
    // 获取最新配置值
    entry, err := cc.Get(key)
    if err != nil {
        return
    }
    
    // 通知所有监听器
    for _, watcher := range watchers {
        go watcher.OnConfigChanged(key, entry.Value)
    }
}

// 配置管理器
type ServiceConfigManager struct {
    configCenter *ConfigCenter
    serviceName  string
    configs      map[string]interface{}
    configsMu    sync.RWMutex
}

// NewServiceConfigManager 创建服务配置管理器
func NewServiceConfigManager(configCenter *ConfigCenter, serviceName string) *ServiceConfigManager {
    manager := &ServiceConfigManager{
        configCenter: configCenter,
        serviceName:  serviceName,
        configs:      make(map[string]interface{}),
    }
    
    // 监听配置变化
    manager.configCenter.Watch(serviceName+".*", manager)
    
    // 加载初始配置
    manager.loadConfigs()
    
    return manager
}

// OnConfigChanged 配置变化回调
func (m *ServiceConfigManager) OnConfigChanged(key string, value interface{}) {
    m.configsMu.Lock()
    defer m.configsMu.Unlock()
    
    m.configs[key] = value
    
    // 可以在这里添加配置热更新逻辑
    fmt.Printf("Config changed: %s = %v\n", key, value)
}

// GetConfig 获取配置值
func (m *ServiceConfigManager) GetConfig(key string, defaultValue interface{}) interface{} {
    m.configsMu.RLock()
    defer m.configsMu.RUnlock()
    
    if value, exists := m.configs[key]; exists {
        return value
    }
    
    // 从配置中心获取
    fullKey := m.serviceName + "." + key
    entry, err := m.configCenter.Get(fullKey)
    if err != nil {
        return defaultValue
    }
    
    m.configs[key] = entry.Value
    return entry.Value
}

// loadConfigs 加载初始配置
func (m *ServiceConfigManager) loadConfigs() {
    // 这里可以批量加载服务相关的所有配置
    // 简化实现，实际可以使用pattern匹配
    commonKeys := []string{
        "database.host", "database.port", "database.name",
        "redis.host", "redis.port",
        "log.level", "log.format",
    }
    
    for _, key := range commonKeys {
        fullKey := m.serviceName + "." + key
        if entry, err := m.configCenter.Get(fullKey); err == nil {
            m.configsMu.Lock()
            m.configs[key] = entry.Value
            m.configsMu.Unlock()
        }
    }
}
```

---

## 🔗 链路追踪监控

### 1. 分布式追踪实现

```go
package tracing

import (
    "context"
    "fmt"
    "net/http"
    "time"
    
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "github.com/uber/jaeger-client-go"
    "github.com/uber/jaeger-client-go/config"
)

// TracingConfig 追踪配置
type TracingConfig struct {
    ServiceName string
    JaegerURL   string
    SampleRate  float64
}

// InitTracing 初始化分布式追踪
func InitTracing(cfg *TracingConfig) (opentracing.Tracer, error) {
    jaegerCfg := config.Configuration{
        ServiceName: cfg.ServiceName,
        Sampler: &config.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: cfg.SampleRate,
        },
        Reporter: &config.ReporterConfig{
            LogSpans:            true,
            BufferFlushInterval: 1 * time.Second,
            LocalAgentHostPort:  cfg.JaegerURL,
        },
    }
    
    tracer, _, err := jaegerCfg.NewTracer()
    if err != nil {
        return nil, err
    }
    
    opentracing.SetGlobalTracer(tracer)
    return tracer, nil
}

// TracingMiddleware HTTP请求追踪中间件
func TracingMiddleware() mvc.HandlerFunc {
    return func(c *mvc.Context) {
        tracer := opentracing.GlobalTracer()
        
        // 从请求头提取Span上下文
        spanCtx, _ := tracer.Extract(
            opentracing.HTTPHeaders,
            opentracing.HTTPHeadersCarrier(c.Request.Header),
        )
        
        // 创建新的Span
        span := tracer.StartSpan(
            fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
            ext.RPCServerOption(spanCtx),
        )
        defer span.Finish()
        
        // 设置标签
        ext.HTTPMethod.Set(span, c.Request.Method)
        ext.HTTPUrl.Set(span, c.Request.URL.String())
        ext.Component.Set(span, "yyhertz-server")
        
        // 将Span上下文传递给下游
        ctx := opentracing.ContextWithSpan(c.Request.Context(), span)
        c.Request = c.Request.WithContext(ctx)
        c.Set("tracing_span", span)
        
        c.Next()
        
        // 设置响应状态
        ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
        if c.Writer.Status() >= 400 {
            ext.Error.Set(span, true)
        }
    }
}

// ServiceClient 带追踪的服务客户端
type TracedServiceClient struct {
    client      ServiceClient
    serviceName string
}

func NewTracedServiceClient(client ServiceClient, serviceName string) *TracedServiceClient {
    return &TracedServiceClient{
        client:      client,
        serviceName: serviceName,
    }
}

func (c *TracedServiceClient) Request(ctx context.Context, method, path string, body interface{}) (*Response, error) {
    span, ctx := opentracing.StartSpanFromContext(ctx, 
        fmt.Sprintf("%s %s", method, path))
    defer span.Finish()
    
    // 设置标签
    ext.SpanKindRPCClient.Set(span)
    ext.HTTPMethod.Set(span, method)
    ext.HTTPUrl.Set(span, path)
    ext.Component.Set(span, "yyhertz-client")
    span.SetTag("service.name", c.serviceName)
    
    // 注入追踪头
    headers := make(http.Header)
    opentracing.GlobalTracer().Inject(
        span.Context(),
        opentracing.HTTPHeaders,
        opentracing.HTTPHeadersCarrier(headers),
    )
    
    // 执行请求
    resp, err := c.client.RequestWithHeaders(method, path, body, headers)
    if err != nil {
        ext.Error.Set(span, true)
        span.LogKV("error", err.Error())
        return nil, err
    }
    
    ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
    return resp, nil
}
```

### 2. 业务追踪实现

```go
// 业务层追踪示例
func (s *OrderService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
    span, ctx := opentracing.StartSpanFromContext(ctx, "OrderService.CreateOrder")
    defer span.Finish()
    
    span.SetTag("user_id", req.UserID)
    span.SetTag("item_count", len(req.Items))
    
    // 1. 验证用户
    span.LogKV("step", "validate_user")
    user, err := s.validateUserWithTracing(ctx, req.UserID)
    if err != nil {
        span.SetTag("error", true)
        span.LogKV("error", err.Error())
        return nil, err
    }
    span.LogKV("user_validated", user.Username)
    
    // 2. 验证商品和库存
    span.LogKV("step", "validate_items")
    orderItems, totalAmount, err := s.validateOrderItemsWithTracing(ctx, req.Items)
    if err != nil {
        span.SetTag("error", true)
        span.LogKV("error", err.Error())
        return nil, err
    }
    span.SetTag("total_amount", totalAmount)
    
    // 3. 创建订单
    span.LogKV("step", "create_order")
    order := &Order{
        UserID:      req.UserID,
        Items:       orderItems,
        TotalAmount: totalAmount,
        Status:      OrderStatusPending,
    }
    
    if err := s.repo.CreateOrder(order); err != nil {
        span.SetTag("error", true)
        span.LogKV("error", err.Error())
        return nil, err
    }
    
    span.SetTag("order_id", order.ID)
    span.LogKV("step", "order_created")
    
    // 4. 异步处理支付
    go func() {
        paymentSpan := opentracing.StartSpan(
            "OrderService.ProcessPayment",
            opentracing.ChildOf(span.Context()),
        )
        defer paymentSpan.Finish()
        
        paymentCtx := opentracing.ContextWithSpan(context.Background(), paymentSpan)
        s.processPaymentWithTracing(paymentCtx, order)
    }()
    
    return order, nil
}

func (s *OrderService) validateUserWithTracing(ctx context.Context, userID uint) (*User, error) {
    span, ctx := opentracing.StartSpanFromContext(ctx, "validate_user")
    defer span.Finish()
    
    span.SetTag("user_id", userID)
    
    // 调用用户服务
    resp, err := s.userClient.Request(ctx, "GET", fmt.Sprintf("/users/%d", userID), nil)
    if err != nil {
        return nil, err
    }
    
    if resp.StatusCode == 404 {
        return nil, fmt.Errorf("用户不存在")
    }
    
    var user User
    if err := resp.JSON(&user); err != nil {
        return nil, err
    }
    
    return &user, nil
}

// 性能监控
type PerformanceCollector struct {
    spans    []SpanMetric
    spansMu  sync.RWMutex
}

type SpanMetric struct {
    OperationName string
    ServiceName   string
    Duration      time.Duration
    Success       bool
    Timestamp     time.Time
    Tags          map[string]interface{}
}

func (pc *PerformanceCollector) CollectSpan(span opentracing.Span) {
    // 从span中提取性能指标
    // 这里需要实现span数据的提取逻辑
}

func (pc *PerformanceCollector) GetMetrics(timeRange time.Duration) []SpanMetric {
    pc.spansMu.RLock()
    defer pc.spansMu.RUnlock()
    
    cutoff := time.Now().Add(-timeRange)
    var metrics []SpanMetric
    
    for _, span := range pc.spans {
        if span.Timestamp.After(cutoff) {
            metrics = append(metrics, span)
        }
    }
    
    return metrics
}
```

---

## 🛡️ 故障容错机制

### 1. 熔断器实现

```go
package circuitbreaker

import (
    "errors"
    "sync"
    "time"
)

// State 熔断器状态
type State int

const (
    StateClosed State = iota
    StateHalfOpen
    StateOpen
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
    maxRequests    uint32
    interval       time.Duration
    timeout        time.Duration
    readyToTrip    func(counts Counts) bool
    onStateChange  func(name string, from State, to State)
    
    mutex      sync.Mutex
    state      State
    generation uint64
    counts     Counts
    expiry     time.Time
}

// Counts 统计计数
type Counts struct {
    Requests             uint32
    TotalSuccesses       uint32
    TotalFailures        uint32
    ConsecutiveSuccesses uint32
    ConsecutiveFailures  uint32
}

// Config 熔断器配置
type Config struct {
    Name          string
    MaxRequests   uint32
    Interval      time.Duration
    Timeout       time.Duration
    ReadyToTrip   func(counts Counts) bool
    OnStateChange func(name string, from State, to State)
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config Config) *CircuitBreaker {
    cb := &CircuitBreaker{
        maxRequests:   config.MaxRequests,
        interval:      config.Interval,
        timeout:       config.Timeout,
        readyToTrip:   config.ReadyToTrip,
        onStateChange: config.OnStateChange,
        state:         StateClosed,
    }
    
    if cb.maxRequests == 0 {
        cb.maxRequests = 1
    }
    
    if cb.interval == 0 {
        cb.interval = time.Minute
    }
    
    if cb.timeout == 0 {
        cb.timeout = 60 * time.Second
    }
    
    if cb.readyToTrip == nil {
        cb.readyToTrip = func(counts Counts) bool {
            return counts.ConsecutiveFailures > 5
        }
    }
    
    return cb
}

// Execute 执行函数调用
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
    generation, err := cb.beforeRequest()
    if err != nil {
        return nil, err
    }
    
    defer func() {
        if r := recover(); r != nil {
            cb.afterRequest(generation, false)
            panic(r)
        }
    }()
    
    result, err := req()
    cb.afterRequest(generation, err == nil)
    return result, err
}

// beforeRequest 请求前检查
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    now := time.Now()
    state, generation := cb.currentState(now)
    
    if state == StateOpen {
        return generation, errors.New("circuit breaker is open")
    } else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
        return generation, errors.New("circuit breaker is half open with too many requests")
    }
    
    cb.counts.Requests++
    return generation, nil
}

// afterRequest 请求后处理
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    now := time.Now()
    state, generation := cb.currentState(now)
    if generation != before {
        return
    }
    
    if success {
        cb.onSuccess(state, now)
    } else {
        cb.onFailure(state, now)
    }
}

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
    cb.counts.TotalSuccesses++
    cb.counts.ConsecutiveSuccesses++
    cb.counts.ConsecutiveFailures = 0
    
    if state == StateHalfOpen {
        cb.setState(StateClosed, now)
    }
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
    cb.counts.TotalFailures++
    cb.counts.ConsecutiveFailures++
    cb.counts.ConsecutiveSuccesses = 0
    
    if cb.readyToTrip(cb.counts) {
        cb.setState(StateOpen, now)
    }
}

// currentState 获取当前状态
func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
    switch cb.state {
    case StateClosed:
        if !cb.expiry.IsZero() && cb.expiry.Before(now) {
            cb.toNewGeneration(now)
        }
    case StateOpen:
        if cb.expiry.Before(now) {
            cb.setState(StateHalfOpen, now)
        }
    }
    return cb.state, cb.generation
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state State, now time.Time) {
    if cb.state == state {
        return
    }
    
    prev := cb.state
    cb.state = state
    cb.toNewGeneration(now)
    
    if cb.onStateChange != nil {
        cb.onStateChange("", prev, state)
    }
}

// toNewGeneration 进入新一代
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
    cb.generation++
    cb.counts = Counts{}
    
    var zero time.Time
    switch cb.state {
    case StateClosed:
        if cb.interval == 0 {
            cb.expiry = zero
        } else {
            cb.expiry = now.Add(cb.interval)
        }
    case StateOpen:
        cb.expiry = now.Add(cb.timeout)
    default:
        cb.expiry = zero
    }
}

// State 获取当前状态
func (cb *CircuitBreaker) State() State {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    now := time.Now()
    state, _ := cb.currentState(now)
    return state
}

// Counts 获取统计信息
func (cb *CircuitBreaker) Counts() Counts {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    return cb.counts
}
```

### 2. 重试机制实现

```go
package retry

import (
    "context"
    "errors"
    "math"
    "time"
)

// RetryPolicy 重试策略
type RetryPolicy struct {
    MaxAttempts   int
    InitialDelay  time.Duration
    MaxDelay      time.Duration
    Multiplier    float64
    Jitter        bool
    RetryableFunc func(error) bool
}

// DefaultRetryPolicy 默认重试策略
var DefaultRetryPolicy = &RetryPolicy{
    MaxAttempts:  3,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     10 * time.Second,
    Multiplier:   2.0,
    Jitter:       true,
    RetryableFunc: func(err error) bool {
        // 默认所有错误都重试
        return true
    },
}

// Retrier 重试器
type Retrier struct {
    policy *RetryPolicy
}

// NewRetrier 创建重试器
func NewRetrier(policy *RetryPolicy) *Retrier {
    if policy == nil {
        policy = DefaultRetryPolicy
    }
    return &Retrier{policy: policy}
}

// Execute 执行重试逻辑
func (r *Retrier) Execute(ctx context.Context, fn func() error) error {
    var lastErr error
    
    for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // 检查是否可重试
        if !r.policy.RetryableFunc(err) {
            return err
        }
        
        // 最后一次尝试失败，不再重试
        if attempt >= r.policy.MaxAttempts {
            break
        }
        
        // 计算延迟时间
        delay := r.calculateDelay(attempt)
        
        // 等待重试
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            // 继续重试
        }
    }
    
    return lastErr
}

// ExecuteWithResult 执行重试逻辑并返回结果
func (r *Retrier) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
    var lastErr error
    var result interface{}
    
    for attempt := 1; attempt <= r.policy.MaxAttempts; attempt++ {
        var err error
        result, err = fn()
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        if !r.policy.RetryableFunc(err) {
            return result, err
        }
        
        if attempt >= r.policy.MaxAttempts {
            break
        }
        
        delay := r.calculateDelay(attempt)
        
        select {
        case <-ctx.Done():
            return result, ctx.Err()
        case <-time.After(delay):
        }
    }
    
    return result, lastErr
}

// calculateDelay 计算延迟时间
func (r *Retrier) calculateDelay(attempt int) time.Duration {
    delay := time.Duration(float64(r.policy.InitialDelay) * math.Pow(r.policy.Multiplier, float64(attempt-1)))
    
    if delay > r.policy.MaxDelay {
        delay = r.policy.MaxDelay
    }
    
    if r.policy.Jitter {
        // 添加随机抖动，避免惊群效应
        jitter := time.Duration(float64(delay) * (0.5 + 0.5*rand.Float64()))
        delay = jitter
    }
    
    return delay
}

// 服务客户端集成重试和熔断
type ResilientServiceClient struct {
    client         ServiceClient
    circuitBreaker *CircuitBreaker
    retrier        *Retrier
}

func NewResilientServiceClient(client ServiceClient) *ResilientServiceClient {
    // 配置熔断器
    cbConfig := Config{
        Name:        "service-client",
        MaxRequests: 10,
        Interval:    60 * time.Second,
        Timeout:     30 * time.Second,
        ReadyToTrip: func(counts Counts) bool {
            return counts.Requests >= 3 && counts.TotalFailures >= 2
        },
    }
    
    // 配置重试策略
    retryPolicy := &RetryPolicy{
        MaxAttempts:  3,
        InitialDelay: 200 * time.Millisecond,
        MaxDelay:     2 * time.Second,
        Multiplier:   2.0,
        Jitter:       true,
        RetryableFunc: func(err error) bool {
            // 只对特定类型的错误进行重试
            return isRetryableError(err)
        },
    }
    
    return &ResilientServiceClient{
        client:         client,
        circuitBreaker: NewCircuitBreaker(cbConfig),
        retrier:        NewRetrier(retryPolicy),
    }
}

func (c *ResilientServiceClient) Request(ctx context.Context, method, path string, body interface{}) (*Response, error) {
    return c.retrier.ExecuteWithResult(ctx, func() (interface{}, error) {
        return c.circuitBreaker.Execute(func() (interface{}, error) {
            return c.client.Request(method, path, body)
        })
    })
}

func isRetryableError(err error) bool {
    // 判断错误是否可重试
    // 例如：网络错误、超时、5xx服务器错误等
    if err == nil {
        return false
    }
    
    // 这里可以根据具体错误类型判断
    errStr := err.Error()
    return strings.Contains(errStr, "timeout") ||
           strings.Contains(errStr, "connection") ||
           strings.Contains(errStr, "network")
}
```

---

## 📦 部署运维实践

### 1. Docker容器化部署

```dockerfile
# 多阶段构建Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# 运行时镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# 从builder阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/migrations ./migrations

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# 启动应用
CMD ["./main"]
```

### 2. Kubernetes部署配置

```yaml
# user-service-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
  labels:
    app: user-service
spec:
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: user-service:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        - name: REDIS_URL
          value: "redis:6379"
        - name: SERVICE_NAME
          value: "user-service"
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
          requests:
            cpu: 250m
            memory: 256Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: user-service
spec:
  selector:
    app: user-service
  ports:
  - protocol: TCP
    port: 8080
    targetPort: 8080
  type: ClusterIP
```

### 3. 监控配置

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    scrape_configs:
    - job_name: 'microservices'
      kubernetes_sd_configs:
      - role: pod
      relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: (user-service|order-service|product-service)
      - source_labels: [__meta_kubernetes_pod_ip]
        target_label: __address__
        replacement: ${1}:8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prometheus
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prometheus
  template:
    metadata:
      labels:
        app: prometheus
    spec:
      containers:
      - name: prometheus
        image: prom/prometheus:latest
        ports:
        - containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/prometheus
      volumes:
      - name: config
        configMap:
          name: prometheus-config
```

---

## 🎬 完整项目案例

### 1. 电商微服务架构

```
电商微服务系统架构
┌─────────────────────────────────────────────────────────┐
│                     API Gateway                        │
│                   (统一入口/认证/限流)                    │
└─────────────────────────┬───────────────────────────────┘
                          │
        ┌─────────────────┼─────────────────┐
        │                 │                 │
   ┌────▼────┐      ┌────▼────┐      ┌────▼────┐
   │ 用户服务  │      │ 商品服务  │      │ 订单服务  │
   │         │      │         │      │         │
   │ - 注册   │      │ - 商品   │      │ - 下单   │
   │ - 登录   │      │ - 分类   │      │ - 支付   │
   │ - 资料   │      │ - 搜索   │      │ - 物流   │
   └─────────┘      └─────────┘      └─────────┘
        │                 │                 │
   ┌────▼────┐      ┌────▼────┐      ┌────▼────┐
   │ 用户数据库 │      │ 商品数据库 │      │ 订单数据库 │
   └─────────┘      └─────────┘      └─────────┘

        ┌─────────────────┐
        │   消息队列/事件总线  │ 
        │   (服务间通信)     │
        └─────────────────┘
                 │
     ┌───────────┼───────────┐
┌────▼────┐ ┌───▼───┐ ┌────▼────┐
│ 库存服务  │ │ 支付服务│ │ 通知服务  │
│         │ │       │ │         │
│ - 扣减   │ │ - 支付 │ │ - 短信   │
│ - 回滚   │ │ - 退款 │ │ - 邮件   │
│ - 预占   │ │ - 对账 │ │ - 推送   │
└─────────┘ └───────┘ └─────────┘
```

### 2. 服务启动配置

```go
// main.go - 用户服务启动入口
package main

import (
    "log"
    
    "github.com/zsy619/yyhertz/framework/microservice"
    "github.com/zsy619/yyhertz/framework/mvc"
    "your-app/user"
    "your-app/config"
)

func main() {
    // 初始化配置
    if err := config.Init(); err != nil {
        log.Fatal("Failed to initialize config:", err)
    }
    
    // 初始化数据库
    if err := initDatabase(); err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    
    // 初始化Redis
    if err := initRedis(); err != nil {
        log.Fatal("Failed to initialize Redis:", err)
    }
    
    // 初始化消息队列
    if err := initMessageQueue(); err != nil {
        log.Fatal("Failed to initialize message queue:", err)
    }
    
    // 初始化链路追踪
    tracer, err := initTracing("user-service")
    if err != nil {
        log.Fatal("Failed to initialize tracing:", err)
    }
    defer tracer.Close()
    
    // 创建微服务实例
    service := microservice.NewService(&microservice.Config{
        Name:    "user-service",
        Version: "v1.0.0",
        Port:    8001,
    })
    
    // 注册服务实现
    userService := user.NewUserService()
    service.RegisterService(userService)
    
    // 添加中间件
    service.Use(
        middleware.Recovery(),
        middleware.Logger(),
        middleware.CORS(),
        middleware.Tracing(),
        middleware.Metrics(),
    )
    
    // 启动服务
    if err := service.Start(); err != nil {
        log.Fatal("Failed to start service:", err)
    }
}

func initDatabase() error {
    // 数据库初始化逻辑
    return nil
}

func initRedis() error {
    // Redis初始化逻辑
    return nil
}

func initMessageQueue() error {
    // 消息队列初始化逻辑
    return nil
}

func initTracing(serviceName string) (io.Closer, error) {
    // 链路追踪初始化逻辑
    return nil, nil
}
```

### 3. 完整服务示例

```go
// user/service.go
package user

import (
    "context"
    "github.com/zsy619/yyhertz/framework/microservice"
    "github.com/zsy619/yyhertz/framework/mvc"
)

type UserService struct {
    microservice.BaseService
    repo         *UserRepository
    eventBus     *events.EventBus
    configMgr    *config.ServiceConfigManager
}

func NewUserService() *UserService {
    return &UserService{
        repo:      NewUserRepository(),
        eventBus:  events.DefaultEventBus,
        configMgr: config.NewServiceConfigManager("user-service"),
    }
}

func (s *UserService) ServiceInfo() *microservice.ServiceInfo {
    return &microservice.ServiceInfo{
        Name:        "user-service",
        Version:     "v1.0.0",
        Description: "用户管理微服务",
        Port:        8001,
        HealthCheck: "/health",
        Endpoints: []microservice.Endpoint{
            {Path: "/users", Methods: []string{"GET", "POST"}},
            {Path: "/users/{id}", Methods: []string{"GET", "PUT", "DELETE"}},
        },
        Dependencies: []string{"database", "redis", "message-queue"},
    }
}

func (s *UserService) RegisterRoutes(app *mvc.Application) {
    api := app.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.GET("", s.GetUsers)
            users.POST("", s.CreateUser)
            users.GET("/:id", s.GetUser)
            users.PUT("/:id", s.UpdateUser)
            users.DELETE("/:id", s.DeleteUser)
        }
    }
    
    // 健康检查端点
    app.GET("/health", s.HealthCheck)
    app.GET("/ready", s.ReadinessCheck)
}

func (s *UserService) GetUsers(c *mvc.Context) {
    // 实现用户列表查询
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    
    users, total, err := s.repo.FindUsers(page, pageSize)
    if err != nil {
        c.ErrorJSON(500, "查询失败")
        return
    }
    
    c.JSON(map[string]interface{}{
        "users": users,
        "total": total,
        "pagination": map[string]int{
            "page":      page,
            "page_size": pageSize,
            "total_pages": (total + pageSize - 1) / pageSize,
        },
    })
}

func (s *UserService) CreateUser(c *mvc.Context) {
    var req CreateUserRequest
    if err := c.BindJSON(&req); err != nil {
        c.ErrorJSON(400, "参数错误")
        return
    }
    
    // 参数验证
    if err := validator.Validate(req); err != nil {
        c.ValidationErrorJSON(err)
        return
    }
    
    // 创建用户
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Phone:    req.Phone,
    }
    
    if err := s.repo.CreateUser(user); err != nil {
        c.ErrorJSON(500, "创建失败")
        return
    }
    
    // 发布用户创建事件
    event := events.NewUserCreatedEvent(user.ID, user.Username, user.Email)
    s.eventBus.Publish(event)
    
    c.JSON(user)
}

func (s *UserService) HealthCheck(c *mvc.Context) {
    checks := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now(),
        "version":   s.ServiceInfo().Version,
    }
    
    // 检查数据库连接
    if err := s.repo.Ping(); err != nil {
        checks["database"] = "unhealthy"
        checks["status"] = "unhealthy"
        c.JSON(503, checks)
        return
    }
    checks["database"] = "healthy"
    
    // 检查Redis连接
    if err := s.checkRedis(); err != nil {
        checks["redis"] = "unhealthy"
        checks["status"] = "unhealthy"
        c.JSON(503, checks)
        return
    }
    checks["redis"] = "healthy"
    
    c.JSON(checks)
}

func (s *UserService) ReadinessCheck(c *mvc.Context) {
    // 就绪检查，确保服务已准备好接收流量
    if !s.isReady() {
        c.JSON(503, map[string]string{
            "status": "not ready",
        })
        return
    }
    
    c.JSON(map[string]string{
        "status": "ready",
    })
}

func (s *UserService) isReady() bool {
    // 检查服务是否准备就绪
    return s.repo != nil && s.eventBus != nil
}

func (s *UserService) checkRedis() error {
    // 检查Redis连接状态
    return nil
}

// 启动时初始化
func (s *UserService) OnStart() error {
    // 服务启动时的初始化逻辑
    log.Println("User service starting...")
    
    // 注册事件监听器
    s.eventBus.Subscribe("order.created", &UserEventHandler{s})
    
    return nil
}

// 停止时清理
func (s *UserService) OnStop() error {
    // 服务停止时的清理逻辑
    log.Println("User service stopping...")
    
    if s.repo != nil {
        s.repo.Close()
    }
    
    return nil
}
```

---

## 📝 总结

通过本教程，你已经掌握了使用YYHertz构建企业级微服务系统的完整技能：

### 🎯 核心能力
- **服务拆分设计** - 领域驱动的微服务架构
- **服务注册发现** - 自动化的服务管理
- **API网关实现** - 统一的服务入口
- **服务间通信** - 事件驱动架构
- **配置中心管理** - 动态配置更新
- **链路追踪监控** - 分布式系统可观测性
- **故障容错机制** - 熔断、重试、降级
- **容器化部署** - Docker + Kubernetes

### 💡 架构原则
- **单一职责** - 每个服务专注于特定业务领域
- **服务自治** - 独立部署、数据隔离
- **去中心化** - 避免单点故障
- **容错设计** - 优雅处理故障
- **可观测性** - 完善的监控体系

### 🚀 最佳实践
- **渐进式拆分** - 从单体逐步演进到微服务
- **数据一致性** - 事件驱动的最终一致性
- **服务边界** - 基于业务能力划分服务
- **API版本管理** - 向后兼容的接口设计
- **安全策略** - 服务间认证授权

---

<div align="center">

**🏗️ 构建现代化微服务架构，拥抱云原生时代！**

**从单体到微服务，从开发到运维，YYHertz助力企业数字化转型！🚀**

</div>