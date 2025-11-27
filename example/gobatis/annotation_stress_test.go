package main

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/mybatis"
	mapper "github.com/zsy619/yyhertz/framework/mybatis/mapper"
)

// TestUser annotation测试用户实体
type TestUser struct {
	ID       int64            `column:"id" pk:"true" auto_incr:"true"`
	Username string           `column:"username" validate:"required,min=3,max=50"`
	Email    string           `column:"email" validate:"required,email"`
	Age      int              `column:"age" validate:"min=0,max=120"`
	Status   string           `column:"status" default:"active"`
	Profile  *TestUserProfile `association:"select=selectUserProfile,column=id,foreignKey=user_id" lazy:"true" cache:"true"`
	Orders   []*TestOrder     `collection:"select=selectUserOrders,column=id,foreignKey=user_id,ofType=Order" lazy:"true" cache:"true"`
}

// TestUserProfile 测试用户资料实体
type TestUserProfile struct {
	ID          int64  `column:"id" pk:"true" auto_incr:"true"`
	UserID      int64  `column:"user_id"`
	Avatar      string `column:"avatar"`
	Description string `column:"description"`
}

// TestOrder 测试订单实体
type TestOrder struct {
	ID     int64   `column:"id" pk:"true" auto_incr:"true"`
	UserID int64   `column:"user_id"`
	Amount float64 `column:"amount"`
	Status string  `column:"status"`
}

// AnnotationStressTester annotation压力测试器
type AnnotationStressTester struct {
	db             *sql.DB
	gormDB         *gorm.DB
	annotationSess *mybatis.AnnotationDrivenSession
	stats          *AnnotationStressStats
}

// AnnotationStressStats annotation压力测试统计
type AnnotationStressStats struct {
	// 操作统计
	TotalOperations   int64 `json:"total_operations"`
	SuccessOperations int64 `json:"success_operations"`
	FailedOperations  int64 `json:"failed_operations"`

	// 性能统计
	TotalLatency int64 `json:"total_latency_ms"`
	MinLatency   int64 `json:"min_latency_ms"`
	MaxLatency   int64 `json:"max_latency_ms"`

	// 并发统计
	ActiveGoroutines int64 `json:"active_goroutines"`
	MaxGoroutines    int64 `json:"max_goroutines"`

	// 缓存统计
	CacheHits   int64 `json:"cache_hits"`
	CacheMisses int64 `json:"cache_misses"`

	// 资源统计
	MemoryUsageBytes int64 `json:"memory_usage_bytes"`
	MaxMemoryBytes   int64 `json:"max_memory_bytes"`

	// 错误统计
	ParseErrors         int64 `json:"parse_errors"`
	SQLGenerationErrors int64 `json:"sql_generation_errors"`
	DatabaseErrors      int64 `json:"database_errors"`

	mutex sync.RWMutex
}

// NewAnnotationStressTester 创建annotation压力测试器
func NewAnnotationStressTester(dbPath string) (*AnnotationStressTester, error) {
	// 初始化SQLite数据库
	var connString string
	if dbPath == ":memory:" {
		// 对内存数据库使用共享缓存
		connString = "file::memory:?mode=memory&cache=shared"
	} else {
		connString = dbPath + "?cache=shared&mode=rwc"
	}

	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(500)
	db.SetMaxIdleConns(100)
	db.SetConnMaxLifetime(30 * time.Minute)

	// 初始化GORM
	gormDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize GORM: %w", err)
	}

	// 创建测试表
	if err := createAnnotationTestTables(db); err != nil {
		return nil, fmt.Errorf("failed to create test tables: %w", err)
	}

	// 创建XML会话（模拟）
	xmlSession := &MockXMLSession{db: db}

	// 创建annotation驱动会话
	annotationSess := mybatis.NewAnnotationDrivenSession(xmlSession)

	tester := &AnnotationStressTester{
		db:             db,
		gormDB:         gormDB,
		annotationSess: annotationSess,
		stats: &AnnotationStressStats{
			MinLatency: 999999,
		},
	}

	return tester, nil
}

