// Package mybatis XML Mapper会话支持
//
// 扩展SimpleSession以支持XML Mapper文件，提供与MyBatis兼容的使用体验
package mybatis

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/zsy619/yyhertz/framework/mybatis/mapper"
)

// XMLSession XML支持的会话接口
type XMLSession interface {
	// 继承SimpleSession的基础CRUD方法
	SelectOne(ctx context.Context, sql string, args ...any) (any, error)
	SelectList(ctx context.Context, sql string, args ...any) ([]any, error)
	SelectPage(ctx context.Context, sql string, page PageRequest, args ...any) (*PageResult, error)
	Insert(ctx context.Context, sql string, args ...any) (int64, error)
	Update(ctx context.Context, sql string, args ...any) (int64, error)
	Delete(ctx context.Context, sql string, args ...any) (int64, error)

	// 批量操作方法
	BatchInsert(ctx context.Context, sql string, batchArgs [][]any) (int64, error)
	BatchUpdate(ctx context.Context, sql string, batchArgs [][]any) (int64, error)
	BatchDelete(ctx context.Context, sql string, batchArgs [][]any) (int64, error)

	// 存储过程调用方法
	CallStoredProc(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error)
	CallStoredProcWithMultiResults(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error)

	// 配置方法 - 返回XMLSession类型以支持方法链
	Debug(enabled bool) XMLSession
	DryRun(enabled bool) XMLSession
	AddBeforeHook(hook BeforeHook) XMLSession
	AddAfterHook(hook AfterHook) XMLSession

	// 缓存管理方法
	EnableCache(config *CacheConfig) XMLSession
	DisableCache() XMLSession
	ClearCache() XMLSession
	GetCacheStats() map[string]any

	// 懒加载管理方法
	EnableLazyLoading(config *LazyLoadConfiguration) XMLSession
	DisableLazyLoading() XMLSession
	RegisterAssociation(typeName string, mapping *AssociationMapping) XMLSession

	// XML映射器管理
	LoadMapperXML(xmlPath string) error
	LoadMapperXMLFromString(xmlContent string) error
	LoadMapperDirectory(dirPath string) error

	// 通过语句ID执行（MyBatis风格）
	SelectOneByID(ctx context.Context, statementId string, parameter any) (any, error)
	SelectListByID(ctx context.Context, statementId string, parameter any) ([]any, error)
	SelectPageByID(ctx context.Context, statementId string, parameter any, page PageRequest) (*PageResult, error)
	InsertByID(ctx context.Context, statementId string, parameter any) (int64, error)
	UpdateByID(ctx context.Context, statementId string, parameter any) (int64, error)
	DeleteByID(ctx context.Context, statementId string, parameter any) (int64, error)

	// 批量操作（基于语句ID）
	BatchInsertByID(ctx context.Context, statementId string, parameters []any) (int64, error)
	BatchUpdateByID(ctx context.Context, statementId string, parameters []any) (int64, error)
	BatchDeleteByID(ctx context.Context, statementId string, parameters []any) (int64, error)

	// Mapper信息查询
	GetStatement(statementId string) *mapper.XMLMappedStatement
	GetResultMap(resultMapId string) *mapper.XMLResultMap
	GetNamespaces() []string
	GetStatementIds(namespace string) []string
}

// xmlSession XML会话实现
type xmlSession struct {
	SimpleSession
	parsers          map[string]*mapper.MapperXMLParser // namespace -> parser
	dynamicBuilder   *mapper.DynamicSqlBuilder
	resultMapper     *mapper.ResultMapper // 新增的结果映射器
	lazyLoadManager  *LazyLoadManager     // 懒加载管理器
	lazyLoadExecutor *LazyLoadingExecutor // 懒加载执行器
}

