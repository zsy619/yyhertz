package core

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

// ============= 控制器管理方法 =============

// SetControllerName 设置控制器名称
func (c *BaseController) SetControllerName(name string) {
	c.ControllerName = name
}

// GetControllerName 获取控制器名称（自动初始化）
func (c *BaseController) GetControllerName() string {
	// 使用新的自动检测机制
	c.autoDetectAndSetControllerName()
	return c.ControllerName
}

// SetActionName 设置动作名称
func (c *BaseController) SetActionName(name string) {
	c.ActionName = name
}

// GetActionName 获取当前动作名称（自动初始化）
func (c *BaseController) GetActionName() string {
	c.ensureInitialized()
	return c.ActionName
}

// SetAppController 设置应用控制器引用
func (c *BaseController) SetAppController(controller IController) {
	c.AppController = controller
}

// GetAppController 获取应用控制器引用
func (c *BaseController) GetAppController() IController {
	return c.AppController
}

// SetControllerInstance 手动设置控制器实例（推荐方式）
func (c *BaseController) SetControllerInstance(controller IController) {
	c.SetAppController(controller)

	// 通过反射自动设置ControllerName
	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	typeName := controllerType.Name()
	if strings.HasSuffix(typeName, "Controller") {
		c.ControllerName = typeName[:len(typeName)-10]
		c.initialized = true
	}
}

// GetControllerAndAction 从路由中解析控制器和动作
func (c *BaseController) GetControllerAndAction() (string, string) {
	return c.ControllerName, c.ActionName
}

// SetControllerAndAction 设置控制器和动作
func (c *BaseController) SetControllerAndAction(controller, action string) {
	c.ControllerName = controller
	c.ActionName = action
}

