// Package main 集成测试
//
// 测试MyBatis框架与其他组件的集成功能
package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mybatis"
	"github.com/zsy619/yyhertz/framework/mybatis/session"
)

// IntegrationTestSuite 集成测试套件
type IntegrationTestSuite struct {
	suite.Suite
	dbSetup    *DatabaseSetup
	mybatis    *mybatis.MyBatis
	userMapper UserMapper
}

// SetupSuite 设置测试套件
func (suite *IntegrationTestSuite) SetupSuite() {
	// 初始化数据库设置
	suite.dbSetup = NewDatabaseSetup(DefaultDatabaseConfig())

	// 设置完整测试环境
	err := suite.dbSetup.SetupCompleteTestEnvironment()
	suite.Require().NoError(err)

	// 创建MyBatis实例
	mybatisConfig := mybatis.NewConfiguration()
	// 使用默认数据库配置
	dbConfig := DefaultDatabaseConfig()
	// 创建正确的数据库配置
	dbCfg := &config.DatabaseConfig{}
	dbCfg.Primary.Driver = "sqlite"
	dbCfg.Primary.Database = dbConfig.Path

	// 设置MyBatis配置
	dbCfg.MyBatis.MapperLocations = "mappers/*.xml"
	dbCfg.GORM.Enable = true

	mybatisConfig.SetDatabaseConfig(dbCfg)
	mybatisConfig.CacheEnabled = true
	mybatisConfig.LazyLoadingEnabled = true
	mybatisConfig.MapUnderscoreToCamelCase = true

	suite.mybatis, err = mybatis.NewMyBatis(mybatisConfig)
	suite.Require().NoError(err)

	// 创建映射器
	session := suite.mybatis.OpenSession()
	suite.userMapper = NewUserMapper(session)
}

// TearDownSuite 清理测试套件
func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.dbSetup != nil {
		suite.dbSetup.TeardownCompleteTestEnvironment()
	}
}

// TestMyBatisWithGORM 测试MyBatis与GORM的集成
func (suite *IntegrationTestSuite) TestMyBatisWithGORM() {
	t := suite.T()

	t.Run("MyBatis和GORM共享数据库连接", func(t *testing.T) {
		// 使用MyBatis插入用户
		user := &User{
			Name:   "MyBatis用户",
			Email:  "mybatis@example.com",
			Age:    25,
			Status: "active",
		}

		id, err := suite.userMapper.Insert(user)
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))

		// 使用GORM验证数据
		var gormUser User
		err = suite.dbSetup.DB.Where("email = ?", "mybatis@example.com").First(&gormUser).Error
		assert.NoError(t, err)
		assert.Equal(t, "MyBatis用户", gormUser.Name)
		assert.Equal(t, "mybatis@example.com", gormUser.Email)

		// 使用GORM更新数据
		err = suite.dbSetup.DB.Model(&gormUser).Update("name", "GORM更新用户").Error
		assert.NoError(t, err)

		// 使用MyBatis验证更新
		updatedUser, err := suite.userMapper.SelectByEmail("mybatis@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "GORM更新用户", updatedUser.Name)

		fmt.Printf("MyBatis与GORM集成测试成功: 用户ID=%d\n", id)
	})
}

