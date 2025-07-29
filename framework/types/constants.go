package types

// 任务执行状态常量
const (
	TASK_SUCCESS = 0  // 任务执行成功
	TASK_ERROR   = -1 // 任务执行出错
	TASK_TIMEOUT = -2 // 任务执行超时
)

// CodeResult 结果代码类型
type CodeResult int

// 基础结果代码
const (
	CodeSuccess    CodeResult = iota // 成功
	CodeError                        // 失败
	CodeInvalid                      // 无效
	CodeNoAuth                       // 无权限
	CodeParamError                   // 参数错误
	CodeNoLogin                      // 未登录
	CodeNoData                       // 无数据
	CodeFatal                        // 严重错误
)

// HTTP状态码常量
const (
	Code400 CodeResult = 400 + iota // Bad Request
	Code401                         // Unauthorized
	Code402                         // Payment Required
	Code403                         // Forbidden
	Code404                         // Not Found
)

const (
	Code500 CodeResult = 500 + iota // Internal Server Error
	Code501                         // Not Implemented
	Code502                         // Bad Gateway
	Code503                         // Service Unavailable
	Code504                         // Gateway Timeout
	Code505                         // HTTP Version Not Supported
)

// 注意：DefaultSchoolCode 已移动到 business_constants.go

// 文件上传相关常量
const (
	MaxFileSize     = 10 << 20 // 10MB
	MaxMemory       = 32 << 20 // 32MB
	UploadDir       = "./uploads"
	AllowedFileExts = ".jpg,.jpeg,.png,.gif,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar"
)

// Session相关常量
const (
	DefaultSessionName     = "HERTZ_SESSION_ID"
	DefaultSessionLifetime = 30 * 60 // 30分钟（秒）
	SessionCookieMaxAge    = 30 * 60 // 30分钟（秒）
)

// 缓存相关常量
const (
	DefaultCacheExpire = 10 * 60 // 10分钟（秒）
	MaxCacheSize       = 1000    // 最大缓存项数
)

// 验证码相关常量
const (
	DefaultCaptchaLength = 5
	DefaultCaptchaWidth  = 120
	DefaultCaptchaHeight = 40
	CaptchaExpire        = 5 * 60 // 5分钟（秒）
)

// 数据库相关常量
const (
	DefaultPageSize = 20
	MaxPageSize     = 1000
)

// 日志级别常量
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
	LogLevelFatal = "fatal"
)

// 环境常量
const (
	EnvDevelopment = "development"
	EnvProduction  = "production"
	EnvTesting     = "testing"
)

// API相关常量
const (
	ContentTypeJSON = "application/json"
	ContentTypeForm = "application/x-www-form-urlencoded"
	ContentTypeHTML = "text/html"
	ContentTypeXML  = "application/xml"
)

// 加密相关常量
const (
	DefaultHashSalt   = "hertz-mvc-salt"
	DefaultEncryptKey = "hertz-mvc-key-16"
	BCryptCost        = 10
)

// 时间格式常量
const (
	TimeFormatDefault    = "2006-01-02 15:04:05"
	TimeFormatDate       = "2006-01-02"
	TimeFormatTime       = "15:04:05"
	TimeFormatISO8601    = "2006-01-02T15:04:05Z07:00"
	TimeFormatRFC3339    = "2006-01-02T15:04:05Z07:00"
	TimeFormatUnix       = "1136239445"
	TimeFormatChinese    = "2006年01月02日 15时04分05秒"
	TimeFormatCompact    = "20060102150405"
	TimeFormatDateOnly   = "20060102"
)

// 正则表达式常量
const (
	RegexEmail    = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	RegexPhone    = `^1[3-9]\d{9}$`
	RegexIDCard   = `^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`
	RegexIP       = `^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`
	RegexURL      = `^https?://[^\s/$.?#].[^\s]*$`
	RegexUsername = `^[a-zA-Z0-9_]{3,20}$`
	RegexPassword = `^.{6,20}$`
)