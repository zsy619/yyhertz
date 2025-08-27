// Package mapper 动态SQL构建器 - 重构版本
//
// 完全重写的动态SQL引擎，支持：
// 1. 基于XML解析器的标签解析（替代正则表达式）
// 2. 强大的表达式求值引擎
// 3. 完整的MyBatis动态SQL标签支持
// 4. 参数映射和类型转换
// 5. SQL缓存和性能优化
package mapper

import (
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

// ================================
// 核心接口定义
// ================================

// DynamicSqlBuilder 动态SQL构建器 - 新版本
type DynamicSqlBuilder struct {
	exprEvaluator ExpressionEvaluator
	nodeParser    SqlNodeParser
	cache         map[string]*ParsedSqlNode // SQL解析结果缓存
}

// SqlNode SQL节点接口
type SqlNode interface {
	Apply(context *DynamicContext) error
}

// ExpressionEvaluator 表达式求值器接口
type ExpressionEvaluator interface {
	EvaluateBoolean(expression string, parameters any) bool
	EvaluateObject(expression string, parameters any) any
	EvaluateIterable(expression string, parameters any) []any
}

// SqlNodeParser SQL节点解析器接口
type SqlNodeParser interface {
	Parse(xml string) (SqlNode, error)
}

// DynamicContext 动态上下文
type DynamicContext struct {
	Parameters   any                    // 输入参数
	ParamMap     map[string]any        // 参数映射
	SqlBuilder   *strings.Builder      // SQL构建器
	ArgsList     []any                 // 参数列表
	UniqueNumber int                   // 唯一编号生成器
	Variables    map[string]any        // 局部变量（用于bind标签）
}

// ParsedSqlNode 解析后的SQL节点（用于缓存）
type ParsedSqlNode struct {
	RootNode SqlNode
	CreateAt time.Time
}

// ================================
// 动态SQL构建器实现
// ================================

// NewDynamicSqlBuilder 创建新的动态SQL构建器
func NewDynamicSqlBuilder() *DynamicSqlBuilder {
	return &DynamicSqlBuilder{
		exprEvaluator: NewOgnlExpressionEvaluator(),
		nodeParser:    NewXmlSqlNodeParser(),
		cache:         make(map[string]*ParsedSqlNode),
	}
}

// Build 构建动态SQL
func (builder *DynamicSqlBuilder) Build(sqlTemplate string, parameter any) (string, []any, error) {
	// 检查缓存
	cacheKey := sqlTemplate
	if cachedNode, exists := builder.cache[cacheKey]; exists {
		return builder.applyNode(cachedNode.RootNode, parameter)
	}

	// 解析SQL模板
	rootNode, err := builder.nodeParser.Parse(sqlTemplate)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse SQL template: %w", err)
	}

	// 缓存解析结果
	builder.cache[cacheKey] = &ParsedSqlNode{
		RootNode: rootNode,
		CreateAt: time.Now(),
	}

	// 应用节点生成SQL
	return builder.applyNode(rootNode, parameter)
}

// applyNode 应用SQL节点生成最终SQL
func (builder *DynamicSqlBuilder) applyNode(node SqlNode, parameter any) (string, []any, error) {
	context := &DynamicContext{
		Parameters:   parameter,
		ParamMap:     builder.buildParameterMap(parameter),
		SqlBuilder:   &strings.Builder{},
		ArgsList:     make([]any, 0),
		UniqueNumber: 0,
		Variables:    make(map[string]any),
	}

	err := node.Apply(context)
	if err != nil {
		return "", nil, err
	}

	return context.SqlBuilder.String(), context.ArgsList, nil
}

// buildParameterMap 构建参数映射
func (builder *DynamicSqlBuilder) buildParameterMap(parameter any) map[string]any {
	paramMap := make(map[string]any)

	if parameter == nil {
		return paramMap
	}

	// 如果已经是map，直接返回
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
			if field.IsExported() && v.Field(i).CanInterface() {
				// 支持json标签
				fieldName := field.Name
				if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
					fieldName = strings.Split(jsonTag, ",")[0]
				}
				paramMap[fieldName] = v.Field(i).Interface()
				paramMap[field.Name] = v.Field(i).Interface() // 同时支持原始字段名
			}
		}
	} else {
		// 基本类型或其他类型
		paramMap["value"] = parameter
		paramMap["_parameter"] = parameter
	}

	return paramMap
}

