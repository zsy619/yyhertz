package controller

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hertzlogrus "github.com/hertz-contrib/logger/logrus"
)

// 类型别名定义
type RequestContext = app.RequestContext
type HandlerFunc = func(context.Context, *RequestContext)

// 控制器接口定义
type IController interface {
	Init()
	Prepare()
	Finish()
}

// App 应用结构
type App struct {
	*server.Hertz
	ViewPath   string
	StaticPath string
	startTime  time.Time
	address    string
}

// NewApp 创建新的应用实例
func NewApp() *App {
	h := server.Default()

	logger := hertzlogrus.NewLogger()
	hlog.SetLogger(logger)

	return &App{
		Hertz:      h,
		ViewPath:   "views",
		StaticPath: "static",
		startTime:  time.Now(),
		address:    ":8080",
	}
}

// SetViewPath 设置视图路径
func (app *App) SetViewPath(path string) {
	app.ViewPath = path
}

// SetStaticPath 设置静态文件路径
func (app *App) SetStaticPath(path string) {
	app.StaticPath = path
	app.Static("/static", path)
}

// Use 添加中间件
func (app *App) Use(middleware ...HandlerFunc) {
	for _, m := range middleware {
		app.Hertz.Use(m)
	}
}

// Run 启动服务器
func (app *App) Run(addr ...string) {
	if len(addr) > 0 {
		app.address = addr[0]
	}
	
	// 启动服务器，忽略地址参数，使用默认配置
	app.Hertz.Spin()
}
