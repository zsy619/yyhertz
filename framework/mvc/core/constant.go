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
// 功能分组说明（共42个功能模块）：
// - 控制器生命周期：框架管理的控制器初始化、准备、完成、销毁方法
// - 控制器管理：控制器元信息和实例管理相关方法
// - 响应输出：各种格式的数据输出和响应处理方法（JSON、XML、文件流等）
// - 请求处理：HTTP请求解析、参数获取、数据绑定等方法
// - 文件处理：文件上传、下载、验证等文件操作方法
// - 模板引擎：模板渲染、配置管理、主题切换、Include引擎等
// - 缓存和性能：HTTP缓存控制、性能优化相关方法
// - 安全防护：CSRF防护、输入验证、加密、权限管理等安全相关方法
// - 功能模块：国际化、日志、调试、邮件、队列等功能性方法
// - 统一管理器：集成的统一管理器和框架扩展方法
// - WebSocket支持：WebSocket连接处理和实时通信方法
//
// 维护说明：
// - 当 BaseController 添加新的公共方法时，需要同步更新 ReservedMethods
// - 所有框架内部方法都应该添加到相应的功能分类中
// - 保持分类清晰，便于维护和理解
// - 新增方法应该根据功能放入对应的分组中
//
// ===========================================================================

// ReservedMethods 定义BaseController中需要跳过自动路由注册的所有公共方法
//
// 这些方法应该被排除在自动路由注册之外，因为它们是框架内部功能方法，
// 不应该作为HTTP路由端点暴露给外部访问。
//
// 基于实际扫描BaseController的所有公共方法重新构建，确保完整性和准确性。
// 总计包含 530+ 个方法（最后更新：2025-08-23，包含所有新增的统一管理器、认证、模板引擎增强等方法），按功能模块进行清晰分组管理。
//
// 维护说明：
// - 当BaseController添加新的公共方法时，必须同步更新此列表
// - 所有框架内部功能方法都应添加到对应的分类中
// - 保持分类清晰，便于维护和理解
var ReservedMethods = map[string]bool{
	// ========== 1. 控制器生命周期方法 ==========
	// 框架自动调用的生命周期管理方法
	"Init": true, "Prepare": true, "Finish": true, "Destroy": true,
	"Reset": true, "InitWithContext": true, "AutoInit": true,
	// 扩展生命周期方法（新增）
	"QuickInit": true, "InitWithName": true, "ResetExecutionState": true, "ShouldStopExecution": true,

	// ========== 2. 控制器管理和元信息方法 ==========
	// 控制器实例管理和元信息访问方法
	"SetControllerName": true, "GetControllerName": true, "ControllerName": true,
	"SetActionName": true, "GetActionName": true, "ActionName": true,
	"SetAppController": true, "GetAppController": true,
	"SetControllerInstance": true, "GetControllerAndAction": true,
	"SetControllerAndAction": true, "IsValidAction": true,
	"GetAvailableActions": true,

	// ========== 3. 基础响应输出方法 ==========
	// 基础HTTP响应和数据输出方法
	"JSON": true, "JSONWithStatus": true, "JSONOK": true, "JSONError": true,
	"JSONSuccess": true, "JSONPage": true, "JSONStatus": true,
	"String": true, "StringWithStatus": true, "Write": true,
	"Abort": true, "CustomAbort": true, "StopRun": true, "Error": true, "Errorf": true,
	// 扩展响应方法（新增）
	"AbortWithStatus": true,

	// ========== 4. 高级响应格式方法 ==========
	// 多种数据格式的响应输出方法
	"XML": true, "XMLWithStatus": true, "YAML": true, "YAMLWithStatus": true,
	"IndentedJSON": true, "IndentedJSONWithStatus": true,
	"JSONP": true, "JSONPWithStatus": true,
	"DataWithStatus": true, "Status": true, "NoContent": true,
	"EmptyResponse": true, "RawResponse": true, "ServeFormatted": true,
	"ServeJSON": true, "ServeJSONP": true, "ServeXML": true, "ServeYAML": true,

	// ========== 5. 流式响应和文件服务方法 ==========
	// 文件下载、流式输出和静态资源服务方法
	"Stream": true, "StreamFile": true, "StreamReader": true,
	"File": true, "FileAttachment": true, "Inline": true,
	"ServeFile": true, "Download": true,

	// ========== 6. 重定向方法 ==========
	// HTTP重定向相关方法
	"Redirect": true, "RedirectPermanent": true, "RedirectTemporary": true,
	"RedirectSeeOther": true,

	// ========== 7. 响应状态判断方法 ==========
	// HTTP响应状态码判断和检查方法
	"IsOk": true, "IsSuccessful": true, "IsRedirect": true,
	"IsClientError": true, "IsServerError": true, "IsForbidden": true,

	// ========== 8. HTTP头部操作方法 ==========
	// HTTP请求头和响应头操作方法
	"AddHeader": true, "SetHeader": true, "GetResponseHeader": true,
	"SetResponseHeader": true, "SetContentType": true, "GetHeader": true,

	// ========== 9. 模板渲染方法 ==========
	// 模板引擎和视图渲染相关方法
	"Render": true, "RenderHTML": true, "RenderWithLayout": true,
	"RenderBytes": true, "RenderString": true, "RenderWithViewName": true,
	"RenderTemplate": true, "RenderTemplateComponent": true,
	"RenderTemplateWithLayout": true,
	// 扩展模板渲染方法（新增）
	"RenderUnifiedTemplate": true, "RenderHTMLWithIncludes": true,

	// ========== 10. 模板配置和管理方法 ==========
	// 模板配置、主题管理和模板函数管理方法
	"SetTplName": true, "GetTplName": true, "SetLayout": true, "GetLayout": true,
	"AddTplFunc": true, "GetTemplateManager": true, "SetTemplatePath": true,
	"SetTemplateTheme": true, "GetTemplateTheme": true, "AddTemplateFunction": true,
	// 扩展模板管理方法（新增）
	"GetTemplateIncludeEngine": true, "SetTemplateIncludeEngine": true,
	"ListAvailableTemplates": true, "AddBeegoTemplateFunctions": true, "CreateTemplateDefinition": true,

	// ========== 11. Cookie操作方法 ==========
	// Cookie读写和管理方法
	"SetCookie": true, "GetCookie": true, "DeleteCookie": true,
	"HasCookie": true, "SetSecureCookie": true, "GetSecureCookie": true,

	// ========== 12. Session管理方法 ==========
	// 会话管理和状态维护方法
	"SetSession": true, "GetSession": true, "DeleteSession": true,
	"DestroySession": true, "HasSession": true, "GetSessionID": true,
	"RegenerateSessionID": true,

	// ========== 13. 基础请求参数获取方法 ==========
	// 基础HTTP参数解析和获取方法
	"GetString": true, "GetInt": true, "GetInt32": true, "GetInt64": true,
	"GetInt8": true, "GetInt16": true, "GetUint8": true, "GetUint16": true,
	"GetUint32": true, "GetUint64": true,
	"GetBool": true, "GetFloat": true, "GetFloat32": true, "GetFloat64": true,
	"GetForm": true, "GetQuery": true, "GetParam": true, "GetUserAgent": true,
	"GetClientIP": true, "GetStringTrim": true, "GetSafeString": true,
	"GetMap": true, "GetMapNoPathParams": true,
	"SaveToFile": true, "SaveToFileWithBuffer": true,
	"GetRequestBody": true, "GetRequestBodyString": true,
	"GetScheme": true, "GetHost": true, "GetPort": true,
	"GetReferer": true, "GetRemoteAddr": true, "GetRealIP": true,

	// ========== 14. 表单解析和多值参数方法 ==========
	// 表单数据解析和多值参数处理方法
	"ParseForm": true, "GetStrings": true, "GetFormStrings": true, "GetQueryStrings": true,
	"ParseFormToMap": true, "GetFormValues": true, "HasFormValue": true,

	// ========== 15. 扩展请求参数获取方法 ==========
	// 高级参数获取和类型转换方法
	"GetFormValue": true, "GetFormValueDefault": true, "GetPostFormInt": true,
	"GetPostFormFloat64": true, "GetPostFormBool": true, "GetParamNames": true,
	"GetParamValues": true, "GetAllParams": true, "GetParamInt": true,
	"GetParamInt64": true, "GetParamFloat64": true, "GetParamBool": true,
	"GetParamDefault": true, "GetQueryAll": true, "GetQueryMap": true,
	"GetQueryDefault": true, "GetQueryInt64": true, "GetQueryFloat64": true,
	"GetQueryBool": true,

	// ========== 16. Tuple参数获取方法 ==========
	// 批量参数获取的Tuple方法
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
	// Float32 Tuples
	"GetFloat32Tuple2": true, "GetFloat32Tuple3": true, "GetFloat32Tuple4": true,
	"GetFloat32Tuple5": true, "GetFloat32Tuple6": true, "GetFloat32Tuple7": true,
	"GetFloat32Tuple8": true, "GetFloat32Tuple9": true, "GetFloat32Tuple10": true,
	// Float64 Tuples (新增)
	"GetFloat64Tuple2": true, "GetFloat64Tuple3": true, "GetFloat64Tuple4": true,
	"GetFloat64Tuple5": true, "GetFloat64Tuple6": true, "GetFloat64Tuple7": true,
	"GetFloat64Tuple8": true, "GetFloat64Tuple9": true, "GetFloat64Tuple10": true,
	// String Tuples
	"GetSafeStringTuple2": true, "GetSafeStringTuple3": true, "GetSafeStringTuple4": true,
	"GetSafeStringTuple5": true, "GetSafeStringTuple6": true, "GetSafeStringTuple7": true,
	"GetSafeStringTuple8": true, "GetSafeStringTuple9": true,

	// ========== 17. HTTP请求类型判断方法 ==========
	// HTTP方法和AJAX请求判断方法
	"IsAjax": true, "IsMethod": true, "IsGet": true, "IsPost": true,
	"IsPut": true, "IsDelete": true, "IsPatch": true, "IsHead": true,
	"IsOptions": true,

	// ========== 18. 请求内容类型判断方法 ==========
	// Content-Type和MIME类型判断方法
	"IsJSON": true, "IsXML": true, "IsForm": true, "IsMultipart": true,
	"IsUpload": true,

	// ========== 19. 数据绑定和解析方法 ==========
	// 请求数据自动绑定到结构体的方法
	"Bind": true, "ShouldBind": true, "BindJSON": true, "ShouldBindJSON": true,
	"BindXML": true, "BindYAML": true,
	"BindForm": true, "BindProtobuf": true,
	"BindQuery": true, "ShouldBindQuery": true,
	"BindAndValidate": true, "ShouldBindAndValidate": true,

	// ========== 20. 高级请求处理方法 ==========
	// 原始请求体处理和多部分表单处理方法
	"GetRawBody": true, "GetRawData": true, "GetBodySize": true,
	"HasBody": true, "GetBodyString": true, "GetMultipartForm": true,
	"ParseMultipartForm": true, "SetMaxMemory": true,

	// ========== 21. 文件上传和处理方法 ==========
	// 文件上传、验证和处理相关方法
	"SaveUploadedFile": true, "GetFile": true, "HasFile": true,
	"GetFileSize": true, "GetFileName": true, "GetFileHeader": true,
	"GetFiles": true, "SaveMultipleFiles": true, "ValidateFileSize": true,
	"ValidateFileExtension": true, "ValidateFileType": true,
	"ValidateFileUpload": true, "SetFileResponseHeaders": true,
	"GetContentTypeByExtension": true, "AddFileAttachment": true,

	// ========== 22. 数据存储和操作方法 ==========
	// 控制器内部数据存储和管理方法
	"SetData": true, "GetData": true, "DelData": true,
	// 扩展数据管理方法（新增）
	"GetContextData": true, "SetContextData": true, "GetTypedContextData": true,

	// ========== 23. 路由和URL构建方法 ==========
	// 路由映射和URL生成相关方法
	"AddMethodMapping": true, "GetMethodMapping": true, "SetMethodMapping": true,
	"GetMappedMethod": true, "SetRoutePattern": true, "GetRoutePattern": true,
	"SetRouteParam": true, "GetRouteParam": true, "GetRouteParams": true,
	"SetRouteParams": true, "URLFor": true, "BuildURL": true,
	"URLMapping": true, "HandlerFunc": true, "AddURLMapping": true,
	"GetURLMappings": true, "BuildLocalizedURL": true,

	// ========== 24. HTTP缓存控制方法 ==========
	// HTTP缓存策略和缓存控制方法
	"SetETag": true, "SetLastModified": true, "SetCacheControl": true,
	"SetMaxAge": true, "SetNoCache": true, "SetPrivateCache": true,
	"SetPublicCache": true, "NotModified": true, "CheckIfNoneMatch": true,
	"CheckIfModifiedSince": true, "HandleConditionalRequest": true,
	"SetExpires": true, "SetExpiresFromNow": true, "SetCacheForMinutes": true,
	"SetCacheForHours": true, "SetCacheForDays": true, "SetImmutableCache": true,
	"SetShortCache": true, "SetMediumCache": true, "SetLongCache": true,
	"IsCacheableMethod": true, "ShouldCache": true, "GetClientCachePreference": true,
	"IsClientNoCacheRequest": true,

	// ========== 25. 压缩和性能优化方法 ==========
	// HTTP压缩、性能监控和优化相关方法
	"SetGzipResponse": true, "EnableGzipCompression": true,
	"GenerateContentHash": true, "GenerateVersionETag": true,
	"SetContentHashETag": true, "StartPerformanceTimer": true,
	"EndPerformanceTimer": true, "SetPerformanceHeaders": true,
	"AddPreloadLink": true, "AddPrefetchLink": true, "AddPreconnectLink": true,
	"ServerPush": true,

	// ========== 26. 国际化和本地化方法 ==========
	// 多语言支持和本地化处理方法
	"SetLanguage": true, "GetLanguage": true, "GetDefaultLanguage": true,
	"DetectLanguageFromHeader": true, "T": true, "Translate": true,
	"SetLocale": true, "GetLocale": true, "FormatNumber": true,
	"FormatCurrency": true, "FormatDate": true, "FormatDateTime": true,
	"GetSupportedLanguages": true, "IsSupportedLanguage": true,
	"GetLanguageName": true, "SwitchLanguage": true, "SetLanguageCookie": true,
	"GetLanguageFromCookie": true, "GetCurrentURL": true,
	"GetLanguageDirection": true, "GetTranslationFile": true,

	// ========== 27. 安全防护方法 ==========
	// Web安全防护和HTTP安全头设置方法
	"SetSecurityHeaders": true, "SetContentSecurityPolicy": true,
	"SetXFrameOptions": true, "SetXContentTypeOptions": true,
	"SetXXSSProtection": true, "SetReferrerPolicy": true,
	"SetPermissionsPolicy": true, "SetStrictTransportSecurity": true,
	"RequireHTTPS": true, "IsHTTPS": true, "CheckIPWhitelist": true,
	"CheckIPBlacklist": true, "RateLimitCheck": true,

	// ========== 28. CSRF/XSRF防护方法 ==========
	// 跨站请求伪造防护相关方法
	"GenerateCSRFToken": true, "GetCSRFToken": true, "ValidateCSRFToken": true,
	"RequireCSRFToken": true, "CSRFError": true, "XSRFToken": true,
	"CheckXSRFCookie": true, "EnableXSRF": true, "DisableXSRF": true,
	// 统一CSRF防护方法（新增）
	"GenerateUnifiedCSRFToken": true, "ValidateUnifiedCSRFToken": true,

	// ========== 29. 密码和加密方法 ==========
	// 密码哈希、加密解密和安全token生成方法
	"HashPassword": true, "VerifyPassword": true, "GenerateSalt": true,
	"GenerateSecureToken": true, "SetSecureSessionCookie": true,

	// ========== 30. 输入验证和清理方法 ==========
	// 数据验证、输入清理和安全过滤方法
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

	// ========== 31. 日志记录方法 ==========
	// 应用日志记录和错误跟踪方法
	"LogInfo": true, "LogInfof": true, "LogError": true, "LogErrorf": true,
	"LogWarn": true, "LogDebug": true, "LogDebugf": true,
	"LogFetal": true, "LogFetalf": true, "LogPanic": true, "LogPanicsf": true,
	"LogValidationError": true, "LogSecurityEvent": true,
	"LogFailedLogin": true, "LogSuspiciousActivity": true,

	// ========== 32. 调试和监控方法 ==========
	// 调试信息收集和性能监控方法
	"GetDebugInfo": true, "StartProfiler": true, "EndProfiler": true,
	"GetProfilerResult": true, "LogPerformance": true, "DumpRequest": true,
	"DumpStackTrace": true, "LogDebugError": true, "DebugJSON": true,
	"DebugHeaders": true, "IsDebugMode": true, "GenerateRequestID": true,
	"PrintDebugInfo": true, "HealthCheck": true,

	// ========== 33. 格式化输出方法（fmt包兼容） ==========
	// 与标准库fmt包兼容的格式化输出方法
	"Print": true, "Printf": true, "Println": true,
	"Sprint": true, "Sprintf": true, "Sprintln": true,
	"Fprint": true, "Fprintf": true, "Fprintln": true,
	"Scan": true, "Scanf": true, "Scanln": true,
	"Sscan": true, "Sscanf": true, "Sscanln": true,
	"FormatOutput": true,

	// ========== 34. 断言和测试辅助方法 ==========
	// 单元测试和断言验证辅助方法
	"Assert": true, "AssertNotNil": true, "AssertEqual": true,

	// ========== 35. 指标收集方法 ==========
	// 性能指标和统计数据收集方法
	"RecordMetric": true, "IncrementCounter": true, "RecordTiming": true,
	"RecordGauge": true,

	// ========== 36. 中间件和优化管理方法 ==========
	// 中间件管理和性能优化控制方法
	"EnableOptimization": true, "DisableOptimization": true,
	"IsOptimizationEnabled": true, "GetMiddleware": true,
	"SetMiddleware": true, "AddMiddleware": true,

	// ========== 37. 统一管理器方法 ==========
	// 统一管理器和框架集成相关方法
	"GetUnifiedManager": true,

	// ========== 38. 邮件发送方法 ==========
	// SMTP邮件发送和邮件队列管理方法
	"SendMail": true, "SendSimpleMail": true, "SendHTMLMail": true,
	"SendMailWithAttachment": true, "SendTemplateMail": true,
	"SendBulkMail": true, "SendMailToList": true, "QueueMail": true,
	"ProcessMailQueue": true, "UpdateMailConfig": true,
	"TestMailConnection": true, "GetMailStats": true, "GetMailLog": true,
	"CreateAttachment": true, "FormatEmailAddress": true,

	// ========== 39. 队列和任务调度方法 ==========
	// 异步任务队列和后台作业调度方法
	"Dispatch": true, "DispatchNow": true, "DispatchLater": true,
	"DispatchToQueue": true, "DispatchEmail": true, "DispatchNotification": true,
	"RegisterJobHandler": true, "GetJobHandler": true, "ProcessQueue": true,
	"GetQueueInfo": true, "GetAllQueues": true, "ClearQueue": true,
	"PurgeFailedJobs": true, "GetJob": true, "GetJobsByType": true,
	"GetJobsByStatus": true, "InitDefaultJobHandlers": true,
	"GetQueueMetrics": true, "ExportQueueData": true, "ImportQueueData": true,

	// ========== 40. 用户认证和权限管理方法 ==========
	// 用户登录、权限验证和会话管理方法
	"IsAdminLogin": true, "SetAdminId": true,
	// 统一用户认证方法（新增）
	"LoginUser": true, "LogoutUser": true, "GetCurrentUser": true, "IsUserAuthenticated": true,

	// ========== 41. 分页处理方法 ==========
	// 数据分页和页面信息处理方法
	"GetPageInfo": true, "GetPageInfoByParam": true, "GetPageInfoDefault": true,

	// ========== 42. WebSocket 支持方法 ==========
	// WebSocket 连接处理和管理相关方法
	"HandleWebSocket": true, "IsWebSocketRequest": true, "SetWebSocketUpgrader": true,
	"CreateWebSocketEchoHandler": true, "CreateWebSocketJSONHandler": true,
}

