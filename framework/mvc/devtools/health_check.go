// Package devtools 提供MVC框架的健康检查功能
//
// 健康检查中间件用于监控应用和系统组件的健康状态，提供：
// - 数据库连接检查
// - 外部服务连通性检查
// - 系统资源状态检查
// - 自定义健康检查项
// - HTTP健康检查端点
// - 详细的健康报告
//
// 功能特性：
// - 支持多种检查类型
// - 异步并发检查
// - 可配置的检查间隔和超时
// - 检查结果缓存
// - 状态变化通知
// - 与监控系统集成
package devtools

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// HealthStatus 健康状态枚举
type HealthStatus string

const (
	// HealthStatusHealthy 健康状态
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusDegraded 降级状态
	HealthStatusDegraded HealthStatus = "degraded"
	// HealthStatusUnhealthy 不健康状态
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusUnknown 未知状态
	HealthStatusUnknown HealthStatus = "unknown"
)

// HealthCheckType 健康检查类型
type HealthCheckType string

const (
	// HealthCheckTypeStartup 启动检查
	HealthCheckTypeStartup HealthCheckType = "startup"
	// HealthCheckTypeReadiness 就绪检查
	HealthCheckTypeReadiness HealthCheckType = "readiness"
	// HealthCheckTypeLiveness 存活检查
	HealthCheckTypeLiveness HealthCheckType = "liveness"
)

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// Name 检查器名称
	Name() string
	// Check 执行健康检查
	Check(ctx context.Context) HealthCheckResult
	// Type 检查类型
	Type() HealthCheckType
	// Timeout 检查超时时间
	Timeout() time.Duration
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	Name      string                 `json:"name"`              // 检查器名称
	Status    HealthStatus           `json:"status"`            // 健康状态
	Message   string                 `json:"message,omitempty"` // 状态消息
	Error     string                 `json:"error,omitempty"`   // 错误信息
	Timestamp time.Time              `json:"timestamp"`         // 检查时间
	Duration  time.Duration          `json:"duration"`          // 检查耗时
	Details   map[string]interface{} `json:"details,omitempty"` // 详细信息
	Type      HealthCheckType        `json:"type"`              // 检查类型
}

// OverallHealthStatus 整体健康状态
type OverallHealthStatus struct {
	Status      HealthStatus        `json:"status"`       // 整体状态
	Timestamp   time.Time           `json:"timestamp"`    // 检查时间
	Duration    time.Duration       `json:"duration"`     // 总检查耗时
	TotalChecks int                 `json:"total_checks"` // 总检查数
	Passed      int                 `json:"passed"`       // 通过检查数
	Failed      int                 `json:"failed"`       // 失败检查数
	Degraded    int                 `json:"degraded"`     // 降级检查数
	Checks      []HealthCheckResult `json:"checks"`       // 各项检查结果
	SystemInfo  SystemHealthInfo    `json:"system_info"`  // 系统信息
}

// SystemHealthInfo 系统健康信息
type SystemHealthInfo struct {
	Uptime      time.Duration `json:"uptime"`       // 运行时间
	Goroutines  int           `json:"goroutines"`   // 协程数
	MemoryUsage uint64        `json:"memory_usage"` // 内存使用量(字节)
	CPUCount    int           `json:"cpu_count"`    // CPU核心数
	GCCount     uint32        `json:"gc_count"`     // GC次数
	Version     string        `json:"version"`      // 应用版本
	Environment string        `json:"environment"`  // 运行环境
	PID         int           `json:"pid"`          // 进程ID
}

// HealthCheckMiddleware 健康检查中间件
type HealthCheckMiddleware struct {
	mu             sync.RWMutex
	checkers       map[string]HealthChecker
	enabled        bool
	checkInterval  time.Duration
	defaultTimeout time.Duration
	lastCheckTime  time.Time
	lastStatus     *OverallHealthStatus
	cache          map[string]*HealthCheckResult
	cacheExpiry    time.Duration
	startTime      time.Time
	version        string
	environment    string
	onStatusChange func(HealthStatus, HealthStatus) // 状态变化回调
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled        bool                             // 是否启用
	CheckInterval  time.Duration                    // 检查间隔
	DefaultTimeout time.Duration                    // 默认超时时间
	CacheExpiry    time.Duration                    // 缓存过期时间
	Version        string                           // 应用版本
	Environment    string                           // 运行环境
	OnStatusChange func(HealthStatus, HealthStatus) // 状态变化回调
}

