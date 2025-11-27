// Package gin - 路由参数约束系统
// 提供强大的路由参数验证和类型约束功能
package gin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// =============================================================================
// 约束类型定义
// =============================================================================

// RouteConstraint 路由约束接口
type RouteConstraint interface {
	// Match 检查参数值是否匹配约束
	Match(paramName, value string) bool
	
	// Error 返回约束错误信息
	Error(paramName, value string) string
	
	// Transform 转换参数值（可选）
	Transform(value string) (any, error)
}

// ConstraintType 约束类型
type ConstraintType int

const (
	ConstraintTypeString ConstraintType = iota
	ConstraintTypeInt
	ConstraintTypeFloat
	ConstraintTypeBool
	ConstraintTypeDate
	ConstraintTypeRegex
	ConstraintTypeEnum
	ConstraintTypeRange
	ConstraintTypeLength
	ConstraintTypeCustom
)

// =============================================================================
// 内置约束实现
// =============================================================================

// IntConstraint 整数约束
type IntConstraint struct {
	Min *int64 // 最小值
	Max *int64 // 最大值
}

// NewIntConstraint 创建整数约束
func NewIntConstraint(min, max *int64) *IntConstraint {
	return &IntConstraint{Min: min, Max: max}
}

// Match 检查是否为有效整数
func (c *IntConstraint) Match(paramName, value string) bool {
	num, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	
	if c.Min != nil && num < *c.Min {
		return false
	}
	
	if c.Max != nil && num > *c.Max {
		return false
	}
	
	return true
}

// Error 返回错误信息
func (c *IntConstraint) Error(paramName, value string) string {
	if c.Min != nil && c.Max != nil {
		return fmt.Sprintf("参数 %s 的值 '%s' 必须是 %d 到 %d 之间的整数", 
			paramName, value, *c.Min, *c.Max)
	} else if c.Min != nil {
		return fmt.Sprintf("参数 %s 的值 '%s' 必须是大于等于 %d 的整数", 
			paramName, value, *c.Min)
	} else if c.Max != nil {
		return fmt.Sprintf("参数 %s 的值 '%s' 必须是小于等于 %d 的整数", 
			paramName, value, *c.Max)
	}
	return fmt.Sprintf("参数 %s 的值 '%s' 必须是有效整数", paramName, value)
}

// Transform 转换为整数
func (c *IntConstraint) Transform(value string) (any, error) {
	return strconv.ParseInt(value, 10, 64)
}

// FloatConstraint 浮点数约束
type FloatConstraint struct {
	Min *float64 // 最小值
	Max *float64 // 最大值
}

// NewFloatConstraint 创建浮点数约束
func NewFloatConstraint(min, max *float64) *FloatConstraint {
	return &FloatConstraint{Min: min, Max: max}
}

// Match 检查是否为有效浮点数
func (c *FloatConstraint) Match(paramName, value string) bool {
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	
	if c.Min != nil && num < *c.Min {
		return false
	}
	
	if c.Max != nil && num > *c.Max {
		return false
	}
	
	return true
}

// Error 返回错误信息
func (c *FloatConstraint) Error(paramName, value string) string {
	if c.Min != nil && c.Max != nil {
		return fmt.Sprintf("参数 %s 的值 '%s' 必须是 %.2f 到 %.2f 之间的数字", 
			paramName, value, *c.Min, *c.Max)
	} else if c.Min != nil {
		return fmt.Sprintf("参数 %s 的值 '%s' 必须是大于等于 %.2f 的数字", 
			paramName, value, *c.Min)
	} else if c.Max != nil {
		return fmt.Sprintf("参数 %s 的值 '%s' 必须是小于等于 %.2f 的数字", 
			paramName, value, *c.Max)
	}
	return fmt.Sprintf("参数 %s 的值 '%s' 必须是有效数字", paramName, value)
}

// Transform 转换为浮点数
func (c *FloatConstraint) Transform(value string) (any, error) {
	return strconv.ParseFloat(value, 64)
}

// RegexConstraint 正则表达式约束
type RegexConstraint struct {
	Pattern *regexp.Regexp
	Message string // 自定义错误信息
}

// NewRegexConstraint 创建正则约束
func NewRegexConstraint(pattern string, message ...string) *RegexConstraint {
	re := regexp.MustCompile(pattern)
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return &RegexConstraint{Pattern: re, Message: msg}
}

