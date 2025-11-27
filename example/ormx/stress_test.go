package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zsy619/yyhertz/framework/orm"
)

// StressTestMetrics 压力测试指标
type StressTestMetrics struct {
	TotalRequests     int64         // 总请求数
	SuccessRequests   int64         // 成功请求数
	FailedRequests    int64         // 失败请求数
	TotalDuration     time.Duration // 总耗时
	MinDuration       time.Duration // 最小耗时
	MaxDuration       time.Duration // 最大耗时
	AverageDuration   time.Duration // 平均耗时
	RequestsPerSecond float64       // 每秒请求数
	ErrorRate         float64       // 错误率
	StartTime         time.Time     // 开始时间
	EndTime           time.Time     // 结束时间
}

// StressTestResult 压力测试结果
type StressTestResult struct {
	TestName    string             // 测试名称
	Concurrency int                // 并发数
	Duration    time.Duration      // 测试持续时间
	Metrics     *StressTestMetrics // 测试指标
	MemStats    runtime.MemStats   // 内存统计
	CPUUsage    float64            // CPU使用率
	Errors      []string           // 错误列表
}

// StressTestConfig 压力测试配置
type StressTestConfig struct {
	Concurrency      int           // 并发数
	Duration         time.Duration // 测试持续时间
	RequestsPerGo    int           // 每个goroutine的请求数
	WarmupTime       time.Duration // 预热时间
	EnableMonitoring bool          // 是否启用监控
	PrintInterval    time.Duration // 打印间隔
}

// DefaultStressTestConfig 默认压力测试配置
func DefaultStressTestConfig() *StressTestConfig {
	return &StressTestConfig{
		Concurrency:      10,
		Duration:         30 * time.Second,
		RequestsPerGo:    1000,
		WarmupTime:       5 * time.Second,
		EnableMonitoring: true,
		PrintInterval:    5 * time.Second,
	}
}

// StressTestUser 压力测试用户模型
type StressTestUser struct {
	orm.BaseModel
	Username string `gorm:"size:50;index" json:"username"`
	Email    string `gorm:"size:100;uniqueIndex" json:"email"`
	Password string `gorm:"size:255" json:"-"`
	Status   string `gorm:"size:20;index" json:"status"`
	Age      int    `json:"age"`
	City     string `gorm:"size:50;index" json:"city"`
}

// StressTestOrder 压力测试订单模型
type StressTestOrder struct {
	orm.BaseModel
	UserID    uint    `gorm:"index" json:"user_id"`
	OrderNo   string  `gorm:"size:32;uniqueIndex" json:"order_no"`
	Amount    float64 `gorm:"type:decimal(10,2)" json:"amount"`
	Status    string  `gorm:"size:20;index" json:"status"`
	CreatedBy string  `gorm:"size:50" json:"created_by"`
}

// StressTestSuite 压力测试套件
type StressTestSuite struct {
	orm     *orm.ORM
	config  *StressTestConfig
	results []*StressTestResult
	mutex   sync.RWMutex
}

// NewStressTestSuite 创建压力测试套件
func NewStressTestSuite(config *StressTestConfig) (*StressTestSuite, error) {
	if config == nil {
		config = DefaultStressTestConfig()
	}

	// 创建数据库配置
	dbConfig := &orm.DatabaseConfig{
		Type:         "sqlite",
		Database:     "stress_test.db",
		MaxIdleConns: config.Concurrency * 2,
		MaxOpenConns: config.Concurrency * 4,
		MaxLifetime:  3600,
		LogLevel:     "silent", // 关闭日志以提高性能
		SlowQuery:    1000,
	}

	// 创建 ORM 实例
	ormInstance, err := orm.NewORM(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("创建ORM实例失败: %w", err)
	}

	// 自动迁移
	if err := ormInstance.AutoMigrate(&StressTestUser{}, &StressTestOrder{}); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	return &StressTestSuite{
		orm:     ormInstance,
		config:  config,
		results: make([]*StressTestResult, 0),
	}, nil
}