// NewXMLSession 创建支持XML的会话
func NewXMLSession(db *gorm.DB) XMLSession {
	simpleSession := NewSimpleSession(db)
	lazyManager := NewLazyLoadManager(nil) // 使用默认配置

	session := &xmlSession{
		SimpleSession:    simpleSession,
		parsers:          make(map[string]*mapper.MapperXMLParser),
		dynamicBuilder:   mapper.NewDynamicSqlBuilder(),
		resultMapper:     mapper.NewResultMapper(), // 初始化结果映射器
		lazyLoadManager:  lazyManager,
		lazyLoadExecutor: NewLazyLoadingExecutor(simpleSession, lazyManager),
	}

	return session
}

// NewXMLSessionWithHooks 创建带钩子的XML会话
func NewXMLSessionWithHooks(db *gorm.DB, enableDebug bool) XMLSession {
	simpleSession := NewSimpleSession(db)
	lazyManager := NewLazyLoadManager(nil) // 使用默认配置

	session := &xmlSession{
		SimpleSession:    simpleSession,
		parsers:          make(map[string]*mapper.MapperXMLParser),
		dynamicBuilder:   mapper.NewDynamicSqlBuilder(),
		resultMapper:     mapper.NewResultMapper(), // 初始化结果映射器
		lazyLoadManager:  lazyManager,
		lazyLoadExecutor: NewLazyLoadingExecutor(simpleSession, lazyManager),
	}

	// 配置调试模式和常用钩子
	if enableDebug {
		session = session.Debug(true).(*xmlSession)
	}

	// 添加常用钩子
	session = session.AddBeforeHook(AuditHook()).(*xmlSession)

	// 添加性能监控钩子（100ms慢查询阈值）
	beforeHook, afterHook := PerformanceHook(100 * time.Millisecond)
	session = session.AddBeforeHook(beforeHook).AddAfterHook(afterHook).(*xmlSession)

	if enableDebug {
		debugBefore, debugAfter := DebugHook()
		session = session.AddBeforeHook(debugBefore).AddAfterHook(debugAfter).(*xmlSession)
	}

	return session
}

// LoadMapperXML 加载XML映射文件
func (xs *xmlSession) LoadMapperXML(xmlPath string) error {
	parser := mapper.NewMapperXMLParser()
	if err := parser.ParseXMLFile(xmlPath); err != nil {
		return fmt.Errorf("failed to load mapper XML %s: %w", xmlPath, err)
	}

	namespace := parser.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("mapper XML file must have a namespace: %s", xmlPath)
	}

	xs.parsers[namespace] = parser
	log.Printf("[XML Mapper] Loaded mapper: %s with %d statements", namespace, len(parser.GetAllStatements()))

	return nil
}

// LoadMapperXMLFromString 从字符串加载XML映射
func (xs *xmlSession) LoadMapperXMLFromString(xmlContent string) error {
	parser := mapper.NewMapperXMLParser()
	if err := parser.ParseXMLReader(strings.NewReader(xmlContent)); err != nil {
		return fmt.Errorf("failed to parse mapper XML content: %w", err)
	}

	namespace := parser.GetNamespace()
	if namespace == "" {
		return fmt.Errorf("mapper XML content must have a namespace")
	}

	xs.parsers[namespace] = parser
	log.Printf("[XML Mapper] Loaded mapper from string: %s with %d statements", namespace, len(parser.GetAllStatements()))

	return nil
}

// LoadMapperDirectory 批量加载目录下的映射文件
func (xs *xmlSession) LoadMapperDirectory(dirPath string) error {
	parsers, err := mapper.LoadMapperDirectory(dirPath)
	if err != nil {
		return fmt.Errorf("failed to load mapper directory %s: %w", dirPath, err)
	}

	for namespace, parser := range parsers {
		xs.parsers[namespace] = parser
		log.Printf("[XML Mapper] Loaded mapper: %s with %d statements", namespace, len(parser.GetAllStatements()))
	}

	log.Printf("[XML Mapper] Total loaded %d mappers from directory: %s", len(parsers), dirPath)
	return nil
}

