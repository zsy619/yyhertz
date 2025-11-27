package context

import (
	"github.com/cloudwego/hertz/pkg/app"
)

// ResponseWriter 响应写入器接口
type ResponseWriter interface {
	Header() map[string]string
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
	Status() int
	Size() int
	Written() bool
}

// responseWriter 响应写入器实现
type responseWriter struct {
	RequestContext *app.RequestContext
	status         int
	size           int
	written        bool
}

// Header 获取响应头
func (w *responseWriter) Header() map[string]string {
	headers := make(map[string]string)
	if w.RequestContext != nil {
		w.RequestContext.Response.Header.VisitAll(func(key, value []byte) {
			headers[string(key)] = string(value)
		})
	}
	return headers
}

// Write 写入响应数据
func (w *responseWriter) Write(data []byte) (int, error) {
	if w.RequestContext != nil {
		w.size += len(data)
		w.written = true
		return w.RequestContext.Write(data)
	}
	return 0, nil
}

// WriteHeader 设置状态码
func (w *responseWriter) WriteHeader(statusCode int) {
	if w.RequestContext != nil && !w.written {
		w.status = statusCode
		w.RequestContext.SetStatusCode(statusCode)
	}
}

// Status 获取状态码
func (w *responseWriter) Status() int {
	return w.status
}

// Size 获取响应大小
func (w *responseWriter) Size() int {
	return w.size
}

// Written 是否已写入
func (w *responseWriter) Written() bool {
	return w.written
}