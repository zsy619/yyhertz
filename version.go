package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc"
	"github.com/zsy619/yyhertz/framework/mvc/middleware"
)

const (
	// 框架版本信息
	FrameworkName    = "Hertz MVC"
	FrameworkVersion = "1.0.0"
	BuildDate        = "2024-07-29"

	// 依赖版本
	HertzVersion = "v0.10.1"
	GoVersion    = "1.24+"

	// 作者信息
	Author     = "YYHertz Team"
	License    = "Apache 2.0"
	Repository = "https://github.com/zsy619/yyhertz"
	Homepage   = "https://yyhertz.hn24365.com"

	// 构建信息
	BuildMode = "release"
)

// VersionInfo 版本信息结构体
type VersionInfo struct {
	Framework    string            `json:"framework"`
	Version      string            `json:"version"`
	BuildDate    string            `json:"build_date"`
	BuildTime    string            `json:"build_time"`
	GoVersion    string            `json:"go_version"`
	Platform     string            `json:"platform"`
	Arch         string            `json:"arch"`
	Dependencies map[string]string `json:"dependencies"`
	Author       string            `json:"author"`
	License      string            `json:"license"`
	Repository   string            `json:"repository"`
	Homepage     string            `json:"homepage"`
}

// GetVersionInfo 获取完整版本信息
func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Framework: FrameworkName,
		Version:   FrameworkVersion,
		BuildDate: BuildDate,
		BuildTime: time.Now().Format("2006-01-02 15:04:05"),
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		Dependencies: map[string]string{
			"hertz": HertzVersion,
			"go":    runtime.Version(),
		},
		Author:     Author,
		License:    License,
		Repository: Repository,
		Homepage:   Homepage,
	}
}

// GetVersionString 获取版本字符串
func GetVersionString() string {
	return fmt.Sprintf("%s %s", FrameworkName, FrameworkVersion)
}

// GetBuildInfo 获取构建信息
func GetBuildInfo() string {
	return fmt.Sprintf("%s %s (built with %s on %s/%s at %s)",
		FrameworkName,
		FrameworkVersion,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		BuildDate,
	)
}

// PrintVersion 打印版本信息
func PrintVersion() {
	info := GetVersionInfo()
	fmt.Printf("🚀 %s Framework\n", info.Framework)
	fmt.Printf("📦 Version: %s\n", info.Version)
	fmt.Printf("🗓️  Build Date: %s\n", info.BuildDate)
	fmt.Printf("🔧 Go Version: %s\n", info.GoVersion)
	fmt.Printf("💻 Platform: %s/%s\n", info.Platform, info.Arch)
	fmt.Printf("⚡ Powered by CloudWeGo-Hertz %s\n", HertzVersion)
	fmt.Printf("👥 Author: %s\n", info.Author)
	fmt.Printf("📄 License: %s\n", info.License)
	fmt.Printf("🌐 Homepage: %s\n", info.Homepage)
	fmt.Printf("📚 Repository: %s\n", info.Repository)
}

// PrintBanner 打印启动横幅
func PrintBanner() {
	fmt.Println(`
██   ██ ███████ ██████  ████████ ███████     ███    ███ ██    ██  ██████ 
██   ██ ██      ██   ██    ██    ██          ████  ████ ██    ██ ██      
███████ █████   ██████     ██    ███████     ██ ████ ██ ██    ██ ██      
██   ██ ██      ██   ██    ██         ██     ██  ██  ██  ██  ██  ██      
██   ██ ███████ ██   ██    ██    ███████     ██      ██   ████    ██████`)
	fmt.Printf("                    %s Framework v%s\n", FrameworkName, FrameworkVersion)
	fmt.Printf("                基于CloudWeGo-Hertz的类Beego框架\n")
	fmt.Printf("                    Build: %s | %s\n", BuildDate, runtime.Version())
}

// GetFeatures 获取框架特性列表
func GetFeatures() []string {
	return []string{
		"🎯 基于Controller的架构设计",
		"⚡ 高性能HTTP服务器(基于Hertz)",
		"🔄 自动路由注册机制",
		"🛡️  丰富的中间件支持",
		"📊 RESTful API设计",
		"🔧 简化的参数绑定",
		"📝 生命周期钩子方法",
		"🎨 JSON响应格式化",
		"📋 请求日志记录",
		"🌐 CORS跨域支持",
	}
}

// GetSystemInfo 获取系统运行信息
func GetSystemInfo() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"go_version":      runtime.Version(),
		"go_os":           runtime.GOOS,
		"go_arch":         runtime.GOARCH,
		"cpu_count":       runtime.NumCPU(),
		"goroutine_count": runtime.NumGoroutine(),
		"memory_usage": map[string]any{
			"alloc_mb":       bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":         bToMb(m.Sys),
			"num_gc":         m.NumGC,
		},
		"framework": map[string]string{
			"name":    FrameworkName,
			"version": FrameworkVersion,
			"hertz":   HertzVersion,
		},
	}
}

