// Package mybatis XML Mapper功能测试
package mybatis

import (
	"context"
	"log"
	"testing"
	"time"
)

// 测试XML解析和基本功能
func TestXMLMapperBasicParsing(t *testing.T) {
	// 创建XML会话
	session := NewXMLMapper(nil) // DryRun模式，不需要真实数据库
	session.DryRun(true) // 启用DryRun
	
	// 测试XML内容
	testMapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="TestMapper">
  <select id="getUserById" parameterType="long" resultType="User">
    SELECT id, name, email FROM users WHERE id = #{id}
  </select>
  
  <select id="findUsers" parameterType="map" resultType="User">
    SELECT * FROM users
    <where>
      <if test="name != null">
        AND name LIKE #{name}
      </if>
      <if test="email != null">
        AND email = #{email}
      </if>
    </where>
  </select>
  
  <insert id="insertUser" parameterType="User">
    INSERT INTO users (name, email, create_at) 
    VALUES (#{name}, #{email}, #{createAt})
  </insert>
</mapper>`
	
	// 加载XML
	err := session.LoadMapperXMLFromString(testMapperXML)
	if err != nil {
		t.Fatalf("Failed to load XML: %v", err)
	}
	
	// 验证命名空间
	namespaces := session.GetNamespaces()
	if len(namespaces) != 1 || namespaces[0] != "TestMapper" {
		t.Fatalf("Expected namespace 'TestMapper', got: %v", namespaces)
	}
	
	// 验证语句
	statementIds := session.GetStatementIds("TestMapper")
	expectedIds := []string{"TestMapper.getUserById", "TestMapper.findUsers", "TestMapper.insertUser"}
	
	if len(statementIds) != len(expectedIds) {
		t.Fatalf("Expected %d statements, got %d", len(expectedIds), len(statementIds))
	}
	
	log.Println("TestXMLMapperBasicParsing passed")
}

// 测试DryRun模式下的SQL生成
func TestXMLMapperDryRun(t *testing.T) {
	session := NewXMLMapper(nil)
	session.DryRun(true)
	session.Debug(true)
	
	testMapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="DryRunMapper">
  <select id="getUser" parameterType="long" resultType="User">
    SELECT * FROM users WHERE id = #{id}
  </select>
  
  <insert id="createUser" parameterType="User">
    INSERT INTO users (name, email) VALUES (#{name}, #{email})
  </insert>
</mapper>`
	
	err := session.LoadMapperXMLFromString(testMapperXML)
	if err != nil {
		t.Fatalf("Failed to load XML: %v", err)
	}
	
	ctx := context.Background()
	
	// 测试查询（DryRun模式）
	result, err := session.SelectOneByID(ctx, "DryRunMapper.getUser", 123)
	if err != nil {
		t.Fatalf("DryRun select failed: %v", err)
	}
	
	// DryRun应该返回nil
	if result != nil {
		t.Fatal("DryRun should return nil result")
	}
	
	// 测试插入（DryRun模式）
	user := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	}
	
	affectedRows, err := session.InsertByID(ctx, "DryRunMapper.createUser", user)
	if err != nil {
		t.Fatalf("DryRun insert failed: %v", err)
	}
	
	// DryRun应该返回0
	if affectedRows != 0 {
		t.Fatalf("DryRun should return 0 affected rows, got %d", affectedRows)
	}
	
	log.Println("TestXMLMapperDryRun passed")
}