// IsValidAction 检查动作是否有效
func (c *BaseController) IsValidAction(action string) bool {
	// 检查是否是保留方法
	if ReservedMethods[action] {
		return false
	}

	// 检查动作名称是否符合规范（首字母大写）
	if len(action) == 0 {
		return false
	}

	firstChar := action[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}

// GetAvailableActions 获取可用的动作列表
func (c *BaseController) GetAvailableActions() []string {
	if c.AppController == nil {
		return []string{}
	}

	// 使用反射获取控制器的所有公共方法
	return getControllerMethods(c.AppController)
}

// AutoInit 通用自动初始化方法（用户友好版）
func (c *BaseController) AutoInit() {
	// 自动设置AppController为当前实例
	// 这个技巧是关键：通过any来获取实际的控制器实例
	if c.AppController == nil {
		// 由于Go的类型系统限制，我们无法直接从BaseController获取外层结构体
		// 但我们可以通过调用栈检测来推断控制器类型并设置ControllerName
		c.autoInitializeControllerName()
	}
}

// ============= 内部方法和辅助函数 =============

// ensureInitialized 确保控制器已初始化（智能版本）
func (c *BaseController) ensureInitialized() {
	// 检查是否需要初始化控制器名称
	if c.ControllerName == "" || c.ControllerName == "UnknownController" {
		// 首先尝试通过AppController获取
		if c.AppController != nil {
			c.ControllerName = ExtractControllerName(c.AppController)
		} else {
			// 通过调用栈检测
			detectedName := c.detectControllerNameFromStack()
			if detectedName != "BaseController" && detectedName != "Base" {
				// 成功检测到具体控制器类型
				c.ControllerName = detectedName
			} else {
				// 无法检测到具体类型，给出提示
				// 在直接调用情况下，我们推荐用户先调用SetAppController
				c.ControllerName = "UnknownController"
			}
		}
	}

	// 动作名称需要每次重新检测，因为它会随调用的方法而变化
	c.ActionName = c.detectCurrentAction()

	// 初始化方法映射（如果未设置）
	if len(c.MethodMapping) == 0 {
		controllerRef := c.AppController
		if controllerRef == nil {
			c.MethodMapping = make(map[string]string)
		} else {
			c.MethodMapping = CreateDefaultMethodMapping(controllerRef)
		}
	}
}

// autoInitializeControllerInfo 自动初始化控制器信息（改进版）
func (c *BaseController) autoInitializeControllerInfo() {
	// 优先通过调用栈检测实际的控制器类型
	detectedName := c.detectControllerNameFromStack()

	// 自动设置控制器名称（如果未设置）
	if c.ControllerName == "" {
		if detectedName != "BaseController" {
			// 成功检测到具体控制器类型
			c.ControllerName = detectedName
		} else if c.AppController != nil {
			// 备用方案：使用AppController引用
			c.ControllerName = ExtractControllerName(c.AppController)
		} else {
			// 最后备用方案：使用调用栈检测的结果
			c.ControllerName = detectedName
		}
	}

	// 自动设置动作名称（通过调用栈分析）
	if c.ActionName == "" {
		c.ActionName = c.detectCurrentAction()
	}

	// 自动创建方法映射（如果未设置）
	if len(c.MethodMapping) == 0 {
		controllerRef := c.AppController
		if controllerRef == nil {
			// 无法获取具体控制器类型，创建基础映射
			c.MethodMapping = make(map[string]string)
		} else {
			c.MethodMapping = CreateDefaultMethodMapping(controllerRef)
		}
	}
}

// autoInitializeControllerName 自动初始化控制器名称（如果尚未初始化）
func (c *BaseController) autoInitializeControllerName() {
	if c.initialized {
		return // 已经初始化过了
	}

	// 通过调用栈分析或其他方式来获取
	detectedName := c.detectActualControllerType()

	if detectedName != "" && detectedName != "BaseController" && detectedName != "Base" {
		c.ControllerName = detectedName
		c.initialized = true
		return
	}

	// 备选方案：如果AppController已设置，使用它
	if c.AppController != nil {
		c.ControllerName = ExtractControllerName(c.AppController)
		c.initialized = true
		return
	}

	// 最后才设置为未知
	if c.ControllerName == "" {
		c.ControllerName = "UnknownController"
		c.initialized = true
	}
}

// autoDetectAndSetControllerName 自动检测并设置Controller名称（终极版）
func (c *BaseController) autoDetectAndSetControllerName() {
	if c.initialized && c.ControllerName != "" && c.ControllerName != "UnknownController" {
		fmt.Println("Controller name already initialized:", c.ControllerName)
		return // 已经正确初始化了
	}

	{
		// controllerType := reflect.TypeOf(c)
		// for controllerType.Kind() == reflect.Ptr {
		// 	controllerType = controllerType.Elem()
		// }
		// controllerName := controllerType.Name()

		// // 检查是否有Controller后缀
		// for suffix := range ControllerNameSuffixReserved {
		// 	if strings.HasSuffix(controllerName, suffix) {
		// 		controllerName = strings.TrimSuffix(controllerName, suffix)
		// 		c.ControllerName = controllerName
		// 		c.initialized = true
		// 		fmt.Println("Controller name set from suffix:", c.ControllerName)
		// 		return
		// 	}
		// }
	}

	// 方法1: 通过全局注册表查找
	outerType := c.getOuterStructType()
	if outerType != nil {
		if name, exists := GetRegisteredControllerName(outerType); exists {
			c.ControllerName = name
			c.initialized = true
			return
		}

		// 方法2: 直接从类型名称推导
		typeName := outerType.Name()
		if strings.HasSuffix(typeName, "Controller") {
			c.ControllerName = typeName[:len(typeName)-10]
			c.initialized = true

			// 同时注册到全局表中，以便下次快速访问
			RegisterControllerType(outerType, c.ControllerName)
			return
		}
	}

	// 方法3: 备选方案 - 调用栈检测
	if detectedName := c.detectTypeFromRuntimeInfo(); detectedName != "" {
		c.ControllerName = detectedName
		c.initialized = true
		return
	}

	// 最后设置为未知
	if c.ControllerName == "" {
		c.ControllerName = "UnknownController"
		c.initialized = true
	}
}

// detectActualControllerType 检测实际的控制器类型（增强版）
func (c *BaseController) detectActualControllerType() string {
	// 方法1: 通过调用栈检测
	stackName := c.detectFromCallStack()
	if stackName != "" {
		return stackName
	}

	// 方法2: 通过反射检测外层结构体（新增）
	reflectionName := c.detectFromReflection()
	if reflectionName != "" {
		return reflectionName
	}

	return ""
}

// detectFromCallStack 从调用栈检测控制器类型
func (c *BaseController) detectFromCallStack() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		funcName := frame.Function
		if funcName == "" {
			continue
		}

		// 跳过框架内部方法
		if strings.Contains(funcName, "autoInitializeControllerName") ||
			strings.Contains(funcName, "detectActualControllerType") ||
			strings.Contains(funcName, "detectFromCallStack") ||
			strings.Contains(funcName, "detectFromReflection") ||
			strings.Contains(funcName, "BaseController") {
			continue
		}

		// 查找包含Controller的调用
		if strings.Contains(funcName, "Controller") {
			parts := strings.Split(funcName, ".")
			for _, part := range parts {
				if strings.Contains(part, "Controller") && strings.Contains(part, "(") {
					controllerType := strings.Trim(part, "(*)")
					if strings.HasSuffix(controllerType, "Controller") {
						name := controllerType[:len(controllerType)-10]
						return name
					}
				}
			}
		}
	}

	return ""
}

