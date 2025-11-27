package orm

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// TransactionManager 事务管理器
type TransactionManager struct {
	db *gorm.DB
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(db *gorm.DB) *TransactionManager {
	if db == nil {
		db = GetDefaultORM().DB()
	}
	return &TransactionManager{db: db}
}

// GetDefaultTransactionManager 获取默认事务管理器
func GetDefaultTransactionManager() *TransactionManager {
	return NewTransactionManager(GetDefaultORM().DB())
}

// Transaction 执行事务
func (tm *TransactionManager) Transaction(fn func(tx *gorm.DB) error) error {
	return tm.db.Transaction(fn)
}

// TransactionWithContext 使用上下文执行事务
func (tm *TransactionManager) TransactionWithContext(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(fn)
}

// Begin 开始事务
func (tm *TransactionManager) Begin() *gorm.DB {
	return tm.db.Begin()
}

// BeginWithContext 使用上下文开始事务
func (tm *TransactionManager) BeginWithContext(ctx context.Context) *gorm.DB {
	return tm.db.WithContext(ctx).Begin()
}

// Commit 提交事务
func (tm *TransactionManager) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

// Rollback 回滚事务
func (tm *TransactionManager) Rollback(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// ============= 事务装饰器和辅助函数 =============

// TransactionFunc 事务函数类型
type TransactionFunc func(tx *gorm.DB) error

// WithTransaction 事务装饰器
func WithTransaction(fn TransactionFunc) error {
	return GetDefaultTransactionManager().Transaction(fn)
}

// WithTransactionContext 带上下文的事务装饰器
func WithTransactionContext(ctx context.Context, fn TransactionFunc) error {
	return GetDefaultTransactionManager().TransactionWithContext(ctx, fn)
}

// SafeTransaction 安全事务执行，带错误恢复
func SafeTransaction(fn TransactionFunc) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("transaction panic recovered: %v", r)
			config.Errorf("Transaction panic: %v", r)
		}
	}()

	return WithTransaction(func(tx *gorm.DB) error {
		config.Debug("Starting safe transaction")

		err := fn(tx)
		if err != nil {
			config.Errorf("Transaction failed: %v", err)
			return err
		}

		config.Debug("Safe transaction completed successfully")
		return nil
	})
}

// ============= 事务上下文管理 =============

type transactionContextKey struct{}

// TransactionContext 事务上下文
type TransactionContext struct {
	TX *gorm.DB
}

// WithTransactionInContext 将事务存储到上下文中
func WithTransactionInContext(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, transactionContextKey{}, &TransactionContext{TX: tx})
}

// TransactionFromContext 从上下文中获取事务
func TransactionFromContext(ctx context.Context) (*gorm.DB, bool) {
	if txCtx, ok := ctx.Value(transactionContextKey{}).(*TransactionContext); ok && txCtx.TX != nil {
		return txCtx.TX, true
	}
	return nil, false
}

// GetDBFromContext 从上下文获取数据库连接（优先使用事务）
func GetDBFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := TransactionFromContext(ctx); ok {
		return tx
	}
	return GetDefaultORM().DB().WithContext(ctx)
}

// ============= 高级事务模式 =============

// NestedTransactionManager 嵌套事务管理器
type NestedTransactionManager struct {
	*TransactionManager
	savepoints []string
}

// NewNestedTransactionManager 创建嵌套事务管理器
func NewNestedTransactionManager(db *gorm.DB) *NestedTransactionManager {
	return &NestedTransactionManager{
		TransactionManager: NewTransactionManager(db),
		savepoints:         make([]string, 0),
	}
}

// SavePoint 创建保存点
func (ntm *NestedTransactionManager) SavePoint(tx *gorm.DB, name string) error {
	err := tx.SavePoint(name).Error
	if err == nil {
		ntm.savepoints = append(ntm.savepoints, name)
		config.Debugf("Created savepoint: %s", name)
	}
	return err
}

// RollbackTo 回滚到保存点
func (ntm *NestedTransactionManager) RollbackTo(tx *gorm.DB, name string) error {
	err := tx.RollbackTo(name).Error
	if err == nil {
		// 移除该保存点之后的所有保存点
		for i, sp := range ntm.savepoints {
			if sp == name {
				ntm.savepoints = ntm.savepoints[:i+1]
				break
			}
		}
		config.Debugf("Rolled back to savepoint: %s", name)
	}
	return err
}

