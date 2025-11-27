// Package mapper OGNL风格表达式求值器
//
// 实现MyBatis兼容的表达式求值功能，支持：
// 1. 属性访问（obj.field, obj["field"]）
// 2. 数组/切片访问（arr[0], arr[index]）
// 3. 比较运算符（==, !=, >, <, >=, <=）
// 4. 逻辑运算符（and, or, not）
// 5. 空值检查（null, not null）
// 6. 字符串操作
package mapper

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// OgnlExpressionEvaluator OGNL风格的表达式求值器
type OgnlExpressionEvaluator struct {
	operatorMap map[string]func(left, right any) bool
}

// NewOgnlExpressionEvaluator 创建OGNL表达式求值器
func NewOgnlExpressionEvaluator() *OgnlExpressionEvaluator {
	evaluator := &OgnlExpressionEvaluator{
		operatorMap: make(map[string]func(left, right any) bool),
	}

	// 注册比较运算符
	evaluator.operatorMap["=="] = evaluator.equals
	evaluator.operatorMap["!="] = evaluator.notEquals
	evaluator.operatorMap[">"] = evaluator.greaterThan
	evaluator.operatorMap["<"] = evaluator.lessThan
	evaluator.operatorMap[">="] = evaluator.greaterThanOrEqual
	evaluator.operatorMap["<="] = evaluator.lessThanOrEqual

	return evaluator
}

// EvaluateBoolean 求值布尔表达式
func (eval *OgnlExpressionEvaluator) EvaluateBoolean(expression string, parameters any) bool {
	expression = strings.TrimSpace(expression)
	
	// 处理逻辑运算符
	if strings.Contains(expression, " and ") {
		return eval.evaluateLogicalAnd(expression, parameters)
	}
	if strings.Contains(expression, " or ") {
		return eval.evaluateLogicalOr(expression, parameters)
	}
	
	// 处理not运算符
	if strings.HasPrefix(expression, "not ") {
		return !eval.EvaluateBoolean(expression[4:], parameters)
	}

	// 处理空值检查
	if strings.HasSuffix(expression, " != null") {
		property := strings.TrimSpace(expression[:len(expression)-8])
		value := eval.EvaluateObject(property, parameters)
		return value != nil && !isEmptyValue(value)
	}
	if strings.HasSuffix(expression, " == null") || strings.HasSuffix(expression, " is null") {
		endPos := len(expression) - 8
		if strings.HasSuffix(expression, " is null") {
			endPos = len(expression) - 9
		}
		property := strings.TrimSpace(expression[:endPos])
		value := eval.EvaluateObject(property, parameters)
		return value == nil || isEmptyValue(value)
	}

	// 处理比较运算符
	for op, compareFunc := range eval.operatorMap {
		if strings.Contains(expression, " "+op+" ") {
			parts := strings.SplitN(expression, " "+op+" ", 2)
			if len(parts) == 2 {
				left := eval.EvaluateObject(strings.TrimSpace(parts[0]), parameters)
				right := eval.parseRightOperand(strings.TrimSpace(parts[1]), parameters)
				return compareFunc(left, right)
			}
		}
	}

	// 简单属性检查（非空且非假值）
	value := eval.EvaluateObject(expression, parameters)
	return isTrueValue(value)
}

// EvaluateObject 求值对象表达式
func (eval *OgnlExpressionEvaluator) EvaluateObject(expression string, parameters any) any {
	expression = strings.TrimSpace(expression)
	
	if expression == "" {
		return nil
	}

	// 字面量处理
	if strings.HasPrefix(expression, "'") && strings.HasSuffix(expression, "'") {
		return expression[1 : len(expression)-1]
	}
	if strings.HasPrefix(expression, "\"") && strings.HasSuffix(expression, "\"") {
		return expression[1 : len(expression)-1]
	}

	// 数字字面量
	if num, err := strconv.ParseInt(expression, 10, 64); err == nil {
		return num
	}
	if num, err := strconv.ParseFloat(expression, 64); err == nil {
		return num
	}

	// 布尔字面量
	if expression == "true" {
		return true
	}
	if expression == "false" {
		return false
	}

	// null字面量
	if expression == "null" {
		return nil
	}

	// 属性访问
	return eval.evaluatePropertyAccess(expression, parameters)
}

// EvaluateIterable 求值可迭代表达式
func (eval *OgnlExpressionEvaluator) EvaluateIterable(expression string, parameters any) []any {
	value := eval.EvaluateObject(expression, parameters)
	return eval.toIterableSlice(value)
}

// evaluatePropertyAccess 求值属性访问表达式
func (eval *OgnlExpressionEvaluator) evaluatePropertyAccess(expression string, parameters any) any {
	parts := strings.Split(expression, ".")
	current := parameters
	
	for _, part := range parts {
		current = eval.getProperty(current, part)
		if current == nil {
			return nil
		}
	}
	
	return current
}

// getProperty 获取对象属性
func (eval *OgnlExpressionEvaluator) getProperty(obj any, property string) any {
	if obj == nil {
		return nil
	}

	// 处理数组/切片索引访问 obj[index]
	if strings.Contains(property, "[") && strings.HasSuffix(property, "]") {
		return eval.getIndexedProperty(obj, property)
	}

	// Map访问
	if m, ok := obj.(map[string]any); ok {
		return m[property]
	}

	// 结构体字段访问
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {
		// 首先尝试直接字段名
		if field := v.FieldByName(property); field.IsValid() && field.CanInterface() {
			return field.Interface()
		}

		// 尝试JSON标签匹配
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				tagName := strings.Split(jsonTag, ",")[0]
				if tagName == property && field.IsExported() {
					return v.Field(i).Interface()
				}
			}
		}

		// 尝试忽略大小写匹配
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if strings.EqualFold(field.Name, property) && field.IsExported() {
				return v.Field(i).Interface()
			}
		}
	}

	return nil
}