// SelectOneByID 通过语句ID查询单条记录
func (xs *xmlSession) SelectOneByID(ctx context.Context, statementId string, parameter any) (any, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return nil, fmt.Errorf("statement not found: %s", statementId)
	}

	if stmt.StatementType != mapper.StatementTypeSelect {
		return nil, fmt.Errorf("statement %s is not a SELECT statement", statementId)
	}

	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}

	// 执行查询
	result, err := xs.SimpleSession.SelectOne(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// 如果没有结果，直接返回
	if result == nil {
		return nil, nil
	}

	// 应用ResultMap映射（如果有的话）
	if stmt.ResultMap != "" {
		mapped, err := xs.applyResultMap(result, stmt.ResultMap)
		if err != nil {
			return nil, err
		}
		result = mapped
	}

	// 处理懒加载（创建代理对象）
	if xs.lazyLoadExecutor != nil {
		processedResult, err := xs.lazyLoadExecutor.ProcessResult(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("failed to process lazy loading: %w", err)
		}
		if processedResult != nil {
			return processedResult, nil
		}
	}

	return result, nil
}

// SelectListByID 通过语句ID查询多条记录
func (xs *xmlSession) SelectListByID(ctx context.Context, statementId string, parameter any) ([]any, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return nil, fmt.Errorf("statement not found: %s", statementId)
	}

	if stmt.StatementType != mapper.StatementTypeSelect {
		return nil, fmt.Errorf("statement %s is not a SELECT statement", statementId)
	}

	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}

	// 执行查询
	results, err := xs.SimpleSession.SelectList(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// 应用ResultMap映射（如果有的话）
	if stmt.ResultMap != "" {
		mappedResults := make([]any, len(results))
		for i, result := range results {
			mapped, err := xs.applyResultMap(result, stmt.ResultMap)
			if err != nil {
				return nil, fmt.Errorf("failed to apply result map: %w", err)
			}
			mappedResults[i] = mapped
		}
		results = mappedResults
	}

	// 处理懒加载（创建代理对象）
	if xs.lazyLoadExecutor != nil {
		processedResults, err := xs.lazyLoadExecutor.ProcessResult(ctx, results)
		if err != nil {
			return nil, fmt.Errorf("failed to process lazy loading: %w", err)
		}
		if processedResults != nil {
			if processedSlice, ok := processedResults.([]any); ok {
				return processedSlice, nil
			}
		}
	}

	return results, nil
}

// SelectPageByID 通过语句ID分页查询
func (xs *xmlSession) SelectPageByID(ctx context.Context, statementId string, parameter any, page PageRequest) (*PageResult, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return nil, fmt.Errorf("statement not found: %s", statementId)
	}

	if stmt.StatementType != mapper.StatementTypeSelect {
		return nil, fmt.Errorf("statement %s is not a SELECT statement", statementId)
	}

	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}

	// 执行分页查询
	pageResult, err := xs.SimpleSession.SelectPage(ctx, sql, page, args...)
	if err != nil {
		return nil, err
	}

	// 应用ResultMap映射（如果有的话）
	if stmt.ResultMap != "" {
		mappedItems := make([]any, len(pageResult.Items))
		for i, item := range pageResult.Items {
			mapped, err := xs.applyResultMap(item, stmt.ResultMap)
			if err != nil {
				return nil, fmt.Errorf("failed to apply result map: %w", err)
			}
			mappedItems[i] = mapped
		}
		pageResult.Items = mappedItems
	}

	return pageResult, nil
}

// InsertByID 通过语句ID插入记录
func (xs *xmlSession) InsertByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return 0, fmt.Errorf("statement not found: %s", statementId)
	}

	if stmt.StatementType != mapper.StatementTypeInsert {
		return 0, fmt.Errorf("statement %s is not an INSERT statement", statementId)
	}

	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}

	return xs.SimpleSession.Insert(ctx, sql, args...)
}

