// Package config 映射器代理实现
//
// 提供动态代理功能，为映射器接口生成运行时代理对象
package config

import (
	"fmt"
	"log"
	"reflect"
	"sync"
)

// MethodStats 方法统计信息
type MethodStats struct {
	CallCount    int64
	TotalTime    int64
	SuccessCount int64
	ErrorCount   int64
}

// MapperProxy 映射器代理
type MapperProxy struct {
	sqlSession      any
	mapperInterface reflect.Type
	methodCache     map[string]*MapperMethod

	// 统计信息
	statsMutex sync.RWMutex
	stats      map[string]*MethodStats
}

// MapperInvocationHandler 映射器调用处理器
type MapperInvocationHandler struct {
	proxy *MapperProxy
}

// NewMapperProxy 创建映射器代理
func NewMapperProxy(mapperInterface reflect.Type, methodCache map[string]*MapperMethod, sqlSession any) any {
	proxy := &MapperProxy{
		sqlSession:      sqlSession,
		mapperInterface: mapperInterface,
		methodCache:     methodCache,
	}

	// 创建动态代理实例
	return createProxy(mapperInterface, proxy)
}

// createProxy 创建代理实例
func createProxy(mapperInterface reflect.Type, proxy *MapperProxy) any {
	// 创建方法映射
	methodMap := make(map[string]reflect.Value)

	// 为接口的每个方法创建对应的函数实现
	for i := 0; i < mapperInterface.NumMethod(); i++ {
		method := mapperInterface.Method(i)
		methodName := method.Name

		funcImpl := reflect.MakeFunc(method.Type, func(args []reflect.Value) []reflect.Value {
			return proxy.invoke(methodName, args)
		})

		methodMap[methodName] = funcImpl
	}

	// 创建一个实现接口的包装器
	wrapper := &MapperProxyWrapper{
		proxy:           proxy,
		mapperInterface: mapperInterface,
		methodMap:       methodMap,
	}

	return wrapper
}

// MapperProxyWrapper 映射器代理包装器
type MapperProxyWrapper struct {
	proxy           *MapperProxy
	mapperInterface reflect.Type
	methodMap       map[string]reflect.Value
}

// 实现接口方法的通用调用
func (w *MapperProxyWrapper) Call(methodName string, args ...any) ([]any, error) {
	// 转换参数为reflect.Value
	reflectArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		if arg == nil {
			reflectArgs[i] = reflect.Zero(reflect.TypeOf((*any)(nil)).Elem())
		} else {
			reflectArgs[i] = reflect.ValueOf(arg)
		}
	}

	// 调用代理方法
	results := w.proxy.invoke(methodName, reflectArgs)

	// 转换返回值
	returnValues := make([]any, len(results))
	for i, result := range results {
		if result.IsValid() {
			returnValues[i] = result.Interface()
		} else {
			returnValues[i] = nil
		}
	}

	// 检查最后一个返回值是否为error
	if len(returnValues) > 0 {
		if err, ok := returnValues[len(returnValues)-1].(error); ok {
			return returnValues[:len(returnValues)-1], err
		}
	}

	return returnValues, nil
}

// 实现接口的所有方法
func (w *MapperProxyWrapper) ImplementsMethod(methodName string) bool {
	_, exists := w.methodMap[methodName]
	return exists
}

// 获取方法实现
func (w *MapperProxyWrapper) GetMethod(methodName string) (reflect.Value, bool) {
	method, exists := w.methodMap[methodName]
	return method, exists
}

// 添加通用方法调用支持
func (w *MapperProxyWrapper) CallMethod(methodName string, args ...reflect.Value) []reflect.Value {
	if method, exists := w.methodMap[methodName]; exists {
		return method.Call(args)
	}
	return []reflect.Value{reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())}
}