// 测试真实数据库操作
func TestXMLMapperRealDatabase(t *testing.T) {
	t.Skip("Skipping real database test - requires actual database setup")
	db := setupTestDB()
	session := NewXMLSessionWithHooks(db, true)
	
	// 加载测试mapper
	testMapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="UserTestMapper">
  <select id="getAllUsers" resultType="User">
    SELECT id, name, email, create_at FROM users ORDER BY id
  </select>
  
  <select id="getUserById" parameterType="long" resultType="User">
    SELECT id, name, email, create_at FROM users WHERE id = #{id}
  </select>
  
  <select id="findUsersByName" parameterType="string" resultType="User">
    SELECT * FROM users WHERE name LIKE #{name} ORDER BY id
  </select>
  
  <insert id="createUser" parameterType="map">
    INSERT INTO users (name, email, create_at) 
    VALUES (#{name}, #{email}, #{createAt})
  </insert>
  
  <update id="updateUserName" parameterType="map">
    UPDATE users SET name = #{name} WHERE id = #{id}
  </update>
  
  <delete id="deleteUser" parameterType="long">
    DELETE FROM users WHERE id = #{id}
  </delete>
</mapper>`
	
	err := session.LoadMapperXMLFromString(testMapperXML)
	if err != nil {
		t.Fatalf("Failed to load XML: %v", err)
	}
	
	ctx := context.Background()
	
	// 测试查询所有用户
	allUsers, err := session.SelectListByID(ctx, "UserTestMapper.getAllUsers", nil)
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}
	
	initialCount := len(allUsers)
	log.Printf("Initial user count: %d", initialCount)
	
	// 测试插入用户
	newUser := map[string]interface{}{
		"name":     "XML Test User",
		"email":    "xmltest@example.com",
		"createAt": time.Now(),
	}
	
	affectedRows, err := session.InsertByID(ctx, "UserTestMapper.createUser", newUser)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	
	if affectedRows != 1 {
		t.Fatalf("Expected 1 affected row, got %d", affectedRows)
	}
	
	// 验证用户被插入
	updatedUsers, err := session.SelectListByID(ctx, "UserTestMapper.getAllUsers", nil)
	if err != nil {
		t.Fatalf("Failed to get updated users: %v", err)
	}
	
	if len(updatedUsers) != initialCount+1 {
		t.Fatalf("Expected %d users after insert, got %d", initialCount+1, len(updatedUsers))
	}
	
	// 测试条件查询
	foundUsers, err := session.SelectListByID(ctx, "UserTestMapper.findUsersByName", "%XML%")
	if err != nil {
		t.Fatalf("Failed to find users by name: %v", err)
	}
	
	if len(foundUsers) == 0 {
		t.Fatal("Expected to find XML test users")
	}
	
	log.Println("TestXMLMapperRealDatabase passed")
}

// 测试ResultMap功能
func TestXMLMapperResultMap(t *testing.T) {
	t.Skip("Skipping ResultMap test - requires actual database setup")
	db := setupTestDB()
	session := NewXMLMapper(db)
	
	// 带ResultMap的XML
	resultMapXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="ResultMapTestMapper">
  <resultMap id="userResultMap" type="User">
    <id property="id" column="user_id"/>
    <result property="name" column="user_name"/>
    <result property="email" column="user_email"/>
    <result property="createAt" column="create_time" javaType="date"/>
  </resultMap>
  
  <select id="getUserWithResultMap" parameterType="long" resultMap="userResultMap">
    SELECT id as user_id, name as user_name, email as user_email, create_at as create_time
    FROM users WHERE id = #{id}
  </select>
</mapper>`
	
	err := session.LoadMapperXMLFromString(resultMapXML)
	if err != nil {
		t.Fatalf("Failed to load ResultMap XML: %v", err)
	}
	
	ctx := context.Background()
	
	// 测试ResultMap查询
	result, err := session.SelectOneByID(ctx, "ResultMapTestMapper.getUserWithResultMap", 1)
	if err != nil {
		t.Fatalf("ResultMap query failed: %v", err)
	}
	
	if result != nil {
		log.Printf("ResultMap query result: %+v", result)
	}
	
	// 验证ResultMap定义
	resultMap := session.GetResultMap("ResultMapTestMapper.userResultMap")
	if resultMap == nil {
		t.Fatal("ResultMap not found")
	}
	
	if len(resultMap.IDMappings) == 0 {
		t.Fatal("Expected ID mappings in ResultMap")
	}
	
	if len(resultMap.ResultMappings) == 0 {
		t.Fatal("Expected result mappings in ResultMap")
	}
	
	log.Println("TestXMLMapperResultMap passed")
}

// 测试动态SQL功能
func TestXMLMapperDynamicSQL(t *testing.T) {
	t.Skip("Skipping dynamic SQL test - requires actual database setup")
	db := setupTestDB()
	session := NewXMLMapper(db)
	session.Debug(true)
	
	// 动态SQL的XML
	dynamicXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="DynamicSQLMapper">
  <select id="findUsersDynamically" parameterType="map" resultType="User">
    SELECT * FROM users
    <where>
      <if test="id != null">
        AND id = #{id}
      </if>
      <if test="name != null">
        AND name LIKE #{name}
      </if>
      <if test="email != null">
        AND email = #{email}
      </if>
    </where>
    ORDER BY id
  </select>
  
  <update id="updateUserDynamically" parameterType="map">
    UPDATE users
    <set>
      <if test="name != null">name = #{name},</if>
      <if test="email != null">email = #{email},</if>
    </set>
    WHERE id = #{id}
  </update>
</mapper>`
	
	err := session.LoadMapperXMLFromString(dynamicXML)
	if err != nil {
		t.Fatalf("Failed to load dynamic SQL XML: %v", err)
	}
	
	ctx := context.Background()
	
	// 测试不同的查询条件组合
	testCases := []map[string]interface{}{
		{"id": 1},
		{"name": "%John%"},
		{"name": "%Jane%", "email": "jane@example.com"},
		{}, // 空条件
	}
	
	for i, testCase := range testCases {
		results, err := session.SelectListByID(ctx, "DynamicSQLMapper.findUsersDynamically", testCase)
		if err != nil {
			t.Fatalf("Dynamic SQL test case %d failed: %v", i, err)
		}
		
		log.Printf("Test case %d results: %d users found", i, len(results))
	}
	
	log.Println("TestXMLMapperDynamicSQL passed")
}

// 测试错误处理
func TestXMLMapperErrorHandling(t *testing.T) {
	session := NewXMLMapper(nil)
	
	// 测试加载无效XML
	invalidXML := `<invalid>xml</invalid>`
	err := session.LoadMapperXMLFromString(invalidXML)
	if err == nil {
		t.Fatal("Expected error for invalid XML, but got none")
	}
	
	// 测试查询不存在的语句
	ctx := context.Background()
	_, err = session.SelectOneByID(ctx, "NonExistent.statement", nil)
	if err == nil {
		t.Fatal("Expected error for non-existent statement, but got none")
	}
	
	// 测试类型不匹配
	validXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="ErrorTestMapper">
  <select id="testSelect" resultType="User">
    SELECT * FROM users
  </select>
</mapper>`
	
	err = session.LoadMapperXMLFromString(validXML)
	if err != nil {
		t.Fatalf("Failed to load valid XML: %v", err)
	}
	
	// 尝试用SELECT语句执行INSERT操作
	_, err = session.InsertByID(ctx, "ErrorTestMapper.testSelect", nil)
	if err == nil {
		t.Fatal("Expected error for wrong statement type, but got none")
	}
	
	log.Println("TestXMLMapperErrorHandling passed")
}

// 性能测试
func TestXMLMapperPerformance(t *testing.T) {
	t.Skip("Skipping performance test - requires actual database setup")
	db := setupTestDB()
	session := NewXMLMapper(db)
	
	// 简单的性能测试XML
	perfXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="PerfTestMapper">
  <select id="getUser" parameterType="long" resultType="User">
    SELECT * FROM users WHERE id = #{id}
  </select>
</mapper>`
	
	err := session.LoadMapperXMLFromString(perfXML)
	if err != nil {
		t.Fatalf("Failed to load performance test XML: %v", err)
	}
	
	ctx := context.Background()
	
	// 执行多次查询测试性能
	iterations := 100
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		_, err := session.SelectOneByID(ctx, "PerfTestMapper.getUser", int64(i%3+1))
		if err != nil {
			t.Fatalf("Performance test query %d failed: %v", i, err)
		}
	}
	
	duration := time.Since(start)
	avgTime := duration / time.Duration(iterations)
	
	log.Printf("Performance test: %d queries in %v, avg: %v per query", 
		iterations, duration, avgTime)
	
	// 简单的性能断言（平均每次查询不超过10ms）
	if avgTime > 10*time.Millisecond {
		t.Logf("Warning: Average query time %v may be too slow", avgTime)
	}
	
	log.Println("TestXMLMapperPerformance passed")
}


// 初始化函数，用于XML测试的准备工作
func init() {
	log.Println("Initializing MyBatis XML Mapper tests...")
}