// NewHealthCheckMiddleware 创建健康检查中间件
func NewHealthCheckMiddleware(config *HealthCheckConfig) *HealthCheckMiddleware {
	if config == nil {
		config = &HealthCheckConfig{}
	}

	// 设置默认值
	if config.CheckInterval <= 0 {
		config.CheckInterval = 30 * time.Second
	}
	if config.DefaultTimeout <= 0 {
		config.DefaultTimeout = 5 * time.Second
	}
	if config.CacheExpiry <= 0 {
		config.CacheExpiry = 10 * time.Second
	}
	if config.Version == "" {
		config.Version = "unknown"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}

	hm := &HealthCheckMiddleware{
		checkers:       make(map[string]HealthChecker),
		enabled:        config.Enabled,
		checkInterval:  config.CheckInterval,
		defaultTimeout: config.DefaultTimeout,
		cache:          make(map[string]*HealthCheckResult),
		cacheExpiry:    config.CacheExpiry,
		startTime:      time.Now(),
		version:        config.Version,
		environment:    config.Environment,
		onStatusChange: config.OnStatusChange,
	}

	// 添加默认的系统检查器
	hm.AddChecker(&SystemResourceChecker{})

	return hm
}

// AddChecker 添加健康检查器
func (hm *HealthCheckMiddleware) AddChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.checkers[checker.Name()] = checker
	config.Debugf("添加健康检查器: %s", checker.Name())
}

// RemoveChecker 移除健康检查器
func (hm *HealthCheckMiddleware) RemoveChecker(name string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	delete(hm.checkers, name)
	delete(hm.cache, name)
	config.Debugf("移除健康检查器: %s", name)
}

// GetCheckers 获取所有检查器
func (hm *HealthCheckMiddleware) GetCheckers() []string {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	names := make([]string, 0, len(hm.checkers))
	for name := range hm.checkers {
		names = append(names, name)
	}
	return names
}

// Enable 启用健康检查
func (hm *HealthCheckMiddleware) Enable() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.enabled = true
}

// Disable 禁用健康检查
func (hm *HealthCheckMiddleware) Disable() {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.enabled = false
}

// IsEnabled 检查是否启用
func (hm *HealthCheckMiddleware) IsEnabled() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.enabled
}

