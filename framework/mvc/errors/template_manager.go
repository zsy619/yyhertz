package errors

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"time"
)

// ============= 模板管理器实现 =============

// DefaultTemplateManager 默认模板管理器
type DefaultTemplateManager struct {
	templates    map[string]*template.Template
	mu           sync.RWMutex
	enableCache  bool
	hotReload    bool
	templateDir  string
	funcs        template.FuncMap
}

// NewDefaultTemplateManager 创建默认模板管理器
func NewDefaultTemplateManager() *DefaultTemplateManager {
	manager := &DefaultTemplateManager{
		templates:   make(map[string]*template.Template),
		enableCache: true,
		hotReload:   false,
		funcs:       make(template.FuncMap),
	}
	
	// 注册默认函数
	manager.registerDefaultFuncs()
	
	// 加载内置模板
	manager.loadBuiltinTemplates()
	
	return manager
}

// RenderTemplate 渲染模板
func (m *DefaultTemplateManager) RenderTemplate(name string, data interface{}) (string, error) {
	var tmpl *template.Template
	var exists bool
	
	if m.enableCache {
		m.mu.RLock()
		tmpl, exists = m.templates[name]
		m.mu.RUnlock()
	}
	
	if !exists || !m.enableCache {
		// 重新加载模板
		if err := m.reloadTemplate(name); err != nil {
			return "", fmt.Errorf("failed to load template %s: %w", name, err)
		}
		
		m.mu.RLock()
		tmpl, exists = m.templates[name]
		m.mu.RUnlock()
		
		if !exists {
			return "", fmt.Errorf("template %s not found", name)
		}
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}
	
	return buf.String(), nil
}

// LoadTemplate 加载模板
func (m *DefaultTemplateManager) LoadTemplate(name string, content string) error {
	tmpl, err := template.New(name).Funcs(m.funcs).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}
	
	m.mu.Lock()
	m.templates[name] = tmpl
	m.mu.Unlock()
	
	return nil
}

// ReloadTemplates 重新加载所有模板
func (m *DefaultTemplateManager) ReloadTemplates() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 清空现有模板
	m.templates = make(map[string]*template.Template)
	
	// 重新加载内置模板
	return m.loadBuiltinTemplates()
}

// SetTemplateDir 设置模板目录
func (m *DefaultTemplateManager) SetTemplateDir(dir string) {
	m.templateDir = dir
}

// EnableCache 启用/禁用缓存
func (m *DefaultTemplateManager) EnableCache(enable bool) {
	m.enableCache = enable
}

// EnableHotReload 启用/禁用热重载
func (m *DefaultTemplateManager) EnableHotReload(enable bool) {
	m.hotReload = enable
}

// RegisterFunc 注册模板函数
func (m *DefaultTemplateManager) RegisterFunc(name string, fn interface{}) {
	m.funcs[name] = fn
}

// reloadTemplate 重新加载单个模板
func (m *DefaultTemplateManager) reloadTemplate(name string) error {
	// 这里可以从文件系统或其他地方加载模板
	// 现在先从内置模板加载
	return m.loadBuiltinTemplate(name)
}

// loadBuiltinTemplates 加载内置模板
func (m *DefaultTemplateManager) loadBuiltinTemplates() error {
	templates := map[string]string{
		"error.html":     errorHTMLTemplate,
		"error.json":     errorJSONTemplate,
		"error.xml":      errorXMLTemplate,
		"error.minimal":  errorMinimalTemplate,
	}
	
	for name, content := range templates {
		if err := m.LoadTemplate(name, content); err != nil {
			return fmt.Errorf("failed to load builtin template %s: %w", name, err)
		}
	}
	
	return nil
}

// loadBuiltinTemplate 加载单个内置模板
func (m *DefaultTemplateManager) loadBuiltinTemplate(name string) error {
	switch name {
	case "error.html":
		return m.LoadTemplate(name, errorHTMLTemplate)
	case "error.json":
		return m.LoadTemplate(name, errorJSONTemplate)
	case "error.xml":
		return m.LoadTemplate(name, errorXMLTemplate)
	case "error.minimal":
		return m.LoadTemplate(name, errorMinimalTemplate)
	default:
		return fmt.Errorf("unknown builtin template: %s", name)
	}
}

// registerDefaultFuncs 注册默认模板函数
func (m *DefaultTemplateManager) registerDefaultFuncs() {
	m.funcs["formatTime"] = func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	}
	
	m.funcs["formatTimeUnix"] = func(t time.Time) int64 {
		return t.Unix()
	}
	
	m.funcs["join"] = func(slice []string, sep string) string {
		if len(slice) == 0 {
			return ""
		}
		result := slice[0]
		for i := 1; i < len(slice); i++ {
			result += sep + slice[i]
		}
		return result
	}
	
	m.funcs["safeHTML"] = func(s string) template.HTML {
		return template.HTML(s)
	}
	
	m.funcs["add"] = func(a, b int) int {
		return a + b
	}
	
	m.funcs["contains"] = func(slice []string, item string) bool {
		for _, s := range slice {
			if s == item {
				return true
			}
		}
		return false
	}
	
	m.funcs["default"] = func(value, defaultValue interface{}) interface{} {
		if value == nil {
			return defaultValue
		}
		return value
	}
}

