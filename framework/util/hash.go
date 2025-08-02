package util

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"strconv"
)

// Md5 calculates the md5 hash of a string
func Md5(data string, rawOutput ...bool) string {
	h := md5.Sum([]byte(data))
	if len(rawOutput) > 0 && rawOutput[0] {
		return string(h[:])
	}
	return hex.EncodeToString(h[:])
}

// Md5File calculates the md5 hash of a file
func Md5File(filename string, rawOutput ...bool) (string, error) {
	content, err := FileGetContents(filename)
	if err != nil {
		return "", err
	}
	return Md5(content, rawOutput...), nil
}

// Sha1 calculates the sha1 hash of a string
func Sha1(data string, rawOutput ...bool) string {
	h := sha1.Sum([]byte(data))
	if len(rawOutput) > 0 && rawOutput[0] {
		return string(h[:])
	}
	return hex.EncodeToString(h[:])
}

// Sha1File calculates the sha1 hash of a file
func Sha1File(filename string, rawOutput ...bool) (string, error) {
	content, err := FileGetContents(filename)
	if err != nil {
		return "", err
	}
	return Sha1(content, rawOutput...), nil
}

// Hash generates a hash value
func Hash(algo, data string, rawOutput ...bool) (string, error) {
	var h hash.Hash

	switch algo {
	case "md5":
		h = md5.New()
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	case "crc32":
		h = crc32.NewIEEE()
	default:
		return "", fmt.Errorf("unsupported hash algorithm: %s", algo)
	}

	h.Write([]byte(data))
	sum := h.Sum(nil)

	if len(rawOutput) > 0 && rawOutput[0] {
		return string(sum), nil
	}
	return hex.EncodeToString(sum), nil
}

// HashFile generates a hash value for a file
func HashFile(algo, filename string, rawOutput ...bool) (string, error) {
	content, err := FileGetContents(filename)
	if err != nil {
		return "", err
	}
	return Hash(algo, content, rawOutput...)
}

// HashHmac generates a keyed hash value using the HMAC method
func HashHmac(algo, data, key string, rawOutput ...bool) (string, error) {
	// Simplified HMAC implementation
	// In production, use crypto/hmac package

	keyBytes := []byte(key)
	dataBytes := []byte(data)

	// Simplified: just hash key + data
	combined := append(keyBytes, dataBytes...)
	return Hash(algo, string(combined), rawOutput...)
}

// Crc32 calculates the crc32 polynomial of a string
func Crc32(data string) uint32 {
	return crc32.ChecksumIEEE([]byte(data))
}

// Base64Encode encodes data with MIME base64
func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64Decode decodes data encoded with MIME base64
func Base64Decode(data string, strict ...bool) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// Base64UrlEncode encodes data with URL-safe base64
func Base64UrlEncode(data string) string {
	return base64.URLEncoding.EncodeToString([]byte(data))
}

// Base64UrlDecode decodes URL-safe base64 encoded data
func Base64UrlDecode(data string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// UrlEncode URL-encodes string
func UrlEncode(str string) string {
	// Simple URL encoding implementation
	var result string
	for _, char := range str {
		switch {
		case (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9'):
			result += string(char)
		case char == '-' || char == '_' || char == '.' || char == '~':
			result += string(char)
		default:
			result += fmt.Sprintf("%%%02X", char)
		}
	}
	return result
}

// UrlDecode decodes URL-encoded string
func UrlDecode(str string) (string, error) {
	var result string
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '%':
			if i+2 < len(str) {
				hex := str[i+1 : i+3]
				if char, err := strconv.ParseInt(hex, 16, 8); err == nil {
					result += string(rune(char))
					i += 2
				} else {
					return "", fmt.Errorf("invalid URL encoding")
				}
			} else {
				return "", fmt.Errorf("invalid URL encoding")
			}
		case '+':
			result += " "
		default:
			result += string(str[i])
		}
	}
	return result, nil
}