// RunAllStressTests 运行所有压力测试
func (s *StressTestSuite) RunAllStressTests() {
	fmt.Println("=== YYHertz ORM 压力测试套件 ===")

	// 预热数据库
	s.warmupDatabase()

	// 运行各种压力测试
	tests := []struct {
		name string
		fn   func() *StressTestResult
	}{
		{"创建操作压力测试", s.runCreateStressTest},
		{"查询操作压力测试", s.runQueryStressTest},
		{"更新操作压力测试", s.runUpdateStressTest},
		{"删除操作压力测试", s.runDeleteStressTest},
		{"混合操作压力测试", s.runMixedStressTest},
		{"事务操作压力测试", s.runTransactionStressTest},
		{"分页查询压力测试", s.runPaginationStressTest},
		{"连接池压力测试", s.runConnectionPoolStressTest},
		{"高并发读写测试", s.runConcurrentReadWriteTest},
		{"长时间运行测试", s.runLongRunningTest},
	}

	for _, test := range tests {
		fmt.Printf("开始运行: %s\n", test.name)
		result := test.fn()
		s.addResult(result)
		s.printTestResult(result)
		
		// 测试间隔，让系统恢复
		time.Sleep(2 * time.Second)
		fmt.Println()
	}

	// 打印总结报告
	s.printSummaryReport()
}

// warmupDatabase 预热数据库
func (s *StressTestSuite) warmupDatabase() {
	fmt.Println("预热数据库...")
	crud := orm.GetSimpleCRUD()

	// 创建一些测试数据
	for i := 0; i < 100; i++ {
		user := &StressTestUser{
			Username: fmt.Sprintf("warmup_user_%d", i),
			Email:    fmt.Sprintf("warmup_%d@example.com", i),
			Password: "password123",
			Status:   "active",
			Age:      rand.Intn(50) + 18,
			City:     fmt.Sprintf("City_%d", rand.Intn(10)),
		}
		crud.Create(user)
	}

	fmt.Println("✅ 数据库预热完成")
}

// runCreateStressTest 创建操作压力测试
func (s *StressTestSuite) runCreateStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					user := &StressTestUser{
						Username: fmt.Sprintf("stress_user_%d_%d", workerID, time.Now().UnixNano()),
						Email:    fmt.Sprintf("stress_%d_%d@example.com", workerID, time.Now().UnixNano()),
						Password: "password123",
						Status:   "active",
						Age:      rand.Intn(50) + 18,
						City:     fmt.Sprintf("City_%d", rand.Intn(10)),
					}

					err := crud.Create(user)
					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "创建操作压力测试",
		Concurrency: s.config.Concurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runQueryStressTest 查询操作压力测试
func (s *StressTestSuite) runQueryStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()
			queryTypes := []string{"by_status", "by_age", "by_city", "by_email"}

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					var users []StressTestUser
					var err error

					queryType := queryTypes[rand.Intn(len(queryTypes))]
					switch queryType {
					case "by_status":
						err = crud.Find(&users, "status = ?", "active")
					case "by_age":
						err = crud.Find(&users, "age > ?", rand.Intn(30)+18)
					case "by_city":
						err = crud.Find(&users, "city = ?", fmt.Sprintf("City_%d", rand.Intn(10)))
					case "by_email":
						err = crud.First(&users, "email LIKE ?", "%@example.com")
					}

					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "查询操作压力测试",
		Concurrency: s.config.Concurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runUpdateStressTest 更新操作压力测试
