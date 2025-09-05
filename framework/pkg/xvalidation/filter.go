package xvalidation

import (
	"html"
	"net/url"
	"regexp"
	"strings"
)

// StripTags removes HTML tags from string
func StripTags(s string) string {
	// Simple HTML tag removal
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// HtmlSpecialChars converts special characters to HTML entities
func HtmlSpecialChars(s string) string {
	return html.EscapeString(s)
}

// HtmlEntityDecode decodes HTML entities
func HtmlEntityDecode(s string) string {
	return html.UnescapeString(s)
}

// UrlEncode URL-encodes string
func UrlEncode(str string) string {
	return url.QueryEscape(str)
}

// UrlDecode decodes URL-encoded string
func UrlDecode(str string) (string, error) {
	return url.QueryUnescape(str)
}

// FilterInput filters and sanitizes input
func FilterInput(input string) string {
	// Basic input filtering
	filtered := strings.TrimSpace(input)
	filtered = HtmlSpecialChars(filtered)
	return filtered
}