package mybatis_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDataSourceRouter(t *testing.T) {
	t.Run("DataSourceRouter_Basic", func(t *testing.T) {
		// 创建路由器
		config := &mybatis.DataSourceConfig{
			ReadWriteSplit: true,
			LoadBalance:    mybatis.LoadBalanceRoundRobin,
			HealthCheck:    mybatis.DefaultHealthCheckConfig(),
		}
		router := mybatis.NewDataSourceRouter(config)
		defer router.Close()

		// 创建测试数据库
		masterDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create master database: %v", err)
		}

		slaveDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create slave database: %v", err)
		}

		// 添加数据源
		masterDS := &mybatis.DataSource{
			Name:     "master",
			Type:     mybatis.DataSourceTypeMaster,
			Weight:   10,
			DB:       masterDB,
			IsActive: true,
		}
		
		slaveDS := &mybatis.DataSource{
			Name:     "slave",
			Type:     mybatis.DataSourceTypeSlave,
			Weight:   10,
			DB:       slaveDB,
			IsActive: true,
		}

		err = router.AddDataSource(masterDS)
		if err != nil {
			t.Fatalf("Failed to add master datasource: %v", err)
		}

		err = router.AddDataSource(slaveDS)
		if err != nil {
			t.Fatalf("Failed to add slave datasource: %v", err)
		}

		// 测试读操作路由到从库
		ctx := context.Background()
		readDS, err := router.GetDataSource(ctx, mybatis.OperationTypeRead)
		if err != nil {
			t.Fatalf("Failed to get read datasource: %v", err)
		}

		if readDS.Type != mybatis.DataSourceTypeSlave {
			t.Errorf("Expected slave datasource for read, got %s", readDS.Type)
		}

		// 测试写操作路由到主库
		writeDS, err := router.GetDataSource(ctx, mybatis.OperationTypeWrite)
		if err != nil {
			t.Fatalf("Failed to get write datasource: %v", err)
		}

		if writeDS.Type != mybatis.DataSourceTypeMaster {
			t.Errorf("Expected master datasource for write, got %s", writeDS.Type)
		}

		// 测试统计信息
		stats := router.GetStats()
		if stats.TotalRequests != 2 {
			t.Errorf("Expected 2 total requests, got %d", stats.TotalRequests)
		}

		if stats.ReadRequests != 1 {
			t.Errorf("Expected 1 read request, got %d", stats.ReadRequests)
		}

		if stats.WriteRequests != 1 {
			t.Errorf("Expected 1 write request, got %d", stats.WriteRequests)
		}
	})

	t.Run("DataSourceRouter_LoadBalance", func(t *testing.T) {
		config := &mybatis.DataSourceConfig{
			ReadWriteSplit: true,
			LoadBalance:    mybatis.LoadBalanceRandom,
		}
		router := mybatis.NewDataSourceRouter(config)
		defer router.Close()

		// 创建多个从库
		for i := 0; i < 3; i++ {
			db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatalf("Failed to create database %d: %v", i, err)
			}

			ds := &mybatis.DataSource{
				Name:     fmt.Sprintf("slave_%d", i),
				Type:     mybatis.DataSourceTypeSlave,
				Weight:   10,
				DB:       db,
				IsActive: true,
			}

			err = router.AddDataSource(ds)
			if err != nil {
				t.Fatalf("Failed to add datasource %d: %v", i, err)
			}
		}

		// 测试负载均衡
		ctx := context.Background()
		selectedSources := make(map[string]int)
		
		for i := 0; i < 30; i++ {
			ds, err := router.GetDataSource(ctx, mybatis.OperationTypeRead)
			if err != nil {
				t.Fatalf("Failed to get datasource: %v", err)
			}
			selectedSources[ds.Name]++
		}

		// 验证负载分布（随机策略下应该相对均匀）
		if len(selectedSources) < 2 {
			t.Error("Load balance should distribute requests across multiple slaves")
		}

		t.Logf("Load balance distribution: %v", selectedSources)
	})

	t.Run("DataSourceRouter_HealthCheck", func(t *testing.T) {
		config := &mybatis.DataSourceConfig{
			ReadWriteSplit: false,
			HealthCheck: &mybatis.HealthCheckConfig{
				Interval:   100 * time.Millisecond,
				Timeout:    50 * time.Millisecond,
				MaxRetries: 1,
				TestSQL:    "SELECT 1",
				Enabled:    true,
			},
		}
		router := mybatis.NewDataSourceRouter(config)
		defer router.Close()

		// 添加一个正常的数据源
		goodDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create good database: %v", err)
		}

		goodDS := &mybatis.DataSource{
			Name:     "good",
			Type:     mybatis.DataSourceTypeMaster,
			Weight:   10,
			DB:       goodDB,
			IsActive: true,
		}

		err = router.AddDataSource(goodDS)
		if err != nil {
			t.Fatalf("Failed to add good datasource: %v", err)
		}

		// 等待健康检查执行
		time.Sleep(200 * time.Millisecond)

		// 验证数据源仍然活跃
		if !goodDS.IsActive {
			t.Error("Good datasource should remain active after health check")
		}
	})

	t.Run("DataSourceRouter_ContextRouting", func(t *testing.T) {
		router := mybatis.NewDataSourceRouter(nil)
		defer router.Close()

		// 添加自定义数据源
		customDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create custom database: %v", err)
		}

		customDS := &mybatis.DataSource{
			Name:     "custom",
			Type:     mybatis.DataSourceTypeCustom,
			Weight:   10,
			DB:       customDB,
			IsActive: true,
		}

		err = router.AddDataSource(customDS)
		if err != nil {
			t.Fatalf("Failed to add custom datasource: %v", err)
		}

		// 测试上下文指定数据源
		ctx := mybatis.SetDataSourceContext(context.Background(), "custom")
		ds, err := router.GetDataSource(ctx, mybatis.OperationTypeRead)
		if err != nil {
			t.Fatalf("Failed to get custom datasource: %v", err)
		}

		if ds.Name != "custom" {
			t.Errorf("Expected custom datasource, got %s", ds.Name)
		}
	})
}

