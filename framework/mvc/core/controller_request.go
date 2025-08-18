package core

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ============= 参数获取方法 =============

// GetString 获取字符串参数
func (c *BaseController) GetString(key string, def ...string) string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}

	// 备用方法
	if val := c.Ctx.Query(key); val != "" {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetInt 获取整数参数
func (c *BaseController) GetInt(key string, def ...int) int {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetInt32 获取整数参数
func (c *BaseController) GetInt32(key string, def ...int32) int32 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return int32(i)
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetInt64 获取整数参数
func (c *BaseController) GetInt64(key string, def ...int64) int64 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.ParseInt(string(val), 10, 64); err == nil {
			return i
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetInt8 获取8位整数参数（Beego兼容）
func (c *BaseController) GetInt8(key string, def ...int8) (int8, error) {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	i64, err := strconv.ParseInt(val, 10, 8)
	return int8(i64), err
}

// GetInt16 获取16位整数参数（Beego兼容）
func (c *BaseController) GetInt16(key string, def ...int16) (int16, error) {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	i64, err := strconv.ParseInt(val, 10, 16)
	return int16(i64), err
}

// GetUint8 获取8位无符号整数参数（Beego兼容）
func (c *BaseController) GetUint8(key string, def ...uint8) (uint8, error) {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	u64, err := strconv.ParseUint(val, 10, 8)
	return uint8(u64), err
}

// GetUint16 获取16位无符号整数参数（Beego兼容）
func (c *BaseController) GetUint16(key string, def ...uint16) (uint16, error) {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	u64, err := strconv.ParseUint(val, 10, 16)
	return uint16(u64), err
}

// GetUint32 获取32位无符号整数参数（Beego兼容）
func (c *BaseController) GetUint32(key string, def ...uint32) (uint32, error) {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	u64, err := strconv.ParseUint(val, 10, 32)
	return uint32(u64), err
}

// GetUint64 获取64位无符号整数参数（Beego兼容）
func (c *BaseController) GetUint64(key string, def ...uint64) (uint64, error) {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0], nil
		}
		return 0, nil
	}

	return strconv.ParseUint(val, 10, 64)
}

// GetFloat32 获取浮点数参数
func (c *BaseController) GetFloat32(key string, def ...float32) float32 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.ParseFloat(string(val), 32); err == nil {
			return float32(i)
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetFloat64 获取浮点数参数
func (c *BaseController) GetFloat64(key string, def ...float64) float64 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0
	}
	if val := c.Ctx.Query(key); val != "" {
		if i, err := strconv.ParseFloat(string(val), 64); err == nil {
			return i
		}
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}

// GetParam 获取路径参数
func (c *BaseController) GetParam(key string) string {
	if c.Ctx == nil {
		return ""
	}
	return c.Ctx.Param(key)
}

// GetForm 获取表单参数
func (c *BaseController) GetForm(key string, def ...string) string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return ""
	}
	if val := c.Ctx.PostForm(key); val != "" {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

// GetBool 获取布尔参数（Beego兼容）
func (c *BaseController) GetBool(key string, def ...bool) bool {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return false
	}

	val := c.Ctx.Query(key)
	if val == "" {
		if len(def) > 0 {
			return def[0]
		}
		return false
	}

	// 转换字符串到布尔值
	switch strings.ToLower(val) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		if len(def) > 0 {
			return def[0]
		}
		return false
	}
}

// GetFloat 获取浮点数参数（Beego兼容）
func (c *BaseController) GetFloat(key string, def ...float64) float64 {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return 0.0
	}

	if val := c.Ctx.Query(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}

	if len(def) > 0 {
		return def[0]
	}
	return 0.0
}

// GetQuery 获取查询参数（Beego兼容）
func (c *BaseController) GetQuery(key string, def ...string) string {
	return c.GetString(key, def...)
}

