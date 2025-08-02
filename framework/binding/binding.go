// Package binding 参数绑定系统
// 借鉴Gin框架的绑定机制，支持多种数据源的自动绑定和验证
package binding

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

// Binding 绑定接口
type Binding interface {
	Name() string
	Bind(*app.RequestContext, any) error
}

// BindingBody 需要读取请求体的绑定接口
type BindingBody interface {
	Binding
	BindBody([]byte, any) error
}

// BindingUri URI绑定接口
type BindingUri interface {
	Name() string
	BindUri(map[string][]string, any) error
}

// 预定义的绑定器实例
var (
	JSON          = jsonBinding{}
	XML           = xmlBinding{}
	Form          = formBinding{}
	Query         = queryBinding{}
	FormPost      = formPostBinding{}
	FormMultipart = formMultipartBinding{}
	ProtoBuf      = protobufBinding{}
	MsgPack       = msgpackBinding{}
	YAML          = yamlBinding{}
	Uri           = uriBinding{}
	Header        = headerBinding{}
)

// 默认验证器
var Validator StructValidator = &defaultValidator{}

// Default 根据HTTP方法和内容类型返回默认绑定器
func Default(method, contentType string) Binding {
	if method == "GET" {
		return Form
	}

	switch contentType {
	case "application/json":
		return JSON
	case "application/xml", "text/xml":
		return XML
	case "application/x-protobuf":
		return ProtoBuf
	case "application/x-msgpack":
		return MsgPack
	case "application/x-yaml", "text/yaml":
		return YAML
	case "multipart/form-data":
		return FormMultipart
	default: // case "application/x-www-form-urlencoded":
		return Form
	}
}

// StructValidator 结构体验证器接口
type StructValidator interface {
	ValidateStruct(any) error
	Engine() any
}

// defaultValidator 默认验证器实现
type defaultValidator struct {
	once     bool
	validate *validator.Validate
}

func (v *defaultValidator) ValidateStruct(obj any) error {
	if !v.once {
		v.lazyinit()
	}
	return v.validate.Struct(obj)
}

func (v *defaultValidator) Engine() any {
	if !v.once {
		v.lazyinit()
	}
	return v.validate
}

func (v *defaultValidator) lazyinit() {
	v.validate = validator.New()
	v.validate.SetTagName("binding")
	v.once = true
}

// JSON绑定器
type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(req *app.RequestContext, obj any) error {
	return decodeJSON(req.Request.Body(), obj)
}

func (jsonBinding) BindBody(body []byte, obj any) error {
	return decodeJSON(body, obj)
}

