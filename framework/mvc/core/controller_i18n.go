package core

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 国际化基础方法 =============

// SetLanguage 设置当前语言
func (c *BaseController) SetLanguage(lang string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to set language")
		return
	}

	// 标准化语言代码（转小写并处理下划线）
	lang = strings.ToLower(strings.Replace(lang, "_", "-", -1))

	// 存储在上下文数据中
	c.SetData("current_language", lang)

	// 设置响应头
	c.SetHeader("Content-Language", lang)
}

// GetLanguage 获取当前语言
func (c *BaseController) GetLanguage() string {
	if lang, ok := c.GetData("current_language").(string); ok {
		return lang
	}

	// 从Accept-Language头部检测
	if detectedLang := c.DetectLanguageFromHeader(); detectedLang != "" {
		return detectedLang
	}

	// 返回默认语言
	return c.GetDefaultLanguage()
}

// GetDefaultLanguage 获取默认语言
func (c *BaseController) GetDefaultLanguage() string {
	// 可以从配置中读取，这里硬编码为中文
	return "zh-cn"
}

// DetectLanguageFromHeader 从Accept-Language头部检测语言
func (c *BaseController) DetectLanguageFromHeader() string {
	if c.Ctx == nil {
		return ""
	}

	acceptLang := c.GetHeader("Accept-Language")
	if acceptLang == "" {
		return ""
	}

	// 解析Accept-Language头部
	languages := c.parseAcceptLanguage(acceptLang)

	// 返回权重最高的支持语言
	supportedLangs := c.GetSupportedLanguages()
	for _, lang := range languages {
		for _, supported := range supportedLangs {
			if c.isLanguageMatch(lang.Code, supported) {
				return supported
			}
		}
	}

	return ""
}

// LanguagePreference Accept-Language的语言偏好结构
type LanguagePreference struct {
	Code   string
	Weight float64
}

// parseAcceptLanguage 解析Accept-Language头部
func (c *BaseController) parseAcceptLanguage(acceptLang string) []LanguagePreference {
	var preferences []LanguagePreference

	parts := strings.Split(acceptLang, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		var code string
		var weight float64 = 1.0

		if strings.Contains(part, ";q=") {
			segments := strings.Split(part, ";q=")
			code = strings.TrimSpace(segments[0])
			if w, err := strconv.ParseFloat(strings.TrimSpace(segments[1]), 64); err == nil {
				weight = w
			}
		} else {
			code = part
		}

		// 标准化语言代码
		code = strings.ToLower(strings.Replace(code, "_", "-", -1))

		preferences = append(preferences, LanguagePreference{
			Code:   code,
			Weight: weight,
		})
	}

	// 按权重排序
	sort.Slice(preferences, func(i, j int) bool {
		return preferences[i].Weight > preferences[j].Weight
	})

	return preferences
}

// isLanguageMatch 检查语言是否匹配
func (c *BaseController) isLanguageMatch(requested, supported string) bool {
	requested = strings.ToLower(requested)
	supported = strings.ToLower(supported)

	// 完全匹配
	if requested == supported {
		return true
	}

	// 语言主代码匹配（如zh匹配zh-cn）
	reqParts := strings.Split(requested, "-")
	supParts := strings.Split(supported, "-")

	return reqParts[0] == supParts[0]
}

// ============= 消息翻译方法 =============

// T 翻译消息（主要方法）
func (c *BaseController) T(key string, args ...any) string {
	return c.Translate(key, args...)
}

// Translate 翻译消息
func (c *BaseController) Translate(key string, args ...any) string {
	lang := c.GetLanguage()

	// 获取翻译消息
	message := c.getTranslationMessage(lang, key)

	// 如果找不到翻译，尝试使用默认语言
	if message == "" {
		defaultLang := c.GetDefaultLanguage()
		if lang != defaultLang {
			message = c.getTranslationMessage(defaultLang, key)
		}
	}

	// 如果仍然找不到，返回键名
	if message == "" {
		message = key
	}

	// 格式化消息
	if len(args) > 0 {
		return fmt.Sprintf(message, args...)
	}

	return message
}

