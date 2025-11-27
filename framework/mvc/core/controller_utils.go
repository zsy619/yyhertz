package core

import (
	"reflect"
	"strings"
	"sync"
)

// ============= 辅助函数 =============

// 方法缓存，提高性能
var methodCache = make(map[string][]string)
var cacheMutex = make(map[string]*sync.RWMutex)

// getControllerMethods 使用反射获取控制器的所有可用方法（带缓存）
func getControllerMethods(controller IController) []string {
	if controller == nil {
		return []string{}
	}

	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	typeName := controllerType.String()

	// 检查缓存
	if mutex, exists := cacheMutex[typeName]; exists {
		mutex.RLock()
		if cached, found := methodCache[typeName]; found {
			mutex.RUnlock()
			return cached
		}
		mutex.RUnlock()
	} else {
		cacheMutex[typeName] = &sync.RWMutex{}
	}

	// 获取写锁进行缓存更新
	mutex := cacheMutex[typeName]
	mutex.Lock()
	defer mutex.Unlock()

	// 双重检查，防止并发问题
	if cached, found := methodCache[typeName]; found {
		return cached
	}

	methods := []string{}
	controllerValue := reflect.ValueOf(controller)
	controllerTypeRef := controllerValue.Type()

	// 遍历所有方法
	numMethods := controllerTypeRef.NumMethod()
	for i := 0; i < numMethods; i++ {
		method := controllerTypeRef.Method(i)
		methodName := method.Name

		// 跳过非公共方法（首字母小写）
		if len(methodName) == 0 || methodName[0] < 'A' || methodName[0] > 'Z' {
			continue
		}

		// 跳过保留方法
		if ReservedMethods[methodName] {
			continue
		}

		// 检查方法签名是否合适（接收者+无参数，无返回值）
		methodType := method.Type
		if methodType.NumIn() == 1 && methodType.NumOut() == 0 {
			methods = append(methods, methodName)
		}
	}

	// 缓存结果
	methodCache[typeName] = methods
	return methods
}

// ExtractControllerName 从控制器类型中提取控制器名称（保持原始大小写）
func ExtractControllerName(controller IController) string {
	if controller == nil {
		return ""
	}

	controllerType := reflect.TypeOf(controller)
	if controllerType.Kind() == reflect.Ptr {
		controllerType = controllerType.Elem()
	}

	name := controllerType.Name()
	// 移除Controller等后缀，但保持原始大小写
	switch {
	case strings.HasSuffix(name, "Controller"):
		name = name[:len(name)-10]
	case strings.HasSuffix(name, "Ctrl"):
		name = name[:len(name)-4]
	case strings.HasSuffix(name, "Handler"):
		name = name[:len(name)-7]
	}

	// 不再转换为小写，保持原始大小写
	return name
}

// ExtractActionName 从方法名中提取动作名称（保持原始大小写）
func ExtractActionName(methodName string) string {
	if len(methodName) < 3 {
		return methodName // 保持原样，不转换
	}

	// 移除HTTP方法前缀
	prefixes := []string{"Get", "Post", "Put", "Delete", "Patch", "Head", "Options"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(methodName, prefix) {
			action := methodName[len(prefix):]
			if len(action) > 0 {
				// 保持原始大小写，不做任何转换
				return action
			}
		}
	}

	// 如果没有HTTP前缀，保持原样
	return methodName
}

// ValidateMethodMapping 验证方法映射的有效性
func ValidateMethodMapping(mapping map[string]string, controller IController) map[string]string {
	if mapping == nil || controller == nil {
		return mapping
	}

	availableMethods := getControllerMethods(controller)
	methodSet := make(map[string]bool)
	for _, method := range availableMethods {
		methodSet[method] = true
	}

	validMapping := make(map[string]string)
	for httpMethod, controllerMethod := range mapping {
		if methodSet[controllerMethod] {
			validMapping[httpMethod] = controllerMethod
		}
	}

	return validMapping
}

// CreateDefaultMethodMapping 创建默认的方法映射（改进版）
func CreateDefaultMethodMapping(controller IController) map[string]string {
	if controller == nil {
		return make(map[string]string)
	}

	mapping := make(map[string]string)
	methods := getControllerMethods(controller)

	// 优先级映射：优先选择Index类方法作为默认GET方法
	priorities := map[string][]string{
		"GET":     {"GetIndex", "GetList", "GetShow", "GetInfo"},
		"POST":    {"PostCreate", "PostStore", "PostSave"},
		"PUT":     {"PutUpdate", "PutModify", "PutEdit"},
		"DELETE":  {"DeleteRemove", "DeleteDestroy", "DeleteDelete"},
		"PATCH":   {"PatchUpdate", "PatchModify"},
		"HEAD":    {"HeadIndex", "HeadInfo"},
		"OPTIONS": {"OptionsIndex", "OptionsInfo"},
	}

	// 创建方法集合用于快速查找
	methodSet := make(map[string]bool)
	for _, method := range methods {
		methodSet[method] = true
	}

	// 按优先级分配映射
	for httpMethod, candidates := range priorities {
		for _, candidate := range candidates {
			if methodSet[candidate] {
				mapping[httpMethod] = candidate
				break // 找到第一个匹配的就停止
			}
		}
	}

	// 为剩余方法创建映射（没有被优先级覆盖的）
	for _, method := range methods {
		// 根据方法名前缀判断HTTP方法
		if strings.HasPrefix(method, "Get") && mapping["GET"] == "" {
			mapping["GET"] = method
		} else if strings.HasPrefix(method, "Post") && mapping["POST"] == "" {
			mapping["POST"] = method
		} else if strings.HasPrefix(method, "Put") && mapping["PUT"] == "" {
			mapping["PUT"] = method
		} else if strings.HasPrefix(method, "Delete") && mapping["DELETE"] == "" {
			mapping["DELETE"] = method
		} else if strings.HasPrefix(method, "Patch") && mapping["PATCH"] == "" {
			mapping["PATCH"] = method
		} else if strings.HasPrefix(method, "Head") && mapping["HEAD"] == "" {
			mapping["HEAD"] = method
		} else if strings.HasPrefix(method, "Options") && mapping["OPTIONS"] == "" {
			mapping["OPTIONS"] = method
		}
	}

	return mapping
}