// ================================
// XML SQL节点解析器实现
// ================================

// XmlSqlNodeParser XML SQL节点解析器
type XmlSqlNodeParser struct {
	evaluator ExpressionEvaluator
}

// NewXmlSqlNodeParser 创建XML SQL节点解析器
func NewXmlSqlNodeParser() *XmlSqlNodeParser {
	return &XmlSqlNodeParser{
		evaluator: NewOgnlExpressionEvaluator(),
	}
}

// Parse 解析SQL模板
func (parser *XmlSqlNodeParser) Parse(sqlTemplate string) (SqlNode, error) {
	// 包装为根节点以便XML解析
	wrappedXml := fmt.Sprintf("<root>%s</root>", sqlTemplate)

	decoder := xml.NewDecoder(strings.NewReader(wrappedXml))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		return input, nil // 简化字符集处理
	}

	return parser.parseNode(decoder)
}

// parseNode 解析XML节点
func (parser *XmlSqlNodeParser) parseNode(decoder *xml.Decoder) (SqlNode, error) {
	nodes := make([]SqlNode, 0)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("XML parsing error: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			node, err := parser.parseStartElement(t, decoder)
			if err != nil {
				return nil, err
			}
			if node != nil {
				nodes = append(nodes, node)
			}

		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" {
				nodes = append(nodes, &StaticTextSqlNode{Text: text})
			}

		case xml.EndElement:
			// 处理结束标签
			break
		}
	}

	if len(nodes) == 0 {
		return &StaticTextSqlNode{Text: ""}, nil
	} else if len(nodes) == 1 {
		return nodes[0], nil
	} else {
		return &MixedSqlNode{Contents: nodes}, nil
	}
}

// parseStartElement 解析开始标签
func (parser *XmlSqlNodeParser) parseStartElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	switch element.Name.Local {
	case "root":
		return parser.parseNode(decoder)
	case "if":
		return parser.parseIfElement(element, decoder)
	case "choose":
		return parser.parseChooseElement(element, decoder)
	case "when":
		return parser.parseWhenElement(element, decoder)
	case "otherwise":
		return parser.parseOtherwiseElement(element, decoder)
	case "where":
		return parser.parseWhereElement(element, decoder)
	case "set":
		return parser.parseSetElement(element, decoder)
	case "foreach":
		return parser.parseForeachElement(element, decoder)
	case "trim":
		return parser.parseTrimElement(element, decoder)
	case "bind":
		return parser.parseBindElement(element, decoder)
	default:
		// 跳过未知标签，但解析其内容
		return parser.parseNode(decoder)
	}
}

// parseIfElement 解析if标签
func (parser *XmlSqlNodeParser) parseIfElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	// 获取test属性
	var testExpr string
	for _, attr := range element.Attr {
		if attr.Name.Local == "test" {
			testExpr = attr.Value
			break
		}
	}

	if testExpr == "" {
		return nil, fmt.Errorf("if element requires 'test' attribute")
	}

	// 解析内容
	content, err := parser.parseNode(decoder)
	if err != nil {
		return nil, err
	}

	return &IfSqlNode{
		Test:     testExpr,
		Contents: content,
		evaluator: parser.evaluator,
	}, nil
}

// parseChooseElement 解析choose标签
func (parser *XmlSqlNodeParser) parseChooseElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	whenNodes := make([]*WhenSqlNode, 0)
	var otherwiseNode SqlNode

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "when":
				node, err := parser.parseWhenElement(t, decoder)
				if err != nil {
					return nil, err
				}
				if whenNode, ok := node.(*WhenSqlNode); ok {
					whenNodes = append(whenNodes, whenNode)
				}
			case "otherwise":
				otherwiseNode, err = parser.parseOtherwiseElement(t, decoder)
				if err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "choose" {
				return &ChooseSqlNode{
					WhenNodes:     whenNodes,
					OtherwiseNode: otherwiseNode,
				}, nil
			}
		}
	}
}

// parseWhenElement 解析when标签
func (parser *XmlSqlNodeParser) parseWhenElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	var testExpr string
	for _, attr := range element.Attr {
		if attr.Name.Local == "test" {
			testExpr = attr.Value
			break
		}
	}

	content, err := parser.parseNode(decoder)
	if err != nil {
		return nil, err
	}

	return &WhenSqlNode{
		Test:      testExpr,
		Contents:  content,
		evaluator: parser.evaluator,
	}, nil
}

