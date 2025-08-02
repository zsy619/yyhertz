package orm

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 连接池使用示例 =============

// RunPoolExamples 运行连接池示例
func RunPoolExamples() error {
	config.Info("Starting connection pool examples...")
	
	// 1. 单节点连接池示例
	config.Info("=== Single Node Pool Example ===")
	if err := singleNodePoolExample(); err != nil {
		return fmt.Errorf("single node pool example failed: %w", err)
	}
	
	// 2. 主从连接池示例
	config.Info("=== Master-Slave Pool Example ===")
	if err := masterSlavePoolExample(); err != nil {
		return fmt.Errorf("master-slave pool example failed: %w", err)
	}
	
	// 3. 集群连接池示例
	config.Info("=== Cluster Pool Example ===")
	if err := clusterPoolExample(); err != nil {
		return fmt.Errorf("cluster pool example failed: %w", err)
	}
	
	// 4. 增强ORM示例
	config.Info("=== Enhanced ORM Example ===")
	if err := enhancedORMExample(); err != nil {
		return fmt.Errorf("enhanced ORM example failed: %w", err)
	}
	
	// 5. 性能测试示例
	config.Info("=== Performance Test Example ===")
	if err := performanceTestExample(); err != nil {
		return fmt.Errorf("performance test example failed: %w", err)
	}
	
	// 6. 监控指标示例
	config.Info("=== Metrics Example ===")
	if err := metricsExample(); err != nil {
		return fmt.Errorf("metrics example failed: %w", err)
	}
	
	config.Info("Connection pool examples completed successfully!")
	return nil
}

// singleNodePoolExample 单节点连接池示例
func singleNodePoolExample() error {
	// 创建数据库配置
	dbConfig := &DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 5,
		MaxOpenConns: 10,
		MaxLifetime:  3600,
	}
	
	// 创建连接池配置
	poolConfig := DefaultPoolConfig()
	poolConfig.MaxIdleConns = 5
	poolConfig.MaxOpenConns = 10
	poolConfig.HealthCheckEnabled = true
	poolConfig.MetricsEnabled = true
	
	// 创建数据库节点
	node := &DatabaseNode{
		ID:       "primary",
		Config:   dbConfig,
		Weight:   10,
		IsMaster: true,
	}
	
	// 创建连接池
	pool, err := NewMultiNodePool(poolConfig, []*DatabaseNode{node})
	if err != nil {
		return err
	}
	defer pool.Close()
	
	// 测试连接获取
	ctx := context.Background()
	
	for i := 0; i < 5; i++ {
		db, err := pool.GetConnection(ctx)
		if err != nil {
			return fmt.Errorf("failed to get connection %d: %w", i, err)
		}
		
		// 执行简单查询
		var result int
		if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		
		config.Infof("Connection %d: query result = %d", i, result)
		
		// 释放连接
		pool.ReleaseConnection(db)
	}
	
	// 健康检查
	if err := pool.HealthCheck(ctx); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	
	// 获取统计信息
	stats := pool.Stats()
	config.Infof("Pool stats: Total=%d, Active=%d, Failed=%d", 
		stats.TotalConnections, stats.OpenConnections, stats.FailedConnections)
	
	return nil
}

// masterSlavePoolExample 主从连接池示例
func masterSlavePoolExample() error {
	// 主库配置
	masterConfig := &DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 5,
		MaxOpenConns: 10,
	}
	
	// 从库配置
	slaveConfig := &DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 3,
		MaxOpenConns: 8,
	}
	
	// 创建主从连接池
	pool, err := CreateMasterSlavePool(masterConfig, slaveConfig)
	if err != nil {
		return err
	}
	defer pool.Close()
	
	ctx := context.Background()
	
	// 测试主库连接
	masterDB, err := pool.GetMasterConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get master connection: %w", err)
	}
	
	var masterResult int
	if err := masterDB.Raw("SELECT 1").Scan(&masterResult).Error; err != nil {
		return fmt.Errorf("master query failed: %w", err)
	}
	config.Infof("Master connection: query result = %d", masterResult)
	
	// 测试从库连接
	slaveDB, err := pool.GetSlaveConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to get slave connection: %w", err)
	}
	
	var slaveResult int
	if err := slaveDB.Raw("SELECT 1").Scan(&slaveResult).Error; err != nil {
		return fmt.Errorf("slave query failed: %w", err)
	}
	config.Infof("Slave connection: query result = %d", slaveResult)
	
	return nil
}

