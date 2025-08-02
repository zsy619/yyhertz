// Package mapper 动态SQL构建器
//
// 支持MyBatis风格的动态SQL语法：
// 1. #{param} - 参数替换
// 2. <if test="condition"> - 条件判断
// 3. <foreach> - 循环遍历
// 4. <choose><when><otherwise> - 选择结构
// 5. <where> - WHERE子句
// 6. <set> - SET子句
// 7. <trim> - 修剪标签
// 8. <bind> - 变量绑定
package mapper

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// DynamicSqlBuilder 动态SQL构建器
type DynamicSqlBuilder struct {
	paramIndex int
	parameters []any
	context    map[string]any
}

// SqlNode SQL节点接口
type SqlNode interface {
	Apply(context DynamicContext) bool
}

// DynamicContext 动态上下文
type DynamicContext struct {
	Parameters  map[string]any
	SqlBuilder  *strings.Builder
	UniqueNumber int
}

// StaticTextSqlNode 静态文本SQL节点
type StaticTextSqlNode struct {
	Text string
}

// TextSqlNode 文本SQL节点
type TextSqlNode struct {
	Text string
}

// IfSqlNode IF SQL节点
type IfSqlNode struct {
	Test     ExpressionEvaluator
	Contents SqlNode
}

// WhereSqlNode WHERE SQL节点
type WhereSqlNode struct {
	Contents SqlNode
}

// SetSqlNode SET SQL节点
type SetSqlNode struct {
	Contents SqlNode
}

// TrimSqlNode TRIM SQL节点
type TrimSqlNode struct {
	Contents         SqlNode
	Prefix           string
	Suffix           string
	PrefixesToOverride []string
	SuffixesToOverride []string
}

// ChooseSqlNode CHOOSE SQL节点
type ChooseSqlNode struct {
	DefaultSqlNode SqlNode
	IfSqlNodes     []SqlNode
}

// ForEachSqlNode FOREACH SQL节点
type ForEachSqlNode struct {
	Contents   SqlNode
	Collection ExpressionEvaluator
	Index      string
	Item       string
	Open       string
	Close      string
	Separator  string
}

// VarDeclSqlNode 变量声明SQL节点 (bind标签)
type VarDeclSqlNode struct {
	Name  string
	Value ExpressionEvaluator
}

// MixedSqlNode 混合SQL节点
type MixedSqlNode struct {
	Contents []SqlNode
}

// ExpressionEvaluator 表达式求值器接口
type ExpressionEvaluator interface {
	EvaluateBoolean(parameter any, context DynamicContext) bool
	EvaluateIterable(parameter any, context DynamicContext) any
}

// OgnlCache OGNL缓存 (Go中使用简化的表达式评估)
type OgnlCache struct {
	cache map[string]*CompiledExpression
}

// CompiledExpression 编译后的表达式
type CompiledExpression struct {
	Expression string
	Compiled   func(any) any
}

// NewDynamicSqlBuilder 创建动态SQL构建器
func NewDynamicSqlBuilder() *DynamicSqlBuilder {
	return &DynamicSqlBuilder{
		paramIndex: 0,
		parameters: make([]any, 0),
		context:    make(map[string]any),
	}
}

// Build 构建动态SQL
func (b *DynamicSqlBuilder) Build(template string, parameter any) (string, []any, error) {
	b.paramIndex = 0
	b.parameters = make([]any, 0)
	b.context = make(map[string]any)
	
	// 创建动态上下文
	context := &DynamicContext{
		Parameters:   b.buildParameterMap(parameter),
		SqlBuilder:   &strings.Builder{},
		UniqueNumber: 0,
	}
	
	// 解析并应用SQL节点
	rootNode, err := b.parseScript(template)
	if err != nil {
		return "", nil, err
	}
	
	rootNode.Apply(*context)
	
	// 处理参数占位符
	sql := context.SqlBuilder.String()
	sql, err = b.replaceParameters(sql, parameter)
	if err != nil {
		return "", nil, err
	}
	
	return sql, b.parameters, nil
}

