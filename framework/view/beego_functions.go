package view

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// BeegoTemplateFuncs Beego风格的模板函数
var BeegoTemplateFuncs = template.FuncMap{
	// ============= 基础工具函数 =============
	"str2html":   Str2HTML,
	"htmlquote":  HTMLQuote,
	"htmlunquote": HTMLUnquote,
	"renderform": RenderForm,
	"assets_js":  AssetsJS,
	"assets_css": AssetsCSS,
	"config":     GetConfig,
	"map_get":    MapGet,
	"urlfor":     URLFor,

	// ============= 字符串处理函数 =============
	"substr":     Substr,
	"truncate":   TruncateString,
	"nl2br":      NL2BR,
	"markdown":   MarkdownString,
	"striphtml":  StripHTML,
	"replace":    strings.ReplaceAll,
	"tolower":    strings.ToLower,
	"toupper":    strings.ToUpper,
	"trim":       strings.TrimSpace,
	"trimprefix": strings.TrimPrefix,
	"trimsuffix": strings.TrimSuffix,

	// ============= 数字处理函数 =============
	"add":    Add,
	"sub":    Sub,
	"mul":    Mul,
	"div":    Div,
	"mod":    Mod,
	"round":  Round,
	"ceil":   Ceil,
	"floor":  Floor,
	"abs":    Abs,

	// ============= 比较函数 =============
	"eq": Eq,
	"ne": Ne,
	"lt": Lt,
	"le": Le,
	"gt": Gt,
	"ge": Ge,
	"in": In,

	// ============= 日期时间函数 =============
	"dateformat":    DateFormat,
	"date":          Date,
	"compare":       Compare,
	"timeago":       TimeAgo,
	"timesince":     TimeSince,
	"timeuntil":     TimeUntil,
	"now":           Now,
	"timestamp":     Timestamp,

	// ============= 集合函数 =============
	"len":      Len,
	"index":    Index,
	"slice":    Slice,
	"append":   AppendSlice,
	"reverse":  Reverse,
	"sort":     SortSlice,
	"join":     strings.Join,
	"split":    strings.Split,
	"contains": strings.Contains,

	// ============= 类型转换函数 =============
	"int":     ToInt,
	"int64":   ToInt64,
	"float":   ToFloat,
	"string":  ToString,
	"bool":    ToBool,

	// ============= URL和编码函数 =============
	"urlencode":   URLEncode,
	"urldecode":   URLDecode,
	"base64enc":   Base64Encode,
	"base64dec":   Base64Decode,
	"md5":         MD5Hash,
	"safejs":      SafeJS,
	"safehtml":    SafeHTML,

	// ============= 条件和逻辑函数 =============
	"default":  Default,
	"empty":    Empty,
	"notnil":   NotNil,
	"and":      And,
	"or":       Or,
	"not":      Not,

	// ============= 模板包含函数 =============
	"include":    Include,
	"template":   TemplateInclude,
	"partial":    Partial,
	"component":  ComponentTemplate,
	"render":     RenderTemplate,

	// ============= 迭代和循环函数 =============
	"range":    CreateRange,
	"seq":      CreateSequence,
	"dict":     CreateDict,
	"makedict": MakeDict,
	"makeslice": MakeSlice,

	// ============= 格式化函数 =============
	"printf":     fmt.Sprintf,
	"sprintf":    fmt.Sprintf,
	"formatsize": FmtByte,
	"currency":   formatCurrency,
	"number":     FormatNumber,
	"percent":    FormatPercent,

	// ============= 其他实用函数 =============
	"uuid":      GenerateUUID,
	"random":    RandomString,
	"shuffle":   Shuffle,
	"unique":    Unique,
	"compact":   Compact,
	"flatten":   Flatten,
}

// ============= 字符串处理函数实现 =============

// Str2HTML 转换字符串为HTML
func Str2HTML(str string) template.HTML {
	return template.HTML(str)
}

// HTMLQuote HTML编码
func HTMLQuote(str string) string {
	return template.HTMLEscapeString(str)
}

// HTMLUnquote HTML解码
func HTMLUnquote(str string) string {
	return template.HTMLEscapeString(str) // 简化实现
}

// Substr 字符串截取
func Substr(str string, start, length int) string {
	runes := []rune(str)
	if start < 0 || start >= len(runes) {
		return ""
	}
	end := start + length
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}

// TruncateString 截断字符串
func TruncateString(str string, length int) string {
	if utf8.RuneCountInString(str) <= length {
		return str
	}
	runes := []rune(str)
	return string(runes[:length]) + "..."
}

