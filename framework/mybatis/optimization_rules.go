// Package mybatis SQL优化规则引擎
//
// 提供智能SQL优化建议：
// 1. 索引优化规则
// 2. 查询重写规则
// 3. 表结构优化规则
// 4. 配置优化规则
package mybatis

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// OptimizationRuleEngine 优化规则引擎
type OptimizationRuleEngine struct {
	rules           []OptimizationRule
	ruleRegistry    map[string]OptimizationRule
	enabledRules    map[string]bool
	ruleWeights     map[string]int
	mutex           sync.RWMutex
}

// IndexOptimizationRule 索引优化规则
type IndexOptimizationRule struct {
	name        string
	priority    int
	patterns    []*regexp.Regexp
	suggestions map[string][]string
}

// QueryRewriteRule 查询重写规则
type QueryRewriteRule struct {
	name        string
	priority    int
	pattern     *regexp.Regexp
	replacement string
	condition   func(*SlowQueryRecord) bool
}

// TableStructureRule 表结构优化规则
type TableStructureRule struct {
	name         string
	priority     int
	analyzer     TableAnalyzer
	suggestions  []StructureSuggestion
}

// ConfigOptimizationRule 配置优化规则
type ConfigOptimizationRule struct {
	name         string
	priority     int
	configType   string
	thresholds   map[string]float64
	suggestions  map[string]string
}

// PerformanceRule 性能规则
type PerformanceRule struct {
	name            string
	priority        int
	durationThreshold time.Duration
	memoryThreshold   int64
	analyzer          PerformanceAnalyzer
}

// StructureSuggestion 结构建议
type StructureSuggestion struct {
	Type        string `json:"type"`        // INDEX, PARTITION, NORMALIZATION
	Table       string `json:"table"`
	Column      string `json:"column"`
	Description string `json:"description"`
	Impact      string `json:"impact"`      // HIGH, MEDIUM, LOW
	Effort      string `json:"effort"`      // HIGH, MEDIUM, LOW
}

// TableAnalyzer 表分析器接口
type TableAnalyzer interface {
	AnalyzeTable(tableName string) (*TableAnalysis, error)
}

// PerformanceAnalyzer 性能分析器接口
type PerformanceAnalyzer interface {
	AnalyzePerformance(record *SlowQueryRecord) (*PerformanceAnalysis, error)
}

// TableAnalysis 表分析结果
type TableAnalysis struct {
	TableName     string              `json:"table_name"`
	RowCount      int64               `json:"row_count"`
	TableSize     int64               `json:"table_size_bytes"`
	IndexCount    int                 `json:"index_count"`
	Indexes       []IndexAnalysis     `json:"indexes"`
	Columns       []ColumnAnalysis    `json:"columns"`
	Partitions    []PartitionAnalysis `json:"partitions"`
	LastAnalyzed  time.Time           `json:"last_analyzed"`
}

// IndexAnalysis 索引分析
type IndexAnalysis struct {
	Name        string   `json:"name"`
	Columns     []string `json:"columns"`
	Type        string   `json:"type"`
	Size        int64    `json:"size_bytes"`
	Cardinality int64    `json:"cardinality"`
	Usage       int64    `json:"usage_count"`
	LastUsed    time.Time `json:"last_used"`
	IsRedundant bool     `json:"is_redundant"`
}

// ColumnAnalysis 列分析
type ColumnAnalysis struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	Nullable     bool    `json:"nullable"`
	Indexed      bool    `json:"indexed"`
	Cardinality  int64   `json:"cardinality"`
	Selectivity  float64 `json:"selectivity"`
	AvgLength    int     `json:"avg_length"`
	NullCount    int64   `json:"null_count"`
}

// PartitionAnalysis 分区分析
type PartitionAnalysis struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Column   string `json:"column"`
	Size     int64  `json:"size_bytes"`
	RowCount int64  `json:"row_count"`
}

