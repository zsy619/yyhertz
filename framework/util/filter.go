package util

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Filter constants
const (
	// Validate filters
	FILTER_VALIDATE_BOOLEAN = 258
	FILTER_VALIDATE_EMAIL   = 274
	FILTER_VALIDATE_FLOAT   = 259
	FILTER_VALIDATE_INT     = 257
	FILTER_VALIDATE_IP      = 275
	FILTER_VALIDATE_REGEXP  = 272
	FILTER_VALIDATE_URL     = 273
	FILTER_VALIDATE_DOMAIN  = 277
	FILTER_VALIDATE_MAC     = 276

	// Sanitize filters
	FILTER_SANITIZE_EMAIL         = 517
	FILTER_SANITIZE_ENCODED       = 514
	FILTER_SANITIZE_MAGIC_QUOTES  = 521
	FILTER_SANITIZE_NUMBER_FLOAT  = 520
	FILTER_SANITIZE_NUMBER_INT    = 519
	FILTER_SANITIZE_SPECIAL_CHARS = 515
	FILTER_SANITIZE_STRING        = 513
	FILTER_SANITIZE_STRIPPED      = 513
	FILTER_SANITIZE_URL           = 518
	FILTER_SANITIZE_ADD_SLASHES   = 523

	// Flags
	FILTER_FLAG_STRIP_LOW         = 4
	FILTER_FLAG_STRIP_HIGH        = 8
	FILTER_FLAG_STRIP_BACKTICK    = 512
	FILTER_FLAG_ALLOW_FRACTION    = 4096
	FILTER_FLAG_ALLOW_THOUSAND    = 8192
	FILTER_FLAG_ALLOW_SCIENTIFIC  = 16384
	FILTER_FLAG_NO_ENCODE_QUOTES  = 1
	FILTER_FLAG_ENCODE_LOW        = 32
	FILTER_FLAG_ENCODE_HIGH       = 64
	FILTER_FLAG_ENCODE_AMP        = 2
	FILTER_FLAG_IPV4              = 1048576
	FILTER_FLAG_IPV6              = 2097152
	FILTER_FLAG_NO_PRIV_RANGE     = 8388608
	FILTER_FLAG_NO_RES_RANGE      = 4194304
	FILTER_FLAG_SCHEME_REQUIRED   = 65536
	FILTER_FLAG_HOST_REQUIRED     = 131072
	FILTER_FLAG_PATH_REQUIRED     = 262144
	FILTER_FLAG_QUERY_REQUIRED    = 524288
	FILTER_FLAG_EMPTY_STRING_NULL = 256

	// Callback filter
	FILTER_CALLBACK = 1024

	// Default filter
	FILTER_DEFAULT    = 516
	FILTER_UNSAFE_RAW = 512
)

// FilterInput gets external variables and optionally filters them
func FilterInput(inputType int, variableName string, filter int, flags ...any) any {
	const (
		INPUT_GET    = 0
		INPUT_POST   = 1
		INPUT_COOKIE = 2
		INPUT_SERVER = 3
		INPUT_ENV    = 4
	)

	// In a real implementation, this would get data from the appropriate source
	// For now, return a placeholder
	value := "example_value"

	return FilterVar(value, filter, flags...)
}

// FilterInputArray gets external variables and optionally filters them
func FilterInputArray(inputType int, definition map[string]map[string]any) map[string]any {
	result := make(map[string]any)

	// In a real implementation, this would process the input array
	for key, filterDef := range definition {
		filter := filterDef["filter"].(int)
		flags := filterDef["flags"]

		// Get the value (placeholder)
		value := "example_value"
		result[key] = FilterVar(value, filter, flags)
	}

	return result
}

// FilterVar filters a variable with a specified filter (enhanced version)
func FilterVar(value any, filter int, flags ...any) any {
	str := ""
	if value != nil {
		str = Strval(value)
	}

	var flag int
	var options map[string]any

	// Process flags and options
	for _, f := range flags {
		switch v := f.(type) {
		case int:
			flag = v
		case map[string]any:
			options = v
		}
	}

	// Handle different filters
	switch filter {
	case FILTER_VALIDATE_BOOLEAN:
		return validateBoolean(str, flag)
	case FILTER_VALIDATE_EMAIL:
		return validateEmail(str, flag)
	case FILTER_VALIDATE_FLOAT:
		return validateFloat(str, flag)
	case FILTER_VALIDATE_INT:
		return validateInt(str, flag, options)
	case FILTER_VALIDATE_IP:
		return validateIP(str, flag)
	case FILTER_VALIDATE_REGEXP:
		return validateRegexp(str, options)
	case FILTER_VALIDATE_URL:
		return validateURL(str, flag)
	case FILTER_VALIDATE_DOMAIN:
		return validateDomain(str, flag)
	case FILTER_VALIDATE_MAC:
		return validateMAC(str)

	case FILTER_SANITIZE_EMAIL:
		return sanitizeEmail(str)
	case FILTER_SANITIZE_ENCODED:
		return sanitizeEncoded(str)
	case FILTER_SANITIZE_NUMBER_FLOAT:
		return sanitizeNumberFloat(str, flag)
	case FILTER_SANITIZE_NUMBER_INT:
		return sanitizeNumberInt(str)
	case FILTER_SANITIZE_SPECIAL_CHARS:
		return sanitizeSpecialChars(str, flag)
	case FILTER_SANITIZE_STRING:
		return sanitizeString(str, flag)
	case FILTER_SANITIZE_URL:
		return sanitizeURL(str)
	case FILTER_SANITIZE_ADD_SLASHES:
		return Addslashes(str)

	case FILTER_CALLBACK:
		if callback, exists := options["callback"]; exists {
			if fn, ok := callback.(func(string) any); ok {
				return fn(str)
			}
		}
		return false

	case FILTER_DEFAULT, FILTER_UNSAFE_RAW:
		return str

	default:
		return str
	}
}

