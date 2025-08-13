package context

import (
	"net/url"
	"strconv"
	"strings"
)

// ============= URL和网络信息增强功能 =============

// BaseURL 获取基础URL（协议+主机名+端口）
// 返回类似 "https://example.com:8080" 的基础URL
func (ctx *Context) BaseURL() string {
	if !ctx.ensureRequest() {
		return ""
	}

	scheme := ctx.Protocol()
	host := ctx.Hostname()
	port := ctx.Port()

	baseURL := scheme + "://" + host
	
	// 只在非标准端口时添加端口号
	if port != "" {
		standardPorts := map[string]string{
			"http":  "80",
			"https": "443",
		}
		if standardPorts[scheme] != port {
			baseURL += ":" + port
		}
	}

	return baseURL
}

// OriginalURL 获取完整的原始请求URL
// 包括协议、主机名、端口、路径和查询参数
func (ctx *Context) OriginalURL() string {
	if !ctx.ensureRequest() {
		return ""
	}

	baseURL := ctx.BaseURL()
	path := ctx.Path()
	queryString := safeStringConvert(ctx.request.URI().QueryString())

	originalURL := baseURL + path
	if queryString != "" {
		originalURL += "?" + queryString
	}

	return originalURL
}

// Protocol 获取请求协议（http 或 https）
func (ctx *Context) Protocol() string {
	if !ctx.ensureRequest() {
		return "http"
	}

	// 检查是否为HTTPS
	if ctx.isSecureConnection() {
		return "https"
	}

	// 检查X-Forwarded-Proto头
	if proto := ctx.Header("X-Forwarded-Proto"); proto != "" {
		return strings.ToLower(proto)
	}

	// 检查X-Forwarded-SSL头
	if ssl := ctx.Header("X-Forwarded-SSL"); ssl == "on" {
		return "https"
	}

	// 检查X-Url-Scheme头
	if scheme := ctx.Header("X-Url-Scheme"); scheme != "" {
		return strings.ToLower(scheme)
	}

	return "http"
}

// Hostname 获取主机名（不包含端口）
func (ctx *Context) Hostname() string {
	if !ctx.ensureRequest() {
		return ""
	}

	host := ctx.Host()
	
	// 移除端口号
	if colonIndex := strings.LastIndex(host, ":"); colonIndex > 0 {
		// 检查是否为IPv6地址
		if !strings.HasPrefix(host, "[") {
			host = host[:colonIndex]
		}
	}

	return host
}

// Port 获取端口号
func (ctx *Context) Port() string {
	if !ctx.ensureRequest() {
		return ""
	}

	host := ctx.Host()

	// 检查Host头中的端口
	if colonIndex := strings.LastIndex(host, ":"); colonIndex > 0 {
		// 检查是否为IPv6地址
		if !strings.HasPrefix(host, "[") {
			return host[colonIndex+1:]
		}
	}

	// 检查X-Forwarded-Port头
	if port := ctx.Header("X-Forwarded-Port"); port != "" {
		return port
	}

	// 根据协议返回默认端口
	switch ctx.Protocol() {
	case "https":
		return "443"
	case "http":
		return "80"
	default:
		return "80"
	}
}

// Subdomains 获取子域名数组
// 例如: api.v1.example.com 返回 ["api", "v1"]
func (ctx *Context) Subdomains(offset ...int) []string {
	hostname := ctx.Hostname()
	if hostname == "" {
		return []string{}
	}

	// 默认偏移量为2 (去掉主域名的两部分，如 example.com)
	offsetValue := 2
	if len(offset) > 0 && offset[0] > 0 {
		offsetValue = offset[0]
	}

	parts := strings.Split(hostname, ".")
	if len(parts) <= offsetValue {
		return []string{}
	}

	// 返回子域名部分
	subdomains := parts[:len(parts)-offsetValue]
	
	// 反转数组以获得正确的顺序
	for i, j := 0, len(subdomains)-1; i < j; i, j = i+1, j-1 {
		subdomains[i], subdomains[j] = subdomains[j], subdomains[i]
	}

	return subdomains
}

// ============= IP地址和地理位置信息 =============