// GetUserAgent 获取User-Agent
func (c *BaseController) GetUserAgent() string {
	if c.Ctx == nil {
		return ""
	}
	return string(c.Ctx.Request().GetHeader("User-Agent"))
}

// GetHeader 获取请求头
func (c *BaseController) GetHeader(key string) string {
	if c.Ctx == nil {
		return ""
	}
	return string(c.Ctx.Request().GetHeader(key))
}

// GetClientIP 获取客户端IP地址
func (c *BaseController) GetClientIP() string {
	if c.Ctx == nil {
		return ""
	}
	// 尝试从X-Forwarded-For获取真实IP
	if xff := c.Ctx.Request().GetHeader("X-Forwarded-For"); len(xff) > 0 {
		xffStr := string(xff)
		ips := strings.Split(xffStr, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 尝试从X-Real-IP获取
	if xri := c.Ctx.Request().GetHeader("X-Real-IP"); len(xri) > 0 {
		return string(xri)
	}

	// 从RemoteAddr获取
	remoteAddr := c.Ctx.Request().RemoteAddr().String()
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return host
}

// ============= HTTP方法判断 =============

// IsAjax 判断是否为AJAX请求
func (c *BaseController) IsAjax() bool {
	if c.Ctx == nil {
		return false
	}
	return string(c.Ctx.Request().GetHeader("X-Requested-With")) == "XMLHttpRequest"
}

// IsMethod 判断HTTP方法
func (c *BaseController) IsMethod(method string) bool {
	if c.Ctx == nil {
		return false
	}
	return strings.EqualFold(string(c.Ctx.Request().Method()), method)
}

// IsPost 判断是否为POST请求
func (c *BaseController) IsPost() bool {
	return c.IsMethod("POST")
}

// IsGet 判断是否为GET请求
func (c *BaseController) IsGet() bool {
	return c.IsMethod("GET")
}

// IsPut 判断是否为PUT请求
func (c *BaseController) IsPut() bool {
	return c.IsMethod("PUT")
}

// IsDelete 判断是否为DELETE请求
func (c *BaseController) IsDelete() bool {
	return c.IsMethod("DELETE")
}

// IsPatch 判断是否为PATCH请求
func (c *BaseController) IsPatch() bool {
	return c.IsMethod("PATCH")
}

// IsHead 判断是否为HEAD请求
func (c *BaseController) IsHead() bool {
	return c.IsMethod("HEAD")
}

// IsOptions 判断是否为OPTIONS请求
func (c *BaseController) IsOptions() bool {
	return c.IsMethod("OPTIONS")
}

// ============= 表单解析和多值参数方法 (Beego兼容) =============

// ParseForm 解析表单数据，包括URL查询参数和POST表单数据
// 这个方法会解析请求中的表单数据，使其可以通过其他方法访问
// 兼容 beego 的 ParseForm 方法
func (c *BaseController) ParseForm(obj any) error {
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return nil
	}

	// 获取原始请求
	req := c.Ctx.Request()

	// 解析查询参数和表单数据
	// Hertz 会自动解析这些数据，我们这里主要是为了兼容性
	// 实际的解析工作已经由 Hertz 框架完成

	// 如果传入了对象，尝试绑定表单数据到对象
	if obj != nil {
		// 使用 Hertz 的绑定功能
		return req.Bind(obj)
	}

	return nil
}

// GetStrings 获取同名参数的所有值，支持多值参数
// 按照 beego 的实现机制，会同时检查 URL 查询参数和 POST 表单数据
// 兼容 beego 的 GetStrings 方法
func (c *BaseController) GetStrings(key string, def ...[]string) []string {
	if c.Ctx == nil || c.Ctx.Request() == nil {
		if len(def) > 0 {
			return def[0]
		}
		return []string{}
	}

	var values []string

	// 1. 从查询参数中获取所有同名值
	queryValues := c.getQueryStrings(key)
	values = append(values, queryValues...)

	// 2. 从POST表单数据中获取所有同名值
	formValues := c.getFormStrings(key)
	values = append(values, formValues...)

	// 3. 如果没有找到任何值，尝试从路径参数获取
	if len(values) == 0 {
		if paramValue := c.GetParam(key); paramValue != "" {
			values = append(values, paramValue)
		}
	}

	// 如果仍然没有值且提供了默认值，返回默认值
	if len(values) == 0 && len(def) > 0 {
		return def[0]
	}

	return values
}

// getQueryStrings 从查询参数中获取所有同名的值
// 内部辅助方法，用于支持 GetStrings
func (c *BaseController) getQueryStrings(key string) []string {
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return []string{}
	}

	var values []string

	// 遍历所有查询参数，收集同名的值
	c.Ctx.Request().QueryArgs().VisitAll(func(k, v []byte) {
		if string(k) == key {
			values = append(values, string(v))
		}
	})

	return values
}

