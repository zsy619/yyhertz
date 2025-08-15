package core

import (
	"strings"
)

// ============= 静态文件处理方法 =============

// SetStaticPath 设置静态文件路径
// 参数：localDir - 静态文件本地目录（相对应用所在目录）
//
//	urlPath - URL路径（可选），如果不提供则自动推导
//
// 示例：SetStaticPath("public", "/static") 或 SetStaticPath("public")
func (app *App) SetStaticPath(localDir string, urlPath ...string) {
	if app.StaticPaths == nil {
		app.StaticPaths = make(map[string]string)
	}

	// 确定URL路径
	var targetUrlPath string
	if len(urlPath) > 0 && urlPath[0] != "" {
		targetUrlPath = urlPath[0]
	} else {
		// 自动推导：移除 "./" 前缀，确保以 "/" 开头
		cleanDir := strings.TrimLeft(strings.TrimPrefix(localDir, "./"), "/")
		if cleanDir == "" {
			targetUrlPath = "/static" // 默认URL路径
		} else {
			targetUrlPath = "/" + cleanDir
		}
	}

	// 确保URL路径以/开头
	if !strings.HasPrefix(targetUrlPath, "/") {
		targetUrlPath = "/" + targetUrlPath
	}

	// 只有当路径不存在或者发生变化时才注册
	if existing, exists := app.StaticPaths[targetUrlPath]; !exists || existing != localDir {
		app.StaticPaths[targetUrlPath] = localDir
		// 注册静态文件路由
		app.Static(targetUrlPath, localDir)
	}
}

// SetStaticPaths 设置多个静态文件路径映射
func (app *App) SetStaticPaths(pathMap map[string]string) {
	app.StaticPaths = make(map[string]string)
	for localPath, urlPath := range pathMap {
		app.SetStaticPath(localPath, urlPath)
	}
}

// AddStaticPath 添加单个静态路径映射
func (app *App) AddStaticPath(localPath, urlPath string) {
	app.SetStaticPath(localPath, urlPath)
}

// AddStaticPaths 添加多个静态路径映射
func (app *App) AddStaticPaths(pathMap map[string]string) {
	app.SetStaticPaths(pathMap)
}

// GetStaticPath 获取默认静态文件路径（向后兼容）
func (app *App) GetStaticPath() string {
	if path, exists := app.StaticPaths["/static"]; exists {
		return path
	}
	// 返回第一个路径作为默认值
	for _, path := range app.StaticPaths {
		return path
	}
	return "./static"
}

// GetStaticPaths 获取所有静态路径映射
func (app *App) GetStaticPaths() map[string]string {
	return app.StaticPaths
}