package mvc

// 重新导出核心功能，保持向后兼容
import (
	"path"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/annotation"
	"github.com/zsy619/yyhertz/framework/mvc/comment"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/core"
	"github.com/zsy619/yyhertz/framework/mvc/router"
	"github.com/zsy619/yyhertz/framework/mvc/session"
	"github.com/zsy619/yyhertz/framework/util"
)

// 类型别名，保持向后兼容
type App = core.App
type RequestContext = core.RequestContext
type HandlerFunc = core.HandlerFunc
type IController = core.IController

// 重新导出常用功能
var (
	once sync.Once

	mutex sync.RWMutex

	GetAppInstance      = core.GetAppInstance
	NewApp              = core.NewApp
	NewAppWithLogConfig = core.NewAppWithLogConfig
	AdaptHandler        = core.AdaptHandler

	HertzApp *App

	IsInitComplete = false // 是否完成初始化
)

// Session相关类型别名
type SessionConfig = session.Config
type SessionManager = session.Manager
type SessionStore = session.Store

var (
	DefaultSessionConfig = session.DefaultConfig
	NewSessionManager    = session.NewManager
)

// Cookie相关类型别名
type CookieConfig = cookie.Config
type CookieOptions = cookie.Options
type CookieHelper = cookie.Helper

var (
	DefaultCookieConfig  = cookie.DefaultConfig
	DefaultCookieOptions = cookie.DefaultOptions
	NewCookieHelper      = cookie.NewHelper
)

// Router相关类型别名
type Router = router.Router
type RouterGroup = router.Group

var (
	NewRouter = router.NewRouter
	NewGroup  = router.NewGroup
)

func init() {
	// 请帮忙实现如下需求：
	// 1. 判断根目录下是否存在 conf/app.yaml 文件，如果不存在按照默认配置生成
	// 2. 判断根目录下是否存在 conf/log.yaml 文件，如果不存在按照默认配置生成
	// 3. 如果根目录下是否存在 conf/template.yaml 文件，如果不存在则加载默认配置生成
	appConf := path.Join(".", "conf", "app.yaml")
	// 判断文件是否存在
	if isExists := util.FileExists(appConf); !isExists {
		// 文件不存在，生成默认配置
		appConfig := config.AppConfig{}
		cm := config.NewViperConfigManager(appConfig)
		_ = cm.Initialize()

		config.WatchConfig(appConfig)
	}
	templateConf := path.Join(".", "conf", "template.yaml")
	if isExists := util.FileExists(templateConf); !isExists {
		// 文件不存在，生成默认配置
		templateConfig := config.TemplateConfig{}
		cm := config.NewViperConfigManager(templateConfig)
		_ = cm.Initialize()

		config.WatchConfig(templateConfig)
	}
	authConf := path.Join(".", "conf", "auth.yaml")
	if isExists := util.FileExists(authConf); !isExists {
		// 文件不存在，生成默认配置
		authConfig := config.AuthConfig{}
		cm := config.NewViperConfigManager(authConfig)
		_ = cm.Initialize()

		config.WatchConfig(authConfig)
	}
	// 初始化全局Hertz应用实例
	once.Do(func() {
		mutex.Lock()
		defer mutex.Unlock()

		// 创建全局Hertz应用实例
		HertzApp = GetAppInstance()

		// 创建注解应用
		AnnotationApp = annotation.NewAnnotationWithApp(HertzApp)

		// 创建注释注解应用
		CommentApp = comment.NewCommentWithApp(HertzApp)

		// 注释注解应用
		IsInitComplete = true
	})
}
