package validation

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ============= 格式验证器实现 =============

// validateEmail 邮箱验证
func (v *Validator) validateEmail(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	_, err := mail.ParseAddress(str)
	return err == nil
}

// validateURL URL验证
func (v *Validator) validateURL(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	_, err := url.ParseRequestURI(str)
	return err == nil
}

// validateAlpha 字母验证
func (v *Validator) validateAlpha(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	matched, _ := regexp.MatchString("^[a-zA-Z]+$", str)
	return matched
}

// validateAlphaNum 字母数字验证
func (v *Validator) validateAlphaNum(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", str)
	return matched
}

// validateNumeric 数字验证
func (v *Validator) validateNumeric(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

// validateInteger 整数验证
func (v *Validator) validateInteger(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

// validateDecimal 小数验证
func (v *Validator) validateDecimal(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	if param != "" {
		// 指定小数位数
		parts := strings.Split(str, ".")
		if len(parts) != 2 {
			return false
		}

		decimalPlaces, err := strconv.Atoi(param)
		if err != nil {
			return false
		}

		if len(parts[1]) != decimalPlaces {
			return false
		}
	}

	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

// ============= IP地址验证器 =============

// validateIP IP地址验证
func (v *Validator) validateIP(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	return net.ParseIP(str) != nil
}

// validateIPv4 IPv4地址验证
func (v *Validator) validateIPv4(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	ip := net.ParseIP(str)
	return ip != nil && ip.To4() != nil
}

// validateIPv6 IPv6地址验证
func (v *Validator) validateIPv6(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	ip := net.ParseIP(str)
	return ip != nil && ip.To4() == nil
}

// validateMAC MAC地址验证
func (v *Validator) validateMAC(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	_, err := net.ParseMAC(str)
	return err == nil
}

// validateUUID UUID验证
func (v *Validator) validateUUID(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// UUID格式: xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(str)
}

// ============= 日期时间验证器 =============

// validateDate 日期验证
func (v *Validator) validateDate(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	layout := "2006-01-02"
	if param != "" {
		layout = param
	}

	_, err := time.Parse(layout, str)
	return err == nil
}

// validateDateTime 日期时间验证
func (v *Validator) validateDateTime(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	layout := "2006-01-02 15:04:05"
	if param != "" {
		layout = param
	}

	_, err := time.Parse(layout, str)
	return err == nil
}

// validateTime 时间验证
func (v *Validator) validateTime(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	layout := "15:04:05"
	if param != "" {
		layout = param
	}

	_, err := time.Parse(layout, str)
	return err == nil
}

// validateBefore 日期早于验证
func (v *Validator) validateBefore(value any, param string) bool {
	str := v.toString(value)
	if str == "" || param == "" {
		return false
	}

	layout := "2006-01-02"

	// 解析目标日期
	targetTime, err := time.Parse(layout, str)
	if err != nil {
		return false
	}

	// 解析比较日期
	compareTime, err := time.Parse(layout, param)
	if err != nil {
		return false
	}

	return targetTime.Before(compareTime)
}

// validateAfter 日期晚于验证
func (v *Validator) validateAfter(value any, param string) bool {
	str := v.toString(value)
	if str == "" || param == "" {
		return false
	}

	layout := "2006-01-02"

	// 解析目标日期
	targetTime, err := time.Parse(layout, str)
	if err != nil {
		return false
	}

	// 解析比较日期
	compareTime, err := time.Parse(layout, param)
	if err != nil {
		return false
	}

	return targetTime.After(compareTime)
}

// ============= 字符串验证器 =============

// validateRegex 正则表达式验证
func (v *Validator) validateRegex(value any, param string) bool {
	str := v.toString(value)
	if str == "" || param == "" {
		return false
	}

	matched, err := regexp.MatchString(param, str)
	return err == nil && matched
}

// validateContains 包含验证
func (v *Validator) validateContains(value any, param string) bool {
	str := v.toString(value)
	if str == "" || param == "" {
		return false
	}

	return strings.Contains(str, param)
}

// validateStartsWith 开头验证
func (v *Validator) validateStartsWith(value any, param string) bool {
	str := v.toString(value)
	if str == "" || param == "" {
		return false
	}

	return strings.HasPrefix(str, param)
}

// validateEndsWith 结尾验证
func (v *Validator) validateEndsWith(value any, param string) bool {
	str := v.toString(value)
	if str == "" || param == "" {
		return false
	}

	return strings.HasSuffix(str, param)
}

// validateIn 枚举验证
func (v *Validator) validateIn(value any, param string) bool {
	str := v.toString(value)
	if str == "" || param == "" {
		return false
	}

	options := strings.Split(param, ",")
	for _, option := range options {
		if strings.TrimSpace(option) == str {
			return true
		}
	}

	return false
}

// validateNotIn 不在枚举中验证
func (v *Validator) validateNotIn(value any, param string) bool {
	return !v.validateIn(value, param)
}

// ============= 数字验证器 =============

// validatePositive 正数验证
func (v *Validator) validatePositive(value any, param string) bool {
	val := v.toFloat64(value)
	if val == nil {
		return false
	}

	return *val > 0
}

// validateNegative 负数验证
func (v *Validator) validateNegative(value any, param string) bool {
	val := v.toFloat64(value)
	if val == nil {
		return false
	}

	return *val < 0
}

// validateNonZero 非零验证
func (v *Validator) validateNonZero(value any, param string) bool {
	val := v.toFloat64(value)
	if val == nil {
		return false
	}

	return *val != 0
}

// ============= 中国特有验证器 =============

// validateMobile 手机号码验证
func (v *Validator) validateMobile(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// 中国手机号码格式：1[3-9]\d{9}
	mobileRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	return mobileRegex.MatchString(str)
}

// validatePhone 固定电话验证
func (v *Validator) validatePhone(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// 中国固定电话格式：区号-号码，如：010-12345678
	phoneRegex := regexp.MustCompile(`^0\d{2,3}-\d{7,8}$`)
	return phoneRegex.MatchString(str)
}

// validateIDCard 身份证号码验证
func (v *Validator) validateIDCard(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// 简化的身份证验证，只验证长度和格式
	if len(str) == 15 {
		// 15位身份证
		matched, _ := regexp.MatchString(`^\d{15}$`, str)
		return matched
	} else if len(str) == 18 {
		// 18位身份证
		matched, _ := regexp.MatchString(`^\d{17}[\dXx]$`, str)
		return matched
	}

	return false
}

// validateZipCode 邮政编码验证
func (v *Validator) validateZipCode(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// 中国邮政编码格式：6位数字
	zipCodeRegex := regexp.MustCompile(`^\d{6}$`)
	return zipCodeRegex.MatchString(str)
}

// ============= 辅助方法 =============

// toString 转换为字符串
func (v *Validator) toString(value any) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return ""
	}
}

// ============= 高级验证器 =============

// ValidateCard 银行卡号验证（Luhn算法）
func (v *Validator) validateCard(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// 移除空格和连字符
	str = strings.ReplaceAll(str, " ", "")
	str = strings.ReplaceAll(str, "-", "")

	// 检查是否都是数字
	matched, _ := regexp.MatchString(`^\d+$`, str)
	if !matched {
		return false
	}

	// Luhn算法验证
	return v.luhnCheck(str)
}

// luhnCheck Luhn算法检查
func (v *Validator) luhnCheck(cardNumber string) bool {
	var sum int
	alternate := false

	// 从右到左遍历
	for i := len(cardNumber) - 1; i >= 0; i-- {
		digit := int(cardNumber[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// ValidateJSON JSON格式验证
func (v *Validator) validateJSON(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// 简单的JSON格式检查
	str = strings.TrimSpace(str)
	return (strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")) ||
		(strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]"))
}

// ValidateBase64 Base64格式验证
func (v *Validator) validateBase64(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// Base64格式验证
	base64Regex := regexp.MustCompile(`^[A-Za-z0-9+/]*={0,2}$`)
	return base64Regex.MatchString(str) && len(str)%4 == 0
}

// ValidateHexColor 十六进制颜色验证
func (v *Validator) validateHexColor(value any, param string) bool {
	str := v.toString(value)
	if str == "" {
		return false
	}

	// 十六进制颜色格式：#RRGGBB 或 #RGB
	hexColorRegex := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	return hexColorRegex.MatchString(str)
}
