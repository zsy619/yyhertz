package captcha_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/zsy619/yyhertz/framework/mvc/captcha"
)

// Example 展示如何在Hertz应用中使用验证码
func Example() {
	// 1. 创建验证码配置
	config := captcha.DefaultConfig()
	config.Width = 120
	config.Height = 40
	config.Length = 4
	config.TTL = 300 // 5分钟过期
	
	// 2. 创建存储
	store := captcha.NewMemoryStore()
	defer store.Close()
	
	// 3. 创建验证码生成器
	generator := captcha.NewGenerator(config, store)
	
	// 4. 创建Hertz应用
	h := server.Default()
	
	// 5. 注册路由
	
	// 生成验证码
	h.GET("/captcha/generate", captcha.GenerateHandler(generator))
	
	// 获取验证码图片
	h.GET("/captcha/image/:id", captcha.ImageHandler(generator))
	
	// 验证验证码
	h.POST("/captcha/verify", captcha.VerifyHandler(generator))
	
	// 需要验证码保护的路由
	middlewareConfig := &captcha.MiddlewareConfig{
		SkipPaths: []string{"/captcha/", "/public/"},
	}
	captchaMiddleware := captcha.NewMiddleware(generator, middlewareConfig)
	
	// 应用中间件到需要保护的路由组
	protected := h.Group("/api", captchaMiddleware.Handler())
	{
		protected.POST("/login", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(http.StatusOK, utils.H{
				"message": "登录成功",
			})
		})
		
		protected.POST("/register", func(c context.Context, ctx *app.RequestContext) {
			ctx.JSON(http.StatusOK, utils.H{
				"message": "注册成功",
			})
		})
	}
	
	// 启动服务器
	log.Println("验证码服务器启动在 :8080")
	h.Spin()
}

// TestCaptchaGeneration 测试验证码生成
func TestCaptchaGeneration(t *testing.T) {
	config := captcha.DefaultConfig()
	store := captcha.NewMemoryStore()
	defer store.Close()
	
	generator := captcha.NewGenerator(config, store)
	
	// 生成验证码
	cap, err := generator.Generate()
	if err != nil {
		t.Fatalf("生成验证码失败: %v", err)
	}
	
	// 验证生成的验证码
	if cap.ID == "" {
		t.Error("验证码ID不能为空")
	}
	
	if cap.Code == "" {
		t.Error("验证码不能为空")
	}
	
	if len(cap.Image) == 0 {
		t.Error("验证码图片不能为空")
	}
	
	fmt.Printf("生成验证码成功: ID=%s, Code=%s\n", cap.ID, cap.Code)
}

// TestCaptchaVerification 测试验证码验证
func TestCaptchaVerification(t *testing.T) {
	config := captcha.DefaultConfig()
	store := captcha.NewMemoryStore()
	defer store.Close()
	
	generator := captcha.NewGenerator(config, store)
	
	// 生成验证码
	cap, err := generator.Generate()
	if err != nil {
		t.Fatalf("生成验证码失败: %v", err)
	}
	
	// 正确验证
	if !generator.Verify(cap.ID, cap.Code) {
		t.Error("正确的验证码验证失败")
	}
	
	// 错误验证（验证码已被消费）
	if generator.Verify(cap.ID, cap.Code) {
		t.Error("验证码应该在验证后被删除")
	}
	
	// 生成新的验证码用于错误验证测试
	cap2, _ := generator.Generate()
	
	// 错误的验证码
	if generator.Verify(cap2.ID, "wrong") {
		t.Error("错误的验证码不应该验证成功")
	}
	
	// 错误的ID
	if generator.Verify("wrong_id", cap2.Code) {
		t.Error("错误的ID不应该验证成功")
	}
}

// TestCaptchaStore 测试不同的存储实现
func TestCaptchaStore(t *testing.T) {
	config := captcha.DefaultConfig()
	
	// 测试内存存储
	t.Run("MemoryStore", func(t *testing.T) {
		store := captcha.NewMemoryStore()
		defer store.Close()
		
		generator := captcha.NewGenerator(config, store)
		testStoreBasicOperations(t, generator)
	})
	
	// 测试Session存储（模拟）
	t.Run("SessionStore", func(t *testing.T) {
		sessionData := make(map[string]*captcha.Captcha)
		
		store := captcha.NewSessionStore(
			"captcha_data",
			func() interface{} { return sessionData },
			func(key string, value interface{}) error {
				sessionData = value.(map[string]*captcha.Captcha)
				return nil
			},
		)
		
		generator := captcha.NewGenerator(config, store)
		testStoreBasicOperations(t, generator)
	})
}

// testStoreBasicOperations 测试存储的基本操作
func testStoreBasicOperations(t *testing.T, generator *captcha.Generator) {
	// 生成验证码
	cap, err := generator.Generate()
	if err != nil {
		t.Fatalf("生成验证码失败: %v", err)
	}
	
	// 获取图片
	image, err := generator.GetImage(cap.ID)
	if err != nil {
		t.Fatalf("获取验证码图片失败: %v", err)
	}
	
	if len(image) == 0 {
		t.Error("验证码图片不能为空")
	}
	
	// 验证
	if !generator.Verify(cap.ID, cap.Code) {
		t.Error("验证码验证失败")
	}
}

// BenchmarkCaptchaGeneration 性能测试：验证码生成
func BenchmarkCaptchaGeneration(b *testing.B) {
	config := captcha.DefaultConfig()
	store := captcha.NewMemoryStore()
	defer store.Close()
	
	generator := captcha.NewGenerator(config, store)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.Generate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCaptchaVerification 性能测试：验证码验证
func BenchmarkCaptchaVerification(b *testing.B) {
	config := captcha.DefaultConfig()
	store := captcha.NewMemoryStore()
	defer store.Close()
	
	generator := captcha.NewGenerator(config, store)
	
	// 预生成一些验证码
	captchas := make([]*captcha.Captcha, 1000)
	for i := 0; i < 1000; i++ {
		cap, _ := generator.Generate()
		captchas[i] = cap
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cap := captchas[i%1000]
		generator.Verify(cap.ID, cap.Code)
	}
}
