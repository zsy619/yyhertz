// Package examples Gin风格的使用示例
// 展示如何使用增强后的YZHertz框架
package examples

import (
	"net/http"
	"time"

	"github.com/zsy619/yyhertz/framework/gin"
)

// User 用户模型
type User struct {
	ID       int64  `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"gte=0,lte=130"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// QueryParams 查询参数
type QueryParams struct {
	Page     int    `form:"page" binding:"required,gte=1"`
	PageSize int    `form:"page_size" binding:"required,gte=1,lte=100"`
	Keyword  string `form:"keyword"`
}

// PathParams 路径参数
type PathParams struct {
	ID int64 `uri:"id" binding:"required,gte=1"`
}

// GinStyleExample Gin风格示例
func GinStyleExample() {
	// 创建路由引擎，带默认中间件（Logger + Recovery）
	r := gin.Default()

	// 添加自定义中间件
	r.Use(customMiddleware())

	// ============= 基础路由 =============

	// 简单的GET路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]string{
			"message": "pong",
		})
	})

	// 带路径参数的路由
	r.GET("/users/:id", getUserHandler)

	// 带查询参数的路由
	r.GET("/search", searchHandler)

	// POST路由（JSON绑定）
	r.POST("/users", createUserHandler)

	// PUT路由（表单绑定）
	r.PUT("/users/:id", updateUserHandler)

	// DELETE路由
	r.DELETE("/users/:id", deleteUserHandler)

	// ============= 路由组 =============

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthCheckHandler)
		v1.POST("/login", loginHandler)

		// 需要认证的路由组
		authorized := v1.Group("/")
		authorized.Use(authMiddleware())
		{
			authorized.GET("/profile", getProfileHandler)
			authorized.PUT("/profile", updateProfileHandler)
			authorized.POST("/logout", logoutHandler)
		}
	}

	// API v2 路由组
	v2 := r.Group("/api/v2")
	v2.Use(rateLimitMiddleware()) // 限流中间件
	{
		v2.GET("/users", listUsersHandler)
		v2.GET("/users/:id/posts", getUserPostsHandler)
	}

	// ============= 静态文件 =============

	// 静态文件服务
	r.Static("/static", "./static")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// ============= 错误处理 =============

	// 404处理
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, map[string]string{
			"error": "Page not found",
			"path":  string(c.URI().Path()),
		})
	})

	// 405处理
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, map[string]string{
			"error":  "Method not allowed",
			"method": string(c.Method()),
		})
	})

	// 启动服务器
	r.Run(":8080")
}

// ============= 处理器函数 =============

// getUserHandler 获取用户处理器
func getUserHandler(c *gin.Context) {
	// 绑定路径参数
	var params PathParams
	if err := c.ShouldBindUri(&params); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 模拟从数据库获取用户
	user := User{
		ID:    params.ID,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	c.JSON(http.StatusOK, user)
}

// searchHandler 搜索处理器
func searchHandler(c *gin.Context) {
	// 绑定查询参数
	var params QueryParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 设置默认值
	if params.PageSize == 0 {
		params.PageSize = 10
	}

	c.JSON(http.StatusOK, map[string]any{
		"page":      params.Page,
		"page_size": params.PageSize,
		"keyword":   params.Keyword,
		"results":   []string{"result1", "result2", "result3"},
		"total":     100,
	})
}

// createUserHandler 创建用户处理器
func createUserHandler(c *gin.Context) {
	// 绑定JSON数据
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 模拟保存用户到数据库
	user.ID = time.Now().Unix()

	c.JSON(http.StatusCreated, user)
}

// updateUserHandler 更新用户处理器
func updateUserHandler(c *gin.Context) {
	// 绑定路径参数
	var pathParams PathParams
	if err := c.ShouldBindUri(&pathParams); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 绑定表单数据
	var user User
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	user.ID = pathParams.ID

	c.JSON(http.StatusOK, user)
}

// deleteUserHandler 删除用户处理器
func deleteUserHandler(c *gin.Context) {
	id := c.Param("id")

	// 模拟删除用户
	c.JSON(http.StatusOK, map[string]string{
		"message": "User deleted successfully",
		"id":      id,
	})
}

// healthCheckHandler 健康检查处理器
func healthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"uptime":    "5 minutes",
	})
}

// loginHandler 登录处理器
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	// 模拟验证用户
	if req.Username == "admin" && req.Password == "password" {
		c.JSON(http.StatusOK, map[string]string{
			"message": "Login successful",
			"token":   "jwt-token-here",
		})
	} else {
		c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials",
		})
	}
}

// getProfileHandler 获取用户资料处理器
func getProfileHandler(c *gin.Context) {
	// 从中间件获取用户信息
	userID := c.GetString("user_id")

	c.JSON(http.StatusOK, map[string]any{
		"user_id": userID,
		"name":    "John Doe",
		"email":   "john@example.com",
	})
}

// updateProfileHandler 更新用户资料处理器
func updateProfileHandler(c *gin.Context) {
	userID := c.GetString("user_id")

	var profile map[string]any
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"message": "Profile updated successfully",
		"user_id": userID,
		"profile": profile,
	})
}

// logoutHandler 登出处理器
func logoutHandler(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"message": "Logout successful",
	})
}

// listUsersHandler 用户列表处理器
func listUsersHandler(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{
		"users": []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 25},
			{ID: 2, Name: "Bob", Email: "bob@example.com", Age: 30},
		},
		"total": 2,
	})
}

// getUserPostsHandler 获取用户文章处理器
func getUserPostsHandler(c *gin.Context) {
	userID := c.Param("id")

	c.JSON(http.StatusOK, map[string]any{
		"user_id": userID,
		"posts": []map[string]any{
			{"id": 1, "title": "Post 1", "content": "Content 1"},
			{"id": 2, "title": "Post 2", "content": "Content 2"},
		},
	})
}

// ============= 中间件函数 =============

// customMiddleware 自定义中间件
func customMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求开始时间
		start := time.Now()

		// 设置自定义头
		c.Header("X-Custom-Header", "YYHertz-Framework")

		// 继续处理
		c.Next()

		// 计算处理时间
		latency := time.Since(start)
		c.Header("X-Response-Time", latency.String())
	}
}

// authMiddleware 认证中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{
				"error": "Authorization header required",
			})
			return
		}

		// 模拟token验证
		if token != "Bearer valid-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
			return
		}

		// 设置用户信息
		c.Set("user_id", "12345")
		c.Set("username", "john_doe")

		c.Next()
	}
}

// rateLimitMiddleware 限流中间件
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单的IP限流逻辑
		clientIP := c.ClientIP()

		// 模拟限流检查
		if clientIP == "192.168.1.100" {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, map[string]string{
				"error": "Rate limit exceeded",
			})
			return
		}

		c.Next()
	}
}
