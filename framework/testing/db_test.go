package testing

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/orm"
)

// DBTestSuite 数据库测试套件
type DBTestSuite struct {
	*BaseTestSuite
	DB           *gorm.DB
	ORM          *orm.ORM
	EnhancedORM  *orm.EnhancedORM
	TestDB       *sql.DB
	transactions []*gorm.DB
	cleanup      []func()
	mutex        sync.Mutex
}

// NewDBTestSuite 创建数据库测试套件
func NewDBTestSuite() *DBTestSuite {
	return &DBTestSuite{
		BaseTestSuite: NewBaseTestSuite(),
		transactions:  make([]*gorm.DB, 0),
		cleanup:       make([]func(), 0),
	}
}

// SetUp 设置数据库测试环境
func (dts *DBTestSuite) SetUp(t *testing.T) {
	dts.BaseTestSuite.SetUp(t)

	// 创建测试数据库配置
	dbConfig := &orm.DatabaseConfig{
		Type:         "sqlite",
		Database:     ":memory:",
		MaxIdleConns: 5,
		MaxOpenConns: 10,
		MaxLifetime:  3600,     // 1小时，单位秒
		LogLevel:     "silent", // 测试时不输出SQL日志
		SlowQuery:    1000,     // 慢查询阈值1秒
	}

	// 创建ORM实例
	var err error
	dts.ORM, err = orm.NewORM(dbConfig)
	if err != nil {
		t.Fatalf("创建ORM失败: %v", err)
	}

	dts.DB = dts.ORM.DB()
	if dts.DB == nil {
		t.Fatal("获取数据库连接失败")
	}

	// 创建增强ORM实例（如果存在的话）
	if dts.EnhancedORM != nil {
		poolConfig := orm.DefaultPoolConfig()
		if poolConfig != nil {
			poolConfig.MetricsEnabled = false     // 测试时关闭指标收集
			poolConfig.HealthCheckEnabled = false // 测试时关闭健康检查
		}

		dts.EnhancedORM, err = orm.NewEnhancedORM(dbConfig, poolConfig)
		if err != nil {
			config.Warnf("创建增强ORM失败，将跳过: %v", err)
			dts.EnhancedORM = nil
		}
	}

	// 获取底层数据库连接
	dts.TestDB, err = dts.DB.DB()
	if err != nil {
		t.Fatalf("获取底层数据库连接失败: %v", err)
	}

	config.Infof("数据库测试套件初始化完成，使用驱动: %s", dbConfig.Type)
}

// TearDown 清理数据库测试环境
func (dts *DBTestSuite) TearDown(t *testing.T) {
	dts.mutex.Lock()
	defer dts.mutex.Unlock()

	// 回滚所有事务
	for _, tx := range dts.transactions {
		tx.Rollback()
	}
	dts.transactions = nil

	// 执行清理函数
	for i := len(dts.cleanup) - 1; i >= 0; i-- {
		func() {
			defer func() {
				if r := recover(); r != nil {
					config.Errorf("清理函数发生panic: %v", r)
				}
			}()
			dts.cleanup[i]()
		}()
	}
	dts.cleanup = nil

	// 关闭数据库连接
	if dts.EnhancedORM != nil {
		dts.EnhancedORM.Close()
	}

	if dts.ORM != nil {
		dts.ORM.Close()
	}

	dts.BaseTestSuite.TearDown(t)
	config.Info("数据库测试套件清理完成")
}

// BeginTransaction 开始事务
func (dts *DBTestSuite) BeginTransaction() *gorm.DB {
	dts.mutex.Lock()
	defer dts.mutex.Unlock()

	tx := dts.DB.Begin()
	dts.transactions = append(dts.transactions, tx)
	return tx
}

// RollbackTransaction 回滚事务
func (dts *DBTestSuite) RollbackTransaction(tx *gorm.DB) {
	dts.mutex.Lock()
	defer dts.mutex.Unlock()

	tx.Rollback()

	// 从事务列表中移除
	for i, t := range dts.transactions {
		if t == tx {
			dts.transactions = append(dts.transactions[:i], dts.transactions[i+1:]...)
			break
		}
	}
}

