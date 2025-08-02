package util

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Cookie management functions

// SetCookie sends a cookie to the client
func SetCookie(name, value string, options ...map[string]any) bool {
	// This function sets cookie data that would be sent in HTTP headers
	// In a real web application, you would need access to http.ResponseWriter

	// Default options
	expire := 0
	path := "/"
	domain := ""
	secure := false
	httponly := false
	samesite := "Lax"

	// Parse options
	if len(options) > 0 {
		opts := options[0]
		if v, exists := opts["expire"]; exists {
			if exp, ok := v.(int); ok {
				expire = exp
			}
		}
		if v, exists := opts["path"]; exists {
			if p, ok := v.(string); ok {
				path = p
			}
		}
		if v, exists := opts["domain"]; exists {
			if d, ok := v.(string); ok {
				domain = d
			}
		}
		if v, exists := opts["secure"]; exists {
			if s, ok := v.(bool); ok {
				secure = s
			}
		}
		if v, exists := opts["httponly"]; exists {
			if h, ok := v.(bool); ok {
				httponly = h
			}
		}
		if v, exists := opts["samesite"]; exists {
			if ss, ok := v.(string); ok {
				samesite = ss
			}
		}
	}

	// Store cookie information for later retrieval
	cookieData := map[string]any{
		"value":    value,
		"expire":   expire,
		"path":     path,
		"domain":   domain,
		"secure":   secure,
		"httponly": httponly,
		"samesite": samesite,
	}

	storeCookieData(name, cookieData)
	return true
}

// SetRawCookie sends a cookie without URL encoding the value
func SetRawCookie(name, value string, options ...map[string]any) bool {
	// Similar to SetCookie but without URL encoding
	return SetCookie(name, value, options...)
}

