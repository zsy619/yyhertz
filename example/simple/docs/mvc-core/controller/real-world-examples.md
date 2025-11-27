# 🌍 实际应用场景示例

本文档展示YYHertz控制器在真实项目中的应用场景，包括电商系统、用户管理和CMS系统的完整实现。

## 🛒 电商系统应用

### 商品管理控制器

```go
package controllers

import (
    "strconv"
    "time"
    "github.com/zsy619/yyhertz/framework/mvc/core"
)

type ProductController struct {
    core.BaseController
    productService *ProductService
    categoryService *CategoryService
    inventoryService *InventoryService
}

func (c *ProductController) Prepare() {
    // 🔐 管理员权限验证
    if c.needsAdminAccess() {
        c.RequireRole("admin")
    }
    
    // 🛡️ XSRF保护
    if c.needsXSRFProtection() {
        c.EnableXSRF(3600)
        if !c.CheckXSRFCookie() {
            c.JSONForbidden("CSRF令牌验证失败")
            c.StopRun()
            return
        }
    }
}

// GET /admin/products - 商品列表页面
func (c *ProductController) GetIndex() {
    // 📊 获取查询参数
    page := c.GetInt("page", 1)
    pageSize := c.GetInt("page_size", 20)
    category := c.GetString("category")
    status := c.GetString("status")
    keyword := c.GetString("keyword")
    
    // 🔍 构建查询条件
    query := ProductQuery{
        Page:     page,
        PageSize: pageSize,
        Category: category,
        Status:   status,
        Keyword:  keyword,
    }
    
    // 📦 获取商品数据
    products, total, err := c.productService.GetProducts(query)
    if err != nil {
        c.LogError("获取商品列表失败", map[string]any{"error": err.Error()})
        c.Error(500, "获取商品列表失败")
        return
    }
    
    // 📋 获取分类列表
    categories, _ := c.categoryService.GetAllCategories()
    
    // 📊 设置模板数据
    pagination := c.calculatePagination(page, pageSize, total)
    c.SetData("Products", products)
    c.SetData("Categories", categories)
    c.SetData("Pagination", pagination)
    c.SetData("Query", query)
    c.SetData("Title", "商品管理")
    
    c.SetLayout("layouts/admin.html")
    c.Render("admin/products/index.html")
}

// POST /admin/products - 创建商品
func (c *ProductController) PostCreate() {
    var req CreateProductRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("请求数据格式错误")
        return
    }
    
    // ✅ 数据验证
    if err := c.ValidateStruct(&req); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
    
    // 🏷️ 处理商品图片
    images := c.handleProductImages()
    req.Images = images
    
    // 💾 创建商品
    product, err := c.productService.CreateProduct(&req)
    if err != nil {
        c.LogError("创建商品失败", map[string]any{
            "error": err.Error(),
            "product": req.Name,
        })
        c.JSONInternalServerError("创建商品失败")
        return
    }
    
    // 📦 初始化库存
    if req.InitialStock > 0 {
        c.inventoryService.InitializeStock(product.ID, req.InitialStock)
    }
    
    c.LogInfo("商品创建成功", map[string]any{
        "product_id": product.ID,
        "name": product.Name,
        "admin_id": c.GetCurrentUserID(),
    })
    
    c.JSONSuccess("商品创建成功", product)
}

// PUT /admin/products/:id/inventory - 更新库存
func (c *ProductController) PutUpdateInventory() {
    productID := c.GetParam("id")
    operation := c.GetString("operation") // "add", "subtract", "set"
    quantity := c.GetInt("quantity")
    reason := c.GetString("reason")
    
    if quantity <= 0 {
        c.JSONBadRequest("数量必须大于0")
        return
    }
    
    // 📦 执行库存操作
    var err error
    switch operation {
    case "add":
        err = c.inventoryService.AddStock(productID, quantity, reason)
    case "subtract":
        err = c.inventoryService.SubtractStock(productID, quantity, reason)
    case "set":
        err = c.inventoryService.SetStock(productID, quantity, reason)
    default:
        c.JSONBadRequest("不支持的操作类型")
        return
    }
    
    if err != nil {
        c.JSONError("库存更新失败: " + err.Error())
        return
    }
    
    // 📊 获取更新后的库存信息
    inventory, _ := c.inventoryService.GetProductInventory(productID)
    
    c.LogInfo("库存更新", map[string]any{
        "product_id": productID,
        "operation": operation,
        "quantity": quantity,
        "current_stock": inventory.Quantity,
    })
    
    c.JSONSuccess("库存更新成功", inventory)
}

// 处理商品图片上传
func (c *ProductController) handleProductImages() []string {
    files := c.GetFiles("images")
    if len(files) == 0 {
        return []string{}
    }
    
    var imagePaths []string
    for _, file := range files {
        // 📏 检查文件
        if !c.isValidImageFile(file) {
            continue
        }
        
        // 💾 保存图片
        path, err := c.saveProductImage(file)
        if err != nil {
            c.LogWarn("图片保存失败", map[string]any{
                "filename": file.Filename,
                "error": err.Error(),
            })
            continue
        }
        
        imagePaths = append(imagePaths, path)
    }
    
    return imagePaths
}
```

