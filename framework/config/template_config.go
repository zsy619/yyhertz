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
	ViewPaths     []string                `json:"view_paths" yaml:"view_paths" mapstructure:"view_paths"`             // 视图文件搜索路径
	LayoutPath    string                  `json:"layout_path" yaml:"layout_path" mapstructure:"layout_path"`          // 布局文件路径
	ComponentPath string                  `json:"component_path" yaml:"component_path" mapstructure:"component_path"` // 组件文件路径
	Extension     string                  `json:"extension" yaml:"extension" mapstructure:"extension"`                // 模板文件扩展名

	// 模板分隔符配置
	DelimLeft  string `json:"delim_left" yaml:"delim_left" mapstructure:"delim_left"`    // 左分隔符
	DelimRight string `json:"delim_right" yaml:"delim_right" mapstructure:"delim_right"` // 右分隔符

	// 功能开关配置
	EnableCache    bool `json:"enable_cache" yaml:"enable_cache" mapstructure:"enable_cache"`          // 启用模板缓存
	EnableReload   bool `json:"enable_reload" yaml:"enable_reload" mapstructure:"enable_reload"`       // 启用热重载
	EnableCompress bool `json:"enable_compress" yaml:"enable_compress" mapstructure:"enable_compress"` // 启用压缩

	// 主题配置
	CurrentTheme string                  `json:"current_theme" yaml:"current_theme" mapstructure:"current_theme"` // 当前主题
	Themes       map[string]*ThemeConfig `json:"themes" yaml:"themes" mapstructure:"themes"`                       // 主题配置映射
}

