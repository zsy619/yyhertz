package devtools

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	Timestamp    time.Time `json:"timestamp"`
	RequestCount int64     `json:"request_count"`
	ErrorCount   int64     `json:"error_count"`
	AvgResponse  float64   `json:"avg_response_time"`
	MaxResponse  float64   `json:"max_response_time"`
	MinResponse  float64   `json:"min_response_time"`
	Memory       struct {
		Alloc      uint64 `json:"alloc"`
		TotalAlloc uint64 `json:"total_alloc"`
		Sys        uint64 `json:"sys"`
		NumGC      uint32 `json:"num_gc"`
	} `json:"memory"`
	Goroutines int     `json:"goroutines"`
	CPUUsage   float64 `json:"cpu_usage"`
}

// EndpointMetrics 端点指标
type EndpointMetrics struct {
	Path       string        `json:"path"`
	Method     string        `json:"method"`
	Count      int64         `json:"count"`
	ErrorCount int64         `json:"error_count"`
	TotalTime  time.Duration `json:"total_time"`
	AvgTime    time.Duration `json:"avg_time"`
	MaxTime    time.Duration `json:"max_time"`
	MinTime    time.Duration `json:"min_time"`
	LastAccess time.Time     `json:"last_access"`
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	mu              sync.RWMutex
	enabled         bool
	startTime       time.Time
	requestCount    int64
	errorCount      int64
	totalResponse   time.Duration
	maxResponse     time.Duration
	minResponse     time.Duration
	endpoints       map[string]*EndpointMetrics
	metricsHistory  []PerformanceMetrics
	maxHistorySize  int
	collectInterval time.Duration
	stopCh          chan struct{}
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		enabled:         true,
		startTime:       time.Now(),
		endpoints:       make(map[string]*EndpointMetrics),
		metricsHistory:  make([]PerformanceMetrics, 0),
		maxHistorySize:  1000,
		collectInterval: 10 * time.Second,
		stopCh:          make(chan struct{}),
		minResponse:     time.Hour, // 初始化为一个很大的值
	}
}

// Start 启动性能监控
func (pm *PerformanceMonitor) Start() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.enabled {
		go pm.collectMetrics()
	}
}

// Stop 停止性能监控
func (pm *PerformanceMonitor) Stop() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.enabled {
		close(pm.stopCh)
		pm.enabled = false
	}
}

// Middleware 性能监控中间件
func (pm *PerformanceMonitor) Middleware() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !pm.enabled {
			c.Next(ctx)
			return
		}

		startTime := time.Now()
		path := string(c.Path())
		method := string(c.Method())
		endpointKey := fmt.Sprintf("%s %s", method, path)

		// 执行下一个中间件
		c.Next(ctx)

		duration := time.Since(startTime)
		statusCode := c.Response.StatusCode()
		isError := statusCode >= 400

		// 更新全局统计
		pm.mu.Lock()
		pm.requestCount++
		if isError {
			pm.errorCount++
		}
		pm.totalResponse += duration
		if duration > pm.maxResponse {
			pm.maxResponse = duration
		}
		if duration < pm.minResponse {
			pm.minResponse = duration
		}

		// 更新端点统计
		endpoint, exists := pm.endpoints[endpointKey]
		if !exists {
			endpoint = &EndpointMetrics{
				Path:       path,
				Method:     method,
				MinTime:    duration,
				MaxTime:    duration,
				LastAccess: startTime,
			}
			pm.endpoints[endpointKey] = endpoint
		}

		endpoint.Count++
		if isError {
			endpoint.ErrorCount++
		}
		endpoint.TotalTime += duration
		endpoint.AvgTime = time.Duration(int64(endpoint.TotalTime) / endpoint.Count)
		if duration > endpoint.MaxTime {
			endpoint.MaxTime = duration
		}
		if duration < endpoint.MinTime {
			endpoint.MinTime = duration
		}
		endpoint.LastAccess = startTime

		pm.mu.Unlock()
	}
}