// TestTransactionIntegration 测试事务集成
func (suite *IntegrationTestSuite) TestTransactionIntegration() {
	t := suite.T()

	t.Run("跨多个映射器的事务", func(t *testing.T) {
		// 使用事务执行多个操作
		err := suite.mybatis.ExecuteWithTransaction(func(session session.SqlSession) error {
			userMapper := NewUserMapper(session)

			// 插入第一个用户
			user1 := &User{
				Name:   "事务用户1",
				Email:  "tx1@example.com",
				Age:    25,
				Status: "active",
			}

			id1, err := userMapper.Insert(user1)
			if err != nil {
				return err
			}

			// 插入第二个用户
			user2 := &User{
				Name:   "事务用户2",
				Email:  "tx2@example.com",
				Age:    26,
				Status: "active",
			}

			id2, err := userMapper.Insert(user2)
			if err != nil {
				return err
			}

			// 更新第一个用户
			user1.ID = id1
			user1.Age = 30
			_, err = userMapper.Update(user1)
			if err != nil {
				return err
			}

			fmt.Printf("事务中创建用户: ID1=%d, ID2=%d\n", id1, id2)
			return nil
		})

		assert.NoError(t, err)

		// 验证事务提交成功
		user1, err := suite.userMapper.SelectByEmail("tx1@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user1)
		assert.Equal(t, 30, user1.Age) // 验证更新生效

		user2, err := suite.userMapper.SelectByEmail("tx2@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user2)
		assert.Equal(t, "事务用户2", user2.Name)
	})

	t.Run("事务回滚测试", func(t *testing.T) {
		originalCount, err := suite.userMapper.SelectCount(&UserQuery{})
		require.NoError(t, err)

		// 执行会失败的事务
		err = suite.mybatis.ExecuteWithTransaction(func(session session.SqlSession) error {
			userMapper := NewUserMapper(session)

			// 插入用户
			user := &User{
				Name:   "回滚测试用户",
				Email:  "rollback@example.com",
				Age:    25,
				Status: "active",
			}

			_, err := userMapper.Insert(user)
			if err != nil {
				return err
			}

			// 故意返回错误触发回滚
			return fmt.Errorf("模拟业务错误")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "模拟业务错误")

		// 验证数据已回滚
		newCount, err := suite.userMapper.SelectCount(&UserQuery{})
		assert.NoError(t, err)
		assert.Equal(t, originalCount, newCount)

		// 验证用户不存在
		user, err := suite.userMapper.SelectByEmail("rollback@example.com")
		assert.NoError(t, err)
		assert.Nil(t, user)

		fmt.Println("事务回滚测试成功")
	})
}

// TestCacheIntegration 测试缓存集成
func (suite *IntegrationTestSuite) TestCacheIntegration() {
	t := suite.T()

	t.Run("多级缓存测试", func(t *testing.T) {
		// 第一次查询（从数据库）
		start1 := time.Now()
		user1, err := suite.userMapper.SelectById(1)
		duration1 := time.Since(start1)
		assert.NoError(t, err)
		assert.NotNil(t, user1)

		// 第二次查询（从缓存）
		start2 := time.Now()
		user2, err := suite.userMapper.SelectById(1)
		duration2 := time.Since(start2)
		assert.NoError(t, err)
		assert.NotNil(t, user2)

		// 验证数据一致性
		assert.Equal(t, user1.ID, user2.ID)
		assert.Equal(t, user1.Name, user2.Name)
		assert.Equal(t, user1.Email, user2.Email)

		fmt.Printf("缓存性能测试: 第一次=%v, 第二次=%v\n", duration1, duration2)
	})

	t.Run("缓存失效测试", func(t *testing.T) {
		// 查询用户建立缓存
		user, err := suite.userMapper.SelectById(2)
		require.NoError(t, err)
		require.NotNil(t, user)

		originalName := user.Name

		// 更新用户
		user.Name = originalName + "_缓存更新"
		_, err = suite.userMapper.Update(user)
		assert.NoError(t, err)

		// 再次查询，应该获取到更新后的数据
		updatedUser, err := suite.userMapper.SelectById(2)
		assert.NoError(t, err)
		assert.Equal(t, originalName+"_缓存更新", updatedUser.Name)

		// 恢复原数据
		user.Name = originalName
		suite.userMapper.Update(user)

		fmt.Println("缓存失效测试成功")
	})
}

// TestBatchIntegration 测试批量操作集成
func (suite *IntegrationTestSuite) TestBatchIntegration() {
	t := suite.T()

	t.Run("大批量数据操作", func(t *testing.T) {
		// 准备大量测试数据
		batchSize := 100
		users := make([]*User, batchSize)

		for i := 0; i < batchSize; i++ {
			users[i] = &User{
				Name:   fmt.Sprintf("批量用户%d", i+1),
				Email:  fmt.Sprintf("batch%d@example.com", i+1),
				Age:    20 + (i % 50),
				Status: []string{"active", "inactive"}[i%2],
			}
		}

		// 批量插入
		start := time.Now()
		affected, err := suite.userMapper.BatchInsert(users)
		insertDuration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, int64(batchSize), affected)

		// 验证插入成功
		query := &UserQuery{
			Keyword:  "批量用户",
			PageSize: batchSize + 10,
		}

		insertedUsers, err := suite.userMapper.SelectList(query)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(insertedUsers), batchSize)

		fmt.Printf("批量插入 %d 个用户，耗时: %v\n", batchSize, insertDuration)

		// 收集ID用于批量删除
		var ids []int64
		for _, user := range insertedUsers {
			if user.Name[:4] == "批量用户" {
				ids = append(ids, user.ID)
			}
		}

		// 批量删除
		start = time.Now()
		deleted, err := suite.userMapper.BatchDelete(ids)
		deleteDuration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, int64(len(ids)), deleted)

		fmt.Printf("批量删除 %d 个用户，耗时: %v\n", len(ids), deleteDuration)
	})
}

