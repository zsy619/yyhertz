// Package gin - 增强渲染系统
// 提供高性能的模板渲染、缓存和优化功能
package gin

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"html/template"
	
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/gin/render"
)

// =============================================================================
// 渲染器接口增强
// =============================================================================

// EnhancedRenderer 增强渲染器接口
type EnhancedRenderer interface {
	render.Render

	// 缓存相关
	SetCacheKey(key string) EnhancedRenderer
	GetCacheKey() string
	EnableCache() EnhancedRenderer
	DisableCache() EnhancedRenderer

	// 压缩相关
	EnableCompression() EnhancedRenderer
	DisableCompression() EnhancedRenderer

	// 性能监控
	GetRenderTime() time.Duration
	GetCacheHit() bool
}

// =============================================================================
// 基础渲染器
// =============================================================================

// BaseRenderer 基础渲染器实现
type BaseRenderer struct {
	cacheKey           string
	cacheEnabled       bool
	compressionEnabled bool
	renderTime         time.Duration
	cacheHit           bool
}

// Render 实现基础渲染方法
func (r *BaseRenderer) Render(c *app.RequestContext) error {
	// 基础渲染器的默认实现
	return nil
}

// WriteContentType 实现内容类型写入
func (r *BaseRenderer) WriteContentType(c *app.RequestContext) {
	// 基础渲染器的默认实现
}

// SetCacheKey 设置缓存键
func (r *BaseRenderer) SetCacheKey(key string) EnhancedRenderer {
	r.cacheKey = key
	return r
}

// GetCacheKey 获取缓存键
func (r *BaseRenderer) GetCacheKey() string {
	return r.cacheKey
}

// EnableCache 启用缓存
func (r *BaseRenderer) EnableCache() EnhancedRenderer {
	r.cacheEnabled = true
	return r
}

// DisableCache 禁用缓存
func (r *BaseRenderer) DisableCache() EnhancedRenderer {
	r.cacheEnabled = false
	return r
}

// EnableCompression 启用压缩
func (r *BaseRenderer) EnableCompression() EnhancedRenderer {
	r.compressionEnabled = true
	return r
}

// DisableCompression 禁用压缩
func (r *BaseRenderer) DisableCompression() EnhancedRenderer {
	r.compressionEnabled = false
	return r
}

// GetRenderTime 获取渲染时间
func (r *BaseRenderer) GetRenderTime() time.Duration {
	return r.renderTime
}

// GetCacheHit 获取缓存命中状态
func (r *BaseRenderer) GetCacheHit() bool {
	return r.cacheHit
}

// =============================================================================
// HTML渲染器增强
// =============================================================================

// EnhancedHTMLRenderer 增强HTML渲染器
type EnhancedHTMLRenderer struct {
	BaseRenderer
	Template *template.Template
	Name     string
	Data     any

	// 模板选项
	Funcs      template.FuncMap
	Layout     string
	Partials   []string
	MinifyHTML bool
}

