package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	// 默认配置文件名和类型
	AppConfigName = "app"

	// 默认引擎配置名
	TemplateConfigName = "template"

	// 默认认证配置名
	AuthConfigName = "auth"

	// 默认日志配置名
	LogConfigName = "log"

	// 默认会话配置名
	SessionConfigName = "session"

	// 默认数据库配置名
	DatabaseConfigName = "database"

	TLSConfigName = "tls"

	RedisConfigName = "redis"

	// MyBatis配置名
	MyBatisConfigName = "mybatis"
)

// ConfigInterface 定义所有配置类型需要实现的接口
type ConfigInterface interface {
	GetConfigName() string
	SetDefaults(v *viper.Viper)
	GenerateDefaultContent() string
}

// ViperConfigManager 泛型Viper配置管理器
type ViperConfigManager[T ConfigInterface] struct {
	config      T // 泛型配置类型
	viper       *viper.Viper
	configPaths []string
	configName  string
	configType  string
	envPrefix   string
	initialized bool
	mu          sync.RWMutex
}

// 全局泛型配置管理器存储
var ConfigManagers sync.Map

// NewViperConfigManager 创建新的泛型配置管理器
func NewViperConfigManager[T ConfigInterface](config T) *ViperConfigManager[T] {
	configName := config.GetConfigName()
	if configName == "" {
		fmt.Printf("配置名称不能为空 - config: %v\n", config)
		Panic("配置名称不能为空")
	}
	// configPaths: []string{".", "./conf", "/etc/yyhertz", "$HOME/.yyhertz"},

	return &ViperConfigManager[T]{
		config:      config,
		viper:       viper.New(),
		configPaths: []string{"./conf"},
		configName:  configName,
		configType:  "yaml",
		envPrefix:   "YYHERTZ",
		initialized: false,
	}
}

// GetConfigManager 获取泛型配置管理器实例（单例模式）
func GetViperConfigManager[T ConfigInterface](config T) *ViperConfigManager[T] {
	configName := config.GetConfigName()

	if value, ok := ConfigManagers.Load(configName); ok {
		if manager, ok := value.(*ViperConfigManager[T]); ok {
			return manager
		}
	}

	manager := NewViperConfigManager(config)
	_ = manager.Initialize()
	ConfigManagers.Store(configName, manager)

	return manager
}

// 根据 configName获取配置管理器实例（兼容性函数）
func GetViperConfigManagerWithName[T ConfigInterface](name string) *ViperConfigManager[T] {
	if value, ok := ConfigManagers.Load(name); ok {
		return value.(*ViperConfigManager[T])
	}
	return nil
}

// Initialize 初始化配置管理器
func (gcm *ViperConfigManager[T]) Initialize() error {
	gcm.mu.Lock()
	defer gcm.mu.Unlock()

	return gcm.initializeInternal()
}

// ensureInitialized 确保配置已初始化（线程安全）
func (gcm *ViperConfigManager[T]) ensureInitialized() {
	if !gcm.initialized {
		_ = gcm.Initialize()
	}
}

// initializeInternal 内部初始化方法（不加锁）
func (gcm *ViperConfigManager[T]) initializeInternal() error {
	if gcm.initialized {
		return nil
	}

	// 设置配置文件名和类型
	gcm.viper.SetConfigName(gcm.configName)
	gcm.viper.SetConfigType(gcm.configType)

	// 添加配置文件搜索路径
	for _, path := range gcm.configPaths {
		gcm.viper.AddConfigPath(path)
	}

	// 设置环境变量前缀
	gcm.viper.SetEnvPrefix(gcm.envPrefix)
	gcm.viper.AutomaticEnv()
	gcm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 使用配置结构体设置默认值
	gcm.config.SetDefaults(gcm.viper)

	// 尝试读取配置文件
	if err := gcm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，使用默认配置并创建示例配置文件
			fmt.Printf("配置文件未找到，使用默认配置并尝试创建示例配置文件 - config_name: %s, config_type: %s, paths: %v\n",
				gcm.configName, gcm.configType, gcm.configPaths)

			if err := gcm.createDefaultConfigFile(); err != nil {
				fmt.Printf("创建默认配置文件失败 - error: %s\n", err.Error())
			}
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	} else {
		fmt.Printf("配置文件加载成功 - config_file: %s\n", gcm.viper.ConfigFileUsed())
	}

	gcm.initialized = true
	return nil
}

