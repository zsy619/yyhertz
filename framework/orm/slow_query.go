// Package orm 提供基于GORM的数据库ORM集成
package orm

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// SlowQueryMonitor 慢查询监控器
type SlowQueryMonitor struct {
	// 慢查询阈值
	threshold time.Duration
	// 慢查询记录
	queries []*SlowQueryRecord
	// 最大记录数
	maxRecords int
	// 互斥锁
	mutex sync.RWMutex
	// 是否启用
	enabled bool
	// 是否自动打印
	autoPrint bool
	// 自动打印间隔
	printInterval time.Duration
	// 停止通道
	stopChan chan struct{}
}

// SlowQueryRecord 慢查询记录
type SlowQueryRecord struct {
	// SQL语句
	SQL string `json:"sql"`
	// 参数
	Params []interface{} `json:"params,omitempty"`
	// 执行时间
	Duration time.Duration `json:"duration"`
	// 执行时间点
	Timestamp time.Time `json:"timestamp"`
	// 调用栈
	CallStack string `json:"call_stack,omitempty"`
	// 数据库名称
	Database string `json:"database"`
	// 表名
	Table string `json:"table"`
	// 操作类型
	Operation string `json:"operation"`
	// 影响行数
	RowsAffected int64 `json:"rows_affected"`
}

// NewSlowQueryMonitor 创建慢查询监控器
func NewSlowQueryMonitor(threshold time.Duration, maxRecords int) *SlowQueryMonitor {
	if threshold <= 0 {
		threshold = time.Millisecond * 500 // 默认500毫秒
	}

	if maxRecords <= 0 {
		maxRecords = 100 // 默认最多保存100条记录
	}

	return &SlowQueryMonitor{
		threshold:     threshold,
		queries:       make([]*SlowQueryRecord, 0, maxRecords),
		maxRecords:    maxRecords,
		enabled:       true,
		autoPrint:     false,
		printInterval: time.Minute * 5,
		stopChan:      make(chan struct{}),
	}
}

// Start 启动慢查询监控
func (m *SlowQueryMonitor) Start() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.enabled = true
	config.Infof("慢查询监控已启动，阈值: %v", m.threshold)
}

// Stop 停止慢查询监控
func (m *SlowQueryMonitor) Stop() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.enabled = false
	config.Info("慢查询监控已停止")
}

// EnableAutoPrint 启用自动打印
func (m *SlowQueryMonitor) EnableAutoPrint(interval time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if interval <= 0 {
		interval = time.Minute * 5 // 默认5分钟
	}

	m.autoPrint = true
	m.printInterval = interval

	// 启动自动打印协程
	go m.autoPrintLoop()
}

// DisableAutoPrint 禁用自动打印
func (m *SlowQueryMonitor) DisableAutoPrint() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.autoPrint = false
	close(m.stopChan)
	m.stopChan = make(chan struct{})
}

// autoPrintLoop 自动打印循环
func (m *SlowQueryMonitor) autoPrintLoop() {
	ticker := time.NewTicker(m.printInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.PrintSlowQueries(10) // 打印前10条最慢的查询
		case <-m.stopChan:
			return
		}
	}
}

// RecordQuery 记录查询
func (m *SlowQueryMonitor) RecordQuery(sql string, duration time.Duration, params ...interface{}) {
	m.mutex.RLock()
	if !m.enabled || duration < m.threshold {
		m.mutex.RUnlock()
		return
	}
	m.mutex.RUnlock()

	// 解析SQL获取表名和操作类型
	table, operation := parseSQL(sql)

	record := &SlowQueryRecord{
		SQL:          sql,
		Params:       params,
		Duration:     duration,
		Timestamp:    time.Now(),
		CallStack:    getCallStack(),
		Database:     "", // 可以从上下文中获取
		Table:        table,
		Operation:    operation,
		RowsAffected: 0, // 可以从结果中获取
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 添加记录
	m.queries = append(m.queries, record)

	// 如果超过最大记录数，删除最早的记录
	if len(m.queries) > m.maxRecords {
		m.queries = m.queries[1:]
	}

	// 记录慢查询日志
	config.Warnf("检测到慢查询: %s [%v]", sql, duration)
}

// GetSlowQueries 获取慢查询记录
func (m *SlowQueryMonitor) GetSlowQueries() []*SlowQueryRecord {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 创建副本
	result := make([]*SlowQueryRecord, len(m.queries))
	copy(result, m.queries)

	return result
}

// GetTopSlowQueries 获取最慢的N条查询
func (m *SlowQueryMonitor) GetTopSlowQueries(n int) []*SlowQueryRecord {
	queries := m.GetSlowQueries()

	// 按执行时间排序
	sort.Slice(queries, func(i, j int) bool {
		return queries[i].Duration > queries[j].Duration
	})

	// 返回前N条
	if n > 0 && n < len(queries) {
		return queries[:n]
	}
	return queries
}

// PrintSlowQueries 打印慢查询记录
func (m *SlowQueryMonitor) PrintSlowQueries(n int) {
	queries := m.GetTopSlowQueries(n)
	if len(queries) == 0 {
		config.Info("没有慢查询记录")
		return
	}

	config.Infof("=== 慢查询统计 (阈值: %v) ===", m.threshold)
	for i, q := range queries {
		config.Infof("%d. [%v] %s", i+1, q.Duration, q.SQL)
		if len(q.Params) > 0 {
			config.Infof("   参数: %v", q.Params)
		}
		config.Infof("   表: %s, 操作: %s, 时间: %s", q.Table, q.Operation, q.Timestamp.Format("2006-01-02 15:04:05"))
		if q.CallStack != "" {
			config.Infof("   调用栈: %s", q.CallStack)
		}
	}
	config.Info("=== 慢查询统计结束 ===")
}

// ClearQueries 清空查询记录
func (m *SlowQueryMonitor) ClearQueries() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.queries = make([]*SlowQueryRecord, 0, m.maxRecords)
	config.Info("慢查询记录已清空")
}

