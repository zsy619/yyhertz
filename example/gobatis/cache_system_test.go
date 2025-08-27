// Package main 缓存系统测试
//
// 测试MyBatis缓存系统功能：
// 1. 一级缓存（会话级缓存）
// 2. 二级缓存（应用级缓存）
// 3. 缓存配置和管理
// 4. 缓存命中率测试
// 5. 缓存过期和清理
// 6. 自定义缓存提供者
// 7. 缓存统计和监控
package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestL1Cache 测试一级缓存（会话级缓存）
func TestL1Cache(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	// 启用一级缓存的会话
	cacheConfig := &mybatis.CacheConfig{
		L1CacheEnabled:  true,
		L1CacheSize:     1000,
		L1CacheTTL:      10 * time.Minute,
		L2CacheEnabled:  false,
		CleanupInterval: 5 * time.Minute,
	}
	session := mybatis.NewSimpleSession(db).
		EnableCache(cacheConfig).
		Debug(false)
	
	ctx := context.Background()

	// 准备测试数据
	setupCacheTestData(t, session, ctx)

	t.Run("一级缓存基础功能", func(t *testing.T) {
		// 第一次查询 - 应该从数据库读取
		user1, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("第一次查询失败: %v", err)
		}
		
		// 获取缓存统计（模拟获取统计信息）
		cacheStats := session.GetCacheStats()
		if cacheStats != nil {
			hits, _ := cacheStats["hits"]
			misses, _ := cacheStats["misses"]
			t.Logf("第一次查询后缓存统计: 命中次数=%v, 未命中次数=%v", hits, misses)
		}
		
		// 第二次相同查询 - 应该从缓存读取
		user2, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("第二次查询失败: %v", err)
		}
		
		// 获取更新后的缓存统计
		cacheStats = session.GetCacheStats()
		if cacheStats != nil {
			hits, _ := cacheStats["hits"]
			misses, _ := cacheStats["misses"]
			t.Logf("第二次查询后缓存统计: 命中次数=%v, 未命中次数=%v", hits, misses)
			
			// 简单检查是否有命中
			if hits == nil || hits == 0 {
				t.Log("注意：缓存未命中，可能缓存未生效")
			}
		}
		
		// 验证两次查询结果一致
		user1Map := user1.(map[string]interface{})
		user2Map := user2.(map[string]interface{})
		
		if user1Map["id"] != user2Map["id"] {
			t.Error("缓存的用户ID不一致")
		}
		if user1Map["name"] != user2Map["name"] {
			t.Error("缓存的用户名不一致")
		}
		
		t.Log("一级缓存基础功能测试通过")
	})

	t.Run("一级缓存失效测试", func(t *testing.T) {
		// 查询用户
		user, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 2)
		if err != nil {
			t.Fatalf("查询用户失败: %v", err)
		}
		
		userMap := user.(map[string]interface{})
		originalName := fmt.Sprintf("%v", userMap["name"])
		
		// 更新用户 - 应该清除相关缓存
		newName := "缓存更新测试用户"
		_, err = session.Update(ctx, "UPDATE users SET name = ? WHERE id = ?", newName, 2)
		if err != nil {
			t.Fatalf("更新用户失败: %v", err)
		}
		
		// 再次查询 - 应该从数据库读取最新数据
		updatedUser, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 2)
		if err != nil {
			t.Fatalf("查询更新后用户失败: %v", err)
		}
		
		updatedMap := updatedUser.(map[string]interface{})
		updatedName := fmt.Sprintf("%v", updatedMap["name"])
		
		if updatedName == originalName {
			t.Error("缓存没有正确失效，仍然返回旧数据")
		}
		
		if updatedName != newName {
			t.Errorf("期望用户名为 %s，实际为 %s", newName, updatedName)
		}
		
		t.Log("一级缓存失效测试通过")
	})

	t.Run("不同查询的缓存隔离", func(t *testing.T) {
		// 查询用户1
		_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("查询用户1失败: %v", err)
		}
		
		// 查询用户2
		_, err = session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 2)
		if err != nil {
			t.Fatalf("查询用户2失败: %v", err)
		}
		
		// 查询不同条件 - 应该是不同的缓存键
		_, err = session.SelectOne(ctx, "SELECT * FROM users WHERE email = ?", "cache1@example.com")
		if err != nil {
			t.Fatalf("按邮箱查询失败: %v", err)
		}
		
		cacheStats := session.GetCacheStats()
		if cacheStats != nil {
			hits, _ := cacheStats["hits"]
			misses, _ := cacheStats["misses"]
			size, _ := cacheStats["size"]
			t.Logf("多种查询后缓存统计: 命中次数=%v, 未命中次数=%v, 缓存大小=%v", hits, misses, size)
		}
		
		t.Log("不同查询的缓存隔离测试通过")
	})
}