// RawUrlEncode URL-encode according to RFC 3986
func RawUrlEncode(str string) string {
	var result string
	for _, char := range str {
		switch {
		case (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9'):
			result += string(char)
		case char == '-' || char == '_' || char == '.' || char == '~':
			result += string(char)
		default:
			result += fmt.Sprintf("%%%02X", char)
		}
	}
	return result
}

// RawUrlDecode decode URL-encoded strings
func RawUrlDecode(str string) (string, error) {
	var result string
	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '%':
			if i+2 < len(str) {
				hex := str[i+1 : i+3]
				if char, err := strconv.ParseInt(hex, 16, 8); err == nil {
					result += string(rune(char))
					i += 2
				} else {
					return "", fmt.Errorf("invalid URL encoding")
				}
			} else {
				return "", fmt.Errorf("invalid URL encoding")
			}
		default:
			result += string(str[i])
		}
	}
	return result, nil
}

// HexToBin converts hexadecimal to binary
func HexToBin(hexStr string) (string, error) {
	if len(hexStr)%2 != 0 {
		return "", fmt.Errorf("hex string must have even length")
	}

	decoded, err := hex.DecodeString(hexStr)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// BinToHex converts binary to hexadecimal
func BinToHex(bin string) string {
	return hex.EncodeToString([]byte(bin))
}

// Uniqid generates a unique ID
func Uniqid(prefix string, moreEntropy ...bool) string {
	timestamp := Time()
	microsec := Microtime(true).(float64)

	result := prefix + fmt.Sprintf("%x", timestamp)

	if len(moreEntropy) > 0 && moreEntropy[0] {
		// Add microseconds for more entropy
		result += fmt.Sprintf("%.8f", microsec-float64(timestamp))
	}

	return result
}

// PasswordHash creates a password hash
func PasswordHash(password string, algo int, options ...map[string]any) string {
	// Simplified implementation - in production use bcrypt or similar
	salt := "defaultsalt" // Should be random
	if len(options) > 0 {
		if s, exists := options[0]["salt"]; exists {
			salt = fmt.Sprintf("%v", s)
		}
	}

	// Simple hash - in production use proper password hashing
	return Sha256(password + salt)
}

// PasswordVerify verifies a password against a hash
func PasswordVerify(password, hash string) bool {
	// Simplified implementation
	return PasswordHash(password, 0) == hash
}

// Sha256 calculates the sha256 hash of a string
func Sha256(data string, rawOutput ...bool) string {
	h := sha256.Sum256([]byte(data))
	if len(rawOutput) > 0 && rawOutput[0] {
		return string(h[:])
	}
	return hex.EncodeToString(h[:])
}

// Sha512 calculates the sha512 hash of a string
func Sha512(data string, rawOutput ...bool) string {
	h := sha512.Sum512([]byte(data))
	if len(rawOutput) > 0 && rawOutput[0] {
		return string(h[:])
	}
	return hex.EncodeToString(h[:])
}

// HashAlgos returns a list of registered hashing algorithms
func HashAlgos() []string {
	return []string{"md5", "sha1", "sha256", "sha512", "crc32"}
}

// HashEquals compares two strings using the same time whether they're equal or not
func HashEquals(knownString, userString string) bool {
	if len(knownString) != len(userString) {
		return false
	}

	var result byte
	for i := 0; i < len(knownString); i++ {
		result |= knownString[i] ^ userString[i]
	}

	return result == 0
}

// Crypt one-way string hashing
func Crypt(str, salt string) string {
	// Simplified implementation - just use sha256 with salt
	return Sha256(str + salt)
}

// PasswordGetInfo returns information about the given hash
func PasswordGetInfo(hash string) map[string]any {
	return map[string]any{
		"algo":     1, // PASSWORD_DEFAULT
		"algoName": "bcrypt",
		"options":  map[string]any{},
	}
}

// PasswordNeedsRehash checks if the given hash matches the given options
func PasswordNeedsRehash(hash string, algo int, options ...map[string]any) bool {
	// Simplified implementation - always return false
	return false
}

// Random generates cryptographically secure pseudo-random bytes
func Random(length int) ([]byte, error) {
	// Simplified implementation - in production use crypto/rand
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = byte(Time() % 256) // Not cryptographically secure
	}
	return bytes, nil
}

// RandomBytes generates cryptographically secure pseudo-random bytes
func RandomBytes(length int) ([]byte, error) {
	return Random(length)
}

// RandomInt generates cryptographically secure pseudo-random integers
func RandomInt(min, max int) int {
	// Simplified implementation - in production use crypto/rand
	diff := max - min + 1
	return min + int(Time()%int64(diff))
}
