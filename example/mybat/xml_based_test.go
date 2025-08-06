// Package mybat 基于XML的MyBatis测试用例
//
// 演示如何使用XML配置文件进行MyBatis映射和查询
package mybat

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	frameworkConfig "github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mybatis"
	"github.com/zsy619/yyhertz/framework/mybatis/config"
)

// 简化的配置结构，用于测试
type TestDatabaseConfig struct {
	Primary TestPrimaryConfig `json:"primary"`
	GORM    TestGORMConfig    `json:"gorm"`
	MyBatis TestMyBatisConfig `json:"mybatis"`
}

type TestPrimaryConfig struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Charset  string `json:"charset"`
}

type TestGORMConfig struct {
	LogLevel string `json:"log_level"`
}

type TestMyBatisConfig struct {
	MapperLocations []string `json:"mapper_locations"`
}

// 简化的MyBatis实例，用于测试
type TestMyBatis struct {
	config *TestDatabaseConfig
}

func NewTestMyBatis(config *TestDatabaseConfig) (*TestMyBatis, error) {
	return &TestMyBatis{config: config}, nil
}

func (mb *TestMyBatis) OpenSession() *TestSqlSession {
	return &TestSqlSession{}
}

// 简化的SqlSession，用于测试
type TestSqlSession struct{}

func (s *TestSqlSession) Close() error {
	return nil
}

// 简化的Configuration，用于测试
type TestConfiguration struct {
	MapUnderscoreToCamelCase bool
	LazyLoadingEnabled       bool
	CacheEnabled             bool
}

func NewTestConfiguration() *TestConfiguration {
	return &TestConfiguration{}
}

func (c *TestConfiguration) SetDatabaseConfig(config *TestDatabaseConfig) {
	// 设置数据库配置
}

// XMLBasedTestSuite 基于XML的测试套件
type XMLBasedTestSuite struct {
	suite.Suite
	dbSetup       *DatabaseSetup
	mybatis       *mybatis.MyBatis
	xmlLoader     *XMLMapperLoader
	configBuilder *ConfigurationBuilder
	userMapper    UserMapper
}

// SetupSuite 设置测试套件
func (suite *XMLBasedTestSuite) SetupSuite() {
	// 初始化数据库设置
	suite.dbSetup = NewDatabaseSetup(DefaultDatabaseConfig())

	// 设置完整测试环境
	err := suite.dbSetup.SetupCompleteTestEnvironment()
	suite.Require().NoError(err)

	// 创建配置构建器
	suite.configBuilder = NewConfigurationBuilder(
		"mybatis-config.xml",
		"mappers",
	)

	// 加载属性文件
	err = suite.configBuilder.LoadProperties("database.properties")
	suite.Require().NoError(err)

	// 构建配置
	mybatisConfig, err := suite.configBuilder.Build()
	suite.Require().NoError(err)

	// 获取数据库配置
	dbConfig, err := frameworkConfig.GetDatabaseConfig()
	suite.Require().NoError(err)
	mybatisConfig.SetDatabaseConfig(dbConfig)

	// 创建MyBatis实例
	suite.mybatis, err = mybatis.NewMyBatis(mybatisConfig)
	suite.Require().NoError(err)

	// 创建映射器
	session := suite.mybatis.OpenSession()
	suite.userMapper = NewUserMapper(session)

	fmt.Println("=== XML Based MyBatis Test Suite Started ===")
}

// TearDownSuite 清理测试套件
func (suite *XMLBasedTestSuite) TearDownSuite() {
	if suite.dbSetup != nil {
		suite.dbSetup.TeardownCompleteTestEnvironment()
	}
	fmt.Println("=== XML Based MyBatis Test Suite Completed ===")
}