func (s *StressTestSuite) runUpdateStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	// 准备一些用户数据用于更新
	crud := orm.GetSimpleCRUD()
	var userIDs []uint
	var users []StressTestUser
	crud.Find(&users, "status = ?", "active")
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	if len(userIDs) == 0 {
		return &StressTestResult{
			TestName: "更新操作压力测试",
			Errors:   []string{"没有找到可更新的用户"},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					userID := userIDs[rand.Intn(len(userIDs))]
					user := &StressTestUser{}
					user.ID = userID

					updateFields := map[string]interface{}{
						"age":  rand.Intn(50) + 18,
						"city": fmt.Sprintf("UpdatedCity_%d", rand.Intn(20)),
					}

					err := crud.Updates(user, updateFields)
					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "更新操作压力测试",
		Concurrency: s.config.Concurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runDeleteStressTest 删除操作压力测试
func (s *StressTestSuite) runDeleteStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					// 创建一个用户用于删除
					user := &StressTestUser{
						Username: fmt.Sprintf("delete_user_%d_%d", workerID, time.Now().UnixNano()),
						Email:    fmt.Sprintf("delete_%d_%d@example.com", workerID, time.Now().UnixNano()),
						Password: "password123",
						Status:   "pending_delete",
						Age:      25,
						City:     "DeleteCity",
					}
					crud.Create(user)

					start := time.Now()
					err := crud.Delete(user)
					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "删除操作压力测试",
		Concurrency: s.config.Concurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runMixedStressTest 混合操作压力测试
func (s *StressTestSuite) runMixedStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					var err error

					// 随机选择操作类型
					operation := rand.Intn(4)
					switch operation {
					case 0: // 创建
						user := &StressTestUser{
							Username: fmt.Sprintf("mixed_user_%d_%d", workerID, time.Now().UnixNano()),
							Email:    fmt.Sprintf("mixed_%d_%d@example.com", workerID, time.Now().UnixNano()),
							Password: "password123",
							Status:   "active",
							Age:      rand.Intn(50) + 18,
							City:     fmt.Sprintf("City_%d", rand.Intn(10)),
						}
						err = crud.Create(user)

					case 1: // 查询
						var users []StressTestUser
						err = crud.Find(&users, "status = ?", "active")

					case 2: // 更新
						var users []StressTestUser
						if crud.Find(&users, "status = ?", "active") == nil && len(users) > 0 {
							user := users[rand.Intn(len(users))]
							err = crud.Update(&user, "age", rand.Intn(50)+18)
						}

					case 3: // 统计
						_, err = crud.Count(&StressTestUser{})
					}

					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "混合操作压力测试",
		Concurrency: s.config.Concurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runTransactionStressTest 事务操作压力测试
func (s *StressTestSuite) runTransactionStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					err := orm.QuickTransaction(func(crud *orm.SimpleCRUD) error {
						// 在事务中创建用户和订单
						user := &StressTestUser{
							Username: fmt.Sprintf("tx_user_%d_%d", workerID, time.Now().UnixNano()),
							Email:    fmt.Sprintf("tx_%d_%d@example.com", workerID, time.Now().UnixNano()),
							Password: "password123",
							Status:   "active",
							Age:      rand.Intn(50) + 18,
							City:     fmt.Sprintf("City_%d", rand.Intn(10)),
						}
						if err := crud.Create(user); err != nil {
							return err
						}

						order := &StressTestOrder{
							UserID:    user.ID,
							OrderNo:   fmt.Sprintf("ORD_%d_%d", workerID, time.Now().UnixNano()),
							Amount:    rand.Float64() * 1000,
							Status:    "pending",
							CreatedBy: fmt.Sprintf("worker_%d", workerID),
						}
						return crud.Create(order)
					})

					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "事务操作压力测试",
		Concurrency: s.config.Concurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runPaginationStressTest 分页查询压力测试
func (s *StressTestSuite) runPaginationStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					var users []StressTestUser
					page := rand.Intn(10) + 1
					pageSize := rand.Intn(20) + 10

					_, err := crud.Paginate(&StressTestUser{}, page, pageSize, &users)
					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "分页查询压力测试",
		Concurrency: s.config.Concurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runConnectionPoolStressTest 连接池压力测试
func (s *StressTestSuite) runConnectionPoolStressTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	// 使用更高的并发数测试连接池
	highConcurrency := s.config.Concurrency * 3
	var wg sync.WaitGroup
	wg.Add(highConcurrency)

	for i := 0; i < highConcurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					// 执行不同类型的操作以测试连接池
					operations := []func() error{
						func() error {
							var users []StressTestUser
							return crud.Find(&users, "status = ?", "active")
						},
						func() error {
							count, err := crud.Count(&StressTestUser{})
							_ = count
							return err
						},
						func() error {
							user := &StressTestUser{
								Username: fmt.Sprintf("pool_user_%d", workerID),
								Email:    fmt.Sprintf("pool_%d@example.com", workerID),
								Status:   "active",
								Age:      25,
							}
							return crud.Create(user)
						},
					}

					operation := operations[rand.Intn(len(operations))]
					err := operation()
					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "连接池压力测试",
		Concurrency: highConcurrency,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runConcurrentReadWriteTest 并发读写测试
func (s *StressTestSuite) runConcurrentReadWriteTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	ctx, cancel := context.WithTimeout(context.Background(), s.config.Duration)
	defer cancel()

	var wg sync.WaitGroup
	
	// 读取goroutines
	readGoroutines := s.config.Concurrency * 2
	wg.Add(readGoroutines)
	for i := 0; i < readGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					var users []StressTestUser
					err := crud.Find(&users, "status = ?", "active")
					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	// 写入goroutines
	writeGoroutines := s.config.Concurrency
	wg.Add(writeGoroutines)
	for i := 0; i < writeGoroutines; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					user := &StressTestUser{
						Username: fmt.Sprintf("rw_user_%d_%d", workerID, time.Now().UnixNano()),
						Email:    fmt.Sprintf("rw_%d_%d@example.com", workerID, time.Now().UnixNano()),
						Password: "password123",
						Status:   "active",
						Age:      rand.Intn(50) + 18,
						City:     fmt.Sprintf("City_%d", rand.Intn(10)),
					}
					err := crud.Create(user)
					duration := time.Since(start)
					durations = append(durations, duration)

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	return &StressTestResult{
		TestName:    "并发读写测试",
		Concurrency: readGoroutines + writeGoroutines,
		Duration:    s.config.Duration,
		Metrics:     metrics,
		Errors:      errors,
	}
}

// runLongRunningTest 长时间运行测试
func (s *StressTestSuite) runLongRunningTest() *StressTestResult {
	metrics := &StressTestMetrics{StartTime: time.Now()}
	var errors []string
	var durations []time.Duration

	// 长时间运行测试持续时间更长
	longDuration := 2 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), longDuration)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(s.config.Concurrency)

	// 记录内存使用情况
	var memStatsBefore, memStatsAfter runtime.MemStats
	runtime.ReadMemStats(&memStatsBefore)

	for i := 0; i < s.config.Concurrency; i++ {
		go func(workerID int) {
			defer wg.Done()
			crud := orm.GetSimpleCRUD()
			operationCount := 0

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()
					var err error

					// 循环执行不同类型的操作
					switch operationCount % 4 {
					case 0: // 创建
						user := &StressTestUser{
							Username: fmt.Sprintf("long_user_%d_%d", workerID, operationCount),
							Email:    fmt.Sprintf("long_%d_%d@example.com", workerID, operationCount),
							Password: "password123",
							Status:   "active",
							Age:      rand.Intn(50) + 18,
							City:     fmt.Sprintf("City_%d", rand.Intn(10)),
						}
						err = crud.Create(user)

					case 1: // 查询
						var users []StressTestUser
						err = crud.Find(&users, "status = ? AND age > ?", "active", rand.Intn(30)+18)

					case 2: // 更新
						var users []StressTestUser
						if crud.Find(&users, "status = ?", "active") == nil && len(users) > 0 {
							user := users[rand.Intn(len(users))]
							err = crud.Update(&user, "age", rand.Intn(50)+18)
						}

					case 3: // 分页
						var users []StressTestUser
						_, err = crud.Paginate(&StressTestUser{}, rand.Intn(10)+1, 20, &users)
					}

					duration := time.Since(start)
					durations = append(durations, duration)
					operationCount++

					atomic.AddInt64(&metrics.TotalRequests, 1)
					if err != nil {
						atomic.AddInt64(&metrics.FailedRequests, 1)
						errors = append(errors, err.Error())
					} else {
						atomic.AddInt64(&metrics.SuccessRequests, 1)
					}

					// 偶尔强制垃圾回收
					if operationCount%1000 == 0 {
						runtime.GC()
					}
				}
			}
		}(i)
	}

	wg.Wait()
	runtime.ReadMemStats(&memStatsAfter)
	metrics.EndTime = time.Now()
	s.calculateMetrics(metrics, durations)

	result := &StressTestResult{
		TestName:    "长时间运行测试",
		Concurrency: s.config.Concurrency,
		Duration:    longDuration,
		Metrics:     metrics,
		MemStats:    memStatsAfter,
		Errors:      errors,
	}

	// 添加内存使用信息
	memUsage := float64(memStatsAfter.Alloc-memStatsBefore.Alloc) / 1024 / 1024
	result.Errors = append(result.Errors, fmt.Sprintf("内存使用增长: %.2f MB", memUsage))

	return result
}