// FilterVarArray filters multiple variables
func FilterVarArray(data map[string]any, definition map[string]map[string]any) map[string]any {
	result := make(map[string]any)

	for key, value := range data {
		if filterDef, exists := definition[key]; exists {
			filter := filterDef["filter"].(int)
			flags := filterDef["flags"]
			result[key] = FilterVar(value, filter, flags)
		} else {
			result[key] = value
		}
	}

	return result
}

// FilterHasSanitizeFilter checks if filter is a sanitize filter
func FilterHasSanitizeFilter(filterId int) bool {
	return filterId >= 513 && filterId <= 523
}

// FilterId returns the filter ID belonging to a named filter
func FilterId(filterName string) int {
	filterMap := map[string]int{
		"boolean": FILTER_VALIDATE_BOOLEAN,
		"email":   FILTER_VALIDATE_EMAIL,
		"float":   FILTER_VALIDATE_FLOAT,
		"int":     FILTER_VALIDATE_INT,
		"ip":      FILTER_VALIDATE_IP,
		"regexp":  FILTER_VALIDATE_REGEXP,
		"url":     FILTER_VALIDATE_URL,
		"domain":  FILTER_VALIDATE_DOMAIN,
		"mac":     FILTER_VALIDATE_MAC,
	}

	if id, exists := filterMap[filterName]; exists {
		return id
	}
	return 0
}

// FilterList returns a list of all supported filters
func FilterList() []string {
	return []string{
		"boolean", "email", "float", "int", "ip", "regexp", "url", "domain", "mac",
		"sanitize_email", "sanitize_encoded", "sanitize_number_float", "sanitize_number_int",
		"sanitize_special_chars", "sanitize_string", "sanitize_url", "callback", "unsafe_raw",
	}
}

// Validation functions

func validateBoolean(value string, flag int) any {
	lower := strings.ToLower(value)
	switch lower {
	case "1", "true", "on", "yes":
		return true
	case "0", "false", "off", "no", "":
		return false
	default:
		if flag&FILTER_FLAG_EMPTY_STRING_NULL != 0 && value == "" {
			return nil
		}
		return false
	}
}

func validateEmail(value string, flag int) any {
	// Basic email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(value) {
		return value
	}
	return false
}

func validateFloat(value string, flag int) any {
	// Remove thousand separators if allowed
	if flag&FILTER_FLAG_ALLOW_THOUSAND != 0 {
		value = strings.ReplaceAll(value, ",", "")
	}

	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f
	}
	return false
}

func validateInt(value string, flag int, options map[string]any) any {
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		// Check min/max range if provided
		if options != nil {
			if minVal, exists := options["min_range"]; exists {
				if min, ok := minVal.(int64); ok && i < min {
					return false
				}
			}
			if maxVal, exists := options["max_range"]; exists {
				if max, ok := maxVal.(int64); ok && i > max {
					return false
				}
			}
		}
		return i
	}
	return false
}

func validateIP(value string, flag int) any {
	ip := net.ParseIP(value)
	if ip == nil {
		return false
	}

	// Check IPv4/IPv6 flags
	if flag&FILTER_FLAG_IPV4 != 0 && ip.To4() == nil {
		return false
	}
	if flag&FILTER_FLAG_IPV6 != 0 && (ip.To4() != nil || ip.To16() == nil) {
		return false
	}

	// Check private/reserved ranges
	if flag&FILTER_FLAG_NO_PRIV_RANGE != 0 && isPrivateIP(ip) {
		return false
	}
	if flag&FILTER_FLAG_NO_RES_RANGE != 0 && isReservedIP(ip) {
		return false
	}

	return value
}

func validateRegexp(value string, options map[string]any) any {
	if options == nil {
		return false
	}

	regexpVal, exists := options["regexp"]
	if !exists {
		return false
	}

	pattern, ok := regexpVal.(string)
	if !ok {
		return false
	}

	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return false
	}

	if matched {
		return value
	}
	return false
}

