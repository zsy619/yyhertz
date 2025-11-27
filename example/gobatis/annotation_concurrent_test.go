package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zsy619/yyhertz/framework/mybatis"
)

// TestAnnotationConcurrent 并发annotation测试
func TestAnnotationConcurrent(t *testing.T) {
	// 创建文件数据库
	db, err := sql.Open("sqlite3", "/tmp/annotation_test.db")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		db.Close()
		// 清理测试数据库
		//os.Remove("/tmp/annotation_test.db")
	}()
	
	// 清理并创建测试表
	db.Exec(`DROP TABLE IF EXISTS testuser`)
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
	concurrency := 10
	operationsPerGoroutine := 5
	
	var wg sync.WaitGroup
	var successCount, failureCount int64
	var mu sync.Mutex
	
	t.Logf("开始并发测试: %d协程 x %d操作", concurrency, operationsPerGoroutine)
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			for j := 0; j < operationsPerGoroutine; j++ {
				user := &TestUser{
					Username: fmt.Sprintf("worker_%d_op_%d", workerID, j),
					Email:    fmt.Sprintf("worker_%d_%d@test.com", workerID, j),
					Age:      20 + j,
					Status:   "active",
				}
				
				id, err := annotationSess.Insert(ctx, user)
				
				mu.Lock()
				if err != nil {
					t.Logf("Worker %d, Operation %d: Insert failed: %v", workerID, j, err)
					failureCount++
				} else {
					t.Logf("Worker %d, Operation %d: Insert success, ID: %d", workerID, j, id)
					successCount++
					
					// 尝试查询验证
					result, err := annotationSess.SelectByID(ctx, &TestUser{}, id)
					if err != nil {
						t.Logf("Worker %d, Operation %d: SelectByID failed: %v", workerID, j, err)
					} else if retrievedUser, ok := result.(*TestUser); ok {
						t.Logf("Worker %d, Operation %d: SelectByID success: %s", workerID, j, retrievedUser.Username)
					}
				}
				mu.Unlock()
			}
		}(i)
	}
	
	wg.Wait()
	
	totalOps := successCount + failureCount
	successRate := float64(successCount) / float64(totalOps) * 100
	
	t.Logf("并发测试完成:")
	t.Logf("- 总操作数: %d", totalOps)
	t.Logf("- 成功操作: %d", successCount)
	t.Logf("- 失败操作: %d", failureCount)
	t.Logf("- 成功率: %.2f%%", successRate)
	
	if successRate < 80.0 {
		t.Errorf("Success rate too low: %.2f%%, expected >= 80%%", successRate)
	}
}