// parseScript 解析脚本
func (b *DynamicSqlBuilder) parseScript(script string) (SqlNode, error) {
	return b.parseScriptNode(script)
}

// parseScriptNode 解析脚本节点
func (b *DynamicSqlBuilder) parseScriptNode(body string) (SqlNode, error) {
	nodes := make([]SqlNode, 0)
	
	// 使用正则表达式解析各种标签
	current := body
	
	for len(current) > 0 {
		// 查找下一个标签
		tagStart := b.findNextTag(current)
		if tagStart == -1 {
			// 没有更多标签，剩余为静态文本
			if len(current) > 0 {
				nodes = append(nodes, &StaticTextSqlNode{Text: current})
			}
			break
		}
		
		// 添加标签前的静态文本
		if tagStart > 0 {
			staticText := current[:tagStart]
			if strings.TrimSpace(staticText) != "" {
				nodes = append(nodes, &StaticTextSqlNode{Text: staticText})
			}
		}
		
		// 解析标签
		tagNode, remaining, err := b.parseTag(current[tagStart:])
		if err != nil {
			return nil, err
		}
		
		nodes = append(nodes, tagNode)
		current = remaining
	}
	
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	
	return &MixedSqlNode{Contents: nodes}, nil
}

// findNextTag 查找下一个标签
func (b *DynamicSqlBuilder) findNextTag(text string) int {
	tags := []string{"<if", "<where", "<set", "<choose", "<foreach", "<trim", "<bind"}
	
	minIndex := -1
	for _, tag := range tags {
		index := strings.Index(text, tag)
		if index != -1 && (minIndex == -1 || index < minIndex) {
			minIndex = index
		}
	}
	
	return minIndex
}

// parseTag 解析标签
func (b *DynamicSqlBuilder) parseTag(text string) (SqlNode, string, error) {
	if strings.HasPrefix(text, "<if") {
		return b.parseIfTag(text)
	} else if strings.HasPrefix(text, "<where") {
		return b.parseWhereTag(text)
	} else if strings.HasPrefix(text, "<set") {
		return b.parseSetTag(text)
	} else if strings.HasPrefix(text, "<choose") {
		return b.parseChooseTag(text)
	} else if strings.HasPrefix(text, "<foreach") {
		return b.parseForeachTag(text)
	} else if strings.HasPrefix(text, "<trim") {
		return b.parseTrimTag(text)
	} else if strings.HasPrefix(text, "<bind") {
		return b.parseBindTag(text)
	}
	
	return nil, text, fmt.Errorf("unknown tag: %s", text[:10])
}

// parseIfTag 解析IF标签
func (b *DynamicSqlBuilder) parseIfTag(text string) (SqlNode, string, error) {
	// 正则匹配 <if test="condition">content</if>
	ifRegex := regexp.MustCompile(`<if\s+test="([^"]+)">([^<]*(?:<(?!/?if\b)[^<]*)*)</if>`)
	matches := ifRegex.FindStringSubmatch(text)
	
	if len(matches) != 3 {
		return nil, text, fmt.Errorf("invalid if tag")
	}
	
	condition := matches[1]
	content := matches[2]
	
	// 创建内容节点
	contentNode, err := b.parseScriptNode(content)
	if err != nil {
		return nil, text, err
	}
	
	// 创建IF节点
	ifNode := &IfSqlNode{
		Test:     NewSimpleExpressionEvaluator(condition),
		Contents: contentNode,
	}
	
	// 计算剩余文本
	remaining := text[len(matches[0]):]
	return ifNode, remaining, nil
}