// ControllerNameSuffixReserved 控制器名称后缀保留字（统一在此定义）
var ControllerNameSuffixReserved = map[string]bool{
	"Controller": true,
	"Ctrl":       true,
	"Handler":    true,
}

var MethodNameSuffixReserved = map[string]bool{
	"Handler": true,
	"Handle":  true,
	"Action":  true,
}

const AdminIdKey = "adminId"

// ============================================================================
// ReservedMethods 统计和验证工具
// ============================================================================

// GetReservedMethodsCount 获取保留方法总数
func GetReservedMethodsCount() int {
	return len(ReservedMethods)
}

// GetReservedMethodsByCategory 按分类获取保留方法（用于调试和维护）
// 注意：这是一个辅助函数，主要用于开发时的统计和验证
func GetReservedMethodsByCategory() map[string]int {
	categories := map[string]int{
		"生命周期方法":      7,  // Init, Prepare, Finish, Destroy, Reset, InitWithContext, AutoInit
		"控制器管理方法":     8,  // ControllerName, ActionName, SetControllerName, GetControllerName等
		"响应输出方法":      15, // JSON, String, Write, Abort等基础响应方法
		"高级响应格式方法":    12, // XML, YAML, JSONP, IndentedJSON等
		"流式响应方法":      7,  // Stream, File, Download等
		"重定向方法":       4,  // Redirect相关
		"HTTP状态判断":     6,  // IsOk, IsSuccessful等
		"HTTP头操作":      6,  // AddHeader, SetHeader等
		"模板渲染方法":      11, // Render, RenderHTML, RenderTemplate等
		"模板配置管理":      10, // SetTplName, GetLayout, AddTemplateFunction等
		"Cookie操作":      6,  // SetCookie, GetCookie等
		"Session管理":     7,  // SetSession, GetSession等
		"基础参数获取":      16, // GetString, GetInt等
		"表单解析":        7,  // ParseForm, GetStrings等
		"扩展参数获取":      12, // GetFormValue, GetParamInt等
		"Tuple参数获取":    30, // 各种Tuple方法
		"HTTP请求判断":     7,  // IsAjax, IsGet, IsPost等
		"内容类型判断":      5,  // IsJSON, IsXML等
		"数据绑定":        10, // Bind, ShouldBind等
		"高级请求处理":      8,  // GetRawBody, GetBodyString等
		"文件上传处理":      14, // SaveUploadedFile, GetFile等
		"数据存储操作":      6,  // SetData, GetData, GetContextData等
		"路由URL构建":     12, // AddMethodMapping, URLFor等
		"HTTP缓存控制":     18, // SetETag, SetCacheControl等
		"压缩性能优化":      11, // SetGzipResponse, GenerateContentHash等
		"国际化本地化":      18, // SetLanguage, Translate等
		"Web安全防护":     8,  // SetSecurityHeaders, RequireHTTPS等
		"CSRF防护":       11, // GenerateCSRFToken, ValidateCSRFToken等
		"密码加密":        5,  // HashPassword, VerifyPassword等
		"输入验证清理":      20, // Validate, ValidateStruct等
		"日志记录":        12, // LogInfo, LogError等
		"调试监控":        11, // GetDebugInfo, StartProfiler等
		"格式化输出":       14, // Print, Printf等fmt兼容方法
		"断言测试":        3,  // Assert相关
		"指标收集":        4,  // RecordMetric, IncrementCounter等
		"中间件管理":       6,  // GetMiddleware, SetMiddleware等
		"统一管理器":       1,  // GetUnifiedManager
		"邮件发送":        13, // SendMail, QueueMail等
		"队列任务调度":      17, // Dispatch, ProcessQueue等
		"用户认证权限":      6,  // IsAdminLogin, LoginUser等
		"分页处理":        3,  // GetPageInfo等
		"WebSocket支持":   5,  // HandleWebSocket等
	}
	return categories
}

// IsReservedMethod 检查指定方法是否为保留方法
func IsReservedMethod(methodName string) bool {
	return ReservedMethods[methodName]
}

// ValidateReservedMethodsIntegrity 验证保留方法的完整性（开发时使用）
// 返回：(当前总数, 预期总数, 是否一致)
func ValidateReservedMethodsIntegrity() (int, int, bool) {
	actual := len(ReservedMethods)
	expected := 530 // 基于最新扫描的BaseController方法数
	return actual, expected, actual >= expected
}
