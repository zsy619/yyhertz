package mapper

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/zsy619/yyhertz/framework/util/mapper/strategies"
)

// FastMapper 高性能对象映射器实现
type FastMapper struct {
	mu       sync.RWMutex
	strategy MappingStrategy
	cache    *TypeCache
	config   *MapConfig
	stats    *MapperStats

	// 不同策略的映射器
	reflectionMapper *strategies.ReflectionMapper
	jsonMapper       *strategies.JSONMapper
	codegenMapper    *strategies.CodegenMapper
}

// NewMapper 创建新的映射器实例
func NewMapper(options ...MapperOption) Mapper {
	// 直接使用MapConfig
	config := &MapConfig{
		Strategy:      StrategyAuto,
		TagName:       "json",
		CaseSensitive: false,
		DeepCopy:      true,
		ZeroFields:    true,
		Converters:    make(map[reflect.Type]TypeConverter),
		MaxDepth:      10,
		IgnoreFields:  []string{},
	}

	// 应用配置选项
	for _, opt := range options {
		opt(config)
	}

	mapper := &FastMapper{
		strategy:         config.Strategy,
		cache:            NewTypeCache(),
		config:           config,
		stats:            &MapperStats{},
		reflectionMapper: strategies.NewReflectionMapper(),
		jsonMapper:       strategies.NewJSONMapper(),
		codegenMapper:    strategies.NewCodegenMapper(),
	}

	return mapper
}

// Map 将源对象映射到目标对象
func (m *FastMapper) Map(src, dst any) error {
	return m.MapWithConfig(src, dst, m.config)
}

// MapWithConfig 使用配置映射对象
func (m *FastMapper) MapWithConfig(src, dst any, config *MapConfig) error {
	if src == nil {
		return fmt.Errorf("source cannot be nil")
	}
	if dst == nil {
		return fmt.Errorf("destination cannot be nil")
	}

	// 检查目标是否为指针
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	// 更新统计
	m.mu.Lock()
	m.stats.TotalMaps++
	m.mu.Unlock()

	// 选择映射策略
	strategy := config.Strategy
	if strategy == StrategyAuto {
		strategy = m.selectOptimalStrategy(src, dst)
	}

	// 执行映射
	var err error
	switch strategy {
	case StrategyReflection:
		err = m.reflectionMapper.Map(src, dst, config)
	case StrategyJSON:
		err = m.jsonMapper.Map(src, dst, config)
	case StrategyCodegen:
		err = m.codegenMapper.Map(src, dst, config)
	default:
		err = fmt.Errorf("unsupported mapping strategy: %d", strategy)
	}

	// 更新统计
	m.mu.Lock()
	if err != nil {
		m.stats.FailedMaps++
	} else {
		m.stats.SuccessfulMaps++
	}
	m.mu.Unlock()

	return err
}

// MapSlice 批量映射切片
func (m *FastMapper) MapSlice(src, dst any) error {
	srcValue := reflect.ValueOf(src)
	dstValue := reflect.ValueOf(dst)

	// 检查源是否为切片
	if srcValue.Kind() != reflect.Slice {
		return fmt.Errorf("source must be a slice")
	}

	// 检查目标是否为切片指针
	if dstValue.Kind() != reflect.Ptr || dstValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("destination must be a pointer to slice")
	}

	srcLen := srcValue.Len()
	if srcLen == 0 {
		return nil
	}

	// 获取切片元素类型
	_ = srcValue.Type().Elem() // srcElemType 未使用
	dstElemType := dstValue.Elem().Type().Elem()

	// 创建目标切片
	dstSlice := reflect.MakeSlice(dstValue.Elem().Type(), srcLen, srcLen)

	// 逐个映射元素
	for i := 0; i < srcLen; i++ {
		srcElem := srcValue.Index(i).Interface()
		dstElem := reflect.New(dstElemType).Interface()

		if err := m.Map(srcElem, dstElem); err != nil {
			return fmt.Errorf("error mapping slice element at index %d: %w", i, err)
		}

		dstSlice.Index(i).Set(reflect.ValueOf(dstElem).Elem())
	}

	dstValue.Elem().Set(dstSlice)
	return nil
}

// SetStrategy 设置映射策略
func (m *FastMapper) SetStrategy(strategy MappingStrategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.strategy = strategy
	m.config.Strategy = strategy
}

// GetStats 获取性能统计
func (m *FastMapper) GetStats() *MapperStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回统计的副本
	return &MapperStats{
		TotalMaps:      m.stats.TotalMaps,
		SuccessfulMaps: m.stats.SuccessfulMaps,
		FailedMaps:     m.stats.FailedMaps,
		AverageTime:    m.stats.AverageTime,
		CacheHits:      m.stats.CacheHits,
		CacheMisses:    m.stats.CacheMisses,
	}
}

// selectOptimalStrategy 选择最优映射策略
func (m *FastMapper) selectOptimalStrategy(src, dst any) MappingStrategy {
	srcType := reflect.TypeOf(src)
	dstType := reflect.TypeOf(dst)

	// 对于指针类型，获取实际类型
	if srcType.Kind() == reflect.Ptr {
		srcType = srcType.Elem()
	}
	if dstType.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
	}

	// 如果是简单结构体且没有复杂嵌套，优先使用JSON策略
	if m.isSimpleStruct(srcType) && m.isSimpleStruct(dstType) {
		return StrategyJSON
	}

	// 如果有代码生成的映射器可用，优先使用
	if m.codegenMapper.HasMapping(srcType, dstType) {
		return StrategyCodegen
	}

	// 默认使用优化的反射映射器
	return StrategyReflection
}

// isSimpleStruct 判断是否为简单结构体（没有复杂嵌套）
func (m *FastMapper) isSimpleStruct(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	// 检查字段深度
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		switch field.Type.Kind() {
		case reflect.Struct:
			// 允许一层结构体嵌套
			return false
		case reflect.Slice, reflect.Array:
			// 检查切片/数组元素类型
			elemType := field.Type.Elem()
			if elemType.Kind() == reflect.Struct || elemType.Kind() == reflect.Ptr {
				return false
			}
		case reflect.Map:
			// Map类型认为是复杂类型
			return false
		case reflect.Interface:
			// 接口类型认为是复杂类型
			return false
		}
	}

	return true
}

// DefaultMapper 默认全局映射器实例
var DefaultMapper = NewMapper()

// Map 使用默认映射器进行映射
func Map(src, dst any) error {
	return DefaultMapper.Map(src, dst)
}

// MapWithConfig 使用配置和默认映射器进行映射
func MapWithConfig(src, dst any, config *MapConfig) error {
	return DefaultMapper.MapWithConfig(src, dst, config)
}

// MapSlice 使用默认映射器进行切片映射
func MapSlice(src, dst any) error {
	return DefaultMapper.MapSlice(src, dst)
}