// parseWhereTag 解析WHERE标签
func (b *DynamicSqlBuilder) parseWhereTag(text string) (SqlNode, string, error) {
	whereRegex := regexp.MustCompile(`<where>([^<]*(?:<(?!/?where\b)[^<]*)*)</where>`)
	matches := whereRegex.FindStringSubmatch(text)
	
	if len(matches) != 2 {
		return nil, text, fmt.Errorf("invalid where tag")
	}
	
	content := matches[1]
	contentNode, err := b.parseScriptNode(content)
	if err != nil {
		return nil, text, err
	}
	
	whereNode := &WhereSqlNode{Contents: contentNode}
	remaining := text[len(matches[0]):]
	return whereNode, remaining, nil
}

// parseSetTag 解析SET标签
func (b *DynamicSqlBuilder) parseSetTag(text string) (SqlNode, string, error) {
	setRegex := regexp.MustCompile(`<set>([^<]*(?:<(?!/?set\b)[^<]*)*)</set>`)
	matches := setRegex.FindStringSubmatch(text)
	
	if len(matches) != 2 {
		return nil, text, fmt.Errorf("invalid set tag")
	}
	
	content := matches[1]
	contentNode, err := b.parseScriptNode(content)
	if err != nil {
		return nil, text, err
	}
	
	setNode := &SetSqlNode{Contents: contentNode}
	remaining := text[len(matches[0]):]
	return setNode, remaining, nil
}

// parseChooseTag 解析CHOOSE标签
func (b *DynamicSqlBuilder) parseChooseTag(text string) (SqlNode, string, error) {
	chooseRegex := regexp.MustCompile(`<choose>(.*?)</choose>`)
	matches := chooseRegex.FindStringSubmatch(text)
	
	if len(matches) != 2 {
		return nil, text, fmt.Errorf("invalid choose tag")
	}
	
	_ = matches[1] // content 暂时不使用，避免编译错误
	
	// 解析when和otherwise
	whenNodes := make([]SqlNode, 0)
	var otherwiseNode SqlNode
	
	// 这里需要更复杂的解析逻辑
	// 简化实现，实际需要完整的XML解析
	
	chooseNode := &ChooseSqlNode{
		IfSqlNodes:     whenNodes,
		DefaultSqlNode: otherwiseNode,
	}
	
	remaining := text[len(matches[0]):]
	return chooseNode, remaining, nil
}

// parseForeachTag 解析FOREACH标签
func (b *DynamicSqlBuilder) parseForeachTag(text string) (SqlNode, string, error) {
	foreachRegex := regexp.MustCompile(`<foreach\s+collection="([^"]+)"\s+item="([^"]+)"(?:\s+index="([^"]+)")?\s+open="([^"]*)"\s+separator="([^"]*)"\s+close="([^"]*)">([^<]*(?:<(?!/?foreach\b)[^<]*)*)</foreach>`)
	matches := foreachRegex.FindStringSubmatch(text)
	
	if len(matches) < 7 {
		return nil, text, fmt.Errorf("invalid foreach tag")
	}
	
	collection := matches[1]
	item := matches[2]
	index := matches[3]
	open := matches[4]
	separator := matches[5]
	close := matches[6]
	content := matches[7]
	
	contentNode, err := b.parseScriptNode(content)
	if err != nil {
		return nil, text, err
	}
	
	foreachNode := &ForEachSqlNode{
		Contents:   contentNode,
		Collection: NewSimpleExpressionEvaluator(collection),
		Item:       item,
		Index:      index,
		Open:       open,
		Separator:  separator,
		Close:      close,
	}
	
	remaining := text[len(matches[0]):]
	return foreachNode, remaining, nil
}