// getTranslationMessage 获取翻译消息（需要实现具体的翻译存储逻辑）
func (c *BaseController) getTranslationMessage(lang, key string) string {
	// 这里应该从翻译文件或数据库中获取翻译
	// 示例实现：使用内存映射
	translations := c.getTranslations(lang)
	if msg, exists := translations[key]; exists {
		return msg
	}
	return ""
}

// getTranslations 获取指定语言的所有翻译（示例实现）
func (c *BaseController) getTranslations(lang string) map[string]string {
	// 这里应该实现真正的翻译加载逻辑
	// 示例数据
	translations := make(map[string]string)

	switch lang {
	case "zh-cn":
		translations["welcome"] = "欢迎"
		translations["hello"] = "你好"
		translations["goodbye"] = "再见"
		translations["error.not_found"] = "未找到"
		translations["error.forbidden"] = "禁止访问"
		translations["success.saved"] = "保存成功"
	case "en-us", "en":
		translations["welcome"] = "Welcome"
		translations["hello"] = "Hello"
		translations["goodbye"] = "Goodbye"
		translations["error.not_found"] = "Not Found"
		translations["error.forbidden"] = "Forbidden"
		translations["success.saved"] = "Saved Successfully"
	case "ja-jp", "ja":
		translations["welcome"] = "いらっしゃいませ"
		translations["hello"] = "こんにちは"
		translations["goodbye"] = "さようなら"
		translations["error.not_found"] = "見つかりません"
		translations["error.forbidden"] = "禁止されています"
		translations["success.saved"] = "保存成功"
	}

	return translations
}

// ============= 语言环境设置方法 =============

// SetLocale 设置地区（语言+地区）
func (c *BaseController) SetLocale(locale string) {
	c.SetData("current_locale", locale)

	// 从locale中提取语言部分
	if lang := c.extractLanguageFromLocale(locale); lang != "" {
		c.SetLanguage(lang)
	}
}

// GetLocale 获取当前地区
func (c *BaseController) GetLocale() string {
	if locale, ok := c.GetData("current_locale").(string); ok {
		return locale
	}

	// 基于语言构建默认locale
	lang := c.GetLanguage()
	return c.buildDefaultLocale(lang)
}

// extractLanguageFromLocale 从locale中提取语言
func (c *BaseController) extractLanguageFromLocale(locale string) string {
	// locale格式通常是 zh_CN, en_US, ja_JP等
	if parts := strings.Split(locale, "_"); len(parts) > 0 {
		return strings.ToLower(parts[0])
	}
	return strings.ToLower(locale)
}

// buildDefaultLocale 基于语言构建默认locale
func (c *BaseController) buildDefaultLocale(lang string) string {
	localeMap := map[string]string{
		"zh":    "zh_CN",
		"zh-cn": "zh_CN",
		"en":    "en_US",
		"en-us": "en_US",
		"ja":    "ja_JP",
		"ja-jp": "ja_JP",
		"ko":    "ko_KR",
		"ko-kr": "ko_KR",
		"fr":    "fr_FR",
		"fr-fr": "fr_FR",
		"de":    "de_DE",
		"de-de": "de_DE",
	}

	if locale, exists := localeMap[lang]; exists {
		return locale
	}

	return "en_US" // 默认值
}

// ============= 数字和日期格式化方法 =============

// FormatNumber 格式化数字
func (c *BaseController) FormatNumber(number float64, precision int) string {
	locale := c.GetLocale()

	// 根据locale格式化数字
	switch {
	case strings.HasPrefix(locale, "zh"):
		// 中文：使用逗号分隔千位
		return c.formatNumberWithSeparator(number, precision, ",", ".")
	case strings.HasPrefix(locale, "en"):
		// 英文：使用逗号分隔千位，点作小数分隔符
		return c.formatNumberWithSeparator(number, precision, ",", ".")
	case strings.HasPrefix(locale, "de"):
		// 德语：使用点分隔千位，逗号作小数分隔符
		return c.formatNumberWithSeparator(number, precision, ".", ",")
	case strings.HasPrefix(locale, "fr"):
		// 法语：使用空格分隔千位，逗号作小数分隔符
		return c.formatNumberWithSeparator(number, precision, " ", ",")
	default:
		return c.formatNumberWithSeparator(number, precision, ",", ".")
	}
}

