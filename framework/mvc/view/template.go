package view

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/pkg/xmath"
	"github.com/zsy619/yyhertz/framework/pkg/xstring"
)

var (
	errBadComparisonType = errors.New("invalid type for comparison")
	errBadComparison     = errors.New("incompatible types for comparison")
	errNoComparison      = errors.New("missing argument for comparison")
)

type kind int

const (
	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	stringKind
	uintKind
)

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}
	return invalidKind, errBadComparisonType
}

// builtinTemplateFuncs 内置模板函数映射（私有）
var builtinTemplateFuncs = template.FuncMap{
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
	"add":           xmath.Add,
	"sub":           xmath.Sub,
	"mul":           xmath.Mul,
	"div":           xmath.Div,
	"mod":           xmath.Mod,
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
	"toInt":         xmath.ToInt,
	"toFloat":       xmath.ToFloat64,
	"upper":         strings.ToUpper,
	"lower":         strings.ToLower,
	"title":         xstring.ToTitleCase,
	"trim":          strings.TrimSpace,
	"replace":       strings.ReplaceAll,
	"split":         strings.Split,
	"join":          strings.Join,
	"hasPrefix":     strings.HasPrefix,
	"hasSuffix":     strings.HasSuffix,
	"len":           Len,
	"index":         Index,
	"slice":         Slice,

	// CSRF Token 相关函数
	"csrf":       GetCSRFTokenFromContext,
	"csrf_token": GetCSRFTokenFromContext,
}

// ============= 访问接口 =============

// GetBuiltinTemplateFunctions 获取内置模板函数映射（只读副本）
func GetBuiltinTemplateFunctions() template.FuncMap {
	funcMapCopy := make(template.FuncMap, len(builtinTemplateFuncs))
	for name, fn := range builtinTemplateFuncs {
		funcMapCopy[name] = fn
	}
	return funcMapCopy
}

// TemplateFuncs 向后兼容的全局变量
var TemplateFuncs = builtinTemplateFuncs

// GetBuiltinFunctionNames 获取所有内置函数名称
func GetBuiltinFunctionNames() []string {
	names := make([]string, 0, len(builtinTemplateFuncs))
	for name := range builtinTemplateFuncs {
		names = append(names, name)
	}
	return names
}

// HasBuiltinFunction 检查是否有指定的内置函数
func HasBuiltinFunction(name string) bool {
	_, exists := builtinTemplateFuncs[name]
	return exists
}

// GetBuiltinFunctionCount 获取内置函数数量
func GetBuiltinFunctionCount() int {
	return len(builtinTemplateFuncs)
}

// ============= 增强的模板加载函数 =============

// LoadTemplate 加载模板文件（增强版，支持缓存和统一函数管理）
func LoadTemplate(templatePath string, data map[string]any) (string, error) {
	return LoadTemplateWithOptions(templatePath, data, &TemplateLoadOptions{})
}

// LoadTemplateWithLayout 加载带布局的模板（增强版）
func LoadTemplateWithLayout(layoutPath, templatePath string, data map[string]any) (string, error) {
	return LoadTemplateWithOptions(templatePath, data, &TemplateLoadOptions{
		LayoutPath: layoutPath,
	})
}

// TemplateLoadOptions 模板加载选项
type TemplateLoadOptions struct {
	LayoutPath  string           // 布局文件路径
	EnableCache bool             // 启用缓存（默认true）
	CustomFuncs template.FuncMap // 自定义函数
	DelimLeft   string           // 左分隔符（默认{{）
	DelimRight  string           // 右分隔符（默认}}）
	StrictMode  bool             // 严格模式（错误时停止）
	ContextData map[string]any   // 上下文数据
	Theme       string           // 主题名称
	Debug       bool             // 调试模式
}

