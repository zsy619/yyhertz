package comment

import (
	"strings"
	
	"github.com/zsy619/yyhertz/framework/mvc/routing"
)

// AutoDiscovery 自动发现和注册控制器
type AutoDiscovery struct {
	app              *App
	scanPaths        []string
	excludePaths     []string
	controllerSuffix string
}

// NewAutoDiscovery 创建自动发现器
func NewAutoDiscovery(app *App) *AutoDiscovery {
	return &AutoDiscovery{
		app:              app,
		scanPaths:        make([]string, 0),
		excludePaths:     make([]string, 0),
		controllerSuffix: "Controller",
	}
}

// WithScanPaths 设置扫描路径
func (ad *AutoDiscovery) WithScanPaths(paths ...string) *AutoDiscovery {
	ad.scanPaths = append(ad.scanPaths, paths...)
	return ad
}

// WithExcludePaths 设置排除路径
func (ad *AutoDiscovery) WithExcludePaths(paths ...string) *AutoDiscovery {
	ad.excludePaths = append(ad.excludePaths, paths...)
	return ad
}

// WithControllerSuffix 设置控制器后缀
func (ad *AutoDiscovery) WithControllerSuffix(suffix string) *AutoDiscovery {
	ad.controllerSuffix = suffix
	return ad
}

// Discover 执行自动发现
func (ad *AutoDiscovery) Discover() error {
	// 扫描指定路径
	for _, path := range ad.scanPaths {
		if ad.shouldExclude(path) {
			continue
		}
		
		err := GetGlobalParser().ScanPackage(path)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// shouldExclude 检查是否应该排除路径
func (ad *AutoDiscovery) shouldExclude(path string) bool {
	for _, excludePath := range ad.excludePaths {
		if path == excludePath {
			return true
		}
	}
	return false
}

// RouteCollector 路由收集器，用于收集和分析路由信息
type RouteCollector struct {
	routes []*RouteInfo
}

// NewRouteCollector 创建路由收集器
func NewRouteCollector() *RouteCollector {
	return &RouteCollector{
		routes: make([]*RouteInfo, 0),
	}
}

// CollectFromApp 从应用收集路由
func (rc *RouteCollector) CollectFromApp(app *App) *RouteCollector {
	routes := app.GetRoutes()
	rc.routes = append(rc.routes, routes...)
	return rc
}

// CollectFromGlobal 从全局注解收集路由
func (rc *RouteCollector) CollectFromGlobal() *RouteCollector {
	parser := GetGlobalParser()
	
	for _, controllerInfo := range parser.ControllerInfos {
		methods := parser.GetControllerMethods(controllerInfo.PackageName, controllerInfo.TypeName)
		
		for _, methodInfo := range methods {
			route := &RouteInfo{
				Path:        routing.CombinePath(controllerInfo.BasePath, methodInfo.Path),
				HTTPMethod:  methodInfo.HTTPMethod,
				PackageName: methodInfo.PackageName,
				TypeName:    methodInfo.TypeName,
				MethodName:  methodInfo.MethodName,
				Description: methodInfo.Description,
				Params:      methodInfo.Params,
				Middlewares: methodInfo.Middlewares,
			}
			rc.routes = append(rc.routes, route)
		}
	}
	
	return rc
}

// FilterByHTTPMethod 按HTTP方法过滤
func (rc *RouteCollector) FilterByHTTPMethod(method string) []*RouteInfo {
	var filtered []*RouteInfo
	for _, route := range rc.routes {
		if route.HTTPMethod == method {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

// FilterByPath 按路径过滤
func (rc *RouteCollector) FilterByPath(pathPattern string) []*RouteInfo {
	var filtered []*RouteInfo
	for _, route := range rc.routes {
		if route.Path == pathPattern {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

// FilterByController 按控制器过滤
func (rc *RouteCollector) FilterByController(typeName string) []*RouteInfo {
	var filtered []*RouteInfo
	for _, route := range rc.routes {
		if route.TypeName == typeName {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

// GetAllRoutes 获取所有路由
func (rc *RouteCollector) GetAllRoutes() []*RouteInfo {
	return rc.routes
}

// GetRouteCount 获取路由数量
func (rc *RouteCollector) GetRouteCount() int {
	return len(rc.routes)
}

// GetControllerCount 获取控制器数量
func (rc *RouteCollector) GetControllerCount() int {
	controllers := make(map[string]bool)
	for _, route := range rc.routes {
		key := route.PackageName + "." + route.TypeName
		controllers[key] = true
	}
	return len(controllers)
}

// GetMethodCount 获取方法数量统计
func (rc *RouteCollector) GetMethodCount() map[string]int {
	counts := make(map[string]int)
	for _, route := range rc.routes {
		counts[route.HTTPMethod]++
	}
	return counts
}

// RouteAnalyzer 路由分析器
type RouteAnalyzer struct {
	collector *RouteCollector
}

// NewRouteAnalyzer 创建路由分析器
func NewRouteAnalyzer(collector *RouteCollector) *RouteAnalyzer {
	return &RouteAnalyzer{
		collector: collector,
	}
}

// AnalyzeDuplicates 分析重复路由
func (ra *RouteAnalyzer) AnalyzeDuplicates() [][]string {
	routeMap := make(map[string][]string)
	
	for _, route := range ra.collector.routes {
		key := route.HTTPMethod + " " + route.Path
		controllerMethod := route.TypeName + "." + route.MethodName
		routeMap[key] = append(routeMap[key], controllerMethod)
	}
	
	var duplicates [][]string
	for key, methods := range routeMap {
		if len(methods) > 1 {
			duplicate := append([]string{key}, methods...)
			duplicates = append(duplicates, duplicate)
		}
	}
	
	return duplicates
}

// AnalyzeRESTfulness 分析RESTful风格
func (ra *RouteAnalyzer) AnalyzeRESTfulness() map[string][]string {
	restPatterns := make(map[string][]string)
	
	for _, route := range ra.collector.routes {
		// 简单的RESTful模式检测
		if route.HTTPMethod == "GET" && !strings.Contains(route.Path, "{") {
			restPatterns["LIST"] = append(restPatterns["LIST"], route.Path)
		} else if route.HTTPMethod == "GET" && strings.Contains(route.Path, "{") {
			restPatterns["GET"] = append(restPatterns["GET"], route.Path)
		} else if route.HTTPMethod == "POST" {
			restPatterns["CREATE"] = append(restPatterns["CREATE"], route.Path)
		} else if route.HTTPMethod == "PUT" || route.HTTPMethod == "PATCH" {
			restPatterns["UPDATE"] = append(restPatterns["UPDATE"], route.Path)
		} else if route.HTTPMethod == "DELETE" {
			restPatterns["DELETE"] = append(restPatterns["DELETE"], route.Path)
		}
	}
	
	return restPatterns
}

// GenerateRouteMap 生成路由映射表
func (ra *RouteAnalyzer) GenerateRouteMap() map[string]map[string]string {
	routeMap := make(map[string]map[string]string)
	
	for _, route := range ra.collector.routes {
		if routeMap[route.TypeName] == nil {
			routeMap[route.TypeName] = make(map[string]string)
		}
		
		key := route.HTTPMethod + " " + route.Path
		routeMap[route.TypeName][key] = route.MethodName
	}
	
	return routeMap
}