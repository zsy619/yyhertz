package core

// ===========================================================================
// 常量定义文件
// ===========================================================================
//
// 本文件包含了 MVC 框架核心模块的常量定义，主要包括：
//
// 1. ReservedMethods: BaseController 保留方法列表
//    - 这些方法不应该被暴露为自动路由端点
//    - 包含生命周期方法、框架内部方法、工具方法等
//    - 基于实际扫描的 432 个 BaseController 方法重建，确保完整性
//
// 2. ControllerNameSuffixReserved: 控制器名称后缀保留字
//    - 用于控制器名称的规范化处理
//    - 在路由注册时会被自动移除
//
// 分组说明：
// - 控制器生命周期：框架管理的控制器初始化、准备、完成、销毁方法
// - 控制器管理：控制器元信息和实例管理相关方法
// - 响应输出：各种格式的数据输出和响应处理方法
// - 请求处理：HTTP请求解析、参数获取、数据绑定等方法
// - 文件处理：文件上传、下载、验证等文件操作方法
// - 缓存和性能：HTTP缓存控制、性能优化相关方法
// - 安全防护：CSRF防护、输入验证、加密等安全相关方法
// - 功能模块：国际化、日志、调试、邮件、队列等功能性方法
//
// 维护说明：
// - 当 BaseController 添加新的公共方法时，需要同步更新 ReservedMethods
// - 所有框架内部方法都应该添加到相应的功能分类中
// - 保持分类清晰，便于维护和理解
// - 新增方法应该根据功能放入对应的分组中
//
// ===========================================================================