// detectFromReflection 通过反射检测外层结构体类型（高级版）
func (c *BaseController) detectFromReflection() string {
	// 新策略：通过查找包含当前BaseController的外层结构体

	// 遍历调用栈，寻找可能的外层结构体线索
	pc := make([]uintptr, 30)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		// 查找可能包含Controller类型信息的函数名
		funcName := frame.Function
		if funcName == "" {
			continue
		}

		// 跳过框架内部函数
		if c.isFrameworkInternalFunction(funcName) {
			continue
		}

		// 尝试从函数名中提取Controller类型信息
		if controllerName := c.extractControllerNameFromFuncName(funcName); controllerName != "" {
			return controllerName
		}
	}

	return ""
}

// getOuterStructType 获取外层结构体类型（核心方法）
func (c *BaseController) getOuterStructType() reflect.Type {
	// 这是最关键的方法：通过内存地址和反射获取外层结构体类型

	// 方法1: 通过unsafe指针转换尝试获取外层类型
	// 由于Go的嵌入机制，BaseController和外层结构体共享相同的起始地址

	// 获取当前BaseController的地址
	baseAddr := uintptr(unsafe.Pointer(c))

	// 尝试通过反射系统获取这个地址对应的实际类型
	return c.inferTypeFromAddress(baseAddr)
}

// inferTypeFromAddress 从地址推断类型（实验性方法）
func (c *BaseController) inferTypeFromAddress(addr uintptr) reflect.Type {
	// 在Go中，由于类型安全的限制，我们无法直接从地址推断任意类型
	// 但我们可以使用一些运行时技巧

	// 检查是否可以通过调用栈上下文推断类型
	return c.inferTypeFromCallStack()
}

// inferTypeFromCallStack 从调用栈推断类型
func (c *BaseController) inferTypeFromCallStack() reflect.Type {
	// 遍历调用栈，寻找可能的类型信息
	pc := make([]uintptr, 30)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])

	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		funcName := frame.Function
		if funcName == "" {
			continue
		}

		// 跳过框架内部函数
		if c.isFrameworkInternalFunction(funcName) {
			continue
		}

		// 尝试从函数名中提取Controller类型信息
		if strings.Contains(funcName, "Controller") {
			// 解析函数名，例如: main.(*HomeController).method
			if controllerTypeName := c.extractControllerTypeNameFromFunc(funcName); controllerTypeName != "" {
				// 尝试通过名称查找已知的类型
				// 这需要一个类型注册机制
				return nil // 暂时返回nil，让调用者使用其他方法
			}
		}
	}

	return nil
}

// 全局Controller类型注册机制
var (
	controllerTypeRegistry = make(map[reflect.Type]string)
	registryMutex          sync.RWMutex
)