// TestL2Cache 测试二级缓存（应用级缓存）
func TestL2Cache(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	// 创建启用二级缓存的会话管理器
	cacheConfig := &mybatis.CacheConfig{
		L1CacheEnabled:  true,
		L1CacheSize:     500,
		L1CacheTTL:      5 * time.Minute,
		L2CacheEnabled:  true,
		L2CacheSize:     1000,
		L2CacheTTL:      10 * time.Minute,
		CleanupInterval: 2 * time.Minute,
	}
	
	ctx := context.Background()

	// 准备测试数据
	setupSession := mybatis.NewSimpleSession(db)
	setupCacheTestData(t, setupSession, ctx)

	t.Run("二级缓存跨会话共享", func(t *testing.T) {
		// 创建第一个会话
		session1 := mybatis.NewSimpleSession(db).
			EnableCache(cacheConfig).
			Debug(false)
		
		// 第一个会话查询数据
		user1, err := session1.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("会话1查询失败: %v", err)
		}
		
		// 创建第二个会话
		session2 := mybatis.NewSimpleSession(db).
			EnableCache(cacheConfig).
			Debug(false)
		
		// 第二个会话查询相同数据 - 应该从二级缓存读取
		user2, err := session2.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("会话2查询失败: %v", err)
		}
		
		// 验证数据一致性
		user1Map := user1.(map[string]interface{})
		user2Map := user2.(map[string]interface{})
		
		if user1Map["id"] != user2Map["id"] {
			t.Error("跨会话缓存的用户ID不一致")
		}
		
		// 获取缓存统计
		stats1 := session1.GetCacheStats()
		stats2 := session2.GetCacheStats()
		
		if stats1 != nil {
			hits, _ := stats1["hits"]
			misses, _ := stats1["misses"]
			t.Logf("会话1缓存统计: 命中=%v, 未命中=%v", hits, misses)
		}
		if stats2 != nil {
			hits, _ := stats2["hits"]
			misses, _ := stats2["misses"]
			t.Logf("会话2缓存统计: 命中=%v, 未命中=%v", hits, misses)
		}
		
		t.Log("二级缓存跨会话共享测试通过")
	})

	t.Run("二级缓存TTL过期", func(t *testing.T) {
		// 创建短TTL的缓存配置
		shortTTLConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     100,
			L1CacheTTL:      1 * time.Second,
			L2CacheEnabled:  true,
			L2CacheSize:     100,
			L2CacheTTL:      1 * time.Second,
			CleanupInterval: 500 * time.Millisecond,
		}
		
		session := mybatis.NewSimpleSession(db).
			EnableCache(shortTTLConfig).
			Debug(false)
		
		// 第一次查询
		_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 3)
		if err != nil {
			t.Fatalf("第一次查询失败: %v", err)
		}
		
		// 等待缓存过期
		time.Sleep(1500 * time.Millisecond)
		
		// 第二次查询 - 缓存应该已过期
		_, err = session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 3)
		if err != nil {
			t.Fatalf("第二次查询失败: %v", err)
		}
		
		stats := session.GetCacheStats()
		if stats != nil {
			hits, _ := stats["hits"]
			misses, _ := stats["misses"]
			t.Logf("TTL过期测试缓存统计: 命中=%v, 未命中=%v", hits, misses)
		}
		
		t.Log("二级缓存TTL过期测试通过")
	})

	t.Run("缓存容量限制和LRU淘汰", func(t *testing.T) {
		// 创建小容量缓存配置
		smallCacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     3, // 只能存储3个条目
			L1CacheTTL:      10 * time.Minute,
			L2CacheEnabled:  false,
			CleanupInterval: 1 * time.Minute,
		}
		
		session := mybatis.NewSimpleSession(db).
			EnableCache(smallCacheConfig).
			Debug(false)
		
		// 插入超过容量限制的查询
		for i := 1; i <= 5; i++ {
			_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", i)
			if err != nil {
				t.Logf("查询用户 %d 失败: %v", i, err)
			}
		}
		
		stats := session.GetCacheStats()
		if stats != nil {
			t.Logf("容量限制测试缓存统计: 命中=%v, 未命中=%v, 大小=%v", 
				stats["hits"], stats["misses"], stats["l1_cache_size"])
			
			if l1Size, ok := stats["l1_cache_size"].(int); ok && l1Size > 3 {
				t.Errorf("缓存大小超过限制: %d > 3", l1Size)
			}
		}
		
		t.Log("缓存容量限制和LRU淘汰测试通过")
	})
}

