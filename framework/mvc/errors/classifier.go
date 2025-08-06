package errors

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	mvccontext "github.com/zsy619/yyhertz/framework/mvc/context"
)

// ErrorCategory 错误分类
type ErrorCategory int

const (
	// 业务错误分类
	CategoryBusiness     ErrorCategory = iota // 业务逻辑错误
	CategoryValidation                        // 参数验证错误
	CategoryAuthentication                   // 认证错误
	CategoryAuthorization                    // 权限错误
	CategoryRateLimit                        // 限流错误
	
	// 系统错误分类
	CategorySystem       // 系统内部错误
	CategoryNetwork      // 网络错误
	CategoryTimeout      // 超时错误
	CategoryDatabase     // 数据库错误
	CategoryExternal     // 外部服务错误
	
	// 客户端错误分类
	CategoryClientError  // 客户端错误
	CategoryBadRequest   // 请求格式错误
	CategoryNotFound     // 资源不存在
	CategoryConflict     // 数据冲突
	
	// 未知错误分类
	CategoryUnknown      // 未知错误
)

// ErrorSeverity 错误严重等级
type ErrorSeverity int

const (
	SeverityLow      ErrorSeverity = iota // 低等级 - 可以忽略
	SeverityMedium                        // 中等级 - 需要关注
	SeverityHigh                          // 高等级 - 需要处理
	SeverityCritical                      // 严重等级 - 立即处理
)

// ErrorClassification 错误分类结果
type ErrorClassification struct {
	Original     error                  // 原始错误
	Category     ErrorCategory          // 错误分类
	Severity     ErrorSeverity          // 严重等级
	Retryable    bool                   // 是否可重试
	Timeout      *time.Duration         // 建议超时时间
	Context      map[string]interface{} // 分类上下文
	Score        float64                // 分类置信度
	Classifier   string                 // 分类器名称
	ClassifiedAt time.Time              // 分类时间
}

// ErrorClassifier 错误分类器接口
type ErrorClassifier interface {
	Classify(err error, ctx *mvccontext.EnhancedContext) *ErrorClassification
	CanClassify(err error) bool
	Priority() int
	Name() string
}

// ClassificationRule 分类规则
type ClassificationRule struct {
	Name       string                 // 规则名称
	Matcher    ErrorMatcher           // 匹配器
	Category   ErrorCategory          // 分类
	Severity   ErrorSeverity          // 严重等级
	Retryable  bool                   // 是否可重试
	Metadata   map[string]interface{} // 元数据
}

// ErrorMatcher 错误匹配器
type ErrorMatcher interface {
	Match(err error) bool
}

// IntelligentClassifier 智能错误分类器
type IntelligentClassifier struct {
	rules       []ClassificationRule // 分类规则
	patterns    []*PatternMatcher    // 模式匹配器
	statistics  ClassifierStats      // 统计信息
	config      ClassifierConfig     // 配置
	mu          sync.RWMutex         // 读写锁
	learningData map[string]*LearningEntry // 学习数据
}

// ClassifierStats 分类器统计
type ClassifierStats struct {
	TotalClassified   int64                         // 总分类次数
	CategoryCounts    map[ErrorCategory]int64       // 各分类数量
	SeverityCounts    map[ErrorSeverity]int64       // 各等级数量
	AccuracyRate      float64                       // 准确率
	AverageScore      float64                       // 平均置信度
	ClassificationTime time.Duration                // 平均分类时间
}

// ClassifierConfig 分类器配置
type ClassifierConfig struct {
	EnableLearning       bool          // 启用机器学习
	EnablePatternMatch   bool          // 启用模式匹配
	EnableStatistics     bool          // 启用统计
	LearningThreshold    float64       // 学习阈值
	MaxLearningEntries   int           // 最大学习条目
	PatternCacheSize     int           // 模式缓存大小
}

// LearningEntry 学习条目
type LearningEntry struct {
	Pattern      string            // 错误模式
	Category     ErrorCategory     // 正确分类
	Severity     ErrorSeverity     // 正确严重等级
	Confidence   float64           // 置信度
	UpdatedAt    time.Time         // 更新时间
	UsageCount   int64             // 使用次数
}

// PatternMatcher 模式匹配器
type PatternMatcher struct {
	Pattern      string           // 匹配模式
	Category     ErrorCategory    // 分类
	Severity     ErrorSeverity    // 严重等级
	Retryable    bool            // 是否可重试
	Confidence   float64         // 置信度
}

