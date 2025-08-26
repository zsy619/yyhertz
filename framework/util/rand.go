package util

import (
	"crypto/rand"
	r "math/rand"
	"time"
)

var alphaNum = []byte(`0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`)

// generateRandomBytes 生成一个指定长度的随机字节切片。
//
// 参数：
//   - alphabets: 可选的自定义字符集，用于生成随机字节。如果为空，则使用默认的字母数字字符集。
//   - n: 生成的字节切片的长度。
//
// 返回值：
//   - []byte: 生成的随机字节切片。
//
// 行为：
//   - 如果未提供自定义字符集，则使用默认的字母数字字符集。
//   - 使用系统的随机数生成器生成随机字节。如果系统随机数生成失败，则回退到伪随机数生成器。
//   - 生成的随机字节基于提供的字符集，确保每个字节都在字符集范围内。
//
// 注意：
//   - 此函数是内部实现的一部分，通常不直接暴露给外部调用者。
func RandomCreateBytes(n int, alphabets ...byte) []byte {
	if len(alphabets) == 0 {
		alphabets = alphaNum
	}
	bytes := make([]byte, n)
	var randBy bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randBy = true
	}
	for i, b := range bytes {
		if randBy {
			bytes[i] = alphabets[r.Intn(len(alphabets))]
		} else {
			bytes[i] = alphabets[b%byte(len(alphabets))]
		}
	}
	return bytes
}