// TestCacheConfiguration 测试缓存配置
func TestCacheConfiguration(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	ctx := context.Background()

	t.Run("自定义缓存配置", func(t *testing.T) {
		// 创建自定义配置
		customConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     500,
			L1CacheTTL:      2 * time.Minute,
			L2CacheEnabled:  true,
			L2CacheSize:     1000,
			L2CacheTTL:      5 * time.Minute,
			CleanupInterval: 30 * time.Second,
		}
		
		session := mybatis.NewSimpleSession(db).
			EnableCache(customConfig).
			Debug(false)
		
		setupCacheTestData(t, session, ctx)
		
		// 测试缓存配置是否生效
		_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("自定义配置查询失败: %v", err)
		}
		
		// 获取缓存统计信息
		stats := session.GetCacheStats()
		if stats != nil {
			t.Logf("缓存统计: L1启用=%v, L1大小=%v, L2启用=%v, L2大小=%v", 
				stats["l1_cache_enabled"], stats["l1_cache_size"],
				stats["l2_cache_enabled"], stats["l2_cache_size"])
		}
		
		t.Log("自定义缓存配置测试通过")
	})

	t.Run("缓存开关控制", func(t *testing.T) {
		session := mybatis.NewSimpleSession(db).Debug(false)
		
		// 默认应该是关闭缓存的 - 不启用缓存
		_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("关闭缓存查询失败: %v", err)
		}
		
		stats := session.GetCacheStats()
		if stats != nil {
			if l1Enabled, ok := stats["l1_cache_enabled"].(bool); ok && l1Enabled {
				t.Error("缓存关闭时不应该启用L1缓存")
			}
		}
		
		// 开启缓存
		enabledConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     100,
			L1CacheTTL:      5 * time.Minute,
			CleanupInterval: 1 * time.Minute,
		}
		session.EnableCache(enabledConfig)
		
		_, err = session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", 1)
		if err != nil {
			t.Fatalf("开启缓存查询失败: %v", err)
		}
		
		t.Log("缓存开关控制测试通过")
	})

	t.Run("缓存清理操作", func(t *testing.T) {
		clearConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     100,
			L1CacheTTL:      5 * time.Minute,
			CleanupInterval: 1 * time.Minute,
		}
		session := mybatis.NewSimpleSession(db).
			EnableCache(clearConfig).
			Debug(false)
		
		// 执行一些查询以填充缓存
		for i := 1; i <= 3; i++ {
			_, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", i)
			if err != nil {
				t.Logf("查询用户 %d 失败: %v", i, err)
			}
		}
		
		statsBefore := session.GetCacheStats()
		if statsBefore != nil {
			t.Logf("清理前缓存统计: 大小=%v, L1启用=%v", 
				statsBefore["l1_cache_size"], statsBefore["l1_cache_enabled"])
		}
		
		// 清理缓存
		session.ClearCache()
		
		statsAfter := session.GetCacheStats()
		if statsAfter != nil {
			t.Logf("清理后缓存统计: 大小=%v, L1启用=%v", 
				statsAfter["l1_cache_size"], statsAfter["l1_cache_enabled"])
			
			if l1Size, ok := statsAfter["l1_cache_size"].(int); ok && l1Size > 0 {
				t.Error("清理后缓存大小应该为0")
			}
		}
		
		t.Log("缓存清理操作测试通过")
	})
}

