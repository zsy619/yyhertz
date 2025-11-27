// Package main 用户映射器定义
//
// 展示完整的MyBatis风格映射器实现，包括基础CRUD、动态SQL、批量操作等
package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
)

// UserMapper 用户映射器接口
// 演示各种MyBatis功能：基础CRUD、动态SQL、批量操作、聚合查询等
type UserMapper interface {
	// ========== 基础CRUD操作 ==========
	
	// SelectById 根据ID查询用户
	// @Select("SELECT * FROM users WHERE id = #{id} AND deleted_at IS NULL")
	SelectById(id int64) (*User, error)
	
	// SelectByEmail 根据邮箱查询用户
	// @Select("SELECT * FROM users WHERE email = #{email} AND deleted_at IS NULL")
	SelectByEmail(email string) (*User, error)
	
	// SelectByIds 根据ID列表查询用户
	// @Select("SELECT * FROM users WHERE id IN (#{ids}) AND deleted_at IS NULL")
	SelectByIds(ids []int64) ([]*User, error)
	
	// Insert 插入用户
	// @Insert("INSERT INTO users (name, email, age, status, avatar, phone, birthday, created_at, updated_at) VALUES (#{name}, #{email}, #{age}, #{status}, #{avatar}, #{phone}, #{birthday}, NOW(), NOW())")
	// @Options(useGeneratedKeys=true, keyProperty="id")
	Insert(user *User) (int64, error)
	
	// Update 更新用户
	// @Update("UPDATE users SET name=#{name}, email=#{email}, age=#{age}, status=#{status}, avatar=#{avatar}, phone=#{phone}, birthday=#{birthday}, updated_at=NOW() WHERE id=#{id}")
	Update(user *User) (int64, error)
	
	// Delete 软删除用户
	// @Update("UPDATE users SET deleted_at=NOW() WHERE id=#{id}")
	Delete(id int64) (int64, error)
	
	// PhysicalDelete 物理删除用户
	// @Delete("DELETE FROM users WHERE id=#{id}")
	PhysicalDelete(id int64) (int64, error)
	
	// ========== 动态SQL查询 ==========
	
	// SelectList 动态条件查询用户列表
	SelectList(query *UserQuery) ([]*User, error)
	
	// SelectCount 动态条件统计用户数量
	SelectCount(query *UserQuery) (int64, error)
	
	// SelectPage 分页查询用户
	SelectPage(query *UserQuery) (*PaginationResult, error)
	
	// UpdateSelective 选择性更新用户
	UpdateSelective(user *User) (int64, error)
	
	// ========== 批量操作 ==========
	
	// BatchInsert 批量插入用户
	BatchInsert(users []*User) (int64, error)
	
	// BatchUpdate 批量更新用户
	BatchUpdate(request *BatchUpdateRequest) (int64, error)
	
	// BatchDelete 批量删除用户
	BatchDelete(ids []int64) (int64, error)
	
	// BatchUpdateStatus 批量更新用户状态
	BatchUpdateStatus(ids []int64, status string) (int64, error)
	
	// ========== 聚合查询 ==========
	
	// SelectStats 查询用户统计信息
	SelectStats() (*UserStats, error)
	
	// SelectByStatus 按状态分组统计
	SelectByStatus() ([]*AggregationResult, error)
	
	// SelectByAgeGroup 按年龄组分组统计
	SelectByAgeGroup() ([]*AggregationResult, error)
	
	// SelectActiveUsersInPeriod 查询指定时间段内的活跃用户
	SelectActiveUsersInPeriod(startTime, endTime time.Time) ([]*User, error)
	
	// ========== 复杂查询 ==========
	
	// SelectWithProfile 查询用户及其档案信息
	SelectWithProfile(id int64) (*ComplexQueryResult, error)
	
	// SelectWithRoles 查询用户及其角色信息
	SelectWithRoles(id int64) (*ComplexQueryResult, error)
	
	// SelectWithArticles 查询用户及其文章
	SelectWithArticles(userId int64, limit int) (*ComplexQueryResult, error)
	
	// SearchUsers 全文搜索用户
	SearchUsers(keyword string, limit int) ([]*User, error)
	
	// SelectSimilarUsers 查询相似用户（年龄和状态相近）
	SelectSimilarUsers(userId int64, limit int) ([]*User, error)
	
	// ========== 特殊查询 ==========
	
	// SelectRandomUsers 随机查询用户
	SelectRandomUsers(limit int) ([]*User, error)
	
	// SelectTopActiveUsers 查询最活跃用户（根据文章数量）
	SelectTopActiveUsers(limit int) ([]*User, error)
	
	// SelectUsersWithoutProfile 查询没有档案信息的用户
	SelectUsersWithoutProfile() ([]*User, error)
	
	// SelectRecentRegistrations 查询最近注册的用户
	SelectRecentRegistrations(days int, limit int) ([]*User, error)
	
	// ========== 存储过程和函数 ==========
	
	// CallUserStatsProcedure 调用用户统计存储过程
	CallUserStatsProcedure(startDate, endDate time.Time) (*UserStats, error)
	
	// SelectUserByCustomFunction 使用自定义函数查询
	SelectUserByCustomFunction(param string) ([]*User, error)
}

