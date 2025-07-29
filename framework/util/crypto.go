package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// ReverseString 字符串反转并转换为小写字母
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return strings.ToLower(string(runes))
}

// MD5Hash 生成MD5哈希
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// SHA256Hash 生成SHA256哈希
func SHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// SimpleEncrypt 简单加密(基于异或)
func SimpleEncrypt(text, key string) string {
	if key == "" {
		key = "hertz-mvc-default-key"
	}
	
	result := make([]byte, len(text))
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	
	for i, b := range []byte(text) {
		result[i] = b ^ keyBytes[i%keyLen]
	}
	
	return hex.EncodeToString(result)
}

// SimpleDecrypt 简单解密(基于异或)
func SimpleDecrypt(encrypted, key string) (string, error) {
	if key == "" {
		key = "hertz-mvc-default-key"
	}
	
	data, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	
	result := make([]byte, len(data))
	keyBytes := []byte(key)
	keyLen := len(keyBytes)
	
	for i, b := range data {
		result[i] = b ^ keyBytes[i%keyLen]
	}
	
	return string(result), nil
}

// PasswordEncrypt 密码加密
func PasswordEncrypt(password, salt string) string {
	key := ReverseString(salt)
	return SimpleEncrypt(password, key)
}

// PasswordDecrypt 密码解密
func PasswordDecrypt(encrypted, salt string) (string, error) {
	key := ReverseString(salt)
	return SimpleDecrypt(encrypted, key)
}