// invoke 调用映射器方法
func (mp *MapperProxy) invoke(methodName string, args []reflect.Value) []reflect.Value {
	// 检查方法名是否为空
	if methodName == "" {
		log.Printf("Empty method name provided")
		return []reflect.Value{reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())}
	}

	// 获取或创建映射器方法
	mapperMethod := mp.cachedMapperMethod(methodName)
	if mapperMethod == nil {
		log.Printf("Failed to create mapper method for: %s", methodName)
		return []reflect.Value{reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())}
	}

	// 记录方法调用
	log.Printf("Invoking mapper method: %s", methodName)

	// 执行映射器方法
	results := mapperMethod.execute(mp.sqlSession, args)

	// 检查执行结果
	if len(results) == 0 {
		log.Printf("Mapper method %s returned empty results", methodName)
		return []reflect.Value{reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())}
	}

	// 检查是否有错误返回
	if len(results) > 0 {
		lastResult := results[len(results)-1]
		if lastResult.Type().String() == "error" && !lastResult.IsNil() {
			log.Printf("Mapper method %s returned error: %v", methodName, lastResult.Interface())
		}
	}

	return results
}

// cachedMapperMethod 获取缓存的映射器方法
func (mp *MapperProxy) cachedMapperMethod(methodName string) *MapperMethod {
	mapperMethod, exists := mp.methodCache[methodName]
	if !exists {
		mapperMethod = mp.createMapperMethod(methodName)
		mp.methodCache[methodName] = mapperMethod
	}
	return mapperMethod
}

// createMapperMethod 创建映射器方法
func (mp *MapperProxy) createMapperMethod(methodName string) *MapperMethod {
	// 验证方法签名
	if err := mp.validateMethodSignature(methodName); err != nil {
		log.Printf("Invalid method signature for %s: %v", methodName, err)
		return nil
	}

	// 获取方法信息
	method, exists := mp.mapperInterface.MethodByName(methodName)
	if !exists {
		return nil
	}

	// 创建SQL命令
	sqlCommand := &SqlCommand{
		Name: fmt.Sprintf("%s.%s", mp.mapperInterface.Name(), methodName),
		Type: mp.getSqlCommandType(methodName),
	}

	// 创建方法签名
	methodSignature := &MethodSignature{
		ReturnsMany:     mp.returnsMany(method.Type),
		ReturnsMap:      mp.returnsMap(method.Type),
		ReturnsVoid:     mp.returnsVoid(method.Type),
		ReturnsCursor:   false, // Go中不需要游标
		ReturnsOptional: false, // Go中不需要Optional
	}

	return &MapperMethod{
		Command:         sqlCommand,
		MethodSignature: methodSignature,
	}
}


// getSqlCommandType 获取SQL命令类型
func (mp *MapperProxy) getSqlCommandType(methodName string) SqlCommandType {
	// 根据方法名推断SQL类型
	switch {
	case len(methodName) >= 6 && (methodName[:6] == "insert" || methodName[:6] == "Insert"):
		return SqlCommandTypeInsert
	case len(methodName) >= 6 && (methodName[:6] == "update" || methodName[:6] == "Update"):
		return SqlCommandTypeUpdate
	case len(methodName) >= 6 && (methodName[:6] == "delete" || methodName[:6] == "Delete"):
		return SqlCommandTypeDelete
	case len(methodName) >= 6 && (methodName[:6] == "select" || methodName[:6] == "Select"):
		return SqlCommandTypeSelect
	case len(methodName) >= 4 && (methodName[:4] == "list" || methodName[:4] == "List"):
		return SqlCommandTypeSelect
	case len(methodName) >= 3 && (methodName[:3] == "get" || methodName[:3] == "Get"):
		return SqlCommandTypeSelect
	case len(methodName) >= 4 && (methodName[:4] == "find" || methodName[:4] == "Find"):
		return SqlCommandTypeSelect
	case len(methodName) >= 5 && (methodName[:5] == "query" || methodName[:5] == "Query"):
		return SqlCommandTypeSelect
	case len(methodName) >= 3 && (methodName[:3] == "add" || methodName[:3] == "Add"):
		return SqlCommandTypeInsert
	case len(methodName) >= 3 && (methodName[:3] == "put" || methodName[:3] == "Put"):
		return SqlCommandTypeUpdate
	case len(methodName) >= 6 && (methodName[:6] == "remove" || methodName[:6] == "Remove"):
		return SqlCommandTypeDelete
	default:
		return SqlCommandTypeSelect // 默认为查询
	}
}

// returnsMany 检查是否返回多个结果
func (mp *MapperProxy) returnsMany(methodType reflect.Type) bool {
	if methodType.NumOut() == 0 {
		return false
	}

	returnType := methodType.Out(0)
	return returnType.Kind() == reflect.Slice || returnType.Kind() == reflect.Array
}

