// Package devtools 提供Metrics指标收集中间件
//
// Metrics收集中间件用于收集和暴露应用性能指标，提供：
// - Prometheus格式指标输出
// - 自定义业务指标收集
// - 系统级指标监控
// - 多维度标签支持
// - 实时指标聚合
// - 指标导出接口
//
// 功能特性：
// - 兼容Prometheus格式
// - 高性能计数器
// - 内存友好的指标存储
// - 支持多种指标类型
// - 自动垃圾回收
package devtools

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// MetricType 指标类型
type MetricType string

const (
	// MetricTypeCounter 计数器类型
	MetricTypeCounter MetricType = "counter"
	// MetricTypeGauge 瞬时值类型
	MetricTypeGauge MetricType = "gauge"
	// MetricTypeHistogram 直方图类型
	MetricTypeHistogram MetricType = "histogram"
	// MetricTypeSummary 摘要类型
	MetricTypeSummary MetricType = "summary"
)

// Metric 指标接口
type Metric interface {
	Name() string
	Help() string
	Type() MetricType
	Labels() map[string]string
	Value() float64
	Collect() []*MetricSample
}

// MetricSample 指标样本
type MetricSample struct {
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
}

// Counter 计数器指标
type Counter struct {
	name   string
	help   string
	labels map[string]string
	value  int64
}

// NewCounter 创建计数器
func NewCounter(name, help string, labels map[string]string) *Counter {
	if labels == nil {
		labels = make(map[string]string)
	}
	return &Counter{
		name:   name,
		help:   help,
		labels: labels,
	}
}

func (c *Counter) Name() string                 { return c.name }
func (c *Counter) Help() string                 { return c.help }
func (c *Counter) Type() MetricType             { return MetricTypeCounter }
func (c *Counter) Labels() map[string]string    { return c.labels }
func (c *Counter) Value() float64               { return float64(atomic.LoadInt64(&c.value)) }

// Inc 增加计数
func (c *Counter) Inc() {
	atomic.AddInt64(&c.value, 1)
}

// Add 增加指定值
func (c *Counter) Add(delta float64) {
	if delta < 0 {
		return // 计数器不能减少
	}
	atomic.AddInt64(&c.value, int64(delta))
}

func (c *Counter) Collect() []*MetricSample {
	return []*MetricSample{
		{
			Name:      c.name,
			Labels:    c.labels,
			Value:     c.Value(),
			Timestamp: time.Now(),
		},
	}
}

// Gauge 瞬时值指标
type Gauge struct {
	name   string
	help   string
	labels map[string]string
	value  int64 // 使用int64存储float64*1000000来避免并发问题
}

// NewGauge 创建瞬时值指标
func NewGauge(name, help string, labels map[string]string) *Gauge {
	if labels == nil {
		labels = make(map[string]string)
	}
	return &Gauge{
		name:   name,
		help:   help,
		labels: labels,
	}
}

func (g *Gauge) Name() string                 { return g.name }
func (g *Gauge) Help() string                 { return g.help }
func (g *Gauge) Type() MetricType             { return MetricTypeGauge }
func (g *Gauge) Labels() map[string]string    { return g.labels }
func (g *Gauge) Value() float64               { return float64(atomic.LoadInt64(&g.value)) / 1000000 }

// Set 设置值
func (g *Gauge) Set(value float64) {
	atomic.StoreInt64(&g.value, int64(value*1000000))
}

// Inc 增加1
func (g *Gauge) Inc() {
	atomic.AddInt64(&g.value, 1000000)
}

// Dec 减少1
func (g *Gauge) Dec() {
	atomic.AddInt64(&g.value, -1000000)
}

// Add 增加指定值
func (g *Gauge) Add(delta float64) {
	atomic.AddInt64(&g.value, int64(delta*1000000))
}

func (g *Gauge) Collect() []*MetricSample {
	return []*MetricSample{
		{
			Name:      g.name,
			Labels:    g.labels,
			Value:     g.Value(),
			Timestamp: time.Now(),
		},
	}
}

// Histogram 直方图指标
type Histogram struct {
	name    string
	help    string
	labels  map[string]string
	buckets []float64
	counts  []int64
	sum     int64
	count   int64
	mu      sync.RWMutex
}

// NewHistogram 创建直方图
func NewHistogram(name, help string, labels map[string]string, buckets []float64) *Histogram {
	if labels == nil {
		labels = make(map[string]string)
	}
	if buckets == nil {
		// 默认桶
		buckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	}
	
	// 确保桶是排序的
	sort.Float64s(buckets)
	
	return &Histogram{
		name:    name,
		help:    help,
		labels:  labels,
		buckets: buckets,
		counts:  make([]int64, len(buckets)+1), // +1 for +Inf bucket
	}
}

func (h *Histogram) Name() string                 { return h.name }
func (h *Histogram) Help() string                 { return h.help }
func (h *Histogram) Type() MetricType             { return MetricTypeHistogram }
func (h *Histogram) Labels() map[string]string    { return h.labels }
func (h *Histogram) Value() float64               { return float64(atomic.LoadInt64(&h.count)) }

// Observe 观察一个值
func (h *Histogram) Observe(value float64) {
	atomic.AddInt64(&h.count, 1)
	atomic.AddInt64(&h.sum, int64(value*1000000))
	
	// 找到合适的桶
	for i, bucket := range h.buckets {
		if value <= bucket {
			atomic.AddInt64(&h.counts[i], 1)
			return
		}
	}
	// +Inf桶
	atomic.AddInt64(&h.counts[len(h.buckets)], 1)
}

