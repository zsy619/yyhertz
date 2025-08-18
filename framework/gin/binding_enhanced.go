// Package gin - 增强数据绑定系统
// 提供强大的数据绑定、验证和转换功能
package gin

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// 绑定接口定义
// =============================================================================

// Binding 绑定接口
type Binding interface {
	Name() string
	Bind(*http.Request, any) error
}

// 注意：BindingBody 接口已在context_advanced.go中定义

// BindingURI URI绑定接口  
type BindingURI interface {
	Name() string
	BindURI(map[string][]string, any) error
}

// StructValidator 结构体验证器接口
type StructValidator interface {
	ValidateStruct(any) error
	Engine() any
}

// =============================================================================
// 绑定类型常量 (扩展)
// =============================================================================

const (
	// 注意：基础MIME常量已在context_advanced.go中定义
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
	MIMETOML              = "application/toml"
)

// =============================================================================
// JSON绑定实现
// =============================================================================

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *http.Request, obj any) error {
	if req == nil || req.Body == nil {
		return fmt.Errorf("invalid request")
	}
	return decodeJSON(req.Body, obj)
}

func (jsonBinding) BindBody(body []byte, obj any) error {
	return json.Unmarshal(body, obj)
}

func decodeJSON(body io.Reader, obj any) error {
	decoder := json.NewDecoder(body)
	if EnableDecoderUseNumber {
		decoder.UseNumber()
	}
	if EnableDecoderDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}

// =============================================================================
// XML绑定实现
// =============================================================================

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(req *http.Request, obj any) error {
	return decodeXML(req.Body, obj)
}

func (xmlBinding) BindBody(body []byte, obj any) error {
	return xml.Unmarshal(body, obj)
}

func decodeXML(body io.Reader, obj any) error {
	decoder := xml.NewDecoder(body)
	if err := decoder.Decode(obj); err != nil {
		return err
	}
	return validate(obj)
}

// =============================================================================
// Form绑定实现
// =============================================================================

type formBinding struct{}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := req.ParseMultipartForm(defaultMemory); err != nil && err != http.ErrNotMultipart {
		return err
	}
	if err := mapForm(obj, req.Form); err != nil {
		return err
	}
	return validate(obj)
}

// =============================================================================
// Query绑定实现
// =============================================================================

type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *http.Request, obj any) error {
	values := req.URL.Query()
	if err := mapForm(obj, values); err != nil {
		return err
	}
	return validate(obj)
}

// =============================================================================
// Header绑定实现
// =============================================================================

type headerBinding struct{}

func (headerBinding) Name() string {
	return "header"
}

func (headerBinding) Bind(req *http.Request, obj any) error {
	if err := mapHeader(obj, req.Header); err != nil {
		return err
	}
	return validate(obj)
}

// =============================================================================
// URI绑定实现
// =============================================================================

type uriBinding struct{}

func (uriBinding) Name() string {
	return "uri"
}

func (uriBinding) BindURI(m map[string][]string, obj any) error {
	if err := mapURI(obj, m); err != nil {
		return err
	}
	return validate(obj)
}

// =============================================================================
// Multipart绑定实现
// =============================================================================

type multipartBinding struct{}

func (multipartBinding) Name() string {
	return "multipart"
}

func (multipartBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	
	if err := mappingByPtr(obj, (*multipartRequest)(req), "form"); err != nil {
		return err
	}
	
	return validate(obj)
}

// =============================================================================
// 增强绑定配置
// =============================================================================

var (
	// 全局绑定配置
	EnableDecoderUseNumber               = false
	EnableDecoderDisallowUnknownFields   = false
	defaultMemory                        = int64(32 << 20) // 32 MB
	
	// 绑定实例
	JSON          = jsonBinding{}
	XML           = xmlBinding{}
	Form          = formBinding{}
	Query         = queryBinding{}
	FormPost      = formPostBinding{}
	FormMultipart = formMultipartBinding{}
	Header        = headerBinding{}
	URI           = uriBinding{}
)

// =============================================================================
// 表单数据映射
// =============================================================================

func mapForm(ptr any, form map[string][]string) error {
	return mappingByPtr(ptr, formSource(form), "form")
}

