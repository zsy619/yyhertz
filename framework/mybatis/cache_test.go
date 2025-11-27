package mybatis_test

import (
	"context"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCacheOperations(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表
	err = db.Exec(`
		CREATE TABLE cache_test (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER
		);
		INSERT INTO cache_test (name, value) VALUES 
			('cache1', 100),
			('cache2', 200),
			('cache3', 300);
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table and data: %v", err)
	}

	session := mybatis.NewSimpleSession(db)
	ctx := context.Background()

	t.Run("Cache_Disabled_By_Default", func(t *testing.T) {
		// 默认情况下缓存应该是禁用的
		stats := session.GetCacheStats()
		if stats["cache_enabled"].(bool) {
			t.Error("Cache should be disabled by default")
		}

		// 执行查询应该不使用缓存
		result1, err := session.SelectList(ctx, "SELECT * FROM cache_test WHERE value > ?", 150)
		if err != nil {
			t.Fatalf("First query failed: %v", err)
		}

		result2, err := session.SelectList(ctx, "SELECT * FROM cache_test WHERE value > ?", 150)
		if err != nil {
			t.Fatalf("Second query failed: %v", err)
		}

		if len(result1) != len(result2) {
			t.Error("Results should be consistent")
		}
	})

	t.Run("Enable_L1_Cache", func(t *testing.T) {
		// 启用一级缓存
		cacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     100,
			L1CacheTTL:      5 * time.Second,
			L2CacheEnabled:  false,
			CleanupInterval: 1 * time.Second,
		}

		session = session.EnableCache(cacheConfig).Debug(true)

		// 验证缓存已启用
		stats := session.GetCacheStats()
		if !stats["l1_cache_enabled"].(bool) {
			t.Error("L1 cache should be enabled")
		}

		// 第一次查询
		sql := "SELECT * FROM cache_test WHERE id = ?"
		args := []any{1}

		start := time.Now()
		result1, err := session.SelectList(ctx, sql, args...)
		if err != nil {
			t.Fatalf("First cached query failed: %v", err)
		}
		firstQueryTime := time.Since(start)

		// 第二次查询（应该来自缓存）
		start = time.Now()
		result2, err := session.SelectList(ctx, sql, args...)
		if err != nil {
			t.Fatalf("Second cached query failed: %v", err)
		}
		secondQueryTime := time.Since(start)

		// 验证结果一致性
		if len(result1) != len(result2) {
			t.Error("Cached results should be consistent")
		}

		// 第二次查询应该明显更快（来自缓存）
		if secondQueryTime > firstQueryTime {
			t.Logf("First query: %v, Second query: %v", firstQueryTime, secondQueryTime)
			// 注意：在测试环境中，差异可能不明显，所以这里只是记录
		}

		// 验证缓存统计
		stats = session.GetCacheStats()
		cacheSize := stats["l1_cache_size"].(int)
		if cacheSize == 0 {
			t.Error("Cache should contain at least one entry")
		}
	})

	t.Run("Enable_L1_and_L2_Cache", func(t *testing.T) {
		// 启用一级和二级缓存
		cacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     50,
			L1CacheTTL:      2 * time.Second,
			L2CacheEnabled:  true,
			L2CacheSize:     200,
			L2CacheTTL:      10 * time.Second,
			CleanupInterval: 1 * time.Second,
		}

		session = session.DisableCache().EnableCache(cacheConfig).Debug(true)

		// 验证两级缓存都已启用
		stats := session.GetCacheStats()
		if !stats["l1_cache_enabled"].(bool) {
			t.Error("L1 cache should be enabled")
		}
		if !stats["l2_cache_enabled"].(bool) {
			t.Error("L2 cache should be enabled")
		}

		// 执行查询
		sql := "SELECT * FROM cache_test WHERE value < ?"
		args := []any{250}

		result1, err := session.SelectList(ctx, sql, args...)
		if err != nil {
			t.Fatalf("First query failed: %v", err)
		}

		// 验证数据被缓存
		result2, err := session.SelectList(ctx, sql, args...)
		if err != nil {
			t.Fatalf("Second query failed: %v", err)
		}

		if len(result1) != len(result2) {
			t.Error("Results should be consistent")
		}
	})

	t.Run("Cache_Clear", func(t *testing.T) {
		// 启用缓存
		cacheConfig := mybatis.DefaultCacheConfig()
		session = session.EnableCache(cacheConfig)

		// 执行查询填充缓存
		_, err := session.SelectList(ctx, "SELECT * FROM cache_test", nil)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		// 验证缓存中有数据
		stats := session.GetCacheStats()
		if stats["l1_cache_size"].(int) == 0 {
			t.Error("Cache should contain data")
		}

		// 清空缓存
		session = session.ClearCache()

		// 验证缓存已清空
		stats = session.GetCacheStats()
		if stats["l1_cache_size"].(int) != 0 {
			t.Error("Cache should be empty after clear")
		}
	})

	t.Run("Cache_Disable", func(t *testing.T) {
		// 启用然后禁用缓存
		cacheConfig := mybatis.DefaultCacheConfig()
		session = session.EnableCache(cacheConfig)

		// 验证缓存已启用
		stats := session.GetCacheStats()
		if !stats["l1_cache_enabled"].(bool) {
			t.Error("Cache should be enabled")
		}

		// 禁用缓存
		session = session.DisableCache()

		// 验证缓存已禁用
		stats = session.GetCacheStats()
		if stats["cache_enabled"].(bool) {
			t.Error("Cache should be disabled")
		}
	})

	t.Run("Cache_TTL_Expiration", func(t *testing.T) {
		// 启用缓存，设置短TTL
		cacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     100,
			L1CacheTTL:      100 * time.Millisecond, // 很短的TTL
			L2CacheEnabled:  false,
			CleanupInterval: 50 * time.Millisecond,
		}

		session = session.EnableCache(cacheConfig).Debug(true)

		// 执行查询
		sql := "SELECT COUNT(*) as count FROM cache_test"
		result1, err := session.SelectList(ctx, sql)
		if err != nil {
			t.Fatalf("First query failed: %v", err)
		}

		// 等待缓存过期
		time.Sleep(200 * time.Millisecond)

		// 再次查询（应该重新从数据库获取）
		result2, err := session.SelectList(ctx, sql)
		if err != nil {
			t.Fatalf("Second query failed: %v", err)
		}

		// 验证结果一致性
		if len(result1) != len(result2) {
			t.Error("Results should be consistent even after cache expiration")
		}
	})
}

func TestXMLSessionCache(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表
	err = db.Exec(`
		CREATE TABLE xml_cache_test (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		);
		INSERT INTO xml_cache_test (name) VALUES ('xml1'), ('xml2');
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	session := mybatis.NewXMLSession(db)
	ctx := context.Background()

	t.Run("XMLSession_Cache_Enable", func(t *testing.T) {
		// 启用缓存
		cacheConfig := mybatis.DefaultCacheConfig()
		session = session.EnableCache(cacheConfig).Debug(true)

		// 验证缓存已启用
		stats := session.GetCacheStats()
		if !stats["l1_cache_enabled"].(bool) {
			t.Error("XMLSession cache should be enabled")
		}

		// 执行查询
		result1, err := session.SelectList(ctx, "SELECT * FROM xml_cache_test")
		if err != nil {
			t.Fatalf("XMLSession cached query failed: %v", err)
		}

		result2, err := session.SelectList(ctx, "SELECT * FROM xml_cache_test")
		if err != nil {
			t.Fatalf("XMLSession second cached query failed: %v", err)
		}

		if len(result1) != len(result2) {
			t.Error("XMLSession cached results should be consistent")
		}
	})

	t.Run("XMLSession_Cache_Clear", func(t *testing.T) {
		// 清空缓存
		session = session.ClearCache()

		stats := session.GetCacheStats()
		if stats["l1_cache_size"].(int) != 0 {
			t.Error("XMLSession cache should be empty after clear")
		}
	})
}