func (h *Histogram) Collect() []*MetricSample {
	samples := make([]*MetricSample, 0, len(h.buckets)+3)
	now := time.Now()
	
	// 累积计数
	cumulative := int64(0)
	for i, bucket := range h.buckets {
		cumulative += atomic.LoadInt64(&h.counts[i])
		labels := make(map[string]string)
		for k, v := range h.labels {
			labels[k] = v
		}
		labels["le"] = fmt.Sprintf("%g", bucket)
		
		samples = append(samples, &MetricSample{
			Name:      h.name + "_bucket",
			Labels:    labels,
			Value:     float64(cumulative),
			Timestamp: now,
		})
	}
	
	// +Inf桶
	cumulative += atomic.LoadInt64(&h.counts[len(h.buckets)])
	infLabels := make(map[string]string)
	for k, v := range h.labels {
		infLabels[k] = v
	}
	infLabels["le"] = "+Inf"
	samples = append(samples, &MetricSample{
		Name:      h.name + "_bucket",
		Labels:    infLabels,
		Value:     float64(cumulative),
		Timestamp: now,
	})
	
	// 总数
	samples = append(samples, &MetricSample{
		Name:      h.name + "_count",
		Labels:    h.labels,
		Value:     float64(atomic.LoadInt64(&h.count)),
		Timestamp: now,
	})
	
	// 总和
	samples = append(samples, &MetricSample{
		Name:      h.name + "_sum",
		Labels:    h.labels,
		Value:     float64(atomic.LoadInt64(&h.sum)) / 1000000,
		Timestamp: now,
	})
	
	return samples
}

// EndpointMetrics 端点指标(从performance_monitor迁移)
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

// GCMetrics GC相关指标(从runtime_metrics迁移)
type GCMetrics struct {
	NumGC          uint32        `json:"num_gc"`           // GC次数
	PauseTotal     time.Duration `json:"pause_total"`      // 总暂停时间
	PauseNs        []uint64      `json:"pause_ns"`         // 暂停时间数组
	PauseEnd       []uint64      `json:"pause_end"`        // 暂停结束时间
	LastGC         time.Time     `json:"last_gc"`          // 最后一次GC时间
	NextGC         uint64        `json:"next_gc"`          // 下次GC阈值
	TargetHeap     uint64        `json:"target_heap"`      // 目标堆大小
	CPUFraction    float64       `json:"cpu_fraction"`     // GC CPU占用率
	EnablePercent  bool          `json:"enable_percent"`   // 是否启用GC
	DebugGC        bool          `json:"debug_gc"`         // 是否调试GC
	
	// 扩展统计
	AvgPauseTime   time.Duration `json:"avg_pause_time"`   // 平均暂停时间
	MaxPauseTime   time.Duration `json:"max_pause_time"`   // 最大暂停时间
	MinPauseTime   time.Duration `json:"min_pause_time"`   // 最小暂停时间
	FrequencyPerSec float64      `json:"frequency_per_sec"` // 每秒GC频率
	TrendDirection  string       `json:"trend_direction"`  // 趋势方向
}

// RuntimeMemoryMetrics 详细内存相关指标(从runtime_metrics迁移)
type RuntimeMemoryMetrics struct {
	Alloc         uint64 `json:"alloc"`           // 当前分配的堆内存
	TotalAlloc    uint64 `json:"total_alloc"`     // 累计分配的堆内存
	Sys           uint64 `json:"sys"`             // 系统内存占用
	Lookups       uint64 `json:"lookups"`         // 指针查找次数
	Mallocs       uint64 `json:"mallocs"`         // 内存分配次数
	Frees         uint64 `json:"frees"`           // 内存释放次数
	
	// 堆内存详细信息
	HeapAlloc    uint64 `json:"heap_alloc"`     // 堆分配内存
	HeapSys      uint64 `json:"heap_sys"`       // 堆系统内存
	HeapIdle     uint64 `json:"heap_idle"`      // 堆空闲内存
	HeapInuse    uint64 `json:"heap_inuse"`     // 堆使用内存
	HeapReleased uint64 `json:"heap_released"`  // 堆释放内存
	HeapObjects  uint64 `json:"heap_objects"`   // 堆对象数量
	
	// 栈内存信息
	StackInuse uint64 `json:"stack_inuse"`     // 栈使用内存
	StackSys   uint64 `json:"stack_sys"`       // 栈系统内存
	
	// 其他系统内存
	MSpanInuse  uint64 `json:"mspan_inuse"`    // MSpan使用内存
	MSpanSys    uint64 `json:"mspan_sys"`      // MSpan系统内存
	MCacheInuse uint64 `json:"mcache_inuse"`   // MCache使用内存
	MCacheSys   uint64 `json:"mcache_sys"`     // MCache系统内存
	BuckHashSys uint64 `json:"buck_hash_sys"`  // 分析哈希表内存
	GCSys       uint64 `json:"gc_sys"`         // GC元数据内存
	OtherSys    uint64 `json:"other_sys"`      // 其他系统内存
	
	// 计算字段
	AllocRate     float64 `json:"alloc_rate"`     // 分配速率 (bytes/sec)
	GCOverhead    float64 `json:"gc_overhead"`    // GC开销百分比
	MemoryEfficiency float64 `json:"memory_efficiency"` // 内存效率
}

// RuntimeMetrics 运行时指标(从runtime_metrics迁移)
type RuntimeMetrics struct {
	Timestamp    time.Time     `json:"timestamp"`     // 时间戳
	NumGoroutine int           `json:"num_goroutine"` // 协程数量
	NumCPU       int           `json:"num_cpu"`       // CPU核心数
	NumCgoCall   int64         `json:"num_cgo_call"`  // CGO调用次数
	
	GCMetrics     *GCMetrics     `json:"gc_metrics"`     // GC指标
	MemoryMetrics *RuntimeMemoryMetrics `json:"memory_metrics"` // 内存指标
	
	// 调度器信息
	GOMAXPROCS    int     `json:"gomaxprocs"`     // 最大P数量
	NumP          int     `json:"num_p"`          // 当前P数量
	
	// 运行时版本信息
	GoVersion     string  `json:"go_version"`     // Go版本
	Compiler      string  `json:"compiler"`       // 编译器
	Architecture  string  `json:"architecture"`   // 架构
	
	// 健康评估
	HealthScore   int     `json:"health_score"`   // 健康评分 (0-100)
	HealthStatus  string  `json:"health_status"`  // 健康状态
	Recommendations []string `json:"recommendations"` // 优化建议
}

