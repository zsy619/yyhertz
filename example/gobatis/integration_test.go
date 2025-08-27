// Package main 集成测试
//
// 测试MyBatis框架与其他组件的集成功能
package main

import (
	"context"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestMyBatisIntegration 测试MyBatis集成功能
func TestMyBatisIntegration(t *testing.T) {
	// 初始化测试数据库
	db, err := InitializeTestDatabase()
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	ctx := context.Background()

	t.Run("SimpleSession与XMLSession集成测试", func(t *testing.T) {
		// 创建SimpleSession
		simpleSession := mybatis.NewSimpleSession(db)

		// 准备数据
		userID, err := simpleSession.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			"集成测试用户", "integration@example.com", 30, "active")
		if err != nil {
			t.Fatalf("SimpleSession插入数据失败: %v", err)
		}

		// 使用XMLSession查询相同数据
		xmlSession := mybatis.NewXMLMapper(db)
		
		// 加载简单的XML映射
		mapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="IntegrationTest">
    <select id="selectById" parameterType="map" resultType="map">
        SELECT * FROM users WHERE id = #{id}
    </select>
</mapper>`

		err = xmlSession.LoadMapperXMLFromString(mapperXML)
		if err != nil {
			t.Fatalf("加载XML映射失败: %v", err)
		}

		// 查询数据
		user, err := xmlSession.SelectOneByID(ctx, "IntegrationTest.selectById", 
			map[string]any{"id": userID})
		if err != nil {
			t.Fatalf("XMLSession查询数据失败: %v", err)
		}

		userMap := user.(map[string]any)
		if userMap["name"] != "集成测试用户" {
			t.Errorf("期望用户名为'集成测试用户'，但得到'%v'", userMap["name"])
		}

		t.Log("SimpleSession与XMLSession集成测试通过")
	})

	t.Run("缓存与会话集成测试", func(t *testing.T) {
		// 创建启用缓存的会话
		cacheConfig := &mybatis.CacheConfig{
			L1CacheEnabled:  true,
			L1CacheSize:     1000,
			L1CacheTTL:      5 * time.Minute,
			L2CacheEnabled:  false,
			CleanupInterval: 1 * time.Minute,
		}
		
		session := mybatis.NewSimpleSession(db).
			EnableCache(cacheConfig).
			Debug(false)

		// 插入测试数据
		userID, err := session.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			"缓存集成用户", "cache_integration@example.com", 28, "active")
		if err != nil {
			t.Fatalf("插入数据失败: %v", err)
		}

		// 第一次查询
		user1, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
		if err != nil {
			t.Fatalf("第一次查询失败: %v", err)
		}

		// 第二次查询（应该从缓存获取）
		user2, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
		if err != nil {
			t.Fatalf("第二次查询失败: %v", err)
		}

		// 验证数据一致性
		user1Map := user1.(map[string]any)
		user2Map := user2.(map[string]any)
		
		if user1Map["name"] != user2Map["name"] {
			t.Error("缓存数据不一致")
		}

		t.Log("缓存与会话集成测试通过")
	})

	t.Run("多数据源集成测试", func(t *testing.T) {
		// 创建两个不同的会话（模拟不同数据源）
		session1 := mybatis.NewSimpleSession(db).Debug(false)
		session2 := mybatis.NewSimpleSession(db).Debug(false)

		// 在第一个会话中插入数据
		userID, err := session1.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			"多数据源用户1", "multi1@example.com", 25, "active")
		if err != nil {
			t.Fatalf("会话1插入数据失败: %v", err)
		}

		// 在第二个会话中查询数据
		user, err := session2.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
		if err != nil {
			t.Fatalf("会话2查询数据失败: %v", err)
		}

		userMap := user.(map[string]any)
		if userMap["name"] != "多数据源用户1" {
			t.Errorf("期望用户名为'多数据源用户1'，但得到'%v'", userMap["name"])
		}

		t.Log("多数据源集成测试通过")
	})
}

// TestTransactionIntegration 测试事务集成功能
func TestTransactionIntegration(t *testing.T) {
	// 初始化测试数据库
	db, err := InitializeTestDatabase()
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	}()

	ctx := context.Background()

	t.Run("事务回滚测试", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 创建事务会话
		txSession := mybatis.NewSimpleSession(tx)

		// 插入数据
		userID, err := txSession.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			"事务测试用户", "transaction@example.com", 27, "active")
		if err != nil {
			t.Fatalf("事务中插入数据失败: %v", err)
		}

		// 验证数据在事务中可见
		user, err := txSession.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
		if err != nil {
			t.Fatalf("事务中查询数据失败: %v", err)
		}
		
		userMap := user.(map[string]any)
		if userMap["name"] != "事务测试用户" {
			t.Errorf("事务中数据不正确")
		}

		// 回滚事务
		tx.Rollback()

		// 创建新会话验证数据已回滚
		session := mybatis.NewSimpleSession(db)
		user, err = session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
		if err == nil {
			t.Error("事务回滚后数据仍然存在，回滚失败")
		}

		t.Log("事务回滚测试通过")
	})

	t.Run("事务提交测试", func(t *testing.T) {
		// 开始事务
		tx := db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// 创建事务会话
		txSession := mybatis.NewSimpleSession(tx)

		// 插入数据
		userID, err := txSession.Insert(ctx,
			"INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)",
			"事务提交用户", "commit@example.com", 29, "active")
		if err != nil {
			t.Fatalf("事务中插入数据失败: %v", err)
		}

		// 提交事务
		tx.Commit()

		// 创建新会话验证数据已提交
		session := mybatis.NewSimpleSession(db)
		user, err := session.SelectOne(ctx, "SELECT * FROM users WHERE id = ?", userID)
		if err != nil {
			t.Fatalf("事务提交后查询数据失败: %v", err)
		}

		userMap := user.(map[string]any)
		if userMap["name"] != "事务提交用户" {
			t.Errorf("事务提交数据不正确")
		}

		t.Log("事务提交测试通过")
	})
}