// Package mybatis 事务追踪和管理
//
// 使用context.Context实现Go风格的事务管理，避免ThreadLocal
package mybatis

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

// TransactionManager 事务管理器
type TransactionManager struct {
	db      *gorm.DB
	tracker *TransactionTracker
}

// TransactionTracker 事务追踪器
type TransactionTracker struct {
	transactions map[string]*TransactionInfo
	mutex        sync.RWMutex
}

// TransactionInfo 事务信息
type TransactionInfo struct {
	ID        string            // 事务ID
	StartTime time.Time         // 开始时间
	UserID    string            // 用户ID
	Status    TransactionStatus // 事务状态
	Operations []Operation      // 操作记录
}

// TransactionStatus 事务状态
type TransactionStatus int

const (
	TransactionActive TransactionStatus = iota
	TransactionCommitted
	TransactionRollbacked
)

func (s TransactionStatus) String() string {
	switch s {
	case TransactionActive:
		return "ACTIVE"
	case TransactionCommitted:
		return "COMMITTED"
	case TransactionRollbacked:
		return "ROLLBACKED"
	default:
		return "UNKNOWN"
	}
}

// Operation 操作记录
type Operation struct {
	Type      string    // 操作类型
	SQL       string    // SQL语句
	Args      []interface{} // 参数
	Timestamp time.Time // 执行时间
	Duration  time.Duration // 执行耗时
	Error     error     // 错误信息
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	return &TransactionManager{
		db:      db,
		tracker: NewTransactionTracker(),
	}
}

// NewTransactionTracker 创建事务追踪器
func NewTransactionTracker() *TransactionTracker {
	return &TransactionTracker{
		transactions: make(map[string]*TransactionInfo),
	}
}

// BeginTransaction 开始事务
func (tm *TransactionManager) BeginTransaction(ctx context.Context, userID string) (context.Context, error) {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	
	txID := generateTransactionID()
	txInfo := &TransactionInfo{
		ID:         txID,
		StartTime:  time.Now(),
		UserID:     userID,
		Status:     TransactionActive,
		Operations: make([]Operation, 0),
	}
	
	tm.tracker.addTransaction(txInfo)
	
	// 将事务信息存储到context中
	ctx = context.WithValue(ctx, TxKey, tx)
	ctx = context.WithValue(ctx, "tx_id", txID)
	ctx = context.WithValue(ctx, "tx_info", txInfo)
	
	log.Printf("[TRANSACTION] Started transaction %s for user %s", txID, userID)
	return ctx, nil
}

// CommitTransaction 提交事务
func (tm *TransactionManager) CommitTransaction(ctx context.Context) error {
	tx, txInfo, err := tm.getTransactionFromContext(ctx)
	if err != nil {
		return err
	}
	
	err = tx.Commit().Error
	if err != nil {
		txInfo.Status = TransactionRollbacked
		tm.tracker.updateTransaction(txInfo)
		return fmt.Errorf("failed to commit transaction %s: %w", txInfo.ID, err)
	}
	
	txInfo.Status = TransactionCommitted
	tm.tracker.updateTransaction(txInfo)
	
	duration := time.Since(txInfo.StartTime)
	log.Printf("[TRANSACTION] Committed transaction %s for user %s in %v", 
		txInfo.ID, txInfo.UserID, duration)
	
	return nil
}

// RollbackTransaction 回滚事务
func (tm *TransactionManager) RollbackTransaction(ctx context.Context) error {
	tx, txInfo, err := tm.getTransactionFromContext(ctx)
	if err != nil {
		return err
	}
	
	err = tx.Rollback().Error
	if err != nil {
		return fmt.Errorf("failed to rollback transaction %s: %w", txInfo.ID, err)
	}
	
	txInfo.Status = TransactionRollbacked
	tm.tracker.updateTransaction(txInfo)
	
	duration := time.Since(txInfo.StartTime)
	log.Printf("[TRANSACTION] Rollbacked transaction %s for user %s in %v", 
		txInfo.ID, txInfo.UserID, duration)
	
	return nil
}

// ExecuteInTransaction 在事务中执行操作
func (tm *TransactionManager) ExecuteInTransaction(ctx context.Context, userID string, fn func(context.Context, SimpleSession) error) error {
	// 检查是否已经在事务中
	if IsInTransaction(ctx) {
		// 已经在事务中，直接执行
		session := NewSimpleSession(GetTransactionDB(ctx))
		return fn(ctx, session)
	}
	
	// 开始新事务
	txCtx, err := tm.BeginTransaction(ctx, userID)
	if err != nil {
		return err
	}
	
	// 创建使用事务DB的session
	session := NewSimpleSession(GetTransactionDB(txCtx))
	
	// 执行操作
	err = fn(txCtx, session)
	if err != nil {
		// 回滚事务
		if rollbackErr := tm.RollbackTransaction(txCtx); rollbackErr != nil {
			log.Printf("[ERROR] Failed to rollback transaction: %v", rollbackErr)
		}
		return err
	}
	
	// 提交事务
	return tm.CommitTransaction(txCtx)
}

