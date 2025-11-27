# 🔄 事务管理

事务管理是确保数据一致性和完整性的核心机制。YYHertz MVC框架基于GORM提供了强大的事务管理功能，支持声明式和编程式事务处理。

## 🌟 核心特性

### ✨ 事务功能
- **🔒 ACID保证** - 原子性、一致性、隔离性、持久性
- **🎯 多种事务模式** - 手动事务、自动事务、嵌套事务
- **🔄 事务传播** - 支持多种事务传播行为
- **💾 事务隔离** - 可配置的隔离级别
- **🎭 事务回滚** - 智能回滚和恢复机制

### 🎪 高级功能
- **📊 事务监控** - 事务性能和状态监控
- **🔍 事务日志** - 详细的事务执行日志
- **⚡ 性能优化** - 事务批处理和优化
- **🛡️ 安全保障** - 事务超时和死锁检测

## 🚀 基础事务操作

### 1. 手动事务管理

```go
// controllers/user_controller.go
package controllers

import (
    "github.com/zsy619/yyhertz/framework/mvc"
    "github.com/zsy619/yyhertz/database"
    "github.com/zsy619/yyhertz/models"
    "gorm.io/gorm"
)

type UserController struct {
    mvc.BaseController
}

// PostCreateWithProfile 创建用户和用户资料（手动事务）
func (c *UserController) PostCreateWithProfile() {
    var req struct {
        User    models.User        `json:"user"`
        Profile models.UserProfile `json:"profile"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 开始事务
    tx := database.DB().Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    // 创建用户
    if err := tx.Create(&req.User).Error; err != nil {
        tx.Rollback()
        c.Error(500, "创建用户失败: "+err.Error())
        return
    }
    
    // 设置关联ID
    req.Profile.UserID = req.User.ID
    
    // 创建用户资料
    if err := tx.Create(&req.Profile).Error; err != nil {
        tx.Rollback()
        c.Error(500, "创建用户资料失败: "+err.Error())
        return
    }
    
    // 提交事务
    if err := tx.Commit().Error; err != nil {
        c.Error(500, "事务提交失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code":    0,
        "message": "创建成功",
        "data": map[string]interface{}{
            "user":    req.User,
            "profile": req.Profile,
        },
    })
}
```

### 2. 事务回调函数

```go
// 使用GORM的Transaction方法
func (c *UserController) PostCreateWithTransaction() {
    var req struct {
        User    models.User        `json:"user"`
        Profile models.UserProfile `json:"profile"`
    }
    
    if err := c.BindJSON(&req); err != nil {
        c.Error(400, "参数错误: "+err.Error())
        return
    }
    
    // 在事务中执行操作
    err := database.DB().Transaction(func(tx *gorm.DB) error {
        // 创建用户
        if err := tx.Create(&req.User).Error; err != nil {
            return err
        }
        
        // 设置关联ID
        req.Profile.UserID = req.User.ID
        
        // 创建用户资料
        if err := tx.Create(&req.Profile).Error; err != nil {
            return err
        }
        
        // 返回nil提交事务，返回error自动回滚
        return nil
    })
    
    if err != nil {
        c.Error(500, "事务执行失败: "+err.Error())
        return
    }
    
    c.JSON(map[string]interface{}{
        "code":    0,
        "message": "创建成功",
        "data": map[string]interface{}{
            "user":    req.User,
            "profile": req.Profile,
        },
    })
}
```

## 🏗️ 声明式事务管理

### 1. 事务装饰器

```go
// transaction/decorator.go
package transaction

import (
    "context"
    "fmt"
    "reflect"
    "runtime"
    
    "gorm.io/gorm"
    "github.com/zsy619/yyhertz/database"
)

// Transactional 事务装饰器
type Transactional struct {
    Propagation Propagation
    Isolation   Isolation
    ReadOnly    bool
    Timeout     int // 超时时间（秒）
}

type Propagation int

const (
    PropagationRequired Propagation = iota
    PropagationRequiresNew
    PropagationSupports
    PropagationNotSupported
    PropagationNever
    PropagationNested
)

type Isolation int

