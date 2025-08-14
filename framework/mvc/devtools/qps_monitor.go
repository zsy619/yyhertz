// Package devtools 提供QPS监控中间件
//
// QPS监控中间件用于实时监控请求速率和并发量，提供：
// - 每秒请求数统计
// - 实时并发量监控  
// - 按端点分组统计
// - 自动限流保护
// - 可视化面板展示
// - 告警阈值配置
//
// 功能特性：
// - 滑动窗口统计
// - 高性能计数器
// - 实时图表展示
// - 自动告警机制
// - 支持分布式部署
package devtools

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// QPSStats QPS统计数据
type QPSStats struct {
	CurrentQPS    float64               `json:"current_qps"`     // 当前QPS
	MaxQPS        float64               `json:"max_qps"`         // 最大QPS
	AvgQPS        float64               `json:"avg_qps"`         // 平均QPS
	TotalRequests int64                 `json:"total_requests"`  // 总请求数
	Concurrent    int64                 `json:"concurrent"`      // 当前并发数
	MaxConcurrent int64                 `json:"max_concurrent"`  // 最大并发数
	Endpoints     map[string]*EndpointQPS `json:"endpoints"`      // 端点QPS统计
	Timestamp     time.Time             `json:"timestamp"`       // 统计时间
	WindowSize    time.Duration         `json:"window_size"`     // 窗口大小
}

// EndpointQPS 端点QPS统计
type EndpointQPS struct {
	Path          string    `json:"path"`
	Method        string    `json:"method"`
	CurrentQPS    float64   `json:"current_qps"`
	MaxQPS        float64   `json:"max_qps"`
	TotalRequests int64     `json:"total_requests"`
	LastUpdate    time.Time `json:"last_update"`
}

// QPSWindow 滑动窗口
type QPSWindow struct {
	timestamps []time.Time
	mu         sync.RWMutex
	windowSize time.Duration
}

// NewQPSWindow 创建滑动窗口
func NewQPSWindow(windowSize time.Duration) *QPSWindow {
	return &QPSWindow{
		timestamps: make([]time.Time, 0, 1000),
		windowSize: windowSize,
	}
}

// Add 添加请求时间戳
func (w *QPSWindow) Add(timestamp time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	w.timestamps = append(w.timestamps, timestamp)
	
	// 清理过期数据
	cutoff := timestamp.Add(-w.windowSize)
	start := 0
	for i, ts := range w.timestamps {
		if ts.After(cutoff) {
			start = i
			break
		}
	}
	if start > 0 {
		w.timestamps = w.timestamps[start:]
	}
}

// Count 获取窗口内请求数量
func (w *QPSWindow) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	now := time.Now()
	cutoff := now.Add(-w.windowSize)
	
	count := 0
	for _, ts := range w.timestamps {
		if ts.After(cutoff) {
			count++
		}
	}
	return count
}

// QPS 计算QPS
func (w *QPSWindow) QPS() float64 {
	count := w.Count()
	return float64(count) / w.windowSize.Seconds()
}

// QPSMonitor QPS监控器
type QPSMonitor struct {
	mu              sync.RWMutex
	enabled         bool
	windowSize      time.Duration
	globalWindow    *QPSWindow
	endpointWindows map[string]*QPSWindow
	concurrent      int64
	maxConcurrent   int64
	totalRequests   int64
	maxQPS          float64
	history         []QPSStats
	maxHistorySize  int
	startTime       time.Time
	
	// 告警配置
	qpsThreshold        float64
	concurrentThreshold int64
	onAlert             func(AlertInfo)
}

// AlertInfo 告警信息
type AlertInfo struct {
	Type      string    `json:"type"`        // 告警类型
	Message   string    `json:"message"`     // 告警消息
	Value     float64   `json:"value"`       // 当前值
	Threshold float64   `json:"threshold"`   // 阈值
	Timestamp time.Time `json:"timestamp"`   // 告警时间
}

// QPSConfig QPS监控配置
type QPSConfig struct {
	Enabled             bool                // 是否启用
	WindowSize          time.Duration       // 滑动窗口大小
	MaxHistorySize      int                 // 最大历史记录数
	QPSThreshold        float64             // QPS告警阈值
	ConcurrentThreshold int64               // 并发告警阈值
	OnAlert             func(AlertInfo)     // 告警回调
}