// TestXMLMapperLoading 测试XML映射器加载
func (suite *XMLBasedTestSuite) TestXMLMapperLoading() {
	t := suite.T()

	t.Run("加载单个XML映射器文件", func(t *testing.T) {
		loader := NewXMLMapperLoader("mappers", config.NewConfiguration())

		mapper, err := loader.LoadMapper("UserMapper.xml")
		assert.NoError(t, err)
		assert.NotNil(t, mapper)
		assert.Equal(t, "UserMapper", mapper.Namespace)

		// 验证映射器内容
		assert.NotEmpty(t, mapper.Selects)
		assert.NotEmpty(t, mapper.Inserts)
		assert.NotEmpty(t, mapper.Updates)
		assert.NotEmpty(t, mapper.Deletes)
		assert.NotEmpty(t, mapper.ResultMaps)
		assert.NotEmpty(t, mapper.SqlFragments)

		fmt.Printf("成功加载映射器: %s\n", mapper.Namespace)
		fmt.Printf("包含 %d 个查询语句\n", len(mapper.Selects))
		fmt.Printf("包含 %d 个插入语句\n", len(mapper.Inserts))
		fmt.Printf("包含 %d 个更新语句\n", len(mapper.Updates))
		fmt.Printf("包含 %d 个删除语句\n", len(mapper.Deletes))
		fmt.Printf("包含 %d 个结果映射\n", len(mapper.ResultMaps))
		fmt.Printf("包含 %d 个SQL片段\n", len(mapper.SqlFragments))
	})

	t.Run("加载所有XML映射器文件", func(t *testing.T) {
		loader := NewXMLMapperLoader("mappers", config.NewConfiguration())

		mappers, err := loader.LoadAllMappers()
		assert.NoError(t, err)
		assert.NotEmpty(t, mappers)

		fmt.Printf("成功加载 %d 个映射器文件\n", len(mappers))
		for _, mapper := range mappers {
			fmt.Printf("  - %s\n", mapper.Namespace)
		}
	})
}

// TestXMLConfigurationLoading 测试XML配置加载
func (suite *XMLBasedTestSuite) TestXMLConfigurationLoading() {
	t := suite.T()

	t.Run("加载MyBatis主配置文件", func(t *testing.T) {
		loader := NewXMLConfigLoader("mybatis-config.xml")

		cfg, err := loader.LoadConfiguration()
		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		fmt.Println("成功加载MyBatis主配置文件")
	})

	t.Run("加载属性文件", func(t *testing.T) {
		properties, err := LoadPropertiesFile("database.properties")
		assert.NoError(t, err)
		assert.NotEmpty(t, properties)

		// 验证关键属性
		assert.Contains(t, properties, "database.driver")
		assert.Contains(t, properties, "database.url")
		assert.Contains(t, properties, "database.username")

		fmt.Printf("成功加载 %d 个属性配置\n", len(properties))
		for key, value := range properties {
			if !contains(key, "password") { // 不打印密码
				fmt.Printf("  %s = %s\n", key, value)
			}
		}
	})
}

// TestXMLBasedCRUD 测试基于XML的CRUD操作
func (suite *XMLBasedTestSuite) TestXMLBasedCRUD() {
	t := suite.T()

	t.Run("XML配置的基础查询", func(t *testing.T) {
		// 使用XML配置的selectById查询
		user, err := suite.userMapper.SelectById(1)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(1), user.ID)

		fmt.Printf("XML查询结果: 用户 %s (%s)\n", user.Name, user.Email)
	})

	t.Run("XML配置的插入操作", func(t *testing.T) {
		newUser := &User{
			Name:   "XML测试用户",
			Email:  "xml-test@example.com",
			Age:    28,
			Status: "active",
			Phone:  "13900000001",
		}

		// 使用XML配置的insert语句
		id, err := suite.userMapper.Insert(newUser)
		assert.NoError(t, err)
		assert.Greater(t, id, int64(0))

		// 验证插入成功
		insertedUser, err := suite.userMapper.SelectByEmail("xml-test@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, insertedUser)
		assert.Equal(t, "XML测试用户", insertedUser.Name)

		fmt.Printf("XML插入成功: 用户ID=%d\n", id)
	})

	t.Run("XML配置的更新操作", func(t *testing.T) {
		// 先查询用户
		user, err := suite.userMapper.SelectByEmail("xml-test@example.com")
		require.NoError(t, err)
		require.NotNil(t, user)

		// 使用XML配置的update语句
		user.Age = 29
		user.Status = "inactive"
		affected, err := suite.userMapper.Update(user)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证更新成功
		updatedUser, err := suite.userMapper.SelectById(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 29, updatedUser.Age)
		assert.Equal(t, "inactive", updatedUser.Status)

		fmt.Printf("XML更新成功: 用户年龄=%d, 状态=%s\n", updatedUser.Age, updatedUser.Status)
	})

	t.Run("XML配置的删除操作", func(t *testing.T) {
		// 创建待删除用户
		testUser := &User{
			Name:   "待删除XML用户",
			Email:  "delete-xml@example.com",
			Age:    25,
			Status: "active",
		}

		id, err := suite.userMapper.Insert(testUser)
		require.NoError(t, err)

		// 使用XML配置的delete语句
		affected, err := suite.userMapper.Delete(id)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证删除成功（软删除）
		deletedUser, err := suite.userMapper.SelectById(id)
		assert.NoError(t, err)
		assert.Nil(t, deletedUser)

		fmt.Printf("XML软删除成功: 用户ID=%d\n", id)
	})
}