// TestConcurrencyIntegration 测试并发集成
func (suite *IntegrationTestSuite) TestConcurrencyIntegration() {
	t := suite.T()

	t.Run("并发读写测试", func(t *testing.T) {
		concurrency := 10
		iterations := 20

		// 创建测试用户
		testUser := &User{
			Name:   "并发测试用户",
			Email:  "concurrent@example.com",
			Age:    25,
			Status: "active",
		}

		id, err := suite.userMapper.Insert(testUser)
		require.NoError(t, err)

		// 并发读取
		readChan := make(chan error, concurrency*iterations)
		var wg sync.WaitGroup
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					user, err := suite.userMapper.SelectById(id)
					if err != nil {
						readChan <- err
						return
					}
					if user == nil {
						readChan <- fmt.Errorf("user not found")
						return
					}
					readChan <- nil
				}
			}()
		}

		// 并发写入
		writeChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				user := &User{
					Name:   fmt.Sprintf("并发写入用户%d", index),
					Email:  fmt.Sprintf("write%d@example.com", index),
					Age:    25 + index,
					Status: "active",
				}

				_, err := suite.userMapper.Insert(user)
				writeChan <- err
			}(i)
		}

		// 等待所有读取操作完成
		wg.Wait()
		close(readChan)

		// 检查结果
		for err := range readChan {
			assert.NoError(t, err)
		}

		for i := 0; i < concurrency; i++ {
			err := <-writeChan
			assert.NoError(t, err)
		}

		fmt.Printf("并发测试完成: %d个goroutine, 每个%d次读取\n", concurrency, iterations)
	})
}

// TestRealWorldScenario 测试真实场景
func (suite *IntegrationTestSuite) TestRealWorldScenario() {
	t := suite.T()

	t.Run("用户注册完整流程", func(t *testing.T) {
		ctx := context.Background()

		// 模拟用户注册流程
		err := suite.mybatis.ExecuteWithTransaction(func(session session.SqlSession) error {
			userMapper := NewUserMapper(session)

			// 1. 检查邮箱是否已存在
			existingUser, err := userMapper.SelectByEmail("newuser@example.com")
			if err != nil {
				return err
			}
			if existingUser != nil {
				return fmt.Errorf("邮箱已存在")
			}

			// 2. 创建用户
			newUser := &User{
				Name:   "新注册用户",
				Email:  "newuser@example.com",
				Age:    25,
				Status: "active",
				Phone:  "13900000000",
			}

			userID, err := userMapper.Insert(newUser)
			if err != nil {
				return err
			}

			// 3. 创建用户档案
			profile := &UserProfile{
				UserID:     userID,
				Bio:        "这是一个新用户",
				Location:   "北京",
				Company:    "测试公司",
				Occupation: "工程师",
			}

			// 使用GORM插入档案（模拟其他服务）
			db := session.GetConnection()
			if err := db.WithContext(ctx).Create(profile).Error; err != nil {
				return err
			}

			// 4. 分配默认角色
			role := &UserRole{
				UserID:      userID,
				RoleName:    "user",
				Permissions: `["read", "write"]`,
			}

			if err := db.WithContext(ctx).Create(role).Error; err != nil {
				return err
			}

			fmt.Printf("用户注册成功: ID=%d\n", userID)
			return nil
		})

		assert.NoError(t, err)

		// 验证注册结果
		user, err := suite.userMapper.SelectByEmail("newuser@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "新注册用户", user.Name)

		// 验证关联数据
		var profile UserProfile
		err = suite.dbSetup.DB.Where("user_id = ?", user.ID).First(&profile).Error
		assert.NoError(t, err)
		assert.Equal(t, "这是一个新用户", profile.Bio)

		var role UserRole
		err = suite.dbSetup.DB.Where("user_id = ?", user.ID).First(&role).Error
		assert.NoError(t, err)
		assert.Equal(t, "user", role.RoleName)

		fmt.Println("用户注册完整流程验证成功")
	})

	t.Run("数据分析查询场景", func(t *testing.T) {
		// 模拟数据分析场景

		// 1. 用户统计
		stats, err := suite.userMapper.SelectStats()
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// 2. 状态分布
		statusStats, err := suite.userMapper.SelectByStatus()
		assert.NoError(t, err)
		assert.NotEmpty(t, statusStats)

		// 3. 年龄分布
		ageStats, err := suite.userMapper.SelectByAgeGroup()
		assert.NoError(t, err)
		assert.NotEmpty(t, ageStats)

		// 4. 最近注册用户
		recentUsers, err := suite.userMapper.SelectRecentRegistrations(30, 10)
		assert.NoError(t, err)

		// 5. 活跃用户
		activeUsers, err := suite.userMapper.SelectTopActiveUsers(5)
		assert.NoError(t, err)

		fmt.Printf("数据分析结果:\n")
		fmt.Printf("  总用户: %d, 活跃用户: %d\n", stats.TotalUsers, stats.ActiveUsers)
		fmt.Printf("  状态分布: %d种\n", len(statusStats))
		fmt.Printf("  年龄分布: %d组\n", len(ageStats))
		fmt.Printf("  最近注册: %d人\n", len(recentUsers))
		fmt.Printf("  活跃用户: %d人\n", len(activeUsers))
	})
}

// TestErrorRecovery 测试错误恢复
func (suite *IntegrationTestSuite) TestErrorRecovery() {
	t := suite.T()

	t.Run("连接中断恢复", func(t *testing.T) {
		// 检查连接状态
		err := suite.dbSetup.CheckDatabaseConnection()
		assert.NoError(t, err)

		// 执行正常查询
		user, err := suite.userMapper.SelectById(1)
		assert.NoError(t, err)
		assert.NotNil(t, user)

		// 模拟连接恢复后的操作
		newUser := &User{
			Name:   "恢复测试用户",
			Email:  "recovery@example.com",
			Age:    25,
			Status: "active",
		}

		id, err := suite.userMapper.Insert(newUser)
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))

		fmt.Println("连接恢复测试成功")
	})
}

