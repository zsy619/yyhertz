package session

import (
	"time"
	
	"github.com/zsy619/yyhertz/framework/config"
)

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

// LoadFromConfig 从配置文件加载Session配置
func LoadFromConfig() *Config {
	// 获取session配置
	sessionConfig, err := config.LoadConfigWithGeneric[config.SessionConfig]("session")
	if err != nil {
		config.Warnf("Failed to load session config, using defaults: %v", err)
		return DefaultConfig()
	}

	// 解析时间间隔
	var cleanInterval time.Duration
	if sessionConfig.Session.Cleanup.Enable {
		if interval, err := time.ParseDuration(sessionConfig.Session.Cleanup.Interval); err == nil {
			cleanInterval = interval
		} else {
			cleanInterval = 10 * time.Minute // 默认值
		}
	} else {
		cleanInterval = 0 // 禁用清理
	}

	sessionCfg := &Config{
		Enabled:       sessionConfig.Middleware.Session.Enable,
		CookieName:    sessionConfig.Session.Name,
		CookiePath:    sessionConfig.Session.Cookie.Path,
		CookieDomain:  sessionConfig.Session.Cookie.Domain,
		MaxAge:        sessionConfig.Session.MaxAge,
		Secure:        sessionConfig.Session.Cookie.Secure,
		HttpOnly:      sessionConfig.Session.Cookie.HttpOnly,
		SameSite:      sessionConfig.Session.Cookie.SameSite,
		CleanInterval: cleanInterval,
	}

	config.Infof("Session config loaded from session.yaml: enabled=%t, name=%s", 
		sessionCfg.Enabled, sessionCfg.CookieName)
	
	return sessionCfg
}