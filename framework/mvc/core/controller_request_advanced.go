package core

import (
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 高级请求处理方法 =============

// Bind 智能绑定数据 (根据Content-Type自动选择)
func (c *BaseController) Bind(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.Bind(obj)
}

// ShouldBind 安全绑定数据 (不会在失败时终止)
func (c *BaseController) ShouldBind(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.ShouldBind(obj)
}

// BindJSON 绑定JSON数据
func (c *BaseController) BindJSON(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.BindJSON(obj)
}

// ShouldBindJSON 安全绑定JSON数据
func (c *BaseController) ShouldBindJSON(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.ShouldBindJSON(obj)
}

// BindXML 绑定XML数据
func (c *BaseController) BindXML(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.BindXML(obj)
}

// BindQuery 绑定查询参数到结构体
func (c *BaseController) BindQuery(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.ShouldBindQuery(obj)
}

// ShouldBindQuery 安全绑定查询参数
func (c *BaseController) ShouldBindQuery(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.ShouldBindQuery(obj)
}

// ============= 数据验证方法 =============

// Validate 验证绑定的数据
func (c *BaseController) Validate(obj any) error {
	// 这里可以集成验证库，如go-playground/validator
	// 目前实现基础验证
	return c.validateStruct(obj)
}

// ValidateStruct 验证结构体
func (c *BaseController) ValidateStruct(obj any) error {
	return c.validateStruct(obj)
}

// BindAndValidate 绑定并验证数据
func (c *BaseController) BindAndValidate(obj any) error {
	if err := c.Bind(obj); err != nil {
		return fmt.Errorf("bind error: %v", err)
	}
	if err := c.Validate(obj); err != nil {
		return fmt.Errorf("validation error: %v", err)
	}
	return nil
}

// ShouldBindAndValidate 安全绑定并验证数据
func (c *BaseController) ShouldBindAndValidate(obj any) error {
	if err := c.ShouldBind(obj); err != nil {
		return fmt.Errorf("bind error: %v", err)
	}
	if err := c.Validate(obj); err != nil {
		return fmt.Errorf("validation error: %v", err)
	}
	return nil
}

// validateStruct 基础结构体验证
func (c *BaseController) validateStruct(obj any) error {
	if obj == nil {
		return fmt.Errorf("validation object is nil")
	}
	
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("validation object must be a struct")
	}
	
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)
		
		// 检查required标签
		if tag := fieldType.Tag.Get("validate"); tag != "" {
			if strings.Contains(tag, "required") {
				if c.isEmptyValue(field) {
					return fmt.Errorf("field %s is required", fieldType.Name)
				}
			}
		}
	}
	
	return nil
}

// isEmptyValue 检查值是否为空
func (c *BaseController) isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
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
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	}
	return false
}

// ============= 请求体处理方法 =============

// GetRawBody 获取原始请求体
func (c *BaseController) GetRawBody() ([]byte, error) {
	if c.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	return c.Ctx.RawBody()
}

// GetRawData 获取原始请求数据 (兼容gin命名)
func (c *BaseController) GetRawData() ([]byte, error) {
	return c.GetRawBody()
}

// GetBodySize 获取请求体大小
func (c *BaseController) GetBodySize() int64 {
	if c.Ctx == nil || c.Ctx.Request == nil {
		return 0
	}
	
	body, err := c.Ctx.Request.Body()
	if err != nil {
		return 0
	}
	return int64(len(body))
}

// HasBody 检查是否有请求体
func (c *BaseController) HasBody() bool {
	return c.GetBodySize() > 0
}

// GetBodyString 获取请求体字符串
func (c *BaseController) GetBodyString() (string, error) {
	body, err := c.GetRawBody()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// ============= 多部分表单处理方法 =============

// GetMultipartForm 获取多部分表单
func (c *BaseController) GetMultipartForm() (*multipart.Form, error) {
	if c.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	return c.Ctx.Request.MultipartForm()
}

// ParseMultipartForm 解析多部分表单
func (c *BaseController) ParseMultipartForm(maxMemory int64) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.ParseMultipartForm(maxMemory)
}

// SetMaxMemory 设置多部分表单最大内存
func (c *BaseController) SetMaxMemory(size int64) {
	// 这个方法主要用于设置解析表单时的最大内存限制
	// 实际的限制设置需要在解析时指定
	config.Infof("Max memory for multipart form set to: %d bytes", size)
}

// ============= 高级参数获取方法 =============

// GetQueryAll 获取所有同名查询参数
func (c *BaseController) GetQueryAll(key string) []string {
	if c.Ctx == nil {
		return nil
	}
	return c.Ctx.QueryAll(key)
}

// GetQueryMap 获取查询参数映射
func (c *BaseController) GetQueryMap() map[string][]string {
	if c.Ctx == nil {
		return make(map[string][]string)
	}
	return c.Ctx.QueryMap()
}

// GetQueryDefault 获取带默认值的查询参数
func (c *BaseController) GetQueryDefault(key, defaultValue string) string {
	if c.Ctx == nil {
		return defaultValue
	}
	return c.Ctx.QueryDefault(key, defaultValue)
}

// GetQueryInt64 获取int64类型查询参数
func (c *BaseController) GetQueryInt64(key string) (int64, error) {
	if c.Ctx == nil {
		return 0, fmt.Errorf("context is nil")
	}
	return c.Ctx.QueryInt64(key)
}

