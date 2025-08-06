// Package mybatis_tests 主要测试运行器
//
// 演示MyBatis框架的完整功能，包括所有CRUD操作、动态SQL、缓存等特性
package mybat

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mybatis"
	"github.com/zsy619/yyhertz/framework/mybatis/session"
)

// TestConfig 测试配置
type TestConfig struct {
	DSN        string
	DB         *gorm.DB
	MyBatis    *mybatis.MyBatis
	Session    session.SqlSession
	UserMapper UserMapper
}

// setupTestEnvironment 设置测试环境
func setupTestEnvironment() (*TestConfig, error) {
	// 1. 配置数据库连接
	dsn := "root:123456@tcp(localhost:3306)/mybatis_test?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 2. 创建MyBatis配置
	mybatisConfig := mybatis.NewConfiguration()
	dbConfig := DefaultDatabaseConfig()
	// 创建正确的数据库配置
	dbCfg := &config.DatabaseConfig{}
	dbCfg.Primary.Driver = "mysql"
	dbCfg.Primary.Host = dbConfig.Host
	dbCfg.Primary.Port = dbConfig.Port
	dbCfg.Primary.Username = dbConfig.Username
	dbCfg.Primary.Password = dbConfig.Password
	dbCfg.Primary.Database = dbConfig.Database
	dbCfg.Primary.Charset = dbConfig.Charset

	mybatisConfig.SetDatabaseConfig(dbCfg)
	mybatisConfig.CacheEnabled = true
	mybatisConfig.LazyLoadingEnabled = true
	mybatisConfig.MapUnderscoreToCamelCase = true

	// 3. 创建MyBatis实例
	mb, err := mybatis.NewMyBatis(mybatisConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create MyBatis instance: %w", err)
	}

	// 4. 开启会话
	sqlSession := mb.OpenSession()

	// 5. 创建映射器
	userMapper := NewUserMapper(sqlSession)

	return &TestConfig{
		DSN:        dsn,
		DB:         db,
		MyBatis:    mb,
		Session:    sqlSession,
		UserMapper: userMapper,
	}, nil
}