// CheckHealth 执行健康检查
func (hm *HealthCheckMiddleware) CheckHealth(ctx context.Context, checkType ...HealthCheckType) *OverallHealthStatus {
	startTime := time.Now()

	hm.mu.RLock()
	if !hm.enabled {
		hm.mu.RUnlock()
		return &OverallHealthStatus{
			Status:      HealthStatusUnknown,
			Timestamp:   startTime,
			Duration:    time.Since(startTime),
			TotalChecks: 0,
		}
	}

	// 过滤检查器
	var checkersToRun []HealthChecker
	filterType := HealthCheckTypeLiveness // 默认类型
	if len(checkType) > 0 {
		filterType = checkType[0]
	}

	for _, checker := range hm.checkers {
		if checker.Type() == filterType {
			checkersToRun = append(checkersToRun, checker)
		}
	}
	hm.mu.RUnlock()

	// 并发执行检查
	results := make([]HealthCheckResult, len(checkersToRun))
	var wg sync.WaitGroup

	for i, checker := range checkersToRun {
		wg.Add(1)
		go func(index int, c HealthChecker) {
			defer wg.Done()

			// 检查缓存
			if cachedResult := hm.getCachedResult(c.Name()); cachedResult != nil {
				results[index] = *cachedResult
				return
			}

			// 执行检查
			checkCtx, cancel := context.WithTimeout(ctx, c.Timeout())
			defer cancel()

			result := c.Check(checkCtx)
			results[index] = result

			// 缓存结果
			hm.setCachedResult(c.Name(), &result)
		}(i, checker)
	}

	wg.Wait()

	// 计算整体状态
	status := &OverallHealthStatus{
		Timestamp:   startTime,
		Duration:    time.Since(startTime),
		TotalChecks: len(results),
		Checks:      results,
		SystemInfo:  hm.getSystemInfo(),
	}

	// 统计各种状态
	for _, result := range results {
		switch result.Status {
		case HealthStatusHealthy:
			status.Passed++
		case HealthStatusDegraded:
			status.Degraded++
		case HealthStatusUnhealthy:
			status.Failed++
		}
	}

	// 确定整体状态
	if status.Failed > 0 {
		status.Status = HealthStatusUnhealthy
	} else if status.Degraded > 0 {
		status.Status = HealthStatusDegraded
	} else if status.Passed > 0 {
		status.Status = HealthStatusHealthy
	} else {
		status.Status = HealthStatusUnknown
	}

	// 检查状态变化
	hm.mu.Lock()
	oldStatus := HealthStatusUnknown
	if hm.lastStatus != nil {
		oldStatus = hm.lastStatus.Status
	}
	if oldStatus != status.Status && hm.onStatusChange != nil {
		go hm.onStatusChange(oldStatus, status.Status)
	}
	hm.lastStatus = status
	hm.lastCheckTime = startTime
	hm.mu.Unlock()

	return status
}

// getCachedResult 获取缓存结果
func (hm *HealthCheckMiddleware) getCachedResult(name string) *HealthCheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if result, exists := hm.cache[name]; exists {
		if time.Since(result.Timestamp) < hm.cacheExpiry {
			return result
		}
		// 清理过期缓存
		delete(hm.cache, name)
	}
	return nil
}

// setCachedResult 设置缓存结果
func (hm *HealthCheckMiddleware) setCachedResult(name string, result *HealthCheckResult) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.cache[name] = result
}

// getSystemInfo 获取系统信息
func (hm *HealthCheckMiddleware) getSystemInfo() SystemHealthInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemHealthInfo{
		Uptime:      time.Since(hm.startTime),
		Goroutines:  runtime.NumGoroutine(),
		MemoryUsage: m.Alloc,
		CPUCount:    runtime.NumCPU(),
		GCCount:     m.NumGC,
		Version:     hm.version,
		Environment: hm.environment,
		PID:         0, // TODO: 获取进程ID
	}
}

// GetLastStatus 获取最后一次检查状态
func (hm *HealthCheckMiddleware) GetLastStatus() *OverallHealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.lastStatus
}

// SystemResourceChecker 系统资源检查器
type SystemResourceChecker struct{}

func (src *SystemResourceChecker) Name() string {
	return "system_resources"
}

func (src *SystemResourceChecker) Type() HealthCheckType {
	return HealthCheckTypeLiveness
}

func (src *SystemResourceChecker) Timeout() time.Duration {
	return 2 * time.Second
}

func (src *SystemResourceChecker) Check(ctx context.Context) HealthCheckResult {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	result := HealthCheckResult{
		Name:      src.Name(),
		Type:      src.Type(),
		Timestamp: start,
		Duration:  time.Since(start),
		Details: map[string]interface{}{
			"memory_alloc_mb": float64(m.Alloc) / 1024 / 1024,
			"memory_sys_mb":   float64(m.Sys) / 1024 / 1024,
			"gc_count":        m.NumGC,
			"goroutines":      runtime.NumGoroutine(),
			"cpu_cores":       runtime.NumCPU(),
		},
	}

	// 简单的资源检查逻辑
	memoryUsageMB := float64(m.Alloc) / 1024 / 1024
	goroutineCount := runtime.NumGoroutine()

	if memoryUsageMB > 1000 || goroutineCount > 10000 {
		result.Status = HealthStatusUnhealthy
		result.Message = "系统资源使用过高"
	} else if memoryUsageMB > 500 || goroutineCount > 5000 {
		result.Status = HealthStatusDegraded
		result.Message = "系统资源使用较高"
	} else {
		result.Status = HealthStatusHealthy
		result.Message = "系统资源正常"
	}

	return result
}