// Render 实现渲染接口
func (r *EnhancedHTMLRenderer) Render(c *app.RequestContext) error {
	start := time.Now()
	defer func() {
		r.renderTime = time.Since(start)
	}()

	// 检查缓存
	if r.cacheEnabled && r.cacheKey != "" {
		if cached := renderCache.Get(r.cacheKey); cached != nil {
			r.cacheHit = true
			return r.writeWithCompression(c, cached)
		}
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := r.Template.ExecuteTemplate(&buf, r.Name, r.Data); err != nil {
		return err
	}

	result := buf.Bytes()

	// HTML压缩
	if r.MinifyHTML {
		result = minifyHTML(result)
	}

	// 缓存结果
	if r.cacheEnabled && r.cacheKey != "" {
		renderCache.Set(r.cacheKey, result, 5*time.Minute)
	}

	return r.writeWithCompression(c, result)
}

// WriteContentType 设置内容类型
func (r *EnhancedHTMLRenderer) WriteContentType(c *app.RequestContext) {
	c.Header("Content-Type", "text/html; charset=utf-8")
}

// writeWithCompression 带压缩的写入
func (r *BaseRenderer) writeWithCompression(c *app.RequestContext, data []byte) error {
	if r.compressionEnabled {
		// 设置压缩头
		c.Header("Content-Encoding", "gzip")
		
		// 创建gzip写入器
		var buf bytes.Buffer
		gzw := gzip.NewWriter(&buf)
		defer gzw.Close()
		
		_, err := gzw.Write(data)
		if err != nil {
			return err
		}
		
		gzw.Close()
		c.Write(buf.Bytes())
		return nil
	}

	c.Write(data)
	return nil
}

// =============================================================================
// JSON渲染器增强
// =============================================================================

// EnhancedJSONRenderer 增强JSON渲染器
type EnhancedJSONRenderer struct {
	BaseRenderer
	Data   any
	Indent string
	Prefix string

	// JSON选项
	HTMLEscape    bool
	SortKeys      bool
	CustomEncoder func(any) ([]byte, error)
}

// Render 实现渲染接口
func (r *EnhancedJSONRenderer) Render(c *app.RequestContext) error {
	start := time.Now()
	defer func() {
		r.renderTime = time.Since(start)
	}()

	// 检查缓存
	if r.cacheEnabled && r.cacheKey != "" {
		if cached := renderCache.Get(r.cacheKey); cached != nil {
			r.cacheHit = true
			return r.writeWithCompression(c, cached)
		}
	}

	// 编码JSON
	var result []byte
	var err error

	if r.CustomEncoder != nil {
		result, err = r.CustomEncoder(r.Data)
	} else {
		if r.Indent != "" {
			result, err = json.MarshalIndent(r.Data, r.Prefix, r.Indent)
		} else {
			result, err = json.Marshal(r.Data)
		}
	}

	if err != nil {
		return err
	}

	// HTML转义
	if r.HTMLEscape {
		var buf bytes.Buffer
		json.HTMLEscape(&buf, result)
		result = buf.Bytes()
	}

	// 缓存结果
	if r.cacheEnabled && r.cacheKey != "" {
		renderCache.Set(r.cacheKey, result, 5*time.Minute)
	}

	return r.writeWithCompression(c, result)
}

// WriteContentType 设置内容类型
func (r *EnhancedJSONRenderer) WriteContentType(c *app.RequestContext) {
	c.Header("Content-Type", "application/json; charset=utf-8")
}

// =============================================================================
// 渲染缓存
// =============================================================================

// RenderCache 渲染缓存接口
type RenderCache interface {
	Get(key string) []byte
	Set(key string, value []byte, expiration time.Duration)
	Delete(key string)
	Clear()
	Stats() CacheStats
}

// CacheStats 缓存统计
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Size        int       `json:"size"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	mu      sync.RWMutex
	items   map[string]*cacheItem
	stats   CacheStats
	maxSize int
	cleanup time.Duration
}

type cacheItem struct {
	data       []byte
	expiration time.Time
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache(maxSize int, cleanup time.Duration) *MemoryCache {
	cache := &MemoryCache{
		items:   make(map[string]*cacheItem),
		maxSize: maxSize,
		cleanup: cleanup,
	}

	// 启动清理协程
	go cache.cleanupLoop()

	return cache
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists || time.Now().After(item.expiration) {
		c.stats.Misses++
		return nil
	}

	c.stats.Hits++
	return item.data
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, value []byte, expiration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查大小限制
	if len(c.items) >= c.maxSize {
		c.evictOldest()
	}

	c.items[key] = &cacheItem{
		data:       value,
		expiration: time.Now().Add(expiration),
	}
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*cacheItem)
}

// Stats 获取统计信息
func (c *MemoryCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	stats.Size = len(c.items)
	return stats
}

// evictOldest 淘汰最旧的缓存项
func (c *MemoryCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestKey == "" || item.expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiration
		}
	}

	if oldestKey != "" {
		delete(c.items, oldestKey)
	}
}

// cleanupLoop 清理循环
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(c.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanupExpired()
	}
}

// cleanupExpired 清理过期项
func (c *MemoryCache) cleanupExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			delete(c.items, key)
		}
	}

	c.stats.LastCleanup = now
}

// 全局渲染缓存
var renderCache RenderCache = NewMemoryCache(1000, 5*time.Minute)

// =============================================================================
// 模板管理器
// =============================================================================

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*template.Template
	mu        sync.RWMutex

	// 配置
	BaseDir    string
	Extension  string
	Funcs      template.FuncMap
	AutoReload bool

	// 监控
	loadTimes map[string]time.Time
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager(baseDir string) *TemplateManager {
	return &TemplateManager{
		templates:  make(map[string]*template.Template),
		loadTimes:  make(map[string]time.Time),
		BaseDir:    baseDir,
		Extension:  ".html",
		AutoReload: false,
	}
}

// LoadTemplate 加载模板
func (tm *TemplateManager) LoadTemplate(name string) (*template.Template, error) {
	tm.mu.RLock()
	tmpl, exists := tm.templates[name]
	loadTime := tm.loadTimes[name]
	tm.mu.RUnlock()

	// 检查是否需要重新加载
	if exists && !tm.AutoReload {
		return tmpl, nil
	}

	// 构建文件路径
	fileName := filepath.Join(tm.BaseDir, name+tm.Extension)

	// 检查文件修改时间
	if exists && tm.AutoReload {
		// TODO: 检查文件修改时间
		_ = loadTime
	}

	// 加载模板
	tmpl, err := template.New(name).Funcs(tm.Funcs).ParseFiles(fileName)
	if err != nil {
		return nil, err
	}

	// 缓存模板
	tm.mu.Lock()
	tm.templates[name] = tmpl
	tm.loadTimes[name] = time.Now()
	tm.mu.Unlock()

	return tmpl, nil
}

// GetTemplate 获取模板
func (tm *TemplateManager) GetTemplate(name string) (*template.Template, error) {
	return tm.LoadTemplate(name)
}

// ClearCache 清除模板缓存
func (tm *TemplateManager) ClearCache() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.templates = make(map[string]*template.Template)
	tm.loadTimes = make(map[string]time.Time)
}

// 全局模板管理器
var defaultTemplateManager = NewTemplateManager("templates")

// =============================================================================
// 渲染工厂
// =============================================================================

// RenderFactory 渲染工厂
type RenderFactory struct {
	htmlCache   bool
	jsonCache   bool
	compression bool
	htmlMinify  bool
	templateMgr *TemplateManager
}

// NewRenderFactory 创建渲染工厂
func NewRenderFactory() *RenderFactory {
	return &RenderFactory{
		templateMgr: defaultTemplateManager,
	}
}

// EnableHTMLCache 启用HTML缓存
func (rf *RenderFactory) EnableHTMLCache() *RenderFactory {
	rf.htmlCache = true
	return rf
}

// EnableJSONCache 启用JSON缓存
func (rf *RenderFactory) EnableJSONCache() *RenderFactory {
	rf.jsonCache = true
	return rf
}

// EnableCompression 启用压缩
func (rf *RenderFactory) EnableCompression() *RenderFactory {
	rf.compression = true
	return rf
}

// EnableHTMLMinify 启用HTML压缩
func (rf *RenderFactory) EnableHTMLMinify() *RenderFactory {
	rf.htmlMinify = true
	return rf
}

// HTML 创建HTML渲染器
func (rf *RenderFactory) HTML(name string, data any) EnhancedRenderer {
	tmpl, err := rf.templateMgr.GetTemplate(name)
	if err != nil {
		// 返回错误渲染器
		return &ErrorRenderer{Error: err}
	}

	renderer := &EnhancedHTMLRenderer{
		BaseRenderer: BaseRenderer{
			cacheEnabled:       rf.htmlCache,
			compressionEnabled: rf.compression,
		},
		Template:   tmpl,
		Name:       name,
		Data:       data,
		MinifyHTML: rf.htmlMinify,
	}

	return renderer
}

// JSON 创建JSON渲染器
func (rf *RenderFactory) JSON(data any) EnhancedRenderer {
	renderer := &EnhancedJSONRenderer{
		BaseRenderer: BaseRenderer{
			cacheEnabled:       rf.jsonCache,
			compressionEnabled: rf.compression,
		},
		Data: data,
	}

	return renderer
}

// =============================================================================
// 错误渲染器
// =============================================================================

// ErrorRenderer 错误渲染器
type ErrorRenderer struct {
	Error error
}

// Render 实现渲染接口
func (r *ErrorRenderer) Render(c *app.RequestContext) error {
	return r.Error
}

// WriteContentType 设置内容类型
func (r *ErrorRenderer) WriteContentType(c *app.RequestContext) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
}

// 实现EnhancedRenderer接口
func (r *ErrorRenderer) SetCacheKey(key string) EnhancedRenderer { return r }
func (r *ErrorRenderer) GetCacheKey() string                     { return "" }
func (r *ErrorRenderer) EnableCache() EnhancedRenderer           { return r }
func (r *ErrorRenderer) DisableCache() EnhancedRenderer          { return r }
func (r *ErrorRenderer) EnableCompression() EnhancedRenderer     { return r }
func (r *ErrorRenderer) DisableCompression() EnhancedRenderer    { return r }
func (r *ErrorRenderer) GetRenderTime() time.Duration            { return 0 }
func (r *ErrorRenderer) GetCacheHit() bool                       { return false }

// =============================================================================
// Context渲染方法增强
// =============================================================================

// RenderEnhanced 使用增强渲染器渲染
func (c *Context) RenderEnhanced(code int, r EnhancedRenderer) {
	// 设置状态码
	c.Status(code)

	// 设置内容类型
	if render, ok := r.(render.Render); ok {
		render.WriteContentType(c.RequestContext)
	}

	// 设置缓存头
	if r.GetCacheKey() != "" && r.GetCacheHit() {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}

	// 设置压缩头
	if c.Request().Header.Get("Accept-Encoding") != "" {
		if strings.Contains(c.Request().Header.Get("Accept-Encoding"), "gzip") {
			c.Header("Content-Encoding", "gzip")
		}
	}

	// 渲染内容
	if err := r.Render(c.RequestContext); err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// 设置性能头
	if renderTime := r.GetRenderTime(); renderTime > 0 {
		c.Header("X-Render-Time", renderTime.String())
	}
}

// HTMLEnhanced 增强HTML渲染
func (c *Context) HTMLEnhanced(code int, name string, data any) {
	renderer := defaultRenderFactory.HTML(name, data)
	c.RenderEnhanced(code, renderer)
}

// JSONEnhanced 增强JSON渲染
func (c *Context) JSONEnhanced(code int, data any) {
	renderer := defaultRenderFactory.JSON(data)
	c.RenderEnhanced(code, renderer)
}

// JSONIndentEnhanced 增强缩进JSON渲染
func (c *Context) JSONIndentEnhanced(code int, data any, indent string) {
	renderer := &EnhancedJSONRenderer{
		BaseRenderer: BaseRenderer{
			cacheEnabled:       defaultRenderFactory.jsonCache,
			compressionEnabled: defaultRenderFactory.compression,
		},
		Data:   data,
		Indent: indent,
	}
	c.RenderEnhanced(code, renderer)
}

// 默认渲染工厂
var defaultRenderFactory = NewRenderFactory()

// =============================================================================
// HTML压缩工具
// =============================================================================

// minifyHTML 简单的HTML压缩
func minifyHTML(data []byte) []byte {
	// 简化实现：去除多余空格和换行
	content := string(data)

	// 去除多余的空白字符
	content = strings.ReplaceAll(content, "\n", "")
	content = strings.ReplaceAll(content, "\t", "")

	// 去除标签间的多余空格
	for strings.Contains(content, "  ") {
		content = strings.ReplaceAll(content, "  ", " ")
	}

	return []byte(content)
}

// =============================================================================
// 渲染性能监控
// =============================================================================

// RenderMonitor 渲染性能监控
type RenderMonitor struct {
	mu            sync.RWMutex
	totalRequests int64
	cacheHits     int64
	totalTime     time.Duration
	avgTime       time.Duration
}

// Record 记录渲染性能
func (m *RenderMonitor) Record(renderTime time.Duration, cacheHit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests++
	m.totalTime += renderTime
	m.avgTime = m.totalTime / time.Duration(m.totalRequests)

	if cacheHit {
		m.cacheHits++
	}
}

// GetStats 获取统计信息
func (m *RenderMonitor) GetStats() map[string]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cacheHitRate := float64(0)
	if m.totalRequests > 0 {
		cacheHitRate = float64(m.cacheHits) / float64(m.totalRequests) * 100
	}

	return map[string]any{
		"total_requests": m.totalRequests,
		"cache_hits":     m.cacheHits,
		"cache_hit_rate": fmt.Sprintf("%.2f%%", cacheHitRate),
		"total_time":     m.totalTime.String(),
		"avg_time":       m.avgTime.String(),
	}
}

// Reset 重置统计
func (m *RenderMonitor) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests = 0
	m.cacheHits = 0
	m.totalTime = 0
	m.avgTime = 0
}

// 全局渲染监控器
var globalRenderMonitor = &RenderMonitor{}

// GetRenderStats 获取全局渲染统计
func GetRenderStats() map[string]any {
	return globalRenderMonitor.GetStats()
}

// GetCacheStats 获取缓存统计
func GetCacheStats() CacheStats {
	return renderCache.Stats()
}

// =============================================================================
// 渲染中间件
// =============================================================================

// RenderMonitorMiddleware 渲染监控中间件
func RenderMonitorMiddleware() HandlerFunc {
	return func(c *Context) {
		start := time.Now()

		c.Next()

		// 记录渲染性能
		renderTime := time.Since(start)
		cacheHit := c.GetHeader("X-Cache") == "HIT"
		globalRenderMonitor.Record(renderTime, cacheHit)
	}
}

// CompressionMiddleware 压缩中间件
func CompressionMiddleware() HandlerFunc {
	return func(c *Context) {
		// 检查客户端是否支持gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// 设置gzip写入器
		gz := gzip.NewWriter(c.Writer())
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// 替换写入器
		// 注意：这里需要实现一个gzip写入器包装器
		c.Next()
	}
}
