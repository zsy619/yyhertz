package core

import (
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/zsy619/yyhertz/framework/config"
)

// ============= HTTP缓存控制方法 =============

// SetETag 设置ETag头
func (c *BaseController) SetETag(etag string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to set ETag")
		return
	}
	// 确保ETag被引号包围
	if !strings.HasPrefix(etag, "\"") || !strings.HasSuffix(etag, "\"") {
		etag = "\"" + etag + "\""
	}
	c.SetHeader("ETag", etag)
}

// SetLastModified 设置Last-Modified头
func (c *BaseController) SetLastModified(t time.Time) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to set Last-Modified")
		return
	}
	c.SetHeader("Last-Modified", t.UTC().Format(time.RFC1123))
}

// SetCacheControl 设置Cache-Control头
func (c *BaseController) SetCacheControl(directive string) {
	c.SetHeader("Cache-Control", directive)
}

// SetMaxAge 设置缓存最大年龄（秒）
func (c *BaseController) SetMaxAge(seconds int) {
	c.SetCacheControl(fmt.Sprintf("max-age=%d", seconds))
}

// SetNoCache 设置禁用缓存
func (c *BaseController) SetNoCache() {
	c.SetCacheControl("no-cache, no-store, must-revalidate")
	c.SetHeader("Pragma", "no-cache")
	c.SetHeader("Expires", "0")
}

// SetPrivateCache 设置私有缓存（只允许客户端缓存）
func (c *BaseController) SetPrivateCache(maxAge int) {
	c.SetCacheControl(fmt.Sprintf("private, max-age=%d", maxAge))
}

// SetPublicCache 设置公共缓存（允许代理缓存）
func (c *BaseController) SetPublicCache(maxAge int) {
	c.SetCacheControl(fmt.Sprintf("public, max-age=%d", maxAge))
}

// ============= 条件请求处理方法 =============

// NotModified 返回304未修改响应
func (c *BaseController) NotModified() {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to return NotModified")
		return
	}
	c.Ctx.AbortWithStatus(consts.StatusNotModified)
}

// CheckIfNoneMatch 检查If-None-Match头（ETag条件）
func (c *BaseController) CheckIfNoneMatch(etag string) bool {
	if c.Ctx == nil {
		return false
	}
	
	ifNoneMatch := c.GetHeader("If-None-Match")
	if ifNoneMatch == "" {
		return false
	}
	
	// 确保ETag被引号包围
	if !strings.HasPrefix(etag, "\"") || !strings.HasSuffix(etag, "\"") {
		etag = "\"" + etag + "\""
	}
	
	// 检查是否匹配（支持多个ETag）
	etags := strings.Split(ifNoneMatch, ",")
	for _, tag := range etags {
		tag = strings.TrimSpace(tag)
		if tag == etag || tag == "*" {
			return true
		}
	}
	return false
}

// CheckIfModifiedSince 检查If-Modified-Since头（时间条件）
func (c *BaseController) CheckIfModifiedSince(lastModified time.Time) bool {
	if c.Ctx == nil {
		return true
	}
	
	ifModifiedSince := c.GetHeader("If-Modified-Since")
	if ifModifiedSince == "" {
		return true
	}
	
	since, err := time.Parse(time.RFC1123, ifModifiedSince)
	if err != nil {
		return true
	}
	
	// 比较时间（忽略毫秒）
	return lastModified.Truncate(time.Second).After(since.Truncate(time.Second))
}

// HandleConditionalRequest 处理条件请求
func (c *BaseController) HandleConditionalRequest(etag string, lastModified time.Time) bool {
	// 设置缓存相关头
	c.SetETag(etag)
	c.SetLastModified(lastModified)
	
	// 检查ETag条件
	if c.CheckIfNoneMatch(etag) {
		c.NotModified()
		return true
	}
	
	// 检查时间条件
	if !c.CheckIfModifiedSince(lastModified) {
		c.NotModified()
		return true
	}
	
	return false
}

// ============= 数据压缩方法 =============

