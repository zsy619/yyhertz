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
	"time"

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
		return Query  // GET 请求使用 Query 绑定器处理查询参数
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
	
	// 注册自定义验证器（如果需要）
	// 确保UUID验证器可用
	v.setupCustomValidators()
	
	v.once = true
}

// setupCustomValidators 设置自定义验证器
func (v *defaultValidator) setupCustomValidators() {
	// UUID验证器应该已经内置在validator包中
	// 这里可以添加其他自定义验证器
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
	// 检查结构体是否有time_format标签的时间字段
	if hasCustomTimeFields(obj) {
		// 如果有自定义时间字段，使用两阶段解析
		if err := decodeJSONWithCustomTime(body, obj); err != nil {
			return err
		}
	} else {
		// 没有自定义时间字段，使用标准JSON解析
		if err := json.Unmarshal(body, obj); err != nil {
			return err
		}
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
	// 解析表单数据
	values := make(url.Values)
	req.PostArgs().VisitAll(func(key, value []byte) {
		values.Add(string(key), string(value))
	})
	
	// 使用自定义的表单映射逻辑（支持时间解析）
	if err := mapForm(obj, values); err != nil {
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
	if ptr == nil {
		return fmt.Errorf("mapUri: 目标对象不能为nil")
	}
	
	// 检查是否为指针类型
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		return fmt.Errorf("mapUri: 目标对象必须是指针类型，得到 %T", ptr)
	}
	
	// 检查指针指向的是否为结构体
	elem := reflect.TypeOf(ptr).Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("mapUri: 目标对象必须指向结构体，得到 %s", elem.Kind())
	}
	
	// 增强错误处理的mapFormByTag调用
	err := mapFormByTag(ptr, m, "uri")
	if err != nil {
		return fmt.Errorf("mapUri: URI参数映射失败 - %w", err)
	}
	
	return nil
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
	case reflect.Struct:
		// 检查是否是 time.Time 类型
		if value.Type() == reflect.TypeOf(time.Time{}) {
			return setTimeField(val, value, field)
		}
		return fmt.Errorf("unsupported struct type: %v", value.Type())
	case reflect.Ptr:
		return setWithProperType(val, value.Elem(), field)
	default:
		return fmt.Errorf("unknown type: %v", value.Kind())
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

// setTimeField 设置时间字段
//
// 支持 time_format 和 time_utc 标签解析时间值
// 兼容 Gin 框架的时间绑定行为
//
// 参数：
//   - val: 要解析的时间字符串
//   - value: 目标 time.Time 字段的反射值
//   - field: 结构体字段信息（包含标签）
//
// 返回：
//   - error: 解析错误
func setTimeField(val string, value reflect.Value, field reflect.StructField) error {
	if val == "" {
		// 空值设置为零值
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	// 从结构体标签中获取时间格式和时区设置
	timeFormat := field.Tag.Get("time_format")
	timeUTC := field.Tag.Get("time_utc")

	// 解析时间
	parsedTime, err := parseTimeWithFormat(val, timeFormat, timeUTC)
	if err != nil {
		return fmt.Errorf("时间解析失败: %w", err)
	}

	// 设置字段值
	value.Set(reflect.ValueOf(parsedTime))
	return nil
}

// parseTimeWithFormat 根据格式解析时间字符串
//
// 支持多种时间格式：
//   - 自定义格式 (如 "2006-01-02")
//   - Unix 时间戳 ("unix", "unixNano", "unixMilli", "uNiXmIcRo")
//   - 默认 RFC3339 格式
//
// 参数：
//   - timeStr: 时间字符串
//   - format: 时间格式（time_format 标签值）
//   - utc: 是否使用 UTC 时区（time_utc 标签值）
//
// 返回：
//   - time.Time: 解析后的时间
//   - error: 解析错误
func parseTimeWithFormat(timeStr, format, utc string) (time.Time, error) {
	var location *time.Location = time.Local

	// 处理 time_utc 标签
	if utc == "1" || strings.ToLower(utc) == "true" {
		location = time.UTC
	}

	// 如果没有指定格式，使用 RFC3339
	if format == "" {
		format = time.RFC3339
	}

	// 处理特殊的 Unix 时间戳格式
	switch strings.ToLower(format) {
	case "unix":
		// Unix 秒级时间戳
		if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			return time.Unix(timestamp, 0).In(location), nil
		}
		return time.Time{}, fmt.Errorf("无效的 Unix 时间戳: %s", timeStr)

	case "unixnano":
		// Unix 纳秒级时间戳
		if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			return time.Unix(0, timestamp).In(location), nil
		}
		return time.Time{}, fmt.Errorf("无效的 Unix 纳秒时间戳: %s", timeStr)

	case "unixmilli":
		// Unix 毫秒级时间戳
		if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			return time.Unix(timestamp/1000, (timestamp%1000)*1000000).In(location), nil
		}
		return time.Time{}, fmt.Errorf("无效的 Unix 毫秒时间戳: %s", timeStr)

	case "unixmicro":
		// Unix 微秒级时间戳
		if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
			return time.Unix(timestamp/1000000, (timestamp%1000000)*1000).In(location), nil
		}
		return time.Time{}, fmt.Errorf("无效的 Unix 微秒时间戳: %s", timeStr)

	default:
		// 使用自定义格式解析
		if location == time.UTC {
			return time.ParseInLocation(format, timeStr, time.UTC)
		}
		return time.ParseInLocation(format, timeStr, time.Local)
	}
}

