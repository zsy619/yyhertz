package main

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestAnnotationSimpleOperations 简单annotation操作测试
func TestAnnotationSimpleOperations(t *testing.T) {
	// 创建内存数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	// 创建测试表
	_, err = db.Exec(`CREATE TABLE testuser (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username VARCHAR(50) NOT NULL,
		email VARCHAR(100) NOT NULL,
		age INTEGER NOT NULL,
		status VARCHAR(20) NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}
	
	// 创建MockXMLSession
	xmlSession := &MockXMLSession{db: db}
	
	// 创建AnnotationDrivenSession
	annotationSess := mybatis.NewAnnotationDrivenSession(xmlSession)
	
	ctx := context.Background()
	
	// 测试插入
	user := &TestUser{
		Username: "simple_test_user",
		Email:    "simple@test.com",
		Age:      30,
		Status:   "active",
	}
	
	id, err := annotationSess.Insert(ctx, user)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	
	t.Logf("插入成功，ID: %d", id)
	
	// 测试查询
	result, err := annotationSess.SelectByID(ctx, &TestUser{}, id)
	if err != nil {
		t.Fatalf("SelectByID failed: %v", err)
	}
	
	if retrievedUser, ok := result.(*TestUser); ok {
		t.Logf("查询成功 - ID: %d, Username: %s, Email: %s", retrievedUser.ID, retrievedUser.Username, retrievedUser.Email)
		
		// 验证数据
		if retrievedUser.Username != user.Username {
			t.Errorf("Username mismatch: got %s, want %s", retrievedUser.Username, user.Username)
		}
	} else {
		t.Errorf("Invalid result type: %T", result)
	}
	
	// 测试更新
	user.ID = id
	user.Age = 31
	affected, err := annotationSess.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	t.Logf("更新成功，影响行数: %d", affected)
	
	// 验证更新
	result, err = annotationSess.SelectByID(ctx, &TestUser{}, id)
	if err != nil {
		t.Fatalf("SelectByID after update failed: %v", err)
	}
	
	if updatedUser, ok := result.(*TestUser); ok {
		if updatedUser.Age != 31 {
			t.Errorf("Age not updated: got %d, want %d", updatedUser.Age, 31)
		} else {
			t.Log("更新验证成功")
		}
	}
	
	// 测试删除
	affected, err = annotationSess.Delete(ctx, user)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	t.Logf("删除成功，影响行数: %d", affected)
	
	// 验证删除
	_, err = annotationSess.SelectByID(ctx, &TestUser{}, id)
	if err == nil {
		t.Error("Record should be deleted but still found")
	} else {
		t.Log("删除验证成功")
	}
	
	t.Log("所有annotation操作测试通过")
}