func mapHeader(ptr any, h map[string][]string) error {
	return mappingByPtr(ptr, headerSource(h), "header")
}

func mapURI(ptr any, m map[string][]string) error {
	return mappingByPtr(ptr, uriSource(m), "uri")
}

// =============================================================================
// 数据源接口
// =============================================================================

type setter interface {
	TrySet(value reflect.Value, field reflect.StructField, key string, opt setOptions) (bool, error)
}

type formSource map[string][]string
type headerSource map[string][]string
type uriSource map[string][]string

func (form formSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (bool, error) {
	return setByForm(value, field, form, tagValue, opt)
}

func (hs headerSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (bool, error) {
	return setByForm(value, field, map[string][]string(hs), tagValue, opt)
}

func (us uriSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (bool, error) {
	return setByForm(value, field, map[string][]string(us), tagValue, opt)
}

// =============================================================================
// 反射映射核心逻辑
// =============================================================================

type setOptions struct {
	isDefaultExists bool
	defaultValue    string
}

func mappingByPtr(ptr any, setter setter, tag string) error {
	_, err := mapping(reflect.ValueOf(ptr), emptyField, setter, tag)
	return err
}

var emptyField = reflect.StructField{}

func mapping(value reflect.Value, field reflect.StructField, setter setter, tag string) (bool, error) {
	if value.Kind() == reflect.Ptr {
		var isNew bool
		vPtr := value
		if value.IsNil() {
			isNew = true
			vPtr = reflect.New(value.Type().Elem())
		}
		isSet, err := mapping(vPtr.Elem(), field, setter, tag)
		if err != nil {
			return false, err
		}
		if isNew && isSet {
			value.Set(vPtr)
		}
		return isSet, nil
	}

	if value.Kind() != reflect.Struct || !field.Anonymous {
		ok, err := tryToSetValue(value, field, setter, tag)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}

	if value.Kind() == reflect.Struct {
		tValue := value.Type()

		var isSet bool
		for i := 0; i < value.NumField(); i++ {
			sf := tValue.Field(i)
			if sf.PkgPath != "" && !sf.Anonymous { // unexported
				continue
			}
			ok, err := mapping(value.Field(i), sf, setter, tag)
			if err != nil {
				return false, err
			}
			isSet = isSet || ok
		}
		return isSet, nil
	}
	return false, nil
}

func tryToSetValue(value reflect.Value, field reflect.StructField, setter setter, tag string) (bool, error) {
	var tagValue string
	var setOpt setOptions

	tagValue = field.Tag.Get(tag)
	tagValue, opts := head(tagValue, ",")

	if tagValue == "-" {
		return false, nil
	}

	var opt string
	for len(opts) > 0 {
		opt, opts = head(opts, ",")

		if k, v := head(opt, "="); k == "default" {
			setOpt.isDefaultExists = true
			setOpt.defaultValue = v
		}
	}

	return setter.TrySet(value, field, tagValue, setOpt)
}

func setByForm(value reflect.Value, field reflect.StructField, form map[string][]string, tagValue string, opt setOptions) (bool, error) {
	vs, ok := form[tagValue]
	if !ok && !opt.isDefaultExists {
		return false, nil
	}

	switch value.Kind() {
	case reflect.Slice:
		if !ok {
			vs = []string{opt.defaultValue}
		}
		return true, setSlice(vs, value, field)
	case reflect.Array:
		if !ok {
			vs = []string{opt.defaultValue}
		}
		if len(vs) != value.Len() {
			return false, fmt.Errorf("array size mismatch")
		}
		return true, setArray(vs, value, field)
	default:
		var val string
		if !ok {
			val = opt.defaultValue
		} else if len(vs) > 0 {
			val = vs[0]
		}
		return true, setWithProperType(val, value, field)
	}
}

func setWithProperType(val string, value reflect.Value, field reflect.StructField) error {
	switch value.Kind() {
	case reflect.Int:
		return setIntField(val, 0, value)
	case reflect.Int8:
		return setIntField(val, 8, value)
	case reflect.Int16:
		return setIntField(val, 16, value)
	case reflect.Int32:
		return setIntField(val, 32, value)
	case reflect.Int64:
		switch value.Interface().(type) {
		case time.Duration:
			return setTimeDuration(val, value, field)
		}
		return setIntField(val, 64, value)
	case reflect.Uint:
		return setUintField(val, 0, value)
	case reflect.Uint8:
		return setUintField(val, 8, value)
	case reflect.Uint16:
		return setUintField(val, 16, value)
	case reflect.Uint32:
		return setUintField(val, 32, value)
	case reflect.Uint64:
		return setUintField(val, 64, value)
	case reflect.Bool:
		return setBoolField(val, value)
	case reflect.Float32:
		return setFloatField(val, 32, value)
	case reflect.Float64:
		return setFloatField(val, 64, value)
	case reflect.String:
		value.SetString(val)
	case reflect.Struct:
		switch value.Interface().(type) {
		case time.Time:
			return setTimeField(val, field, value)
		}
		return json.Unmarshal(stringToBytes(val), value.Addr().Interface())
	case reflect.Map:
		return json.Unmarshal(stringToBytes(val), value.Addr().Interface())
	case reflect.Ptr:
		if !value.Elem().IsValid() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setWithProperType(val, value.Elem(), field)
	default:
		return fmt.Errorf("unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	switch tf := strings.ToLower(timeFormat); tf {
	case "unix", "unixnano":
		return setTimeFieldWithFormat(val, tf, value)
	default:
		if val == "" {
			value.Set(reflect.ValueOf(time.Time{}))
			return nil
		}
		l := time.UTC
		if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
			l = time.UTC
		}
		if locTag := structField.Tag.Get("time_location"); locTag != "" {
			loc, err := time.LoadLocation(locTag)
			if err != nil {
				return err
			}
			l = loc
		}
		t, err := time.ParseInLocation(timeFormat, val, l)
		value.Set(reflect.ValueOf(t))
		return err
	}
}

func setTimeFieldWithFormat(val, format string, value reflect.Value) error {
	switch format {
	case "unix":
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(time.Unix(i, 0)))
	case "unixnano":
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(time.Unix(0, i)))
	}
	return nil
}

func setTimeDuration(val string, value reflect.Value, field reflect.StructField) error {
	d, err := time.ParseDuration(val)
	if err != nil {
		return err
	}
	value.Set(reflect.ValueOf(d))
	return nil
}

func setSlice(vals []string, value reflect.Value, field reflect.StructField) error {
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, slice, field)
	if err != nil {
		return err
	}
	value.Set(slice)
	return nil
}

