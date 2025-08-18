package context

import (
	"encoding/xml"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"gopkg.in/yaml.v3"
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

// BindYAML 绑定YAML数据到结构体
func (ctx *Context) BindYAML(obj any) error {
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

	return yaml.Unmarshal(body, obj)
}

// BindProtobuf 绑定Protobuf数据到消息对象
func (ctx *Context) BindProtobuf(obj any) error {
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

	// 检查对象是否实现了proto.Message接口
	if protoMsg, ok := obj.(proto.Message); ok {
		return proto.Unmarshal(body, protoMsg)
	}
	
	return &BindingError{
		Type:    "BIND_ERROR",
		Message: "BindProtobuf target must implement proto.Message interface",
	}
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
	case strings.Contains(contentType, "yaml"), strings.Contains(contentType, "yml"):
		return ctx.BindYAML(obj)
	case strings.Contains(contentType, "protobuf"), strings.Contains(contentType, "x-protobuf"):
		return ctx.BindProtobuf(obj)
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

// ShouldBindYAML 安全YAML绑定
func (ctx *Context) ShouldBindYAML(obj any) error {
	return ctx.BindYAML(obj)
}

// ShouldBindProtobuf 安全Protobuf绑定
func (ctx *Context) ShouldBindProtobuf(obj any) error {
	return ctx.BindProtobuf(obj)
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

// // ParseForm 解析表单数据
// func (ctx *Context) ParseForm() error {
// 	if !ctx.ensureRequest() {
// 		return ErrRequestNotFound
// 	}

// 	// Hertz的RequestContext会自动解析表单数据
// 	// 这里主要是为了兼容性
// 	return nil
// }

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
	if obj == nil {
		return nil
	}

	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &BindingError{
			Type:    "BIND_ERROR",
			Message: "BindHeader target must be a non-nil pointer to struct",
		}
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return &BindingError{
			Type:    "BIND_ERROR",
			Message: "BindHeader target must be a struct",
		}
	}

	rt := rv.Type()

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		// 仅处理可导出字段
		if sf.PkgPath != "" {
			continue
		}

		// 嵌入匿名结构体，递归处理
		if sf.Anonymous && sf.Type.Kind() == reflect.Struct {
			if err := ctx.bindHeadersToStruct(rv.Field(i).Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		// 读取 header tag
		headerName := sf.Tag.Get("header")
		if headerName == "-" {
			continue
		}
		if headerName == "" {
			// 默认使用字段名
			headerName = sf.Name
		}

		val := ctx.Header(headerName)
		if val == "" {
			// 没有该请求头则跳过
			continue
		}

		fv := rv.Field(i)
		if !fv.CanSet() {
			continue
		}

		// 支持指针类型：若为nil则先分配
		target := fv
		if fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				elem := reflect.New(fv.Type().Elem())
				fv.Set(elem)
			}
			target = fv.Elem()
		}

		// 根据字段类型设置值（简化实现）
		switch target.Kind() {
		case reflect.String:
			target.SetString(val)
		case reflect.Bool:
			if b, ok := parseBool(val); ok {
				target.SetBool(b)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if iv, ok := parseInt64(val); ok {
				if target.OverflowInt(iv) {
					// 溢出则跳过
					continue
				}
				target.SetInt(iv)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if iv, ok := parseInt64(val); ok && iv >= 0 {
				u := uint64(iv)
				if target.OverflowUint(u) {
					continue
				}
				target.SetUint(u)
			}
		case reflect.Float32, reflect.Float64:
			if fv, ok := parseFloat64(val); ok {
				if target.OverflowFloat(fv) {
					continue
				}
				target.SetFloat(fv)
			}
		case reflect.Slice:
			// 简化：以逗号分隔
			if target.Type().Elem().Kind() == reflect.String {
				parts := splitAndTrim(val, ",")
				s := reflect.MakeSlice(target.Type(), len(parts), len(parts))
				for i := range parts {
					s.Index(i).SetString(parts[i])
				}
				target.Set(s)
			}
		// 其他复杂类型不处理
		default:
			// 跳过不支持的类型
			continue
		}
	}

	return nil
}

// 工具函数：按分隔符切分并去空白
func splitAndTrim(s, sep string) []string {
	if s == "" {
		return nil
	}
	raw := strings.Split(s, sep)
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
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
