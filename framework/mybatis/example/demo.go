// Package example MyBatis框架演示程序
//
// 展示如何使用MyBatis-Go框架进行完整的数据库操作
package example

import (
	"fmt"
	"log"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis/config"
	"github.com/zsy619/yyhertz/framework/mybatis/session"
	frameworkConfig "github.com/zsy619/yyhertz/framework/config"
)

// MyBatisDemo MyBatis演示程序
type MyBatisDemo struct {
	sqlSessionFactory session.SqlSessionFactory
	configuration     *config.Configuration
}

// NewMyBatisDemo 创建MyBatis演示程序
func NewMyBatisDemo() (*MyBatisDemo, error) {
	// 创建配置
	configuration := config.NewConfiguration()
	
	// 设置数据库配置 (使用SQLite作为演示)
	dbConfig := &frameworkConfig.DatabaseConfig{}
	dbConfig.Primary.Driver = "sqlite"
	dbConfig.Primary.Database = "mybatis_demo.db"
	dbConfig.Primary.MaxIdleConns = 5
	dbConfig.Primary.MaxOpenConns = 10
	dbConfig.Primary.ConnMaxLifetime = "1h"
	dbConfig.Primary.LogLevel = "info"
	configuration.SetDatabaseConfig(dbConfig)
	
	// 注册用户映射器
	err := RegisterUserMapperStatements(configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to register user mapper: %w", err)
	}
	
	// 创建SQL会话工厂
	sqlSessionFactory, err := session.NewDefaultSqlSessionFactory(configuration)
	if err != nil {
		return nil, fmt.Errorf("failed to create sql session factory: %w", err)
	}
	
	demo := &MyBatisDemo{
		sqlSessionFactory: sqlSessionFactory,
		configuration:     configuration,
	}
	
	// 初始化数据库表
	err = demo.initDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}
	
	return demo, nil
}

// initDatabase 初始化数据库表
func (demo *MyBatisDemo) initDatabase() error {
	sqlSession := demo.sqlSessionFactory.OpenSession()
	defer sqlSession.Close()
	
	db := sqlSession.GetConnection()
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}
	
	// 创建用户表
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			age INTEGER NOT NULL,
			status VARCHAR(20) DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	
	err := db.Exec(createTableSQL).Error
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	
	log.Println("✅ 数据库表初始化成功")
	return nil
}

// RunDemo 运行演示程序
func (demo *MyBatisDemo) RunDemo() error {
	log.Println("🚀 开始MyBatis-Go框架演示...")
	
	// 1. 演示基本CRUD操作
	err := demo.demonstrateCRUD()
	if err != nil {
		return fmt.Errorf("CRUD demonstration failed: %w", err)
	}
	
	// 2. 演示动态SQL查询
	err = demo.demonstrateDynamicSQL()
	if err != nil {
		return fmt.Errorf("dynamic SQL demonstration failed: %w", err)
	}
	
	// 3. 演示批量操作
	err = demo.demonstrateBatchOperations()
	if err != nil {
		return fmt.Errorf("batch operations demonstration failed: %w", err)
	}
	
	// 4. 演示事务操作
	err = demo.demonstrateTransactions()
	if err != nil {
		return fmt.Errorf("transaction demonstration failed: %w", err)
	}
	
	// 5. 演示缓存机制
	err = demo.demonstrateCaching()
	if err != nil {
		return fmt.Errorf("caching demonstration failed: %w", err)
	}
	
	log.Println("🎉 MyBatis-Go框架演示完成！")
	return nil
}