func setArray(vals []string, value reflect.Value, field reflect.StructField) error {
	for i, s := range vals {
		err := setWithProperType(s, value.Index(i), field)
		if err != nil {
			return err
		}
	}
	return nil
}

// =============================================================================
// 工具函数
// =============================================================================

func head(str, sep string) (head string, tail string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, ""
	}
	return str[:idx], str[idx+len(sep):]
}

func stringToBytes(s string) []byte {
	return []byte(s)
}

// =============================================================================
// 验证器集成
// =============================================================================

var Validator StructValidator = &defaultValidator{}

type defaultValidator struct{}

func (*defaultValidator) ValidateStruct(obj any) error {
	if obj == nil {
		return nil
	}
	
	value := reflect.ValueOf(obj)
	valueType := value.Type()
	
	if value.Kind() == reflect.Ptr && !value.IsNil() {
		value = value.Elem()
		valueType = value.Type()
	}
	
	if value.Kind() != reflect.Struct {
		return nil
	}
	
	return validateStruct(value, valueType)
}

func (*defaultValidator) Engine() any {
	return nil
}

func validateStruct(value reflect.Value, valueType reflect.Type) error {
	for i := 0; i < value.NumField(); i++ {
		field := valueType.Field(i)
		fieldValue := value.Field(i)
		
		// 检查required标签
		if tag := field.Tag.Get("binding"); tag != "" {
			if strings.Contains(tag, "required") {
				if isEmptyValue(fieldValue) {
					return fmt.Errorf("field %s is required", field.Name)
				}
			}
		}
		
		// 递归验证嵌套结构体
		if fieldValue.Kind() == reflect.Struct {
			if err := validateStruct(fieldValue, fieldValue.Type()); err != nil {
				return err
			}
		} else if fieldValue.Kind() == reflect.Ptr && !fieldValue.IsNil() && fieldValue.Elem().Kind() == reflect.Struct {
			if err := validateStruct(fieldValue.Elem(), fieldValue.Elem().Type()); err != nil {
				return err
			}
		}
	}
	return nil
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String, reflect.Array:
		return v.Len() == 0
	case reflect.Map, reflect.Slice:
		return v.Len() == 0 || v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

func validate(obj any) error {
	if Validator == nil {
		return nil
	}
	return Validator.ValidateStruct(obj)
}

// =============================================================================
// 扩展绑定类型
// =============================================================================

type formPostBinding struct{}
type formMultipartBinding struct{}

func (formPostBinding) Name() string {
	return "form-urlencoded"
}

func (formPostBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	if err := mapForm(obj, req.PostForm); err != nil {
		return err
	}
	return validate(obj)
}

func (formMultipartBinding) Name() string {
	return "multipart/form-data"
}

func (formMultipartBinding) Bind(req *http.Request, obj any) error {
	if err := req.ParseMultipartForm(defaultMemory); err != nil {
		return err
	}
	if err := mappingByPtr(obj, (*multipartRequest)(req), "form"); err != nil {
		return err
	}
	return validate(obj)
}

type multipartRequest http.Request

func (r *multipartRequest) TrySet(value reflect.Value, field reflect.StructField, key string, opt setOptions) (bool, error) {
	req := (*http.Request)(r)
	
	if req.MultipartForm == nil {
		return false, nil
	}
	
	if values := req.MultipartForm.Value[key]; len(values) > 0 {
		return setByForm(value, field, map[string][]string{key: values}, key, opt)
	}
	
	if files := req.MultipartForm.File[key]; len(files) > 0 {
		return setByMultipartFormFile(value, field, files)
	}
	
	return false, nil
}

func setByMultipartFormFile(value reflect.Value, field reflect.StructField, files []*multipart.FileHeader) (bool, error) {
	switch value.Kind() {
	case reflect.Ptr:
		switch value.Interface().(type) {
		case *multipart.FileHeader:
			value.Set(reflect.ValueOf(files[0]))
			return true, nil
		}
	case reflect.Struct:
		switch value.Interface().(type) {
		case multipart.FileHeader:
			value.Set(reflect.ValueOf(*files[0]))
			return true, nil
		}
	case reflect.Slice:
		slice := reflect.MakeSlice(value.Type(), len(files), len(files))
		isSet := false
		for i, file := range files {
			if _, err := setByMultipartFormFile(slice.Index(i), field, []*multipart.FileHeader{file}); err != nil {
				return false, err
			}
			isSet = true
		}
		if isSet {
			value.Set(slice)
		}
		return isSet, nil
	case reflect.Array:
		return false, fmt.Errorf("unsupported field type for file array")
	}
	return false, fmt.Errorf("unsupported field type for file")
}

// =============================================================================
// Context绑定方法增强
// =============================================================================

// 在Context中添加增强的绑定方法

// ShouldBindJSONEnhanced JSON绑定，不会中止请求
func (c *Context) ShouldBindJSONEnhanced(obj any) error {
	return c.ShouldBindWith(obj, JSON)
}

// ShouldBindXMLEnhanced XML绑定，不会中止请求
func (c *Context) ShouldBindXMLEnhanced(obj any) error {
	return c.ShouldBindWith(obj, XML)
}

// ShouldBindQueryEnhanced Query参数绑定，不会中止请求
func (c *Context) ShouldBindQueryEnhanced(obj any) error {
	return c.ShouldBindWith(obj, Query)
}

// ShouldBindHeaderEnhanced Header绑定，不会中止请求
func (c *Context) ShouldBindHeaderEnhanced(obj any) error {
	return c.ShouldBindWith(obj, Header)
}

// ShouldBindUriEnhanced URI参数绑定，不会中止请求
func (c *Context) ShouldBindUriEnhanced(obj any) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return URI.BindURI(m, obj)
}