func validateURL(value string, flag int) any {
	// Basic URL validation
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(value) {
		return false
	}

	// Check additional requirements
	if flag&FILTER_FLAG_PATH_REQUIRED != 0 && !strings.Contains(value, "/") {
		return false
	}
	if flag&FILTER_FLAG_QUERY_REQUIRED != 0 && !strings.Contains(value, "?") {
		return false
	}

	return value
}

func validateDomain(value string, flag int) any {
	// Basic domain validation
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	if domainRegex.MatchString(value) && len(value) <= 253 {
		return value
	}
	return false
}

func validateMAC(value string) any {
	// MAC address validation
	macRegex := regexp.MustCompile(`^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$`)
	if macRegex.MatchString(value) {
		return value
	}
	return false
}

// Sanitization functions

func sanitizeEmail(value string) string {
	// Remove characters not allowed in email
	var result strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '@' || r == '.' || r == '_' || r == '-' || r == '+' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func sanitizeEncoded(value string) string {
	// URL encode special characters
	return UrlEncode(value)
}

func sanitizeNumberFloat(value string, flag int) string {
	var result strings.Builder

	for _, r := range value {
		if unicode.IsDigit(r) || r == '.' || r == '-' || r == '+' {
			result.WriteRune(r)
		} else if flag&FILTER_FLAG_ALLOW_FRACTION != 0 && r == '/' {
			result.WriteRune(r)
		} else if flag&FILTER_FLAG_ALLOW_THOUSAND != 0 && r == ',' {
			result.WriteRune(r)
		} else if flag&FILTER_FLAG_ALLOW_SCIENTIFIC != 0 && (r == 'e' || r == 'E') {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func sanitizeNumberInt(value string) string {
	var result strings.Builder

	for _, r := range value {
		if unicode.IsDigit(r) || r == '-' || r == '+' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func sanitizeSpecialChars(value string, flag int) string {
	result := value

	// HTML encode special characters
	result = strings.ReplaceAll(result, "&", "&amp;")
	result = strings.ReplaceAll(result, "<", "&lt;")
	result = strings.ReplaceAll(result, ">", "&gt;")

	if flag&FILTER_FLAG_NO_ENCODE_QUOTES == 0 {
		result = strings.ReplaceAll(result, "\"", "&quot;")
		result = strings.ReplaceAll(result, "'", "&#39;")
	}

	if flag&FILTER_FLAG_STRIP_LOW != 0 {
		// Remove low ASCII characters (0-31)
		var cleaned strings.Builder
		for _, r := range result {
			if r >= 32 {
				cleaned.WriteRune(r)
			}
		}
		result = cleaned.String()
	}

	if flag&FILTER_FLAG_STRIP_HIGH != 0 {
		// Remove high ASCII characters (127-255)
		var cleaned strings.Builder
		for _, r := range result {
			if r < 127 {
				cleaned.WriteRune(r)
			}
		}
		result = cleaned.String()
	}

	return result
}

func sanitizeString(value string, flag int) string {
	result := value

	// Strip tags
	result = StripTags(result)

	if flag&FILTER_FLAG_STRIP_LOW != 0 {
		var cleaned strings.Builder
		for _, r := range result {
			if r >= 32 {
				cleaned.WriteRune(r)
			}
		}
		result = cleaned.String()
	}

	if flag&FILTER_FLAG_STRIP_HIGH != 0 {
		var cleaned strings.Builder
		for _, r := range result {
			if r < 127 {
				cleaned.WriteRune(r)
			}
		}
		result = cleaned.String()
	}

	if flag&FILTER_FLAG_STRIP_BACKTICK != 0 {
		result = strings.ReplaceAll(result, "`", "")
	}

	return result
}

func sanitizeURL(value string) string {
	// Remove characters not allowed in URLs
	var result strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) ||
			r == ':' || r == '/' || r == '?' || r == '#' || r == '[' || r == ']' ||
			r == '@' || r == '!' || r == '$' || r == '&' || r == '\'' || r == '(' ||
			r == ')' || r == '*' || r == '+' || r == ',' || r == ';' || r == '=' ||
			r == '%' || r == '-' || r == '.' || r == '_' || r == '~' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Helper functions for IP validation

func isPrivateIP(ip net.IP) bool {
	if ip.To4() != nil {
		// IPv4 private ranges
		return ip.IsPrivate()
	}
	// IPv6 private ranges would be checked here
	return false
}

func isReservedIP(ip net.IP) bool {
	if ip.To4() != nil {
		// Check for reserved IPv4 ranges
		if ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() {
			return true
		}
	}
	return false
}

// Additional filter functions

// FilterHasVar checks if variable of specified type exists
func FilterHasVar(inputType int, variableName string) bool {
	// In a real implementation, this would check if the variable exists
	// in the specified input type (GET, POST, etc.)
	return true // Placeholder
}

// Custom filter function type
type FilterFunc func(string) any

// RegisterFilter registers a custom filter (not in PHP but useful for Go)
func RegisterFilter(name string, filter FilterFunc) {
	// In a real implementation, this would register the filter
	// for use with FILTER_CALLBACK
}
