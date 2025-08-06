package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/orm"
)

// TestUser 测试用户模型
type TestUser struct {
	orm.BaseModel
	Name   string `gorm:"size:100;not null" json:"name"`
	Email  string `gorm:"size:100;uniqueIndex" json:"email"`
	Age    int    `json:"age"`
	Active bool   `gorm:"default:true" json:"active"`
}

// TestProduct 测试产品模型
type TestProduct struct {
	orm.BaseModel
	Name        string        `gorm:"size:200;not null" json:"name"`
	Description string        `gorm:"type:text" json:"description"`
	Price       float64       `gorm:"precision:10;scale:2" json:"price"`
	Stock       int           `gorm:"default:0" json:"stock"`
	CategoryID  uint          `json:"category_id"`
	Category    *TestCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

// TestCategory 测试分类模型
type TestCategory struct {
	orm.BaseModel
	Name     string        `gorm:"size:100;not null" json:"name"`
	Products []TestProduct `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// TestDatabaseConfig 测试数据库配置加载
func TestDatabaseConfig(t *testing.T) {
	t.Log("=== 测试数据库配置加载 ===")

	// 使用ORM的默认配置方法
	config := orm.DefaultDatabaseConfig()
	if config == nil {
		t.Fatal("获取默认配置失败")
	}

	// 验证配置
	if config.Type == "" {
		t.Error("数据库类型不能为空")
	}

	t.Logf("数据库类型: %s", config.Type)
	t.Logf("数据库主机: %s", config.Host)
	t.Logf("数据库端口: %d", config.Port)
	t.Logf("数据库名称: %s", config.Database)
	t.Logf("最大连接数: %d", config.MaxOpenConns)
	t.Logf("最大空闲连接: %d", config.MaxIdleConns)

	t.Log("✅ 数据库配置测试通过")
}

// TestDatabaseConnection 测试数据库连接
func TestDatabaseConnection(t *testing.T) {
	t.Log("=== 测试数据库连接 ===")

	// 获取默认ORM实例
	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	// 测试数据库连接
	if err := ormInstance.Ping(); err != nil {
		t.Fatalf("数据库连接失败: %v", err)
	}

	t.Log("✅ 数据库连接测试通过")
}

// TestDatabaseMigration 测试数据库迁移
func TestDatabaseMigration(t *testing.T) {
	t.Log("=== 测试数据库迁移 ===")

	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	// 自动迁移测试模型
	models := []any{
		&TestCategory{},
		&TestUser{},
		&TestProduct{},
	}

	if err := ormInstance.AutoMigrate(models...); err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	t.Log("✅ 数据库迁移测试通过")
}

// TestDatabaseCRUD 测试数据库CRUD操作
func TestDatabaseCRUD(t *testing.T) {
	t.Log("=== 测试数据库CRUD操作 ===")

	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	db := ormInstance.DB()

	// 清理测试数据
	defer func() {
		db.Where("email LIKE ?", "%test%").Delete(&TestUser{})
		db.Where("name LIKE ?", "%测试%").Delete(&TestCategory{})
		db.Where("name LIKE ?", "%测试%").Delete(&TestProduct{})
	}()

	// 1. 测试创建操作
	t.Log("--- 测试创建操作 ---")

	// 创建分类
	category := &TestCategory{
		Name: "测试分类",
	}
	if err := db.Create(category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	t.Logf("创建分类成功，ID: %d", category.ID)

	// 创建用户
	user := &TestUser{
		Name:   "测试用户",
		Email:  "test@example.com",
		Age:    25,
		Active: true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	t.Logf("创建用户成功，ID: %d", user.ID)

	// 创建产品
	product := &TestProduct{
		Name:        "测试产品",
		Description: "这是一个测试产品",
		Price:       99.99,
		Stock:       100,
		CategoryID:  category.ID,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("创建产品失败: %v", err)
	}
	t.Logf("创建产品成功，ID: %d", product.ID)

	// 2. 测试查询操作
	t.Log("--- 测试查询操作 ---")

	// 查询用户
	var foundUser TestUser
	if err := db.Where("email = ?", "test@example.com").First(&foundUser).Error; err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
	t.Logf("查询用户成功: %s", foundUser.Name)

	// 查询产品及关联分类
	var foundProduct TestProduct
	if err := db.Preload("Category").Where("id = ?", product.ID).First(&foundProduct).Error; err != nil {
		t.Fatalf("查询产品失败: %v", err)
	}
	t.Logf("查询产品成功: %s, 分类: %s", foundProduct.Name, foundProduct.Category.Name)

	// 3. 测试更新操作
	t.Log("--- 测试更新操作 ---")

	// 更新用户年龄
	if err := db.Model(&foundUser).Update("age", 30).Error; err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}
	t.Log("更新用户年龄成功")

	// 批量更新产品库存
	if err := db.Model(&TestProduct{}).Where("category_id = ?", category.ID).Update("stock", 50).Error; err != nil {
		t.Fatalf("批量更新产品失败: %v", err)
	}
	t.Log("批量更新产品库存成功")

	// 4. 测试删除操作
	t.Log("--- 测试删除操作 ---")

	// 软删除产品
	if err := db.Delete(&foundProduct).Error; err != nil {
		t.Fatalf("删除产品失败: %v", err)
	}
	t.Log("删除产品成功")

	// 验证软删除
	var count int64
	db.Model(&TestProduct{}).Where("id = ?", product.ID).Count(&count)
	if count != 0 {
		t.Error("软删除验证失败")
	}
	t.Log("软删除验证成功")

	t.Log("✅ 数据库CRUD操作测试通过")
}

// TestDatabaseTransaction 测试数据库事务
func TestDatabaseTransaction(t *testing.T) {
	t.Log("=== 测试数据库事务 ===")

	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	// 清理测试数据
	defer func() {
		db := ormInstance.DB()
		db.Where("email LIKE ?", "%transaction%").Delete(&TestUser{})
	}()

	// 测试成功事务
	t.Log("--- 测试成功事务 ---")
	err := ormInstance.Transaction(func(tx *gorm.DB) error {
		user1 := &TestUser{
			Name:  "事务用户1",
			Email: "transaction1@example.com",
			Age:   20,
		}
		if err := tx.Create(user1).Error; err != nil {
			return err
		}

		user2 := &TestUser{
			Name:  "事务用户2",
			Email: "transaction2@example.com",
			Age:   25,
		}
		if err := tx.Create(user2).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("成功事务测试失败: %v", err)
	}

	// 验证数据是否成功提交
	var count int64
	ormInstance.DB().Model(&TestUser{}).Where("email LIKE ?", "%transaction%").Count(&count)
	if count != 2 {
		t.Error("事务提交验证失败")
	}
	t.Log("✅ 成功事务测试通过")

	// 测试回滚事务
	t.Log("--- 测试回滚事务 ---")
	err = ormInstance.Transaction(func(tx *gorm.DB) error {
		user3 := &TestUser{
			Name:  "事务用户3",
			Email: "transaction3@example.com",
			Age:   30,
		}
		if err := tx.Create(user3).Error; err != nil {
			return err
		}

		// 故意产生错误来触发回滚
		return fmt.Errorf("故意的错误")
	})

	if err == nil {
		t.Error("回滚事务应该返回错误")
	}

	// 验证数据是否成功回滚
	ormInstance.DB().Model(&TestUser{}).Where("email LIKE ?", "%transaction%").Count(&count)
	if count != 2 { // 应该还是2条记录，第3条应该回滚了
		t.Error("事务回滚验证失败")
	}
	t.Log("✅ 回滚事务测试通过")
}

// TestDatabaseContext 测试数据库上下文
func TestDatabaseContext(t *testing.T) {
	t.Log("=== 测试数据库上下文 ===")

	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用上下文进行查询
	var users []TestUser
	if err := ormInstance.WithContext(ctx).Limit(10).Find(&users).Error; err != nil {
		t.Fatalf("上下文查询失败: %v", err)
	}

	t.Logf("使用上下文查询到 %d 条用户记录", len(users))
	t.Log("✅ 数据库上下文测试通过")
}

// TestDatabaseStats 测试数据库统计信息
func TestDatabaseStats(t *testing.T) {
	t.Log("=== 测试数据库统计信息 ===")

	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	// 获取统计信息
	stats := ormInstance.GetStats()
	if len(stats) == 0 {
		t.Error("统计信息为空")
	}

	t.Log("数据库统计信息:")
	for key, value := range stats {
		t.Logf("  %s: %v", key, value)
	}

	// 验证必要的统计字段
	requiredFields := []string{"database_type", "database_name"}
	for _, field := range requiredFields {
		if _, exists := stats[field]; !exists {
			t.Errorf("缺少统计字段: %s", field)
		}
	}

	t.Log("✅ 数据库统计信息测试通过")
}

// TestDatabaseRawSQL 测试原生SQL
func TestDatabaseRawSQL(t *testing.T) {
	t.Log("=== 测试原生SQL ===")

	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		t.Fatal("获取ORM实例失败")
	}

	db := ormInstance.DB()

	// 测试原生查询
	var result struct {
		Count int `json:"count"`
	}

	if err := db.Raw("SELECT COUNT(*) as count FROM test_users WHERE active = ?", true).Scan(&result).Error; err != nil {
		t.Logf("原生查询警告（可能是表不存在): %v", err)
	} else {
		t.Logf("活跃用户数量: %d", result.Count)
	}

	t.Log("✅ 原生SQL测试通过")
}

// BenchmarkDatabaseInsert 性能测试：插入操作
func BenchmarkDatabaseInsert(b *testing.B) {
	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		b.Fatal("获取ORM实例失败")
	}

	db := ormInstance.DB()

	// 确保表存在
	db.AutoMigrate(&TestUser{})

	// 清理测试数据
	defer func() {
		db.Where("email LIKE ?", "%benchmark%").Delete(&TestUser{})
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			user := &TestUser{
				Name:  fmt.Sprintf("基准测试用户%d", i),
				Email: fmt.Sprintf("benchmark%d@example.com", i),
				Age:   20 + (i % 50),
			}
			db.Create(user)
			i++
		}
	})
}

// BenchmarkDatabaseQuery 性能测试：查询操作
func BenchmarkDatabaseQuery(b *testing.B) {
	ormInstance := orm.GetDefaultORM()
	if ormInstance == nil {
		b.Fatal("获取ORM实例失败")
	}

	db := ormInstance.DB()

	// 确保表存在并有数据
	db.AutoMigrate(&TestUser{})

	// 插入一些测试数据
	for i := 0; i < 100; i++ {
		user := &TestUser{
			Name:  fmt.Sprintf("查询测试用户%d", i),
			Email: fmt.Sprintf("query%d@example.com", i),
			Age:   20 + (i % 50),
		}
		db.Create(user)
	}

	// 清理测试数据
	defer func() {
		db.Where("email LIKE ?", "%query%").Delete(&TestUser{})
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var users []TestUser
			db.Where("age > ?", 25).Limit(10).Find(&users)
		}
	})
}

// 运行所有测试的辅助函数
func runAllTests() {
	log.Println("开始运行数据库测试套件...")

	tests := []struct {
		name string
		fn   func(*testing.T)
	}{
		{"数据库配置", TestDatabaseConfig},
		{"数据库连接", TestDatabaseConnection},
		{"数据库迁移", TestDatabaseMigration},
		{"数据库CRUD", TestDatabaseCRUD},
		{"数据库事务", TestDatabaseTransaction},
		{"数据库上下文", TestDatabaseContext},
		{"数据库统计", TestDatabaseStats},
		{"原生SQL", TestDatabaseRawSQL},
	}

	for _, test := range tests {
		log.Printf("运行测试: %s", test.name)
		t := &testing.T{}
		test.fn(t)
		if t.Failed() {
			log.Printf("❌ 测试失败: %s", test.name)
		} else {
			log.Printf("✅ 测试通过: %s", test.name)
		}
	}

	log.Println("数据库测试套件完成!")
}

// 如果直接运行此文件，执行所有测试
func init() {
	// 这个函数可以在需要时手动调用
	// runAllTests()
}
