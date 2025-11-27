package main

import (
	"fmt"
	"testing"

	"github.com/zsy619/yyhertz/framework/orm"
)

// SimpleBenchmarkUser 简化的基准测试用户模型
type SimpleBenchmarkUser struct {
	ID     uint   `gorm:"primaryKey"`
	Name   string `gorm:"size:100"`
	Email  string `gorm:"size:100;uniqueIndex"`
	Age    int
	Status string `gorm:"size:20"`
}

// setupSimpleBenchmarkDB 设置简化基准测试数据库
func setupSimpleBenchmarkDB(b *testing.B) *orm.ORM {
	config := &orm.DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		LogLevel:     "silent",
	}

	ormInstance, err := orm.NewORM(config)
	if err != nil {
		b.Fatalf("创建ORM失败: %v", err)
	}

	// 自动迁移
	if err := ormInstance.AutoMigrate(&SimpleBenchmarkUser{}); err != nil {
		b.Fatalf("自动迁移失败: %v", err)
	}

	return ormInstance
}

// BenchmarkSimpleCRUDCreate 简化创建操作基准测试
func BenchmarkSimpleCRUDCreate(b *testing.B) {
	ormInstance := setupSimpleBenchmarkDB(b)
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &SimpleBenchmarkUser{
			Name:   fmt.Sprintf("User%d", i),
			Email:  fmt.Sprintf("user%d@test.com", i),
			Age:    25 + i%50,
			Status: "active",
		}
		if err := crud.Create(user); err != nil {
			b.Fatalf("创建用户失败: %v", err)
		}
	}
}

// BenchmarkSimpleCRUDRead 简化读取操作基准测试
func BenchmarkSimpleCRUDRead(b *testing.B) {
	ormInstance := setupSimpleBenchmarkDB(b)
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	// 准备测试数据
	for i := 0; i < 1000; i++ {
		user := &SimpleBenchmarkUser{
			Name:   fmt.Sprintf("TestUser%d", i),
			Email:  fmt.Sprintf("testuser%d@test.com", i),
			Age:    25 + i%50,
			Status: "active",
		}
		crud.Create(user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user SimpleBenchmarkUser
		id := uint(i%1000 + 1)
		if err := crud.First(&user, id); err != nil {
			b.Fatalf("查询用户失败: %v", err)
		}
	}
}

// BenchmarkSimpleCRUDUpdate 简化更新操作基准测试
func BenchmarkSimpleCRUDUpdate(b *testing.B) {
	ormInstance := setupSimpleBenchmarkDB(b)
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	// 准备测试数据
	var users []SimpleBenchmarkUser
	for i := 0; i < b.N; i++ {
		user := &SimpleBenchmarkUser{
			Name:   fmt.Sprintf("UpdateUser%d", i),
			Email:  fmt.Sprintf("updateuser%d@test.com", i),
			Age:    25,
			Status: "active",
		}
		crud.Create(user)
		users = append(users, *user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user := &users[i]
		if err := crud.Update(user, "age", 30+i%20); err != nil {
			b.Fatalf("更新用户失败: %v", err)
		}
	}
}

// BenchmarkSimpleQuery 简化查询操作基准测试
func BenchmarkSimpleQuery(b *testing.B) {
	ormInstance := setupSimpleBenchmarkDB(b)
	crud := orm.NewSimpleCRUD(ormInstance.DB())

	// 准备测试数据
	statuses := []string{"active", "inactive", "pending"}
	for i := 0; i < 1000; i++ {
		user := &SimpleBenchmarkUser{
			Name:   fmt.Sprintf("QueryUser%d", i),
			Email:  fmt.Sprintf("queryuser%d@test.com", i),
			Age:    25 + i%50,
			Status: statuses[i%3],
		}
		crud.Create(user)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []SimpleBenchmarkUser
		status := statuses[i%3]
		if err := crud.Find(&users, "status = ?", status); err != nil {
			b.Fatalf("查询用户失败: %v", err)
		}
	}
}