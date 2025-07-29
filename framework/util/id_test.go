package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	t.Run("生成通用ID", func(t *testing.T) {
		id := GenerateID("test")
		
		assert.NotEmpty(t, id)
		assert.True(t, strings.HasPrefix(id, "test_"))
		
		// 生成两个ID应该不同
		id2 := GenerateID("test")
		assert.NotEqual(t, id, id2)
	})
}

func TestGenerateSpecificIDs(t *testing.T) {
	t.Run("生成用户ID", func(t *testing.T) {
		id := GenerateUserID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "user_"))
	})
	
	t.Run("生成管理员ID", func(t *testing.T) {
		id := GenerateAdminID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "admin_"))
	})
	
	t.Run("生成账户ID", func(t *testing.T) {
		id := GenerateAccountID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "account_"))
	})
	
	t.Run("生成角色ID", func(t *testing.T) {
		id := GenerateRoleID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "role_"))
	})
	
	t.Run("生成权限ID", func(t *testing.T) {
		id := GeneratePermissionID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "perm_"))
	})
	
	t.Run("生成会话ID", func(t *testing.T) {
		id := GenerateSessionID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "session_"))
	})
	
	t.Run("生成令牌ID", func(t *testing.T) {
		id := GenerateTokenID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "token_"))
	})
	
	t.Run("生成文件ID", func(t *testing.T) {
		id := GenerateFileID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "file_"))
	})
	
	t.Run("生成日志ID", func(t *testing.T) {
		id := GenerateLogID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "log_"))
	})
	
	t.Run("生成任务ID", func(t *testing.T) {
		id := GenerateTaskID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "task_"))
	})
	
	t.Run("生成订单ID", func(t *testing.T) {
		id := GenerateOrderID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "order_"))
	})
	
	t.Run("生成产品ID", func(t *testing.T) {
		id := GenerateProductID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "product_"))
	})
	
	t.Run("生成分类ID", func(t *testing.T) {
		id := GenerateCategoryID()
		assert.NotEmpty(t, id.String())
		assert.True(t, strings.HasPrefix(id.String(), "category_"))
	})
}

func TestUUID(t *testing.T) {
	t.Run("生成UUID", func(t *testing.T) {
		uuid := UUID()
		
		assert.NotEmpty(t, uuid)
		// UUID格式应该包含连字符
		assert.Contains(t, uuid, "-")
		
		// 生成两个UUID应该不同
		uuid2 := UUID()
		assert.NotEqual(t, uuid, uuid2)
	})
}

func TestShortID(t *testing.T) {
	t.Run("生成短ID", func(t *testing.T) {
		id := ShortID()
		
		assert.NotEmpty(t, id)
		assert.Equal(t, 12, len(id)) // 6字节转16进制为12个字符
		
		// 生成两个短ID应该不同
		id2 := ShortID()
		assert.NotEqual(t, id, id2)
	})
}

func TestNumericID(t *testing.T) {
	t.Run("生成数字ID", func(t *testing.T) {
		id := NumericID()
		
		assert.Greater(t, id, int64(0))
		
		// 生成两个ID应该不同(时间递增)
		id2 := NumericID()
		assert.GreaterOrEqual(t, id2, id)
	})
}

func TestIDStringMethods(t *testing.T) {
	t.Run("ID类型String方法", func(t *testing.T) {
		tests := []struct {
			name string
			id   interface{ String() string }
		}{
			{"AccountID", AccountID("test-account")},
			{"UserID", UserID("test-user")},
			{"AdminID", AdminID("test-admin")},
			{"RoleID", RoleID("test-role")},
			{"PermissionID", PermissionID("test-permission")},
			{"SessionID", SessionID("test-session")},
			{"TokenID", TokenID("test-token")},
			{"FileID", FileID("test-file")},
			{"LogID", LogID("test-log")},
			{"TaskID", TaskID("test-task")},
			{"OrderID", OrderID("test-order")},
			{"ProductID", ProductID("test-product")},
			{"CategoryID", CategoryID("test-category")},
		}
		
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := tt.id.String()
				assert.True(t, strings.HasPrefix(result, "test-"))
			})
		}
	})
}