package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMD5Functions(t *testing.T) {
	t.Run("MD5生成正确的哈希值", func(t *testing.T) {
		input := []byte("hello world")
		expected := "5eb63bbbe01eeed093cb22bb8f5acdc3"
		result := MD5(input)
		assert.Equal(t, expected, result)
	})
	
	t.Run("MD5String生成正确的哈希值", func(t *testing.T) {
		input := "hello world"
		expected := "5eb63bbbe01eeed093cb22bb8f5acdc3"
		result := MD5String(input)
		assert.Equal(t, expected, result)
	})
	
	t.Run("空字符串MD5", func(t *testing.T) {
		expected := "d41d8cd98f00b204e9800998ecf8427e"
		result := MD5String("")
		assert.Equal(t, expected, result)
	})
}

func TestEmailValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"有效邮箱", "test@example.com", true},
		{"有效邮箱带数字", "user123@test.co.uk", true},
		{"有效邮箱带特殊字符", "user.name+tag@example.com", true},
		{"无效邮箱-缺少@", "testexample.com", false},
		{"无效邮箱-缺少域名", "test@", false},
		{"无效邮箱-缺少用户名", "@example.com", false},
		{"无效邮箱-多个@", "test@@example.com", false},
		{"无效邮箱-空字符串", "", false},
		{"无效邮箱-过长", strings.Repeat("a", 250) + "@example.com", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmail(tt.email)
			assert.Equal(t, tt.expected, result, "测试用例: %s", tt.name)
		})
	}
}