// parseOtherwiseElement 解析otherwise标签
func (parser *XmlSqlNodeParser) parseOtherwiseElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	return parser.parseNode(decoder)
}

// parseWhereElement 解析where标签
func (parser *XmlSqlNodeParser) parseWhereElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	content, err := parser.parseNode(decoder)
	if err != nil {
		return nil, err
	}

	return &WhereSqlNode{Contents: content}, nil
}

// parseSetElement 解析set标签
func (parser *XmlSqlNodeParser) parseSetElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	content, err := parser.parseNode(decoder)
	if err != nil {
		return nil, err
	}

	return &SetSqlNode{Contents: content}, nil
}

// parseForeachElement 解析foreach标签
func (parser *XmlSqlNodeParser) parseForeachElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	var collection, item, index, open, separator, close string

	for _, attr := range element.Attr {
		switch attr.Name.Local {
		case "collection":
			collection = attr.Value
		case "item":
			item = attr.Value
		case "index":
			index = attr.Value
		case "open":
			open = attr.Value
		case "separator":
			separator = attr.Value
		case "close":
			close = attr.Value
		}
	}

	content, err := parser.parseNode(decoder)
	if err != nil {
		return nil, err
	}

	return &ForeachSqlNode{
		Collection: collection,
		Item:       item,
		Index:      index,
		Open:       open,
		Separator:  separator,
		Close:      close,
		Contents:   content,
		evaluator:  parser.evaluator,
	}, nil
}

// parseTrimElement 解析trim标签
func (parser *XmlSqlNodeParser) parseTrimElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	var prefix, suffix, prefixOverrides, suffixOverrides string

	for _, attr := range element.Attr {
		switch attr.Name.Local {
		case "prefix":
			prefix = attr.Value
		case "suffix":
			suffix = attr.Value
		case "prefixOverrides":
			prefixOverrides = attr.Value
		case "suffixOverrides":
			suffixOverrides = attr.Value
		}
	}

	content, err := parser.parseNode(decoder)
	if err != nil {
		return nil, err
	}

	return &TrimSqlNode{
		Prefix:           prefix,
		Suffix:           suffix,
		PrefixOverrides:  strings.Split(prefixOverrides, "|"),
		SuffixOverrides:  strings.Split(suffixOverrides, "|"),
		Contents:         content,
	}, nil
}

// parseBindElement 解析bind标签
func (parser *XmlSqlNodeParser) parseBindElement(element xml.StartElement, decoder *xml.Decoder) (SqlNode, error) {
	var name, value string

	for _, attr := range element.Attr {
		switch attr.Name.Local {
		case "name":
			name = attr.Value
		case "value":
			value = attr.Value
		}
	}

	return &BindSqlNode{
		Name:      name,
		Value:     value,
		evaluator: parser.evaluator,
	}, nil
}

// ================================
// SQL节点实现
// ================================

// StaticTextSqlNode 静态文本节点
type StaticTextSqlNode struct {
	Text string
}

func (node *StaticTextSqlNode) Apply(context *DynamicContext) error {
	// 首先处理字符串替换 ${param}
	processedText := node.processStringSubstitution(node.Text, context)
	
	// 然后处理参数占位符 #{param}
	finalText, args := node.processParameters(processedText, context)
	
	context.SqlBuilder.WriteString(finalText)
	context.ArgsList = append(context.ArgsList, args...)
	return nil
}

// processStringSubstitution 处理字符串替换 ${param}
func (node *StaticTextSqlNode) processStringSubstitution(text string, context *DynamicContext) string {
	result := text
	
	// 查找所有${param}占位符
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end = start + end

		paramExpr := result[start+2 : end]
		paramValue := node.getParameterValue(paramExpr, context)
		
		// 安全检查：记录字符串替换使用情况
		node.logStringSubstitutionWarning(paramExpr, paramValue)
		
		// 直接字符串替换
		stringValue := node.convertToString(paramValue)
		result = result[:start] + stringValue + result[end+1:]
	}

	return result
}