// getFormStrings 从POST表单数据中获取所有同名的值
// 内部辅助方法，用于支持 GetStrings
func (c *BaseController) getFormStrings(key string) []string {
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return []string{}
	}

	var values []string

	// 遍历所有POST表单参数，收集同名的值
	c.Ctx.Request().PostArgs().VisitAll(func(k, v []byte) {
		if string(k) == key {
			values = append(values, string(v))
		}
	})

	return values
}

// GetFormStrings 专门从表单数据中获取字符串数组 (扩展方法)
// 只检查 POST 表单数据，不检查查询参数
func (c *BaseController) GetFormStrings(key string, def ...[]string) []string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return []string{}
	}

	values := c.getFormStrings(key)

	if len(values) == 0 && len(def) > 0 {
		return def[0]
	}

	return values
}

// GetQueryStrings 专门从查询参数中获取字符串数组 (扩展方法)
// 只检查 URL 查询参数，不检查表单数据
func (c *BaseController) GetQueryStrings(key string, def ...[]string) []string {
	if c.Ctx == nil {
		if len(def) > 0 {
			return def[0]
		}
		return []string{}
	}

	values := c.getQueryStrings(key)

	if len(values) == 0 && len(def) > 0 {
		return def[0]
	}

	return values
}

// ParseFormToMap 解析表单数据到 map[string][]string (扩展方法)
// 返回所有表单数据的映射，包括查询参数和POST数据
func (c *BaseController) ParseFormToMap() map[string][]string {
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return make(map[string][]string)
	}

	result := make(map[string][]string)

	// 解析查询参数
	c.Ctx.Request().QueryArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		value := string(v)
		result[key] = append(result[key], value)
	})

	// 解析POST表单数据
	c.Ctx.Request().PostArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		value := string(v)
		result[key] = append(result[key], value)
	})

	return result
}

// GetFormValues 获取表单的所有值 (类似 http.Request.Form)
// 兼容标准库的 url.Values 类型
func (c *BaseController) GetFormValues() url.Values {
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return make(url.Values)
	}

	values := make(url.Values)

	// 添加查询参数
	c.Ctx.Request().QueryArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		value := string(v)
		values.Add(key, value)
	})

	// 添加POST表单数据
	c.Ctx.Request().PostArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		value := string(v)
		values.Add(key, value)
	})

	return values
}

// HasFormValue 检查是否存在指定的表单参数 (扩展方法)
// 同时检查查询参数和POST表单数据
func (c *BaseController) HasFormValue(key string) bool {
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return false
	}

	// 检查查询参数
	if c.Ctx.Request().QueryArgs().Has(key) {
		return true
	}

	// 检查POST表单数据
	if c.Ctx.Request().PostArgs().Has(key) {
		return true
	}

	// 检查路径参数
	if c.GetParam(key) != "" {
		return true
	}

	return false
}

// ============= Beego兼容的数据绑定方法 =============

// BindYAML 绑定YAML数据到结构体（Beego兼容）
// 读取HTTP请求体中的YAML数据并将其绑定到指定的Go结构体对象
func (c *BaseController) BindYAML(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.BindYAML(obj)
}