// clusterPoolExample 集群连接池示例
func clusterPoolExample() error {
	// 创建多个数据库配置
	configs := []*DatabaseConfig{
		{Type: "sqlite", Database: ":memory:", MaxIdleConns: 5, MaxOpenConns: 10}, // 主库
		{Type: "sqlite", Database: ":memory:", MaxIdleConns: 3, MaxOpenConns: 8},  // 从库1
		{Type: "sqlite", Database: ":memory:", MaxIdleConns: 3, MaxOpenConns: 8},  // 从库2
	}
	
	// 创建集群连接池
	pool, err := CreateClusterPool(configs)
	if err != nil {
		return err
	}
	defer pool.Close()
	
	ctx := context.Background()
	
	// 测试多次连接获取（验证负载均衡）
	for i := 0; i < 10; i++ {
		db, err := pool.GetConnection(ctx)
		if err != nil {
			return fmt.Errorf("failed to get connection %d: %w", i, err)
		}
		
		var result int
		if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
			return fmt.Errorf("query %d failed: %w", i, err)
		}
		
		config.Infof("Cluster connection %d: query result = %d", i, result)
		pool.ReleaseConnection(db)
	}
	
	return nil
}

// enhancedORMExample 增强ORM示例
func enhancedORMExample() error {
	// 创建数据库配置
	dbConfig := &DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 5,
		MaxOpenConns: 10,
	}
	
	// 创建连接池配置
	poolConfig := DefaultPoolConfig()
	poolConfig.MetricsEnabled = true
	poolConfig.HealthCheckEnabled = true
	
	// 创建增强ORM
	enhancedORM, err := NewEnhancedORM(dbConfig, poolConfig)
	if err != nil {
		return err
	}
	defer enhancedORM.Close()
	
	ctx := context.Background()
	
	// 创建示例表
	type User struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100"`
		Email string `gorm:"size:100;uniqueIndex"`
	}
	
	// 获取写连接进行迁移
	writeDB, err := enhancedORM.GetWriteDB(ctx)
	if err != nil {
		return err
	}
	
	if err := writeDB.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	
	// 创建用户
	user := &User{Name: "John Doe", Email: "john@example.com"}
	if err := enhancedORM.Create(ctx, user); err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	config.Infof("Created user: %+v", user)
	
	// 查询用户
	var foundUser User
	if err := enhancedORM.First(ctx, &foundUser, "email = ?", "john@example.com"); err != nil {
		return fmt.Errorf("find user failed: %w", err)
	}
	config.Infof("Found user: %+v", foundUser)
	
	// 更新用户
	if err := enhancedORM.Update(ctx, &foundUser, "name", "Jane Doe"); err != nil {
		return fmt.Errorf("update user failed: %w", err)
	}
	config.Info("Updated user name")
	
	// 计数
	count, err := enhancedORM.Count(ctx, &User{})
	if err != nil {
		return fmt.Errorf("count users failed: %w", err)
	}
	config.Infof("Total users: %d", count)
	
	// 事务示例
	err = enhancedORM.Transaction(ctx, func(tx *gorm.DB) error {
		user2 := &User{Name: "Bob Smith", Email: "bob@example.com"}
		if err := tx.Create(user2).Error; err != nil {
			return err
		}
		
		user3 := &User{Name: "Alice Brown", Email: "alice@example.com"}
		if err := tx.Create(user3).Error; err != nil {
			return err
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	config.Info("Transaction completed")
	
	// 最终计数
	finalCount, err := enhancedORM.Count(ctx, &User{})
	if err != nil {
		return fmt.Errorf("final count failed: %w", err)
	}
	config.Infof("Final user count: %d", finalCount)
	
	return nil
}

// performanceTestExample 性能测试示例
func performanceTestExample() error {
	// 创建数据库配置
	dbConfig := &DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 10,
		MaxOpenConns: 50,
	}
	
	// 创建连接池配置
	poolConfig := DefaultPoolConfig()
	poolConfig.MaxIdleConns = 10
	poolConfig.MaxOpenConns = 50
	poolConfig.MetricsEnabled = true
	
	// 创建增强ORM
	enhancedORM, err := NewEnhancedORM(dbConfig, poolConfig)
	if err != nil {
		return err
	}
	defer enhancedORM.Close()
	
	ctx := context.Background()
	
	// 创建测试表
	type TestRecord struct {
		ID    uint   `gorm:"primaryKey"`
		Name  string `gorm:"size:100"`
		Value int
	}
	
	writeDB, err := enhancedORM.GetWriteDB(ctx)
	if err != nil {
		return err
	}
	
	if err := writeDB.AutoMigrate(&TestRecord{}); err != nil {
		return err
	}
	
	// 并发写入测试
	config.Info("Starting concurrent write test...")
	
	const numGoroutines = 10
	const recordsPerGoroutine = 100
	
	start := time.Now()
	var wg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < recordsPerGoroutine; j++ {
				record := &TestRecord{
					Name:  fmt.Sprintf("Worker%d-Record%d", workerID, j),
					Value: workerID*1000 + j,
				}
				
				if err := enhancedORM.Create(ctx, record); err != nil {
					config.Errorf("Failed to create record: %v", err)
					return
				}
			}
		}(i)
	}
	
	wg.Wait()
	writeDuration := time.Since(start)
	
	totalRecords := numGoroutines * recordsPerGoroutine
	config.Infof("Created %d records in %v (%.2f records/sec)", 
		totalRecords, writeDuration, float64(totalRecords)/writeDuration.Seconds())
	
	// 并发读取测试
	config.Info("Starting concurrent read test...")
	
	start = time.Now()
	var readWg sync.WaitGroup
	
	for i := 0; i < numGoroutines; i++ {
		readWg.Add(1)
		go func(workerID int) {
			defer readWg.Done()
			
			for j := 0; j < recordsPerGoroutine; j++ {
				var records []TestRecord
				if err := enhancedORM.Find(ctx, &records, "value >= ?", workerID*1000); err != nil {
					config.Errorf("Failed to read records: %v", err)
					return
				}
			}
		}(i)
	}
	
	readWg.Wait()
	readDuration := time.Since(start)
	
	totalReads := numGoroutines * recordsPerGoroutine
	config.Infof("Performed %d reads in %v (%.2f reads/sec)", 
		totalReads, readDuration, float64(totalReads)/readDuration.Seconds())
	
	return nil
}