// HealthCheckPanel 健康检查面板
type HealthCheckPanel struct {
	middleware *HealthCheckMiddleware
}

// NewHealthCheckPanel 创建健康检查面板
func NewHealthCheckPanel(middleware *HealthCheckMiddleware) *HealthCheckPanel {
	return &HealthCheckPanel{
		middleware: middleware,
	}
}

// RegisterRoutes 注册健康检查路由
func (hcp *HealthCheckPanel) RegisterRoutes(engine any) {
	// 类型断言，支持不同的引擎类型
	var healthGroup *route.RouterGroup

	// 尝试不同的类型断言
	if h, ok := engine.(*route.Engine); ok {
		healthGroup = h.Group("/yyhertz/health")
	} else {
		config.Error("无法注册健康检查路由，未知引擎类型")
		return
	}

	// 注册路由的通用方法
	registerRoute := func(method, path string, handler func(ctx context.Context, c *app.RequestContext)) {
		switch method {
		case "GET":
			healthGroup.GET(path, handler)
		case "POST":
			healthGroup.POST(path, handler)
		default:
			config.Warnf("不支持的HTTP方法: %s", method)
		}
	}

	// Kubernetes风格的健康检查端点
	registerRoute("GET", "/live", hcp.livenessCheck)   // 存活检查
	registerRoute("GET", "/ready", hcp.readinessCheck) // 就绪检查
	registerRoute("GET", "/startup", hcp.startupCheck) // 启动检查

	// 通用健康检查端点
	registerRoute("GET", "/", hcp.healthCheck)       // 全面健康检查
	registerRoute("GET", "/status", hcp.healthCheck) // 别名

	// 健康检查面板
	registerRoute("GET", "/panel", hcp.healthPanel)

	// 管理接口
	registerRoute("GET", "/checkers", hcp.getCheckers)   // 获取检查器列表
	registerRoute("POST", "/enable", hcp.enableHealth)   // 启用健康检查
	registerRoute("POST", "/disable", hcp.disableHealth) // 禁用健康检查
}

// livenessCheck 存活检查 - 用于K8s liveness probe
func (hcp *HealthCheckPanel) livenessCheck(ctx context.Context, c *app.RequestContext) {
	status := hcp.middleware.CheckHealth(ctx, HealthCheckTypeLiveness)

	if status.Status == HealthStatusHealthy || status.Status == HealthStatusDegraded {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

// readinessCheck 就绪检查 - 用于K8s readiness probe
func (hcp *HealthCheckPanel) readinessCheck(ctx context.Context, c *app.RequestContext) {
	status := hcp.middleware.CheckHealth(ctx, HealthCheckTypeReadiness)

	if status.Status == HealthStatusHealthy {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

// startupCheck 启动检查 - 用于K8s startup probe
func (hcp *HealthCheckPanel) startupCheck(ctx context.Context, c *app.RequestContext) {
	status := hcp.middleware.CheckHealth(ctx, HealthCheckTypeStartup)

	if status.Status == HealthStatusHealthy {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

// healthCheck 全面健康检查
func (hcp *HealthCheckPanel) healthCheck(ctx context.Context, c *app.RequestContext) {
	status := hcp.middleware.CheckHealth(ctx)

	// 根据overall status决定HTTP状态码
	var httpStatus int
	switch status.Status {
	case HealthStatusHealthy:
		httpStatus = http.StatusOK
	case HealthStatusDegraded:
		httpStatus = http.StatusOK // 降级状态仍返回200，但在响应体中标注
	case HealthStatusUnhealthy:
		httpStatus = http.StatusServiceUnavailable
	default:
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, status)
}

// getCheckers 获取检查器列表
func (hcp *HealthCheckPanel) getCheckers(ctx context.Context, c *app.RequestContext) {
	checkers := hcp.middleware.GetCheckers()

	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"enabled":  hcp.middleware.IsEnabled(),
			"checkers": checkers,
			"count":    len(checkers),
		},
	})
}

// enableHealth 启用健康检查
func (hcp *HealthCheckPanel) enableHealth(ctx context.Context, c *app.RequestContext) {
	hcp.middleware.Enable()

	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "健康检查已启用",
		"enabled": true,
	})
}

// disableHealth 禁用健康检查
func (hcp *HealthCheckPanel) disableHealth(ctx context.Context, c *app.RequestContext) {
	hcp.middleware.Disable()

	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "健康检查已禁用",
		"enabled": false,
	})
}