// IPs 获取客户端IP地址列表（包括代理链）
func (ctx *Context) IPs() []string {
	if !ctx.ensureRequest() {
		return []string{}
	}

	var ips []string

	// 检查X-Forwarded-For头
	if xForwardedFor := ctx.Header("X-Forwarded-For"); xForwardedFor != "" {
		for _, ip := range strings.Split(xForwardedFor, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				ips = append(ips, ip)
			}
		}
	}

	// 检查X-Real-IP头
	if xRealIP := ctx.Header("X-Real-IP"); xRealIP != "" {
		ips = append(ips, xRealIP)
	}

	// 添加直接连接的IP
	if clientIP := ctx.ClientIP(); clientIP != "" {
		ips = append(ips, clientIP)
	}

	// 去重
	return removeDuplicateIPs(ips)
}

// IsIPv4 检查客户端IP是否为IPv4
func (ctx *Context) IsIPv4() bool {
	ip := ctx.ClientIP()
	return ip != "" && !strings.Contains(ip, ":")
}

// IsIPv6 检查客户端IP是否为IPv6
func (ctx *Context) IsIPv6() bool {
	ip := ctx.ClientIP()
	return ip != "" && strings.Contains(ip, ":")
}

// IsLocalhost 检查是否为本地请求
func (ctx *Context) IsLocalhost() bool {
	hostname := ctx.Hostname()
	ip := ctx.ClientIP()
	
	return hostname == "localhost" ||
		hostname == "127.0.0.1" ||
		hostname == "::1" ||
		ip == "127.0.0.1" ||
		ip == "::1" ||
		strings.HasPrefix(ip, "192.168.") ||
		strings.HasPrefix(ip, "10.") ||
		strings.HasPrefix(ip, "172.16.") ||
		strings.HasPrefix(ip, "172.17.") ||
		strings.HasPrefix(ip, "172.18.") ||
		strings.HasPrefix(ip, "172.19.") ||
		strings.HasPrefix(ip, "172.20.") ||
		strings.HasPrefix(ip, "172.21.") ||
		strings.HasPrefix(ip, "172.22.") ||
		strings.HasPrefix(ip, "172.23.") ||
		strings.HasPrefix(ip, "172.24.") ||
		strings.HasPrefix(ip, "172.25.") ||
		strings.HasPrefix(ip, "172.26.") ||
		strings.HasPrefix(ip, "172.27.") ||
		strings.HasPrefix(ip, "172.28.") ||
		strings.HasPrefix(ip, "172.29.") ||
		strings.HasPrefix(ip, "172.30.") ||
		strings.HasPrefix(ip, "172.31.")
}

// ============= URL解析和构建 =============

// ParsedURL 获取解析后的URL对象
func (ctx *Context) ParsedURL() (*url.URL, error) {
	if !ctx.ensureRequest() {
		return nil, &ContextError{
			Code:    "REQUEST_NOT_FOUND",
			Message: "Request context not found",
		}
	}

	originalURL := ctx.OriginalURL()
	return url.Parse(originalURL)
}

// URLFor 构建指向指定路径的完整URL
func (ctx *Context) URLFor(path string, params ...map[string]string) string {
	baseURL := ctx.BaseURL()
	
	// 确保路径以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	fullURL := baseURL + path

	// 添加查询参数
	if len(params) > 0 && params[0] != nil {
		values := url.Values{}
		for key, value := range params[0] {
			values.Add(key, value)
		}
		if queryString := values.Encode(); queryString != "" {
			fullURL += "?" + queryString
		}
	}

	return fullURL
}

// AbsoluteURL 将相对URL转换为绝对URL
func (ctx *Context) AbsoluteURL(relativeURL string) string {
	if strings.HasPrefix(relativeURL, "http://") || strings.HasPrefix(relativeURL, "https://") {
		// 已经是绝对URL
		return relativeURL
	}

	baseURL := ctx.BaseURL()

	if strings.HasPrefix(relativeURL, "/") {
		// 相对于根路径
		return baseURL + relativeURL
	}

	// 相对于当前路径
	currentPath := ctx.Path()
	if strings.HasSuffix(currentPath, "/") {
		return baseURL + currentPath + relativeURL
	}

	// 移除文件名，保留目录
	lastSlash := strings.LastIndex(currentPath, "/")
	if lastSlash >= 0 {
		return baseURL + currentPath[:lastSlash+1] + relativeURL
	}

	return baseURL + "/" + relativeURL
}

