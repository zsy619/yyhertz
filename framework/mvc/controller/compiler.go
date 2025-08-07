package controller

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/mvc/context"
)

// ControllerCompiler 控制器编译器 - 预编译控制器减少反射调用
type ControllerCompiler struct {
	cache      sync.Map                    // 编译缓存
	methods    sync.Map                    // 方法缓存
	lifecycle  *LifecycleManager          // 生命周期管理器
	precompiled map[string]*CompiledController // 预编译的控制器
	mu         sync.RWMutex               // 读写锁
}

// CompiledController 预编译的控制器
type CompiledController struct {
	Type         reflect.Type                    // 控制器类型
	Methods      map[string]*CompiledMethod      // 编译后的方法
	Instance     interface{}                     // 控制器实例
	Pool         *ControllerPool                 // 控制器池
	Metadata     *ControllerMetadata             // 控制器元数据
	CreatedAt    time.Time                      // 创建时间
}

// CompiledMethod 预编译的方法
type CompiledMethod struct {
	Name         string                 // 方法名
	Handler      MethodHandler          // 编译后的处理器
	ParamBinder  *ParameterBinder       // 参数绑定器
	Validator    *MethodValidator       // 方法验证器
	HTTPMethods  []string               // 支持的HTTP方法
	Path         string                 // 路径模式
	Middleware   []string               // 中间件列表
	CacheEnabled bool                   // 是否启用缓存
}

// MethodHandler 方法处理器类型
type MethodHandler func(ctx *context.Context, controller interface{}) error

// ControllerMetadata 控制器元数据
type ControllerMetadata struct {
	Name         string            // 控制器名称
	Package      string            // 包名
	Methods      []string          // 方法列表
	Tags         map[string]string // 标签信息
	Dependencies []string          // 依赖列表
	Cached       bool              // 是否缓存
}

// CompilerConfig 编译器配置
type CompilerConfig struct {
	EnableCache      bool          // 启用缓存
	CacheSize        int           // 缓存大小
	PrecompileAll    bool          // 预编译所有控制器
	OptimizeLevel    int           // 优化级别 (0-3)
	EnableLifecycle  bool          // 启用生命周期管理
	PoolSize         int           // 控制器池大小
	MaxIdleTime      time.Duration // 最大空闲时间
}

// DefaultCompilerConfig 默认编译器配置
func DefaultCompilerConfig() *CompilerConfig {
	return &CompilerConfig{
		EnableCache:     true,
		CacheSize:       1000,
		PrecompileAll:   false,
		OptimizeLevel:   2,
		EnableLifecycle: true,
		PoolSize:        50,
		MaxIdleTime:     30 * time.Minute,
	}
}

// NewControllerCompiler 创建控制器编译器
func NewControllerCompiler(config *CompilerConfig) *ControllerCompiler {
	if config == nil {
		config = DefaultCompilerConfig()
	}

	compiler := &ControllerCompiler{
		precompiled: make(map[string]*CompiledController),
		lifecycle:   NewLifecycleManager(config),
	}

	return compiler
}

// Compile 编译控制器
func (cc *ControllerCompiler) Compile(controller interface{}) (*CompiledController, error) {
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	controllerName := controllerType.Name()
	
	// 检查缓存
	if cached, exists := cc.getFromCache(controllerName); exists {
		return cached, nil
	}

	// 编译控制器
	compiled, err := cc.compileController(controller, controllerType)
	if err != nil {
		return nil, fmt.Errorf("failed to compile controller %s: %w", controllerName, err)
	}

	// 缓存结果
	cc.cache.Store(controllerName, compiled)
	
	return compiled, nil
}