// CommitTransaction 提交事务
func (dts *DBTestSuite) CommitTransaction(tx *gorm.DB) error {
	dts.mutex.Lock()
	defer dts.mutex.Unlock()

	err := tx.Commit().Error

	// 从事务列表中移除
	for i, t := range dts.transactions {
		if t == tx {
			dts.transactions = append(dts.transactions[:i], dts.transactions[i+1:]...)
			break
		}
	}

	return err
}

// AddCleanup 添加清理函数
func (dts *DBTestSuite) AddCleanup(fn func()) {
	dts.mutex.Lock()
	defer dts.mutex.Unlock()
	dts.cleanup = append(dts.cleanup, fn)
}

// CreateTable 创建测试表
func (dts *DBTestSuite) CreateTable(model any) error {
	return dts.DB.AutoMigrate(model)
}

// DropTable 删除测试表
func (dts *DBTestSuite) DropTable(model any) error {
	return dts.DB.Migrator().DropTable(model)
}

// TruncateTable 清空测试表
func (dts *DBTestSuite) TruncateTable(tableName string) error {
	return dts.DB.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error
}

// SeedData 种子数据
func (dts *DBTestSuite) SeedData(data any) error {
	return dts.DB.Create(data).Error
}

// CountRecords 统计记录数
func (dts *DBTestSuite) CountRecords(model any, conditions ...any) int64 {
	var count int64
	query := dts.DB.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	query.Count(&count)
	return count
}

// RecordExists 检查记录是否存在
func (dts *DBTestSuite) RecordExists(model any, conditions ...any) bool {
	return dts.CountRecords(model, conditions...) > 0
}

// LoadFixture 加载测试夹具
func (dts *DBTestSuite) LoadFixture(fixturePath string) error {
	// 读取SQL文件并执行
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		return fmt.Errorf("读取夹具文件失败: %w", err)
	}

	// 按分号分割SQL语句
	statements := strings.Split(string(content), ";")

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if err := dts.DB.Exec(stmt).Error; err != nil {
			return fmt.Errorf("执行SQL语句失败: %w", err)
		}
	}

	return nil
}

// ============= 数据库断言工具 =============

// DBAssertion 数据库断言工具
type DBAssertion struct {
	t  *testing.T
	db *gorm.DB
}

// NewDBAssertion 创建数据库断言工具
func NewDBAssertion(t *testing.T, db *gorm.DB) *DBAssertion {
	return &DBAssertion{t: t, db: db}
}

// AssertRecordCount 断言记录数量
func (dba *DBAssertion) AssertRecordCount(model any, expected int64, conditions ...any) {
	var count int64
	query := dba.db.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	query.Count(&count)

	if count != expected {
		dba.t.Helper()
		dba.t.Errorf("期望 %d 条记录，实际得到 %d 条", expected, count)
	}
}

// AssertRecordExists 断言记录存在
func (dba *DBAssertion) AssertRecordExists(model any, conditions ...any) {
	var count int64
	query := dba.db.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	query.Count(&count)

	if count == 0 {
		dba.t.Helper()
		dba.t.Error("期望记录存在，但实际不存在")
	}
}

// AssertRecordNotExists 断言记录不存在
func (dba *DBAssertion) AssertRecordNotExists(model any, conditions ...any) {
	var count int64
	query := dba.db.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	query.Count(&count)

	if count > 0 {
		dba.t.Helper()
		dba.t.Errorf("期望记录不存在，但找到了 %d 条记录", count)
	}
}

// AssertFieldValue 断言字段值
func (dba *DBAssertion) AssertFieldValue(model any, field string, expected any, conditions ...any) {
	query := dba.db.Model(model)

	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}

	var result map[string]any
	if err := query.Select(field).First(&result).Error; err != nil {
		dba.t.Helper()
		dba.t.Errorf("查询字段值失败: %v", err)
		return
	}

	actual, exists := result[field]
	if !exists {
		dba.t.Helper()
		dba.t.Errorf("结果中未找到字段 '%s'", field)
		return
	}

	if actual != expected {
		dba.t.Helper()
		dba.t.Errorf("期望字段 '%s' 的值为 %v，实际得到 %v", field, expected, actual)
	}
}