// logStringSubstitutionWarning 记录字符串替换警告
func (node *StaticTextSqlNode) logStringSubstitutionWarning(paramName string, value any) {
	// TODO: 使用更好的日志系统
	fmt.Printf("[SECURITY WARNING] Using string substitution ${%s} with value: %v. "+
		"This may cause SQL injection. Consider using #{%s} instead.\n", 
		paramName, value, paramName)
}

// convertToString 转换为字符串
func (node *StaticTextSqlNode) convertToString(value any) string {
	if value == nil {
		return "NULL"
	}
	
	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (node *StaticTextSqlNode) processParameters(text string, context *DynamicContext) (string, []any) {
	// 简化实现：查找#{param}并替换为?
	args := make([]any, 0)
	result := text

	// 使用简单的字符串处理替代复杂正则
	for {
		start := strings.Index(result, "#{")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}
		end = start + end

		paramExpr := result[start+2 : end]
		paramValue := node.getParameterValue(paramExpr, context)
		args = append(args, paramValue)

		result = result[:start] + "?" + result[end+1:]
	}

	return result, args
}

func (node *StaticTextSqlNode) getParameterValue(expression string, context *DynamicContext) any {
	// 首先检查局部变量
	if value, exists := context.Variables[expression]; exists {
		return value
	}

	// 然后检查参数映射
	if value, exists := context.ParamMap[expression]; exists {
		return value
	}

	// 最后尝试复杂表达式解析
	evaluator := NewOgnlExpressionEvaluator()
	return evaluator.EvaluateObject(expression, context.Parameters)
}

// MixedSqlNode 混合节点
type MixedSqlNode struct {
	Contents []SqlNode
}

func (node *MixedSqlNode) Apply(context *DynamicContext) error {
	for _, content := range node.Contents {
		if err := content.Apply(context); err != nil {
			return err
		}
	}
	return nil
}

// IfSqlNode IF节点
type IfSqlNode struct {
	Test      string
	Contents  SqlNode
	evaluator ExpressionEvaluator
}

func (node *IfSqlNode) Apply(context *DynamicContext) error {
	if node.evaluator.EvaluateBoolean(node.Test, context.Parameters) {
		return node.Contents.Apply(context)
	}
	return nil
}

// ChooseSqlNode CHOOSE节点
type ChooseSqlNode struct {
	WhenNodes     []*WhenSqlNode
	OtherwiseNode SqlNode
}

func (node *ChooseSqlNode) Apply(context *DynamicContext) error {
	for _, whenNode := range node.WhenNodes {
		if whenNode.evaluator.EvaluateBoolean(whenNode.Test, context.Parameters) {
			return whenNode.Contents.Apply(context)
		}
	}

	if node.OtherwiseNode != nil {
		return node.OtherwiseNode.Apply(context)
	}

	return nil
}

// WhenSqlNode WHEN节点
type WhenSqlNode struct {
	Test      string
	Contents  SqlNode
	evaluator ExpressionEvaluator
}

func (node *WhenSqlNode) Apply(context *DynamicContext) error {
	if node.evaluator.EvaluateBoolean(node.Test, context.Parameters) {
		return node.Contents.Apply(context)
	}
	return nil
}

// WhereSqlNode WHERE节点
type WhereSqlNode struct {
	Contents SqlNode
}

func (node *WhereSqlNode) Apply(context *DynamicContext) error {
	oldLength := context.SqlBuilder.Len()
	
	err := node.Contents.Apply(context)
	if err != nil {
		return err
	}

	newLength := context.SqlBuilder.Len()
	if newLength > oldLength {
		// 有内容被添加
		addedContent := context.SqlBuilder.String()[oldLength:]
		
		// 移除开头的AND或OR
		trimmed := strings.TrimSpace(addedContent)
		if strings.HasPrefix(strings.ToUpper(trimmed), "AND ") {
			trimmed = strings.TrimSpace(trimmed[4:])
		} else if strings.HasPrefix(strings.ToUpper(trimmed), "OR ") {
			trimmed = strings.TrimSpace(trimmed[3:])
		}

		if trimmed != "" {
			// 重新构建SQL
			existingContent := context.SqlBuilder.String()[:oldLength]
			context.SqlBuilder.Reset()
			context.SqlBuilder.WriteString(existingContent)
			context.SqlBuilder.WriteString("WHERE ")
			context.SqlBuilder.WriteString(trimmed)
		}
	}

	return nil
}