// UpdateByID 通过语句ID更新记录
func (xs *xmlSession) UpdateByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return 0, fmt.Errorf("statement not found: %s", statementId)
	}

	if stmt.StatementType != mapper.StatementTypeUpdate {
		return 0, fmt.Errorf("statement %s is not an UPDATE statement", statementId)
	}

	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}

	return xs.SimpleSession.Update(ctx, sql, args...)
}

// DeleteByID 通过语句ID删除记录
func (xs *xmlSession) DeleteByID(ctx context.Context, statementId string, parameter any) (int64, error) {
	stmt := xs.getStatementByID(statementId)
	if stmt == nil {
		return 0, fmt.Errorf("statement not found: %s", statementId)
	}

	if stmt.StatementType != mapper.StatementTypeDelete {
		return 0, fmt.Errorf("statement %s is not a DELETE statement", statementId)
	}

	// 构建最终的SQL
	sql, args, err := xs.buildSQL(stmt, parameter)
	if err != nil {
		return 0, fmt.Errorf("failed to build SQL for %s: %w", statementId, err)
	}

	return xs.SimpleSession.Delete(ctx, sql, args...)
}

// GetStatement 获取语句定义
func (xs *xmlSession) GetStatement(statementId string) *mapper.XMLMappedStatement {
	return xs.getStatementByID(statementId)
}

// GetResultMap 获取ResultMap定义
func (xs *xmlSession) GetResultMap(resultMapId string) *mapper.XMLResultMap {
	return xs.getResultMapByID(resultMapId)
}

// GetNamespaces 获取所有已加载的命名空间
func (xs *xmlSession) GetNamespaces() []string {
	namespaces := make([]string, 0, len(xs.parsers))
	for namespace := range xs.parsers {
		namespaces = append(namespaces, namespace)
	}
	return namespaces
}

// GetStatementIds 获取指定命名空间下的所有语句ID
func (xs *xmlSession) GetStatementIds(namespace string) []string {
	parser, exists := xs.parsers[namespace]
	if !exists {
		return []string{}
	}

	statements := parser.GetAllStatements()
	ids := make([]string, 0, len(statements))
	for statementId := range statements {
		ids = append(ids, statementId)
	}
	return ids
}

// 内部方法

// getStatementByID 通过ID获取语句
func (xs *xmlSession) getStatementByID(statementId string) *mapper.XMLMappedStatement {
	parts := strings.SplitN(statementId, ".", 2)
	if len(parts) != 2 {
		return nil
	}

	namespace := parts[0]
	parser, exists := xs.parsers[namespace]
	if !exists {
		return nil
	}

	return parser.GetStatement(statementId)
}

// getResultMapByID 通过ID获取ResultMap
func (xs *xmlSession) getResultMapByID(resultMapId string) *mapper.XMLResultMap {
	parts := strings.SplitN(resultMapId, ".", 2)
	if len(parts) != 2 {
		return nil
	}

	namespace := parts[0]
	parser, exists := xs.parsers[namespace]
	if !exists {
		return nil
	}

	return parser.GetResultMap(resultMapId)
}

// buildSQL 构建最终的SQL语句
func (xs *xmlSession) buildSQL(stmt *mapper.XMLMappedStatement, parameter any) (string, []any, error) {
	sql := stmt.SQL

	// 检查是否包含动态SQL
	if xs.containsDynamicSQL(sql) {
		// 使用动态SQL构建器
		builtSQL, args, err := xs.dynamicBuilder.Build(sql, parameter)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build dynamic SQL: %w", err)
		}
		return builtSQL, args, nil
	} else {
		// 静态SQL，直接处理参数占位符
		builtSQL, args, err := xs.processStaticSQL(sql, parameter)
		if err != nil {
			return "", nil, fmt.Errorf("failed to process static SQL: %w", err)
		}
		return builtSQL, args, nil
	}
}