// calculateMetrics 计算测试指标
func (s *StressTestSuite) calculateMetrics(metrics *StressTestMetrics, durations []time.Duration) {
	if len(durations) == 0 {
		return
	}

	// 计算总耗时
	totalDuration := metrics.EndTime.Sub(metrics.StartTime)
	metrics.TotalDuration = totalDuration

	// 计算最小、最大、平均耗时
	metrics.MinDuration = durations[0]
	metrics.MaxDuration = durations[0]
	var totalNs int64

	for _, duration := range durations {
		if duration < metrics.MinDuration {
			metrics.MinDuration = duration
		}
		if duration > metrics.MaxDuration {
			metrics.MaxDuration = duration
		}
		totalNs += duration.Nanoseconds()
	}

	metrics.AverageDuration = time.Duration(totalNs / int64(len(durations)))

	// 计算每秒请求数
	seconds := totalDuration.Seconds()
	if seconds > 0 {
		metrics.RequestsPerSecond = float64(metrics.TotalRequests) / seconds
	}

	// 计算错误率
	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests) * 100
	}
}

// addResult 添加测试结果
func (s *StressTestSuite) addResult(result *StressTestResult) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.results = append(s.results, result)
}

// printTestResult 打印测试结果
func (s *StressTestSuite) printTestResult(result *StressTestResult) {
	fmt.Printf("✅ %s 完成\n", result.TestName)
	fmt.Printf("   并发数: %d\n", result.Concurrency)
	fmt.Printf("   持续时间: %v\n", result.Duration)
	fmt.Printf("   总请求数: %d\n", result.Metrics.TotalRequests)
	fmt.Printf("   成功请求: %d\n", result.Metrics.SuccessRequests)
	fmt.Printf("   失败请求: %d\n", result.Metrics.FailedRequests)
	fmt.Printf("   每秒请求数: %.2f\n", result.Metrics.RequestsPerSecond)
	fmt.Printf("   错误率: %.2f%%\n", result.Metrics.ErrorRate)
	fmt.Printf("   平均耗时: %v\n", result.Metrics.AverageDuration)
	fmt.Printf("   最小耗时: %v\n", result.Metrics.MinDuration)
	fmt.Printf("   最大耗时: %v\n", result.Metrics.MaxDuration)

	if len(result.Errors) > 0 {
		fmt.Printf("   错误信息: %d条错误\n", len(result.Errors))
		if len(result.Errors) <= 5 {
			for _, err := range result.Errors {
				fmt.Printf("     - %s\n", err)
			}
		}
	}
}

