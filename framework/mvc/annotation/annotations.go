package annotation

import (
	"reflect"
	"strings"
)

// RestController struct级别注解，表示这是一个REST控制器
// 使用方式: type UserController struct { `rest:""`}
const RestControllerTag = "rest"

// Controller struct级别注解，表示这是一个MVC控制器
// 使用方式: type UserController struct { `controller:""`}
const ControllerTag = "controller"

// RequestMapping struct级别注解，为整个控制器设置基础路径
// 使用方式: type UserController struct { `mapping:"/api/users"`}
const RequestMappingTag = "mapping"

// Service struct级别注解，表示这是一个服务组件
// 使用方式: type UserService struct { `service:""`}
const ServiceTag = "service"

// Repository struct级别注解，表示这是一个数据访问组件
// 使用方式: type UserRepository struct { `repository:""`}
const RepositoryTag = "repository"

// Component struct级别注解，表示这是一个通用组件
// 使用方式: type UserComponent struct { `component:""`}
const ComponentTag = "component"

// HTTP方法注解常量
const (
	GetMappingTag     = "get"      // GET请求映射
	PostMappingTag    = "post"     // POST请求映射
	PutMappingTag     = "put"      // PUT请求映射
	DeleteMappingTag  = "delete"   // DELETE请求映射
	PatchMappingTag   = "patch"    // PATCH请求映射
	HeadMappingTag    = "head"     // HEAD请求映射
	OptionsMappingTag = "options"  // OPTIONS请求映射
	AnyMappingTag     = "any"      // 任意方法映射
)

// 参数注解常量
const (
	PathVariableTag   = "path"     // 路径变量 @PathVariable
	RequestParamTag   = "param"    // 请求参数 @RequestParam  
	RequestBodyTag    = "body"     // 请求体 @RequestBody
	RequestHeaderTag  = "header"   // 请求头 @RequestHeader
	CookieValueTag    = "cookie"   // Cookie值 @CookieValue
)

// 验证注解常量
const (
	ValidTag     = "valid"     // 参数验证
	RequiredTag  = "required"  // 必需参数
	MinTag       = "min"       // 最小值
	MaxTag       = "max"       // 最大值
	SizeTag      = "size"      // 大小限制
	PatternTag   = "pattern"   // 正则模式
)

// ControllerInfo 控制器信息
type ControllerInfo struct {
	Type         reflect.Type            // 控制器类型
	Name         string                  // 控制器名称
	BasePath     string                  // 基础路径
	IsRest       bool                    // 是否为REST控制器
	Methods      map[string]*MethodInfo  // 方法信息映射
	Tags         map[string]string       // 所有标签信息
}

// MethodInfo 方法信息
type MethodInfo struct {
	Name         string                 // 方法名称
	Path         string                 // 路径
	HTTPMethod   string                 // HTTP方法
	Params       []*ParamInfo           // 参数信息
	Tags         map[string]string      // 方法标签
	Method       reflect.Method         // 反射方法信息
}

// ParamInfo 参数信息
type ParamInfo struct {
	Name        string            // 参数名称
	Type        reflect.Type      // 参数类型
	Kind        ParamKind         // 参数类型
	Required    bool              // 是否必需
	DefaultVal  string            // 默认值
	Validation  *ValidationInfo   // 验证信息
	Tags        map[string]string // 参数标签
}

// ParamKind 参数类型枚举
type ParamKind int

const (
	ParamKindPath ParamKind = iota
	ParamKindQuery
	ParamKindBody
	ParamKindHeader
	ParamKindCookie
	ParamKindForm
)

// ValidationInfo 验证信息
type ValidationInfo struct {
	Required bool   // 是否必需
	Min      *int   // 最小值
	Max      *int   // 最大值
	Size     string // 大小限制 "min,max"
	Pattern  string // 正则模式
}

// ParseControllerTags 解析控制器struct标签
func ParseControllerTags(structType reflect.Type) (*ControllerInfo, error) {
	if structType.Kind() != reflect.Struct {
		return nil, nil
	}

	info := &ControllerInfo{
		Type:    structType,
		Name:    structType.Name(),
		Methods: make(map[string]*MethodInfo),
		Tags:    make(map[string]string),
	}

	// 解析struct标签
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		tagString := string(field.Tag)
		
		// 解析各种标签，包括空值标签
		if strings.Contains(tagString, RestControllerTag) {
			restTag := field.Tag.Get(RestControllerTag)
			info.IsRest = true
			info.Tags[RestControllerTag] = restTag
		}
		
		if strings.Contains(tagString, ControllerTag) {
			controllerTag := field.Tag.Get(ControllerTag)
			info.Tags[ControllerTag] = controllerTag
		}
		
		if strings.Contains(tagString, RequestMappingTag) {
			mappingTag := field.Tag.Get(RequestMappingTag)
			if mappingTag != "" {
				info.BasePath = mappingTag
			}
			info.Tags[RequestMappingTag] = mappingTag
		}
		
		if strings.Contains(tagString, ServiceTag) {
			serviceTag := field.Tag.Get(ServiceTag)
			info.Tags[ServiceTag] = serviceTag
		}
		
		if strings.Contains(tagString, RepositoryTag) {
			repoTag := field.Tag.Get(RepositoryTag)
			info.Tags[RepositoryTag] = repoTag
		}
		
		if strings.Contains(tagString, ComponentTag) {
			compTag := field.Tag.Get(ComponentTag)
			info.Tags[ComponentTag] = compTag
		}
	}

	// 规范化基础路径
	if info.BasePath != "" {
		info.BasePath = normalizePath(info.BasePath)
	}

	return info, nil
}

// ParseMethodTags 解析方法标签
func ParseMethodTags(method reflect.Method) (*MethodInfo, error) {
	methodType := method.Type
	
	info := &MethodInfo{
		Name:   method.Name,
		Method: method,
		Params: make([]*ParamInfo, 0),
		Tags:   make(map[string]string),
	}

	// 这里需要通过其他方式获取方法标签，因为Go的reflect无法直接获取方法标签
	// 我们需要实现一个注册机制来存储方法标签信息
	
	// 解析参数
	for i := 1; i < methodType.NumIn(); i++ { // 跳过receiver
		paramType := methodType.In(i)
		paramInfo := &ParamInfo{
			Name: getParamName(i),
			Type: paramType,
			Tags: make(map[string]string),
		}
		
		info.Params = append(info.Params, paramInfo)
	}

	return info, nil
}

// normalizePath 规范化路径
func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	
	if strings.HasSuffix(path, "/") && path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	
	return path
}

// getParamName 获取参数名称（简化实现）
func getParamName(index int) string {
	return ""
}

// CombinePath 组合路径
func CombinePath(basePath, methodPath string) string {
	basePath = normalizePath(basePath)
	methodPath = normalizePath(methodPath)
	
	if basePath == "" {
		return methodPath
	}
	
	if methodPath == "" || methodPath == "/" {
		return basePath
	}
	
	return basePath + methodPath
}