// LoadTemplateWithOptions 带选项的模板加载（核心实现）
func LoadTemplateWithOptions(templatePath string, data map[string]any, options *TemplateLoadOptions) (string, error) {
	if options == nil {
		options = &TemplateLoadOptions{}
	}

	// 设置默认值
	if options.DelimLeft == "" {
		options.DelimLeft = "{{"
	}
	if options.DelimRight == "" {
		options.DelimRight = "}}"
	}
	if !options.StrictMode {
		options.EnableCache = true // 默认启用缓存
	}

	// 获取统一的函数管理器
	manager := GetGlobalFunctionManager()

	// 合并函数映射：内置函数 + 全局函数 + 自定义函数
	mergedFuncs := manager.GetMergedFunctions(options.CustomFuncs)

	// 准备渲染数据
	renderData := prepareTemplateData(data, options)

	var tmpl *template.Template
	var err error

	if options.LayoutPath != "" {
		// 使用布局模板
		tmpl, err = createTemplateWithLayout(templatePath, options.LayoutPath, mergedFuncs, options)
	} else {
		// 单独模板
		tmpl, err = createSingleTemplate(templatePath, mergedFuncs, options)
	}

	if err != nil {
		if options.Debug {
			return "", fmt.Errorf("template creation failed for %s: %w\nOptions: %+v", templatePath, err, options)
		}
		return "", fmt.Errorf("template creation failed: %w", err)
	}

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderData); err != nil {
		if options.StrictMode {
			return "", fmt.Errorf("template execution failed for %s: %w", templatePath, err)
		}

		// 非严格模式下，返回错误信息但不中断
		if options.Debug {
			return fmt.Sprintf("<!-- Template execution error: %s -->", err.Error()), nil
		}
		return "", fmt.Errorf("template execution error: %w", err)
	}

	result := buf.String()

	// 调试信息
	if options.Debug {
		debugInfo := fmt.Sprintf("<!-- Template: %s, Layout: %s, Theme: %s -->",
			templatePath, options.LayoutPath, options.Theme)
		result = debugInfo + "\n" + result
	}

	return result, nil
}

// createSingleTemplate 创建单个模板
func createSingleTemplate(templatePath string, funcs template.FuncMap, options *TemplateLoadOptions) (*template.Template, error) {
	templateName := filepath.Base(templatePath)

	tmpl := template.New(templateName).
		Delims(options.DelimLeft, options.DelimRight).
		Funcs(funcs)

	_, err := tmpl.ParseFiles(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template file %s: %w", templatePath, err)
	}

	return tmpl, nil
}

// createTemplateWithLayout 创建带布局的模板
func createTemplateWithLayout(templatePath, layoutPath string, funcs template.FuncMap, options *TemplateLoadOptions) (*template.Template, error) {
	// 使用layout作为主模板名
	tmpl := template.New("layout").
		Delims(options.DelimLeft, options.DelimRight).
		Funcs(funcs)

	// 先解析布局文件，再解析内容文件
	_, err := tmpl.ParseFiles(layoutPath, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template files (layout: %s, template: %s): %w", layoutPath, templatePath, err)
	}

	return tmpl, nil
}

// prepareTemplateData 准备模板数据
func prepareTemplateData(data map[string]any, options *TemplateLoadOptions) map[string]any {
	if data == nil {
		data = make(map[string]any)
	}

	// 添加上下文数据
	if options.ContextData != nil {
		for k, v := range options.ContextData {
			if _, exists := data[k]; !exists { // 不覆盖用户数据
				data[k] = v
			}
		}
	}

	// 添加系统数据
	data["__template_theme__"] = options.Theme
	data["__template_debug__"] = options.Debug
	data["__template_timestamp__"] = time.Now().Unix()

	// 添加便捷的CSRF token访问
	if provider := GetCSRFTokenProvider(); provider != nil {
		if _, exists := data["csrf_token"]; !exists {
			data["csrf_token"] = provider.GenerateSimpleToken()
		}
		if _, exists := data["csrf"]; !exists {
			data["csrf"] = data["csrf_token"]
		}
	}

	return data
}

// ============= Beego风格的便捷函数 =============

// LoadTemplateFromString 从字符串加载模板（Beego风格）
func LoadTemplateFromString(templateContent string, data map[string]any) (string, error) {
	return LoadTemplateFromStringWithOptions(templateContent, data, &TemplateLoadOptions{})
}

// LoadTemplateFromStringWithOptions 从字符串加载模板（带选项）
func LoadTemplateFromStringWithOptions(templateContent string, data map[string]any, options *TemplateLoadOptions) (string, error) {
	if options == nil {
		options = &TemplateLoadOptions{}
	}

	// 设置默认值
	if options.DelimLeft == "" {
		options.DelimLeft = "{{"
	}
	if options.DelimRight == "" {
		options.DelimRight = "}}"
	}

	// 获取统一的函数管理器
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(options.CustomFuncs)

	// 创建模板
	tmpl := template.New("string_template").
		Delims(options.DelimLeft, options.DelimRight).
		Funcs(mergedFuncs)

	_, err := tmpl.Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template string: %w", err)
	}

	// 准备渲染数据
	renderData := prepareTemplateData(data, options)

	// 渲染模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, renderData); err != nil {
		return "", fmt.Errorf("template execution error: %w", err)
	}

	return buf.String(), nil
}

