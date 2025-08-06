// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// CachedORM 带缓存的ORM
type CachedORM struct {
	*ORM
	cacheManager *CacheManager
	cacheConfig  *CachedORMConfig
}

// CachedORMConfig 带缓存的ORM配置
type CachedORMConfig struct {
	// 是否启用缓存
	Enabled bool `json:"enabled" yaml:"enabled"`
	// 默认过期时间
	DefaultExpiration time.Duration `json:"default_expiration" yaml:"default_expiration"`
	// 是否缓存查询结果
	CacheQueries bool `json:"cache_queries" yaml:"cache_queries"`
	// 是否缓存单条记录
	CacheSingleRecords bool `json:"cache_single_records" yaml:"cache_single_records"`
	// 是否自动清除缓存
	AutoInvalidate bool `json:"auto_invalidate" yaml:"auto_invalidate"`
	// 缓存键前缀
	KeyPrefix string `json:"key_prefix" yaml:"key_prefix"`
	// 缓存命中日志
	LogCacheHits bool `json:"log_cache_hits" yaml:"log_cache_hits"`
}

// DefaultCachedORMConfig 默认带缓存的ORM配置
func DefaultCachedORMConfig() *CachedORMConfig {
	return &CachedORMConfig{
		Enabled:            true,
		DefaultExpiration:  time.Minute * 5,
		CacheQueries:       true,
		CacheSingleRecords: true,
		AutoInvalidate:     true,
		KeyPrefix:          "orm:",
		LogCacheHits:       true,
	}
}

// NewCachedORM 创建带缓存的ORM
func NewCachedORM(orm *ORM, cacheManager *CacheManager, config *CachedORMConfig) *CachedORM {
	if orm == nil {
		orm = GetDefaultORM()
	}

	if cacheManager == nil {
		cacheManager = GetGlobalCacheManager()
	}

	if config == nil {
		config = DefaultCachedORMConfig()
	}

	return &CachedORM{
		ORM:          orm,
		cacheManager: cacheManager,
		cacheConfig:  config,
	}
}

// buildCacheKey 构建缓存键
func (c *CachedORM) buildCacheKey(model interface{}, id interface{}) string {
	modelType := reflect.TypeOf(model)
	modelName := modelType.String()
	if modelType.Kind() == reflect.Ptr {
		modelName = modelType.Elem().String()
	}

	return fmt.Sprintf("%s%s:%v", c.cacheConfig.KeyPrefix, modelName, id)
}

// buildQueryCacheKey 构建查询缓存键
func (c *CachedORM) buildQueryCacheKey(query string, args ...interface{}) string {
	argsStr := ""
	for _, arg := range args {
		argsStr += fmt.Sprintf(":%v", arg)
	}
	return fmt.Sprintf("%squery:%s%s", c.cacheConfig.KeyPrefix, query, argsStr)
}

// First 查询第一条记录（带缓存）
func (c *CachedORM) First(dest interface{}, conds ...interface{}) error {
	if !c.cacheConfig.Enabled || !c.cacheConfig.CacheSingleRecords {
		return c.ORM.DB().First(dest, conds...).Error
	}

	// 尝试从缓存获取
	var id interface{}
	if len(conds) > 0 {
		id = conds[0]
	} else {
		// 如果没有提供条件，尝试从目标对象获取ID
		val := reflect.ValueOf(dest)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
			if val.Kind() == reflect.Struct {
				idField := val.FieldByName("ID")
				if idField.IsValid() && idField.CanInterface() {
					id = idField.Interface()
				}
			}
		}
	}

	if id != nil {
		cacheKey := c.buildCacheKey(dest, id)
		found, err := c.cacheManager.Get(cacheKey, dest)
		if err != nil {
			config.Warnf("从缓存获取记录失败: %v", err)
		} else if found {
			if c.cacheConfig.LogCacheHits {
				config.Debugf("缓存命中: %s", cacheKey)
			}
			return nil
		}
	}

	// 缓存未命中，从数据库查询
	if err := c.ORM.DB().First(dest, conds...).Error; err != nil {
		return err
	}

	// 缓存查询结果
	if id != nil {
		cacheKey := c.buildCacheKey(dest, id)
		if err := c.cacheManager.Set(cacheKey, dest, c.cacheConfig.DefaultExpiration); err != nil {
			config.Warnf("缓存记录失败: %v", err)
		}
	}

	return nil
}

// Find 查询多条记录（带缓存）
func (c *CachedORM) Find(dest interface{}, conds ...interface{}) error {
	if !c.cacheConfig.Enabled || !c.cacheConfig.CacheQueries {
		return c.ORM.DB().Find(dest, conds...).Error
	}

	// 构建缓存键
	query := ""
	args := make([]interface{}, 0)
	if len(conds) > 0 {
		query, _ = conds[0].(string)
		if len(conds) > 1 {
			args = conds[1:]
		}
	}

	cacheKey := c.buildQueryCacheKey(query, args...)
	found, err := c.cacheManager.Get(cacheKey, dest)
	if err != nil {
		config.Warnf("从缓存获取查询结果失败: %v", err)
	} else if found {
		if c.cacheConfig.LogCacheHits {
			config.Debugf("查询缓存命中: %s", cacheKey)
		}
		return nil
	}

	// 缓存未命中，从数据库查询
	if err := c.ORM.DB().Find(dest, conds...).Error; err != nil {
		return err
	}

	// 缓存查询结果
	if err := c.cacheManager.Set(cacheKey, dest, c.cacheConfig.DefaultExpiration); err != nil {
		config.Warnf("缓存查询结果失败: %v", err)
	}

	return nil
}

