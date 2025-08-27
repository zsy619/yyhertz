// Package main XML基础测试
//
// 测试MyBatis框架的XML映射功能
package main

import (
	"context"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestXMLBasicFunctionality 测试XML基础功能
func TestXMLBasicFunctionality(t *testing.T) {
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

	// 创建XMLSession
	xmlSession := mybatis.NewXMLMapper(db)
	ctx := context.Background()

	// 准备XML映射
	mapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="UserMapper">
    <insert id="insertUser" parameterType="map">
        INSERT INTO users (name, email, age, status) VALUES (#{name}, #{email}, #{age}, #{status})
    </insert>
    
    <select id="selectById" parameterType="map" resultType="map">
        SELECT * FROM users WHERE id = #{id}
    </select>
    
    <select id="selectAll" resultType="map">
        SELECT * FROM users ORDER BY id
    </select>
    
    <update id="updateUser" parameterType="map">
        UPDATE users SET name = #{name}, age = #{age} WHERE id = #{id}
    </update>
    
    <delete id="deleteUser" parameterType="map">
        DELETE FROM users WHERE id = #{id}
    </delete>
</mapper>`

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(mapperXML)
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	t.Run("XML插入操作测试", func(t *testing.T) {
		// 插入用户
		params := map[string]any{
			"name":   "XML测试用户",
			"email":  "xml@test.com", 
			"age":    25,
			"status": "active",
		}

		userID, err := xmlSession.InsertByID(ctx, "UserMapper.insertUser", params)
		if err != nil {
			t.Fatalf("XML插入用户失败: %v", err)
		}

		if userID == 0 {
			t.Error("插入用户ID不应该为0")
		}

		t.Logf("XML插入用户成功，ID: %d", userID)
	})

	t.Run("XML查询操作测试", func(t *testing.T) {
		// 先插入一个测试用户
		insertParams := map[string]any{
			"name":   "查询测试用户",
			"email":  "query@test.com",
			"age":    30,
			"status": "active",
		}

		userID, err := xmlSession.InsertByID(ctx, "UserMapper.insertUser", insertParams)
		if err != nil {
			t.Fatalf("插入测试用户失败: %v", err)
		}

		// 根据ID查询
		queryParams := map[string]any{"id": userID}
		user, err := xmlSession.SelectOneByID(ctx, "UserMapper.selectById", queryParams)
		if err != nil {
			t.Fatalf("XML查询用户失败: %v", err)
		}

		userMap := user.(map[string]any)
		if userMap["name"] != "查询测试用户" {
			t.Errorf("期望用户名为'查询测试用户'，但得到'%v'", userMap["name"])
		}

		t.Log("XML查询操作测试通过")
	})

	t.Run("XML更新操作测试", func(t *testing.T) {
		// 先插入一个测试用户
		insertParams := map[string]any{
			"name":   "更新测试用户",
			"email":  "update@test.com",
			"age":    25,
			"status": "active",
		}

		userID, err := xmlSession.InsertByID(ctx, "UserMapper.insertUser", insertParams)
		if err != nil {
			t.Fatalf("插入测试用户失败: %v", err)
		}

		// 更新用户
		updateParams := map[string]any{
			"id":   userID,
			"name": "已更新用户",
			"age":  26,
		}

		affected, err := xmlSession.UpdateByID(ctx, "UserMapper.updateUser", updateParams)
		if err != nil {
			t.Fatalf("XML更新用户失败: %v", err)
		}

		if affected != 1 {
			t.Errorf("期望更新1行，但实际更新了%d行", affected)
		}

		// 验证更新结果
		queryParams := map[string]any{"id": userID}
		user, err := xmlSession.SelectOneByID(ctx, "UserMapper.selectById", queryParams)
		if err != nil {
			t.Fatalf("查询更新后的用户失败: %v", err)
		}

		userMap := user.(map[string]any)
		if userMap["name"] != "已更新用户" {
			t.Errorf("期望用户名为'已更新用户'，但得到'%v'", userMap["name"])
		}

		t.Log("XML更新操作测试通过")
	})

	t.Run("XML删除操作测试", func(t *testing.T) {
		// 先插入一个测试用户
		insertParams := map[string]any{
			"name":   "删除测试用户",
			"email":  "delete@test.com",
			"age":    28,
			"status": "active",
		}

		userID, err := xmlSession.InsertByID(ctx, "UserMapper.insertUser", insertParams)
		if err != nil {
			t.Fatalf("插入测试用户失败: %v", err)
		}

		// 删除用户
		deleteParams := map[string]any{"id": userID}
		affected, err := xmlSession.DeleteByID(ctx, "UserMapper.deleteUser", deleteParams)
		if err != nil {
			t.Fatalf("XML删除用户失败: %v", err)
		}

		if affected != 1 {
			t.Errorf("期望删除1行，但实际删除了%d行", affected)
		}

		// 验证删除结果
		user, err := xmlSession.SelectOneByID(ctx, "UserMapper.selectById", deleteParams)
		if err == nil && user != nil {
			t.Error("用户删除后仍然存在")
		}

		t.Log("XML删除操作测试通过")
	})

	t.Run("XML查询全部测试", func(t *testing.T) {
		// 插入多个测试用户
		testUsers := []map[string]any{
			{"name": "用户1", "email": "user1@test.com", "age": 20, "status": "active"},
			{"name": "用户2", "email": "user2@test.com", "age": 25, "status": "inactive"},
			{"name": "用户3", "email": "user3@test.com", "age": 30, "status": "active"},
		}

		for i, userParams := range testUsers {
			_, err := xmlSession.InsertByID(ctx, "UserMapper.insertUser", userParams)
			if err != nil {
				t.Fatalf("插入测试用户%d失败: %v", i+1, err)
			}
		}

		// 查询所有用户
		users, err := xmlSession.SelectListByID(ctx, "UserMapper.selectAll", nil)
		if err != nil {
			t.Fatalf("XML查询所有用户失败: %v", err)
		}

		if len(users) < len(testUsers) {
			t.Errorf("期望至少查询到%d个用户，但只查询到%d个", len(testUsers), len(users))
		}

		t.Logf("XML查询全部测试通过，查询到%d个用户", len(users))
	})
}

// TestXMLDynamicSQL 测试XML动态SQL功能
func TestXMLDynamicSQL(t *testing.T) {
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

	// 创建XMLSession
	xmlSession := mybatis.NewXMLMapper(db)
	ctx := context.Background()

	// 准备动态SQL映射
	dynamicMapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="DynamicMapper">
    <insert id="insertUser" parameterType="map">
        INSERT INTO users (name, email, age, status) VALUES (#{name}, #{email}, #{age}, #{status})
    </insert>
    
    <select id="selectByCondition" parameterType="map" resultType="map">
        SELECT * FROM users
        <where>
            <if test="name != null and name != ''">
                AND name LIKE CONCAT('%', #{name}, '%')
            </if>
            <if test="status != null">
                AND status = #{status}
            </if>
            <if test="ageMin > 0">
                AND age >= #{ageMin}
            </if>
            <if test="ageMax > 0">
                AND age <= #{ageMax}
            </if>
        </where>
        ORDER BY id
    </select>
</mapper>`

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(dynamicMapperXML)
	if err != nil {
		t.Fatalf("加载动态XML映射失败: %v", err)
	}

	// 准备测试数据
	testUsers := []map[string]any{
		{"name": "张三", "email": "zhangsan@test.com", "age": 25, "status": "active"},
		{"name": "李四", "email": "lisi@test.com", "age": 30, "status": "inactive"},
		{"name": "王五", "email": "wangwu@test.com", "age": 35, "status": "active"},
		{"name": "赵六", "email": "zhaoliu@test.com", "age": 28, "status": "pending"},
	}

	for _, userParams := range testUsers {
		_, err := xmlSession.InsertByID(ctx, "DynamicMapper.insertUser", userParams)
		if err != nil {
			t.Fatalf("插入测试数据失败: %v", err)
		}
	}

	t.Run("按姓名动态查询", func(t *testing.T) {
		params := map[string]any{
			"name": "张",
		}

		users, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectByCondition", params)
		if err != nil {
			t.Fatalf("动态SQL查询失败: %v", err)
		}

		if len(users) < 1 {
			t.Error("期望查询到至少1个包含'张'字的用户")
		}

		t.Logf("按姓名动态查询通过，查询到%d个用户", len(users))
	})

	t.Run("按状态动态查询", func(t *testing.T) {
		params := map[string]any{
			"status": "active",
		}

		users, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectByCondition", params)
		if err != nil {
			t.Fatalf("动态SQL查询失败: %v", err)
		}

		if len(users) < 1 {
			t.Error("期望查询到至少1个active状态的用户")
		}

		// 验证所有查询结果都是active状态
		for _, user := range users {
			userMap := user.(map[string]any)
			if userMap["status"] != "active" {
				t.Errorf("查询结果中包含非active用户: %v", userMap["status"])
			}
		}

		t.Logf("按状态动态查询通过，查询到%d个active用户", len(users))
	})

	t.Run("按年龄范围动态查询", func(t *testing.T) {
		params := map[string]any{
			"ageMin": 25,
			"ageMax": 30,
		}

		users, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectByCondition", params)
		if err != nil {
			t.Fatalf("动态SQL查询失败: %v", err)
		}

		if len(users) < 1 {
			t.Error("期望查询到至少1个年龄在25-30之间的用户")
		}

		// 验证所有查询结果都在年龄范围内
		for _, user := range users {
			userMap := user.(map[string]any)
			age := userMap["age"]
			if age.(int64) < 25 || age.(int64) > 30 {
				t.Errorf("查询结果中包含年龄超出范围的用户: %v", age)
			}
		}

		t.Logf("按年龄范围动态查询通过，查询到%d个用户", len(users))
	})

	t.Run("多条件组合动态查询", func(t *testing.T) {
		params := map[string]any{
			"name":   "张",
			"status": "active",
			"ageMin": 20,
		}

		users, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectByCondition", params)
		if err != nil {
			t.Fatalf("动态SQL查询失败: %v", err)
		}

		t.Logf("多条件组合动态查询完成，查询到%d个用户", len(users))
	})
}

// TestXMLPerformance 测试XML映射性能
func TestXMLPerformance(t *testing.T) {
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

	// 创建XMLSession
	xmlSession := mybatis.NewXMLMapper(db)
	ctx := context.Background()

	// 准备性能测试映射
	perfMapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="PerfMapper">
    <insert id="insertUser" parameterType="map">
        INSERT INTO users (name, email, age, status) VALUES (#{name}, #{email}, #{age}, #{status})
    </insert>
    
    <select id="selectById" parameterType="map" resultType="map">
        SELECT * FROM users WHERE id = #{id}
    </select>
    
    <select id="selectByStatus" parameterType="map" resultType="map">
        SELECT * FROM users WHERE status = #{status} LIMIT 100
    </select>
</mapper>`

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(perfMapperXML)
	if err != nil {
		t.Fatalf("加载性能测试XML映射失败: %v", err)
	}

	t.Run("XML批量插入性能测试", func(t *testing.T) {
		const batchSize = 100
		start := time.Now()

		for i := 0; i < batchSize; i++ {
			params := map[string]any{
				"name":   "性能用户" + string(rune(i)),
				"email":  "perf" + string(rune(i)) + "@test.com",
				"age":    20 + (i % 50),
				"status": []string{"active", "inactive", "pending"}[i%3],
			}

			_, err := xmlSession.InsertByID(ctx, "PerfMapper.insertUser", params)
			if err != nil {
				t.Logf("插入性能测试数据%d失败: %v", i, err)
			}
		}

		duration := time.Since(start)
		opsPerSecond := float64(batchSize) / duration.Seconds()

		t.Logf("XML批量插入性能测试:")
		t.Logf("  插入%d条记录耗时: %v", batchSize, duration)
		t.Logf("  平均每秒操作数: %.2f ops/s", opsPerSecond)

		if opsPerSecond < 10 {
			t.Logf("警告：XML插入性能较低 (%.2f ops/s)", opsPerSecond)
		}
	})

	t.Run("XML查询性能测试", func(t *testing.T) {
		const queryCount = 100
		start := time.Now()

		for i := 0; i < queryCount; i++ {
			params := map[string]any{
				"status": []string{"active", "inactive", "pending"}[i%3],
			}

			_, err := xmlSession.SelectListByID(ctx, "PerfMapper.selectByStatus", params)
			if err != nil {
				t.Logf("性能测试查询%d失败: %v", i, err)
			}
		}

		duration := time.Since(start)
		opsPerSecond := float64(queryCount) / duration.Seconds()

		t.Logf("XML查询性能测试:")
		t.Logf("  执行%d次查询耗时: %v", queryCount, duration)
		t.Logf("  平均每秒操作数: %.2f ops/s", opsPerSecond)

		if opsPerSecond < 100 {
			t.Logf("警告：XML查询性能较低 (%.2f ops/s)", opsPerSecond)
		}
	})
}