// BindForm 绑定表单数据到结构体（Beego兼容）
// 读取HTTP表单数据（包括查询参数和POST表单）并将其绑定到指定的Go结构体对象
func (c *BaseController) BindForm(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}
	return c.Ctx.BindForm(obj)
}

// BindProtobuf 绑定Protobuf数据到消息对象（Beego兼容）
// 读取HTTP请求体中的Protobuf二进制数据并将其绑定到指定的Proto消息对象
func (c *BaseController) BindProtobuf(obj any) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")  
	}
	return c.Ctx.BindProtobuf(obj)
}

// ============= 表单数据映射方法 =============

// GetMap 获取所有表单数据并返回 map[string]any
// 支持排除指定字段，同时处理查询参数和POST表单数据
// 兼容 beego 框架的 GetMap 方法
func (c *BaseController) GetMap(exclude ...string) (data map[string]any) {
	data = make(map[string]any)
	
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return data
	}

	// 创建排除字段的快速查找map
	excludeMap := make(map[string]bool)
	for _, key := range exclude {
		excludeMap[key] = true
	}

	// 收集所有参数到临时map中，支持多值
	allParams := make(map[string][]string)

	// 1. 收集查询参数
	c.Ctx.Request().QueryArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		if !excludeMap[key] {
			value := string(v)
			allParams[key] = append(allParams[key], value)
		}
	})

	// 2. 收集POST表单数据
	c.Ctx.Request().PostArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		if !excludeMap[key] {
			value := string(v)
			allParams[key] = append(allParams[key], value)
		}
	})

	// 3. 处理路径参数（自动获取所有路径参数）
	// 使用 Context 的 ParamMap() 方法自动获取所有实际存在的路径参数
	// 路径参数通常是单值的，优先级低于查询参数和表单参数
	if c.Ctx != nil {
		pathParams := c.Ctx.ParamMap()
		for paramName, paramValue := range pathParams {
			if !excludeMap[paramName] && paramValue != "" {
				// 路径参数优先级较低，只有在其他地方没有该参数时才添加
				if _, exists := allParams[paramName]; !exists {
					allParams[paramName] = []string{paramValue}
				}
			}
		}
	}
	
	// 4. 转换为 map[string]any，智能处理单值和多值
	for key, values := range allParams {
		if len(values) == 0 {
			continue
		} else if len(values) == 1 {
			// 单值情况，尝试智能类型转换
			data[key] = c.convertToAny(values[0])
		} else {
			// 多值情况，转换为[]any
			anyValues := make([]any, len(values))
			for i, v := range values {
				anyValues[i] = c.convertToAny(v)
			}
			data[key] = anyValues
		}
	}

	return data
}

// convertToAny 智能转换字符串到合适的类型
// 内部辅助方法，用于支持 GetMap 的类型转换
func (c *BaseController) convertToAny(value string) any {
	if value == "" {
		return ""
	}

	// 尝试转换为布尔值
	if lowerVal := strings.ToLower(value); lowerVal == "true" || lowerVal == "false" {
		return lowerVal == "true"
	}

	// 尝试转换为整数
	if intVal, err := strconv.Atoi(value); err == nil {
		return intVal
	}

	// 尝试转换为浮点数
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal
	}

	// 默认返回字符串
	return value
}

