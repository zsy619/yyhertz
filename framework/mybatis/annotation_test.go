package mybatis_test

import (
	"context"
	"strings"
	"testing"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// AnnotationUser 使用注解的用户实体
type AnnotationUser struct {
	ID       int64  `column:"id" pk:"true" auto_incr:"true"`
	Name     string `column:"name" validate:"required,min=2,max=50"`
	Email    string `column:"email" validate:"required,email"`
	Age      int    `column:"age" default:"0"`
	Status   string `column:"status" default:"active"`
	IsActive bool   `column:"is_active" default:"true"`
	Ignored  string `ignore:"true"` // 这个字段会被忽略
}

// AnnotationUserWithTable 带表名注解的用户实体
type AnnotationUserWithTable struct {
	TableName string `table:"custom_users"` // 表名标签
	ID        int64  `column:"user_id" pk:"true" auto_incr:"true"`
	Username  string `column:"username"`
	Password  string `column:"password"`
}

// AnnotationUserWithAssociation 带关联的用户实体
type AnnotationUserWithAssociation struct {
	ID      int64                       `column:"id" pk:"true"`
	Name    string                      `column:"name"`
	Profile *AnnotationUserProfile      `association:"select=SELECT * FROM user_profiles WHERE user_id = ?,column=id,foreignKey=id" lazy:"true"`
	Orders  []*AnnotationOrder          `collection:"select=SELECT * FROM orders WHERE user_id = ?,column=id,foreignKey=id,ofType=AnnotationOrder" lazy:"true"`
}

// AnnotationUserProfile 用户资料实体
type AnnotationUserProfile struct {
	ID     int64  `column:"id" pk:"true"`
	UserID int64  `column:"user_id"`
	Bio    string `column:"bio"`
}

// AnnotationOrder 订单实体
type AnnotationOrder struct {
	ID     int64   `column:"id" pk:"true"`
	UserID int64   `column:"user_id"`
	Title  string  `column:"title"`
	Amount float64 `column:"amount"`
}

func TestAnnotationSystem(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表
	err = db.Exec(`
		CREATE TABLE annotationuser (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			age INTEGER DEFAULT 0,
			status TEXT DEFAULT 'active',
			is_active BOOLEAN DEFAULT 1
		);
		
		CREATE TABLE custom_users (
			user_id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			password TEXT NOT NULL
		);
	`).Error
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}

	xmlSession := mybatis.NewXMLSession(db)
	session := mybatis.NewAnnotationDrivenSession(xmlSession)
	ctx := context.Background()

	t.Run("TagParser_ParseStruct", func(t *testing.T) {
		parser := mybatis.NewTagParser()
		
		// 解析基本用户结构
		user := &AnnotationUser{}
		tableInfo, err := parser.ParseStruct(user)
		if err != nil {
			t.Fatalf("Failed to parse struct: %v", err)
		}

		// 验证表名
		expectedTableName := "annotationuser"
		if tableInfo.Name != expectedTableName {
			t.Errorf("Expected table name %s, got %s", expectedTableName, tableInfo.Name)
		}

		// 验证字段数量（忽略 Ignored 字段）
		expectedColumns := 6 // ID, Name, Email, Age, Status, IsActive
		if len(tableInfo.Columns) != expectedColumns {
			t.Errorf("Expected %d columns, got %d", expectedColumns, len(tableInfo.Columns))
		}

		// 验证主键
		if len(tableInfo.PrimaryKeys) != 1 || tableInfo.PrimaryKeys[0] != "id" {
			t.Errorf("Expected primary key 'id', got %v", tableInfo.PrimaryKeys)
		}

		// 验证自增字段
		idColumn := tableInfo.Columns["ID"]
		if idColumn == nil || !idColumn.IsAutoIncr {
			t.Error("ID field should be auto increment")
		}

		// 验证默认值
		ageColumn := tableInfo.Columns["Age"]
		if ageColumn == nil || ageColumn.DefaultValue == nil {
			t.Error("Age field should have default value")
		}

		// 验证忽略字段不存在
		if _, exists := tableInfo.Columns["Ignored"]; exists {
			t.Error("Ignored field should not be in columns")
		}
	})

	t.Run("TagParser_CustomTableName", func(t *testing.T) {
		parser := mybatis.NewTagParser()
		
		userWithTable := &AnnotationUserWithTable{}
		tableInfo, err := parser.ParseStruct(userWithTable)
		if err != nil {
			t.Fatalf("Failed to parse struct with table tag: %v", err)
		}

		expectedTableName := "custom_users"
		if tableInfo.Name != expectedTableName {
			t.Errorf("Expected table name %s, got %s", expectedTableName, tableInfo.Name)
		}

		// 验证自定义列名
		idColumn := tableInfo.Columns["ID"]
		if idColumn == nil || idColumn.ColumnName != "user_id" {
			t.Error("ID field should map to user_id column")
		}
	})

	t.Run("SQLGenerator_Insert", func(t *testing.T) {
		generator := mybatis.NewSQLGenerator()
		
		user := &AnnotationUser{
			Name:     "张三",
			Email:    "zhangsan@example.com",
			Age:      25,
			Status:   "active",
			IsActive: true,
		}

		sql, args, err := generator.GenerateInsertSQL(user)
		if err != nil {
			t.Fatalf("Failed to generate insert SQL: %v", err)
		}

		t.Logf("Generated INSERT SQL: %s", sql)
		t.Logf("Args: %v", args)

		// 验证SQL格式
		if !annotationTestContains(sql, "INSERT INTO") || !annotationTestContains(sql, "annotationuser") {
			t.Error("Generated SQL should be a valid INSERT statement")
		}

		// 验证参数数量（不包括自增ID）
		expectedArgCount := 5 // name, email, age, status, is_active
		if len(args) != expectedArgCount {
			t.Errorf("Expected %d arguments, got %d", expectedArgCount, len(args))
		}
	})

	t.Run("SQLGenerator_Update", func(t *testing.T) {
		generator := mybatis.NewSQLGenerator()
		
		user := &AnnotationUser{
			ID:       1,
			Name:     "李四",
			Email:    "lisi@example.com",
			Age:      30,
			Status:   "inactive",
			IsActive: false,
		}

		sql, args, err := generator.GenerateUpdateSQL(user)
		if err != nil {
			t.Fatalf("Failed to generate update SQL: %v", err)
		}

		t.Logf("Generated UPDATE SQL: %s", sql)
		t.Logf("Args: %v", args)

		// 验证SQL格式
		if !annotationTestContains(sql, "UPDATE") || !annotationTestContains(sql, "SET") || !annotationTestContains(sql, "WHERE") {
			t.Error("Generated SQL should be a valid UPDATE statement")
		}

		// 验证包含主键条件
		if !annotationTestContains(sql, "id = ?") {
			t.Error("UPDATE SQL should include primary key condition")
		}
	})

	t.Run("SQLGenerator_Select", func(t *testing.T) {
		generator := mybatis.NewSQLGenerator()
		
		user := &AnnotationUser{}
		sql, err := generator.GenerateSelectSQL(user, "age > 18")
		if err != nil {
			t.Fatalf("Failed to generate select SQL: %v", err)
		}

		t.Logf("Generated SELECT SQL: %s", sql)

		// 验证SQL格式
		if !annotationTestContains(sql, "SELECT") || !annotationTestContains(sql, "FROM") || !annotationTestContains(sql, "WHERE") {
			t.Error("Generated SQL should be a valid SELECT statement")
		}

		// 验证包含条件
		if !annotationTestContains(sql, "age > 18") {
			t.Error("SELECT SQL should include condition")
		}
	})

	t.Run("SQLGenerator_Delete", func(t *testing.T) {
		generator := mybatis.NewSQLGenerator()
		
		user := &AnnotationUser{
			ID: 1,
		}

		sql, args, err := generator.GenerateDeleteSQL(user)
		if err != nil {
			t.Fatalf("Failed to generate delete SQL: %v", err)
		}

		t.Logf("Generated DELETE SQL: %s", sql)
		t.Logf("Args: %v", args)

		// 验证SQL格式
		if !annotationTestContains(sql, "DELETE FROM") || !annotationTestContains(sql, "WHERE") {
			t.Error("Generated SQL should be a valid DELETE statement")
		}

		// 验证包含主键条件
		if !annotationTestContains(sql, "id = ?") {
			t.Error("DELETE SQL should include primary key condition")
		}
	})

	t.Run("AnnotationDrivenSession_CRUD", func(t *testing.T) {
		// 测试插入
		user := &AnnotationUser{
			Name:     "王五",
			Email:    "wangwu@example.com",
			Age:      28,
			Status:   "active",
			IsActive: true,
		}

		insertResult, err := session.Insert(ctx, user)
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}

		if insertResult != 1 {
			t.Errorf("Expected 1 affected row, got %d", insertResult)
		}

		// 测试查询所有
		allUsers, err := session.SelectAll(ctx, &AnnotationUser{})
		if err != nil {
			t.Fatalf("Failed to select all users: %v", err)
		}

		if len(allUsers) != 1 {
			t.Errorf("Expected 1 user, got %d", len(allUsers))
		}

		// 测试按ID查询
		foundUser, err := session.SelectByID(ctx, &AnnotationUser{}, 1)
		if err != nil {
			t.Fatalf("Failed to select user by ID: %v", err)
		}

		if foundUser == nil {
			t.Error("Should find user by ID")
		}

		// 测试条件查询
		activeUsers, err := session.SelectWhere(ctx, &AnnotationUser{}, "status = ?", "active")
		if err != nil {
			t.Fatalf("Failed to select users by condition: %v", err)
		}

		if len(activeUsers) != 1 {
			t.Errorf("Expected 1 active user, got %d", len(activeUsers))
		}

		// 测试更新
		updateUser := &AnnotationUser{
			ID:       1,
			Name:     "王五修改",
			Email:    "wangwu_updated@example.com",
			Age:      29,
			Status:   "inactive",
			IsActive: false,
		}

		updateResult, err := session.Update(ctx, updateUser)
		if err != nil {
			t.Fatalf("Failed to update user: %v", err)
		}

		if updateResult != 1 {
			t.Errorf("Expected 1 updated row, got %d", updateResult)
		}

		// 测试删除
		deleteResult, err := session.Delete(ctx, &AnnotationUser{ID: 1})
		if err != nil {
			t.Fatalf("Failed to delete user: %v", err)
		}

		if deleteResult != 1 {
			t.Errorf("Expected 1 deleted row, got %d", deleteResult)
		}

		// 验证删除成功
		remainingUsers, err := session.SelectAll(ctx, &AnnotationUser{})
		if err != nil {
			t.Fatalf("Failed to verify deletion: %v", err)
		}

		if len(remainingUsers) != 0 {
			t.Errorf("Expected 0 users after deletion, got %d", len(remainingUsers))
		}
	})
}