// metricsExample 监控指标示例
func metricsExample() error {
	// 获取全局增强ORM
	enhancedORM := GetGlobalEnhancedORM()
	
	ctx := context.Background()
	
	// 执行一些操作来生成指标
	type MetricTest struct {
		ID   uint   `gorm:"primaryKey"`
		Name string `gorm:"size:100"`
	}
	
	writeDB, err := enhancedORM.GetWriteDB(ctx)
	if err != nil {
		return err
	}
	
	if err := writeDB.AutoMigrate(&MetricTest{}); err != nil {
		return err
	}
	
	// 创建一些记录
	for i := 0; i < 10; i++ {
		record := &MetricTest{Name: fmt.Sprintf("Test-%d", i)}
		if err := enhancedORM.Create(ctx, record); err != nil {
			config.Errorf("Failed to create test record: %v", err)
		}
	}
	
	// 执行一些查询
	for i := 0; i < 20; i++ {
		var records []MetricTest
		if err := enhancedORM.Find(ctx, &records); err != nil {
			config.Errorf("Failed to query test records: %v", err)
		}
	}
	
	// 获取并打印指标
	config.Info("=== Connection Pool Metrics ===")
	
	// 连接池统计
	poolStats := enhancedORM.GetPoolStats()
	config.Infof("Pool Stats:")
	config.Infof("  Max Open Connections: %d", poolStats.MaxOpenConnections)
	config.Infof("  Open Connections: %d", poolStats.OpenConnections)
	config.Infof("  In Use: %d", poolStats.InUse)
	config.Infof("  Idle: %d", poolStats.Idle)
	config.Infof("  Total Connections: %d", poolStats.TotalConnections)
	config.Infof("  Total Queries: %d", poolStats.TotalQueries)
	config.Infof("  Slow Queries: %d", poolStats.SlowQueries)
	config.Infof("  Failed Connections: %d", poolStats.FailedConnections)
	config.Infof("  Healthy Nodes: %d", poolStats.HealthyNodes)
	config.Infof("  Unhealthy Nodes: %d", poolStats.UnhealthyNodes)
	
	// 详细指标
	metrics := enhancedORM.GetMetrics()
	if metrics != nil {
		config.Infof("Detailed Metrics:")
		config.Infof("  Total Connections: %d", metrics.TotalConnections)
		config.Infof("  Active Connections: %d", metrics.ActiveConnections)
		config.Infof("  Total Queries: %d", metrics.TotalQueries)
		config.Infof("  Slow Queries: %d", metrics.SlowQueries)
		config.Infof("  Failed Queries: %d", metrics.FailedQueries)
		config.Infof("  Average Query Time: %v", metrics.AverageQueryTime)
		
		// 节点指标
		for nodeID, nodeMetrics := range metrics.NodeMetrics {
			config.Infof("  Node %s:", nodeID)
			config.Infof("    Total Connections: %d", nodeMetrics.TotalConnections)
			config.Infof("    Total Queries: %d", nodeMetrics.TotalQueries)
			config.Infof("    Average Response Time: %v", nodeMetrics.AverageResponseTime)
		}
		
		// 时间窗口指标
		if metrics.WindowMetrics != nil {
			config.Infof("  Window Metrics:")
			config.Infof("    1 Min - QPS: %.2f, Queries: %d", 
				metrics.WindowMetrics.OneMinute.QPS, metrics.WindowMetrics.OneMinute.TotalQueries)
			config.Infof("    5 Min - QPS: %.2f, Queries: %d", 
				metrics.WindowMetrics.FiveMinutes.QPS, metrics.WindowMetrics.FiveMinutes.TotalQueries)
		}
	}
	
	return nil
}