// createDefaultConfigFile 创建默认配置文件
func (gcm *ViperConfigManager[T]) createDefaultConfigFile() error {
	// 使用第一个配置路径，或默认使用 ./conf
	configDir := "./conf"
	if len(gcm.configPaths) > 1 {
		configDir = filepath.Join(gcm.configPaths[1])
	}
	configFile := filepath.Join(configDir, gcm.configName+"."+gcm.configType)

	// 确保配置目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 检查文件是否已存在
	if _, err := os.Stat(configFile); err == nil {
		return nil // 文件已存在，不需要创建
	}

	// 写入配置文件
	if err := os.WriteFile(configFile, []byte(gcm.config.GenerateDefaultContent()), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("默认配置文件创建成功 - config_file: %s\n", configFile)

	return nil
}

// GetConfig 获取完整配置
func (gcm *ViperConfigManager[T]) GetConfig() (*T, error) {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	var config T
	if err := gcm.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}

// Get 获取配置值
func (gcm *ViperConfigManager[T]) Get(key string) any {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.Get(key)
}

// GetString 获取字符串配置值
func (gcm *ViperConfigManager[T]) GetString(key string) string {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.GetString(key)
}

// GetInt 获取整数配置值
func (gcm *ViperConfigManager[T]) GetInt(key string) int {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.GetInt(key)
}

// GetIntSlice 获取整数数组配置值
func (cm *ViperConfigManager[T]) GetIntSlice(key string) []int {
	cm.ensureInitialized()

	// 尝试直接获取 []int
	if slice, ok := cm.Get(key).([]int); ok {
		return slice
	}

	// 尝试从其他类型转换
	value := cm.Get(key)
	if value == nil {
		return []int{} // 返回空切片
	}

	// 使用反射处理其他类型
	out, _ := convertToIntSlice(value)
	if out == nil {
		return []int{} // 返回空切片
	}
	return out
}

// GetStringSlice 获取字符串数组配置值
func (gcm *ViperConfigManager[T]) GetStringSlice(key string) []string {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.GetStringSlice(key)
}

// GetBool 获取布尔配置值
func (gcm *ViperConfigManager[T]) GetBool(key string) bool {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.GetBool(key)
}

func (cm *ViperConfigManager[T]) GetBoolSlice(key string) []bool {
	cm.ensureInitialized()

	// 尝试直接获取 []bool
	if slice, ok := cm.Get(key).([]bool); ok {
		return slice
	}

	// 尝试从 []any 转换
	if ifaceSlice, ok := cm.Get(key).([]any); ok {
		out, _ := convertInterfaceSliceToBool(ifaceSlice)
		if out == nil {
			return []bool{} // 返回空切片
		}
		return out
	}

	// 尝试从字符串解析
	if str, ok := cm.Get(key).(string); ok {
		out, _ := parseBoolSliceFromString(str)
		if out == nil {
			return []bool{} // 返回空切片
		}
		return out
	}

	// 尝试从其他类型转换
	value := cm.Get(key)
	if value == nil {
		return []bool{} // 返回空切片
	}

	// 使用反射处理其他类型
	out, _ := convertToBoolSlice(value)
	if out == nil {
		return []bool{} // 返回空切片
	}
	return out
}

// GetFloat64 获取浮点数配置值
func (gcm *ViperConfigManager[T]) GetFloat64(key string) float64 {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.GetFloat64(key)
}

// GetFloat64Slice 获取浮点数数组配置值
func (gcm *ViperConfigManager[T]) GetFloat64Slice(key string) []float64 {
	gcm.ensureInitialized()

	// 尝试直接获取 []float64
	if slice, ok := gcm.Get(key).([]float64); ok {
		return slice
	}

	// 尝试从 []any 转换
	if ifaceSlice, ok := gcm.Get(key).([]any); ok {
		result := make([]float64, 0, len(ifaceSlice))
		for _, v := range ifaceSlice {
			switch val := v.(type) {
			case float64:
				result = append(result, val)
			case float32:
				result = append(result, float64(val))
			case int:
				result = append(result, float64(val))
			case int64:
				result = append(result, float64(val))
			case string:
				if f, err := parseFloat(val); err == nil {
					result = append(result, f)
				}
			}
		}
		return result
	}

	return []float64{}
}

// GetDuration 获取时间间隔配置值
func (gcm *ViperConfigManager[T]) GetDuration(key string) time.Duration {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.GetDuration(key)
}

// GetDurationSlice 获取时间间隔数组配置值
func (gcm *ViperConfigManager[T]) GetDurationSlice(key string) []time.Duration {
	gcm.ensureInitialized()

	// 尝试直接获取 []time.Duration
	if slice, ok := gcm.Get(key).([]time.Duration); ok {
		return slice
	}

	// 尝试从 []any 转换
	if ifaceSlice, ok := gcm.Get(key).([]any); ok {
		result := make([]time.Duration, 0, len(ifaceSlice))
		for _, v := range ifaceSlice {
			switch val := v.(type) {
			case time.Duration:
				result = append(result, val)
			case string:
				if d, err := time.ParseDuration(val); err == nil {
					result = append(result, d)
				}
			case int64:
				result = append(result, time.Duration(val))
			case int:
				result = append(result, time.Duration(val))
			}
		}
		return result
	}

	// 尝试从 []string 转换
	if stringSlice := gcm.GetStringSlice(key); len(stringSlice) > 0 {
		result := make([]time.Duration, 0, len(stringSlice))
		for _, s := range stringSlice {
			if d, err := time.ParseDuration(s); err == nil {
				result = append(result, d)
			}
		}
		return result
	}

	return []time.Duration{}
}

// GetTime 获取时间配置值
func (gcm *ViperConfigManager[T]) GetTime(key string) time.Time {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.GetTime(key)
}

// GetTimeSlice 获取时间数组配置值
func (gcm *ViperConfigManager[T]) GetTimeSlice(key string) []time.Time {
	gcm.ensureInitialized()

	// 尝试直接获取 []time.Time
	if slice, ok := gcm.Get(key).([]time.Time); ok {
		return slice
	}

	// 尝试从 []any 转换
	if ifaceSlice, ok := gcm.Get(key).([]any); ok {
		result := make([]time.Time, 0, len(ifaceSlice))
		for _, v := range ifaceSlice {
			switch val := v.(type) {
			case time.Time:
				result = append(result, val)
			case string:
				// 尝试多种时间格式解析
				formats := []string{
					time.RFC3339,
					time.RFC3339Nano,
					"2006-01-02 15:04:05",
					"2006-01-02",
					"15:04:05",
				}
				for _, format := range formats {
					if t, err := time.Parse(format, val); err == nil {
						result = append(result, t)
						break
					}
				}
			}
		}
		return result
	}

	// 尝试从 []string 转换
	if stringSlice := gcm.GetStringSlice(key); len(stringSlice) > 0 {
		result := make([]time.Time, 0, len(stringSlice))
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02",
			"15:04:05",
		}
		for _, s := range stringSlice {
			for _, format := range formats {
				if t, err := time.Parse(format, s); err == nil {
					result = append(result, t)
					break
				}
			}
		}
		return result
	}

	return []time.Time{}
}

