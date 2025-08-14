// Package devtools 提供安全监控指标功能
//
// 安全监控模块用于检测和监控应用的安全事件，提供：
// - 恶意请求检测和拦截
// - 认证失败统计分析
// - IP地址威胁情报分析
// - 异常访问模式识别
// - SQL注入攻击检测
// - XSS攻击检测
// - CSRF攻击防护
// - 暴力破解检测
//
// 功能特性：
// - 实时威胁检测
// - IP黑名单自动维护
// - 攻击模式机器学习
// - 安全事件告警
// - 地理位置分析
// - 用户行为异常检测
package devtools

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// ThreatLevel 威胁级别枚举
type ThreatLevel string

const (
	ThreatLevelLow      ThreatLevel = "low"      // 低风险
	ThreatLevelMedium   ThreatLevel = "medium"   // 中等风险
	ThreatLevelHigh     ThreatLevel = "high"     // 高风险
	ThreatLevelCritical ThreatLevel = "critical" // 严重风险
)

// AttackType 攻击类型枚举
type AttackType string

const (
	AttackTypeSQLInjection   AttackType = "sql_injection"   // SQL注入
	AttackTypeXSS            AttackType = "xss"             // 跨站脚本
	AttackTypeCSRF           AttackType = "csrf"            // 跨站请求伪造
	AttackTypeBruteForce     AttackType = "brute_force"     // 暴力破解
	AttackTypePathTraversal  AttackType = "path_traversal"  // 路径遍历
	AttackTypeCommandInjection AttackType = "command_injection" // 命令注入
	AttackTypeDDoS           AttackType = "ddos"            // DDoS攻击
	AttackTypeScanning       AttackType = "scanning"        // 扫描探测
	AttackTypeUnauthorized   AttackType = "unauthorized"    // 未授权访问
	AttackTypeSuspicious     AttackType = "suspicious"      // 可疑活动
)

// SecurityEvent 安全事件
type SecurityEvent struct {
	ID          string      `json:"id"`           // 事件ID
	Timestamp   time.Time   `json:"timestamp"`    // 发生时间
	IP          string      `json:"ip"`           // 来源IP
	UserAgent   string      `json:"user_agent"`   // 用户代理
	Method      string      `json:"method"`       // HTTP方法
	URL         string      `json:"url"`          // 请求URL
	AttackType  AttackType  `json:"attack_type"`  // 攻击类型
	ThreatLevel ThreatLevel `json:"threat_level"` // 威胁级别
	Description string      `json:"description"`  // 事件描述
	Payload     string      `json:"payload"`      // 攻击载荷
	Blocked     bool        `json:"blocked"`      // 是否被阻止
	Score       int         `json:"score"`        // 威胁评分
}

// IPThreatInfo IP威胁信息
type IPThreatInfo struct {
	IP             string      `json:"ip"`              // IP地址
	Country        string      `json:"country"`         // 国家
	Region         string      `json:"region"`          // 地区
	City           string      `json:"city"`            // 城市
	ISP            string      `json:"isp"`             // ISP
	ThreatScore    int         `json:"threat_score"`    // 威胁评分
	IsBlacklisted  bool        `json:"is_blacklisted"` // 是否在黑名单
	AttackCount    int64       `json:"attack_count"`   // 攻击次数
	LastAttack     time.Time   `json:"last_attack"`    // 最后攻击时间
	AttackTypes    []AttackType `json:"attack_types"`  // 攻击类型
	IsWhitelisted  bool        `json:"is_whitelisted"` // 是否在白名单
	FirstSeen      time.Time   `json:"first_seen"`     // 首次发现时间
}

// UserSecurityStats 用户安全统计
type UserSecurityStats struct {
	UserID         string    `json:"user_id"`         // 用户ID
	LoginAttempts  int64     `json:"login_attempts"`  // 登录尝试次数
	LoginFailures  int64     `json:"login_failures"`  // 登录失败次数
	LoginSuccesses int64     `json:"login_successes"` // 登录成功次数
	LastLogin      time.Time `json:"last_login"`      // 最后登录时间
	LastFailure    time.Time `json:"last_failure"`    // 最后失败时间
	SuspiciousIPs  []string  `json:"suspicious_ips"`  // 可疑IP列表
	IsLocked       bool      `json:"is_locked"`       // 是否被锁定
	LockTime       time.Time `json:"lock_time"`       // 锁定时间
}

// SecurityMetrics 安全指标统计
type SecurityMetrics struct {
	// 攻击统计
	TotalAttacks      int64                    `json:"total_attacks"`      // 总攻击次数
	BlockedAttacks    int64                    `json:"blocked_attacks"`    // 被阻止的攻击
	AttacksByType     map[AttackType]int64     `json:"attacks_by_type"`    // 按类型分类的攻击
	AttacksByLevel    map[ThreatLevel]int64    `json:"attacks_by_level"`   // 按级别分类的攻击
	
	// 认证统计
	AuthAttempts      int64                    `json:"auth_attempts"`      // 认证尝试次数
	AuthFailures      int64                    `json:"auth_failures"`      // 认证失败次数
	AuthSuccesses     int64                    `json:"auth_successes"`     // 认证成功次数
	
	// IP统计
	UniqueAttackerIPs int64                    `json:"unique_attacker_ips"` // 唯一攻击者IP数
	BlacklistedIPs    int64                    `json:"blacklisted_ips"`     // 黑名单IP数
	WhitelistedIPs    int64                    `json:"whitelisted_ips"`     // 白名单IP数
	
	// 地理统计
	AttacksByCountry  map[string]int64         `json:"attacks_by_country"`  // 按国家分类的攻击
	TopAttackerIPs    []string                 `json:"top_attacker_ips"`    // 顶级攻击者IP
	
	// 时间统计
	AttacksLastHour   int64                    `json:"attacks_last_hour"`   // 最近一小时攻击数
	AttacksLast24h    int64                    `json:"attacks_last_24h"`    // 最近24小时攻击数
	AttacksLastWeek   int64                    `json:"attacks_last_week"`   // 最近一周攻击数
	
	// 防护统计
	BlockRate         float64                  `json:"block_rate"`          // 阻断率
	FalsePositiveRate float64                  `json:"false_positive_rate"` // 误报率
}

