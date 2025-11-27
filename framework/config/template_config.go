package config

import (
	"sync"

	"github.com/spf13/viper"
)

var (
	// 全局模板实例
	GlobalTemplate *TemplateConfig
	// 初始化锁
	templateOnce sync.Once
	// 配置文件是否已加载
	templateConfigLoaded bool
	// 配置文件加载锁
	templateConfigOnce sync.Once
)

// ThemeConfig 主题配置
type ThemeConfig struct {
	Name          string            `json:"name" yaml:"name" mapstructure:"name"`
	ViewPaths     []string          `json:"view_paths" yaml:"view_paths" mapstructure:"view_paths"`
	LayoutPath    string            `json:"layout_path" yaml:"layout_path" mapstructure:"layout_path"`
	ComponentPath string            `json:"component_path" yaml:"component_path" mapstructure:"component_path"`
	StaticPath    string            `json:"static_path" yaml:"static_path" mapstructure:"static_path"`
	Enabled       bool              `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	Default       bool              `json:"default" yaml:"default" mapstructure:"default"`
	Variables     map[string]string `json:"variables" yaml:"variables" mapstructure:"variables"`
}

// TemplateConfig 模板引擎配置结构
type TemplateConfig struct {
	// 基础路径配置
	Paths struct {
		ViewPaths     []string          `json:"view_paths" yaml:"view_paths" mapstructure:"view_paths"`             // 视图文件搜索路径
		LayoutPath    string            `json:"layout_path" yaml:"layout_path" mapstructure:"layout_path"`          // 布局文件路径
		ComponentPath string            `json:"component_path" yaml:"component_path" mapstructure:"component_path"` // 组件文件路径
		Extension     string            `json:"extension" yaml:"extension" mapstructure:"extension"`                // 模板文件扩展名
		StaticPaths   map[string]string `json:"static_paths" yaml:"static_paths" mapstructure:"static_paths"`       // URL路径 -> 本地路径映射
	} `json:"paths" yaml:"paths" mapstructure:"paths"`

	// 模板语法配置
	Syntax struct {
		DelimLeft  string `json:"delim_left" yaml:"delim_left" mapstructure:"delim_left"`    // 左分隔符
		DelimRight string `json:"delim_right" yaml:"delim_right" mapstructure:"delim_right"` // 右分隔符
	} `json:"syntax" yaml:"syntax" mapstructure:"syntax"`

	// 缓存配置
	Cache struct {
		EnableCache   bool   `json:"enable_cache" yaml:"enable_cache" mapstructure:"enable_cache"`       // 启用模板缓存
		CacheSize     int    `json:"cache_size" yaml:"cache_size" mapstructure:"cache_size"`             // 模板缓存大小限制
		CacheTTL      int    `json:"cache_ttl" yaml:"cache_ttl" mapstructure:"cache_ttl"`                // 缓存过期时间（秒）
		CacheStrategy string `json:"cache_strategy" yaml:"cache_strategy" mapstructure:"cache_strategy"` // 缓存策略: lru, lfu, fifo
	} `json:"cache" yaml:"cache" mapstructure:"cache"`

	// 性能优化配置
	Performance struct {
		Precompile      bool `json:"precompile" yaml:"precompile" mapstructure:"precompile"`                   // 预编译模板
		CompileParallel bool `json:"compile_parallel" yaml:"compile_parallel" mapstructure:"compile_parallel"` // 并行编译
		CompileWorkers  int  `json:"compile_workers" yaml:"compile_workers" mapstructure:"compile_workers"`    // 编译工作线程数
		EnableCompress  bool `json:"enable_compress" yaml:"enable_compress" mapstructure:"enable_compress"`    // 启用压缩
		MinifyHTML      bool `json:"minify_html" yaml:"minify_html" mapstructure:"minify_html"`                // HTML压缩
		MinifyCSS       bool `json:"minify_css" yaml:"minify_css" mapstructure:"minify_css"`                   // CSS压缩
		MinifyJS        bool `json:"minify_js" yaml:"minify_js" mapstructure:"minify_js"`                      // JS压缩
		BundleAssets    bool `json:"bundle_assets" yaml:"bundle_assets" mapstructure:"bundle_assets"`          // 资源打包
	} `json:"performance" yaml:"performance" mapstructure:"performance"`

	// 国际化配置
	I18n struct {
		Enabled          bool     `json:"enabled" yaml:"enabled" mapstructure:"enabled"`                                  // 启用国际化
		DefaultLocale    string   `json:"default_locale" yaml:"default_locale" mapstructure:"default_locale"`             // 默认语言
		SupportedLocales []string `json:"supported_locales" yaml:"supported_locales" mapstructure:"supported_locales"`    // 支持的语言列表
		LocalePath       string   `json:"locale_path" yaml:"locale_path" mapstructure:"locale_path"`                      // 语言文件路径
		FallbackLocale   string   `json:"fallback_locale" yaml:"fallback_locale" mapstructure:"fallback_locale"`          // 回退语言
		DetectFromHeader bool     `json:"detect_from_header" yaml:"detect_from_header" mapstructure:"detect_from_header"` // 从请求头检测语言
	} `json:"i18n" yaml:"i18n" mapstructure:"i18n"`

	// 错误处理配置
	Error struct {
		Templates   map[string]string `json:"templates" yaml:"templates" mapstructure:"templates"`          // 错误模板映射
		ShowDetails bool              `json:"show_details" yaml:"show_details" mapstructure:"show_details"` // 显示错误详情
		LogLevel    string            `json:"log_level" yaml:"log_level" mapstructure:"log_level"`          // 错误日志级别
	} `json:"error" yaml:"error" mapstructure:"error"`

	// 开发调试配置
	Debug struct {
		Mode             bool `json:"mode" yaml:"mode" mapstructure:"mode"`                                        // 调试模式
		Toolbar          bool `json:"toolbar" yaml:"toolbar" mapstructure:"toolbar"`                               // 调试工具栏
		TemplateComments bool `json:"template_comments" yaml:"template_comments" mapstructure:"template_comments"` // 模板注释
		SourceMap        bool `json:"source_map" yaml:"source_map" mapstructure:"source_map"`                      // 源码映射
		Profiling        bool `json:"profiling" yaml:"profiling" mapstructure:"profiling"`                         // 性能分析
	} `json:"debug" yaml:"debug" mapstructure:"debug"`

	// 资源管理配置
	Assets struct {
		Versioning      bool   `json:"versioning" yaml:"versioning" mapstructure:"versioning"`                   // 资源版本控制
		VersionStrategy string `json:"version_strategy" yaml:"version_strategy" mapstructure:"version_strategy"` // 版本策略: hash, timestamp, manual
		ManifestPath    string `json:"manifest_path" yaml:"manifest_path" mapstructure:"manifest_path"`          // 资源清单文件
		CDNEnabled      bool   `json:"cdn_enabled" yaml:"cdn_enabled" mapstructure:"cdn_enabled"`                // 启用CDN
		CDNUrl          string `json:"cdn_url" yaml:"cdn_url" mapstructure:"cdn_url"`                            // CDN基础URL
		CDNFallback     bool   `json:"cdn_fallback" yaml:"cdn_fallback" mapstructure:"cdn_fallback"`             // CDN回退到本地
	} `json:"assets" yaml:"assets" mapstructure:"assets"`

	// 布局和组件配置
	Layout struct {
		DefaultLayout      string `json:"default_layout" yaml:"default_layout" mapstructure:"default_layout"`                // 默认布局
		LayoutSuffix       string `json:"layout_suffix" yaml:"layout_suffix" mapstructure:"layout_suffix"`                   // 布局文件后缀
		PartialPrefix      string `json:"partial_prefix" yaml:"partial_prefix" mapstructure:"partial_prefix"`                // 片段文件前缀
		ComponentNamespace string `json:"component_namespace" yaml:"component_namespace" mapstructure:"component_namespace"` // 组件命名空间
		ComponentCache     bool   `json:"component_cache" yaml:"component_cache" mapstructure:"component_cache"`             // 组件缓存
		ComponentTimeout   int    `json:"component_timeout" yaml:"component_timeout" mapstructure:"component_timeout"`       // 组件渲染超时（毫秒）
	} `json:"layout" yaml:"layout" mapstructure:"layout"`

	// 渲染配置
	Render struct {
		MaxRenderDepth int  `json:"max_render_depth" yaml:"max_render_depth" mapstructure:"max_render_depth"` // 最大渲染深度（防止无限递归）
		RenderTimeout  int  `json:"render_timeout" yaml:"render_timeout" mapstructure:"render_timeout"`       // 渲染超时（毫秒）
		BufferSize     int  `json:"buffer_size" yaml:"buffer_size" mapstructure:"buffer_size"`                // 渲染缓冲区大小
		Streaming      bool `json:"streaming" yaml:"streaming" mapstructure:"streaming"`                      // 流式渲染
	} `json:"render" yaml:"render" mapstructure:"render"`

	// 热重载配置
	Reload struct {
		Enabled         bool     `json:"enabled" yaml:"enabled" mapstructure:"enabled"`                            // 启用热重载
		WatchInterval   int      `json:"watch_interval" yaml:"watch_interval" mapstructure:"watch_interval"`       // 文件监控间隔（毫秒）
		WatchExtensions []string `json:"watch_extensions" yaml:"watch_extensions" mapstructure:"watch_extensions"` // 监控的文件扩展名
		WatchExclude    []string `json:"watch_exclude" yaml:"watch_exclude" mapstructure:"watch_exclude"`          // 排除的目录
		DebounceDelay   int      `json:"debounce_delay" yaml:"debounce_delay" mapstructure:"debounce_delay"`       // 防抖延迟（毫秒）
	} `json:"reload" yaml:"reload" mapstructure:"reload"`

	// 插件系统配置
	Plugins struct {
		Enabled bool   `json:"enabled" yaml:"enabled" mapstructure:"enabled"` // 启用插件
		Path    string `json:"path" yaml:"path" mapstructure:"path"`          // 插件目录
	} `json:"plugins" yaml:"plugins" mapstructure:"plugins"`

	// 高级功能配置
	Advanced struct {
		StrictMode        bool   `json:"strict_mode" yaml:"strict_mode" mapstructure:"strict_mode"`                      // 严格模式
		CompatibilityMode string `json:"compatibility_mode" yaml:"compatibility_mode" mapstructure:"compatibility_mode"` // 兼容模式: "beego", "gin", ""
	} `json:"advanced" yaml:"advanced" mapstructure:"advanced"`

	// 日志监控配置
	Monitor struct {
		TemplateLogEnabled    bool   `json:"template_log_enabled" yaml:"template_log_enabled" mapstructure:"template_log_enabled"`          // 模板日志
		TemplateLogLevel      string `json:"template_log_level" yaml:"template_log_level" mapstructure:"template_log_level"`                // 日志级别
		TemplateLogFile       string `json:"template_log_file" yaml:"template_log_file" mapstructure:"template_log_file"`                   // 日志文件
		MetricsEnabled        bool   `json:"metrics_enabled" yaml:"metrics_enabled" mapstructure:"metrics_enabled"`                         // 启用指标收集
		MetricsEndpoint       string `json:"metrics_endpoint" yaml:"metrics_endpoint" mapstructure:"metrics_endpoint"`                      // 指标端点
		SlowTemplateThreshold int    `json:"slow_template_threshold" yaml:"slow_template_threshold" mapstructure:"slow_template_threshold"` // 慢模板阈值（毫秒）
	} `json:"monitor" yaml:"monitor" mapstructure:"monitor"`

	// 主题配置
	Theme struct {
		Current string                  `json:"current" yaml:"current" mapstructure:"current"` // 当前主题
		Themes  map[string]*ThemeConfig `json:"themes" yaml:"themes" mapstructure:"themes"`    // 主题配置映射
	} `json:"theme" yaml:"theme" mapstructure:"theme"`
}

// DefaultTemplateConfig 返回全局的模板配置实例。
// 该函数使用单例模式确保全局配置只初始化一次，并通过 viper 加载默认配置。
// 如果配置尚未加载，会调用 SetDefaults 方法设置默认值。
// 返回的实例为 *TemplateConfig 类型，确保全局唯一性。
func DefaultTemplateConfig() *TemplateConfig {
	templateOnce.Do(func() {
		template := &TemplateConfig{}

		// 基础路径配置
		template.Paths.ViewPaths = []string{"views", "templates"}
		template.Paths.LayoutPath = "views/layouts"
		template.Paths.ComponentPath = "views/components"
		template.Paths.Extension = ".html"
		template.Paths.StaticPaths = map[string]string{
			"/static":    "./static",    // 基础静态文件
			"/assets":    "./assets",    // 前端资源
			"/uploads":   "./uploads",   // 用户上传
			"/downloads": "./downloads", // 用户下载
			"/public":    "./public",    // 公共资源
		}

		// 模板语法配置
		template.Syntax.DelimLeft = "{{"
		template.Syntax.DelimRight = "}}"

		// 缓存配置
		template.Cache.EnableCache = true
		template.Cache.CacheSize = 1000
		template.Cache.CacheTTL = 3600
		template.Cache.CacheStrategy = "lru"

		// 性能优化配置
		template.Performance.Precompile = false
		template.Performance.CompileParallel = false
		template.Performance.CompileWorkers = 4
		template.Performance.EnableCompress = false
		template.Performance.MinifyHTML = false
		template.Performance.MinifyCSS = false
		template.Performance.MinifyJS = false
		template.Performance.BundleAssets = false

		// 国际化配置
		template.I18n.Enabled = false
		template.I18n.DefaultLocale = "zh-CN"
		template.I18n.SupportedLocales = []string{"zh-CN", "en-US"}
		template.I18n.LocalePath = "./locales"
		template.I18n.FallbackLocale = "en-US"
		template.I18n.DetectFromHeader = true

		// 错误处理配置
		template.Error.Templates = map[string]string{
			"404": "errors/404.html",
			"500": "errors/500.html",
			"403": "errors/403.html",
		}
		template.Error.ShowDetails = false
		template.Error.LogLevel = "error"

		// 开发调试配置
		template.Debug.Mode = false
		template.Debug.Toolbar = false
		template.Debug.TemplateComments = false
		template.Debug.SourceMap = false
		template.Debug.Profiling = false

		// 资源管理配置
		template.Assets.Versioning = false
		template.Assets.VersionStrategy = "hash"
		template.Assets.ManifestPath = "./manifest.json"
		template.Assets.CDNEnabled = false
		template.Assets.CDNUrl = ""
		template.Assets.CDNFallback = true

		// 布局和组件配置
		template.Layout.DefaultLayout = "main"
		template.Layout.LayoutSuffix = "_layout"
		template.Layout.PartialPrefix = "_"
		template.Layout.ComponentNamespace = "components"
		template.Layout.ComponentCache = true
		template.Layout.ComponentTimeout = 5000

		// 渲染配置
		template.Render.MaxRenderDepth = 10
		template.Render.RenderTimeout = 30000
		template.Render.BufferSize = 4096
		template.Render.Streaming = false

		// 热重载配置
		template.Reload.Enabled = true
		template.Reload.WatchInterval = 1000
		template.Reload.WatchExtensions = []string{".html", ".css", ".js"}
		template.Reload.WatchExclude = []string{"node_modules", ".git"}
		template.Reload.DebounceDelay = 500

		// 插件系统配置
		template.Plugins.Enabled = false
		template.Plugins.Path = "./plugins"

		// 高级功能配置
		template.Advanced.StrictMode = false
		template.Advanced.CompatibilityMode = ""

		// 日志监控配置
		template.Monitor.TemplateLogEnabled = false
		template.Monitor.TemplateLogLevel = "info"
		template.Monitor.TemplateLogFile = "./logs/template.log"
		template.Monitor.MetricsEnabled = false
		template.Monitor.MetricsEndpoint = "/metrics"
		template.Monitor.SlowTemplateThreshold = 1000

		// 主题配置
		template.Theme.Current = "default"
		template.Theme.Themes = map[string]*ThemeConfig{
			"default": {
				Name:          "default",
				ViewPaths:     []string{"views"},
				LayoutPath:    "views/layouts",
				ComponentPath: "views/components",
				StaticPath:    "static",
				Enabled:       true,
				Default:       true,
				Variables:     make(map[string]string),
			},
		}
	})

	templateConfigOnce.Do(func() {
		if !templateConfigLoaded {
			v := viper.New()
			GlobalTemplate.SetDefaults(v)
			templateConfigLoaded = true
		}
	})

	return GlobalTemplate
}

// GetConfigName 实现 ConfigInterface 接口
func (c TemplateConfig) GetConfigName() string {
	return TemplateConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c TemplateConfig) SetDefaults(v *viper.Viper) {
	// ========== 基础路径配置 ==========
	v.SetDefault("paths.view_paths", []string{"views", "templates"})
	v.SetDefault("paths.layout_path", "views/layouts")
	v.SetDefault("paths.component_path", "views/components")
	v.SetDefault("paths.extension", ".html")
	v.SetDefault("paths.static_paths", map[string]string{
		"/static":    "./static",
		"/assets":    "./assets",
		"/uploads":   "./uploads",
		"/downloads": "./downloads",
		"/public":    "./public",
	})

	// ========== 模板语法配置 ==========
	v.SetDefault("syntax.delim_left", "{{")
	v.SetDefault("syntax.delim_right", "}}")

	// ========== 缓存配置 ==========
	v.SetDefault("cache.enable_cache", true)
	v.SetDefault("cache.cache_size", 1000)
	v.SetDefault("cache.cache_ttl", 3600)
	v.SetDefault("cache.cache_strategy", "lru")

	// ========== 性能优化配置 ==========
	v.SetDefault("performance.precompile", false)
	v.SetDefault("performance.compile_parallel", false)
	v.SetDefault("performance.compile_workers", 4)
	v.SetDefault("performance.enable_compress", false)
	v.SetDefault("performance.minify_html", false)
	v.SetDefault("performance.minify_css", false)
	v.SetDefault("performance.minify_js", false)
	v.SetDefault("performance.bundle_assets", false)

	// ========== 国际化配置 ==========
	v.SetDefault("i18n.enabled", false)
	v.SetDefault("i18n.default_locale", "zh-CN")
	v.SetDefault("i18n.supported_locales", []string{"zh-CN", "en-US"})
	v.SetDefault("i18n.locale_path", "./locales")
	v.SetDefault("i18n.fallback_locale", "en-US")
	v.SetDefault("i18n.detect_from_header", true)

	// ========== 错误处理配置 ==========
	v.SetDefault("error.templates", map[string]string{
		"404": "errors/404.html",
		"500": "errors/500.html",
		"403": "errors/403.html",
	})
	v.SetDefault("error.show_details", false)
	v.SetDefault("error.log_level", "error")

	// ========== 开发调试配置 ==========
	v.SetDefault("debug.mode", false)
	v.SetDefault("debug.toolbar", false)
	v.SetDefault("debug.template_comments", false)
	v.SetDefault("debug.source_map", false)
	v.SetDefault("debug.profiling", false)

	// ========== 资源管理配置 ==========
	v.SetDefault("assets.versioning", false)
	v.SetDefault("assets.version_strategy", "hash")
	v.SetDefault("assets.manifest_path", "./manifest.json")
	v.SetDefault("assets.cdn_enabled", false)
	v.SetDefault("assets.cdn_url", "")
	v.SetDefault("assets.cdn_fallback", true)

	// ========== 布局和组件配置 ==========
	v.SetDefault("layout.default_layout", "main")
	v.SetDefault("layout.layout_suffix", "_layout")
	v.SetDefault("layout.partial_prefix", "_")
	v.SetDefault("layout.component_namespace", "components")
	v.SetDefault("layout.component_cache", true)
	v.SetDefault("layout.component_timeout", 5000)

	// ========== 渲染配置 ==========
	v.SetDefault("render.max_render_depth", 10)
	v.SetDefault("render.render_timeout", 30000)
	v.SetDefault("render.buffer_size", 4096)
	v.SetDefault("render.streaming", false)

	// ========== 热重载配置 ==========
	v.SetDefault("reload.enabled", true)
	v.SetDefault("reload.watch_interval", 1000)
	v.SetDefault("reload.watch_extensions", []string{".html", ".css", ".js"})
	v.SetDefault("reload.watch_exclude", []string{"node_modules", ".git"})
	v.SetDefault("reload.debounce_delay", 500)

	// ========== 插件系统配置 ==========
	v.SetDefault("plugins.enabled", false)
	v.SetDefault("plugins.path", "./plugins")

	// ========== 高级功能配置 ==========
	v.SetDefault("advanced.strict_mode", false)
	v.SetDefault("advanced.compatibility_mode", "")

	// ========== 日志监控配置 ==========
	v.SetDefault("monitor.template_log_enabled", false)
	v.SetDefault("monitor.template_log_level", "info")
	v.SetDefault("monitor.template_log_file", "./logs/template.log")
	v.SetDefault("monitor.metrics_enabled", false)
	v.SetDefault("monitor.metrics_endpoint", "/metrics")
	v.SetDefault("monitor.slow_template_threshold", 1000)

	// ========== 主题配置 ==========
	v.SetDefault("theme.current", "default")
	v.SetDefault("theme.themes.default.name", "default")
	v.SetDefault("theme.themes.default.view_paths", []string{"views"})
	v.SetDefault("theme.themes.default.layout_path", "views/layouts")
	v.SetDefault("theme.themes.default.component_path", "views/components")
	v.SetDefault("theme.themes.default.static_path", "static")
	v.SetDefault("theme.themes.default.enabled", true)
	v.SetDefault("theme.themes.default.default", true)
	v.SetDefault("theme.themes.default.variables", map[string]string{})
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c TemplateConfig) GenerateDefaultContent() string {
	return `# YYHertz Template Engine Configuration
# 模板引擎配置文件 - 分组结构配置示例

# ========== 基础路径配置 ==========
paths:
  view_paths: ["views", "templates"]  # 视图文件搜索路径
  layout_path: "views/layouts"        # 布局文件路径
  component_path: "views/components"  # 组件文件路径
  extension: ".html"                  # 模板文件扩展名
  # 静态资源配置 - URL路径到本地路径的映射
  static_paths:
    "/static": "./static"             # 静态文件路径
    "/assets": "./assets"             # 资源文件路径
    "/uploads": "./uploads"           # 上传文件路径
    "/downloads": "./downloads"       # 下载文件路径
    "/public": "./public"             # 公共资源路径

# ========== 模板语法配置 ==========
syntax:
  delim_left: "{{"                    # 左分隔符
  delim_right: "}}"                   # 右分隔符

# ========== 缓存配置 ==========
cache:
  enable_cache: true                  # 启用模板缓存
  cache_size: 1000                    # 缓存条目数量限制
  cache_ttl: 3600                     # 缓存过期时间（秒）
  cache_strategy: "lru"               # 缓存策略: lru, lfu, fifo

# ========== 性能优化配置 ==========
performance:
  precompile: false                   # 预编译模板
  compile_parallel: false             # 并行编译
  compile_workers: 4                  # 编译工作线程数
  enable_compress: false              # 启用压缩
  minify_html: false                  # HTML压缩
  minify_css: false                   # CSS压缩
  minify_js: false                    # JS压缩
  bundle_assets: false                # 资源打包

# ========== 国际化配置 ==========
i18n:
  enabled: false                      # 启用国际化
  default_locale: "zh-CN"             # 默认语言
  supported_locales: ["zh-CN", "en-US"] # 支持的语言列表
  locale_path: "./locales"            # 语言文件路径
  fallback_locale: "en-US"            # 回退语言
  detect_from_header: true            # 从请求头检测语言

# ========== 错误处理配置 ==========
error:
  templates:                          # 错误页面模板映射
    "404": "errors/404.html"
    "500": "errors/500.html"
    "403": "errors/403.html"
  show_details: false                 # 显示错误详情（开发模式建议true）
  log_level: "error"                  # 错误日志级别

# ========== 开发调试配置 ==========
debug:
  mode: false                         # 调试模式
  toolbar: false                      # 调试工具栏
  template_comments: false            # 模板注释
  source_map: false                   # 源码映射
  profiling: false                    # 性能分析

# ========== 资源管理配置 ==========
assets:
  versioning: false                   # 资源版本控制
  version_strategy: "hash"            # 版本策略: hash, timestamp, manual
  manifest_path: "./manifest.json"    # 资源清单文件
  cdn_enabled: false                  # 启用CDN
  cdn_url: ""                         # CDN基础URL
  cdn_fallback: true                  # CDN回退到本地

# ========== 布局和组件配置 ==========
layout:
  default_layout: "main"              # 默认布局
  layout_suffix: "_layout"            # 布局文件后缀
  partial_prefix: "_"                 # 片段文件前缀
  component_namespace: "components"   # 组件命名空间
  component_cache: true               # 组件缓存
  component_timeout: 5000             # 组件渲染超时（毫秒）

# ========== 渲染配置 ==========
render:
  max_render_depth: 10                # 最大渲染深度（防止无限递归）
  render_timeout: 30000               # 渲染超时（毫秒）
  buffer_size: 4096                   # 渲染缓冲区大小
  streaming: false                    # 流式渲染

# ========== 热重载配置 ==========
reload:
  enabled: true                       # 启用热重载
  watch_interval: 1000                # 文件监控间隔（毫秒）
  watch_extensions: [".html", ".css", ".js"] # 监控的文件扩展名
  watch_exclude: ["node_modules", ".git"] # 排除的目录
  debounce_delay: 500                 # 防抖延迟（毫秒）

# ========== 插件系统配置 ==========
plugins:
  enabled: false                      # 启用插件
  path: "./plugins"                   # 插件目录

# ========== 高级功能配置 ==========
advanced:
  strict_mode: false                  # 严格模式
  compatibility_mode: ""              # 兼容模式: "beego", "gin", ""

# ========== 日志监控配置 ==========
monitor:
  template_log_enabled: false         # 模板日志
  template_log_level: "info"          # 日志级别
  template_log_file: "./logs/template.log" # 日志文件
  metrics_enabled: false              # 启用指标收集
  metrics_endpoint: "/metrics"        # 指标端点
  slow_template_threshold: 1000       # 慢模板阈值（毫秒）

# ========== 主题配置 ==========
theme:
  current: "default"                  # 当前主题
  # 主题映射配置
  themes:
    default:
      name: "default"                 # 主题名称
      view_paths: ["views"]           # 主题视图路径
      layout_path: "views/layouts"    # 主题布局路径
      component_path: "views/components" # 主题组件路径
      static_path: "static"           # 静态资源路径
      enabled: true                   # 是否启用
      default: true                   # 是否为默认主题
      variables: {}                   # 主题变量

# ========== 配置示例和使用说明 ==========
#
# 1. 生产环境优化配置:
#    cache:
#      enable_cache: true            # 启用缓存提高性能
#    reload:
#      enabled: false                # 禁用热重载
#    performance:
#      enable_compress: true         # 启用压缩
#      minify_html: true             # HTML压缩
#      precompile: true              # 预编译模板
#      compile_parallel: true        # 并行编译
#
# 2. 开发环境调试配置:
#    debug:
#      mode: true                    # 开启调试模式
#      toolbar: true                 # 显示调试工具栏
#      template_comments: true       # 添加模板注释
#      source_map: true              # 生成源码映射
#    error:
#      show_details: true            # 显示详细错误信息
#
# 3. 自定义主题配置:
#    theme:
#      themes:
#        dark:
#          name: "dark"
#          view_paths: ["themes/dark/views"]
#          layout_path: "themes/dark/layouts"
#          component_path: "themes/dark/components"
#          enabled: true
#          variables:
#            primary_color: "#1a1a1a"
#            secondary_color: "#333333"
#
# 4. 国际化配置:
#    i18n:
#      enabled: true
#      supported_locales: ["zh-CN", "en-US", "ja-JP", "ko-KR"]
#      default_locale: "zh-CN"
#      fallback_locale: "en-US"
#
# 5. CDN和资源优化:
#    assets:
#      cdn_enabled: true
#      cdn_url: "https://cdn.example.com"
#      versioning: true
#      version_strategy: "hash"
#      bundle_assets: true
#
# 6. 监控和性能分析:
#    monitor:
#      metrics_enabled: true
#      template_log_enabled: true
#      slow_template_threshold: 500
#    debug:
#      profiling: true
`
}

// LoadTemplateConfig 从配置文件加载模板配置
func LoadTemplateConfig() *TemplateConfig {
	// 先获取默认配置
	cfg := DefaultTemplateConfig()

	// 尝试从配置文件读取设置
	manager := GetViperConfigManager(*cfg)
	configPtr, err := manager.GetConfig()
	if err != nil {
		// 如果加载失败，返回默认配置
		return cfg
	}

	return configPtr
}