func TestAssociationAnnotations(t *testing.T) {
	t.Run("Association_Parsing", func(t *testing.T) {
		parser := mybatis.NewTagParser()
		
		user := &AnnotationUserWithAssociation{}
		tableInfo, err := parser.ParseStruct(user)
		if err != nil {
			t.Fatalf("Failed to parse struct with associations: %v", err)
		}

		// 验证关联信息
		if len(tableInfo.Associations) != 1 {
			t.Errorf("Expected 1 association, got %d", len(tableInfo.Associations))
		}

		if len(tableInfo.Collections) != 1 {
			t.Errorf("Expected 1 collection, got %d", len(tableInfo.Collections))
		}

		// 验证关联配置
		profileAssoc := tableInfo.Associations[0]
		if profileAssoc.PropertyName != "Profile" {
			t.Errorf("Expected association property 'Profile', got %s", profileAssoc.PropertyName)
		}

		if !profileAssoc.LazyLoad {
			t.Error("Profile association should be lazy loaded")
		}

		// 验证集合配置
		ordersCollection := tableInfo.Collections[0]
		if ordersCollection.PropertyName != "Orders" {
			t.Errorf("Expected collection property 'Orders', got %s", ordersCollection.PropertyName)
		}

		if ordersCollection.OfType != "AnnotationOrder" {
			t.Errorf("Expected collection ofType 'AnnotationOrder', got %s", ordersCollection.OfType)
		}
	})
}

// annotationTestContains 辅助函数，避免与其他测试文件中的contains函数冲突
func annotationTestContains(s, substr string) bool {
	return strings.Contains(s, substr)
}