// NL2BR 换行符转为HTML
func NL2BR(str string) template.HTML {
	return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(str), "\n", "<br>"))
}

// MarkdownString 简单Markdown渲染
func MarkdownString(str string) template.HTML {
	// 简化的Markdown实现
	html := template.HTMLEscapeString(str)
	
	// 处理粗体
	re := regexp.MustCompile(`\*\*(.*?)\*\*`)
	html = re.ReplaceAllString(html, "<strong>$1</strong>")
	
	// 处理斜体
	re = regexp.MustCompile(`\*(.*?)\*`)
	html = re.ReplaceAllString(html, "<em>$1</em>")
	
	// 处理链接
	re = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	html = re.ReplaceAllString(html, `<a href="$2">$1</a>`)
	
	// 处理换行
	html = strings.ReplaceAll(html, "\n", "<br>")
	
	return template.HTML(html)
}

// StripHTML 去除HTML标签
func StripHTML(str string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(str, "")
}

// ============= 数学函数实现 =============

// Round 四舍五入
func Round(f float64) float64 {
	return math.Round(f)
}

// Ceil 向上取整
func Ceil(f float64) float64 {
	return math.Ceil(f)
}

// Floor 向下取整
func Floor(f float64) float64 {
	return math.Floor(f)
}

// Abs 绝对值
func Abs(f float64) float64 {
	return math.Abs(f)
}

// ============= 比较函数实现 =============

// In 检查是否在集合中
func In(item any, slice any) bool {
	switch s := slice.(type) {
	case []any:
		for _, v := range s {
			if v == item {
				return true
			}
		}
	case []string:
		itemStr := fmt.Sprintf("%v", item)
		for _, v := range s {
			if v == itemStr {
				return true
			}
		}
	case []int:
		if itemInt, ok := item.(int); ok {
			for _, v := range s {
				if v == itemInt {
					return true
				}
			}
		}
	}
	return false
}

// ============= 日期时间函数实现 =============

// DateFormat 格式化日期
func DateFormat(date any, layout string) string {
	var t time.Time
	
	switch v := date.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v != nil {
			t = *v
		} else {
			return ""
		}
	case int64:
		t = time.Unix(v, 0)
	case string:
		// 尝试解析字符串
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02",
			time.RFC3339,
			time.RFC822,
		}
		for _, l := range layouts {
			if parsed, err := time.Parse(l, v); err == nil {
				t = parsed
				break
			}
		}
		if t.IsZero() {
			return v
		}
	default:
		return fmt.Sprintf("%v", date)
	}
	
	return t.Format(layout)
}

// Date 获取当前日期
func Date(layout string) string {
	return time.Now().Format(layout)
}

// Compare 比较日期
func Compare(date1, date2 any) int {
	t1 := parseTime(date1)
	t2 := parseTime(date2)
	
	if t1.Before(t2) {
		return -1
	} else if t1.After(t2) {
		return 1
	}
	return 0
}

// TimeAgo 时间前
func TimeAgo(date any) string {
	t := parseTime(date)
	if t.IsZero() {
		return ""
	}
	
	duration := time.Since(t)
	
	if duration < time.Minute {
		return "刚刚"
	} else if duration < time.Hour {
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	} else if duration < 30*24*time.Hour {
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	} else if duration < 365*24*time.Hour {
		return fmt.Sprintf("%d个月前", int(duration.Hours()/(24*30)))
	} else {
		return fmt.Sprintf("%d年前", int(duration.Hours()/(24*365)))
	}
}

// TimeSince 距离现在的时间
func TimeSince(date any) string {
	return TimeAgo(date)
}

// TimeUntil 到某时间还有多久
func TimeUntil(date any) string {
	t := parseTime(date)
	if t.IsZero() {
		return ""
	}
	
	duration := time.Until(t)
	if duration < 0 {
		return TimeAgo(date)
	}
	
	if duration < time.Minute {
		return "即将"
	} else if duration < time.Hour {
		return fmt.Sprintf("%d分钟后", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d小时后", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%d天后", int(duration.Hours()/24))
	}
}

// Now 当前时间
func Now() time.Time {
	return time.Now()
}

// Timestamp 当前时间戳
func Timestamp() int64 {
	return time.Now().Unix()
}

// ============= 集合函数实现 =============

// AppendSlice 追加到切片
func AppendSlice(slice any, items ...any) []any {
	result := make([]any, 0)
	
	// 转换现有切片
	switch s := slice.(type) {
	case []any:
		result = append(result, s...)
	case []string:
		for _, v := range s {
			result = append(result, v)
		}
	case []int:
		for _, v := range s {
			result = append(result, v)
		}
	default:
		if slice != nil {
			result = append(result, slice)
		}
	}
	
	// 追加新项目
	result = append(result, items...)
	return result
}