// ReservedMethods 定义需要跳过的BaseController和生命周期方法（导出为公共变量）
// 基于实际扫描的432个BaseController方法重建，确保完整性和准确性
// 这些方法应该被排除在自动路由注册之外，避免框架内部方法被暴露为路由端点
var ReservedMethods = map[string]bool{
	// ============= 控制器生命周期方法 =============
	"Init": true, "Prepare": true, "Finish": true, "Destroy": true,
	"Reset": true, "InitWithContext": true, "AutoInit": true,

	// ============= 控制器管理和元信息方法 =============
	"SetControllerName": true, "GetControllerName": true,
	"SetActionName": true, "GetActionName": true,
	"SetAppController": true, "GetAppController": true,
	"SetControllerInstance": true, "GetControllerAndAction": true,
	"SetControllerAndAction": true, "IsValidAction": true,
	"GetAvailableActions": true,

	// ============= 基础响应输出方法 =============
	"JSON": true, "JSONWithStatus": true, "JSONOK": true, "JSONError": true,
	"JSONSuccess": true, "JSONPage": true, "JSONStatus": true,
	"String": true, "StringWithStatus": true, "Write": true,
	"Abort": true, "CustomAbort": true, "StopRun": true, "Error": true,

	// ============= 高级响应格式方法 =============
	"XML": true, "XMLWithStatus": true, "YAML": true, "YAMLWithStatus": true,
	"IndentedJSON": true, "IndentedJSONWithStatus": true,
	"JSONP": true, "JSONPWithStatus": true,
	"DataWithStatus": true, "Status": true, "NoContent": true,
	"EmptyResponse": true, "RawResponse": true,

	// ============= 流式响应和文件服务方法 =============
	"Stream": true, "StreamFile": true, "StreamReader": true,
	"File": true, "FileAttachment": true, "Inline": true,
	"ServeFile": true, "Download": true,

	// ============= 重定向方法 =============
	"Redirect": true, "RedirectPermanent": true, "RedirectTemporary": true,
	"RedirectSeeOther": true,

	// ============= 响应状态判断方法 =============
	"IsOk": true, "IsSuccessful": true, "IsRedirect": true,
	"IsClientError": true, "IsServerError": true, "IsForbidden": true,

	// ============= HTTP头部操作方法 =============
	"AddHeader": true, "SetHeader": true, "GetResponseHeader": true,
	"SetResponseHeader": true, "SetContentType": true,

	// ============= 模板渲染方法 =============
	"Render": true, "RenderHTML": true, "RenderWithLayout": true,
	"RenderBytes": true, "RenderString": true, "RenderWithViewName": true,
	"RenderTemplate": true, "RenderTemplateComponent": true,
	"RenderTemplateWithLayout": true, "RenderHTMLWithIncludes": true,

	// ============= 模板配置和管理方法 =============
	"SetTplName": true, "GetTplName": true, "SetLayout": true, "GetLayout": true,
	"AddTplFunc": true, "GetTemplateManager": true, "SetTemplatePath": true,
	"SetTemplateTheme": true, "GetTemplateTheme": true, "AddTemplateFunction": true,
	"AddBeegoTemplateFunctions": true, "CreateTemplateDefinition": true,
	"ListAvailableTemplates": true, "SetTemplateIncludeEngine": true,
	"GetTemplateIncludeEngine": true,

	// ============= Cookie操作方法 =============
	"SetCookie": true, "GetCookie": true, "DeleteCookie": true,
	"HasCookie": true, "SetSecureCookie": true, "GetSecureCookie": true,

	// ============= Session管理方法 =============
	"SetSession": true, "GetSession": true, "DeleteSession": true,
	"DestroySession": true, "HasSession": true, "GetSessionID": true,
	"RegenerateSessionID": true,

	// ============= 基础请求参数获取方法 =============
	"GetString": true, "GetInt": true, "GetInt32": true, "GetInt64": true,
	"GetBool": true, "GetFloat": true, "GetForm": true, "GetQuery": true,
	"GetParam": true, "GetHeader": true, "GetUserAgent": true, "GetClientIP": true,
	"GetStringTrim": true, "GetSafeString": true,

	// ============= 扩展请求参数获取方法 =============
	"GetFormValue": true, "GetFormValueDefault": true, "GetPostFormInt": true,
	"GetPostFormFloat64": true, "GetPostFormBool": true, "GetParamNames": true,
	"GetParamValues": true, "GetAllParams": true, "GetParamInt": true,
	"GetParamInt64": true, "GetParamFloat64": true, "GetParamBool": true,
	"GetParamDefault": true, "GetQueryAll": true, "GetQueryMap": true,
	"GetQueryDefault": true, "GetQueryInt64": true, "GetQueryFloat64": true,
	"GetQueryBool": true,

	// ============= Tuple参数获取方法 =============
	// Integer Tuples
	"GetIntTuple2": true, "GetIntTuple3": true, "GetIntTuple4": true,
	"GetIntTuple5": true, "GetIntTuple6": true, "GetIntTuple7": true,
	"GetIntTuple8": true, "GetIntTuple9": true,
	// Int32 Tuples
	"GetInt32Tuple2": true, "GetInt32Tuple3": true, "GetInt32Tuple4": true,
	"GetInt32Tuple5": true, "GetInt32Tuple6": true, "GetInt32Tuple7": true,
	"GetInt32Tuple8": true, "GetInt32Tuple9": true,
	// Int64 Tuples
	"GetInt64Tuple2": true, "GetInt64Tuple3": true, "GetInt64Tuple4": true,
	"GetInt64Tuple5": true, "GetInt64Tuple6": true, "GetInt64Tuple7": true,
	"GetInt64Tuple8": true, "GetInt64Tuple9": true,
	// String Tuples
	"GetSafeStringTuple2": true, "GetSafeStringTuple3": true, "GetSafeStringTuple4": true,
	"GetSafeStringTuple5": true, "GetSafeStringTuple6": true, "GetSafeStringTuple7": true,
	"GetSafeStringTuple8": true, "GetSafeStringTuple9": true,

	// ============= HTTP请求类型判断方法 =============
	"IsAjax": true, "IsMethod": true, "IsGet": true, "IsPost": true,
	"IsPut": true, "IsDelete": true, "IsPatch": true, "IsHead": true,
	"IsOptions": true,

	// ============= 请求内容类型判断方法 =============
	"IsJSON": true, "IsXML": true, "IsForm": true, "IsMultipart": true,
	"IsUpload": true,

	// ============= 数据绑定和解析方法 =============
	"Bind": true, "ShouldBind": true, "BindJSON": true, "ShouldBindJSON": true,
	"BindXML": true, "BindQuery": true, "ShouldBindQuery": true,
	"BindAndValidate": true, "ShouldBindAndValidate": true,

	// ============= 高级请求处理方法 =============
	"GetRawBody": true, "GetRawData": true, "GetBodySize": true,
	"HasBody": true, "GetBodyString": true, "GetMultipartForm": true,
	"ParseMultipartForm": true, "SetMaxMemory": true,

	// ============= 文件上传和处理方法 =============
	"SaveUploadedFile": true, "GetFile": true, "HasFile": true,
	"GetFileSize": true, "GetFileName": true, "GetFileHeader": true,
	"GetFiles": true, "SaveMultipleFiles": true, "ValidateFileSize": true,
	"ValidateFileExtension": true, "ValidateFileType": true,
	"ValidateFileUpload": true, "SetFileResponseHeaders": true,
	"GetContentTypeByExtension": true, "AddFileAttachment": true,

	// ============= 数据操作方法 =============
	"SetData": true, "GetData": true, "DelData": true,

	// ============= 路由和URL构建方法 =============
	"AddMethodMapping": true, "GetMethodMapping": true, "SetMethodMapping": true,
	"GetMappedMethod": true, "SetRoutePattern": true, "GetRoutePattern": true,
	"SetRouteParam": true, "GetRouteParam": true, "GetRouteParams": true,
	"SetRouteParams": true, "URLFor": true, "BuildURL": true,
	"URLMapping": true, "HandlerFunc": true, "AddURLMapping": true,
	"GetURLMappings": true, "BuildLocalizedURL": true,

	// ============= 缓存控制方法 =============
	"SetETag": true, "SetLastModified": true, "SetCacheControl": true,
	"SetMaxAge": true, "SetNoCache": true, "SetPrivateCache": true,
	"SetPublicCache": true, "NotModified": true, "CheckIfNoneMatch": true,
	"CheckIfModifiedSince": true, "HandleConditionalRequest": true,
	"SetExpires": true, "SetExpiresFromNow": true, "SetCacheForMinutes": true,
	"SetCacheForHours": true, "SetCacheForDays": true, "SetImmutableCache": true,
	"SetShortCache": true, "SetMediumCache": true, "SetLongCache": true,
	"IsCacheableMethod": true, "ShouldCache": true, "GetClientCachePreference": true,
	"IsClientNoCacheRequest": true,

	// ============= 压缩和性能优化方法 =============
	"SetGzipResponse": true, "EnableGzipCompression": true,
	"GenerateContentHash": true, "GenerateVersionETag": true,
	"SetContentHashETag": true, "StartPerformanceTimer": true,
	"EndPerformanceTimer": true, "SetPerformanceHeaders": true,
	"AddPreloadLink": true, "AddPrefetchLink": true, "AddPreconnectLink": true,
	"ServerPush": true,

	// ============= 国际化和本地化方法 =============
	"SetLanguage": true, "GetLanguage": true, "GetDefaultLanguage": true,
	"DetectLanguageFromHeader": true, "T": true, "Translate": true,
	"SetLocale": true, "GetLocale": true, "FormatNumber": true,
	"FormatCurrency": true, "FormatDate": true, "FormatDateTime": true,
	"GetSupportedLanguages": true, "IsSupportedLanguage": true,
	"GetLanguageName": true, "SwitchLanguage": true, "SetLanguageCookie": true,
	"GetLanguageFromCookie": true, "GetCurrentURL": true,
	"GetLanguageDirection": true, "GetTranslationFile": true,

	// ============= 安全防护方法 =============
	"SetSecurityHeaders": true, "SetContentSecurityPolicy": true,
	"SetXFrameOptions": true, "SetXContentTypeOptions": true,
	"SetXXSSProtection": true, "SetReferrerPolicy": true,
	"SetPermissionsPolicy": true, "SetStrictTransportSecurity": true,
	"RequireHTTPS": true, "IsHTTPS": true, "CheckIPWhitelist": true,
	"CheckIPBlacklist": true, "RateLimitCheck": true,

	// ============= CSRF/XSRF防护方法 =============
	"GenerateCSRFToken": true, "GetCSRFToken": true, "ValidateCSRFToken": true,
	"RequireCSRFToken": true, "CSRFError": true, "XSRFToken": true,
	"CheckXSRFCookie": true, "EnableXSRF": true, "DisableXSRF": true,

	// ============= 密码和加密方法 =============
	"HashPassword": true, "VerifyPassword": true, "GenerateSalt": true,
	"GenerateSecureToken": true, "SetSecureSessionCookie": true,

	// ============= 输入验证和清理方法 =============
	"Validate": true, "ValidateStruct": true, "ValidateBatch": true,
	"ValidateRequired": true, "ValidateLength": true, "ValidateRange": true,
	"ValidatePattern": true, "ValidateIn": true, "ValidateEmailFormat": true,
	"ValidatePhoneFormat": true, "ValidateIDCardFormat": true,
	"ValidateURLFormat": true, "ValidateIPFormat": true,
	"ValidatePasswordStrength": true, "ValidateForm": true,
	"HasValidationErrors": true, "GetFirstValidationError": true,
	"SetValidationErrors": true, "RegisterValidator": true,
	"ApplyCustomValidator": true, "ValidateDateTime": true,
	"ValidateJSON": true, "ValidateNumeric": true, "ValidateInteger": true,
	"ValidateBoolean": true, "CreateValidationResult": true,
	"ReturnValidationResult": true, "ValidateEmail": true,
	"ValidatePhone": true, "ValidateURL": true, "ValidateIPAddress": true,
	"ValidateSessionTimeout": true, "SanitizeHTML": true, "SanitizeFilename": true,

	// ============= 日志记录方法 =============
	"LogInfo": true, "LogInfof": true, "LogError": true, "LogErrorf": true,
	"LogWarn": true, "LogDebug": true, "LogDebugf": true,
	"LogFetal": true, "LogFetalf": true, "LogPanic": true, "LogPanicsf": true,
	"LogValidationError": true, "LogSecurityEvent": true,
	"LogFailedLogin": true, "LogSuspiciousActivity": true,

	// ============= 调试和监控方法 =============
	"GetDebugInfo": true, "StartProfiler": true, "EndProfiler": true,
	"GetProfilerResult": true, "LogPerformance": true, "DumpRequest": true,
	"DumpStackTrace": true, "LogDebugError": true, "DebugJSON": true,
	"DebugHeaders": true, "IsDebugMode": true, "GenerateRequestID": true,
	"PrintDebugInfo": true, "HealthCheck": true,

	// ============= 断言和测试辅助方法 =============
	"Assert": true, "AssertNotNil": true, "AssertEqual": true,

	// ============= 指标收集方法 =============
	"RecordMetric": true, "IncrementCounter": true, "RecordTiming": true,
	"RecordGauge": true,

	// ============= 中间件和优化管理方法 =============
	"EnableOptimization": true, "DisableOptimization": true,
	"IsOptimizationEnabled": true, "GetMiddleware": true,
	"SetMiddleware": true, "AddMiddleware": true,

	// ============= 邮件发送方法 =============
	"SendMail": true, "SendSimpleMail": true, "SendHTMLMail": true,
	"SendMailWithAttachment": true, "SendTemplateMail": true,
	"SendBulkMail": true, "SendMailToList": true, "QueueMail": true,
	"ProcessMailQueue": true, "UpdateMailConfig": true,
	"TestMailConnection": true, "GetMailStats": true, "GetMailLog": true,
	"CreateAttachment": true, "FormatEmailAddress": true,

	// ============= 队列和任务调度方法 =============
	"Dispatch": true, "DispatchNow": true, "DispatchLater": true,
	"DispatchToQueue": true, "DispatchEmail": true, "DispatchNotification": true,
	"RegisterJobHandler": true, "GetJobHandler": true, "ProcessQueue": true,
	"GetQueueInfo": true, "GetAllQueues": true, "ClearQueue": true,
	"PurgeFailedJobs": true, "GetJob": true, "GetJobsByType": true,
	"GetJobsByStatus": true, "InitDefaultJobHandlers": true,
	"GetQueueMetrics": true, "ExportQueueData": true, "ImportQueueData": true,
}

// ControllerNameSuffixReserved 控制器名称后缀保留字（统一在此定义）
var ControllerNameSuffixReserved = map[string]bool{
	"Controller": true,
	"Ctrl":       true,
}

const AdminIdKey = "adminId"