// RecordOperation 记录操作
func (tm *TransactionManager) RecordOperation(ctx context.Context, operationType, sql string, args []interface{}, duration time.Duration, err error) {
	if !IsInTransaction(ctx) {
		return
	}
	
	txInfo := GetTransactionInfo(ctx)
	if txInfo == nil {
		return
	}
	
	operation := Operation{
		Type:      operationType,
		SQL:       sql,
		Args:      args,
		Timestamp: time.Now(),
		Duration:  duration,
		Error:     err,
	}
	
	tm.tracker.addOperation(txInfo.ID, operation)
}

// getTransactionFromContext 从context获取事务信息
func (tm *TransactionManager) getTransactionFromContext(ctx context.Context) (*gorm.DB, *TransactionInfo, error) {
	tx, ok := ctx.Value(TxKey).(*gorm.DB)
	if !ok {
		return nil, nil, fmt.Errorf("no active transaction found in context")
	}
	
	txInfo, ok := ctx.Value("tx_info").(*TransactionInfo)
	if !ok {
		return nil, nil, fmt.Errorf("no transaction info found in context")
	}
	
	return tx, txInfo, nil
}

// 事务追踪器方法

// addTransaction 添加事务
func (tt *TransactionTracker) addTransaction(txInfo *TransactionInfo) {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()
	tt.transactions[txInfo.ID] = txInfo
}

// updateTransaction 更新事务
func (tt *TransactionTracker) updateTransaction(txInfo *TransactionInfo) {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()
	tt.transactions[txInfo.ID] = txInfo
}

// addOperation 添加操作记录
func (tt *TransactionTracker) addOperation(txID string, operation Operation) {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()
	
	if txInfo, exists := tt.transactions[txID]; exists {
		txInfo.Operations = append(txInfo.Operations, operation)
	}
}

// GetTransaction 获取事务信息
func (tt *TransactionTracker) GetTransaction(txID string) (*TransactionInfo, bool) {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()
	txInfo, exists := tt.transactions[txID]
	return txInfo, exists
}

// GetAllTransactions 获取所有事务信息
func (tt *TransactionTracker) GetAllTransactions() map[string]*TransactionInfo {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()
	
	result := make(map[string]*TransactionInfo)
	for k, v := range tt.transactions {
		result[k] = v
	}
	return result
}

// GetActiveTransactions 获取活跃事务
func (tt *TransactionTracker) GetActiveTransactions() []*TransactionInfo {
	tt.mutex.RLock()
	defer tt.mutex.RUnlock()
	
	var active []*TransactionInfo
	for _, txInfo := range tt.transactions {
		if txInfo.Status == TransactionActive {
			active = append(active, txInfo)
		}
	}
	return active
}

// CleanupOldTransactions 清理旧事务记录
func (tt *TransactionTracker) CleanupOldTransactions(maxAge time.Duration) int {
	tt.mutex.Lock()
	defer tt.mutex.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	cleaned := 0
	
	for txID, txInfo := range tt.transactions {
		if txInfo.Status != TransactionActive && txInfo.StartTime.Before(cutoff) {
			delete(tt.transactions, txID)
			cleaned++
		}
	}
	
	return cleaned
}

// 工具函数

// IsInTransaction 检查是否在事务中
func IsInTransaction(ctx context.Context) bool {
	_, ok := ctx.Value(TxKey).(*gorm.DB)
	return ok
}

// GetTransactionDB 获取事务DB
func GetTransactionDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(TxKey).(*gorm.DB); ok {
		return tx
	}
	return nil
}

// GetTransactionInfo 获取事务信息
func GetTransactionInfo(ctx context.Context) *TransactionInfo {
	if txInfo, ok := ctx.Value("tx_info").(*TransactionInfo); ok {
		return txInfo
	}
	return nil
}

// GetTransactionID 获取事务ID
func GetTransactionID(ctx context.Context) string {
	if txID, ok := ctx.Value("tx_id").(string); ok {
		return txID
	}
	return ""
}

// generateTransactionID 生成事务ID
func generateTransactionID() string {
	return fmt.Sprintf("tx_%d", time.Now().UnixNano())
}

// TransactionAwareSession 支持事务的会话包装器
type TransactionAwareSession struct {
	SimpleSession
	tm *TransactionManager
}

// NewTransactionAwareSession 创建支持事务的会话
func NewTransactionAwareSession(db *gorm.DB) *TransactionAwareSession {
	return &TransactionAwareSession{
		SimpleSession: NewSimpleSession(db),
		tm:            NewTransactionManager(db),
	}
}

// BeginTransaction 开始事务
func (tas *TransactionAwareSession) BeginTransaction(ctx context.Context, userID string) (context.Context, error) {
	return tas.tm.BeginTransaction(ctx, userID)
}

// CommitTransaction 提交事务
func (tas *TransactionAwareSession) CommitTransaction(ctx context.Context) error {
	return tas.tm.CommitTransaction(ctx)
}

// RollbackTransaction 回滚事务
func (tas *TransactionAwareSession) RollbackTransaction(ctx context.Context) error {
	return tas.tm.RollbackTransaction(ctx)
}

// ExecuteInTransaction 在事务中执行
func (tas *TransactionAwareSession) ExecuteInTransaction(ctx context.Context, userID string, fn func(context.Context, SimpleSession) error) error {
	return tas.tm.ExecuteInTransaction(ctx, userID, fn)
}

// GetTransactionTracker 获取事务追踪器
func (tas *TransactionAwareSession) GetTransactionTracker() *TransactionTracker {
	return tas.tm.tracker
}