// GetMapNoPathParams 获取表单数据但不包含路径参数
// 只收集查询参数和POST表单数据，排除路径参数
// 适用于需要明确区分表单数据和路径参数的场景
func (c *BaseController) GetMapNoPathParams(exclude ...string) (data map[string]any) {
	data = make(map[string]any)
	
	if c.Ctx == nil || c.Ctx.Request() == nil {
		return data
	}

	// 创建排除字段的快速查找map
	excludeMap := make(map[string]bool)
	for _, key := range exclude {
		excludeMap[key] = true
	}

	// 收集所有参数到临时map中，支持多值
	allParams := make(map[string][]string)

	// 1. 收集查询参数
	c.Ctx.Request().QueryArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		if !excludeMap[key] {
			value := string(v)
			allParams[key] = append(allParams[key], value)
		}
	})

	// 2. 收集POST表单数据
	c.Ctx.Request().PostArgs().VisitAll(func(k, v []byte) {
		key := string(k)
		if !excludeMap[key] {
			value := string(v)
			allParams[key] = append(allParams[key], value)
		}
	})

	// 注意：这个方法故意不包含路径参数

	// 3. 转换为 map[string]any，智能处理单值和多值
	for key, values := range allParams {
		if len(values) == 0 {
			continue
		} else if len(values) == 1 {
			// 单值情况，尝试智能类型转换
			data[key] = c.convertToAny(values[0])
		} else {
			// 多值情况，转换为[]any
			anyValues := make([]any, len(values))
			for i, v := range values {
				anyValues[i] = c.convertToAny(v)
			}
			data[key] = anyValues
		}
	}

	return data
}
// ============= 文件处理和请求体方法 =============

// SaveToFile 将上传的文件保存到指定路径（Beego兼容）
// fromFile: 表单中文件字段名
// toFile: 保存到的文件路径
// 兼容 beego 的 SaveToFile 方法
func (c *BaseController) SaveToFile(fromFile, toFile string) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}

	// 获取上传的文件
	file, fileHeader, err := c.Ctx.FormFile(fromFile)
	if err != nil {
		return fmt.Errorf("failed to get form file '%s': %v", fromFile, err)
	}
	defer file.Close()

	// 确保目标目录存在
	dir := filepath.Dir(toFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %v", dir, err)
	}

	// 创建目标文件
	dst, err := os.OpenFile(toFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file '%s': %v", toFile, err)
	}
	defer dst.Close()

	// 复制文件内容
	_, err = io.Copy(dst, file)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	// 记录文件信息到日志（可选）
	_ = fileHeader // 使用 fileHeader 避免未使用变量警告

	return nil
}

// SaveToFileWithBuffer 使用指定缓冲区将上传的文件保存到指定路径
// fromFile: 表单中文件字段名
// toFile: 保存到的文件路径  
// buf: 复制时使用的缓冲区
// 这是 SaveToFile 的高性能版本，允许重用缓冲区
func (c *BaseController) SaveToFileWithBuffer(fromFile, toFile string, buf []byte) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}

	// 获取上传的文件
	file, fileHeader, err := c.Ctx.FormFile(fromFile)
	if err != nil {
		return fmt.Errorf("failed to get form file '%s': %v", fromFile, err)
	}
	defer file.Close()

	// 确保目标目录存在
	dir := filepath.Dir(toFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %v", dir, err)
	}

	// 创建目标文件
	dst, err := os.OpenFile(toFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create destination file '%s': %v", toFile, err)
	}
	defer dst.Close()

	// 使用指定缓冲区复制文件内容
	if len(buf) > 0 {
		_, err = io.CopyBuffer(dst, file, buf)
	} else {
		_, err = io.Copy(dst, file)
	}
	
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	// 记录文件信息到日志（可选）
	_ = fileHeader // 使用 fileHeader 避免未使用变量警告

	return nil
}

// GetRequestBody 获取原始请求体数据
// 返回请求体的字节数据和可能的错误
// 注意：请求体只能读取一次，重复调用会返回空数据
func (c *BaseController) GetRequestBody() ([]byte, error) {
	if c.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	// 使用 Context 的 RawBody 方法读取请求体
	return c.Ctx.RawBody()
}

// GetRequestBodyString 获取请求体的字符串形式
// 这是 GetRequestBody 的便捷方法
func (c *BaseController) GetRequestBodyString() (string, error) {
	body, err := c.GetRequestBody()
	if err != nil {
		return "", err
	}
	return string(body), nil
}