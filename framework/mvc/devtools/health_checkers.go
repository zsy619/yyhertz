// Package devtools 提供常用的健康检查器实现
//
// 健康检查器集合包含：
// - 数据库连接检查器
// - Redis缓存检查器
// - HTTP服务检查器
// - 磁盘空间检查器
// - 自定义检查器
//
// 功能特性：
// - 支持多种数据库类型
// - 连接池状态检查
// - 网络连通性检查
// - 资源使用率监控
// - 可配置的检查参数
package devtools

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// DatabaseChecker 数据库健康检查器
type DatabaseChecker struct {
	name     string
	db       *sql.DB
	query    string
	timeout  time.Duration
	checkType HealthCheckType
}

// NewDatabaseChecker 创建数据库检查器
func NewDatabaseChecker(name string, db *sql.DB, options ...DatabaseCheckerOption) *DatabaseChecker {
	checker := &DatabaseChecker{
		name:     name,
		db:       db,
		query:    "SELECT 1",
		timeout:  5 * time.Second,
		checkType: HealthCheckTypeLiveness,
	}
	
	for _, option := range options {
		option(checker)
	}
	
	return checker
}

// DatabaseCheckerOption 数据库检查器选项
type DatabaseCheckerOption func(*DatabaseChecker)

// WithDatabaseQuery 设置检查查询
func WithDatabaseQuery(query string) DatabaseCheckerOption {
	return func(checker *DatabaseChecker) {
		checker.query = query
	}
}

// WithDatabaseTimeout 设置检查超时
func WithDatabaseTimeout(timeout time.Duration) DatabaseCheckerOption {
	return func(checker *DatabaseChecker) {
		checker.timeout = timeout
	}
}

// WithDatabaseCheckType 设置检查类型
func WithDatabaseCheckType(checkType HealthCheckType) DatabaseCheckerOption {
	return func(checker *DatabaseChecker) {
		checker.checkType = checkType
	}
}

func (dc *DatabaseChecker) Name() string {
	return dc.name
}

func (dc *DatabaseChecker) Type() HealthCheckType {
	return dc.checkType
}

func (dc *DatabaseChecker) Timeout() time.Duration {
	return dc.timeout
}

func (dc *DatabaseChecker) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()
	
	result := HealthCheckResult{
		Name:      dc.name,
		Type:      dc.checkType,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}
	
	// 检查数据库是否为nil
	if dc.db == nil {
		result.Status = HealthStatusUnhealthy
		result.Error = "database connection is nil"
		result.Duration = time.Since(start)
		return result
	}
	
	// 获取连接池统计
	stats := dc.db.Stats()
	result.Details["open_connections"] = stats.OpenConnections
	result.Details["in_use"] = stats.InUse
	result.Details["idle"] = stats.Idle
	result.Details["wait_count"] = stats.WaitCount
	result.Details["wait_duration"] = stats.WaitDuration.String()
	result.Details["max_idle_closed"] = stats.MaxIdleClosed
	result.Details["max_lifetime_closed"] = stats.MaxLifetimeClosed
	
	// 执行检查查询
	queryCtx, cancel := context.WithTimeout(ctx, dc.timeout)
	defer cancel()
	
	var testResult sql.NullString
	err := dc.db.QueryRowContext(queryCtx, dc.query).Scan(&testResult)
	
	result.Duration = time.Since(start)
	
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Error = err.Error()
		result.Message = "数据库查询失败"
		return result
	}
	
	// 检查连接池状态
	if stats.OpenConnections == 0 {
		result.Status = HealthStatusUnhealthy
		result.Message = "没有可用的数据库连接"
		return result
	}
	
	// 检查是否有太多等待的连接
	if stats.WaitCount > 100 {
		result.Status = HealthStatusDegraded
		result.Message = "数据库连接池压力较大"
		return result
	}
	
	result.Status = HealthStatusHealthy
	result.Message = "数据库连接正常"
	return result
}

// RedisChecker Redis健康检查器
type RedisChecker struct {
	name      string
	addr      string
	password  string
	db        int
	timeout   time.Duration
	checkType HealthCheckType
}