const (
    IsolationDefault Isolation = iota
    IsolationReadUncommitted
    IsolationReadCommitted
    IsolationRepeatableRead
    IsolationSerializable
)

// WithTransaction 事务方法装饰器
func WithTransaction(config Transactional) func(interface{}) interface{} {
    return func(fn interface{}) interface{} {
        fnValue := reflect.ValueOf(fn)
        fnType := fnValue.Type()
        
        if fnType.Kind() != reflect.Func {
            panic("WithTransaction can only be applied to functions")
        }
        
        return reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
            return executeInTransaction(config, fnValue, args)
        }).Interface()
    }
}

func executeInTransaction(config Transactional, fn reflect.Value, args []reflect.Value) []reflect.Value {
    db := database.DB()
    
    // 根据传播行为决定事务处理方式
    switch config.Propagation {
    case PropagationRequired:
        return executeRequired(db, config, fn, args)
    case PropagationRequiresNew:
        return executeRequiresNew(db, config, fn, args)
    case PropagationSupports:
        return executeSupports(db, config, fn, args)
    case PropagationNotSupported:
        return executeNotSupported(config, fn, args)
    case PropagationNever:
        return executeNever(config, fn, args)
    case PropagationNested:
        return executeNested(db, config, fn, args)
    default:
        return executeRequired(db, config, fn, args)
    }
}

func executeRequired(db *gorm.DB, config Transactional, fn reflect.Value, args []reflect.Value) []reflect.Value {
    // 检查是否已在事务中
    if isInTransaction(db) {
        return fn.Call(args)
    }
    
    // 开始新事务
    return executeInNewTransaction(db, config, fn, args)
}

func executeRequiresNew(db *gorm.DB, config Transactional, fn reflect.Value, args []reflect.Value) []reflect.Value {
    // 总是开始新事务
    return executeInNewTransaction(db, config, fn, args)
}

func executeInNewTransaction(db *gorm.DB, config Transactional, fn reflect.Value, args []reflect.Value) []reflect.Value {
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            panic(r)
        }
    }()
    
    // 设置事务隔离级别
    if config.Isolation != IsolationDefault {
        setTransactionIsolation(tx, config.Isolation)
    }
    
    // 替换args中的DB实例为事务实例
    newArgs := replaceDBInArgs(args, tx)
    
    // 执行函数
    results := fn.Call(newArgs)
    
    // 检查返回值中是否有错误
    if hasError(results) {
        tx.Rollback()
        return results
    }
    
    // 提交事务
    if err := tx.Commit().Error; err != nil {
        // 构造错误返回值
        return createErrorResult(fn.Type(), err)
    }
    
    return results
}

// 辅助函数
func isInTransaction(db *gorm.DB) bool {
    // 检查是否在事务中的逻辑
    return false // 简化实现
}

func setTransactionIsolation(tx *gorm.DB, isolation Isolation) {
    var level string
    switch isolation {
    case IsolationReadUncommitted:
        level = "READ UNCOMMITTED"
    case IsolationReadCommitted:
        level = "READ COMMITTED"
    case IsolationRepeatableRead:
        level = "REPEATABLE READ"
    case IsolationSerializable:
        level = "SERIALIZABLE"
    default:
        return
    }
    
    tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", level))
}

func replaceDBInArgs(args []reflect.Value, tx *gorm.DB) []reflect.Value {
    // 在参数中查找并替换*gorm.DB实例
    newArgs := make([]reflect.Value, len(args))
    copy(newArgs, args)
    
    for i, arg := range args {
        if arg.Type() == reflect.TypeOf((*gorm.DB)(nil)) {
            newArgs[i] = reflect.ValueOf(tx)
        }
    }
    
    return newArgs
}

func hasError(results []reflect.Value) bool {
    for _, result := range results {
        if result.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
            if !result.IsNil() {
                return true
            }
        }
    }
    return false
}

func createErrorResult(fnType reflect.Type, err error) []reflect.Value {
    results := make([]reflect.Value, fnType.NumOut())
    for i := 0; i < fnType.NumOut(); i++ {
        outType := fnType.Out(i)
        if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
            results[i] = reflect.ValueOf(err)
        } else {
            results[i] = reflect.Zero(outType)
        }
    }
    return results
}
```

### 2. 服务层事务应用

```go
// services/user_service.go
package services

