package config

import (
	"fmt"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/util"
)

// 控制台颜色常量
const (
	// 前景色
	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"

	// 背景色
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"

	// 样式
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Underline = "\033[4m"
)

// colorPrint 根据不同平台输出彩色文本
func colorPrint(text string, color string) string {
	// Windows平台下不使用ANSI颜色代码
	if runtime.GOOS == "windows" {
		return text
	}
	return color + text + Reset
}

// printBanner 打印启动banner
func printBanner() {
	// 使用大号ASCII艺术字体
	banner := `

	██╗   ██╗██╗   ██╗██╗  ██╗███████╗██████╗ ████████╗███████╗
	╚██╗ ██╔╝╚██╗ ██╔╝██║  ██║██╔════╝██╔══██╗╚══██╔══╝╚══███╔╝
	 ╚████╔╝  ╚████╔╝ ███████║█████╗  ██████╔╝   ██║     ███╔╝ 
	  ╚██╔╝    ╚██╔╝  ██╔══██║██╔══╝  ██╔══██╗   ██║    ███╔╝  
	   ██║      ██║   ██║  ██║███████╗██║  ██║   ██║   ███████╗
	   ╚═╝      ╚═╝   ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝   ╚═╝   ╚══════╝
`

	// 定义不同颜色 - 使用浅色为主
	colors := []string{FgYellow, FgCyan, FgWhite, FgGreen, FgMagenta, FgBlue}
	lines := []string{}

	// 将banner分割成行
	for i, line := range strings.Split(banner, "\n") {
		if line == "" {
			lines = append(lines, "")
			continue
		}
		// 每行使用不同颜色
		color := colors[i%len(colors)]
		lines = append(lines, colorPrint(line, color))
	}

	// 打印彩色banner
	for _, line := range lines {
		fmt.Println(line)
	}

	// 打印系统信息
	fmt.Println()
	printSystemInfo()
}

// printSystemInfo 打印系统信息
func printSystemInfo() {
	// 获取系统信息
	goVersion := runtime.Version()
	goOS := runtime.GOOS
	goArch := runtime.GOARCH
	numCPU := runtime.NumCPU()

	// 打印系统信息 - 使用浅色
	fmt.Println(colorPrint("系统信息:", FgYellow))
	fmt.Printf("%s %s\n", colorPrint("Go 版 本:", FgCyan), goVersion)
	fmt.Printf("%s %s/%s\n", colorPrint("操作系统:", FgCyan), goOS, goArch)
	fmt.Printf("%s %d\n", colorPrint("CPU核心数:", FgCyan), numCPU)

	// 打印框架信息 - 使用浅色
	fmt.Println()
	fmt.Println(colorPrint("框架信息:", FgYellow))
	fmt.Printf("%s %s\n", colorPrint("框架名称:", FgCyan), colorPrint("YYHertz MVC", FgWhite))
	fmt.Printf("%s %s\n", colorPrint("版    本:", FgCyan), "v1.0.0")
	fmt.Printf("%s %s\n", colorPrint("作    者:", FgCyan), "YYHertz Team")
	fmt.Println()
}

func InitConfig[T ConfigInterface](cnf T) {
	appConf := path.Join(".", "conf", fmt.Sprintf("%s.yaml", cnf.GetConfigName()))
	// 判断文件是否存在
	if isExists := util.FileExists(appConf); !isExists {
		// 文件不存在，生成默认配置
		cm := NewViperConfigManager(cnf)
		err := cm.Initialize()
		if err != nil {
			panic(err)
		}
		WatchConfig(cnf)
	}
}

// 初始化配置注册表
func init() {
	// 打印启动banner
	go func() {
		printBanner()
	}()

	InitConfig(&AppConfig{})
	InitConfig(&TemplateConfig{})
	InitConfig(&AuthConfig{})
	InitConfig(&LogConfig{})
	InitConfig(&DatabaseConfig{})
	InitConfig(&RedisConfig{})
	InitConfig(&SessionConfig{})
	InitConfig(&TLSServerConfig{})
	InitConfig(&MyBatisConfig{})

	RegisterConfigName[AppConfig](AppConfigName)
	RegisterConfigName[TemplateConfig](TemplateConfigName)
	RegisterConfigName[AuthConfig](AuthConfigName)
	RegisterConfigName[TLSServerConfig](TLSConfigName)
	RegisterConfigName[LogConfig](LogConfigName)
	RegisterConfigName[DatabaseConfig](DatabaseConfigName)
	RegisterConfigName[SessionConfig](SessionConfigName)
	RegisterConfigName[RedisConfig](RedisConfigName)
	RegisterConfigName[MyBatisConfig](MyBatisConfigName)
}

