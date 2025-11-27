package xmapper

import (
	"reflect"
	"strings"
	"sync"
	"time"
)

// TypeCache 类型信息缓存系统
type TypeCache struct {
	mu              sync.RWMutex
	structInfoCache map[reflect.Type]*CachedStructInfo // 结构体信息缓存
	fieldMapCache   map[string]*CachedFieldMapping     // 字段映射缓存
	conversionCache map[string]TypeConverter           // 类型转换器缓存
	accessStats     map[string]*AccessStats            // 访问统计
	cacheHits       int64                              // 缓存命中次数
	cacheMisses     int64                              // 缓存未命中次数
	maxCacheSize    int                                // 最大缓存大小
	cleanupInterval time.Duration                      // 清理间隔
	lastCleanup     time.Time                          // 上次清理时间
}

// CachedStructInfo 缓存的结构体信息
type CachedStructInfo struct {
	Type           reflect.Type
	Fields         []CachedFieldInfo
	FieldMap       map[string]int
	TagMap         map[string]map[string]string // tagName -> fieldName -> tagValue
	JSONMappable   bool
	IsSimpleStruct bool
	HasPointers    bool
	HasSlices      bool
	HasMaps        bool
	CreatedAt      time.Time
	LastAccessed   time.Time
	AccessCount    int64
}

// CachedFieldInfo 缓存的字段信息
type CachedFieldInfo struct {
	Field       reflect.StructField
	Index       []int
	Tags        map[string]string
	IsExported  bool
	IsEmbedded  bool
	HasJSON     bool
	JSONName    string
	JSONOptions []string
	TypeInfo    *TypeInfo
}

// CachedFieldMapping 缓存的字段映射
type CachedFieldMapping struct {
	SrcType         reflect.Type
	DstType         reflect.Type
	SrcField        reflect.StructField
	DstField        reflect.StructField
	SrcIndex        []int
	DstIndex        []int
	NeedsConversion bool
	Converter       TypeConverter
	ConversionCost  int // 转换成本估算
	CreatedAt       time.Time
	LastUsed        time.Time
	UseCount        int64
}

// AccessStats 访问统计
type AccessStats struct {
	Key        string
	HitCount   int64
	MissCount  int64
	TotalTime  time.Duration
	AvgTime    time.Duration
	LastAccess time.Time
	CreatedAt  time.Time
}

// TypeInfo 类型信息
type TypeInfo struct {
	Kind           reflect.Kind
	Size           uintptr
	Align          int
	FieldAlign     int
	IsComparable   bool
	IsBasicType    bool
	IsNumericType  bool
	IsStringType   bool
	IsTimeType     bool
	ConversionCost int
}

// NewTypeCache 创建类型缓存
func NewTypeCache() *TypeCache {
	cache := &TypeCache{
		structInfoCache: make(map[reflect.Type]*CachedStructInfo),
		fieldMapCache:   make(map[string]*CachedFieldMapping),
		conversionCache: make(map[string]TypeConverter),
		accessStats:     make(map[string]*AccessStats),
		maxCacheSize:    10000, // 默认最大缓存10000条
		cleanupInterval: 5 * time.Minute,
		lastCleanup:     time.Now(),
	}

	// 启动后台清理协程
	go cache.backgroundCleanup()

	return cache
}

// GetStructInfo 获取结构体信息（使用缓存）
func (tc *TypeCache) GetStructInfo(t reflect.Type, tagName string) *CachedStructInfo {
	tc.mu.RLock()
	if info, exists := tc.structInfoCache[t]; exists {
		// 更新访问统计
		info.LastAccessed = time.Now()
		info.AccessCount++
		tc.cacheHits++
		tc.mu.RUnlock()
		return info
	}
	tc.mu.RUnlock()

	// 缓存未命中，创建新的结构体信息
	tc.mu.Lock()
	defer tc.mu.Unlock()

	// 双重检查
	if info, exists := tc.structInfoCache[t]; exists {
		info.LastAccessed = time.Now()
		info.AccessCount++
		tc.cacheHits++
		return info
	}

	tc.cacheMisses++

	// 创建新的结构体信息
	info := tc.buildStructInfo(t, tagName)

	// 检查缓存大小限制
	if len(tc.structInfoCache) >= tc.maxCacheSize {
		tc.evictOldestEntries()
	}

	tc.structInfoCache[t] = info
	return info
}

// CacheStats 缓存统计信息
type CacheStats struct {
	CacheHits         int64
	CacheMisses       int64
	HitRatio          float64
	StructInfoCount   int
	FieldMappingCount int
	ConverterCount    int
	MaxCacheSize      int
	LastCleanup       time.Time
}

// GetStats 获取缓存统计信息
func (tc *TypeCache) GetStats() *CacheStats {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	return &CacheStats{
		CacheHits:         tc.cacheHits,
		CacheMisses:       tc.cacheMisses,
		HitRatio:          tc.calculateHitRatio(),
		StructInfoCount:   len(tc.structInfoCache),
		FieldMappingCount: len(tc.fieldMapCache),
		ConverterCount:    len(tc.conversionCache),
		MaxCacheSize:      tc.maxCacheSize,
		LastCleanup:       tc.lastCleanup,
	}
}

// Clear 清空缓存
func (tc *TypeCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.structInfoCache = make(map[reflect.Type]*CachedStructInfo)
	tc.fieldMapCache = make(map[string]*CachedFieldMapping)
	tc.conversionCache = make(map[string]TypeConverter)
	tc.accessStats = make(map[string]*AccessStats)
	tc.cacheHits = 0
	tc.cacheMisses = 0
}