func TestEmailBytesValidation(t *testing.T) {
	tests := []struct {
		name     string
		email    []byte
		expected bool
	}{
		{"有效邮箱字节", []byte("test@example.com"), true},
		{"无效邮箱字节", []byte("invalid.email"), false},
		{"空字节数组", []byte{}, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmailBytes(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPhoneValidation(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{"有效手机号13开头", "13812345678", true},
		{"有效手机号15开头", "15912345678", true},
		{"有效手机号18开头", "18812345678", true},
		{"有效手机号19开头", "19912345678", true},
		{"无效手机号12开头", "12812345678", false},
		{"无效手机号长度不够", "1381234567", false},
		{"无效手机号长度过长", "138123456789", false},
		{"无效手机号包含字母", "1381234567a", false},
		{"空字符串", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPhone(tt.phone)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIDCardValidation(t *testing.T) {
	tests := []struct {
		name     string
		idCard   string
		expected bool
	}{
		{"有效身份证18位", "110101199003078515", true},
		{"有效身份证末位X", "11010119900307851X", true},
		{"无效身份证长度17位", "11010119900307851", false},
		{"无效身份证长度19位", "1101011990030785151", false},
		{"无效身份证年份", "110101200203078515", true}, // 2002年是有效的
		{"无效身份证月份", "110101199013078515", false}, // 13月无效
		{"无效身份证日期", "110101199002328515", false}, // 2月32日无效
		{"空字符串", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsIDCard(tt.idCard)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"有效用户名字母", "testuser", true},
		{"有效用户名数字", "user123", true},
		{"有效用户名下划线", "test_user", true},
		{"有效用户名混合", "test_user_123", true},
		{"无效用户名太短", "ab", false},
		{"无效用户名太长", strings.Repeat("a", 25), false},
		{"无效用户名特殊字符", "test-user", false},
		{"无效用户名空格", "test user", false},
		{"空字符串", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUsername(tt.username)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"有效密码6位", "123456", true},
		{"有效密码20位", "12345678901234567890", true},
		{"无效密码太短", "12345", false},
		{"无效密码太长", strings.Repeat("a", 25), false},
		{"空字符串", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStrongPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"强密码", "Aa1@2345", true},
		{"强密码复杂", "MyP@ssw0rd123", true},
		{"弱密码-缺少大写", "aa1@2345", false},
		{"弱密码-缺少小写", "AA1@2345", false},
		{"弱密码-缺少数字", "Aa@bcdef", false},
		{"弱密码-缺少特殊字符", "Aa123456", false},
		{"弱密码-长度不够", "Aa1@234", false},
		{"空字符串", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsStrongPassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWeakPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"弱密码-常见密码", "123456", true},
		{"弱密码-password", "password", true},
		{"弱密码-太短", "12345", true},
		{"弱密码-纯数字", "1234567890", true},
		{"弱密码-纯字母", "abcdefgh", true},
		{"强密码", "MyP@ssw0rd123", false},
		{"中等密码", "mypassword123", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWeakPassword(tt.password)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNumericValidation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"整数", "123", true},
		{"浮点数", "123.45", true},
		{"负数", "-123", true},
		{"负浮点数", "-123.45", true},
		{"科学计数法", "1.23e10", true},
		{"字母", "abc", false},
		{"混合", "123abc", false},
		{"空字符串", "", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNumeric(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCharacterTypeValidation(t *testing.T) {
	t.Run("IsAlpha测试", func(t *testing.T) {
		assert.True(t, IsAlpha("abc"))
		assert.True(t, IsAlpha("ABC"))
		assert.True(t, IsAlpha("AbC"))
		assert.False(t, IsAlpha("abc123"))
		assert.False(t, IsAlpha("abc "))
		assert.False(t, IsAlpha(""))
	})
	
	t.Run("IsAlphaNumeric测试", func(t *testing.T) {
		assert.True(t, IsAlphaNumeric("abc123"))
		assert.True(t, IsAlphaNumeric("ABC123"))
		assert.True(t, IsAlphaNumeric("abc"))
		assert.True(t, IsAlphaNumeric("123"))
		assert.False(t, IsAlphaNumeric("abc-123"))
		assert.False(t, IsAlphaNumeric("abc 123"))
		assert.False(t, IsAlphaNumeric(""))
	})
	
	t.Run("字符特征检测", func(t *testing.T) {
		password := "MyP@ssw0rd"
		assert.True(t, HasLowerCase(password))
		assert.True(t, HasUpperCase(password))
		assert.True(t, HasDigit(password))
		assert.True(t, HasSpecialChar(password))
	})
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"移除HTML标签",
			"<script>alert('xss')</script>Hello<b>World</b>",
			"alert(xss)HelloWorld",
		},
		{
			"移除SQL注入字符",
			"SELECT * FROM users WHERE id = '1'; DROP TABLE users;--",
			"SELECT * FROM users WHERE id = 1 DROP TABLE users",
		},
		{
			"移除JavaScript",
			"<script>alert('xss')</script>javascript:alert('xss')",
			"alert(xss)alert(xss)",
		},
		{
			"正常文本",
			"  Hello World  ",
			"Hello World",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTMLEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"转义HTML特殊字符",
			"<script>alert(\"Hello & World\");</script>",
			"&lt;script&gt;alert(&quot;Hello &amp; World&quot;);&lt;/script&gt;",
		},
		{
			"转义单引号",
			"It's a test",
			"It&#39;s a test",
		},
		{
			"正常文本",
			"Hello World",
			"Hello World",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
			
			// 测试反转义
			unescaped := UnescapeHTML(result)
			assert.Equal(t, tt.input, unescaped)
		})
	}
}

func TestFileValidation(t *testing.T) {
	t.Run("ValidateFileExtension", func(t *testing.T) {
		allowedExts := []string{"jpg", "png", "gif", "pdf"}
		
		assert.True(t, ValidateFileExtension("image.jpg", allowedExts))
		assert.True(t, ValidateFileExtension("image.JPG", allowedExts))
		assert.True(t, ValidateFileExtension("document.pdf", allowedExts))
		assert.False(t, ValidateFileExtension("script.exe", allowedExts))
		assert.False(t, ValidateFileExtension("noextension", allowedExts))
	})
	
	t.Run("GetFileExt", func(t *testing.T) {
		assert.Equal(t, ".jpg", GetFileExt("image.jpg"))
		assert.Equal(t, ".pdf", GetFileExt("document.pdf"))
		assert.Equal(t, "", GetFileExt("noextension"))
		assert.Equal(t, ".gz", GetFileExt("archive.tar.gz"))
	})
	
	t.Run("ValidateFileSize", func(t *testing.T) {
		maxSize := int64(1024 * 1024) // 1MB
		
		assert.True(t, ValidateFileSize(1024, maxSize))
		assert.True(t, ValidateFileSize(maxSize, maxSize))
		assert.False(t, ValidateFileSize(maxSize+1, maxSize))
		assert.False(t, ValidateFileSize(0, maxSize))
		assert.False(t, ValidateFileSize(-1, maxSize))
	})
}

func TestValidationResult(t *testing.T) {
	t.Run("创建和使用ValidationResult", func(t *testing.T) {
		vr := NewValidationResult()
		
		assert.True(t, vr.Valid)
		assert.Empty(t, vr.Errors)
		assert.False(t, vr.HasErrors())
		assert.Equal(t, "", vr.GetFirstError())
		
		// 添加错误
		vr.AddError("用户名不能为空")
		vr.AddError("密码长度不够")
		
		assert.False(t, vr.Valid)
		assert.True(t, vr.HasErrors())
		assert.Len(t, vr.Errors, 2)
		assert.Equal(t, "用户名不能为空", vr.GetFirstError())
	})
}

func TestValidationHelpers(t *testing.T) {
	t.Run("ValidateLength", func(t *testing.T) {
		assert.True(t, ValidateLength("hello", 3, 10))
		assert.True(t, ValidateLength("hello", 5, 5))
		assert.False(t, ValidateLength("hello", 6, 10))
		assert.False(t, ValidateLength("hello", 1, 4))
	})
	
	t.Run("ValidateRange", func(t *testing.T) {
		assert.True(t, ValidateRange(5.0, 1.0, 10.0))
		assert.True(t, ValidateRange(1.0, 1.0, 10.0))
		assert.True(t, ValidateRange(10.0, 1.0, 10.0))
		assert.False(t, ValidateRange(0.5, 1.0, 10.0))
		assert.False(t, ValidateRange(10.5, 1.0, 10.0))
	})
	
	t.Run("ValidateIn", func(t *testing.T) {
		options := []string{"red", "green", "blue"}
		
		assert.True(t, ValidateIn("red", options))
		assert.True(t, ValidateIn("green", options))
		assert.False(t, ValidateIn("yellow", options))
		assert.False(t, ValidateIn("", options))
	})
	
	t.Run("ValidateNotIn", func(t *testing.T) {
		options := []string{"admin", "root", "system"}
		
		assert.True(t, ValidateNotIn("user", options))
		assert.False(t, ValidateNotIn("admin", options))
	})
	
	t.Run("ValidateRegex", func(t *testing.T) {
		pattern := `^\d{4}-\d{2}-\d{2}$` // YYYY-MM-DD 日期格式
		
		assert.True(t, ValidateRegex("2024-01-01", pattern))
		assert.True(t, ValidateRegex("2023-12-31", pattern))
		assert.False(t, ValidateRegex("24-1-1", pattern))
		assert.False(t, ValidateRegex("invalid-date", pattern))
	})
}

// 性能测试
func BenchmarkIsEmail(b *testing.B) {
	email := "test.user+tag@example.com"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsEmail(email)
	}
}

func BenchmarkIsPhone(b *testing.B) {
	phone := "13812345678"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsPhone(phone)
	}
}

func BenchmarkIsStrongPassword(b *testing.B) {
	password := "MyStr0ng!P@ssw0rd"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsStrongPassword(password)
	}
}

func BenchmarkSanitizeString(b *testing.B) {
	input := "<script>alert('test')</script>SELECT * FROM users WHERE id='1';"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeString(input)
	}
}