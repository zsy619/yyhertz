// Package plugin 分页插件实现
//
// 提供自动分页功能，支持多种数据库的分页语法
package plugin

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// PaginationPlugin 分页插件
type PaginationPlugin struct {
	*BasePlugin
	dialectType    string // 数据库方言类型
	defaultPageSize int   // 默认页面大小
	maxPageSize     int   // 最大页面大小
}

// PageRequest 分页请求
type PageRequest struct {
	PageNum  int `json:"pageNum"`  // 页码，从1开始
	PageSize int `json:"pageSize"` // 每页大小
}

// PageResult 分页结果
type PageResult struct {
	List       []any `json:"list"`       // 数据列表
	Total      int64 `json:"total"`      // 总记录数
	PageNum    int   `json:"pageNum"`    // 当前页码
	PageSize   int   `json:"pageSize"`   // 每页大小
	TotalPages int   `json:"totalPages"` // 总页数
	HasNext    bool  `json:"hasNext"`    // 是否有下一页
	HasPrev    bool  `json:"hasPrev"`    // 是否有上一页
}

// NewPaginationPlugin 创建分页插件
func NewPaginationPlugin() *PaginationPlugin {
	plugin := &PaginationPlugin{
		BasePlugin:      NewBasePlugin("pagination", 1),
		dialectType:     "mysql",
		defaultPageSize: 20,
		maxPageSize:     1000,
	}
	return plugin
}

// Intercept 拦截方法调用
func (plugin *PaginationPlugin) Intercept(invocation *Invocation) (any, error) {
	// 检查是否需要分页
	pageRequest := plugin.extractPageRequest(invocation.Args)
	if pageRequest == nil {
		// 不需要分页，直接执行原方法
		return invocation.Proceed()
	}
	
	// 验证分页参数
	if err := plugin.validatePageRequest(pageRequest); err != nil {
		return nil, err
	}
	
	// 执行分页查询
	return plugin.executePageQuery(invocation, pageRequest)
}

// Plugin 包装目标对象
func (plugin *PaginationPlugin) Plugin(target any) any {
	// 这里应该创建代理对象，简化实现
	return target
}

// SetProperties 设置插件属性
func (plugin *PaginationPlugin) SetProperties(properties map[string]any) {
	plugin.BasePlugin.SetProperties(properties)
	
	if dialectType := plugin.GetPropertyString("dialectType", ""); dialectType != "" {
		plugin.dialectType = dialectType
	}
	
	plugin.defaultPageSize = plugin.GetPropertyInt("defaultPageSize", 20)
	plugin.maxPageSize = plugin.GetPropertyInt("maxPageSize", 1000)
}

// extractPageRequest 从参数中提取分页请求
func (plugin *PaginationPlugin) extractPageRequest(args []any) *PageRequest {
	for _, arg := range args {
		if pageReq, ok := arg.(*PageRequest); ok {
			return pageReq
		}
		
		// 检查是否是包含分页信息的map
		if argMap, ok := arg.(map[string]any); ok {
			if pageNum, hasPageNum := argMap["pageNum"]; hasPageNum {
				if pageSize, hasPageSize := argMap["pageSize"]; hasPageSize {
					return &PageRequest{
						PageNum:  plugin.toInt(pageNum, 1),
						PageSize: plugin.toInt(pageSize, plugin.defaultPageSize),
					}
				}
			}
		}
		
		// 使用反射检查结构体字段
		if plugin.hasPageFields(arg) {
			return plugin.extractPageFromStruct(arg)
		}
	}
	
	return nil
}

// hasPageFields 检查是否包含分页字段
func (plugin *PaginationPlugin) hasPageFields(obj any) bool {
	if obj == nil {
		return false
	}
	
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return false
	}
	
	t := v.Type()
	hasPageNum := false
	hasPageSize := false
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := strings.ToLower(field.Name)
		
		if fieldName == "pagenum" || fieldName == "page" {
			hasPageNum = true
		}
		if fieldName == "pagesize" || fieldName == "size" {
			hasPageSize = true
		}
	}
	
	return hasPageNum && hasPageSize
}

// extractPageFromStruct 从结构体中提取分页信息
func (plugin *PaginationPlugin) extractPageFromStruct(obj any) *PageRequest {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return nil
	}
	
	pageReq := &PageRequest{
		PageNum:  1,
		PageSize: plugin.defaultPageSize,
	}
	
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := strings.ToLower(field.Name)
		fieldValue := v.Field(i)
		
		if !fieldValue.CanInterface() {
			continue
		}
		
		switch fieldName {
		case "pagenum", "page":
			pageReq.PageNum = plugin.toInt(fieldValue.Interface(), 1)
		case "pagesize", "size":
			pageReq.PageSize = plugin.toInt(fieldValue.Interface(), plugin.defaultPageSize)
		}
	}
	
	return pageReq
}

// toInt 转换为整数
func (plugin *PaginationPlugin) toInt(value any, defaultValue int) int {
	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float32:
		return int(v)
	case float64:
		return int(v)
	case string:
		// 尝试解析字符串
		var i int
		if n, _ := fmt.Sscanf(v, "%d", &i); n == 1 {
			return i
		}
	}
	return defaultValue
}

