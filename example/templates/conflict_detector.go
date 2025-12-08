package main

import (
	"fmt"
	"sort"
	"strings"
)

// Go内置动作冲突检测和修复工具
type ConflictDetector struct {
	// Go模板内置动作/函数集合，用于快速判断是否与自定义函数重名
	goBuiltins map[string]bool
	// 当前工程内的模板函数映射：funcName -> 实际实现（用于报告与建议）
	currentFuncs map[string]string
	// 已检测到的冲突列表（包含风险级别、说明、建议等）
	conflicts []ConflictInfo
	// 针对每个冲突函数名的重命名建议集合
	suggestions map[string][]string
}

type ConflictInfo struct {
	FuncName    string
	ActualFunc  string
	RiskLevel   string
	Suggestion  []string
	Description string
}

// 初始化冲突检测器
func NewConflictDetector() *ConflictDetector {
	// Go模板内置动作 - 完整列表
	builtins := map[string]bool{
		// 控制结构
		"if":       true,
		"else":     true,
		"end":      true,
		"range":    true,
		"with":     true,
		"template": true,
		"define":   true,
		"block":    true,

		// 逻辑运算
		"and": true,
		"or":  true,
		"not": true,

		// 比较运算
		"eq": true,
		"ne": true,
		"lt": true,
		"le": true,
		"gt": true,
		"ge": true,

		// 内置函数
		"len":     true,
		"index":   true,
		"slice":   true,
		"printf":  true,
		"print":   true,
		"println": true,
		"call":    true,

		// HTML/URL 相关
		"html":     true,
		"js":       true,
		"urlquery": true,
	}

	// 当前函数映射 - 从分析报告中提取
	currentFuncs := map[string]string{
		"and":    "And",
		"eq":     "Eq",
		"ge":     "Ge",
		"gt":     "Gt",
		"index":  "Index",
		"le":     "Le",
		"len":    "Len",
		"lt":     "Lt",
		"ne":     "Ne",
		"not":    "Not",
		"or":     "Or",
		"printf": "fmt.Sprintf",
		"slice":  "Slice",
	}

	return &ConflictDetector{
		goBuiltins:   builtins,
		currentFuncs: currentFuncs,
		conflicts:    make([]ConflictInfo, 0),
		suggestions:  make(map[string][]string),
	}
}

// 检测所有冲突
func (cd *ConflictDetector) DetectConflicts() {
	for funcName, actualFunc := range cd.currentFuncs {
		if cd.goBuiltins[funcName] {
			conflict := ConflictInfo{
				FuncName:    funcName,
				ActualFunc:  actualFunc,
				RiskLevel:   cd.assessRisk(funcName),
				Suggestion:  cd.generateSuggestions(funcName, actualFunc),
				Description: cd.getConflictDescription(funcName),
			}
			cd.conflicts = append(cd.conflicts, conflict)
		}
	}

	// 按风险级别排序
	sort.Slice(cd.conflicts, func(i, j int) bool {
		riskOrder := map[string]int{"Critical": 3, "High": 2, "Medium": 1, "Low": 0}
		return riskOrder[cd.conflicts[i].RiskLevel] > riskOrder[cd.conflicts[j].RiskLevel]
	})
}

// 评估冲突风险级别
func (cd *ConflictDetector) assessRisk(funcName string) string {
	// 高风险：经常使用的内置动作
	highRisk := []string{"and", "or", "not", "eq", "ne", "lt", "le", "gt", "ge", "len", "index"}

	// 中风险：较常用的内置函数
	mediumRisk := []string{"printf", "slice"}

	// 低风险：不太常用的内置函数
	// lowRisk := []string{"print", "println", "call", "html", "js", "urlquery"}

	for _, high := range highRisk {
		if funcName == high {
			return "Critical"
		}
	}

	for _, medium := range mediumRisk {
		if funcName == medium {
			return "High"
		}
	}

	return "Medium"
}