// formatNumberWithSeparator 使用指定分隔符格式化数字
func (c *BaseController) formatNumberWithSeparator(number float64, precision int, thousandSep, decimalSep string) string {
	// 简单实现，实际项目中建议使用专业的本地化库
	format := fmt.Sprintf("%%.%df", precision)
	str := fmt.Sprintf(format, number)

	parts := strings.Split(str, ".")

	// 处理整数部分的千位分隔符
	intPart := parts[0]
	if len(intPart) > 3 {
		var result []string
		for len(intPart) > 3 {
			result = append([]string{intPart[len(intPart)-3:]}, result...)
			intPart = intPart[:len(intPart)-3]
		}
		result = append([]string{intPart}, result...)
		intPart = strings.Join(result, thousandSep)
	}

	if len(parts) > 1 && precision > 0 {
		return intPart + decimalSep + parts[1]
	}

	return intPart
}

// FormatCurrency 格式化货币
func (c *BaseController) FormatCurrency(amount float64) string {
	locale := c.GetLocale()

	switch {
	case strings.HasPrefix(locale, "zh"):
		return "¥" + c.FormatNumber(amount, 2)
	case strings.HasPrefix(locale, "en_US"):
		return "$" + c.FormatNumber(amount, 2)
	case strings.HasPrefix(locale, "en_GB"):
		return "£" + c.FormatNumber(amount, 2)
	case strings.HasPrefix(locale, "ja"):
		return "¥" + c.FormatNumber(amount, 0)
	case strings.HasPrefix(locale, "de"):
		return c.FormatNumber(amount, 2) + " €"
	case strings.HasPrefix(locale, "fr"):
		return c.FormatNumber(amount, 2) + " €"
	default:
		return "$" + c.FormatNumber(amount, 2)
	}
}

// FormatDate 格式化日期
func (c *BaseController) FormatDate(t time.Time) string {
	locale := c.GetLocale()

	switch {
	case strings.HasPrefix(locale, "zh"):
		return t.Format("2006年01月02日")
	case strings.HasPrefix(locale, "en_US"):
		return t.Format("01/02/2006")
	case strings.HasPrefix(locale, "en_GB"):
		return t.Format("02/01/2006")
	case strings.HasPrefix(locale, "ja"):
		return t.Format("2006年01月02日")
	case strings.HasPrefix(locale, "de"):
		return t.Format("02.01.2006")
	case strings.HasPrefix(locale, "fr"):
		return t.Format("02/01/2006")
	default:
		return t.Format("2006-01-02")
	}
}

// FormatDateTime 格式化日期时间
func (c *BaseController) FormatDateTime(t time.Time) string {
	locale := c.GetLocale()

	switch {
	case strings.HasPrefix(locale, "zh"):
		return t.Format("2006年01月02日 15:04:05")
	case strings.HasPrefix(locale, "en_US"):
		return t.Format("01/02/2006 3:04:05 PM")
	case strings.HasPrefix(locale, "en_GB"):
		return t.Format("02/01/2006 15:04:05")
	case strings.HasPrefix(locale, "ja"):
		return t.Format("2006年01月02日 15:04:05")
	case strings.HasPrefix(locale, "de"):
		return t.Format("02.01.2006 15:04:05")
	case strings.HasPrefix(locale, "fr"):
		return t.Format("02/01/2006 15:04:05")
	default:
		return t.Format("2006-01-02 15:04:05")
	}
}

// ============= 语言切换和管理方法 =============

