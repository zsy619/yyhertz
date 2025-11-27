package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// 模板函数分析工具
type TemplateFunctionAnalyzer struct {
	functions      map[string]string // 函数名 -> 实际函数
	goBuiltins     []string          // Go 内置动作
	namingPatterns map[string][]string // 命名模式 -> 函数列表
}

// Go 模板内置动作列表
var goBuiltinActions = []string{
	"template", "define", "block", "range", "if", "else", "end", "with", "call",
	"and", "or", "not", "eq", "ne", "lt", "le", "gt", "ge", "index", "len", "print", "printf", "println",
	"html", "js", "urlquery", "call", "slice",
}

// 分析当前注册的所有函数
func (analyzer *TemplateFunctionAnalyzer) analyzeFunctions() {
	// 从beego_functions.go中提取的实际函数列表
	functions := map[string]string{
		"abs":              "xmath.Abs",
		"add":              "xmath.Add", 
		"and":              "And",
		"append":           "AppendSlice",
		"assets_css":       "AssetsCSS",
		"assets_js":        "AssetsJS",
		"base64dec":        "Base64Decode",
		"base64Dec":        "Base64Decode",
		"base64enc":        "Base64Encode", 
		"base64Enc":        "Base64Encode",
		"bool":             "xmath.ToBool",
		"ceil":             "xmath.Ceil",
		"compact":          "Compact",
		"compare":          "Compare",
		"comparenot":       "CompareNot",
		"compareNot":       "CompareNot",
		"component":        "ComponentTemplate",
		"config":           "GetConfig",
		"contains":         "xstring.Contains",
		"csrf":             "GetCSRFTokenFromContext",
		"csrf_token":       "GetCSRFTokenFromContext",
		"currency":         "formatCurrency",
		"date":             "Date",
		"dateformat":       "DateFormat",
		"default":          "Default",
		"div":              "xmath.Div",
		"empty":            "Empty",
		"eq":               "Eq",
		"first":            "First",
		"flatten":          "Flatten",
		"float":            "xmath.ToFloat64",
		"floor":            "xmath.Floor",
		"formatfilesize":   "FmtByte",
		"formatFileSize":   "FmtByte",
		"formatsize":       "FmtByte",
		"formatSize":       "FmtByte",
		"ge":               "Ge",
		"gt":               "Gt",
		"hasPrefix":        "xstring.HasPrefix",
		"hasSuffix":        "xstring.HasSuffix",
		"html2str":         "HTML2str",
		"htmlquote":        "HTMLQuote",
		"htmlunquote":      "HTMLUnquote",
		"i18n":             "I18n",
		"in":               "In",
		"include":          "Include",
		"includeTmpl":      "Include",
		"index":            "Index",
		"int":              "xmath.ToInt",
		"int64":            "xmath.ToInt64",
		"join":             "strings.Join",
		"last":             "Last",
		"le":               "Le",
		"len":              "Len",
		"lt":               "Lt",
		"makedict":         "MakeDict",
		"makeDict":         "MakeDict",
		"makerange":        "CreateRange",
		"makeRange":        "CreateRange",
		"makeseq":          "CreateSequence",
		"makeSeq":          "CreateSequence",
		"makeslice":        "MakeSlice",
		"makeSlice":        "MakeSlice",
		"map_get":          "MapGet",
		"markdown":         "MarkdownString",
		"md5":              "MD5Hash",
		"mod":              "xmath.Mod",
		"mul":              "xmath.Mul",
		"ne":               "Ne",
		"nl2br":            "NL2BR",
		"not":              "Not",
		"notnil":           "NotNil",
		"notNil":           "NotNil",
		"notnull":          "NotNil",
		"notNull":          "NotNil",
		"now":              "Now",
		"number":           "FormatNumber",
		"or":               "Or",
		"partial":          "Partial",
		"percent":          "FormatPercent",
		"printf":           "fmt.Sprintf",
		"prop":             "GetProp",
		"random":           "RandomString",
		"raw":              "RawHTML",
		"render":           "RenderTemplate",
		"renderform":       "RenderForm",
		"renderPartial":    "RenderTemplate",
		"replace":          "strings.ReplaceAll",
		"reverse":          "Reverse",
		"round":            "xmath.Round",
		"safehtml":         "SafeHTML",
		"safeHtml":         "SafeHTML",
		"safejs":           "SafeJS",
		"safeJs":           "SafeJS",
		"shuffle":          "Shuffle",
		"slice":            "Slice",
		"slot":             "GetSlot",
		"sort":             "SortSlice",
		"split":            "strings.Split",
		"sprintf":          "fmt.Sprintf",
		"str2html":         "Str2HTML",
		"string":           "ToString",
		"striphtml":        "StripHTML",
		"sub":              "xmath.Sub",
		"substr":           "xstring.Substr",
		"Substring":        "xstring.Substr",
		"templatefunc":     "TemplateInclude",
		"templateinclude":  "TemplateInclude",
		"timeago":          "TimeAgo",
		"timeAgo":          "TimeAgo",
		"timesince":        "TimeSince",
		"timeSince":        "TimeSince",
		"timestamp":        "Timestamp",
		"timeuntil":        "TimeUntil",
		"timeUntil":        "TimeUntil",
		"tmpl_include":     "Include",
		"tocapital":        "xstring.CapitalizeFirst",
		"tolower":          "strings.ToLower",
		"totitle":          "strings.ToTitle",
		"toupper":          "strings.ToUpper",
		"trans":            "I18n",
		"trim":             "strings.TrimSpace",
		"trimprefix":       "strings.TrimPrefix",
		"trimPrefixSlash":  "xstring.TrimPrefixSlash",
		"trimSlash":        "xstring.TrimSlash",
		"trimsuffix":       "strings.TrimSuffix",
		"trimSuffixSlash":  "xstring.TrimSuffixSlash",
		"truncate":         "TruncateString",
		"unescaped":        "RawHTML",
		"unique":           "Unique",
		"urldecode":        "URLDecode",
		"urlDecode":        "URLDecode",
		"urlencode":        "URLEncode",
		"urlEncode":        "URLEncode",
		"urlfor":           "URLFor",
		"uuid":             "GenerateUUID",
	}

	analyzer.functions = functions
	analyzer.goBuiltins = goBuiltinActions
	analyzer.namingPatterns = make(map[string][]string)
}