### 订单处理控制器

```go
type OrderController struct {
    core.BaseController
    orderService *OrderService
    paymentService *PaymentService
    notificationService *NotificationService
}

// POST /orders - 创建订单
func (c *OrderController) PostCreate() {
    userID := c.GetCurrentUserID()
    if userID == 0 {
        c.JSONUnauthorized("请先登录")
        return
    }
    
    var req CreateOrderRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("订单数据格式错误")
        return
    }
    
    // 🛒 验证购物车
    if len(req.Items) == 0 {
        c.JSONBadRequest("购物车不能为空")
        return
    }
    
    // 💰 计算订单金额
    orderTotal, err := c.orderService.CalculateOrderTotal(req.Items)
    if err != nil {
        c.JSONError("订单金额计算失败")
        return
    }
    
    // 📦 检查库存
    if err := c.orderService.CheckInventory(req.Items); err != nil {
        c.JSONError("库存不足: " + err.Error())
        return
    }
    
    // 🎫 创建订单
    order, err := c.orderService.CreateOrder(&OrderInfo{
        UserID:      userID,
        Items:       req.Items,
        Total:       orderTotal,
        Address:     req.ShippingAddress,
        PaymentMethod: req.PaymentMethod,
    })
    
    if err != nil {
        c.JSONInternalServerError("订单创建失败")
        return
    }
    
    // 🔒 锁定库存
    c.inventoryService.ReserveStock(order.ID, req.Items)
    
    // 📧 发送订单确认
    c.sendOrderConfirmation(order)
    
    c.JSONSuccess("订单创建成功", map[string]any{
        "order_id": order.ID,
        "total": order.Total,
        "payment_url": c.generatePaymentURL(order),
    })
}

// POST /orders/:id/pay - 处理支付
func (c *OrderController) PostPay() {
    orderID := c.GetParam("id")
    paymentMethod := c.GetString("payment_method")
    
    // 🔍 获取订单信息
    order, err := c.orderService.GetOrder(orderID)
    if err != nil {
        c.JSONNotFound("订单不存在")
        return
    }
    
    // 🔐 权限检查
    if order.UserID != c.GetCurrentUserID() {
        c.JSONForbidden("无权限访问此订单")
        return
    }
    
    // 💳 处理支付
    paymentResult, err := c.paymentService.ProcessPayment(&PaymentRequest{
        OrderID: orderID,
        Amount:  order.Total,
        Method:  paymentMethod,
        UserID:  order.UserID,
    })
    
    if err != nil {
        c.JSONError("支付处理失败: " + err.Error())
        return
    }
    
    if paymentResult.Success {
        // 📦 确认库存扣减
        c.inventoryService.ConfirmStockReservation(orderID)
        
        // 📊 更新订单状态
        c.orderService.UpdateOrderStatus(orderID, "paid")
        
        // 📧 发送支付成功通知
        c.sendPaymentSuccessNotification(order)
        
        c.JSONSuccess("支付成功", paymentResult)
    } else {
        // 🔓 释放库存
        c.inventoryService.ReleaseStockReservation(orderID)
        c.JSONError("支付失败: " + paymentResult.Message)
    }
}
```

## 👥 用户管理系统

### 用户认证控制器