// NewIntelligentClassifier 创建智能分类器
func NewIntelligentClassifier() *IntelligentClassifier {
	classifier := &IntelligentClassifier{
		rules:        make([]ClassificationRule, 0),
		patterns:     make([]*PatternMatcher, 0),
		statistics:   ClassifierStats{
			CategoryCounts: make(map[ErrorCategory]int64),
			SeverityCounts: make(map[ErrorSeverity]int64),
		},
		config:       DefaultClassifierConfig(),
		learningData: make(map[string]*LearningEntry),
	}
	
	// 初始化默认规则
	classifier.initDefaultRules()
	
	// 初始化模式匹配器
	classifier.initPatternMatchers()
	
	return classifier
}

// DefaultClassifierConfig 默认分类器配置
func DefaultClassifierConfig() ClassifierConfig {
	return ClassifierConfig{
		EnableLearning:     true,
		EnablePatternMatch: true,
		EnableStatistics:   true,
		LearningThreshold:  0.8,
		MaxLearningEntries: 10000,
		PatternCacheSize:   1000,
	}
}

// Classify 分类错误
func (c *IntelligentClassifier) Classify(err error, ctx *mvccontext.EnhancedContext) *ErrorClassification {
	start := time.Now()
	defer func() {
		if c.config.EnableStatistics {
			c.updateClassificationStats(time.Since(start))
		}
	}()
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.statistics.TotalClassified++
	
	// 首先尝试精确规则匹配
	if classification := c.classifyByRules(err, ctx); classification != nil {
		classification.Classifier = "rule-based"
		c.updateStats(classification)
		return classification
	}
	
	// 然后尝试模式匹配
	if c.config.EnablePatternMatch {
		if classification := c.classifyByPatterns(err, ctx); classification != nil {
			classification.Classifier = "pattern-based"
			c.updateStats(classification)
			return classification
		}
	}
	
	// 最后尝试机器学习
	if c.config.EnableLearning {
		if classification := c.classifyByLearning(err, ctx); classification != nil {
			classification.Classifier = "ml-based"
			c.updateStats(classification)
			return classification
		}
	}
	
	// 默认分类
	classification := c.getDefaultClassification(err)
	classification.Classifier = "default"
	c.updateStats(classification)
	
	return classification
}

// classifyByRules 基于规则分类
func (c *IntelligentClassifier) classifyByRules(err error, ctx *mvccontext.EnhancedContext) *ErrorClassification {
	for _, rule := range c.rules {
		if rule.Matcher.Match(err) {
			return &ErrorClassification{
				Original:     err,
				Category:     rule.Category,
				Severity:     rule.Severity,
				Retryable:    rule.Retryable,
				Context:      rule.Metadata,
				Score:        1.0, // 规则匹配置信度最高
				ClassifiedAt: time.Now(),
			}
		}
	}
	
	return nil
}

// classifyByPatterns 基于模式分类
func (c *IntelligentClassifier) classifyByPatterns(err error, ctx *mvccontext.EnhancedContext) *ErrorClassification {
	errMsg := err.Error()
	
	var bestMatch *PatternMatcher
	var bestScore float64
	
	for _, pattern := range c.patterns {
		score := c.calculatePatternScore(errMsg, pattern.Pattern)
		if score > bestScore && score > 0.5 { // 最低置信度阈值
			bestScore = score
			bestMatch = pattern
		}
	}
	
	if bestMatch != nil {
		return &ErrorClassification{
			Original:     err,
			Category:     bestMatch.Category,
			Severity:     bestMatch.Severity,
			Retryable:    bestMatch.Retryable,
			Score:        bestScore,
			ClassifiedAt: time.Now(),
		}
	}
	
	return nil
}

// classifyByLearning 基于机器学习分类
func (c *IntelligentClassifier) classifyByLearning(err error, ctx *mvccontext.EnhancedContext) *ErrorClassification {
	errMsg := err.Error()
	pattern := c.extractPattern(errMsg)
	
	if entry, exists := c.learningData[pattern]; exists {
		entry.UsageCount++
		
		if entry.Confidence >= c.config.LearningThreshold {
			return &ErrorClassification{
				Original:     err,
				Category:     entry.Category,
				Severity:     entry.Severity,
				Retryable:    c.isRetryableByCategory(entry.Category),
				Score:        entry.Confidence,
				ClassifiedAt: time.Now(),
			}
		}
	}
	
	return nil
}

