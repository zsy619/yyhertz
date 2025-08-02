package util

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// 常用正则表达式
var (
	emailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phonePattern    = regexp.MustCompile(`^1[3-9]\d{9}$`)
	idCardPattern   = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}((0[1-9])|(1[0-2]))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$`)
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
	passwordPattern = regexp.MustCompile(`^.{6,20}$`)
	urlPattern      = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	ipPattern       = regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
)

// MD5 生成MD5哈希值
func MD5(data []byte) string {
	hash := md5.New()
	hash.Write(data)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// MD5String 生成字符串的MD5哈希值
func MD5String(s string) string {
	return MD5([]byte(s))
}

// IsEmail 验证邮箱格式
func IsEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailPattern.MatchString(email)
}

// IsEmailBytes 验证邮箱格式(字节数组)
func IsEmailBytes(b []byte) bool {
	return emailPattern.Match(b)
}

// IsPhone 验证手机号格式
func IsPhone(phone string) bool {
	return phonePattern.MatchString(phone)
}

// IsIDCard 验证身份证号格式
func IsIDCard(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}
	return idCardPattern.MatchString(idCard)
}

// IsUsername 验证用户名格式
func IsUsername(username string) bool {
	return usernamePattern.MatchString(username)
}

// IsPassword 验证密码格式
func IsPassword(password string) bool {
	return passwordPattern.MatchString(password)
}

// IsURL 验证URL格式
func IsURL(url string) bool {
	return urlPattern.MatchString(url)
}

// IsIPAddress 验证IP地址格式
func IsIPAddress(ip string) bool {
	return ipPattern.MatchString(ip)
}

// IsInteger 检查字符串是否为整数
func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// IsAlpha 检查字符串是否只包含字母
func IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return s != ""
}

// IsAlphaNumeric 检查字符串是否只包含字母和数字
func IsAlphaNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

// IsLowerCase 检查字符串是否全为小写
func IsLowerCase(s string) bool {
	return s == strings.ToLower(s) && s != ""
}

// IsUpperCase 检查字符串是否全为大写
func IsUpperCase(s string) bool {
	return s == strings.ToUpper(s) && s != ""
}

// HasLowerCase 检查字符串是否包含小写字母
func HasLowerCase(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

// HasUpperCase 检查字符串是否包含大写字母
func HasUpperCase(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// HasDigit 检查字符串是否包含数字
func HasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// HasSpecialChar 检查字符串是否包含特殊字符
func HasSpecialChar(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// IsStrongPassword 检查是否为强密码(至少8位，包含大小写字母、数字和特殊字符)
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	
	return HasLowerCase(password) && 
		   HasUpperCase(password) && 
		   HasDigit(password) && 
		   HasSpecialChar(password)
}

// IsWeakPassword 检查是否为弱密码
func IsWeakPassword(password string) bool {
	if len(password) < 6 {
		return true
	}
	
	// 检查常见弱密码
	weakPasswords := []string{
		"123456", "password", "123456789", "12345678", "12345", "1234567",
		"qwerty", "abc123", "111111", "123123", "admin", "letmein",
		"welcome", "monkey", "1234567890", "password123", "123abc",
	}
	
	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if lowerPassword == weak {
			return true
		}
	}
	
	// 检查是否只包含数字或字母
	if IsNumeric(password) || IsAlpha(password) {
		return true
	}
	
	return false
}

// ValidateLength 验证字符串长度
func ValidateLength(s string, min, max int) bool {
	length := len(s)
	return length >= min && length <= max
}

// ValidateRange 验证数字范围
func ValidateRange(value, min, max float64) bool {
	return value >= min && value <= max
}

// ValidateIn 验证值是否在给定的选项中
func ValidateIn(value string, options []string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}

// ValidateNotIn 验证值是否不在给定的选项中
func ValidateNotIn(value string, options []string) bool {
	return !ValidateIn(value, options)
}

// ValidateRegex 使用正则表达式验证
func ValidateRegex(s, pattern string) bool {
	matched, err := regexp.MatchString(pattern, s)
	return err == nil && matched
}

// SanitizeString 清理字符串，移除危险字符
func SanitizeString(s string) string {
	// 移除HTML标签
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	s = htmlRegex.ReplaceAllString(s, "")
	
	// 移除SQL注入相关字符
	sqlChars := []string{"'", "\"", ";", "--", "/*", "*/", "xp_", "sp_"}
	for _, char := range sqlChars {
		s = strings.ReplaceAll(s, char, "")
	}
	
	// 移除JavaScript相关字符（不改变原字符串大小写）
	jsChars := []string{"<script", "</script>", "javascript:", "onload=", "onerror="}
	lowerS := strings.ToLower(s)
	for _, char := range jsChars {
		if strings.Contains(lowerS, char) {
			// 使用正则表达式进行大小写不敏感的替换
			regex := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(char))
			s = regex.ReplaceAllString(s, "")
			lowerS = strings.ToLower(s) // 更新小写版本
		}
	}
	
	return strings.TrimSpace(s)
}

// EscapeHTML 转义HTML特殊字符
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// UnescapeHTML 反转义HTML特殊字符
func UnescapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&amp;", "&")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&#39;", "'")
	return s
}

// ValidateFileExtension 验证文件扩展名
func ValidateFileExtension(filename string, allowedExts []string) bool {
	ext := strings.ToLower(GetFileExt(filename))
	if ext != "" && ext[0] == '.' {
		ext = ext[1:] // 移除点号
	}
	
	for _, allowedExt := range allowedExts {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}
	return false
}

// GetFileExt 获取文件扩展名
func GetFileExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
		if filename[i] == '/' || filename[i] == '\\' {
			break
		}
	}
	return ""
}

// ValidateFileSize 验证文件大小
func ValidateFileSize(size, maxSize int64) bool {
	return size > 0 && size <= maxSize
}

// ValidationResult 验证结果结构
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// NewValidationResult 创建验证结果
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: make([]string, 0),
	}
}

// AddError 添加错误信息
func (vr *ValidationResult) AddError(error string) {
	vr.Valid = false
	vr.Errors = append(vr.Errors, error)
}

// HasErrors 是否有错误
func (vr *ValidationResult) HasErrors() bool {
	return len(vr.Errors) > 0
}

// GetFirstError 获取第一个错误
func (vr *ValidationResult) GetFirstError() string {
	if len(vr.Errors) > 0 {
		return vr.Errors[0]
	}
	return ""
}