// GetSupportedLanguages 获取支持的语言列表
func (c *BaseController) GetSupportedLanguages() []string {
	// 这里应该从配置中读取，示例硬编码
	return []string{"zh-cn", "en-us", "ja-jp", "ko-kr", "fr-fr", "de-de"}
}

// IsSupportedLanguage 检查是否为支持的语言
func (c *BaseController) IsSupportedLanguage(lang string) bool {
	lang = strings.ToLower(lang)
	supportedLangs := c.GetSupportedLanguages()

	for _, supported := range supportedLangs {
		if c.isLanguageMatch(lang, supported) {
			return true
		}
	}

	return false
}

// GetLanguageName 获取语言的本地化名称
func (c *BaseController) GetLanguageName(langCode string) string {
	langNames := map[string]string{
		"zh-cn": "简体中文",
		"zh-tw": "繁體中文",
		"en-us": "English",
		"en-gb": "English (UK)",
		"ja-jp": "日本語",
		"ko-kr": "한국어",
		"fr-fr": "Français",
		"de-de": "Deutsch",
		"es-es": "Español",
		"it-it": "Italiano",
		"pt-pt": "Português",
		"ru-ru": "Русский",
	}

	if name, exists := langNames[strings.ToLower(langCode)]; exists {
		return name
	}

	return langCode
}

// SwitchLanguage 切换语言并重定向
func (c *BaseController) SwitchLanguage(lang string, redirectURL ...string) {
	if !c.IsSupportedLanguage(lang) {
		config.Warnf("Unsupported language: %s", lang)
		return
	}

	c.SetLanguage(lang)

	// 设置语言Cookie，便于下次访问记住用户偏好
	c.SetLanguageCookie(lang)

	// 重定向到指定URL或当前页面
	var url string
	if len(redirectURL) > 0 && redirectURL[0] != "" {
		url = redirectURL[0]
	} else {
		url = c.GetCurrentURL()
	}

	c.Redirect(url)
}

// SetLanguageCookie 设置语言偏好Cookie
func (c *BaseController) SetLanguageCookie(lang string) {
	// 设置1年有效期的语言Cookie
	expiry := time.Now().AddDate(1, 0, 0)
	maxAge := int(time.Until(expiry).Seconds())

	// 使用临时方法，避免导入cookie包
	if c.Ctx != nil && c.Ctx.Request() != nil {
		cookieStr := fmt.Sprintf("language=%s; Max-Age=%d; Path=/; HttpOnly", lang, maxAge)
		c.SetHeader("Set-Cookie", cookieStr)
	}
}

// GetLanguageFromCookie 从Cookie获取语言偏好
func (c *BaseController) GetLanguageFromCookie() string {
	return c.GetCookie("language")
}

// ============= 辅助方法 =============

// GetCurrentURL 获取当前URL
func (c *BaseController) GetCurrentURL() string {
	if c.Ctx == nil {
		return "/"
	}
	return c.Ctx.Request().URI().String()
}

// GetLanguageDirection 获取语言方向（LTR或RTL）
func (c *BaseController) GetLanguageDirection() string {
	lang := c.GetLanguage()

	// RTL语言列表
	rtlLanguages := []string{"ar", "he", "fa", "ur"}

	langCode := strings.Split(lang, "-")[0]
	for _, rtlLang := range rtlLanguages {
		if langCode == rtlLang {
			return "rtl"
		}
	}

	return "ltr"
}

// BuildLocalizedURL 构建本地化URL
func (c *BaseController) BuildLocalizedURL(path string, lang ...string) string {
	var targetLang string
	if len(lang) > 0 {
		targetLang = lang[0]
	} else {
		targetLang = c.GetLanguage()
	}

	// 简单实现：在路径前添加语言代码
	return fmt.Sprintf("/%s%s", targetLang, path)
}

// GetTranslationFile 获取翻译文件路径
func (c *BaseController) GetTranslationFile(lang string) string {
	// 构建翻译文件路径
	return filepath.Join("locales", lang+".json")
}
