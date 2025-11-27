package mybatis_test

import (
	"context"
	"testing"

	"github.com/zsy619/yyhertz/framework/mybatis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStoredProcOperations(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建测试表和函数
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_table (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER
		);
		INSERT INTO test_table (name, value) VALUES 
			('test1', 10),
			('test2', 20),
			('test3', 30);
	`).Error
	if err != nil {
		t.Fatalf("Failed to create table and data: %v", err)
	}

	session := mybatis.NewSimpleSession(db)
	ctx := context.Background()

	t.Run("CallStoredProc_Simple", func(t *testing.T) {
		// 测试简单的存储过程调用（SQLite不原生支持存储过程，这里模拟调用场景）
		// 使用参数化查询模拟存储过程
		params := []mybatis.ProcParam{
			{Name: "input_value", Value: 15, Direction: "IN"},
		}

		// 由于SQLite不支持真正的存储过程，我们只测试DryRun模式
		session = session.Debug(true).DryRun(true)
		
		result, err := session.CallStoredProc(ctx, "test_procedure", params)
		if err != nil {
			t.Fatalf("CallStoredProc failed: %v", err)
		}

		// DryRun模式应该返回空结果
		if result == nil {
			t.Error("Expected result to be non-nil")
		}

		if result.RowsAffected != 0 {
			t.Errorf("Expected 0 affected rows in DryRun, got %d", result.RowsAffected)
		}

		if len(result.ResultSets) != 0 {
			t.Errorf("Expected empty result sets in DryRun, got %d", len(result.ResultSets))
		}
	})

	t.Run("CallStoredProcWithMultiResults_Complex", func(t *testing.T) {
		// 测试多结果集存储过程调用
		params := []mybatis.ProcParam{
			{Name: "min_value", Value: 15, Direction: "IN"},
			{Name: "max_value", Value: 25, Direction: "IN"},
			{Name: "output_count", Value: nil, Direction: "OUT"},
		}

		// 使用DryRun模式测试
		session = session.Debug(true).DryRun(true)
		
		result, err := session.CallStoredProcWithMultiResults(ctx, "complex_procedure", params)
		if err != nil {
			t.Fatalf("CallStoredProcWithMultiResults failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result to be non-nil")
		}

		// DryRun模式应该返回空结果
		if result.RowsAffected != 0 {
			t.Errorf("Expected 0 affected rows in DryRun, got %d", result.RowsAffected)
		}
	})

	t.Run("StoredProc_Parameters", func(t *testing.T) {
		// 测试不同类型的参数
		params := []mybatis.ProcParam{
			{Name: "in_param", Value: "test_input", Direction: "IN"},
			{Name: "out_param", Value: nil, Direction: "OUT"},
			{Name: "inout_param", Value: 42, Direction: "INOUT"},
		}

		session = session.Debug(true).DryRun(true)
		
		result, err := session.CallStoredProc(ctx, "param_test_procedure", params)
		if err != nil {
			t.Fatalf("Parameter test failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result to be non-nil")
		}

		// 验证输出参数映射存在（即使是空的）
		if result.OutputParams == nil {
			t.Error("Expected OutputParams map to be initialized")
		}
	})

	t.Run("StoredProc_EmptyParams", func(t *testing.T) {
		// 测试无参数存储过程
		params := []mybatis.ProcParam{}

		session = session.Debug(true).DryRun(true)
		
		result, err := session.CallStoredProc(ctx, "no_param_procedure", params)
		if err != nil {
			t.Fatalf("Empty params test failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result to be non-nil")
		}
	})

	t.Run("StoredProc_InvalidDirection", func(t *testing.T) {
		// 测试无效参数方向
		params := []mybatis.ProcParam{
			{Name: "invalid_param", Value: "test", Direction: "INVALID"},
		}

		session = session.Debug(true).DryRun(true)
		
		_, err := session.CallStoredProc(ctx, "test_procedure", params)
		if err == nil {
			t.Error("Expected error for invalid parameter direction")
		}

		if !contains(err.Error(), "unsupported parameter direction") {
			t.Errorf("Expected 'unsupported parameter direction' error, got: %v", err)
		}
	})
}

func TestXMLSessionStoredProc(t *testing.T) {
	// 设置内存SQLite数据库
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	session := mybatis.NewXMLSession(db)
	ctx := context.Background()

	t.Run("XMLSession_CallStoredProc", func(t *testing.T) {
		params := []mybatis.ProcParam{
			{Name: "test_param", Value: "xml_test", Direction: "IN"},
		}

		// 使用DryRun模式
		session = session.Debug(true).DryRun(true)
		
		result, err := session.CallStoredProc(ctx, "xml_test_procedure", params)
		if err != nil {
			t.Fatalf("XMLSession CallStoredProc failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result to be non-nil")
		}
	})

	t.Run("XMLSession_CallStoredProcWithMultiResults", func(t *testing.T) {
		params := []mybatis.ProcParam{
			{Name: "multi_param", Value: 100, Direction: "IN"},
		}

		session = session.Debug(true).DryRun(true)
		
		result, err := session.CallStoredProcWithMultiResults(ctx, "xml_multi_procedure", params)
		if err != nil {
			t.Fatalf("XMLSession CallStoredProcWithMultiResults failed: %v", err)
		}

		if result == nil {
			t.Error("Expected result to be non-nil")
		}
	})
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		   (s == substr || len(s) >= len(substr) && 
		   	(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		   	 findInString(s, substr)))
}

func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}