// SecurityMetricsConfig 安全监控配置
type SecurityMetricsConfig struct {
	Enabled              bool          `json:"enabled"`               // 是否启用
	EnableSQLInjectionDetection bool   `json:"enable_sql_detection"`  // 启用SQL注入检测
	EnableXSSDetection   bool          `json:"enable_xss_detection"`   // 启用XSS检测
	EnableBruteForceDetection bool     `json:"enable_brute_force_detection"` // 启用暴力破解检测
	
	// 阈值配置
	MaxFailedLogins      int           `json:"max_failed_logins"`      // 最大登录失败次数
	BruteForceWindow     time.Duration `json:"brute_force_window"`     // 暴力破解检测窗口
	IPLockDuration       time.Duration `json:"ip_lock_duration"`       // IP锁定时长
	UserLockDuration     time.Duration `json:"user_lock_duration"`     // 用户锁定时长
	
	// 记录配置
	MaxSecurityEvents    int           `json:"max_security_events"`    // 最大安全事件记录数
	MaxIPThreatInfo      int           `json:"max_ip_threat_info"`     // 最大IP威胁信息数
	CleanupInterval      time.Duration `json:"cleanup_interval"`       // 清理间隔
	
	// 告警配置
	EnableAlerts         bool          `json:"enable_alerts"`          // 启用告警
	AlertThreshold       int           `json:"alert_threshold"`        // 告警阈值
	OnSecurityEvent      func(SecurityEvent) `json:"-"`          // 安全事件回调
}

// SecurityMetricsCollector 安全指标收集器
type SecurityMetricsCollector struct {
	mu                  sync.RWMutex
	config              *SecurityMetricsConfig
	enabled             bool
	startTime           time.Time
	
	// 指标数据
	metrics             *SecurityMetrics
	securityEvents      []SecurityEvent
	ipThreatInfo        map[string]*IPThreatInfo
	userSecurityStats   map[string]*UserSecurityStats
	
	// 检测规则
	sqlInjectionRegexes []*regexp.Regexp
	xssRegexes         []*regexp.Regexp
	pathTraversalRegexes []*regexp.Regexp
	
	// 黑白名单
	blacklistedIPs      map[string]time.Time
	whitelistedIPs      map[string]bool
	
	// 收集器控制
	cleanupTicker       *time.Ticker
	stopChan            chan struct{}
	
	// 事件队列
	eventQueue          chan SecurityEvent
	eventProcessor      *EventProcessor
}

// EventProcessor 事件处理器
type EventProcessor struct {
	mu           sync.RWMutex
	rules        []DetectionRule
	isProcessing bool
}

// DetectionRule 检测规则接口
type DetectionRule interface {
	Name() string
	Detect(req *http.Request, payload string) (bool, ThreatLevel, string)
	Priority() int
}

// NewSecurityMetricsCollector 创建安全指标收集器
func NewSecurityMetricsCollector(config *SecurityMetricsConfig) *SecurityMetricsCollector {
	if config == nil {
		config = &SecurityMetricsConfig{
			Enabled:                     true,
			EnableSQLInjectionDetection: true,
			EnableXSSDetection:          true,
			EnableBruteForceDetection:   true,
			MaxFailedLogins:             5,
			BruteForceWindow:            15 * time.Minute,
			IPLockDuration:              1 * time.Hour,
			UserLockDuration:            30 * time.Minute,
			MaxSecurityEvents:           10000,
			MaxIPThreatInfo:             5000,
			CleanupInterval:             1 * time.Hour,
			EnableAlerts:                true,
			AlertThreshold:              10,
		}
	}
	
	collector := &SecurityMetricsCollector{
		config:            config,
		enabled:           config.Enabled,
		startTime:         time.Now(),
		stopChan:          make(chan struct{}),
		eventQueue:        make(chan SecurityEvent, 1000),
		
		metrics: &SecurityMetrics{
			AttacksByType:    make(map[AttackType]int64),
			AttacksByLevel:   make(map[ThreatLevel]int64),
			AttacksByCountry: make(map[string]int64),
			TopAttackerIPs:   make([]string, 0),
		},
		
		securityEvents:    make([]SecurityEvent, 0, config.MaxSecurityEvents),
		ipThreatInfo:      make(map[string]*IPThreatInfo),
		userSecurityStats: make(map[string]*UserSecurityStats),
		blacklistedIPs:    make(map[string]time.Time),
		whitelistedIPs:    make(map[string]bool),
		
		eventProcessor: &EventProcessor{
			rules: make([]DetectionRule, 0),
		},
	}
	
	// 初始化检测规则
	collector.initializeDetectionRules()
	
	// 初始化事件处理器
	collector.initializeEventProcessor()
	
	return collector
}