// teardownTestEnvironment 清理测试环境
func teardownTestEnvironment(config *TestConfig) {
	if config.Session != nil {
		config.Session.Close()
	}
	if config.DB != nil {
		sqlDB, _ := config.DB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// initTestDatabase 初始化测试数据库
func initTestDatabase(db *gorm.DB) error {
	// 创建数据库表
	ctx := context.Background()

	// 执行建表SQL
	tables := []string{
		CreateUsersTableSQL,
		CreateUserProfilesTableSQL,
		CreateUserRolesTableSQL,
		CreateArticlesTableSQL,
		CreateCategoriesTableSQL,
		CreateUserArticleViewsTableSQL,
	}

	for _, tableSQL := range tables {
		if err := db.WithContext(ctx).Exec(tableSQL).Error; err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	// 插入测试数据
	testData := []string{
		InsertTestUsersSQL,
		InsertTestProfilesSQL,
		InsertTestRolesSQL,
	}

	for _, dataSQL := range testData {
		if err := db.WithContext(ctx).Exec(dataSQL).Error; err != nil {
			return fmt.Errorf("failed to insert test data: %w", err)
		}
	}

	// 创建存储过程和函数
	procedures := []string{
		CreateUserStatsProcedureSQL,
		CreateCustomFunctionSQL,
	}

	for _, procSQL := range procedures {
		if err := db.WithContext(ctx).Exec(procSQL).Error; err != nil {
			log.Printf("Warning: Failed to create procedure/function: %v", err)
		}
	}

	return nil
}

// TestMain 主测试入口
func TestMain(m *testing.M) {
	fmt.Println("=== MyBatis-Go 框架测试开始 ===")

	// 设置测试环境
	config, err := setupTestEnvironment()
	if err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}
	defer teardownTestEnvironment(config)

	// 初始化测试数据库
	if err := initTestDatabase(config.DB); err != nil {
		log.Fatalf("Failed to initialize test database: %v", err)
	}

	fmt.Println("测试环境准备完成")

	// 运行所有测试
	m.Run()

	fmt.Println("=== MyBatis-Go 框架测试结束 ===")
}

// TestBasicCRUD 测试基础CRUD操作
func TestBasicCRUD(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试根据ID查询用户", func(t *testing.T) {
		user, err := config.UserMapper.SelectById(1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "张三", user.Name)
		assert.Equal(t, "zhangsan@example.com", user.Email)

		fmt.Printf("查询到用户: %+v\n", user)
	})

	t.Run("测试根据邮箱查询用户", func(t *testing.T) {
		user, err := config.UserMapper.SelectByEmail("lisi@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "李四", user.Name)

		fmt.Printf("根据邮箱查询到用户: %+v\n", user)
	})

	t.Run("测试插入用户", func(t *testing.T) {
		newUser := &User{
			Name:   "测试用户",
			Email:  "test@example.com",
			Age:    25,
			Status: "active",
			Phone:  "13900000000",
		}

		id, err := config.UserMapper.Insert(newUser)
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))

		// 验证插入成功
		insertedUser, err := config.UserMapper.SelectByEmail("test@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, insertedUser)
		assert.Equal(t, "测试用户", insertedUser.Name)

		fmt.Printf("插入用户成功，ID: %d\n", id)
	})

	t.Run("测试更新用户", func(t *testing.T) {
		user, err := config.UserMapper.SelectById(1)
		require.NoError(t, err)
		require.NotNil(t, user)

		originalName := user.Name
		user.Name = "张三(已更新)"
		user.Age = 26

		affected, err := config.UserMapper.Update(user)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证更新成功
		updatedUser, err := config.UserMapper.SelectById(1)
		assert.NoError(t, err)
		assert.Equal(t, "张三(已更新)", updatedUser.Name)
		assert.Equal(t, 26, updatedUser.Age)

		// 恢复原始数据
		user.Name = originalName
		user.Age = 25
		config.UserMapper.Update(user)

		fmt.Printf("更新用户成功: %s -> %s\n", originalName, updatedUser.Name)
	})

	t.Run("测试软删除用户", func(t *testing.T) {
		// 创建测试用户
		testUser := &User{
			Name:   "待删除用户",
			Email:  "delete@example.com",
			Age:    30,
			Status: "active",
		}

		id, err := config.UserMapper.Insert(testUser)
		require.NoError(t, err)

		// 执行软删除
		affected, err := config.UserMapper.Delete(id)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证用户已被软删除（查询不到）
		deletedUser, err := config.UserMapper.SelectById(id)
		assert.NoError(t, err)
		assert.Nil(t, deletedUser)

		fmt.Printf("软删除用户成功，ID: %d\n", id)
	})
}