// TestIntegrationSuite 运行集成测试套件
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// TestCompleteWorkflow 测试完整工作流
func TestCompleteWorkflow(t *testing.T) {
	// 设置测试环境
	dbSetup := NewDatabaseSetup(DefaultDatabaseConfig())
	err := dbSetup.SetupCompleteTestEnvironment()
	require.NoError(t, err)
	defer dbSetup.TeardownCompleteTestEnvironment()

	// 创建MyBatis实例
	mybatisConfig := mybatis.NewConfiguration()
	dbConfig := DefaultDatabaseConfig()
	// 创建正确的数据库配置
	dbCfg := &config.DatabaseConfig{}
	dbCfg.Primary.Driver = "sqlite"
	dbCfg.Primary.Database = dbConfig.Path

	// 设置MyBatis配置
	dbCfg.MyBatis.MapperLocations = "mappers/*.xml"
	dbCfg.GORM.Enable = true
	dbCfg.Primary.LogLevel = "info"

	mybatisConfig.SetDatabaseConfig(dbCfg)

	mb, err := mybatis.NewMyBatis(mybatisConfig)
	require.NoError(t, err)

	t.Run("端到端业务流程", func(t *testing.T) {
		// 1. 用户管理流程
		session := mb.OpenSession()
		defer session.Close()

		userMapper := NewUserMapper(session)

		// 创建用户
		user := &User{
			Name:   "端到端测试用户",
			Email:  "e2e@example.com",
			Age:    28,
			Status: "active",
		}

		userID, err := userMapper.Insert(user)
		assert.NoError(t, err)

		// 查询用户
		retrievedUser, err := userMapper.SelectById(userID)
		assert.NoError(t, err)
		assert.Equal(t, user.Name, retrievedUser.Name)

		// 更新用户
		retrievedUser.Age = 29
		affected, err := userMapper.Update(retrievedUser)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证更新
		updatedUser, err := userMapper.SelectById(userID)
		assert.NoError(t, err)
		assert.Equal(t, 29, updatedUser.Age)

		// 动态查询
		query := &UserQuery{
			Name:     "端到端",
			Status:   "active",
			AgeMin:   25,
			AgeMax:   35,
			PageSize: 10,
		}

		users, err := userMapper.SelectList(query)
		assert.NoError(t, err)
		assert.NotEmpty(t, users)

		// 分页查询
		pageResult, err := userMapper.SelectPage(query)
		assert.NoError(t, err)
		assert.NotNil(t, pageResult)
		assert.Greater(t, pageResult.Total, int64(0))

		// 聚合查询
		stats, err := userMapper.SelectStats()
		assert.NoError(t, err)
		assert.Greater(t, stats.TotalUsers, int64(0))

		// 软删除
		affected, err = userMapper.Delete(userID)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证删除
		deletedUser, err := userMapper.SelectById(userID)
		assert.NoError(t, err)
		assert.Nil(t, deletedUser)

		fmt.Printf("端到端测试完成: 用户ID=%d\n", userID)
	})

	t.Run("性能压力测试", func(t *testing.T) {
		session := mb.OpenSession()
		defer session.Close()

		userMapper := NewUserMapper(session)

		// 压力测试参数
		testCount := 50

		start := time.Now()

		// 并发执行多种操作
		for i := 0; i < testCount; i++ {
			// 插入
			user := &User{
				Name:   fmt.Sprintf("压力测试用户%d", i),
				Email:  fmt.Sprintf("stress%d@example.com", i),
				Age:    20 + (i % 30),
				Status: "active",
			}

			userID, err := userMapper.Insert(user)
			assert.NoError(t, err)

			// 查询
			_, err = userMapper.SelectById(userID)
			assert.NoError(t, err)

			// 更新
			user.Age = user.Age + 1
			user.ID = userID
			_, err = userMapper.Update(user)
			assert.NoError(t, err)
		}

		duration := time.Since(start)

		// 批量查询
		query := &UserQuery{
			Keyword:  "压力测试",
			PageSize: testCount + 10,
		}

		stressUsers, err := userMapper.SelectList(query)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(stressUsers), testCount)

		// 批量删除
		var ids []int64
		for _, user := range stressUsers {
			if len(user.Name) >= 6 && user.Name[:6] == "压力测试用户" {
				ids = append(ids, user.ID)
			}
		}

		_, err = userMapper.BatchDelete(ids)
		assert.NoError(t, err)

		fmt.Printf("压力测试完成: %d次操作，耗时: %v，平均: %v/次\n",
			testCount*3, duration, duration/time.Duration(testCount*3))
	})
}

