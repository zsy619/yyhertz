package captcha

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

// IntegrationExample 完整的集成示例，展示如何在实际项目中使用验证码
func IntegrationExample() {
	// 1. 创建Hertz应用
	h := server.Default(server.WithHostPorts(":8080"))

	// 2. 配置验证码
	config := &Config{
		Width:   150,
		Height:  50,
		Length:  5,
		TTL:     600, // 10分钟
		Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ",
	}

	// 3. 创建存储（可根据需要选择不同存储方式）
	store := NewMemoryStore()
	defer store.Close()

	// 4. 创建验证码生成器
	generator := NewGenerator(config, store)

	// 5. 注册验证码相关路由
	captchaGroup := h.Group("/captcha")
	{
		// 生成验证码ID
		captchaGroup.GET("/generate", GenerateHandler(generator))
		
		// 获取验证码图片
		captchaGroup.GET("/image/:id", ImageHandler(generator))
		
		// 验证验证码（独立验证接口）
		captchaGroup.POST("/verify", VerifyHandler(generator))
	}

	// 6. 创建验证码中间件
	middlewareConfig := &MiddlewareConfig{
		// 跳过验证码相关路径和公共资源
		SkipPaths: []string{
			"/captcha/",
			"/public/",
			"/static/",
			"/health",
		},
		// 自定义错误处理
		ErrorHandler: func(c context.Context, ctx *app.RequestContext, err error) {
			ctx.Header("Content-Type", "application/json")
			
			if captchaErr, ok := err.(*CaptchaError); ok {
				ctx.JSON(http.StatusBadRequest, utils.H{
					"success": false,
					"code":    captchaErr.Code,
					"message": captchaErr.Message,
					"data":    nil,
				})
			} else {
				ctx.JSON(http.StatusInternalServerError, utils.H{
					"success": false,
					"code":    ErrCodeInternal,
					"message": "系统内部错误",
					"data":    nil,
				})
			}
		},
		// 验证成功处理
		SuccessHandler: func(c context.Context, ctx *app.RequestContext) {
			// 可以在这里记录日志、更新统计等
			// log.Printf("验证码验证成功，来源IP: %s", ctx.ClientIP())
		},
	}

	captchaMiddleware := NewMiddleware(generator, middlewareConfig)

	// 7. 需要验证码保护的API路由
	apiGroup := h.Group("/api", captchaMiddleware.Handler())
	{
		// 用户认证相关接口
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/login", loginHandler)
			authGroup.POST("/register", registerHandler)
			authGroup.POST("/forgot-password", forgotPasswordHandler)
		}

		// 敏感操作接口
		userGroup := apiGroup.Group("/user")
		{
			userGroup.POST("/change-password", changePasswordHandler)
			userGroup.DELETE("/delete-account", deleteAccountHandler)
		}
	}

	// 8. 不需要验证码的公共接口
	publicGroup := h.Group("/public")
	{
		publicGroup.GET("/info", publicInfoHandler)
		publicGroup.GET("/health", healthCheckHandler)
	}

	// 9. 静态文件服务（包含验证码演示页面）
	h.Static("/static", "./static")
	h.GET("/", func(c context.Context, ctx *app.RequestContext) {
		ctx.HTML(http.StatusOK, "index.html", utils.H{
			"title": "验证码演示",
		})
	})

	// 10. 启动服务器
	h.Spin()
}

// 示例处理函数
func loginHandler(c context.Context, ctx *app.RequestContext) {
	username := string(ctx.PostForm("username"))
	password := string(ctx.PostForm("password"))

	// 模拟登录逻辑
	if username == "admin" && password == "password" {
		ctx.JSON(http.StatusOK, utils.H{
			"success": true,
			"message": "登录成功",
			"data": utils.H{
				"user_id": 1,
				"username": username,
				"token": "mock_jwt_token",
			},
		})
	} else {
		ctx.JSON(http.StatusUnauthorized, utils.H{
			"success": false,
			"message": "用户名或密码错误",
			"data": nil,
		})
	}
}

