// Package mybatis 简化版MyBatis实现
//
// 基于Gobatis设计理念，追求简洁性和Go语言惯用法
package mybatis

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SimpleSession 简化版会话接口 - 核心CRUD操作
type SimpleSession interface {
	// 基础CRUD方法
	SelectOne(ctx context.Context, sql string, args ...any) (any, error)
	SelectList(ctx context.Context, sql string, args ...any) ([]any, error)
	SelectPage(ctx context.Context, sql string, page PageRequest, args ...any) (*PageResult, error)
	Insert(ctx context.Context, sql string, args ...any) (int64, error)
	Update(ctx context.Context, sql string, args ...any) (int64, error)
	Delete(ctx context.Context, sql string, args ...any) (int64, error)

	// 批量操作方法
	BatchInsert(ctx context.Context, sql string, batchArgs [][]any) (int64, error)
	BatchUpdate(ctx context.Context, sql string, batchArgs [][]any) (int64, error)
	BatchDelete(ctx context.Context, sql string, batchArgs [][]any) (int64, error)

	// 存储过程调用方法
	CallStoredProc(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error)
	CallStoredProcWithMultiResults(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error)

	// 配置方法 - 返回自身类型以支持方法链
	Debug(enabled bool) SimpleSession
	DryRun(enabled bool) SimpleSession

	// 缓存管理方法
	EnableCache(config *CacheConfig) SimpleSession
	DisableCache() SimpleSession
	ClearCache() SimpleSession
	GetCacheStats() map[string]any

	// 钩子方法 - 返回自身类型以支持方法链
	AddBeforeHook(hook BeforeHook) SimpleSession
	AddAfterHook(hook AfterHook) SimpleSession
	
	// 性能监控方法
	EnablePerformanceMonitor(config *PerformanceConfig) SimpleSession
	DisablePerformanceMonitor() SimpleSession
	GetPerformanceStats() map[string]*SQLStats
	GetSlowQueries(limit int) []*SlowQuery
	GetPerformanceReport() *StatisticsReport
	ClearPerformanceStats() SimpleSession
}

// SessionConfig 会话配置
type SessionConfig struct {
	DryRun bool
	Debug  bool
	Logger *log.Logger
}

// defaultSession 默认会话实现
type defaultSession struct {
	db               *gorm.DB
	config           SessionConfig
	beforeHooks      []BeforeHook
	afterHooks       []AfterHook
	cacheManager     *CacheManager
	performanceMonitor *PerformanceMonitor // 性能监控器
}

// BeforeHook 执行前钩子
type BeforeHook func(ctx context.Context, sql string, args []any) error

// AfterHook 执行后钩子
type AfterHook func(ctx context.Context, result any, duration time.Duration, err error)

// PageRequest 分页请求
type PageRequest struct {
	Page int `json:"page"` // 页码，从1开始
	Size int `json:"size"` // 每页大小
}

// PageResult 分页结果
type PageResult struct {
	Items      []any `json:"items"`      // 数据列表
	Total      int64 `json:"total"`      // 总记录数
	Page       int   `json:"page"`       // 当前页码
	Size       int   `json:"size"`       // 每页大小
	TotalPages int   `json:"totalPages"` // 总页数
}

// StoredProcResult 存储过程执行结果
type StoredProcResult struct {
	OutputParams map[string]any   `json:"outputParams"` // 输出参数
	ResultSets   [][]map[string]any `json:"resultSets"`   // 多结果集
	RowsAffected int64              `json:"rowsAffected"` // 影响行数
}

// ProcParam 存储过程参数
type ProcParam struct {
	Name      string `json:"name"`      // 参数名
	Value     any    `json:"value"`     // 参数值
	Direction string `json:"direction"` // 参数方向: IN, OUT, INOUT
}