```go
type AuthController struct {
    core.BaseController
    userService *UserService
    authService *AuthService
    captchaService *CaptchaService
}

// POST /auth/register - 用户注册
func (c *AuthController) PostRegister() {
    var req UserRegisterRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("注册信息格式错误")
        return
    }
    
    // 🤖 验证码检查
    if !c.captchaService.Verify(req.CaptchaID, req.CaptchaCode) {
        c.JSONBadRequest("验证码错误")
        return
    }
    
    // ✅ 数据验证
    if err := c.ValidateStruct(&req); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
    
    // 🔍 检查邮箱是否已存在
    if exists := c.userService.EmailExists(req.Email); exists {
        c.JSONBadRequest("邮箱已被注册")
        return
    }
    
    // 📝 创建用户
    user, err := c.userService.CreateUser(&req)
    if err != nil {
        c.LogError("用户注册失败", map[string]any{
            "email": req.Email,
            "error": err.Error(),
        })
        c.JSONInternalServerError("注册失败")
        return
    }
    
    // 📧 发送欢迎邮件
    c.sendWelcomeEmail(user)
    
    // 📝 记录注册事件
    c.LogInfo("用户注册成功", map[string]any{
        "user_id": user.ID,
        "email": user.Email,
        "ip": c.GetClientIP(),
    })
    
    c.JSONSuccess("注册成功", map[string]any{
        "user_id": user.ID,
        "username": user.Username,
    })
}

// POST /auth/login - 用户登录
func (c *AuthController) PostLogin() {
    var req UserLoginRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("登录信息格式错误")
        return
    }
    
    // 🛡️ 登录限流检查
    clientIP := c.GetClientIP()
    if c.rateLimiter.IsBlocked(clientIP) {
        c.JSONError("登录尝试过于频繁，请稍后再试", nil, 429)
        return
    }
    
    // 🔐 验证用户身份
    user, err := c.authService.Authenticate(req.Email, req.Password)
    if err != nil {
        // 📊 记录登录失败
        c.rateLimiter.RecordFailure(clientIP)
        c.recordLoginAttempt(false, req.Email)
        
        c.JSONUnauthorized("邮箱或密码错误")
        return
    }
    
    // ✅ 检查账户状态
    if user.Status != "active" {
        c.JSONForbidden("账户已被禁用")
        return
    }
    
    // 🎫 生成访问令牌
    token, err := c.authService.GenerateAccessToken(user)
    if err != nil {
        c.JSONInternalServerError("令牌生成失败")
        return
    }
    
    // 💾 保存登录信息
    c.SetSession("user_id", user.ID)
    c.SetSession("username", user.Username)
    c.SetSession("role", user.Role)
    c.SetSession("login_time", time.Now())
    c.SetSession("ip_address", clientIP)
    
    // 🍪 设置登录Cookie
    c.SetSecureCookie("jwt_secret", "access_token", token, 7*24*3600)
    
    // 📊 更新最后登录时间
    c.userService.UpdateLastLogin(user.ID, clientIP)
    
    // 📝 记录登录成功
    c.recordLoginAttempt(true, req.Email)
    c.LogInfo("用户登录成功", map[string]any{
        "user_id": user.ID,
        "email": user.Email,
        "ip": clientIP,
    })
    
    c.JSONSuccess("登录成功", map[string]any{
        "access_token": token,
        "user": map[string]any{
            "id": user.ID,
            "username": user.Username,
            "email": user.Email,
            "role": user.Role,
        },
    })
}
```

## 📰 CMS内容管理

### 文章管理控制器

