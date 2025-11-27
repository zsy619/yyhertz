// Package main XMLSession映射器测试
//
// 测试XMLSession的XML映射功能：
// 1. XML映射文件加载
// 2. 动态SQL执行
// 3. 参数映射（#{}和${}）
// 4. 结果映射
// 5. 分页查询
// 6. 复杂查询条件
package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestXMLSessionBasic 测试XML映射器基础功能
func TestXMLSessionBasic(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	// 创建XMLSession
	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	t.Run("加载XML映射", func(t *testing.T) {
		// 加载XML映射
		err := xmlSession.LoadMapperXMLFromString(getUserMapperXML())
		if err != nil {
			t.Fatalf("加载XML映射失败: %v", err)
		}
		
		t.Log("成功加载XML映射文件")
	})

	t.Run("基础查询操作", func(t *testing.T) {
		// 先插入测试数据
		setupXMLTestData(t, xmlSession, ctx)

		// 通过ID查询用户
		user, err := xmlSession.SelectOneByID(ctx, "UserMapper.selectById", 1)
		if err != nil {
			t.Fatalf("XML查询失败: %v", err)
		}
		
		if user == nil {
			t.Fatal("期望查询到用户，但结果为nil")
		}
		
		// 验证结果
		userMap, ok := user.(map[string]interface{})
		if !ok {
			t.Fatalf("期望结果类型为map[string]interface{}，得到: %T", user)
		}
		
		t.Logf("XML查询成功: %+v", userMap)
	})

	t.Run("动态SQL条件查询", func(t *testing.T) {
		// 测试动态SQL查询
		query := UserQuery{
			Status: "active",
			AgeMin: 25,
			AgeMax: 35,
		}
		
		users, err := xmlSession.SelectListByID(ctx, "UserMapper.selectByCondition", query)
		if err != nil {
			t.Fatalf("动态SQL查询失败: %v", err)
		}
		
		t.Logf("动态SQL查询成功，共查询到 %d 条记录", len(users))
		
		// 验证查询条件生效（这里只是基本验证）
		for i, user := range users {
			userMap := user.(map[string]interface{})
			if userMap["status"] != "active" {
				t.Errorf("记录 %d 状态不符合查询条件: %v", i, userMap["status"])
			}
		}
	})
}