// collectMetrics 收集性能指标
func (pm *PerformanceMonitor) collectMetrics() {
	ticker := time.NewTicker(pm.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.recordMetrics()
		case <-pm.stopCh:
			return
		}
	}
}

// recordMetrics 记录性能指标
func (pm *PerformanceMonitor) recordMetrics() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	avgResponse := float64(0)
	if pm.requestCount > 0 {
		avgResponse = float64(pm.totalResponse.Nanoseconds()) / float64(pm.requestCount) / 1e6 // 转换为毫秒
	}

	metrics := PerformanceMetrics{
		Timestamp:    time.Now(),
		RequestCount: pm.requestCount,
		ErrorCount:   pm.errorCount,
		AvgResponse:  avgResponse,
		MaxResponse:  float64(pm.maxResponse.Nanoseconds()) / 1e6,
		MinResponse:  float64(pm.minResponse.Nanoseconds()) / 1e6,
		Goroutines:   runtime.NumGoroutine(),
		CPUUsage:     pm.getCPUUsage(),
	}

	metrics.Memory.Alloc = m.Alloc
	metrics.Memory.TotalAlloc = m.TotalAlloc
	metrics.Memory.Sys = m.Sys
	metrics.Memory.NumGC = m.NumGC

	// 添加到历史记录
	pm.metricsHistory = append(pm.metricsHistory, metrics)

	// 限制历史记录大小
	if len(pm.metricsHistory) > pm.maxHistorySize {
		pm.metricsHistory = pm.metricsHistory[1:]
	}
}

// getCPUUsage 获取CPU使用率（简化版本）
func (pm *PerformanceMonitor) getCPUUsage() float64 {
	// 这里是一个简化的CPU使用率计算
	// 实际项目中可能需要更复杂的实现
	return float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10
}

// GetMetrics 获取当前性能指标
func (pm *PerformanceMonitor) GetMetrics() PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.metricsHistory) == 0 {
		pm.recordMetrics()
	}

	if len(pm.metricsHistory) > 0 {
		return pm.metricsHistory[len(pm.metricsHistory)-1]
	}

	return PerformanceMetrics{}
}

// GetHistory 获取历史指标
func (pm *PerformanceMonitor) GetHistory(limit int) []PerformanceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if limit <= 0 || limit > len(pm.metricsHistory) {
		limit = len(pm.metricsHistory)
	}

	start := len(pm.metricsHistory) - limit
	result := make([]PerformanceMetrics, limit)
	copy(result, pm.metricsHistory[start:])
	return result
}

// GetEndpoints 获取端点统计
func (pm *PerformanceMonitor) GetEndpoints() map[string]*EndpointMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*EndpointMetrics)
	for k, v := range pm.endpoints {
		// 创建副本
		endpoint := *v
		result[k] = &endpoint
	}
	return result
}

// Reset 重置统计
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.startTime = time.Now()
	pm.requestCount = 0
	pm.errorCount = 0
	pm.totalResponse = 0
	pm.maxResponse = 0
	pm.minResponse = time.Hour
	pm.endpoints = make(map[string]*EndpointMetrics)
	pm.metricsHistory = make([]PerformanceMetrics, 0)
}

// PerformancePanel 性能监控面板
type PerformancePanel struct {
	monitor *PerformanceMonitor
}

// NewPerformancePanel 创建性能面板
func NewPerformancePanel(monitor *PerformanceMonitor) *PerformancePanel {
	return &PerformancePanel{
		monitor: monitor,
	}
}