import (
    "github.com/zsy619/yyhertz/models"
    "github.com/zsy619/yyhertz/transaction"
    "gorm.io/gorm"
)

type UserService struct {
    db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
    return &UserService{db: db}
}

// CreateUserWithProfile 创建用户和资料（声明式事务）
var CreateUserWithProfile = transaction.WithTransaction(transaction.Transactional{
    Propagation: transaction.PropagationRequired,
    Isolation:   transaction.IsolationReadCommitted,
    ReadOnly:    false,
    Timeout:     30,
})(func(s *UserService, user *models.User, profile *models.UserProfile) error {
    // 创建用户
    if err := s.db.Create(user).Error; err != nil {
        return err
    }
    
    // 设置关联ID
    profile.UserID = user.ID
    
    // 创建用户资料
    if err := s.db.Create(profile).Error; err != nil {
        return err
    }
    
    return nil
})

// UpdateUserBalance 更新用户余额（需要事务）
var UpdateUserBalance = transaction.WithTransaction(transaction.Transactional{
    Propagation: transaction.PropagationRequired,
    Isolation:   transaction.IsolationSerializable,
})(func(s *UserService, userID uint, amount float64) error {
    var user models.User
    
    // 锁定记录
    if err := s.db.Set("gorm:query_option", "FOR UPDATE").First(&user, userID).Error; err != nil {
        return err
    }
    
    // 检查余额
    if user.Balance+amount < 0 {
        return fmt.Errorf("余额不足")
    }
    
    // 更新余额
    user.Balance += amount
    if err := s.db.Save(&user).Error; err != nil {
        return err
    }
    
    // 记录交易历史
    transaction := models.Transaction{
        UserID: userID,
        Amount: amount,
        Type:   "balance_update",
    }
    
    if err := s.db.Create(&transaction).Error; err != nil {
        return err
    }
    
    return nil
})
```

## 🔧 高级事务模式

### 1. 分布式事务（Saga模式）

```go
// transaction/saga.go
package transaction

import (
    "context"
    "fmt"
    "log"
)

// SagaStep 事务步骤
type SagaStep struct {
    Name         string
    Execute      func(ctx context.Context) error
    Compensate   func(ctx context.Context) error
    CanRetry     bool
    MaxRetries   int
}

// Saga 分布式事务协调器
type Saga struct {
    steps       []SagaStep
    executedSteps []int
    ctx         context.Context
}

func NewSaga(ctx context.Context) *Saga {
    return &Saga{
        ctx:           ctx,
        steps:         make([]SagaStep, 0),
        executedSteps: make([]int, 0),
    }
}

// AddStep 添加事务步骤
func (s *Saga) AddStep(step SagaStep) *Saga {
    s.steps = append(s.steps, step)
    return s
}

// Execute 执行Saga事务
func (s *Saga) Execute() error {
    for i, step := range s.steps {
        if err := s.executeStep(i, step); err != nil {
            // 执行失败，开始补偿
            if compensateErr := s.compensate(); compensateErr != nil {
                log.Printf("Saga compensation failed: %v", compensateErr)
                return fmt.Errorf("saga execution failed: %v, compensation failed: %v", err, compensateErr)
            }
            return fmt.Errorf("saga execution failed: %v", err)
        }
        s.executedSteps = append(s.executedSteps, i)
    }
    
    log.Printf("Saga completed successfully with %d steps", len(s.steps))
    return nil
}

func (s *Saga) executeStep(index int, step SagaStep) error {
    retries := 0
    maxRetries := step.MaxRetries
    if maxRetries == 0 {
        maxRetries = 1
    }
    
    for retries < maxRetries {
        log.Printf("Executing saga step %d: %s (attempt %d)", index, step.Name, retries+1)
        
        if err := step.Execute(s.ctx); err != nil {
            if step.CanRetry && retries < maxRetries-1 {
                retries++
                log.Printf("Step %s failed, retrying... (%d/%d)", step.Name, retries, maxRetries)
                continue
            }
            return fmt.Errorf("step %s failed: %w", step.Name, err)
        }
        
        log.Printf("Step %s completed successfully", step.Name)
        return nil
    }
    
    return fmt.Errorf("step %s failed after %d retries", step.Name, maxRetries)
}