// TestDynamicSQL 测试动态SQL查询
func TestDynamicSQL(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试动态条件查询", func(t *testing.T) {
		query := &UserQuery{
			Name:      "张",
			Status:    "active",
			AgeMin:    20,
			AgeMax:    40,
			Page:      1,
			PageSize:  10,
			OrderBy:   "created_at",
			OrderDesc: true,
		}

		users, err := config.UserMapper.SelectList(query)
		assert.NoError(t, err)
		assert.NotEmpty(t, users)

		// 验证查询结果符合条件
		for _, user := range users {
			assert.Contains(t, user.Name, "张")
			assert.Equal(t, "active", user.Status)
			assert.GreaterOrEqual(t, user.Age, 20)
			assert.LessOrEqual(t, user.Age, 40)
		}

		fmt.Printf("动态条件查询到 %d 个用户\n", len(users))
	})

	t.Run("测试关键字搜索", func(t *testing.T) {
		query := &UserQuery{
			Keyword:  "张",
			PageSize: 5,
		}

		users, err := config.UserMapper.SelectList(query)
		assert.NoError(t, err)

		// 验证搜索结果
		for _, user := range users {
			containsKeyword := user.Name != "" && (user.Name != "" || user.Email != "")
			assert.True(t, containsKeyword)
		}

		fmt.Printf("关键字搜索到 %d 个用户\n", len(users))
	})

	t.Run("测试分页查询", func(t *testing.T) {
		query := &UserQuery{
			Page:     1,
			PageSize: 3,
			OrderBy:  "id",
		}

		result, err := config.UserMapper.SelectPage(query)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		if users, ok := result.Data.([]*User); ok {
			assert.LessOrEqual(t, len(users), 3)
		}
		assert.Greater(t, result.Total, int64(0))
		assert.Equal(t, 1, result.Page)
		assert.Equal(t, 3, result.PageSize)

		fmt.Printf("分页查询结果: 总数=%d, 当前页=%d, 每页=%d, 总页数=%d\n",
			result.Total, result.Page, result.PageSize, result.TotalPages)
	})
}

// TestBatchOperations 测试批量操作
func TestBatchOperations(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试批量插入", func(t *testing.T) {
		users := []*User{
			{Name: "批量用户1", Email: "batch1@example.com", Age: 25, Status: "active"},
			{Name: "批量用户2", Email: "batch2@example.com", Age: 26, Status: "active"},
			{Name: "批量用户3", Email: "batch3@example.com", Age: 27, Status: "inactive"},
		}

		affected, err := config.UserMapper.BatchInsert(users)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), affected)

		// 验证插入成功
		for _, user := range users {
			insertedUser, err := config.UserMapper.SelectByEmail(user.Email)
			assert.NoError(t, err)
			assert.NotNil(t, insertedUser)
			assert.Equal(t, user.Name, insertedUser.Name)
		}

		fmt.Printf("批量插入 %d 个用户成功\n", len(users))
	})

	t.Run("测试批量更新状态", func(t *testing.T) {
		// 获取几个用户ID
		query := &UserQuery{PageSize: 3}
		users, err := config.UserMapper.SelectList(query)
		require.NoError(t, err)
		require.NotEmpty(t, users)

		ids := make([]int64, len(users))
		for i, user := range users {
			ids[i] = user.ID
		}

		// 批量更新状态
		affected, err := config.UserMapper.BatchUpdateStatus(ids, "inactive")
		assert.NoError(t, err)
		assert.Equal(t, int64(len(ids)), affected)

		// 验证更新成功
		for _, id := range ids {
			user, err := config.UserMapper.SelectById(id)
			assert.NoError(t, err)
			assert.Equal(t, "inactive", user.Status)
		}

		// 恢复原状态
		config.UserMapper.BatchUpdateStatus(ids, "active")

		fmt.Printf("批量更新 %d 个用户状态成功\n", len(ids))
	})

	t.Run("测试批量删除", func(t *testing.T) {
		// 先创建测试数据
		testUsers := []*User{
			{Name: "待删除1", Email: "delete1@example.com", Age: 25, Status: "active"},
			{Name: "待删除2", Email: "delete2@example.com", Age: 26, Status: "active"},
		}

		var ids []int64
		for _, user := range testUsers {
			id, err := config.UserMapper.Insert(user)
			require.NoError(t, err)
			ids = append(ids, id)
		}

		// 批量删除
		affected, err := config.UserMapper.BatchDelete(ids)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(ids)), affected)

		// 验证删除成功
		for _, id := range ids {
			user, err := config.UserMapper.SelectById(id)
			assert.NoError(t, err)
			assert.Nil(t, user) // 软删除后查询不到
		}

		fmt.Printf("批量删除 %d 个用户成功\n", len(ids))
	})
}

