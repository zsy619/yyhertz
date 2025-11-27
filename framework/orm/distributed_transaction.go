// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/config"
)

// DistributedTransactionManager 分布式事务管理器
type DistributedTransactionManager struct {
	// 事务参与者
	participants []*TransactionParticipant
	// 事务超时时间
	timeout time.Duration
	// 事务ID
	transactionID string
	// 事务状态
	status string
	// 互斥锁
	mutex sync.RWMutex
	// 上下文
	ctx context.Context
	// 取消函数
	cancel context.CancelFunc
}

// TransactionParticipant 事务参与者
type TransactionParticipant struct {
	// 数据库连接
	DB *gorm.DB
	// 数据库名称
	Name string
	// 事务
	TX *gorm.DB
	// 准备状态
	Prepared bool
	// 提交状态
	Committed bool
	// 回滚状态
	RolledBack bool
	// 错误信息
	Error error
}

// NewDistributedTransactionManager 创建分布式事务管理器
func NewDistributedTransactionManager(timeout time.Duration) *DistributedTransactionManager {
	if timeout <= 0 {
		timeout = time.Second * 30 // 默认30秒超时
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	return &DistributedTransactionManager{
		participants:  make([]*TransactionParticipant, 0),
		timeout:       timeout,
		transactionID: fmt.Sprintf("tx-%d", time.Now().UnixNano()),
		status:        "created",
		ctx:           ctx,
		cancel:        cancel,
	}
}

// AddParticipant 添加事务参与者
func (dtm *DistributedTransactionManager) AddParticipant(db *gorm.DB, name string) *DistributedTransactionManager {
	dtm.mutex.Lock()
	defer dtm.mutex.Unlock()

	if dtm.status != "created" {
		config.Warnf("事务已经开始，无法添加参与者")
		return dtm
	}

	participant := &TransactionParticipant{
		DB:   db,
		Name: name,
	}

	dtm.participants = append(dtm.participants, participant)
	return dtm
}

// Begin 开始事务
func (dtm *DistributedTransactionManager) Begin() error {
	dtm.mutex.Lock()
	defer dtm.mutex.Unlock()

	if dtm.status != "created" {
		return fmt.Errorf("事务状态错误: %s", dtm.status)
	}

	if len(dtm.participants) == 0 {
		return errors.New("没有事务参与者")
	}

	// 为每个参与者开始事务
	for _, p := range dtm.participants {
		tx := p.DB.WithContext(dtm.ctx).Begin()
		if tx.Error != nil {
			p.Error = tx.Error
			config.Errorf("参与者 %s 开始事务失败: %v", p.Name, tx.Error)
			// 回滚已经开始的事务
			dtm.rollback()
			return tx.Error
		}
		p.TX = tx
	}

	dtm.status = "begun"
	config.Infof("分布式事务 %s 已开始，参与者数量: %d", dtm.transactionID, len(dtm.participants))
	return nil
}

// Prepare 准备阶段（两阶段提交的第一阶段）
func (dtm *DistributedTransactionManager) Prepare() error {
	dtm.mutex.Lock()
	defer dtm.mutex.Unlock()

	if dtm.status != "begun" {
		return fmt.Errorf("事务状态错误: %s", dtm.status)
	}

	// 准备阶段，检查所有参与者是否可以提交
	for _, p := range dtm.participants {
		// 这里可以执行一些准备工作，例如检查约束条件等
		// 在实际的两阶段提交中，这一步会将事务写入预提交日志
		// 但在GORM中，我们只能模拟这个过程
		p.Prepared = true
	}

	dtm.status = "prepared"
	config.Infof("分布式事务 %s 准备完成", dtm.transactionID)
	return nil
}

// Commit 提交事务（两阶段提交的第二阶段）
func (dtm *DistributedTransactionManager) Commit() error {
	dtm.mutex.Lock()
	defer dtm.mutex.Unlock()

	if dtm.status != "prepared" && dtm.status != "begun" {
		return fmt.Errorf("事务状态错误: %s", dtm.status)
	}

	var commitErrors []error

	// 提交所有参与者的事务
	for _, p := range dtm.participants {
		if p.TX == nil {
			continue
		}

		err := p.TX.Commit().Error
		if err != nil {
			p.Error = err
			commitErrors = append(commitErrors, fmt.Errorf("参与者 %s 提交失败: %w", p.Name, err))
			config.Errorf("参与者 %s 提交失败: %v", p.Name, err)
		} else {
			p.Committed = true
			config.Debugf("参与者 %s 提交成功", p.Name)
		}
	}

	if len(commitErrors) > 0 {
		dtm.status = "commit_failed"
		// 在实际的两阶段提交中，如果有参与者提交失败，
		// 我们应该尝试回滚所有参与者，但这可能导致数据不一致
		// 在这里，我们只记录错误，并返回第一个错误
		return commitErrors[0]
	}

	dtm.status = "committed"
	config.Infof("分布式事务 %s 提交成功", dtm.transactionID)
	return nil
}

// Rollback 回滚事务
func (dtm *DistributedTransactionManager) Rollback() error {
	dtm.mutex.Lock()
	defer dtm.mutex.Unlock()

	return dtm.rollback()
}

// rollback 内部回滚方法
func (dtm *DistributedTransactionManager) rollback() error {
	if dtm.status == "committed" || dtm.status == "rolledback" {
		return fmt.Errorf("事务状态错误: %s", dtm.status)
	}

	var rollbackErrors []error

	// 回滚所有参与者的事务
	for _, p := range dtm.participants {
		if p.TX == nil || p.Committed || p.RolledBack {
			continue
		}

		err := p.TX.Rollback().Error
		if err != nil {
			p.Error = err
			rollbackErrors = append(rollbackErrors, fmt.Errorf("参与者 %s 回滚失败: %w", p.Name, err))
			config.Errorf("参与者 %s 回滚失败: %v", p.Name, err)
		} else {
			p.RolledBack = true
			config.Debugf("参与者 %s 回滚成功", p.Name)
		}
	}

	dtm.status = "rolledback"
	config.Infof("分布式事务 %s 已回滚", dtm.transactionID)

	if len(rollbackErrors) > 0 {
		// 返回第一个错误
		return rollbackErrors[0]
	}

	return nil
}

// Close 关闭事务管理器
func (dtm *DistributedTransactionManager) Close() {
	dtm.mutex.Lock()
	defer dtm.mutex.Unlock()

	// 取消上下文
	if dtm.cancel != nil {
		dtm.cancel()
	}

	// 如果事务还没有提交或回滚，尝试回滚
	if dtm.status != "committed" && dtm.status != "rolledback" {
		_ = dtm.rollback()
	}

	config.Infof("分布式事务 %s 已关闭", dtm.transactionID)
}

// Status 获取事务状态
func (dtm *DistributedTransactionManager) Status() string {
	dtm.mutex.RLock()
	defer dtm.mutex.RUnlock()
	return dtm.status
}

// TransactionID 获取事务ID
func (dtm *DistributedTransactionManager) TransactionID() string {
	return dtm.transactionID
}

// WithContext 使用自定义上下文
func (dtm *DistributedTransactionManager) WithContext(ctx context.Context) *DistributedTransactionManager {
	dtm.mutex.Lock()
	defer dtm.mutex.Unlock()

	if dtm.status != "created" {
		config.Warnf("事务已经开始，无法更改上下文")
		return dtm
	}

	// 取消旧的上下文
	if dtm.cancel != nil {
		dtm.cancel()
	}

	// 创建新的上下文
	newCtx, cancel := context.WithTimeout(ctx, dtm.timeout)
	dtm.ctx = newCtx
	dtm.cancel = cancel

	return dtm
}

// Execute 执行事务函数
func (dtm *DistributedTransactionManager) Execute(fn func() error) error {
	// 开始事务
	if err := dtm.Begin(); err != nil {
		return err
	}

	// 确保事务最终会被关闭
	defer dtm.Close()

	// 执行业务逻辑
	err := fn()
	if err != nil {
		// 业务逻辑失败，回滚事务
		rollbackErr := dtm.Rollback()
		if rollbackErr != nil {
			// 回滚也失败了，返回组合错误
			return fmt.Errorf("业务逻辑失败: %v, 回滚失败: %v", err, rollbackErr)
		}
		return err
	}

	// 准备提交
	if err := dtm.Prepare(); err != nil {
		// 准备失败，回滚事务
		_ = dtm.Rollback()
		return err
	}

	// 提交事务
	return dtm.Commit()
}

// ============= 便捷函数 =============

// ExecuteDistributedTransaction 执行分布式事务
func ExecuteDistributedTransaction(timeout time.Duration, fn func(dtm *DistributedTransactionManager) error) error {
	dtm := NewDistributedTransactionManager(timeout)
	defer dtm.Close()

	return fn(dtm)
}

// ExecuteAcrossDBs 在多个数据库之间执行事务
func ExecuteAcrossDBs(dbs map[string]*gorm.DB, fn func(txMap map[string]*gorm.DB) error) error {
	return ExecuteDistributedTransaction(time.Second*30, func(dtm *DistributedTransactionManager) error {
		// 添加所有数据库作为参与者
		for name, db := range dbs {
			dtm.AddParticipant(db, name)
		}

		// 开始事务
		if err := dtm.Begin(); err != nil {
			return err
		}

		// 创建事务映射
		txMap := make(map[string]*gorm.DB)
		for _, p := range dtm.participants {
			txMap[p.Name] = p.TX
		}

		// 执行业务逻辑
		err := fn(txMap)
		if err != nil {
			// 业务逻辑失败，回滚事务
			_ = dtm.Rollback()
			return err
		}

		// 准备提交
		if err := dtm.Prepare(); err != nil {
			// 准备失败，回滚事务
			_ = dtm.Rollback()
			return err
		}

		// 提交事务
		return dtm.Commit()
	})
}
