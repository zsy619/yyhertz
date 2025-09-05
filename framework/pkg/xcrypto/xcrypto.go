// Package xcrypto 提供了加密、哈希、ID生成等安全相关的工具函数
//
// 主要功能：
//   - 哈希函数：MD5、SHA1、SHA256、SHA512、CRC32、Bcrypt等
//   - 加密解密：AES、DES、Base64、URL编码等
//   - ID生成：UUID、雪花ID、随机ID等
//   - 数字摘要：HMAC签名验证等
//
// 基础用法：
//
//	import "your-project/framework/pkg/xcrypto"
//
//	// 哈希运算
//	md5Hash := xcrypto.Md5("hello world")
//	sha1Hash := xcrypto.Sha1("hello world")
//
//	// AES加密解密
//	key := "your-32-byte-key-for-aes-256"
//	encrypted := xcrypto.AesEncrypt("plain text", key)
//	decrypted := xcrypto.AesDecrypt(encrypted, key)
//
//	// ID生成
//	uuid := xcrypto.UUID()
//	snowflakeId := xcrypto.SnowflakeId()
//
//	// Base64编码
//	encoded := xcrypto.Base64Encode("hello world")
//	decoded := xcrypto.Base64Decode(encoded)
//
// 密码加密示例：
//
//	// Bcrypt密码加密
//	hashedPassword := xcrypto.BcryptHash("user_password")
//	isValid := xcrypto.BcryptCheck("user_password", hashedPassword)
//
// HMAC签名示例：
//
//	// HMAC-SHA256签名
//	signature := xcrypto.HmacSha256("data", "secret_key")
//	isValid := xcrypto.HmacSha256Verify("data", "secret_key", signature)
//
package xcrypto