// ============= 批量操作事务支持 =============

// BatchOperation 批量操作接口
type BatchOperation interface {
	Execute(tx *gorm.DB) error
	GetDescription() string
}

// BatchOperationFunc 批量操作函数类型
type BatchOperationFunc struct {
	Fn   func(tx *gorm.DB) error
	Desc string
}

// Execute 执行操作
func (bof BatchOperationFunc) Execute(tx *gorm.DB) error {
	return bof.Fn(tx)
}

// GetDescription 获取描述
func (bof BatchOperationFunc) GetDescription() string {
	return bof.Desc
}

// BatchTransactionManager 批量事务管理器
type BatchTransactionManager struct {
	*TransactionManager
	operations []BatchOperation
}

// NewBatchTransactionManager 创建批量事务管理器
func NewBatchTransactionManager(db *gorm.DB) *BatchTransactionManager {
	return &BatchTransactionManager{
		TransactionManager: NewTransactionManager(db),
		operations:         make([]BatchOperation, 0),
	}
}

// AddOperation 添加操作
func (btm *BatchTransactionManager) AddOperation(op BatchOperation) *BatchTransactionManager {
	btm.operations = append(btm.operations, op)
	return btm
}

// AddOperationFunc 添加操作函数
func (btm *BatchTransactionManager) AddOperationFunc(desc string, fn func(tx *gorm.DB) error) *BatchTransactionManager {
	return btm.AddOperation(BatchOperationFunc{Fn: fn, Desc: desc})
}

// Execute 执行所有操作
func (btm *BatchTransactionManager) Execute() error {
	if len(btm.operations) == 0 {
		return fmt.Errorf("no operations to execute")
	}

	return btm.Transaction(func(tx *gorm.DB) error {
		config.Infof("Executing batch transaction with %d operations", len(btm.operations))

		for i, op := range btm.operations {
			config.Debugf("Executing operation %d/%d: %s", i+1, len(btm.operations), op.GetDescription())

			if err := op.Execute(tx); err != nil {
				config.Errorf("Batch operation failed at step %d (%s): %v", i+1, op.GetDescription(), err)
				return fmt.Errorf("batch operation failed at step %d (%s): %w", i+1, op.GetDescription(), err)
			}

			config.Debugf("Operation %d completed successfully", i+1)
		}

		config.Info("All batch operations completed successfully")
		return nil
	})
}

// Clear 清空操作列表
func (btm *BatchTransactionManager) Clear() *BatchTransactionManager {
	btm.operations = btm.operations[:0]
	return btm
}

// ============= 事务统计和监控 =============

// TransactionStats 事务统计信息
type TransactionStats struct {
	TotalTransactions    int64 `json:"total_transactions"`
	SuccessTransactions  int64 `json:"success_transactions"`
	FailedTransactions   int64 `json:"failed_transactions"`
	ActiveTransactions   int64 `json:"active_transactions"`
	AverageExecutionTime int64 `json:"average_execution_time_ms"`
}

// TransactionMonitor 事务监控器
type TransactionMonitor struct {
	stats *TransactionStats
}

// NewTransactionMonitor 创建事务监控器
func NewTransactionMonitor() *TransactionMonitor {
	return &TransactionMonitor{
		stats: &TransactionStats{},
	}
}

// GetStats 获取统计信息
func (tm *TransactionMonitor) GetStats() *TransactionStats {
	return tm.stats
}

// ============= 便捷函数 =============

// CreateInTransaction 在事务中创建记录
func CreateInTransaction(model any) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return tx.Create(model).Error
	})
}

// UpdateInTransaction 在事务中更新记录
func UpdateInTransaction(model any) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return tx.Save(model).Error
	})
}

// DeleteInTransaction 在事务中删除记录
func DeleteInTransaction(model any, conditions ...any) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return tx.Delete(model, conditions...).Error
	})
}

// BulkCreateInTransaction 在事务中批量创建记录
func BulkCreateInTransaction(models any, batchSize int) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(models, batchSize).Error
	})
}

// ExecuteInTransaction 在事务中执行SQL
func ExecuteInTransaction(sql string, values ...any) error {
	return WithTransaction(func(tx *gorm.DB) error {
		return tx.Exec(sql, values...).Error
	})
}
