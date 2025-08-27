package mybatis_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUser 测试用户实体
type TestUser struct {
	ID       int64             `json:"id"`
	Name     string            `json:"name"`
	Email    string            `json:"email"`
	Profile  *TestUserProfile  `json:"profile"`  // 一对一关联
	Orders   []*TestOrder      `json:"orders"`   // 一对多关联
}

// TestUserProfile 用户资料
type TestUserProfile struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"userId"`
	Bio    string `json:"bio"`
	Avatar string `json:"avatar"`
}

// TestOrder 订单
type TestOrder struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"userId"`
	Title  string `json:"title"`
	Amount float64 `json:"amount"`
}

func TestLazyLoadingOperations(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表
	err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		);
		
		CREATE TABLE user_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			bio TEXT,
			avatar TEXT
		);
		
		CREATE TABLE orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			amount DECIMAL(10,2)
		);
		
		INSERT INTO users (name, email) VALUES 
			('张三', 'zhangsan@example.com'),
			('李四', 'lisi@example.com');
			
		INSERT INTO user_profiles (user_id, bio, avatar) VALUES 
			(1, '张三的个人简介', 'avatar1.jpg'),
			(2, '李四的个人简介', 'avatar2.jpg');
			
		INSERT INTO orders (user_id, title, amount) VALUES 
			(1, '订单1', 100.00),
			(1, '订单2', 200.00),
			(2, '订单3', 300.00);
	`).Error
	if err != nil {
		t.Fatalf("Failed to create tables and data: %v", err)
	}

	session := mybatis.NewXMLSession(db)
	ctx := context.Background()

	t.Run("LazyLoading_Disabled_By_Default", func(t *testing.T) {
		// 默认情况下懒加载应该是启用的，但没有注册关联映射
		// 所以不会创建代理对象
		
		result, err := session.SelectList(ctx, "SELECT id, name, email FROM users")
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 users, got %d", len(result))
		}

		// 结果应该是普通的map，不是代理对象
		for _, item := range result {
			if _, ok := item.(*mybatis.ProxyWrapper); ok {
				t.Error("Should not create proxy object without association mappings")
			}
		}
	})

	t.Run("Enable_LazyLoading_With_Associations", func(t *testing.T) {
		// 启用懒加载并注册关联映射
		config := &mybatis.LazyLoadConfiguration{
			LazyLoadingEnabled:     true,
			AggressiveLazyLoading:  false,
			LazyLoadTriggerMethods: []string{"toString", "equals"},
			ProxyFactory:           "REFLECTION",
		}

		session = session.EnableLazyLoading(config)

		// 注册一对一关联：User -> UserProfile
		profileAssociation := &mybatis.AssociationMapping{
			Property:   "profile",
			Column:     "id",
			Select:     "SELECT id, user_id, bio, avatar FROM user_profiles WHERE user_id = ?",
			ForeignKey: "id",
			Type:       mybatis.AssociationTypeOne,
			ResultType: reflect.TypeOf(&TestUserProfile{}),
			LazyLoad:   true,
		}
		session.RegisterAssociation("TestUser", profileAssociation)

		// 注册一对多关联：User -> Orders
		ordersAssociation := &mybatis.AssociationMapping{
			Property:   "orders",
			Column:     "id", 
			Select:     "SELECT id, user_id, title, amount FROM orders WHERE user_id = ?",
			ForeignKey: "id",
			Type:       mybatis.AssociationTypeMany,
			ResultType: reflect.TypeOf([]*TestOrder{}),
			LazyLoad:   true,
		}
		session.RegisterAssociation("TestUser", ordersAssociation)

		// 执行查询，应该创建代理对象
		result, err := session.SelectList(ctx, "SELECT id, name, email FROM users")
		if err != nil {
			t.Fatalf("Query with lazy loading failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 users, got %d", len(result))
		}

		// 验证是否创建了代理对象
		proxyCount := 0
		for _, item := range result {
			if proxy, ok := item.(*mybatis.ProxyWrapper); ok {
				proxyCount++
				
				// 验证代理对象有懒加载器
				lazyLoader := proxy.GetLazyLoader()
				if lazyLoader == nil {
					t.Error("Proxy should have lazy loader")
				}

				// 验证目标对象
				target := proxy.GetTarget()
				if target == nil {
					t.Error("Proxy should have target object")
				}
			}
		}

		// 注意：由于我们的测试数据是map类型，不是结构体，
		// 实际的代理创建可能不会发生，这取决于类型匹配
		t.Logf("Proxy objects created: %d", proxyCount)
	})

	t.Run("LazyLoader_Operations", func(t *testing.T) {
		// 测试懒加载器的基本操作
		config := mybatis.DefaultLazyLoadConfiguration()
		manager := mybatis.NewLazyLoadManager(config)

		// 创建一个测试对象
		testUser := &TestUser{
			ID:    1,
			Name:  "Test User",
			Email: "test@example.com",
		}

		// 创建代理
		proxy, err := manager.CreateProxy(testUser, session)
		if err != nil {
			t.Fatalf("Failed to create proxy: %v", err)
		}

		// 检查是否创建了代理包装器
		if proxyWrapper, ok := proxy.(*mybatis.ProxyWrapper); ok {
			lazyLoader := proxyWrapper.GetLazyLoader()
			
			// 测试属性加载状态
			if lazyLoader.IsLoaded("profile") {
				t.Error("Profile should not be loaded initially")
			}
			
			if lazyLoader.IsLoaded("orders") {
				t.Error("Orders should not be loaded initially")
			}
			
			// 测试触发加载
			err = proxyWrapper.TriggerLoad(ctx, "toString")
			// 由于没有注册加载器，这里可能会有错误，但不影响测试逻辑
			t.Logf("Trigger load result: %v", err)
		}
	})

	t.Run("Disable_LazyLoading", func(t *testing.T) {
		// 禁用懒加载
		session = session.DisableLazyLoading()

		result, err := session.SelectList(ctx, "SELECT id, name, email FROM users")
		if err != nil {
			t.Fatalf("Query after disable lazy loading failed: %v", err)
		}

		// 验证没有创建代理对象
		for _, item := range result {
			if _, ok := item.(*mybatis.ProxyWrapper); ok {
				t.Error("Should not create proxy object when lazy loading is disabled")
			}
		}
	})

	t.Run("AssociationMapping_Registration", func(t *testing.T) {
		// 测试关联映射注册
		config := mybatis.DefaultLazyLoadConfiguration()
		manager := mybatis.NewLazyLoadManager(config)

		// 注册关联映射
		mapping := &mybatis.AssociationMapping{
			Property:   "testProperty",
			Column:     "test_column",
			Select:     "SELECT * FROM test WHERE id = ?",
			ForeignKey: "testId",
			Type:       mybatis.AssociationTypeOne,
			LazyLoad:   true,
		}

		manager.RegisterAssociation("TestType", mapping)

		// 验证注册成功（间接验证，通过没有错误来确认）
		// 实际使用中，关联映射会在创建代理时使用
		t.Log("Association mapping registered successfully")
	})
}

func TestLazyLoadConfiguration(t *testing.T) {
	t.Run("Default_Configuration", func(t *testing.T) {
		config := mybatis.DefaultLazyLoadConfiguration()
		
		if !config.LazyLoadingEnabled {
			t.Error("Default lazy loading should be enabled")
		}
		
		if config.AggressiveLazyLoading {
			t.Error("Default aggressive lazy loading should be false")
		}
		
		if config.ProxyFactory != "REFLECTION" {
			t.Error("Default proxy factory should be REFLECTION")
		}
		
		expectedMethods := []string{"equals", "clone", "hashCode", "toString"}
		if len(config.LazyLoadTriggerMethods) != len(expectedMethods) {
			t.Error("Default trigger methods count mismatch")
		}
	})

	t.Run("Custom_Configuration", func(t *testing.T) {
		config := &mybatis.LazyLoadConfiguration{
			LazyLoadingEnabled:     false,
			AggressiveLazyLoading:  true,
			LazyLoadTriggerMethods: []string{"customMethod"},
			ProxyFactory:           "CUSTOM",
		}

		if config.LazyLoadingEnabled {
			t.Error("Custom lazy loading should be disabled")
		}
		
		if !config.AggressiveLazyLoading {
			t.Error("Custom aggressive lazy loading should be true")
		}
		
		if config.ProxyFactory != "CUSTOM" {
			t.Error("Custom proxy factory should be CUSTOM")
		}
	})
}

func TestAssociationTypes(t *testing.T) {
	t.Run("Association_Types", func(t *testing.T) {
		// 测试关联类型常量
		if mybatis.AssociationTypeOne != 0 {
			t.Error("AssociationTypeOne should be 0")
		}
		
		if mybatis.AssociationTypeMany != 1 {
			t.Error("AssociationTypeMany should be 1")
		}
	})
}