// NewRedisChecker 创建Redis检查器
func NewRedisChecker(name, addr string, options ...RedisCheckerOption) *RedisChecker {
	checker := &RedisChecker{
		name:      name,
		addr:      addr,
		timeout:   5 * time.Second,
		checkType: HealthCheckTypeLiveness,
	}
	
	for _, option := range options {
		option(checker)
	}
	
	return checker
}

// RedisCheckerOption Redis检查器选项
type RedisCheckerOption func(*RedisChecker)

// WithRedisPassword 设置Redis密码
func WithRedisPassword(password string) RedisCheckerOption {
	return func(checker *RedisChecker) {
		checker.password = password
	}
}

// WithRedisDB 设置Redis数据库
func WithRedisDB(db int) RedisCheckerOption {
	return func(checker *RedisChecker) {
		checker.db = db
	}
}

// WithRedisTimeout 设置检查超时
func WithRedisTimeout(timeout time.Duration) RedisCheckerOption {
	return func(checker *RedisChecker) {
		checker.timeout = timeout
	}
}

// WithRedisCheckType 设置检查类型
func WithRedisCheckType(checkType HealthCheckType) RedisCheckerOption {
	return func(checker *RedisChecker) {
		checker.checkType = checkType
	}
}

func (rc *RedisChecker) Name() string {
	return rc.name
}

func (rc *RedisChecker) Type() HealthCheckType {
	return rc.checkType
}

func (rc *RedisChecker) Timeout() time.Duration {
	return rc.timeout
}

func (rc *RedisChecker) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()
	
	result := HealthCheckResult{
		Name:      rc.name,
		Type:      rc.checkType,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}
	
	// 简单的TCP连接检查（模拟Redis检查）
	conn, err := net.DialTimeout("tcp", rc.addr, rc.timeout)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis连接失败"
		result.Duration = time.Since(start)
		return result
	}
	defer conn.Close()
	
	// 设置读写超时
	conn.SetDeadline(time.Now().Add(rc.timeout))
	
	// 发送PING命令
	_, err = conn.Write([]byte("PING\r\n"))
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis PING命令发送失败"
		result.Duration = time.Since(start)
		return result
	}
	
	// 读取响应
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Error = err.Error()
		result.Message = "Redis PING响应读取失败"
		result.Duration = time.Since(start)
		return result
	}
	
	response := string(buffer[:n])
	result.Details["response"] = response
	result.Details["addr"] = rc.addr
	
	result.Duration = time.Since(start)
	
	// 检查响应
	if len(response) > 0 && (response[0] == '+' || response[0] == ':') {
		result.Status = HealthStatusHealthy
		result.Message = "Redis连接正常"
	} else {
		result.Status = HealthStatusDegraded
		result.Message = "Redis响应异常"
	}
	
	return result
}

// HTTPServiceChecker HTTP服务健康检查器
type HTTPServiceChecker struct {
	name           string
	url            string
	method         string
	expectedStatus int
	timeout        time.Duration
	checkType      HealthCheckType
	headers        map[string]string
}

// NewHTTPServiceChecker 创建HTTP服务检查器
func NewHTTPServiceChecker(name, url string, options ...HTTPServiceCheckerOption) *HTTPServiceChecker {
	checker := &HTTPServiceChecker{
		name:           name,
		url:            url,
		method:         "GET",
		expectedStatus: 200,
		timeout:        10 * time.Second,
		checkType:      HealthCheckTypeReadiness,
		headers:        make(map[string]string),
	}
	
	for _, option := range options {
		option(checker)
	}
	
	return checker
}

// HTTPServiceCheckerOption HTTP服务检查器选项
type HTTPServiceCheckerOption func(*HTTPServiceChecker)

// WithHTTPMethod 设置HTTP方法
func WithHTTPMethod(method string) HTTPServiceCheckerOption {
	return func(checker *HTTPServiceChecker) {
		checker.method = method
	}
}

