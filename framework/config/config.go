package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// Config 配置管理器
type ConfigManager struct {
	configs map[string]string
	mutex   sync.RWMutex
}

// NewConfigManager 创建配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs: make(map[string]string),
	}
}

// Set 设置配置项
func (c *ConfigManager) Set(key, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configs[key] = value
}

// Get 获取配置项
func (c *ConfigManager) Get(key, defaultValue string) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	if value, exists := c.configs[key]; exists {
		return value
	}
	
	// 尝试从环境变量获取
	if envValue := os.Getenv(strings.ToUpper(strings.ReplaceAll(key, ".", "_"))); envValue != "" {
		return envValue
	}
	
	return defaultValue
}

// GetBool 获取布尔类型配置
func (c *ConfigManager) GetBool(key string, defaultValue bool) bool {
	value := c.Get(key, "")
	if value == "" {
		return defaultValue
	}
	
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	
	return defaultValue
}

// GetInt 获取整型配置
func (c *ConfigManager) GetInt(key string, defaultValue int) int {
	value := c.Get(key, "")
	if value == "" {
		return defaultValue
	}
	
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	
	return defaultValue
}

// GetInt64 获取int64类型配置
func (c *ConfigManager) GetInt64(key string, defaultValue int64) int64 {
	value := c.Get(key, "")
	if value == "" {
		return defaultValue
	}
	
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}
	
	return defaultValue
}

// GetFloat64 获取float64类型配置
func (c *ConfigManager) GetFloat64(key string, defaultValue float64) float64 {
	value := c.Get(key, "")
	if value == "" {
		return defaultValue
	}
	
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}
	
	return defaultValue
}

// GetSlice 获取切片类型配置(逗号分隔)
func (c *ConfigManager) GetSlice(key string, defaultValue []string) []string {
	value := c.Get(key, "")
	if value == "" {
		return defaultValue
	}
	
	return strings.Split(value, ",")
}

// LoadFromEnv 从环境变量加载配置
func (c *ConfigManager) LoadFromEnv(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) != 2 {
			continue
		}
		
		key := pair[0]
		value := pair[1]
		
		// 如果指定了前缀，只加载匹配的环境变量
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			continue
		}
		
		// 转换环境变量名为配置key格式
		configKey := strings.ToLower(strings.ReplaceAll(key, "_", "."))
		if prefix != "" {
			configKey = strings.TrimPrefix(configKey, strings.ToLower(prefix)+".")
		}
		
		c.configs[configKey] = value
	}
}

// LoadFromMap 从map加载配置
func (c *ConfigManager) LoadFromMap(data map[string]string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	for key, value := range data {
		c.configs[key] = value
	}
}

// GetAll 获取所有配置
func (c *ConfigManager) GetAll() map[string]string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	result := make(map[string]string)
	for k, v := range c.configs {
		result[k] = v
	}
	
	return result
}

// Exists 检查配置项是否存在
func (c *ConfigManager) Exists(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	_, exists := c.configs[key]
	return exists
}

// Delete 删除配置项
func (c *ConfigManager) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.configs, key)
}

// Clear 清空所有配置
func (c *ConfigManager) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.configs = make(map[string]string)
}

// 全局配置管理器实例
var defaultConfig = NewConfigManager()

// 便捷函数
func Set(key, value string) {
	defaultConfig.Set(key, value)
}

func Get(key, defaultValue string) string {
	return defaultConfig.Get(key, defaultValue)
}

func GetBool(key string, defaultValue bool) bool {
	return defaultConfig.GetBool(key, defaultValue)
}

func GetInt(key string, defaultValue int) int {
	return defaultConfig.GetInt(key, defaultValue)
}

func GetInt64(key string, defaultValue int64) int64 {
	return defaultConfig.GetInt64(key, defaultValue)
}

func GetFloat64(key string, defaultValue float64) float64 {
	return defaultConfig.GetFloat64(key, defaultValue)
}

func GetSlice(key string, defaultValue []string) []string {
	return defaultConfig.GetSlice(key, defaultValue)
}

func LoadFromEnv(prefix string) {
	defaultConfig.LoadFromEnv(prefix)
}

func LoadFromMap(data map[string]string) {
	defaultConfig.LoadFromMap(data)
}

func GetAll() map[string]string {
	return defaultConfig.GetAll()
}

func Exists(key string) bool {
	return defaultConfig.Exists(key)
}

func Delete(key string) {
	defaultConfig.Delete(key)
}

func Clear() {
	defaultConfig.Clear()
}

// 初始化默认配置
func init() {
	// 设置一些默认配置
	defaultConfig.Set("app.name", "Hertz MVC")
	defaultConfig.Set("app.version", "1.0.0")
	defaultConfig.Set("server.port", "8888")
	defaultConfig.Set("server.host", "0.0.0.0")
	
	// 从环境变量加载配置
	defaultConfig.LoadFromEnv("")
}