// 检测Go内置动作冲突
func (analyzer *TemplateFunctionAnalyzer) findGoBuiltinConflicts() []string {
	var conflicts []string
	for funcName := range analyzer.functions {
		for _, builtin := range analyzer.goBuiltins {
			if funcName == builtin {
				conflicts = append(conflicts, funcName)
				break
			}
		}
	}
	sort.Strings(conflicts)
	return conflicts
}

// 分析命名模式
func (analyzer *TemplateFunctionAnalyzer) analyzeNamingPatterns() map[string][]string {
	patterns := make(map[string][]string)
	
	// 按实际函数分组，找出别名
	funcGroups := make(map[string][]string)
	for funcName, actualFunc := range analyzer.functions {
		funcGroups[actualFunc] = append(funcGroups[actualFunc], funcName)
	}
	
	// 分析命名模式
	for actualFunc, aliases := range funcGroups {
		if len(aliases) > 1 {
			sort.Strings(aliases)
			patterns[actualFunc] = aliases
		}
	}
	
	return patterns
}

// 检测缺失的标准别名
func (analyzer *TemplateFunctionAnalyzer) findMissingAliases() map[string][]string {
	missing := make(map[string][]string)
	
	// 按实际函数分组
	funcGroups := make(map[string][]string)
	for funcName, actualFunc := range analyzer.functions {
		funcGroups[actualFunc] = append(funcGroups[actualFunc], funcName)
	}
	
	// 检查每个函数组的标准化别名
	for actualFunc, aliases := range funcGroups {
		if len(aliases) == 1 {
			continue // 只有一个名称的函数，跳过
		}
		
		// 检查是否有驼峰和下划线版本
		hasLowerCase := false
		hasCamelCase := false
		
		for _, alias := range aliases {
			if isAllLowerCase(alias) {
				hasLowerCase = true
			}
			if isCamelCase(alias) {
				hasCamelCase = true
			}
		}
		
		var missingAliases []string
		
		// 基于现有别名推断可能缺失的别名
		if hasLowerCase && !hasCamelCase {
			// 尝试生成驼峰版本
			for _, alias := range aliases {
				if isAllLowerCase(alias) {
					camelCase := toCamelCase(alias)
					if camelCase != alias && !contains(aliases, camelCase) {
						missingAliases = append(missingAliases, camelCase)
					}
				}
			}
		}
		
		if hasCamelCase && !hasLowerCase {
			// 尝试生成小写版本
			for _, alias := range aliases {
				if isCamelCase(alias) {
					lowerCase := strings.ToLower(alias)
					if lowerCase != alias && !contains(aliases, lowerCase) {
						missingAliases = append(missingAliases, lowerCase)
					}
				}
			}
		}
		
		if len(missingAliases) > 0 {
			missing[actualFunc] = missingAliases
		}
	}
	
	return missing
}