// BenchmarkIntegration 集成性能基准测试
func BenchmarkIntegration(b *testing.B) {
	// 设置测试环境
	dbSetup := NewDatabaseSetup(DefaultDatabaseConfig())
	err := dbSetup.SetupCompleteTestEnvironment()
	if err != nil {
		b.Fatalf("Failed to setup test environment: %v", err)
	}
	defer dbSetup.TeardownCompleteTestEnvironment()

	// 创建MyBatis实例
	mybatisConfig := mybatis.NewConfiguration()
	dbConfig := DefaultDatabaseConfig()
	// 创建正确的数据库配置
	dbCfg := &config.DatabaseConfig{}
	dbCfg.Primary.Driver = "sqlite"
	dbCfg.Primary.Database = dbConfig.Path

	// 设置MyBatis配置
	dbCfg.MyBatis.MapperLocations = "mappers/*.xml"
	dbCfg.GORM.Enable = true
	dbCfg.Primary.LogLevel = "info"

	mybatisConfig.SetDatabaseConfig(dbCfg)

	mb, err := mybatis.NewMyBatis(mybatisConfig)
	if err != nil {
		b.Fatalf("Failed to create MyBatis instance: %v", err)
	}

	session := mb.OpenSession()
	defer session.Close()

	userMapper := NewUserMapper(session)

	b.Run("CompleteUserWorkflow", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// 创建用户
			user := &User{
				Name:   fmt.Sprintf("BenchUser%d", i),
				Email:  fmt.Sprintf("bench%d@example.com", i),
				Age:    25,
				Status: "active",
			}

			userID, err := userMapper.Insert(user)
			if err != nil {
				b.Fatal(err)
			}

			// 查询用户
			_, err = userMapper.SelectById(userID)
			if err != nil {
				b.Fatal(err)
			}

			// 更新用户
			user.ID = userID
			user.Age = 26
			_, err = userMapper.Update(user)
			if err != nil {
				b.Fatal(err)
			}

			// 删除用户
			_, err = userMapper.Delete(userID)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