// PerformanceMetrics 性能指标(从performance_monitor迁移)
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

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu               sync.RWMutex
	metrics          map[string]Metric
	enabled          bool
	httpRequests     *Counter
	httpDuration     *Histogram
	httpErrors       *Counter
	goRoutines       *Gauge
	memoryUsage      *Gauge
	cpuUsage         *Gauge
	startTime        time.Time
	uptime           *Gauge
	
	// 端点监控相关字段(从performance_monitor迁移)
	requestCount     int64
	errorCount       int64
	totalResponse    time.Duration
	maxResponse      time.Duration
	minResponse      time.Duration
	endpoints        map[string]*EndpointMetrics
	metricsHistory   []PerformanceMetrics
	maxHistorySize   int
	collectInterval  time.Duration
	stopCh           chan struct{}
	
	// 运行时监控相关字段(从runtime_metrics迁移)
	runtimeHistory   []RuntimeMetrics
	gcTrendData      []float64
	memTrendData     []float64
	lastCollectTime  time.Time
}

// MetricsConfig 指标配置
type MetricsConfig struct {
	Enabled       bool   // 是否启用
	Namespace     string // 指标命名空间
	Subsystem     string // 指标子系统
	IncludeSystem bool   // 是否包含系统指标
}

// NewMetricsCollector 创建指标收集器
func NewMetricsCollector(config *MetricsConfig) *MetricsCollector {
	if config == nil {
		config = &MetricsConfig{
			Enabled:       true,
			Namespace:     "yyhertz",
			Subsystem:     "http",
			IncludeSystem: true,
		}
	}
	
	mc := &MetricsCollector{
		metrics:         make(map[string]Metric),
		enabled:         config.Enabled,
		startTime:       time.Now(),
		endpoints:       make(map[string]*EndpointMetrics),
		metricsHistory:  make([]PerformanceMetrics, 0),
		maxHistorySize:  1000,
		collectInterval: 10 * time.Second,
		stopCh:          make(chan struct{}),
		minResponse:     time.Hour, // 初始化为一个很大的值
		runtimeHistory:  make([]RuntimeMetrics, 0),
		gcTrendData:     make([]float64, 0, 100),
		memTrendData:    make([]float64, 0, 100),
	}
	
	prefix := ""
	if config.Namespace != "" {
		prefix = config.Namespace + "_"
	}
	if config.Subsystem != "" {
		prefix += config.Subsystem + "_"
	}
	
	// 注册默认的HTTP指标
	mc.httpRequests = NewCounter(
		prefix+"requests_total",
		"Total number of HTTP requests",
		map[string]string{},
	)
	mc.RegisterMetric(mc.httpRequests)
	
	mc.httpDuration = NewHistogram(
		prefix+"request_duration_seconds",
		"HTTP request duration in seconds",
		map[string]string{},
		[]float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30},
	)
	mc.RegisterMetric(mc.httpDuration)
	
	mc.httpErrors = NewCounter(
		prefix+"errors_total",
		"Total number of HTTP errors",
		map[string]string{},
	)
	mc.RegisterMetric(mc.httpErrors)
	
	// 注册系统指标
	if config.IncludeSystem {
		mc.goRoutines = NewGauge(
			prefix+"goroutines",
			"Number of goroutines",
			map[string]string{},
		)
		mc.RegisterMetric(mc.goRoutines)
		
		mc.memoryUsage = NewGauge(
			prefix+"memory_usage_bytes",
			"Memory usage in bytes",
			map[string]string{},
		)
		mc.RegisterMetric(mc.memoryUsage)
		
		mc.uptime = NewGauge(
			prefix+"uptime_seconds",
			"Uptime in seconds",
			map[string]string{},
		)
		mc.RegisterMetric(mc.uptime)
		
		// 启动系统指标收集
		go mc.collectSystemMetrics()
	}
	
	return mc
}

// RegisterMetric 注册指标
func (mc *MetricsCollector) RegisterMetric(metric Metric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metrics[metric.Name()] = metric
}

// UnregisterMetric 注销指标
func (mc *MetricsCollector) UnregisterMetric(name string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	delete(mc.metrics, name)
}

// GetMetric 获取指标
func (mc *MetricsCollector) GetMetric(name string) Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.metrics[name]
}

// Handler 指标收集中间件(集成端点监控功能)
func (mc *MetricsCollector) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !mc.enabled {
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
		durationSeconds := duration.Seconds()
		statusCode := c.Response.StatusCode()
		isError := statusCode >= 400
		
		// 创建带标签的指标(原有功能)
		labels := map[string]string{
			"method": method,
			"path":   path,
			"status": strconv.Itoa(statusCode),
		}
		
		// 记录请求(原有功能)
		requestCounter := mc.getOrCreateCounter("http_requests_total", "Total HTTP requests", labels)
		requestCounter.Inc()
		
		// 记录响应时间(原有功能)
		durationHisto := mc.getOrCreateHistogram("http_request_duration_seconds", "HTTP request duration", labels)
		durationHisto.Observe(durationSeconds)
		
		// 记录错误(原有功能)
		if isError {
			errorCounter := mc.getOrCreateCounter("http_errors_total", "Total HTTP errors", labels)
			errorCounter.Inc()
		}
		
		// 端点监控功能(从performance_monitor迁移)
		mc.mu.Lock()
		mc.requestCount++
		if isError {
			mc.errorCount++
		}
		mc.totalResponse += duration
		if duration > mc.maxResponse {
			mc.maxResponse = duration
		}
		if duration < mc.minResponse {
			mc.minResponse = duration
		}

		// 更新端点统计
		endpoint, exists := mc.endpoints[endpointKey]
		if !exists {
			endpoint = &EndpointMetrics{
				Path:       path,
				Method:     method,
				MinTime:    duration,
				MaxTime:    duration,
				LastAccess: startTime,
			}
			mc.endpoints[endpointKey] = endpoint
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

		mc.mu.Unlock()
	}
}

// getOrCreateCounter 获取或创建计数器
func (mc *MetricsCollector) getOrCreateCounter(name, help string, labels map[string]string) *Counter {
	key := mc.buildMetricKey(name, labels)
	mc.mu.RLock()
	if metric, exists := mc.metrics[key]; exists {
		mc.mu.RUnlock()
		if counter, ok := metric.(*Counter); ok {
			return counter
		}
	}
	mc.mu.RUnlock()
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// 双重检查
	if metric, exists := mc.metrics[key]; exists {
		if counter, ok := metric.(*Counter); ok {
			return counter
		}
	}
	
	counter := NewCounter(name, help, labels)
	mc.metrics[key] = counter
	return counter
}