// 私有方法实现
func (tc *TypeCache) buildStructInfo(t reflect.Type, tagName string) *CachedStructInfo {
	now := time.Now()

	info := &CachedStructInfo{
		Type:         t,
		Fields:       make([]CachedFieldInfo, 0, t.NumField()),
		FieldMap:     make(map[string]int),
		TagMap:       make(map[string]map[string]string),
		CreatedAt:    now,
		LastAccessed: now,
		AccessCount:  1,
	}

	// 初始化标签映射
	if tagName != "" {
		info.TagMap[tagName] = make(map[string]string)
	}

	// 分析字段
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldInfo := CachedFieldInfo{
			Field:      field,
			Index:      field.Index,
			Tags:       make(map[string]string),
			IsExported: field.IsExported(),
			IsEmbedded: field.Anonymous,
			TypeInfo:   tc.analyzeType(field.Type),
		}

		// 处理标签
		if tagName != "" {
			if tagValue := field.Tag.Get(tagName); tagValue != "" {
				fieldInfo.Tags[tagName] = tagValue

				// 解析JSON标签
				if tagName == "json" {
					fieldInfo.HasJSON = true
					parts := strings.Split(tagValue, ",")
					if len(parts) > 0 && parts[0] != "" {
						fieldInfo.JSONName = parts[0]
						if len(parts) > 1 {
							fieldInfo.JSONOptions = parts[1:]
						}

						// 添加到标签映射
						info.TagMap[tagName][parts[0]] = field.Name
					}
				}
			}
		}

		// 更新结构体特征
		switch field.Type.Kind() {
		case reflect.Ptr:
			info.HasPointers = true
		case reflect.Slice, reflect.Array:
			info.HasSlices = true
		case reflect.Map:
			info.HasMaps = true
		}

		info.Fields = append(info.Fields, fieldInfo)
		info.FieldMap[field.Name] = i
	}

	// 分析结构体特征
	info.JSONMappable = tc.isJSONMappable(t)
	info.IsSimpleStruct = tc.isSimpleStruct(t)

	return info
}

func (tc *TypeCache) analyzeType(t reflect.Type) *TypeInfo {
	info := &TypeInfo{
		Kind:         t.Kind(),
		Size:         t.Size(),
		Align:        t.Align(),
		FieldAlign:   t.FieldAlign(),
		IsComparable: t.Comparable(),
	}

	// 分析类型特征
	switch t.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		info.IsBasicType = true

		if t.Kind() != reflect.Bool && t.Kind() != reflect.String {
			info.IsNumericType = true
		}

		if t.Kind() == reflect.String {
			info.IsStringType = true
		}

		info.ConversionCost = 1

	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			info.IsTimeType = true
			info.ConversionCost = 3
		} else {
			info.ConversionCost = 5
		}

	case reflect.Slice, reflect.Array:
		info.ConversionCost = 10

	case reflect.Map:
		info.ConversionCost = 15

	case reflect.Ptr:
		info.ConversionCost = 2

	default:
		info.ConversionCost = 20
	}

	return info
}

func (tc *TypeCache) calculateHitRatio() float64 {
	total := tc.cacheHits + tc.cacheMisses
	if total == 0 {
		return 0
	}
	return float64(tc.cacheHits) / float64(total)
}

func (tc *TypeCache) isJSONMappable(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true

	case reflect.Struct:
		if t == reflect.TypeOf(time.Time{}) {
			return true
		}
		// 检查结构体字段
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() && !tc.isJSONMappable(field.Type) {
				return false
			}
		}
		return true

	case reflect.Slice, reflect.Array:
		return tc.isJSONMappable(t.Elem())

	case reflect.Map:
		return t.Key().Kind() == reflect.String && tc.isJSONMappable(t.Elem())

	case reflect.Ptr:
		return tc.isJSONMappable(t.Elem())

	case reflect.Interface:
		return true

	default:
		return false
	}
}

func (tc *TypeCache) isSimpleStruct(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	// 字段数量限制
	if t.NumField() > 20 {
		return false
	}

	// 检查字段类型复杂度
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !tc.isSimpleType(field.Type) {
			return false
		}
	}

	return true
}

func (tc *TypeCache) isSimpleType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true

	case reflect.Struct:
		return t == reflect.TypeOf(time.Time{})

	case reflect.Ptr:
		return tc.isSimpleType(t.Elem())

	default:
		return false
	}
}

func (tc *TypeCache) evictOldestEntries() {
	// 简单的LRU淘汰策略：删除最旧的20%条目
	toDelete := len(tc.structInfoCache) / 5
	if toDelete == 0 {
		toDelete = 1
	}

	// 找出最旧的条目
	var oldest []*struct {
		Type reflect.Type
		Time time.Time
	}

	for t, info := range tc.structInfoCache {
		oldest = append(oldest, &struct {
			Type reflect.Type
			Time time.Time
		}{t, info.LastAccessed})
	}

	// 删除最旧的条目
	for i := 0; i < toDelete && i < len(oldest); i++ {
		delete(tc.structInfoCache, oldest[i].Type)
	}
}

func (tc *TypeCache) backgroundCleanup() {
	ticker := time.NewTicker(tc.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		tc.performCleanup()
	}
}

func (tc *TypeCache) performCleanup() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-1 * time.Hour) // 删除1小时未访问的条目

	// 清理结构体信息缓存
	for t, info := range tc.structInfoCache {
		if info.LastAccessed.Before(cutoff) {
			delete(tc.structInfoCache, t)
		}
	}

	tc.lastCleanup = now
}