// compileController 执行控制器编译
func (cc *ControllerCompiler) compileController(controller interface{}, controllerType reflect.Type) (*CompiledController, error) {
	controllerName := controllerType.Name()
	
	compiled := &CompiledController{
		Type:      controllerType,
		Methods:   make(map[string]*CompiledMethod),
		Instance:  controller,
		Pool:      NewControllerPool(controllerType, 10),
		Metadata:  cc.extractMetadata(controllerType),
		CreatedAt: time.Now(),
	}

	// 编译所有公开方法
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		
		// 跳过非公开方法和基础方法
		if !method.IsExported() || cc.isBaseMethod(method.Name) {
			continue
		}

		compiledMethod, err := cc.compileMethod(method, controllerType)
		if err != nil {
			return nil, fmt.Errorf("failed to compile method %s.%s: %w", controllerName, method.Name, err)
		}

		if compiledMethod != nil {
			compiled.Methods[method.Name] = compiledMethod
		}
	}

	return compiled, nil
}

// compileMethod 编译单个方法
func (cc *ControllerCompiler) compileMethod(method reflect.Method, controllerType reflect.Type) (*CompiledMethod, error) {
	methodType := method.Type
	
	// 检查方法签名
	if methodType.NumIn() < 1 { // 至少需要接收者
		return nil, nil
	}

	// 提取HTTP方法和路径
	httpMethods, path := cc.extractRouteInfo(method.Name)
	if len(httpMethods) == 0 {
		return nil, nil // 不是路由方法
	}

	// 创建参数绑定器
	paramBinder, err := NewParameterBinder(methodType)
	if err != nil {
		return nil, fmt.Errorf("failed to create parameter binder: %w", err)
	}

	// 创建方法验证器
	validator := NewMethodValidator(methodType)

	// 创建优化的方法处理器
	handler := cc.createOptimizedHandler(method, paramBinder, validator)

	compiled := &CompiledMethod{
		Name:        method.Name,
		Handler:     handler,
		ParamBinder: paramBinder,
		Validator:   validator,
		HTTPMethods: httpMethods,
		Path:        path,
		Middleware:  cc.extractMiddleware(method),
	}

	return compiled, nil
}

// createOptimizedHandler 创建优化的方法处理器
func (cc *ControllerCompiler) createOptimizedHandler(method reflect.Method, binder *ParameterBinder, validator *MethodValidator) MethodHandler {
	return func(ctx *context.Context, controller interface{}) error {
		// 1. 参数绑定和验证
		params, err := binder.BindParameters(ctx)
		if err != nil {
			return fmt.Errorf("parameter binding failed: %w", err)
		}

		// 2. 参数验证
		if err := validator.ValidateParameters(params); err != nil {
			return fmt.Errorf("parameter validation failed: %w", err)
		}

		// 3. 调用方法（减少反射调用）
		methodValue := reflect.ValueOf(controller).MethodByName(method.Name)
		if !methodValue.IsValid() {
			return fmt.Errorf("method %s not found", method.Name)
		}

		// 构造调用参数
		args := make([]reflect.Value, len(params))
		for i, param := range params {
			args[i] = reflect.ValueOf(param)
		}

		// 执行方法调用
		results := methodValue.Call(args)
		
		// 处理返回值
		return cc.handleMethodResult(ctx, results)
	}
}

// handleMethodResult 处理方法返回值
func (cc *ControllerCompiler) handleMethodResult(ctx *context.Context, results []reflect.Value) error {
	if len(results) == 0 {
		return nil
	}

	// 检查最后一个返回值是否为error
	lastResult := results[len(results)-1]
	if lastResult.Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !lastResult.IsNil() {
			return lastResult.Interface().(error)
		}
	}

	// 处理其他返回值（如数据响应）
	if len(results) > 1 {
		for i := 0; i < len(results)-1; i++ {
			result := results[i]
			if result.IsValid() && !result.IsNil() {
				// 将返回值写入响应
				ctx.JSON(200, result.Interface())
				break
			}
		}
	}

	return nil
}

