package yyhertz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBaseControllerEnhancedMethods 测试增强的BaseController方法
func TestBaseControllerEnhancedMethods(t *testing.T) {
	controller := NewBaseController()

	t.Run("测试参数类型转换方法", func(t *testing.T) {
		// 测试没有Context时的默认值
		assert.Equal(t, int64(0), controller.GetInt64("test", 0))
		assert.Equal(t, int64(100), controller.GetInt64("test", 100))
		assert.Equal(t, 0.0, controller.GetFloat("test", 0.0))
		assert.Equal(t, 3.14, controller.GetFloat("test", 3.14))
		assert.Equal(t, false, controller.GetBool("test", false))
		assert.Equal(t, true, controller.GetBool("test", true))
	})

	t.Run("测试HTTP方法判断", func(t *testing.T) {
		// 没有Context时应该返回false
		assert.False(t, controller.IsPost())
		assert.False(t, controller.IsGet())
		assert.False(t, controller.IsPut())
		assert.False(t, controller.IsDelete())
		assert.False(t, controller.IsPatch())
		assert.False(t, controller.IsAjax())
	})

	t.Run("测试验证方法", func(t *testing.T) {
		// 测试必填字段验证
		fields := map[string]string{
			"用户名": "",
			"邮箱":  "test@example.com",
			"密码":  "",
		}
		errors := controller.ValidateRequired(fields)
		assert.Len(t, errors, 2) // 用户名和密码为空
		assert.Contains(t, errors, "用户名不能为空")
		assert.Contains(t, errors, "密码不能为空")

		// 测试邮箱验证
		assert.True(t, controller.ValidateEmail("test@example.com"))
		assert.True(t, controller.ValidateEmail("user.name+tag@domain.co.uk"))
		assert.False(t, controller.ValidateEmail("invalid-email"))
		assert.False(t, controller.ValidateEmail(""))

		// 测试手机号验证
		assert.True(t, controller.ValidatePhone("13812345678"))
		assert.True(t, controller.ValidatePhone("15987654321"))
		assert.False(t, controller.ValidatePhone("12345678901"))
		assert.False(t, controller.ValidatePhone("1381234567"))
		assert.False(t, controller.ValidatePhone(""))
	})

	t.Run("测试会话方法", func(t *testing.T) {
		// 测试设置和获取会话数据
		controller.SetSession("user_id", 123)
		controller.SetSession("username", "testuser")

		assert.Equal(t, 123, controller.GetSession("user_id"))
		assert.Equal(t, "testuser", controller.GetSession("username"))
		assert.Nil(t, controller.GetSession("nonexistent"))
	})

	t.Run("测试日志方法", func(t *testing.T) {
		// 这些方法不应该panic
		assert.NotPanics(t, func() {
			controller.LogInfo("测试信息日志")
			controller.LogError("测试错误日志")
			controller.LogDebug("测试调试日志")
		})
	})

	t.Run("测试调试方法", func(t *testing.T) {
		// 测试请求信息dump
		dumpInfo := controller.DumpRequest()
		assert.Contains(t, dumpInfo, "error")
		assert.Equal(t, "context is nil", dumpInfo["error"])
	})
}

func TestValidationMethods(t *testing.T) {
	controller := NewBaseController()

	t.Run("邮箱验证测试", func(t *testing.T) {
		validEmails := []string{
			"test@example.com",
			"user.name@domain.com",
			"user+tag@example.co.uk",
			"123@test.org",
		}

		for _, email := range validEmails {
			assert.True(t, controller.ValidateEmail(email), "应该验证通过: %s", email)
		}

		invalidEmails := []string{
			"",
			"invalid",
			"@domain.com",
			"user@",
			"user@domain",
			"user.domain.com",
		}

		for _, email := range invalidEmails {
			assert.False(t, controller.ValidateEmail(email), "应该验证失败: %s", email)
		}
	})

	t.Run("手机号验证测试", func(t *testing.T) {
		validPhones := []string{
			"13812345678",
			"15987654321",
			"18123456789",
			"19876543210",
		}

		for _, phone := range validPhones {
			assert.True(t, controller.ValidatePhone(phone), "应该验证通过: %s", phone)
		}

		invalidPhones := []string{
			"",
			"12345678901",  // 12开头
			"1381234567",   // 长度不够
			"138123456789", // 长度过长
			"1381234567a",  // 包含字母
			"21812345678",  // 2开头
		}

		for _, phone := range invalidPhones {
			assert.False(t, controller.ValidatePhone(phone), "应该验证失败: %s", phone)
		}
	})
}