// 全局便捷函数，用于快速获取不同类型的配置

// GetAppConfig 获取应用配置
func GetAppConfig() (*AppConfig, error) {
	manager := GetViperConfigManager(AppConfig{})
	return manager.GetConfig()
}

// GetTemplateConfig 获取模板配置
func GetTemplateConfig() (*TemplateConfig, error) {
	manager := GetViperConfigManager(TemplateConfig{})
	return manager.GetConfig()
}

// GetAuthConfig 获取认证配置
func GetAuthConfig() (*AuthConfig, error) {
	manager := GetViperConfigManager(AuthConfig{})
	return manager.GetConfig()
}

// GetTLSConfig 获取TLS配置
func GetTLSConfig() (*TLSServerConfig, error) {
	manager := GetViperConfigManager(TLSServerConfig{})
	return manager.GetConfig()
}

// GetLogConfig 获取日志配置
func GetLogConfig() (*LogConfig, error) {
	manager := GetViperConfigManager(LogConfig{})
	return manager.GetConfig()
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() (*DatabaseConfig, error) {
	manager := GetViperConfigManager(DatabaseConfig{})
	return manager.GetConfig()
}

func GetRedisConfig() (*RedisConfig, error) {
	manager := GetViperConfigManager(RedisConfig{})
	return manager.GetConfig()
}

// GetMyBatisConfig 获取MyBatis配置
func GetMyBatisConfig() (*MyBatisConfig, error) {
	manager := GetViperConfigManager(MyBatisConfig{})
	return manager.GetConfig()
}

// GetAppConfigManager 获取应用配置管理器
func GetAppConfigManager() *ViperConfigManager[AppConfig] {
	return GetViperConfigManager(AppConfig{})
}

// GetTemplateConfigManager 获取模板配置管理器
func GetTemplateConfigManager() *ViperConfigManager[TemplateConfig] {
	return GetViperConfigManager(TemplateConfig{})
}

// GetAuthConfigManager 获取认证配置管理器
func GetAuthConfigManager() *ViperConfigManager[AuthConfig] {
	return GetViperConfigManager(AuthConfig{})
}

// GetTLSConfigManager 获取TLS配置管理器
func GetTLSConfigManager() *ViperConfigManager[TLSServerConfig] {
	return GetViperConfigManager(TLSServerConfig{})
}

// GetLogConfigManager 获取日志配置管理器
func GetLogConfigManager() *ViperConfigManager[LogConfig] {
	return GetViperConfigManager(LogConfig{})
}

// GetDatabaseConfigManager 获取数据库配置管理器
func GetDatabaseConfigManager() *ViperConfigManager[DatabaseConfig] {
	return GetViperConfigManager(DatabaseConfig{})
}

// GetRedisConfigManager 获取Redis配置管理器
func GetRedisConfigManager() *ViperConfigManager[RedisConfig] {
	return GetViperConfigManager(RedisConfig{})
}

// GetMyBatisConfigManager 获取MyBatis配置管理器
func GetMyBatisConfigManager() *ViperConfigManager[MyBatisConfig] {
	return GetViperConfigManager(MyBatisConfig{})
}

// 泛型配置函数 - 主要API

// GetConfigValue 获取指定配置的值（泛型版本）
func GetConfigValue[T ConfigInterface](config T, key string, defaults ...any) any {
	manager := GetViperConfigManager(config)
	value := manager.Get(key)
	if value == nil && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// GetConfigString 获取指定配置的字符串值（泛型版本）
func GetConfigString[T ConfigInterface](config T, key string, defaults ...string) string {
	manager := GetViperConfigManager(config)
	value := manager.GetString(key)
	if value == "" && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// GetConfigInt 获取指定配置的整数值（泛型版本）
func GetConfigInt[T ConfigInterface](config T, key string, defaults ...int) int {
	manager := GetViperConfigManager(config)
	value := manager.GetInt(key)
	if value == 0 && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// GetConfigBool 获取指定配置的布尔值（泛型版本）
func GetConfigBool[T ConfigInterface](config T, key string, defaults ...bool) bool {
	manager := GetViperConfigManager(config)
	if !manager.IsSet(key) && len(defaults) > 0 {
		return defaults[0]
	}
	return manager.GetBool(key)
}

// GetConfigStringSlice 获取指定配置的字符串切片值（泛型版本）
func GetConfigStringSlice[T ConfigInterface](config T, key string, defaults ...[]string) []string {
	manager := GetViperConfigManager(config)
	value := manager.GetStringSlice(key)
	if len(value) == 0 && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// GetConfigIntSlice 获取指定配置的整数切片值（泛型版本）
func GetConfigIntSlice[T ConfigInterface](config T, key string, defaults ...[]int) []int {
	manager := GetViperConfigManager(config)
	value := manager.GetIntSlice(key)
	if len(value) == 0 && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// 使用类型方法注册表
var configNameRegistry = make(map[reflect.Type]string)

func RegisterConfigName[T ConfigInterface](name string) {
	var zero T
	t := reflect.TypeOf(zero)
	configNameRegistry[t] = name
}

// GetConfigBoolSlice 获取指定配置的布尔切片值（泛型版本）
func GetConfigBoolSlice[T ConfigInterface](config T, key string, defaults ...[]bool) []bool {
	manager := GetViperConfigManager(config)
	value := manager.GetBoolSlice(key)
	if len(value) == 0 && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// GetConfigFloat64Slice 获取指定配置的浮点数切片值（泛型版本）
func GetConfigFloat64Slice[T ConfigInterface](config T, key string, defaults ...[]float64) []float64 {
	manager := GetViperConfigManager(config)
	value := manager.GetFloat64Slice(key)
	if len(value) == 0 && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// GetConfigDurationSlice 获取指定配置的时间间隔切片值（泛型版本）
func GetConfigDurationSlice[T ConfigInterface](config T, key string, defaults ...[]time.Duration) []time.Duration {
	manager := GetViperConfigManager(config)
	value := manager.GetDurationSlice(key)
	if len(value) == 0 && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// GetConfigTimeSlice 获取指定配置的时间切片值（泛型版本）
func GetConfigTimeSlice[T ConfigInterface](config T, key string, defaults ...[]time.Time) []time.Time {
	manager := GetViperConfigManager(config)
	value := manager.GetTimeSlice(key)
	if len(value) == 0 && len(defaults) > 0 {
		return defaults[0]
	}
	return value
}

// SetConfigValue 设置指定配置的值（泛型版本）
func SetConfigValue[T ConfigInterface](config T, key string, value any) {
	manager := GetViperConfigManager(config)
	manager.Set(key, value)
}

// WatchConfig 监听指定配置文件变化（泛型版本）
func WatchConfig[T ConfigInterface](config T) {
	manager := GetViperConfigManager(config)
	manager.WatchConfig()
}

// 便捷的特定配置类型函数

// GetAppConfigInt 获取应用配置的整数值
func GetAppConfigInt(key string) int {
	return GetConfigInt(AppConfig{}, key)
}

// GetAppConfigString 获取应用配置的字符串值
func GetAppConfigString(key string) string {
	return GetConfigString(AppConfig{}, key)
}

// GetAppConfigBool 获取应用配置的布尔值
func GetAppConfigBool(key string) bool {
	return GetConfigBool(AppConfig{}, key)
}

// GetAppConfigStringSlice 获取应用配置的字符串切片值
func GetAppConfigStringSlice(key string) []string {
	return GetConfigStringSlice(AppConfig{}, key)
}

// GetAppConfigIntSlice 获取应用配置的整数切片值
func GetAppConfigIntSlice(key string) []int {
	return GetConfigIntSlice(AppConfig{}, key)
}

// GetAppConfigBoolSlice 获取应用配置的布尔切片值
func GetAppConfigBoolSlice(key string) []bool {
	return GetConfigBoolSlice(AppConfig{}, key)
}

// GetAppConfigFloat64Slice 获取应用配置的浮点数切片值
func GetAppConfigFloat64Slice(key string) []float64 {
	return GetConfigFloat64Slice(AppConfig{}, key)
}

// GetAppConfigDurationSlice 获取应用配置的时间间隔切片值
func GetAppConfigDurationSlice(key string) []time.Duration {
	return GetConfigDurationSlice(AppConfig{}, key)
}

// GetAppConfigTimeSlice 获取应用配置的时间切片值
func GetAppConfigTimeSlice(key string) []time.Time {
	return GetConfigTimeSlice(AppConfig{}, key)
}

// GetLogConfigString 获取日志配置的字符串值
func GetLogConfigString(key string) string {
	return GetConfigString(LogConfig{}, key)
}

// GetLogConfigInt 获取日志配置的整数值
func GetLogConfigInt(key string) int {
	return GetConfigInt(LogConfig{}, key)
}

// GetLogConfigBool 获取日志配置的布尔值
func GetLogConfigBool(key string) bool {
	return GetConfigBool(LogConfig{}, key)
}

// GetLogConfigStringSlice 获取日志配置的字符串切片值
func GetLogConfigStringSlice(key string) []string {
	return GetConfigStringSlice(LogConfig{}, key)
}

// 基于反射的新获取方式 - 支持默认值

// getConfigManagerByType 通过类型获取配置管理器（泛型辅助函数）
func getConfigManagerByType[T ConfigInterface]() (*ViperConfigManager[T], bool) {
	var zero T
	t := reflect.TypeOf(zero)
	// 从注册表获取配置名称
	configName, ok := configNameRegistry[t]
	if !ok {
		fmt.Printf("未注册的配置类型: %v\n", t)
		return nil, false
	}

	// 获取对应的Viper管理器
	vmg := GetViperConfigManagerWithName[T](configName)
	if vmg == nil {
		fmt.Printf("未找到配置管理器 - config_name: %s\n", configName)
		return nil, false
	}

	return vmg, true
}

// GetConfigStringWithDefaults 通过反射获取配置字符串值
func GetConfigStringWithDefaults[T ConfigInterface](key string, defaults ...string) string {
	var defaultValue string
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetString(key)
	if outValue == "" {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigIntWithDefaults 通过反射获取配置整数值
func GetConfigIntWithDefaults[T ConfigInterface](key string, defaults ...int) int {
	var defaultValue int
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetInt(key)
	if outValue == 0 && !vmg.IsSet(key) {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigBoolWithDefaults 通过反射获取配置布尔值
func GetConfigBoolWithDefaults[T ConfigInterface](key string, defaults ...bool) bool {
	var defaultValue bool
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	if !vmg.IsSet(key) {
		return defaultValue
	}
	return vmg.GetBool(key)
}

// GetConfigStringSliceWithDefaults 通过反射获取配置字符串切片值
func GetConfigStringSliceWithDefaults[T ConfigInterface](key string, defaults ...[]string) []string {
	var defaultValue []string
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetStringSlice(key)
	if len(outValue) == 0 {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigIntSliceWithDefaults 通过反射获取配置整数切片值
func GetConfigIntSliceWithDefaults[T ConfigInterface](key string, defaults ...[]int) []int {
	var defaultValue []int
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetIntSlice(key)
	if len(outValue) == 0 {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigFloat64WithDefaults 通过反射获取配置浮点数值
func GetConfigFloat64WithDefaults[T ConfigInterface](key string, defaults ...float64) float64 {
	var defaultValue float64
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetFloat64(key)
	if outValue == 0.0 && !vmg.IsSet(key) {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigDurationWithDefaults 通过反射获取配置时间间隔值
func GetConfigDurationWithDefaults[T ConfigInterface](key string, defaults ...time.Duration) time.Duration {
	var defaultValue time.Duration
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetDuration(key)
	if outValue == 0 && !vmg.IsSet(key) {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigBoolSliceWithDefaults 通过反射获取配置布尔切片值
func GetConfigBoolSliceWithDefaults[T ConfigInterface](key string, defaults ...[]bool) []bool {
	var defaultValue []bool
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetBoolSlice(key)
	if len(outValue) == 0 {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigFloat64SliceWithDefaults 通过反射获取配置浮点数切片值
func GetConfigFloat64SliceWithDefaults[T ConfigInterface](key string, defaults ...[]float64) []float64 {
	var defaultValue []float64
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetFloat64Slice(key)
	if len(outValue) == 0 {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigDurationSliceWithDefaults 通过反射获取配置时间间隔切片值
func GetConfigDurationSliceWithDefaults[T ConfigInterface](key string, defaults ...[]time.Duration) []time.Duration {
	var defaultValue []time.Duration
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetDurationSlice(key)
	if len(outValue) == 0 {
		outValue = defaultValue
	}
	return outValue
}

// GetConfigTimeSliceWithDefaults 通过反射获取配置时间切片值
func GetConfigTimeSliceWithDefaults[T ConfigInterface](key string, defaults ...[]time.Time) []time.Time {
	var defaultValue []time.Time
	if len(defaults) > 0 {
		defaultValue = defaults[0]
	}

	vmg, ok := getConfigManagerByType[T]()
	if !ok {
		return defaultValue
	}

	outValue := vmg.GetTimeSlice(key)
	if len(outValue) == 0 {
		outValue = defaultValue
	}
	return outValue
}
