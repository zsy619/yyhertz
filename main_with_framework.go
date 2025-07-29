package main

import (
	"context"
	"flag"
	"log"
	"time"

	"hertz-controller/framework/controller"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// 系统控制器 - 版本和健康检查接口
type SystemController struct {
	controller.BaseController
}

func (c *SystemController) GetVersion() {
	c.JSON(GetVersionInfo())
}

func (c *SystemController) GetHealth() {
	c.JSON(GetHealthStatus())
}

func (c *SystemController) GetInfo() {
	c.JSON(map[string]any{
		"version":  GetVersionInfo(),
		"system":   GetSystemInfo(),
		"features": GetFeatures(),
	})
}

// 用户控制器
type UserController struct {
	controller.BaseController
}

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (c *UserController) GetIndex() {
	users := []User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", CreatedAt: "2024-01-15"},
		{ID: 2, Name: "李四", Email: "lisi@example.com", CreatedAt: "2024-02-20"},
		{ID: 3, Name: "王五", Email: "wangwu@example.com", CreatedAt: "2024-03-10"},
	}
	
	c.JSON(map[string]any{
		"success": true,
		"message": "用户列表获取成功",
		"data":    users,
		"total":   len(users),
	})
}

func (c *UserController) GetInfo() {
	userId := c.GetString("id", "1")
	name := c.GetString("name", "默认用户")
	
	user := User{
		ID:        1,
		Name:      name,
		Email:     "user@example.com",
		CreatedAt: "2024-01-15",
	}
	
	c.JSON(map[string]any{
		"success":  true,
		"message":  "用户信息获取成功",
		"data":     user,
		"query_id": userId,
	})
}

func (c *UserController) PostCreate() {
	name := c.GetForm("name")
	email := c.GetForm("email")
	
	if name == "" || email == "" {
		c.JSON(map[string]any{
			"success": false,
			"message": "用户名和邮箱不能为空",
		})
		return
	}
	
	user := User{
		ID:        4,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now().Format("2006-01-02"),
	}
	
	c.JSON(map[string]any{
		"success": true,
		"message": "用户创建成功",
		"data":    user,
	})
}

// 首页控制器
type HomeController struct {
	controller.BaseController
}

func (c *HomeController) GetIndex() {
	c.JSON(map[string]any{
		"message":    "🚀 欢迎使用Hertz MVC框架！",
		"framework":  FrameworkName,
		"version":    FrameworkVersion,
		"powered_by": "CloudWeGo-Hertz " + HertzVersion,
		"features":   GetFeatures(),
		"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		"build_info": GetBuildInfo(),
	})
}

// 日志中间件
func LoggerMiddleware() controller.HandlerFunc {
	return func(c context.Context, ctx *controller.RequestContext) {
		start := time.Now()
		method := string(ctx.Method())
		path := string(ctx.Path())
		clientIP := ctx.ClientIP()
		
		log.Printf("📝 [%s] %s %s %s", time.Now().Format("15:04:05"), clientIP, method, path)
		
		ctx.Next(c)
		
		latency := time.Since(start)
		status := ctx.Response.StatusCode()
		log.Printf("✅ %s %s - %d - %v", method, path, status, latency)
	}
}

func main() {
	var (
		showVersion = flag.Bool("version", false, "显示版本信息")
		showBanner  = flag.Bool("banner", true, "显示启动横幅")
		port        = flag.String("port", "8888", "服务器端口")
	)
	flag.Parse()
	
	// 显示版本信息并退出
	if *showVersion {
		PrintVersion()
		return
	}
	
	// 显示启动横幅
	if *showBanner {
		PrintBanner()
	}
	
	// 创建应用实例
	app := controller.NewApp()

	// 添加中间件
	app.Use(LoggerMiddleware())

	// 注册控制器
	userController := &UserController{}
	homeController := &HomeController{}
	systemController := &SystemController{}
	
	app.RegisterController("/user", userController)
	app.RegisterController("/home", homeController)
	app.RegisterController("/system", systemController)

	// 首页路由
	app.GET("/", controller.HandlerFunc(func(ctx context.Context, c *controller.RequestContext) {
		homeCtrl := &HomeController{}
		homeCtrl.Ctx = c
		homeCtrl.Data = make(map[string]any)
		homeCtrl.GetIndex()
	}))

	// API文档路由
	app.GET("/api", controller.HandlerFunc(func(ctx context.Context, c *controller.RequestContext) {
		c.JSON(consts.StatusOK, map[string]any{
			"title":   "Hertz MVC API 文档",
			"version": GetVersionString(),
			"build":   GetBuildInfo(),
			"endpoints": map[string]any{
				"系统接口": map[string]string{
					"GET /":               "首页信息",
					"GET /api":            "API文档",
					"GET /system/version": "版本信息",
					"GET /system/health":  "健康检查",
					"GET /system/info":    "系统信息",
				},
				"业务接口": map[string]string{
					"GET /home/index":   "首页",
					"GET /user/index":   "用户列表",
					"GET /user/info":    "用户详情 (参数: id, name)",
					"POST /user/create": "创建用户 (参数: name, email)",
				},
			},
		})
	}))

	log.Printf("🚀 %s 启动成功!", GetVersionString())
	log.Printf("📍 服务器地址: http://localhost:%s", *port)
	log.Printf("🕐 启动时间: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Println("")
	log.Println("📋 可用路由:")
	log.Println("系统接口:")
	log.Println("  GET    /                 - 首页")
	log.Println("  GET    /api              - API文档")
	log.Println("  GET    /system/version   - 版本信息")
	log.Println("  GET    /system/health    - 健康检查")
	log.Println("  GET    /system/info      - 系统信息")
	log.Println("业务接口:")
	log.Println("  GET    /home/index       - 首页信息")
	log.Println("  GET    /user/index       - 用户列表")
	log.Println("  GET    /user/info        - 用户信息")
	log.Println("  POST   /user/create      - 创建用户")
	log.Println("")
	log.Println("💡 测试命令:")
	log.Printf("curl http://localhost:%s/\n", *port)
	log.Printf("curl http://localhost:%s/system/version\n", *port)
	log.Printf("curl http://localhost:%s/user/index\n", *port)
	log.Printf("curl -X POST http://localhost:%s/user/create -d 'name=张三&email=test@example.com'\n", *port)

	app.Spin()
}