// WithHTTPExpectedStatus 设置期望的状态码
func WithHTTPExpectedStatus(status int) HTTPServiceCheckerOption {
	return func(checker *HTTPServiceChecker) {
		checker.expectedStatus = status
	}
}

// WithHTTPTimeout 设置超时时间
func WithHTTPTimeout(timeout time.Duration) HTTPServiceCheckerOption {
	return func(checker *HTTPServiceChecker) {
		checker.timeout = timeout
	}
}

// WithHTTPCheckType 设置检查类型
func WithHTTPCheckType(checkType HealthCheckType) HTTPServiceCheckerOption {
	return func(checker *HTTPServiceChecker) {
		checker.checkType = checkType
	}
}

// WithHTTPHeaders 设置请求头
func WithHTTPHeaders(headers map[string]string) HTTPServiceCheckerOption {
	return func(checker *HTTPServiceChecker) {
		checker.headers = headers
	}
}

func (hsc *HTTPServiceChecker) Name() string {
	return hsc.name
}

func (hsc *HTTPServiceChecker) Type() HealthCheckType {
	return hsc.checkType
}

func (hsc *HTTPServiceChecker) Timeout() time.Duration {
	return hsc.timeout
}

func (hsc *HTTPServiceChecker) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()
	
	result := HealthCheckResult{
		Name:      hsc.name,
		Type:      hsc.checkType,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: hsc.timeout,
	}
	
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, hsc.method, hsc.url, nil)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Error = err.Error()
		result.Message = "HTTP请求创建失败"
		result.Duration = time.Since(start)
		return result
	}
	
	// 设置请求头
	for key, value := range hsc.headers {
		req.Header.Set(key, value)
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Error = err.Error()
		result.Message = "HTTP请求失败"
		result.Duration = time.Since(start)
		return result
	}
	defer resp.Body.Close()
	
	result.Duration = time.Since(start)
	result.Details["url"] = hsc.url
	result.Details["method"] = hsc.method
	result.Details["status_code"] = resp.StatusCode
	result.Details["response_time_ms"] = result.Duration.Milliseconds()
	
	// 检查状态码
	if resp.StatusCode == hsc.expectedStatus {
		result.Status = HealthStatusHealthy
		result.Message = fmt.Sprintf("HTTP服务正常 (状态码: %d)", resp.StatusCode)
	} else {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("HTTP服务状态码异常: 期望 %d, 实际 %d", hsc.expectedStatus, resp.StatusCode)
	}
	
	return result
}

// DiskSpaceChecker 磁盘空间检查器
type DiskSpaceChecker struct {
	name            string
	path            string
	warningPercent  float64
	criticalPercent float64
	timeout         time.Duration
	checkType       HealthCheckType
}

// NewDiskSpaceChecker 创建磁盘空间检查器
func NewDiskSpaceChecker(name, path string, options ...DiskSpaceCheckerOption) *DiskSpaceChecker {
	checker := &DiskSpaceChecker{
		name:            name,
		path:            path,
		warningPercent:  80.0,
		criticalPercent: 90.0,
		timeout:         5 * time.Second,
		checkType:       HealthCheckTypeLiveness,
	}
	
	for _, option := range options {
		option(checker)
	}
	
	return checker
}

// DiskSpaceCheckerOption 磁盘空间检查器选项
type DiskSpaceCheckerOption func(*DiskSpaceChecker)

// WithDiskWarningPercent 设置警告百分比
func WithDiskWarningPercent(percent float64) DiskSpaceCheckerOption {
	return func(checker *DiskSpaceChecker) {
		checker.warningPercent = percent
	}
}

// WithDiskCriticalPercent 设置严重百分比
func WithDiskCriticalPercent(percent float64) DiskSpaceCheckerOption {
	return func(checker *DiskSpaceChecker) {
		checker.criticalPercent = percent
	}
}

// WithDiskTimeout 设置超时时间
func WithDiskTimeout(timeout time.Duration) DiskSpaceCheckerOption {
	return func(checker *DiskSpaceChecker) {
		checker.timeout = timeout
	}
}