// RegisterRoutes 注册性能监控路由
func (pp *PerformancePanel) RegisterRoutes(engine any) {
	// 类型断言，支持不同的引擎类型
	var perfGroup *route.RouterGroup

	// 尝试不同的类型断言
	if h, ok := engine.(*route.Engine); ok {
		perfGroup = h.Group("/performance")
	} else {
		log.Println("无法注册性能路由，未知引擎类型")
		return // 如果类型不匹配，直接返回
	}

	// 注册路由的通用方法
	registerRoute := func(method, path string, handler func(ctx context.Context, c *app.RequestContext)) {
		switch method {
		case "GET":
			perfGroup.GET(path, handler)
		case "POST":
			perfGroup.POST(path, handler)
		case "PUT":
			perfGroup.PUT(path, handler)
		case "DELETE":
			perfGroup.DELETE(path, handler)
		case "PATCH":
			perfGroup.PATCH(path, handler)
		case "HEAD":
			perfGroup.HEAD(path, handler)
		case "OPTIONS":
			perfGroup.OPTIONS(path, handler)
		default:
			log.Printf("不支持的HTTP方法: %s", method)
		}
	}

	// 获取当前性能指标
	registerRoute("GET", "/metrics", pp.getMetrics)

	// 获取历史指标
	registerRoute("GET", "/history", pp.getHistory)

	// 获取端点统计
	registerRoute("GET", "/endpoints", pp.getEndpoints)

	// 重置统计
	registerRoute("POST", "/reset", pp.resetMetrics)

	// 性能监控面板页面
	registerRoute("GET", "/panel", pp.performancePanel)
}

// getMetrics 获取性能指标
func (pp *PerformancePanel) getMetrics(ctx context.Context, c *app.RequestContext) {
	metrics := pp.monitor.GetMetrics()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": metrics,
	})
}

// getHistory 获取历史指标
func (pp *PerformancePanel) getHistory(ctx context.Context, c *app.RequestContext) {
	limit := 100 // 默认返回最近100条记录
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 100
		}
	}

	history := pp.monitor.GetHistory(limit)
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": history,
	})
}

// getEndpoints 获取端点统计
func (pp *PerformancePanel) getEndpoints(ctx context.Context, c *app.RequestContext) {
	endpoints := pp.monitor.GetEndpoints()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": endpoints,
	})
}

// resetMetrics 重置指标
func (pp *PerformancePanel) resetMetrics(ctx context.Context, c *app.RequestContext) {
	pp.monitor.Reset()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "性能指标已重置",
	})
}