```go
type ArticleController struct {
    core.BaseController
    articleService *ArticleService
    categoryService *CategoryService
    tagService *TagService
}

// GET /admin/articles - 文章管理页面
func (c *ArticleController) GetIndex() {
    // 📊 查询参数
    page := c.GetInt("page", 1)
    status := c.GetString("status", "all")
    category := c.GetString("category")
    author := c.GetString("author")
    keyword := c.GetString("keyword")
    
    query := ArticleQuery{
        Page:     page,
        PageSize: 15,
        Status:   status,
        Category: category,
        Author:   author,
        Keyword:  keyword,
    }
    
    // 📚 获取文章列表
    articles, total, err := c.articleService.GetArticles(query)
    if err != nil {
        c.Error(500, "获取文章列表失败")
        return
    }
    
    // 📋 获取辅助数据
    categories, _ := c.categoryService.GetAllCategories()
    authors, _ := c.userService.GetAuthors()
    
    // 📊 统计信息
    stats, _ := c.articleService.GetStatistics()
    
    // 🎨 模板数据
    c.SetData("Articles", articles)
    c.SetData("Categories", categories)
    c.SetData("Authors", authors)
    c.SetData("Stats", stats)
    c.SetData("Query", query)
    c.SetData("Pagination", c.calculatePagination(page, 15, total))
    c.SetData("Title", "文章管理")
    
    c.SetLayout("layouts/admin.html")
    c.Render("admin/articles/index.html")
}

// POST /admin/articles - 创建文章
func (c *ArticleController) PostCreate() {
    var req CreateArticleRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("文章数据格式错误")
        return
    }
    
    // ✅ 验证数据
    if err := c.ValidateStruct(&req); err != nil {
        c.JSONBadRequest("数据验证失败: " + err.Error())
        return
    }
    
    // 🏷️ 处理标签
    tags, err := c.tagService.ProcessTags(req.Tags)
    if err != nil {
        c.JSONError("标签处理失败")
        return
    }
    
    // 🖼️ 处理封面图片
    coverImage := ""
    if file := c.GetFile("cover_image"); file != nil {
        path, err := c.uploadCoverImage(file)
        if err == nil {
            coverImage = path
        }
    }
    
    // 📝 创建文章
    article := &Article{
        Title:       req.Title,
        Content:     req.Content,
        Summary:     req.Summary,
        CategoryID:  req.CategoryID,
        AuthorID:    c.GetCurrentUserID(),
        Status:      req.Status,
        CoverImage:  coverImage,
        Tags:        tags,
        PublishTime: req.PublishTime,
        SEOTitle:    req.SEOTitle,
        SEOKeywords: req.SEOKeywords,
        SEODescription: req.SEODescription,
    }
    
    savedArticle, err := c.articleService.CreateArticle(article)
    if err != nil {
        c.JSONInternalServerError("文章创建失败")
        return
    }
    
    // 📊 如果是发布状态，触发相关操作
    if req.Status == "published" {
        // 🔄 更新站点地图
        c.updateSitemap()
        
        // 📧 通知订阅者
        c.notifySubscribers(savedArticle)
        
        // 🔍 提交搜索引擎
        c.submitToSearchEngines(savedArticle)
    }
    
    c.LogInfo("文章创建成功", map[string]any{
        "article_id": savedArticle.ID,
        "title": savedArticle.Title,
        "author_id": c.GetCurrentUserID(),
        "status": savedArticle.Status,
    })
    
    c.JSONSuccess("文章创建成功", savedArticle)
}

// PUT /admin/articles/:id/status - 批量更新文章状态
func (c *ArticleController) PutBatchUpdateStatus() {
    var req BatchUpdateStatusRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSONBadRequest("请求格式错误")
        return
    }
    
    if len(req.ArticleIDs) == 0 {
        c.JSONBadRequest("请选择要更新的文章")
        return
    }
    
    // 📊 执行批量更新
    updatedCount, err := c.articleService.BatchUpdateStatus(req.ArticleIDs, req.Status)
    if err != nil {
        c.JSONError("批量更新失败: " + err.Error())
        return
    }
    
    // 📝 记录操作日志
    c.LogInfo("批量更新文章状态", map[string]any{
        "article_ids": req.ArticleIDs,
        "status": req.Status,
        "updated_count": updatedCount,
        "operator_id": c.GetCurrentUserID(),
    })
    
    c.JSONSuccess("批量更新成功", map[string]any{
        "updated_count": updatedCount,
    })
}
```

## 🎯 应用场景总结

### 1. 电商系统特点
- **库存管理** - 实时库存跟踪和预留机制
- **订单处理** - 完整的订单生命周期管理
- **支付集成** - 多种支付方式支持
- **用户体验** - 购物车、愿望清单等功能

### 2. 用户管理特点
- **安全认证** - 多层安全验证机制
- **权限控制** - 基于角色的权限管理
- **用户体验** - 注册、登录、找回密码流程
- **数据保护** - 敏感信息加密存储

### 3. CMS系统特点
- **内容管理** - 文章、媒体、分类管理
- **工作流程** - 内容审核发布流程
- **SEO优化** - 搜索引擎优化功能
- **用户交互** - 评论、订阅等功能

每个应用场景都充分利用了YYHertz控制器的核心功能，展示了框架的灵活性和强大能力。