func (s *Saga) compensate() error {
    log.Printf("Starting saga compensation for %d executed steps", len(s.executedSteps))
    
    // 逆序执行补偿操作
    for i := len(s.executedSteps) - 1; i >= 0; i-- {
        stepIndex := s.executedSteps[i]
        step := s.steps[stepIndex]
        
        if step.Compensate != nil {
            log.Printf("Compensating step %d: %s", stepIndex, step.Name)
            if err := step.Compensate(s.ctx); err != nil {
                log.Printf("Compensation for step %s failed: %v", step.Name, err)
                return err
            }
            log.Printf("Step %s compensated successfully", step.Name)
        }
    }
    
    log.Printf("Saga compensation completed")
    return nil
}

// 使用示例
func CreateOrderSaga(ctx context.Context, order *models.Order) error {
    saga := NewSaga(ctx)
    
    saga.AddStep(SagaStep{
        Name: "ReserveInventory",
        Execute: func(ctx context.Context) error {
            return reserveInventory(order.Items)
        },
        Compensate: func(ctx context.Context) error {
            return releaseInventory(order.Items)
        },
        CanRetry:   true,
        MaxRetries: 3,
    }).AddStep(SagaStep{
        Name: "ProcessPayment",
        Execute: func(ctx context.Context) error {
            return processPayment(order.PaymentInfo)
        },
        Compensate: func(ctx context.Context) error {
            return refundPayment(order.PaymentInfo)
        },
        CanRetry:   true,
        MaxRetries: 2,
    }).AddStep(SagaStep{
        Name: "CreateOrder",
        Execute: func(ctx context.Context) error {
            return createOrder(order)
        },
        Compensate: func(ctx context.Context) error {
            return cancelOrder(order.ID)
        },
        CanRetry:   false,
    }).AddStep(SagaStep{
        Name: "SendNotification",
        Execute: func(ctx context.Context) error {
            return sendOrderNotification(order)
        },
        Compensate: func(ctx context.Context) error {
            return sendCancellationNotification(order)
        },
        CanRetry:   true,
        MaxRetries: 5,
    })
    
    return saga.Execute()
}
```

### 2. 事务监控和指标

```go
// transaction/monitor.go
package transaction

import (
    "sync"
    "time"
)

type TransactionMetrics struct {
    TotalTransactions    int64         `json:"total_transactions"`
    SuccessfulTransactions int64       `json:"successful_transactions"`
    FailedTransactions   int64         `json:"failed_transactions"`
    AverageLatency       time.Duration `json:"average_latency"`
    MaxLatency          time.Duration `json:"max_latency"`
    ActiveTransactions   int64         `json:"active_transactions"`
    mutex               sync.RWMutex
}

var metrics = &TransactionMetrics{}

// RecordTransaction 记录事务指标
func RecordTransaction(duration time.Duration, success bool) {
    metrics.mutex.Lock()
    defer metrics.mutex.Unlock()
    
    metrics.TotalTransactions++
    
    if success {
        metrics.SuccessfulTransactions++
    } else {
        metrics.FailedTransactions++
    }
    
    // 更新延迟统计
    if duration > metrics.MaxLatency {
        metrics.MaxLatency = duration
    }
    
    // 简单的移动平均
    metrics.AverageLatency = (metrics.AverageLatency + duration) / 2
}

// GetMetrics 获取事务指标
func GetMetrics() TransactionMetrics {
    metrics.mutex.RLock()
    defer metrics.mutex.RUnlock()
    
    return *metrics
}

// TransactionMonitor 事务监控器
type TransactionMonitor struct {
    startTime time.Time
    name      string
}

