package xmapper

import (
	"reflect"
	"time"
)

// MapperConfig 映射器配置结构
type MapperConfig struct {
	// 映射策略选择
	Strategy MappingStrategy

	// 标签名称配置
	TagName string

	// 大小写敏感设置
	CaseSensitive bool

	// 深拷贝模式
	DeepCopy bool

	// 忽略的字段列表
	IgnoreFields []string

	// 缓存配置
	CacheConfig CacheConfig

	// 性能配置
	PerformanceConfig PerformanceConfig

	// 调试配置
	DebugConfig DebugConfig
}

// CacheConfig 缓存配置
type CacheConfig struct {
	// 启用缓存
	Enabled bool

	// 类型缓存大小
	TypeCacheSize int

	// 字段缓存大小
	FieldCacheSize int

	// 缓存TTL
	TTL time.Duration

	// 清理间隔
	CleanupInterval time.Duration

	// LRU淘汰阈值
	EvictionThreshold float64
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	// 启用性能统计
	EnableStats bool

	// 启用性能监控
	EnableMonitoring bool

	// 复杂度阈值（用于策略自动选择）
	ComplexityThreshold int

	// 批量处理阈值
	BatchThreshold int

	// 工作池大小
	WorkerPoolSize int

	// 超时设置
	Timeout time.Duration
}

// DebugConfig 调试配置
type DebugConfig struct {
	// 启用调试模式
	Enabled bool

	// 启用详细日志
	VerboseLogging bool

	// 启用性能跟踪
	EnableTracing bool

	// 启用内存分析
	EnableProfiling bool
}

// DefaultMapperConfig 返回默认映射器配置
func DefaultMapperConfig() *MapperConfig {
	return &MapperConfig{
		Strategy:      StrategyAuto,
		TagName:       "json",
		CaseSensitive: true,
		DeepCopy:      true,
		IgnoreFields:  []string{},
		CacheConfig: CacheConfig{
			Enabled:           true,
			TypeCacheSize:     1000,
			FieldCacheSize:    5000,
			TTL:               30 * time.Minute,
			CleanupInterval:   5 * time.Minute,
			EvictionThreshold: 0.8,
		},
		PerformanceConfig: PerformanceConfig{
			EnableStats:         true,
			EnableMonitoring:    false,
			ComplexityThreshold: 10,
			BatchThreshold:      100,
			WorkerPoolSize:      4,
			Timeout:             30 * time.Second,
		},
		DebugConfig: DebugConfig{
			Enabled:         false,
			VerboseLogging:  false,
			EnableTracing:   false,
			EnableProfiling: false,
		},
	}
}

// DefaultMapConfig 返回默认映射配置
func DefaultMapConfig() *MapConfig {
	return &MapConfig{
		Strategy:      StrategyAuto,
		TagName:       "json",
		CaseSensitive: true,
		DeepCopy:      true,
		ZeroFields:    true,
		Converters:    make(map[reflect.Type]TypeConverter),
		MaxDepth:      10,
		IgnoreFields:  []string{},
	}
}

// PresetConfig 预设配置类型
type PresetConfig string

const (
	// PresetDevelopment 开发环境预设
	PresetDevelopment PresetConfig = "development"
	// PresetProduction 生产环境预设
	PresetProduction PresetConfig = "production"
	// PresetTesting 测试环境预设
	PresetTesting PresetConfig = "testing"
	// PresetHighPerformance 高性能预设
	PresetHighPerformance PresetConfig = "high_performance"
	// PresetLowMemory 低内存预设
	PresetLowMemory PresetConfig = "low_memory"
)

// GetDevelopmentConfig 开发环境配置
func GetDevelopmentConfig() *MapperConfig {
	config := DefaultMapperConfig()
	config.DebugConfig.Enabled = true
	config.DebugConfig.VerboseLogging = true
	config.DebugConfig.EnableTracing = true
	config.PerformanceConfig.EnableMonitoring = true
	return config
}

// GetProductionConfig 生产环境配置
func GetProductionConfig() *MapperConfig {
	config := DefaultMapperConfig()
	config.DebugConfig.Enabled = false
	config.DebugConfig.VerboseLogging = false
	config.PerformanceConfig.EnableStats = true
	config.PerformanceConfig.EnableMonitoring = false
	config.CacheConfig.TypeCacheSize = 2000
	config.CacheConfig.FieldCacheSize = 10000
	return config
}

// GetTestingConfig 测试环境配置
func GetTestingConfig() *MapperConfig {
	config := DefaultMapperConfig()
	config.CacheConfig.Enabled = true
	config.PerformanceConfig.EnableStats = true
	config.DebugConfig.Enabled = false
	config.CacheConfig.TTL = 5 * time.Minute
	config.CacheConfig.CleanupInterval = 1 * time.Minute
	return config
}

// GetHighPerformanceConfig 高性能配置
func GetHighPerformanceConfig() *MapperConfig {
	config := DefaultMapperConfig()
	config.Strategy = StrategyJSON // 默认使用JSON策略获得更好性能
	config.CacheConfig.TypeCacheSize = 5000
	config.CacheConfig.FieldCacheSize = 20000
	config.CacheConfig.TTL = 60 * time.Minute
	config.PerformanceConfig.WorkerPoolSize = 8
	config.PerformanceConfig.BatchThreshold = 50
	config.DebugConfig.Enabled = false
	return config
}