// getOrCreateHistogram 获取或创建直方图
func (mc *MetricsCollector) getOrCreateHistogram(name, help string, labels map[string]string) *Histogram {
	key := mc.buildMetricKey(name, labels)
	mc.mu.RLock()
	if metric, exists := mc.metrics[key]; exists {
		mc.mu.RUnlock()
		if histogram, ok := metric.(*Histogram); ok {
			return histogram
		}
	}
	mc.mu.RUnlock()
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// 双重检查
	if metric, exists := mc.metrics[key]; exists {
		if histogram, ok := metric.(*Histogram); ok {
			return histogram
		}
	}
	
	histogram := NewHistogram(name, help, labels, nil)
	mc.metrics[key] = histogram
	return histogram
}

// buildMetricKey 构建指标键
func (mc *MetricsCollector) buildMetricKey(name string, labels map[string]string) string {
	key := name
	if len(labels) > 0 {
		var labelPairs []string
		for k, v := range labels {
			labelPairs = append(labelPairs, k+"="+v)
		}
		sort.Strings(labelPairs)
		key += "{" + strings.Join(labelPairs, ",") + "}"
	}
	return key
}

// collectSystemMetrics 收集系统指标
func (mc *MetricsCollector) collectSystemMetrics() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	// CPU使用率计算相关变量
	var lastCPUTime time.Duration
	var lastSampleTime time.Time = time.Now()
	
	for range ticker.C {
		if !mc.enabled {
			continue
		}
		
		// 更新协程数
		if mc.goRoutines != nil {
			mc.goRoutines.Set(float64(runtime.NumGoroutine()))
		}
		
		// 更新详细内存指标
		if mc.memoryUsage != nil {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			
			// 基础内存使用
			mc.memoryUsage.Set(float64(m.Alloc))
			
			// 创建或获取详细内存指标
			mc.getOrCreateGauge("system_memory_heap_alloc_bytes", "Heap allocated bytes", nil).Set(float64(m.Alloc))
			mc.getOrCreateGauge("system_memory_heap_sys_bytes", "Heap system bytes", nil).Set(float64(m.HeapSys))
			mc.getOrCreateGauge("system_memory_heap_idle_bytes", "Heap idle bytes", nil).Set(float64(m.HeapIdle))
			mc.getOrCreateGauge("system_memory_heap_inuse_bytes", "Heap in-use bytes", nil).Set(float64(m.HeapInuse))
			mc.getOrCreateGauge("system_memory_heap_released_bytes", "Heap released bytes", nil).Set(float64(m.HeapReleased))
			mc.getOrCreateGauge("system_memory_heap_objects", "Number of heap objects", nil).Set(float64(m.HeapObjects))
			mc.getOrCreateGauge("system_memory_stack_inuse_bytes", "Stack in-use bytes", nil).Set(float64(m.StackInuse))
			mc.getOrCreateGauge("system_memory_stack_sys_bytes", "Stack system bytes", nil).Set(float64(m.StackSys))
			mc.getOrCreateGauge("system_memory_mspan_inuse_bytes", "MSpan in-use bytes", nil).Set(float64(m.MSpanInuse))
			mc.getOrCreateGauge("system_memory_mspan_sys_bytes", "MSpan system bytes", nil).Set(float64(m.MSpanSys))
			mc.getOrCreateGauge("system_memory_mcache_inuse_bytes", "MCache in-use bytes", nil).Set(float64(m.MCacheInuse))
			mc.getOrCreateGauge("system_memory_mcache_sys_bytes", "MCache system bytes", nil).Set(float64(m.MCacheSys))
			mc.getOrCreateGauge("system_memory_buck_hash_sys_bytes", "Profiling bucket hash table bytes", nil).Set(float64(m.BuckHashSys))
			mc.getOrCreateGauge("system_memory_gc_sys_bytes", "GC metadata bytes", nil).Set(float64(m.GCSys))
			mc.getOrCreateGauge("system_memory_other_sys_bytes", "Other system bytes", nil).Set(float64(m.OtherSys))
			mc.getOrCreateGauge("system_memory_total_alloc_bytes", "Total allocated bytes", nil).Set(float64(m.TotalAlloc))
			mc.getOrCreateGauge("system_memory_sys_bytes", "Total system bytes", nil).Set(float64(m.Sys))
			mc.getOrCreateGauge("system_memory_mallocs_total", "Total number of mallocs", nil).Set(float64(m.Mallocs))
			mc.getOrCreateGauge("system_memory_frees_total", "Total number of frees", nil).Set(float64(m.Frees))
			mc.getOrCreateGauge("system_gc_num", "Number of completed GC cycles", nil).Set(float64(m.NumGC))
			mc.getOrCreateGauge("system_gc_cpu_fraction", "Fraction of CPU time used by GC", nil).Set(m.GCCPUFraction)
		}
		
		// CPU使用率监控（基于协程时间）
		currentTime := time.Now()
		currentCPUTime := mc.getCPUTime()
		
		if !lastSampleTime.IsZero() {
			timeDiff := currentTime.Sub(lastSampleTime)
			cpuDiff := currentCPUTime - lastCPUTime
			
			if timeDiff > 0 {
				cpuUsage := float64(cpuDiff) / float64(timeDiff) * 100
				mc.getOrCreateGauge("system_cpu_usage_percent", "CPU usage percentage", nil).Set(cpuUsage)
			}
		}
		
		lastCPUTime = currentCPUTime
		lastSampleTime = currentTime
		
		// CPU核心信息
		numCPU := runtime.NumCPU()
		mc.getOrCreateGauge("system_cpu_cores", "Number of CPU cores", nil).Set(float64(numCPU))
		mc.getOrCreateGauge("system_goroutines_per_cpu", "Number of goroutines per CPU core", nil).Set(float64(runtime.NumGoroutine()) / float64(numCPU))
		
		// 更新运行时间
		if mc.uptime != nil {
			mc.uptime.Set(time.Since(mc.startTime).Seconds())
		}
	}
}

// getCPUTime 获取CPU时间（简化实现）
func (mc *MetricsCollector) getCPUTime() time.Duration {
	// 这里使用简化的CPU时间计算
	// 实际生产环境中可能需要更精确的实现
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return time.Duration(m.PauseTotalNs) * time.Nanosecond
}