// containsDynamicSQL 检查是否包含动态SQL标签
func (xs *xmlSession) containsDynamicSQL(sql string) bool {
	dynamicTags := []string{"<if", "<where", "<set", "<choose", "<foreach", "<trim", "<bind"}
	for _, tag := range dynamicTags {
		if strings.Contains(sql, tag) {
			return true
		}
	}
	return false
}

// processStaticSQL 处理静态SQL的参数占位符
func (xs *xmlSession) processStaticSQL(sql string, parameter any) (string, []any, error) {
	args := make([]any, 0)

	// 简单的#{param}占位符替换
	result := sql

	// 查找所有#{xxx}占位符
	paramPattern := `#\{([^}]+)\}`
	matches := regexp.MustCompile(paramPattern).FindAllStringSubmatch(sql, -1)

	for _, match := range matches {
		if len(match) > 1 {
			paramName := match[1]
			value := xs.getParameterValue(parameter, paramName)
			args = append(args, value)
		}
	}

	// 将所有#{xxx}替换为?
	result = regexp.MustCompile(paramPattern).ReplaceAllString(result, "?")

	return result, args, nil
}

// getParameterValue 从参数对象中获取指定属性的值
func (xs *xmlSession) getParameterValue(parameter any, propertyPath string) any {
	if parameter == nil {
		return nil
	}

	// 如果是map类型
	if paramMap, ok := parameter.(map[string]any); ok {
		return paramMap[propertyPath]
	}

	// 使用反射获取结构体字段值
	return xs.getFieldValue(parameter, propertyPath)
}

// getFieldValue 使用反射获取字段值
func (xs *xmlSession) getFieldValue(obj any, fieldPath string) any {
	if obj == nil {
		return nil
	}

	parts := strings.Split(fieldPath, ".")
	current := obj

	for _, part := range parts {
		if current == nil {
			return nil
		}

		v := reflect.ValueOf(current)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return nil
		}

		field := v.FieldByName(part)
		if !field.IsValid() || !field.CanInterface() {
			return nil
		}

		current = field.Interface()
	}

	return current
}

// applyResultMap 应用ResultMap映射
func (xs *xmlSession) applyResultMap(result any, resultMapId string) (any, error) {
	resultMap := xs.getResultMapByID(resultMapId)
	if resultMap == nil {
		return result, nil // 如果没有找到ResultMap，直接返回原结果
	}

	// 使用新的结果映射器处理单个结果
	if resultData, ok := result.(map[string]any); ok {
		return xs.resultMapper.MapResult(resultData, resultMap)
	}

	// 如果不是map格式的结果，保持原有逻辑
	return result, nil
}

// ConvertValue 根据TargetType转换值
func (xs *xmlSession) ConvertValue(value any, targetType string) any {
	if targetType == "" || value == nil {
		return value
	}

	switch strings.ToLower(targetType) {
	case "string", "go.string":
		return fmt.Sprintf("%v", value)
	case "int", "integer", "go.int":
		if str, ok := value.(string); ok {
			if i, err := strconv.Atoi(str); err == nil {
				return i
			}
		}
		return value
	case "long", "go.long":
		if str, ok := value.(string); ok {
			if l, err := strconv.ParseInt(str, 10, 64); err == nil {
				return l
			}
		}
		return value
	case "double", "go.double":
		if str, ok := value.(string); ok {
			if d, err := strconv.ParseFloat(str, 64); err == nil {
				return d
			}
		}
		return value
	case "boolean", "go.bool":
		if str, ok := value.(string); ok {
			if b, err := strconv.ParseBool(str); err == nil {
				return b
			}
		}
		return value
	case "date", "time.time":
		if str, ok := value.(string); ok {
			// 尝试解析常见的日期格式
			formats := []string{
				"2006-01-02 15:04:05",
				"2006-01-02",
				"15:04:05",
			}
			for _, format := range formats {
				if t, err := time.Parse(format, str); err == nil {
					return t
				}
			}
		}
		return value
	default:
		return value
	}
}

// XMLSession配置方法实现 - 返回XMLSession类型以支持方法链