// GetLowMemoryConfig 低内存配置
func GetLowMemoryConfig() *MapperConfig {
	config := DefaultMapperConfig()
	config.CacheConfig.TypeCacheSize = 200
	config.CacheConfig.FieldCacheSize = 1000
	config.CacheConfig.TTL = 5 * time.Minute
	config.CacheConfig.CleanupInterval = 2 * time.Minute
	config.CacheConfig.EvictionThreshold = 0.6
	config.PerformanceConfig.WorkerPoolSize = 2
	config.PerformanceConfig.EnableStats = false
	config.DebugConfig.Enabled = false
	return config
}

// ValidateConfig 验证配置有效性
func ValidateConfig(config *MapperConfig) error {
	if config.CacheConfig.TypeCacheSize <= 0 {
		config.CacheConfig.TypeCacheSize = 1000
	}

	if config.CacheConfig.FieldCacheSize <= 0 {
		config.CacheConfig.FieldCacheSize = 5000
	}

	if config.CacheConfig.TTL <= 0 {
		config.CacheConfig.TTL = 30 * time.Minute
	}

	if config.CacheConfig.CleanupInterval <= 0 {
		config.CacheConfig.CleanupInterval = 5 * time.Minute
	}

	if config.CacheConfig.EvictionThreshold <= 0 || config.CacheConfig.EvictionThreshold >= 1 {
		config.CacheConfig.EvictionThreshold = 0.8
	}

	if config.PerformanceConfig.ComplexityThreshold <= 0 {
		config.PerformanceConfig.ComplexityThreshold = 10
	}

	if config.PerformanceConfig.BatchThreshold <= 0 {
		config.PerformanceConfig.BatchThreshold = 100
	}

	if config.PerformanceConfig.WorkerPoolSize <= 0 {
		config.PerformanceConfig.WorkerPoolSize = 4
	}

	if config.PerformanceConfig.Timeout <= 0 {
		config.PerformanceConfig.Timeout = 30 * time.Second
	}

	if config.TagName == "" {
		config.TagName = "json"
	}

	return nil
}

// Clone 克隆配置
func (c *MapperConfig) Clone() *MapperConfig {
	clone := *c

	// 深拷贝切片
	clone.IgnoreFields = make([]string, len(c.IgnoreFields))
	copy(clone.IgnoreFields, c.IgnoreFields)

	return &clone
}

// IsDebugMode 检查是否为调试模式
func (c *MapperConfig) IsDebugMode() bool {
	return c.DebugConfig.Enabled
}

// IsCacheEnabled 检查缓存是否启用
func (c *MapperConfig) IsCacheEnabled() bool {
	return c.CacheConfig.Enabled
}

// IsStatsEnabled 检查统计是否启用
func (c *MapperConfig) IsStatsEnabled() bool {
	return c.PerformanceConfig.EnableStats
}

// GetIgnoreFieldsMap 获取忽略字段的快速查找Map
func (c *MapperConfig) GetIgnoreFieldsMap() map[string]bool {
	ignoreMap := make(map[string]bool, len(c.IgnoreFields))
	for _, field := range c.IgnoreFields {
		ignoreMap[field] = true
	}
	return ignoreMap
}

// ShouldIgnoreField 检查是否应该忽略指定字段
func (c *MapperConfig) ShouldIgnoreField(fieldName string) bool {
	for _, ignored := range c.IgnoreFields {
		if ignored == fieldName {
			return true
		}
	}
	return false
}

// GetEffectiveStrategy 获取有效的映射策略
func (c *MapperConfig) GetEffectiveStrategy(complexity int) MappingStrategy {
	if c.Strategy == StrategyAuto {
		if complexity <= c.PerformanceConfig.ComplexityThreshold {
			return StrategyJSON
		}
		return StrategyReflection
	}
	return c.Strategy
}

// WithStrategy 设置策略
func (c *MapConfig) WithStrategy(strategy MappingStrategy) *MapConfig {
	c.Strategy = strategy
	return c
}

// WithIgnoreFields 设置忽略字段
func (c *MapConfig) WithIgnoreFields(fields ...string) *MapConfig {
	c.IgnoreFields = fields
	return c
}

// WithTagName 设置标签名
func (c *MapConfig) WithTagName(tag string) *MapConfig {
	c.TagName = tag
	return c
}

// WithCaseSensitive 设置大小写敏感
func (c *MapConfig) WithCaseSensitive(sensitive bool) *MapConfig {
	c.CaseSensitive = sensitive
	return c
}

// WithDeepCopy 设置深拷贝
func (c *MapConfig) WithDeepCopy(deep bool) *MapConfig {
	c.DeepCopy = deep
	return c
}

// WithZeroFields 设置零值字段映射
func (c *MapConfig) WithZeroFields(zero bool) *MapConfig {
	c.ZeroFields = zero
	return c
}

// WithMaxDepth 设置最大深度
func (c *MapConfig) WithMaxDepth(depth int) *MapConfig {
	c.MaxDepth = depth
	return c
}

// WithConverter 添加类型转换器
func (c *MapConfig) WithConverter(t reflect.Type, converter TypeConverter) *MapConfig {
	if c.Converters == nil {
		c.Converters = make(map[reflect.Type]TypeConverter)
	}
	c.Converters[t] = converter
	return c
}
