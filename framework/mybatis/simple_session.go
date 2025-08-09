// Package mybatis 简化版MyBatis实现
//
// 基于Gobatis设计理念，追求简洁性和Go语言惯用法
package mybatis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SimpleSession 简化版会话接口 - 只保留最核心的方法
type SimpleSession interface {
	SelectOne(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
	SelectList(ctx context.Context, sql string, args ...interface{}) ([]interface{}, error)
	SelectPage(ctx context.Context, sql string, page PageRequest, args ...interface{}) (*PageResult, error)
	Insert(ctx context.Context, sql string, args ...interface{}) (int64, error)
	Update(ctx context.Context, sql string, args ...interface{}) (int64, error)
	Delete(ctx context.Context, sql string, args ...interface{}) (int64, error)
	
	// 钩子方法
	AddBeforeHook(hook BeforeHook) SimpleSession
	AddAfterHook(hook AfterHook) SimpleSession
	
	// 配置方法
	DryRun(enabled bool) SimpleSession
	Debug(enabled bool) SimpleSession
}

// SessionConfig 会话配置
type SessionConfig struct {
	DryRun bool
	Debug  bool
	Logger *log.Logger
}

// defaultSession 默认会话实现
type defaultSession struct {
	db          *gorm.DB
	config      SessionConfig
	beforeHooks []BeforeHook
	afterHooks  []AfterHook
}

// BeforeHook 执行前钩子
type BeforeHook func(ctx context.Context, sql string, args []interface{}) error

// AfterHook 执行后钩子
type AfterHook func(ctx context.Context, result interface{}, duration time.Duration, err error)

// PageRequest 分页请求
type PageRequest struct {
	Page int `json:"page"` // 页码，从1开始
	Size int `json:"size"` // 每页大小
}

// PageResult 分页结果
type PageResult struct {
	Items      []interface{} `json:"items"`      // 数据列表
	Total      int64         `json:"total"`      // 总记录数
	Page       int           `json:"page"`       // 当前页码
	Size       int           `json:"size"`       // 每页大小
	TotalPages int           `json:"totalPages"` // 总页数
}

// NewSimpleSession 创建简化版会话
func NewSimpleSession(db *gorm.DB) SimpleSession {
	return &defaultSession{
		db: db,
		config: SessionConfig{
			Logger: log.Default(),
		},
		beforeHooks: make([]BeforeHook, 0),
		afterHooks:  make([]AfterHook, 0),
	}
}

// DryRun 设置DryRun模式
func (s *defaultSession) DryRun(enabled bool) SimpleSession {
	s.config.DryRun = enabled
	return s
}

// Debug 设置Debug模式
func (s *defaultSession) Debug(enabled bool) SimpleSession {
	s.config.Debug = enabled
	return s
}

// AddBeforeHook 添加执行前钩子
func (s *defaultSession) AddBeforeHook(hook BeforeHook) SimpleSession {
	s.beforeHooks = append(s.beforeHooks, hook)
	return s
}

// AddAfterHook 添加执行后钩子
func (s *defaultSession) AddAfterHook(hook AfterHook) SimpleSession {
	s.afterHooks = append(s.afterHooks, hook)
	return s
}

// SelectOne 查询单条记录
func (s *defaultSession) SelectOne(ctx context.Context, sql string, args ...interface{}) (interface{}, error) {
	results, err := s.SelectList(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	
	if len(results) == 0 {
		return nil, nil
	}
	
	if len(results) > 1 {
		return nil, fmt.Errorf("expected one result but found %d", len(results))
	}
	
	return results[0], nil
}

// SelectList 查询多条记录
func (s *defaultSession) SelectList(ctx context.Context, sql string, args ...interface{}) ([]interface{}, error) {
	startTime := time.Now()
	
	// 执行前钩子
	for _, hook := range s.beforeHooks {
		if err := hook(ctx, sql, args); err != nil {
			return nil, fmt.Errorf("before hook error: %w", err)
		}
	}
	
	var result []interface{}
	var err error
	
	if s.config.DryRun {
		// DryRun模式：只打印SQL，不实际执行
		s.logSQL("[DryRun]", sql, args)
		result = make([]interface{}, 0) // 返回空结果
	} else {
		// 实际执行查询
		if s.config.Debug {
			s.logSQL("[Debug]", sql, args)
		}
		
		var rows []map[string]interface{}
		err = s.db.Raw(sql, args...).Scan(&rows).Error
		if err != nil {
			s.logError("Query failed", err)
		} else {
			// 转换结果
			result = make([]interface{}, len(rows))
			for i, row := range rows {
				result[i] = row
			}
		}
	}
	
	duration := time.Since(startTime)
	
	// 执行后钩子
	for _, hook := range s.afterHooks {
		hook(ctx, result, duration, err)
	}
	
	return result, err
}

// SelectPage 分页查询
func (s *defaultSession) SelectPage(ctx context.Context, sql string, page PageRequest, args ...interface{}) (*PageResult, error) {
	// 参数验证
	if page.Page < 1 {
		page.Page = 1
	}
	if page.Size < 1 {
		page.Size = 10
	}
	if page.Size > 1000 {
		page.Size = 1000 // 防止过大的分页
	}
	
	startTime := time.Now()
	
	// 执行前钩子
	for _, hook := range s.beforeHooks {
		if err := hook(ctx, fmt.Sprintf("PAGE: %s", sql), args); err != nil {
			return nil, fmt.Errorf("before hook error: %w", err)
		}
	}
	
	var total int64
	var items []interface{}
	var err error
	
	if s.config.DryRun {
		// DryRun模式
		s.logSQL("[DryRun Count]", s.buildCountSQL(sql), args)
		s.logSQL("[DryRun Page]", s.buildPageSQL(sql, page), args)
		total = 0
		items = make([]interface{}, 0)
	} else {
		// 1. 查询总数
		countSQL := s.buildCountSQL(sql)
		if s.config.Debug {
			s.logSQL("[Debug Count]", countSQL, args)
		}
		
		err = s.db.Raw(countSQL, args...).Scan(&total).Error
		if err != nil {
			s.logError("Count query failed", err)
			return nil, err
		}
		
		// 2. 分页查询
		if total > 0 {
			pageSQL := s.buildPageSQL(sql, page)
			if s.config.Debug {
				s.logSQL("[Debug Page]", pageSQL, args)
			}
			
			var rows []map[string]interface{}
			err = s.db.Raw(pageSQL, args...).Scan(&rows).Error
			if err != nil {
				s.logError("Page query failed", err)
				return nil, err
			}
			
			// 转换结果
			items = make([]interface{}, len(rows))
			for i, row := range rows {
				items[i] = row
			}
		} else {
			items = make([]interface{}, 0)
		}
	}
	
	result := &PageResult{
		Items:      items,
		Total:      total,
		Page:       page.Page,
		Size:       page.Size,
		TotalPages: int((total + int64(page.Size) - 1) / int64(page.Size)),
	}
	
	duration := time.Since(startTime)
	
	// 执行后钩子
	for _, hook := range s.afterHooks {
		hook(ctx, result, duration, err)
	}
	
	return result, err
}

// Insert 插入记录
func (s *defaultSession) Insert(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	return s.executeUpdate(ctx, "INSERT", sql, args...)
}

// Update 更新记录
func (s *defaultSession) Update(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	return s.executeUpdate(ctx, "UPDATE", sql, args...)
}

// Delete 删除记录
func (s *defaultSession) Delete(ctx context.Context, sql string, args ...interface{}) (int64, error) {
	return s.executeUpdate(ctx, "DELETE", sql, args...)
}

// executeUpdate 执行更新操作
func (s *defaultSession) executeUpdate(ctx context.Context, operation, sql string, args ...interface{}) (int64, error) {
	startTime := time.Now()
	
	// 执行前钩子
	for _, hook := range s.beforeHooks {
		if err := hook(ctx, sql, args); err != nil {
			return 0, fmt.Errorf("before hook error: %w", err)
		}
	}
	
	var affectedRows int64
	var err error
	
	if s.config.DryRun {
		// DryRun模式：只打印SQL，不实际执行
		s.logSQL(fmt.Sprintf("[DryRun %s]", operation), sql, args)
		affectedRows = 0 // DryRun返回0
	} else {
		// 实际执行
		if s.config.Debug {
			s.logSQL(fmt.Sprintf("[Debug %s]", operation), sql, args)
		}
		
		result := s.db.Exec(sql, args...)
		err = result.Error
		if err != nil {
			s.logError(fmt.Sprintf("%s failed", operation), err)
		} else {
			affectedRows = result.RowsAffected
		}
	}
	
	duration := time.Since(startTime)
	
	// 执行后钩子
	for _, hook := range s.afterHooks {
		hook(ctx, affectedRows, duration, err)
	}
	
	return affectedRows, err
}

// buildCountSQL 构建count查询SQL
func (s *defaultSession) buildCountSQL(sql string) string {
	// 移除ORDER BY子句
	upperSQL := strings.ToUpper(sql)
	if orderByIndex := strings.LastIndex(upperSQL, "ORDER BY"); orderByIndex != -1 {
		// 简单检查ORDER BY是否在括号外
		if !s.isInsideParentheses(sql, orderByIndex) {
			sql = sql[:orderByIndex]
		}
	}
	
	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS count_table", sql)
}

// buildPageSQL 构建分页查询SQL
func (s *defaultSession) buildPageSQL(sql string, page PageRequest) string {
	offset := (page.Page - 1) * page.Size
	return fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, page.Size, offset)
}

// isInsideParentheses 检查位置是否在括号内
func (s *defaultSession) isInsideParentheses(sql string, pos int) bool {
	openCount := 0
	for i := 0; i < pos; i++ {
		switch sql[i] {
		case '(':
			openCount++
		case ')':
			openCount--
		}
	}
	return openCount > 0
}

// logSQL 记录SQL日志
func (s *defaultSession) logSQL(prefix, sql string, args []interface{}) {
	if len(args) > 0 {
		s.config.Logger.Printf("%s SQL: %s\nArgs: %+v", prefix, sql, args)
	} else {
		s.config.Logger.Printf("%s SQL: %s", prefix, sql)
	}
}

// logError 记录错误日志
func (s *defaultSession) logError(message string, err error) {
	s.config.Logger.Printf("ERROR: %s - %v", message, err)
}