// bToMb 字节转MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// IsDebugMode 检查是否为调试模式
func IsDebugMode() bool {
	return BuildMode == "debug"
}

// GetHealthStatus 获取健康状态
func GetHealthStatus() map[string]any {
	return map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"uptime":    time.Since(time.Now()).String(), // 这里实际应用中需要记录启动时间
		"version":   FrameworkVersion,
		"framework": FrameworkName,
	}
}

// CheckDependencies 检查依赖版本兼容性
func CheckDependencies() bool {
	// 检查Go版本
	goVer := runtime.Version()
	if goVer < "go1.18" {
		fmt.Printf("⚠️  Warning: Go version %s may not be fully supported. Recommend Go 1.18+\n", goVer)
		return false
	}

	fmt.Printf("✅ Go version %s is supported\n", goVer)
	return true
}

// =============== 控制器定义 ===============

// 系统控制器 - 版本和健康检查接口
type SystemController struct {
	mvc.BaseController
}

func (c *SystemController) GetVersion() {
	config.Info("获取版本信息请求")
	c.JSON(GetVersionInfo())
}

func (c *SystemController) GetHealth() {
	config.Info("健康检查请求")
	c.JSON(GetHealthStatus())
}

func (c *SystemController) GetInfo() {
	config.Info("获取系统信息请求")
	c.JSON(map[string]any{
		"version":  GetVersionInfo(),
		"system":   GetSystemInfo(),
		"features": GetFeatures(),
	})
}

// 用户控制器
type UserController struct {
	mvc.BaseController
}

