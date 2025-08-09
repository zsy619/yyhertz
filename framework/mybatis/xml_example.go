// Package mybatis XML Mapper使用示例
//
// 展示如何使用XML Mapper文件，体现与Java MyBatis的兼容性
package mybatis

import (
	"context"
	"log"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserXMLService 使用XML Mapper的用户服务
type UserXMLService struct {
	session XMLSession
}

// UserQueryXML 用户查询参数
type UserQueryXML struct {
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Status   string    `json:"status"`
	CreateAtStart *time.Time `json:"createAtStart"`
	CreateAtEnd   *time.Time `json:"createAtEnd"`
	PageNum  int       `json:"pageNum"`
	PageSize int       `json:"pageSize"`
}

// NewUserXMLService 创建使用XML Mapper的用户服务
func NewUserXMLService(db *gorm.DB) (*UserXMLService, error) {
	session := NewXMLSessionWithHooks(db, true)
	
	service := &UserXMLService{
		session: session,
	}
	
	// 加载XML映射文件
	// 这里可以从不同位置加载
	if err := service.loadMappers(); err != nil {
		return nil, err
	}
	
	return service, nil
}

// loadMappers 加载映射文件
func (s *UserXMLService) loadMappers() error {
	// 方式1: 直接加载单个XML文件
	// err := s.session.LoadMapperXML("./mappers/UserMapper.xml")
	
	// 方式2: 从字符串加载（用于演示）
	userMapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="com.example.UserMapper">

  <!-- 基础查询 -->
  <select id="getUserById" parameterType="long" resultType="User">
    SELECT id, name, email, status, create_at 
    FROM users 
    WHERE id = #{id}
  </select>
  
  <!-- 带ResultMap的查询 -->
  <resultMap id="userResultMap" type="User">
    <id property="id" column="user_id"/>
    <result property="name" column="user_name"/>
    <result property="email" column="user_email"/>
    <result property="status" column="user_status"/>
    <result property="createAt" column="create_time" javaType="date"/>
  </resultMap>
  
  <select id="getUserWithMapping" parameterType="long" resultMap="userResultMap">
    SELECT user_id, user_name, user_email, user_status, create_time
    FROM users 
    WHERE user_id = #{id}
  </select>
  
  <!-- 动态SQL查询 -->
  <select id="findUsersByCondition" parameterType="UserQueryXML" resultType="User">
    SELECT id, name, email, status, create_at FROM users
    <where>
      <if test="name != null and name != ''">
        AND name LIKE #{name}
      </if>
      <if test="email != null and email != ''">
        AND email = #{email}
      </if>
      <if test="status != null and status != ''">
        AND status = #{status}
      </if>
      <if test="createAtStart != null">
        AND create_at >= #{createAtStart}
      </if>
      <if test="createAtEnd != null">
        AND create_at &lt;= #{createAtEnd}
      </if>
    </where>
    ORDER BY create_at DESC
  </select>
  
  <!-- 统计查询 -->
  <select id="countUsersByStatus" parameterType="string" resultType="int">
    SELECT COUNT(*) FROM users WHERE status = #{status}
  </select>
  
  <!-- 插入操作 -->
  <insert id="insertUser" parameterType="User" useGeneratedKeys="true" keyProperty="id">
    INSERT INTO users (name, email, status, create_at)
    VALUES (#{name}, #{email}, #{status}, #{createAt})
  </insert>
  
  <!-- 批量插入 -->
  <insert id="insertUsers" parameterType="list">
    INSERT INTO users (name, email, status, create_at) VALUES
    <foreach collection="list" item="user" separator=",">
      (#{user.name}, #{user.email}, #{user.status}, #{user.createAt})
    </foreach>
  </insert>
  
  <!-- 动态更新 -->
  <update id="updateUser" parameterType="User">
    UPDATE users 
    <set>
      <if test="name != null">name = #{name},</if>
      <if test="email != null">email = #{email},</if>
      <if test="status != null">status = #{status},</if>
    </set>
    WHERE id = #{id}
  </update>
  
  <!-- 条件删除 -->
  <delete id="deleteUsersByStatus" parameterType="string">
    DELETE FROM users WHERE status = #{status}
  </delete>

</mapper>`
	
	if err := s.session.LoadMapperXMLFromString(userMapperXML); err != nil {
		return err
	}
	
	// 方式3: 批量加载目录（实际使用时推荐）
	// err := s.session.LoadMapperDirectory("./mappers")
	
	return nil
}

// GetUserById 根据ID获取用户
func (s *UserXMLService) GetUserById(ctx context.Context, id int64) (*User, error) {
	result, err := s.session.SelectOneByID(ctx, "com.example.UserMapper.getUserById", id)
	if err != nil {
		return nil, err
	}
	
	if result == nil {
		return nil, nil
	}
	
	// 简单的结果转换
	if userMap, ok := result.(map[string]interface{}); ok {
		return mapToUser(userMap), nil
	}
	
	return nil, nil
}

// GetUserWithMapping 使用ResultMap获取用户
func (s *UserXMLService) GetUserWithMapping(ctx context.Context, id int64) (*User, error) {
	result, err := s.session.SelectOneByID(ctx, "com.example.UserMapper.getUserWithMapping", id)
	if err != nil {
		return nil, err
	}
	
	if result == nil {
		return nil, nil
	}
	
	if userMap, ok := result.(map[string]interface{}); ok {
		return mapToUser(userMap), nil
	}
	
	return nil, nil
}

// FindUsersByCondition 根据条件查询用户（动态SQL）
func (s *UserXMLService) FindUsersByCondition(ctx context.Context, query UserQueryXML) ([]*User, error) {
	results, err := s.session.SelectListByID(ctx, "com.example.UserMapper.findUsersByCondition", query)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0, len(results))
	for _, result := range results {
		if userMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(userMap))
		}
	}
	
	return users, nil
}

// FindUsersByConditionWithPage 分页查询用户
func (s *UserXMLService) FindUsersByConditionWithPage(ctx context.Context, query UserQueryXML) (*PageResult, error) {
	pageReq := PageRequest{
		Page: query.PageNum,
		Size: query.PageSize,
	}
	
	pageResult, err := s.session.SelectPageByID(ctx, "com.example.UserMapper.findUsersByCondition", query, pageReq)
	if err != nil {
		return nil, err
	}
	
	// 转换分页结果中的用户数据
	users := make([]interface{}, len(pageResult.Items))
	for i, item := range pageResult.Items {
		if userMap, ok := item.(map[string]interface{}); ok {
			users[i] = mapToUser(userMap)
		} else {
			users[i] = item
		}
	}
	pageResult.Items = users
	
	return pageResult, nil
}

// CountUsersByStatus 统计指定状态的用户数量
func (s *UserXMLService) CountUsersByStatus(ctx context.Context, status string) (int64, error) {
	result, err := s.session.SelectOneByID(ctx, "com.example.UserMapper.countUsersByStatus", status)
	if err != nil {
		return 0, err
	}
	
	if countResult, ok := result.(map[string]interface{}); ok {
		if count, exists := countResult["COUNT(*)"]; exists {
			if countVal, ok := count.(int64); ok {
				return countVal, nil
			}
		}
	}
	
	return 0, nil
}

// CreateUser 创建用户
func (s *UserXMLService) CreateUser(ctx context.Context, user *User) (int64, error) {
	if user.CreateAt.IsZero() {
		user.CreateAt = time.Now()
	}
	
	return s.session.InsertByID(ctx, "com.example.UserMapper.insertUser", user)
}

// CreateUsers 批量创建用户
func (s *UserXMLService) CreateUsers(ctx context.Context, users []*User) (int64, error) {
	for _, user := range users {
		if user.CreateAt.IsZero() {
			user.CreateAt = time.Now()
		}
	}
	
	return s.session.InsertByID(ctx, "com.example.UserMapper.insertUsers", users)
}

// UpdateUser 更新用户（动态SQL）
func (s *UserXMLService) UpdateUser(ctx context.Context, user *User) (int64, error) {
	return s.session.UpdateByID(ctx, "com.example.UserMapper.updateUser", user)
}

// DeleteUsersByStatus 根据状态删除用户
func (s *UserXMLService) DeleteUsersByStatus(ctx context.Context, status string) (int64, error) {
	return s.session.DeleteByID(ctx, "com.example.UserMapper.deleteUsersByStatus", status)
}

// ShowMapperInfo 显示已加载的Mapper信息
func (s *UserXMLService) ShowMapperInfo() {
	namespaces := s.session.GetNamespaces()
	log.Printf("已加载的命名空间数量: %d", len(namespaces))
	
	for _, namespace := range namespaces {
		statementIds := s.session.GetStatementIds(namespace)
		log.Printf("命名空间: %s, 语句数量: %d", namespace, len(statementIds))
		
		for _, statementId := range statementIds {
			stmt := s.session.GetStatement(statementId)
			if stmt != nil {
				log.Printf("  - %s [%s] %s", statementId, stmt.StatementType, 
					truncateSQL(stmt.SQL, 50))
			}
		}
	}
}

// truncateSQL 截断SQL用于显示
func truncateSQL(sql string, maxLen int) string {
	// 移除多余的空白字符
	sql = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(sql), " ")
	
	if len(sql) <= maxLen {
		return sql
	}
	
	return sql[:maxLen] + "..."
}


// 示例使用函数

// ExampleXMLMapperBasicUsage XML Mapper基础使用示例
func ExampleXMLMapperBasicUsage(db *gorm.DB) {
	// 创建服务
	userService, err := NewUserXMLService(db)
	if err != nil {
		log.Printf("Error creating XML service: %v", err)
		return
	}
	
	// 显示Mapper信息
	userService.ShowMapperInfo()
	
	ctx := context.Background()
	
	// 基础查询
	user, err := userService.GetUserById(ctx, 1)
	if err != nil {
		log.Printf("Error getting user: %v", err)
		return
	}
	
	if user != nil {
		log.Printf("Found user: %+v", user)
	}
	
	// 动态查询
	query := UserQueryXML{
		Name:   "%john%",
		Status: "active",
	}
	
	users, err := userService.FindUsersByCondition(ctx, query)
	if err != nil {
		log.Printf("Error finding users: %v", err)
		return
	}
	
	log.Printf("Found %d users matching condition", len(users))
}

// ExampleXMLMapperAdvancedUsage XML Mapper高级功能示例
func ExampleXMLMapperAdvancedUsage(db *gorm.DB) {
	userService, _ := NewUserXMLService(db)
	ctx := context.Background()
	
	// 分页查询
	query := UserQueryXML{
		Status:   "active",
		PageNum:  1,
		PageSize: 10,
	}
	
	pageResult, err := userService.FindUsersByConditionWithPage(ctx, query)
	if err != nil {
		log.Printf("Error in paged query: %v", err)
		return
	}
	
	log.Printf("Page result: %d items, total: %d, pages: %d", 
		len(pageResult.Items), pageResult.Total, pageResult.TotalPages)
	
	// 统计查询
	count, err := userService.CountUsersByStatus(ctx, "active")
	if err != nil {
		log.Printf("Error counting users: %v", err)
		return
	}
	
	log.Printf("Active users count: %d", count)
	
	// 创建用户
	newUser := &User{
		Name:  "XML User",
		Email: "xml@example.com",
	}
	
	_, err = userService.CreateUser(ctx, newUser)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return
	}
	
	log.Println("User created successfully via XML mapper")
}

// ExampleDirectXMLSession 直接使用XML会话示例
func ExampleDirectXMLSession(db *gorm.DB) {
	// 直接创建XML会话
	session := NewXMLMapper(db)
	
	// 加载XML（实际使用中应该从文件加载）
	simpleMapperXML := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE mapper PUBLIC "-//mybatis.org//DTD Mapper 3.0//EN" 
  "http://mybatis.org/dtd/mybatis-3-mapper.dtd">
<mapper namespace="SimpleMapper">
  <select id="getUserCount" resultType="int">
    SELECT COUNT(*) FROM users
  </select>
  
  <select id="getActiveUsers" resultType="User">
    SELECT * FROM users WHERE status = 'active' LIMIT 10
  </select>
</mapper>`
	
	err := session.LoadMapperXMLFromString(simpleMapperXML)
	if err != nil {
		log.Printf("Error loading XML: %v", err)
		return
	}
	
	ctx := context.Background()
	
	// 直接使用语句ID查询
	countResult, err := session.SelectOneByID(ctx, "SimpleMapper.getUserCount", nil)
	if err != nil {
		log.Printf("Error getting count: %v", err)
		return
	}
	
	log.Printf("Total users: %+v", countResult)
	
	// 查询活跃用户列表
	activeUsers, err := session.SelectListByID(ctx, "SimpleMapper.getActiveUsers", nil)
	if err != nil {
		log.Printf("Error getting active users: %v", err)
		return
	}
	
	log.Printf("Found %d active users", len(activeUsers))
}