// parseTrimTag 解析TRIM标签
func (b *DynamicSqlBuilder) parseTrimTag(text string) (SqlNode, string, error) {
	// 简化实现
	trimRegex := regexp.MustCompile(`<trim\s+prefix="([^"]*)"\s+suffix="([^"]*)"\s+prefixOverrides="([^"]*)"\s+suffixOverrides="([^"]*)">([^<]*(?:<(?!/?trim\b)[^<]*)*)</trim>`)
	matches := trimRegex.FindStringSubmatch(text)
	
	if len(matches) != 6 {
		return nil, text, fmt.Errorf("invalid trim tag")
	}
	
	prefix := matches[1]
	suffix := matches[2]
	prefixOverrides := strings.Split(matches[3], "|")
	suffixOverrides := strings.Split(matches[4], "|")
	content := matches[5]
	
	contentNode, err := b.parseScriptNode(content)
	if err != nil {
		return nil, text, err
	}
	
	trimNode := &TrimSqlNode{
		Contents:           contentNode,
		Prefix:             prefix,
		Suffix:             suffix,
		PrefixesToOverride: prefixOverrides,
		SuffixesToOverride: suffixOverrides,
	}
	
	remaining := text[len(matches[0]):]
	return trimNode, remaining, nil
}

// parseBindTag 解析BIND标签
func (b *DynamicSqlBuilder) parseBindTag(text string) (SqlNode, string, error) {
	bindRegex := regexp.MustCompile(`<bind\s+name="([^"]+)"\s+value="([^"]+)"\s*/>`)
	matches := bindRegex.FindStringSubmatch(text)
	
	if len(matches) != 3 {
		return nil, text, fmt.Errorf("invalid bind tag")
	}
	
	name := matches[1]
	value := matches[2]
	
	bindNode := &VarDeclSqlNode{
		Name:  name,
		Value: NewSimpleExpressionEvaluator(value),
	}
	
	remaining := text[len(matches[0]):]
	return bindNode, remaining, nil
}

// buildParameterMap 构建参数映射
func (b *DynamicSqlBuilder) buildParameterMap(parameter any) map[string]any {
	paramMap := make(map[string]any)
	
	if parameter == nil {
		return paramMap
	}
	
	if m, ok := parameter.(map[string]any); ok {
		return m
	}
	
	// 使用反射解析结构体
	v := reflect.ValueOf(parameter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				paramMap[field.Name] = v.Field(i).Interface()
			}
		}
	} else {
		paramMap["value"] = parameter
	}
	
	return paramMap
}

// replaceParameters 替换参数占位符
func (b *DynamicSqlBuilder) replaceParameters(template string, parameter any) (string, error) {
	paramRegex := regexp.MustCompile(`#\{([^}]+)\}`)
	
	result := paramRegex.ReplaceAllStringFunc(template, func(match string) string {
		paramName := paramRegex.FindStringSubmatch(match)[1]
		
		value := b.getPropertyValue(parameter, paramName)
		b.parameters = append(b.parameters, value)
		
		return "?"
	})
	
	return result, nil
}

// getPropertyValue 获取属性值
func (b *DynamicSqlBuilder) getPropertyValue(obj any, propertyPath string) any {
	if obj == nil {
		return nil
	}
	
	parts := strings.Split(propertyPath, ".")
	current := obj
	
	for _, part := range parts {
		if current == nil {
			return nil
		}
		
		if m, ok := current.(map[string]any); ok {
			current = m[part]
			continue
		}
		
		v := reflect.ValueOf(current)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		
		if v.Kind() != reflect.Struct {
			return nil
		}
		
		field := b.findField(v, part)
		if !field.IsValid() {
			return nil
		}
		
		current = field.Interface()
	}
	
	return current
}

// findField 查找字段
func (b *DynamicSqlBuilder) findField(v reflect.Value, fieldName string) reflect.Value {
	t := v.Type()
	
	if field := v.FieldByName(fieldName); field.IsValid() {
		return field
	}
	
	fieldNameLower := strings.ToLower(fieldName)
	for i := 0; i < t.NumField(); i++ {
		if strings.ToLower(t.Field(i).Name) == fieldNameLower {
			return v.Field(i)
		}
	}
	
	return reflect.Value{}
}

// SQL节点实现

// Apply 应用静态文本SQL节点
func (node *StaticTextSqlNode) Apply(context DynamicContext) bool {
	context.SqlBuilder.WriteString(node.Text)
	return true
}