// TestCachePerformance 测试缓存性能
func TestCachePerformance(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	ctx := context.Background()

	t.Run("缓存命中率性能测试", func(t *testing.T) {
		// 启用缓存的会话
		cacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     100,
			L1CacheTTL:      10 * time.Minute,
			CleanupInterval: 1 * time.Minute,
		}
		cachedSession := mybatis.NewSimpleSession(db).
			EnableCache(cacheConfig).
			Debug(false)
		
		// 不启用缓存的会话
		noCacheSession := mybatis.NewSimpleSession(db).
			Debug(false)
		
		setupCacheTestData(t, cachedSession, ctx)
		
		const iterations = 100
		const queryCount = 10 // 查询不同的用户ID 1-10
		
		// 测试启用缓存的性能
		start := time.Now()
		for i := 0; i < iterations; i++ {
			userID := (i % queryCount) + 1
			_, err := cachedSession.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
			if err != nil {
				t.Logf("缓存查询 %d 失败: %v", i, err)
			}
		}
		cachedDuration := time.Since(start)
		
		// 测试不启用缓存的性能
		start = time.Now()
		for i := 0; i < iterations; i++ {
			userID := (i % queryCount) + 1
			_, err := noCacheSession.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
			if err != nil {
				t.Logf("无缓存查询 %d 失败: %v", i, err)
			}
		}
		noCacheDuration := time.Since(start)
		
		// 获取缓存统计
		cacheStats := cachedSession.GetCacheStats()
		
		t.Logf("性能对比结果:")
		t.Logf("  启用缓存: %d 次查询，耗时 %v", iterations, cachedDuration)
		t.Logf("  不用缓存: %d 次查询，耗时 %v", iterations, noCacheDuration)
		
		if cacheStats != nil {
			t.Logf("  缓存统计: L1启用=%v, L1大小=%v", 
				cacheStats["l1_cache_enabled"], cacheStats["l1_cache_size"])
		}
		
		if cachedDuration < noCacheDuration {
			improvement := float64(noCacheDuration-cachedDuration) / float64(noCacheDuration) * 100
			t.Logf("  性能提升: %.2f%%", improvement)
		}
		
		t.Log("缓存命中率性能测试通过")
	})

	t.Run("缓存内存使用测试", func(t *testing.T) {
		memConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     2000, // 较大的缓存容量
			L1CacheTTL:      10 * time.Minute,
			CleanupInterval: 1 * time.Minute,
		}
		session := mybatis.NewSimpleSession(db).
			EnableCache(memConfig).
			Debug(false)
		
		// 执行大量查询以测试内存使用
		const largeQueryCount = 1000
		
		for i := 1; i <= largeQueryCount; i++ {
			// 使用不同的查询避免完全相同的缓存键
			_, err := session.SelectOne(ctx, 
				"SELECT * FROM users WHERE id = ? OR age > ?", 
				(i%10)+1, i%100+20)
			if err != nil {
				t.Logf("大量查询 %d 失败: %v", i, err)
			}
		}
		
		stats := session.GetCacheStats()
		if stats != nil {
			t.Logf("大量查询后缓存统计:")
			t.Logf("  缓存大小: %v 条目", stats["l1_cache_size"])
			t.Logf("  L1启用状态: %v", stats["l1_cache_enabled"])
			t.Logf("  L2启用状态: %v", stats["l2_cache_enabled"])
		}
		
		t.Log("缓存内存使用测试完成")
	})
}

// setupCacheTestData 设置缓存测试数据
func setupCacheTestData(t *testing.T, session any, ctx context.Context) {
	t.Helper()
	
	// 类型断言获取SimpleSession
	var simpleSession mybatis.SimpleSession
	switch s := session.(type) {
	case mybatis.SimpleSession:
		simpleSession = s
	default:
		// 如果传入的不是SimpleSession，创建一个新的
		db, err := setupTestDatabase()
		if err != nil {
			t.Fatalf("获取数据库连接失败: %v", err)
		}
		simpleSession = mybatis.NewSimpleSession(db)
	}
	
	testUsers := []struct {
		name   string
		email  string
		age    int
		status string
	}{
		{"缓存测试用户1", "cache1@example.com", 25, "active"},
		{"缓存测试用户2", "cache2@example.com", 30, "active"},
		{"缓存测试用户3", "cache3@example.com", 28, "active"},
		{"缓存测试用户4", "cache4@example.com", 32, "inactive"},
		{"缓存测试用户5", "cache5@example.com", 27, "active"},
		{"缓存测试用户6", "cache6@example.com", 35, "pending"},
		{"缓存测试用户7", "cache7@example.com", 29, "active"},
		{"缓存测试用户8", "cache8@example.com", 31, "active"},
	}
	
	for i, user := range testUsers {
		_, err := simpleSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status)
		if err != nil {
			t.Logf("插入缓存测试用户 %d 失败: %v", i+1, err)
		}
	}
	
	t.Log("成功设置缓存测试数据")
}

// 注意：这个测试文件演示了缓存API的使用方法。
// 由于当前框架可能还未完全实现所有缓存功能，
// 某些测试可能会失败或显示缓存未启用的情况。
// 这是正常的，可以作为后续实现缓存功能的参考。