// Debug 设置调试模式
func (xs *xmlSession) Debug(enabled bool) XMLSession {
	xs.SimpleSession = xs.SimpleSession.Debug(enabled)
	return xs
}

// DryRun 设置DryRun模式
func (xs *xmlSession) DryRun(enabled bool) XMLSession {
	xs.SimpleSession = xs.SimpleSession.DryRun(enabled)
	return xs
}

// AddBeforeHook 添加执行前钩子
func (xs *xmlSession) AddBeforeHook(hook BeforeHook) XMLSession {
	xs.SimpleSession = xs.SimpleSession.AddBeforeHook(hook)
	return xs
}

// AddAfterHook 添加执行后钩子
func (xs *xmlSession) AddAfterHook(hook AfterHook) XMLSession {
	xs.SimpleSession = xs.SimpleSession.AddAfterHook(hook)
	return xs
}

// 批量操作方法实现 - 委托给SimpleSession

// BatchInsert 批量插入记录
func (xs *xmlSession) BatchInsert(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return xs.SimpleSession.BatchInsert(ctx, sql, batchArgs)
}

// BatchUpdate 批量更新记录
func (xs *xmlSession) BatchUpdate(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return xs.SimpleSession.BatchUpdate(ctx, sql, batchArgs)
}

// BatchDelete 批量删除记录
func (xs *xmlSession) BatchDelete(ctx context.Context, sql string, batchArgs [][]any) (int64, error) {
	return xs.SimpleSession.BatchDelete(ctx, sql, batchArgs)
}

// 基于ID的批量操作方法实现

// BatchInsertByID 基于语句ID批量插入记录
func (xs *xmlSession) BatchInsertByID(ctx context.Context, statementId string, parameters []any) (int64, error) {
	return xs.executeBatchByID(ctx, statementId, parameters, "INSERT")
}

// BatchUpdateByID 基于语句ID批量更新记录
func (xs *xmlSession) BatchUpdateByID(ctx context.Context, statementId string, parameters []any) (int64, error) {
	return xs.executeBatchByID(ctx, statementId, parameters, "UPDATE")
}

// BatchDeleteByID 基于语句ID批量删除记录
func (xs *xmlSession) BatchDeleteByID(ctx context.Context, statementId string, parameters []any) (int64, error) {
	return xs.executeBatchByID(ctx, statementId, parameters, "DELETE")
}

// executeBatchByID 执行基于语句ID的批量操作
func (xs *xmlSession) executeBatchByID(ctx context.Context, statementId string, parameters []any, operationType string) (int64, error) {
	if len(parameters) == 0 {
		return 0, nil
	}

	// 获取语句配置
	statement := xs.GetStatement(statementId)
	if statement == nil {
		return 0, fmt.Errorf("statement not found: %s", statementId)
	}

	// 验证操作类型
	if !strings.EqualFold(statement.StatementType.String(), operationType) {
		return 0, fmt.Errorf("statement %s is of type %s, but expected %s", statementId, statement.StatementType.String(), operationType)
	}

	var totalAffectedRows int64
	var err error

	// 为每个参数构建SQL并收集批量参数
	batchArgs := make([][]any, 0, len(parameters))
	var finalSQL string

	for i, param := range parameters {
		// 构建动态SQL
		sql, args, buildErr := xs.dynamicBuilder.Build(statement.SQL, param)
		if buildErr != nil {
			return 0, fmt.Errorf("failed to build SQL for parameter %d: %w", i, buildErr)
		}

		// 第一次构建时保存SQL模板
		if i == 0 {
			finalSQL = sql
		} else if sql != finalSQL {
			// 如果SQL不一致，则不能使用批量操作，需要逐个执行
			log.Printf("Warning: SQL inconsistency detected for %s at parameter %d, falling back to individual execution", statementId, i)
			return xs.executeBatchByIDIndividually(ctx, statement, parameters, operationType)
		}

		batchArgs = append(batchArgs, args)
	}

	// 使用批量操作
	switch operationType {
	case "INSERT":
		totalAffectedRows, err = xs.BatchInsert(ctx, finalSQL, batchArgs)
	case "UPDATE":
		totalAffectedRows, err = xs.BatchUpdate(ctx, finalSQL, batchArgs)
	case "DELETE":
		totalAffectedRows, err = xs.BatchDelete(ctx, finalSQL, batchArgs)
	default:
		return 0, fmt.Errorf("unsupported operation type: %s", operationType)
	}

	return totalAffectedRows, err
}