// getIndexedProperty 获取索引属性 obj[index]
func (eval *OgnlExpressionEvaluator) getIndexedProperty(obj any, property string) any {
	bracketStart := strings.Index(property, "[")
	bracketEnd := strings.LastIndex(property, "]")
	
	if bracketStart == -1 || bracketEnd == -1 || bracketEnd <= bracketStart {
		return nil
	}

	baseProp := property[:bracketStart]
	indexExpr := property[bracketStart+1 : bracketEnd]

	// 获取基础对象
	baseObj := obj
	if baseProp != "" {
		baseObj = eval.getProperty(obj, baseProp)
	}

	if baseObj == nil {
		return nil
	}

	// 解析索引
	var index int
	if indexNum, err := strconv.Atoi(indexExpr); err == nil {
		index = indexNum
	} else {
		// 索引是变量，需要求值
		indexValue := eval.EvaluateObject(indexExpr, obj)
		if indexInt, ok := indexValue.(int); ok {
			index = indexInt
		} else if indexInt64, ok := indexValue.(int64); ok {
			index = int(indexInt64)
		} else {
			return nil
		}
	}

	// 根据类型访问索引
	v := reflect.ValueOf(baseObj)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if index >= 0 && index < v.Len() {
			return v.Index(index).Interface()
		}
	case reflect.Map:
		// Map的键访问
		keyValue := reflect.ValueOf(index)
		if result := v.MapIndex(keyValue); result.IsValid() {
			return result.Interface()
		}
	}

	return nil
}

// 比较运算符实现

func (eval *OgnlExpressionEvaluator) equals(left, right any) bool {
	return compareValues(left, right) == 0
}

func (eval *OgnlExpressionEvaluator) notEquals(left, right any) bool {
	return compareValues(left, right) != 0
}

func (eval *OgnlExpressionEvaluator) greaterThan(left, right any) bool {
	return compareValues(left, right) > 0
}

func (eval *OgnlExpressionEvaluator) lessThan(left, right any) bool {
	return compareValues(left, right) < 0
}

func (eval *OgnlExpressionEvaluator) greaterThanOrEqual(left, right any) bool {
	return compareValues(left, right) >= 0
}

func (eval *OgnlExpressionEvaluator) lessThanOrEqual(left, right any) bool {
	return compareValues(left, right) <= 0
}

// evaluateLogicalAnd 求值逻辑AND表达式
func (eval *OgnlExpressionEvaluator) evaluateLogicalAnd(expression string, parameters any) bool {
	parts := strings.Split(expression, " and ")
	for _, part := range parts {
		if !eval.EvaluateBoolean(strings.TrimSpace(part), parameters) {
			return false
		}
	}
	return true
}

// evaluateLogicalOr 求值逻辑OR表达式
func (eval *OgnlExpressionEvaluator) evaluateLogicalOr(expression string, parameters any) bool {
	parts := strings.Split(expression, " or ")
	for _, part := range parts {
		if eval.EvaluateBoolean(strings.TrimSpace(part), parameters) {
			return true
		}
	}
	return false
}

// parseRightOperand 解析右操作数
func (eval *OgnlExpressionEvaluator) parseRightOperand(operand string, parameters any) any {
	// 如果是字面量，直接返回
	if strings.HasPrefix(operand, "'") && strings.HasSuffix(operand, "'") {
		return operand[1 : len(operand)-1]
	}
	if strings.HasPrefix(operand, "\"") && strings.HasSuffix(operand, "\"") {
		return operand[1 : len(operand)-1]
	}

	// 数字字面量
	if num, err := strconv.ParseInt(operand, 10, 64); err == nil {
		return num
	}
	if num, err := strconv.ParseFloat(operand, 64); err == nil {
		return num
	}

	// 布尔字面量
	if operand == "true" {
		return true
	}
	if operand == "false" {
		return false
	}

	// 否则作为属性访问
	return eval.EvaluateObject(operand, parameters)
}

// toIterableSlice 转换为可迭代切片
func (eval *OgnlExpressionEvaluator) toIterableSlice(value any) []any {
	if value == nil {
		return []any{}
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result
	case reflect.Map:
		result := make([]any, 0, v.Len())
		for _, key := range v.MapKeys() {
			result = append(result, v.MapIndex(key).Interface())
		}
		return result
	default:
		// 单个元素当作长度为1的切片
		return []any{value}
	}
}

// 工具函数

// compareValues 比较两个值，返回-1, 0, 1
func compareValues(left, right any) int {
	if left == nil && right == nil {
		return 0
	}
	if left == nil {
		return -1
	}
	if right == nil {
		return 1
	}

	// 尝试数字比较
	leftNum, leftIsNum := toNumber(left)
	rightNum, rightIsNum := toNumber(right)
	if leftIsNum && rightIsNum {
		if leftNum < rightNum {
			return -1
		} else if leftNum > rightNum {
			return 1
		}
		return 0
	}

	// 字符串比较
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)
	if leftStr < rightStr {
		return -1
	} else if leftStr > rightStr {
		return 1
	}
	return 0
}

// toNumber 转换为数字
func toNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

// isEmptyValue 检查是否为空值
func isEmptyValue(value any) bool {
	if value == nil {
		return true
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// isTrueValue 检查是否为真值
func isTrueValue(value any) bool {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != ""
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int() != 0
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(v).Uint() != 0
	case float32, float64:
		return reflect.ValueOf(v).Float() != 0
	default:
		// 非空对象为真
		return !isEmptyValue(v)
	}
}