// healthPanel 健康检查面板页面
func (hcp *HealthCheckPanel) healthPanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 健康检查面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .status-card { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .status-healthy { border-left: 5px solid #28a745; }
        .status-degraded { border-left: 5px solid #ffc107; }
        .status-unhealthy { border-left: 5px solid #dc3545; }
        .status-unknown { border-left: 5px solid #6c757d; }
        .status-indicator { display: inline-block; padding: 4px 12px; border-radius: 20px; font-size: 12px; font-weight: bold; text-transform: uppercase; }
        .healthy { background: #d4edda; color: #155724; }
        .degraded { background: #fff3cd; color: #856404; }
        .unhealthy { background: #f8d7da; color: #721c24; }
        .unknown { background: #e2e3e5; color: #383d41; }
        .checks-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .check-item { background: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .check-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
        .check-details { font-size: 14px; color: #666; }
        .system-info { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; }
        .info-item { text-align: center; }
        .info-value { font-size: 1.5em; font-weight: bold; color: #007bff; }
        .info-label { color: #666; font-size: 14px; margin-top: 5px; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .controls { margin-bottom: 20px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 健康检查面板</h1>
        <div class="controls">
            <button class="btn btn-primary" onclick="refreshHealth()">刷新状态</button>
            <button class="btn btn-success" onclick="enableHealth()">启用检查</button>
            <button class="btn btn-danger" onclick="disableHealth()">禁用检查</button>
        </div>
    </div>

    <div id="overallStatus" class="status-card">
        <div style="text-align: center; color: #666;">加载中...</div>
    </div>

    <div id="checksGrid" class="checks-grid">
        <!-- 检查项将在这里动态生成 -->
    </div>

    <div class="system-info">
        <h3>系统信息</h3>
        <div id="systemInfo" class="info-grid">
            <!-- 系统信息将在这里动态生成 -->
        </div>
    </div>

    <script>
        function loadHealthStatus() {
            fetch('/yyhertz/health/')
                .then(response => response.json())
                .then(data => {
                    updateOverallStatus(data);
                    updateChecksGrid(data.checks || []);
                    updateSystemInfo(data.system_info || {});
                })
                .catch(error => {
                    console.error('加载健康状态失败:', error);
                    document.getElementById('overallStatus').innerHTML = 
                        '<div style="text-align: center; color: red;">加载失败</div>';
                });
        }

        function updateOverallStatus(status) {
            const statusEl = document.getElementById('overallStatus');
            const statusClass = 'status-' + status.status;
            const indicatorClass = status.status;
            
            statusEl.className = 'status-card ' + statusClass;
            statusEl.innerHTML = 
                '<div style="display: flex; justify-content: space-between; align-items: center;">' +
                    '<div>' +
                        '<h2>整体状态: <span class="status-indicator ' + indicatorClass + '">' + status.status + '</span></h2>' +
                        '<p>检查时间: ' + new Date(status.timestamp).toLocaleString() + '</p>' +
                        '<p>检查耗时: ' + status.duration + '</p>' +
                    '</div>' +
                    '<div style="text-align: right;">' +
                        '<div style="font-size: 24px; font-weight: bold;">通过率</div>' +
                        '<div style="font-size: 36px; color: #007bff;">' + 
                            Math.round((status.passed / status.total_checks) * 100) + '%' +
                        '</div>' +
                        '<div style="color: #666;">' + status.passed + '/' + status.total_checks + ' 项通过</div>' +
                    '</div>' +
                '</div>';
        }

        function updateChecksGrid(checks) {
            const grid = document.getElementById('checksGrid');
            
            if (checks.length === 0) {
                grid.innerHTML = '<div style="text-align: center; color: #666; grid-column: 1 / -1;">暂无检查项</div>';
                return;
            }

            let html = '';
            checks.forEach(check => {
                const statusClass = check.status;
                html += 
                    '<div class="check-item">' +
                        '<div class="check-header">' +
                            '<h4>' + check.name + '</h4>' +
                            '<span class="status-indicator ' + statusClass + '">' + check.status + '</span>' +
                        '</div>' +
                        '<div class="check-details">' +
                            '<p><strong>类型:</strong> ' + check.type + '</p>' +
                            '<p><strong>消息:</strong> ' + (check.message || '无') + '</p>' +
                            '<p><strong>检查时间:</strong> ' + new Date(check.timestamp).toLocaleString() + '</p>' +
                            '<p><strong>耗时:</strong> ' + check.duration + '</p>' +
                            (check.error ? '<p style="color: red;"><strong>错误:</strong> ' + check.error + '</p>' : '') +
                            (check.details ? '<pre style="background: #f8f9fa; padding: 10px; border-radius: 4px; font-size: 12px;">' + 
                                JSON.stringify(check.details, null, 2) + '</pre>' : '') +
                        '</div>' +
                    '</div>';
            });
            
            grid.innerHTML = html;
        }

        function updateSystemInfo(systemInfo) {
            const infoEl = document.getElementById('systemInfo');
            
            infoEl.innerHTML = 
                '<div class="info-item">' +
                    '<div class="info-value">' + formatDuration(systemInfo.uptime || 0) + '</div>' +
                    '<div class="info-label">运行时间</div>' +
                '</div>' +
                '<div class="info-item">' +
                    '<div class="info-value">' + (systemInfo.goroutines || 0) + '</div>' +
                    '<div class="info-label">协程数</div>' +
                '</div>' +
                '<div class="info-item">' +
                    '<div class="info-value">' + formatBytes(systemInfo.memory_usage || 0) + '</div>' +
                    '<div class="info-label">内存使用</div>' +
                '</div>' +
                '<div class="info-item">' +
                    '<div class="info-value">' + (systemInfo.cpu_count || 0) + '</div>' +
                    '<div class="info-label">CPU核心数</div>' +
                '</div>' +
                '<div class="info-item">' +
                    '<div class="info-value">' + (systemInfo.gc_count || 0) + '</div>' +
                    '<div class="info-label">GC次数</div>' +
                '</div>' +
                '<div class="info-item">' +
                    '<div class="info-value">' + (systemInfo.version || 'unknown') + '</div>' +
                    '<div class="info-label">版本</div>' +
                '</div>';
        }

        function formatDuration(nanoseconds) {
            const seconds = Math.floor(nanoseconds / 1e9);
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const remainingSeconds = seconds % 60;
            
            if (hours > 0) {
                return hours + 'h ' + minutes + 'm';
            } else if (minutes > 0) {
                return minutes + 'm ' + remainingSeconds + 's';
            } else {
                return remainingSeconds + 's';
            }
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function refreshHealth() {
            loadHealthStatus();
        }

        function enableHealth() {
            fetch('/yyhertz/health/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('健康检查已启用');
                    loadHealthStatus();
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableHealth() {
            if (confirm('确定要禁用健康检查吗？')) {
                fetch('/yyhertz/health/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('健康检查已禁用');
                        loadHealthStatus();
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        // 页面加载时自动加载健康状态
        window.onload = function() {
            loadHealthStatus();
            // 每30秒自动刷新
            setInterval(loadHealthStatus, 30000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}

// Handler 返回健康检查中间件处理函数
func (hm *HealthCheckMiddleware) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 健康检查中间件通常不需要拦截所有请求
		// 这里只是简单地传递到下一个中间件
		c.Next(ctx)
	}
}
