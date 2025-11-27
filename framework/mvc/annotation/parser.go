package annotation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// AnnotationParser 注解解析器
type AnnotationParser struct {
	registry *MethodRegistry
}

// NewAnnotationParser 创建注解解析器
func NewAnnotationParser() *AnnotationParser {
	return &AnnotationParser{
		registry: GetRegistry(),
	}
}

// ParseController 解析控制器注解
func (p *AnnotationParser) ParseController(controller core.IController) (*ControllerInfo, error) {
	controllerType := reflect.TypeOf(controller)
	originalType := controllerType // 保存原始类型用于方法查找
	
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	if controllerType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("controller must be a struct, got %s", controllerType.Kind())
	}

	// 解析struct标签
	info, err := ParseControllerTags(controllerType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse controller tags: %w", err)
	}

	if info == nil {
		// 如果没有注解标签，创建默认信息
		info = &ControllerInfo{
			Type:    controllerType,
			Name:    controllerType.Name(),
			Methods: make(map[string]*MethodInfo),
			Tags:    make(map[string]string),
		}
	}

	// 解析方法 - 使用原始指针类型来查找方法
	err = p.parseMethods(info, originalType)
	if err != nil {
		return nil, fmt.Errorf("failed to parse methods: %w", err)
	}

	// 注册控制器信息
	p.registry.RegisterControllerInfo(controllerType, info)

	return info, nil
}

// parseMethods 解析控制器方法
func (p *AnnotationParser) parseMethods(info *ControllerInfo, controllerType reflect.Type) error {
	// 获取结构体类型用于映射查找
	structType := controllerType
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	
	// 获取注册的方法映射（基于结构体类型）
	mappings := p.registry.GetControllerMappings(structType)
	
	// 创建方法映射的索引
	mappingIndex := make(map[string]*MethodMapping)
	for _, mapping := range mappings {
		mappingIndex[mapping.MethodName] = mapping
	}

	// 遍历所有方法（使用controllerType来查找方法）
	numMethods := controllerType.NumMethod()
	
	for i := 0; i < numMethods; i++ {
		method := controllerType.Method(i)
		
		// 跳过非公开方法
		if !method.IsExported() {
			continue
		}

		methodInfo := &MethodInfo{
			Name:   method.Name,
			Method: method,
			Tags:   make(map[string]string),
			Params: make([]*ParamInfo, 0),
		}

		// 检查是否有注册的映射
		if mapping, exists := mappingIndex[method.Name]; exists {
			methodInfo.Path = mapping.Path
			methodInfo.HTTPMethod = mapping.HTTPMethod
			methodInfo.Tags = mapping.Tags

			// 解析参数信息
			err := p.parseMethodParameters(methodInfo, method, mapping)
			if err != nil {
				return fmt.Errorf("failed to parse parameters for method %s: %w", method.Name, err)
			}
			
			// 只添加有映射的方法
			info.Methods[method.Name] = methodInfo
		}
	}

	return nil
}

// parseMethodParameters 解析方法参数
func (p *AnnotationParser) parseMethodParameters(methodInfo *MethodInfo, method reflect.Method, mapping *MethodMapping) error {
	methodType := method.Type
	
	// 创建参数映射的索引
	paramMappingIndex := make(map[string]*ParamMapping)
	for _, paramMapping := range mapping.Params {
		paramMappingIndex[paramMapping.Name] = paramMapping
	}

	// 遍历方法参数（跳过receiver）
	for i := 1; i < methodType.NumIn(); i++ {
		paramType := methodType.In(i)
		
		paramInfo := &ParamInfo{
			Name: fmt.Sprintf("param%d", i),
			Type: paramType,
			Tags: make(map[string]string),
		}

		// 根据参数类型推断参数种类
		paramInfo.Kind = p.inferParamKind(paramType)

		// 如果有对应的参数映射，使用映射信息
		if paramMapping, exists := paramMappingIndex[paramInfo.Name]; exists {
			paramInfo.Required = paramMapping.Required
			paramInfo.DefaultVal = paramMapping.DefaultValue
			paramInfo.Validation = paramMapping.Validation
		}

		methodInfo.Params = append(methodInfo.Params, paramInfo)
	}

	return nil
}

