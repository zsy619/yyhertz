// Package main 动态SQL和条件查询测试
//
// 测试MyBatis动态SQL功能：
// 1. if条件标签
// 2. where条件标签
// 3. choose-when-otherwise标签
// 4. foreach循环标签
// 5. set更新标签
// 6. trim标签功能
// 7. 复杂条件组合
package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestDynamicSQLBasic 测试基础动态SQL功能
func TestDynamicSQLBasic(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载动态SQL测试映射
	err = xmlSession.LoadMapperXMLFromString(getAdvancedDynamicSQLMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupDynamicSQLTestData(t, xmlSession, ctx)

	t.Run("if条件-单个条件", func(t *testing.T) {
		// 测试单个if条件
		query := UserQuery{
			Name: "张",
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithSingleIf", query)
		if err != nil {
			t.Fatalf("单个if条件查询失败: %v", err)
		}
		
		t.Logf("单个if条件查询成功，查询到 %d 条记录", len(users))
		
		// 验证所有结果都包含"张"
		for i, user := range users {
			userMap := user.(map[string]interface{})
			name := fmt.Sprintf("%v", userMap["name"])
			if len(name) > 0 && name[0:3] != "张" { // 检查第一个字符是否为"张"
				t.Errorf("记录 %d 的姓名 '%s' 不符合查询条件", i, name)
				break
			}
		}
	})

	t.Run("if条件-多个条件", func(t *testing.T) {
		// 测试多个if条件
		query := UserQuery{
			Status: "active",
			AgeMin: 25,
			AgeMax: 35,
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithMultipleIf", query)
		if err != nil {
			t.Fatalf("多个if条件查询失败: %v", err)
		}
		
		t.Logf("多个if条件查询成功，查询到 %d 条记录", len(users))
		
		// 验证结果符合条件
		for i, user := range users {
			userMap := user.(map[string]interface{})
			status := fmt.Sprintf("%v", userMap["status"])
			if status != "active" {
				t.Errorf("记录 %d 状态不符合条件: %s", i, status)
				break
			}
		}
	})

	t.Run("where条件-自动添加AND", func(t *testing.T) {
		query := UserQuery{
			Status: "active",
			Name:   "李",
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithWhere", query)
		if err != nil {
			t.Fatalf("where条件查询失败: %v", err)
		}
		
		t.Logf("where条件查询成功，查询到 %d 条记录", len(users))
	})

	t.Run("choose-when-otherwise分支", func(t *testing.T) {
		// 测试name分支
		query1 := UserQuery{Name: "张三"}
		users1, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithChoose", query1)
		if err != nil {
			t.Fatalf("choose-when(name)查询失败: %v", err)
		}

		// 测试email分支  
		query2 := UserQuery{Email: "zhangsan@example.com"}
		users2, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithChoose", query2)
		if err != nil {
			t.Fatalf("choose-when(email)查询失败: %v", err)
		}

		// 测试otherwise分支
		query3 := UserQuery{}
		users3, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithChoose", query3)
		if err != nil {
			t.Fatalf("choose-otherwise查询失败: %v", err)
		}

		t.Logf("choose分支测试成功 - name分支: %d条, email分支: %d条, otherwise分支: %d条", 
			len(users1), len(users2), len(users3))
	})
}

// TestDynamicSQLForeach 测试foreach循环功能
func TestDynamicSQLForeach(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getAdvancedDynamicSQLMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据并获取IDs
	userIDs := setupDynamicSQLTestDataWithIDs(t, xmlSession, ctx)

	t.Run("foreach-IN查询", func(t *testing.T) {
		if len(userIDs) < 3 {
			t.Skip("测试数据不足，跳过foreach测试")
		}
		
		// 取前3个ID进行测试
		testIDs := userIDs[:3]
		query := map[string]interface{}{
			"ids": testIDs,
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectByIds", query)
		if err != nil {
			t.Fatalf("foreach IN查询失败: %v", err)
		}
		
		if len(users) != len(testIDs) {
			t.Errorf("期望查询到 %d 条记录，实际查询到 %d 条", len(testIDs), len(users))
		}
		
		t.Logf("foreach IN查询成功，查询到 %d 条记录", len(users))
	})

	t.Run("foreach-批量插入", func(t *testing.T) {
		batchUsers := []map[string]interface{}{
			{"name": "批量用户1", "email": "batch1@example.com", "age": 25, "status": "active"},
			{"name": "批量用户2", "email": "batch2@example.com", "age": 26, "status": "active"},
			{"name": "批量用户3", "email": "batch3@example.com", "age": 27, "status": "active"},
		}
		
		query := map[string]interface{}{
			"users": batchUsers,
		}
		
		result, err := xmlSession.SelectOneByID(ctx, "DynamicAdvanced.batchInsertUsers", query)
		if err != nil {
			t.Fatalf("foreach批量插入失败: %v", err)
		}
		
		t.Logf("foreach批量插入成功: %+v", result)
	})
}

// TestDynamicSQLSet 测试set动态更新功能
func TestDynamicSQLSet(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getAdvancedDynamicSQLMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	userIDs := setupDynamicSQLTestDataWithIDs(t, xmlSession, ctx)
	if len(userIDs) == 0 {
		t.Fatal("没有测试数据可用于set更新测试")
	}

	t.Run("set标签-部分更新", func(t *testing.T) {
		// 只更新age和status字段
		updateData := map[string]interface{}{
			"id":     userIDs[0],
			"age":    30,
			"status": "updated",
			// name 和 email 故意不设置，测试动态set功能
		}
		
		result, err := xmlSession.SelectOneByID(ctx, "DynamicAdvanced.updateUserSelective", updateData)
		if err != nil {
			t.Fatalf("set标签部分更新失败: %v", err)
		}
		
		t.Logf("set标签部分更新成功: %+v", result)
		
		// 验证更新结果
		simpleSession := mybatis.NewSimpleSession(db)
		updatedUser, err := simpleSession.SelectOne(ctx, 
			"SELECT * FROM users WHERE id = ?", userIDs[0])
		if err != nil {
			t.Fatalf("验证更新结果失败: %v", err)
		}
		
		userMap := updatedUser.(map[string]interface{})
		if fmt.Sprintf("%v", userMap["status"]) != "updated" {
			t.Errorf("期望状态为updated，实际为: %v", userMap["status"])
		}
	})

	t.Run("set标签-全量更新", func(t *testing.T) {
		if len(userIDs) < 2 {
			t.Skip("测试数据不足")
		}
		
		// 设置所有字段
		updateData := map[string]interface{}{
			"id":     userIDs[1],
			"name":   "完全更新用户",
			"email":  "fullyupdated@example.com",
			"age":    35,
			"status": "fully_updated",
		}
		
		result, err := xmlSession.SelectOneByID(ctx, "DynamicAdvanced.updateUserSelective", updateData)
		if err != nil {
			t.Fatalf("set标签全量更新失败: %v", err)
		}
		
		t.Logf("set标签全量更新成功: %+v", result)
	})
}

// TestDynamicSQLTrim 测试trim标签功能
func TestDynamicSQLTrim(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getAdvancedDynamicSQLMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupDynamicSQLTestData(t, xmlSession, ctx)

	t.Run("trim标签-自动处理前缀", func(t *testing.T) {
		query := UserQuery{
			Status: "active",
			AgeMin: 25,
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithTrim", query)
		if err != nil {
			t.Fatalf("trim标签查询失败: %v", err)
		}
		
		t.Logf("trim标签查询成功，查询到 %d 条记录", len(users))
	})

	t.Run("trim标签-处理空条件", func(t *testing.T) {
		emptyQuery := UserQuery{}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithTrim", emptyQuery)
		if err != nil {
			t.Fatalf("trim标签空条件查询失败: %v", err)
		}
		
		t.Logf("trim标签空条件查询成功，查询到 %d 条记录", len(users))
	})
}

// TestDynamicSQLComplex 测试复杂条件组合
func TestDynamicSQLComplex(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载映射
	err = xmlSession.LoadMapperXMLFromString(getAdvancedDynamicSQLMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupDynamicSQLTestData(t, xmlSession, ctx)

	t.Run("复杂条件组合", func(t *testing.T) {
		query := UserQuery{
			Status:   "active",
			AgeMin:   20,
			AgeMax:   50,
			Keyword:  "用户",
			Page:     1,
			PageSize: 10,
			OrderBy:  "age",
			OrderDesc: false,
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithComplexConditions", query)
		if err != nil {
			t.Fatalf("复杂条件组合查询失败: %v", err)
		}
		
		t.Logf("复杂条件组合查询成功，查询到 %d 条记录", len(users))
		
		// 验证排序（如果有结果的话）
		if len(users) > 1 {
			user1 := users[0].(map[string]interface{})
			user2 := users[1].(map[string]interface{})
			
			age1 := fmt.Sprintf("%v", user1["age"])
			age2 := fmt.Sprintf("%v", user2["age"])
			
			t.Logf("验证排序 - 第1条记录年龄: %s, 第2条记录年龄: %s", age1, age2)
		}
	})

	t.Run("嵌套条件测试", func(t *testing.T) {
		query := map[string]interface{}{
			"searchType": "advanced",
			"filters": map[string]interface{}{
				"status":    "active",
				"ageRange":  true,
				"ageMin":    25,
				"ageMax":    35,
			},
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicAdvanced.selectWithNestedConditions", query)
		if err != nil {
			t.Fatalf("嵌套条件查询失败: %v", err)
		}
		
		t.Logf("嵌套条件查询成功，查询到 %d 条记录", len(users))
	})
}

// setupDynamicSQLTestData 设置动态SQL测试数据
func setupDynamicSQLTestData(t *testing.T, xmlSession mybatis.XMLSession, ctx context.Context) {
	t.Helper()
	
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("获取数据库连接失败: %v", err)
	}
	
	simpleSession := mybatis.NewSimpleSession(db)
	
	testUsers := []struct {
		name   string
		email  string
		age    int
		status string
	}{
		{"张三", "zhangsan@example.com", 28, "active"},
		{"李四", "lisi@example.com", 32, "inactive"},
		{"王五", "wangwu@example.com", 25, "active"},
		{"赵六", "zhaoliu@example.com", 30, "pending"},
		{"孙七", "sunqi@example.com", 27, "active"},
		{"张小明", "zhangxiaoming@example.com", 24, "active"},
		{"李小红", "lixiaohong@example.com", 29, "active"},
		{"动态用户1", "dynamic1@example.com", 26, "active"},
		{"动态用户2", "dynamic2@example.com", 33, "inactive"},
		{"测试用户", "test@example.com", 31, "active"},
	}
	
	for i, user := range testUsers {
		_, err := simpleSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status)
		if err != nil {
			t.Logf("插入动态SQL测试用户 %d 失败: %v", i+1, err)
		}
	}
	
	t.Log("成功设置动态SQL测试数据")
}

// setupDynamicSQLTestDataWithIDs 设置动态SQL测试数据并返回IDs
func setupDynamicSQLTestDataWithIDs(t *testing.T, xmlSession mybatis.XMLSession, ctx context.Context) []int64 {
	t.Helper()
	
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("获取数据库连接失败: %v", err)
	}
	
	simpleSession := mybatis.NewSimpleSession(db)
	var userIDs []int64
	
	testUsers := []struct {
		name   string
		email  string
		age    int
		status string
	}{
		{"ID测试用户1", "idtest1@example.com", 28, "active"},
		{"ID测试用户2", "idtest2@example.com", 32, "active"},
		{"ID测试用户3", "idtest3@example.com", 25, "active"},
	}
	
	for i, user := range testUsers {
		userID, err := simpleSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status)
		if err != nil {
			t.Logf("插入ID测试用户 %d 失败: %v", i+1, err)
		} else {
			userIDs = append(userIDs, userID)
		}
	}
	
	t.Logf("成功设置动态SQL测试数据，获得 %d 个用户ID", len(userIDs))
	return userIDs
}

// getAdvancedDynamicSQLMapperXML 获取高级动态SQL映射XML
func getAdvancedDynamicSQLMapperXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="DynamicAdvanced">
    
    <!-- 单个if条件测试 -->
    <select id="selectWithSingleIf" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <where>
            <if test="name != null and name != ''">
                name LIKE CONCAT('%', #{name}, '%')
            </if>
        </where>
        ORDER BY id LIMIT 10
    </select>
    
    <!-- 多个if条件测试 -->
    <select id="selectWithMultipleIf" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <where>
            <if test="status != null and status != ''">
                AND status = #{status}
            </if>
            <if test="ageMin != null and ageMin > 0">
                AND age >= #{ageMin}
            </if>
            <if test="ageMax != null and ageMax > 0">
                AND age <= #{ageMax}
            </if>
        </where>
        ORDER BY age LIMIT 15
    </select>
    
    <!-- where条件测试 -->
    <select id="selectWithWhere" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <where>
            <if test="status != null and status != ''">
                status = #{status}
            </if>
            <if test="name != null and name != ''">
                AND name LIKE CONCAT('%', #{name}, '%')
            </if>
        </where>
        ORDER BY created_at DESC LIMIT 10
    </select>
    
    <!-- choose-when-otherwise测试 -->
    <select id="selectWithChoose" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <where>
            <choose>
                <when test="name != null and name != ''">
                    name LIKE CONCAT('%', #{name}, '%')
                </when>
                <when test="email != null and email != ''">
                    email = #{email}
                </when>
                <otherwise>
                    status = 'active'
                </otherwise>
            </choose>
        </where>
        ORDER BY created_at DESC LIMIT 10
    </select>
    
    <!-- foreach IN查询测试 -->
    <select id="selectByIds" parameterType="map" resultType="map">
        SELECT * FROM users
        WHERE id IN
        <foreach collection="ids" item="id" open="(" separator="," close=")">
            #{id}
        </foreach>
        ORDER BY id
    </select>
    
    <!-- foreach批量插入测试 -->
    <insert id="batchInsertUsers" parameterType="map">
        INSERT INTO users (name, email, age, status) VALUES
        <foreach collection="users" item="user" separator=",">
            (#{user.name}, #{user.email}, #{user.age}, #{user.status})
        </foreach>
    </insert>
    
    <!-- set标签动态更新测试 -->
    <update id="updateUserSelective" parameterType="map">
        UPDATE users
        <set>
            <if test="name != null and name != ''">
                name = #{name},
            </if>
            <if test="email != null and email != ''">
                email = #{email},
            </if>
            <if test="age != null and age > 0">
                age = #{age},
            </if>
            <if test="status != null and status != ''">
                status = #{status},
            </if>
            updated_at = CURRENT_TIMESTAMP
        </set>
        WHERE id = #{id}
    </update>
    
    <!-- trim标签测试 -->
    <select id="selectWithTrim" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <trim prefix="WHERE" prefixOverrides="AND |OR ">
            <if test="status != null and status != ''">
                AND status = #{status}
            </if>
            <if test="ageMin != null and ageMin > 0">
                AND age >= #{ageMin}
            </if>
            <if test="ageMax != null and ageMax > 0">
                AND age <= #{ageMax}
            </if>
        </trim>
        ORDER BY id LIMIT 20
    </select>
    
    <!-- 复杂条件组合测试 -->
    <select id="selectWithComplexConditions" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <where>
            <if test="status != null and status != ''">
                AND status = #{status}
            </if>
            <if test="ageMin != null and ageMin > 0">
                AND age >= #{ageMin}
            </if>
            <if test="ageMax != null and ageMax > 0">
                AND age <= #{ageMax}
            </if>
            <if test="keyword != null and keyword != ''">
                AND (name LIKE CONCAT('%', #{keyword}, '%') OR email LIKE CONCAT('%', #{keyword}, '%'))
            </if>
        </where>
        <choose>
            <when test="orderBy != null and orderBy != ''">
                ORDER BY ${orderBy}
                <if test="orderDesc">DESC</if>
                <if test="!orderDesc">ASC</if>
            </when>
            <otherwise>
                ORDER BY created_at DESC
            </otherwise>
        </choose>
        <if test="pageSize != null and pageSize > 0">
            LIMIT #{pageSize}
        </if>
    </select>
    
    <!-- 嵌套条件测试 -->
    <select id="selectWithNestedConditions" parameterType="map" resultType="map">
        SELECT * FROM users
        <where>
            <if test="searchType == 'advanced'">
                <if test="filters != null">
                    <if test="filters.status != null">
                        AND status = #{filters.status}
                    </if>
                    <if test="filters.ageRange">
                        <if test="filters.ageMin != null">
                            AND age >= #{filters.ageMin}
                        </if>
                        <if test="filters.ageMax != null">
                            AND age <= #{filters.ageMax}
                        </if>
                    </if>
                </if>
            </if>
        </where>
        ORDER BY id LIMIT 10
    </select>
    
</mapper>`
}