package session

import "time"

// Config Session配置
type Config struct {
	Enabled       bool          // 是否启用Session
	CookieName    string        // Session Cookie名称
	CookiePath    string        // Cookie路径
	CookieDomain  string        // Cookie域名
	MaxAge        int           // 过期时间（秒）
	Secure        bool          // 是否仅HTTPS
	HttpOnly      bool          // 是否仅HTTP
	SameSite      string        // SameSite策略
	CleanInterval time.Duration // 清理间隔
}

// DefaultConfig 默认Session配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:       true,
		CookieName:    "session_id",
		CookiePath:    "/",
		CookieDomain:  "",
		MaxAge:        3600, // 1小时
		Secure:        false,
		HttpOnly:      true,
		SameSite:      "Lax",
		CleanInterval: 10 * time.Minute,
	}
}