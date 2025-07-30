package errors

import (
	"errors"
	"fmt"
)

// =============== 错误处理类型定义 ===============

// ErrNo 自定义错误类型(来自FreeCar项目)
type ErrNo struct {
	ErrCode int64  `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

// Error 实现error接口
func (e ErrNo) Error() string {
	return fmt.Sprintf("err_code=%d, err_msg=%s", e.ErrCode, e.ErrMsg)
}

// NewErrNo 创建新的错误
func NewErrNo(code int64, msg string) ErrNo {
	return ErrNo{
		ErrCode: code,
		ErrMsg:  msg,
	}
}

// WithMessage 修改错误消息
func (e ErrNo) WithMessage(msg string) ErrNo {
	return ErrNo{
		ErrCode: e.ErrCode,
		ErrMsg:  msg,
	}
}

// Response 标准响应结构(来自FreeCar项目)
type Response struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

// 预定义的错误码
var (
	// 成功
	SuccessErrNo = NewErrNo(0, "Success")

	// 系统错误 (10000-10999)
	ServiceError    = NewErrNo(10001, "Service is unable to start successfully")
	ParamError      = NewErrNo(10002, "Wrong Parameter has been given")
	AuthorizeFail   = NewErrNo(10003, "Authorization failed")
	TooManyRequests = NewErrNo(10004, "Too many requests")
	ForbiddenError  = NewErrNo(10005, "Forbidden")

	// 用户相关错误 (20000-20999)
	UserNotExist     = NewErrNo(20001, "User does not exists")
	UserAlreadyExist = NewErrNo(20002, "User already exists")
	UserCreateError  = NewErrNo(20003, "User create error")
	UserLoginError   = NewErrNo(20004, "User login error")
	UserUpdateError  = NewErrNo(20005, "User update error")
	UserDeleteError  = NewErrNo(20006, "User delete error")
	InvalidPassword  = NewErrNo(20007, "Invalid password")
	UserDisabled     = NewErrNo(20008, "User is disabled")

	// 权限相关错误 (30000-30999)
	PermissionDenied = NewErrNo(30001, "Permission denied")
	TokenExpired     = NewErrNo(30002, "Token expired")
	TokenInvalid     = NewErrNo(30003, "Token invalid")
	TokenMissing     = NewErrNo(30004, "Token missing")

	// 数据相关错误 (40000-40999)
	DataNotFound     = NewErrNo(40001, "Data not found")
	DataCreateError  = NewErrNo(40002, "Data create error")
	DataUpdateError  = NewErrNo(40003, "Data update error")
	DataDeleteError  = NewErrNo(40004, "Data delete error")
	DataAlreadyExist = NewErrNo(40005, "Data already exists")
	DatabaseError    = NewErrNo(40006, "Database error")

	// 文件相关错误 (50000-50999)
	FileNotFound    = NewErrNo(50001, "File not found")
	FileUploadError = NewErrNo(50002, "File upload error")
	FileDeleteError = NewErrNo(50003, "File delete error")
	FileSizeError   = NewErrNo(50004, "File size error")
	FileTypeError   = NewErrNo(50005, "File type error")

	// 网络相关错误 (60000-60999)
	NetworkError    = NewErrNo(60001, "Network error")
	TimeoutError    = NewErrNo(60002, "Request timeout")
	ConnectionError = NewErrNo(60003, "Connection error")

	// RPC相关错误 (70000-70999)
	RPCError        = NewErrNo(70001, "RPC call error")
	ServiceNotFound = NewErrNo(70002, "Service not found")

	// 缓存相关错误 (80000-80999)
	CacheError    = NewErrNo(80001, "Cache error")
	CacheNotFound = NewErrNo(80002, "Cache not found")

	// 配置相关错误 (90000-90999)
	ConfigError    = NewErrNo(90001, "Configuration error")
	ConfigNotFound = NewErrNo(90002, "Configuration not found")
)

func NewSystemError(msg string) error {
	return errors.New(msg)
}
