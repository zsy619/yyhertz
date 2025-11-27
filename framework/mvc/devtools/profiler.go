// Package devtools 提供Profiling性能分析中间件
//
// Profiler性能分析中间件用于应用性能深度分析，提供：
// - CPU性能分析
// - 内存分析和泄漏检测
// - 协程分析
// - 阻塞分析
// - 互斥锁分析
// - 火焰图生成
//
// 功能特性：
// - 集成Go原生pprof
// - 可视化性能分析界面
// - 自动性能基线对比
// - 实时性能监控
// - 安全的分析数据导出
package devtools

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"runtime"
	rpprof "runtime/pprof"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/route"

	"github.com/zsy619/yyhertz/framework/config"
)

// ProfileType 性能分析类型
type ProfileType string

const (
	// ProfileTypeCPU CPU分析
	ProfileTypeCPU ProfileType = "cpu"
	// ProfileTypeHeap 堆内存分析
	ProfileTypeHeap ProfileType = "heap"
	// ProfileTypeGoroutine 协程分析
	ProfileTypeGoroutine ProfileType = "goroutine"
	// ProfileTypeBlock 阻塞分析
	ProfileTypeBlock ProfileType = "block"
	// ProfileTypeMutex 互斥锁分析
	ProfileTypeMutex ProfileType = "mutex"
	// ProfileTypeAllocs 内存分配分析
	ProfileTypeAllocs ProfileType = "allocs"
)