// SetGzipResponse 设置Gzip压缩响应
func (c *BaseController) SetGzipResponse(data []byte) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	
	// 检查客户端是否支持gzip
	acceptEncoding := c.GetHeader("Accept-Encoding")
	if !strings.Contains(acceptEncoding, "gzip") {
		// 客户端不支持gzip，直接返回原始数据
		c.Ctx.Write(data)
		return nil
	}
	
	// 设置压缩相关头
	c.SetHeader("Content-Encoding", "gzip")
	c.SetHeader("Vary", "Accept-Encoding")
	
	// 压缩数据
	var compressedData strings.Builder
	gzipWriter := gzip.NewWriter(&compressedData)
	
	if _, err := gzipWriter.Write(data); err != nil {
		return fmt.Errorf("failed to compress data: %v", err)
	}
	
	if err := gzipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %v", err)
	}
	
	// 写入压缩后的数据
	c.Ctx.Write([]byte(compressedData.String()))
	return nil
}

// EnableGzipCompression 启用Gzip压缩
func (c *BaseController) EnableGzipCompression() {
	c.SetHeader("Content-Encoding", "gzip")
	c.SetHeader("Vary", "Accept-Encoding")
}

// ============= 内容哈希和版本控制方法 =============

// GenerateContentHash 生成内容MD5哈希作为ETag
func (c *BaseController) GenerateContentHash(content []byte) string {
	hash := md5.Sum(content)
	return fmt.Sprintf("%x", hash)
}

// GenerateVersionETag 生成版本化的ETag
func (c *BaseController) GenerateVersionETag(content []byte, version string) string {
	contentHash := c.GenerateContentHash(content)
	return fmt.Sprintf("%s-%s", version, contentHash)
}

// SetContentHashETag 根据内容生成并设置ETag
func (c *BaseController) SetContentHashETag(content []byte) {
	etag := c.GenerateContentHash(content)
	c.SetETag(etag)
}

// ============= 性能优化方法 =============

// SetExpires 设置Expires头
func (c *BaseController) SetExpires(t time.Time) {
	c.SetHeader("Expires", t.UTC().Format(time.RFC1123))
}

// SetExpiresFromNow 从现在开始设置过期时间
func (c *BaseController) SetExpiresFromNow(duration time.Duration) {
	expires := time.Now().Add(duration)
	c.SetExpires(expires)
}

// SetCacheForMinutes 设置缓存分钟数
func (c *BaseController) SetCacheForMinutes(minutes int) {
	c.SetMaxAge(minutes * 60)
	c.SetExpiresFromNow(time.Duration(minutes) * time.Minute)
}

// SetCacheForHours 设置缓存小时数
func (c *BaseController) SetCacheForHours(hours int) {
	c.SetCacheForMinutes(hours * 60)
}

// SetCacheForDays 设置缓存天数
func (c *BaseController) SetCacheForDays(days int) {
	c.SetCacheForHours(days * 24)
}

// ============= 客户端缓存策略方法 =============

// SetImmutableCache 设置不可变资源缓存（适用于版本化资源）
func (c *BaseController) SetImmutableCache() {
	// 设置1年的长期缓存，标记为不可变
	c.SetCacheControl("public, max-age=31536000, immutable")
}

// SetShortCache 设置短期缓存（适用于经常变化的内容）
func (c *BaseController) SetShortCache() {
	// 设置5分钟缓存，允许重新验证
	c.SetCacheControl("public, max-age=300, must-revalidate")
}

// SetMediumCache 设置中期缓存（适用于偶尔变化的内容）
func (c *BaseController) SetMediumCache() {
	// 设置1小时缓存，允许重新验证
	c.SetCacheControl("public, max-age=3600, must-revalidate")
}

// SetLongCache 设置长期缓存（适用于静态资源）
func (c *BaseController) SetLongCache() {
	// 设置30天缓存
	c.SetCacheControl("public, max-age=2592000")
}

// ============= 缓存验证辅助方法 =============