// SetSqlNode SET节点
type SetSqlNode struct {
	Contents SqlNode
}

func (node *SetSqlNode) Apply(context *DynamicContext) error {
	oldLength := context.SqlBuilder.Len()
	
	err := node.Contents.Apply(context)
	if err != nil {
		return err
	}

	newLength := context.SqlBuilder.Len()
	if newLength > oldLength {
		addedContent := context.SqlBuilder.String()[oldLength:]
		
		// 移除末尾的逗号
		trimmed := strings.TrimSpace(addedContent)
		trimmed = strings.TrimSuffix(trimmed, ",")

		if trimmed != "" {
			existingContent := context.SqlBuilder.String()[:oldLength]
			context.SqlBuilder.Reset()
			context.SqlBuilder.WriteString(existingContent)
			context.SqlBuilder.WriteString("SET ")
			context.SqlBuilder.WriteString(trimmed)
		}
	}

	return nil
}

// ForeachSqlNode FOREACH节点
type ForeachSqlNode struct {
	Collection string
	Item       string
	Index      string
	Open       string
	Separator  string
	Close      string
	Contents   SqlNode
	evaluator  ExpressionEvaluator
}

func (node *ForeachSqlNode) Apply(context *DynamicContext) error {
	items := node.evaluator.EvaluateIterable(node.Collection, context.Parameters)
	if len(items) == 0 {
		return nil
	}

	// 写入开始标记
	context.SqlBuilder.WriteString(node.Open)

	// 遍历集合
	for i, item := range items {
		if i > 0 && node.Separator != "" {
			context.SqlBuilder.WriteString(node.Separator)
		}

		// 设置循环变量
		oldItem := context.Variables[node.Item]
		oldIndex := context.Variables[node.Index]
		
		context.Variables[node.Item] = item
		if node.Index != "" {
			context.Variables[node.Index] = i
		}

		// 应用内容
		err := node.Contents.Apply(context)
		
		// 恢复变量
		if oldItem != nil {
			context.Variables[node.Item] = oldItem
		} else {
			delete(context.Variables, node.Item)
		}
		if oldIndex != nil && node.Index != "" {
			context.Variables[node.Index] = oldIndex
		} else if node.Index != "" {
			delete(context.Variables, node.Index)
		}

		if err != nil {
			return err
		}
	}

	// 写入结束标记
	context.SqlBuilder.WriteString(node.Close)

	return nil
}

// TrimSqlNode TRIM节点
type TrimSqlNode struct {
	Prefix          string
	Suffix          string
	PrefixOverrides []string
	SuffixOverrides []string
	Contents        SqlNode
}

func (node *TrimSqlNode) Apply(context *DynamicContext) error {
	oldLength := context.SqlBuilder.Len()
	
	err := node.Contents.Apply(context)
	if err != nil {
		return err
	}

	newLength := context.SqlBuilder.Len()
	if newLength > oldLength {
		addedContent := context.SqlBuilder.String()[oldLength:]
		trimmed := strings.TrimSpace(addedContent)

		// 移除前缀
		for _, prefix := range node.PrefixOverrides {
			if prefix != "" && strings.HasPrefix(strings.ToUpper(trimmed), strings.ToUpper(prefix)) {
				trimmed = strings.TrimSpace(trimmed[len(prefix):])
				break
			}
		}

		// 移除后缀
		for _, suffix := range node.SuffixOverrides {
			if suffix != "" && strings.HasSuffix(strings.ToUpper(trimmed), strings.ToUpper(suffix)) {
				trimmed = strings.TrimSpace(trimmed[:len(trimmed)-len(suffix)])
				break
			}
		}

		if trimmed != "" {
			existingContent := context.SqlBuilder.String()[:oldLength]
			context.SqlBuilder.Reset()
			context.SqlBuilder.WriteString(existingContent)
			context.SqlBuilder.WriteString(node.Prefix)
			context.SqlBuilder.WriteString(trimmed)
			context.SqlBuilder.WriteString(node.Suffix)
		}
	}

	return nil
}

// BindSqlNode BIND节点
type BindSqlNode struct {
	Name      string
	Value     string
	evaluator ExpressionEvaluator
}

func (node *BindSqlNode) Apply(context *DynamicContext) error {
	value := node.evaluator.EvaluateObject(node.Value, context.Parameters)
	context.Variables[node.Name] = value
	return nil
}