// extractRouteInfo 提取路由信息
func (cc *ControllerCompiler) extractRouteInfo(methodName string) ([]string, string) {
	// 解析方法名获取HTTP方法
	methodName = strings.ToLower(methodName)
	
	var httpMethods []string
	var path string

	switch {
	case strings.HasPrefix(methodName, "get"):
		httpMethods = []string{"GET"}
		path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "get"))
	case strings.HasPrefix(methodName, "post"):
		httpMethods = []string{"POST"}
		path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "post"))
	case strings.HasPrefix(methodName, "put"):
		httpMethods = []string{"PUT"}
		path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "put"))
	case strings.HasPrefix(methodName, "delete"):
		httpMethods = []string{"DELETE"}
		path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "delete"))
	case strings.HasPrefix(methodName, "patch"):
		httpMethods = []string{"PATCH"}
		path = "/" + strings.ToLower(strings.TrimPrefix(methodName, "patch"))
	default:
		// 不是标准的HTTP方法，跳过
		return nil, ""
	}

	// 处理空路径
	if path == "/" {
		path = "/index"
	}

	return httpMethods, path
}

// extractMetadata 提取控制器元数据
func (cc *ControllerCompiler) extractMetadata(controllerType reflect.Type) *ControllerMetadata {
	metadata := &ControllerMetadata{
		Name:    controllerType.Name(),
		Package: controllerType.PkgPath(),
		Methods: make([]string, 0),
		Tags:    make(map[string]string),
	}

	// 提取方法列表
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		if method.IsExported() {
			metadata.Methods = append(metadata.Methods, method.Name)
		}
	}

	return metadata
}

// extractMiddleware 提取中间件信息
func (cc *ControllerCompiler) extractMiddleware(method reflect.Method) []string {
	// 这里可以通过标签或其他方式提取中间件信息
	// 暂时返回空列表
	return []string{}
}

// isBaseMethod 判断是否为基础方法
func (cc *ControllerCompiler) isBaseMethod(methodName string) bool {
	baseMethods := []string{
		"Prepare", "Finish", "Init", "URLMapping", "HandlerFunc",
		"CheckXSRFCookie", "XSRFToken", "CheckXSRFCookie", "XSRFFormHTML",
		"GetString", "GetStrings", "GetInt", "GetInt8", "GetInt16", "GetInt32", "GetInt64",
		"GetBool", "GetFloat", "GetFile", "SaveToFile", "StartSession", "SetSession",
		"GetSession", "DelSession", "SessionRegenerateID", "DestroySession",
		"IsAjax", "GetSecureCookie", "SetSecureCookie", "XSRFToken", "CheckXSRFCookie",
		"Redirect", "Abort", "CustomAbort", "StopRun", "URLFor", "ServeJSON",
		"ServeJSONP", "ServeXML", "ServeFormatted", "Input", "ParseForm", "GetControllerAndAction",
	}

	for _, baseMethod := range baseMethods {
		if methodName == baseMethod {
			return true
		}
	}
	return false
}

// getFromCache 从缓存获取编译结果
func (cc *ControllerCompiler) getFromCache(controllerName string) (*CompiledController, bool) {
	if value, exists := cc.cache.Load(controllerName); exists {
		return value.(*CompiledController), true
	}
	return nil, false
}

// GetCompiledController 获取编译后的控制器
func (cc *ControllerCompiler) GetCompiledController(controllerName string) (*CompiledController, bool) {
	return cc.getFromCache(controllerName)
}

// PrecompileAll 预编译所有已注册的控制器
func (cc *ControllerCompiler) PrecompileAll(controllers []interface{}) error {
	for _, controller := range controllers {
		if _, err := cc.Compile(controller); err != nil {
			return fmt.Errorf("failed to precompile controller: %w", err)
		}
	}
	return nil
}

// Stats 编译器统计信息
type CompilerStats struct {
	CompiledControllers int           // 编译的控制器数量
	CompiledMethods     int           // 编译的方法数量
	CacheHitRate        float64       // 缓存命中率
	AverageCompileTime  time.Duration // 平均编译时间
	TotalMemoryUsage    int64         // 总内存使用量
}

// GetStats 获取编译器统计信息
func (cc *ControllerCompiler) GetStats() *CompilerStats {
	stats := &CompilerStats{}
	
	cc.cache.Range(func(key, value interface{}) bool {
		stats.CompiledControllers++
		compiled := value.(*CompiledController)
		stats.CompiledMethods += len(compiled.Methods)
		return true
	})

	return stats
}