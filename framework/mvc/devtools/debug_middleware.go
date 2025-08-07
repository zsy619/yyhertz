package devtools

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/route"
)

// DebugInfo 调试信息
type DebugInfo struct {
	RequestID  string            `json:"request_id"`
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Headers    map[string]string `json:"headers"`
	Query      map[string]string `json:"query"`
	Body       string            `json:"body"`
	Response   string            `json:"response"`
	StatusCode int               `json:"status_code"`
	Duration   time.Duration     `json:"duration"`
	Memory     MemoryInfo        `json:"memory"`
	Goroutines int               `json:"goroutines"`
	Stack      []StackFrame      `json:"stack,omitempty"`
	Middleware []MiddlewareInfo  `json:"middleware"`
	Controller string            `json:"controller,omitempty"`
	Action     string            `json:"action,omitempty"`
	Errors     []string          `json:"errors,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// MemoryInfo 内存信息
type MemoryInfo struct {
	Alloc      uint64 `json:"alloc"`       // 当前分配的内存
	TotalAlloc uint64 `json:"total_alloc"` // 总分配的内存
	Sys        uint64 `json:"sys"`         // 系统内存
	NumGC      uint32 `json:"num_gc"`      // GC次数
}

// StackFrame 堆栈帧
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// MiddlewareInfo 中间件信息
type MiddlewareInfo struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Order    int           `json:"order"`
}

// DebugMiddleware 调试中间件
type DebugMiddleware struct {
	enabled     bool
	maxBodySize int64
	storage     *DebugStorage
	mu          sync.RWMutex
}

// DebugStorage 调试信息存储
type DebugStorage struct {
	requests map[string]*DebugInfo
	mu       sync.RWMutex
	maxSize  int
}

// NewDebugStorage 创建调试存储
func NewDebugStorage(maxSize int) *DebugStorage {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &DebugStorage{
		requests: make(map[string]*DebugInfo),
		maxSize:  maxSize,
	}
}

// Store 存储调试信息
func (ds *DebugStorage) Store(info *DebugInfo) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// 如果超过最大大小，删除最旧的记录
	if len(ds.requests) >= ds.maxSize {
		var oldestID string
		var oldestTime time.Time

		for id, req := range ds.requests {
			if oldestID == "" || req.Timestamp.Before(oldestTime) {
				oldestID = id
				oldestTime = req.Timestamp
			}
		}

		if oldestID != "" {
			delete(ds.requests, oldestID)
		}
	}

	ds.requests[info.RequestID] = info
}

// Get 获取调试信息
func (ds *DebugStorage) Get(requestID string) (*DebugInfo, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	info, exists := ds.requests[requestID]
	return info, exists
}

// GetAll 获取所有调试信息
func (ds *DebugStorage) GetAll() map[string]*DebugInfo {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	result := make(map[string]*DebugInfo)
	for k, v := range ds.requests {
		result[k] = v
	}
	return result
}

// Clear 清空调试信息
func (ds *DebugStorage) Clear() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.requests = make(map[string]*DebugInfo)
}

// NewDebugMiddleware 创建调试中间件
func NewDebugMiddleware() *DebugMiddleware {
	return &DebugMiddleware{
		enabled:     true,
		maxBodySize: 1024 * 1024, // 1MB
		storage:     NewDebugStorage(1000),
	}
}

// Enable 启用调试
func (dm *DebugMiddleware) Enable() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.enabled = true
}

// Disable 禁用调试
func (dm *DebugMiddleware) Disable() {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.enabled = false
}

// IsEnabled 检查是否启用
func (dm *DebugMiddleware) IsEnabled() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.enabled
}

// SetMaxBodySize 设置最大请求体大小
func (dm *DebugMiddleware) SetMaxBodySize(size int64) {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	dm.maxBodySize = size
}

// Handler 中间件处理函数
func (dm *DebugMiddleware) Handler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		if !dm.IsEnabled() {
			c.Next(ctx)
			return
		}

		startTime := time.Now()

		// 生成请求ID
		requestID := dm.generateRequestID(c)
		c.Header("X-Debug-Request-ID", requestID)

		// 收集请求信息
		debugInfo := &DebugInfo{
			RequestID:  requestID,
			Method:     string(c.Method()),
			Path:       string(c.Path()),
			Headers:    dm.extractHeaders(c),
			Query:      dm.extractQuery(c),
			Body:       dm.extractBody(c),
			Timestamp:  startTime,
			Middleware: []MiddlewareInfo{},
		}

		// 收集内存信息
		debugInfo.Memory = dm.getMemoryInfo()
		debugInfo.Goroutines = runtime.NumGoroutine()

		// 在上下文中设置调试信息
		c.Set("debug_info", debugInfo)

		// 创建响应缓冲区来捕获响应
		var responseBuffer bytes.Buffer

		// 执行下一个中间件
		c.Next(ctx)

		// 收集响应信息
		debugInfo.StatusCode = c.Response.StatusCode()
		debugInfo.Response = responseBuffer.String()
		debugInfo.Duration = time.Since(startTime)

		// 收集错误信息
		if errors := c.Errors; len(errors) > 0 {
			debugInfo.Errors = make([]string, len(errors))
			for i, err := range errors {
				debugInfo.Errors[i] = err.Error()
			}
		}

		// 收集堆栈信息（如果有错误）
		if len(debugInfo.Errors) > 0 {
			debugInfo.Stack = dm.getStackTrace()
		}

		// 存储调试信息
		dm.storage.Store(debugInfo)

		// 记录调试日志
		dm.logDebugInfo(debugInfo)
	}
}

// generateRequestID 生成请求ID
func (dm *DebugMiddleware) generateRequestID(c *app.RequestContext) string {
	// 尝试从请求头获取
	if id := c.GetHeader("X-Request-ID"); len(id) > 0 {
		return string(id)
	}

	// 生成新的ID
	return fmt.Sprintf("debug_%d", time.Now().UnixNano())
}

// extractHeaders 提取请求头
func (dm *DebugMiddleware) extractHeaders(c *app.RequestContext) map[string]string {
	headers := make(map[string]string)
	c.Request.Header.VisitAll(func(key, value []byte) {
		headers[string(key)] = string(value)
	})
	return headers
}

// extractQuery 提取查询参数
func (dm *DebugMiddleware) extractQuery(c *app.RequestContext) map[string]string {
	query := make(map[string]string)
	c.QueryArgs().VisitAll(func(key, value []byte) {
		query[string(key)] = string(value)
	})
	return query
}

// extractBody 提取请求体
func (dm *DebugMiddleware) extractBody(c *app.RequestContext) string {
	body := c.Request.Body()
	if len(body) == 0 {
		return ""
	}

	if int64(len(body)) > dm.maxBodySize {
		return fmt.Sprintf("[请求体过大，已截断] 大小: %d bytes", len(body))
	}

	// 检查是否为二进制数据
	if dm.isBinary(body) {
		return fmt.Sprintf("[二进制数据] 大小: %d bytes", len(body))
	}

	return string(body)
}

// isBinary 检查是否为二进制数据
func (dm *DebugMiddleware) isBinary(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

// getMemoryInfo 获取内存信息
func (dm *DebugMiddleware) getMemoryInfo() MemoryInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryInfo{
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		NumGC:      m.NumGC,
	}
}

// getStackTrace 获取堆栈跟踪
func (dm *DebugMiddleware) getStackTrace() []StackFrame {
	var frames []StackFrame

	// 获取调用栈
	pc := make([]uintptr, 32)
	n := runtime.Callers(3, pc) // 跳过前3个帧

	if n > 0 {
		callersFrames := runtime.CallersFrames(pc[:n])
		for {
			frame, more := callersFrames.Next()

			frames = append(frames, StackFrame{
				Function: frame.Function,
				File:     frame.File,
				Line:     frame.Line,
			})

			if !more {
				break
			}
		}
	}

	return frames
}

// logDebugInfo 记录调试信息
func (dm *DebugMiddleware) logDebugInfo(info *DebugInfo) {
	hlog.Infof("[DEBUG] %s %s - %d - %v - Memory: %d KB - Goroutines: %d",
		info.Method,
		info.Path,
		info.StatusCode,
		info.Duration,
		info.Memory.Alloc/1024,
		info.Goroutines,
	)
}

// GetStorage 获取存储
func (dm *DebugMiddleware) GetStorage() *DebugStorage {
	return dm.storage
}

// DebugPanel 调试面板
type DebugPanel struct {
	middleware *DebugMiddleware
}

// NewDebugPanel 创建调试面板
func NewDebugPanel(middleware *DebugMiddleware) *DebugPanel {
	return &DebugPanel{
		middleware: middleware,
	}
}

// RegisterRoutes 注册调试路由
func (dp *DebugPanel) RegisterRoutes(engine any) {
	// 类型断言，支持不同的引擎类型
	var debugGroup *route.RouterGroup

	// 尝试不同的类型断言
	if h, ok := engine.(*route.Engine); ok {
		debugGroup = h.Group("/debug")
	} else {
		log.Println("无法注册调试路由，未知引擎类型")
		return // 如果类型不匹配，直接返回
	}

	// 注册路由的通用方法
	registerRoute := func(method, path string, handler func(ctx context.Context, c *app.RequestContext)) {
		switch method {
		case "GET":
			debugGroup.GET(path, handler)
		case "POST":
			debugGroup.POST(path, handler)
		case "PUT":
			debugGroup.PUT(path, handler)
		case "DELETE":
			debugGroup.DELETE(path, handler)
		case "PATCH":
			debugGroup.PATCH(path, handler)
		case "HEAD":
			debugGroup.HEAD(path, handler)
		case "OPTIONS":
			debugGroup.OPTIONS(path, handler)
		default:
			log.Printf("不支持的HTTP方法: %s", method)
		}
	}

	// 获取所有调试信息
	registerRoute("GET", "/requests", dp.getAllRequests)

	// 获取特定请求的调试信息
	registerRoute("GET", "/requests/:id", dp.getRequest)

	// 清空调试信息
	registerRoute("DELETE", "/requests", dp.clearRequests)

	// 调试面板页面
	registerRoute("GET", "/panel", dp.debugPanel)

	// 启用/禁用调试
	registerRoute("POST", "/toggle", dp.toggleDebug)
}

// getAllRequests 获取所有请求
func (dp *DebugPanel) getAllRequests(ctx context.Context, c *app.RequestContext) {
	requests := dp.middleware.storage.GetAll()
	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": requests,
	})
}

// getRequest 获取特定请求
func (dp *DebugPanel) getRequest(ctx context.Context, c *app.RequestContext) {
	requestID := c.Param("id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求ID不能为空",
		})
		return
	}

	info, exists := dp.middleware.storage.Get(requestID)
	if !exists {
		c.JSON(http.StatusNotFound, map[string]any{
			"code":    404,
			"message": "请求信息不存在",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": info,
	})
}

// clearRequests 清空请求
func (dp *DebugPanel) clearRequests(ctx context.Context, c *app.RequestContext) {
	dp.middleware.storage.Clear()
	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "调试信息已清空",
	})
}

// debugPanel 调试面板页面
func (dp *DebugPanel) debugPanel(ctx context.Context, c *app.RequestContext) {
	html := `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YYHertz 调试面板</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .header { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .controls { margin-bottom: 20px; }
        .btn { padding: 8px 16px; margin-right: 10px; border: none; border-radius: 4px; cursor: pointer; }
        .btn-primary { background: #007bff; color: white; }
        .btn-danger { background: #dc3545; color: white; }
        .requests { background: white; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .request-item { padding: 15px; border-bottom: 1px solid #eee; cursor: pointer; }
        .request-item:hover { background: #f8f9fa; }
        .request-method { display: inline-block; padding: 2px 8px; border-radius: 3px; font-size: 12px; font-weight: bold; }
        .GET { background: #28a745; color: white; }
        .POST { background: #007bff; color: white; }
        .PUT { background: #ffc107; color: black; }
        .DELETE { background: #dc3545; color: white; }
        .request-details { display: none; margin-top: 10px; padding: 10px; background: #f8f9fa; border-radius: 4px; }
        pre { background: #f1f1f1; padding: 10px; border-radius: 4px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>YYHertz 调试面板</h1>
        <div class="controls">
            <button class="btn btn-primary" onclick="loadRequests()">刷新</button>
            <button class="btn btn-danger" onclick="clearRequests()">清空</button>
            <button class="btn" onclick="toggleDebug()">切换调试状态</button>
        </div>
    </div>
    
    <div class="requests" id="requests">
        <div style="padding: 20px; text-align: center; color: #666;">
            加载中...
        </div>
    </div>

    <script>
        function loadRequests() {
            fetch('/debug/requests')
                .then(response => response.json())
                .then(data => {
                    const container = document.getElementById('requests');
                    if (Object.keys(data.data).length === 0) {
                        container.innerHTML = '<div style="padding: 20px; text-align: center; color: #666;">暂无调试信息</div>';
                        return;
                    }
                    
                    let html = '';
                    Object.values(data.data).forEach(request => {
                        html += '<div class="request-item" onclick="toggleDetails(\'' + request.request_id + '\')">' +
                            '<div>' +
                                '<span class="request-method ' + request.method + '">' + request.method + '</span>' +
                                '<strong>' + request.path + '</strong>' +
                                '<span style="float: right; color: #666;">' + new Date(request.timestamp).toLocaleString() + '</span>' +
                            '</div>' +
                            '<div style="margin-top: 5px; font-size: 14px; color: #666;">' +
                                '状态: ' + request.status_code + ' | 耗时: ' + request.duration + ' | 内存: ' + Math.round(request.memory.alloc/1024) + 'KB' +
                            '</div>' +
                            '<div class="request-details" id="details-' + request.request_id + '">' +
                                '<h4>请求详情</h4>' +
                                '<p><strong>请求ID:</strong> ' + request.request_id + '</p>' +
                                '<p><strong>控制器:</strong> ' + (request.controller || 'N/A') + '</p>' +
                                '<p><strong>动作:</strong> ' + (request.action || 'N/A') + '</p>' +
                                '<h5>请求头</h5>' +
                                '<pre>' + JSON.stringify(request.headers, null, 2) + '</pre>' +
                                '<h5>查询参数</h5>' +
                                '<pre>' + JSON.stringify(request.query, null, 2) + '</pre>' +
                                '<h5>请求体</h5>' +
                                '<pre>' + (request.body || '无') + '</pre>' +
                                '<h5>响应</h5>' +
                                '<pre>' + (request.response || '无') + '</pre>' +
                                (request.errors && request.errors.length > 0 ? 
                                    '<h5>错误信息</h5><pre style="color: red;">' + request.errors.join('\\n') + '</pre>'
                                : '') +
                                (request.stack && request.stack.length > 0 ? 
                                    '<h5>堆栈跟踪</h5><pre>' + request.stack.map(frame => frame.function + ' (' + frame.file + ':' + frame.line + ')').join('\\n') + '</pre>'
                                : '') +
                            '</div>' +
                        '</div>';
                    });
                    
                    container.innerHTML = html;
                })
                .catch(error => {
                    console.error('加载请求失败:', error);
                    document.getElementById('requests').innerHTML = '<div style="padding: 20px; text-align: center; color: red;">加载失败</div>';
                });
        }
        
        function toggleDetails(requestId) {
            const details = document.getElementById('details-' + requestId);
            if (details.style.display === 'none' || details.style.display === '') {
                details.style.display = 'block';
            } else {
                details.style.display = 'none';
            }
        }
        
        function clearRequests() {
            if (confirm('确定要清空所有调试信息吗？')) {
                fetch('/debug/requests', { method: 'DELETE' })
                    .then(response => response.json())
                    .then(data => {
                        alert('调试信息已清空');
                        loadRequests();
                    })
                    .catch(error => {
                        console.error('清空失败:', error);
                        alert('清空失败');
                    });
            }
        }
        
        function toggleDebug() {
            fetch('/debug/toggle', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    alert('调试状态已切换');
                })
                .catch(error => {
                    console.error('切换失败:', error);
                    alert('切换失败');
                });
        }
        
        // 页面加载时自动加载请求
        window.onload = function() {
            loadRequests();
            // 每5秒自动刷新
            setInterval(loadRequests, 5000);
        };
    </script>
</body>
</html>`

	c.SetContentType("text/html; charset=utf-8")
	c.WriteString(html)
}

// toggleDebug 切换调试状态
func (dp *DebugPanel) toggleDebug(ctx context.Context, c *app.RequestContext) {
	if dp.middleware.IsEnabled() {
		dp.middleware.Disable()
	} else {
		dp.middleware.Enable()
	}

	c.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "调试状态已切换",
		"enabled": dp.middleware.IsEnabled(),
	})
}
