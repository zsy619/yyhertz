package core

// ReservedMethods 定义需要跳过的BaseController和生命周期方法（导出为公共变量）
// 只包含实际存在的方法，确保准确性
var ReservedMethods = map[string]bool{
	// ============= 生命周期方法 =============
	"Init": true, "Prepare": true, "Finish": true,
	"Destroy": true, "Reset": true, "InitWithContext": true,
	"AutoInit": true,

	// ============= 基础响应方法 =============
	"JSON": true, "String": true, "Write": true,
	"JSONWithStatus": true, "StringWithStatus": true,
	"JSONOK": true, "JSONError": true, "JSONSuccess": true,
	"JSONPage": true, "JSONStatus": true,
	"Redirect": true, "Error": true, "Abort": true, "CustomAbort": true, "StopRun": true,

	// ============= 高级响应方法 =============
	"XML": true, "XMLWithStatus": true, "YAML": true, "YAMLWithStatus": true,
	"IndentedJSON": true, "IndentedJSONWithStatus": true, "DataWithStatus": true,
	"Status": true, "NoContent": true, "Stream": true,
	"RedirectPermanent": true, "RedirectTemporary": true, "RedirectSeeOther": true,
	"IsOk": true, "IsSuccessful": true, "IsRedirect": true, "IsClientError": true,
	"IsServerError": true, "IsForbidden": true, "AddHeader": true, "SetHeader": true,
	"GetResponseHeader": true, "SetContentType": true, "JSONP": true, "JSONPWithStatus": true, "SetResponseHeader": true,
	"EmptyResponse": true, "RawResponse": true,

	// ============= 模板渲染方法 =============
	"Render": true, "RenderHTML": true, "RenderWithLayout": true,
	"RenderBytes": true, "RenderString": true, "RenderWithViewName": true,
	"RenderTemplate": true, "RenderTemplateComponent": true, "RenderTemplateWithLayout": true,
	"RenderHTMLWithIncludes": true,

	// ============= 模板配置方法 =============
	"SetTplName": true, "GetTplName": true, "SetLayout": true, "GetLayout": true,
	"AddTplFunc": true, "GetTemplateManager": true, "SetTemplatePath": true,
	"SetTemplateTheme": true, "GetTemplateTheme": true, "AddTemplateFunction": true,
	"AddBeegoTemplateFunctions": true, "CreateTemplateDefinition": true, "ListAvailableTemplates": true,
	"SetTemplateIncludeEngine": true, "GetTemplateIncludeEngine": true,

	// ============= Cookie操作方法 =============
	"SetCookie": true, "GetCookie": true, "DeleteCookie": true, "HasCookie": true,
	"SetSecureCookie": true, "GetSecureCookie": true,

	// ============= Session操作方法 =============
	"SetSession": true, "GetSession": true, "DeleteSession": true, "DestroySession": true,
	"HasSession": true, "GetSessionID": true, "RegenerateSessionID": true,

	// ============= 请求参数获取方法 =============
	"GetForm": true, "GetQuery": true, "GetParam": true, "GetString": true,
	"GetInt": true, "GetBool": true, "GetFloat": true,
	"GetHeader": true, "GetUserAgent": true, "GetClientIP": true,

	// ============= 高级请求处理方法 =============
	"Bind": true, "ShouldBind": true, "BindJSON": true, "ShouldBindJSON": true,
	"BindXML": true, "BindQuery": true, "ShouldBindQuery": true, "Validate": true,
	"ValidateStruct": true, "BindAndValidate": true, "ShouldBindAndValidate": true,
	"GetRawBody": true, "GetRawData": true, "GetBodySize": true, "HasBody": true,
	"GetBodyString": true, "GetMultipartForm": true, "ParseMultipartForm": true,
	"SetMaxMemory": true, "GetQueryAll": true, "GetQueryMap": true, "GetQueryDefault": true,
	"GetQueryInt64": true, "GetQueryFloat64": true, "GetQueryBool": true,
	"GetFormValue": true, "GetFormValueDefault": true, "GetPostFormInt": true,
	"GetPostFormFloat64": true, "GetPostFormBool": true, "GetParamNames": true,
	"GetParamValues": true, "GetAllParams": true, "GetParamInt": true, "GetParamInt64": true,
	"GetParamFloat64": true, "GetParamBool": true, "GetParamDefault": true,
	"IsJSON": true, "IsXML": true, "IsForm": true, "IsMultipart": true, "IsUpload": true,

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
	"IsValidAction": true, "GetAvailableActions": true,

	// ============= 路由和映射方法 =============
	"AddMethodMapping": true, "GetMethodMapping": true, "SetMethodMapping": true, "GetMappedMethod": true,
	"SetRoutePattern": true, "GetRoutePattern": true, "SetRouteParam": true, "GetRouteParam": true,
	"GetRouteParams": true, "SetRouteParams": true,
	"URLFor": true, "BuildURL": true, "URLMapping": true, "HandlerFunc": true,
	"AddURLMapping": true, "GetURLMappings": true,

	// ============= 文件处理方法 =============
	"File": true, "FileAttachment": true, "Inline": true, "ServeFile": true, "Download": true,
	"SaveUploadedFile": true, "GetFile": true, "HasFile": true, "GetFileSize": true,
	"GetFileName": true, "GetFileHeader": true, "GetFiles": true, "SaveMultipleFiles": true,
	"ValidateFileSize": true, "ValidateFileExtension": true, "ValidateFileType": true,
	"StreamFile": true, "StreamReader": true, "SetFileResponseHeaders": true,
	"GetContentTypeByExtension": true, "ValidateFileUpload": true, "AddFileAttachment": true,

	// ============= 缓存和性能方法 =============
	"SetETag": true, "SetLastModified": true, "SetCacheControl": true, "SetMaxAge": true,
	"SetNoCache": true, "SetPrivateCache": true, "SetPublicCache": true, "NotModified": true,
	"CheckIfNoneMatch": true, "CheckIfModifiedSince": true, "HandleConditionalRequest": true,
	"SetGzipResponse": true, "EnableGzipCompression": true, "GenerateContentHash": true,
	"GenerateVersionETag": true, "SetContentHashETag": true, "SetExpires": true,
	"SetExpiresFromNow": true, "SetCacheForMinutes": true, "SetCacheForHours": true,
	"SetCacheForDays": true, "SetImmutableCache": true, "SetShortCache": true,
	"SetMediumCache": true, "SetLongCache": true, "IsCacheableMethod": true,
	"ShouldCache": true, "GetClientCachePreference": true, "IsClientNoCacheRequest": true,
	"StartPerformanceTimer": true, "EndPerformanceTimer": true, "SetPerformanceHeaders": true,
	"AddPreloadLink": true, "AddPrefetchLink": true, "AddPreconnectLink": true, "ServerPush": true,

	// ============= 国际化方法 =============
	"SetLanguage": true, "GetLanguage": true, "GetDefaultLanguage": true,
	"DetectLanguageFromHeader": true, "T": true, "Translate": true, "SetLocale": true,
	"GetLocale": true, "FormatNumber": true, "FormatCurrency": true, "FormatDate": true,
	"FormatDateTime": true, "GetSupportedLanguages": true, "IsSupportedLanguage": true,
	"GetLanguageName": true, "SwitchLanguage": true, "SetLanguageCookie": true,
	"GetLanguageFromCookie": true, "GetCurrentURL": true, "GetLanguageDirection": true,
	"BuildLocalizedURL": true, "GetTranslationFile": true,

	// ============= 安全增强方法 =============
	"SetSecurityHeaders": true, "SetContentSecurityPolicy": true, "SetXFrameOptions": true,
	"SetXContentTypeOptions": true, "SetXXSSProtection": true, "SetReferrerPolicy": true,
	"SetPermissionsPolicy": true, "SetStrictTransportSecurity": true,
	"GenerateCSRFToken": true, "GetCSRFToken": true, "ValidateCSRFToken": true,
	"RequireCSRFToken": true, "CSRFError": true, "SanitizeHTML": true,
	"ValidateEmail": true, "ValidatePhone": true, "ValidateURL": true,
	"ValidateIPAddress": true, "SanitizeFilename": true, "CheckIPWhitelist": true,
	"CheckIPBlacklist": true, "RateLimitCheck": true, "RequireHTTPS": true,
	"IsHTTPS": true, "HashPassword": true, "VerifyPassword": true, "GenerateSalt": true,
	"GenerateSecureToken": true, "SetSecureSessionCookie": true,
	"ValidateSessionTimeout": true, "LogSecurityEvent": true,
	"LogFailedLogin": true, "LogSuspiciousActivity": true,

	// ============= XSRF/CSRF安全方法 =============
	"XSRFToken": true, "CheckXSRFCookie": true, "EnableXSRF": true, "DisableXSRF": true,

	// ============= 调试和监控方法 =============
	"GetDebugInfo": true, "StartProfiler": true, "EndProfiler": true,
	"GetProfilerResult": true, "LogPerformance": true, "DumpRequest": true,
	"DumpStackTrace": true, "LogDebugError": true, "DebugJSON": true, "DebugHeaders": true,
	"IsDebugMode": true, "GenerateRequestID": true, "PrintDebugInfo": true,
	"Assert": true, "AssertNotNil": true, "AssertEqual": true, "RecordMetric": true,
	"IncrementCounter": true, "RecordTiming": true, "RecordGauge": true,
	"HealthCheck": true,

	// ============= 日志方法 =============
	"LogInfo": true, "LogInfof": true, "LogError": true, "LogErrorf": true,
	"LogWarn": true, "LogDebug": true, "LogDebugf": true,
	"LogFetal": true, "LogFetalf": true, "LogPanic": true, "LogPanicsf": true,
	"LogValidationError": true,

	// ============= 数据验证方法 =============
	"ValidateRequired": true, "ValidateLength": true, "ValidateRange": true, "ValidatePattern": true,
	"ValidateIn": true, "ValidateEmailFormat": true, "ValidatePhoneFormat": true, "ValidateIDCardFormat": true,
	"ValidateURLFormat": true, "ValidateIPFormat": true, "ValidatePasswordStrength": true, "ValidateBatch": true,
	"ValidateForm": true, "HasValidationErrors": true, "GetFirstValidationError": true, "SetValidationErrors": true,
	"RegisterValidator": true, "ApplyCustomValidator": true, "ValidateDateTime": true, "ValidateJSON": true,
	"ValidateNumeric": true, "ValidateInteger": true, "ValidateBoolean": true, "CreateValidationResult": true,
	"ReturnValidationResult": true,

	// ============= 中间件和优化方法 =============
	"EnableOptimization": true, "DisableOptimization": true, "IsOptimizationEnabled": true,
	"GetMiddleware": true, "SetMiddleware": true, "AddMiddleware": true,

	// ============= 邮件发送方法 =============
	"SendMail": true, "SendSimpleMail": true, "SendHTMLMail": true, "SendMailWithAttachment": true,
	"SendTemplateMail": true, "SendBulkMail": true, "SendMailToList": true, "QueueMail": true,
	"ProcessMailQueue": true, "UpdateMailConfig": true, "TestMailConnection": true, "GetMailStats": true,
	"GetMailLog": true, "CreateAttachment": true, "FormatEmailAddress": true,

	// ============= 队列任务方法 =============
	"Dispatch": true, "DispatchNow": true, "DispatchLater": true, "DispatchToQueue": true,
	"DispatchEmail": true, "DispatchNotification": true, "RegisterJobHandler": true, "GetJobHandler": true,
	"ProcessQueue": true, "GetQueueInfo": true, "GetAllQueues": true, "ClearQueue": true, "PurgeFailedJobs": true,
	"GetJob": true, "GetJobsByType": true, "GetJobsByStatus": true, "InitDefaultJobHandlers": true,
	"GetQueueMetrics": true, "ExportQueueData": true, "ImportQueueData": true,
}

var ControllerNameSuffixReserved = map[string]bool{
	"Controller": true,
	"Control":    true,
	"Ctrl":       true,
	"Handler":    true,
}
