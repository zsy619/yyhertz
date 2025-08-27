package main

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestAnnotationBasicDebug 基础annotation调试测试
func TestAnnotationBasicDebug(t *testing.T) {
	// 创建内存数据库
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	// 创建测试表（使用annotation解析器生成的表名）
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
	
	// 测试TagParser
	parser := mybatis.NewTagParser()
	user := &TestUser{}
	
	tableInfo, err := parser.ParseStruct(user)
	if err != nil {
		t.Fatalf("ParseStruct failed: %v", err)
	}
	
	t.Logf("解析结果 - 表名: %s", tableInfo.Name)
	t.Logf("解析结果 - 字段数: %d", len(tableInfo.Columns))
	for fieldName, columnInfo := range tableInfo.Columns {
		t.Logf("字段: %s -> 列: %s (主键: %t, 自增: %t)", fieldName, columnInfo.ColumnName, columnInfo.IsPrimaryKey, columnInfo.IsAutoIncr)
	}
	
	// 测试SQL生成器
	sqlGen := mybatis.NewSQLGenerator()
	
	testUser := &TestUser{
		Username: "debug_user",
		Email:    "debug@test.com",
		Age:      25,
		Status:   "active",
	}
	
	insertSQL, args, err := sqlGen.GenerateInsertSQL(testUser)
	if err != nil {
		t.Fatalf("GenerateInsertSQL failed: %v", err)
	}
	
	t.Logf("生成的INSERT SQL: %s", insertSQL)
	t.Logf("参数: %v", args)
	
	// 直接执行SQL测试
	result, err := db.ExecContext(context.Background(), insertSQL, args...)
	if err != nil {
		t.Fatalf("Execute INSERT failed: %v", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("LastInsertId failed: %v", err)
	}
	
	t.Logf("插入成功，ID: %d", id)
	
	// 测试查询SQL生成
	selectSQL, err := sqlGen.GenerateSelectSQL(testUser, "id = ?")
	if err != nil {
		t.Fatalf("GenerateSelectSQL failed: %v", err)
	}
	
	t.Logf("生成的SELECT SQL: %s", selectSQL)
	
	// 执行查询测试  
	row := db.QueryRowContext(context.Background(), selectSQL, id)
	var retrievedUser TestUser
	// 根据最新生成的SQL字段顺序进行Scan: id, username, email, age, status
	err = row.Scan(&retrievedUser.ID, &retrievedUser.Username, &retrievedUser.Email, &retrievedUser.Age, &retrievedUser.Status)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	
	t.Logf("查询成功 - ID: %d, Username: %s, Email: %s", retrievedUser.ID, retrievedUser.Username, retrievedUser.Email)
	
	// 验证数据正确性
	if retrievedUser.Username != testUser.Username {
		t.Errorf("Username mismatch: got %s, want %s", retrievedUser.Username, testUser.Username)
	}
	if retrievedUser.Email != testUser.Email {
		t.Errorf("Email mismatch: got %s, want %s", retrievedUser.Email, testUser.Email)
	}
	
	t.Log("基础annotation功能验证成功")
}