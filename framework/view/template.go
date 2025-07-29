package view

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TemplateFuncs 模板函数映射
var TemplateFuncs = template.FuncMap{
	"makeSlice":     MakeSlice,
	"concatString":  ConcatString,
	"containString": ContainString,
	"authContain":   AuthContain,
	"fmtByte":       FmtByte,
	"fmtFloat":      FmtFloat,
	"fmtFloat2":     FmtFloat2,
	"fmtFloat3":     FmtFloat3,
	"fmtFloat4":     FmtFloat4,
	"fmtFloat5":     FmtFloat5,
	"fmtString":     FmtString,
	"getTime":       GetTime,
	"getTimestamp":  GetTimestamp,
	"formatTime":    FormatTime,
	"add":           Add,
	"sub":           Sub,
	"mul":           Mul,
	"div":           Div,
	"mod":           Mod,
	"eq":            Eq,
	"ne":            Ne,
	"lt":            Lt,
	"le":            Le,
	"gt":            Gt,
	"ge":            Ge,
	"and":           And,
	"or":            Or,
	"not":           Not,
	"default":       Default,
	"toString":      ToString,
	"toInt":         ToInt,
	"toFloat":       ToFloat,
	"upper":         strings.ToUpper,
	"lower":         strings.ToLower,
	"title":         strings.Title,
	"trim":          strings.TrimSpace,
	"replace":       strings.ReplaceAll,
	"split":         strings.Split,
	"join":          strings.Join,
	"hasPrefix":     strings.HasPrefix,
	"hasSuffix":     strings.HasSuffix,
	"len":           Len,
	"index":         Index,
	"slice":         Slice,
}

// LoadTemplate 加载模板文件
func LoadTemplate(templatePath string, data map[string]any) (string, error) {
	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(TemplateFuncs).ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}
	
	return buf.String(), nil
}

// LoadTemplateWithLayout 加载带布局的模板
func LoadTemplateWithLayout(layoutPath, templatePath string, data map[string]any) (string, error) {
	tmpl, err := template.New("layout").Funcs(TemplateFuncs).ParseFiles(layoutPath, templatePath)
	if err != nil {
		return "", fmt.Errorf("template parse error: %w", err)
	}
	
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		return "", fmt.Errorf("template execute error: %w", err)
	}
	
	return buf.String(), nil
}

// 工具函数

// MakeSlice 创建切片
func MakeSlice(args ...any) []any {
	return args
}

// ConcatString 连接字符串
func ConcatString(strs ...string) string {
	var result strings.Builder
	for _, str := range strs {
		result.WriteString(str)
	}
	return result.String()
}

// ContainString 检查字符串是否包含子字符串
func ContainString(s, substr string) bool {
	return strings.Contains(","+s+",", ","+substr+",")
}

// AuthContain 检查权限包含
func AuthContain(s string, in int) bool {
	s = "," + s + ","
	substr := "," + strconv.Itoa(in) + ","
	return strings.Contains(s, substr)
}

// FmtByte 格式化字节
func FmtByte(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%.2fB", float64(size))
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(size)/(1024*1024))
	} else if size < 1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fGB", float64(size)/(1024*1024*1024))
	} else if size < 1024*1024*1024*1024*1024 {
		return fmt.Sprintf("%.2fTB", float64(size)/(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fPB", float64(size)/(1024*1024*1024*1024*1024))
	}
}

// FmtFloat 格式化浮点数
func FmtFloat(value float64, decimals int) string {
	return fmt.Sprintf("%.*f", decimals, value)
}

// FmtFloat2 格式化浮点数(2位小数)
func FmtFloat2(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

// FmtFloat3 格式化浮点数(3位小数)
func FmtFloat3(value float64) string {
	return fmt.Sprintf("%.3f", value)
}

// FmtFloat4 格式化浮点数(4位小数)
func FmtFloat4(value float64) string {
	return fmt.Sprintf("%.4f", value)
}

// FmtFloat5 格式化浮点数(5位小数)
func FmtFloat5(value float64) string {
	return fmt.Sprintf("%.5f", value)
}

// FmtString 格式化字符串
func FmtString(value string, width int) string {
	return fmt.Sprintf("%*s", width, value)
}

// GetTime 获取当前时间
func GetTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// GetTimestamp 获取当前时间戳
func GetTimestamp() int64 {
	return time.Now().Unix()
}

// FormatTime 格式化时间
func FormatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

// 数学运算函数
func Add(a, b any) any {
	return toFloat64(a) + toFloat64(b)
}

func Sub(a, b any) any {
	return toFloat64(a) - toFloat64(b)
}

func Mul(a, b any) any {
	return toFloat64(a) * toFloat64(b)
}

func Div(a, b any) any {
	bVal := toFloat64(b)
	if bVal == 0 {
		return 0
	}
	return toFloat64(a) / bVal
}

func Mod(a, b any) any {
	return int(toFloat64(a)) % int(toFloat64(b))
}

// 比较函数
func Eq(a, b any) bool {
	return a == b
}

func Ne(a, b any) bool {
	return a != b
}

func Lt(a, b any) bool {
	return toFloat64(a) < toFloat64(b)
}

func Le(a, b any) bool {
	return toFloat64(a) <= toFloat64(b)
}

func Gt(a, b any) bool {
	return toFloat64(a) > toFloat64(b)
}

func Ge(a, b any) bool {
	return toFloat64(a) >= toFloat64(b)
}

// 逻辑函数
func And(a, b bool) bool {
	return a && b
}

func Or(a, b bool) bool {
	return a || b
}

func Not(a bool) bool {
	return !a
}

// Default 返回默认值
func Default(defaultValue, value any) any {
	if value == nil || value == "" {
		return defaultValue
	}
	return value
}

// 类型转换函数
func ToString(v any) string {
	return fmt.Sprintf("%v", v)
}

func ToInt(v any) int {
	return int(toFloat64(v))
}

func ToFloat(v any) float64 {
	return toFloat64(v)
}

// 集合函数
func Len(v any) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case []any:
		return len(val)
	case map[string]any:
		return len(val)
	default:
		return 0
	}
}

func Index(slice any, index int) any {
	if s, ok := slice.([]any); ok && index >= 0 && index < len(s) {
		return s[index]
	}
	return nil
}

func Slice(slice any, start, end int) any {
	if s, ok := slice.([]any); ok {
		if start < 0 {
			start = 0
		}
		if end > len(s) {
			end = len(s)
		}
		if start >= end {
			return []any{}
		}
		return s[start:end]
	}
	return nil
}

// 辅助函数
func toFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0
}