// Match 检查是否匹配正则表达式
func (c *RegexConstraint) Match(paramName, value string) bool {
	return c.Pattern.MatchString(value)
}

// Error 返回错误信息
func (c *RegexConstraint) Error(paramName, value string) string {
	if c.Message != "" {
		return fmt.Sprintf("参数 %s 的值 '%s' %s", paramName, value, c.Message)
	}
	return fmt.Sprintf("参数 %s 的值 '%s' 格式不正确", paramName, value)
}

// Transform 原样返回字符串
func (c *RegexConstraint) Transform(value string) (any, error) {
	return value, nil
}

// EnumConstraint 枚举约束
type EnumConstraint struct {
	Values      []string // 允许的值
	CaseSensitive bool   // 是否区分大小写
}

// NewEnumConstraint 创建枚举约束
func NewEnumConstraint(values []string, caseSensitive bool) *EnumConstraint {
	return &EnumConstraint{Values: values, CaseSensitive: caseSensitive}
}

// Match 检查是否在枚举值中
func (c *EnumConstraint) Match(paramName, value string) bool {
	for _, v := range c.Values {
		if c.CaseSensitive {
			if v == value {
				return true
			}
		} else {
			if strings.EqualFold(v, value) {
				return true
			}
		}
	}
	return false
}

// Error 返回错误信息
func (c *EnumConstraint) Error(paramName, value string) string {
	return fmt.Sprintf("参数 %s 的值 '%s' 必须是以下值之一: %s", 
		paramName, value, strings.Join(c.Values, ", "))
}

// Transform 原样返回字符串
func (c *EnumConstraint) Transform(value string) (any, error) {
	return value, nil
}

// LengthConstraint 长度约束
type LengthConstraint struct {
	Min *int // 最小长度
	Max *int // 最大长度
}

// NewLengthConstraint 创建长度约束
func NewLengthConstraint(min, max *int) *LengthConstraint {
	return &LengthConstraint{Min: min, Max: max}
}

// Match 检查字符串长度
func (c *LengthConstraint) Match(paramName, value string) bool {
	length := len(value)
	
	if c.Min != nil && length < *c.Min {
		return false
	}
	
	if c.Max != nil && length > *c.Max {
		return false
	}
	
	return true
}

// Error 返回错误信息
func (c *LengthConstraint) Error(paramName, value string) string {
	if c.Min != nil && c.Max != nil {
		return fmt.Sprintf("参数 %s 的长度必须在 %d 到 %d 之间", 
			paramName, *c.Min, *c.Max)
	} else if c.Min != nil {
		return fmt.Sprintf("参数 %s 的长度必须至少为 %d", paramName, *c.Min)
	} else if c.Max != nil {
		return fmt.Sprintf("参数 %s 的长度不能超过 %d", paramName, *c.Max)
	}
	return fmt.Sprintf("参数 %s 长度不符合要求", paramName)
}

// Transform 原样返回字符串
func (c *LengthConstraint) Transform(value string) (any, error) {
	return value, nil
}

// DateConstraint 日期约束
type DateConstraint struct {
	Format string     // 日期格式
	Min    *time.Time // 最小日期
	Max    *time.Time // 最大日期
}

// NewDateConstraint 创建日期约束
func NewDateConstraint(format string, min, max *time.Time) *DateConstraint {
	return &DateConstraint{Format: format, Min: min, Max: max}
}

// Match 检查是否为有效日期
func (c *DateConstraint) Match(paramName, value string) bool {
	date, err := time.Parse(c.Format, value)
	if err != nil {
		return false
	}
	
	if c.Min != nil && date.Before(*c.Min) {
		return false
	}
	
	if c.Max != nil && date.After(*c.Max) {
		return false
	}
	
	return true
}

// Error 返回错误信息
func (c *DateConstraint) Error(paramName, value string) string {
	return fmt.Sprintf("参数 %s 的值 '%s' 必须是格式为 %s 的有效日期", 
		paramName, value, c.Format)
}

// Transform 转换为时间对象
func (c *DateConstraint) Transform(value string) (any, error) {
	return time.Parse(c.Format, value)
}

// =============================================================================
// 复合约束
// =============================================================================

// CompositeConstraint 复合约束
type CompositeConstraint struct {
	Constraints []RouteConstraint
	Operator    ConstraintOperator // AND/OR
}

// ConstraintOperator 约束操作符
type ConstraintOperator int

const (
	ConstraintOperatorAND ConstraintOperator = iota
	ConstraintOperatorOR
)

