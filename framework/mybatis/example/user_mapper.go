// Package example MyBatis框架使用示例
//
// 演示如何使用MyBatis-Go框架进行数据库操作
package example

import (
	"reflect"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis/config"
	"github.com/zsy619/yyhertz/framework/mybatis/session"
)

// User 用户模型
type User struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Age       int       `json:"age" db:"age"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UserQuery 用户查询参数
type UserQuery struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	AgeMin   int    `json:"age_min"`
	AgeMax   int    `json:"age_max"`
	Status   string `json:"status"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// UserMapper 用户映射器接口
type UserMapper interface {
	// SelectById 根据ID查询用户
	SelectById(id int64) (*User, error)
	
	// SelectByEmail 根据邮箱查询用户
	SelectByEmail(email string) (*User, error)
	
	// SelectList 查询用户列表
	SelectList(query *UserQuery) ([]*User, error)
	
	// SelectCount 查询用户总数
	SelectCount(query *UserQuery) (int64, error)
	
	// Insert 插入用户
	Insert(user *User) (int64, error)
	
	// Update 更新用户
	Update(user *User) (int64, error)
	
	// UpdateSelective 选择性更新用户
	UpdateSelective(user *User) (int64, error)
	
	// Delete 删除用户
	Delete(id int64) (int64, error)
	
	// BatchInsert 批量插入用户
	BatchInsert(users []*User) (int64, error)
}

// UserMapperImpl 用户映射器实现 (用于注册SQL映射)
type UserMapperImpl struct {
	sqlSession session.SqlSession
}

// NewUserMapper 创建用户映射器
func NewUserMapper(sqlSession session.SqlSession) UserMapper {
	return &UserMapperImpl{
		sqlSession: sqlSession,
	}
}

// 实现UserMapper接口的方法

// SelectById 根据ID查询用户
func (mapper *UserMapperImpl) SelectById(id int64) (*User, error) {
	result, err := mapper.sqlSession.SelectOne("UserMapper.SelectById", id)
	if err != nil {
		return nil, err
	}
	
	if user, ok := result.(*User); ok {
		return user, nil
	}
	
	return nil, nil
}

// SelectByEmail 根据邮箱查询用户
func (mapper *UserMapperImpl) SelectByEmail(email string) (*User, error) {
	result, err := mapper.sqlSession.SelectOne("UserMapper.SelectByEmail", email)
	if err != nil {
		return nil, err
	}
	
	if user, ok := result.(*User); ok {
		return user, nil
	}
	
	return nil, nil
}

// SelectList 查询用户列表
func (mapper *UserMapperImpl) SelectList(query *UserQuery) ([]*User, error) {
	results, err := mapper.sqlSession.SelectList("UserMapper.SelectList", query)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, len(results))
	for i, result := range results {
		if user, ok := result.(*User); ok {
			users[i] = user
		}
	}
	
	return users, nil
}

// SelectCount 查询用户总数
func (mapper *UserMapperImpl) SelectCount(query *UserQuery) (int64, error) {
	result, err := mapper.sqlSession.SelectOne("UserMapper.SelectCount", query)
	if err != nil {
		return 0, err
	}
	
	if count, ok := result.(int64); ok {
		return count, nil
	}
	
	return 0, nil
}

// Insert 插入用户
func (mapper *UserMapperImpl) Insert(user *User) (int64, error) {
	return mapper.sqlSession.Insert("UserMapper.Insert", user)
}

// Update 更新用户
func (mapper *UserMapperImpl) Update(user *User) (int64, error) {
	return mapper.sqlSession.Update("UserMapper.Update", user)
}

// UpdateSelective 选择性更新用户
func (mapper *UserMapperImpl) UpdateSelective(user *User) (int64, error) {
	return mapper.sqlSession.Update("UserMapper.UpdateSelective", user)
}

// Delete 删除用户
func (mapper *UserMapperImpl) Delete(id int64) (int64, error) {
	return mapper.sqlSession.Delete("UserMapper.Delete", id)
}

// BatchInsert 批量插入用户
func (mapper *UserMapperImpl) BatchInsert(users []*User) (int64, error) {
	return mapper.sqlSession.Insert("UserMapper.BatchInsert", users)
}

// RegisterUserMapperStatements 注册用户映射器的SQL语句
func RegisterUserMapperStatements(configuration *config.Configuration) error {
	// 注册映射器
	userMapperType := reflect.TypeOf((*UserMapper)(nil)).Elem()
	err := configuration.GetMapperRegistry().RegisterMapper(userMapperType)
	if err != nil {
		return err
	}
	
	// 这里应该注册各种SQL语句到MappedStatement
	// 由于简化实现，这里只是示例代码结构
	
	return nil
}

// SQL语句常量 (实际应该在XML文件或注解中定义)
const (
	SelectByIdSQL = `
		SELECT id, name, email, age, status, created_at, updated_at 
		FROM users 
		WHERE id = #{id}
	`
	
	SelectByEmailSQL = `
		SELECT id, name, email, age, status, created_at, updated_at 
		FROM users 
		WHERE email = #{email}
	`
	
	SelectListSQL = `
		SELECT id, name, email, age, status, created_at, updated_at 
		FROM users 
		<where>
			<if test="name != null and name != ''">
				AND name LIKE CONCAT('%', #{name}, '%')
			</if>
			<if test="email != null and email != ''">
				AND email = #{email}
			</if>
			<if test="ageMin > 0">
				AND age >= #{ageMin}
			</if>
			<if test="ageMax > 0">
				AND age <= #{ageMax}
			</if>
			<if test="status != null and status != ''">
				AND status = #{status}
			</if>
		</where>
		ORDER BY id DESC
		<if test="page > 0 and pageSize > 0">
			LIMIT #{pageSize} OFFSET #{offset}
		</if>
	`
	
	SelectCountSQL = `
		SELECT COUNT(*) 
		FROM users 
		<where>
			<if test="name != null and name != ''">
				AND name LIKE CONCAT('%', #{name}, '%')
			</if>
			<if test="email != null and email != ''">
				AND email = #{email}
			</if>
			<if test="ageMin > 0">
				AND age >= #{ageMin}
			</if>
			<if test="ageMax > 0">
				AND age <= #{ageMax}
			</if>
			<if test="status != null and status != ''">
				AND status = #{status}
			</if>
		</where>
	`
	
	InsertSQL = `
		INSERT INTO users (name, email, age, status, created_at, updated_at)
		VALUES (#{name}, #{email}, #{age}, #{status}, #{createdAt}, #{updatedAt})
	`
	
	UpdateSQL = `
		UPDATE users 
		SET name = #{name}, email = #{email}, age = #{age}, 
		    status = #{status}, updated_at = #{updatedAt}
		WHERE id = #{id}
	`
	
	UpdateSelectiveSQL = `
		UPDATE users 
		<set>
			<if test="name != null and name != ''">
				name = #{name},
			</if>
			<if test="email != null and email != ''">
				email = #{email},
			</if>
			<if test="age > 0">
				age = #{age},
			</if>
			<if test="status != null and status != ''">
				status = #{status},
			</if>
			updated_at = NOW()
		</set>
		WHERE id = #{id}
	`
	
	DeleteSQL = `
		DELETE FROM users WHERE id = #{id}
	`
	
	BatchInsertSQL = `
		INSERT INTO users (name, email, age, status, created_at, updated_at)
		VALUES
		<foreach collection="users" item="user" separator=",">
			(#{user.name}, #{user.email}, #{user.age}, #{user.status}, #{user.createdAt}, #{user.updatedAt})
		</foreach>
	`
)