// StartTransaction 开始监控事务
func StartTransaction(name string) *TransactionMonitor {
    metrics.mutex.Lock()
    metrics.ActiveTransactions++
    metrics.mutex.Unlock()
    
    return &TransactionMonitor{
        startTime: time.Now(),
        name:      name,
    }
}

// End 结束事务监控
func (tm *TransactionMonitor) End(success bool) {
    duration := time.Since(tm.startTime)
    
    metrics.mutex.Lock()
    metrics.ActiveTransactions--
    metrics.mutex.Unlock()
    
    RecordTransaction(duration, success)
    
    if !success {
        log.Printf("Transaction %s failed after %v", tm.name, duration)
    } else {
        log.Printf("Transaction %s completed in %v", tm.name, duration)
    }
}
```

## 🎯 事务最佳实践

### 1. 事务边界设计

```go
// 正确的事务边界设计
func (s *UserService) TransferMoney(fromUserID, toUserID uint, amount float64) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 1. 锁定源用户账户
        var fromUser models.User
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            First(&fromUser, fromUserID).Error; err != nil {
            return err
        }
        
        // 2. 锁定目标用户账户
        var toUser models.User
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            First(&toUser, toUserID).Error; err != nil {
            return err
        }
        
        // 3. 检查余额
        if fromUser.Balance < amount {
            return fmt.Errorf("余额不足")
        }
        
        // 4. 执行转账
        fromUser.Balance -= amount
        toUser.Balance += amount
        
        // 5. 保存用户信息
        if err := tx.Save(&fromUser).Error; err != nil {
            return err
        }
        
        if err := tx.Save(&toUser).Error; err != nil {
            return err
        }
        
        // 6. 记录转账历史
        transfer := models.Transfer{
            FromUserID: fromUserID,
            ToUserID:   toUserID,
            Amount:     amount,
            Status:     "completed",
        }
        
        if err := tx.Create(&transfer).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### 2. 事务超时处理

```go
// transaction/timeout.go
package transaction

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    "gorm.io/gorm"
)

// WithTimeout 为事务添加超时控制
func WithTimeout(db *gorm.DB, timeout time.Duration, fn func(*gorm.DB) error) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    // 为GORM设置上下文
    db = db.WithContext(ctx)
    
    // 使用channel来处理超时
    errChan := make(chan error, 1)
    
    go func() {
        errChan <- fn(db)
    }()
    
    select {
    case err := <-errChan:
        return err
    case <-ctx.Done():
        return fmt.Errorf("transaction timeout after %v", timeout)
    }
}

// 使用示例
func (s *UserService) CreateUserWithTimeout(user *models.User) error {
    return WithTimeout(s.db, 30*time.Second, func(db *gorm.DB) error {
        return db.Transaction(func(tx *gorm.DB) error {
            // 长时间运行的事务操作
            if err := tx.Create(user).Error; err != nil {
                return err
            }
            
            // 模拟一些耗时操作
            time.Sleep(1 * time.Second)
            
            return nil
        })
    })
}
```

### 3. 死锁检测和处理

```go
// transaction/deadlock.go
package transaction

import (
    "errors"
    "strings"
    "time"
    
    "gorm.io/gorm"
)

// DeadlockRetry 死锁重试装饰器
func DeadlockRetry(maxRetries int, retryDelay time.Duration) func(func() error) error {
    return func(fn func() error) error {
        var lastErr error
        
        for i := 0; i <= maxRetries; i++ {
            lastErr = fn()
            
            if lastErr == nil {
                return nil
            }
            
            // 检查是否是死锁错误
            if !isDeadlockError(lastErr) {
                return lastErr
            }
            
            if i < maxRetries {
                log.Printf("Deadlock detected, retrying in %v (attempt %d/%d)", retryDelay, i+1, maxRetries)
                time.Sleep(retryDelay)
            }
        }
        
        return fmt.Errorf("operation failed after %d retries due to deadlock: %w", maxRetries, lastErr)
    }
}

func isDeadlockError(err error) bool {
    if err == nil {
        return false
    }
    
    errStr := strings.ToLower(err.Error())
    deadlockKeywords := []string{
        "deadlock",
        "lock wait timeout",
        "40001", // PostgreSQL serialization failure
        "1213",  // MySQL deadlock
    }
    
    for _, keyword := range deadlockKeywords {
        if strings.Contains(errStr, keyword) {
            return true
        }
    }
    
    return false
}

// 使用示例
func (s *UserService) TransferMoneyWithDeadlockRetry(fromUserID, toUserID uint, amount float64) error {
    return DeadlockRetry(3, 100*time.Millisecond)(func() error {
        return s.TransferMoney(fromUserID, toUserID, amount)
    })
}
```