// PerformanceAnalysis 性能分析结果
type PerformanceAnalysis struct {
	QueryComplexity   string                 `json:"query_complexity"`   // SIMPLE, MODERATE, COMPLEX
	ResourceUsage     string                 `json:"resource_usage"`     // LOW, MEDIUM, HIGH
	Bottlenecks       []string               `json:"bottlenecks"`
	Recommendations   []string               `json:"recommendations"`
	EstimatedImprovement float64             `json:"estimated_improvement"`
	Details           map[string]interface{} `json:"details"`
}

// NewOptimizationRuleEngine 创建优化规则引擎
func NewOptimizationRuleEngine() *OptimizationRuleEngine {
	engine := &OptimizationRuleEngine{
		rules:        make([]OptimizationRule, 0),
		ruleRegistry: make(map[string]OptimizationRule),
		enabledRules: make(map[string]bool),
		ruleWeights:  make(map[string]int),
	}
	
	// 注册默认规则
	engine.registerDefaultRules()
	
	return engine
}

// registerDefaultRules 注册默认规则
func (ore *OptimizationRuleEngine) registerDefaultRules() {
	// 索引优化规则
	ore.RegisterRule(&IndexOptimizationRule{
		name:     "missing_index_where_clause",
		priority: 10,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)WHERE\s+(\w+)\s*=`),
			regexp.MustCompile(`(?i)WHERE\s+(\w+)\s+IN\s*\(`),
		},
		suggestions: map[string][]string{
			"equality": {"Consider adding an index on the column used in WHERE clause for equality comparison"},
			"in_clause": {"Consider adding an index on the column used in IN clause"},
		},
	})
	
	ore.RegisterRule(&IndexOptimizationRule{
		name:     "missing_index_order_by",
		priority: 8,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)ORDER\s+BY\s+(\w+)`),
		},
		suggestions: map[string][]string{
			"order_by": {"Consider adding an index on the ORDER BY column to avoid sorting"},
		},
	})
	
	ore.RegisterRule(&IndexOptimizationRule{
		name:     "missing_index_join",
		priority: 9,
		patterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)JOIN\s+\w+\s+ON\s+\w+\.(\w+)\s*=\s*\w+\.(\w+)`),
		},
		suggestions: map[string][]string{
			"join": {"Consider adding indexes on both sides of the JOIN condition"},
		},
	})
	
	// 查询重写规则
	ore.RegisterRule(&QueryRewriteRule{
		name:     "select_star_optimization",
		priority: 7,
		pattern:  regexp.MustCompile(`(?i)SELECT\s+\*\s+FROM`),
		condition: func(record *SlowQueryRecord) bool {
			return record.Duration > 100*time.Millisecond
		},
	})
	
	ore.RegisterRule(&QueryRewriteRule{
		name:     "limit_without_order_optimization",
		priority: 6,
		pattern:  regexp.MustCompile(`(?i)LIMIT\s+\d+`),
		condition: func(record *SlowQueryRecord) bool {
			return !strings.Contains(strings.ToUpper(record.SQL), "ORDER BY")
		},
	})
	
	// 性能规则
	ore.RegisterRule(&PerformanceRule{
		name:              "high_memory_usage",
		priority:          8,
		durationThreshold: 500 * time.Millisecond,
		memoryThreshold:   50 * 1024 * 1024, // 50MB
		analyzer:          &DefaultPerformanceAnalyzer{},
	})
	
	// 配置优化规则
	ore.RegisterRule(&ConfigOptimizationRule{
		name:       "connection_pool_optimization",
		priority:   5,
		configType: "connection_pool",
		thresholds: map[string]float64{
			"max_connections": 100,
			"idle_timeout":    300, // seconds
		},
		suggestions: map[string]string{
			"high_connection_usage": "Consider increasing connection pool size or optimizing query performance",
			"high_idle_time":        "Consider reducing connection idle timeout to free up resources",
		},
	})
}

// RegisterRule 注册规则
func (ore *OptimizationRuleEngine) RegisterRule(rule OptimizationRule) {
	ore.mutex.Lock()
	defer ore.mutex.Unlock()
	
	ruleType := rule.GetRuleType()
	ore.rules = append(ore.rules, rule)
	ore.ruleRegistry[ruleType] = rule
	ore.enabledRules[ruleType] = true
	ore.ruleWeights[ruleType] = rule.Priority()
	
	// 按优先级排序
	sort.Slice(ore.rules, func(i, j int) bool {
		return ore.rules[i].Priority() > ore.rules[j].Priority()
	})
}

// EnableRule 启用规则
func (ore *OptimizationRuleEngine) EnableRule(ruleType string) {
	ore.mutex.Lock()
	defer ore.mutex.Unlock()
	ore.enabledRules[ruleType] = true
}

// DisableRule 禁用规则
func (ore *OptimizationRuleEngine) DisableRule(ruleType string) {
	ore.mutex.Lock()
	defer ore.mutex.Unlock()
	ore.enabledRules[ruleType] = false
}

// AnalyzeQuery 分析查询并生成优化建议
func (ore *OptimizationRuleEngine) AnalyzeQuery(record *SlowQueryRecord) []string {
	ore.mutex.RLock()
	defer ore.mutex.RUnlock()
	
	allSuggestions := make([]string, 0)
	
	// 应用所有启用的规则
	for _, rule := range ore.rules {
		if !ore.enabledRules[rule.GetRuleType()] {
			continue
		}
		
		suggestions := rule.Analyze(record)
		allSuggestions = append(allSuggestions, suggestions...)
	}
	
	// 去重和排序
	uniqueSuggestions := ore.deduplicate(allSuggestions)
	return ore.prioritizeSuggestions(uniqueSuggestions)
}

// deduplicate 去重建议
func (ore *OptimizationRuleEngine) deduplicate(suggestions []string) []string {
	seen := make(map[string]bool)
	unique := make([]string, 0)
	
	for _, suggestion := range suggestions {
		if !seen[suggestion] {
			seen[suggestion] = true
			unique = append(unique, suggestion)
		}
	}
	
	return unique
}

// prioritizeSuggestions 按优先级排序建议
func (ore *OptimizationRuleEngine) prioritizeSuggestions(suggestions []string) []string {
	// 根据建议类型的优先级排序
	sort.Slice(suggestions, func(i, j int) bool {
		// 按建议的重要性排序（简化实现）
		priorityMap := map[string]int{
			"index":     10,
			"query":     8,
			"structure": 6,
			"config":    4,
		}
		
		pi := ore.getSuggestionPriority(suggestions[i], priorityMap)
		pj := ore.getSuggestionPriority(suggestions[j], priorityMap)
		
		return pi > pj
	})
	
	return suggestions
}

// getSuggestionPriority 获取建议优先级
func (ore *OptimizationRuleEngine) getSuggestionPriority(suggestion string, priorityMap map[string]int) int {
	suggestion = strings.ToLower(suggestion)
	
	for keyword, priority := range priorityMap {
		if strings.Contains(suggestion, keyword) {
			return priority
		}
	}
	
	return 1 // 默认优先级
}

// IndexOptimizationRule 方法实现

// Analyze 分析索引优化
func (ior *IndexOptimizationRule) Analyze(record *SlowQueryRecord) []string {
	suggestions := make([]string, 0)
	
	for _, pattern := range ior.patterns {
		if pattern.MatchString(record.SQL) {
			// 根据匹配的模式生成建议
			if strings.Contains(ior.name, "where_clause") {
				suggestions = append(suggestions, ior.suggestions["equality"]...)
			} else if strings.Contains(ior.name, "order_by") {
				suggestions = append(suggestions, ior.suggestions["order_by"]...)
			} else if strings.Contains(ior.name, "join") {
				suggestions = append(suggestions, ior.suggestions["join"]...)
			}
		}
	}
	
	// 添加特定表的索引建议
	for _, tableName := range record.TableNames {
		suggestions = append(suggestions, 
			fmt.Sprintf("Consider analyzing table '%s' for missing indexes", tableName))
	}
	
	return suggestions
}

// GetRuleType 获取规则类型
func (ior *IndexOptimizationRule) GetRuleType() string {
	return ior.name
}

// Priority 获取优先级
func (ior *IndexOptimizationRule) Priority() int {
	return ior.priority
}

// QueryRewriteRule 方法实现

// Analyze 分析查询重写
func (qrr *QueryRewriteRule) Analyze(record *SlowQueryRecord) []string {
	suggestions := make([]string, 0)
	
	// 检查条件
	if qrr.condition != nil && !qrr.condition(record) {
		return suggestions
	}
	
	// 检查模式匹配
	if qrr.pattern.MatchString(record.SQL) {
		switch qrr.name {
		case "select_star_optimization":
			suggestions = append(suggestions, 
				"Avoid SELECT * and specify only needed columns to reduce data transfer and memory usage")
		case "limit_without_order_optimization":
			suggestions = append(suggestions, 
				"Use ORDER BY with LIMIT to ensure consistent and predictable results")
		}
	}
	
	return suggestions
}

// GetRuleType 获取规则类型
func (qrr *QueryRewriteRule) GetRuleType() string {
	return qrr.name
}

// Priority 获取优先级
func (qrr *QueryRewriteRule) Priority() int {
	return qrr.priority
}

// PerformanceRule 方法实现

// Analyze 分析性能规则
func (pr *PerformanceRule) Analyze(record *SlowQueryRecord) []string {
	suggestions := make([]string, 0)
	
	// 检查时间阈值
	if record.Duration >= pr.durationThreshold {
		suggestions = append(suggestions, 
			fmt.Sprintf("Query execution time (%v) exceeds recommended threshold (%v)", 
				record.Duration, pr.durationThreshold))
	}
	
	// 检查内存阈值
	if record.MemoryUsed >= pr.memoryThreshold {
		suggestions = append(suggestions, 
			fmt.Sprintf("Query memory usage (%d bytes) exceeds recommended threshold (%d bytes)", 
				record.MemoryUsed, pr.memoryThreshold))
	}
	
	// 使用分析器进行深度分析
	if pr.analyzer != nil {
		analysis, err := pr.analyzer.AnalyzePerformance(record)
		if err == nil {
			suggestions = append(suggestions, analysis.Recommendations...)
		}
	}
	
	return suggestions
}

// GetRuleType 获取规则类型
func (pr *PerformanceRule) GetRuleType() string {
	return pr.name
}

// Priority 获取优先级
func (pr *PerformanceRule) Priority() int {
	return pr.priority
}

// ConfigOptimizationRule 方法实现

// Analyze 分析配置优化
func (cor *ConfigOptimizationRule) Analyze(record *SlowQueryRecord) []string {
	suggestions := make([]string, 0)
	
	// 基于查询模式分析配置
	switch cor.configType {
	case "connection_pool":
		if record.Duration > 1*time.Second {
			suggestions = append(suggestions, cor.suggestions["high_connection_usage"])
		}
	}
	
	return suggestions
}

// GetRuleType 获取规则类型
func (cor *ConfigOptimizationRule) GetRuleType() string {
	return cor.name
}

// Priority 获取优先级
func (cor *ConfigOptimizationRule) Priority() int {
	return cor.priority
}

// DefaultPerformanceAnalyzer 默认性能分析器实现
type DefaultPerformanceAnalyzer struct{}

// AnalyzePerformance 分析性能
func (dpa *DefaultPerformanceAnalyzer) AnalyzePerformance(record *SlowQueryRecord) (*PerformanceAnalysis, error) {
	analysis := &PerformanceAnalysis{
		Bottlenecks:     make([]string, 0),
		Recommendations: make([]string, 0),
		Details:         make(map[string]interface{}),
	}
	
	// 分析查询复杂度
	analysis.QueryComplexity = dpa.analyzeComplexity(record.SQL)
	
	// 分析资源使用
	analysis.ResourceUsage = dpa.analyzeResourceUsage(record)
	
	// 识别瓶颈
	bottlenecks := dpa.identifyBottlenecks(record)
	analysis.Bottlenecks = bottlenecks
	
	// 生成建议
	recommendations := dpa.generateRecommendations(record, bottlenecks)
	analysis.Recommendations = recommendations
	
	// 估算改进效果
	analysis.EstimatedImprovement = dpa.estimateImprovement(record, recommendations)
	
	// 添加详细信息
	analysis.Details = map[string]interface{}{
		"query_length":     len(record.SQL),
		"table_count":      len(record.TableNames),
		"duration_ms":      record.Duration.Milliseconds(),
		"memory_used_mb":   record.MemoryUsed / (1024 * 1024),
	}
	
	return analysis, nil
}

// analyzeComplexity 分析查询复杂度
func (dpa *DefaultPerformanceAnalyzer) analyzeComplexity(sql string) string {
	upperSQL := strings.ToUpper(sql)
	
	complexity := 0
	
	// 检查复杂操作
	complexPatterns := []string{
		"JOIN", "UNION", "SUBQUERY", "GROUP BY", "HAVING", 
		"ORDER BY", "DISTINCT", "CASE WHEN",
	}
	
	for _, pattern := range complexPatterns {
		if strings.Contains(upperSQL, pattern) {
			complexity++
		}
	}
	
	// 检查函数调用
	functionPatterns := []string{
		"COUNT(", "SUM(", "AVG(", "MAX(", "MIN(",
		"SUBSTRING(", "CONCAT(", "DATE(",
	}
	
	for _, pattern := range functionPatterns {
		if strings.Contains(upperSQL, pattern) {
			complexity++
		}
	}
	
	if complexity >= 5 {
		return "COMPLEX"
	} else if complexity >= 2 {
		return "MODERATE"
	}
	
	return "SIMPLE"
}

// analyzeResourceUsage 分析资源使用
func (dpa *DefaultPerformanceAnalyzer) analyzeResourceUsage(record *SlowQueryRecord) string {
	memoryMB := record.MemoryUsed / (1024 * 1024)
	durationMs := record.Duration.Milliseconds()
	
	if memoryMB > 100 || durationMs > 1000 {
		return "HIGH"
	} else if memoryMB > 20 || durationMs > 200 {
		return "MEDIUM"
	}
	
	return "LOW"
}

// identifyBottlenecks 识别瓶颈
func (dpa *DefaultPerformanceAnalyzer) identifyBottlenecks(record *SlowQueryRecord) []string {
	bottlenecks := make([]string, 0)
	
	// 时间瓶颈
	if record.Duration > 500*time.Millisecond {
		bottlenecks = append(bottlenecks, "Execution time too long")
	}
	
	// 内存瓶颈
	if record.MemoryUsed > 50*1024*1024 {
		bottlenecks = append(bottlenecks, "High memory consumption")
	}
	
	// SQL模式瓶颈
	upperSQL := strings.ToUpper(record.SQL)
	if strings.Contains(upperSQL, "SELECT *") {
		bottlenecks = append(bottlenecks, "SELECT * may fetch unnecessary columns")
	}
	
	if strings.Contains(upperSQL, "LIMIT") && !strings.Contains(upperSQL, "ORDER BY") {
		bottlenecks = append(bottlenecks, "LIMIT without ORDER BY may produce inconsistent results")
	}
	
	if strings.Contains(upperSQL, "LIKE '%") {
		bottlenecks = append(bottlenecks, "Leading wildcard in LIKE prevents index usage")
	}
	
	return bottlenecks
}

// generateRecommendations 生成建议
func (dpa *DefaultPerformanceAnalyzer) generateRecommendations(record *SlowQueryRecord, bottlenecks []string) []string {
	recommendations := make([]string, 0)
	
	for _, bottleneck := range bottlenecks {
		switch {
		case strings.Contains(bottleneck, "Execution time"):
			recommendations = append(recommendations, "Consider adding appropriate indexes or optimizing query logic")
		case strings.Contains(bottleneck, "memory consumption"):
			recommendations = append(recommendations, "Optimize query to reduce memory usage, consider pagination")
		case strings.Contains(bottleneck, "SELECT *"):
			recommendations = append(recommendations, "Replace SELECT * with specific column names")
		case strings.Contains(bottleneck, "LIMIT without ORDER BY"):
			recommendations = append(recommendations, "Add ORDER BY clause when using LIMIT")
		case strings.Contains(bottleneck, "Leading wildcard"):
			recommendations = append(recommendations, "Avoid leading wildcards in LIKE patterns, consider full-text search")
		}
	}
	
	// 基于查询类型的通用建议
	switch record.QueryType {
	case "SELECT":
		if len(record.TableNames) > 1 {
			recommendations = append(recommendations, "Ensure join conditions use indexed columns")
		}
	case "INSERT":
		recommendations = append(recommendations, "Consider batch inserts for better performance")
	case "UPDATE", "DELETE":
		recommendations = append(recommendations, "Ensure WHERE clause uses indexed columns")
	}
	
	return recommendations
}

// estimateImprovement 估算改进效果
func (dpa *DefaultPerformanceAnalyzer) estimateImprovement(record *SlowQueryRecord, recommendations []string) float64 {
	improvement := 0.0
	
	// 基于建议类型估算改进百分比
	for _, recommendation := range recommendations {
		switch {
		case strings.Contains(recommendation, "index"):
			improvement += 0.4 // 40%改进
		case strings.Contains(recommendation, "SELECT *"):
			improvement += 0.2 // 20%改进
		case strings.Contains(recommendation, "batch"):
			improvement += 0.3 // 30%改进
		case strings.Contains(recommendation, "pagination"):
			improvement += 0.5 // 50%改进
		default:
			improvement += 0.1 // 10%改进
		}
	}
	
	// 限制最大改进为90%
	if improvement > 0.9 {
		improvement = 0.9
	}
	
	return improvement
}

// GetRuleStatistics 获取规则统计信息
func (ore *OptimizationRuleEngine) GetRuleStatistics() map[string]interface{} {
	ore.mutex.RLock()
	defer ore.mutex.RUnlock()
	
	stats := map[string]interface{}{
		"total_rules":   len(ore.rules),
		"enabled_rules": 0,
		"rule_details":  make([]map[string]interface{}, 0),
	}
	
	enabledCount := 0
	ruleDetails := make([]map[string]interface{}, 0)
	
	for _, rule := range ore.rules {
		ruleType := rule.GetRuleType()
		enabled := ore.enabledRules[ruleType]
		
		if enabled {
			enabledCount++
		}
		
		ruleDetails = append(ruleDetails, map[string]interface{}{
			"type":     ruleType,
			"priority": rule.Priority(),
			"enabled":  enabled,
		})
	}
	
	stats["enabled_rules"] = enabledCount
	stats["rule_details"] = ruleDetails
	
	return stats
}

// UpdateRulePriority 更新规则优先级
func (ore *OptimizationRuleEngine) UpdateRulePriority(ruleType string, priority int) error {
	ore.mutex.Lock()
	defer ore.mutex.Unlock()
	
	if _, exists := ore.ruleRegistry[ruleType]; !exists {
		return fmt.Errorf("rule type '%s' not found", ruleType)
	}
	
	ore.ruleWeights[ruleType] = priority
	
	// 重新排序规则
	sort.Slice(ore.rules, func(i, j int) bool {
		return ore.rules[i].Priority() > ore.rules[j].Priority()
	})
	
	return nil
}