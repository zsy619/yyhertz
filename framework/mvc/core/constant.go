package core

// ReservedMethods 定义需要跳过的BaseController和生命周期方法（导出为公共变量）
var ReservedMethods = map[string]bool{
	// ============= 生命周期方法 =============
	"Init": true, "Prepare": true, "Finish": true,

	// ============= 基础响应方法 =============
	"JSON": true, "String": true,
	"JSONWithStatus": true, "StringWithStatus": true,
	"JSONOK": true, "JSONError": true, "JSONSuccess": true,
	"JSONPage": true, "JSONStatus": true,
	"Redirect": true, "Error": true,

	// ============= 模板渲染方法 =============
	"Render": true, "RenderHTML": true, "RenderWithLayout": true,
	"RenderBytes": true, "RenderString": true, "RenderWithViewName": true,
	"RenderTemplate": true, "RenderTemplateComponent": true, "RenderTemplateWithLayout": true,

	// ============= 模板配置方法 =============
	"SetTplName": true, "GetTplName": true, "SetLayout": true, "GetLayout": true,
	"AddTplFunc": true, "GetTemplateManager": true, "SetTemplatePath": true,
	"SetTemplateTheme": true, "GetTemplateTheme": true, "AddTemplateFunction": true,

	// ============= Cookie操作方法 =============
	"SetCookie": true, "GetCookie": true, "DeleteCookie": true, "HasCookie": true,
	"SetSecureCookie": true, "GetSecureCookie": true,

	// ============= Session操作方法 =============
	"SetSession": true, "GetSession": true, "DeleteSession": true,
	"HasSession": true, "GetSessionID": true,

	// ============= 请求参数获取方法 =============
	"GetForm": true, "GetQuery": true, "GetParam": true, "GetString": true,
	"GetInt": true, "GetBool": true, "GetFloat": true,
	"GetHeader": true, "GetUserAgent": true, "GetClientIP": true,

	// ============= HTTP方法判断 =============
	"IsAjax": true, "IsMethod": true, "IsGet": true, "IsPost": true,
	"IsPut": true, "IsDelete": true, "IsPatch": true, "IsHead": true, "IsOptions": true,

	// ============= 数据操作方法 =============
	"SetData": true, "GetData": true, "DelData": true,

	// ============= 控制器管理方法 =============
	"SetControllerName": true, "GetControllerName": true,
	"SetActionName": true, "GetActionName": true,
	"SetAppController": true, "GetAppController": true,
	"SetControllerInstance": true, "GetControllerAndAction": true, "SetControllerAndAction": true,
	"IsValidAction": true, "GetAvailableActions": true, "AutoInit": true,

	// ============= 路由和映射方法 =============
	"AddMethodMapping": true, "GetMethodMapping": true, "SetMethodMapping": true, "GetMappedMethod": true,
	"SetRoutePattern": true, "GetRoutePattern": true, "SetRouteParam": true, "GetRouteParam": true,
	"GetRouteParams": true, "SetRouteParams": true,
	"URLFor": true, "BuildURL": true,

	// ============= XSRF/CSRF安全方法 =============
	"XSRFToken": true, "CheckXSRFCookie": true, "EnableXSRF": true, "DisableXSRF": true,

	// ============= URL映射和处理器方法 =============
	"URLMapping": true, "HandlerFunc": true, "AddURLMapping": true, "GetURLMappings": true,

	// ============= 流程控制方法 =============
	"StopRun": true, "Abort": true, "CustomAbort": true,

	// ============= 日志方法 =============
	"LogInfo": true, "LogInfof": true, "LogError": true, "LogErrorf": true,
	"LogWarn": true, "LogDebug": true, "LogDebugf": true,
	"LogFetal": true, "LogFetalf": true, "LogPanic": true, "LogPanicsf": true,
}

var ControllerNameSuffixReserved = map[string]bool{
	"Controller": true,
	"Ctrl":       true,
	"Handler":    true,
}