// Create 创建记录（自动清除缓存）
func (c *CachedORM) Create(value interface{}) error {
	err := c.ORM.DB().Create(value).Error
	if err == nil && c.cacheConfig.Enabled && c.cacheConfig.AutoInvalidate {
		// 获取ID
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
			if val.Kind() == reflect.Struct {
				idField := val.FieldByName("ID")
				if idField.IsValid() && idField.CanInterface() {
					id := idField.Interface()
					cacheKey := c.buildCacheKey(value, id)
					if err := c.cacheManager.Delete(cacheKey); err != nil {
						config.Warnf("清除缓存失败: %v", err)
					}
				}
			}
		}
	}
	return err
}

// Save 保存记录（自动清除缓存）
func (c *CachedORM) Save(value interface{}) error {
	err := c.ORM.DB().Save(value).Error
	if err == nil && c.cacheConfig.Enabled && c.cacheConfig.AutoInvalidate {
		// 获取ID
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Ptr && !val.IsNil() {
			val = val.Elem()
			if val.Kind() == reflect.Struct {
				idField := val.FieldByName("ID")
				if idField.IsValid() && idField.CanInterface() {
					id := idField.Interface()
					cacheKey := c.buildCacheKey(value, id)
					if err := c.cacheManager.Delete(cacheKey); err != nil {
						config.Warnf("清除缓存失败: %v", err)
					}
				}
			}
		}
	}
	return err
}

// Delete 删除记录（自动清除缓存）
func (c *CachedORM) Delete(value interface{}, conds ...interface{}) error {
	// 先获取ID
	var id interface{}
	if len(conds) > 0 {
		id = conds[0]
	}

	err := c.ORM.DB().Delete(value, conds...).Error
	if err == nil && c.cacheConfig.Enabled && c.cacheConfig.AutoInvalidate && id != nil {
		cacheKey := c.buildCacheKey(value, id)
		if err := c.cacheManager.Delete(cacheKey); err != nil {
			config.Warnf("清除缓存失败: %v", err)
		}
	}
	return err
}

// ClearCache 清空缓存
func (c *CachedORM) ClearCache() error {
	if !c.cacheConfig.Enabled {
		return nil
	}
	return c.cacheManager.Clear()
}

// ClearModelCache 清空模型缓存
func (c *CachedORM) ClearModelCache(model interface{}) error {
	if !c.cacheConfig.Enabled {
		return nil
	}

	// 这里需要实现一个前缀匹配的删除
	// 由于内存缓存不支持前缀删除，这里简单实现为清空所有缓存
	// 在实际使用中，可以使用Redis等支持前缀匹配的缓存提供者
	return c.cacheManager.Clear()
}

// WithContext 使用上下文
func (c *CachedORM) WithContext(ctx context.Context) *gorm.DB {
	return c.ORM.WithContext(ctx)
}

// 全局缓存ORM实例
var (
	globalCachedORM *CachedORM
	cachedORMOnce   sync.Once
)

// GetGlobalCachedORM 获取全局缓存ORM实例
func GetGlobalCachedORM() *CachedORM {
	cachedORMOnce.Do(func() {
		globalCachedORM = NewCachedORM(GetDefaultORM(), GetGlobalCacheManager(), DefaultCachedORMConfig())
	})
	return globalCachedORM
}

// SetGlobalCachedORM 设置全局缓存ORM实例
func SetGlobalCachedORM(orm *CachedORM) {
	if globalCachedORM != nil {
		// 不需要关闭底层ORM，因为它可能被其他地方使用
	}
	globalCachedORM = orm
}

// ============= 便捷函数 =============

// FirstWithCache 使用缓存查询第一条记录
func FirstWithCache(dest interface{}, conds ...interface{}) error {
	return GetGlobalCachedORM().First(dest, conds...)
}

// FindWithCache 使用缓存查询多条记录
func FindWithCache(dest interface{}, conds ...interface{}) error {
	return GetGlobalCachedORM().Find(dest, conds...)
}

// CreateWithCache 使用缓存创建记录
func CreateWithCache(value interface{}) error {
	return GetGlobalCachedORM().Create(value)
}

// SaveWithCache 使用缓存保存记录
func SaveWithCache(value interface{}) error {
	return GetGlobalCachedORM().Save(value)
}

// DeleteWithCache 使用缓存删除记录
func DeleteWithCache(value interface{}, conds ...interface{}) error {
	return GetGlobalCachedORM().Delete(value, conds...)
}

// ClearAllCache 清空所有缓存
func ClearAllCache() error {
	return GetGlobalCachedORM().ClearCache()
}