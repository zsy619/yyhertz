package main

import (
	"fmt"
	"runtime"
	"time"
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
	Author      = "CloudWeGo Team"
	License     = "Apache 2.0"
	Repository  = "https://github.com/cloudwego/hertz"
	Homepage    = "https://www.cloudwego.io/zh/docs/hertz/"
	
	// 构建信息
	BuildMode = "release"
)

// VersionInfo 版本信息结构体
type VersionInfo struct {
	Framework   string            `json:"framework"`
	Version     string            `json:"version"`
	BuildDate   string            `json:"build_date"`
	BuildTime   string            `json:"build_time"`
	GoVersion   string            `json:"go_version"`
	Platform    string            `json:"platform"`
	Arch        string            `json:"arch"`
	Dependencies map[string]string `json:"dependencies"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Repository  string            `json:"repository"`
	Homepage    string            `json:"homepage"`
}

// GetVersionInfo 获取完整版本信息
func GetVersionInfo() *VersionInfo {
	return &VersionInfo{
		Framework:  FrameworkName,
		Version:    FrameworkVersion,
		BuildDate:  BuildDate,
		BuildTime:  time.Now().Format("2006-01-02 15:04:05"),
		GoVersion:  runtime.Version(),
		Platform:   runtime.GOOS,
		Arch:       runtime.GOARCH,
		Dependencies: map[string]string{
			"hertz":   HertzVersion,
			"go":      runtime.Version(),
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
██   ██ ███████ ██   ██    ██    ███████     ██      ██   ████    ██████ 
`)
	fmt.Printf("                    %s Framework v%s\n", FrameworkName, FrameworkVersion)
	fmt.Printf("                基于CloudWeGo-Hertz的类Beego框架\n")
	fmt.Printf("                    Build: %s | %s\n", BuildDate, runtime.Version())
	fmt.Println()
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
		"go_version":     runtime.Version(),
		"go_os":          runtime.GOOS,
		"go_arch":        runtime.GOARCH,
		"cpu_count":      runtime.NumCPU(),
		"goroutine_count": runtime.NumGoroutine(),
		"memory_usage": map[string]any{
			"alloc_mb":      bToMb(m.Alloc),
			"total_alloc_mb": bToMb(m.TotalAlloc),
			"sys_mb":        bToMb(m.Sys),
			"num_gc":        m.NumGC,
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

// init 初始化函数
func init() {
	// 可以在这里进行一些初始化检查
	if !CheckDependencies() {
		fmt.Println("⚠️  Some dependencies may not be compatible")
	}
}