// TestXMLSessionParameterMapping 测试参数映射功能
func TestXMLSessionParameterMapping(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(getParameterTestMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	t.Run("#{} 参数映射测试", func(t *testing.T) {
		// 插入测试数据
		setupXMLTestData(t, xmlSession, ctx)

		// 测试预处理参数映射
		user, err := xmlSession.SelectOneByID(ctx, "ParameterMapper.selectByIdPrepared", 1)
		if err != nil {
			t.Fatalf("#{}参数映射失败: %v", err)
		}
		
		if user == nil {
			t.Error("期望查询到用户，但结果为nil")
		}
		
		t.Log("#{}参数映射测试通过")
	})

	t.Run("${} 字符串替换测试", func(t *testing.T) {
		// 测试字符串直接替换（注意：这里只用于安全的场景，如表名、列名）
		users, err := xmlSession.SelectListByID(ctx, "ParameterMapper.selectByFieldValue", map[string]interface{}{
			"field": "status",
			"value": "active",
		})
		if err != nil {
			t.Fatalf("${}字符串替换失败: %v", err)
		}
		
		t.Logf("${}字符串替换测试通过，查询到 %d 条记录", len(users))
	})

	t.Run("复合参数映射测试", func(t *testing.T) {
		// 测试同时使用#{}和${}的场景
		query := map[string]interface{}{
			"tableName": "users",      // 用于${}替换
			"status":    "active",     // 用于#{}预处理
			"minAge":    25,           // 用于#{}预处理
		}
		
		users, err := xmlSession.SelectListByID(ctx, "ParameterMapper.selectByMixedParams", query)
		if err != nil {
			t.Fatalf("复合参数映射失败: %v", err)
		}
		
		t.Logf("复合参数映射测试通过，查询到 %d 条记录", len(users))
	})
}

// TestXMLSessionDynamicSQL 测试动态SQL功能
func TestXMLSessionDynamicSQL(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(getDynamicSQLMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 插入测试数据
	setupXMLTestData(t, xmlSession, ctx)

	t.Run("if条件测试", func(t *testing.T) {
		// 测试有条件的查询
		query := UserQuery{
			Name:   "张",
			Status: "active",
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectWithIf", query)
		if err != nil {
			t.Fatalf("if条件查询失败: %v", err)
		}
		
		t.Logf("if条件查询成功，查询到 %d 条记录", len(users))

		// 测试无条件的查询
		emptyQuery := UserQuery{}
		users2, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectWithIf", emptyQuery)
		if err != nil {
			t.Fatalf("无条件查询失败: %v", err)
		}
		
		t.Logf("无条件查询成功，查询到 %d 条记录", len(users2))
		
		// 验证有条件查询的记录数应该 <= 无条件查询
		if len(users) > len(users2) {
			t.Error("有条件查询的记录数不应该大于无条件查询")
		}
	})

	t.Run("where条件测试", func(t *testing.T) {
		query := UserQuery{
			Status: "active",
			AgeMin: 20,
			AgeMax: 50,
		}
		
		users, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectWithWhere", query)
		if err != nil {
			t.Fatalf("where条件查询失败: %v", err)
		}
		
		t.Logf("where条件查询成功，查询到 %d 条记录", len(users))
	})

	t.Run("choose-when-otherwise测试", func(t *testing.T) {
		// 测试choose分支
		query1 := UserQuery{Name: "张三"}
		users1, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectWithChoose", query1)
		if err != nil {
			t.Fatalf("choose-when查询失败: %v", err)
		}
		
		query2 := UserQuery{Email: "test@example.com"}
		users2, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectWithChoose", query2)
		if err != nil {
			t.Fatalf("choose-when查询失败: %v", err)
		}
		
		query3 := UserQuery{} // otherwise分支
		users3, err := xmlSession.SelectListByID(ctx, "DynamicMapper.selectWithChoose", query3)
		if err != nil {
			t.Fatalf("choose-otherwise查询失败: %v", err)
		}
		
		t.Logf("choose测试通过 - name分支: %d条, email分支: %d条, otherwise分支: %d条", 
			len(users1), len(users2), len(users3))
	})
}

// TestXMLSessionPagination 测试XML分页查询
func TestXMLSessionPagination(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载XML映射
	err = xmlSession.LoadMapperXMLFromString(getUserMapperXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 准备测试数据
	setupXMLPaginationData(t, xmlSession, ctx, 30)

	t.Run("XML分页查询", func(t *testing.T) {
		query := UserQuery{
			Status: "active",
		}
		
		pageRequest := mybatis.PageRequest{
			Page: 1,
			Size: 5,
		}
		
		pageResult, err := xmlSession.SelectPageByID(ctx, "UserMapper.selectByCondition", query, pageRequest)
		if err != nil {
			t.Fatalf("XML分页查询失败: %v", err)
		}

		// 验证分页结果
		if pageResult.Page != 1 {
			t.Errorf("期望当前页为1，得到: %d", pageResult.Page)
		}
		if pageResult.Size != 5 {
			t.Errorf("期望页面大小为5，得到: %d", pageResult.Size)
		}
		if len(pageResult.Items) == 0 {
			t.Error("期望查询到数据，但结果为空")
		}
		
		t.Logf("XML分页查询成功 - 总记录数: %d, 当前页: %d/%d, 当前页数据: %d条", 
			pageResult.Total, pageResult.Page, pageResult.TotalPages, len(pageResult.Items))
	})
}

// TestXMLSessionResultMapping 测试结果映射
func TestXMLSessionResultMapping(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("设置测试数据库失败: %v", err)
	}
	defer cleanupDatabase(db)

	xmlSession := mybatis.NewXMLMapper(db).Debug(false)
	ctx := context.Background()

	// 加载包含结果映射的XML
	err = xmlSession.LoadMapperXMLFromString(getResultMappingXML())
	if err != nil {
		t.Fatalf("加载XML映射失败: %v", err)
	}

	// 插入测试数据
	setupXMLTestData(t, xmlSession, ctx)

	t.Run("基础结果映射", func(t *testing.T) {
		user, err := xmlSession.SelectOneByID(ctx, "ResultMapper.selectUserWithMapping", 1)
		if err != nil {
			t.Fatalf("结果映射查询失败: %v", err)
		}
		
		if user == nil {
			t.Fatal("期望查询到用户，但结果为nil")
		}
		
		// 验证结果映射
		userMap := user.(map[string]interface{})
		if _, exists := userMap["id"]; !exists {
			t.Error("结果映射中缺少id字段")
		}
		if _, exists := userMap["name"]; !exists {
			t.Error("结果映射中缺少name字段")
		}
		
		t.Logf("基础结果映射成功: %+v", userMap)
	})

	t.Run("别名结果映射", func(t *testing.T) {
		user, err := xmlSession.SelectOneByID(ctx, "ResultMapper.selectUserWithAlias", 1)
		if err != nil {
			t.Fatalf("别名结果映射查询失败: %v", err)
		}
		
		if user == nil {
			t.Fatal("期望查询到用户，但结果为nil")
		}
		
		userMap := user.(map[string]interface{})
		// 验证别名字段
		if _, exists := userMap["user_id"]; !exists {
			t.Error("别名结果映射中缺少user_id字段")
		}
		if _, exists := userMap["user_name"]; !exists {
			t.Error("别名结果映射中缺少user_name字段")
		}
		
		t.Logf("别名结果映射成功: %+v", userMap)
	})
}

// setupXMLTestData 设置XML测试数据
func setupXMLTestData(t *testing.T, session mybatis.XMLSession, ctx context.Context) {
	t.Helper()
	
	// 由于我们需要插入数据，直接使用底层数据库连接
	// 这里我们重新获取数据库连接来创建SimpleSession
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("获取数据库连接失败: %v", err)
	}
	
	// 使用简化版Session插入数据（因为我们主要测试查询）
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
	}
	
	for i, user := range testUsers {
		_, err := simpleSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			user.name, user.email, user.age, user.status)
		if err != nil {
			t.Logf("插入测试用户 %d 失败: %v", i+1, err)
		}
	}
	
	t.Log("成功设置XML测试数据")
}

