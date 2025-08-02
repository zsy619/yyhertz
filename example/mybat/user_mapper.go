// Package mybatis_tests 用户映射器定义
//
// 展示完整的MyBatis风格映射器实现，包括基础CRUD、动态SQL、批量操作等
package mybat

import (
	"reflect"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis/session"
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
	sqlSession session.SqlSession
}

// NewUserMapper 创建用户映射器
func NewUserMapper(sqlSession session.SqlSession) UserMapper {
	return &UserMapperImpl{
		sqlSession: sqlSession,
	}
}

// ========== 基础CRUD操作实现 ==========

func (m *UserMapperImpl) SelectById(id int64) (*User, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.SelectById", id)
	if err != nil {
		return nil, err
	}
	if user, ok := result.(*User); ok {
		return user, nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SelectByEmail(email string) (*User, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.SelectByEmail", email)
	if err != nil {
		return nil, err
	}
	if user, ok := result.(*User); ok {
		return user, nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SelectByIds(ids []int64) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectByIds", map[string]any{"ids": ids})
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

func (m *UserMapperImpl) Insert(user *User) (int64, error) {
	return m.sqlSession.Insert("UserMapper.Insert", user)
}

func (m *UserMapperImpl) Update(user *User) (int64, error) {
	return m.sqlSession.Update("UserMapper.Update", user)
}

func (m *UserMapperImpl) Delete(id int64) (int64, error) {
	return m.sqlSession.Update("UserMapper.Delete", id)
}

func (m *UserMapperImpl) PhysicalDelete(id int64) (int64, error) {
	return m.sqlSession.Delete("UserMapper.PhysicalDelete", id)
}

// ========== 动态SQL查询实现 ==========

func (m *UserMapperImpl) SelectList(query *UserQuery) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectList", query)
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

func (m *UserMapperImpl) SelectCount(query *UserQuery) (int64, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.SelectCount", query)
	if err != nil {
		return 0, err
	}
	if count, ok := result.(int64); ok {
		return count, nil
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
	return m.sqlSession.Update("UserMapper.UpdateSelective", user)
}

// ========== 批量操作实现 ==========

func (m *UserMapperImpl) BatchInsert(users []*User) (int64, error) {
	return m.sqlSession.Insert("UserMapper.BatchInsert", map[string]any{"users": users})
}

func (m *UserMapperImpl) BatchUpdate(request *BatchUpdateRequest) (int64, error) {
	return m.sqlSession.Update("UserMapper.BatchUpdate", request)
}

func (m *UserMapperImpl) BatchDelete(ids []int64) (int64, error) {
	return m.sqlSession.Update("UserMapper.BatchDelete", map[string]any{"ids": ids})
}

func (m *UserMapperImpl) BatchUpdateStatus(ids []int64, status string) (int64, error) {
	return m.sqlSession.Update("UserMapper.BatchUpdateStatus", map[string]any{
		"ids":    ids,
		"status": status,
	})
}

// ========== 聚合查询实现 ==========

func (m *UserMapperImpl) SelectStats() (*UserStats, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.SelectStats", nil)
	if err != nil {
		return nil, err
	}
	if stats, ok := result.(*UserStats); ok {
		return stats, nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SelectByStatus() ([]*AggregationResult, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectByStatus", nil)
	if err != nil {
		return nil, err
	}
	
	aggregations := make([]*AggregationResult, len(results))
	for i, result := range results {
		if agg, ok := result.(*AggregationResult); ok {
			aggregations[i] = agg
		}
	}
	return aggregations, nil
}

func (m *UserMapperImpl) SelectByAgeGroup() ([]*AggregationResult, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectByAgeGroup", nil)
	if err != nil {
		return nil, err
	}
	
	aggregations := make([]*AggregationResult, len(results))
	for i, result := range results {
		if agg, ok := result.(*AggregationResult); ok {
			aggregations[i] = agg
		}
	}
	return aggregations, nil
}

func (m *UserMapperImpl) SelectActiveUsersInPeriod(startTime, endTime time.Time) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectActiveUsersInPeriod", map[string]any{
		"startTime": startTime,
		"endTime":   endTime,
	})
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

// ========== 复杂查询实现 ==========

func (m *UserMapperImpl) SelectWithProfile(id int64) (*ComplexQueryResult, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.SelectWithProfile", id)
	if err != nil {
		return nil, err
	}
	if complex, ok := result.(*ComplexQueryResult); ok {
		return complex, nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SelectWithRoles(id int64) (*ComplexQueryResult, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.SelectWithRoles", id)
	if err != nil {
		return nil, err
	}
	if complex, ok := result.(*ComplexQueryResult); ok {
		return complex, nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SelectWithArticles(userId int64, limit int) (*ComplexQueryResult, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.SelectWithArticles", map[string]any{
		"userId": userId,
		"limit":  limit,
	})
	if err != nil {
		return nil, err
	}
	if complex, ok := result.(*ComplexQueryResult); ok {
		return complex, nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SearchUsers(keyword string, limit int) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SearchUsers", map[string]any{
		"keyword": keyword,
		"limit":   limit,
	})
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

func (m *UserMapperImpl) SelectSimilarUsers(userId int64, limit int) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectSimilarUsers", map[string]any{
		"userId": userId,
		"limit":  limit,
	})
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

// ========== 特殊查询实现 ==========

func (m *UserMapperImpl) SelectRandomUsers(limit int) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectRandomUsers", map[string]any{
		"limit": limit,
	})
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

func (m *UserMapperImpl) SelectTopActiveUsers(limit int) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectTopActiveUsers", map[string]any{
		"limit": limit,
	})
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

func (m *UserMapperImpl) SelectUsersWithoutProfile() ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectUsersWithoutProfile", nil)
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

func (m *UserMapperImpl) SelectRecentRegistrations(days int, limit int) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectRecentRegistrations", map[string]any{
		"days":  days,
		"limit": limit,
	})
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

// ========== 存储过程和函数实现 ==========

func (m *UserMapperImpl) CallUserStatsProcedure(startDate, endDate time.Time) (*UserStats, error) {
	result, err := m.sqlSession.SelectOne("UserMapper.CallUserStatsProcedure", map[string]any{
		"startDate": startDate,
		"endDate":   endDate,
	})
	if err != nil {
		return nil, err
	}
	if stats, ok := result.(*UserStats); ok {
		return stats, nil
	}
	return nil, nil
}

func (m *UserMapperImpl) SelectUserByCustomFunction(param string) ([]*User, error) {
	results, err := m.sqlSession.SelectList("UserMapper.SelectUserByCustomFunction", map[string]any{
		"param": param,
	})
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

// RegisterUserMapper 注册用户映射器
func RegisterUserMapper(session session.SqlSession) (UserMapper, error) {
	// 在实际应用中，这里会注册所有的SQL映射
	return NewUserMapper(session), nil
}

// GetUserMapperType 获取用户映射器类型
func GetUserMapperType() reflect.Type {
	return reflect.TypeOf((*UserMapper)(nil)).Elem()
}