// GetStats 获取统计信息
func (m *SlowQueryMonitor) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{})
	stats["threshold"] = m.threshold.String()
	stats["total_records"] = len(m.queries)
	stats["enabled"] = m.enabled

	// 按表分组统计
	tableStats := make(map[string]int)
	// 按操作分组统计
	opStats := make(map[string]int)
	// 计算平均执行时间
	var totalDuration time.Duration

	for _, q := range m.queries {
		tableStats[q.Table]++
		opStats[q.Operation]++
		totalDuration += q.Duration
	}

	var avgDuration time.Duration
	if len(m.queries) > 0 {
		avgDuration = totalDuration / time.Duration(len(m.queries))
	}

	stats["average_duration"] = avgDuration.String()
	stats["table_stats"] = tableStats
	stats["operation_stats"] = opStats

	return stats
}

// SetThreshold 设置慢查询阈值
func (m *SlowQueryMonitor) SetThreshold(threshold time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.threshold = threshold
	config.Infof("慢查询阈值已更新为: %v", threshold)
}

// GetThreshold 获取慢查询阈值
func (m *SlowQueryMonitor) GetThreshold() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.threshold
}

// IsEnabled 检查是否启用
func (m *SlowQueryMonitor) IsEnabled() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.enabled
}

// GetRecordCount 获取记录数量
func (m *SlowQueryMonitor) GetRecordCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return len(m.queries)
}

// parseSQL 解析SQL获取表名和操作类型
func parseSQL(sql string) (table string, operation string) {
	// 转换为大写便于匹配
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))
	
	// 确定操作类型
	switch {
	case strings.HasPrefix(upperSQL, "SELECT"):
		operation = "SELECT"
		// 提取FROM后的表名
		if fromIndex := strings.Index(upperSQL, " FROM "); fromIndex != -1 {
			remaining := strings.TrimSpace(upperSQL[fromIndex+6:])
			parts := strings.Fields(remaining)
			if len(parts) > 0 {
				table = strings.ToLower(parts[0])
			}
		}
	case strings.HasPrefix(upperSQL, "INSERT"):
		operation = "INSERT"
		// 提取INTO后的表名
		if intoIndex := strings.Index(upperSQL, " INTO "); intoIndex != -1 {
			remaining := strings.TrimSpace(upperSQL[intoIndex+6:])
			parts := strings.Fields(remaining)
			if len(parts) > 0 {
				table = strings.ToLower(parts[0])
			}
		}
	case strings.HasPrefix(upperSQL, "UPDATE"):
		operation = "UPDATE"
		// 提取UPDATE后的表名
		parts := strings.Fields(upperSQL)
		if len(parts) > 1 {
			table = strings.ToLower(parts[1])
		}
	case strings.HasPrefix(upperSQL, "DELETE"):
		operation = "DELETE"
		// 提取FROM后的表名
		if fromIndex := strings.Index(upperSQL, " FROM "); fromIndex != -1 {
			remaining := strings.TrimSpace(upperSQL[fromIndex+6:])
			parts := strings.Fields(remaining)
			if len(parts) > 0 {
				table = strings.ToLower(parts[0])
			}
		}
	case strings.HasPrefix(upperSQL, "CREATE"):
		operation = "CREATE"
	case strings.HasPrefix(upperSQL, "DROP"):
		operation = "DROP"
	case strings.HasPrefix(upperSQL, "ALTER"):
		operation = "ALTER"
	default:
		operation = "OTHER"
	}

	// 如果没有提取到表名，设置为unknown
	if table == "" {
		table = "unknown"
	}

	return table, operation
}

