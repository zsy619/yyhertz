package core

import (
	"fmt"
	"reflect"
	"strings"
)

// ============= 路由和映射方法 =============

// AddMethodMapping 添加HTTP方法映射
func (c *BaseController) AddMethodMapping(httpMethod, controllerMethod string) {
	if c.MethodMapping == nil {
		c.MethodMapping = make(map[string]string)
	}
	c.MethodMapping[strings.ToUpper(httpMethod)] = controllerMethod
}

// GetMethodMapping 获取方法映射（自动初始化）
func (c *BaseController) GetMethodMapping() map[string]string {
	c.ensureInitialized()
	return c.MethodMapping
}

// SetMethodMapping 设置完整的方法映射
func (c *BaseController) SetMethodMapping(mapping map[string]string) {
	c.MethodMapping = mapping
}

// GetMappedMethod 根据HTTP方法获取对应的控制器方法名
func (c *BaseController) GetMappedMethod(httpMethod string) string {
	if c.MethodMapping == nil {
		return ""
	}
	return c.MethodMapping[strings.ToUpper(httpMethod)]
}

// SetRoutePattern 设置路由模式
func (c *BaseController) SetRoutePattern(pattern string) {
	c.RoutePattern = pattern
}

// GetRoutePattern 获取路由模式
func (c *BaseController) GetRoutePattern() string {
	return c.RoutePattern
}

// SetRouteParam 设置路由参数
func (c *BaseController) SetRouteParam(key, value string) {
	if c.RouteParams == nil {
		c.RouteParams = make(map[string]string)
	}
	c.RouteParams[key] = value
}

// GetRouteParam 获取路由参数
func (c *BaseController) GetRouteParam(key string) string {
	if c.RouteParams == nil {
		return ""
	}
	return c.RouteParams[key]
}

// GetRouteParams 获取所有路由参数
func (c *BaseController) GetRouteParams() map[string]string {
	return c.RouteParams
}

// SetRouteParams 设置所有路由参数
func (c *BaseController) SetRouteParams(params map[string]string) {
	c.RouteParams = params
}

// URLFor 生成URL（Beego兼容）
func (c *BaseController) URLFor(endpoint string, values ...any) string {
	if c.URLGenerator != nil {
		return c.URLGenerator(endpoint, values...)
	}

	// 默认简单实现
	result := endpoint
	for i, value := range values {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// BuildURL 构建URL
func (c *BaseController) BuildURL(controller, action string, params ...any) string {
	url := "/" + controller + "/" + action
	if len(params) > 0 {
		url += "?"
		for i := 0; i < len(params); i += 2 {
			if i > 0 {
				url += "&"
			}
			key := fmt.Sprintf("%v", params[i])
			value := ""
			if i+1 < len(params) {
				value = fmt.Sprintf("%v", params[i+1])
			}
			url += key + "=" + value
		}
	}
	return url
}

// ============= URL映射和处理器方法 =============

// URLMapping 注册URL映射（Beego ControllerInterface兼容）
func (c *BaseController) URLMapping() {
	// 默认实现：自动生成基于方法名的URL映射
	if c.URLMappings == nil {
		c.URLMappings = make(map[string]string)
	}
	
	// 获取控制器的所有公共方法
	if c.AppController != nil {
		controllerType := reflect.TypeOf(c.AppController)
		if controllerType.Kind() == reflect.Ptr {
			controllerType = controllerType.Elem()
		}
		
		for i := 0; i < controllerType.NumMethod(); i++ {
			method := controllerType.Method(i)
			methodName := method.Name
			
			// 跳过保留方法
			if ReservedMethods[methodName] {
				continue
			}
			
			// 根据方法名生成URL模式
			if pattern := c.generateURLPattern(methodName); pattern != "" {
				c.URLMappings[pattern] = methodName
			}
		}
	}
}

// AddURLMapping 添加URL映射
func (c *BaseController) AddURLMapping(pattern, method string) {
	if c.URLMappings == nil {
		c.URLMappings = make(map[string]string)
	}
	c.URLMappings[pattern] = method
}

// GetURLMappings 获取所有URL映射
func (c *BaseController) GetURLMappings() map[string]string {
	if c.URLMappings == nil {
		c.URLMapping() // 自动初始化
	}
	return c.URLMappings
}

// HandlerFunc 检查处理器函数是否存在（Beego ControllerInterface兼容）
func (c *BaseController) HandlerFunc(fn string) bool {
	if c.HandlerFuncs == nil {
		// 初始化时自动检测所有可用的处理器函数
		c.initializeHandlerFuncs()
	}
	
	return c.HandlerFuncs[fn]
}

// ============= 内部辅助方法 =============

// generateURLPattern 根据方法名生成URL模式
func (c *BaseController) generateURLPattern(methodName string) string {
	// 移除HTTP方法前缀
	actionName := methodName
	httpPrefixes := []string{"Get", "Post", "Put", "Delete", "Patch", "Head", "Options"}
	
	for _, prefix := range httpPrefixes {
		if strings.HasPrefix(methodName, prefix) {
			actionName = methodName[len(prefix):]
			break
		}
	}
	
	if actionName == "" || actionName == "Index" {
		return "/"
	}
	
	// 转换为小写并添加斜杠
	return "/" + strings.ToLower(actionName)
}

// initializeHandlerFuncs 初始化处理器函数映射
func (c *BaseController) initializeHandlerFuncs() {
	c.HandlerFuncs = make(map[string]bool)
	
	if c.AppController != nil {
		controllerType := reflect.TypeOf(c.AppController)
		if controllerType.Kind() == reflect.Ptr {
			controllerType = controllerType.Elem()
		}
		
		for i := 0; i < controllerType.NumMethod(); i++ {
			method := controllerType.Method(i)
			methodName := method.Name
			
			// 只包含公共方法且不是保留方法
			if methodName[0] >= 'A' && methodName[0] <= 'Z' && !ReservedMethods[methodName] {
				c.HandlerFuncs[methodName] = true
			}
		}
	}
}