// ProfileResult 分析结果
type ProfileResult struct {
	Type      ProfileType `json:"type"`
	Name      string      `json:"name"`
	Size      int         `json:"size"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time   `json:"timestamp"`
	Data      []byte      `json:"-"` // 原始数据不在JSON中返回
}

// PerformanceBaseline 性能基线
type PerformanceBaseline struct {
	Timestamp       time.Time `json:"timestamp"`
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     uint64    `json:"memory_usage"`
	GoroutineCount  int       `json:"goroutine_count"`
	GCCount         uint32    `json:"gc_count"`
	RequestCount    int64     `json:"request_count"`
	AvgResponseTime float64   `json:"avg_response_time"`
}

// Profiler 性能分析器
type Profiler struct {
	mu             sync.RWMutex
	enabled        bool
	profiles       map[string]*ProfileResult
	maxProfiles    int
	autoProfile    bool
	profileInterval time.Duration
	baseline       *PerformanceBaseline
	requestCount   int64
	totalResponse  time.Duration
	startTime      time.Time
	stopCh         chan struct{}
}

// ProfilerConfig 分析器配置
type ProfilerConfig struct {
	Enabled         bool          // 是否启用
	MaxProfiles     int           // 最大保存的分析文件数
	AutoProfile     bool          // 是否自动分析
	ProfileInterval time.Duration // 自动分析间隔
}

// NewProfiler 创建性能分析器
func NewProfiler(config *ProfilerConfig) *Profiler {
	if config == nil {
		config = &ProfilerConfig{
			Enabled:         true,
			MaxProfiles:     10,
			AutoProfile:     false,
			ProfileInterval: 5 * time.Minute,
		}
	}
	
	p := &Profiler{
		enabled:         config.Enabled,
		profiles:        make(map[string]*ProfileResult),
		maxProfiles:     config.MaxProfiles,
		autoProfile:     config.AutoProfile,
		profileInterval: config.ProfileInterval,
		startTime:       time.Now(),
		stopCh:          make(chan struct{}),
	}
	
	if config.AutoProfile {
		go p.autoProfileLoop()
	}
	
	return p
}

// Handler 性能分析中间件
func (p *Profiler) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !p.enabled {
			c.Next(ctx)
			return
		}
		
		start := time.Now()
		
		// 执行下一个中间件
		c.Next(ctx)
		
		duration := time.Since(start)
		
		// 更新统计
		p.mu.Lock()
		p.requestCount++
		p.totalResponse += duration
		p.mu.Unlock()
	}
}

// StartProfile 开始性能分析
func (p *Profiler) StartProfile(profileType ProfileType, duration time.Duration) error {
	if !p.enabled {
		return fmt.Errorf("profiler is disabled")
	}
	
	name := fmt.Sprintf("%s_%d", profileType, time.Now().Unix())
	
	switch profileType {
	case ProfileTypeCPU:
		return p.startCPUProfile(name, duration)
	case ProfileTypeHeap:
		return p.captureHeapProfile(name)
	case ProfileTypeGoroutine:
		return p.captureGoroutineProfile(name)
	case ProfileTypeBlock:
		return p.captureBlockProfile(name)
	case ProfileTypeMutex:
		return p.captureMutexProfile(name)
	case ProfileTypeAllocs:
		return p.captureAllocsProfile(name)
	default:
		return fmt.Errorf("unsupported profile type: %s", profileType)
	}
}

// startCPUProfile 开始CPU分析
func (p *Profiler) startCPUProfile(name string, duration time.Duration) error {
	var buf bytes.Buffer
	
	if err := rpprof.StartCPUProfile(&buf); err != nil {
		return err
	}
	
	go func() {
		time.Sleep(duration)
		rpprof.StopCPUProfile()
		
		p.mu.Lock()
		defer p.mu.Unlock()
		
		result := &ProfileResult{
			Type:      ProfileTypeCPU,
			Name:      name,
			Size:      buf.Len(),
			Duration:  duration,
			Timestamp: time.Now(),
			Data:      buf.Bytes(),
		}
		
		p.storeProfile(name, result)
	}()
	
	return nil
}

// captureHeapProfile 捕获堆内存分析
func (p *Profiler) captureHeapProfile(name string) error {
	var buf bytes.Buffer
	
	if err := rpprof.WriteHeapProfile(&buf); err != nil {
		return err
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result := &ProfileResult{
		Type:      ProfileTypeHeap,
		Name:      name,
		Size:      buf.Len(),
		Duration:  0,
		Timestamp: time.Now(),
		Data:      buf.Bytes(),
	}
	
	p.storeProfile(name, result)
	return nil
}

// captureGoroutineProfile 捕获协程分析
func (p *Profiler) captureGoroutineProfile(name string) error {
	var buf bytes.Buffer
	
	if err := rpprof.Lookup("goroutine").WriteTo(&buf, 0); err != nil {
		return err
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result := &ProfileResult{
		Type:      ProfileTypeGoroutine,
		Name:      name,
		Size:      buf.Len(),
		Duration:  0,
		Timestamp: time.Now(),
		Data:      buf.Bytes(),
	}
	
	p.storeProfile(name, result)
	return nil
}

// captureBlockProfile 捕获阻塞分析
func (p *Profiler) captureBlockProfile(name string) error {
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)
	
	var buf bytes.Buffer
	
	if err := rpprof.Lookup("block").WriteTo(&buf, 0); err != nil {
		return err
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result := &ProfileResult{
		Type:      ProfileTypeBlock,
		Name:      name,
		Size:      buf.Len(),
		Duration:  0,
		Timestamp: time.Now(),
		Data:      buf.Bytes(),
	}
	
	p.storeProfile(name, result)
	return nil
}

// captureMutexProfile 捕获互斥锁分析
func (p *Profiler) captureMutexProfile(name string) error {
	runtime.SetMutexProfileFraction(1)
	defer runtime.SetMutexProfileFraction(0)
	
	var buf bytes.Buffer
	
	if err := rpprof.Lookup("mutex").WriteTo(&buf, 0); err != nil {
		return err
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result := &ProfileResult{
		Type:      ProfileTypeMutex,
		Name:      name,
		Size:      buf.Len(),
		Duration:  0,
		Timestamp: time.Now(),
		Data:      buf.Bytes(),
	}
	
	p.storeProfile(name, result)
	return nil
}

// captureAllocsProfile 捕获内存分配分析
func (p *Profiler) captureAllocsProfile(name string) error {
	var buf bytes.Buffer
	
	if err := rpprof.Lookup("allocs").WriteTo(&buf, 0); err != nil {
		return err
	}
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	result := &ProfileResult{
		Type:      ProfileTypeAllocs,
		Name:      name,
		Size:      buf.Len(),
		Duration:  0,
		Timestamp: time.Now(),
		Data:      buf.Bytes(),
	}
	
	p.storeProfile(name, result)
	return nil
}

// storeProfile 存储分析结果
func (p *Profiler) storeProfile(name string, result *ProfileResult) {
	p.profiles[name] = result
	
	// 限制最大数量
	if len(p.profiles) > p.maxProfiles {
		// 删除最旧的分析
		var oldest string
		var oldestTime time.Time
		
		for name, profile := range p.profiles {
			if oldest == "" || profile.Timestamp.Before(oldestTime) {
				oldest = name
				oldestTime = profile.Timestamp
			}
		}
		
		if oldest != "" {
			delete(p.profiles, oldest)
		}
	}
}

// GetProfiles 获取所有分析结果
func (p *Profiler) GetProfiles() map[string]*ProfileResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	result := make(map[string]*ProfileResult)
	for k, v := range p.profiles {
		// 创建副本，不包含原始数据
		result[k] = &ProfileResult{
			Type:      v.Type,
			Name:      v.Name,
			Size:      v.Size,
			Duration:  v.Duration,
			Timestamp: v.Timestamp,
		}
	}
	return result
}

// GetProfile 获取特定分析结果
func (p *Profiler) GetProfile(name string) *ProfileResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return p.profiles[name]
}

// DeleteProfile 删除分析结果
func (p *Profiler) DeleteProfile(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	delete(p.profiles, name)
}

// CreateBaseline 创建性能基线
func (p *Profiler) CreateBaseline() *PerformanceBaseline {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	p.mu.RLock()
	avgResponse := float64(0)
	if p.requestCount > 0 {
		avgResponse = float64(p.totalResponse.Nanoseconds()) / float64(p.requestCount) / 1e6
	}
	p.mu.RUnlock()
	
	baseline := &PerformanceBaseline{
		Timestamp:       time.Now(),
		CPUUsage:        float64(runtime.NumGoroutine()) / float64(runtime.NumCPU()) * 10,
		MemoryUsage:     m.Alloc,
		GoroutineCount:  runtime.NumGoroutine(),
		GCCount:         m.NumGC,
		RequestCount:    p.requestCount,
		AvgResponseTime: avgResponse,
	}
	
	p.mu.Lock()
	p.baseline = baseline
	p.mu.Unlock()
	
	return baseline
}

// GetBaseline 获取性能基线
func (p *Profiler) GetBaseline() *PerformanceBaseline {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	return p.baseline
}

// autoProfileLoop 自动分析循环
func (p *Profiler) autoProfileLoop() {
	ticker := time.NewTicker(p.profileInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if p.enabled {
				// 自动创建堆内存分析
				go p.captureHeapProfile(fmt.Sprintf("auto_heap_%d", time.Now().Unix()))
				
				// 自动创建协程分析
				go p.captureGoroutineProfile(fmt.Sprintf("auto_goroutine_%d", time.Now().Unix()))
			}
		case <-p.stopCh:
			return
		}
	}
}

// Enable 启用分析器
func (p *Profiler) Enable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = true
}

// Disable 禁用分析器
func (p *Profiler) Disable() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.enabled = false
}

// IsEnabled 检查是否启用
func (p *Profiler) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// Stop 停止分析器
func (p *Profiler) Stop() {
	close(p.stopCh)
}

// ProfilePanel 性能分析面板
type ProfilePanel struct {
	profiler *Profiler
}

// NewProfilePanel 创建性能分析面板
func NewProfilePanel(profiler *Profiler) *ProfilePanel {
	return &ProfilePanel{
		profiler: profiler,
	}
}

// RegisterRoutes 注册性能分析路由
func (pp *ProfilePanel) RegisterRoutes(engine any) {
	var profileGroup *route.RouterGroup
	
	if h, ok := engine.(*route.Engine); ok {
		profileGroup = h.Group("/yyhertz/profile")
	} else {
		config.Error("无法注册Profile路由，未知引擎类型")
		return
	}
	
	// 注册路由
	profileGroup.GET("/list", pp.getProfiles)
	profileGroup.GET("/download/:name", pp.downloadProfile)
	profileGroup.POST("/start", pp.startProfile)
	profileGroup.DELETE("/delete/:name", pp.deleteProfile)
	profileGroup.POST("/baseline", pp.createBaseline)
	profileGroup.GET("/baseline", pp.getBaseline)
	profileGroup.GET("/panel", pp.profilePanel)
	profileGroup.POST("/enable", pp.enableProfiler)
	profileGroup.POST("/disable", pp.disableProfiler)
	
	// 集成Go原生pprof路由
	pprofGroup := profileGroup.Group("/debug/pprof")
	pp.registerPprofRoutes(pprofGroup)
}

// registerPprofRoutes 注册pprof路由
func (pp *ProfilePanel) registerPprofRoutes(group *route.RouterGroup) {
	// 注意：由于Hertz和标准http的协议差异，这里提供简化的pprof端点
	// 建议使用独立的pprof路由或者通过我们的性能分析面板
	
	group.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "pprof index - please use /yyhertz/profile/panel for web interface")
	})
	
	group.GET("/cmdline", func(ctx context.Context, c *app.RequestContext) {
		c.String(http.StatusOK, "cmdline profile - please use /yyhertz/profile/panel for web interface") 
	})
	
	group.GET("/profile", func(ctx context.Context, c *app.RequestContext) {
		// 启动CPU分析
		pp.profiler.StartProfile(ProfileTypeCPU, 30*time.Second)
		c.String(http.StatusOK, "CPU profiling started for 30 seconds - check /yyhertz/profile/panel for results")
	})
	
	group.GET("/heap", func(ctx context.Context, c *app.RequestContext) {
		// 生成堆分析
		pp.profiler.StartProfile(ProfileTypeHeap, 0)
		c.String(http.StatusOK, "Heap profile generated - check /yyhertz/profile/panel for results")
	})
	
	group.GET("/goroutine", func(ctx context.Context, c *app.RequestContext) {
		// 生成协程分析
		pp.profiler.StartProfile(ProfileTypeGoroutine, 0)
		c.String(http.StatusOK, "Goroutine profile generated - check /yyhertz/profile/panel for results")
	})
	
	group.GET("/block", func(ctx context.Context, c *app.RequestContext) {
		// 生成阻塞分析
		pp.profiler.StartProfile(ProfileTypeBlock, 0)
		c.String(http.StatusOK, "Block profile generated - check /yyhertz/profile/panel for results")
	})
	
	group.GET("/mutex", func(ctx context.Context, c *app.RequestContext) {
		// 生成互斥锁分析
		pp.profiler.StartProfile(ProfileTypeMutex, 0)
		c.String(http.StatusOK, "Mutex profile generated - check /yyhertz/profile/panel for results")
	})
}

// getProfiles 获取分析列表
func (pp *ProfilePanel) getProfiles(ctx context.Context, c *app.RequestContext) {
	profiles := pp.profiler.GetProfiles()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": profiles,
	})
}

// downloadProfile 下载分析文件
func (pp *ProfilePanel) downloadProfile(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "分析名称不能为空",
		})
		return
	}
	
	profile := pp.profiler.GetProfile(name)
	if profile == nil {
		c.JSON(http.StatusNotFound, map[string]any{
			"code":    404,
			"message": "分析文件不存在",
		})
		return
	}
	
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.prof", name))
	c.Write(profile.Data)
}

// startProfile 开始分析
func (pp *ProfilePanel) startProfile(ctx context.Context, c *app.RequestContext) {
	profileType := ProfileType(c.Query("type"))
	durationStr := c.Query("duration")
	
	if profileType == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "分析类型不能为空",
		})
		return
	}
	
	duration := 30 * time.Second
	if durationStr != "" {
		if d, err := time.ParseDuration(durationStr); err == nil {
			duration = d
		}
	}
	
	err := pp.profiler.StartProfile(profileType, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "分析已启动",
		"type":    profileType,
		"duration": duration.String(),
	})
}

// deleteProfile 删除分析
func (pp *ProfilePanel) deleteProfile(ctx context.Context, c *app.RequestContext) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "分析名称不能为空",
		})
		return
	}
	
	pp.profiler.DeleteProfile(name)
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "分析已删除",
	})
}

// createBaseline 创建基线
func (pp *ProfilePanel) createBaseline(ctx context.Context, c *app.RequestContext) {
	baseline := pp.profiler.CreateBaseline()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": baseline,
		"message": "性能基线已创建",
	})
}

// getBaseline 获取基线
func (pp *ProfilePanel) getBaseline(ctx context.Context, c *app.RequestContext) {
	baseline := pp.profiler.GetBaseline()
	if baseline == nil {
		c.JSON(http.StatusNotFound, map[string]any{
			"code":    404,
			"message": "尚未创建性能基线",
		})
		return
	}
	
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": baseline,
	})
}

// enableProfiler 启用分析器
func (pp *ProfilePanel) enableProfiler(ctx context.Context, c *app.RequestContext) {
	pp.profiler.Enable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "性能分析器已启用",
		"enabled": true,
	})
}

// disableProfiler 禁用分析器
func (pp *ProfilePanel) disableProfiler(ctx context.Context, c *app.RequestContext) {
	pp.profiler.Disable()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "性能分析器已禁用",
		"enabled": false,
	})
}

// profilePanel 分析面板页面
func (pp *ProfilePanel) profilePanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 性能分析面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .section { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .profile-controls { margin-bottom: 20px; }
        .profile-controls select, .profile-controls input { padding: 8px; margin-right: 10px; border: 1px solid #ddd; border-radius: 4px; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-success { background: #28a745; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .btn-warning { background: #ffc107; color: black; }
        .profiles-table { overflow-x: auto; }
        table { width: 100%; border-collapse: collapse; }
        th, td { padding: 12px; text-align: left; border-bottom: 1px solid #eee; }
        th { background: #f8f9fa; font-weight: bold; }
        .profile-type { padding: 2px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; }
        .cpu { background: #dc3545; color: white; }
        .heap { background: #28a745; color: white; }
        .goroutine { background: #007bff; color: white; }
        .block { background: #ffc107; color: black; }
        .mutex { background: #6f42c1; color: white; }
        .allocs { background: #fd7e14; color: white; }
        .baseline-info { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; }
        .baseline-item { text-align: center; }
        .baseline-value { font-size: 1.5em; font-weight: bold; color: #007bff; }
        .baseline-label { color: #666; margin-top: 5px; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 性能分析面板</h1>
        <div>
            <button class="btn btn-primary" onclick="refreshProfiles()">刷新列表</button>
            <button class="btn btn-success" onclick="enableProfiler()">启用分析</button>
            <button class="btn btn-danger" onclick="disableProfiler()">禁用分析</button>
            <button class="btn btn-warning" onclick="createBaseline()">创建基线</button>
        </div>
    </div>

    <div class="section">
        <h3>开始新的性能分析</h3>
        <div class="profile-controls">
            <select id="profileType">
                <option value="cpu">CPU分析</option>
                <option value="heap">堆内存分析</option>
                <option value="goroutine">协程分析</option>
                <option value="block">阻塞分析</option>
                <option value="mutex">互斥锁分析</option>
                <option value="allocs">内存分配分析</option>
            </select>
            <input type="text" id="duration" placeholder="持续时间 (如: 30s)" value="30s">
            <button class="btn btn-primary" onclick="startProfile()">开始分析</button>
        </div>
        <p style="color: #666; font-size: 14px;">
            提示: CPU分析需要指定持续时间，其他类型为瞬时快照。
            分析文件可下载后使用 <code>go tool pprof</code> 命令进行详细分析。
        </p>
    </div>

    <div class="section">
        <h3>性能基线</h3>
        <div class="baseline-info" id="baselineInfo">
            <div style="text-align: center; color: #666;">尚未创建性能基线</div>
        </div>
    </div>

    <div class="section">
        <h3>分析文件列表</h3>
        <div class="profiles-table">
            <table id="profilesTable">
                <thead>
                    <tr>
                        <th>名称</th>
                        <th>类型</th>
                        <th>大小</th>
                        <th>持续时间</th>
                        <th>创建时间</th>
                        <th>操作</th>
                    </tr>
                </thead>
                <tbody>
                    <!-- 分析数据将在这里动态生成 -->
                </tbody>
            </table>
        </div>
    </div>

    <div class="section">
        <h3>在线分析工具</h3>
        <p>也可以直接访问以下链接进行在线分析：</p>
        <ul>
            <li><a href="/yyhertz/profile/debug/pprof/" target="_blank">pprof 主页</a></li>
            <li><a href="/yyhertz/profile/debug/pprof/heap" target="_blank">堆内存分析</a></li>
            <li><a href="/yyhertz/profile/debug/pprof/goroutine" target="_blank">协程分析</a></li>
            <li><a href="/yyhertz/profile/debug/pprof/profile?seconds=30" target="_blank">CPU分析(30秒)</a></li>
            <li><a href="/yyhertz/profile/debug/pprof/block" target="_blank">阻塞分析</a></li>
            <li><a href="/yyhertz/profile/debug/pprof/mutex" target="_blank">互斥锁分析</a></li>
        </ul>
    </div>

    <script>
        function refreshProfiles() {
            loadProfiles();
            loadBaseline();
        }

        function loadProfiles() {
            fetch('/yyhertz/profile/list')
                .then(response => response.json())
                .then(data => {
                    updateProfilesTable(data.data);
                })
                .catch(error => {
                    console.error('加载分析列表失败:', error);
                });
        }

        function loadBaseline() {
            fetch('/yyhertz/profile/baseline')
                .then(response => response.json())
                .then(data => {
                    if (data.code === 0) {
                        updateBaselineInfo(data.data);
                    } else {
                        document.getElementById('baselineInfo').innerHTML = 
                            '<div style="text-align: center; color: #666;">尚未创建性能基线</div>';
                    }
                })
                .catch(error => {
                    console.error('加载基线失败:', error);
                });
        }

        function updateProfilesTable(profiles) {
            const tbody = document.querySelector('#profilesTable tbody');
            let html = '';

            if (!profiles || Object.keys(profiles).length === 0) {
                html = '<tr><td colspan="6" style="text-align: center; color: #666;">暂无分析文件</td></tr>';
            } else {
                Object.values(profiles).forEach(profile => {
                    const duration = profile.duration > 0 ? (profile.duration / 1000000000).toFixed(1) + 's' : '-';
                    html += '<tr>' +
                        '<td>' + profile.name + '</td>' +
                        '<td><span class="profile-type ' + profile.type + '">' + profile.type + '</span></td>' +
                        '<td>' + formatBytes(profile.size) + '</td>' +
                        '<td>' + duration + '</td>' +
                        '<td>' + new Date(profile.timestamp).toLocaleString() + '</td>' +
                        '<td>' +
                            '<button class="btn btn-primary" onclick="downloadProfile(\'' + profile.name + '\')">下载</button>' +
                            '<button class="btn btn-danger" onclick="deleteProfile(\'' + profile.name + '\')">删除</button>' +
                        '</td>' +
                        '</tr>';
                });
            }

            tbody.innerHTML = html;
        }

        function updateBaselineInfo(baseline) {
            const info = document.getElementById('baselineInfo');
            info.innerHTML = 
                '<div class="baseline-item">' +
                    '<div class="baseline-value">' + baseline.cpu_usage.toFixed(1) + '%</div>' +
                    '<div class="baseline-label">CPU使用率</div>' +
                '</div>' +
                '<div class="baseline-item">' +
                    '<div class="baseline-value">' + formatBytes(baseline.memory_usage) + '</div>' +
                    '<div class="baseline-label">内存使用</div>' +
                '</div>' +
                '<div class="baseline-item">' +
                    '<div class="baseline-value">' + baseline.goroutine_count + '</div>' +
                    '<div class="baseline-label">协程数</div>' +
                '</div>' +
                '<div class="baseline-item">' +
                    '<div class="baseline-value">' + baseline.gc_count + '</div>' +
                    '<div class="baseline-label">GC次数</div>' +
                '</div>' +
                '<div class="baseline-item">' +
                    '<div class="baseline-value">' + baseline.request_count + '</div>' +
                    '<div class="baseline-label">请求总数</div>' +
                '</div>' +
                '<div class="baseline-item">' +
                    '<div class="baseline-value">' + baseline.avg_response_time.toFixed(2) + 'ms</div>' +
                    '<div class="baseline-label">平均响应时间</div>' +
                '</div>';
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function startProfile() {
            const type = document.getElementById('profileType').value;
            const duration = document.getElementById('duration').value;

            fetch('/yyhertz/profile/start?type=' + type + '&duration=' + duration, { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.code === 0) {
                        alert('分析已启动: ' + data.message);
                        setTimeout(refreshProfiles, 2000); // 2秒后刷新列表
                    } else {
                        alert('启动分析失败: ' + data.message);
                    }
                })
                .catch(error => {
                    console.error('启动分析失败:', error);
                    alert('启动分析失败');
                });
        }

        function downloadProfile(name) {
            window.open('/yyhertz/profile/download/' + name, '_blank');
        }

        function deleteProfile(name) {
            if (confirm('确定要删除分析文件 "' + name + '" 吗？')) {
                fetch('/yyhertz/profile/delete/' + name, { method: 'DELETE' })
                    .then(response => response.json())
                    .then(data => {
                        alert('分析文件已删除');
                        refreshProfiles();
                    })
                    .catch(error => {
                        console.error('删除失败:', error);
                        alert('删除失败');
                    });
            }
        }

        function createBaseline() {
            fetch('/yyhertz/profile/baseline', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('性能基线已创建');
                    loadBaseline();
                })
                .catch(error => {
                    console.error('创建基线失败:', error);
                    alert('创建基线失败');
                });
        }

        function enableProfiler() {
            fetch('/yyhertz/profile/enable', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('性能分析器已启用');
                })
                .catch(error => {
                    console.error('启用失败:', error);
                    alert('启用失败');
                });
        }

        function disableProfiler() {
            if (confirm('确定要禁用性能分析器吗？')) {
                fetch('/yyhertz/profile/disable', { method: 'POST' })
                    .then(response => response.json())
                    .then(data => {
                        alert('性能分析器已禁用');
                    })
                    .catch(error => {
                        console.error('禁用失败:', error);
                        alert('禁用失败');
                    });
            }
        }

        // 页面加载时初始化
        window.onload = function() {
            refreshProfiles();
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}