// returnsMap 检查是否返回Map
func (mp *MapperProxy) returnsMap(methodType reflect.Type) bool {
	if methodType.NumOut() == 0 {
		return false
	}

	returnType := methodType.Out(0)
	return returnType.Kind() == reflect.Map
}

// returnsVoid 检查是否无返回值
func (mp *MapperProxy) returnsVoid(methodType reflect.Type) bool {
	return methodType.NumOut() == 0 ||
		(methodType.NumOut() == 1 && methodType.Out(0).String() == "error")
}


// getMethodType 获取方法类型
func (mp *MapperProxy) getMethodType(methodName string) reflect.Type {
	method, exists := mp.mapperInterface.MethodByName(methodName)
	if !exists {
		return nil
	}
	return method.Type
}

// validateMethodSignature 验证方法签名是否符合MyBatis规范
func (mp *MapperProxy) validateMethodSignature(methodName string) error {
	methodType := mp.getMethodType(methodName)
	if methodType == nil {
		return fmt.Errorf("method %s not found", methodName)
	}

	// 检查参数数量（最多支持2个参数：参数对象和RowBounds）
	if methodType.NumIn() > 2 {
		return fmt.Errorf("method %s has too many parameters, maximum 2 allowed", methodName)
	}

	// 检查返回值数量（最多支持2个返回值：结果对象和error）
	if methodType.NumOut() > 2 {
		return fmt.Errorf("method %s has too many return values, maximum 2 allowed", methodName)
	}

	// 如果有返回值，最后一个必须是error类型
	if methodType.NumOut() == 2 {
		if methodType.Out(1).String() != "error" {
			return fmt.Errorf("method %s second return value must be error type", methodName)
		}
	}

	return nil
}

// execute 执行映射器方法
func (mm *MapperMethod) execute(sqlSession any, args []reflect.Value) []reflect.Value {
	// 转换参数
	param := mm.convertArgsToSqlCommandParam(args)

	// 根据命令类型执行相应操作
	switch mm.Command.Type {
	case SqlCommandTypeInsert:
		return mm.executeInsert(sqlSession, param)
	case SqlCommandTypeUpdate:
		return mm.executeUpdate(sqlSession, param)
	case SqlCommandTypeDelete:
		return mm.executeDelete(sqlSession, param)
	case SqlCommandTypeSelect:
		return mm.executeSelect(sqlSession, param)
	default:
		return []reflect.Value{reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())}
	}
}

// convertArgsToSqlCommandParam 转换参数为SQL命令参数
func (mm *MapperMethod) convertArgsToSqlCommandParam(args []reflect.Value) any {
	if len(args) == 0 {
		return nil
	}

	if len(args) == 1 {
		// 如果参数是nil，返回nil
		if !args[0].IsValid() || (args[0].Kind() == reflect.Ptr && args[0].IsNil()) {
			return nil
		}
		return args[0].Interface()
	}

	// 多参数情况，转换为map
	paramMap := make(map[string]any)
	for i, arg := range args {
		// 处理nil参数
		if !arg.IsValid() || (arg.Kind() == reflect.Ptr && arg.IsNil()) {
			paramMap[fmt.Sprintf("param%d", i+1)] = nil
		} else {
			paramMap[fmt.Sprintf("param%d", i+1)] = arg.Interface()
		}
	}
	return paramMap
}

