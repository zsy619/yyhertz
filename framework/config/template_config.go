package config

import (
	"github.com/spf13/viper"
)

// TemplateConfig 模板引擎配置结构
type TemplateConfig struct {
	// 模板引擎基础配置
	Engine struct {
		Type         string   `mapstructure:"type" yaml:"type" json:"type"`                            // html, pug, handlebars, etc.
		Directory    string   `mapstructure:"directory" yaml:"directory" json:"directory"`             // 模板文件目录
		Extension    string   `mapstructure:"extension" yaml:"extension" json:"extension"`             // 模板文件扩展名
		Reload       bool     `mapstructure:"reload" yaml:"reload" json:"reload"`                      // 是否自动重载
		Debug        bool     `mapstructure:"debug" yaml:"debug" json:"debug"`                         // 调试模式
		Charset      string   `mapstructure:"charset" yaml:"charset" json:"charset"`                   // 字符编码
		Delimiters   []string `mapstructure:"delimiters" yaml:"delimiters" json:"delimiters"`          // 模板分隔符
		FuncMap      []string `mapstructure:"func_map" yaml:"func_map" json:"func_map"`                // 自定义函数
		LayoutDir    string   `mapstructure:"layout_dir" yaml:"layout_dir" json:"layout_dir"`          // 布局文件目录
		PartialsDir  string   `mapstructure:"partials_dir" yaml:"partials_dir" json:"partials_dir"`    // 部分模板目录
		DisableCache bool     `mapstructure:"disable_cache" yaml:"disable_cache" json:"disable_cache"` // 禁用缓存
	} `mapstructure:"engine" yaml:"engine" json:"engine"`

	// 渲染配置
	Render struct {
		HTMLEscape   bool     `mapstructure:"html_escape" yaml:"html_escape" json:"html_escape"`       // HTML转义
		CompressHTML bool     `mapstructure:"compress_html" yaml:"compress_html" json:"compress_html"` // 压缩HTML
		IndentJSON   bool     `mapstructure:"indent_json" yaml:"indent_json" json:"indent_json"`       // JSON缩进
		IndentXML    bool     `mapstructure:"indent_xml" yaml:"indent_xml" json:"indent_xml"`          // XML缩进
		DefaultType  string   `mapstructure:"default_type" yaml:"default_type" json:"default_type"`    // 默认渲染类型
		Headers      []string `mapstructure:"headers" yaml:"headers" json:"headers"`                   // 默认响应头
	} `mapstructure:"render" yaml:"render" json:"render"`

	// 缓存配置
	Cache struct {
		Enable    bool   `mapstructure:"enable" yaml:"enable" json:"enable"`             // 启用缓存
		Type      string `mapstructure:"type" yaml:"type" json:"type"`                   // 缓存类型: memory, redis, file
		TTL       int    `mapstructure:"ttl" yaml:"ttl" json:"ttl"`                      // 缓存TTL(秒)
		MaxSize   int    `mapstructure:"max_size" yaml:"max_size" json:"max_size"`       // 最大缓存条目数
		KeyPrefix string `mapstructure:"key_prefix" yaml:"key_prefix" json:"key_prefix"` // 缓存键前缀
		CacheDir  string `mapstructure:"cache_dir" yaml:"cache_dir" json:"cache_dir"`    // 文件缓存目录
	} `mapstructure:"cache" yaml:"cache" json:"cache"`

	// 静态资源配置
	Static struct {
		Enable        bool     `mapstructure:"enable" yaml:"enable" json:"enable"`                         // 启用静态文件服务
		Root          string   `mapstructure:"root" yaml:"root" json:"root"`                               // 静态文件根目录
		Index         []string `mapstructure:"index" yaml:"index" json:"index"`                            // 默认首页文件
		Browse        bool     `mapstructure:"browse" yaml:"browse" json:"browse"`                         // 启用目录浏览
		Compress      bool     `mapstructure:"compress" yaml:"compress" json:"compress"`                   // 启用压缩
		CacheDuration int      `mapstructure:"cache_duration" yaml:"cache_duration" json:"cache_duration"` // 缓存时长(秒)
		Exclude       []string `mapstructure:"exclude" yaml:"exclude" json:"exclude"`                      // 排除的文件/目录
	} `mapstructure:"static" yaml:"static" json:"static"`

	// 组件配置
	Components struct {
		Enable       bool     `mapstructure:"enable" yaml:"enable" json:"enable"`                      // 启用组件系统
		Directory    string   `mapstructure:"directory" yaml:"directory" json:"directory"`             // 组件目录
		AutoRegister bool     `mapstructure:"auto_register" yaml:"auto_register" json:"auto_register"` // 自动注册
		Namespace    string   `mapstructure:"namespace" yaml:"namespace" json:"namespace"`             // 组件命名空间
		Extensions   []string `mapstructure:"extensions" yaml:"extensions" json:"extensions"`          // 组件文件扩展名
	} `mapstructure:"components" yaml:"components" json:"components"`

	// 开发配置
	Development struct {
		LiveReload    bool     `mapstructure:"live_reload" yaml:"live_reload" json:"live_reload"`          // 实时重载
		AutoBuild     bool     `mapstructure:"auto_build" yaml:"auto_build" json:"auto_build"`             // 自动构建
		WatchPatterns []string `mapstructure:"watch_patterns" yaml:"watch_patterns" json:"watch_patterns"` // 监听文件模式
		BuildCommand  string   `mapstructure:"build_command" yaml:"build_command" json:"build_command"`    // 构建命令
		DevServer     struct {
			Port        int    `mapstructure:"port" yaml:"port" json:"port"`                         // 开发服务器端口
			Host        string `mapstructure:"host" yaml:"host" json:"host"`                         // 开发服务器主机
			Proxy       string `mapstructure:"proxy" yaml:"proxy" json:"proxy"`                      // 代理地址
			OpenBrowser bool   `mapstructure:"open_browser" yaml:"open_browser" json:"open_browser"` // 自动打开浏览器
		} `mapstructure:"dev_server" yaml:"dev_server" json:"dev_server"`
	} `mapstructure:"development" yaml:"development" json:"development"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c TemplateConfig) GetConfigName() string {
	return TemplateConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c TemplateConfig) SetDefaults(v *viper.Viper) {
	// 模板引擎默认配置
	v.SetDefault("engine.type", "html")
	v.SetDefault("engine.directory", "./views")
	v.SetDefault("engine.extension", ".html")
	v.SetDefault("engine.reload", true)
	v.SetDefault("engine.debug", false)
	v.SetDefault("engine.charset", "UTF-8")
	v.SetDefault("engine.delimiters", []string{"{{", "}}"})
	v.SetDefault("engine.func_map", []string{})
	v.SetDefault("engine.layout_dir", "./views/layout")
	v.SetDefault("engine.partials_dir", "./views/partials")
	v.SetDefault("engine.disable_cache", false)

	// 渲染默认配置
	v.SetDefault("render.html_escape", true)
	v.SetDefault("render.compress_html", false)
	v.SetDefault("render.indent_json", false)
	v.SetDefault("render.indent_xml", false)
	v.SetDefault("render.default_type", "text/html")
	v.SetDefault("render.headers", []string{})

	// 缓存默认配置
	v.SetDefault("cache.enable", true)
	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.ttl", 3600)
	v.SetDefault("cache.max_size", 1000)
	v.SetDefault("cache.key_prefix", "template:")
	v.SetDefault("cache.cache_dir", "./cache/templates")

	// 静态资源默认配置
	v.SetDefault("static.enable", true)
	v.SetDefault("static.root", "./static")
	v.SetDefault("static.index", []string{"index.html", "index.htm"})
	v.SetDefault("static.browse", false)
	v.SetDefault("static.compress", true)
	v.SetDefault("static.cache_duration", 86400)
	v.SetDefault("static.exclude", []string{".git", ".svn", ".DS_Store"})

	// 组件默认配置
	v.SetDefault("components.enable", false)
	v.SetDefault("components.directory", "./views/components")
	v.SetDefault("components.auto_register", true)
	v.SetDefault("components.namespace", "")
	v.SetDefault("components.extensions", []string{".html", ".htm"})

	// 开发环境默认配置
	v.SetDefault("development.live_reload", true)
	v.SetDefault("development.auto_build", false)
	v.SetDefault("development.watch_patterns", []string{"*.html", "*.css", "*.js"})
	v.SetDefault("development.build_command", "")
	v.SetDefault("development.dev_server.port", 3000)
	v.SetDefault("development.dev_server.host", "localhost")
	v.SetDefault("development.dev_server.proxy", "")
	v.SetDefault("development.dev_server.open_browser", false)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c TemplateConfig) GenerateDefaultContent() string {
	return `# YYHertz Template Engine Configuration
# 模板引擎配置文件

# 模板引擎基础配置
engine:
  type: "html"                    # 模板引擎类型: html, pug, handlebars
  directory: "./views"            # 模板文件目录
  extension: ".html"              # 模板文件扩展名
  reload: true                    # 开发环境下自动重载
  debug: false                    # 调试模式
  charset: "UTF-8"                # 字符编码
  delimiters: ["{{", "}}"]        # 模板分隔符
  func_map: []                    # 自定义模板函数
  layout_dir: "./views/layout"    # 布局文件目录
  partials_dir: "./views/partials" # 部分模板目录
  disable_cache: false            # 禁用缓存

# 渲染配置
render:
  html_escape: true               # HTML自动转义
  compress_html: false           # 压缩HTML输出
  indent_json: false             # JSON格式化输出
  indent_xml: false              # XML格式化输出
  default_type: "text/html"      # 默认Content-Type
  headers: []                    # 默认响应头

# 缓存配置
cache:
  enable: true                   # 启用模板缓存
  type: "memory"                 # 缓存类型: memory, redis, file
  ttl: 3600                      # 缓存生存时间(秒)
  max_size: 1000                 # 内存缓存最大条目数
  key_prefix: "template:"        # 缓存键前缀
  cache_dir: "./cache/templates" # 文件缓存目录

# 静态资源配置
static:
  enable: true                   # 启用静态文件服务
  root: "./static"               # 静态文件根目录
  index: ["index.html", "index.htm"] # 默认首页文件
  browse: false                  # 启用目录浏览
  compress: true                 # 启用gzip压缩
  cache_duration: 86400          # 浏览器缓存时长(秒)
  exclude: [".git", ".svn", ".DS_Store"] # 排除的文件/目录

# 组件系统配置
components:
  enable: false                  # 启用组件系统
  directory: "./views/components" # 组件文件目录
  auto_register: true            # 自动注册组件
  namespace: ""                  # 组件命名空间
  extensions: [".html", ".htm"]  # 组件文件扩展名

# 开发环境配置
development:
  live_reload: true              # 实时重载
  auto_build: false              # 自动构建
  watch_patterns: ["*.html", "*.css", "*.js"] # 监听文件模式
  build_command: ""              # 构建命令
  
  # 开发服务器配置
  dev_server:
    port: 3000                   # 开发服务器端口
    host: "localhost"            # 开发服务器主机
    proxy: ""                    # 代理后端服务地址
    open_browser: false          # 自动打开浏览器
`
}