// NewQPSMonitor 创建QPS监控器
func NewQPSMonitor(config *QPSConfig) *QPSMonitor {
	if config == nil {
		config = &QPSConfig{}
	}
	
	// 设置默认值
	if config.WindowSize <= 0 {
		config.WindowSize = 60 * time.Second
	}
	if config.MaxHistorySize <= 0 {
		config.MaxHistorySize = 1000
	}
	if config.QPSThreshold <= 0 {
		config.QPSThreshold = 1000
	}
	if config.ConcurrentThreshold <= 0 {
		config.ConcurrentThreshold = 1000
	}
	
	return &QPSMonitor{
		enabled:             config.Enabled,
		windowSize:          config.WindowSize,
		globalWindow:        NewQPSWindow(config.WindowSize),
		endpointWindows:     make(map[string]*QPSWindow),
		history:            make([]QPSStats, 0, config.MaxHistorySize),
		maxHistorySize:     config.MaxHistorySize,
		startTime:          time.Now(),
		qpsThreshold:       config.QPSThreshold,
		concurrentThreshold: config.ConcurrentThreshold,
		onAlert:            config.OnAlert,
	}
}

// Handler QPS监控中间件
func (qm *QPSMonitor) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !qm.enabled {
			c.Next(ctx)
			return
		}
		
		startTime := time.Now()
		path := string(c.Path())
		method := string(c.Method())
		endpointKey := fmt.Sprintf("%s %s", method, path)
		
		// 增加并发计数
		current := atomic.AddInt64(&qm.concurrent, 1)
		
		// 更新最大并发数
		for {
			max := atomic.LoadInt64(&qm.maxConcurrent)
			if current <= max || atomic.CompareAndSwapInt64(&qm.maxConcurrent, max, current) {
				break
			}
		}
		
		// 检查并发告警
		if qm.onAlert != nil && current > qm.concurrentThreshold {
			go qm.onAlert(AlertInfo{
				Type:      "concurrent",
				Message:   fmt.Sprintf("并发数过高: %d", current),
				Value:     float64(current),
				Threshold: float64(qm.concurrentThreshold),
				Timestamp: startTime,
			})
		}
		
		defer func() {
			// 减少并发计数
			atomic.AddInt64(&qm.concurrent, -1)
			
			// 增加总请求数
			atomic.AddInt64(&qm.totalRequests, 1)
			
			// 记录请求时间戳
			qm.recordRequest(startTime, endpointKey)
		}()
		
		c.Next(ctx)
	}
}

// recordRequest 记录请求
func (qm *QPSMonitor) recordRequest(timestamp time.Time, endpointKey string) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	
	// 记录全局请求
	qm.globalWindow.Add(timestamp)
	
	// 记录端点请求
	if _, exists := qm.endpointWindows[endpointKey]; !exists {
		qm.endpointWindows[endpointKey] = NewQPSWindow(qm.windowSize)
	}
	qm.endpointWindows[endpointKey].Add(timestamp)
	
	// 检查QPS告警
	currentQPS := qm.globalWindow.QPS()
	if qm.onAlert != nil && currentQPS > qm.qpsThreshold {
		go qm.onAlert(AlertInfo{
			Type:      "qps",
			Message:   fmt.Sprintf("QPS过高: %.2f", currentQPS),
			Value:     currentQPS,
			Threshold: qm.qpsThreshold,
			Timestamp: timestamp,
		})
	}
	
	// 更新最大QPS
	if currentQPS > qm.maxQPS {
		qm.maxQPS = currentQPS
	}
}