// executeInsert 执行插入操作
func (mm *MapperMethod) executeInsert(sqlSession any, param any) []reflect.Value {
	// 使用反射调用SqlSession的Insert方法
	sessionValue := reflect.ValueOf(sqlSession)
	if !sessionValue.IsValid() || sessionValue.Kind() != reflect.Interface && sessionValue.Kind() != reflect.Ptr {
		err := fmt.Errorf("invalid sqlSession type for insert operation")
		if mm.MethodSignature.ReturnsVoid {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf(int64(0))),
			reflect.ValueOf(err),
		}
	}

	// 通过反射调用Insert方法
	insertMethod := sessionValue.MethodByName("Insert")
	if !insertMethod.IsValid() {
		err := fmt.Errorf("sqlSession does not have Insert method")
		if mm.MethodSignature.ReturnsVoid {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf(int64(0))),
			reflect.ValueOf(err),
		}
	}

	// 调用Insert方法
	args := []reflect.Value{
		reflect.ValueOf(mm.Command.Name),
		reflect.ValueOf(param),
	}
	results := insertMethod.Call(args)

	// 记录SQL执行日志  
	getConfigMethod := sessionValue.MethodByName("GetConfiguration")
	if getConfigMethod.IsValid() {
		configResults := getConfigMethod.Call([]reflect.Value{})
		if len(configResults) > 0 && !configResults[0].IsNil() {
			log.Printf("Executing INSERT statement: %s with parameters: %v", mm.Command.Name, param)
		}
	}

	if mm.MethodSignature.ReturnsVoid {
		if len(results) > 1 && !results[1].IsNil() {
			return []reflect.Value{results[1]}
		}
		return []reflect.Value{}
	}

	if len(results) >= 2 {
		return []reflect.Value{results[0], results[1]}
	}
	return []reflect.Value{
		reflect.Zero(reflect.TypeOf(int64(0))),
		reflect.ValueOf(fmt.Errorf("unexpected return values from Insert method")),
	}
}

// executeUpdate 执行更新操作
func (mm *MapperMethod) executeUpdate(sqlSession any, param any) []reflect.Value {
	// 使用反射调用SqlSession的Update方法
	sessionValue := reflect.ValueOf(sqlSession)
	if !sessionValue.IsValid() || sessionValue.Kind() != reflect.Interface && sessionValue.Kind() != reflect.Ptr {
		err := fmt.Errorf("invalid sqlSession type for update operation")
		if mm.MethodSignature.ReturnsVoid {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf(int64(0))),
			reflect.ValueOf(err),
		}
	}

	// 通过反射调用Update方法
	updateMethod := sessionValue.MethodByName("Update")
	if !updateMethod.IsValid() {
		err := fmt.Errorf("sqlSession does not have Update method")
		if mm.MethodSignature.ReturnsVoid {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf(int64(0))),
			reflect.ValueOf(err),
		}
	}

	// 调用Update方法
	args := []reflect.Value{
		reflect.ValueOf(mm.Command.Name),
		reflect.ValueOf(param),
	}
	results := updateMethod.Call(args)

	// 记录SQL执行日志
	getConfigMethod := sessionValue.MethodByName("GetConfiguration")
	if getConfigMethod.IsValid() {
		configResults := getConfigMethod.Call([]reflect.Value{})
		if len(configResults) > 0 && !configResults[0].IsNil() {
			log.Printf("Executing UPDATE statement: %s with parameters: %v", mm.Command.Name, param)
		}
	}

	if mm.MethodSignature.ReturnsVoid {
		if len(results) > 1 && !results[1].IsNil() {
			return []reflect.Value{results[1]}
		}
		return []reflect.Value{}
	}

	if len(results) >= 2 {
		return []reflect.Value{results[0], results[1]}
	}
	return []reflect.Value{
		reflect.Zero(reflect.TypeOf(int64(0))),
		reflect.ValueOf(fmt.Errorf("unexpected return values from Update method")),
	}
}

// executeDelete 执行删除操作
func (mm *MapperMethod) executeDelete(sqlSession any, param any) []reflect.Value {
	// 使用反射调用SqlSession的Delete方法
	sessionValue := reflect.ValueOf(sqlSession)
	if !sessionValue.IsValid() || sessionValue.Kind() != reflect.Interface && sessionValue.Kind() != reflect.Ptr {
		err := fmt.Errorf("invalid sqlSession type for delete operation")
		if mm.MethodSignature.ReturnsVoid {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf(int64(0))),
			reflect.ValueOf(err),
		}
	}

	// 通过反射调用Delete方法
	deleteMethod := sessionValue.MethodByName("Delete")
	if !deleteMethod.IsValid() {
		err := fmt.Errorf("sqlSession does not have Delete method")
		if mm.MethodSignature.ReturnsVoid {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf(int64(0))),
			reflect.ValueOf(err),
		}
	}

	// 调用Delete方法
	args := []reflect.Value{
		reflect.ValueOf(mm.Command.Name),
		reflect.ValueOf(param),
	}
	results := deleteMethod.Call(args)

	// 记录SQL执行日志
	getConfigMethod := sessionValue.MethodByName("GetConfiguration")
	if getConfigMethod.IsValid() {
		configResults := getConfigMethod.Call([]reflect.Value{})
		if len(configResults) > 0 && !configResults[0].IsNil() {
			log.Printf("Executing DELETE statement: %s with parameters: %v", mm.Command.Name, param)
		}
	}

	if mm.MethodSignature.ReturnsVoid {
		if len(results) > 1 && !results[1].IsNil() {
			return []reflect.Value{results[1]}
		}
		return []reflect.Value{}
	}

	if len(results) >= 2 {
		return []reflect.Value{results[0], results[1]}
	}
	return []reflect.Value{
		reflect.Zero(reflect.TypeOf(int64(0))),
		reflect.ValueOf(fmt.Errorf("unexpected return values from Delete method")),
	}
}