// ============= 模板内容定义 =============

// errorHTMLTemplate HTML错误页面模板（简化版）
const errorHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - 错误页面</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 0; padding: 20px; background: #f5f5f5; line-height: 1.6; }
        .container { max-width: 800px; margin: 0 auto; }
        .header { background: white; padding: 30px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); text-align: center; }
        .error-card { background: white; padding: 30px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); border-left: 5px solid #dc3545; }
        .error-header { display: flex; align-items: center; margin-bottom: 20px; flex-wrap: wrap; }
        .status-icon { font-size: 48px; margin-right: 20px; }
        .status-info h1 { margin: 0; color: #333; font-size: 2.5em; }
        .status-info p { margin: 5px 0 0 0; color: #666; font-size: 1.2em; }
        .error-details { background: #f8f9fa; padding: 20px; border-radius: 6px; margin: 20px 0; }
        .suggestions { list-style: none; padding: 0; }
        .suggestions li { background: #e3f2fd; padding: 12px 16px; margin: 8px 0; border-radius: 4px; border-left: 4px solid #2196f3; }
        .actions { text-align: center; margin: 30px 0; }
        .btn { padding: 12px 24px; margin: 0 8px; border: none; border-radius: 4px; cursor: pointer; font-size: 16px; text-decoration: none; display: inline-block; transition: background-color 0.3s; }
        .btn-primary { background: #007bff; color: white; }
        .btn-secondary { background: #6c757d; color: white; }
        .footer { text-align: center; color: #6c757d; padding: 20px; font-size: 14px; }
        @media (max-width: 768px) {
            body { padding: 10px; }
            .header, .error-card { padding: 20px; }
            .error-header { flex-direction: column; text-align: center; }
            .status-icon { margin-right: 0; margin-bottom: 10px; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.FrameworkName | default "YYHertz Framework"}}</h1>
            <p>应用遇到了一个问题，请查看下面的详细信息</p>
        </div>
        
        <div class="error-card">
            <div class="error-header">
                <div class="status-icon">{{.Icon | default "❌"}}</div>
                <div class="status-info">
                    <h1>{{.StatusCode}} {{.StatusText}}</h1>
                    <p>{{.StatusText}}</p>
                </div>
            </div>
            
            <div class="error-message">
                <p style="font-size: 1.1em; color: #333; margin-bottom: 20px;">{{.ErrorMessage}}</p>
            </div>
            
            <div class="error-details">
                <h3>🔍 请求详情</h3>
                <p><strong>请求路径:</strong> {{.RequestPath}}</p>
                <p><strong>请求方法:</strong> {{.RequestMethod}}</p>
                <p><strong>时间:</strong> {{formatTime .Timestamp}}</p>
                {{if .RequestID}}<p><strong>请求ID:</strong> {{.RequestID}}</p>{{end}}
            </div>
            
            {{if .Suggestions}}
            <div>
                <h3>💡 解决建议</h3>
                <ul class="suggestions">
                    {{range .Suggestions}}
                    <li>{{.}}</li>
                    {{end}}
                </ul>
            </div>
            {{end}}
            
            <div class="actions">
                <a href="javascript:history.back()" class="btn btn-secondary">返回上页</a>
                <a href="/" class="btn btn-primary">返回首页</a>
            </div>
        </div>
        
        <div class="footer">
            <p>&copy; {{.FrameworkName | default "YYHertz Framework"}}. 技术支持团队随时为您服务。</p>
        </div>
    </div>
</body>
</html>`

// errorJSONTemplate JSON错误模板
const errorJSONTemplate = `{
    "code": {{.StatusCode}},
    "message": "{{.ErrorMessage}}",
    "success": false,
    "path": "{{.RequestPath}}",
    "method": "{{.RequestMethod}}",
    "timestamp": {{formatTimeUnix .Timestamp}}{{if .RequestID}},
    "request_id": "{{.RequestID}}"{{end}}{{if .Suggestions}},
    "suggestions": [{{range $i, $s := .Suggestions}}{{if $i}}, {{end}}"{{$s}}"{{end}}]{{end}}{{if .Details}},
    "details": {{.Details}}{{end}}
}`

// errorXMLTemplate XML错误模板
const errorXMLTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<error>
    <code>{{.StatusCode}}</code>
    <title>{{.StatusText}}</title>
    <message>{{.ErrorMessage}}</message>
    <path>{{.RequestPath}}</path>
    <method>{{.RequestMethod}}</method>
    <timestamp>{{formatTime .Timestamp}}</timestamp>{{if .RequestID}}
    <request_id>{{.RequestID}}</request_id>{{end}}{{if .Suggestions}}
    <suggestions>{{range .Suggestions}}
        <suggestion>{{.}}</suggestion>{{end}}
    </suggestions>{{end}}
</error>`

// errorMinimalTemplate 最小化模板（用于高性能场景）
const errorMinimalTemplate = `Error {{.StatusCode}}: {{.ErrorMessage}}`

// ============= 全局模板管理器 =============

var globalTemplateManager = NewDefaultTemplateManager()

// GetGlobalTemplateManager 获取全局模板管理器
func GetGlobalTemplateManager() TemplateManager {
	return globalTemplateManager
}

// RenderTemplate 渲染模板（全局方法）
func RenderTemplate(name string, data interface{}) (string, error) {
	return globalTemplateManager.RenderTemplate(name, data)
}

// ============= 模板数据结构 =============

// TemplateData 模板数据结构
type TemplateData struct {
	*ErrorContext
	
	// 扩展字段
	Title         string `json:"title"`
	FrameworkName string `json:"framework_name"`
	Icon          string `json:"icon"`
	ShowDebug     bool   `json:"show_debug"`
	Language      string `json:"language"`
	
	// 支持信息
	SupportEmail string `json:"support_email,omitempty"`
	SupportPhone string `json:"support_phone,omitempty"`
	
	// 配置信息
	EnableRetry   bool `json:"enable_retry"`
	RetryInterval int  `json:"retry_interval"`
}

// NewTemplateData 创建模板数据
func NewTemplateData(errorCtx *ErrorContext) *TemplateData {
	return &TemplateData{
		ErrorContext:  errorCtx,
		Title:         "错误页面",
		FrameworkName: "YYHertz Framework",
		Icon:          "❌",
		ShowDebug:     false,
		Language:      "zh-CN",
		EnableRetry:   false,
		RetryInterval: 5,
	}
}

// SetConfig 设置配置
func (td *TemplateData) SetConfig(config *StatusConfig) {
	if config != nil {
		td.Icon = config.Icon
		td.ShowDebug = config.ShowDetails
		td.EnableRetry = config.Retryable
	}
}

// SetFrameworkInfo 设置框架信息
func (td *TemplateData) SetFrameworkInfo(name, title string) {
	td.FrameworkName = name
	td.Title = title
}

// SetSupportInfo 设置支持信息
func (td *TemplateData) SetSupportInfo(email, phone string) {
	td.SupportEmail = email
	td.SupportPhone = phone
}

// ============= 模板缓存优化 =============

// TemplateCache 模板缓存
type TemplateCache struct {
	cache       map[string]CacheEntry
	mu          sync.RWMutex
	maxSize     int
	defaultTTL  time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	content   string
	timestamp time.Time
	ttl       time.Duration
	hits      int64
}

// NewTemplateCache 创建模板缓存
func NewTemplateCache(maxSize int, defaultTTL time.Duration) *TemplateCache {
	return &TemplateCache{
		cache:      make(map[string]CacheEntry),
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
	}
}

// Get 获取缓存内容
func (tc *TemplateCache) Get(key string) (string, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	
	entry, exists := tc.cache[key]
	if !exists {
		return "", false
	}
	
	// 检查是否过期
	if time.Since(entry.timestamp) > entry.ttl {
		// 异步清理过期条目
		go tc.cleanupExpired(key)
		return "", false
	}
	
	// 更新命中次数
	entry.hits++
	tc.cache[key] = entry
	
	return entry.content, true
}

// Set 设置缓存内容
func (tc *TemplateCache) Set(key, content string, ttl time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	if ttl <= 0 {
		ttl = tc.defaultTTL
	}
	
	// 如果缓存已满，移除最少使用的条目
	if len(tc.cache) >= tc.maxSize {
		tc.evictLRU()
	}
	
	tc.cache[key] = CacheEntry{
		content:   content,
		timestamp: time.Now(),
		ttl:       ttl,
		hits:      1,
	}
}

// evictLRU 驱逐最少使用的条目
func (tc *TemplateCache) evictLRU() {
	var minHits int64 = -1
	var evictKey string
	
	for key, entry := range tc.cache {
		if minHits == -1 || entry.hits < minHits {
			minHits = entry.hits
			evictKey = key
		}
	}
	
	if evictKey != "" {
		delete(tc.cache, evictKey)
	}
}

// cleanupExpired 清理过期条目
func (tc *TemplateCache) cleanupExpired(key string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	
	if entry, exists := tc.cache[key]; exists {
		if time.Since(entry.timestamp) > entry.ttl {
			delete(tc.cache, key)
		}
	}
}