// NewCompositeConstraint 创建复合约束
func NewCompositeConstraint(operator ConstraintOperator, constraints ...RouteConstraint) *CompositeConstraint {
	return &CompositeConstraint{
		Constraints: constraints,
		Operator:    operator,
	}
}

// Match 检查复合约束
func (c *CompositeConstraint) Match(paramName, value string) bool {
	if c.Operator == ConstraintOperatorAND {
		// 所有约束都必须匹配
		for _, constraint := range c.Constraints {
			if !constraint.Match(paramName, value) {
				return false
			}
		}
		return true
	} else {
		// 至少一个约束匹配
		for _, constraint := range c.Constraints {
			if constraint.Match(paramName, value) {
				return true
			}
		}
		return false
	}
}

// Error 返回错误信息
func (c *CompositeConstraint) Error(paramName, value string) string {
	var errors []string
	for _, constraint := range c.Constraints {
		if !constraint.Match(paramName, value) {
			errors = append(errors, constraint.Error(paramName, value))
		}
	}
	
	if c.Operator == ConstraintOperatorAND {
		return strings.Join(errors, " 且 ")
	} else {
		return strings.Join(errors, " 或 ")
	}
}

// Transform 使用第一个匹配的约束进行转换
func (c *CompositeConstraint) Transform(value string) (any, error) {
	for _, constraint := range c.Constraints {
		if result, err := constraint.Transform(value); err == nil {
			return result, nil
		}
	}
	return value, fmt.Errorf("没有约束能够转换值: %s", value)
}

// =============================================================================
// 约束注册和管理
// =============================================================================

// ConstraintRegistry 约束注册表
type ConstraintRegistry struct {
	constraints map[string]map[string]RouteConstraint // [route][param] -> constraint
}

// NewConstraintRegistry 创建约束注册表
func NewConstraintRegistry() *ConstraintRegistry {
	return &ConstraintRegistry{
		constraints: make(map[string]map[string]RouteConstraint),
	}
}

// AddConstraint 添加路由参数约束
func (r *ConstraintRegistry) AddConstraint(route, param string, constraint RouteConstraint) {
	if r.constraints[route] == nil {
		r.constraints[route] = make(map[string]RouteConstraint)
	}
	r.constraints[route][param] = constraint
}

// GetConstraint 获取参数约束
func (r *ConstraintRegistry) GetConstraint(route, param string) (RouteConstraint, bool) {
	if routeConstraints, exists := r.constraints[route]; exists {
		if constraint, exists := routeConstraints[param]; exists {
			return constraint, true
		}
	}
	return nil, false
}

// ValidateParams 验证路由参数
func (r *ConstraintRegistry) ValidateParams(route string, params Params) []ConstraintError {
	var errors []ConstraintError
	
	routeConstraints, exists := r.constraints[route]
	if !exists {
		return errors // 没有约束，验证通过
	}
	
	for paramName, constraint := range routeConstraints {
		// 查找参数值
		paramValue := ""
		found := false
		for _, param := range params {
			if param.Key == paramName {
				paramValue = param.Value
				found = true
				break
			}
		}
		
		if !found {
			errors = append(errors, ConstraintError{
				ParamName: paramName,
				Value:     "",
				Message:   fmt.Sprintf("缺少必需的参数: %s", paramName),
			})
			continue
		}
		
		if !constraint.Match(paramName, paramValue) {
			errors = append(errors, ConstraintError{
				ParamName: paramName,
				Value:     paramValue,
				Message:   constraint.Error(paramName, paramValue),
			})
		}
	}
	
	return errors
}

// ConstraintError 约束验证错误
type ConstraintError struct {
	ParamName string // 参数名
	Value     string // 参数值
	Message   string // 错误信息
}

// Error 实现error接口
func (e ConstraintError) Error() string {
	return e.Message
}

// =============================================================================
// 预定义约束
// =============================================================================

// 常用约束快捷创建函数
var (
	// Int 创建整数约束
	Int = func(min, max *int64) RouteConstraint {
		return NewIntConstraint(min, max)
	}
	
	// Float 创建浮点数约束
	Float = func(min, max *float64) RouteConstraint {
		return NewFloatConstraint(min, max)
	}
	
	// Regex 创建正则约束
	Regex = func(pattern string, message ...string) RouteConstraint {
		return NewRegexConstraint(pattern, message...)
	}
	
	// Enum 创建枚举约束
	Enum = func(values []string, caseSensitive bool) RouteConstraint {
		return NewEnumConstraint(values, caseSensitive)
	}
	
	// Length 创建长度约束
	Length = func(min, max *int) RouteConstraint {
		return NewLengthConstraint(min, max)
	}
	
	// Date 创建日期约束
	Date = func(format string, min, max *time.Time) RouteConstraint {
		return NewDateConstraint(format, min, max)
	}
)