// GetCookie gets a cookie value (not in PHP but useful)
func GetCookie(name string, defaultValue ...string) string {
	data := getCookieData(name)
	if data != nil {
		if value, ok := data["value"].(string); ok {
			return value
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return ""
}

// DeleteCookie deletes a cookie (not in PHP but useful)
func DeleteCookie(name string, options ...map[string]any) bool {
	// Set expiration to past date to delete cookie
	opts := make(map[string]any)
	if len(options) > 0 {
		for k, v := range options[0] {
			opts[k] = v
		}
	}
	opts["expire"] = time.Now().Unix() - 3600 // 1 hour ago

	return SetCookie(name, "", opts)
}

// Cookie storage for simulation (in real app, this would be HTTP headers)
var cookieStorage = make(map[string]map[string]any)

func storeCookieData(name string, data map[string]any) {
	cookieStorage[name] = data
}

func getCookieData(name string) map[string]any {
	return cookieStorage[name]
}

// HTTP integration functions

// SetCookieHTTP sets a cookie using http.ResponseWriter
func SetCookieHTTP(w http.ResponseWriter, name, value string, options ...map[string]any) bool {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}

	// Apply options
	if len(options) > 0 {
		opts := options[0]

		if v, exists := opts["expire"]; exists {
			if exp, ok := v.(int); ok {
				if exp > 0 {
					cookie.Expires = time.Unix(int64(exp), 0)
					cookie.MaxAge = exp - int(time.Now().Unix())
				}
			}
		}

		if v, exists := opts["path"]; exists {
			if p, ok := v.(string); ok {
				cookie.Path = p
			}
		}

		if v, exists := opts["domain"]; exists {
			if d, ok := v.(string); ok {
				cookie.Domain = d
			}
		}

		if v, exists := opts["secure"]; exists {
			if s, ok := v.(bool); ok {
				cookie.Secure = s
			}
		}

		if v, exists := opts["httponly"]; exists {
			if h, ok := v.(bool); ok {
				cookie.HttpOnly = h
			}
		}

		if v, exists := opts["samesite"]; exists {
			if ss, ok := v.(string); ok {
				switch strings.ToLower(ss) {
				case "strict":
					cookie.SameSite = http.SameSiteStrictMode
				case "lax":
					cookie.SameSite = http.SameSiteLaxMode
				case "none":
					cookie.SameSite = http.SameSiteNoneMode
				default:
					cookie.SameSite = http.SameSiteDefaultMode
				}
			}
		}
	}

	http.SetCookie(w, cookie)
	return true
}

// GetCookieHTTP gets a cookie from http.Request
func GetCookieHTTP(r *http.Request, name string, defaultValue ...string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return cookie.Value
}

// GetAllCookiesHTTP gets all cookies from http.Request
func GetAllCookiesHTTP(r *http.Request) map[string]string {
	result := make(map[string]string)
	for _, cookie := range r.Cookies() {
		result[cookie.Name] = cookie.Value
	}
	return result
}

// DeleteCookieHTTP deletes a cookie using http.ResponseWriter
func DeleteCookieHTTP(w http.ResponseWriter, name string, options ...map[string]any) bool {
	opts := make(map[string]any)
	if len(options) > 0 {
		for k, v := range options[0] {
			opts[k] = v
		}
	}
	opts["expire"] = int(time.Now().Unix() - 3600) // 1 hour ago
	opts["maxage"] = -1

	return SetCookieHTTP(w, name, "", opts)
}

// Cookie validation and security functions

// ValidateCookieName validates cookie name according to RFC specifications
func ValidateCookieName(name string) bool {
	if name == "" {
		return false
	}

	// Cookie name cannot contain these characters
	invalidChars := "()[]{}/?,@:;\\\"="
	for _, char := range invalidChars {
		if strings.ContainsRune(name, char) {
			return false
		}
	}

	// Cannot contain spaces or control characters
	for _, r := range name {
		if r <= 0x20 || r >= 0x7F {
			return false
		}
	}

	return true
}

// ValidateCookieValue validates cookie value
func ValidateCookieValue(value string) bool {
	// Check for invalid characters in cookie value
	for _, r := range value {
		if r < 0x21 || r > 0x7E || r == '"' || r == ',' || r == ';' || r == '\\' {
			return false
		}
	}
	return true
}

// SecureCookieValue encrypts cookie value for security
func SecureCookieValue(value, secret string) string {
	// Simple encryption using XOR (in production, use proper encryption)
	result := make([]byte, len(value))
	secretBytes := []byte(secret)

	for i, v := range []byte(value) {
		result[i] = v ^ secretBytes[i%len(secretBytes)]
	}

	return Base64Encode(string(result))
}

// UnsecureCookieValue decrypts cookie value
func UnsecureCookieValue(encryptedValue, secret string) (string, error) {
	decoded, err := Base64Decode(encryptedValue)
	if err != nil {
		return "", err
	}

	result := make([]byte, len(decoded))
	secretBytes := []byte(secret)

	for i, v := range []byte(decoded) {
		result[i] = v ^ secretBytes[i%len(secretBytes)]
	}

	return string(result), nil
}

// Cookie utility functions

// CookieSize calculates the size of a cookie in bytes
func CookieSize(name, value string) int {
	// Basic calculation: name=value
	size := len(name) + 1 + len(value) // +1 for '='
	return size
}

// ValidateCookieSize checks if cookie size is within limits
func ValidateCookieSize(name, value string) bool {
	const maxCookieSize = 4096 // 4KB limit for most browsers
	return CookieSize(name, value) <= maxCookieSize
}

// ParseCookieHeader parses raw cookie header string
func ParseCookieHeader(header string) map[string]string {
	cookies := make(map[string]string)

	for _, cookie := range strings.Split(header, ";") {
		cookie = strings.TrimSpace(cookie)
		if cookie == "" {
			continue
		}

		parts := strings.SplitN(cookie, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// URL decode if needed
		if decodedValue, err := url.QueryUnescape(value); err == nil {
			value = decodedValue
		}

		cookies[name] = value
	}

	return cookies
}

// BuildCookieHeader builds a cookie header string
func BuildCookieHeader(cookies map[string]string) string {
	var parts []string

	for name, value := range cookies {
		// URL encode if needed
		encodedValue := url.QueryEscape(value)
		parts = append(parts, name+"="+encodedValue)
	}

	return strings.Join(parts, "; ")
}

// Cookie domain and path utilities

// IsValidCookieDomain checks if domain is valid for cookie
func IsValidCookieDomain(domain, hostDomain string) bool {
	if domain == "" {
		return true // Empty domain means current domain
	}

	// Remove leading dot
	if strings.HasPrefix(domain, ".") {
		domain = domain[1:]
	}

	// Check if host domain ends with cookie domain
	return strings.HasSuffix(hostDomain, domain)
}

// IsValidCookiePath checks if path is valid for cookie
func IsValidCookiePath(cookiePath, requestPath string) bool {
	if cookiePath == "" || cookiePath == "/" {
		return true
	}

	// Cookie path must be a prefix of request path
	return strings.HasPrefix(requestPath, cookiePath)
}

// Advanced cookie functions

// CookieJar manages multiple cookies
type CookieJar struct {
	cookies map[string]*http.Cookie
}

// NewCookieJar creates a new cookie jar
func NewCookieJar() *CookieJar {
	return &CookieJar{
		cookies: make(map[string]*http.Cookie),
	}
}

// SetCookieJar adds a cookie to the jar
func (cj *CookieJar) SetCookieJar(cookie *http.Cookie) {
	cj.cookies[cookie.Name] = cookie
}

// GetCookieJar gets a cookie from the jar
func (cj *CookieJar) GetCookieJar(name string) *http.Cookie {
	return cj.cookies[name]
}

// GetAllCookiesJar gets all cookies from the jar
func (cj *CookieJar) GetAllCookiesJar() map[string]*http.Cookie {
	result := make(map[string]*http.Cookie)
	for k, v := range cj.cookies {
		result[k] = v
	}
	return result
}

// DeleteCookieJar removes a cookie from the jar
func (cj *CookieJar) DeleteCookieJar(name string) {
	delete(cj.cookies, name)
}

// ClearCookieJar removes all cookies from the jar
func (cj *CookieJar) ClearCookieJar() {
	cj.cookies = make(map[string]*http.Cookie)
}

// FilterCookiesJar filters cookies by domain and path
func (cj *CookieJar) FilterCookiesJar(domain, path string) []*http.Cookie {
	var result []*http.Cookie

	for _, cookie := range cj.cookies {
		if IsValidCookieDomain(cookie.Domain, domain) &&
			IsValidCookiePath(cookie.Path, path) {
			result = append(result, cookie)
		}
	}

	return result
}

// Cookie security helpers

// IsSecureCookie checks if cookie should only be sent over HTTPS
func IsSecureCookie(cookie *http.Cookie) bool {
	return cookie.Secure
}

// IsHttpOnlyCookie checks if cookie is HTTP only
func IsHttpOnlyCookie(cookie *http.Cookie) bool {
	return cookie.HttpOnly
}

// GetCookieSameSite gets the SameSite attribute
func GetCookieSameSite(cookie *http.Cookie) string {
	switch cookie.SameSite {
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteLaxMode:
		return "Lax"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return "Default"
	}
}

// Cookie expiration utilities

// IsCookieExpired checks if cookie is expired
func IsCookieExpired(cookie *http.Cookie) bool {
	if cookie.MaxAge < 0 {
		return true
	}

	if !cookie.Expires.IsZero() && cookie.Expires.Before(time.Now()) {
		return true
	}

	return false
}

// GetCookieExpirationTime gets the expiration time of a cookie
func GetCookieExpirationTime(cookie *http.Cookie) time.Time {
	if cookie.MaxAge > 0 {
		return time.Now().Add(time.Duration(cookie.MaxAge) * time.Second)
	}

	if !cookie.Expires.IsZero() {
		return cookie.Expires
	}

	// Session cookie - expires when browser closes
	return time.Time{}
}