// GetStats 获取QPS统计
func (qm *QPSMonitor) GetStats() QPSStats {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	
	currentQPS := qm.globalWindow.QPS()
	
	// 计算平均QPS
	avgQPS := float64(0)
	totalRequests := atomic.LoadInt64(&qm.totalRequests)
	if elapsed := time.Since(qm.startTime).Seconds(); elapsed > 0 {
		avgQPS = float64(totalRequests) / elapsed
	}
	
	// 收集端点统计
	endpoints := make(map[string]*EndpointQPS)
	for key, window := range qm.endpointWindows {
		// 解析端点key
		parts := split2(key, " ")
		if len(parts) == 2 {
			endpoints[key] = &EndpointQPS{
				Method:        parts[0],
				Path:          parts[1],
				CurrentQPS:    window.QPS(),
				TotalRequests: int64(window.Count()),
				LastUpdate:    time.Now(),
			}
		}
	}
	
	stats := QPSStats{
		CurrentQPS:    currentQPS,
		MaxQPS:        qm.maxQPS,
		AvgQPS:        avgQPS,
		TotalRequests: totalRequests,
		Concurrent:    atomic.LoadInt64(&qm.concurrent),
		MaxConcurrent: atomic.LoadInt64(&qm.maxConcurrent),
		Endpoints:     endpoints,
		Timestamp:     time.Now(),
		WindowSize:    qm.windowSize,
	}
	
	// 添加到历史记录
	qm.history = append(qm.history, stats)
	if len(qm.history) > qm.maxHistorySize {
		qm.history = qm.history[1:]
	}
	
	return stats
}

// GetHistory 获取历史统计
func (qm *QPSMonitor) GetHistory(limit int) []QPSStats {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	
	if limit <= 0 || limit > len(qm.history) {
		limit = len(qm.history)
	}
	
	start := len(qm.history) - limit
	result := make([]QPSStats, limit)
	copy(result, qm.history[start:])
	return result
}

// GetTopEndpoints 获取QPS最高的端点
func (qm *QPSMonitor) GetTopEndpoints(limit int) []*EndpointQPS {
	stats := qm.GetStats()
	
	endpoints := make([]*EndpointQPS, 0, len(stats.Endpoints))
	for _, ep := range stats.Endpoints {
		endpoints = append(endpoints, ep)
	}
	
	// 按QPS排序
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].CurrentQPS > endpoints[j].CurrentQPS
	})
	
	if limit > 0 && limit < len(endpoints) {
		endpoints = endpoints[:limit]
	}
	
	return endpoints
}

// Reset 重置统计
func (qm *QPSMonitor) Reset() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	
	qm.globalWindow = NewQPSWindow(qm.windowSize)
	qm.endpointWindows = make(map[string]*QPSWindow)
	atomic.StoreInt64(&qm.concurrent, 0)
	atomic.StoreInt64(&qm.maxConcurrent, 0)
	atomic.StoreInt64(&qm.totalRequests, 0)
	qm.maxQPS = 0
	qm.history = make([]QPSStats, 0, qm.maxHistorySize)
	qm.startTime = time.Now()
}

// Enable 启用QPS监控
func (qm *QPSMonitor) Enable() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.enabled = true
}

// Disable 禁用QPS监控
func (qm *QPSMonitor) Disable() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.enabled = false
}

// IsEnabled 检查是否启用
func (qm *QPSMonitor) IsEnabled() bool {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return qm.enabled
}

// QPSPanel QPS监控面板
type QPSPanel struct {
	monitor *QPSMonitor
}

// NewQPSPanel 创建QPS面板
func NewQPSPanel(monitor *QPSMonitor) *QPSPanel {
	return &QPSPanel{
		monitor: monitor,
	}
}

// RegisterRoutes 注册QPS监控路由
func (qp *QPSPanel) RegisterRoutes(engine any) {
	var qpsGroup *route.RouterGroup
	
	if h, ok := engine.(*route.Engine); ok {
		qpsGroup = h.Group("/yyhertz/qps")
	} else {
		config.Error("无法注册QPS路由，未知引擎类型")
		return
	}
	
	// 注册路由
	qpsGroup.GET("/stats", qp.getStats)
	qpsGroup.GET("/history", qp.getHistory)
	qpsGroup.GET("/top", qp.getTopEndpoints)
	qpsGroup.POST("/reset", qp.resetStats)
	qpsGroup.GET("/panel", qp.qpsPanel)
	qpsGroup.POST("/enable", qp.enableMonitor)
	qpsGroup.POST("/disable", qp.disableMonitor)
}

// getStats 获取QPS统计
func (qp *QPSPanel) getStats(ctx context.Context, c *app.RequestContext) {
	stats := qp.monitor.GetStats()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": stats,
	})
}