// printSummaryReport 打印总结报告
func (s *StressTestSuite) printSummaryReport() {
	fmt.Println("=== 压力测试总结报告 ===")
	fmt.Printf("测试配置: 并发数=%d, 测试时长=%v\n", s.config.Concurrency, s.config.Duration)
	fmt.Printf("总测试项: %d\n", len(s.results))

	var totalRequests, totalSuccessRequests, totalFailedRequests int64
	var totalRPS float64
	var avgErrorRate float64

	for _, result := range s.results {
		totalRequests += result.Metrics.TotalRequests
		totalSuccessRequests += result.Metrics.SuccessRequests
		totalFailedRequests += result.Metrics.FailedRequests
		totalRPS += result.Metrics.RequestsPerSecond
		avgErrorRate += result.Metrics.ErrorRate
	}

	if len(s.results) > 0 {
		totalRPS /= float64(len(s.results))
		avgErrorRate /= float64(len(s.results))
	}

	fmt.Printf("\n总体统计:\n")
	fmt.Printf("  总请求数: %d\n", totalRequests)
	fmt.Printf("  成功请求: %d\n", totalSuccessRequests)
	fmt.Printf("  失败请求: %d\n", totalFailedRequests)
	fmt.Printf("  平均RPS: %.2f\n", totalRPS)
	fmt.Printf("  平均错误率: %.2f%%\n", avgErrorRate)

	// 获取系统统计信息
	stats, err := orm.GetConnectionStats()
	if err == nil {
		fmt.Printf("\n连接池统计:\n")
		fmt.Printf("  最大连接数: %d\n", stats.MaxOpenConnections)
		fmt.Printf("  活跃连接数: %d\n", stats.InUse)
		fmt.Printf("  空闲连接数: %d\n", stats.Idle)
		fmt.Printf("  等待次数: %d\n", stats.WaitCount)
		fmt.Printf("  等待时长: %v\n", stats.WaitDuration)
	}

	fmt.Printf("\n✅ 压力测试完成！\n")
}

// RunStressTestDemo 运行压力测试演示
func RunStressTestDemo() {
	fmt.Println("开始 YYHertz ORM 压力测试...")

	config := &StressTestConfig{
		Concurrency:      5,  // 降低并发数以适应演示
		Duration:         15 * time.Second, // 缩短测试时间
		RequestsPerGo:    100,
		WarmupTime:       2 * time.Second,
		EnableMonitoring: true,
		PrintInterval:    3 * time.Second,
	}

	suite, err := NewStressTestSuite(config)
	if err != nil {
		log.Fatalf("创建压力测试套件失败: %v", err)
	}

	suite.RunAllStressTests()
}

// 注释掉main函数以避免冲突
// func main() {
// 	RunStressTestDemo()
// }