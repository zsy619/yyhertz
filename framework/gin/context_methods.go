// Package gin - Context 方法实现
// Context相关的所有方法实现，包括中间件控制、参数获取、数据绑定、响应渲染等

package gin

import (
	"fmt"
	"strings"

	"github.com/zsy619/yyhertz/framework/gin/binding"
	"github.com/zsy619/yyhertz/framework/gin/render"
)

// SetEngine 设置当前上下文的引擎实例。
// 参数：
//   - engine: 指向 Engine 结构体的指针，表示要设置的引擎。
//
// 说明：
//
//	该方法用于将指定的引擎实例与当前上下文关联，后续操作将基于此引擎执行。
func (c *Context) SetEngine(engine *Engine) {
	c.engine = engine
}

// =============================================================================
// 中间件控制方法
// =============================================================================

// Next 执行处理链中的下一个处理函数
//
// 在中间件中调用Next()会继续执行后续的中间件或最终的处理函数。
// 这是实现中间件链式调用的核心方法。
//
// 示例：
//
//	func middleware(c *gin.Context) {
//		// 前置处理逻辑
//		log.Println("Before request")
//
//		c.Next() // 执行后续处理函数
//
//		// 后置处理逻辑
//		log.Println("After request")
//	}
func (c *Context) Next() {
	c.index++
	// 边界检查和中止状态验证，防止数组越界
	for c.index >= 0 && c.index < int8(len(c.handlers)) && !c.IsAborted() {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort 终止当前请求的处理链
//
// 调用Abort()后，不会执行后续的处理函数，但已经执行的中间件
// 的后置逻辑（Next()之后的代码）仍会继续执行。
//
// 示例：
//
//	func authMiddleware(c *gin.Context) {
//		if !isAuthenticated(c) {
//			c.Abort()
//			return
//		}
//		c.Next()
//	}
func (c *Context) Abort() {
	// 使用安全的中止索引，避免int8溢出问题
	// 设为处理函数链长度，确保不会再执行任何处理函数
	c.index = int8(len(c.handlers))
}

// AbortWithStatus 终止处理链并设置HTTP状态码
//
// 除了终止处理链外，还会设置响应的HTTP状态码。
//
// 参数：
//   - code: HTTP状态码
//
// 示例：
//
//	c.AbortWithStatus(401) // 未授权
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

// AbortWithStatusJSON 终止处理链并返回JSON错误响应
//
// 这是一个便捷方法，同时设置状态码和JSON响应体。
//
// 参数：
//   - code: HTTP状态码
//   - jsonObj: 要序列化为JSON的对象
//
// 示例：
//
//	c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
func (c *Context) AbortWithStatusJSON(code int, jsonObj any) {
	c.Abort()
	c.JSON(code, jsonObj)
}

// IsAborted 检查当前请求是否已被终止
//
// 返回值：
//   - bool: 如果请求已被终止返回true，否则返回false
//
// 示例：
//
//	if c.IsAborted() {
//		return // 请求已被终止，直接返回
//	}
func (c *Context) IsAborted() bool {
	// 改进的中止状态检测：处理溢出和边界情况
	// 如果index >= handlers长度，或者index为负数（溢出情况），都认为已中止
	handlersLen := int8(len(c.handlers))
	return c.index >= handlersLen || c.index < -1
}

// =============================================================================
// 测试和调试辅助方法
// =============================================================================

// GetIndex 获取当前中间件索引（仅用于测试和调试）
func (c *Context) GetIndex() int8 {
	return c.index
}

// GetHandlersCount 获取处理函数链长度（仅用于测试和调试）
func (c *Context) GetHandlersCount() int {
	return len(c.handlers)
}

// =============================================================================
// 键值存储方法
// =============================================================================

// Set 在上下文中存储键值对
//
// 用于在中间件和处理函数之间传递数据。存储的数据仅在当前请求
// 的生命周期内有效。
//
// 参数：
//   - key: 键名
//   - value: 值，可以是任意类型
//
// 示例：
//
//	c.Set("user_id", 123)
//	c.Set("username", "john")
//	c.Set("user", userObj)
func (c *Context) Set(key string, value any) {
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

// Get 从上下文中获取键值对
//
// 返回指定键对应的值和是否存在的标志。
//
// 参数：
//   - key: 键名
//
// 返回值：
//   - value: 键对应的值，如果不存在则为nil
//   - exists: 键是否存在的布尔值
//
// 示例：
//
//	if userID, exists := c.Get("user_id"); exists {
//		// 处理用户ID
//		id := userID.(int)
//	}
func (c *Context) Get(key string) (value any, exists bool) {
	if c.Keys != nil {
		value, exists = c.Keys[key]
	}
	return
}

// MustGet 获取键值对，如果不存在则panic
//
// 与Get不同，MustGet假设键一定存在，如果不存在会panic。
// 适用于确信键存在的场景。
//
// 参数：
//   - key: 键名
//
// 返回值：
//   - any: 键对应的值
//
// 示例：
//
//	userID := c.MustGet("user_id").(int)
func (c *Context) MustGet(key string) any {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString 获取字符串类型的值
//
// 尝试将键对应的值转换为字符串。如果键不存在或类型不匹配，
// 返回空字符串。
//
// 参数：
//   - key: 键名
//
// 返回值：
//   - string: 字符串值，不存在或类型不匹配时返回空字符串
//
// 示例：
//
//	username := c.GetString("username")
func (c *Context) GetString(key string) string {
	if value, exists := c.Get(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// =============================================================================
// 参数获取方法
// =============================================================================

// Param 获取路径参数
//
// 从URL路径中获取命名参数的值。路径参数在路由注册时通过
// 冒号语法定义，如："/user/:id"
//
// 参数：
//   - key: 参数名
//
// 返回值：
//   - string: 参数值，如果不存在返回空字符串
//
// 示例：
//
//	// 路由: /user/:id
//	// 请求: /user/123
//	id := c.Param("id") // 返回 "123"
func (c *Context) Param(key string) string {
	value, _ := c.Params.Get(key)
	return value
}

// Query 获取查询参数
//
// 从URL查询字符串中获取参数值。
//
// 参数：
//   - key: 参数名
//
// 返回值：
//   - string: 参数值，如果不存在返回空字符串
//
// 示例：
//
//	// 请求: /search?q=golang&page=1
//	q := c.Query("q")       // 返回 "golang"
//	page := c.Query("page") // 返回 "1"
func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

// DefaultQuery 获取查询参数，支持默认值
//
// 从URL查询字符串中获取参数值，如果参数不存在则返回默认值。
//
// 参数：
//   - key: 参数名
//   - defaultValue: 默认值
//
// 返回值：
//   - string: 参数值或默认值
//
// 示例：
//
//	// 请求: /search?q=golang
//	q := c.DefaultQuery("q", "")         // 返回 "golang"
//	page := c.DefaultQuery("page", "1")  // 返回 "1"（默认值）
func (c *Context) DefaultQuery(key, defaultValue string) string {
	if value, exists := c.GetQuery(key); exists {
		return value
	}
	return defaultValue
}

// GetQuery 获取查询参数并返回是否存在
//
// 从URL查询字符串中获取参数值，同时返回参数是否存在的标志。
//
// 参数：
//   - key: 参数名
//
// 返回值：
//   - string: 参数值
//   - bool: 参数是否存在
//
// 示例：
//
//	if value, exists := c.GetQuery("optional_param"); exists {
//		// 处理可选参数
//	}
func (c *Context) GetQuery(key string) (string, bool) {
	if values := c.RequestContext.Request.URI().QueryArgs(); values != nil {
		if value := values.Peek(key); value != nil {
			return string(value), true
		}
	}
	return "", false
}

// GetHeader 获取请求头
//
// 从HTTP请求头中获取指定字段的值。
//
// 参数：
//   - key: 请求头字段名
//
// 返回值：
//   - string: 请求头的值，如果不存在返回空字符串
//
// 示例：
//
//	userAgent := c.GetHeader("User-Agent")
//	authToken := c.GetHeader("Authorization")
//	contentType := c.GetHeader("Content-Type")
func (c *Context) GetHeader(key string) string {
	return string(c.RequestContext.Request.Header.Peek(key))
}

// =============================================================================
// 数据绑定方法
// =============================================================================

// Bind 绑定请求数据到结构体（会验证）
//
// 根据Content-Type自动选择合适的绑定方式，并进行数据验证。
// 如果绑定或验证失败，会自动返回400错误响应。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
//
// 示例：
//
//	type User struct {
//		Name  string `json:"name" binding:"required"`
//		Email string `json:"email" binding:"required,email"`
//	}
//
//	var user User
//	if err := c.Bind(&user); err != nil {
//		// 错误已自动处理，返回400响应
//		return
//	}
func (c *Context) Bind(obj any) error {
	b := binding.Default(string(c.RequestContext.Request.Method()), c.GetHeader("Content-Type"))
	return c.MustBindWithHertz(obj, b)
}

// BindJSON 绑定JSON数据到结构体（会验证）
//
// 专门用于绑定JSON格式的请求体数据。如果绑定或验证失败，
// 会自动返回400错误响应。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
//
// 示例：
//
//	var user User
//	if err := c.BindJSON(&user); err != nil {
//		// 错误已自动处理
//		return
//	}
func (c *Context) BindJSON(obj any) error {
	return c.MustBindWithHertz(obj, binding.JSON)
}

// ShouldBindJSON 绑定JSON数据（不会自动返回错误）
//
// 与BindJSON类似，但不会自动返回错误响应，需要手动处理错误。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
//
// 示例：
//
//	var user User
//	if err := c.ShouldBindJSON(&user); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//		return
//	}
func (c *Context) ShouldBindJSON(obj any) error {
	return c.ShouldBindWithHertz(obj, binding.JSON)
}

// ShouldBindQuery 绑定查询参数到结构体
//
// 将URL查询参数绑定到结构体，不会自动返回错误响应。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
//
// 示例：
//
//	type SearchQuery struct {
//		Query string `form:"q" binding:"required"`
//		Page  int    `form:"page"`
//	}
//
//	var query SearchQuery
//	if err := c.ShouldBindQuery(&query); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//		return
//	}
func (c *Context) ShouldBindQuery(obj any) error {
	return c.ShouldBindWithHertz(obj, binding.Query)
}

// ShouldBindUri 绑定URI参数到结构体
//
// 将路径参数绑定到结构体，不会自动返回错误响应。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
//
// 示例：
//
//	type UserID struct {
//		ID int `uri:"id" binding:"required"`
//	}
//
//	var userID UserID
//	if err := c.ShouldBindUri(&userID); err != nil {
//		c.JSON(400, gin.H{"error": err.Error()})
//		return
//	}
func (c *Context) ShouldBindUri(obj any) error {
	// URI绑定需要特殊处理，因为它使用不同的接口
	m := make(map[string][]string)
	for _, param := range c.Params {
		m[param.Key] = []string{param.Value}
	}

	// 调试日志：记录参数提取状态
	if IsDebugging() {
		DebugPrintRoute("ShouldBindUri", "参数映射", map[string]interface{}{
			"c.Params": c.Params,
			"映射表":      m,
			"目标类型":     fmt.Sprintf("%T", obj),
		})
	}

	err := binding.Uri.BindUri(m, obj)

	// 调试日志：记录绑定结果
	if IsDebugging() {
		if err != nil {
			DebugPrintRoute("ShouldBindUri", "绑定失败", map[string]interface{}{
				"错误信息": err.Error(),
				"绑定结果": obj,
			})
		} else {
			DebugPrintRoute("ShouldBindUri", "绑定成功", map[string]interface{}{
				"绑定结果": obj,
			})
		}
	}

	return err
}

// ParamCheckResult 参数检查结果
type ParamCheckResult struct {
	HasParams      bool              `json:"has_params"`      // 是否有参数
	ParamsCount    int               `json:"params_count"`    // 参数数量
	ParamsList     []Param           `json:"params_list"`     // 参数列表
	MissingParams  []string          `json:"missing_params"`  // 缺失的参数
	EmptyParams    []string          `json:"empty_params"`    // 空值参数
	ValidationInfo map[string]string `json:"validation_info"` // 验证信息
	Suggestions    []string          `json:"suggestions"`     // 建议
}

// CheckUriParams 检查URI参数状态
//
// 全面检查URI参数的提取状态，帮助诊断ShouldBindUri问题。
//
// 参数：
//   - expectedParams: 期望的参数名称列表
//
// 返回值：
//   - ParamCheckResult: 详细的检查结果
//
// 示例：
//
//	result := c.CheckUriParams([]string{"name", "id"})
//	if !result.HasParams {
//		c.JSON(400, gin.H{"error": "缺少URI参数", "check_result": result})
//		return
//	}
func (c *Context) CheckUriParams(expectedParams []string) ParamCheckResult {
	result := ParamCheckResult{
		ParamsList:     c.Params,
		ValidationInfo: make(map[string]string),
		Suggestions:    make([]string, 0),
	}

	// 基本信息
	result.HasParams = len(c.Params) > 0
	result.ParamsCount = len(c.Params)

	// 创建参数映射便于查找
	paramMap := make(map[string]string)
	for _, param := range c.Params {
		paramMap[param.Key] = param.Value
	}

	// 检查期望的参数
	for _, expectedParam := range expectedParams {
		if value, exists := paramMap[expectedParam]; exists {
			if value == "" {
				result.EmptyParams = append(result.EmptyParams, expectedParam)
				result.ValidationInfo[expectedParam] = "参数存在但值为空"
			} else {
				result.ValidationInfo[expectedParam] = fmt.Sprintf("参数正常，值: %s", value)
			}
		} else {
			result.MissingParams = append(result.MissingParams, expectedParam)
			result.ValidationInfo[expectedParam] = "参数缺失"
		}
	}

	// 生成建议
	if len(result.MissingParams) > 0 {
		result.Suggestions = append(result.Suggestions,
			fmt.Sprintf("缺少参数: %v，请检查路由定义和URL格式", result.MissingParams))
	}

	if len(result.EmptyParams) > 0 {
		result.Suggestions = append(result.Suggestions,
			fmt.Sprintf("参数值为空: %v，请确保URL中包含实际值", result.EmptyParams))
	}

	if len(c.Params) == 0 {
		result.Suggestions = append(result.Suggestions,
			"没有提取到任何参数，请检查:",
			"1. 路由是否正确定义（如 /:name/:id）",
			"2. 请求URL是否匹配路由模式",
			"3. 参数名称是否一致")
	}

	return result
}

// ShouldBind 根据Content-Type自动绑定（不会自动返回错误）
//
// 根据请求的Content-Type自动选择合适的绑定方式。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
func (c *Context) ShouldBind(obj any) error {
	b := binding.Default(string(c.RequestContext.Request.Method()), c.GetHeader("Content-Type"))
	return c.ShouldBindWithHertz(obj, b)
}

// ShouldBindWithHertz 使用Hertz绑定器绑定数据
func (c *Context) ShouldBindWithHertz(obj any, b binding.Binding) error {
	return b.Bind(c.RequestContext, obj)
}

// MustBindWithHertz 使用Hertz绑定器绑定数据（会自动返回错误）
func (c *Context) MustBindWithHertz(obj any, b binding.Binding) error {
	if err := c.ShouldBindWithHertz(obj, b); err != nil {
		c.AbortWithStatusJSON(400, H{"error": err.Error()})
		return err
	}
	return nil
}

// =============================================================================
// YAML 绑定方法
// =============================================================================

// BindYAML 绑定YAML数据到结构体（会验证并自动返回错误）
//
// 将请求体中的YAML数据绑定到结构体，验证失败时会自动返回400错误。
// 这是MustBindWith的快捷方式，使用YAML绑定器。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误（发生错误时已自动设置响应）
//
// 示例：
//
//	type User struct {
//		Name string `yaml:"name" binding:"required"`
//		Age  int    `yaml:"age" binding:"min=1"`
//	}
//
//	var user User
//	if err := c.BindYAML(&user); err != nil {
//		// 错误已自动处理
//		return
//	}
func (c *Context) BindYAML(obj any) error {
	return c.MustBindWithHertz(obj, binding.YAML)
}

// ShouldBindYAML 绑定YAML数据（不会自动返回错误）
//
// 将请求体中的YAML数据绑定到结构体，不会自动返回错误响应。
// 这是ShouldBindWith的快捷方式，使用YAML绑定器。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
//
// 示例：
//
//	type Config struct {
//		Host string `yaml:"host" binding:"required"`
//		Port int    `yaml:"port" binding:"min=1,max=65535"`
//	}
//
//	var config Config
//	if err := c.ShouldBindYAML(&config); err != nil {
//		c.YAML(400, gin.H{"error": err.Error()})
//		return
//	}
func (c *Context) ShouldBindYAML(obj any) error {
	return c.ShouldBindWithHertz(obj, binding.YAML)
}

// ShouldBindBodyWithYAML 从请求体绑定YAML数据
//
// 直接从请求体的字节数据中绑定YAML，不会自动返回错误响应。
// 适用于需要重复读取请求体或自定义错误处理的场景。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
//
// 示例：
//
//	type Settings struct {
//		Theme    string            `yaml:"theme"`
//		Features map[string]bool   `yaml:"features"`
//	}
//
//	var settings Settings
//	if err := c.ShouldBindBodyWithYAML(&settings); err != nil {
//		c.YAML(422, gin.H{
//			"error": "Invalid YAML format",
//			"details": err.Error(),
//		})
//		return
//	}
func (c *Context) ShouldBindBodyWithYAML(obj any) error {
	body := c.RequestContext.Request.Body()
	return binding.YAML.BindBody(body, obj)
}

// =============================================================================
// XML 绑定方法
// =============================================================================

// BindXML 绑定XML数据到结构体（会验证并自动返回错误）
//
// 将请求体中的XML数据绑定到结构体，验证失败时会自动返回400错误。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误（发生错误时已自动设置响应）
func (c *Context) BindXML(obj any) error {
	return c.MustBindWithHertz(obj, binding.XML)
}

// ShouldBindXML 绑定XML数据（不会自动返回错误）
//
// 将请求体中的XML数据绑定到结构体，不会自动返回错误响应。
//
// 参数：
//   - obj: 要绑定到的结构体指针
//
// 返回值：
//   - error: 绑定或验证错误
func (c *Context) ShouldBindXML(obj any) error {
	return c.ShouldBindWithHertz(obj, binding.XML)
}

// =============================================================================
// 错误处理方法
// =============================================================================

// AbortWithErrorMessage 终止处理链并记录错误消息
//
// 终止当前请求的处理链，设置状态码，并将错误添加到错误列表中。
//
// 参数：
//   - code: HTTP状态码
//   - err: 错误对象
func (c *Context) AbortWithErrorMessage(code int, err error) {
	c.AbortWithStatus(code)
	c.Errors = append(c.Errors, err)
}

// =============================================================================
// 响应渲染方法
// =============================================================================

// JSON 返回JSON响应
//
// 将对象序列化为JSON格式并返回给客户端。
//
// 参数：
//   - code: HTTP状态码
//   - obj: 要序列化的对象
//
// 示例：
//
//	c.JSON(200, gin.H{"message": "success", "data": user})
//	c.JSON(400, gin.H{"error": "validation failed"})
func (c *Context) JSON(code int, obj any) {
	c.Render(code, render.JSON{Data: obj})
}

// AsciiJSON 返回ASCII JSON响应
//
// 将非ASCII字符转换为Unicode转义序列的JSON响应。
// 这确保了输出只包含ASCII字符，适用于需要纯ASCII输出的场景。
//
// 参数：
//   - code: HTTP状态码
//   - obj: 要序列化的数据对象
//
// 示例：
//
//	c.AsciiJSON(200, gin.H{"message": "你好世界"})
//	// 输出: {"message":"\\u4f60\\u597d\\u4e16\\u754c"}
func (c *Context) AsciiJSON(code int, obj any) {
	c.Render(code, render.AsciiJSON{Data: obj})
}

// XML 返回XML响应
//
// 将对象序列化为XML格式并返回给客户端。
//
// 参数：
//   - code: HTTP状态码
//   - obj: 要序列化的对象
//
// 示例：
//
//	c.XML(200, gin.H{"message": "success", "data": user})
//	c.XML(400, gin.H{"error": "validation failed"})
func (c *Context) XML(code int, obj any) {
	c.Render(code, render.XML{Data: obj})
}

// YAML 返回YAML响应
//
// 将对象序列化为YAML格式并返回给客户端。
// YAML格式具有良好的可读性，适合配置文件和人类阅读。
//
// 参数：
//   - code: HTTP状态码
//   - obj: 要序列化的对象
//
// 示例：
//
//	c.YAML(200, gin.H{"message": "success", "data": user})
//	c.YAML(400, gin.H{"error": "validation failed"})
func (c *Context) YAML(code int, obj any) {
	c.Render(code, render.YAML{Data: obj})
}

// String 返回字符串响应
//
// 返回格式化的字符串响应。
//
// 参数：
//   - code: HTTP状态码
//   - format: 格式化字符串
//   - values: 格式化参数
//
// 示例：
//
//	c.String(200, "Hello %s!", "World")
//	c.String(404, "Page not found")
func (c *Context) String(code int, format string, values ...any) {
	c.Render(code, render.String{Format: format, Data: values})
}

// HTML 返回HTML响应
//
// 渲染HTML模板并返回。
//
// 参数：
//   - code: HTTP状态码
//   - name: 模板名称
//   - obj: 模板数据
//
// 示例：
//
//	c.HTML(200, "index.html", gin.H{"title": "Home"})
func (c *Context) HTML(code int, name string, obj any) {
	// 检查是否已设置HTML渲染器
	if c.engine.HTMLRender == nil {
		c.String(code, "HTML template rendering not configured. Use LoadHTMLGlob() or LoadHTMLFiles() first.")
		return
	}

	// 使用渲染器创建实例并渲染
	instance := c.engine.HTMLRender.Instance(name, obj)
	c.Render(code, instance)
}

// Data 返回原始数据响应
//
// 返回原始字节数据，通常用于文件下载或二进制数据。
//
// 参数：
//   - code: HTTP状态码
//   - contentType: 内容类型
//   - data: 原始数据
//
// 示例：
//
//	c.Data(200, "application/octet-stream", fileData)
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{ContentType: contentType, Data: data})
}

// Render 使用指定的渲染器渲染响应
//
// 使用给定的渲染器来生成响应内容。
//
// 参数：
//   - code: HTTP状态码
//   - r: 渲染器
//
// 示例：
//
//	c.Render(200, render.JSON{Data: data})
func (c *Context) Render(code int, r render.Render) {
	c.Status(code)
	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.RequestContext)
		c.Abort()
		return
	}
	if err := r.Render(c.RequestContext); err != nil {
		panic(err)
	}
}

// File 返回文件响应
//
// 发送文件给客户端，通常用于文件下载。
//
// 参数：
//   - filepath: 文件路径
//
// 示例：
//
//	c.File("./uploads/document.pdf")
func (c *Context) File(filepath string) {
	c.RequestContext.Response.Header.Del("Content-Type")
	c.RequestContext.File(filepath)
}

// Header 设置响应头
//
// 设置HTTP响应头的值。
//
// 参数：
//   - key: 响应头名称
//   - value: 响应头值
//
// 示例：
//
//	c.Header("Cache-Control", "no-cache")
//	c.Header("X-Custom-Header", "custom-value")
func (c *Context) Header(key, value string) {
	c.Response.Header.Set(key, value)
}

// Status 设置HTTP状态码
//
// 设置响应的HTTP状态码。
//
// 参数：
//   - code: HTTP状态码
//
// 示例：
//
//	c.Status(201) // Created
//	c.Status(404) // Not Found
func (c *Context) Status(code int) {
	c.Response.SetStatusCode(code)
}

// =============================================================================
// 表单映射方法 (QueryMap & PostFormMap)
// =============================================================================

// QueryMap 从查询字符串中解析映射类型的参数
//
// 解析形如 "key[subkey]=value" 格式的查询参数到映射中。
// 这对于处理复杂的表单数据结构非常有用。
//
// 参数：
//   - key: 映射的基础键名
//
// 返回值：
//   - map[string]string: 解析后的键值映射
//
// 示例：
//
//	// 查询: ?ids[first]=123&ids[second]=456
//	ids := c.QueryMap("ids")
//	// 结果: map[string]string{"first": "123", "second": "456"}
func (c *Context) QueryMap(key string) map[string]string {
	dicts, _ := c.GetQueryMap(key)
	return dicts
}

// GetQueryMap 从查询字符串中解析映射类型的参数并返回是否存在
//
// 与 QueryMap 类似，但会返回映射是否存在的标志。
//
// 参数：
//   - key: 映射的基础键名
//
// 返回值：
//   - map[string]string: 解析后的键值映射
//   - bool: 映射是否存在
//
// 示例：
//
//	if ids, exists := c.GetQueryMap("ids"); exists {
//		// 处理映射数据
//		for k, v := range ids {
//			fmt.Printf("%s: %s\n", k, v)
//		}
//	}
func (c *Context) GetQueryMap(key string) (map[string]string, bool) {
	// 从 Hertz 请求中获取查询参数
	queryArgs := c.RequestContext.Request.URI().QueryArgs()
	if queryArgs == nil {
		return make(map[string]string), false
	}

	// 将查询参数转换为标准格式
	formData := make(map[string][]string)
	queryArgs.VisitAll(func(k, v []byte) {
		key := string(k)
		value := string(v)
		if existing, exists := formData[key]; exists {
			formData[key] = append(existing, value)
		} else {
			formData[key] = []string{value}
		}
	})

	return getMapFromFormData(formData, key)
}

// PostFormMap 从表单数据中解析映射类型的参数
//
// 解析 POST 表单中形如 "key[subkey]=value" 格式的参数到映射中。
// 这对于处理复杂的表单数据结构非常有用。
//
// 参数：
//   - key: 映射的基础键名
//
// 返回值：
//   - map[string]string: 解析后的键值映射
//
// 示例：
//
//	// 表单: names[first]=John&names[last]=Doe
//	names := c.PostFormMap("names")
//	// 结果: map[string]string{"first": "John", "last": "Doe"}
func (c *Context) PostFormMap(key string) map[string]string {
	dicts, _ := c.GetPostFormMap(key)
	return dicts
}

// GetPostFormMap 从表单数据中解析映射类型的参数并返回是否存在
//
// 与 PostFormMap 类似，但会返回映射是否存在的标志。
//
// 参数：
//   - key: 映射的基础键名
//
// 返回值：
//   - map[string]string: 解析后的键值映射
//   - bool: 映射是否存在
//
// 示例：
//
//	if names, exists := c.GetPostFormMap("names"); exists {
//		// 处理映射数据
//		firstName := names["first"]
//		lastName := names["last"]
//	}
func (c *Context) GetPostFormMap(key string) (map[string]string, bool) {
	// 从 Hertz 请求中获取表单数据
	var formData map[string][]string

	// 检查是否为 multipart 表单
	if c.RequestContext.Request.Header.IsPost() {
		// 获取 Content-Type
		contentType := string(c.RequestContext.Request.Header.ContentType())

		if strings.Contains(contentType, "multipart/form-data") {
			// 处理 multipart 表单
			multipartForm, err := c.RequestContext.MultipartForm()
			if err == nil && multipartForm != nil {
				formData = multipartForm.Value
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// 处理 URL 编码表单
			postArgs := c.RequestContext.Request.PostArgs()
			if postArgs != nil {
				formData = make(map[string][]string)
				postArgs.VisitAll(func(k, v []byte) {
					key := string(k)
					value := string(v)
					if existing, exists := formData[key]; exists {
						formData[key] = append(existing, value)
					} else {
						formData[key] = []string{value}
					}
				})
			}
		}
	}

	if formData == nil {
		return make(map[string]string), false
	}

	return getMapFromFormData(formData, key)
}

// getMapFromFormData 从表单数据中提取映射类型的参数
//
// 这是一个辅助函数，用于处理形如 "key[subkey]=value" 格式的表单数据。
// 它解析括号记法并构建嵌套的映射结构。
//
// 参数：
//   - data: 原始表单数据，key-value数组映射
//   - key: 要提取的映射基础键名
//
// 返回值：
//   - map[string]string: 解析后的键值映射
//   - bool: 映射是否存在
//
// 支持的格式：
//   - "users[0][name]=John" -> map["0"]map["name"] = "John"  (注：当前实现仅支持一级)
//   - "config[database]=localhost" -> map["database"] = "localhost"
//   - "tags[]=tag1&tags[]=tag2" -> map["0"]="tag1", map["1"]="tag2"
//
// 注意：
//   - 这个实现兼容 Gin 原版的行为
//   - 空括号会自动生成数字索引
//   - 特殊处理同名键的多个值（数组情况）
func getMapFromFormData(data map[string][]string, key string) (map[string]string, bool) {
	result := make(map[string]string)
	found := false

	// 精确匹配基础键
	if values, exists := data[key]; exists {
		// 直接键匹配，使用第一个值
		if len(values) > 0 {
			result[""] = values[0]
			found = true
		}
	}

	// 查找形如 "key[subkey]" 的模式
	keyPrefix := key + "["

	// 特殊处理：先检查空括号情况 "key[]"
	emptyBracketKey := key + "[]"
	if values, exists := data[emptyBracketKey]; exists {
		// 处理 key[]=value1&key[]=value2 这种数组格式
		for i, value := range values {
			result[fmt.Sprintf("%d", i)] = value
			found = true
		}
	}

	// 处理命名子键
	for formKey, values := range data {
		if strings.HasPrefix(formKey, keyPrefix) && strings.HasSuffix(formKey, "]") && formKey != emptyBracketKey {
			// 提取括号内的子键
			start := len(keyPrefix)
			end := len(formKey) - 1
			if start < end {
				subKey := formKey[start:end]

				// 使用第一个值（兼容 Gin 行为）
				if len(values) > 0 {
					result[subKey] = values[0]
					found = true
				}
			}
		}
	}

	return result, found
}

// =============================================================================
// 辅助函数
// =============================================================================

// bodyAllowedForStatus 检查状态码是否允许响应体
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
