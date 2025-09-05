package xvalidation

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strconv"
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