// initializeDetectionRules 初始化检测规则
func (smc *SecurityMetricsCollector) initializeDetectionRules() {
	// SQL注入检测正则
	if smc.config.EnableSQLInjectionDetection {
		smc.sqlInjectionRegexes = []*regexp.Regexp{
			regexp.MustCompile(`(?i)(union.*select|select.*from|insert.*into|delete.*from|update.*set)`),
			regexp.MustCompile(`(?i)(\bor\b.*=.*\bor\b|\band\b.*=.*\band\b)`),
			regexp.MustCompile(`(?i)(drop\s+table|truncate\s+table|alter\s+table)`),
			regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\(|sp_executesql)`),
		}
	}
	
	// XSS检测正则
	if smc.config.EnableXSSDetection {
		smc.xssRegexes = []*regexp.Regexp{
			regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
			regexp.MustCompile(`(?i)javascript:`),
			regexp.MustCompile(`(?i)on\w+\s*=`),
			regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		}
	}
	
	// 路径遍历检测正则
	smc.pathTraversalRegexes = []*regexp.Regexp{
		regexp.MustCompile(`\.\.\/|\.\.\\`),
		regexp.MustCompile(`\/etc\/passwd|\/etc\/shadow`),
		regexp.MustCompile(`\\windows\\system32`),
	}
}

// initializeEventProcessor 初始化事件处理器
func (smc *SecurityMetricsCollector) initializeEventProcessor() {
	// 添加内置检测规则
	if smc.config.EnableSQLInjectionDetection {
		smc.eventProcessor.rules = append(smc.eventProcessor.rules, &SQLInjectionRule{
			regexes: smc.sqlInjectionRegexes,
		})
	}
	
	if smc.config.EnableXSSDetection {
		smc.eventProcessor.rules = append(smc.eventProcessor.rules, &XSSRule{
			regexes: smc.xssRegexes,
		})
	}
	
	smc.eventProcessor.rules = append(smc.eventProcessor.rules, &PathTraversalRule{
		regexes: smc.pathTraversalRegexes,
	})
	
	// 按优先级排序规则
	sort.Slice(smc.eventProcessor.rules, func(i, j int) bool {
		return smc.eventProcessor.rules[i].Priority() > smc.eventProcessor.rules[j].Priority()
	})
}

// Start 启动安全监控
func (smc *SecurityMetricsCollector) Start() {
	if !smc.enabled {
		return
	}
	
	smc.cleanupTicker = time.NewTicker(smc.config.CleanupInterval)
	
	// 启动事件处理器
	go smc.eventProcessorLoop()
	
	// 启动清理任务
	go smc.cleanupLoop()
}

// Stop 停止安全监控
func (smc *SecurityMetricsCollector) Stop() {
	if smc.cleanupTicker != nil {
		smc.cleanupTicker.Stop()
	}
	close(smc.stopChan)
}

// eventProcessorLoop 事件处理循环
func (smc *SecurityMetricsCollector) eventProcessorLoop() {
	for {
		select {
		case event := <-smc.eventQueue:
			smc.processSecurityEvent(event)
		case <-smc.stopChan:
			return
		}
	}
}

// cleanupLoop 清理循环
func (smc *SecurityMetricsCollector) cleanupLoop() {
	for {
		select {
		case <-smc.cleanupTicker.C:
			smc.cleanup()
		case <-smc.stopChan:
			return
		}
	}
}

// Handler 安全监控中间件
func (smc *SecurityMetricsCollector) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !smc.enabled {
			c.Next(ctx)
			return
		}
		
		start := time.Now()
		clientIP := c.ClientIP()
		method := string(c.Method())
		url := string(c.URI().FullURI())
		userAgent := string(c.UserAgent())
		
		// 检查白名单
		if smc.isWhitelisted(clientIP) {
			c.Next(ctx)
			return
		}
		
		// 检查黑名单
		if smc.isBlacklisted(clientIP) {
			smc.recordSecurityEvent(SecurityEvent{
				ID:          smc.generateEventID(),
				Timestamp:   start,
				IP:          clientIP,
				UserAgent:   userAgent,
				Method:      method,
				URL:         url,
				AttackType:  AttackTypeUnauthorized,
				ThreatLevel: ThreatLevelHigh,
				Description: "Request from blacklisted IP",
				Blocked:     true,
				Score:       80,
			})
			
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		
		// 获取请求体进行分析
		body := c.Request.Body()
		queryArgs := c.QueryArgs()
		payload := fmt.Sprintf("%s %s", string(body), queryArgs.String())
		
		// 威胁检测
		if detected, threatLevel, description := smc.detectThreats(method, url, payload, userAgent); detected {
			blocked := threatLevel == ThreatLevelHigh || threatLevel == ThreatLevelCritical
			
			event := SecurityEvent{
				ID:          smc.generateEventID(),
				Timestamp:   start,
				IP:          clientIP,
				UserAgent:   userAgent,
				Method:      method,
				URL:         url,
				AttackType:  smc.classifyAttackType(payload),
				ThreatLevel: threatLevel,
				Description: description,
				Payload:     payload,
				Blocked:     blocked,
				Score:       smc.calculateThreatScore(threatLevel),
			}
			
			// 记录安全事件
			select {
			case smc.eventQueue <- event:
			default:
				// 队列满时直接处理
				smc.processSecurityEvent(event)
			}
			
			if blocked {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		
		// 执行下一个中间件
		c.Next(ctx)
		
		// 检查响应状态码
		statusCode := c.Response.StatusCode()
		if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
			smc.recordAuthFailure(clientIP, "")
		}
	}
}

// detectThreats 威胁检测
func (smc *SecurityMetricsCollector) detectThreats(method, url, payload, userAgent string) (bool, ThreatLevel, string) {
	maxThreatLevel := ThreatLevelLow
	descriptions := make([]string, 0)
	detected := false
	
	// 使用事件处理器的规则进行检测
	for _, rule := range smc.eventProcessor.rules {
		if isDetected, level, desc := rule.Detect(&http.Request{
			Method: method,
			URL:    nil, // 简化处理
		}, payload); isDetected {
			detected = true
			descriptions = append(descriptions, fmt.Sprintf("%s: %s", rule.Name(), desc))
			
			if smc.compareThreatLevel(level, maxThreatLevel) {
				maxThreatLevel = level
			}
		}
	}
	
	// 用户代理异常检测
	if smc.detectSuspiciousUserAgent(userAgent) {
		detected = true
		descriptions = append(descriptions, "Suspicious user agent detected")
		if maxThreatLevel == ThreatLevelLow {
			maxThreatLevel = ThreatLevelMedium
		}
	}
	
	// URL异常检测
	if smc.detectSuspiciousURL(url) {
		detected = true
		descriptions = append(descriptions, "Suspicious URL pattern detected")
		if maxThreatLevel == ThreatLevelLow {
			maxThreatLevel = ThreatLevelMedium
		}
	}
	
	return detected, maxThreatLevel, strings.Join(descriptions, "; ")
}

// detectSuspiciousUserAgent 检测可疑用户代理
func (smc *SecurityMetricsCollector) detectSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"sqlmap", "nikto", "nmap", "masscan", "zap",
		"burp", "dirb", "gobuster", "wfuzz", "hydra",
	}
	
	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}
	
	return false
}

// detectSuspiciousURL 检测可疑URL
func (smc *SecurityMetricsCollector) detectSuspiciousURL(url string) bool {
	suspiciousPatterns := []string{
		"/admin", "/phpmyadmin", "/.env", "/config",
		"/backup", "/test", "/debug", "/.git",
	}
	
	urlLower := strings.ToLower(url)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}
	
	return false
}

// classifyAttackType 分类攻击类型
func (smc *SecurityMetricsCollector) classifyAttackType(payload string) AttackType {
	payloadLower := strings.ToLower(payload)
	
	// SQL注入检测
	for _, regex := range smc.sqlInjectionRegexes {
		if regex.MatchString(payloadLower) {
			return AttackTypeSQLInjection
		}
	}
	
	// XSS检测
	for _, regex := range smc.xssRegexes {
		if regex.MatchString(payloadLower) {
			return AttackTypeXSS
		}
	}
	
	// 路径遍历检测
	for _, regex := range smc.pathTraversalRegexes {
		if regex.MatchString(payloadLower) {
			return AttackTypePathTraversal
		}
	}
	
	return AttackTypeSuspicious
}

// calculateThreatScore 计算威胁评分
func (smc *SecurityMetricsCollector) calculateThreatScore(level ThreatLevel) int {
	switch level {
	case ThreatLevelLow:
		return 25
	case ThreatLevelMedium:
		return 50
	case ThreatLevelHigh:
		return 75
	case ThreatLevelCritical:
		return 100
	default:
		return 0
	}
}

// compareThreatLevel 比较威胁级别
func (smc *SecurityMetricsCollector) compareThreatLevel(level1, level2 ThreatLevel) bool {
	levelMap := map[ThreatLevel]int{
		ThreatLevelLow:      1,
		ThreatLevelMedium:   2,
		ThreatLevelHigh:     3,
		ThreatLevelCritical: 4,
	}
	
	return levelMap[level1] > levelMap[level2]
}

// processSecurityEvent 处理安全事件
func (smc *SecurityMetricsCollector) processSecurityEvent(event SecurityEvent) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	
	// 更新统计指标
	atomic.AddInt64(&smc.metrics.TotalAttacks, 1)
	if event.Blocked {
		atomic.AddInt64(&smc.metrics.BlockedAttacks, 1)
	}
	
	smc.metrics.AttacksByType[event.AttackType]++
	smc.metrics.AttacksByLevel[event.ThreatLevel]++
	
	// 记录安全事件
	smc.recordSecurityEvent(event)
	
	// 更新IP威胁信息
	smc.updateIPThreatInfo(event.IP, event.AttackType, event.ThreatLevel)
	
	// 检查是否需要拉黑IP
	if event.ThreatLevel == ThreatLevelHigh || event.ThreatLevel == ThreatLevelCritical {
		smc.addToBlacklist(event.IP)
	}
	
	// 触发告警
	if smc.config.EnableAlerts && smc.config.OnSecurityEvent != nil {
		go smc.config.OnSecurityEvent(event)
	}
}

// recordSecurityEvent 记录安全事件
func (smc *SecurityMetricsCollector) recordSecurityEvent(event SecurityEvent) {
	if len(smc.securityEvents) >= smc.config.MaxSecurityEvents {
		// 移除最旧的事件
		smc.securityEvents = smc.securityEvents[1:]
	}
	smc.securityEvents = append(smc.securityEvents, event)
}

// updateIPThreatInfo 更新IP威胁信息
func (smc *SecurityMetricsCollector) updateIPThreatInfo(ip string, attackType AttackType, threatLevel ThreatLevel) {
	info, exists := smc.ipThreatInfo[ip]
	if !exists {
		info = &IPThreatInfo{
			IP:          ip,
			FirstSeen:   time.Now(),
			AttackTypes: make([]AttackType, 0),
		}
		smc.ipThreatInfo[ip] = info
	}
	
	atomic.AddInt64(&info.AttackCount, 1)
	info.LastAttack = time.Now()
	info.ThreatScore += smc.calculateThreatScore(threatLevel)
	
	// 添加攻击类型（去重）
	found := false
	for _, t := range info.AttackTypes {
		if t == attackType {
			found = true
			break
		}
	}
	if !found {
		info.AttackTypes = append(info.AttackTypes, attackType)
	}
	
	// 检查是否需要拉黑
	if info.ThreatScore > 200 || info.AttackCount > 10 {
		info.IsBlacklisted = true
		smc.addToBlacklist(ip)
	}
}

// recordAuthFailure 记录认证失败
func (smc *SecurityMetricsCollector) recordAuthFailure(ip, userID string) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	
	atomic.AddInt64(&smc.metrics.AuthFailures, 1)
	
	// 更新用户统计
	if userID != "" {
		stats, exists := smc.userSecurityStats[userID]
		if !exists {
			stats = &UserSecurityStats{
				UserID:        userID,
				SuspiciousIPs: make([]string, 0),
			}
			smc.userSecurityStats[userID] = stats
		}
		
		atomic.AddInt64(&stats.LoginFailures, 1)
		stats.LastFailure = time.Now()
		
		// 检查是否需要锁定用户
		if stats.LoginFailures >= int64(smc.config.MaxFailedLogins) {
			stats.IsLocked = true
			stats.LockTime = time.Now()
		}
	}
	
	// 检查暴力破解
	if smc.config.EnableBruteForceDetection {
		smc.checkBruteForce(ip)
	}
}

// checkBruteForce 检查暴力破解
func (smc *SecurityMetricsCollector) checkBruteForce(ip string) {
	// 统计最近时间窗口内的失败次数
	now := time.Now()
	windowStart := now.Add(-smc.config.BruteForceWindow)
	
	failures := 0
	for _, event := range smc.securityEvents {
		if event.IP == ip && event.Timestamp.After(windowStart) && 
		   (event.AttackType == AttackTypeBruteForce || event.AttackType == AttackTypeUnauthorized) {
			failures++
		}
	}
	
	if failures >= smc.config.MaxFailedLogins {
		// 记录暴力破解事件
		event := SecurityEvent{
			ID:          smc.generateEventID(),
			Timestamp:   now,
			IP:          ip,
			AttackType:  AttackTypeBruteForce,
			ThreatLevel: ThreatLevelHigh,
			Description: fmt.Sprintf("Brute force attack detected: %d failures in %v", failures, smc.config.BruteForceWindow),
			Blocked:     true,
			Score:       80,
		}
		
		smc.recordSecurityEvent(event)
		smc.addToBlacklist(ip)
	}
}

// isWhitelisted 检查IP是否在白名单
func (smc *SecurityMetricsCollector) isWhitelisted(ip string) bool {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	return smc.whitelistedIPs[ip]
}

// isBlacklisted 检查IP是否在黑名单
func (smc *SecurityMetricsCollector) isBlacklisted(ip string) bool {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	
	expireTime, exists := smc.blacklistedIPs[ip]
	if !exists {
		return false
	}
	
	// 检查是否过期
	if time.Now().After(expireTime) {
		delete(smc.blacklistedIPs, ip)
		return false
	}
	
	return true
}

// addToBlacklist 添加IP到黑名单
func (smc *SecurityMetricsCollector) addToBlacklist(ip string) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	
	expireTime := time.Now().Add(smc.config.IPLockDuration)
	smc.blacklistedIPs[ip] = expireTime
	atomic.AddInt64(&smc.metrics.BlacklistedIPs, 1)
}

// AddToWhitelist 添加IP到白名单
func (smc *SecurityMetricsCollector) AddToWhitelist(ip string) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	
	smc.whitelistedIPs[ip] = true
	atomic.AddInt64(&smc.metrics.WhitelistedIPs, 1)
}

// RemoveFromWhitelist 从白名单移除IP
func (smc *SecurityMetricsCollector) RemoveFromWhitelist(ip string) {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	
	if smc.whitelistedIPs[ip] {
		delete(smc.whitelistedIPs, ip)
		atomic.AddInt64(&smc.metrics.WhitelistedIPs, -1)
	}
}

// cleanup 清理过期数据
func (smc *SecurityMetricsCollector) cleanup() {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	
	now := time.Now()
	
	// 清理过期的黑名单IP
	for ip, expireTime := range smc.blacklistedIPs {
		if now.After(expireTime) {
			delete(smc.blacklistedIPs, ip)
		}
	}
	
	// 清理过期的安全事件（保留最近24小时）
	cutoff := now.Add(-24 * time.Hour)
	validEvents := make([]SecurityEvent, 0)
	for _, event := range smc.securityEvents {
		if event.Timestamp.After(cutoff) {
			validEvents = append(validEvents, event)
		}
	}
	smc.securityEvents = validEvents
	
	// 清理过期的IP威胁信息
	if len(smc.ipThreatInfo) > smc.config.MaxIPThreatInfo {
		// 按最后攻击时间排序，保留最近的
		type ipInfo struct {
			ip   string
			info *IPThreatInfo
		}
		
		var infos []ipInfo
		for ip, info := range smc.ipThreatInfo {
			infos = append(infos, ipInfo{ip: ip, info: info})
		}
		
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].info.LastAttack.After(infos[j].info.LastAttack)
		})
		
		// 保留前MaxIPThreatInfo个
		smc.ipThreatInfo = make(map[string]*IPThreatInfo)
		for i := 0; i < smc.config.MaxIPThreatInfo && i < len(infos); i++ {
			smc.ipThreatInfo[infos[i].ip] = infos[i].info
		}
	}
}

// generateEventID 生成事件ID
func (smc *SecurityMetricsCollector) generateEventID() string {
	return fmt.Sprintf("sec_%d_%d", time.Now().UnixNano(), len(smc.securityEvents))
}

// GetMetrics 获取安全指标
func (smc *SecurityMetricsCollector) GetMetrics() *SecurityMetrics {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	
	// 更新时间相关的统计
	smc.updateTimeBasedStats()
	
	metrics := *smc.metrics
	
	// 复制map数据
	metrics.AttacksByType = make(map[AttackType]int64)
	for k, v := range smc.metrics.AttacksByType {
		metrics.AttacksByType[k] = v
	}
	
	metrics.AttacksByLevel = make(map[ThreatLevel]int64)
	for k, v := range smc.metrics.AttacksByLevel {
		metrics.AttacksByLevel[k] = v
	}
	
	metrics.AttacksByCountry = make(map[string]int64)
	for k, v := range smc.metrics.AttacksByCountry {
		metrics.AttacksByCountry[k] = v
	}
	
	return &metrics
}

// updateTimeBasedStats 更新基于时间的统计
func (smc *SecurityMetricsCollector) updateTimeBasedStats() {
	now := time.Now()
	
	// 统计最近不同时间段的攻击数
	var attacksLastHour, attacksLast24h, attacksLastWeek int64
	
	for _, event := range smc.securityEvents {
		if event.Timestamp.After(now.Add(-time.Hour)) {
			attacksLastHour++
		}
		if event.Timestamp.After(now.Add(-24 * time.Hour)) {
			attacksLast24h++
		}
		if event.Timestamp.After(now.Add(-7 * 24 * time.Hour)) {
			attacksLastWeek++
		}
	}
	
	smc.metrics.AttacksLastHour = attacksLastHour
	smc.metrics.AttacksLast24h = attacksLast24h
	smc.metrics.AttacksLastWeek = attacksLastWeek
	
	// 计算阻断率
	if smc.metrics.TotalAttacks > 0 {
		smc.metrics.BlockRate = float64(smc.metrics.BlockedAttacks) / float64(smc.metrics.TotalAttacks) * 100
	}
}

// GetSecurityEvents 获取安全事件
func (smc *SecurityMetricsCollector) GetSecurityEvents(limit int) []SecurityEvent {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	
	if limit <= 0 || limit > len(smc.securityEvents) {
		limit = len(smc.securityEvents)
	}
	
	// 返回最新的事件
	events := make([]SecurityEvent, limit)
	start := len(smc.securityEvents) - limit
	copy(events, smc.securityEvents[start:])
	
	return events
}

// GetIPThreatInfo 获取IP威胁信息
func (smc *SecurityMetricsCollector) GetIPThreatInfo() map[string]*IPThreatInfo {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	
	info := make(map[string]*IPThreatInfo)
	for k, v := range smc.ipThreatInfo {
		threatInfo := *v
		info[k] = &threatInfo
	}
	
	return info
}

// GetUserSecurityStats 获取用户安全统计
func (smc *SecurityMetricsCollector) GetUserSecurityStats() map[string]*UserSecurityStats {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	
	stats := make(map[string]*UserSecurityStats)
	for k, v := range smc.userSecurityStats {
		userStats := *v
		stats[k] = &userStats
	}
	
	return stats
}

// Reset 重置指标
func (smc *SecurityMetricsCollector) Reset() {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	
	// 重置统计指标
	smc.metrics.TotalAttacks = 0
	smc.metrics.BlockedAttacks = 0
	smc.metrics.AuthAttempts = 0
	smc.metrics.AuthFailures = 0
	smc.metrics.AuthSuccesses = 0
	smc.metrics.UniqueAttackerIPs = 0
	
	// 清空map
	for k := range smc.metrics.AttacksByType {
		smc.metrics.AttacksByType[k] = 0
	}
	for k := range smc.metrics.AttacksByLevel {
		smc.metrics.AttacksByLevel[k] = 0
	}
	for k := range smc.metrics.AttacksByCountry {
		smc.metrics.AttacksByCountry[k] = 0
	}
	
	// 清空事件和威胁信息
	smc.securityEvents = smc.securityEvents[:0]
	smc.ipThreatInfo = make(map[string]*IPThreatInfo)
	smc.userSecurityStats = make(map[string]*UserSecurityStats)
}

// IsEnabled 检查是否启用
func (smc *SecurityMetricsCollector) IsEnabled() bool {
	smc.mu.RLock()
	defer smc.mu.RUnlock()
	return smc.enabled
}

// Enable 启用收集器
func (smc *SecurityMetricsCollector) Enable() {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	smc.enabled = true
}

// Disable 禁用收集器
func (smc *SecurityMetricsCollector) Disable() {
	smc.mu.Lock()
	defer smc.mu.Unlock()
	smc.enabled = false
}

// 内置检测规则实现

// SQLInjectionRule SQL注入检测规则
type SQLInjectionRule struct {
	regexes []*regexp.Regexp
}

func (r *SQLInjectionRule) Name() string { return "SQL Injection Detection" }
func (r *SQLInjectionRule) Priority() int { return 90 }

func (r *SQLInjectionRule) Detect(req *http.Request, payload string) (bool, ThreatLevel, string) {
	for _, regex := range r.regexes {
		if regex.MatchString(strings.ToLower(payload)) {
			return true, ThreatLevelHigh, "Potential SQL injection detected"
		}
	}
	return false, ThreatLevelLow, ""
}

// XSSRule XSS检测规则
type XSSRule struct {
	regexes []*regexp.Regexp
}

func (r *XSSRule) Name() string { return "XSS Detection" }
func (r *XSSRule) Priority() int { return 85 }

func (r *XSSRule) Detect(req *http.Request, payload string) (bool, ThreatLevel, string) {
	for _, regex := range r.regexes {
		if regex.MatchString(strings.ToLower(payload)) {
			return true, ThreatLevelMedium, "Potential XSS attack detected"
		}
	}
	return false, ThreatLevelLow, ""
}

// PathTraversalRule 路径遍历检测规则
type PathTraversalRule struct {
	regexes []*regexp.Regexp
}

func (r *PathTraversalRule) Name() string { return "Path Traversal Detection" }
func (r *PathTraversalRule) Priority() int { return 80 }

func (r *PathTraversalRule) Detect(req *http.Request, payload string) (bool, ThreatLevel, string) {
	for _, regex := range r.regexes {
		if regex.MatchString(payload) {
			return true, ThreatLevelHigh, "Potential path traversal detected"
		}
	}
	return false, ThreatLevelLow, ""
}

// SecurityMetricsPanel 安全监控面板
type SecurityMetricsPanel struct {
	collector *SecurityMetricsCollector
}

// NewSecurityMetricsPanel 创建安全监控面板
func NewSecurityMetricsPanel(collector *SecurityMetricsCollector) *SecurityMetricsPanel {
	return &SecurityMetricsPanel{
		collector: collector,
	}
}

// RegisterRoutes 注册路由
func (smp *SecurityMetricsPanel) RegisterRoutes(engine any) {
	var secGroup *route.RouterGroup
	
	if h, ok := engine.(*route.Engine); ok {
		secGroup = h.Group("/yyhertz/security")
	} else {
		config.Error("无法注册安全监控路由，未知引擎类型")
		return
	}
	
	// 注册路由
	secGroup.GET("/", smp.getSecurityMetrics)
	secGroup.GET("/events", smp.getSecurityEvents)
	secGroup.GET("/threats", smp.getIPThreatInfo)
	secGroup.GET("/users", smp.getUserSecurityStats)
	secGroup.POST("/whitelist", smp.addToWhitelist)
	secGroup.DELETE("/whitelist", smp.removeFromWhitelist)
	secGroup.POST("/reset", smp.resetMetrics)
	secGroup.POST("/enable", smp.enableCollector)
	secGroup.POST("/disable", smp.disableCollector)
	secGroup.GET("/panel", smp.securityPanel)
}

// getSecurityMetrics 获取安全指标
func (smp *SecurityMetricsPanel) getSecurityMetrics(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"metrics":   smp.collector.GetMetrics(),
			"enabled":   smp.collector.IsEnabled(),
			"timestamp": time.Now(),
		},
	})
}

// getSecurityEvents 获取安全事件
func (smp *SecurityMetricsPanel) getSecurityEvents(ctx context.Context, c *app.RequestContext) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed := parseInt(string(l), 100); parsed > 0 {
			limit = parsed
		}
	}
	
	events := smp.collector.GetSecurityEvents(limit)
	
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"events":     events,
			"total":      len(events),
			"timestamp":  time.Now(),
		},
	})
}

// getIPThreatInfo 获取IP威胁信息
func (smp *SecurityMetricsPanel) getIPThreatInfo(ctx context.Context, c *app.RequestContext) {
	threatInfo := smp.collector.GetIPThreatInfo()
	
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"threat_info": threatInfo,
			"total":       len(threatInfo),
			"timestamp":   time.Now(),
		},
	})
}

// getUserSecurityStats 获取用户安全统计
func (smp *SecurityMetricsPanel) getUserSecurityStats(ctx context.Context, c *app.RequestContext) {
	userStats := smp.collector.GetUserSecurityStats()
	
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"user_stats": userStats,
			"total":      len(userStats),
			"timestamp":  time.Now(),
		},
	})
}

// addToWhitelist 添加到白名单
func (smp *SecurityMetricsPanel) addToWhitelist(ctx context.Context, c *app.RequestContext) {
	type request struct {
		IP string `json:"ip"`
	}
	
	var req request
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    1,
			"message": "Invalid request",
		})
		return
	}
	
	// 验证IP格式
	if net.ParseIP(req.IP) == nil {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    1,
			"message": "Invalid IP address",
		})
		return
	}
	
	smp.collector.AddToWhitelist(req.IP)
	
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": fmt.Sprintf("IP %s added to whitelist", req.IP),
	})
}

// removeFromWhitelist 从白名单移除
func (smp *SecurityMetricsPanel) removeFromWhitelist(ctx context.Context, c *app.RequestContext) {
	ip := c.Query("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    1,
			"message": "IP parameter required",
		})
		return
	}
	
	smp.collector.RemoveFromWhitelist(string(ip))
	
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": fmt.Sprintf("IP %s removed from whitelist", ip),
	})
}

// resetMetrics 重置指标
func (smp *SecurityMetricsPanel) resetMetrics(ctx context.Context, c *app.RequestContext) {
	smp.collector.Reset()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "安全指标已重置",
	})
}

// enableCollector 启用收集器
func (smp *SecurityMetricsPanel) enableCollector(ctx context.Context, c *app.RequestContext) {
	smp.collector.Enable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "安全监控已启用",
		"enabled": true,
	})
}

// disableCollector 禁用收集器
func (smp *SecurityMetricsPanel) disableCollector(ctx context.Context, c *app.RequestContext) {
	smp.collector.Disable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "安全监控已禁用",
		"enabled": false,
	})
}

// parseInt 解析整数
func parseInt(s string, defaultValue int) int {
	if v, err := fmt.Sscanf(s, "%d", &defaultValue); err != nil || v != 1 {
		return defaultValue
	}
	return defaultValue
}

// securityPanel 安全监控面板页面
func (smp *SecurityMetricsPanel) securityPanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 安全监控面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-card h3 { margin-top: 0; color: #333; border-bottom: 2px solid #dc3545; padding-bottom: 10px; }
        .metric-item { display: flex; justify-content: space-between; padding: 8px 0; border-bottom: 1px solid #eee; }
        .metric-item:last-child { border-bottom: none; }
        .metric-label { font-weight: bold; color: #555; }
        .metric-value { color: #dc3545; font-weight: bold; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-warning { background: #ffc107; color: black; }
        .events-table { background: white; margin-top: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .events-table table { width: 100%; border-collapse: collapse; }
        .events-table th, .events-table td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        .events-table th { background: #f8f9fa; font-weight: bold; }
        .threat-level { padding: 4px 8px; border-radius: 4px; color: white; font-size: 0.8em; }
        .threat-low { background: #28a745; }
        .threat-medium { background: #ffc107; color: black; }
        .threat-high { background: #fd7e14; }
        .threat-critical { background: #dc3545; }
        .status-indicator { width: 12px; height: 12px; border-radius: 50%; display: inline-block; margin-right: 5px; }
        .status-healthy { background: #28a745; }
        .status-warning { background: #ffc107; }
        .status-error { background: #dc3545; }
        .attack-payload { font-family: monospace; font-size: 0.8em; max-width: 200px; overflow: hidden; text-overflow: ellipsis; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 安全监控面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshMetrics()">刷新数据</button>
            <button class="btn btn-success" onclick="enableCollector()">启用监控</button>
            <button class="btn btn-danger" onclick="disableCollector()">禁用监控</button>
            <button class="btn btn-warning" onclick="resetMetrics()">重置指标</button>
        </div>
    </div>

    <div class="metrics-grid">
        <div class="metric-card">
            <h3><span class="status-indicator status-error"></span>攻击统计</h3>
            <div id="attackMetrics">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-warning"></span>认证统计</h3>
            <div id="authMetrics">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-error"></span>IP威胁统计</h3>
            <div id="ipMetrics">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>

        <div class="metric-card">
            <h3><span class="status-indicator status-warning"></span>时间段攻击</h3>
            <div id="timeMetrics">
                <div style="text-align: center; padding: 20px; color: #666;">加载中...</div>
            </div>
        </div>
    </div>

    <div class="events-table">
        <h3 style="padding: 20px; margin: 0; border-bottom: 1px solid #eee;">最近安全事件</h3>
        <div style="overflow-x: auto;">
            <table id="eventsTable">
                <thead>
                    <tr>
                        <th>时间</th>
                        <th>来源IP</th>
                        <th>攻击类型</th>
                        <th>威胁级别</th>
                        <th>描述</th>
                        <th>载荷</th>
                        <th>状态</th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td colspan="7" style="text-align: center; padding: 20px; color: #666;">加载中...</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>

    <script>
        function refreshMetrics() {
            loadSecurityData();
        }

        function loadSecurityData() {
            Promise.all([
                fetch('/yyhertz/security/'),
                fetch('/yyhertz/security/events?limit=50')
            ])
            .then(responses => Promise.all(responses.map(r => r.json())))
            .then(([metricsResponse, eventsResponse]) => {
                showAttackMetrics(metricsResponse.data.metrics);
                showSecurityEvents(eventsResponse.data.events || []);
            })
            .catch(error => {
                console.error('加载安全数据失败:', error);
            });
        }

        function showAttackMetrics(metrics) {
            // 攻击统计
            const attackContainer = document.getElementById('attackMetrics');
            const blockRate = metrics.block_rate || 0;
            
            attackContainer.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">总攻击数</span><span class="metric-value">' + (metrics.total_attacks || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">被阻止攻击</span><span class="metric-value">' + (metrics.blocked_attacks || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">阻断率</span><span class="metric-value">' + blockRate.toFixed(1) + '%</span></div>' +
                '<div class="metric-item"><span class="metric-label">SQL注入</span><span class="metric-value">' + (metrics.attacks_by_type?.sql_injection || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">XSS攻击</span><span class="metric-value">' + (metrics.attacks_by_type?.xss || 0) + '</span></div>';

            // 认证统计
            const authContainer = document.getElementById('authMetrics');
            const authSuccessRate = metrics.auth_attempts > 0 ? 
                (metrics.auth_successes / metrics.auth_attempts * 100) : 0;
            
            authContainer.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">认证尝试</span><span class="metric-value">' + (metrics.auth_attempts || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">认证成功</span><span class="metric-value">' + (metrics.auth_successes || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">认证失败</span><span class="metric-value">' + (metrics.auth_failures || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">成功率</span><span class="metric-value">' + authSuccessRate.toFixed(1) + '%</span></div>';

            // IP威胁统计
            const ipContainer = document.getElementById('ipMetrics');
            ipContainer.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">攻击者IP数</span><span class="metric-value">' + (metrics.unique_attacker_ips || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">黑名单IP</span><span class="metric-value">' + (metrics.blacklisted_ips || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">白名单IP</span><span class="metric-value">' + (metrics.whitelisted_ips || 0) + '</span></div>';

            // 时间段攻击统计
            const timeContainer = document.getElementById('timeMetrics');
            timeContainer.innerHTML = '' +
                '<div class="metric-item"><span class="metric-label">最近1小时</span><span class="metric-value">' + (metrics.attacks_last_hour || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最近24小时</span><span class="metric-value">' + (metrics.attacks_last_24h || 0) + '</span></div>' +
                '<div class="metric-item"><span class="metric-label">最近一周</span><span class="metric-value">' + (metrics.attacks_last_week || 0) + '</span></div>';
        }

        function showSecurityEvents(events) {
            const tbody = document.querySelector('#eventsTable tbody');
            
            if (events.length === 0) {
                tbody.innerHTML = '<tr><td colspan="7" style="text-align: center; padding: 20px; color: #666;">暂无安全事件</td></tr>';
                return;
            }

            let html = '';
            events.slice(0, 50).forEach(event => {
                const threatClass = 'threat-' + event.threat_level;
                const statusText = event.blocked ? '已阻止' : '已通过';
                const statusClass = event.blocked ? 'status-error' : 'status-warning';
                
                html += '<tr>' +
                    '<td>' + new Date(event.timestamp).toLocaleString() + '</td>' +
                    '<td>' + event.ip + '</td>' +
                    '<td>' + (event.attack_type || 'unknown') + '</td>' +
                    '<td><span class="threat-level ' + threatClass + '">' + event.threat_level + '</span></td>' +
                    '<td>' + (event.description || '') + '</td>' +
                    '<td><div class="attack-payload" title="' + (event.payload || '') + '">' + (event.payload || '').substring(0, 50) + '</div></td>' +
                    '<td><span class="status-indicator ' + statusClass + '"></span>' + statusText + '</td>' +
                    '</tr>';
            });
            
            tbody.innerHTML = html;
        }

        function enableCollector() {
            fetch('/yyhertz/security/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('安全监控已启用');
                    refreshMetrics();
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableCollector() {
            if (confirm('确定要禁用安全监控吗？')) {
                fetch('/yyhertz/security/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('安全监控已禁用');
                        refreshMetrics();
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        function resetMetrics() {
            if (confirm('确定要重置所有安全指标吗？')) {
                fetch('/yyhertz/security/reset', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('安全指标已重置');
                        refreshMetrics();
                    })
                    .catch(error => {
                        console.error('重置失败:', error);
                        alert('重置失败');
                    });
            }
        }

        // 页面加载时初始化
        window.onload = function() {
            loadSecurityData();
            // 每30秒自动刷新
            setInterval(loadSecurityData, 30000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}