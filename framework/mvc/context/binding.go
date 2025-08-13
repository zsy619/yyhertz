package context

import (
	"encoding/xml"
	"strings"
)

// ============= 核心绑定方法 =============

// BindJSON 解析JSON请求体
func (ctx *Context) BindJSON(obj any) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	return ctx.request.BindJSON(obj)
}

// BindXML 解析XML请求体
func (ctx *Context) BindXML(obj any) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	body, err := ctx.request.Body()
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return ErrEmptyBody
	}

	return xml.Unmarshal(body, obj)
}

// BindQuery 绑定查询参数到结构体
func (ctx *Context) BindQuery(obj any) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}
	return ctx.request.BindQuery(obj)
}

// BindForm 绑定表单数据到结构体
func (ctx *Context) BindForm(obj any) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}
	return ctx.request.Bind(obj)
}

// ============= 智能绑定方法 =============

// Bind 智能绑定 (根据Content-Type自动选择)
func (ctx *Context) Bind(obj any) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	contentType := ctx.ContentType()

	// 根据Content-Type选择绑定方式
	switch {
	case strings.Contains(contentType, ContentTypeJSON):
		return ctx.BindJSON(obj)
	case strings.Contains(contentType, "xml"):
		return ctx.BindXML(obj)
	case strings.Contains(contentType, ContentTypeForm):
		return ctx.BindForm(obj)
	case strings.Contains(contentType, ContentTypeMultipart):
		return ctx.BindForm(obj)
	default:
		// 默认尝试JSON绑定
		return ctx.BindJSON(obj)
	}
}

// ============= 安全绑定方法 =============

// ShouldBind 安全绑定 (不会在失败时终止)
func (ctx *Context) ShouldBind(obj any) error {
	return ctx.Bind(obj)
}

// ShouldBindJSON 安全JSON绑定
func (ctx *Context) ShouldBindJSON(obj any) error {
	return ctx.BindJSON(obj)
}

// ShouldBindXML 安全XML绑定
func (ctx *Context) ShouldBindXML(obj any) error {
	return ctx.BindXML(obj)
}

// ShouldBindQuery 安全查询参数绑定
func (ctx *Context) ShouldBindQuery(obj any) error {
	return ctx.BindQuery(obj)
}

// ShouldBindForm 安全表单绑定
func (ctx *Context) ShouldBindForm(obj any) error {
	return ctx.BindForm(obj)
}

// ============= 绑定验证结合方法 =============

// BindAndValidate 绑定数据并验证
func (ctx *Context) BindAndValidate(obj any) error {
	if err := ctx.Bind(obj); err != nil {
		return &BindingError{
			Type:    "BIND_ERROR",
			Message: "Failed to bind request data: " + err.Error(),
			Cause:   err,
		}
	}

	// 如果对象实现了Validator接口，进行验证
	if validator, ok := obj.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return &BindingError{
				Type:    "VALIDATION_ERROR",
				Message: "Validation failed: " + err.Error(),
				Cause:   err,
			}
		}
	}

	return nil
}

// ShouldBindAndValidate 安全的绑定和验证
func (ctx *Context) ShouldBindAndValidate(obj any) error {
	return ctx.BindAndValidate(obj)
}

// ============= 表单解析方法 =============

// ParseForm 解析表单数据
func (ctx *Context) ParseForm() error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// Hertz的RequestContext会自动解析表单数据
	// 这里主要是为了兼容性
	return nil
}

// ParseMultipartForm 解析多部分表单
func (ctx *Context) ParseMultipartForm(maxMemory int64) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// Hertz的RequestContext会自动处理multipart表单
	// 这里主要是为了兼容性
	return nil
}

// ============= 高级绑定功能 =============

// BindHeader 绑定请求头到结构体
func (ctx *Context) BindHeader(obj any) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// 使用反射将请求头绑定到结构体
	// 这是一个简化的实现，实际项目中可能需要更复杂的逻辑
	return ctx.bindHeadersToStruct(obj)
}

// BindUri 绑定URI参数到结构体
func (ctx *Context) BindUri(obj any) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// 将路由参数绑定到结构体
	return ctx.bindUriToStruct(obj)
}

// ============= 绑定辅助方法 =============

// bindHeadersToStruct 将请求头绑定到结构体（简化实现）
func (ctx *Context) bindHeadersToStruct(obj any) error {
	// TODO: 实现请求头到结构体的绑定
	// 这需要使用反射来分析结构体字段的标签
	return nil
}

// bindUriToStruct 将URI参数绑定到结构体（简化实现）
func (ctx *Context) bindUriToStruct(obj any) error {
	// TODO: 实现URI参数到结构体的绑定
	return nil
}

// ============= 验证接口定义 =============

// Validator 验证器接口
// 实现此接口的结构体可以在绑定后自动进行验证
type Validator interface {
	Validate() error
}

// ============= 错误类型定义 =============

// BindingError 绑定错误类型
type BindingError struct {
	Type    string
	Message string
	Cause   error
}

func (e *BindingError) Error() string {
	return e.Message
}

func (e *BindingError) Unwrap() error {
	return e.Cause
}

// ============= 内容检查方法 =============

// HasBody 检查请求是否有body
func (ctx *Context) HasBody() bool {
	if !ctx.ensureRequest() {
		return false
	}

	body, _ := ctx.request.Body()
	return len(body) > 0
}

// GetBodySize 获取请求体大小
func (ctx *Context) GetBodySize() int64 {
	if !ctx.ensureRequest() {
		return 0
	}

	body, _ := ctx.request.Body()
	return int64(len(body))
}

// GetBodyString 获取请求体字符串
func (ctx *Context) GetBodyString() (string, error) {
	body, err := ctx.RawBody()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ============= 高级验证功能 =============

// ValidateStruct 验证结构体
func (ctx *Context) ValidateStruct(obj any) error {
	if validator, ok := obj.(Validator); ok {
		return validator.Validate()
	}
	return nil
}

// ============= MIME类型检测 =============

// GetContentLength 获取Content-Length
func (ctx *Context) GetContentLength() int64 {
	if !ctx.ensureRequest() {
		return 0
	}

	contentLength := ctx.Header(HeaderContentLength)
	if length, ok := parseInt64(contentLength); ok {
		return length
	}
	return 0
}

// IsContentTypeSupported 检查Content-Type是否支持
func (ctx *Context) IsContentTypeSupported(supportedTypes ...string) bool {
	contentType := strings.ToLower(ctx.ContentType())
	
	for _, supported := range supportedTypes {
		if strings.Contains(contentType, strings.ToLower(supported)) {
			return true
		}
	}
	return false
}