// AssertQueryResult 断言查询结果
func (dba *DBAssertion) AssertQueryResult(sql string, expectedRows int, args ...any) {
	rows, err := dba.db.Raw(sql, args...).Rows()
	if err != nil {
		dba.t.Helper()
		dba.t.Errorf("执行查询失败: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if count != expectedRows {
		dba.t.Helper()
		dba.t.Errorf("期望 %d 行，实际得到 %d 行", expectedRows, count)
	}
}

// AssertTableExists 断言表存在
func (dba *DBAssertion) AssertTableExists(tableName string) {
	if !dba.db.Migrator().HasTable(tableName) {
		dba.t.Helper()
		dba.t.Errorf("期望表 '%s' 存在", tableName)
	}
}

// AssertTableNotExists 断言表不存在
func (dba *DBAssertion) AssertTableNotExists(tableName string) {
	if dba.db.Migrator().HasTable(tableName) {
		dba.t.Helper()
		dba.t.Errorf("期望表 '%s' 不存在", tableName)
	}
}

// AssertColumnExists 断言列存在
func (dba *DBAssertion) AssertColumnExists(tableName, columnName string) {
	if !dba.db.Migrator().HasColumn(tableName, columnName) {
		dba.t.Helper()
		dba.t.Errorf("期望表 '%s' 中的列 '%s' 存在", tableName, columnName)
	}
}

// AssertIndexExists 断言索引存在
func (dba *DBAssertion) AssertIndexExists(tableName, indexName string) {
	if !dba.db.Migrator().HasIndex(tableName, indexName) {
		dba.t.Helper()
		dba.t.Errorf("期望表 '%s' 上的索引 '%s' 存在", tableName, indexName)
	}
}

// ============= 测试数据生成器 =============

// TestDataFactory 测试数据工厂
type TestDataFactory struct {
	generators map[string]func() any
	sequences  map[string]int
	mutex      sync.Mutex
}

// NewTestDataFactory 创建测试数据工厂
func NewTestDataFactory() *TestDataFactory {
	return &TestDataFactory{
		generators: make(map[string]func() any),
		sequences:  make(map[string]int),
	}
}

// RegisterGenerator 注册数据生成器
func (tdf *TestDataFactory) RegisterGenerator(name string, generator func() any) {
	tdf.mutex.Lock()
	defer tdf.mutex.Unlock()
	tdf.generators[name] = generator
}

// Generate 生成测试数据
func (tdf *TestDataFactory) Generate(name string) any {
	tdf.mutex.Lock()
	defer tdf.mutex.Unlock()

	generator, exists := tdf.generators[name]
	if !exists {
		panic(fmt.Sprintf("未注册名为 '%s' 的生成器", name))
	}

	return generator()
}

// Sequence 生成序列号
func (tdf *TestDataFactory) Sequence(name string) int {
	tdf.mutex.Lock()
	defer tdf.mutex.Unlock()

	tdf.sequences[name]++
	return tdf.sequences[name]
}

// GenerateBatch 批量生成测试数据
func (tdf *TestDataFactory) GenerateBatch(name string, count int) []any {
	result := make([]any, count)
	for i := 0; i < count; i++ {
		result[i] = tdf.Generate(name)
	}
	return result
}

// ============= 数据库性能测试 =============

// DBBenchmark 数据库性能测试
type DBBenchmark struct {
	db      *gorm.DB
	results map[string]*BenchmarkResult
	mutex   sync.Mutex
}

// BenchmarkResult 性能测试结果
type BenchmarkResult struct {
	Name        string
	Operations  int
	Duration    time.Duration
	TotalTime   time.Duration
	AverageTime time.Duration
	MinTime     time.Duration
	MaxTime     time.Duration
	QPS         float64
}

// NewDBBenchmark 创建数据库性能测试
func NewDBBenchmark(db *gorm.DB) *DBBenchmark {
	return &DBBenchmark{
		db:      db,
		results: make(map[string]*BenchmarkResult),
	}
}

// Run 运行性能测试
func (dbb *DBBenchmark) Run(name string, operations int, fn func(*gorm.DB)) *BenchmarkResult {
	dbb.mutex.Lock()
	defer dbb.mutex.Unlock()

	var totalTime time.Duration
	var minTime = time.Hour
	var maxTime time.Duration

	start := time.Now()

	for i := 0; i < operations; i++ {
		opStart := time.Now()
		fn(dbb.db)
		opDuration := time.Since(opStart)

		totalTime += opDuration
		if opDuration < minTime {
			minTime = opDuration
		}
		if opDuration > maxTime {
			maxTime = opDuration
		}
	}

	duration := time.Since(start)
	averageTime := totalTime / time.Duration(operations)
	qps := float64(operations) / duration.Seconds()

	result := &BenchmarkResult{
		Name:        name,
		Operations:  operations,
		Duration:    duration,
		TotalTime:   totalTime,
		AverageTime: averageTime,
		MinTime:     minTime,
		MaxTime:     maxTime,
		QPS:         qps,
	}

	dbb.results[name] = result
	return result
}

// GetResult 获取测试结果
func (dbb *DBBenchmark) GetResult(name string) *BenchmarkResult {
	dbb.mutex.Lock()
	defer dbb.mutex.Unlock()
	return dbb.results[name]
}

// GetResults 获取所有测试结果
func (dbb *DBBenchmark) GetResults() map[string]*BenchmarkResult {
	dbb.mutex.Lock()
	defer dbb.mutex.Unlock()

	results := make(map[string]*BenchmarkResult)
	for k, v := range dbb.results {
		results[k] = v
	}
	return results
}

// PrintResults 打印测试结果
func (dbb *DBBenchmark) PrintResults() {
	dbb.mutex.Lock()
	defer dbb.mutex.Unlock()

	config.Info("=== 数据库性能测试结果 ===")
	for name, result := range dbb.results {
		config.Infof("测试: %s", name)
		config.Infof("  操作数: %d", result.Operations)
		config.Infof("  总耗时: %v", result.Duration)
		config.Infof("  平均耗时: %v", result.AverageTime)
		config.Infof("  最小耗时: %v", result.MinTime)
		config.Infof("  最大耗时: %v", result.MaxTime)
		config.Infof("  QPS: %.2f", result.QPS)
		config.Info("---")
	}
}

// ============= 便捷函数 =============

// WithTransaction 在事务中执行测试
func WithTransaction(dts *DBTestSuite, fn func(*gorm.DB)) {
	tx := dts.BeginTransaction()
	defer dts.RollbackTransaction(tx)
	fn(tx)
}

// WithCleanup 使用自动清理执行测试
func WithCleanup(dts *DBTestSuite, fn func()) {
	defer func() {
		// 清理所有表
		tables := []string{}
		dts.DB.Raw("SELECT name FROM sqlite_master WHERE type='table'").Pluck("name", &tables)
		for _, table := range tables {
			if !strings.HasPrefix(table, "sqlite_") {
				dts.DB.Exec(fmt.Sprintf("DELETE FROM %s", table))
			}
		}
	}()
	fn()
}

// CreateTestModels 创建常用测试模型
func CreateTestModels(db *gorm.DB) error {
	// 用户模型
	type User struct {
		ID        uint   `gorm:"primaryKey"`
		Name      string `gorm:"size:100"`
		Email     string `gorm:"size:100;uniqueIndex"`
		Age       int
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	// 文章模型
	type Post struct {
		ID        uint   `gorm:"primaryKey"`
		Title     string `gorm:"size:200"`
		Content   string `gorm:"type:text"`
		UserID    uint
		User      User `gorm:"foreignKey:UserID"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	return db.AutoMigrate(&User{}, &Post{})
}

// GetTestDB 获取测试数据库实例
func GetTestDB() (*gorm.DB, func()) {
	dbConfig := &orm.DatabaseConfig{
		Type:     "sqlite",
		Database: ":memory:",
		LogLevel: "silent",
	}

	ormInstance, err := orm.NewORM(dbConfig)
	if err != nil {
		panic(fmt.Sprintf("创建测试数据库失败: %v", err))
	}

	cleanup := func() {
		ormInstance.Close()
	}

	return ormInstance.DB(), cleanup
}