// ShouldBindWith 使用指定绑定器进行绑定
func (c *Context) ShouldBindWith(obj any, b Binding) error {
	return b.Bind(c.Request(), obj)
}

// MustBindWith 使用指定绑定器进行绑定，失败时中止请求
func (c *Context) MustBindWith(obj any, b Binding) error {
	if err := c.ShouldBindWith(obj, b); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind)
		return err
	}
	return nil
}

// BindJSONEnhanced JSON绑定，失败时中止请求
func (c *Context) BindJSONEnhanced(obj any) error {
	return c.MustBindWith(obj, JSON)
}

// BindXMLEnhanced XML绑定，失败时中止请求
func (c *Context) BindXMLEnhanced(obj any) error {
	return c.MustBindWith(obj, XML)
}

// BindQueryEnhanced Query绑定，失败时中止请求
func (c *Context) BindQueryEnhanced(obj any) error {
	return c.MustBindWith(obj, Query)
}

// BindHeaderEnhanced Header绑定，失败时中止请求  
func (c *Context) BindHeaderEnhanced(obj any) error {
	return c.MustBindWith(obj, Header)
}

// BindUriEnhanced URI绑定，失败时中止请求
func (c *Context) BindUriEnhanced(obj any) error {
	if err := c.ShouldBindUriEnhanced(obj); err != nil {
		c.AbortWithError(http.StatusBadRequest, err).SetType(ErrorTypeBind)
		return err
	}
	return nil
}