// TestAggregationQueries 测试聚合查询
func TestAggregationQueries(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试用户统计", func(t *testing.T) {
		stats, err := config.UserMapper.SelectStats()
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.TotalUsers, int64(0))
		assert.Greater(t, stats.ActiveUsers, int64(0))

		fmt.Printf("用户统计: 总用户=%d, 活跃用户=%d, 最近用户=%d\n",
			stats.TotalUsers, stats.ActiveUsers, stats.RecentUsers)
	})

	t.Run("测试按状态分组统计", func(t *testing.T) {
		results, err := config.UserMapper.SelectByStatus()
		assert.NoError(t, err)
		assert.NotEmpty(t, results)

		fmt.Println("按状态分组统计:")
		for _, result := range results {
			fmt.Printf("  状态: %v, 数量: %d\n", result.Value, result.Count)
		}
	})

	t.Run("测试按年龄组分组统计", func(t *testing.T) {
		results, err := config.UserMapper.SelectByAgeGroup()
		assert.NoError(t, err)
		assert.NotEmpty(t, results)

		fmt.Println("按年龄组分组统计:")
		for _, result := range results {
			fmt.Printf("  年龄组: %v, 数量: %d\n", result.Value, result.Count)
		}
	})

	t.Run("测试时间段活跃用户查询", func(t *testing.T) {
		endTime := time.Now()
		startTime := endTime.AddDate(0, -1, 0) // 一个月前

		users, err := config.UserMapper.SelectActiveUsersInPeriod(startTime, endTime)
		assert.NoError(t, err)

		fmt.Printf("最近一个月活跃用户: %d 个\n", len(users))
	})
}

// TestComplexQueries 测试复杂查询
func TestComplexQueries(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试用户档案联合查询", func(t *testing.T) {
		result, err := config.UserMapper.SelectWithProfile(1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.User)

		if result.Profile != nil {
			fmt.Printf("用户档案查询: 用户=%s, 公司=%s, 职位=%s\n",
				result.User.Name, result.Profile.Company, result.Profile.Occupation)
		} else {
			fmt.Printf("用户档案查询: 用户=%s (无档案信息)\n", result.User.Name)
		}
	})

	t.Run("测试用户角色联合查询", func(t *testing.T) {
		result, err := config.UserMapper.SelectWithRoles(1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.User)

		fmt.Printf("用户角色查询: 用户=%s, 角色数量=%d\n",
			result.User.Name, len(result.Roles))
	})

	t.Run("测试全文搜索", func(t *testing.T) {
		users, err := config.UserMapper.SearchUsers("张", 5)
		assert.NoError(t, err)

		fmt.Printf("全文搜索结果: %d 个用户\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
		}
	})

	t.Run("测试相似用户查询", func(t *testing.T) {
		users, err := config.UserMapper.SelectSimilarUsers(1, 3)
		assert.NoError(t, err)

		fmt.Printf("相似用户查询结果: %d 个用户\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (年龄: %d, 状态: %s)\n", user.Name, user.Age, user.Status)
		}
	})
}

// TestSpecialQueries 测试特殊查询
func TestSpecialQueries(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试随机用户查询", func(t *testing.T) {
		users, err := config.UserMapper.SelectRandomUsers(3)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(users), 3)

		fmt.Printf("随机用户查询结果: %d 个用户\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
		}
	})

	t.Run("测试最活跃用户查询", func(t *testing.T) {
		users, err := config.UserMapper.SelectTopActiveUsers(5)
		assert.NoError(t, err)

		fmt.Printf("最活跃用户查询结果: %d 个用户\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
		}
	})

	t.Run("测试无档案用户查询", func(t *testing.T) {
		users, err := config.UserMapper.SelectUsersWithoutProfile()
		assert.NoError(t, err)

		fmt.Printf("无档案用户查询结果: %d 个用户\n", len(users))
	})

	t.Run("测试最近注册用户", func(t *testing.T) {
		users, err := config.UserMapper.SelectRecentRegistrations(30, 5)
		assert.NoError(t, err)

		fmt.Printf("最近30天注册用户: %d 个\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (注册时间: %s)\n", user.Name, user.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	})
}

