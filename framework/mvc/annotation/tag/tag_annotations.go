package tag

import (
	"reflect"
	"strings"
)

// Tag-based annotation system using struct tags
// 基于struct标签的注解系统

// TagController struct标签控制器信息
type TagControllerInfo struct {
	Type         reflect.Type            // 控制器类型
	Name         string                  // 控制器名称
	BasePath     string                  // 基础路径
	IsRest       bool                    // 是否为REST控制器
	IsController bool                    // 是否为MVC控制器
	Methods      map[string]*TagMethodInfo // 方法信息映射
	Tags         map[string]string       // 所有标签信息
}

// TagMethodInfo struct标签方法信息
type TagMethodInfo struct {
	Name         string                 // 方法名称
	Path         string                 // 路径
	HTTPMethod   string                 // HTTP方法
	Params       []*TagParamInfo        // 参数信息
	Tags         map[string]string      // 方法标签
	Method       reflect.Method         // 反射方法信息
}

// TagParamInfo struct标签参数信息
type TagParamInfo struct {
	Name        string            // 参数名称
	Type        reflect.Type      // 参数类型
	Source      TagParamSource    // 参数来源
	Required    bool              // 是否必需
	DefaultVal  string            // 默认值
	Tags        map[string]string // 参数标签
}

// TagParamSource 参数来源枚举
type TagParamSource int

const (
	TagParamSourcePath TagParamSource = iota
	TagParamSourceQuery
	TagParamSourceBody
	TagParamSourceHeader
	TagParamSourceCookie
	TagParamSourceForm
)

// Struct标签常量
const (
	RestControllerTag = "rest"      // REST控制器标签
	ControllerTag     = "controller" // MVC控制器标签
	RequestMappingTag = "mapping"   // 路径映射标签
	ServiceTag        = "service"   // 服务标签
	RepositoryTag     = "repository" // 仓库标签
	ComponentTag      = "component" // 组件标签
)

// ParseTagControllerInfo 解析struct标签控制器信息
func ParseTagControllerInfo(structType reflect.Type) (*TagControllerInfo, error) {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	
	if structType.Kind() != reflect.Struct {
		return nil, nil
	}

	info := &TagControllerInfo{
		Type:    structType,
		Name:    structType.Name(),
		Methods: make(map[string]*TagMethodInfo),
		Tags:    make(map[string]string),
	}

	// 解析struct字段的标签
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// 解析各种标签
		if restTag := field.Tag.Get(RestControllerTag); restTag != "" {
			info.IsRest = true
			info.Tags[RestControllerTag] = restTag
		}
		
		if controllerTag := field.Tag.Get(ControllerTag); controllerTag != "" {
			info.IsController = true
			info.Tags[ControllerTag] = controllerTag
		}
		
		if mappingTag := field.Tag.Get(RequestMappingTag); mappingTag != "" {
			info.BasePath = normalizePath(mappingTag)
			info.Tags[RequestMappingTag] = mappingTag
		}
		
		if serviceTag := field.Tag.Get(ServiceTag); serviceTag != "" {
			info.Tags[ServiceTag] = serviceTag
		}
		
		if repoTag := field.Tag.Get(RepositoryTag); repoTag != "" {
			info.Tags[RepositoryTag] = repoTag
		}
		
		if compTag := field.Tag.Get(ComponentTag); compTag != "" {
			info.Tags[ComponentTag] = compTag
		}
	}

	return info, nil
}

// HasTagAnnotations 检查类型是否有标签注解
func HasTagAnnotations(structType reflect.Type) bool {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	
	if structType.Kind() != reflect.Struct {
		return false
	}
	
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// 检查是否有任何注解标签
		if field.Tag.Get(RestControllerTag) != "" ||
		   field.Tag.Get(ControllerTag) != "" ||
		   field.Tag.Get(RequestMappingTag) != "" ||
		   field.Tag.Get(ServiceTag) != "" ||
		   field.Tag.Get(RepositoryTag) != "" ||
		   field.Tag.Get(ComponentTag) != "" {
			return true
		}
	}
	
	return false
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

// CombinePath 组合路径
func CombinePath(basePath, methodPath string) string {
	basePath = normalizePath(basePath)
	methodPath = normalizePath(methodPath)
	
	if basePath == "" {
		return methodPath
	}
	
	if methodPath == "" {
		return basePath
	}
	
	return basePath + methodPath
}