// processTimeFields 处理JSON绑定后的时间字段
//
// 在标准JSON解析后，对结构体中带有time_format标签的time.Time字段
// 进行重新解析，以支持自定义时间格式和时区设置
//
// 参数：
//   - obj: 已进行JSON解析的结构体指针
//   - jsonData: 原始JSON数据
//
// 返回：
//   - error: 处理错误
func processTimeFields(obj any, jsonData []byte) error {
	if obj == nil {
		return nil
	}
	
	// 解析JSON数据为map，用于获取原始字符串值
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &jsonMap); err != nil {
		return fmt.Errorf("解析JSON到map失败: %w", err)
	}
	
	// 递归处理结构体字段
	return processStructTimeFields(reflect.ValueOf(obj), reflect.TypeOf(obj), jsonMap, "")
}

// processStructTimeFields 递归处理结构体中的时间字段
//
// 参数：
//   - value: 结构体值的反射对象
//   - typ: 结构体类型的反射对象  
//   - jsonMap: JSON数据的map表示
//   - prefix: 字段前缀（用于嵌套结构体）
//
// 返回：
//   - error: 处理错误
func processStructTimeFields(value reflect.Value, typ reflect.Type, jsonMap map[string]interface{}, prefix string) error {
	// 如果是指针，获取其指向的值
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		}
		value = value.Elem()
		typ = typ.Elem()
	}
	
	// 确保是结构体类型
	if value.Kind() != reflect.Struct {
		return nil
	}
	
	// 遍历结构体字段
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		typeField := typ.Field(i)
		
		// 跳过不可设置的字段
		if !field.CanSet() {
			continue
		}
		
		// 获取JSON标签名
		jsonTag := typeField.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			jsonTag = typeField.Name
		}
		
		// 处理标签选项（如omitempty）
		jsonFieldName := strings.Split(jsonTag, ",")[0]
		fullFieldName := jsonFieldName
		if prefix != "" {
			fullFieldName = prefix + "." + jsonFieldName
		}
		
		// 处理time.Time字段
		if field.Type() == reflect.TypeOf(time.Time{}) {
			// 检查是否有time_format标签
			timeFormat := typeField.Tag.Get("time_format")
			timeUTC := typeField.Tag.Get("time_utc")
			
			if timeFormat != "" {
				// 从JSON map中获取原始字符串值
				if rawValue, exists := jsonMap[jsonFieldName]; exists {
					if timeStr, ok := rawValue.(string); ok {
						// 使用自定义格式重新解析时间
						parsedTime, err := parseTimeWithFormat(timeStr, timeFormat, timeUTC)
						if err != nil {
							return fmt.Errorf("字段 %s 时间解析失败: %w", fullFieldName, err)
						}
						field.Set(reflect.ValueOf(parsedTime))
					}
				}
			}
		} else if field.Kind() == reflect.Struct {
			// 递归处理嵌套结构体
			if nestedMap, ok := jsonMap[jsonFieldName].(map[string]interface{}); ok {
				if err := processStructTimeFields(field, field.Type(), nestedMap, fullFieldName); err != nil {
					return err
				}
			}
		} else if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			// 处理指向结构体的指针
			if !field.IsNil() {
				if nestedMap, ok := jsonMap[jsonFieldName].(map[string]interface{}); ok {
					if err := processStructTimeFields(field, field.Type(), nestedMap, fullFieldName); err != nil {
						return err
					}
				}
			}
		}
	}
	
	return nil
}