// TestXMLBasedDynamicSQL 测试基于XML的动态SQL
func (suite *XMLBasedTestSuite) TestXMLBasedDynamicSQL() {
	t := suite.T()

	t.Run("XML配置的动态查询", func(t *testing.T) {
		query := &UserQuery{
			Name:     "XML",
			Status:   "active",
			AgeMin:   20,
			AgeMax:   40,
			Page:     1,
			PageSize: 10,
		}

		// 使用XML配置的selectList查询
		users, err := suite.userMapper.SelectList(query)
		assert.NoError(t, err)

		fmt.Printf("XML动态查询结果: %d 个用户\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (%d岁, %s)\n", user.Name, user.Age, user.Status)
		}
	})

	t.Run("XML配置的条件统计", func(t *testing.T) {
		query := &UserQuery{
			Status: "active",
		}

		// 使用XML配置的selectCount查询
		count, err := suite.userMapper.SelectCount(query)
		assert.NoError(t, err)
		assert.Greater(t, count, int64(0))

		fmt.Printf("XML条件统计结果: %d 个活跃用户\n", count)
	})

	t.Run("XML配置的选择性更新", func(t *testing.T) {
		// 先查询用户
		users, err := suite.userMapper.SelectList(&UserQuery{
			Name:     "XML",
			PageSize: 1,
		})
		require.NoError(t, err)
		require.NotEmpty(t, users)

		user := users[0]
		originalAge := user.Age

		// 只更新年龄字段
		user.Age = originalAge + 1
		user.Name = "" // 清空，测试选择性更新

		// 使用XML配置的updateSelective语句
		affected, err := suite.userMapper.UpdateSelective(user)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), affected)

		// 验证选择性更新（名称不应该被清空）
		updatedUser, err := suite.userMapper.SelectById(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, originalAge+1, updatedUser.Age)
		assert.NotEmpty(t, updatedUser.Name) // 名称应该保持不变

		fmt.Printf("XML选择性更新成功: 年龄 %d -> %d\n", originalAge, updatedUser.Age)
	})
}

// TestXMLBasedBatchOperations 测试基于XML的批量操作
func (suite *XMLBasedTestSuite) TestXMLBasedBatchOperations() {
	t := suite.T()

	t.Run("XML配置的批量插入", func(t *testing.T) {
		users := []*User{
			{Name: "XML批量用户1", Email: "xml-batch1@example.com", Age: 25, Status: "active"},
			{Name: "XML批量用户2", Email: "xml-batch2@example.com", Age: 26, Status: "active"},
			{Name: "XML批量用户3", Email: "xml-batch3@example.com", Age: 27, Status: "inactive"},
		}

		// 使用XML配置的batchInsert语句
		affected, err := suite.userMapper.BatchInsert(users)
		assert.NoError(t, err)
		assert.Equal(t, int64(3), affected)

		// 验证批量插入成功
		for _, user := range users {
			insertedUser, err := suite.userMapper.SelectByEmail(user.Email)
			assert.NoError(t, err)
			assert.NotNil(t, insertedUser)
			assert.Equal(t, user.Name, insertedUser.Name)
		}

		fmt.Printf("XML批量插入成功: %d 个用户\n", len(users))
	})

	t.Run("XML配置的批量状态更新", func(t *testing.T) {
		// 获取XML批量用户的ID
		query := &UserQuery{
			Name:     "XML批量",
			PageSize: 10,
		}
		users, err := suite.userMapper.SelectList(query)
		require.NoError(t, err)
		require.NotEmpty(t, users)

		var ids []int64
		for _, user := range users {
			ids = append(ids, user.ID)
		}

		// 使用XML配置的batchUpdateStatus语句
		affected, err := suite.userMapper.BatchUpdateStatus(ids, "suspended")
		assert.NoError(t, err)
		assert.Equal(t, int64(len(ids)), affected)

		// 验证批量更新成功
		for _, id := range ids {
			user, err := suite.userMapper.SelectById(id)
			assert.NoError(t, err)
			assert.Equal(t, "suspended", user.Status)
		}

		fmt.Printf("XML批量状态更新成功: %d 个用户状态改为suspended\n", len(ids))
	})

	t.Run("XML配置的批量删除", func(t *testing.T) {
		// 获取要删除的用户ID
		query := &UserQuery{
			Name:     "XML批量",
			PageSize: 10,
		}
		users, err := suite.userMapper.SelectList(query)
		require.NoError(t, err)
		require.NotEmpty(t, users)

		var ids []int64
		for _, user := range users {
			ids = append(ids, user.ID)
		}

		// 使用XML配置的batchDelete语句
		affected, err := suite.userMapper.BatchDelete(ids)
		assert.NoError(t, err)
		assert.Equal(t, int64(len(ids)), affected)

		// 验证批量删除成功
		for _, id := range ids {
			user, err := suite.userMapper.SelectById(id)
			assert.NoError(t, err)
			assert.Nil(t, user) // 软删除后查询不到
		}

		fmt.Printf("XML批量删除成功: %d 个用户\n", len(ids))
	})
}