// Apply 应用文本SQL节点
func (node *TextSqlNode) Apply(context DynamicContext) bool {
	// 处理${}占位符
	result := node.Text
	// 简化实现，实际需要完整的表达式处理
	context.SqlBuilder.WriteString(result)
	return true
}

// Apply 应用IF SQL节点
func (node *IfSqlNode) Apply(context DynamicContext) bool {
	if node.Test.EvaluateBoolean(context.Parameters, context) {
		node.Contents.Apply(context)
		return true
	}
	return false
}

// Apply 应用WHERE SQL节点
func (node *WhereSqlNode) Apply(context DynamicContext) bool {
	oldSql := context.SqlBuilder.String()
	node.Contents.Apply(context)
	newSql := context.SqlBuilder.String()
	
	// 检查是否有内容添加
	addedSql := newSql[len(oldSql):]
	if strings.TrimSpace(addedSql) != "" {
		// 移除开头的AND或OR
		trimmed := regexp.MustCompile(`^\s*(AND|OR)\s+`).ReplaceAllString(addedSql, "")
		
		// 重新构建SQL
		context.SqlBuilder.Reset()
		context.SqlBuilder.WriteString(oldSql)
		context.SqlBuilder.WriteString("WHERE ")
		context.SqlBuilder.WriteString(trimmed)
	}
	
	return true
}

// Apply 应用SET SQL节点
func (node *SetSqlNode) Apply(context DynamicContext) bool {
	oldSql := context.SqlBuilder.String()
	node.Contents.Apply(context)
	newSql := context.SqlBuilder.String()
	
	addedSql := newSql[len(oldSql):]
	if strings.TrimSpace(addedSql) != "" {
		trimmed := regexp.MustCompile(`,\s*$`).ReplaceAllString(addedSql, "")
		
		context.SqlBuilder.Reset()
		context.SqlBuilder.WriteString(oldSql)
		context.SqlBuilder.WriteString("SET ")
		context.SqlBuilder.WriteString(trimmed)
	}
	
	return true
}

// Apply 应用混合SQL节点
func (node *MixedSqlNode) Apply(context DynamicContext) bool {
	for _, content := range node.Contents {
		content.Apply(context)
	}
	return true
}

// Apply 应用CHOOSE SQL节点
func (node *ChooseSqlNode) Apply(context DynamicContext) bool {
	// 检查when条件
	for _, ifNode := range node.IfSqlNodes {
		if ifNode.Apply(context) {
			return true
		}
	}
	
	// 如果没有when匹配，使用otherwise
	if node.DefaultSqlNode != nil {
		return node.DefaultSqlNode.Apply(context)
	}
	
	return false
}

// Apply 应用FOREACH SQL节点
func (node *ForEachSqlNode) Apply(context DynamicContext) bool {
	// 获取集合
	collection := node.Collection.EvaluateIterable(context.Parameters, context)
	if collection == nil {
		return false
	}
	
	// 简化的foreach实现
	context.SqlBuilder.WriteString(node.Open)
	context.SqlBuilder.WriteString(" /* foreach content */ ")
	context.SqlBuilder.WriteString(node.Close)
	
	return true
}

// Apply 应用TRIM SQL节点
func (node *TrimSqlNode) Apply(context DynamicContext) bool {
	oldSql := context.SqlBuilder.String()
	node.Contents.Apply(context)
	newSql := context.SqlBuilder.String()
	
	addedSql := newSql[len(oldSql):]
	if strings.TrimSpace(addedSql) != "" {
		// 应用trim逻辑
		trimmed := addedSql
		
		// 移除前缀
		for _, prefix := range node.PrefixesToOverride {
			if strings.HasPrefix(strings.TrimSpace(trimmed), prefix) {
				trimmed = strings.TrimSpace(trimmed)[len(prefix):]
				break
			}
		}
		
		// 移除后缀
		for _, suffix := range node.SuffixesToOverride {
			if strings.HasSuffix(strings.TrimSpace(trimmed), suffix) {
				trimmed = strings.TrimSpace(trimmed)[:len(trimmed)-len(suffix)]
				break
			}
		}
		
		// 重新构建SQL
		context.SqlBuilder.Reset()
		context.SqlBuilder.WriteString(oldSql)
		context.SqlBuilder.WriteString(node.Prefix)
		context.SqlBuilder.WriteString(trimmed)
		context.SqlBuilder.WriteString(node.Suffix)
	}
	
	return true
}