// hasCustomTimeFields 检查结构体是否有带time_format标签的时间字段
//
// 参数：
//   - obj: 要检查的结构体指针
//
// 返回：
//   - bool: 是否包含自定义时间字段
func hasCustomTimeFields(obj any) bool {
	if obj == nil {
		return false
	}
	
	return checkStructForCustomTimeFields(reflect.TypeOf(obj))
}

// checkStructForCustomTimeFields 递归检查结构体类型是否包含自定义时间字段
//
// 参数：
//   - typ: 要检查的类型
//
// 返回：
//   - bool: 是否包含自定义时间字段
func checkStructForCustomTimeFields(typ reflect.Type) bool {
	// 如果是指针，获取其指向的类型
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	
	// 确保是结构体类型
	if typ.Kind() != reflect.Struct {
		return false
	}
	
	// 遍历结构体字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		
		// 检查time.Time字段是否有time_format标签
		if field.Type == reflect.TypeOf(time.Time{}) {
			if field.Tag.Get("time_format") != "" {
				return true
			}
		} else if field.Type.Kind() == reflect.Struct {
			// 递归检查嵌套结构体
			if checkStructForCustomTimeFields(field.Type) {
				return true
			}
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct {
			// 检查指向结构体的指针
			if checkStructForCustomTimeFields(field.Type.Elem()) {
				return true
			}
		}
	}
	
	return false
}

// decodeJSONWithCustomTime 使用自定义时间解析的JSON解码
//
// 采用两阶段解析策略：
// 1. 先解析到map[string]interface{}获取原始值
// 2. 手动构造目标结构体，对时间字段应用自定义解析
//
// 参数：
//   - body: JSON字节数据
//   - obj: 目标结构体指针
//
// 返回：
//   - error: 解析错误
func decodeJSONWithCustomTime(body []byte, obj any) error {
	// 第一阶段：解析JSON到map
	var rawData map[string]interface{}
	if err := json.Unmarshal(body, &rawData); err != nil {
		return fmt.Errorf("JSON解析到map失败: %w", err)
	}
	
	// 第二阶段：手动填充结构体字段
	return populateStructFromMap(obj, rawData)
}

// populateStructFromMap 从map数据填充结构体
//
// 参数：
//   - obj: 目标结构体指针
//   - data: JSON解析后的map数据
//
// 返回：
//   - error: 填充错误
func populateStructFromMap(obj any, data map[string]interface{}) error {
	if obj == nil {
		return fmt.Errorf("目标对象不能为nil")
	}
	
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("目标对象必须是指针类型")
	}
	
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("目标对象必须指向结构体")
	}
	
	typ := value.Type()
	
	// 遍历结构体字段
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		typeField := typ.Field(i)
		
		// 跳过不可设置的字段
		if !field.CanSet() {
			continue
		}
		
		// 获取JSON标签名
		jsonTag := typeField.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			jsonTag = typeField.Name
		}
		
		// 处理标签选项（如omitempty）
		jsonFieldName := strings.Split(jsonTag, ",")[0]
		
		// 从map中获取值
		rawValue, exists := data[jsonFieldName]
		if !exists {
			continue
		}
		
		// 根据字段类型进行设置
		if err := setFieldFromRawValue(field, typeField, rawValue); err != nil {
			return fmt.Errorf("设置字段 %s 失败: %w", jsonFieldName, err)
		}
	}
	
	return nil
}