// 生成修复建议
func (cd *ConflictDetector) generateSuggestions(funcName, actualFunc string) []string {
	suggestions := make([]string, 0)

	switch funcName {
	case "and":
		suggestions = []string{"logicAnd", "andOp", "logical_and", "boolAnd"}
	case "or":
		suggestions = []string{"logicOr", "orOp", "logical_or", "boolOr"}
	case "not":
		suggestions = []string{"logicNot", "notOp", "logical_not", "boolNot"}
	case "eq":
		suggestions = []string{"equal", "equals", "isEqual", "compare_eq"}
	case "ne":
		suggestions = []string{"notEqual", "notEquals", "isNotEqual", "compare_ne"}
	case "lt":
		suggestions = []string{"lessThan", "isLess", "compare_lt", "less"}
	case "le":
		suggestions = []string{"lessEqual", "lessOrEqual", "compare_le", "lte"}
	case "gt":
		suggestions = []string{"greaterThan", "isGreater", "compare_gt", "greater"}
	case "ge":
		suggestions = []string{"greaterEqual", "greaterOrEqual", "compare_ge", "gte"}
	case "len":
		suggestions = []string{"length", "size", "count", "getLen"}
	case "index":
		suggestions = []string{"getIndex", "arrayIndex", "sliceIndex", "indexOf"}
	case "slice":
		suggestions = []string{"subSlice", "arraySlice", "getSlice", "slicePart"}
	case "printf":
		suggestions = []string{"sprintf", "format", "formatStr", "strFormat"}
	default:
		// 通用策略：添加前缀或后缀
		base := strings.ToLower(funcName)
		suggestions = []string{
			"tmpl" + strings.Title(funcName),   // tmplEq, tmplLen
			base + "Func",                      // eqFunc, lenFunc
			"custom" + strings.Title(funcName), // customEq, customLen
			base + "_op",                       // eq_op, len_op
		}
	}

	return suggestions
}

// 获取冲突描述
func (cd *ConflictDetector) getConflictDescription(funcName string) string {
	descriptions := map[string]string{
		"and":    "与Go模板逻辑AND运算符冲突，可能导致逻辑表达式解析错误",
		"or":     "与Go模板逻辑OR运算符冲突，可能导致逻辑表达式解析错误",
		"not":    "与Go模板逻辑NOT运算符冲突，可能导致逻辑表达式解析错误",
		"eq":     "与Go模板等于比较运算符冲突，可能导致条件判断失效",
		"ne":     "与Go模板不等于比较运算符冲突，可能导致条件判断失效",
		"lt":     "与Go模板小于比较运算符冲突，可能导致条件判断失效",
		"le":     "与Go模板小于等于比较运算符冲突，可能导致条件判断失效",
		"gt":     "与Go模板大于比较运算符冲突，可能导致条件判断失效",
		"ge":     "与Go模板大于等于比较运算符冲突，可能导致条件判断失效",
		"len":    "与Go模板内置len函数冲突，可能导致长度计算错误",
		"index":  "与Go模板内置index函数冲突，可能导致索引访问错误",
		"slice":  "与Go模板内置slice函数冲突，可能导致切片操作错误",
		"printf": "与Go模板内置printf函数冲突，可能导致格式化输出错误",
	}

	if desc, exists := descriptions[funcName]; exists {
		return desc
	}
	return fmt.Sprintf("与Go模板内置'%s'功能冲突，可能导致模板解析或执行错误", funcName)
}

// 生成冲突修复代码
func (cd *ConflictDetector) generateFixCode() string {
	var code strings.Builder

	code.WriteString("// 🔧 Go内置动作冲突修复建议\n")
	code.WriteString("// 以下函数名与Go模板内置动作冲突，建议重命名：\n\n")

	for _, conflict := range cd.conflicts {
		code.WriteString(fmt.Sprintf("// ❌ 冲突函数: %s → %s (%s风险)\n",
			conflict.FuncName, conflict.ActualFunc, conflict.RiskLevel))
		code.WriteString(fmt.Sprintf("// 说明: %s\n", conflict.Description))
		code.WriteString("// 建议的安全别名：\n")

		for i, suggestion := range conflict.Suggestion {
			if i == 0 {
				code.WriteString(fmt.Sprintf(`"%s": %s, // 推荐使用`, suggestion, conflict.ActualFunc))
				code.WriteString(fmt.Sprintf(" (替代冲突的'%s')\n", conflict.FuncName))
			} else {
				code.WriteString(fmt.Sprintf(`"%s": %s, // 备选方案%d`, suggestion, conflict.ActualFunc, i))
				code.WriteString("\n")
			}
		}
		code.WriteString("\n")
	}

	return code.String()
}