// ============= 请求来源分析 =============

// IsReferrerSameDomain 检查Referer是否来自同一域名
func (ctx *Context) IsReferrerSameDomain() bool {
	referer := ctx.Referer()
	if referer == "" {
		return false
	}

	refererURL, err := url.Parse(referer)
	if err != nil {
		return false
	}

	currentHostname := ctx.Hostname()
	return refererURL.Hostname() == currentHostname
}

// GetReferrerDomain 获取Referer的域名
func (ctx *Context) GetReferrerDomain() string {
	referer := ctx.Referer()
	if referer == "" {
		return ""
	}

	refererURL, err := url.Parse(referer)
	if err != nil {
		return ""
	}

	return refererURL.Hostname()
}

// ============= 代理和负载均衡信息 =============

// IsBehindProxy 检查是否在代理或负载均衡器后面
func (ctx *Context) IsBehindProxy() bool {
	proxyHeaders := []string{
		"X-Forwarded-For",
		"X-Forwarded-Proto",
		"X-Forwarded-Host",
		"X-Real-IP",
		"X-Forwarded-Port",
		"X-Forwarded-SSL",
	}

	for _, header := range proxyHeaders {
		if ctx.Header(header) != "" {
			return true
		}
	}

	return false
}

// GetProxyInfo 获取代理信息
func (ctx *Context) GetProxyInfo() map[string]string {
	info := make(map[string]string)

	proxyHeaders := map[string]string{
		"forwarded_for":   "X-Forwarded-For",
		"forwarded_proto": "X-Forwarded-Proto",
		"forwarded_host":  "X-Forwarded-Host",
		"forwarded_port":  "X-Forwarded-Port",
		"real_ip":         "X-Real-IP",
		"forwarded_ssl":   "X-Forwarded-SSL",
	}

	for key, header := range proxyHeaders {
		if value := ctx.Header(header); value != "" {
			info[key] = value
		}
	}

	return info
}

// ============= 网络连接信息 =============

// ConnectionInfo 获取连接信息摘要
func (ctx *Context) ConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"protocol":        ctx.Protocol(),
		"hostname":        ctx.Hostname(),
		"port":           ctx.Port(),
		"client_ip":      ctx.ClientIP(),
		"client_ips":     ctx.IPs(),
		"is_secure":      ctx.IsSecure(),
		"is_localhost":   ctx.IsLocalhost(),
		"is_ipv4":        ctx.IsIPv4(),
		"is_ipv6":        ctx.IsIPv6(),
		"behind_proxy":   ctx.IsBehindProxy(),
		"user_agent":     ctx.UserAgent(),
		"referer":        ctx.Referer(),
		"original_url":   ctx.OriginalURL(),
		"base_url":       ctx.BaseURL(),
		"subdomains":     ctx.Subdomains(),
	}
}

// ============= 辅助函数 =============

// removeDuplicateIPs 去除重复的IP地址
func removeDuplicateIPs(ips []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, ip := range ips {
		if !keys[ip] {
			keys[ip] = true
			result = append(result, ip)
		}
	}

	return result
}

// isSecureConnection 检查是否为安全连接（HTTPS）
// 这是一个内部辅助方法，增强了现有的IsSecure方法的功能
func (ctx *Context) isSecureConnection() bool {
	if !ctx.ensureRequest() {
		return false
	}

	// 首先使用现有的IsSecure方法
	if ctx.IsSecure() {
		return true
	}

	// 检查额外的代理头
	if proto := ctx.Header("X-Forwarded-Proto"); proto == "https" {
		return true
	}

	if ssl := ctx.Header("X-Forwarded-SSL"); ssl == "on" {
		return true
	}

	if scheme := ctx.Header("X-Url-Scheme"); scheme == "https" {
		return true
	}

	return false
}

// PortInt 获取端口号（整数形式）
func (ctx *Context) PortInt() int {
	portStr := ctx.Port()
	if portStr == "" {
		return 0
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0
	}

	return port
}