// createAnnotationTestTables 创建annotation测试表
func createAnnotationTestTables(db *sql.DB) error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS testuser (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username VARCHAR(50) NOT NULL,
			email VARCHAR(100) NOT NULL,
			age INTEGER NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS testuserprofile (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			avatar VARCHAR(255),
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES testuser(id)
		)`,
		`CREATE TABLE IF NOT EXISTS testorder (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES testuser(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_testuser_email ON testuser(email)`,
		`CREATE INDEX IF NOT EXISTS idx_testuserprofile_user_id ON testuserprofile(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_testorder_user_id ON testorder(user_id)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to create table: %w", err)
		}
	}

	return nil
}

// MockXMLSession 模拟XML会话实现
type MockXMLSession struct {
	db *sql.DB
}

// 实现XMLSession接口的基础方法

func (m *MockXMLSession) Insert(ctx context.Context, sql string, args ...any) (int64, error) {
	result, err := m.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (m *MockXMLSession) Update(ctx context.Context, sql string, args ...any) (int64, error) {
	result, err := m.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (m *MockXMLSession) Delete(ctx context.Context, sql string, args ...any) (int64, error) {
	result, err := m.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (m *MockXMLSession) SelectOne(ctx context.Context, sql string, args ...any) (any, error) {
	// 先查询以确定列的顺序
	rows, err := m.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no rows found")
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// 创建值切片
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		return nil, err
	}

	// 创建用户对象并根据列名赋值
	user := &TestUser{}
	for i, colName := range columns {
		switch colName {
		case "id":
			if val, ok := values[i].(int64); ok {
				user.ID = val
			}
		case "username":
			if val, ok := values[i].(string); ok {
				user.Username = val
			}
		case "email":
			if val, ok := values[i].(string); ok {
				user.Email = val
			}
		case "age":
			if val, ok := values[i].(int64); ok {
				user.Age = int(val)
			}
		case "status":
			if val, ok := values[i].(string); ok {
				user.Status = val
			}
		}
	}

	return user, nil
}

func (m *MockXMLSession) SelectList(ctx context.Context, sql string, args ...any) ([]any, error) {
	rows, err := m.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []any
	for rows.Next() {
		// 创建值切片
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		err = rows.Scan(valuePtrs...)
		if err != nil {
			continue
		}

		// 创建用户对象并根据列名赋值
		user := &TestUser{}
		for i, colName := range columns {
			switch colName {
			case "id":
				if val, ok := values[i].(int64); ok {
					user.ID = val
				}
			case "username":
				if val, ok := values[i].(string); ok {
					user.Username = val
				}
			case "email":
				if val, ok := values[i].(string); ok {
					user.Email = val
				}
			case "age":
				if val, ok := values[i].(int64); ok {
					user.Age = int(val)
				}
			case "status":
				if val, ok := values[i].(string); ok {
					user.Status = val
				}
			}
		}

		results = append(results, user)
	}

	return results, nil
}

// 添加其他XMLSession接口方法的空实现
func (m *MockXMLSession) SelectPage(ctx context.Context, sql string, page mybatis.PageRequest, args ...any) (*mybatis.PageResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) BatchInsert(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) BatchUpdate(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) BatchDelete(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) CallStoredProc(ctx context.Context, procName string, params []mybatis.ProcParam) (*mybatis.StoredProcResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) CallStoredProcWithMultiResults(ctx context.Context, procName string, params []mybatis.ProcParam) (*mybatis.StoredProcResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) Debug(enabled bool) mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) DryRun(enabled bool) mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) AddBeforeHook(hook mybatis.BeforeHook) mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) AddAfterHook(hook mybatis.AfterHook) mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) EnableCache(config *mybatis.CacheConfig) mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) DisableCache() mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) ClearCache() mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) GetCacheStats() map[string]any {
	return make(map[string]any)
}

func (m *MockXMLSession) BatchDeleteByID(ctx context.Context, statementId string, parameters []any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) RegisterAssociation(typeName string, mapping *mybatis.AssociationMapping) mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) BatchInsertByID(ctx context.Context, statementId string, parameters []any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) BatchUpdateByID(ctx context.Context, statementId string, parameters []any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) EnableLazyLoading(config *mybatis.LazyLoadConfiguration) mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) DisableLazyLoading() mybatis.XMLSession {
	return m
}

func (m *MockXMLSession) LoadMapperXML(xmlPath string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockXMLSession) LoadMapperXMLFromString(xmlContent string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockXMLSession) LoadMapperDirectory(dirPath string) error {
	return fmt.Errorf("not implemented")
}

func (m *MockXMLSession) SelectOneByID(ctx context.Context, statementId string, parameter any) (any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) SelectListByID(ctx context.Context, statementId string, parameter any) ([]any, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) SelectPageByID(ctx context.Context, statementId string, parameter any, page mybatis.PageRequest) (*mybatis.PageResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) InsertByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) UpdateByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) DeleteByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (m *MockXMLSession) GetStatement(statementId string) *mapper.XMLMappedStatement {
	return nil
}

func (m *MockXMLSession) GetResultMap(resultMapId string) *mapper.XMLResultMap {
	return nil
}

func (m *MockXMLSession) GetNamespaces() []string {
	return []string{}
}

func (m *MockXMLSession) GetStatementIds(namespace string) []string {
	return []string{}
}