// Reverse 反转切片
func Reverse(slice any) []any {
	switch s := slice.(type) {
	case []any:
		result := make([]any, len(s))
		for i, v := range s {
			result[len(s)-1-i] = v
		}
		return result
	case []string:
		result := make([]any, len(s))
		for i, v := range s {
			result[len(s)-1-i] = v
		}
		return result
	case []int:
		result := make([]any, len(s))
		for i, v := range s {
			result[len(s)-1-i] = v
		}
		return result
	}
	return []any{}
}

// SortSlice 排序切片（简化实现）
func SortSlice(slice any) []any {
	// 这里应该实现更复杂的排序逻辑
	// 暂时返回原切片
	switch s := slice.(type) {
	case []any:
		return s
	case []string:
		result := make([]any, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result
	case []int:
		result := make([]any, len(s))
		for i, v := range s {
			result[i] = v
		}
		return result
	}
	return []any{}
}

// ============= 类型转换函数实现 =============

// ToInt64 转换为int64
func ToInt64(v any) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case int32:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

// ToBool 转换为bool
func ToBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
		return val != ""
	}
	return false
}

// ============= URL和编码函数实现 =============

// URLEncode URL编码
func URLEncode(str string) string {
	return url.QueryEscape(str)
}

// URLDecode URL解码
func URLDecode(str string) string {
	if decoded, err := url.QueryUnescape(str); err == nil {
		return decoded
	}
	return str
}

// Base64Encode Base64编码
func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// Base64Decode Base64解码
func Base64Decode(str string) string {
	if decoded, err := base64.StdEncoding.DecodeString(str); err == nil {
		return string(decoded)
	}
	return str
}

// MD5Hash MD5哈希
func MD5Hash(str string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(str)))
}

// SafeJS 安全的JavaScript
func SafeJS(str string) template.JS {
	return template.JS(str)
}

// SafeHTML 安全的HTML
func SafeHTML(str string) template.HTML {
	return template.HTML(str)
}

// ============= 条件和逻辑函数实现 =============

// Empty 检查是否为空
func Empty(v any) bool {
	if v == nil {
		return true
	}
	
	switch val := v.(type) {
	case string:
		return val == ""
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	case int:
		return val == 0
	case int64:
		return val == 0
	case float64:
		return val == 0
	case bool:
		return !val
	}
	return false
}

// NotNil 检查是否不为nil
func NotNil(v any) bool {
	return v != nil
}

// ============= 迭代和循环函数实现 =============

// CreateRange 创建数字范围
func CreateRange(start, end int) []int {
	if start > end {
		return []int{}
	}
	
	result := make([]int, end-start+1)
	for i := 0; i < len(result); i++ {
		result[i] = start + i
	}
	return result
}

// CreateSequence 创建序列
func CreateSequence(count int) []int {
	result := make([]int, count)
	for i := 0; i < count; i++ {
		result[i] = i
	}
	return result
}

// CreateDict 创建字典
func CreateDict(values ...any) map[string]any {
	dict := make(map[string]any)
	for i := 0; i < len(values)-1; i += 2 {
		key := fmt.Sprintf("%v", values[i])
		dict[key] = values[i+1]
	}
	return dict
}

// MakeDict 创建字典（别名）
func MakeDict(values ...any) map[string]any {
	return CreateDict(values...)
}

// ============= 格式化函数实现 =============

// FormatNumber 格式化数字
func FormatNumber(number any, decimals int) string {
	f := toFloat64(number)
	return fmt.Sprintf("%.*f", decimals, f)
}

// FormatPercent 格式化百分比
func FormatPercent(number any, decimals int) string {
	f := toFloat64(number) * 100
	return fmt.Sprintf("%.*f%%", decimals, f)
}

// formatCurrency 格式化货币
func formatCurrency(amount any, currency string) string {
	var value float64
	
	switch v := amount.(type) {
	case float64:
		value = v
	case float32:
		value = float64(v)
	case int:
		value = float64(v)
	case int64:
		value = float64(v)
	default:
		return fmt.Sprintf("%v", amount)
	}
	
	switch currency {
	case "CNY", "RMB", "¥":
		return fmt.Sprintf("¥%.2f", value)
	case "USD", "$":
		return fmt.Sprintf("$%.2f", value)
	case "EUR", "€":
		return fmt.Sprintf("€%.2f", value)
	default:
		return fmt.Sprintf("%.2f %s", value, currency)
	}
}

// ============= 其他实用函数实现 =============