// RegisterControllerType 注册Controller类型（包级别函数）
func RegisterControllerType(controllerType reflect.Type, name string) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	controllerTypeRegistry[controllerType] = name
}

// GetRegisteredControllerName 获取注册的Controller名称
func GetRegisteredControllerName(controllerType reflect.Type) (string, bool) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()
	name, exists := controllerTypeRegistry[controllerType]
	return name, exists
}

// isFrameworkInternalFunction 判断是否为框架内部函数
func (c *BaseController) isFrameworkInternalFunction(funcName string) bool {
	internalFunctions := []string{
		"autoInitializeControllerName",
		"detectActualControllerType",
		"detectFromCallStack",
		"detectFromReflection",
		"isFrameworkInternalFunction",
		"extractControllerNameFromFuncName",
		"tryInferControllerTypeFromAddress",
		"detectFromMainContext",
		"detectControllerTypeByReflection",
		"BaseController",
		"GetControllerName",
		"ensureInitialized",
	}

	for _, internal := range internalFunctions {
		if strings.Contains(funcName, internal) {
			return true
		}
	}
	return false
}

// extractControllerNameFromFuncName 从函数名中提取Controller名称
func (c *BaseController) extractControllerNameFromFuncName(funcName string) string {
	// 查找各种可能的Controller函数名模式
	patterns := []string{
		// 模式1: main.(*HomeController).method
		// 模式2: package.(*HomeController).method
		// 模式3: main.HomeController.method
	}

	_ = patterns // 避免未使用变量警告

	if strings.Contains(funcName, "Controller") {
		parts := strings.Split(funcName, ".")
		for _, part := range parts {
			// 处理 (*ControllerName) 格式
			if strings.HasPrefix(part, "(*") && strings.HasSuffix(part, ")") && strings.Contains(part, "Controller") {
				controllerType := strings.Trim(part, "(*)")
				if strings.HasSuffix(controllerType, "Controller") {
					return controllerType[:len(controllerType)-10] // 去掉"Controller"后缀
				}
			}
			// 处理 ControllerName 格式（不带指针）
			if strings.HasSuffix(part, "Controller") && !strings.Contains(part, "(") {
				return part[:len(part)-10] // 去掉"Controller"后缀
			}
		}
	}

	return ""
}

// extractControllerTypeNameFromFunc 从函数名提取Controller类型名称
func (c *BaseController) extractControllerTypeNameFromFunc(funcName string) string {
	// 解析各种可能的函数名格式
	if strings.Contains(funcName, "Controller") {
		parts := strings.Split(funcName, ".")
		for _, part := range parts {
			if strings.Contains(part, "Controller") && strings.Contains(part, "(") {
				// 提取(*HomeController)中的HomeController
				clean := strings.Trim(part, "(*)")
				if strings.HasSuffix(clean, "Controller") {
					return clean
				}
			}
		}
	}
	return ""
}

// detectTypeFromRuntimeInfo 从运行时信息检测类型
func (c *BaseController) detectTypeFromRuntimeInfo() string {
	// 新思路：检查当前goroutine的调用栈信息
	// 寻找可能的Controller实例化或方法调用信息

	pc := make([]uintptr, 50) // 增加调用栈深度
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])

	// 遍历整个调用栈，寻找Controller相关的信息
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		funcName := frame.Function
		if funcName == "" {
			continue
		}

		// 跳过所有框架内部函数
		if c.isFrameworkInternalFunction(funcName) {
			continue
		}

		// 专门查找main函数中的Controller实例化
		if strings.Contains(funcName, "main.main") {
			// 这表示在main函数中直接实例化
			// 我们需要用不同的策略
			continue
		}

		// 查找任何包含Controller的函数调用
		if strings.Contains(funcName, "Controller") {
			if extracted := c.extractControllerNameFromFuncName(funcName); extracted != "" {
				return extracted
			}
		}

		// 查找可能的包名+Controller模式
		// 例如: github.com/xxx/controllers.(*HomeController).method
		if strings.Contains(funcName, "controllers.") {
			parts := strings.Split(funcName, "controllers.")
			if len(parts) > 1 {
				remaining := parts[1]
				if controllerName := c.extractControllerNameFromString(remaining); controllerName != "" {
					return controllerName
				}
			}
		}
	}

	// 如果调用栈检测失败，尝试最后的绝招
	return c.lastResortDetection()
}