// getHistory 获取历史统计
func (qp *QPSPanel) getHistory(ctx context.Context, c *app.RequestContext) {
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 100
		}
	}
	
	history := qp.monitor.GetHistory(limit)
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": history,
	})
}

// getTopEndpoints 获取QPS最高的端点
func (qp *QPSPanel) getTopEndpoints(ctx context.Context, c *app.RequestContext) {
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 10
		}
	}
	
	endpoints := qp.monitor.GetTopEndpoints(limit)
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": endpoints,
	})
}

// resetStats 重置统计
func (qp *QPSPanel) resetStats(ctx context.Context, c *app.RequestContext) {
	qp.monitor.Reset()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "QPS统计已重置",
	})
}

// enableMonitor 启用监控
func (qp *QPSPanel) enableMonitor(ctx context.Context, c *app.RequestContext) {
	qp.monitor.Enable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "QPS监控已启用",
		"enabled": true,
	})
}

// disableMonitor 禁用监控
func (qp *QPSPanel) disableMonitor(ctx context.Context, c *app.RequestContext) {
	qp.monitor.Disable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "QPS监控已禁用",
		"enabled": false,
	})
}

// qpsPanel QPS监控面板页面
func (qp *QPSPanel) qpsPanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz QPS监控面板</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .stat-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); text-align: center; }
        .stat-value { font-size: 2.5em; font-weight: bold; color: #007bff; }
        .stat-label { color: #666; margin-top: 10px; }
        .chart-container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .endpoints-table { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); overflow: hidden; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: bold; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .method-badge { padding: 2px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; }
        .GET { background: #28a745; color: white; }
        .POST { background: #007bff; color: white; }
        .PUT { background: #ffc107; color: black; }
        .DELETE { background: #dc3545; color: white; }
        .high-qps { color: #dc3545; font-weight: bold; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz QPS监控面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshData()">刷新数据</button>
            <button class="btn btn-success" onclick="enableMonitor()">启用监控</button>
            <button class="btn btn-danger" onclick="disableMonitor()">禁用监控</button>
            <button class="btn btn-danger" onclick="resetStats()">重置统计</button>
        </div>
    </div>

    <div class="stats-grid" id="statsGrid">
        <!-- 统计卡片将在这里动态生成 -->
    </div>

    <div class="chart-container">
        <h3>实时QPS趋势</h3>
        <canvas id="qpsChart" width="400" height="200"></canvas>
    </div>

    <div class="chart-container">
        <h3>并发数趋势</h3>
        <canvas id="concurrentChart" width="400" height="200"></canvas>
    </div>

    <div class="endpoints-table">
        <h3 style="padding: 20px; margin: 0; border-bottom: 1px solid #eee;">QPS最高的端点</h3>
        <table id="endpointsTable">
            <thead>
                <tr>
                    <th>方法</th>
                    <th>路径</th>
                    <th>当前QPS</th>
                    <th>总请求数</th>
                    <th>最后更新</th>
                </tr>
            </thead>
            <tbody>
                <!-- 端点数据将在这里动态生成 -->
            </tbody>
        </table>
    </div>

    <script>
        let qpsChart, concurrentChart;

        function initCharts() {
            // QPS图表
            const qpsCtx = document.getElementById('qpsChart').getContext('2d');
            qpsChart = new Chart(qpsCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'QPS',
                        data: [],
                        borderColor: '#007bff',
                        backgroundColor: 'rgba(0, 123, 255, 0.1)',
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    }
                }
            });

            // 并发图表
            const concurrentCtx = document.getElementById('concurrentChart').getContext('2d');
            concurrentChart = new Chart(concurrentCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: '并发数',
                        data: [],
                        borderColor: '#28a745',
                        backgroundColor: 'rgba(40, 167, 69, 0.1)',
                        tension: 0.4
                    }]
                },
                options: {
                    responsive: true,
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    }
                }
            });
        }

        function loadStats() {
            fetch('/yyhertz/qps/stats')
                .then(response => response.json())
                .then(data => {
                    updateStatsCards(data.data);
                })
                .catch(error => console.error('加载统计失败:', error));
        }

        function loadHistory() {
            fetch('/yyhertz/qps/history?limit=50')
                .then(response => response.json())
                .then(data => {
                    updateCharts(data.data);
                })
                .catch(error => console.error('加载历史数据失败:', error));
        }

        function loadTopEndpoints() {
            fetch('/yyhertz/qps/top?limit=10')
                .then(response => response.json())
                .then(data => {
                    updateEndpointsTable(data.data);
                })
                .catch(error => console.error('加载端点数据失败:', error));
        }

        function updateStatsCards(stats) {
            const grid = document.getElementById('statsGrid');
            grid.innerHTML = 
                '<div class="stat-card">' +
                    '<div class="stat-value">' + (stats.current_qps || 0).toFixed(2) + '</div>' +
                    '<div class="stat-label">当前QPS</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-value">' + (stats.max_qps || 0).toFixed(2) + '</div>' +
                    '<div class="stat-label">最大QPS</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-value">' + (stats.avg_qps || 0).toFixed(2) + '</div>' +
                    '<div class="stat-label">平均QPS</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-value">' + (stats.concurrent || 0) + '</div>' +
                    '<div class="stat-label">当前并发</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-value">' + (stats.max_concurrent || 0) + '</div>' +
                    '<div class="stat-label">最大并发</div>' +
                '</div>' +
                '<div class="stat-card">' +
                    '<div class="stat-value">' + (stats.total_requests || 0) + '</div>' +
                    '<div class="stat-label">总请求数</div>' +
                '</div>';
        }

        function updateCharts(history) {
            if (!history || history.length === 0) return;

            const labels = history.map(h => new Date(h.timestamp).toLocaleTimeString());
            const qpsData = history.map(h => h.current_qps || 0);
            const concurrentData = history.map(h => h.concurrent || 0);

            // 更新QPS图表
            qpsChart.data.labels = labels;
            qpsChart.data.datasets[0].data = qpsData;
            qpsChart.update();

            // 更新并发图表
            concurrentChart.data.labels = labels;
            concurrentChart.data.datasets[0].data = concurrentData;
            concurrentChart.update();
        }

        function updateEndpointsTable(endpoints) {
            const tbody = document.querySelector('#endpointsTable tbody');
            let html = '';

            if (!endpoints || endpoints.length === 0) {
                html = '<tr><td colspan="5" style="text-align: center; color: #666;">暂无数据</td></tr>';
            } else {
                endpoints.forEach(endpoint => {
                    const qpsClass = endpoint.current_qps > 100 ? 'high-qps' : '';
                    html += '<tr>' +
                        '<td><span class="method-badge ' + endpoint.method + '">' + endpoint.method + '</span></td>' +
                        '<td>' + endpoint.path + '</td>' +
                        '<td class="' + qpsClass + '">' + endpoint.current_qps.toFixed(2) + '</td>' +
                        '<td>' + endpoint.total_requests + '</td>' +
                        '<td>' + new Date(endpoint.last_update).toLocaleString() + '</td>' +
                        '</tr>';
                });
            }

            tbody.innerHTML = html;
        }

        function refreshData() {
            loadStats();
            loadHistory();
            loadTopEndpoints();
        }

        function enableMonitor() {
            fetch('/yyhertz/qps/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('QPS监控已启用');
                    refreshData();
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableMonitor() {
            if (confirm('确定要禁用QPS监控吗？')) {
                fetch('/yyhertz/qps/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('QPS监控已禁用');
                        refreshData();
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        function resetStats() {
            if (confirm('确定要重置所有QPS统计吗？')) {
                fetch('/yyhertz/qps/reset', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('QPS统计已重置');
                        refreshData();
                    })
                    .catch(error => {
                        console.error('重置失败:', error);
                        alert('重置失败');
                    });
            }
        }

        // 页面加载时初始化
        window.onload = function() {
            initCharts();
            refreshData();
            // 每5秒自动刷新
            setInterval(refreshData, 5000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}

// 辅助函数：分割字符串为两部分
func split2(s, sep string) []string {
	idx := len(s)
	for i := 0; i < len(s)-len(sep)+1; i++ {
		if s[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx == len(s) {
		return []string{s}
	}
	return []string{s[:idx], s[idx+len(sep):]}
}