// getOrCreateGauge 获取或创建Gauge指标
func (mc *MetricsCollector) getOrCreateGauge(name, help string, labels map[string]string) *Gauge {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	key := mc.buildMetricKey(name, labels)
	if metric, exists := mc.metrics[key]; exists {
		if gauge, ok := metric.(*Gauge); ok {
			return gauge
		}
	}
	
	gauge := NewGauge(name, help, labels)
	mc.metrics[key] = gauge
	return gauge
}

// Collect 收集所有指标
func (mc *MetricsCollector) Collect() []*MetricSample {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	var samples []*MetricSample
	for _, metric := range mc.metrics {
		samples = append(samples, metric.Collect()...)
	}
	
	return samples
}

// Enable 启用指标收集
func (mc *MetricsCollector) Enable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = true
}

// Disable 禁用指标收集
func (mc *MetricsCollector) Disable() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.enabled = false
}

// IsEnabled 检查是否启用
func (mc *MetricsCollector) IsEnabled() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.enabled
}

// MetricsPanel 指标面板
type MetricsPanel struct {
	collector *MetricsCollector
}

// NewMetricsPanel 创建指标面板
func NewMetricsPanel(collector *MetricsCollector) *MetricsPanel {
	return &MetricsPanel{
		collector: collector,
	}
}

// RegisterRoutes 注册指标路由
func (mp *MetricsPanel) RegisterRoutes(engine any) {
	var metricsGroup *route.RouterGroup
	
	if h, ok := engine.(*route.Engine); ok {
		metricsGroup = h.Group("/yyhertz/metrics")
	} else {
		config.Error("无法注册Metrics路由，未知引擎类型")
		return
	}
	
	// 注册路由
	metricsGroup.GET("/", mp.getMetrics)
	metricsGroup.GET("/prometheus", mp.getPrometheusMetrics)
	metricsGroup.GET("/json", mp.getJSONMetrics)
	metricsGroup.POST("/enable", mp.enableCollector)
	metricsGroup.POST("/disable", mp.disableCollector)
	metricsGroup.GET("/panel", mp.metricsPanel)
	
	// 性能监控相关路由(从performance_monitor迁移)
	metricsGroup.GET("/performance", mp.getCurrentPerformanceMetrics)
	metricsGroup.GET("/performance/history", mp.getPerformanceHistory)
	metricsGroup.GET("/performance/endpoints", mp.getEndpoints)
	metricsGroup.POST("/performance/reset", mp.resetPerformanceMetrics)
	
	// 运行时监控相关路由(从runtime_metrics迁移)
	metricsGroup.GET("/runtime", mp.getCurrentRuntimeMetrics)
	metricsGroup.GET("/runtime/history", mp.getRuntimeHistory)
	metricsGroup.GET("/runtime/health", mp.getRuntimeHealthStatus)
	metricsGroup.POST("/runtime/gc", mp.triggerGC)
}

// getMetrics 获取指标
func (mp *MetricsPanel) getMetrics(ctx context.Context, c *app.RequestContext) {
	format := c.Query("format")
	if format == "prometheus" {
		mp.getPrometheusMetrics(ctx, c)
	} else {
		mp.getJSONMetrics(ctx, c)
	}
}

// getPrometheusMetrics 获取Prometheus格式指标
func (mp *MetricsPanel) getPrometheusMetrics(ctx context.Context, c *app.RequestContext) {
	samples := mp.collector.Collect()
	
	var output strings.Builder
	metricGroups := make(map[string][]*MetricSample)
	
	// 按指标名分组
	for _, sample := range samples {
		metricGroups[sample.Name] = append(metricGroups[sample.Name], sample)
	}
	
	for metricName, samples := range metricGroups {
		// 写入HELP和TYPE（简化版本）
		if len(samples) > 0 {
			output.WriteString(fmt.Sprintf("# HELP %s Auto-generated metric\n", metricName))
			output.WriteString(fmt.Sprintf("# TYPE %s gauge\n", metricName))
		}
		
		for _, sample := range samples {
			output.WriteString(sample.Name)
			if len(sample.Labels) > 0 {
				output.WriteString("{")
				var labelPairs []string
				for k, v := range sample.Labels {
					labelPairs = append(labelPairs, fmt.Sprintf("%s=\"%s\"", k, v))
				}
				output.WriteString(strings.Join(labelPairs, ","))
				output.WriteString("}")
			}
			output.WriteString(fmt.Sprintf(" %g %d\n", sample.Value, sample.Timestamp.UnixMilli()))
		}
	}
	
	c.SetContentType("text/plain; charset=utf-8")
	c.WriteString(output.String())
}

// getJSONMetrics 获取JSON格式指标
func (mp *MetricsPanel) getJSONMetrics(ctx context.Context, c *app.RequestContext) {
	samples := mp.collector.Collect()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"samples":   samples,
			"timestamp": time.Now(),
			"count":     len(samples),
		},
	})
}

// enableCollector 启用收集器
func (mp *MetricsPanel) enableCollector(ctx context.Context, c *app.RequestContext) {
	mp.collector.Enable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "指标收集已启用",
		"enabled": true,
	})
}

// disableCollector 禁用收集器
func (mp *MetricsPanel) disableCollector(ctx context.Context, c *app.RequestContext) {
	mp.collector.Disable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "指标收集已禁用",
		"enabled": false,
	})
}