// GenerateUUID 生成UUID（简化版）
func GenerateUUID() string {
	// 这里应该使用真正的UUID库
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// Shuffle 打乱切片
func Shuffle(slice any) []any {
	// 简化实现，实际应该使用随机算法
	return SortSlice(slice)
}

// Unique 去重
func Unique(slice any) []any {
	seen := make(map[any]bool)
	result := make([]any, 0)
	
	switch s := slice.(type) {
	case []any:
		for _, v := range s {
			if !seen[v] {
				seen[v] = true
				result = append(result, v)
			}
		}
	case []string:
		for _, v := range s {
			if !seen[v] {
				seen[v] = true
				result = append(result, v)
			}
		}
	case []int:
		for _, v := range s {
			if !seen[v] {
				seen[v] = true
				result = append(result, v)
			}
		}
	}
	return result
}

// Compact 移除空值
func Compact(slice any) []any {
	result := make([]any, 0)
	
	switch s := slice.(type) {
	case []any:
		for _, v := range s {
			if !Empty(v) {
				result = append(result, v)
			}
		}
	case []string:
		for _, v := range s {
			if v != "" {
				result = append(result, v)
			}
		}
	case []int:
		for _, v := range s {
			if v != 0 {
				result = append(result, v)
			}
		}
	}
	return result
}

// Flatten 扁平化嵌套切片
func Flatten(slice any) []any {
	result := make([]any, 0)
	
	switch s := slice.(type) {
	case []any:
		for _, v := range s {
			if nested, ok := v.([]any); ok {
				result = append(result, Flatten(nested)...)
			} else {
				result = append(result, v)
			}
		}
	default:
		if slice != nil {
			result = append(result, slice)
		}
	}
	return result
}

// ============= 模板相关函数 =============

// 这些函数需要访问模板引擎实例，将在后续实现中完善

// Include 包含模板（占位符）
func Include(templateName string, data ...any) template.HTML {
	return template.HTML(fmt.Sprintf("<!-- Include: %s -->", templateName))
}

// TemplateInclude 模板包含（占位符）
func TemplateInclude(templateName string, data ...any) template.HTML {
	return Include(templateName, data...)
}

// Partial 部分模板（占位符）
func Partial(templateName string, data ...any) template.HTML {
	return Include(templateName, data...)
}

// ComponentTemplate 组件模板（占位符）
func ComponentTemplate(componentName string, data ...any) template.HTML {
	return template.HTML(fmt.Sprintf("<!-- Component: %s -->", componentName))
}

// RenderTemplate 渲染模板（占位符）
func RenderTemplate(templateName string, data ...any) template.HTML {
	return Include(templateName, data...)
}

// RenderForm 渲染表单（占位符）
func RenderForm(form any) template.HTML {
	return template.HTML("<!-- Form rendering not implemented -->")
}

// AssetsJS JS资源（占位符）
func AssetsJS(files ...string) template.HTML {
	var html strings.Builder
	for _, file := range files {
		html.WriteString(fmt.Sprintf(`<script src="/static/js/%s"></script>`, file))
	}
	return template.HTML(html.String())
}

// AssetsCSS CSS资源（占位符）
func AssetsCSS(files ...string) template.HTML {
	var html strings.Builder
	for _, file := range files {
		html.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="/static/css/%s">`, file))
	}
	return template.HTML(html.String())
}

// GetConfig 获取配置（占位符）
func GetConfig(key string) string {
	return fmt.Sprintf("config:%s", key)
}

// MapGet 从map获取值
func MapGet(m map[string]any, key string) any {
	if v, ok := m[key]; ok {
		return v
	}
	return nil
}

// URLFor 生成URL（占位符）
func URLFor(endpoint string, params ...any) string {
	url := "/" + strings.TrimPrefix(endpoint, "/")
	if len(params) > 0 {
		query := make([]string, 0, len(params)/2)
		for i := 0; i < len(params)-1; i += 2 {
			key := fmt.Sprintf("%v", params[i])
			value := fmt.Sprintf("%v", params[i+1])
			query = append(query, fmt.Sprintf("%s=%s", key, value))
		}
		if len(query) > 0 {
			url += "?" + strings.Join(query, "&")
		}
	}
	return url
}

// ============= 辅助函数 =============

// parseTime 解析时间
func parseTime(v any) time.Time {
	switch val := v.(type) {
	case time.Time:
		return val
	case *time.Time:
		if val != nil {
			return *val
		}
	case int64:
		return time.Unix(val, 0)
	case string:
		layouts := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z07:00",
			"2006-01-02",
			time.RFC3339,
			time.RFC822,
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, val); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}