// demonstrateCRUD 演示CRUD操作
func (demo *MyBatisDemo) demonstrateCRUD() error {
	log.Println("\n📝 演示基本CRUD操作...")
	
	// 获取SQL会话
	sqlSession := demo.sqlSessionFactory.OpenSession()
	defer sqlSession.Close()
	
	// 获取用户映射器
	userMapper := NewUserMapper(sqlSession)
	
	// 1. 插入用户
	log.Println("1️⃣ 插入新用户...")
	newUser := &User{
		Name:      "张三",
		Email:     "zhangsan@example.com",
		Age:       25,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	insertResult, err := userMapper.Insert(newUser)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	log.Printf("✅ 插入用户成功，影响行数: %d", insertResult)
	
	// 2. 根据邮箱查询用户
	log.Println("2️⃣ 根据邮箱查询用户...")
	foundUser, err := userMapper.SelectByEmail("zhangsan@example.com")
	if err != nil {
		return fmt.Errorf("failed to select user by email: %w", err)
	}
	if foundUser != nil {
		log.Printf("✅ 查询到用户: ID=%d, Name=%s, Email=%s", foundUser.ID, foundUser.Name, foundUser.Email)
	}
	
	// 3. 更新用户
	if foundUser != nil {
		log.Println("3️⃣ 更新用户信息...")
		foundUser.Age = 26
		foundUser.UpdatedAt = time.Now()
		
		updateResult, err := userMapper.Update(foundUser)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		log.Printf("✅ 更新用户成功，影响行数: %d", updateResult)
	}
	
	// 4. 根据ID查询用户
	if foundUser != nil {
		log.Println("4️⃣ 根据ID查询用户...")
		userById, err := userMapper.SelectById(foundUser.ID)
		if err != nil {
			return fmt.Errorf("failed to select user by id: %w", err)
		}
		if userById != nil {
			log.Printf("✅ 查询到用户: ID=%d, Name=%s, Age=%d", userById.ID, userById.Name, userById.Age)
		}
	}
	
	return nil
}

// demonstrateDynamicSQL 演示动态SQL查询
func (demo *MyBatisDemo) demonstrateDynamicSQL() error {
	log.Println("\n🔍 演示动态SQL查询...")
	
	sqlSession := demo.sqlSessionFactory.OpenSession()
	defer sqlSession.Close()
	
	userMapper := NewUserMapper(sqlSession)
	
	// 创建测试数据
	testUsers := []*User{
		{Name: "李四", Email: "lisi@example.com", Age: 30, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "王五", Email: "wangwu@example.com", Age: 28, Status: "inactive", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "赵六", Email: "zhaoliu@example.com", Age: 35, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	
	for _, user := range testUsers {
		_, err := userMapper.Insert(user)
		if err != nil {
			log.Printf("⚠️  插入测试用户失败: %v", err)
		}
	}
	
	// 1. 按名称模糊查询
	log.Println("1️⃣ 按名称模糊查询...")
	query1 := &UserQuery{Name: "李"}
	users1, err := userMapper.SelectList(query1)
	if err != nil {
		return fmt.Errorf("failed to select users by name: %w", err)
	}
	log.Printf("✅ 找到 %d 个包含'李'的用户", len(users1))
	
	// 2. 按年龄范围查询
	log.Println("2️⃣ 按年龄范围查询...")
	query2 := &UserQuery{AgeMin: 25, AgeMax: 30}
	users2, err := userMapper.SelectList(query2)
	if err != nil {
		return fmt.Errorf("failed to select users by age range: %w", err)
	}
	log.Printf("✅ 找到 %d 个年龄在25-30之间的用户", len(users2))
	
	// 3. 按状态查询
	log.Println("3️⃣ 按状态查询...")
	query3 := &UserQuery{Status: "active"}
	users3, err := userMapper.SelectList(query3)
	if err != nil {
		return fmt.Errorf("failed to select users by status: %w", err)
	}
	log.Printf("✅ 找到 %d 个活跃用户", len(users3))
	
	// 4. 复合条件查询
	log.Println("4️⃣ 复合条件查询...")
	query4 := &UserQuery{AgeMin: 25, Status: "active"}
	users4, err := userMapper.SelectList(query4)
	if err != nil {
		return fmt.Errorf("failed to select users by multiple conditions: %w", err)
	}
	log.Printf("✅ 找到 %d 个年龄>=25且状态为active的用户", len(users4))
	
	// 5. 统计查询
	log.Println("5️⃣ 统计查询...")
	count, err := userMapper.SelectCount(&UserQuery{Status: "active"})
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}
	log.Printf("✅ 活跃用户总数: %d", count)
	
	return nil
}

// demonstrateBatchOperations 演示批量操作
func (demo *MyBatisDemo) demonstrateBatchOperations() error {
	log.Println("\n📦 演示批量操作...")
	
	sqlSession := demo.sqlSessionFactory.OpenSession()
	defer sqlSession.Close()
	
	userMapper := NewUserMapper(sqlSession)
	
	// 批量插入用户
	log.Println("1️⃣ 批量插入用户...")
	batchUsers := []*User{
		{Name: "批量用户1", Email: "batch1@example.com", Age: 20, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "批量用户2", Email: "batch2@example.com", Age: 21, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "批量用户3", Email: "batch3@example.com", Age: 22, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "批量用户4", Email: "batch4@example.com", Age: 23, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{Name: "批量用户5", Email: "batch5@example.com", Age: 24, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	
	batchResult, err := userMapper.BatchInsert(batchUsers)
	if err != nil {
		return fmt.Errorf("failed to batch insert users: %w", err)
	}
	log.Printf("✅ 批量插入成功，影响行数: %d", batchResult)
	
	// 查询所有批量插入的用户
	log.Println("2️⃣ 查询批量插入的用户...")
	query := &UserQuery{Name: "批量用户"}
	batchQueryUsers, err := userMapper.SelectList(query)
	if err != nil {
		return fmt.Errorf("failed to query batch users: %w", err)
	}
	log.Printf("✅ 查询到 %d 个批量用户", len(batchQueryUsers))
	
	return nil
}

// demonstrateTransactions 演示事务操作
func (demo *MyBatisDemo) demonstrateTransactions() error {
	log.Println("\n💳 演示事务操作...")
	
	// 1. 成功的事务
	log.Println("1️⃣ 演示成功的事务...")
	err := demo.executeSuccessfulTransaction()
	if err != nil {
		return fmt.Errorf("successful transaction demonstration failed: %w", err)
	}
	
	// 2. 失败回滚的事务
	log.Println("2️⃣ 演示失败回滚的事务...")
	err = demo.executeFailedTransaction()
	if err != nil {
		log.Printf("✅ 事务按预期回滚: %v", err)
	}
	
	return nil
}

// executeSuccessfulTransaction 执行成功的事务
func (demo *MyBatisDemo) executeSuccessfulTransaction() error {
	sqlSession := demo.sqlSessionFactory.OpenSession() // 手动事务
	defer sqlSession.Close()
	
	userMapper := NewUserMapper(sqlSession)
	
	// 在事务中插入多个用户
	transactionUser1 := &User{
		Name: "事务用户1", Email: "tx1@example.com", Age: 30,
		Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	
	transactionUser2 := &User{
		Name: "事务用户2", Email: "tx2@example.com", Age: 31,
		Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	
	_, err := userMapper.Insert(transactionUser1)
	if err != nil {
		sqlSession.Rollback()
		return fmt.Errorf("failed to insert transaction user 1: %w", err)
	}
	
	_, err = userMapper.Insert(transactionUser2)
	if err != nil {
		sqlSession.Rollback()
		return fmt.Errorf("failed to insert transaction user 2: %w", err)
	}
	
	// 提交事务
	err = sqlSession.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	log.Println("✅ 事务提交成功")
	return nil
}

// executeFailedTransaction 执行失败的事务
func (demo *MyBatisDemo) executeFailedTransaction() error {
	sqlSession := demo.sqlSessionFactory.OpenSession()
	defer sqlSession.Close()
	
	userMapper := NewUserMapper(sqlSession)
	
	// 尝试插入重复邮箱的用户（应该失败）
	duplicateUser := &User{
		Name: "重复用户", Email: "tx1@example.com", // 重复邮箱
		Age: 25, Status: "active", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	
	_, err := userMapper.Insert(duplicateUser)
	if err != nil {
		sqlSession.Rollback()
		return fmt.Errorf("transaction rolled back due to duplicate email")
	}
	
	return nil
}

// demonstrateCaching 演示缓存机制
func (demo *MyBatisDemo) demonstrateCaching() error {
	log.Println("\n🗄️ 演示缓存机制...")
	
	sqlSession := demo.sqlSessionFactory.OpenSession()
	defer sqlSession.Close()
	
	userMapper := NewUserMapper(sqlSession)
	
	// 多次查询同一用户，测试一级缓存
	log.Println("1️⃣ 测试一级缓存（同一会话）...")
	startTime := time.Now()
	user1, err := userMapper.SelectByEmail("zhangsan@example.com")
	firstQueryTime := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("first query failed: %w", err)
	}
	
	startTime = time.Now()
	user2, err := userMapper.SelectByEmail("zhangsan@example.com")
	secondQueryTime := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("second query failed: %w", err)
	}
	
	log.Printf("✅ 第一次查询耗时: %v", firstQueryTime)
	log.Printf("✅ 第二次查询耗时: %v", secondQueryTime)
	if user1 != nil && user2 != nil {
		log.Printf("✅ 查询结果一致: %t", user1.ID == user2.ID)
	}
	
	// 清除缓存
	log.Println("2️⃣ 清除缓存...")
	sqlSession.ClearCache()
	
	startTime = time.Now()
	user3, err := userMapper.SelectByEmail("zhangsan@example.com")
	thirdQueryTime := time.Since(startTime)
	if err != nil {
		return fmt.Errorf("third query failed: %w", err)
	}
	
	log.Printf("✅ 清除缓存后查询耗时: %v", thirdQueryTime)
	if user3 != nil {
		log.Printf("✅ 用户信息: ID=%d, Name=%s", user3.ID, user3.Name)
	}
	
	return nil
}

// Close 关闭演示程序
func (demo *MyBatisDemo) Close() error {
	// 这里可以添加清理逻辑
	log.Println("🔚 MyBatis演示程序已关闭")
	return nil
}

// RunMyBatisDemo 运行MyBatis演示的主函数
func RunMyBatisDemo() {
	demo, err := NewMyBatisDemo()
	if err != nil {
		log.Fatalf("❌ 创建MyBatis演示程序失败: %v", err)
	}
	defer demo.Close()
	
	err = demo.RunDemo()
	if err != nil {
		log.Fatalf("❌ 运行MyBatis演示程序失败: %v", err)
	}
}