// BindEnhanced 自动检测内容类型并绑定
func (c *Context) BindEnhanced(obj any) error {
	b := DefaultBinding(c.Request().Method, c.ContentType())
	return c.MustBindWith(obj, b)
}

// ShouldBindEnhanced 自动检测内容类型并绑定，不会中止请求
func (c *Context) ShouldBindEnhanced(obj any) error {
	b := DefaultBinding(c.Request().Method, c.ContentType())
	return c.ShouldBindWith(obj, b)
}

// DefaultBinding 根据HTTP方法和内容类型返回默认绑定器
func DefaultBinding(method, contentType string) Binding {
	if method == http.MethodGet {
		return Form
	}

	switch contentType {
	case MIMEJSON:
		return JSON
	case MIMEXML, MIMEXML2:
		return XML
	case MIMEPOSTForm:
		return FormPost
	case MIMEMultipartPOSTForm:
		return FormMultipart
	default:
		return Form
	}
}

// =============================================================================
// 数据转换和验证工具
// =============================================================================

// Transform 数据转换工具
type Transform struct {
	ToLower   bool
	ToUpper   bool
	Trim      bool
	Strip     string
	MinLength int
	MaxLength int
}

// Apply 应用转换
func (t *Transform) Apply(value string) string {
	result := value
	
	if t.Trim {
		result = strings.TrimSpace(result)
	}
	
	if t.Strip != "" {
		result = strings.ReplaceAll(result, t.Strip, "")
	}
	
	if t.ToLower {
		result = strings.ToLower(result)
	}
	
	if t.ToUpper {
		result = strings.ToUpper(result)
	}
	
	if t.MinLength > 0 && len(result) < t.MinLength {
		result = result + strings.Repeat(" ", t.MinLength-len(result))
	}
	
	if t.MaxLength > 0 && len(result) > t.MaxLength {
		result = result[:t.MaxLength]
	}
	
	return result
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidationErrors 验证错误列表
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// BindingOptions 绑定选项
type BindingOptions struct {
	RequiredByDefault bool        // 默认所有字段都是必需的
	Transform         *Transform  // 数据转换
	Validator         func(any) error // 自定义验证器
}

// EnhancedBind 增强绑定方法
func (c *Context) EnhancedBind(obj any, opts *BindingOptions) error {
	// 先进行标准绑定
	if err := c.ShouldBind(obj); err != nil {
		return err
	}
	
	// 应用选项
	if opts != nil {
		if opts.Transform != nil {
			applyTransform(obj, opts.Transform)
		}
		
		if opts.Validator != nil {
			if err := opts.Validator(obj); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func applyTransform(obj any, transform *Transform) {
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr && !value.IsNil() {
		value = value.Elem()
	}
	
	if value.Kind() != reflect.Struct {
		return
	}
	
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			original := field.String()
			transformed := transform.Apply(original)
			field.SetString(transformed)
		}
	}
}