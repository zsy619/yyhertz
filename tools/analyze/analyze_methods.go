package main

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/zsy619/yyhertz/framework/mvc/core"
)

// 当前的ReservedMethods
var currentReservedMethods = map[string]bool{
	// 生命周期方法
	"Init": true, "Prepare": true, "Finish": true,
	// BaseController基础方法
	"SetData": true, "GetData": true, "DelData": true, "JSON": true, "String": true, "HTML": true,
	"Render": true, "RenderHTML": true, "RenderWithLayout": true, "RenderBytes": true, "RenderString": true,
	"Redirect": true, "Error": true, "SetCookie": true, "GetCookie": true, "DeleteCookie": true,
	"SetSession": true, "GetSession": true, "DeleteSession": true,
	"GetForm": true, "GetQuery": true, "GetParam": true, "GetString": true,
	"GetInt": true, "GetBool": true, "GetFloat": true, "GetUserAgent": true,
	"GetClientIP": true, "IsAjax": true, "IsMethod": true, "GetHeader": true,
	// Beego兼容方法
	"SetTplName": true, "GetTplName": true, "SetLayout": true, "GetLayout": true,
	"AddTplFunc": true, "renderTemplate": true, "renderBasicTemplate": true,
	"StopRun": true, "Abort": true, "CustomAbort": true,
	"SetSecureCookie": true, "GetSecureCookie": true, "encryptCookieValue": true, "decryptCookieValue": true,
	// 控制器管理方法
	"SetControllerName": true, "GetControllerName": true, "SetActionName": true, "GetActionName": true,
	"AddMethodMapping": true, "GetMethodMapping": true, "SetMethodMapping": true, "GetMappedMethod": true,
	"SetRoutePattern": true, "GetRoutePattern": true, "SetRouteParam": true, "GetRouteParam": true,
	"GetRouteParams": true, "SetRouteParams": true, "SetAppController": true, "GetAppController": true,
	"URLFor": true, "BuildURL": true, "GetControllerAndAction": true, "SetControllerAndAction": true,
	"IsValidAction": true, "GetAvailableActions": true, "AutoDetectAction": true,
	// BaseController验证方法
	"ValidateRequired": true, "ValidateEmail": true, "ValidateLength": true,
	"ValidateInt": true, "ValidateFloat": true,
	// BaseController日志方法
	"LogInfo": true, "LogInfof": true, "LogError": true, "LogErrorf": true,
	"LogWarn": true, "LogWarnf": true, "LogDebug": true, "LogDebugf": true,
	// BaseController其他工具方法
	"StringWithStatus": true, "JSONWithStatus": true, "JSONOk": true, "JSONError": true,
	"JSONSuccess": true, "JSONPage": true, "JSONStatus": true,
	"GetSessionID": true, "HasCookie": true, "HasSession": true, "IsGet": true,
	"IsPost": true, "IsPut": true, "IsDelete": true, "IsPatch": true, "IsHead": true,
	"IsOptions": true,
	// 一些BaseController特殊方法 (可能被误当作业务方法)
	"Clear": true, "ClearSession": true, "DestroySession": true,
}

func main() {
	fmt.Println("=== BaseController 方法分析 ===")
	
	// 创建BaseController实例
	controller := core.NewBaseController()
	
	// 获取所有公共方法
	controllerType := reflect.TypeOf(controller)
	
	var allMethods []string
	var missingMethods []string
	
	fmt.Printf("BaseController 类型: %s\n", controllerType)
	fmt.Printf("方法总数: %d\n\n", controllerType.NumMethod())
	
	// 收集所有公共方法
	for i := 0; i < controllerType.NumMethod(); i++ {
		method := controllerType.Method(i)
		methodName := method.Name
		
		// 只统计公共方法（首字母大写）
		if len(methodName) > 0 && methodName[0] >= 'A' && methodName[0] <= 'Z' {
			allMethods = append(allMethods, methodName)
			
			// 检查是否在ReservedMethods中
			if !currentReservedMethods[methodName] {
				missingMethods = append(missingMethods, methodName)
			}
		}
	}
	
	sort.Strings(allMethods)
	sort.Strings(missingMethods)
	
	fmt.Printf("所有公共方法 (%d个):\n", len(allMethods))
	for i, method := range allMethods {
		if i > 0 && i%5 == 0 {
			fmt.Println()
		}
		fmt.Printf("%-25s", method)
	}
	fmt.Println("\n")
	
	fmt.Printf("遗漏的方法 (%d个):\n", len(missingMethods))
	for _, method := range missingMethods {
		fmt.Printf("- %s\n", method)
	}
	
	// 按功能分组遗漏的方法
	fmt.Println("\n=== 按功能分组的遗漏方法 ===")
	categorizeMethod(missingMethods)
}

func categorizeMethod(methods []string) {
	categories := map[string][]string{
		"模板相关":     {},
		"XSRF安全":    {},
		"URL映射":     {},
		"控制器管理":    {},
		"日志相关":     {},
		"工具方法":     {},
		"内部方法":     {},
		"其他":       {},
	}
	
	for _, method := range methods {
		switch {
		case strings.Contains(method, "Template") || strings.Contains(method, "Render") || strings.Contains(method, "Tpl"):
			categories["模板相关"] = append(categories["模板相关"], method)
		case strings.Contains(method, "XSRF") || strings.Contains(method, "CSRF"):
			categories["XSRF安全"] = append(categories["XSRF安全"], method)
		case strings.Contains(method, "URL") || strings.Contains(method, "Mapping") || strings.Contains(method, "Handler"):
			categories["URL映射"] = append(categories["URL映射"], method)
		case strings.Contains(method, "Controller") || strings.Contains(method, "Action") || strings.Contains(method, "App"):
			categories["控制器管理"] = append(categories["控制器管理"], method)
		case strings.Contains(method, "Log") || strings.Contains(method, "Fatal") || strings.Contains(method, "Panic"):
			categories["日志相关"] = append(categories["日志相关"], method)
		case strings.HasPrefix(method, "detect") || strings.HasPrefix(method, "auto") || 
			 strings.HasPrefix(method, "ensure") || strings.HasPrefix(method, "initialize") ||
			 strings.Contains(method, "Internal") || strings.Contains(method, "Framework"):
			categories["内部方法"] = append(categories["内部方法"], method)
		case strings.Contains(method, "Get") || strings.Contains(method, "Set") || 
			 strings.Contains(method, "Add") || strings.Contains(method, "Enable") ||
			 strings.Contains(method, "Disable"):
			categories["工具方法"] = append(categories["工具方法"], method)
		default:
			categories["其他"] = append(categories["其他"], method)
		}
	}
	
	for category, methodList := range categories {
		if len(methodList) > 0 {
			fmt.Printf("\n【%s】(%d个):\n", category, len(methodList))
			for _, method := range methodList {
				fmt.Printf("  - %s\n", method)
			}
		}
	}
}