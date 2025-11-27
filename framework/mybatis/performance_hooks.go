// Package mybatis 性能监控钩子
//
// 提供基于钩子系统的性能监控集成
package mybatis

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	
	"gorm.io/gorm"
)

// executionData 存储执行数据
type executionData struct {
	SQL        string
	Args       []any
	StatementID string
	DataSource string
}

// executionStore 存储当前执行的上下文信息
var executionStore = &sync.Map{}

// PerformanceMonitoringHook 创建性能监控钩子
func PerformanceMonitoringHook(monitor *PerformanceMonitor) (BeforeHook, AfterHook) {
	beforeHook := func(ctx context.Context, sql string, args []any) error {
		if monitor == nil {
			return nil
		}
		
		// 使用context的地址作为键来存储执行信息
		key := fmt.Sprintf("%p", ctx)
		execData := &executionData{
			SQL:         sql,
			Args:        args,
			StatementID: generateStatementID(sql),
			DataSource:  getDataSourceFromSQL(sql),
		}
		
		executionStore.Store(key, execData)
		return nil
	}
	
	afterHook := func(ctx context.Context, result any, duration time.Duration, err error) {
		if monitor == nil {
			return
		}
		
		// 获取并删除执行数据
		key := fmt.Sprintf("%p", ctx)
		if value, exists := executionStore.LoadAndDelete(key); exists {
			if execData, ok := value.(*executionData); ok {
				monitor.RecordExecution(
					execData.StatementID,
					execData.SQL,
					execData.Args,
					duration,
					execData.DataSource,
					err,
				)
				return
			}
		}
		
		// 如果无法获取执行信息，使用默认值
		monitor.RecordExecution(
			"sql.unknown",
			"Unknown SQL",
			[]any{},
			duration,
			"default",
			err,
		)
	}
	
	return beforeHook, afterHook
}

// generateStatementID 从SQL生成语句ID
func generateStatementID(sql string) string {
	sql = strings.TrimSpace(strings.ToUpper(sql))
	
	if strings.HasPrefix(sql, "SELECT") {
		return "sql.select"
	} else if strings.HasPrefix(sql, "INSERT") {
		return "sql.insert"
	} else if strings.HasPrefix(sql, "UPDATE") {
		return "sql.update"
	} else if strings.HasPrefix(sql, "DELETE") {
		return "sql.delete"
	} else {
		return "sql.other"
	}
}

// getDataSourceFromSQL 从SQL中推测数据源类型
func getDataSourceFromSQL(sql string) string {
	sql = strings.TrimSpace(strings.ToUpper(sql))
	
	if strings.HasPrefix(sql, "SELECT") {
		return "read"
	} else {
		return "write"
	}
}

// NewSessionWithPerformanceMonitoring 创建带性能监控的会话
func NewSessionWithPerformanceMonitoring(db *gorm.DB, config *PerformanceConfig) SimpleSession {
	if config == nil {
		config = DefaultPerformanceConfig()
	}
	
	session := NewSimpleSession(db)
	monitor := NewPerformanceMonitor(config)
	
	// 添加性能监控钩子，使用同一个监控器
	beforeHook, afterHook := PerformanceMonitoringHook(monitor)
	session = session.AddBeforeHook(beforeHook).AddAfterHook(afterHook)
	
	// 启用性能监控器，使用同一个监控器实例
	// 不使用session.EnablePerformanceMonitor，因为那会创建新的监控器
	// 而是直接设置监控器
	if defaultSess, ok := session.(*defaultSession); ok {
		defaultSess.performanceMonitor = monitor
	}
	
	return session
}