// 预定义常用约束
var (
	// PositiveInt 正整数约束
	PositiveInt = NewIntConstraint(func() *int64 { i := int64(1); return &i }(), nil)
	
	// NonNegativeInt 非负整数约束
	NonNegativeInt = NewIntConstraint(func() *int64 { i := int64(0); return &i }(), nil)
	
	// Email 邮箱格式约束
	Email = NewRegexConstraint(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "必须是有效的邮箱地址")
	
	// URL URL格式约束
	URL = NewRegexConstraint(`^https?://[^\s/$.?#].[^\s]*$`, "必须是有效的URL")
	
	// UUID UUID格式约束
	UUID = NewRegexConstraint(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, "必须是有效的UUID")
	
	// Alpha 纯字母约束
	Alpha = NewRegexConstraint(`^[a-zA-Z]+$`, "只能包含字母")
	
	// AlphaNum 字母数字约束
	AlphaNum = NewRegexConstraint(`^[a-zA-Z0-9]+$`, "只能包含字母和数字")
	
	// DateISO ISO日期格式约束
	DateISO = NewDateConstraint("2006-01-02", nil, nil)
	
	// DateTime ISO日期时间格式约束
	DateTime = NewDateConstraint("2006-01-02T15:04:05Z", nil, nil)
)

// =============================================================================
// Engine集成
// =============================================================================

// 在Engine中添加约束支持
func (engine *Engine) AddConstraint(route, param string, constraint RouteConstraint) *Engine {
	if engine.router == nil {
		engine.router = NewRouterEngine()
	}
	if engine.router.constraints == nil {
		engine.router.constraints = NewConstraintRegistry()
	}
	engine.router.constraints.AddConstraint(route, param, constraint)
	return engine
}

// WithConstraints 链式添加多个约束
func (engine *Engine) WithConstraints(route string, constraints map[string]RouteConstraint) *Engine {
	for param, constraint := range constraints {
		engine.AddConstraint(route, param, constraint)
	}
	return engine
}

// 扩展RouterEngine支持约束
func (r *RouterEngine) validateConstraints(route string, params Params) []ConstraintError {
	if r.constraints == nil {
		return nil
	}
	return r.constraints.ValidateParams(route, params)
}

// =============================================================================
// 约束验证中间件
// =============================================================================

// ConstraintValidator 约束验证中间件
func ConstraintValidator() HandlerFunc {
	return func(c *Context) {
		// 获取当前路由
		route := c.Request().URL.Path
		
		// 验证参数约束
		if c.engine.router != nil && c.engine.router.constraints != nil {
			errors := c.engine.router.validateConstraints(route, c.Params)
			if len(errors) > 0 {
				// 参数验证失败
				errorMessages := make([]string, len(errors))
				for i, err := range errors {
					errorMessages[i] = err.Message
				}
				
				c.JSON(400, H{
					"error":   "参数验证失败",
					"details": errorMessages,
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// =============================================================================
// 使用示例和工具函数
// =============================================================================

// 示例：如何使用约束系统
/*
func ExampleUsage() {
	r := gin.New()
	
	// 添加约束验证中间件
	r.Use(gin.ConstraintValidator())
	
	// 添加路由约束
	r.AddConstraint("/user/:id", "id", gin.PositiveInt)
	r.AddConstraint("/user/:id/posts/:page", "page", gin.NonNegativeInt)
	r.AddConstraint("/user/:email", "email", gin.Email)
	
	// 复合约束示例
	r.AddConstraint("/product/:category", "category", 
		gin.Enum([]string{"electronics", "clothing", "books"}, false))
	
	// 注册路由
	r.GET("/user/:id", func(c *gin.Context) {
		id := c.Param("id") // 已经验证为正整数
		c.JSON(200, gin.H{"user_id": id})
	})
	
	r.GET("/user/:email", func(c *gin.Context) {
		email := c.Param("email") // 已经验证为有效邮箱
		c.JSON(200, gin.H{"email": email})
	})
}
*/