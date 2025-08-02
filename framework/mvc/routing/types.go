package routing

import (
	"reflect"
)

// RouteInfo 路由信息（统一结构）
type RouteInfo struct {
	Path            string            // 路径
	HTTPMethod      string            // HTTP方法
	PackageName     string            // 包名
	TypeName        string            // 类型名
	ControllerType  reflect.Type      // 控制器类型
	MethodName      string            // 方法名
	Description     string            // 描述
	Params          []*ParamInfo      // 参数信息
	Middlewares     []string          // 中间件
	Tags            map[string]string // 标签
	Source          AnnotationSource  // 注解来源
}

// AnnotationSource 注解来源枚举
type AnnotationSource string

const (
	SourceStructTag   AnnotationSource = "struct_tag"   // struct标签
	SourceComment     AnnotationSource = "comment"      // Go注释
	SourceManual      AnnotationSource = "manual"       // 手动注册
	SourceHybrid      AnnotationSource = "hybrid"       // 混合方式
)

// ParamInfo 参数信息（统一结构）
type ParamInfo struct {
	Name         string      // 参数名
	Source       ParamSource // 参数来源
	Required     bool        // 是否必需
	DefaultValue string      // 默认值
	Description  string      // 描述
	Type         string      // 参数类型
}

// ParamSource 参数来源枚举
type ParamSource string

const (
	ParamSourcePath   ParamSource = "path"   // 路径参数
	ParamSourceQuery  ParamSource = "query"  // 查询参数
	ParamSourceBody   ParamSource = "body"   // 请求体参数
	ParamSourceHeader ParamSource = "header" // 请求头参数
	ParamSourceCookie ParamSource = "cookie" // Cookie参数
	ParamSourceForm   ParamSource = "form"   // 表单参数
)

// ControllerInfo 控制器信息（统一结构）
type ControllerInfo struct {
	PackageName      string            // 包名
	TypeName         string            // 类型名
	ControllerType   reflect.Type      // 控制器类型
	IsRestController bool              // 是否为REST控制器
	IsController     bool              // 是否为MVC控制器
	BasePath         string            // 基础路径
	Description      string            // 描述
	Tags             map[string]string // 标签
	Source           AnnotationSource  // 注解来源
}

// MethodInfo 方法信息（统一结构）
type MethodInfo struct {
	PackageName string          // 包名
	TypeName    string          // 类型名
	MethodName  string          // 方法名
	HTTPMethod  string          // HTTP方法
	Path        string          // 路径
	Description string          // 描述
	Params      []*ParamInfo    // 参数信息
	Middlewares []string        // 中间件
	Tags        map[string]string // 标签
	Source      AnnotationSource // 注解来源
}

// MethodMapping 方法映射信息（用于annotation包）
type MethodMapping struct {
	ControllerType reflect.Type   // 控制器类型
	MethodName     string         // 方法名
	HTTPMethod     string         // HTTP方法
	Path           string         // 路径
	Description    string         // 描述
	Params         []*ParamInfo   // 参数信息
	Middlewares    []string       // 中间件
	Tags           map[string]string // 标签
}

// 构建器接口，用于链式配置
type MethodMappingBuilder interface {
	WithDescription(desc string) MethodMappingBuilder
	WithPathParam(name string, required bool) MethodMappingBuilder
	WithQueryParam(name, defaultValue string, required bool) MethodMappingBuilder
	WithBodyParam(required bool) MethodMappingBuilder
	WithHeaderParam(name, defaultValue string, required bool) MethodMappingBuilder
	WithCookieParam(name, defaultValue string, required bool) MethodMappingBuilder
	WithMiddleware(middlewares ...string) MethodMappingBuilder
	WithTag(key, value string) MethodMappingBuilder
	Build() *MethodMapping
}

// Router接口，定义路由器的基本功能
type Router interface {
	// 注册路由
	RegisterRoute(route *RouteInfo) error
	
	// 获取已注册的路由
	GetRegisteredRoutes() []*RouteInfo
	
	// 检查路由是否存在
	RouteExists(path, method string) bool
}

// Parser接口，定义解析器的基本功能
type Parser interface {
	// 解析控制器
	ParseController(controller interface{}) (*ControllerInfo, error)
	
	// 解析方法
	ParseMethods(controller interface{}) ([]*MethodInfo, error)
	
	// 获取解析器类型
	GetParserType() AnnotationSource
}

// Registry接口，定义注册表的基本功能
type Registry interface {
	// 注册方法映射
	RegisterMethodMapping(mapping *MethodMapping) error
	
	// 获取控制器的所有映射
	GetControllerMappings(controllerType reflect.Type) []*MethodMapping
	
	// 获取特定方法的映射
	GetMethodMapping(controllerType reflect.Type, methodName string) *MethodMapping
	
	// 清空注册表
	Clear()
}