// setFieldFromRawValue 从原始值设置结构体字段
//
// 参数：
//   - field: 目标字段的反射值
//   - typeField: 字段的类型信息
//   - rawValue: 从JSON解析得到的原始值
//
// 返回：
//   - error: 设置错误
func setFieldFromRawValue(field reflect.Value, typeField reflect.StructField, rawValue interface{}) error {
	switch field.Type() {
	case reflect.TypeOf(time.Time{}):
		// 处理time.Time字段
		return setTimeFieldFromRaw(field, typeField, rawValue)
	default:
		// 处理其他类型字段
		return setGenericFieldFromRaw(field, rawValue)
	}
}

// setTimeFieldFromRaw 从原始值设置时间字段
//
// 参数：
//   - field: time.Time字段的反射值
//   - typeField: 字段的类型信息（包含标签）
//   - rawValue: 原始值
//
// 返回：
//   - error: 设置错误
func setTimeFieldFromRaw(field reflect.Value, typeField reflect.StructField, rawValue interface{}) error {
	// 将原始值转换为字符串
	var timeStr string
	switch v := rawValue.(type) {
	case string:
		timeStr = v
	case nil:
		// 空值设置为零值
		field.Set(reflect.ValueOf(time.Time{}))
		return nil
	default:
		return fmt.Errorf("时间字段必须是字符串类型，得到 %T", rawValue)
	}
	
	// 获取时间格式标签
	timeFormat := typeField.Tag.Get("time_format")
	timeUTC := typeField.Tag.Get("time_utc")
	
	if timeFormat != "" {
		// 使用自定义格式解析，带回退机制
		parsedTime, err := parseTimeWithFormat(timeStr, timeFormat, timeUTC)
		if err != nil {
			// 自定义格式失败，尝试标准JSON时间格式作为回退
			fallbackTime, fallbackErr := time.Parse(time.RFC3339, timeStr)
			if fallbackErr != nil {
				return fmt.Errorf("时间解析失败 - 自定义格式: %v, 标准格式: %v", err, fallbackErr)
			}
			// 如果需要UTC转换，应用到回退结果
			if timeUTC == "1" || strings.ToLower(timeUTC) == "true" {
				fallbackTime = fallbackTime.UTC()
			}
			field.Set(reflect.ValueOf(fallbackTime))
		} else {
			field.Set(reflect.ValueOf(parsedTime))
		}
	} else {
		// 使用标准JSON时间格式解析
		parsedTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("标准时间格式解析失败: %w", err)
		}
		field.Set(reflect.ValueOf(parsedTime))
	}
	
	return nil
}

// setGenericFieldFromRaw 从原始值设置通用字段
//
// 参数：
//   - field: 目标字段的反射值
//   - rawValue: 原始值
//
// 返回：
//   - error: 设置错误
func setGenericFieldFromRaw(field reflect.Value, rawValue interface{}) error {
	if rawValue == nil {
		// 空值处理
		field.Set(reflect.Zero(field.Type()))
		return nil
	}
	
	rawValueReflect := reflect.ValueOf(rawValue)
	
	// 类型检查和转换
	if rawValueReflect.Type().ConvertibleTo(field.Type()) {
		field.Set(rawValueReflect.Convert(field.Type()))
		return nil
	}
	
	return fmt.Errorf("无法将 %T 类型的值转换为 %s", rawValue, field.Type())
}