// metricsPanel 指标面板页面
func (mp *MetricsPanel) metricsPanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz Metrics指标面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metrics-container { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric-item { padding: 15px; border-bottom: 1px solid #eee; }
        .metric-name { font-weight: bold; color: #007bff; }
        .metric-value { float: right; font-weight: bold; }
        .metric-labels { font-size: 12px; color: #666; margin-top: 5px; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .format-toggle { margin-bottom: 20px; }
        pre { background: #f8f9fa; padding: 15px; border-radius: 4px; overflow-x: auto; white-space: pre-wrap; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz Metrics指标面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshMetrics()">刷新指标</button>
            <button class="btn btn-success" onclick="enableCollector()">启用收集</button>
            <button class="btn btn-danger" onclick="disableCollector()">禁用收集</button>
        </div>
        <div class="format-toggle">
            <label><input type="radio" name="format" value="json" checked onchange="changeFormat()"> JSON格式</label>
            <label><input type="radio" name="format" value="prometheus" onchange="changeFormat()"> Prometheus格式</label>
        </div>
    </div>

    <div class="metrics-container" id="metricsContainer">
        <div style="padding: 20px; text-align: center; color: #666;">加载中...</div>
    </div>

    <script>
        let currentFormat = 'json';

        function changeFormat() {
            const format = document.querySelector('input[name="format"]:checked').value;
            currentFormat = format;
            refreshMetrics();
        }

        function refreshMetrics() {
            const url = currentFormat === 'prometheus' ? '/yyhertz/metrics/prometheus' : '/yyhertz/metrics/json';
            
            fetch(url)
                .then(response => {
                    if (currentFormat === 'prometheus') {
                        return response.text();
                    } else {
                        return response.json();
                    }
                })
                .then(data => {
                    if (currentFormat === 'prometheus') {
                        showPrometheusFormat(data);
                    } else {
                        showJSONFormat(data);
                    }
                })
                .catch(error => {
                    console.error('加载指标失败:', error);
                    document.getElementById('metricsContainer').innerHTML = 
                        '<div style="padding: 20px; text-align: center; color: red;">加载失败</div>';
                });
        }

        function showJSONFormat(response) {
            const container = document.getElementById('metricsContainer');
            const samples = response.data.samples || [];
            
            if (samples.length === 0) {
                container.innerHTML = '<div style="padding: 20px; text-align: center; color: #666;">暂无指标数据</div>';
                return;
            }

            let html = '';
            samples.forEach(sample => {
                let labelsStr = '';
                if (sample.labels && Object.keys(sample.labels).length > 0) {
                    const labelPairs = Object.entries(sample.labels).map(([k, v]) => k + '=' + v);
                    labelsStr = '<div class="metric-labels">{' + labelPairs.join(', ') + '}</div>';
                }

                html += '<div class="metric-item">' +
                    '<div class="metric-name">' + sample.name + 
                    '<span class="metric-value">' + sample.value + '</span></div>' +
                    labelsStr +
                    '</div>';
            });

            container.innerHTML = html;
        }

        function showPrometheusFormat(data) {
            const container = document.getElementById('metricsContainer');
            container.innerHTML = '<pre>' + data + '</pre>';
        }

        function enableCollector() {
            fetch('/yyhertz/metrics/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('指标收集已启用');
                    refreshMetrics();
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableCollector() {
            if (confirm('确定要禁用指标收集吗？')) {
                fetch('/yyhertz/metrics/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('指标收集已禁用');
                        refreshMetrics();
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        // 页面加载时初始化
        window.onload = function() {
            refreshMetrics();
            // 每30秒自动刷新
            setInterval(refreshMetrics, 30000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}

// ============= 从performance_monitor迁移的方法 =============

// GetCurrentPerformanceMetrics 获取当前性能指标(从performance_monitor迁移)
func (mc *MetricsCollector) GetCurrentPerformanceMetrics() PerformanceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	avgResponse := float64(0)
	if mc.requestCount > 0 {
		avgResponse = float64(mc.totalResponse.Nanoseconds()) / float64(mc.requestCount) / 1e6 // 转换为毫秒
	}
	
	metrics := PerformanceMetrics{
		Timestamp:    time.Now(),
		RequestCount: mc.requestCount,
		ErrorCount:   mc.errorCount,
		AvgResponse:  avgResponse,
		MaxResponse:  float64(mc.maxResponse.Nanoseconds()) / 1e6,
		MinResponse:  float64(mc.minResponse.Nanoseconds()) / 1e6,
		Goroutines:   runtime.NumGoroutine(),
		CPUUsage:     mc.getCPUUsage(),
	}
	
	metrics.Memory.Alloc = m.Alloc
	metrics.Memory.TotalAlloc = m.TotalAlloc
	metrics.Memory.Sys = m.Sys
	metrics.Memory.NumGC = m.NumGC
	
	return metrics
}

// GetPerformanceHistory 获取性能历史指标(从performance_monitor迁移)
func (mc *MetricsCollector) GetPerformanceHistory(limit int) []PerformanceMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	if limit <= 0 || limit > len(mc.metricsHistory) {
		limit = len(mc.metricsHistory)
	}
	
	start := len(mc.metricsHistory) - limit
	result := make([]PerformanceMetrics, limit)
	copy(result, mc.metricsHistory[start:])
	return result
}

// GetEndpoints 获取端点统计(从performance_monitor迁移)
func (mc *MetricsCollector) GetEndpoints() map[string]*EndpointMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*EndpointMetrics)
	for k, v := range mc.endpoints {
		// 创建副本
		endpoint := *v
		result[k] = &endpoint
	}
	return result
}

// ResetPerformanceMetrics 重置性能统计(从performance_monitor迁移)
func (mc *MetricsCollector) ResetPerformanceMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.requestCount = 0
	mc.errorCount = 0
	mc.totalResponse = 0
	mc.maxResponse = 0
	mc.minResponse = time.Hour
	mc.endpoints = make(map[string]*EndpointMetrics)
	mc.metricsHistory = make([]PerformanceMetrics, 0)
}

// getCPUUsage 获取CPU使用率（简化版本）(从performance_monitor迁移)
func (mc *MetricsCollector) getCPUUsage() float64 {
	// 这里是一个简化的CPU使用率计算
	// 实际项目中可能需要更复杂的实现
	return float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10
}

// StartPerformanceCollection 启动性能指标收集(从performance_monitor迁移)
func (mc *MetricsCollector) StartPerformanceCollection() {
	go func() {
		ticker := time.NewTicker(mc.collectInterval)
		defer ticker.Stop()
		
		for range ticker.C {
			if !mc.enabled {
				continue
			}
			
			metrics := mc.GetCurrentPerformanceMetrics()
			
			mc.mu.Lock()
			mc.metricsHistory = append(mc.metricsHistory, metrics)
			if len(mc.metricsHistory) > mc.maxHistorySize {
				mc.metricsHistory = mc.metricsHistory[1:]
			}
			mc.mu.Unlock()
		}
	}()
}

// ============= 性能监控面板路由处理方法(从performance_monitor迁移) =============

// getCurrentPerformanceMetrics 获取当前性能指标
func (mp *MetricsPanel) getCurrentPerformanceMetrics(ctx context.Context, c *app.RequestContext) {
	metrics := mp.collector.GetCurrentPerformanceMetrics()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": metrics,
	})
}

// getPerformanceHistory 获取性能历史指标
func (mp *MetricsPanel) getPerformanceHistory(ctx context.Context, c *app.RequestContext) {
	limit := 100 // 默认返回最近100条记录
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 100
		}
	}

	history := mp.collector.GetPerformanceHistory(limit)
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": history,
	})
}