// QuickRender 快速渲染（Beego风格的简化接口）
func QuickRender(templatePath string, data any) (string, error) {
	// 转换data为map[string]any
	var dataMap map[string]any
	if data != nil {
		if m, ok := data.(map[string]any); ok {
			dataMap = m
		} else {
			dataMap = map[string]any{"data": data}
		}
	}

	return LoadTemplate(templatePath, dataMap)
}

// QuickRenderWithLayout 快速布局渲染
func QuickRenderWithLayout(templatePath, layoutPath string, data any) (string, error) {
	// 转换data为map[string]any
	var dataMap map[string]any
	if data != nil {
		if m, ok := data.(map[string]any); ok {
			dataMap = m
		} else {
			dataMap = map[string]any{"data": data}
		}
	}

	return LoadTemplateWithLayout(layoutPath, templatePath, dataMap)
}

// RenderWithEngine 使用指定引擎渲染（集成现有引擎）
func RenderWithEngine(engine *TemplateEngine, templateName string, data any) (string, error) {
	if engine == nil {
		engine = GetDefaultEngine()
	}

	return engine.Render(templateName, data)
}

// RenderWithLayoutAndEngine 使用指定引擎和布局渲染
func RenderWithLayoutAndEngine(engine *TemplateEngine, templateName, layoutName string, data any) (string, error) {
	if engine == nil {
		engine = GetDefaultEngine()
	}

	return engine.RenderWithLayout(templateName, layoutName, data)
}

// BatchRender 批量渲染模板
func BatchRender(templates []string, data map[string]any) (map[string]string, error) {
	results := make(map[string]string)
	errors := make([]string, 0)

	for _, templatePath := range templates {
		result, err := LoadTemplate(templatePath, data)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", templatePath, err))
			continue
		}
		results[templatePath] = result
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("batch render errors: %s", strings.Join(errors, "; "))
	}

	return results, nil
}

// CompileTemplate 编译模板但不执行（用于预编译缓存）
func CompileTemplate(templatePath string, options *TemplateLoadOptions) (*template.Template, error) {
	if options == nil {
		options = &TemplateLoadOptions{}
	}

	// 获取统一的函数管理器
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(options.CustomFuncs)

	if options.LayoutPath != "" {
		return createTemplateWithLayout(templatePath, options.LayoutPath, mergedFuncs, options)
	} else {
		return createSingleTemplate(templatePath, mergedFuncs, options)
	}
}

// ValidateTemplate 验证模板语法
func ValidateTemplate(templatePath string) error {
	_, err := CompileTemplate(templatePath, &TemplateLoadOptions{
		StrictMode: true,
		Debug:      true,
	})
	return err
}

// ValidateTemplateString 验证模板字符串语法
func ValidateTemplateString(templateContent string) error {
	manager := GetGlobalFunctionManager()
	mergedFuncs := manager.GetMergedFunctions(nil)

	tmpl := template.New("validation").Funcs(mergedFuncs)
	_, err := tmpl.Parse(templateContent)
	return err
}

// GetTemplateInfo 获取模板信息（用于调试）
func GetTemplateInfo(templatePath string) (map[string]any, error) {
	tmpl, err := CompileTemplate(templatePath, &TemplateLoadOptions{})
	if err != nil {
		return nil, err
	}

	info := map[string]any{
		"name":      tmpl.Name(),
		"templates": make([]string, 0),
	}

	// 获取模板中定义的所有子模板
	for _, t := range tmpl.Templates() {
		if t.Name() != "" {
			info["templates"] = append(info["templates"].([]string), t.Name())
		}
	}

	return info, nil
}

// 工具函数

// MakeSlice 创建切片
func MakeSlice(args ...any) []any {
	return args
}

