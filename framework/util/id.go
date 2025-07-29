package util

import (
	"crypto/rand"
	"fmt"
	"time"
)

// 强类型ID类型定义(来自FreeCar项目)
type (
	// AccountID 账户ID
	AccountID string
	// UserID 用户ID  
	UserID string
	// AdminID 管理员ID
	AdminID string
	// RoleID 角色ID
	RoleID string
	// PermissionID 权限ID
	PermissionID string
	// SessionID 会话ID
	SessionID string
	// TokenID 令牌ID
	TokenID string
	// FileID 文件ID
	FileID string
	// LogID 日志ID
	LogID string
	// TaskID 任务ID
	TaskID string
	// OrderID 订单ID
	OrderID string
	// ProductID 产品ID
	ProductID string
	// CategoryID 分类ID
	CategoryID string
)

// String 方法实现
func (id AccountID) String() string    { return string(id) }
func (id UserID) String() string       { return string(id) }
func (id AdminID) String() string      { return string(id) }
func (id RoleID) String() string       { return string(id) }
func (id PermissionID) String() string { return string(id) }
func (id SessionID) String() string    { return string(id) }
func (id TokenID) String() string      { return string(id) }
func (id FileID) String() string       { return string(id) }
func (id LogID) String() string        { return string(id) }
func (id TaskID) String() string       { return string(id) }
func (id OrderID) String() string      { return string(id) }
func (id ProductID) String() string    { return string(id) }
func (id CategoryID) String() string   { return string(id) }

// GenerateID 生成通用ID
func GenerateID(prefix string) string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	
	return fmt.Sprintf("%s_%d_%x", prefix, timestamp, randomBytes)
}

// GenerateUserID 生成用户ID
func GenerateUserID() UserID {
	return UserID(GenerateID("user"))
}

// GenerateAdminID 生成管理员ID
func GenerateAdminID() AdminID {
	return AdminID(GenerateID("admin"))
}

// GenerateAccountID 生成账户ID
func GenerateAccountID() AccountID {
	return AccountID(GenerateID("account"))
}

// GenerateRoleID 生成角色ID
func GenerateRoleID() RoleID {
	return RoleID(GenerateID("role"))
}

// GeneratePermissionID 生成权限ID
func GeneratePermissionID() PermissionID {
	return PermissionID(GenerateID("perm"))
}

// GenerateSessionID 生成会话ID
func GenerateSessionID() SessionID {
	return SessionID(GenerateID("session"))
}

// GenerateTokenID 生成令牌ID
func GenerateTokenID() TokenID {
	return TokenID(GenerateID("token"))
}

// GenerateFileID 生成文件ID
func GenerateFileID() FileID {
	return FileID(GenerateID("file"))
}

// GenerateLogID 生成日志ID
func GenerateLogID() LogID {
	return LogID(GenerateID("log"))
}

// GenerateTaskID 生成任务ID
func GenerateTaskID() TaskID {
	return TaskID(GenerateID("task"))
}

// GenerateOrderID 生成订单ID
func GenerateOrderID() OrderID {
	return OrderID(GenerateID("order"))
}

// GenerateProductID 生成产品ID
func GenerateProductID() ProductID {
	return ProductID(GenerateID("product"))
}

// GenerateCategoryID 生成分类ID
func GenerateCategoryID() CategoryID {
	return CategoryID(GenerateID("category"))
}

// UUID 生成简单的UUID
func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ShortID 生成短ID
func ShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// NumericID 生成数字ID
func NumericID() int64 {
	return time.Now().UnixNano()
}