// IsCacheableMethod 检查HTTP方法是否可缓存
func (c *BaseController) IsCacheableMethod() bool {
	if c.Ctx == nil {
		return false
	}
	
	method := c.Ctx.Method()
	return method == "GET" || method == "HEAD"
}

// ShouldCache 检查是否应该缓存响应
func (c *BaseController) ShouldCache() bool {
	if !c.IsCacheableMethod() {
		return false
	}
	
	// 检查是否已经设置了no-cache
	cacheControl := c.GetResponseHeader("Cache-Control")
	if strings.Contains(strings.ToLower(cacheControl), "no-cache") {
		return false
	}
	
	return true
}

// GetClientCachePreference 获取客户端缓存偏好
func (c *BaseController) GetClientCachePreference() string {
	if c.Ctx == nil {
		return ""
	}
	
	cacheControl := c.GetHeader("Cache-Control")
	return cacheControl
}

// IsClientNoCacheRequest 检查客户端是否请求不使用缓存
func (c *BaseController) IsClientNoCacheRequest() bool {
	cacheControl := c.GetClientCachePreference()
	return strings.Contains(strings.ToLower(cacheControl), "no-cache") ||
		   strings.Contains(strings.ToLower(cacheControl), "max-age=0")
}

// ============= 性能监控方法 =============

// StartPerformanceTimer 开始性能计时
func (c *BaseController) StartPerformanceTimer() time.Time {
	return time.Now()
}

// EndPerformanceTimer 结束性能计时并设置响应头
func (c *BaseController) EndPerformanceTimer(start time.Time) time.Duration {
	duration := time.Since(start)
	
	// 设置服务器处理时间头（毫秒）
	c.SetHeader("X-Response-Time", fmt.Sprintf("%.2fms", float64(duration.Nanoseconds())/1e6))
	
	return duration
}

// SetPerformanceHeaders 设置性能相关的响应头
func (c *BaseController) SetPerformanceHeaders() {
	// DNS预取控制
	c.SetHeader("X-DNS-Prefetch-Control", "on")
	
	// 预加载关键资源
	c.SetHeader("X-Content-Type-Options", "nosniff")
	
	// 启用Keep-Alive
	c.SetHeader("Connection", "keep-alive")
}

// ============= 资源预加载方法 =============

// AddPreloadLink 添加资源预加载链接
func (c *BaseController) AddPreloadLink(href string, as string, crossorigin ...string) {
	link := fmt.Sprintf("<%s>; rel=preload; as=%s", href, as)
	
	if len(crossorigin) > 0 && crossorigin[0] != "" {
		link += fmt.Sprintf("; crossorigin=%s", crossorigin[0])
	}
	
	// 获取现有的Link头
	existingLinks := c.GetResponseHeader("Link")
	if existingLinks != "" {
		link = existingLinks + ", " + link
	}
	
	c.SetHeader("Link", link)
}

// AddPrefetchLink 添加资源预获取链接
func (c *BaseController) AddPrefetchLink(href string) {
	link := fmt.Sprintf("<%s>; rel=prefetch", href)
	
	// 获取现有的Link头
	existingLinks := c.GetResponseHeader("Link")
	if existingLinks != "" {
		link = existingLinks + ", " + link
	}
	
	c.SetHeader("Link", link)
}

// AddPreconnectLink 添加预连接链接
func (c *BaseController) AddPreconnectLink(href string, crossorigin ...bool) {
	link := fmt.Sprintf("<%s>; rel=preconnect", href)
	
	if len(crossorigin) > 0 && crossorigin[0] {
		link += "; crossorigin"
	}
	
	// 获取现有的Link头
	existingLinks := c.GetResponseHeader("Link")
	if existingLinks != "" {
		link = existingLinks + ", " + link
	}
	
	c.SetHeader("Link", link)
}

// ============= 服务端推送方法 =============

// ServerPush 服务端推送资源（HTTP/2）
func (c *BaseController) ServerPush(target string) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	
	// 设置服务端推送头
	c.SetHeader("Link", fmt.Sprintf("<%s>; rel=preload; nopush", target))
	
	config.Infof("Server push initiated for resource: %s", target)
	return nil
}