func TestMultiDataSourceSession(t *testing.T) {
	t.Run("MultiDataSourceSession_Basic", func(t *testing.T) {
		// 创建多数据源会话
		session := mybatis.NewMultiDataSourceSessionWithDefaults()
		
		// 创建测试数据库
		masterDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create master database: %v", err)
		}

		slaveDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create slave database: %v", err)
		}

		// 创建测试表
		err = masterDB.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT NOT NULL
			);
			INSERT INTO users (name, email) VALUES ('Master User', 'master@example.com');
		`).Error
		if err != nil {
			t.Fatalf("Failed to setup master database: %v", err)
		}

		err = slaveDB.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT NOT NULL
			);
			INSERT INTO users (name, email) VALUES ('Slave User', 'slave@example.com');
		`).Error
		if err != nil {
			t.Fatalf("Failed to setup slave database: %v", err)
		}

		// 添加数据源
		masterDS := &mybatis.DataSource{
			Name:     "master",
			Type:     mybatis.DataSourceTypeMaster,
			Weight:   10,
			DB:       masterDB,
			IsActive: true,
		}
		
		slaveDS := &mybatis.DataSource{
			Name:     "slave",
			Type:     mybatis.DataSourceTypeSlave,
			Weight:   10,
			DB:       slaveDB,
			IsActive: true,
		}

		err = session.AddDataSource(masterDS)
		if err != nil {
			t.Fatalf("Failed to add master datasource: %v", err)
		}

		err = session.AddDataSource(slaveDS)
		if err != nil {
			t.Fatalf("Failed to add slave datasource: %v", err)
		}

		ctx := context.Background()

		// 测试读操作（应该路由到从库）
		results, err := session.SelectList(ctx, "SELECT name FROM users")
		if err != nil {
			t.Fatalf("Failed to execute read query: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result from slave, got %d", len(results))
		}

		// 测试写操作（应该路由到主库）
		affected, err := session.Insert(ctx, "INSERT INTO users (name, email) VALUES (?, ?)", "New User", "new@example.com")
		if err != nil {
			t.Fatalf("Failed to execute write query: %v", err)
		}

		if affected != 1 {
			t.Errorf("Expected 1 affected row, got %d", affected)
		}

		// 验证统计信息
		stats := session.GetDataSourceStats()
		if stats.ReadRequests < 1 {
			t.Error("Should have at least 1 read request")
		}

		if stats.WriteRequests < 1 {
			t.Error("Should have at least 1 write request")
		}
	})

	t.Run("MultiDataSourceSession_WithSpecificDataSource", func(t *testing.T) {
		session := mybatis.NewMultiDataSourceSessionWithDefaults()
		
		// 创建自定义数据源
		customDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create custom database: %v", err)
		}

		err = customDB.Exec(`
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL
			);
			INSERT INTO users (name) VALUES ('Custom User');
		`).Error
		if err != nil {
			t.Fatalf("Failed to setup custom database: %v", err)
		}

		customDS := &mybatis.DataSource{
			Name:     "custom",
			Type:     mybatis.DataSourceTypeCustom,
			Weight:   10,
			DB:       customDB,
			IsActive: true,
		}

		err = session.AddDataSource(customDS)
		if err != nil {
			t.Fatalf("Failed to add custom datasource: %v", err)
		}

		// 使用指定数据源进行查询
		specificSession := session.WithDataSource("custom")
		ctx := context.Background()

		results, err := specificSession.SelectList(ctx, "SELECT name FROM users")
		if err != nil {
			t.Fatalf("Failed to execute query with specific datasource: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result from custom datasource, got %d", len(results))
		}
	})

	t.Run("MultiDataSourceSession_Configuration", func(t *testing.T) {
		session := mybatis.NewMultiDataSourceSessionWithDefaults()

		// 测试配置方法
		session = session.DisableReadWriteSplit().
			SetLoadBalanceStrategy(mybatis.LoadBalanceWeighted).
			DisableHealthCheck()

		// 验证配置是否生效（通过行为测试）
		// 这里主要测试方法链是否正常工作
		if session == nil {
			t.Error("Configuration chain should return valid session")
		}
	})
}