// setupXMLPaginationData 设置XML分页测试数据
func setupXMLPaginationData(t *testing.T, session mybatis.XMLSession, ctx context.Context, count int) {
	t.Helper()
	
	// 获取数据库连接来创建SimpleSession
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("获取数据库连接失败: %v", err)
	}
	
	simpleSession := mybatis.NewSimpleSession(db)
	
	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("分页用户%d", i)
		email := fmt.Sprintf("page_%d@example.com", i)
		age := 20 + (i % 40) // 年龄在20-59之间
		status := []string{"active", "inactive", "pending"}[i%3]
		
		_, err := simpleSession.Insert(ctx,
			`INSERT INTO users (name, email, age, status) VALUES (?, ?, ?, ?)`,
			name, email, age, status)
		if err != nil {
			t.Logf("插入分页测试数据 %d 失败: %v", i, err)
		}
	}
	
	t.Logf("成功设置 %d 条XML分页测试数据", count)
}

// getParameterTestMapperXML 获取参数测试映射XML
func getParameterTestMapperXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="ParameterMapper">
    <!-- 预处理参数测试 -->
    <select id="selectByIdPrepared" parameterType="int" resultType="map">
        SELECT * FROM users WHERE id = #{id}
    </select>
    
    <!-- 字符串替换测试（仅用于安全场景如字段名、表名） -->
    <select id="selectByFieldValue" parameterType="map" resultType="map">
        SELECT * FROM users WHERE ${field} = #{value} LIMIT 10
    </select>
    
    <!-- 混合参数测试 -->
    <select id="selectByMixedParams" parameterType="map" resultType="map">
        SELECT * FROM ${tableName} WHERE status = #{status} AND age >= #{minAge} LIMIT 5
    </select>
</mapper>`
}

// getDynamicSQLMapperXML 获取动态SQL测试映射XML
func getDynamicSQLMapperXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="DynamicMapper">
    <!-- if条件测试 -->
    <select id="selectWithIf" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <where>
            <if test="name != null and name != ''">
                AND name LIKE CONCAT('%', #{name}, '%')
            </if>
            <if test="status != null and status != ''">
                AND status = #{status}
            </if>
            <if test="ageMin > 0">
                AND age >= #{ageMin}
            </if>
        </where>
        ORDER BY id LIMIT 20
    </select>
    
    <!-- where条件测试 -->
    <select id="selectWithWhere" parameterType="UserQuery" resultType="map">
        SELECT * FROM users
        <where>
            status = #{status}
            <if test="ageMin > 0">
                AND age >= #{ageMin}
            </if>
            <if test="ageMax > 0">
                AND age <= #{ageMax}
            </if>
        </where>
        ORDER BY age LIMIT 15
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
</mapper>`
}

// getResultMappingXML 获取结果映射测试XML
func getResultMappingXML() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="ResultMapper">
    <!-- 基础结果映射 -->
    <select id="selectUserWithMapping" parameterType="int" resultType="map">
        SELECT id, name, email, age, status, created_at FROM users WHERE id = #{id}
    </select>
    
    <!-- 别名结果映射 -->
    <select id="selectUserWithAlias" parameterType="int" resultType="map">
        SELECT 
            id as user_id, 
            name as user_name, 
            email as user_email, 
            age as user_age,
            status as user_status
        FROM users WHERE id = #{id}
    </select>
</mapper>`
}