// 辅助函数
func isAllLowerCase(s string) bool {
	return s == strings.ToLower(s) && !strings.ContainsAny(s, "_")
}

func isCamelCase(s string) bool {
	matched, _ := regexp.MatchString(`^[a-z][a-zA-Z0-9]*$`, s)
	hasUpper := matched && s != strings.ToLower(s)
	return hasUpper
}

func hasUnderscore(s string) bool {
	return strings.Contains(s, "_")
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) <= 1 {
		// 简单的情况：将第二个单词开始首字母大写
		if len(s) > 1 {
			return s[:1] + strings.ToUpper(s[1:2]) + s[2:]
		}
		return s
	}
	
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 生成分析报告
func (analyzer *TemplateFunctionAnalyzer) generateReport() string {
	var report strings.Builder
	
	report.WriteString("# 🔍 YYHertz 模板函数命名分析报告\n\n")
	
	// 1. 基本统计
	report.WriteString("## 📊 基本统计\n\n")
	report.WriteString(fmt.Sprintf("- **总函数数量**: %d\n", len(analyzer.functions)))
	report.WriteString(fmt.Sprintf("- **Go内置动作数量**: %d\n", len(analyzer.goBuiltins)))
	
	// 2. Go内置动作冲突
	conflicts := analyzer.findGoBuiltinConflicts()
	report.WriteString(fmt.Sprintf("- **潜在冲突数量**: %d\n\n", len(conflicts)))
	
	if len(conflicts) > 0 {
		report.WriteString("## ⚠️ Go内置动作冲突\n\n")
		report.WriteString("以下函数名与Go模板内置动作冲突：\n\n")
		for _, conflict := range conflicts {
			actualFunc := analyzer.functions[conflict]
			report.WriteString(fmt.Sprintf("- `%s` → `%s` ⚠️ **冲突风险高**\n", conflict, actualFunc))
		}
		report.WriteString("\n")
	}
	
	// 3. 命名模式分析
	patterns := analyzer.analyzeNamingPatterns()
	report.WriteString("## 🏷️ 函数别名模式分析\n\n")
	report.WriteString(fmt.Sprintf("发现 %d 个函数有多个别名：\n\n", len(patterns)))
	
	for actualFunc, aliases := range patterns {
		report.WriteString(fmt.Sprintf("### `%s`\n", actualFunc))
		report.WriteString("别名: ")
		for i, alias := range aliases {
			if i > 0 {
				report.WriteString(", ")
			}
			report.WriteString(fmt.Sprintf("`%s`", alias))
		}
		report.WriteString("\n\n")
	}
	
	// 4. 缺失的标准别名
	missing := analyzer.findMissingAliases()
	if len(missing) > 0 {
		report.WriteString("## 🔧 建议添加的标准别名\n\n")
		for actualFunc, missingAliases := range missing {
			existing := patterns[actualFunc]
			report.WriteString(fmt.Sprintf("### `%s`\n", actualFunc))
			report.WriteString("现有别名: ")
			for i, alias := range existing {
				if i > 0 {
					report.WriteString(", ")
				}
				report.WriteString(fmt.Sprintf("`%s`", alias))
			}
			report.WriteString("\n建议添加: ")
			for i, alias := range missingAliases {
				if i > 0 {
					report.WriteString(", ")
				}
				report.WriteString(fmt.Sprintf("`%s`", alias))
			}
			report.WriteString("\n\n")
		}
	}
	
	// 5. 命名规范建议
	report.WriteString("## 💡 命名规范建议\n\n")
	report.WriteString("### 标准化原则\n")
	report.WriteString("1. **避免冲突**: 不使用Go内置动作名称\n")
	report.WriteString("2. **多样性**: 提供驼峰、小写、下划线等多种风格\n")
	report.WriteString("3. **一致性**: 同类功能函数使用相似的命名模式\n")
	report.WriteString("4. **简洁性**: 优先选择简短易记的主要名称\n\n")
	
	report.WriteString("### 推荐的函数别名模式\n")
	report.WriteString("```\n")
	report.WriteString("主函数名: actualFunction\n")
	report.WriteString("驼峰别名: camelCaseAlias (如: timeAgo)\n")
	report.WriteString("小写别名: lowercase (如: timeago)\n")
	report.WriteString("下划线别名: under_score (如: time_ago)\n")
	report.WriteString("简化别名: short (如: ago)\n")
	report.WriteString("```\n\n")
	
	return report.String()
}

func main() {
	analyzer := &TemplateFunctionAnalyzer{}
	analyzer.analyzeFunctions()
	
	report := analyzer.generateReport()
	fmt.Print(report)
}