package context

import (
	"strconv"
	"sync"
	"unsafe"
)

// ============= 通用辅助函数 =============

// ensureRequest 检查RequestContext是否存在
// 所有需要访问Request的方法都应该先调用此方法
func (ctx *Context) ensureRequest() bool {
	return ctx.request != nil
}

// safeStringConvert 安全地将字节数组转换为字符串
// 对空字节数组返回空字符串，避免不必要的分配
func safeStringConvert(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// 使用unsafe进行零拷贝转换（仅在只读场景下安全）
	return *(*string)(unsafe.Pointer(&b))
}

// parseValueWithDefault 解析值并提供默认值
// 如果值为空字符串，返回默认值
func parseValueWithDefault(value, defaultValue string) string {
	if value != "" {
		return value
	}
	return defaultValue
}

// parseInt 安全地解析整数
// 返回解析结果和是否成功的标志
func parseInt(s string) (int, bool) {
	if s == "" {
		return 0, true // 空字符串视为0且成功
	}
	if val, err := strconv.Atoi(s); err == nil {
		return val, true
	}
	return 0, false
}

// parseInt64 安全地解析int64
func parseInt64(s string) (int64, bool) {
	if s == "" {
		return 0, true
	}
	if val, err := strconv.ParseInt(s, 10, 64); err == nil {
		return val, true
	}
	return 0, false
}

// parseFloat64 安全地解析float64
func parseFloat64(s string) (float64, bool) {
	if s == "" {
		return 0.0, true
	}
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val, true
	}
	return 0.0, false
}

// parseBool 安全地解析布尔值
func parseBool(s string) (bool, bool) {
	if s == "" {
		return false, true
	}
	if val, err := strconv.ParseBool(s); err == nil {
		return val, true
	}
	return false, false
}

// ============= 对象池 =============

// 为频繁使用的对象类型创建对象池
var (
	// 字符串切片池
	stringSlicePool = sync.Pool{
		New: func() any {
			return make([]string, 0, 8) // 预分配8个元素
		},
	}
)

// getStringSlice 从池中获取字符串切片
func getStringSlice() []string {
	return stringSlicePool.Get().([]string)
}

// putStringSlice 将字符串切片归还到池中
func putStringSlice(slice []string) {
	if slice != nil {
		slice = slice[:0] // 重置长度但保留容量
		stringSlicePool.Put(slice)
	}
}

// ============= 常量定义 =============

const (
	// HTTP方法常量
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
	MethodTrace   = "TRACE"
	MethodConnect = "CONNECT"

	// 常用的Content-Type常量
	ContentTypeJSON        = "application/json"
	ContentTypeXML         = "application/xml"
	ContentTypeHTML        = "text/html"
	ContentTypeText        = "text/plain"
	ContentTypeForm        = "application/x-www-form-urlencoded"
	ContentTypeMultipart   = "multipart/form-data"
	ContentTypeYAML        = "application/x-yaml"
	ContentTypeOctetStream = "application/octet-stream"

	// 常用的请求头常量
	HeaderContentType        = "Content-Type"
	HeaderContentLength      = "Content-Length"
	HeaderContentDisposition = "Content-Disposition"
	HeaderUserAgent          = "User-Agent"
	HeaderReferer            = "Referer"
	HeaderAuthorization      = "Authorization"
	HeaderAccept             = "Accept"
	HeaderAcceptEncoding     = "Accept-Encoding"
	HeaderAcceptLanguage     = "Accept-Language"
	HeaderCacheControl       = "Cache-Control"
	HeaderConnection         = "Connection"
	HeaderUpgrade            = "Upgrade"
	HeaderXRequestedWith     = "X-Requested-With"
	HeaderXForwardedFor      = "X-Forwarded-For"
	HeaderXRealIP            = "X-Real-IP"

	// HTTP状态码常量
	StatusOK                           = 200
	StatusCreated                      = 201
	StatusAccepted                     = 202
	StatusNoContent                    = 204
	StatusPartialContent               = 206
	StatusMovedPermanently             = 301
	StatusFound                        = 302
	StatusSeeOther                     = 303
	StatusNotModified                  = 304
	StatusTemporaryRedirect            = 307
	StatusPermanentRedirect            = 308
	StatusBadRequest                   = 400
	StatusUnauthorized                 = 401
	StatusForbidden                    = 403
	StatusNotFound                     = 404
	StatusMethodNotAllowed             = 405
	StatusRequestedRangeNotSatisfiable = 416
	StatusInternalServerError          = 500
	StatusNotImplemented               = 501
	StatusBadGateway                   = 502
	StatusServiceUnavailable           = 503
)

// ============= 实用工具函数 =============

// isEmptyString 检查字符串是否为空或只包含空白字符
func isEmptyString(s string) bool {
	return len(s) == 0 || len(s) == len(s)-len(s[:len(s)])
}

// firstNonEmpty 返回第一个非空字符串
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