// extractControllerNameFromString 从字符串中提取Controller名称
func (c *BaseController) extractControllerNameFromString(s string) string {
	// 处理各种可能的格式
	// (*HomeController).method -> Home
	// HomeController.method -> Home

	if strings.Contains(s, "Controller") {
		// 找到Controller关键字的位置
		parts := strings.Split(s, ".")
		for _, part := range parts {
			if strings.Contains(part, "Controller") {
				// 清理(*HomeController)格式
				clean := strings.Trim(part, "(*)")
				if strings.HasSuffix(clean, "Controller") {
					return clean[:len(clean)-10] // 去掉"Controller"后缀
				}
			}
		}
	}
	return ""
}

// lastResortDetection 最后的检测方法
func (c *BaseController) lastResortDetection() string {
	// 在所有方法都失败后，我们尝试一个极其巧妙的方法：
	// 利用Go运行时的类型信息系统

	// 获取当前BaseController实例的地址
	baseAddr := uintptr(unsafe.Pointer(c))
	_ = baseAddr // 避免未使用警告

	// 由于Go的安全限制，我们无法直接从内存地址推断类型
	// 但我们可以尝试检查是否有其他上下文线索

	// 如果所有检测都失败，我们需要让用户知道
	// 返回空字符串，让调用者处理
	return ""
}

// detectControllerNameFromStack 通过调用栈检测控制器名称（改进版）
func (c *BaseController) detectControllerNameFromStack() string {
	// 首先尝试反射检测，这是最可靠的方法
	reflectionName := c.detectControllerNameFromReflection()
	if reflectionName != "BaseController" && reflectionName != "Base" {
		return reflectionName
	}

	// 如果反射检测失败，再尝试调用栈检测
	pc := make([]uintptr, 25)
	n := runtime.Callers(1, pc) // 从当前函数开始获取
	frames := runtime.CallersFrames(pc[:n])

	// 遍历调用栈，找到控制器方法
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		// 解析函数名
		funcName := frame.Function
		if funcName == "" {
			continue
		}

		// 查找控制器类型名称
		// 格式: main.(*DebugAutoController).GetIndex
		if strings.Contains(funcName, "Controller") {
			// 跳过框架内部的Controller相关方法
			if strings.Contains(funcName, "BaseController") {
				continue
			}

			// 跳过一些明显的框架内部方法
			if strings.Contains(funcName, "detectControllerNameFromStack") ||
				strings.Contains(funcName, "ensureInitialized") ||
				strings.Contains(funcName, "autoInitializeControllerInfo") {
				continue
			}

			// 查找包含Controller的部分
			parts := strings.Split(funcName, ".")
			for _, part := range parts {
				// 处理 (*ControllerName) 格式
				if strings.Contains(part, "Controller") {
					// 去掉括号和指针符号
					controllerType := strings.Trim(part, "(*)")
					if strings.HasSuffix(controllerType, "Controller") {
						name := controllerType[:len(controllerType)-10] // 去掉"Controller"
						return name                                     // 保持原始大小写，不转换为小写
					}
				}
			}
		}
	}

	// 最后才返回反射检测的结果
	return reflectionName
}