// getCallStack 获取调用栈信息
func getCallStack() string {
	var stack []string
	
	// 跳过当前函数和RecordQuery函数
	for i := 3; i < 8; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		
		funcName := fn.Name()
		// 过滤掉runtime相关的函数
		if strings.Contains(funcName, "runtime.") {
			continue
		}
		
		// 简化文件路径
		if lastSlash := strings.LastIndex(file, "/"); lastSlash != -1 {
			file = file[lastSlash+1:]
		}
		
		stack = append(stack, fmt.Sprintf("%s:%d", file, line))
	}
	
	if len(stack) == 0 {
		return "无调用栈信息"
	}
	
	return strings.Join(stack, " -> ")
}

// ============= 慢查询分析器 =============

// SlowQueryAnalyzer 慢查询分析器
type SlowQueryAnalyzer struct {
	monitor *SlowQueryMonitor
}

// NewSlowQueryAnalyzer 创建慢查询分析器
func NewSlowQueryAnalyzer(monitor *SlowQueryMonitor) *SlowQueryAnalyzer {
	return &SlowQueryAnalyzer{
		monitor: monitor,
	}
}

// AnalyzeByTable 按表分析慢查询
func (a *SlowQueryAnalyzer) AnalyzeByTable() map[string]*TableAnalysis {
	queries := a.monitor.GetSlowQueries()
	result := make(map[string]*TableAnalysis)
	
	for _, q := range queries {
		if analysis, exists := result[q.Table]; exists {
			analysis.Count++
			analysis.TotalDuration += q.Duration
			if q.Duration > analysis.MaxDuration {
				analysis.MaxDuration = q.Duration
				analysis.SlowestQuery = q.SQL
			}
			if q.Duration < analysis.MinDuration {
				analysis.MinDuration = q.Duration
			}
		} else {
			result[q.Table] = &TableAnalysis{
				Table:         q.Table,
				Count:         1,
				TotalDuration: q.Duration,
				MaxDuration:   q.Duration,
				MinDuration:   q.Duration,
				SlowestQuery:  q.SQL,
			}
		}
	}
	
	// 计算平均时间
	for _, analysis := range result {
		analysis.AvgDuration = analysis.TotalDuration / time.Duration(analysis.Count)
	}
	
	return result
}

// AnalyzeByOperation 按操作类型分析慢查询
func (a *SlowQueryAnalyzer) AnalyzeByOperation() map[string]*OperationAnalysis {
	queries := a.monitor.GetSlowQueries()
	result := make(map[string]*OperationAnalysis)
	
	for _, q := range queries {
		if analysis, exists := result[q.Operation]; exists {
			analysis.Count++
			analysis.TotalDuration += q.Duration
			if q.Duration > analysis.MaxDuration {
				analysis.MaxDuration = q.Duration
				analysis.SlowestQuery = q.SQL
			}
			if q.Duration < analysis.MinDuration {
				analysis.MinDuration = q.Duration
			}
		} else {
			result[q.Operation] = &OperationAnalysis{
				Operation:     q.Operation,
				Count:         1,
				TotalDuration: q.Duration,
				MaxDuration:   q.Duration,
				MinDuration:   q.Duration,
				SlowestQuery:  q.SQL,
			}
		}
	}
	
	// 计算平均时间
	for _, analysis := range result {
		analysis.AvgDuration = analysis.TotalDuration / time.Duration(analysis.Count)
	}
	
	return result
}

// TableAnalysis 表分析结果
type TableAnalysis struct {
	Table         string        `json:"table"`
	Count         int           `json:"count"`
	TotalDuration time.Duration `json:"total_duration"`
	AvgDuration   time.Duration `json:"avg_duration"`
	MaxDuration   time.Duration `json:"max_duration"`
	MinDuration   time.Duration `json:"min_duration"`
	SlowestQuery  string        `json:"slowest_query"`
}

// OperationAnalysis 操作分析结果
type OperationAnalysis struct {
	Operation     string        `json:"operation"`
	Count         int           `json:"count"`
	TotalDuration time.Duration `json:"total_duration"`
	AvgDuration   time.Duration `json:"avg_duration"`
	MaxDuration   time.Duration `json:"max_duration"`
	MinDuration   time.Duration `json:"min_duration"`
	SlowestQuery  string        `json:"slowest_query"`
}