// ConcatString 连接字符串
func ConcatString(strs ...string) string {
	var totalLen int
	for _, s := range strs {
		totalLen += len(s)
	}
	var builder strings.Builder
	builder.Grow(totalLen) // 预分配内存
	for _, s := range strs {
		builder.WriteString(s)
	}
	return builder.String()
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
	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	if size < 1024 {
		return strconv.FormatInt(size, 10) + units[0]
	}

	div, exp := int64(1024), 0
	for n := size / 1024; n >= 1024 && exp < len(units)-2; n /= 1024 {
		div *= 1024
		exp++
	}

	// 使用整数运算避免浮点精度问题
	value := float64(size) / float64(div)

	// 使用 strconv 而不是 fmt.Sprintf 以提高性能
	return strconv.FormatFloat(value, 'f', 2, 64) + units[exp+1]
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

// 比较函数
func Eq(arg1 any, arg2 ...any) (bool, error) {
	v1 := reflect.ValueOf(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	if len(arg2) == 0 {
		return false, errNoComparison
	}
	for _, arg := range arg2 {
		v2 := reflect.ValueOf(arg)
		k2, err := basicKind(v2)
		if err != nil {
			return false, err
		}
		truth := false
		if k1 != k2 {
			// Special case: Can compare integer values regardless of type's sign.
			switch {
			case k1 == intKind && k2 == uintKind:
				truth = v1.Int() >= 0 && uint64(v1.Int()) == v2.Uint()
			case k1 == uintKind && k2 == intKind:
				truth = v2.Int() >= 0 && v1.Uint() == uint64(v2.Int())
			default:
				return false, errBadComparison
			}
			if truth {
				return true, nil
			} else {
				return false, nil
			}
		}
		switch k1 {
		case boolKind:
			truth = v1.Bool() == v2.Bool()
		case complexKind:
			truth = v1.Complex() == v2.Complex()
		case floatKind:
			truth = v1.Float() == v2.Float()
		case intKind:
			truth = v1.Int() == v2.Int()
		case stringKind:
			truth = v1.String() == v2.String()
		case uintKind:
			truth = v1.Uint() == v2.Uint()
		default:
			panic("invalid kind")
		}
		if truth {
			return true, nil
		}
	}
	return false, nil
}

func Ne(arg1, arg2 any) (bool, error) {
	// != is the inverse of ==.
	equal, err := Eq(arg1, arg2)
	return !equal, err
}

func Lt(arg1, arg2 any) (bool, error) {
	v1 := reflect.ValueOf(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	v2 := reflect.ValueOf(arg2)
	k2, err := basicKind(v2)
	if err != nil {
		return false, err
	}
	truth := false
	if k1 != k2 {
		// Special case: Can compare integer values regardless of type's sign.
		switch {
		case k1 == intKind && k2 == uintKind:
			truth = v1.Int() < 0 || uint64(v1.Int()) < v2.Uint()
		case k1 == uintKind && k2 == intKind:
			truth = v2.Int() >= 0 && v1.Uint() < uint64(v2.Int())
		default:
			return false, errBadComparison
		}
	} else {
		switch k1 {
		case boolKind, complexKind:
			return false, errBadComparisonType
		case floatKind:
			truth = v1.Float() < v2.Float()
		case intKind:
			truth = v1.Int() < v2.Int()
		case stringKind:
			truth = v1.String() < v2.String()
		case uintKind:
			truth = v1.Uint() < v2.Uint()
		default:
			return false, errBadComparisonType
		}
	}
	return truth, nil
}

func Le(arg1, arg2 any) (bool, error) {
	// <= is < or ==.
	lessThan, err := Lt(arg1, arg2)
	if lessThan || err != nil {
		return lessThan, err
	}
	return Eq(arg1, arg2)
}

func Gt(arg1, arg2 any) (bool, error) {
	// > is the inverse of <=.
	lessOrEqual, err := Le(arg1, arg2)
	if err != nil {
		return false, err
	}
	return !lessOrEqual, nil
}

func Ge(arg1, arg2 any) (bool, error) {
	// >= is the inverse of <.
	lessThan, err := Lt(arg1, arg2)
	if err != nil {
		return false, err
	}
	return !lessThan, nil
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
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
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

// GetCSRFTokenFromContext 从模板上下文获取CSRF token
// 这个函数用于模板函数 {{csrf}} 和 {{csrf_token}}
func GetCSRFTokenFromContext() string {
	// 获取全局CSRF提供者
	provider := GetCSRFTokenProvider()
	if provider != nil {
		return provider.GenerateSimpleToken()
	}
	return ""
}