func decodeJSON(body []byte, obj any) error {
	if err := json.Unmarshal(body, obj); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// XML绑定器
type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(req *app.RequestContext, obj any) error {
	return decodeXML(req.Request.Body(), obj)
}

func (xmlBinding) BindBody(body []byte, obj any) error {
	return decodeXML(body, obj)
}

func decodeXML(body []byte, obj any) error {
	if err := xml.Unmarshal(body, obj); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// Form绑定器
type formBinding struct{}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind(req *app.RequestContext, obj any) error {
	if err := req.Bind(obj); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// Query绑定器
type queryBinding struct{}

func (queryBinding) Name() string {
	return "query"
}

func (queryBinding) Bind(req *app.RequestContext, obj any) error {
	values := make(url.Values)
	req.URI().QueryArgs().VisitAll(func(key, value []byte) {
		values.Add(string(key), string(value))
	})
	if err := mapForm(obj, values); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// FormPost绑定器
type formPostBinding struct{}

func (formPostBinding) Name() string {
	return "form-urlencoded"
}

func (formPostBinding) Bind(req *app.RequestContext, obj any) error {
	if err := req.Bind(obj); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// FormMultipart绑定器
type formMultipartBinding struct{}

func (formMultipartBinding) Name() string {
	return "multipart/form-data"
}

func (formMultipartBinding) Bind(req *app.RequestContext, obj any) error {
	if err := req.Bind(obj); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// ProtoBuf绑定器
type protobufBinding struct{}

func (protobufBinding) Name() string {
	return "protobuf"
}

func (protobufBinding) Bind(req *app.RequestContext, obj any) error {
	// TODO: 实现protobuf绑定
	return fmt.Errorf("protobuf binding not implemented")
}

func (protobufBinding) BindBody(body []byte, obj any) error {
	// TODO: 实现protobuf绑定
	return fmt.Errorf("protobuf binding not implemented")
}

// MsgPack绑定器
type msgpackBinding struct{}

func (msgpackBinding) Name() string {
	return "msgpack"
}

func (msgpackBinding) Bind(req *app.RequestContext, obj any) error {
	// TODO: 实现msgpack绑定
	return fmt.Errorf("msgpack binding not implemented")
}

func (msgpackBinding) BindBody(body []byte, obj any) error {
	// TODO: 实现msgpack绑定
	return fmt.Errorf("msgpack binding not implemented")
}

// YAML绑定器
type yamlBinding struct{}

func (yamlBinding) Name() string {
	return "yaml"
}

func (yamlBinding) Bind(req *app.RequestContext, obj any) error {
	return decodeYAML(req.Request.Body(), obj)
}

func (yamlBinding) BindBody(body []byte, obj any) error {
	return decodeYAML(body, obj)
}

func decodeYAML(body []byte, obj any) error {
	if err := yaml.Unmarshal(body, obj); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// URI绑定器
type uriBinding struct{}

func (uriBinding) Name() string {
	return "uri"
}

func (uriBinding) BindUri(m map[string][]string, obj any) error {
	if err := mapUri(obj, m); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// Header绑定器
type headerBinding struct{}

func (headerBinding) Name() string {
	return "header"
}

func (headerBinding) Bind(req *app.RequestContext, obj any) error {
	if err := mapHeader(obj, req); err != nil {
		return err
	}
	return Validator.ValidateStruct(obj)
}

// 辅助函数

// mapForm 将表单数据映射到结构体
func mapForm(ptr any, form url.Values) error {
	return mapFormByTag(ptr, form, "form")
}

// mapFormByTag 按标签映射表单数据
func mapFormByTag(ptr any, form url.Values, tag string) error {
	if ptr == nil || len(form) == 0 {
		return nil
	}

	return mapping(ptr, formSource(form), tag)
}

// mapUri 映射URI参数
func mapUri(ptr any, m map[string][]string) error {
	return mapFormByTag(ptr, m, "uri")
}

// mapHeader 映射请求头
func mapHeader(ptr any, req *app.RequestContext) error {
	h := make(map[string][]string)
	req.Request.Header.VisitAll(func(key, value []byte) {
		h[string(key)] = []string{string(value)}
	})
	return mapFormByTag(ptr, h, "header")
}

// formSource form数据源
type formSource map[string][]string

func (f formSource) TrySet(value reflect.Value, field reflect.StructField, tagValue string, opt setOptions) (bool, error) {
	return setByForm(value, field, f, tagValue, opt)
}

// setByForm 通过表单设置值
func setByForm(value reflect.Value, field reflect.StructField, form map[string][]string, tagValue string, opt setOptions) (isSetted bool, err error) {
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
		}

		if len(vs) > 0 {
			val = vs[0]
		}
		return true, setWithProperType(val, value, field)
	}
}

// setOptions 设置选项
type setOptions struct {
	isDefaultExists bool
	defaultValue    string
}

// setSlice 设置切片值
func setSlice(vals []string, value reflect.Value, field reflect.StructField) error {
	slice := reflect.MakeSlice(value.Type(), len(vals), len(vals))
	err := setArray(vals, slice, field)
	if err != nil {
		return err
	}
	value.Set(slice)
	return nil
}

// setArray 设置数组值
func setArray(vals []string, value reflect.Value, field reflect.StructField) error {
	for i, s := range vals {
		err := setWithProperType(s, value.Index(i), field)
		if err != nil {
			return err
		}
	}
	return nil
}

// setWithProperType 设置适当类型的值
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
	case reflect.Ptr:
		return setWithProperType(val, value.Elem(), field)
	default:
		return fmt.Errorf("unknown type")
	}
	return nil
}

// setIntField 设置整数字段
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

// setUintField 设置无符号整数字段
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

// setBoolField 设置布尔字段
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

// setFloatField 设置浮点数字段
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

// mappingByPtr 按指针映射
type mappingByPtr interface {
	TrySet(value reflect.Value, field reflect.StructField, key string, opt setOptions) (bool, error)
}

// mapping 映射函数
func mapping(ptr any, mapper mappingByPtr, tag string) error {
	err := mapFormByTag2(ptr, mapper, tag)
	return err
}

func mapFormByTag2(ptr any, mapper mappingByPtr, tag string) error {
	if ptr == nil {
		return nil
	}
	v := reflect.ValueOf(ptr).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		typeField := t.Field(i)
		structField := v.Field(i)
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()
		inputFieldName := typeField.Tag.Get(tag)

		if inputFieldName == "" {
			inputFieldName = typeField.Name
			// 如果字段是匿名的，递归处理
			if typeField.Anonymous && structFieldKind == reflect.Struct {
				if err := mapFormByTag2(structField.Addr().Interface(), mapper, tag); err != nil {
					return err
				}
				continue
			}
		}

		if inputFieldName == "-" {
			continue
		}

		inputFieldName, opts := head(inputFieldName, ",")
		opt := setOptions{}
		for _, opt_str := range opts {
			opt_str = strings.TrimSpace(opt_str)
			if strings.HasPrefix(opt_str, "default=") {
				opt.defaultValue = opt_str[8:]
				opt.isDefaultExists = true
			}
		}

		if ok, err := tryToSetValue(structField, typeField, mapper, inputFieldName, opt); err != nil {
			return err
		} else if ok {
			continue
		}

		if structFieldKind == reflect.Struct {
			if err := mapFormByTag2(structField.Addr().Interface(), mapper, tag); err != nil {
				return err
			}
		}
	}
	return nil
}

func tryToSetValue(value reflect.Value, field reflect.StructField, mapper mappingByPtr, inputFieldName string, opt setOptions) (bool, error) {
	return mapper.TrySet(value, field, inputFieldName, opt)
}

func head(str, sep string) (head string, tail []string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, []string{}
	}
	return str[:idx], strings.Split(str[idx+len(sep):], sep)
}