// executeSelect 执行查询操作
func (mm *MapperMethod) executeSelect(sqlSession any, param any) []reflect.Value {
	// 使用反射调用SqlSession的查询方法
	sessionValue := reflect.ValueOf(sqlSession)
	if !sessionValue.IsValid() || sessionValue.Kind() != reflect.Interface && sessionValue.Kind() != reflect.Ptr {
		err := fmt.Errorf("invalid sqlSession type")
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf((*any)(nil)).Elem()),
			reflect.ValueOf(err),
		}
	}

	// 记录SQL执行日志
	getConfigMethod := sessionValue.MethodByName("GetConfiguration")
	if getConfigMethod.IsValid() {
		configResults := getConfigMethod.Call([]reflect.Value{})
		if len(configResults) > 0 && !configResults[0].IsNil() {
			log.Printf("Executing SELECT statement: %s with parameters: %v", mm.Command.Name, param)
		}
	}

	if mm.MethodSignature.ReturnsMany {
		// 返回列表 - 调用SelectList方法
		selectListMethod := sessionValue.MethodByName("SelectList")
		if !selectListMethod.IsValid() {
			err := fmt.Errorf("sqlSession does not have SelectList method")
			return []reflect.Value{
				reflect.Zero(reflect.TypeOf([]any{})),
				reflect.ValueOf(err),
			}
		}

		args := []reflect.Value{
			reflect.ValueOf(mm.Command.Name),
			reflect.ValueOf(param),
		}
		results := selectListMethod.Call(args)

		if len(results) >= 2 {
			return []reflect.Value{results[0], results[1]}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf([]any{})),
			reflect.ValueOf(fmt.Errorf("unexpected return values from SelectList method")),
		}
	} else if mm.MethodSignature.ReturnsMap {
		// 返回Map - 调用SelectMap方法
		selectMapMethod := sessionValue.MethodByName("SelectMap")
		if !selectMapMethod.IsValid() {
			err := fmt.Errorf("sqlSession does not have SelectMap method")
			return []reflect.Value{
				reflect.Zero(reflect.TypeOf(map[string]any{})),
				reflect.ValueOf(err),
			}
		}

		args := []reflect.Value{
			reflect.ValueOf(mm.Command.Name),
			reflect.ValueOf(param),
		}
		results := selectMapMethod.Call(args)

		if len(results) >= 2 {
			return []reflect.Value{results[0], results[1]}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf(map[string]any{})),
			reflect.ValueOf(fmt.Errorf("unexpected return values from SelectMap method")),
		}
	} else {
		// 返回单个对象 - 调用SelectOne方法
		selectOneMethod := sessionValue.MethodByName("SelectOne")
		if !selectOneMethod.IsValid() {
			err := fmt.Errorf("sqlSession does not have SelectOne method")
			return []reflect.Value{
				reflect.Zero(reflect.TypeOf((*any)(nil)).Elem()),
				reflect.ValueOf(err),
			}
		}

		args := []reflect.Value{
			reflect.ValueOf(mm.Command.Name),
			reflect.ValueOf(param),
		}
		results := selectOneMethod.Call(args)

		if len(results) >= 2 {
			return []reflect.Value{results[0], results[1]}
		}
		return []reflect.Value{
			reflect.Zero(reflect.TypeOf((*any)(nil)).Elem()),
			reflect.ValueOf(fmt.Errorf("unexpected return values from SelectOne method")),
		}
	}
}