// NewSimpleSession 创建简化版会话
func NewSimpleSession(db *gorm.DB) SimpleSession {
	return &defaultSession{
		db: db,
		config: SessionConfig{
			Logger: log.Default(),
		},
		beforeHooks:  make([]BeforeHook, 0),
		afterHooks:   make([]AfterHook, 0),
		cacheManager: nil, // 默认不启用缓存
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
func (s *defaultSession) SelectOne(ctx context.Context, sql string, args ...any) (any, error) {
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
func (s *defaultSession) SelectList(ctx context.Context, sql string, args ...any) ([]any, error) {
	startTime := time.Now()

	// 检查缓存
	if s.cacheManager != nil {
		if cachedResult, found := s.cacheManager.Get(sql, args); found {
			if s.config.Debug {
				s.config.Logger.Printf("[Cache Hit] SQL: %s", sql)
			}
			
			// 执行后钩子（缓存命中）
			for _, hook := range s.afterHooks {
				hook(ctx, cachedResult, time.Since(startTime), nil)
			}
			
			if result, ok := cachedResult.([]any); ok {
				return result, nil
			}
		}
	}

	// 执行前钩子
	for _, hook := range s.beforeHooks {
		if err := hook(ctx, sql, args); err != nil {
			return nil, fmt.Errorf("before hook error: %w", err)
		}
	}

	var result []any
	var err error

	if s.config.DryRun {
		// DryRun模式：只打印SQL，不实际执行
		s.logSQL("[DryRun]", sql, args)
		result = make([]any, 0) // 返回空结果
	} else {
		// 实际执行查询
		if s.config.Debug {
			s.logSQL("[Debug]", sql, args)
		}

		var rows []map[string]any
		err = s.db.Raw(sql, args...).Scan(&rows).Error
		if err != nil {
			s.logError("Query failed", err)
		} else {
			// 转换结果
			result = make([]any, len(rows))
			for i, row := range rows {
				result[i] = row
			}
			
			// 将结果存入缓存（仅在成功时）
			if s.cacheManager != nil {
				s.cacheManager.Set(sql, args, result)
				if s.config.Debug {
					s.config.Logger.Printf("[Cache Set] SQL: %s", sql)
				}
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
func (s *defaultSession) SelectPage(ctx context.Context, sql string, page PageRequest, args ...any) (*PageResult, error) {
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
	var items []any
	var err error

	if s.config.DryRun {
		// DryRun模式
		s.logSQL("[DryRun Count]", s.buildCountSQL(sql), args)
		s.logSQL("[DryRun Page]", s.buildPageSQL(sql, page), args)
		total = 0
		items = make([]any, 0)
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

			var rows []map[string]any
			err = s.db.Raw(pageSQL, args...).Scan(&rows).Error
			if err != nil {
				s.logError("Page query failed", err)
				return nil, err
			}

			// 转换结果
			items = make([]any, len(rows))
			for i, row := range rows {
				items[i] = row
			}
		} else {
			items = make([]any, 0)
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
func (s *defaultSession) Insert(ctx context.Context, sql string, args ...any) (int64, error) {
	return s.executeUpdate(ctx, "INSERT", sql, args...)
}

// Update 更新记录
func (s *defaultSession) Update(ctx context.Context, sql string, args ...any) (int64, error) {
	return s.executeUpdate(ctx, "UPDATE", sql, args...)
}

// Delete 删除记录
func (s *defaultSession) Delete(ctx context.Context, sql string, args ...any) (int64, error) {
	return s.executeUpdate(ctx, "DELETE", sql, args...)
}

// executeUpdate 执行更新操作
func (s *defaultSession) executeUpdate(ctx context.Context, operation, sql string, args ...any) (int64, error) {
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
func (s *defaultSession) logSQL(prefix, sql string, args []any) {
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

// BatchInsert 批量插入记录
func (s *defaultSession) BatchInsert(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return s.executeBatchUpdate(ctx, "BATCH_INSERT", sql, batchArgs)
}

// BatchUpdate 批量更新记录
func (s *defaultSession) BatchUpdate(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return s.executeBatchUpdate(ctx, "BATCH_UPDATE", sql, batchArgs)
}

// BatchDelete 批量删除记录
func (s *defaultSession) BatchDelete(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return s.executeBatchUpdate(ctx, "BATCH_DELETE", sql, batchArgs)
}

// executeBatchUpdate 执行批量更新操作
func (s *defaultSession) executeBatchUpdate(ctx context.Context, operation, sql string, batchArgs [][]any) (int64, error) {
	if len(batchArgs) == 0 {
		return 0, nil
	}

	startTime := time.Now()
	
	// 执行前钩子 - 对于批量操作，传递第一组参数作为示例
	for _, hook := range s.beforeHooks {
		if err := hook(ctx, fmt.Sprintf("%s (batch size: %d)", sql, len(batchArgs)), batchArgs[0]); err != nil {
			return 0, fmt.Errorf("before hook error: %w", err)
		}
	}

	var totalAffectedRows int64
	var err error

	if s.config.DryRun {
		// DryRun模式：只打印SQL，不实际执行
		s.logBatchSQL(fmt.Sprintf("[DryRun %s]", operation), sql, batchArgs)
		totalAffectedRows = 0
	} else {
		// 实际执行批量操作
		if s.config.Debug {
			s.logBatchSQL(fmt.Sprintf("[Debug %s]", operation), sql, batchArgs)
		}

		// 在事务中执行批量操作以提升性能
		err = s.db.Transaction(func(tx *gorm.DB) error {
			for i, args := range batchArgs {
				result := tx.Exec(sql, args...)
				if result.Error != nil {
					return fmt.Errorf("batch operation failed at index %d: %w", i, result.Error)
				}
				totalAffectedRows += result.RowsAffected
			}
			return nil
		})

		if err != nil {
			s.logError(fmt.Sprintf("%s failed", operation), err)
		}
	}

	duration := time.Since(startTime)

	// 执行后钩子
	for _, hook := range s.afterHooks {
		hook(ctx, totalAffectedRows, duration, err)
	}

	return totalAffectedRows, err
}

// logBatchSQL 记录批量SQL日志
func (s *defaultSession) logBatchSQL(prefix, sql string, batchArgs [][]any) {
	if len(batchArgs) == 0 {
		s.config.Logger.Printf("%s SQL: %s (empty batch)", prefix, sql)
		return
	}

	// 只显示前3组参数，避免日志过长
	displayCount := len(batchArgs)
	if displayCount > 3 {
		displayCount = 3
	}

	s.config.Logger.Printf("%s SQL: %s", prefix, sql)
	s.config.Logger.Printf("Batch size: %d", len(batchArgs))
	
	for i := 0; i < displayCount; i++ {
		s.config.Logger.Printf("  Args[%d]: %+v", i, batchArgs[i])
	}
	
	if len(batchArgs) > 3 {
		s.config.Logger.Printf("  ... and %d more entries", len(batchArgs)-3)
	}
}

// CallStoredProc 调用存储过程（单结果集）
func (s *defaultSession) CallStoredProc(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error) {
	return s.executeStoredProc(ctx, procName, params, false)
}

// CallStoredProcWithMultiResults 调用存储过程（多结果集）
func (s *defaultSession) CallStoredProcWithMultiResults(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error) {
	return s.executeStoredProc(ctx, procName, params, true)
}

// executeStoredProc 执行存储过程的核心实现
func (s *defaultSession) executeStoredProc(ctx context.Context, procName string, params []ProcParam, multiResults bool) (*StoredProcResult, error) {
	startTime := time.Now()

	// 构建CALL语句和参数
	callSQL, args, outputIndexes, err := s.buildStoredProcCall(procName, params)
	if err != nil {
		return nil, fmt.Errorf("failed to build stored proc call: %w", err)
	}

	// 执行前钩子
	for _, hook := range s.beforeHooks {
		if err := hook(ctx, callSQL, args); err != nil {
			return nil, fmt.Errorf("before hook error: %w", err)
		}
	}

	var result *StoredProcResult
	var execErr error

	if s.config.DryRun {
		// DryRun模式：只打印调用，不实际执行
		s.logSQL("[DryRun CALL]", callSQL, args)
		result = &StoredProcResult{
			OutputParams: make(map[string]any),
			ResultSets:   [][]map[string]any{},
			RowsAffected: 0,
		}
	} else {
		// 实际执行存储过程
		if s.config.Debug {
			s.logSQL("[Debug CALL]", callSQL, args)
		}

		if multiResults {
			result, execErr = s.executeStoredProcWithMultiResults(ctx, callSQL, args, params, outputIndexes)
		} else {
			result, execErr = s.executeStoredProcSingleResult(ctx, callSQL, args, params, outputIndexes)
		}

		if execErr != nil {
			s.logError("Stored procedure execution failed", execErr)
		}
	}

	duration := time.Since(startTime)

	// 执行后钩子
	for _, hook := range s.afterHooks {
		hook(ctx, result, duration, execErr)
	}

	return result, execErr
}

// buildStoredProcCall 构建存储过程调用SQL
func (s *defaultSession) buildStoredProcCall(procName string, params []ProcParam) (string, []any, map[int]string, error) {
	if len(params) == 0 {
		return fmt.Sprintf("CALL %s()", procName), []any{}, map[int]string{}, nil
	}

	var placeholders []string
	var args []any
	outputIndexes := make(map[int]string)

	for i, param := range params {
		switch strings.ToUpper(param.Direction) {
		case "IN":
			placeholders = append(placeholders, "?")
			args = append(args, param.Value)
		case "OUT":
			// 对于输出参数，根据数据库类型使用不同的占位符
			placeholders = append(placeholders, "@"+param.Name)
			outputIndexes[i] = param.Name
		case "INOUT":
			// 输入输出参数，先设置值，后面会读取输出
			placeholders = append(placeholders, "?")
			args = append(args, param.Value)
			outputIndexes[i] = param.Name
		default:
			return "", nil, nil, fmt.Errorf("unsupported parameter direction: %s", param.Direction)
		}
	}

	callSQL := fmt.Sprintf("CALL %s(%s)", procName, strings.Join(placeholders, ", "))
	return callSQL, args, outputIndexes, nil
}

// executeStoredProcSingleResult 执行单结果集存储过程
func (s *defaultSession) executeStoredProcSingleResult(ctx context.Context, callSQL string, args []any, params []ProcParam, outputIndexes map[int]string) (*StoredProcResult, error) {
	result := &StoredProcResult{
		OutputParams: make(map[string]any),
		ResultSets:   [][]map[string]any{},
		RowsAffected: 0,
	}

	// 在事务中执行以支持输出参数
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 执行存储过程
		dbResult := tx.Exec(callSQL, args...)
		if dbResult.Error != nil {
			return dbResult.Error
		}
		result.RowsAffected = dbResult.RowsAffected

		// 如果有输出参数，需要查询获取
		if len(outputIndexes) > 0 {
			outputParams, err := s.fetchOutputParams(tx, outputIndexes)
			if err != nil {
				return fmt.Errorf("failed to fetch output parameters: %w", err)
			}
			result.OutputParams = outputParams
		}

		// 尝试获取结果集（如果存储过程返回结果）
		rows, err := tx.Raw("SELECT 1").Rows() // 占位查询检查是否有结果
		if err == nil {
			rows.Close()
			
			// 执行实际查询获取结果
			var resultSet []map[string]any
			err = tx.Raw(callSQL, args...).Scan(&resultSet).Error
			if err == nil && len(resultSet) > 0 {
				result.ResultSets = append(result.ResultSets, resultSet)
			}
		}

		return nil
	})

	return result, err
}

// executeStoredProcWithMultiResults 执行多结果集存储过程
func (s *defaultSession) executeStoredProcWithMultiResults(ctx context.Context, callSQL string, args []any, params []ProcParam, outputIndexes map[int]string) (*StoredProcResult, error) {
	result := &StoredProcResult{
		OutputParams: make(map[string]any),
		ResultSets:   [][]map[string]any{},
		RowsAffected: 0,
	}

	// 多结果集需要使用原生SQL连接
	sqlDB, err := s.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 执行存储过程
	rows, err := sqlDB.QueryContext(ctx, callSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute stored procedure: %w", err)
	}
	defer rows.Close()

	// 处理多个结果集
	for {
		// 获取列信息
		columns, err := rows.Columns()
		if err != nil {
			break
		}

		var resultSet []map[string]any
		
		// 读取当前结果集的所有行
		for rows.Next() {
			// 创建扫描目标
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			// 扫描行数据
			if err := rows.Scan(valuePtrs...); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}

			// 构建行映射
			row := make(map[string]any)
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			resultSet = append(resultSet, row)
		}

		// 将结果集添加到结果中
		if len(resultSet) > 0 {
			result.ResultSets = append(result.ResultSets, resultSet)
		}

		// 检查是否还有更多结果集
		if !rows.NextResultSet() {
			break
		}
	}

	// 处理输出参数（如果有）
	if len(outputIndexes) > 0 {
		outputParams, err := s.fetchOutputParamsNative(sqlDB, ctx, outputIndexes)
		if err != nil {
			s.config.Logger.Printf("Warning: failed to fetch output parameters: %v", err)
		} else {
			result.OutputParams = outputParams
		}
	}

	return result, rows.Err()
}

// fetchOutputParams 获取输出参数值（使用GORM）
func (s *defaultSession) fetchOutputParams(tx *gorm.DB, outputIndexes map[int]string) (map[string]any, error) {
	outputParams := make(map[string]any)
	
	for _, paramName := range outputIndexes {
		var value interface{}
		err := tx.Raw("SELECT @" + paramName).Scan(&value).Error
		if err != nil {
			return nil, fmt.Errorf("failed to fetch output parameter %s: %w", paramName, err)
		}
		outputParams[paramName] = value
	}
	
	return outputParams, nil
}

// fetchOutputParamsNative 获取输出参数值（使用原生连接）
func (s *defaultSession) fetchOutputParamsNative(db *sql.DB, ctx context.Context, outputIndexes map[int]string) (map[string]any, error) {
	outputParams := make(map[string]any)
	
	for _, paramName := range outputIndexes {
		var value interface{}
		err := db.QueryRowContext(ctx, "SELECT @"+paramName).Scan(&value)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch output parameter %s: %w", paramName, err)
		}
		outputParams[paramName] = value
	}
	
	return outputParams, nil
}

// 缓存管理方法实现

// EnableCache 启用缓存
func (s *defaultSession) EnableCache(config *CacheConfig) SimpleSession {
	if config == nil {
		config = DefaultCacheConfig()
	}
	s.cacheManager = NewCacheManager(config)
	return s
}

// DisableCache 禁用缓存
func (s *defaultSession) DisableCache() SimpleSession {
	if s.cacheManager != nil {
		s.cacheManager.Stop()
		s.cacheManager = nil
	}
	return s
}

// ClearCache 清空缓存
func (s *defaultSession) ClearCache() SimpleSession {
	if s.cacheManager != nil {
		s.cacheManager.ClearAll()
	}
	return s
}

// GetCacheStats 获取缓存统计
func (s *defaultSession) GetCacheStats() map[string]any {
	if s.cacheManager != nil {
		return s.cacheManager.GetStats()
	}
	return map[string]any{
		"cache_enabled": false,
	}
}

// 性能监控方法实现

// EnablePerformanceMonitor 启用性能监控
func (s *defaultSession) EnablePerformanceMonitor(config *PerformanceConfig) SimpleSession {
	s.performanceMonitor = NewPerformanceMonitor(config)
	return s
}

// DisablePerformanceMonitor 禁用性能监控
func (s *defaultSession) DisablePerformanceMonitor() SimpleSession {
	if s.performanceMonitor != nil {
		s.performanceMonitor.Close()
		s.performanceMonitor = nil
	}
	return s
}

// GetPerformanceStats 获取性能统计信息
func (s *defaultSession) GetPerformanceStats() map[string]*SQLStats {
	if s.performanceMonitor != nil {
		return s.performanceMonitor.GetStatistics()
	}
	return make(map[string]*SQLStats)
}

// GetSlowQueries 获取慢查询记录
func (s *defaultSession) GetSlowQueries(limit int) []*SlowQuery {
	if s.performanceMonitor != nil {
		return s.performanceMonitor.GetSlowQueries(limit)
	}
	return []*SlowQuery{}
}

// GetPerformanceReport 获取性能报告
func (s *defaultSession) GetPerformanceReport() *StatisticsReport {
	if s.performanceMonitor != nil {
		return s.performanceMonitor.GetStatisticsReport()
	}
	return &StatisticsReport{}
}

// ClearPerformanceStats 清空性能统计
func (s *defaultSession) ClearPerformanceStats() SimpleSession {
	if s.performanceMonitor != nil {
		s.performanceMonitor.ClearStatistics()
	}
	return s
}