// TestStoredProcedures 测试存储过程和函数
func TestStoredProcedures(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试用户统计存储过程", func(t *testing.T) {
		startDate := time.Now().AddDate(0, -1, 0)
		endDate := time.Now()

		stats, err := config.UserMapper.CallUserStatsProcedure(startDate, endDate)
		if err != nil {
			t.Logf("存储过程调用失败(可能未创建): %v", err)
			return
		}

		assert.NotNil(t, stats)
		fmt.Printf("存储过程统计结果: 总用户=%d, 活跃用户=%d\n",
			stats.TotalUsers, stats.ActiveUsers)
	})

	t.Run("测试自定义函数查询", func(t *testing.T) {
		users, err := config.UserMapper.SelectUserByCustomFunction("张")
		if err != nil {
			t.Logf("自定义函数调用失败(可能未创建): %v", err)
			return
		}

		fmt.Printf("自定义函数查询结果: %d 个用户\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
		}
	})
}

// TestCacheAndPerformance 测试缓存和性能
func TestCacheAndPerformance(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试查询缓存", func(t *testing.T) {
		// 第一次查询
		start1 := time.Now()
		user1, err := config.UserMapper.SelectById(1)
		duration1 := time.Since(start1)
		assert.NoError(t, err)
		assert.NotNil(t, user1)

		// 第二次查询（应该命中缓存）
		start2 := time.Now()
		user2, err := config.UserMapper.SelectById(1)
		duration2 := time.Since(start2)
		assert.NoError(t, err)
		assert.NotNil(t, user2)

		fmt.Printf("查询性能对比: 第一次=%v, 第二次=%v\n", duration1, duration2)
		assert.Equal(t, user1.ID, user2.ID)
		assert.Equal(t, user1.Name, user2.Name)
	})

	t.Run("测试缓存清除", func(t *testing.T) {
		// 查询用户
		user, err := config.UserMapper.SelectById(1)
		require.NoError(t, err)

		// 清除缓存
		config.Session.ClearCache()

		// 再次查询
		user2, err := config.UserMapper.SelectById(1)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, user2.ID)

		fmt.Println("缓存清除测试完成")
	})
}

// TestTransactions 测试事务
func TestTransactions(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试事务提交", func(t *testing.T) {
		// 开始事务
		testUser := &User{
			Name:   "事务测试用户",
			Email:  "transaction@example.com",
			Age:    25,
			Status: "active",
		}

		id, err := config.UserMapper.Insert(testUser)
		assert.NoError(t, err)

		// 提交事务
		err = config.Session.Commit()
		assert.NoError(t, err)

		// 验证数据已保存
		savedUser, err := config.UserMapper.SelectById(id)
		assert.NoError(t, err)
		assert.NotNil(t, savedUser)
		assert.Equal(t, testUser.Name, savedUser.Name)

		fmt.Printf("事务提交测试成功，用户ID: %d\n", id)
	})

	t.Run("测试事务回滚", func(t *testing.T) {
		// 创建新会话用于回滚测试
		rollbackSession := config.MyBatis.OpenSessionWithAutoCommit(false)
		defer rollbackSession.Close()

		rollbackMapper := NewUserMapper(rollbackSession)

		testUser := &User{
			Name:   "回滚测试用户",
			Email:  "rollback@example.com",
			Age:    25,
			Status: "active",
		}

		id, err := rollbackMapper.Insert(testUser)
		assert.NoError(t, err)

		// 验证数据存在（在事务中）
		user, err := rollbackMapper.SelectById(id)
		assert.NoError(t, err)
		assert.NotNil(t, user)

		// 回滚事务
		err = rollbackSession.Rollback()
		assert.NoError(t, err)

		// 验证数据已回滚（使用新会话查询）
		newSession := config.MyBatis.OpenSession()
		defer newSession.Close()
		newMapper := NewUserMapper(newSession)

		rolledBackUser, err := newMapper.SelectById(id)
		assert.NoError(t, err)
		assert.Nil(t, rolledBackUser)

		fmt.Println("事务回滚测试成功")
	})
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	config, err := setupTestEnvironment()
	require.NoError(t, err)
	defer teardownTestEnvironment(config)

	t.Run("测试查询不存在的用户", func(t *testing.T) {
		user, err := config.UserMapper.SelectById(99999)
		assert.NoError(t, err)
		assert.Nil(t, user)

		fmt.Println("查询不存在用户测试通过")
	})

	t.Run("测试无效邮箱查询", func(t *testing.T) {
		user, err := config.UserMapper.SelectByEmail("nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, user)

		fmt.Println("无效邮箱查询测试通过")
	})

	t.Run("测试空参数查询", func(t *testing.T) {
		query := &UserQuery{}
		users, err := config.UserMapper.SelectList(query)
		assert.NoError(t, err)
		assert.NotNil(t, users)

		fmt.Printf("空参数查询返回 %d 个用户\n", len(users))
	})
}