// getDefaultClassification 获取默认分类
func (c *IntelligentClassifier) getDefaultClassification(err error) *ErrorClassification {
	// 基于错误类型进行简单推断
	errMsg := strings.ToLower(err.Error())
	
	var category ErrorCategory = CategoryUnknown
	var severity ErrorSeverity = SeverityMedium
	var retryable bool = false
	
	switch {
	case strings.Contains(errMsg, "timeout"):
		category, severity, retryable = CategoryTimeout, SeverityHigh, true
	case strings.Contains(errMsg, "connection"):
		category, severity, retryable = CategoryNetwork, SeverityHigh, true
	case strings.Contains(errMsg, "database") || strings.Contains(errMsg, "sql"):
		category, severity, retryable = CategoryDatabase, SeverityHigh, true
	case strings.Contains(errMsg, "auth"):
		category, severity, retryable = CategoryAuthentication, SeverityMedium, false
	case strings.Contains(errMsg, "permission"):
		category, severity, retryable = CategoryAuthorization, SeverityMedium, false
	case strings.Contains(errMsg, "validation") || strings.Contains(errMsg, "invalid"):
		category, severity, retryable = CategoryValidation, SeverityLow, false
	case strings.Contains(errMsg, "not found"):
		category, severity, retryable = CategoryNotFound, SeverityLow, false
	default:
		category, severity, retryable = CategorySystem, SeverityMedium, false
	}
	
	return &ErrorClassification{
		Original:     err,
		Category:     category,
		Severity:     severity,
		Retryable:    retryable,
		Score:        0.6, // 默认分类置信度较低
		ClassifiedAt: time.Now(),
	}
}

// CanClassify 是否能够分类
func (c *IntelligentClassifier) CanClassify(err error) bool {
	return err != nil
}

// Priority 优先级
func (c *IntelligentClassifier) Priority() int {
	return 100 // 高优先级
}

// Name 分类器名称
func (c *IntelligentClassifier) Name() string {
	return "intelligent-classifier"
}

// initDefaultRules 初始化默认规则
func (c *IntelligentClassifier) initDefaultRules() {
	// ErrNo 业务错误
	c.AddRule(ClassificationRule{
		Name:     "errno-business",
		Matcher:  &TypeMatcher{TargetType: "*errors.ErrNo"},
		Category: CategoryBusiness,
		Severity: SeverityLow,
		Retryable: false,
	})
	
	// 超时错误
	c.AddRule(ClassificationRule{
		Name:     "context-timeout",
		Matcher:  &ContextMatcher{},
		Category: CategoryTimeout,
		Severity: SeverityHigh,
		Retryable: true,
	})
	
	// 网络错误
	c.AddRule(ClassificationRule{
		Name:     "network-error",
		Matcher:  &MessageMatcher{Patterns: []string{"connection", "network", "dial"}},
		Category: CategoryNetwork,
		Severity: SeverityHigh,
		Retryable: true,
	})
}

// initPatternMatchers 初始化模式匹配器
func (c *IntelligentClassifier) initPatternMatchers() {
	patterns := []*PatternMatcher{
		{Pattern: "timeout", Category: CategoryTimeout, Severity: SeverityHigh, Retryable: true, Confidence: 0.9},
		{Pattern: "connection refused", Category: CategoryNetwork, Severity: SeverityHigh, Retryable: true, Confidence: 0.95},
		{Pattern: "database", Category: CategoryDatabase, Severity: SeverityHigh, Retryable: true, Confidence: 0.8},
		{Pattern: "unauthorized", Category: CategoryAuthentication, Severity: SeverityMedium, Retryable: false, Confidence: 0.9},
		{Pattern: "forbidden", Category: CategoryAuthorization, Severity: SeverityMedium, Retryable: false, Confidence: 0.9},
		{Pattern: "validation failed", Category: CategoryValidation, Severity: SeverityLow, Retryable: false, Confidence: 0.85},
		{Pattern: "not found", Category: CategoryNotFound, Severity: SeverityLow, Retryable: false, Confidence: 0.8},
		{Pattern: "rate limit", Category: CategoryRateLimit, Severity: SeverityMedium, Retryable: true, Confidence: 0.9},
	}
	
	c.patterns = patterns
}

// AddRule 添加分类规则
func (c *IntelligentClassifier) AddRule(rule ClassificationRule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.rules = append(c.rules, rule)
}