// WithDiskCheckType 设置检查类型
func WithDiskCheckType(checkType HealthCheckType) DiskSpaceCheckerOption {
	return func(checker *DiskSpaceChecker) {
		checker.checkType = checkType
	}
}

func (dsc *DiskSpaceChecker) Name() string {
	return dsc.name
}

func (dsc *DiskSpaceChecker) Type() HealthCheckType {
	return dsc.checkType
}

func (dsc *DiskSpaceChecker) Timeout() time.Duration {
	return dsc.timeout
}

func (dsc *DiskSpaceChecker) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()
	
	result := HealthCheckResult{
		Name:      dsc.name,
		Type:      dsc.checkType,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}
	
	// 获取磁盘使用情况
	usage, err := dsc.getDiskUsage(dsc.path)
	if err != nil {
		result.Status = HealthStatusUnhealthy
		result.Error = err.Error()
		result.Message = "获取磁盘使用情况失败"
		result.Duration = time.Since(start)
		return result
	}
	
	result.Duration = time.Since(start)
	result.Details["path"] = dsc.path
	result.Details["total_bytes"] = usage.Total
	result.Details["used_bytes"] = usage.Used
	result.Details["free_bytes"] = usage.Free
	result.Details["used_percent"] = usage.UsedPercent
	result.Details["total_gb"] = float64(usage.Total) / (1024 * 1024 * 1024)
	result.Details["used_gb"] = float64(usage.Used) / (1024 * 1024 * 1024)
	result.Details["free_gb"] = float64(usage.Free) / (1024 * 1024 * 1024)
	
	// 判断磁盘使用状态
	if usage.UsedPercent >= dsc.criticalPercent {
		result.Status = HealthStatusUnhealthy
		result.Message = fmt.Sprintf("磁盘使用率过高: %.1f%% (严重阈值: %.1f%%)", usage.UsedPercent, dsc.criticalPercent)
	} else if usage.UsedPercent >= dsc.warningPercent {
		result.Status = HealthStatusDegraded
		result.Message = fmt.Sprintf("磁盘使用率较高: %.1f%% (警告阈值: %.1f%%)", usage.UsedPercent, dsc.warningPercent)
	} else {
		result.Status = HealthStatusHealthy
		result.Message = fmt.Sprintf("磁盘使用率正常: %.1f%%", usage.UsedPercent)
	}
	
	return result
}

// DiskUsage 磁盘使用情况
type DiskUsage struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// getDiskUsage 获取磁盘使用情况
func (dsc *DiskSpaceChecker) getDiskUsage(path string) (*DiskUsage, error) {
	var stat syscall.Statfs_t
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, err
	}
	
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free
	
	var usedPercent float64
	if total > 0 {
		usedPercent = (float64(used) / float64(total)) * 100
	}
	
	return &DiskUsage{
		Total:       total,
		Used:        used,
		Free:        free,
		UsedPercent: usedPercent,
	}, nil
}

// CustomChecker 自定义检查器
type CustomChecker struct {
	name      string
	checkFunc func(ctx context.Context) HealthCheckResult
	timeout   time.Duration
	checkType HealthCheckType
}

// NewCustomChecker 创建自定义检查器
func NewCustomChecker(name string, checkFunc func(ctx context.Context) HealthCheckResult, options ...CustomCheckerOption) *CustomChecker {
	checker := &CustomChecker{
		name:      name,
		checkFunc: checkFunc,
		timeout:   5 * time.Second,
		checkType: HealthCheckTypeLiveness,
	}
	
	for _, option := range options {
		option(checker)
	}
	
	return checker
}

// CustomCheckerOption 自定义检查器选项
type CustomCheckerOption func(*CustomChecker)

// WithCustomTimeout 设置超时时间
func WithCustomTimeout(timeout time.Duration) CustomCheckerOption {
	return func(checker *CustomChecker) {
		checker.timeout = timeout
	}
}

// WithCustomCheckType 设置检查类型
func WithCustomCheckType(checkType HealthCheckType) CustomCheckerOption {
	return func(checker *CustomChecker) {
		checker.checkType = checkType
	}
}