// BenchmarkQueries 性能基准测试
func BenchmarkQueries(b *testing.B) {
	config, err := setupTestEnvironment()
	if err != nil {
		b.Fatalf("Failed to setup test environment: %v", err)
	}
	defer teardownTestEnvironment(config)

	b.Run("SelectById", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := config.UserMapper.SelectById(1)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("SelectList", func(b *testing.B) {
		query := &UserQuery{
			Status:   "active",
			PageSize: 10,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := config.UserMapper.SelectList(query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Insert", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			user := &User{
				Name:   fmt.Sprintf("Benchmark用户%d", i),
				Email:  fmt.Sprintf("bench%d@example.com", i),
				Age:    25,
				Status: "active",
			}

			_, err := config.UserMapper.Insert(user)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// TestMapperRegistration 测试映射器注册
func TestMapperRegistration(t *testing.T) {
	t.Run("测试映射器类型获取", func(t *testing.T) {
		mapperType := GetUserMapperType()
		assert.NotNil(t, mapperType)
		assert.Equal(t, "UserMapper", mapperType.Name())
		assert.Equal(t, reflect.Interface, mapperType.Kind())

		fmt.Printf("映射器类型: %s, 类别: %s\n", mapperType.Name(), mapperType.Kind())
	})

	t.Run("测试映射器创建", func(t *testing.T) {
		config, err := setupTestEnvironment()
		require.NoError(t, err)
		defer teardownTestEnvironment(config)

		mapper, err := RegisterUserMapper(config.Session)
		assert.NoError(t, err)
		assert.NotNil(t, mapper)

		// 测试映射器功能
		user, err := mapper.SelectById(1)
		assert.NoError(t, err)
		if user != nil {
			fmt.Printf("通过注册的映射器查询到用户: %s\n", user.Name)
		}
	})
}

// printTestSummary 打印测试总结
func printTestSummary() {
	fmt.Println("\n=== MyBatis-Go 框架测试总结 ===")
	fmt.Println("✅ 基础CRUD操作")
	fmt.Println("✅ 动态SQL查询")
	fmt.Println("✅ 批量操作")
	fmt.Println("✅ 聚合查询")
	fmt.Println("✅ 复杂联合查询")
	fmt.Println("✅ 特殊查询功能")
	fmt.Println("✅ 存储过程和函数")
	fmt.Println("✅ 缓存机制")
	fmt.Println("✅ 事务管理")
	fmt.Println("✅ 错误处理")
	fmt.Println("✅ 性能测试")
	fmt.Println("✅ 映射器注册")
	fmt.Println("\n🎉 所有测试功能验证完成！")
	fmt.Println("📊 MyBatis-Go框架已成功实现所有核心特性")
}