## 🧪 事务测试

### 1. 事务测试工具

```go
// transaction/testing.go
package transaction

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "gorm.io/gorm"
)

// TestTransaction 事务测试辅助函数
func TestTransaction(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB) error) {
    // 开始事务
    tx := db.Begin()
    defer tx.Rollback() // 测试结束后总是回滚
    
    // 执行测试函数
    err := fn(tx)
    
    // 验证结果
    if err != nil {
        t.Errorf("Transaction test failed: %v", err)
    }
}

// AssertTransactionRollback 测试事务回滚
func AssertTransactionRollback(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB) error) {
    // 记录初始状态
    var initialCount int64
    db.Model(&models.User{}).Count(&initialCount)
    
    // 执行应该失败的事务
    err := db.Transaction(fn)
    
    // 验证事务失败
    assert.Error(t, err)
    
    // 验证数据未被修改
    var finalCount int64
    db.Model(&models.User{}).Count(&finalCount)
    assert.Equal(t, initialCount, finalCount, "Transaction should have been rolled back")
}

// 使用示例
func TestUserTransfer(t *testing.T) {
    db := setupTestDB()
    defer cleanupTestDB(db)
    
    service := NewUserService(db)
    
    // 创建测试用户
    user1 := &models.User{Balance: 100}
    user2 := &models.User{Balance: 50}
    
    db.Create(user1)
    db.Create(user2)
    
    t.Run("Successful Transfer", func(t *testing.T) {
        err := service.TransferMoney(user1.ID, user2.ID, 30)
        assert.NoError(t, err)
        
        // 验证余额
        db.First(user1, user1.ID)
        db.First(user2, user2.ID)
        
        assert.Equal(t, float64(70), user1.Balance)
        assert.Equal(t, float64(80), user2.Balance)
    })
    
    t.Run("Insufficient Balance", func(t *testing.T) {
        AssertTransactionRollback(t, db, func(tx *gorm.DB) error {
            return service.TransferMoney(user2.ID, user1.ID, 200)
        })
    })
}
```

## 📊 事务性能优化

### 1. 批量操作事务

```go
// 批量插入优化
func (s *UserService) CreateUsersBatch(users []models.User, batchSize int) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        for i := 0; i < len(users); i += batchSize {
            end := i + batchSize
            if end > len(users) {
                end = len(users)
            }
            
            batch := users[i:end]
            if err := tx.Create(&batch).Error; err != nil {
                return err
            }
        }
        return nil
    })
}

// 批量更新优化
func (s *UserService) UpdateUsersStatusBatch(userIDs []uint, status string) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        return tx.Model(&models.User{}).
            Where("id IN ?", userIDs).
            Update("status", status).Error
    })
}
```

### 2. 事务预处理语句

```go
// 使用预处理语句提高性能
func (s *UserService) BulkUpdateWithPreparedStmt(updates []UserUpdate) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // 准备更新语句
        stmt := tx.Session(&gorm.Session{PrepareStmt: true})
        
        for _, update := range updates {
            if err := stmt.Model(&models.User{}).
                Where("id = ?", update.ID).
                Updates(update.Data).Error; err != nil {
                return err
            }
        }
        
        return nil
    })
}
```

## 🔗 相关资源

- [GORM快速开始](./gorm-quickstart)
- [数据库配置](./database-config)
- [性能优化建议](../dev-tools/performance)
- [错误处理最佳实践](../advanced/error-handling)

---

> 💡 **提示**: 合理使用事务是保证数据一致性的关键，但过度使用会影响性能。建议根据业务需求选择合适的事务边界和隔离级别。