type User struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func (c *UserController) GetIndex() {
	config.Info("获取用户列表请求")
	users := []User{
		{ID: 1, Name: "张三", Email: "zhangsan@example.com", CreatedAt: "2024-01-15"},
		{ID: 2, Name: "李四", Email: "lisi@example.com", CreatedAt: "2024-02-20"},
		{ID: 3, Name: "王五", Email: "wangwu@example.com", CreatedAt: "2024-03-10"},
	}

	config.WithFields(map[string]any{
		"user_count": len(users),
	}).Info("用户列表获取成功")

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

	config.WithFields(map[string]any{
		"user_id": userId,
		"name":    name,
	}).Info("获取用户信息请求")

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

	config.WithFields(map[string]any{
		"name":  name,
		"email": email,
	}).Info("创建用户请求")

	if name == "" || email == "" {
		config.Warn("用户创建失败：用户名和邮箱不能为空")
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

	config.WithFields(map[string]any{
		"user_id": user.ID,
		"name":    user.Name,
		"email":   user.Email,
	}).Info("用户创建成功")

	c.JSON(map[string]any{
		"success": true,
		"message": "用户创建成功",
		"data":    user,
	})
}

// 首页控制器
type HomeController struct {
	mvc.BaseController
}

func (c *HomeController) GetIndex() {
	config.Info("获取首页信息请求")
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

// =============== 中间件定义 ===============

// 日志中间件
func LoggerMiddleware() mvc.HandlerFunc {
	return func(c context.Context, ctx *mvc.RequestContext) {
		start := time.Now()
		method := string(ctx.Method())
		path := string(ctx.Path())
		clientIP := ctx.ClientIP()

		// 使用单例日志系统记录请求开始
		config.WithFields(map[string]any{
			"method":    method,
			"path":      path,
			"client_ip": clientIP,
			"time":      time.Now().Format("15:04:05"),
		}).Info("📝 HTTP Request Start")

		ctx.Next(c)

		latency := time.Since(start)
		status := ctx.Response.StatusCode()

		// 使用单例日志系统记录请求完成
		config.WithFields(map[string]any{
			"method":   method,
			"path":     path,
			"status":   status,
			"latency":  latency,
			"duration": latency.String(),
		}).Info("✅ HTTP Request Complete")
	}
}

// =============== 主函数 ===============

func main() {
	var (
		showVersion  = flag.Bool("version", false, "显示版本信息")
		showBanner   = flag.Bool("banner", true, "显示启动横幅")
		port         = flag.String("port", "", "服务器端口")
		enableHTTPS  = flag.Bool("https", false, "启用HTTPS")
		certFile     = flag.String("cert", "", "TLS证书文件路径")
		keyFile      = flag.String("key", "", "TLS私钥文件路径")
		requireHTTPS = flag.Bool("require-https", false, "强制要求HTTPS连接")
		// configFile   = flag.String("config", "", "配置文件路径")
	)
	flag.Parse()

	// 显示版本信息并退出
	if *showVersion {
		PrintVersion()
		return
	}

	// 初始化配置管理器
	configManager := config.GetAppConfigManager()

	if err := configManager.Initialize(); err != nil {
		log.Printf("配置初始化失败: %v", err)
		return
	}

	// 启用配置文件监听
	config.WatchConfig(config.AppConfig{})

	// 获取应用配置
	appConfig, err := configManager.GetConfig()
	if err != nil {
		config.WithFields(map[string]any{
			"error": err.Error(),
		}).Fatal("获取配置失败")
	}

	// 命令行参数优先级高于配置文件
	if *port == "" {
		*port = fmt.Sprintf("%d", appConfig.App.Port)
	}
	if !*enableHTTPS && appConfig.TLS.Basic.Enable {
		*enableHTTPS = appConfig.TLS.Basic.Enable
		if *certFile == "" {
			*certFile = appConfig.TLS.Certificate.CertFile
		}
		if *keyFile == "" {
			*keyFile = appConfig.TLS.Certificate.KeyFile
		}
	}

	// 显示启动横幅
	if *showBanner {
		PrintBanner()
	}

	log.Printf("应用配置加载完成: config_file=%s, app_name=%s, environment=%s, debug_mode=%v",
		configManager.ConfigFileUsed(),
		appConfig.App.Name,
		appConfig.App.Environment,
		appConfig.App.Debug)

	// 创建应用实例
	app := mvc.NewApp()

	// 添加中间件
	// TLS安全中间件
	tlsConfig := middleware.DefaultTLSConfig()
	tlsConfig.Enable = *enableHTTPS
	tlsConfig.CertFile = *certFile
	tlsConfig.KeyFile = *keyFile
	tlsConfig.RequireHTTPS = *requireHTTPS
	tlsConfig.HSTSEnabled = true // 启用HSTS

	// 从配置文件合并TLS设置
	if appConfig.TLS.Basic.Enable {
		tlsConfig.Enable = appConfig.TLS.Basic.Enable
		if tlsConfig.CertFile == "" {
			tlsConfig.CertFile = appConfig.TLS.Certificate.CertFile
		}
		if tlsConfig.KeyFile == "" {
			tlsConfig.KeyFile = appConfig.TLS.Certificate.KeyFile
		}
	}

	// 验证TLS配置
	if err := middleware.ValidateTLSConfig(tlsConfig); err != nil {
		config.WithFields(map[string]any{
			"error": err.Error(),
		}).Fatal("TLS配置验证失败")
	}

	app.Use(middleware.TLSSupportMiddleware(tlsConfig))

	// 日志中间件
	app.Use(LoggerMiddleware())

	// 注册控制器
	userController := &UserController{}
	homeController := &HomeController{}
	systemController := &SystemController{}

	app.AutoRouterPrefix("/user", userController)
	app.AutoRouterPrefix("/home", homeController)
	app.AutoRouterPrefix("/system", systemController)

	// 首页路由
	app.GET("/", mvc.HandlerFunc(func(ctx context.Context, c *mvc.RequestContext) {
		homeCtrl := &HomeController{}
		// homeCtrl.Ctx = &ctx  // 暂时注释这行避免类型错误
		homeCtrl.Data = make(map[string]any)
		homeCtrl.GetIndex()
	}))

	// 使用单例日志系统记录服务器启动信息
	config.WithFields(map[string]any{
		"framework":     GetVersionString(),
		"port":          *port,
		"https_enabled": *enableHTTPS,
		"require_https": *requireHTTPS,
		"time":          time.Now().Format("2006-01-02 15:04:05"),
	}).Info("🚀 服务器启动成功")

	// 显示服务器地址（根据HTTPS状态）
	protocol := "http"
	if *enableHTTPS {
		protocol = "https"
	}
	config.Infof("📍 服务器地址: %s://localhost:%s", protocol, *port)
	config.Infof("🕐 启动时间: %s", time.Now().Format("2006-01-02 15:04:05"))

	// 显示TLS状态
	if *enableHTTPS {
		config.WithFields(map[string]any{
			"cert_file":     *certFile,
			"key_file":      *keyFile,
			"require_https": *requireHTTPS,
		}).Info("🔒 HTTPS已启用")
	}

	config.Info("📋 可用路由:")
	config.Info("系统接口:")
	config.Info("  GET    /                 - 首页")
	config.Info("  GET    /api              - API文档")
	config.Info("  GET    /system/version   - 版本信息")
	config.Info("  GET    /system/health    - 健康检查")
	config.Info("  GET    /system/info      - 系统信息")
	config.Info("业务接口:")
	config.Info("  GET    /home/index       - 首页信息")
	config.Info("  GET    /user/index       - 用户列表")
	config.Info("  GET    /user/info        - 用户信息")
	config.Info("  POST   /user/create      - 创建用户")

	config.Info("💡 测试命令:")
	config.Infof("curl http://localhost:%s/", *port)
	config.Infof("curl http://localhost:%s/system/version", *port)
	config.Infof("curl http://localhost:%s/user/index", *port)
	config.Infof("curl -X POST http://localhost:%s/user/create -d 'name=张三&email=test@example.com'", *port)

	app.Spin()
}

// init 初始化函数
func init() {
	// 可以在这里进行一些初始化检查
	if !CheckDependencies() {
		fmt.Println("⚠️  Some dependencies may not be compatible")
	}
}