// inferMethodMapping 自动推断方法映射
func (p *AnnotationParser) inferMethodMapping(methodInfo *MethodInfo, method reflect.Method) {
	methodName := method.Name
	
	// 根据方法名推断HTTP方法和路径
	if strings.HasPrefix(methodName, "Get") {
		methodInfo.HTTPMethod = "GET"
		methodInfo.Path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "Get"))
	} else if strings.HasPrefix(methodName, "Post") {
		methodInfo.HTTPMethod = "POST"
		methodInfo.Path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "Post"))
	} else if strings.HasPrefix(methodName, "Put") {
		methodInfo.HTTPMethod = "PUT"
		methodInfo.Path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "Put"))
	} else if strings.HasPrefix(methodName, "Delete") {
		methodInfo.HTTPMethod = "DELETE"
		methodInfo.Path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "Delete"))
	} else if strings.HasPrefix(methodName, "Patch") {
		methodInfo.HTTPMethod = "PATCH"
		methodInfo.Path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "Patch"))
	} else {
		// 默认为GET方法
		methodInfo.HTTPMethod = "GET"
		methodInfo.Path = "/" + strings.ToLower(methodName)
	}

	// 规范化路径
	methodInfo.Path = normalizePath(methodInfo.Path)
}

// inferParamKind 推断参数类型
func (p *AnnotationParser) inferParamKind(paramType reflect.Type) ParamKind {
	// 根据参数类型推断参数种类
	switch {
	case isContextType(paramType):
		return ParamKindQuery // Context类型通常用于获取查询参数
	case isStructType(paramType):
		return ParamKindBody // 结构体类型通常用作请求体
	case isStringType(paramType):
		return ParamKindQuery // 字符串类型通常用作查询参数
	default:
		return ParamKindQuery // 默认为查询参数
	}
}

// isContextType 检查是否为Context类型
func isContextType(t reflect.Type) bool {
	return t.String() == "*context.Context" || 
		   strings.Contains(t.String(), "Context")
}

// isStructType 检查是否为结构体类型
func isStructType(t reflect.Type) bool {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Struct
}

// isStringType 检查是否为字符串类型
func isStringType(t reflect.Type) bool {
	return t.Kind() == reflect.String
}

// ParseAllControllers 解析所有控制器
func (p *AnnotationParser) ParseAllControllers(controllers ...core.IController) ([]*ControllerInfo, error) {
	var infos []*ControllerInfo
	
	for _, controller := range controllers {
		info, err := p.ParseController(controller)
		if err != nil {
			return nil, fmt.Errorf("failed to parse controller %T: %w", controller, err)
		}
		infos = append(infos, info)
	}
	
	return infos, nil
}

// BuildRouteInfo 构建路由信息
func (p *AnnotationParser) BuildRouteInfo(info *ControllerInfo) []*RouteInfo {
	var routes []*RouteInfo
	
	for _, methodInfo := range info.Methods {
		if methodInfo.HTTPMethod != "" {
			fullPath := CombinePath(info.BasePath, methodInfo.Path)
			
			route := &RouteInfo{
				Path:           fullPath,
				HTTPMethod:     methodInfo.HTTPMethod,
				ControllerType: info.Type,
				MethodName:     methodInfo.Name,
				Method:         methodInfo.Method,
				Params:         methodInfo.Params,
				Tags:           mergeTags(info.Tags, methodInfo.Tags),
			}
			routes = append(routes, route)
		}
	}
	
	return routes
}

// RouteInfo 路由信息
type RouteInfo struct {
	Path           string
	HTTPMethod     string
	ControllerType reflect.Type
	MethodName     string
	Method         reflect.Method
	Params         []*ParamInfo
	Tags           map[string]string
}

// mergeTags 合并标签
func mergeTags(controllerTags, methodTags map[string]string) map[string]string {
	merged := make(map[string]string)
	
	// 复制控制器标签
	for k, v := range controllerTags {
		merged[k] = v
	}
	
	// 复制方法标签（方法标签优先级更高）
	for k, v := range methodTags {
		merged[k] = v
	}
	
	return merged
}