// GetQueryFloat64 获取float64类型查询参数
func (c *BaseController) GetQueryFloat64(key string) (float64, error) {
	if c.Ctx == nil {
		return 0, fmt.Errorf("context is nil")
	}
	return c.Ctx.QueryFloat64(key)
}

// GetQueryBool 获取bool类型查询参数
func (c *BaseController) GetQueryBool(key string) (bool, error) {
	if c.Ctx == nil {
		return false, fmt.Errorf("context is nil")
	}
	return c.Ctx.QueryBool(key)
}

// ============= 高级表单处理方法 =============

// GetFormValue 获取表单值 (GET/POST通用)
func (c *BaseController) GetFormValue(key string) string {
	if c.Ctx == nil {
		return ""
	}
	return c.Ctx.FormValue(key)
}

// GetFormValueDefault 获取带默认值的表单参数
func (c *BaseController) GetFormValueDefault(key, defaultValue string) string {
	if c.Ctx == nil {
		return defaultValue
	}
	return c.Ctx.FormValueDefault(key, defaultValue)
}

// GetPostFormInt 获取int类型POST表单参数
func (c *BaseController) GetPostFormInt(key string) (int, error) {
	if c.Ctx == nil {
		return 0, fmt.Errorf("context is nil")
	}
	return c.Ctx.PostFormInt(key)
}

// GetPostFormFloat64 获取float64类型POST表单参数
func (c *BaseController) GetPostFormFloat64(key string) (float64, error) {
	if c.Ctx == nil {
		return 0, fmt.Errorf("context is nil")
	}
	return c.Ctx.PostFormFloat64(key)
}

// GetPostFormBool 获取bool类型POST表单参数
func (c *BaseController) GetPostFormBool(key string) (bool, error) {
	if c.Ctx == nil {
		return false, fmt.Errorf("context is nil")
	}
	return c.Ctx.PostFormBool(key)
}

// ============= 路由参数增强方法 =============

// GetParamNames 获取所有路由参数名称
func (c *BaseController) GetParamNames() []string {
	if c.Ctx == nil {
		return nil
	}
	
	var names []string
	for _, param := range c.Ctx.Params {
		names = append(names, param.Key)
	}
	return names
}

// GetParamValues 获取所有路由参数值
func (c *BaseController) GetParamValues() []string {
	if c.Ctx == nil {
		return nil
	}
	
	var values []string
	for _, param := range c.Ctx.Params {
		values = append(values, param.Value)
	}
	return values
}

// GetAllParams 获取所有路由参数映射
func (c *BaseController) GetAllParams() map[string]string {
	if c.Ctx == nil {
		return make(map[string]string)
	}
	
	result := make(map[string]string)
	for _, param := range c.Ctx.Params {
		result[param.Key] = param.Value
	}
	return result
}

// GetParamInt 获取int类型路由参数
func (c *BaseController) GetParamInt(key string) (int, error) {
	value := c.GetParam(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s not found", key)
	}
	return strconv.Atoi(value)
}

// GetParamInt64 获取int64类型路由参数
func (c *BaseController) GetParamInt64(key string) (int64, error) {
	value := c.GetParam(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s not found", key)
	}
	return strconv.ParseInt(value, 10, 64)
}

// GetParamFloat64 获取float64类型路由参数
func (c *BaseController) GetParamFloat64(key string) (float64, error) {
	value := c.GetParam(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s not found", key)
	}
	return strconv.ParseFloat(value, 64)
}

// GetParamBool 获取bool类型路由参数
func (c *BaseController) GetParamBool(key string) (bool, error) {
	value := c.GetParam(key)
	if value == "" {
		return false, fmt.Errorf("parameter %s not found", key)
	}
	return strconv.ParseBool(value)
}

// GetParamDefault 获取带默认值的路由参数
func (c *BaseController) GetParamDefault(key, defaultValue string) string {
	value := c.GetParam(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// ============= 请求内容类型判断方法 =============

// IsJSON 判断请求是否为JSON格式
func (c *BaseController) IsJSON() bool {
	if c.Ctx == nil {
		return false
	}
	contentType := c.Ctx.ContentType()
	return strings.Contains(strings.ToLower(contentType), "application/json")
}

// IsXML 判断请求是否为XML格式
func (c *BaseController) IsXML() bool {
	if c.Ctx == nil {
		return false
	}
	contentType := c.Ctx.ContentType()
	return strings.Contains(strings.ToLower(contentType), "application/xml") ||
		   strings.Contains(strings.ToLower(contentType), "text/xml")
}

// IsForm 判断请求是否为表单格式
func (c *BaseController) IsForm() bool {
	if c.Ctx == nil {
		return false
	}
	contentType := c.Ctx.ContentType()
	return strings.Contains(strings.ToLower(contentType), "application/x-www-form-urlencoded")
}

// IsMultipart 判断请求是否为多部分表单
func (c *BaseController) IsMultipart() bool {
	if c.Ctx == nil {
		return false
	}
	contentType := c.Ctx.ContentType()
	return strings.Contains(strings.ToLower(contentType), "multipart/form-data")
}

// IsUpload 判断请求是否包含文件上传
func (c *BaseController) IsUpload() bool {
	return c.IsMultipart()
}