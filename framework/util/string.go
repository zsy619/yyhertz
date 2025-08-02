package util

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"
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

// Strtolower converts string to lowercase
func Strtolower(s string) string {
	return strings.ToLower(s)
}

// Strtoupper converts string to uppercase
func Strtoupper(s string) string {
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
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
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

// Number format a number with grouped thousands
func NumberFormat(number float64, decimals int, decPoint, thousandsSep string) string {
	formatted := fmt.Sprintf("%."+strconv.Itoa(decimals)+"f", number)
	parts := strings.Split(formatted, ".")

	// Add thousands separator
	intPart := parts[0]
	if len(intPart) > 3 {
		var result []rune
		for i, r := range []rune(intPart) {
			if i > 0 && (len(intPart)-i)%3 == 0 {
				result = append(result, []rune(thousandsSep)...)
			}
			result = append(result, r)
		}
		intPart = string(result)
	}

	if decimals > 0 && len(parts) > 1 {
		return intPart + decPoint + parts[1]
	}
	return intPart
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