// Set 设置配置值
func (gcm *ViperConfigManager[T]) Set(key string, value any) {
	gcm.ensureInitialized()

	gcm.mu.Lock()
	defer gcm.mu.Unlock()

	gcm.viper.Set(key, value)
}

// IsSet 检查配置是否已设置
func (gcm *ViperConfigManager[T]) IsSet(key string) bool {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.IsSet(key)
}

// WatchConfig 监听配置文件变化
func (gcm *ViperConfigManager[T]) WatchConfig() {
	gcm.mu.Lock()
	defer gcm.mu.Unlock()

	if !gcm.initialized {
		// 直接调用内部初始化方法，避免重复加锁
		gcm.initializeInternal()
	}

	gcm.viper.WatchConfig()
	gcm.viper.OnConfigChange(func(e fsnotify.Event) {
		log.Printf("配置文件发生变化，重新加载 - file: %s, operation: %s", e.Name, e.Op.String())
		if err := gcm.viper.ReadInConfig(); err != nil {
			log.Printf("重新加载配置文件失败 - error: %s", err.Error())
			_, _ = gcm.GetConfig()
		} else {
			log.Printf("配置文件重新加载成功 - file: %s", gcm.viper.ConfigFileUsed())
		}
	})
}

// ConfigFileUsed 获取当前使用的配置文件路径
func (gcm *ViperConfigManager[T]) ConfigFileUsed() string {
	gcm.ensureInitialized()

	gcm.mu.RLock()
	defer gcm.mu.RUnlock()

	return gcm.viper.ConfigFileUsed()
}

// parseFloat 解析字符串为float64
func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// LoadConfigWithGeneric 泛型加载配置的便捷函数
func LoadConfigWithGeneric[T ConfigInterface](configName string) (T, error) {
	// 创建配置类型的零值实例
	var config T

	// 获取配置管理器
	manager := GetViperConfigManagerWithName[T](configName)
	if manager == nil {
		// 如果管理器不存在，创建一个新的
		manager = NewViperConfigManager(config)
		if err := manager.Initialize(); err != nil {
			return config, err
		}
		ConfigManagers.Store(configName, manager)
	}

	// 获取配置
	configPtr, err := manager.GetConfig()
	if err != nil {
		return config, err
	}

	return *configPtr, nil
}
