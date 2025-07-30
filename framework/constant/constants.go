package constant

// =============== YYHertz 框架常量定义文件 ===============
// 该文件包含了 YYHertz 框架中使用的所有常量定义，包括：
// - 基础系统常量：任务状态、HTTP状态码、安全相关代码等
// - 业务常量：用户类型、权限级别、审核状态、支付状态等
// - 技术常量：文件上传、缓存、日志、环境配置等
// =====================================================

// =============== 基础系统常量 ===============

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

// 安全相关错误代码
const (
	CodeHTTPSRequired CodeResult = 1000 + iota // HTTPS连接是必需的
	CodeTLSError                               // TLS错误
	CodeCertError                              // 证书错误
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

// =============== 技术配置常量 ===============

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
	TimeFormatDefault  = "2006-01-02 15:04:05"
	TimeFormatDate     = "2006-01-02"
	TimeFormatTime     = "15:04:05"
	TimeFormatISO8601  = "2006-01-02T15:04:05Z07:00"
	TimeFormatRFC3339  = "2006-01-02T15:04:05Z07:00"
	TimeFormatUnix     = "1136239445"
	TimeFormatChinese  = "2006年01月02日 15时04分05秒"
	TimeFormatCompact  = "20060102150405"
	TimeFormatDateOnly = "20060102"
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

// =============== 业务常量定义 ===============

// 学校相关常量
const (
	DefaultSchoolCode = 10000 // 默认学校代码
)

// 路径相关常量
const (
	DefaultLoginURI       = "/admin/login"
	DefaultLoginMobileURI = "/mobile/login"
	DefaultLogoutURI      = "/admin/logout"

	SchoolLoginPath = "/school/login"
	AdminLoginPath  = "/admin/login"
)

// CAS相关常量
const (
	CASLoginURI                   = "/cas/login"
	CASLogoutURI                  = "/cas/logout"
	CASValidateURI                = "/cas/validate"
	CASVersion2ServiceValidateURI = "/cas/serviceValidate"
	CASVersion3ServiceValidateURI = "/cas/p3/serviceValidate"
)

// 业务状态码常量
const (
	BusinessCodeSuccess    = 0 // 业务成功
	BusinessCodeError      = 1 // 业务失败
	BusinessCodeProcessing = 2 // 处理中
	BusinessCodePending    = 3 // 待处理
	BusinessCodeCanceled   = 4 // 已取消
	BusinessCodeExpired    = 5 // 已过期
)

// 用户类型常量
const (
	UserTypeAdmin   = "admin"   // 管理员
	UserTypeTeacher = "teacher" // 教师
	UserTypeStudent = "student" // 学生
	UserTypeParent  = "parent"  // 家长
	UserTypeGuest   = "guest"   // 访客
)

// 权限级别常量
const (
	PermissionLevelNone      = 0 // 无权限
	PermissionLevelRead      = 1 // 只读权限
	PermissionLevelWrite     = 2 // 读写权限
	PermissionLevelAdmin     = 3 // 管理员权限
	PermissionLevelSuperUser = 4 // 超级用户权限
)

// 数据状态常量
const (
	DataStatusActive   = 1  // 活跃状态
	DataStatusInactive = 0  // 非活跃状态
	DataStatusDeleted  = -1 // 已删除状态
	DataStatusArchived = -2 // 已归档状态
)

// 操作类型常量
const (
	OperationTypeCreate = "create" // 创建操作
	OperationTypeRead   = "read"   // 读取操作
	OperationTypeUpdate = "update" // 更新操作
	OperationTypeDelete = "delete" // 删除操作
	OperationTypeExport = "export" // 导出操作
	OperationTypeImport = "import" // 导入操作
)

// 通知类型常量
const (
	NotificationTypeSystem  = "system"  // 系统通知
	NotificationTypeWarning = "warning" // 警告通知
	NotificationTypeError   = "error"   // 错误通知
	NotificationTypeInfo    = "info"    // 信息通知
	NotificationTypeSuccess = "success" // 成功通知
)

// 日志操作类型常量
const (
	LogActionLogin    = "login"    // 登录
	LogActionLogout   = "logout"   // 登出
	LogActionCreate   = "create"   // 创建
	LogActionUpdate   = "update"   // 更新
	LogActionDelete   = "delete"   // 删除
	LogActionView     = "view"     // 查看
	LogActionDownload = "download" // 下载
	LogActionUpload   = "upload"   // 上传
	LogActionExport   = "export"   // 导出
	LogActionImport   = "import"   // 导入
)

// 文件类型常量
const (
	FileTypeImage    = "image"    // 图片文件
	FileTypeDocument = "document" // 文档文件
	FileTypeVideo    = "video"    // 视频文件
	FileTypeAudio    = "audio"    // 音频文件
	FileTypeArchive  = "archive"  // 压缩文件
	FileTypeOther    = "other"    // 其他文件
)

// 学期类型常量
const (
	SemesterTypeSpring = "spring" // 春季学期
	SemesterTypeFall   = "fall"   // 秋季学期
	SemesterTypeSummer = "summer" // 夏季学期
	SemesterTypeWinter = "winter" // 冬季学期
)

// 审核状态常量
const (
	AuditStatusPending  = 0 // 待审核
	AuditStatusApproved = 1 // 已通过
	AuditStatusRejected = 2 // 已拒绝
	AuditStatusRevoked  = 3 // 已撤销
)

// 优先级等级常量
const (
	PriorityLow      = 1 // 低优先级
	PriorityNormal   = 2 // 普通优先级
	PriorityHigh     = 3 // 高优先级
	PriorityUrgent   = 4 // 紧急优先级
	PriorityCritical = 5 // 危急优先级
)

// 性别常量
const (
	GenderUnknown = 0 // 未知
	GenderMale    = 1 // 男性
	GenderFemale  = 2 // 女性
	GenderOther   = 3 // 其他
)

// 婚姻状况常量
const (
	MaritalStatusSingle   = "single"   // 单身
	MaritalStatusMarried  = "married"  // 已婚
	MaritalStatusDivorced = "divorced" // 离异
	MaritalStatusWidowed  = "widowed"  // 丧偶
	MaritalStatusOther    = "other"    // 其他
)

// 教育程度常量
const (
	EducationLevelPrimary   = "primary"   // 小学
	EducationLevelJunior    = "junior"    // 初中
	EducationLevelSenior    = "senior"    // 高中
	EducationLevelCollege   = "college"   // 大专
	EducationLevelBachelor  = "bachelor"  // 本科
	EducationLevelMaster    = "master"    // 硕士
	EducationLevelDoctorate = "doctorate" // 博士
	EducationLevelPostdoc   = "postdoc"   // 博士后
	EducationLevelOther     = "other"     // 其他
)

// 考试类型常量
const (
	ExamTypeMidterm = "midterm" // 期中考试
	ExamTypeFinal   = "final"   // 期末考试
	ExamTypeQuiz    = "quiz"    // 小测验
	ExamTypeTest    = "test"    // 测试
	ExamTypeOther   = "other"   // 其他
)

// 成绩等级常量
const (
	GradeLevelA = "A" // 优秀
	GradeLevelB = "B" // 良好
	GradeLevelC = "C" // 中等
	GradeLevelD = "D" // 及格
	GradeLevelF = "F" // 不及格
)

// 课程状态常量
const (
	CourseStatusDraft     = "draft"     // 草稿
	CourseStatusPublished = "published" // 已发布
	CourseStatusActive    = "active"    // 进行中
	CourseStatusCompleted = "completed" // 已完成
	CourseStatusCanceled  = "canceled"  // 已取消
	CourseStatusArchived  = "archived"  // 已归档
)

// 消息状态常量
const (
	MessageStatusUnread  = 0 // 未读
	MessageStatusRead    = 1 // 已读
	MessageStatusReplied = 2 // 已回复
	MessageStatusDeleted = 3 // 已删除
)

// 支付状态常量
const (
	PaymentStatusPending  = "pending"  // 待支付
	PaymentStatusPaid     = "paid"     // 已支付
	PaymentStatusRefunded = "refunded" // 已退款
	PaymentStatusCanceled = "canceled" // 已取消
	PaymentStatusExpired  = "expired"  // 已过期
)

// 设备类型常量
const (
	DeviceTypePC      = "pc"      // 电脑
	DeviceTypeMobile  = "mobile"  // 手机
	DeviceTypeTablet  = "tablet"  // 平板
	DeviceTypeUnknown = "unknown" // 未知设备
)

// API版本常量
const (
	APIVersionV1 = "v1" // API版本1
	APIVersionV2 = "v2" // API版本2
	APIVersionV3 = "v3" // API版本3
)

// =============== 错误码定义 ===============

// 业务错误码常量
const (
	// 用户相关错误码 (1000-1999)
	ErrorCodeUserNotFound       = 1001 // 用户不存在
	ErrorCodeUserExists         = 1002 // 用户已存在
	ErrorCodeInvalidCredentials = 1003 // 凭据无效
	ErrorCodeUserDisabled       = 1004 // 用户已禁用
	ErrorCodePasswordExpired    = 1005 // 密码已过期
	ErrorCodeAccountLocked      = 1006 // 账户已锁定

	// 权限相关错误码 (2000-2999)
	ErrorCodePermissionDenied       = 2001 // 权限被拒绝
	ErrorCodeInsufficientPrivileges = 2002 // 权限不足
	ErrorCodeRoleNotFound           = 2003 // 角色不存在
	ErrorCodeAccessExpired          = 2004 // 访问已过期

	// 数据相关错误码 (3000-3999)
	ErrorCodeDataNotFound      = 3001 // 数据不存在
	ErrorCodeDataExists        = 3002 // 数据已存在
	ErrorCodeDataCorrupted     = 3003 // 数据损坏
	ErrorCodeDataTooLarge      = 3004 // 数据过大
	ErrorCodeInvalidDataFormat = 3005 // 数据格式无效

	// 文件相关错误码 (4000-4999)
	ErrorCodeFileNotFound       = 4001 // 文件不存在
	ErrorCodeFileTooBig         = 4002 // 文件过大
	ErrorCodeInvalidFileType    = 4003 // 文件类型无效
	ErrorCodeFileUploadFailed   = 4004 // 文件上传失败
	ErrorCodeFileDownloadFailed = 4005 // 文件下载失败

	// 网络相关错误码 (5000-5999)
	ErrorCodeNetworkTimeout     = 5001 // 网络超时
	ErrorCodeNetworkUnavailable = 5002 // 网络不可用
	ErrorCodeServiceUnavailable = 5003 // 服务不可用
	ErrorCodeRateLimitExceeded  = 5004 // 超出频率限制
)

// 业务成功码常量
const (
	// 操作成功码 (10000-10999)
	SuccessCodeOperationComplete = 10001 // 操作完成
	SuccessCodeDataSaved         = 10002 // 数据已保存
	SuccessCodeDataUpdated       = 10003 // 数据已更新
	SuccessCodeDataDeleted       = 10004 // 数据已删除
	SuccessCodeFileUploaded      = 10005 // 文件已上传
	SuccessCodeFileDownloaded    = 10006 // 文件已下载
	SuccessCodeMessageSent       = 10007 // 消息已发送
	SuccessCodeEmailSent         = 10008 // 邮件已发送
	SuccessCodeNotificationSent  = 10009 // 通知已发送
)