// 生成详细的冲突报告
func (cd *ConflictDetector) GenerateReport() string {
	var report strings.Builder

	report.WriteString("# 🚨 Go模板内置动作冲突检测报告\n\n")

	// 概要统计
	criticalCount := 0
	highCount := 0
	mediumCount := 0

	for _, conflict := range cd.conflicts {
		switch conflict.RiskLevel {
		case "Critical":
			criticalCount++
		case "High":
			highCount++
		case "Medium":
			mediumCount++
		}
	}

	report.WriteString("## 📊 风险统计\n\n")
	report.WriteString(fmt.Sprintf("- 🔴 **Critical风险**: %d个\n", criticalCount))
	report.WriteString(fmt.Sprintf("- 🟡 **High风险**: %d个\n", highCount))
	report.WriteString(fmt.Sprintf("- 🟢 **Medium风险**: %d个\n", mediumCount))
	report.WriteString(fmt.Sprintf("- **总冲突数**: %d个\n\n", len(cd.conflicts)))

	// 详细冲突列表
	report.WriteString("## 🔍 详细冲突分析\n\n")

	currentRisk := ""
	for _, conflict := range cd.conflicts {
		if conflict.RiskLevel != currentRisk {
			currentRisk = conflict.RiskLevel
			var emoji string
			switch currentRisk {
			case "Critical":
				emoji = "🔴"
			case "High":
				emoji = "🟡"
			case "Medium":
				emoji = "🟢"
			}
			report.WriteString(fmt.Sprintf("### %s %s风险\n\n", emoji, currentRisk))
		}

		report.WriteString(fmt.Sprintf("#### `%s` → `%s`\n", conflict.FuncName, conflict.ActualFunc))
		report.WriteString(fmt.Sprintf("**问题描述**: %s\n\n", conflict.Description))
		report.WriteString("**修复建议**:\n")
		for i, suggestion := range conflict.Suggestion {
			priority := ""
			if i == 0 {
				priority = " (推荐)"
			}
			report.WriteString(fmt.Sprintf("- `%s`%s\n", suggestion, priority))
		}
		report.WriteString("\n")
	}

	// 修复代码示例
	report.WriteString("## 🔧 代码修复示例\n\n")
	report.WriteString("```go\n")
	report.WriteString(cd.generateFixCode())
	report.WriteString("```\n\n")

	// 优先级建议
	report.WriteString("## 🎯 修复优先级建议\n\n")
	report.WriteString("1. **立即修复Critical风险**: 这些冲突会导致模板解析完全失败\n")
	report.WriteString("2. **尽快修复High风险**: 这些冲突会导致功能异常但不会完全阻塞\n")
	report.WriteString("3. **计划修复Medium风险**: 这些冲突在特定场景下可能出现问题\n\n")

	report.WriteString("## 📋 修复检查清单\n\n")
	for i, conflict := range cd.conflicts {
		report.WriteString(fmt.Sprintf("- [ ] **%s**: 重命名 `%s` 为 `%s`\n",
			conflict.RiskLevel, conflict.FuncName, conflict.Suggestion[0]))
		if i < 5 { // 只显示前5个作为示例
			continue
		} else if i == 5 {
			report.WriteString("- [ ] ... (更多项目请参考详细列表)\n")
			break
		}
	}

	return report.String()
}

// func main() {
// 	detector := NewConflictDetector()
// 	detector.DetectConflicts()

// 	report := detector.GenerateReport()
// 	fmt.Print(report)
// }