// detectControllerNameFromReflection 通过反射检测控制器名称（增强版）
func (c *BaseController) detectControllerNameFromReflection() string {
	// 方法1: 优先通过AppController获取实际类型
	if c.AppController != nil {
		actualType := reflect.TypeOf(c.AppController)
		if actualType.Kind() == reflect.Ptr {
			actualType = actualType.Elem()
		}

		// 获取类型名称
		typeName := actualType.Name()

		// 移除Controller后缀，保持原始大小写
		if strings.HasSuffix(typeName, "Controller") {
			typeName = typeName[:len(typeName)-10]
		}

		return typeName
	}

	// 方法2: 通过调用栈分析获取实际类型信息
	// 这是一个增强的反射方法，结合调用栈信息
	pc := make([]uintptr, 15)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])

	// 查找最近的控制器相关调用
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		funcName := frame.Function
		if funcName == "" {
			continue
		}

		// 查找控制器类型名称，跳过BaseController
		if strings.Contains(funcName, "Controller") && !strings.Contains(funcName, "BaseController") {
			// 跳过框架内部方法
			if strings.Contains(funcName, "detectControllerNameFromReflection") ||
				strings.Contains(funcName, "detectControllerNameFromStack") ||
				strings.Contains(funcName, "ensureInitialized") {
				continue
			}

			// 解析函数名获取控制器类型
			// 格式: main.(*TestController).MethodName 或 package.(*TestController).MethodName
			parts := strings.Split(funcName, ".")
			for _, part := range parts {
				if strings.Contains(part, "Controller") && strings.Contains(part, "(") {
					// 提取控制器名称: (*TestController) -> TestController
					controllerType := strings.Trim(part, "(*)")
					if strings.HasSuffix(controllerType, "Controller") {
						name := controllerType[:len(controllerType)-10] // 去掉"Controller"
						return name                                     // 保持原始大小写
					}
				}
			}
		}
	}

	// 方法3: 最后才直接反射当前实例（这通常会返回BaseController）
	actualType := reflect.TypeOf(c)
	if actualType.Kind() == reflect.Ptr {
		actualType = actualType.Elem()
	}

	// 获取类型名称
	typeName := actualType.Name()

	// 移除Controller后缀，保持原始大小写
	if strings.HasSuffix(typeName, "Controller") {
		typeName = typeName[:len(typeName)-10]
	}

	return typeName
}

// detectCurrentAction 通过调用栈自动检测当前执行的动作名称（修复版）
func (c *BaseController) detectCurrentAction() string {
	// 获取调用栈信息
	pc := make([]uintptr, 20)
	n := runtime.Callers(1, pc) // 从当前函数开始获取
	frames := runtime.CallersFrames(pc[:n])

	// 遍历调用栈，找到控制器方法
	for {
		frame, more := frames.Next()
		if !more {
			break
		}

		// 解析函数名
		funcName := frame.Function
		if funcName == "" {
			continue
		}

		// 提取方法名（去掉包名和接收器信息）
		parts := strings.Split(funcName, ".")
		if len(parts) < 2 {
			continue
		}

		methodName := parts[len(parts)-1]

		// 检查是否是控制器方法（不是Init、Prepare等生命周期方法）
		if c.isControllerAction(methodName) {
			// 提取动作名称
			return ExtractActionName(methodName)
		}
	}

	// 如果无法检测到，返回默认值
	return "index"
}

// isControllerAction 判断是否是控制器的业务动作方法
func (c *BaseController) isControllerAction(methodName string) bool {
	// 必须是公共方法（首字母大写）
	if len(methodName) == 0 || methodName[0] < 'A' || methodName[0] > 'Z' {
		return false
	}

	// 不能是保留方法
	if ReservedMethods[methodName] {
		return false
	}

	// 不能是生命周期方法和内部方法
	lifecycleMethods := map[string]bool{
		"Init": true, "Prepare": true, "Finish": true,
		"autoInitializeControllerInfo":  true,
		"detectCurrentAction":           true,
		"detectActualController":        true,
		"detectControllerNameFromStack": true,
		"isControllerAction":            true,
		"ensureInitialized":             true,
		// 其他可能的内部方法
		"ServeHTTP": true, "ServeJSON": true, "ServeXML": true,
	}

	if lifecycleMethods[methodName] {
		return false
	}

	// 必须是HTTP方法前缀的业务方法
	httpPrefixes := []string{"Get", "Post", "Put", "Delete", "Patch", "Head", "Options"}
	for _, prefix := range httpPrefixes {
		if strings.HasPrefix(methodName, prefix) {
			return true
		}
	}

	return false
}