// validatePageRequest 验证分页请求
func (plugin *PaginationPlugin) validatePageRequest(pageReq *PageRequest) error {
	if pageReq.PageNum < 1 {
		return fmt.Errorf("页码必须大于0，当前值: %d", pageReq.PageNum)
	}
	
	if pageReq.PageSize < 1 {
		return fmt.Errorf("每页大小必须大于0，当前值: %d", pageReq.PageSize)
	}
	
	if pageReq.PageSize > plugin.maxPageSize {
		return fmt.Errorf("每页大小不能超过%d，当前值: %d", plugin.maxPageSize, pageReq.PageSize)
	}
	
	return nil
}

// executePageQuery 执行分页查询
func (plugin *PaginationPlugin) executePageQuery(invocation *Invocation, pageReq *PageRequest) (any, error) {
	// 1. 先执行count查询获取总数
	total, err := plugin.executeCountQuery(invocation)
	if err != nil {
		return nil, fmt.Errorf("执行count查询失败: %w", err)
	}
	
	// 2. 如果总数为0，直接返回空结果
	if total == 0 {
		return &PageResult{
			List:       make([]any, 0),
			Total:      0,
			PageNum:    pageReq.PageNum,
			PageSize:   pageReq.PageSize,
			TotalPages: 0,
			HasNext:    false,
			HasPrev:    false,
		}, nil
	}
	
	// 3. 修改原SQL添加分页条件
	err = plugin.addPaginationToInvocation(invocation, pageReq)
	if err != nil {
		return nil, fmt.Errorf("添加分页条件失败: %w", err)
	}
	
	// 4. 执行分页查询
	result, err := invocation.Proceed()
	if err != nil {
		return nil, fmt.Errorf("执行分页查询失败: %w", err)
	}
	
	// 5. 包装结果
	return plugin.wrapPageResult(result, total, pageReq), nil
}

// executeCountQuery 执行count查询
func (plugin *PaginationPlugin) executeCountQuery(invocation *Invocation) (int64, error) {
	// 这里需要根据原SQL生成count查询
	// 简化实现，实际应该解析SQL并生成对应的count语句
	
	// 模拟count查询结果
	return 100, nil // 简化实现
}

// addPaginationToInvocation 为调用添加分页条件
func (plugin *PaginationPlugin) addPaginationToInvocation(invocation *Invocation, pageReq *PageRequest) error {
	// 这里需要修改SQL添加LIMIT和OFFSET
	// 简化实现，实际应该解析和修改SQL语句
	
	offset := (pageReq.PageNum - 1) * pageReq.PageSize
	limit := pageReq.PageSize
	
	// 将分页信息添加到调用参数中
	invocation.Properties["offset"] = offset
	invocation.Properties["limit"] = limit
	invocation.Properties["pageNum"] = pageReq.PageNum
	invocation.Properties["pageSize"] = pageReq.PageSize
	
	return nil
}

// wrapPageResult 包装分页结果
func (plugin *PaginationPlugin) wrapPageResult(result any, total int64, pageReq *PageRequest) *PageResult {
	var list []any
	
	// 转换结果为列表
	if resultList, ok := result.([]any); ok {
		list = resultList
	} else if result != nil {
		list = []any{result}
	} else {
		list = make([]any, 0)
	}
	
	totalPages := int((total + int64(pageReq.PageSize) - 1) / int64(pageReq.PageSize))
	
	return &PageResult{
		List:       list,
		Total:      total,
		PageNum:    pageReq.PageNum,
		PageSize:   pageReq.PageSize,
		TotalPages: totalPages,
		HasNext:    pageReq.PageNum < totalPages,
		HasPrev:    pageReq.PageNum > 1,
	}
}

// generateCountSql 生成count查询SQL
func (plugin *PaginationPlugin) generateCountSql(originalSql string) string {
	// 移除ORDER BY子句
	orderByRegex := regexp.MustCompile(`(?i)\s+order\s+by\s+[^)]*$`)
	sql := orderByRegex.ReplaceAllString(originalSql, "")
	
	// 简单的count包装
	return fmt.Sprintf("SELECT COUNT(*) FROM (%s) tmp_count", sql)
}

// generatePageSql 生成分页查询SQL
func (plugin *PaginationPlugin) generatePageSql(originalSql string, offset, limit int) string {
	switch plugin.dialectType {
	case "mysql":
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", originalSql, limit, offset)
	case "postgresql":
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", originalSql, limit, offset)
	case "sqlite":
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", originalSql, limit, offset)
	case "sqlserver":
		return fmt.Sprintf("%s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", originalSql, offset, limit)
	case "oracle":
		return fmt.Sprintf("SELECT * FROM (SELECT ROWNUM rn, t.* FROM (%s) t WHERE ROWNUM <= %d) WHERE rn > %d", 
			originalSql, offset+limit, offset)
	default:
		// 默认使用MySQL语法
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", originalSql, limit, offset)
	}
}

// PageHelper 分页助手工具类
type PageHelper struct {
	plugin *PaginationPlugin
}

// NewPageHelper 创建分页助手
func NewPageHelper() *PageHelper {
	return &PageHelper{
		plugin: NewPaginationPlugin(),
	}
}

// StartPage 开始分页
func (helper *PageHelper) StartPage(pageNum, pageSize int) *PageRequest {
	return &PageRequest{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
}

// StartPageWithTotal 开始分页（包含总数查询）
func (helper *PageHelper) StartPageWithTotal(pageNum, pageSize int, count bool) *PageRequest {
	pageReq := &PageRequest{
		PageNum:  pageNum,
		PageSize: pageSize,
	}
	// 这里可以添加是否查询总数的标记
	return pageReq
}