func registerHandler(c context.Context, ctx *app.RequestContext) {
	username := string(ctx.PostForm("username"))
	email := string(ctx.PostForm("email"))
	password := string(ctx.PostForm("password"))

	// 模拟注册逻辑
	if username == "" || email == "" || password == "" {
		ctx.JSON(http.StatusBadRequest, utils.H{
			"success": false,
			"message": "请填写完整信息",
			"data": nil,
		})
		return
	}

	ctx.JSON(http.StatusOK, utils.H{
		"success": true,
		"message": "注册成功",
		"data": utils.H{
			"user_id": 2,
			"username": username,
		},
	})
}

func forgotPasswordHandler(c context.Context, ctx *app.RequestContext) {
	email := string(ctx.PostForm("email"))

	ctx.JSON(http.StatusOK, utils.H{
		"success": true,
		"message": "密码重置邮件已发送到 " + email,
		"data": nil,
	})
}

func changePasswordHandler(c context.Context, ctx *app.RequestContext) {
	oldPassword := string(ctx.PostForm("old_password"))
	newPassword := string(ctx.PostForm("new_password"))

	// 模拟密码修改逻辑
	_ = oldPassword
	_ = newPassword

	ctx.JSON(http.StatusOK, utils.H{
		"success": true,
		"message": "密码修改成功",
		"data": nil,
	})
}

func deleteAccountHandler(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, utils.H{
		"success": true,
		"message": "账户已删除",
		"data": nil,
	})
}

func publicInfoHandler(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, utils.H{
		"success": true,
		"message": "公共信息",
		"data": utils.H{
			"version": "1.0.0",
			"description": "这是一个公共接口，不需要验证码",
		},
	})
}

func healthCheckHandler(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, utils.H{
		"status": "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

// 高级配置示例
func AdvancedConfigExample() *Generator {
	// 高性能配置
	config := &Config{
		Width:   200,  // 更大的图片提高识别率
		Height:  60,
		Length:  6,    // 更长的验证码提高安全性
		TTL:     1800, // 30分钟过期时间
		Charset: "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ", // 数字+字母
	}

	// 使用内存存储，适合单机部署
	store := NewMemoryStore()

	return NewGenerator(config, store)
}

// 分布式配置示例（Redis存储）
func DistributedConfigExample() *Generator {
	config := DefaultConfig()

	// Redis存储配置（需要实现Redis存储）
	// redisClient := redis.NewClient(&redis.Options{
	//     Addr: "localhost:6379",
	// })
	// store := NewRedisStore(redisClient, "captcha:", time.Minute*10)

	// 暂时使用内存存储作为示例
	store := NewMemoryStore()

	return NewGenerator(config, store)
}

// 自定义错误处理示例
func CustomErrorHandlerExample() *MiddlewareConfig {
	return &MiddlewareConfig{
		SkipPaths: []string{"/captcha/", "/public/"},
		ErrorHandler: func(c context.Context, ctx *app.RequestContext, err error) {
			// 记录错误日志
			// log.Printf("验证码验证失败: %v, IP: %s, Path: %s", 
			//     err, ctx.ClientIP(), string(ctx.Path()))

			// 根据错误类型返回不同响应
			if captchaErr, ok := err.(*CaptchaError); ok {
				switch captchaErr.Code {
				case ErrCodeMissingParams:
					ctx.JSON(http.StatusBadRequest, utils.H{
						"error": "MISSING_CAPTCHA",
						"message": "请提供验证码",
					})
				case ErrCodeInvalidCaptcha:
					ctx.JSON(http.StatusBadRequest, utils.H{
						"error": "INVALID_CAPTCHA",
						"message": "验证码错误，请重新输入",
					})
				case ErrCodeExpiredCaptcha:
					ctx.JSON(http.StatusBadRequest, utils.H{
						"error": "EXPIRED_CAPTCHA",
						"message": "验证码已过期，请刷新后重试",
					})
				default:
					ctx.JSON(http.StatusBadRequest, utils.H{
						"error": "CAPTCHA_ERROR",
						"message": captchaErr.Message,
					})
				}
			} else {
				ctx.JSON(http.StatusInternalServerError, utils.H{
					"error": "INTERNAL_ERROR",
					"message": "服务器内部错误",
				})
			}
		},
		SuccessHandler: func(c context.Context, ctx *app.RequestContext) {
			// 验证成功后的处理
			// 可以记录成功日志、更新统计信息等
			// log.Printf("验证码验证成功, IP: %s", ctx.ClientIP())
		},
	}
}