func TestControllerDataManagement(t *testing.T) {
	controller := NewBaseController()

	t.Run("测试数据设置和获取", func(t *testing.T) {
		// 测试单个数据设置
		controller.SetData("key1", "value1")
		assert.Equal(t, "value1", controller.Data["key1"])

		// 测试批量数据设置
		batchData := map[string]any{
			"key2": "value2",
			"key3": 123,
			"key4": true,
		}
		controller.SetDatas(batchData)

		assert.Equal(t, "value2", controller.Data["key2"])
		assert.Equal(t, 123, controller.Data["key3"])
		assert.Equal(t, true, controller.Data["key4"])
	})

	t.Run("测试会话数据操作", func(t *testing.T) {
		// 设置会话数据
		controller.SetSession("user_id", 100)
		controller.SetSession("user_role", "admin")

		// 获取会话数据
		assert.Equal(t, 100, controller.GetSession("user_id"))
		assert.Equal(t, "admin", controller.GetSession("user_role"))
		assert.Nil(t, controller.GetSession("nonexistent_key"))

		// 验证会话数据实际存储在Data中
		assert.Equal(t, 100, controller.Data["session_user_id"])
		assert.Equal(t, "admin", controller.Data["session_user_role"])
	})
}

func TestControllerFileOperations(t *testing.T) {
	controller := NewBaseController()

	t.Run("测试文件操作（无Context）", func(t *testing.T) {
		// 当没有Context时，这些方法应该返回错误
		file, err := controller.GetFile("test")
		assert.Nil(t, file)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context is nil")

		files, err := controller.GetFiles("test")
		assert.Nil(t, files)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context is nil")

		err = controller.SaveFile(nil, "test.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context is nil")
	})
}

func TestControllerBindingMethods(t *testing.T) {
	controller := NewBaseController()

	t.Run("测试数据绑定方法（无Context）", func(t *testing.T) {
		var testStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		// 当没有Context时，这些方法应该返回错误
		err := controller.BindJSON(&testStruct)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context is nil")

		err = controller.BindQuery(&testStruct)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context is nil")

		err = controller.BindForm(&testStruct)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context is nil")
	})
}

func TestRequiredFieldValidation(t *testing.T) {
	controller := NewBaseController()

	t.Run("测试必填字段验证", func(t *testing.T) {
		testCases := []struct {
			name     string
			fields   map[string]string
			expected int // 期望的错误数量
		}{
			{
				name: "所有字段都不为空",
				fields: map[string]string{
					"用户名": "admin",
					"邮箱":  "admin@example.com",
					"密码":  "123456",
				},
				expected: 0,
			},
			{
				name: "部分字段为空",
				fields: map[string]string{
					"用户名": "",
					"邮箱":  "admin@example.com",
					"密码":  "",
				},
				expected: 2,
			},
			{
				name: "所有字段都为空",
				fields: map[string]string{
					"用户名": "",
					"邮箱":  "",
					"密码":  "",
				},
				expected: 3,
			},
			{
				name: "字段包含空格",
				fields: map[string]string{
					"用户名": "   ",
					"邮箱":  "admin@example.com",
					"密码":  "123456",
				},
				expected: 1, // 只有空格的字段应该被视为空
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				errors := controller.ValidateRequired(tc.fields)
				assert.Len(t, errors, tc.expected, "测试用例: %s", tc.name)
			})
		}
	})
}

// 基准测试
func BenchmarkBaseControllerValidation(b *testing.B) {
	controller := NewBaseController()

	b.Run("邮箱验证", func(b *testing.B) {
		email := "test@example.com"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.ValidateEmail(email)
		}
	})

	b.Run("手机号验证", func(b *testing.B) {
		phone := "13812345678"
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.ValidatePhone(phone)
		}
	})

	b.Run("必填字段验证", func(b *testing.B) {
		fields := map[string]string{
			"用户名": "admin",
			"邮箱":  "admin@example.com",
			"密码":  "123456",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			controller.ValidateRequired(fields)
		}
	})
}

// 测试边界情况
func TestControllerEdgeCases(t *testing.T) {
	t.Run("nil controller处理", func(t *testing.T) {
		// 这个测试只是为了说明应该避免nil controller
		// 实际使用中不应该出现这种情况
		assert.True(t, true, "应该避免使用nil controller")
	})

	t.Run("未初始化的controller", func(t *testing.T) {
		ctrl := &BaseController{}

		// 初始化之前Data应该为nil
		assert.Nil(t, ctrl.Data)

		// 初始化后应该正常工作
		ctrl.Init()
		assert.NotNil(t, ctrl.Data)
		// logger字段已移除，现在使用单例日志系统
	})
}