// executeBatchByIDIndividually 当SQL不一致时逐个执行操作
func (xs *xmlSession) executeBatchByIDIndividually(ctx context.Context, statement *mapper.XMLMappedStatement, parameters []any, operationType string) (int64, error) {
	var totalAffectedRows int64

	for i, param := range parameters {
		var affectedRows int64
		var err error

		switch operationType {
		case "INSERT":
			affectedRows, err = xs.InsertByID(ctx, statement.ID, param)
		case "UPDATE":
			affectedRows, err = xs.UpdateByID(ctx, statement.ID, param)
		case "DELETE":
			affectedRows, err = xs.DeleteByID(ctx, statement.ID, param)
		}

		if err != nil {
			return totalAffectedRows, fmt.Errorf("individual operation failed at parameter %d: %w", i, err)
		}

		totalAffectedRows += affectedRows
	}

	return totalAffectedRows, nil
}

// 存储过程调用方法实现 - 委托给SimpleSession

// CallStoredProc 调用存储过程（单结果集）
func (xs *xmlSession) CallStoredProc(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error) {
	return xs.SimpleSession.CallStoredProc(ctx, procName, params)
}

// CallStoredProcWithMultiResults 调用存储过程（多结果集）
func (xs *xmlSession) CallStoredProcWithMultiResults(ctx context.Context, procName string, params []ProcParam) (*StoredProcResult, error) {
	return xs.SimpleSession.CallStoredProcWithMultiResults(ctx, procName, params)
}

// 缓存管理方法实现 - 委托给SimpleSession

// EnableCache 启用缓存
func (xs *xmlSession) EnableCache(config *CacheConfig) XMLSession {
	xs.SimpleSession = xs.SimpleSession.EnableCache(config)
	return xs
}

// DisableCache 禁用缓存
func (xs *xmlSession) DisableCache() XMLSession {
	xs.SimpleSession = xs.SimpleSession.DisableCache()
	return xs
}

// ClearCache 清空缓存
func (xs *xmlSession) ClearCache() XMLSession {
	xs.SimpleSession = xs.SimpleSession.ClearCache()
	return xs
}

// GetCacheStats 获取缓存统计
func (xs *xmlSession) GetCacheStats() map[string]any {
	return xs.SimpleSession.GetCacheStats()
}

// 懒加载管理方法实现

// EnableLazyLoading 启用懒加载
func (xs *xmlSession) EnableLazyLoading(config *LazyLoadConfiguration) XMLSession {
	if config == nil {
		config = DefaultLazyLoadConfiguration()
	}
	xs.lazyLoadManager = NewLazyLoadManager(config)
	xs.lazyLoadExecutor = NewLazyLoadingExecutor(xs.SimpleSession, xs.lazyLoadManager)
	return xs
}

// DisableLazyLoading 禁用懒加载
func (xs *xmlSession) DisableLazyLoading() XMLSession {
	xs.lazyLoadManager = NewLazyLoadManager(&LazyLoadConfiguration{
		LazyLoadingEnabled: false,
	})
	xs.lazyLoadExecutor = NewLazyLoadingExecutor(xs.SimpleSession, xs.lazyLoadManager)
	return xs
}

// RegisterAssociation 注册关联映射
func (xs *xmlSession) RegisterAssociation(typeName string, mapping *AssociationMapping) XMLSession {
	xs.lazyLoadManager.RegisterAssociation(typeName, mapping)
	return xs
}