// getEndpoints 获取端点统计
func (mp *MetricsPanel) getEndpoints(ctx context.Context, c *app.RequestContext) {
	endpoints := mp.collector.GetEndpoints()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": endpoints,
	})
}

// resetPerformanceMetrics 重置性能指标
func (mp *MetricsPanel) resetPerformanceMetrics(ctx context.Context, c *app.RequestContext) {
	mp.collector.ResetPerformanceMetrics()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "性能指标已重置",
	})
}

// ============= 从runtime_metrics迁移的方法 =============

// GetCurrentRuntimeMetrics 获取当前运行时指标(从runtime_metrics迁移)
func (mc *MetricsCollector) GetCurrentRuntimeMetrics() RuntimeMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	now := time.Now()
	
	// 创建基础指标
	metrics := RuntimeMetrics{
		Timestamp:    now,
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		NumCgoCall:   runtime.NumCgoCall(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		NumP:         runtime.NumCPU(), // 简化实现
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Architecture: runtime.GOARCH,
	}
	
	// 收集GC指标
	metrics.GCMetrics = mc.collectGCMetrics(&m)
	
	// 收集内存指标
	metrics.MemoryMetrics = mc.collectRuntimeMemoryMetrics(&m)
	
	// 健康评估
	metrics.HealthScore = mc.calculateHealthScore(&metrics)
	metrics.HealthStatus = mc.getHealthStatus(metrics.HealthScore)
	metrics.Recommendations = mc.getOptimizationRecommendations(&metrics)
	
	// 更新趋势数据
	mc.updateTrendData(&metrics)
	
	return metrics
}

// collectGCMetrics 收集GC指标(从runtime_metrics迁移)
func (mc *MetricsCollector) collectGCMetrics(m *runtime.MemStats) *GCMetrics {
	gcMetrics := &GCMetrics{
		NumGC:         m.NumGC,
		PauseTotal:    time.Duration(m.PauseTotalNs),
		NextGC:        m.NextGC,
		CPUFraction:   m.GCCPUFraction,
		EnablePercent: true, // Go默认启用GC
		DebugGC:       false,
	}
	
	// 最后一次GC时间
	if m.NumGC > 0 {
		gcMetrics.LastGC = time.Unix(0, int64(m.LastGC))
	}
	
	// 计算暂停时间统计
	if len(m.PauseNs) > 0 {
		var totalPause, maxPause, minPause uint64
		minPause = ^uint64(0) // 最大值
		count := int(m.NumGC)
		if count > 256 {
			count = 256
		}
		
		for i := 0; i < count; i++ {
			pause := m.PauseNs[(m.NumGC+uint32(256-i))%256]
			if pause > 0 {
				totalPause += pause
				if pause > maxPause {
					maxPause = pause
				}
				if pause < minPause {
					minPause = pause
				}
			}
		}
		
		if count > 0 {
			gcMetrics.AvgPauseTime = time.Duration(totalPause / uint64(count))
			gcMetrics.MaxPauseTime = time.Duration(maxPause)
			gcMetrics.MinPauseTime = time.Duration(minPause)
		}
		
		// 计算GC频率
		if !mc.lastCollectTime.IsZero() {
			duration := time.Since(mc.lastCollectTime).Seconds()
			if duration > 0 {
				gcMetrics.FrequencyPerSec = float64(m.NumGC) / duration
			}
		}
	}
	
	// 趋势分析
	if len(mc.gcTrendData) > 1 {
		recent := mc.gcTrendData[len(mc.gcTrendData)-1]
		previous := mc.gcTrendData[len(mc.gcTrendData)-2]
		
		if recent > previous*1.1 {
			gcMetrics.TrendDirection = "增长"
		} else if recent < previous*0.9 {
			gcMetrics.TrendDirection = "下降"
		} else {
			gcMetrics.TrendDirection = "稳定"
		}
	}
	
	return gcMetrics
}

// collectRuntimeMemoryMetrics 收集详细内存指标(从runtime_metrics迁移)
func (mc *MetricsCollector) collectRuntimeMemoryMetrics(m *runtime.MemStats) *RuntimeMemoryMetrics {
	memMetrics := &RuntimeMemoryMetrics{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		Lookups:      m.Lookups,
		Mallocs:      m.Mallocs,
		Frees:        m.Frees,
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		StackSys:     m.StackSys,
		MSpanInuse:   m.MSpanInuse,
		MSpanSys:     m.MSpanSys,
		MCacheInuse:  m.MCacheInuse,
		MCacheSys:    m.MCacheSys,
		BuckHashSys:  m.BuckHashSys,
		GCSys:        m.GCSys,
		OtherSys:     m.OtherSys,
	}
	
	// 计算分配速率
	if !mc.lastCollectTime.IsZero() && len(mc.runtimeHistory) > 0 {
		duration := time.Since(mc.lastCollectTime).Seconds()
		if duration > 0 {
			lastMetrics := mc.runtimeHistory[len(mc.runtimeHistory)-1]
			allocDiff := m.TotalAlloc - lastMetrics.MemoryMetrics.TotalAlloc
			memMetrics.AllocRate = float64(allocDiff) / duration
		}
	}
	
	// 计算GC开销
	if m.Sys > 0 {
		memMetrics.GCOverhead = (float64(m.GCSys) / float64(m.Sys)) * 100
	}
	
	// 计算内存效率
	if m.HeapSys > 0 {
		memMetrics.MemoryEfficiency = (float64(m.HeapInuse) / float64(m.HeapSys)) * 100
	}
	
	return memMetrics
}

