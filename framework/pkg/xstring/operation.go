package xstring

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Strlen returns the length of a string in bytes
func Strlen(s string) int {
	return len(s)
}

// MbStrlen returns the length of a string in characters
func MbStrlen(s string) int {
	return utf8.RuneCountInString(s)
}

// Substr returns part of a string
func Substr(s string, start int, length ...int) string {
	runes := []rune(s)
	size := len(runes)

	if start < 0 {
		start = size + start
	}
	if start < 0 || start > size {
		return ""
	}

	if len(length) == 0 {
		return string(runes[start:])
	}

	end := start + length[0]
	if end > size {
		end = size
	}
	if end < start {
		return ""
	}

	return string(runes[start:end])
}

// Strpos finds the position of the first occurrence of a substring
func Strpos(haystack, needle string, offset ...int) int {
	start := 0
	if len(offset) > 0 && offset[0] > 0 {
		start = offset[0]
	}

	if start >= len(haystack) {
		return -1
	}

	pos := strings.Index(haystack[start:], needle)
	if pos == -1 {
		return -1
	}
	return start + pos
}

// Strrpos finds the position of the last occurrence of a substring
func Strrpos(haystack, needle string) int {
	return strings.LastIndex(haystack, needle)
}

// Explode splits a string by a string
func Explode(delimiter, s string, limit ...int) []string {
	if len(limit) == 0 {
		return strings.Split(s, delimiter)
	}
	return strings.SplitN(s, delimiter, limit[0])
}

// Implode joins array elements with a string
func Implode(glue string, pieces []string) string {
	return strings.Join(pieces, glue)
}

// StrReplace replaces all occurrences of the search string with the replacement string
func StrReplace(search, replace, subject string, count ...int) string {
	if len(count) == 0 {
		return strings.ReplaceAll(subject, search, replace)
	}
	return strings.Replace(subject, search, replace, count[0])
}

// PregReplace performs a regular expression search and replace
func PregReplace(pattern, replacement, subject string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(subject, replacement), nil
}

// PregMatch performs a regular expression match
func PregMatch(pattern, subject string) (bool, error) {
	matched, err := regexp.MatchString(pattern, subject)
	return matched, err
}

// StrPad pads a string to a certain length with another string
func StrPad(input string, padLength int, padString string, padType int) string {
	const (
		STR_PAD_RIGHT = iota
		STR_PAD_LEFT
		STR_PAD_BOTH
	)

	inputLen := len(input)
	if padLength <= inputLen {
		return input
	}

	padLen := padLength - inputLen
	pad := padString
	if len(pad) == 0 {
		pad = " "
	}

	switch padType {
	case STR_PAD_LEFT:
		return strings.Repeat(pad, padLen/len(pad)+1)[:padLen] + input
	case STR_PAD_BOTH:
		leftPad := padLen / 2
		rightPad := padLen - leftPad
		return strings.Repeat(pad, leftPad/len(pad)+1)[:leftPad] +
			input +
			strings.Repeat(pad, rightPad/len(pad)+1)[:rightPad]
	default: // STR_PAD_RIGHT
		return input + strings.Repeat(pad, padLen/len(pad)+1)[:padLen]
	}
}

// StrRepeat repeats a string
func StrRepeat(input string, multiplier int) string {
	return strings.Repeat(input, multiplier)
}

// Wordwrap wraps a string to a given number of characters
func Wordwrap(str string, width int, breakStr string, cut bool) string {
	if width <= 0 {
		return str
	}

	words := strings.Fields(str)
	if len(words) == 0 {
		return str
	}

	var result strings.Builder
	lineLength := 0

	for i, word := range words {
		if i > 0 {
			if lineLength+1+len(word) > width {
				result.WriteString(breakStr)
				lineLength = 0
			} else {
				result.WriteString(" ")
				lineLength++
			}
		}

		if cut && len(word) > width {
			for len(word) > width {
				result.WriteString(word[:width])
				result.WriteString(breakStr)
				word = word[width:]
				lineLength = 0
			}
		}

		result.WriteString(word)
		lineLength += len(word)
	}

	return result.String()
}

// HasPrefix 是否某字符串开头
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix 是否某字符串结尾
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Contains 是否包含某字符串
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// TrimPrefixSlash 去掉开头的 /
func TrimPrefixSlash(s string) string {
	return strings.TrimPrefix(s, "/")
}

// TrimSuffixSlash 去掉结尾的 /
func TrimSuffixSlash(s string) string {
	return strings.TrimSuffix(s, "/")
}

// TrimSlash 去掉前后的 /
func TrimSlash(s string) string {
	return strings.Trim(s, "/")
}

// Addslashes adds backslashes before certain characters
func Addslashes(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\000", "\\000")
	return s
}

// Stripslashes removes backslashes from a string
func Stripslashes(s string) string {
	s = strings.ReplaceAll(s, "\\\\", "\\")
	s = strings.ReplaceAll(s, "\\'", "'")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\000", "\000")
	return s
}

// Sprintf returns a formatted string
func Sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// Concats 连接字符串
func Concats(strs ...string) string {
	var result strings.Builder
	for _, str := range strs {
		_, _ = result.WriteString(str)
	}
	return result.String()
}

// ContainsCommaStr 检查字符串是否包含子字符串(逗号分隔)
func ContainsCommaStr(s, commaStr string) bool {
	return strings.Contains(","+s+",", ","+commaStr+",")
}