// TestXMLBasedComplexQueries 测试基于XML的复杂查询
func (suite *XMLBasedTestSuite) TestXMLBasedComplexQueries() {
	t := suite.T()

	t.Run("XML配置的关联查询", func(t *testing.T) {
		// 使用XML配置的selectWithProfile查询
		result, err := suite.userMapper.SelectWithProfile(1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.User)

		fmt.Printf("XML关联查询结果: 用户=%s\n", result.User.Name)
		if result.Profile != nil {
			fmt.Printf("  档案信息: 公司=%s, 职位=%s\n", result.Profile.Company, result.Profile.Occupation)
		} else {
			fmt.Printf("  档案信息: 无\n")
		}
	})

	t.Run("XML配置的集合查询", func(t *testing.T) {
		// 使用XML配置的selectWithRoles查询
		result, err := suite.userMapper.SelectWithRoles(1)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.User)

		fmt.Printf("XML集合查询结果: 用户=%s, 角色数量=%d\n", result.User.Name, len(result.Roles))
		for _, role := range result.Roles {
			fmt.Printf("  - 角色: %s\n", role.RoleName)
		}
	})

	t.Run("XML配置的全文搜索", func(t *testing.T) {
		// 使用XML配置的searchUsers查询
		users, err := suite.userMapper.SearchUsers("test", 5)
		assert.NoError(t, err)

		fmt.Printf("XML全文搜索结果: %d 个用户\n", len(users))
		for _, user := range users {
			fmt.Printf("  - %s (%s)\n", user.Name, user.Email)
		}
	})
}

// TestXMLBasedAggregation 测试基于XML的聚合查询
func (suite *XMLBasedTestSuite) TestXMLBasedAggregation() {
	t := suite.T()

	t.Run("XML配置的用户统计", func(t *testing.T) {
		// 使用XML配置的selectStats查询
		stats, err := suite.userMapper.SelectStats()
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.TotalUsers, int64(0))

		fmt.Printf("XML统计查询结果:\n")
		fmt.Printf("  总用户数: %d\n", stats.TotalUsers)
		fmt.Printf("  活跃用户数: %d\n", stats.ActiveUsers)
		fmt.Printf("  最近用户数: %d\n", stats.RecentUsers)
	})

	t.Run("XML配置的状态分组", func(t *testing.T) {
		// 使用XML配置的selectByStatus查询
		results, err := suite.userMapper.SelectByStatus()
		assert.NoError(t, err)
		assert.NotEmpty(t, results)

		fmt.Printf("XML状态分组结果:\n")
		for _, result := range results {
			fmt.Printf("  %s: %d 个用户\n", result.Value, result.Count)
		}
	})

	t.Run("XML配置的年龄分组", func(t *testing.T) {
		// 使用XML配置的selectByAgeGroup查询
		results, err := suite.userMapper.SelectByAgeGroup()
		assert.NoError(t, err)
		assert.NotEmpty(t, results)

		fmt.Printf("XML年龄分组结果:\n")
		for _, result := range results {
			fmt.Printf("  %s: %d 个用户\n", result.Value, result.Count)
		}
	})

	t.Run("XML配置的时间段查询", func(t *testing.T) {
		endTime := time.Now()
		startTime := endTime.AddDate(0, -1, 0) // 一个月前

		// 使用XML配置的selectActiveUsersInPeriod查询
		users, err := suite.userMapper.SelectActiveUsersInPeriod(startTime, endTime)
		assert.NoError(t, err)

		fmt.Printf("XML时间段查询结果: 最近一个月 %d 个活跃用户\n", len(users))
	})
}

