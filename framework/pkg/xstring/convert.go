package xstring

import (
	"html"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ToLower converts string to lowercase
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper converts string to uppercase
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// Ucfirst makes a string's first character uppercase
func Ucfirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

// Ucwords uppercase the first character of each word
func Ucwords(s string) string {
	return strings.Title(s)
}

// CapitalizeFirst 首字母大写（处理Unicode字符）
func CapitalizeFirst(str string) string {
	if str == "" {
		return ""
	}

	runes := []rune(str)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// Trim strips whitespace from the beginning and end of a string
func Trim(s string, chars ...string) string {
	if len(chars) == 0 {
		return strings.TrimSpace(s)
	}
	return strings.Trim(s, chars[0])
}

// Ltrim strips whitespace from the beginning of a string
func Ltrim(s string, chars ...string) string {
	if len(chars) == 0 {
		return strings.TrimLeftFunc(s, unicode.IsSpace)
	}
	return strings.TrimLeft(s, chars[0])
}

// Rtrim strips whitespace from the end of a string
func Rtrim(s string, chars ...string) string {
	if len(chars) == 0 {
		return strings.TrimRightFunc(s, unicode.IsSpace)
	}
	return strings.TrimRight(s, chars[0])
}

// Htmlspecialchars converts special characters to HTML entities
func Htmlspecialchars(s string) string {
	return html.EscapeString(s)
}

// HtmlspecialcharsDecode converts special HTML entities back to characters
func HtmlspecialcharsDecode(s string) string {
	return html.UnescapeString(s)
}

// Nl2br inserts HTML line breaks before all newlines
func Nl2br(s string) string {
	return strings.ReplaceAll(s, "\n", "<br />\n")
}

// StripTags strips HTML from a string
func StripTags(s string) string {
	re := strings.NewReplacer("<", "&lt;", ">", "&gt;")
	return re.Replace(s)
}

// ToTitleCase 转换为标题格式（替代strings.Title）
func ToTitleCase(s string) string {
	if s == "" {
		return s
	}

	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}