// performancePanel 性能监控面板页面
func (pp *PerformancePanel) performancePanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 性能监控面板</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 20px; margin-bottom: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-value { font-size: 2em; font-weight: bold; color: #007bff; }
        .metric-label { color: #666; margin-top: 5px; }
        .chart-container { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .endpoints-table { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); overflow: hidden; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: bold; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .method-badge { padding: 2px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; }
        .GET { background: #28a745; color: white; }
        .POST { background: #007bff; color: white; }
        .PUT { background: #ffc107; color: black; }
        .DELETE { background: #dc3545; color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 性能监控面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshData()">刷新数据</button>
            <button class="btn btn-danger" onclick="resetMetrics()">重置统计</button>
        </div>
    </div>

    <div class="metrics-grid" id="metricsGrid">
        <!-- 指标卡片将在这里动态生成 -->
    </div>

    <div class="chart-container">
        <h3>响应时间趋势</h3>
        <canvas id="responseTimeChart" width="400" height="200"></canvas>
    </div>

    <div class="chart-container">
        <h3>内存使用趋势</h3>
        <canvas id="memoryChart" width="400" height="200"></canvas>
    </div>

    <div class="endpoints-table">
        <h3 style="padding: 20px; margin: 0; border-bottom: 1px solid #eee;">端点统计</h3>
        <table id="endpointsTable">
            <thead>
                <tr>
                    <th>方法</th>
                    <th>路径</th>
                    <th>请求数</th>
                    <th>错误数</th>
                    <th>平均响应时间</th>
                    <th>最大响应时间</th>
                    <th>最后访问</th>
                </tr>
            </thead>
            <tbody>
                <!-- 端点数据将在这里动态生成 -->
            </tbody>
        </table>
    </div>

    <script>
        let responseTimeChart, memoryChart;

        function initCharts() {
            // 响应时间图表
            const responseCtx = document.getElementById('responseTimeChart').getContext('2d');
            responseTimeChart = new Chart(responseCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: '平均响应时间 (ms)',
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

            // 内存使用图表
            const memoryCtx = document.getElementById('memoryChart').getContext('2d');
            memoryChart = new Chart(memoryCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: '内存使用 (MB)',
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

        function loadMetrics() {
            fetch('/performance/metrics')
                .then(response => response.json())
                .then(data => {
                    updateMetricsCards(data.data);
                })
                .catch(error => console.error('加载指标失败:', error));
        }

        function loadHistory() {
            fetch('/performance/history?limit=50')
                .then(response => response.json())
                .then(data => {
                    updateCharts(data.data);
                })
                .catch(error => console.error('加载历史数据失败:', error));
        }

        function loadEndpoints() {
            fetch('/performance/endpoints')
                .then(response => response.json())
                .then(data => {
                    updateEndpointsTable(data.data);
                })
                .catch(error => console.error('加载端点数据失败:', error));
        }

        function updateMetricsCards(metrics) {
            const grid = document.getElementById('metricsGrid');
            grid.innerHTML = 
                '<div class="metric-card">' +
                    '<div class="metric-value">' + (metrics.request_count || 0) + '</div>' +
                    '<div class="metric-label">总请求数</div>' +
                '</div>' +
                '<div class="metric-card">' +
                    '<div class="metric-value">' + (metrics.error_count || 0) + '</div>' +
                    '<div class="metric-label">错误数</div>' +
                '</div>' +
                '<div class="metric-card">' +
                    '<div class="metric-value">' + (metrics.avg_response_time || 0).toFixed(2) + 'ms</div>' +
                    '<div class="metric-label">平均响应时间</div>' +
                '</div>' +
                '<div class="metric-card">' +
                    '<div class="metric-value">' + Math.round((metrics.memory && metrics.memory.alloc || 0) / 1024 / 1024) + 'MB</div>' +
                    '<div class="metric-label">内存使用</div>' +
                '</div>' +
                '<div class="metric-card">' +
                    '<div class="metric-value">' + (metrics.goroutines || 0) + '</div>' +
                    '<div class="metric-label">协程数</div>' +
                '</div>' +
                '<div class="metric-card">' +
                    '<div class="metric-value">' + (metrics.cpu_usage || 0).toFixed(1) + '%</div>' +
                    '<div class="metric-label">CPU使用率</div>' +
                '</div>';
        }

        function updateCharts(history) {
            if (!history || history.length === 0) return;

            const labels = history.map(h => new Date(h.timestamp).toLocaleTimeString());
            const responseTimes = history.map(h => h.avg_response_time || 0);
            const memoryUsage = history.map(h => Math.round((h.memory && h.memory.alloc || 0) / 1024 / 1024));

            // 更新响应时间图表
            responseTimeChart.data.labels = labels;
            responseTimeChart.data.datasets[0].data = responseTimes;
            responseTimeChart.update();

            // 更新内存使用图表
            memoryChart.data.labels = labels;
            memoryChart.data.datasets[0].data = memoryUsage;
            memoryChart.update();
        }

        function updateEndpointsTable(endpoints) {
            const tbody = document.querySelector('#endpointsTable tbody');
            let html = '';

            Object.values(endpoints).forEach(endpoint => {
                html += '<tr>' +
                    '<td><span class="method-badge ' + endpoint.method + '">' + endpoint.method + '</span></td>' +
                    '<td>' + endpoint.path + '</td>' +
                    '<td>' + endpoint.count + '</td>' +
                    '<td>' + endpoint.error_count + '</td>' +
                    '<td>' + (endpoint.avg_time / 1000000).toFixed(2) + 'ms</td>' +
                    '<td>' + (endpoint.max_time / 1000000).toFixed(2) + 'ms</td>' +
                    '<td>' + new Date(endpoint.last_access).toLocaleString() + '</td>' +
                    '</tr>';
            });

            tbody.innerHTML = html || '<tr><td colspan="7" style="text-align: center; color: #666;">暂无数据</td></tr>';
        }

        function refreshData() {
            loadMetrics();
            loadHistory();
            loadEndpoints();
        }

        function resetMetrics() {
            if (confirm('确定要重置所有性能统计吗？')) {
                fetch('/performance/reset', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('性能统计已重置');
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
            // 每10秒自动刷新
            setInterval(refreshData, 10000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}