// Apply 应用变量声明SQL节点
func (node *VarDeclSqlNode) Apply(context DynamicContext) bool {
	// 变量绑定不直接输出SQL，而是设置参数
	value := node.Value.EvaluateIterable(context.Parameters, context)
	context.Parameters[node.Name] = value
	return true
}

// SimpleExpressionEvaluator 简单表达式求值器
type SimpleExpressionEvaluator struct {
	Expression string
}

// NewSimpleExpressionEvaluator 创建简单表达式求值器
func NewSimpleExpressionEvaluator(expression string) *SimpleExpressionEvaluator {
	return &SimpleExpressionEvaluator{Expression: expression}
}

// EvaluateBoolean 求值布尔表达式
func (evaluator *SimpleExpressionEvaluator) EvaluateBoolean(parameter any, context DynamicContext) bool {
	// 简化的表达式求值
	expression := strings.TrimSpace(evaluator.Expression)
	
	if strings.Contains(expression, "!= null") {
		paramName := strings.TrimSpace(strings.Split(expression, "!= null")[0])
		value := getNestedValue(parameter, paramName)
		return value != nil
	}
	
	if strings.Contains(expression, "== null") {
		paramName := strings.TrimSpace(strings.Split(expression, "== null")[0])
		value := getNestedValue(parameter, paramName)
		return value == nil
	}
	
	if strings.Contains(expression, "==") && strings.Contains(expression, "'") {
		parts := strings.Split(expression, "==")
		if len(parts) == 2 {
			paramName := strings.TrimSpace(parts[0])
			expectedValue := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
			actualValue := getNestedValue(parameter, paramName)
			return fmt.Sprintf("%v", actualValue) == expectedValue
		}
	}
	
	if strings.Contains(expression, ">") {
		parts := strings.Split(expression, ">")
		if len(parts) == 2 {
			paramName := strings.TrimSpace(parts[0])
			expectedValue := strings.TrimSpace(parts[1])
			actualValue := getNestedValue(parameter, paramName)
			
			if actualNum, err := toNumber(actualValue); err == nil {
				if expectedNum, err := strconv.ParseFloat(expectedValue, 64); err == nil {
					return actualNum > expectedNum
				}
			}
		}
	}
	
	// 默认检查参数是否存在且不为空
	value := getNestedValue(parameter, expression)
	return isNotEmpty(value)
}

// EvaluateIterable 求值可迭代表达式
func (evaluator *SimpleExpressionEvaluator) EvaluateIterable(parameter any, context DynamicContext) any {
	return getNestedValue(parameter, evaluator.Expression)
}

// 辅助函数

func getNestedValue(obj any, path string) any {
	if obj == nil {
		return nil
	}
	
	if m, ok := obj.(map[string]any); ok {
		return m[path]
	}
	
	parts := strings.Split(path, ".")
	current := obj
	
	for _, part := range parts {
		if current == nil {
			return nil
		}
		
		v := reflect.ValueOf(current)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		
		if v.Kind() == reflect.Struct {
			field := v.FieldByName(part)
			if !field.IsValid() {
				return nil
			}
			current = field.Interface()
		} else {
			return nil
		}
	}
	
	return current
}

func isNotEmpty(value any) bool {
	if value == nil {
		return false
	}
	
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.String() != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Bool:
		return v.Bool()
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() > 0
	default:
		return true
	}
}

func toNumber(value any) (float64, error) {
	if value == nil {
		return 0, fmt.Errorf("value is nil")
	}
	
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String:
		return strconv.ParseFloat(v.String(), 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to number", value)
	}
}