// Learn 学习错误分类
func (c *IntelligentClassifier) Learn(err error, correctCategory ErrorCategory, correctSeverity ErrorSeverity) {
	if !c.config.EnableLearning {
		return
	}
	
	c.mu.Lock()
	defer c.mu.Unlock()
	
	pattern := c.extractPattern(err.Error())
	
	if entry, exists := c.learningData[pattern]; exists {
		// 更新已存在的条目
		entry.Category = correctCategory
		entry.Severity = correctSeverity
		entry.Confidence = c.calculateLearningConfidence(entry.UsageCount + 1)
		entry.UpdatedAt = time.Now()
		entry.UsageCount++
	} else {
		// 创建新的学习条目
		if len(c.learningData) >= c.config.MaxLearningEntries {
			c.evictLearningData()
		}
		
		c.learningData[pattern] = &LearningEntry{
			Pattern:    pattern,
			Category:   correctCategory,
			Severity:   correctSeverity,
			Confidence: 0.5, // 初始置信度
			UpdatedAt:  time.Now(),
			UsageCount: 1,
		}
	}
}

// extractPattern 提取错误模式
func (c *IntelligentClassifier) extractPattern(errMsg string) string {
	// 简化的模式提取：去除数字和时间戳，保留关键词
	pattern := strings.ToLower(errMsg)
	
	// 移除常见的变量部分
	replacements := map[string]string{
		`\d+`: "[NUM]",
		`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`: "[UUID]",
		`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`: "[TIME]",
		`"[^"]*"`: "[STR]",
	}
	
	for old, new := range replacements {
		pattern = strings.ReplaceAll(pattern, old, new)
	}
	
	return pattern
}

// calculatePatternScore 计算模式匹配分数
func (c *IntelligentClassifier) calculatePatternScore(errMsg, pattern string) float64 {
	errMsg = strings.ToLower(errMsg)
	pattern = strings.ToLower(pattern)
	
	if strings.Contains(errMsg, pattern) {
		// 基于匹配长度和位置计算分数
		score := float64(len(pattern)) / float64(len(errMsg))
		
		// 如果在开头匹配，增加分数
		if strings.HasPrefix(errMsg, pattern) {
			score += 0.2
		}
		
		// 限制最大分数
		if score > 1.0 {
			score = 1.0
		}
		
		return score
	}
	
	return 0.0
}

// calculateLearningConfidence 计算学习置信度
func (c *IntelligentClassifier) calculateLearningConfidence(usageCount int64) float64 {
	// 基于使用次数计算置信度，使用对数函数避免过快增长
	confidence := 0.5 + 0.4 * (1.0 - 1.0/float64(usageCount + 1))
	
	if confidence > 0.95 {
		confidence = 0.95 // 最大置信度限制
	}
	
	return confidence
}

// isRetryableByCategory 根据分类判断是否可重试
func (c *IntelligentClassifier) isRetryableByCategory(category ErrorCategory) bool {
	retryableCategories := map[ErrorCategory]bool{
		CategoryTimeout:    true,
		CategoryNetwork:    true,
		CategoryDatabase:   true,
		CategoryExternal:   true,
		CategoryRateLimit:  true,
		CategorySystem:     false,
	}
	
	return retryableCategories[category]
}

// updateStats 更新统计信息
func (c *IntelligentClassifier) updateStats(classification *ErrorClassification) {
	c.statistics.CategoryCounts[classification.Category]++
	c.statistics.SeverityCounts[classification.Severity]++
	c.statistics.AverageScore = (c.statistics.AverageScore * float64(c.statistics.TotalClassified - 1) + 
		classification.Score) / float64(c.statistics.TotalClassified)
}

// updateClassificationStats 更新分类时间统计
func (c *IntelligentClassifier) updateClassificationStats(duration time.Duration) {
	c.statistics.ClassificationTime = (c.statistics.ClassificationTime * 
		time.Duration(c.statistics.TotalClassified - 1) + duration) / 
		time.Duration(c.statistics.TotalClassified)
}

// evictLearningData 清理学习数据
func (c *IntelligentClassifier) evictLearningData() {
	// 简单的LRU清理：删除最少使用和最旧的条目
	var oldestPattern string
	var oldestTime time.Time = time.Now()
	var minUsage int64 = 9223372036854775807 // max int64
	
	for pattern, entry := range c.learningData {
		if entry.UsageCount < minUsage || 
		   (entry.UsageCount == minUsage && entry.UpdatedAt.Before(oldestTime)) {
			minUsage = entry.UsageCount
			oldestTime = entry.UpdatedAt
			oldestPattern = pattern
		}
	}
	
	if oldestPattern != "" {
		delete(c.learningData, oldestPattern)
	}
}