// calculateHealthScore 计算健康评分(从runtime_metrics迁移)
func (mc *MetricsCollector) calculateHealthScore(metrics *RuntimeMetrics) int {
	score := 100
	
	// GC频率检查 (权重: 25%)
	if metrics.GCMetrics.FrequencyPerSec > 10 {
		score -= 25
	} else if metrics.GCMetrics.FrequencyPerSec > 5 {
		score -= 10
	}
	
	// GC暂停时间检查 (权重: 25%)
	if metrics.GCMetrics.AvgPauseTime > 10*time.Millisecond {
		score -= 25
	} else if metrics.GCMetrics.AvgPauseTime > 5*time.Millisecond {
		score -= 10
	}
	
	// 内存效率检查 (权重: 25%)
	if metrics.MemoryMetrics.MemoryEfficiency < 50 {
		score -= 25
	} else if metrics.MemoryMetrics.MemoryEfficiency < 70 {
		score -= 10
	}
	
	// 协程数量检查 (权重: 25%)
	goroutinesPerCPU := float64(metrics.NumGoroutine) / float64(metrics.NumCPU)
	if goroutinesPerCPU > 1000 {
		score -= 25
	} else if goroutinesPerCPU > 500 {
		score -= 10
	}
	
	if score < 0 {
		score = 0
	}
	
	return score
}

// getHealthStatus 获取健康状态(从runtime_metrics迁移)
func (mc *MetricsCollector) getHealthStatus(score int) string {
	switch {
	case score >= 90:
		return "优秀"
	case score >= 75:
		return "良好"
	case score >= 60:
		return "一般"
	case score >= 40:
		return "较差"
	default:
		return "危险"
	}
}

// getOptimizationRecommendations 获取优化建议(从runtime_metrics迁移)
func (mc *MetricsCollector) getOptimizationRecommendations(metrics *RuntimeMetrics) []string {
	var recommendations []string
	
	// GC相关建议
	if metrics.GCMetrics.FrequencyPerSec > 10 {
		recommendations = append(recommendations, "GC频率过高，建议优化内存分配策略")
	}
	
	if metrics.GCMetrics.AvgPauseTime > 10*time.Millisecond {
		recommendations = append(recommendations, "GC暂停时间过长，建议减少长时间持有的对象")
	}
	
	// 内存相关建议
	if metrics.MemoryMetrics.MemoryEfficiency < 50 {
		recommendations = append(recommendations, "内存使用效率低，建议检查内存泄漏")
	}
	
	if metrics.MemoryMetrics.AllocRate > 100*1024*1024 { // 100MB/s
		recommendations = append(recommendations, "内存分配速率过高，建议使用对象池")
	}
	
	// 协程相关建议
	goroutinesPerCPU := float64(metrics.NumGoroutine) / float64(metrics.NumCPU)
	if goroutinesPerCPU > 1000 {
		recommendations = append(recommendations, "协程数量过多，建议使用协程池或限制并发")
	}
	
	// 系统相关建议
	if metrics.MemoryMetrics.HeapSys > 1024*1024*1024 { // 1GB
		recommendations = append(recommendations, "堆内存占用较大，建议定期检查内存使用情况")
	}
	
	return recommendations
}

// updateTrendData 更新趋势数据(从runtime_metrics迁移)
func (mc *MetricsCollector) updateTrendData(metrics *RuntimeMetrics) {
	// 更新GC趋势
	mc.gcTrendData = append(mc.gcTrendData, metrics.GCMetrics.FrequencyPerSec)
	if len(mc.gcTrendData) > 100 {
		mc.gcTrendData = mc.gcTrendData[1:]
	}
	
	// 更新内存趋势
	mc.memTrendData = append(mc.memTrendData, float64(metrics.MemoryMetrics.Alloc))
	if len(mc.memTrendData) > 100 {
		mc.memTrendData = mc.memTrendData[1:]
	}
	
	mc.lastCollectTime = time.Now()
}

// GetRuntimeHistory 获取运行时历史指标(从runtime_metrics迁移)
func (mc *MetricsCollector) GetRuntimeHistory(limit int) []RuntimeMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	if limit <= 0 || limit > len(mc.runtimeHistory) {
		limit = len(mc.runtimeHistory)
	}
	
	start := len(mc.runtimeHistory) - limit
	result := make([]RuntimeMetrics, limit)
	copy(result, mc.runtimeHistory[start:])
	return result
}

// StartRuntimeCollection 启动运行时指标收集(从runtime_metrics迁移)
func (mc *MetricsCollector) StartRuntimeCollection() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			if !mc.enabled {
				continue
			}
			
			metrics := mc.GetCurrentRuntimeMetrics()
			
			mc.mu.Lock()
			mc.runtimeHistory = append(mc.runtimeHistory, metrics)
			if len(mc.runtimeHistory) > mc.maxHistorySize {
				mc.runtimeHistory = mc.runtimeHistory[1:]
			}
			mc.mu.Unlock()
		}
	}()
}

// ============= 运行时监控面板路由处理方法(从runtime_metrics迁移) =============

// getCurrentRuntimeMetrics 获取当前运行时指标
func (mp *MetricsPanel) getCurrentRuntimeMetrics(ctx context.Context, c *app.RequestContext) {
	metrics := mp.collector.GetCurrentRuntimeMetrics()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": metrics,
	})
}

// getRuntimeHistory 获取运行时历史指标
func (mp *MetricsPanel) getRuntimeHistory(ctx context.Context, c *app.RequestContext) {
	limit := 50 // 默认返回最近50条记录
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || l != 1 {
			limit = 50
		}
	}

	history := mp.collector.GetRuntimeHistory(limit)
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": history,
	})
}

// getRuntimeHealthStatus 获取运行时健康状态
func (mp *MetricsPanel) getRuntimeHealthStatus(ctx context.Context, c *app.RequestContext) {
	metrics := mp.collector.GetCurrentRuntimeMetrics()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": map[string]any{
			"health_score": metrics.HealthScore,
			"health_status": metrics.HealthStatus,
			"recommendations": metrics.Recommendations,
		},
	})
}

// triggerGC 触发GC
func (mp *MetricsPanel) triggerGC(ctx context.Context, c *app.RequestContext) {
	runtime.GC()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "GC已触发",
	})
}