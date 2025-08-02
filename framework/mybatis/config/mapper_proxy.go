// Package config 映射器代理实现
//
// 提供动态代理功能，为映射器接口生成运行时代理对象
package config

import (
	"reflect"
	"fmt"
)

// MapperProxy 映射器代理
type MapperProxy struct {
	sqlSession      any
	mapperInterface reflect.Type
	methodCache     map[string]*MapperMethod
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
	// 使用反射创建接口实现
	proxyValue := reflect.New(mapperInterface).Elem()
	
	// 为每个方法创建实现
	for i := 0; i < mapperInterface.NumMethod(); i++ {
		method := mapperInterface.Method(i)
		proxyValue.Field(i).Set(reflect.MakeFunc(method.Type, func(args []reflect.Value) []reflect.Value {
			return proxy.invoke(method.Name, args[1:]) // 跳过接收者参数
		}))
	}
	
	return proxyValue.Interface()
}

// invoke 调用映射器方法
func (mp *MapperProxy) invoke(methodName string, args []reflect.Value) []reflect.Value {
	// 获取或创建映射器方法
	mapperMethod := mp.cachedMapperMethod(methodName)
	if mapperMethod == nil {
		return []reflect.Value{reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())}
	}
	
	// 执行映射器方法
	return mapperMethod.execute(mp.sqlSession, args)
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
		ReturnsMany:   mp.returnsMany(method.Type),
		ReturnsMap:    mp.returnsMap(method.Type),
		ReturnsVoid:   mp.returnsVoid(method.Type),
		ReturnsCursor: false, // Go中不需要游标
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
	case len(methodName) >= 6 && methodName[:6] == "insert" || methodName[:6] == "Insert":
		return SqlCommandTypeInsert
	case len(methodName) >= 6 && methodName[:6] == "update" || methodName[:6] == "Update":
		return SqlCommandTypeUpdate
	case len(methodName) >= 6 && methodName[:6] == "delete" || methodName[:6] == "Delete":
		return SqlCommandTypeDelete
	case len(methodName) >= 6 && methodName[:6] == "select" || methodName[:6] == "Select":
		return SqlCommandTypeSelect
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
		return args[0].Interface()
	}
	
	// 多参数情况，转换为map
	paramMap := make(map[string]any)
	for i, arg := range args {
		paramMap[fmt.Sprintf("param%d", i+1)] = arg.Interface()
	}
	return paramMap
}

// executeInsert 执行插入操作
func (mm *MapperMethod) executeInsert(sqlSession any, param any) []reflect.Value {
	// 这里需要调用sqlSession的Insert方法
	// 简化实现，实际需要类型断言和错误处理
	result := int64(1) // 模拟插入结果
	var err error
	
	if mm.MethodSignature.ReturnsVoid {
		if err != nil {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{}
	}
	
	return []reflect.Value{
		reflect.ValueOf(result),
		reflect.ValueOf(err),
	}
}

// executeUpdate 执行更新操作
func (mm *MapperMethod) executeUpdate(sqlSession any, param any) []reflect.Value {
	result := int64(1) // 模拟更新结果
	var err error
	
	if mm.MethodSignature.ReturnsVoid {
		if err != nil {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{}
	}
	
	return []reflect.Value{
		reflect.ValueOf(result),
		reflect.ValueOf(err),
	}
}

// executeDelete 执行删除操作
func (mm *MapperMethod) executeDelete(sqlSession any, param any) []reflect.Value {
	result := int64(1) // 模拟删除结果
	var err error
	
	if mm.MethodSignature.ReturnsVoid {
		if err != nil {
			return []reflect.Value{reflect.ValueOf(err)}
		}
		return []reflect.Value{}
	}
	
	return []reflect.Value{
		reflect.ValueOf(result),
		reflect.ValueOf(err),
	}
}

// executeSelect 执行查询操作
func (mm *MapperMethod) executeSelect(sqlSession any, param any) []reflect.Value {
	if mm.MethodSignature.ReturnsMany {
		// 返回列表
		var result []any
		var err error
		
		return []reflect.Value{
			reflect.ValueOf(result),
			reflect.ValueOf(err),
		}
	} else if mm.MethodSignature.ReturnsMap {
		// 返回Map
		var result map[string]any
		var err error
		
		return []reflect.Value{
			reflect.ValueOf(result),
			reflect.ValueOf(err),
		}
	} else {
		// 返回单个对象
		var result any
		var err error
		
		return []reflect.Value{
			reflect.ValueOf(result),
			reflect.ValueOf(err),
		}
	}
}