func (cc *CustomChecker) Name() string {
	return cc.name
}

func (cc *CustomChecker) Type() HealthCheckType {
	return cc.checkType
}

func (cc *CustomChecker) Timeout() time.Duration {
	return cc.timeout
}

func (cc *CustomChecker) Check(ctx context.Context) HealthCheckResult {
	if cc.checkFunc == nil {
		return HealthCheckResult{
			Name:      cc.name,
			Type:      cc.checkType,
			Status:    HealthStatusUnhealthy,
			Error:     "check function is nil",
			Message:   "自定义检查函数未设置",
			Timestamp: time.Now(),
			Duration:  0,
		}
	}
	
	return cc.checkFunc(ctx)
}

// 便利函数：创建常用的健康检查器

// CreateMySQLChecker 创建MySQL检查器
func CreateMySQLChecker(name string, db *sql.DB) *DatabaseChecker {
	return NewDatabaseChecker(name, db,
		WithDatabaseQuery("SELECT 1"),
		WithDatabaseTimeout(5*time.Second),
		WithDatabaseCheckType(HealthCheckTypeLiveness),
	)
}

// CreatePostgreSQLChecker 创建PostgreSQL检查器
func CreatePostgreSQLChecker(name string, db *sql.DB) *DatabaseChecker {
	return NewDatabaseChecker(name, db,
		WithDatabaseQuery("SELECT 1"),
		WithDatabaseTimeout(5*time.Second),
		WithDatabaseCheckType(HealthCheckTypeLiveness),
	)
}

// CreateRedisChecker 创建Redis检查器
func CreateRedisChecker(name, addr string) *RedisChecker {
	return NewRedisChecker(name, addr,
		WithRedisTimeout(5*time.Second),
		WithRedisCheckType(HealthCheckTypeLiveness),
	)
}

// CreateHTTPChecker 创建HTTP检查器
func CreateHTTPChecker(name, url string) *HTTPServiceChecker {
	return NewHTTPServiceChecker(name, url,
		WithHTTPMethod("GET"),
		WithHTTPExpectedStatus(200),
		WithHTTPTimeout(10*time.Second),
		WithHTTPCheckType(HealthCheckTypeReadiness),
	)
}

// CreateDiskChecker 创建磁盘检查器
func CreateDiskChecker(name, path string) *DiskSpaceChecker {
	return NewDiskSpaceChecker(name, path,
		WithDiskWarningPercent(80.0),
		WithDiskCriticalPercent(90.0),
		WithDiskTimeout(5*time.Second),
		WithDiskCheckType(HealthCheckTypeLiveness),
	)
}

// 示例：创建一个简单的文件存在检查器
func CreateFileExistsChecker(name, filePath string) *CustomChecker {
	return NewCustomChecker(name, func(ctx context.Context) HealthCheckResult {
		start := time.Now()
		
		result := HealthCheckResult{
			Name:      name,
			Type:      HealthCheckTypeLiveness,
			Timestamp: start,
			Details:   map[string]interface{}{"file_path": filePath},
		}
		
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			result.Status = HealthStatusUnhealthy
			result.Error = err.Error()
			result.Message = fmt.Sprintf("文件不存在: %s", filePath)
		} else if err != nil {
			result.Status = HealthStatusDegraded
			result.Error = err.Error()
			result.Message = fmt.Sprintf("检查文件时出错: %s", filePath)
		} else {
			result.Status = HealthStatusHealthy
			result.Message = fmt.Sprintf("文件存在: %s", filePath)
		}
		
		result.Duration = time.Since(start)
		return result
	}, WithCustomTimeout(2*time.Second))
}

// AddHealthCheckers 批量添加健康检查器到中间件
func AddHealthCheckers(middleware *HealthCheckMiddleware, checkers ...HealthChecker) {
	for _, checker := range checkers {
		middleware.AddChecker(checker)
		config.Debugf("添加健康检查器: %s", checker.Name())
	}
}