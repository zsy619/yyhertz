package mapper

import (
	"reflect"
)

// MappingStrategy 映射策略类型
type MappingStrategy int

const (
	// StrategyAuto 自动选择最优策略
	StrategyAuto MappingStrategy = iota
	// StrategyReflection 使用反射映射
	StrategyReflection
	// StrategyJSON 使用JSON转换映射
	StrategyJSON
	// StrategyCodegen 使用代码生成映射
	StrategyCodegen
)

// Mapper 高性能对象映射器接口
type Mapper interface {
	// Map 将源对象映射到目标对象
	Map(src, dst any) error
	// MapWithConfig 使用配置映射对象
	MapWithConfig(src, dst any, config *MapConfig) error
	// MapSlice 批量映射切片
	MapSlice(src, dst any) error
	// SetStrategy 设置映射策略
	SetStrategy(strategy MappingStrategy)
	// GetStats 获取性能统计
	GetStats() *MapperStats
}

// MapConfig 映射配置
type MapConfig struct {
	// Strategy 映射策略
	Strategy MappingStrategy
	// IgnoreFields 忽略的字段列表
	IgnoreFields []string
	// TagName 标签名称，默认"json"
	TagName string
	// CaseSensitive 是否大小写敏感
	CaseSensitive bool
	// DeepCopy 是否深拷贝
	DeepCopy bool
	// ZeroFields 是否映射零值字段
	ZeroFields bool
	// Converters 自定义类型转换器
	Converters map[reflect.Type]TypeConverter
	// MaxDepth 最大递归深度，防止循环引用
	MaxDepth int
}

// TypeConverter 自定义类型转换器
type TypeConverter func(src any) (any, error)

// MapperStats 映射器性能统计
type MapperStats struct {
	// TotalMaps 总映射次数
	TotalMaps int64
	// SuccessfulMaps 成功映射次数
	SuccessfulMaps int64
	// FailedMaps 失败映射次数
	FailedMaps int64
	// AverageTime 平均执行时间（纳秒）
	AverageTime int64
	// CacheHits 缓存命中次数
	CacheHits int64
	// CacheMisses 缓存未命中次数
	CacheMisses int64
}

// MapperOption 映射器配置选项
type MapperOption func(*MapConfig)

// WithConverter 添加自定义类型转换器
func WithConverter(srcType reflect.Type, converter TypeConverter) MapperOption {
	return func(config *MapConfig) {
		config.Converters[srcType] = converter
	}
}

// WithStrategy 设置映射策略
func WithStrategy(strategy MappingStrategy) MapperOption {
	return func(config *MapConfig) {
		config.Strategy = strategy
	}
}

// WithTagName 设置标签名称
func WithTagName(tagName string) MapperOption {
	return func(config *MapConfig) {
		config.TagName = tagName
	}
}

// WithIgnoreFields 设置忽略字段
func WithIgnoreFields(fields ...string) MapperOption {
	return func(config *MapConfig) {
		config.IgnoreFields = append(config.IgnoreFields, fields...)
	}
}

// WithCaseSensitive 设置大小写敏感
func WithCaseSensitive(sensitive bool) MapperOption {
	return func(config *MapConfig) {
		config.CaseSensitive = sensitive
	}
}

// WithConfig 设置完整配置
func WithConfig(config *MapConfig) MapperOption {
	return func(c *MapConfig) {
		*c = *config
	}
}

// WithDeepCopy 设置深拷贝模式
func WithDeepCopy(deep bool) MapperOption {
	return func(config *MapConfig) {
		config.DeepCopy = deep
	}
}

// WithMaxDepth 设置最大递归深度
func WithMaxDepth(maxDepth int) MapperOption {
	return func(config *MapConfig) {
		config.MaxDepth = maxDepth
	}
}