// DefaultTemplateConfig 返回全局的模板配置实例。
// 该函数使用单例模式确保全局配置只初始化一次，并通过 viper 加载默认配置。
// 如果配置尚未加载，会调用 SetDefaults 方法设置默认值。
// 返回的实例为 *TemplateConfig 类型，确保全局唯一性。
func DefaultTemplateConfig() *TemplateConfig {
	templateOnce.Do(func() {
		GlobalTemplate = &TemplateConfig{
			ViewPaths:      []string{"views", "templates"},
			LayoutPath:     "views/layouts",
			ComponentPath:  "views/components",
			Extension:      ".html",
			DelimLeft:      "{{",
			DelimRight:     "}}",
			EnableCache:    true,
			EnableReload:   true,
			EnableCompress: false,
			CurrentTheme:   "default",
			Themes: map[string]*ThemeConfig{
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
	// 基础路径配置
	v.SetDefault("view_paths", []string{"views", "templates"})
	v.SetDefault("layout_path", "views/layouts")
	v.SetDefault("component_path", "views/components")
	v.SetDefault("extension", ".html")

	// 分隔符配置
	v.SetDefault("delim_left", "{{")
	v.SetDefault("delim_right", "}}")

	// 功能开关配置
	v.SetDefault("enable_cache", true)
	v.SetDefault("enable_reload", true)
	v.SetDefault("enable_compress", false)

	// 主题配置
	v.SetDefault("current_theme", "default")
	
	// 默认主题配置
	v.SetDefault("themes.default.name", "default")
	v.SetDefault("themes.default.view_paths", []string{"views"})
	v.SetDefault("themes.default.layout_path", "views/layouts")
	v.SetDefault("themes.default.component_path", "views/components")
	v.SetDefault("themes.default.static_path", "static")
	v.SetDefault("themes.default.enabled", true)
	v.SetDefault("themes.default.default", true)
	v.SetDefault("themes.default.variables", map[string]string{})
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c TemplateConfig) GenerateDefaultContent() string {
	return `# YYHertz Template Engine Configuration
# 模板引擎配置文件

# 基础路径配置
view_paths: ["views", "templates"]  # 视图文件搜索路径
layout_path: "views/layouts"        # 布局文件路径
component_path: "views/components"  # 组件文件路径
extension: ".html"                  # 模板文件扩展名

# 模板分隔符配置
delim_left: "{{"                    # 左分隔符
delim_right: "}}"                   # 右分隔符

# 功能开关配置
enable_cache: true                  # 启用模板缓存
enable_reload: true                 # 启用热重载
enable_compress: false              # 启用压缩

# 主题配置
current_theme: "default"            # 当前主题

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

# 使用示例:
# 1. 自定义主题配置:
#    themes:
#      dark:
#        name: "dark"
#        view_paths: ["themes/dark/views"]
#        layout_path: "themes/dark/layouts"
#        enabled: true
#        variables:
#          primary_color: "#1a1a1a"
#          secondary_color: "#333333"
#
# 2. 多路径配置:
#    view_paths: ["views", "custom_views", "third_party_views"]
#
# 3. 性能优化:
#    enable_cache: true      # 生产环境建议启用
#    enable_reload: false    # 生产环境建议禁用
#    enable_compress: true   # 启用压缩减少传输大小
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

// ConvertToViewTemplateConfig 转换为view包的TemplateConfig（向后兼容）
func ConvertToViewTemplateConfig(cfg *TemplateConfig) *ViewTemplateConfig {
	if cfg == nil {
		cfg = DefaultTemplateConfig()
	}
	
	return &ViewTemplateConfig{
		ViewPaths:      cfg.ViewPaths,
		LayoutPath:     cfg.LayoutPath,
		ComponentPath:  cfg.ComponentPath,
		Extension:      cfg.Extension,
		DelimLeft:      cfg.DelimLeft,
		DelimRight:     cfg.DelimRight,
		EnableCache:    cfg.EnableCache,
		EnableReload:   cfg.EnableReload,
		EnableCompress: cfg.EnableCompress,
		CurrentTheme:   cfg.CurrentTheme,
		Themes:         convertThemes(cfg.Themes),
	}
}

// ViewTemplateConfig view包的TemplateConfig结构（向后兼容）
type ViewTemplateConfig struct {
	ViewPaths      []string                   `json:"view_paths" yaml:"view_paths"`
	LayoutPath     string                     `json:"layout_path" yaml:"layout_path"`
	ComponentPath  string                     `json:"component_path" yaml:"component_path"`
	Extension      string                     `json:"extension" yaml:"extension"`
	DelimLeft      string                     `json:"delim_left" yaml:"delim_left"`
	DelimRight     string                     `json:"delim_right" yaml:"delim_right"`
	EnableCache    bool                       `json:"enable_cache" yaml:"enable_cache"`
	EnableReload   bool                       `json:"enable_reload" yaml:"enable_reload"`
	EnableCompress bool                       `json:"enable_compress" yaml:"enable_compress"`
	CurrentTheme   string                     `json:"current_theme" yaml:"current_theme"`
	Themes         map[string]*ViewThemeConfig `json:"themes" yaml:"themes"`
}

// ViewThemeConfig view包的ThemeConfig结构（向后兼容）
type ViewThemeConfig struct {
	Name          string            `json:"name"`
	ViewPaths     []string          `json:"view_paths"`
	LayoutPath    string            `json:"layout_path"`
	ComponentPath string            `json:"component_path"`
	StaticPath    string            `json:"static_path"`
	Enabled       bool              `json:"enabled"`
	Default       bool              `json:"default"`
	Variables     map[string]string `json:"variables"`
}

// convertThemes 转换主题配置映射
func convertThemes(themes map[string]*ThemeConfig) map[string]*ViewThemeConfig {
	if themes == nil {
		return nil
	}
	
	viewThemes := make(map[string]*ViewThemeConfig)
	for name, theme := range themes {
		if theme != nil {
			viewThemes[name] = &ViewThemeConfig{
				Name:          theme.Name,
				ViewPaths:     theme.ViewPaths,
				LayoutPath:    theme.LayoutPath,
				ComponentPath: theme.ComponentPath,
				StaticPath:    theme.StaticPath,
				Enabled:       theme.Enabled,
				Default:       theme.Default,
				Variables:     theme.Variables,
			}
		}
	}
	return viewThemes
}

// DefaultViewTemplateConfig 默认view包TemplateConfig（向后兼容）
func DefaultViewTemplateConfig() *ViewTemplateConfig {
	cfg := DefaultTemplateConfig()
	return ConvertToViewTemplateConfig(cfg)
}