// ============= 健康检查示例 =============

// RunHealthCheckExample 运行健康检查示例
func RunHealthCheckExample() error {
	config.Info("Starting health check example...")
	
	// 获取全局增强ORM
	enhancedORM := GetGlobalEnhancedORM()
	
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	
	// 执行健康检查
	if err := enhancedORM.HealthCheck(ctx); err != nil {
		config.Errorf("Health check failed: %v", err)
		return err
	}
	
	config.Info("Health check passed")
	
	// 获取健康状态
	stats := enhancedORM.GetPoolStats()
	config.Infof("Health status: %d healthy nodes, %d unhealthy nodes", 
		stats.HealthyNodes, stats.UnhealthyNodes)
	
	return nil
}

// ============= 故障转移示例 =============

// simulateFailoverExample 模拟故障转移示例
func simulateFailoverExample() error {
	config.Info("Simulating failover scenario...")
	
	// 这个示例展示了如何处理节点故障
	// 在实际应用中，连接池会自动检测故障节点并转移到健康节点
	
	enhancedORM := GetGlobalEnhancedORM()
	ctx := context.Background()
	
	// 执行操作前的健康检查
	if err := enhancedORM.HealthCheck(ctx); err != nil {
		config.Warnf("Initial health check failed: %v", err)
	}
	
	// 尝试执行操作
	type FailoverTest struct {
		ID   uint   `gorm:"primaryKey"`
		Data string
	}
	
	writeDB, err := enhancedORM.GetWriteDB(ctx)
	if err != nil {
		config.Errorf("Failed to get write DB: %v", err)
		return err
	}
	
	if err := writeDB.AutoMigrate(&FailoverTest{}); err != nil {
		config.Errorf("Migration failed: %v", err)
		return err
	}
	
	// 创建测试记录
	record := &FailoverTest{Data: "test data"}
	if err := enhancedORM.Create(ctx, record); err != nil {
		config.Errorf("Create failed: %v", err)
		// 在实际场景中，这里可能会触发故障转移逻辑
	} else {
		config.Info("Operation succeeded despite potential node issues")
	}
	
	return nil
}

// ============= 便捷函数 =============

// CreateExampleDatabase 创建示例数据库
func CreateExampleDatabase() (*EnhancedORM, error) {
	dbConfig := &DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
	}
	
	poolConfig := DefaultPoolConfig()
	poolConfig.MetricsEnabled = true
	poolConfig.HealthCheckEnabled = true
	
	return NewEnhancedORM(dbConfig, poolConfig)
}

// BenchmarkConnectionPool 连接池性能基准测试
func BenchmarkConnectionPool(iterations int) error {
	enhancedORM, err := CreateExampleDatabase()
	if err != nil {
		return err
	}
	defer enhancedORM.Close()
	
	ctx := context.Background()
	
	config.Infof("Running connection pool benchmark with %d iterations", iterations)
	
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		// 获取连接
		db, err := enhancedORM.GetWriteDB(ctx)
		if err != nil {
			return fmt.Errorf("failed to get connection: %w", err)
		}
		
		// 执行简单查询
		var result int
		if err := db.Raw("SELECT 1").Scan(&result).Error; err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		
		// 释放连接 (在GORM中自动管理)
		// enhancedORM.ReleaseConnection(db)
	}
	
	duration := time.Since(start)
	
	config.Infof("Benchmark completed: %d operations in %v", iterations, duration)
	config.Infof("Average time per operation: %v", duration/time.Duration(iterations))
	config.Infof("Operations per second: %.2f", float64(iterations)/duration.Seconds())
	
	// 打印最终指标
	PrintPoolMetrics()
	
	return nil
}