// RecordOperation 记录操作统计
func (s *AnnotationStressStats) RecordOperation(success bool, latency time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	atomic.AddInt64(&s.TotalOperations, 1)

	if success {
		atomic.AddInt64(&s.SuccessOperations, 1)
	} else {
		atomic.AddInt64(&s.FailedOperations, 1)
	}

	latencyMs := latency.Milliseconds()
	atomic.AddInt64(&s.TotalLatency, latencyMs)

	if latencyMs < s.MinLatency {
		s.MinLatency = latencyMs
	}
	if latencyMs > s.MaxLatency {
		s.MaxLatency = latencyMs
	}
}

// UpdateResourceUsage 更新资源使用统计
func (s *AnnotationStressStats) UpdateResourceUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memUsage := int64(m.Alloc)
	atomic.StoreInt64(&s.MemoryUsageBytes, memUsage)

	if memUsage > s.MaxMemoryBytes {
		atomic.StoreInt64(&s.MaxMemoryBytes, memUsage)
	}

	atomic.StoreInt64(&s.ActiveGoroutines, int64(runtime.NumGoroutine()))

	if s.ActiveGoroutines > s.MaxGoroutines {
		atomic.StoreInt64(&s.MaxGoroutines, s.ActiveGoroutines)
	}
}

// TestAnnotationHighConcurrency 高并发annotation操作测试
func TestAnnotationHighConcurrency(t *testing.T) {
	tester, err := NewAnnotationStressTester("/tmp/annotation_stress.db")
	if err != nil {
		t.Fatalf("Failed to create tester: %v", err)
	}
	defer tester.db.Close()

	concurrency := 200
	operationsPerGoroutine := 50
	ctx := context.Background()

	t.Logf("开始高并发annotation测试: %d协程, 每个协程%d次操作", concurrency, operationsPerGoroutine)

	var wg sync.WaitGroup
	startTime := time.Now()

	// 资源监控协程
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				tester.stats.UpdateResourceUsage()
			case <-ctx.Done():
				return
			}
		}
	}()

	// 启动并发测试
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				// 测试插入操作
				user := &TestUser{
					Username: fmt.Sprintf("user_%d_%d", workerID, j),
					Email:    fmt.Sprintf("user_%d_%d@test.com", workerID, j),
					Age:      20 + (workerID % 50),
					Status:   "active",
				}

				opStart := time.Now()
				id, err := tester.annotationSess.Insert(ctx, user)
				latency := time.Since(opStart)

				success := err == nil && id > 0
				tester.stats.RecordOperation(success, latency)

				if !success {
					atomic.AddInt64(&tester.stats.DatabaseErrors, 1)
				}

				// 测试查询操作
				if success {
					opStart = time.Now()
					_, err = tester.annotationSess.SelectByID(ctx, &TestUser{}, id)
					latency = time.Since(opStart)

					tester.stats.RecordOperation(err == nil, latency)

					if err != nil {
						atomic.AddInt64(&tester.stats.DatabaseErrors, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	// 输出测试结果
	t.Logf("高并发annotation测试完成:")
	t.Logf("- 测试耗时: %v", duration)
	t.Logf("- 总操作数: %d", tester.stats.TotalOperations)
	t.Logf("- 成功操作: %d", tester.stats.SuccessOperations)
	t.Logf("- 失败操作: %d", tester.stats.FailedOperations)
	t.Logf("- 成功率: %.2f%%", float64(tester.stats.SuccessOperations)/float64(tester.stats.TotalOperations)*100)
	t.Logf("- 平均延迟: %.2fms", float64(tester.stats.TotalLatency)/float64(tester.stats.TotalOperations))
	t.Logf("- 最小延迟: %dms", tester.stats.MinLatency)
	t.Logf("- 最大延迟: %dms", tester.stats.MaxLatency)
	t.Logf("- QPS: %.2f", float64(tester.stats.TotalOperations)/duration.Seconds())
	t.Logf("- 最大协程数: %d", tester.stats.MaxGoroutines)
	t.Logf("- 最大内存使用: %.2fMB", float64(tester.stats.MaxMemoryBytes)/(1024*1024))
	t.Logf("- 数据库错误: %d", tester.stats.DatabaseErrors)

	// 验证测试结果
	successRate := float64(tester.stats.SuccessOperations) / float64(tester.stats.TotalOperations) * 100
	if successRate < 95.0 {
		t.Errorf("Success rate too low: %.2f%%, expected >= 95%%", successRate)
	}

	avgLatency := float64(tester.stats.TotalLatency) / float64(tester.stats.TotalOperations)
	if avgLatency > 50.0 { // 50ms阈值
		t.Errorf("Average latency too high: %.2fms, expected <= 50ms", avgLatency)
	}
}

// Close 关闭测试器
func (tester *AnnotationStressTester) Close() error {
	return tester.db.Close()
}