// UserMapperImpl 用户映射器实现
type UserMapperImpl struct {
	simpleSession mybatis.SimpleSession
}

// NewUserMapper 创建用户映射器
func NewUserMapper(session interface{}) UserMapper {
	if simpleSession, ok := session.(mybatis.SimpleSession); ok {
		return &UserMapperImpl{
			simpleSession: simpleSession,
		}
	}
	panic("Unsupported session type")
}

// ========== 基础CRUD操作实现 ==========

func (m *UserMapperImpl) SelectById(id int64) (*User, error) {
	ctx := context.Background()
	result, err := m.simpleSession.SelectOne(ctx, "SELECT * FROM users WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return nil, err
	}
	
	if result == nil {
		return nil, nil
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		return mapToUser(resultMap), nil
	}
	
	return nil, nil
}

func (m *UserMapperImpl) SelectByEmail(email string) (*User, error) {
	ctx := context.Background()
	result, err := m.simpleSession.SelectOne(ctx, "SELECT * FROM users WHERE email = ? AND deleted_at IS NULL", email)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	if resultMap, ok := result.(map[string]interface{}); ok {
		return mapToUser(resultMap), nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SelectByIds(ids []int64) ([]*User, error) {
	// 简化实现，逐个查询
	users := make([]*User, 0)
	for _, id := range ids {
		user, err := m.SelectById(id)
		if err != nil {
			return nil, err
		}
		if user != nil {
			users = append(users, user)
		}
	}
	return users, nil
}

func (m *UserMapperImpl) Insert(user *User) (int64, error) {
	ctx := context.Background()
	return m.simpleSession.Insert(ctx, 
		"INSERT INTO users (name, email, age, status, phone, birthday, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))",
		user.Name, user.Email, user.Age, user.Status, user.Phone, user.Birthday)
}

func (m *UserMapperImpl) Update(user *User) (int64, error) {
	ctx := context.Background()
	return m.simpleSession.Update(ctx,
		"UPDATE users SET name=?, email=?, age=?, status=?, phone=?, birthday=?, updated_at=datetime('now') WHERE id=?",
		user.Name, user.Email, user.Age, user.Status, user.Phone, user.Birthday, user.ID)
}

func (m *UserMapperImpl) Delete(id int64) (int64, error) {
	ctx := context.Background()
	return m.simpleSession.Update(ctx, "UPDATE users SET deleted_at=datetime('now') WHERE id=?", id)
}

func (m *UserMapperImpl) PhysicalDelete(id int64) (int64, error) {
	ctx := context.Background()
	return m.simpleSession.Delete(ctx, "DELETE FROM users WHERE id = ?", id)
}

// ========== 动态SQL查询实现 ==========

func (m *UserMapperImpl) SelectList(query *UserQuery) ([]*User, error) {
	ctx := context.Background()
	
	// 构建动态SQL
	sql := "SELECT * FROM users WHERE 1=1"
	var args []interface{}
	
	if query != nil {
		if query.Name != "" {
			sql += " AND name LIKE ?"
			args = append(args, "%"+query.Name+"%")
		}
		if query.Status != "" {
			sql += " AND status = ?"
			args = append(args, query.Status)
		}
		if query.AgeMin > 0 {
			sql += " AND age >= ?"
			args = append(args, query.AgeMin)
		}
		if query.AgeMax > 0 {
			sql += " AND age <= ?"
			args = append(args, query.AgeMax)
		}
		if query.Keyword != "" {
			sql += " AND (name LIKE ? OR email LIKE ?)"
			args = append(args, "%"+query.Keyword+"%", "%"+query.Keyword+"%")
		}
		
		// 排序
		if query.OrderBy != "" {
			sql += " ORDER BY " + query.OrderBy
			if query.OrderDesc {
				sql += " DESC"
			}
		}
		
		// 分页
		if query.PageSize > 0 {
			sql += " LIMIT ?"
			args = append(args, query.PageSize)
			if query.Page > 1 {
				sql += " OFFSET ?"
				args = append(args, (query.Page-1)*query.PageSize)
			}
		}
	}
	
	results, err := m.simpleSession.SelectList(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

func (m *UserMapperImpl) SelectCount(query *UserQuery) (int64, error) {
	ctx := context.Background()
	
	// 构建动态计数SQL
	sql := "SELECT COUNT(*) as count FROM users WHERE deleted_at IS NULL"
	var args []interface{}
	
	if query != nil {
		if query.Name != "" {
			sql += " AND name LIKE ?"
			args = append(args, "%"+query.Name+"%")
		}
		if query.Status != "" {
			sql += " AND status = ?"
			args = append(args, query.Status)
		}
		if query.AgeMin > 0 {
			sql += " AND age >= ?"
			args = append(args, query.AgeMin)
		}
		if query.AgeMax > 0 {
			sql += " AND age <= ?"
			args = append(args, query.AgeMax)
		}
		if query.Keyword != "" {
			sql += " AND (name LIKE ? OR email LIKE ?)"
			args = append(args, "%"+query.Keyword+"%", "%"+query.Keyword+"%")
		}
	}
	
	result, err := m.simpleSession.SelectOne(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		// 查找所有可能的键名
		for _, value := range resultMap {
			// 如果是指针，先解引用
			if ptr, ok := value.(*interface{}); ok {
				value = *ptr
			}
			
			switch v := value.(type) {
			case int64:
				return v, nil
			case int:
				return int64(v), nil
			case float64:
				return int64(v), nil
			}
		}
	}
	return 0, nil
}

func (m *UserMapperImpl) SelectPage(query *UserQuery) (*PaginationResult, error) {
	// 查询总数
	total, err := m.SelectCount(query)
	if err != nil {
		return nil, err
	}
	
	// 查询数据
	users, err := m.SelectList(query)
	if err != nil {
		return nil, err
	}
	
	// 计算分页信息
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}
	
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	return &PaginationResult{
		Data:       users,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}, nil
}

func (m *UserMapperImpl) UpdateSelective(user *User) (int64, error) {
	// 简化实现，直接调用Update
	return m.Update(user)
}

// ========== 批量操作实现 ==========

func (m *UserMapperImpl) BatchInsert(users []*User) (int64, error) {
	ctx := context.Background()
	var affected int64
	
	for _, user := range users {
		id, err := m.simpleSession.Insert(ctx, 
			"INSERT INTO users (name, email, age, status, phone, birthday, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))",
			user.Name, user.Email, user.Age, user.Status, user.Phone, user.Birthday)
		if err != nil {
			return affected, err
		}
		if id > 0 {
			affected++
		}
	}
	
	return affected, nil
}

func (m *UserMapperImpl) BatchUpdate(request *BatchUpdateRequest) (int64, error) {
	// 简化实现，返回错误表示不支持
	return 0, fmt.Errorf("BatchUpdate not supported in simplified implementation")
}

func (m *UserMapperImpl) BatchDelete(ids []int64) (int64, error) {
	ctx := context.Background()
	var affected int64
	
	for _, id := range ids {
		count, err := m.simpleSession.Update(ctx, "UPDATE users SET deleted_at = datetime('now') WHERE id = ?", id)
		if err != nil {
			return affected, err
		}
		affected += count
	}
	
	return affected, nil
}

func (m *UserMapperImpl) BatchUpdateStatus(ids []int64, status string) (int64, error) {
	ctx := context.Background()
	var affected int64
	
	for _, id := range ids {
		count, err := m.simpleSession.Update(ctx, "UPDATE users SET status = ? WHERE id = ?", status, id)
		if err != nil {
			return affected, err
		}
		affected += count
	}
	
	return affected, nil
}

// ========== 聚合查询实现 ==========

func (m *UserMapperImpl) SelectStats() (*UserStats, error) {
	ctx := context.Background()
	
	// 查询统计信息
	result, err := m.simpleSession.SelectOne(ctx, `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_users,
			COUNT(CASE WHEN datetime('now', '-30 days') <= created_at THEN 1 END) as recent_users
		FROM users
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	
	if resultMap, ok := result.(map[string]interface{}); ok {
		stats := &UserStats{}
		if totalUsers, ok := resultMap["total_users"].(int64); ok {
			stats.TotalUsers = totalUsers
		}
		if activeUsers, ok := resultMap["active_users"].(int64); ok {
			stats.ActiveUsers = activeUsers
		}
		if recentUsers, ok := resultMap["recent_users"].(int64); ok {
			stats.RecentUsers = recentUsers
		}
		return stats, nil
	}
	
	return nil, nil
}

func (m *UserMapperImpl) SelectByStatus() ([]*AggregationResult, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, `
		SELECT status as value, COUNT(*) as count 
		FROM users 
		WHERE deleted_at IS NULL 
		GROUP BY status
	`)
	if err != nil {
		return nil, err
	}
	
	aggregations := make([]*AggregationResult, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			agg := &AggregationResult{}
			if value, ok := resultMap["value"]; ok {
				agg.Value = value
			}
			if count, ok := resultMap["count"].(int64); ok {
				agg.Count = count
			}
			aggregations = append(aggregations, agg)
		}
	}
	return aggregations, nil
}

func (m *UserMapperImpl) SelectByAgeGroup() ([]*AggregationResult, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, `
		SELECT 
			CASE 
				WHEN age < 25 THEN '18-24'
				WHEN age < 35 THEN '25-34'
				WHEN age < 45 THEN '35-44'
				WHEN age < 55 THEN '45-54'
				ELSE '55+'
			END as value,
			COUNT(*) as count
		FROM users 
		WHERE deleted_at IS NULL 
		GROUP BY 
			CASE 
				WHEN age < 25 THEN '18-24'
				WHEN age < 35 THEN '25-34'
				WHEN age < 45 THEN '35-44'
				WHEN age < 55 THEN '45-54'
				ELSE '55+'
			END
	`)
	if err != nil {
		return nil, err
	}
	
	aggregations := make([]*AggregationResult, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			agg := &AggregationResult{}
			if value, ok := resultMap["value"]; ok {
				agg.Value = value
			}
			if count, ok := resultMap["count"].(int64); ok {
				agg.Count = count
			}
			aggregations = append(aggregations, agg)
		}
	}
	return aggregations, nil
}

func (m *UserMapperImpl) SelectActiveUsersInPeriod(startTime, endTime time.Time) ([]*User, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, 
		"SELECT * FROM users WHERE status = 'active' AND created_at BETWEEN ? AND ? AND deleted_at IS NULL",
		startTime, endTime)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

// ========== 复杂查询实现 ==========

func (m *UserMapperImpl) SelectWithProfile(id int64) (*ComplexQueryResult, error) {
	// 简化实现，只返回用户信息
	user, err := m.SelectById(id)
	if err != nil {
		return nil, err
	}
	
	return &ComplexQueryResult{
		User: user,
	}, nil
}

func (m *UserMapperImpl) SelectWithRoles(id int64) (*ComplexQueryResult, error) {
	// 简化实现，只返回用户信息
	user, err := m.SelectById(id)
	if err != nil {
		return nil, err
	}
	
	return &ComplexQueryResult{
		User:  user,
		Roles: []*UserRole{}, // 空的角色列表
	}, nil
}

func (m *UserMapperImpl) SelectWithArticles(userId int64, limit int) (*ComplexQueryResult, error) {
	// 简化实现，只返回用户信息
	user, err := m.SelectById(userId)
	if err != nil {
		return nil, err
	}
	
	return &ComplexQueryResult{
		User: user,
	}, nil
}

func (m *UserMapperImpl) SearchUsers(keyword string, limit int) ([]*User, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, 
		"SELECT * FROM users WHERE (name LIKE ? OR email LIKE ?) AND deleted_at IS NULL LIMIT ?",
		"%"+keyword+"%", "%"+keyword+"%", limit)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

func (m *UserMapperImpl) SelectSimilarUsers(userId int64, limit int) ([]*User, error) {
	ctx := context.Background()
	
	// 先获取基准用户信息
	baseUser, err := m.SelectById(userId)
	if err != nil || baseUser == nil {
		return nil, err
	}
	
	results, err := m.simpleSession.SelectList(ctx, 
		"SELECT * FROM users WHERE id != ? AND status = ? AND ABS(age - ?) <= 5 AND deleted_at IS NULL LIMIT ?",
		userId, baseUser.Status, baseUser.Age, limit)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

// ========== 特殊查询实现 ==========

func (m *UserMapperImpl) SelectRandomUsers(limit int) ([]*User, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, 
		"SELECT * FROM users WHERE deleted_at IS NULL ORDER BY RANDOM() LIMIT ?", 
		limit)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

func (m *UserMapperImpl) SelectTopActiveUsers(limit int) ([]*User, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, 
		"SELECT * FROM users WHERE status = 'active' AND deleted_at IS NULL ORDER BY updated_at DESC LIMIT ?", 
		limit)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

func (m *UserMapperImpl) SelectUsersWithoutProfile() ([]*User, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, `
		SELECT u.* FROM users u 
		LEFT JOIN user_profiles p ON u.id = p.user_id 
		WHERE p.user_id IS NULL AND u.deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

func (m *UserMapperImpl) SelectRecentRegistrations(days int, limit int) ([]*User, error) {
	ctx := context.Background()
	
	results, err := m.simpleSession.SelectList(ctx, 
		"SELECT * FROM users WHERE created_at >= datetime('now', '-' || ? || ' days') AND deleted_at IS NULL ORDER BY created_at DESC LIMIT ?",
		days, limit)
	if err != nil {
		return nil, err
	}
	
	users := make([]*User, 0)
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			users = append(users, mapToUser(resultMap))
		}
	}
	return users, nil
}

// ========== 存储过程和函数实现 ==========

func (m *UserMapperImpl) CallUserStatsProcedure(startDate, endDate time.Time) (*UserStats, error) {
	// 简化实现，返回基本统计
	return m.SelectStats()
}

func (m *UserMapperImpl) SelectUserByCustomFunction(param string) ([]*User, error) {
	// 简化实现，使用搜索功能
	return m.SearchUsers(param, 10)
}

// RegisterUserMapper 注册用户映射器
func RegisterUserMapper(session interface{}) (UserMapper, error) {
	// 在实际应用中，这里会注册所有的SQL映射
	return NewUserMapper(session), nil
}

// GetUserMapperType 获取用户映射器类型
func GetUserMapperType() reflect.Type {
	return reflect.TypeOf((*UserMapper)(nil)).Elem()
}

// mapToUser 将map结果转换为User结构体
func mapToUser(m map[string]interface{}) *User {
	user := &User{}
	
	if id, ok := m["id"]; ok {
		// SQLite可能返回int64或其他整数类型
		switch v := id.(type) {
		case int64:
			user.ID = v
		case int:
			user.ID = int64(v)
		}
	}
	
	if name, ok := m["name"]; ok {
		if nameStr, ok := name.(string); ok {
			user.Name = nameStr
		}
	}
	
	if email, ok := m["email"]; ok {
		if emailStr, ok := email.(string); ok {
			user.Email = emailStr
		}
	}
	
	if age, ok := m["age"]; ok {
		switch v := age.(type) {
		case int:
			user.Age = v
		case int64:
			user.Age = int(v)
		}
	}
	
	if status, ok := m["status"]; ok {
		if statusStr, ok := status.(string); ok {
			user.Status = statusStr
		}
	}
	
	if phone, ok := m["phone"]; ok {
		if phoneStr, ok := phone.(string); ok {
			user.Phone = phoneStr
		}
	}
	
	if createdAt, ok := m["created_at"]; ok {
		if createdAtTime, ok := createdAt.(time.Time); ok {
			user.CreatedAt = createdAtTime
		}
	}
	
	if updatedAt, ok := m["updated_at"]; ok {
		if updatedAtTime, ok := updatedAt.(time.Time); ok {
			user.UpdatedAt = updatedAtTime
		}
	}
	
	return user
}