// GetStatistics 获取统计信息
func (c *IntelligentClassifier) GetStatistics() ClassifierStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// 深拷贝统计信息
	stats := ClassifierStats{
		TotalClassified:    c.statistics.TotalClassified,
		CategoryCounts:     make(map[ErrorCategory]int64),
		SeverityCounts:     make(map[ErrorSeverity]int64),
		AccuracyRate:       c.statistics.AccuracyRate,
		AverageScore:       c.statistics.AverageScore,
		ClassificationTime: c.statistics.ClassificationTime,
	}
	
	for k, v := range c.statistics.CategoryCounts {
		stats.CategoryCounts[k] = v
	}
	
	for k, v := range c.statistics.SeverityCounts {
		stats.SeverityCounts[k] = v
	}
	
	return stats
}

// 错误匹配器实现

// TypeMatcher 类型匹配器
type TypeMatcher struct {
	TargetType string
}

func (m *TypeMatcher) Match(err error) bool {
	return fmt.Sprintf("%T", err) == m.TargetType
}

// MessageMatcher 消息匹配器
type MessageMatcher struct {
	Patterns []string
}

func (m *MessageMatcher) Match(err error) bool {
	errMsg := strings.ToLower(err.Error())
	for _, pattern := range m.Patterns {
		if strings.Contains(errMsg, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// ContextMatcher 上下文匹配器
type ContextMatcher struct{}

func (m *ContextMatcher) Match(err error) bool {
	return err == context.DeadlineExceeded || err == context.Canceled
}

// 全局分类器实例
var globalClassifier = NewIntelligentClassifier()

// GetGlobalClassifier 获取全局分类器
func GetGlobalClassifier() *IntelligentClassifier {
	return globalClassifier
}

// ClassifyError 分类错误（全局方法）
func ClassifyError(err error, ctx *mvccontext.EnhancedContext) *ErrorClassification {
	return globalClassifier.Classify(err, ctx)
}

// LearnError 学习错误分类（全局方法）
func LearnError(err error, category ErrorCategory, severity ErrorSeverity) {
	globalClassifier.Learn(err, category, severity)
}

// GetCategoryName 获取分类名称
func GetCategoryName(category ErrorCategory) string {
	names := map[ErrorCategory]string{
		CategoryBusiness:       "Business",
		CategoryValidation:     "Validation",
		CategoryAuthentication: "Authentication",
		CategoryAuthorization:  "Authorization",
		CategoryRateLimit:      "RateLimit",
		CategorySystem:         "System",
		CategoryNetwork:        "Network",
		CategoryTimeout:        "Timeout",
		CategoryDatabase:       "Database",
		CategoryExternal:       "External",
		CategoryClientError:    "ClientError",
		CategoryBadRequest:     "BadRequest",
		CategoryNotFound:       "NotFound",
		CategoryConflict:       "Conflict",
		CategoryUnknown:        "Unknown",
	}
	
	if name, exists := names[category]; exists {
		return name
	}
	return "Unknown"
}

// GetSeverityName 获取严重等级名称
func GetSeverityName(severity ErrorSeverity) string {
	names := map[ErrorSeverity]string{
		SeverityLow:      "Low",
		SeverityMedium:   "Medium",
		SeverityHigh:     "High",
		SeverityCritical: "Critical",
	}
	
	if name, exists := names[severity]; exists {
		return name
	}
	return "Unknown"
}

// PrintClassifierInfo 打印分类器信息
func PrintClassifierInfo() {
	stats := globalClassifier.GetStatistics()
	
	fmt.Println("=== Error Classifier Statistics ===")
	fmt.Printf("Total Classifications: %d\n", stats.TotalClassified)
	fmt.Printf("Average Score: %.2f\n", stats.AverageScore)
	fmt.Printf("Average Classification Time: %v\n", stats.ClassificationTime)
	
	fmt.Println("\nCategory Distribution:")
	for category, count := range stats.CategoryCounts {
		percentage := float64(count) / float64(stats.TotalClassified) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", GetCategoryName(category), count, percentage)
	}
	
	fmt.Println("\nSeverity Distribution:")
	for severity, count := range stats.SeverityCounts {
		percentage := float64(count) / float64(stats.TotalClassified) * 100
		fmt.Printf("  %s: %d (%.1f%%)\n", GetSeverityName(severity), count, percentage)
	}
}