// TestXMLBasedPerformance 测试基于XML的性能
func (suite *XMLBasedTestSuite) TestXMLBasedPerformance() {
	t := suite.T()

	t.Run("XML查询性能测试", func(t *testing.T) {
		iterations := 100

		start := time.Now()
		for i := 0; i < iterations; i++ {
			user, err := suite.userMapper.SelectById(1)
			assert.NoError(t, err)
			assert.NotNil(t, user)
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(iterations)
		fmt.Printf("XML查询性能: %d次查询耗时%v, 平均%v/次\n", iterations, duration, avgDuration)

		// 性能基准：平均每次查询应在10ms以内
		assert.Less(t, avgDuration, 10*time.Millisecond)
	})

	t.Run("XML动态查询性能测试", func(t *testing.T) {
		iterations := 50
		query := &UserQuery{
			Status:   "active",
			AgeMin:   20,
			AgeMax:   50,
			PageSize: 10,
		}

		start := time.Now()
		for i := 0; i < iterations; i++ {
			users, err := suite.userMapper.SelectList(query)
			assert.NoError(t, err)
			assert.NotNil(t, users)
		}
		duration := time.Since(start)

		avgDuration := duration / time.Duration(iterations)
		fmt.Printf("XML动态查询性能: %d次查询耗时%v, 平均%v/次\n", iterations, duration, avgDuration)

		// 性能基准：平均每次动态查询应在20ms以内
		assert.Less(t, avgDuration, 20*time.Millisecond)
	})
}

// TestXMLBasedCache 测试基于XML的缓存
func (suite *XMLBasedTestSuite) TestXMLBasedCache() {
	t := suite.T()

	t.Run("XML查询缓存测试", func(t *testing.T) {
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

		fmt.Printf("XML缓存测试: 第一次=%v, 第二次=%v\n", duration1, duration2)

		// 缓存应该明显提升性能
		if duration1 > time.Microsecond && duration2 > time.Microsecond {
			// 只有在两次查询都有意义的时间时才比较
			fmt.Printf("缓存性能提升: %.2fx\n", float64(duration1)/float64(duration2))
		}
	})
}

// TestXMLBasedSuite 运行基于XML的测试套件
func TestXMLBasedSuite(t *testing.T) {
	suite.Run(t, new(XMLBasedTestSuite))
}

// BenchmarkXMLBasedOperations 基于XML的性能基准测试
func BenchmarkXMLBasedOperations(b *testing.B) {
	// 设置测试环境
	dbSetup := NewDatabaseSetup(DefaultDatabaseConfig())
	err := dbSetup.SetupCompleteTestEnvironment()
	if err != nil {
		b.Fatalf("Failed to setup test environment: %v", err)
	}
	defer dbSetup.TeardownCompleteTestEnvironment()

	// 创建配置构建器
	configBuilder := NewConfigurationBuilder("mybatis-config.xml", "mappers")
	configBuilder.LoadProperties("database.properties")

	// 构建配置
	mybatisConfig, err := configBuilder.Build()
	if err != nil {
		b.Fatalf("Failed to build configuration: %v", err)
	}

	// 获取数据库配置
	dbConfig, err := frameworkConfig.GetDatabaseConfig()
	if err != nil {
		b.Fatalf("Failed to get database config: %v", err)
	}
	mybatisConfig.SetDatabaseConfig(dbConfig)

	// 创建MyBatis实例
	mb, err := mybatis.NewMyBatis(mybatisConfig)
	if err != nil {
		b.Fatalf("Failed to create MyBatis instance: %v", err)
	}

	session := mb.OpenSession()
	defer session.Close()

	userMapper := NewUserMapper(session)

	b.Run("XMLSelectById", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := userMapper.SelectById(1)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("XMLDynamicQuery", func(b *testing.B) {
		query := &UserQuery{
			Status:   "active",
			PageSize: 10,
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := userMapper.SelectList(query)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// 辅助函数
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