// PrintTableAnalysis 打印表分析结果
func (a *SlowQueryAnalyzer) PrintTableAnalysis() {
	analysis := a.AnalyzeByTable()
	if len(analysis) == 0 {
		config.Info("没有慢查询数据可分析")
		return
	}
	
	config.Info("=== 按表分析慢查询 ===")
	for table, data := range analysis {
		config.Infof("表: %s", table)
		config.Infof("  查询次数: %d", data.Count)
		config.Infof("  平均耗时: %v", data.AvgDuration)
		config.Infof("  最大耗时: %v", data.MaxDuration)
		config.Infof("  最小耗时: %v", data.MinDuration)
		config.Infof("  最慢查询: %s", data.SlowestQuery)
		config.Info("---")
	}
	config.Info("=== 表分析结束 ===")
}

// PrintOperationAnalysis 打印操作分析结果
func (a *SlowQueryAnalyzer) PrintOperationAnalysis() {
	analysis := a.AnalyzeByOperation()
	if len(analysis) == 0 {
		config.Info("没有慢查询数据可分析")
		return
	}
	
	config.Info("=== 按操作类型分析慢查询 ===")
	for operation, data := range analysis {
		config.Infof("操作: %s", operation)
		config.Infof("  查询次数: %d", data.Count)
		config.Infof("  平均耗时: %v", data.AvgDuration)
		config.Infof("  最大耗时: %v", data.MaxDuration)
		config.Infof("  最小耗时: %v", data.MinDuration)
		config.Infof("  最慢查询: %s", data.SlowestQuery)
		config.Info("---")
	}
	config.Info("=== 操作分析结束 ===")
}

// ============= 全局实例 =============

// 全局慢查询监控器
var (
	globalSlowQueryMonitor *SlowQueryMonitor
	slowQueryMonitorOnce   sync.Once
)

// GetGlobalSlowQueryMonitor 获取全局慢查询监控器
func GetGlobalSlowQueryMonitor() *SlowQueryMonitor {
	slowQueryMonitorOnce.Do(func() {
		globalSlowQueryMonitor = NewSlowQueryMonitor(time.Millisecond*500, 100)
		globalSlowQueryMonitor.Start()
	})
	return globalSlowQueryMonitor
}

// SetGlobalSlowQueryMonitor 设置全局慢查询监控器
func SetGlobalSlowQueryMonitor(monitor *SlowQueryMonitor) {
	if globalSlowQueryMonitor != nil {
		globalSlowQueryMonitor.Stop()
	}
	globalSlowQueryMonitor = monitor
}

// ============= 便捷函数 =============

// RecordSlowQuery 记录慢查询
func RecordSlowQuery(sql string, duration time.Duration, params ...interface{}) {
	GetGlobalSlowQueryMonitor().RecordQuery(sql, duration, params...)
}

// PrintSlowQueryStats 打印慢查询统计
func PrintSlowQueryStats(n int) {
	GetGlobalSlowQueryMonitor().PrintSlowQueries(n)
}

// GetSlowQueryStats 获取慢查询统计
func GetSlowQueryStats() map[string]interface{} {
	return GetGlobalSlowQueryMonitor().GetStats()
}

// ClearSlowQueryRecords 清空慢查询记录
func ClearSlowQueryRecords() {
	GetGlobalSlowQueryMonitor().ClearQueries()
}

// SetSlowQueryThreshold 设置慢查询阈值
func SetSlowQueryThreshold(threshold time.Duration) {
	GetGlobalSlowQueryMonitor().SetThreshold(threshold)
}

// EnableSlowQueryAutoPrint 启用慢查询自动打印
func EnableSlowQueryAutoPrint(interval time.Duration) {
	GetGlobalSlowQueryMonitor().EnableAutoPrint(interval)
}

// DisableSlowQueryAutoPrint 禁用慢查询自动打印
func DisableSlowQueryAutoPrint() {
	GetGlobalSlowQueryMonitor().DisableAutoPrint()
}

// AnalyzeSlowQueriesByTable 按表分析慢查询
func AnalyzeSlowQueriesByTable() {
	analyzer := NewSlowQueryAnalyzer(GetGlobalSlowQueryMonitor())
	analyzer.PrintTableAnalysis()
}

// AnalyzeSlowQueriesByOperation 按操作类型分析慢查询
func AnalyzeSlowQueriesByOperation() {
	analyzer := NewSlowQueryAnalyzer(GetGlobalSlowQueryMonitor())
	analyzer.PrintOperationAnalysis()
}