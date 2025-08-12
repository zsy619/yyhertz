package constant

// ExecutionLayer 统一的执行层级类型
// 用于定义中间件和过滤器在请求处理生命周期中的执行位置
type ExecutionLayer int

// 统一的执行层级常量
// 这些常量定义了请求处理过程中所有可能的执行时机点
const (
	// LayerBeforeStatic 静态文件处理前
	// 在静态文件处理器执行之前运行，用于静态资源访问控制
	LayerBeforeStatic ExecutionLayer = iota

	// LayerGlobal 全局处理层（路由匹配前）
	// 在路由匹配之前运行，通常用于全局认证、CORS、限流等
	LayerGlobal

	// LayerGroup 路由组处理层
	// 在路由组级别运行，用于路由组特定的中间件处理
	LayerGroup

	// LayerRoute 路由处理层（控制器执行前）
	// 在控制器方法执行之前运行，用于参数验证、权限检查等
	LayerRoute

	// LayerController 控制器处理层（控制器执行后）
	// 在控制器方法执行之后运行，用于响应处理、日志记录等
	LayerController

	// LayerFinishRouter 请求处理完成后
	// 在整个请求处理完成后运行，用于清理资源、性能统计等
	LayerFinishRouter
)

// 过滤器位置常量的映射（向后兼容）
// 过滤器系统使用的是简化的层级概念，不包含路由组层
const (
	// BeforeStatic 静态文件处理前 -> LayerBeforeStatic
	BeforeStatic = int(LayerBeforeStatic)

	// BeforeRouter 路由匹配前 -> LayerGlobal  
	BeforeRouter = int(LayerGlobal)

	// BeforeExec 控制器执行前 -> LayerRoute
	BeforeExec = int(LayerRoute)

	// AfterExec 控制器执行后 -> LayerController
	AfterExec = int(LayerController)

	// FinishRouter 请求处理完成后 -> LayerFinishRouter
	FinishRouter = int(LayerFinishRouter)
)

// ExecutionLayerNames 执行层级名称映射
var ExecutionLayerNames = map[ExecutionLayer]string{
	LayerBeforeStatic: "BeforeStatic",
	LayerGlobal:       "Global",
	LayerGroup:        "Group",
	LayerRoute:        "Route", 
	LayerController:   "Controller",
	LayerFinishRouter: "FinishRouter",
}

// FilterPositionNames 过滤器位置名称映射（向后兼容）
var FilterPositionNames = map[int]string{
	BeforeStatic: "BeforeStatic",
	BeforeRouter: "BeforeRouter", 
	BeforeExec:   "BeforeExec",
	AfterExec:    "AfterExec",
	FinishRouter: "FinishRouter",
}

// GetExecutionLayerName 获取执行层级名称
func GetExecutionLayerName(layer ExecutionLayer) string {
	if name, exists := ExecutionLayerNames[layer]; exists {
		return name
	}
	return "Unknown"
}

// GetFilterPositionName 获取过滤器位置名称（向后兼容）
func GetFilterPositionName(position int) string {
	if name, exists := FilterPositionNames[position]; exists {
		return name
	}
	return "Unknown"
}

// IsValidExecutionLayer 验证执行层级是否有效
func IsValidExecutionLayer(layer ExecutionLayer) bool {
	return layer >= LayerBeforeStatic && layer <= LayerFinishRouter
}

// IsValidFilterPosition 验证过滤器位置是否有效（向后兼容）
func IsValidFilterPosition(position int) bool {
	return position >= BeforeStatic && position <= FinishRouter
}

// FilterPositionToLayer 将过滤器位置转换为执行层级
func FilterPositionToLayer(position int) ExecutionLayer {
	switch position {
	case BeforeStatic:
		return LayerBeforeStatic
	case BeforeRouter:
		return LayerGlobal
	case BeforeExec:
		return LayerRoute
	case AfterExec:
		return LayerController
	case FinishRouter:
		return LayerFinishRouter
	default:
		return LayerGlobal // 默认返回全局层级
	}
}

// LayerToFilterPosition 将执行层级转换为过滤器位置（仅支持过滤器相关层级）
func LayerToFilterPosition(layer ExecutionLayer) int {
	switch layer {
	case LayerBeforeStatic:
		return BeforeStatic
	case LayerGlobal:
		return BeforeRouter
	case LayerRoute:
		return BeforeExec
	case LayerController:
		return AfterExec
	case LayerFinishRouter:
		return FinishRouter
	default:
		return BeforeRouter // 默认返回路由前位置
	}
}

// GetAllExecutionLayers 获取所有执行层级
func GetAllExecutionLayers() []ExecutionLayer {
	return []ExecutionLayer{
		LayerBeforeStatic,
		LayerGlobal,
		LayerGroup,
		LayerRoute,
		LayerController,
		LayerFinishRouter,
	}
}

// GetFilterExecutionLayers 获取过滤器相关的执行层级
func GetFilterExecutionLayers() []ExecutionLayer {
	return []ExecutionLayer{
		LayerBeforeStatic,
		LayerGlobal,
		LayerRoute,
		LayerController,
		LayerFinishRouter,
	}
}