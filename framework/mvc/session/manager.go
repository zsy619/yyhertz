package session

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

// Manager Session管理器
type Manager struct {
	config *Config
}

// NewManager 创建Session管理器
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}
	return &Manager{
		config: config,
	}
}

// NewManagerFromConfig 从配置文件创建Session管理器
func NewManagerFromConfig() *Manager {
	config := LoadFromConfig()
	return &Manager{
		config: config,
	}
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *Config {
	return m.config
}

// SetConfig 设置配置
func (m *Manager) SetConfig(config *Config) {
	if config != nil {
		m.config = config
	}
}

// IsEnabled 检查Session是否启用
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}

// Enable 启用Session
func (m *Manager) Enable() {
	m.config.Enabled = true
}

// Disable 禁用Session
func (m *Manager) Disable() {
	m.config.Enabled = false
}

// generateSessionID 生成Session ID
func (m *Manager) generateSessionID() string {
	return fmt.Sprintf("sess_%d_%d", time.Now().UnixNano(), time.Now().UnixNano()%100000)
}

// GetOrCreateSession 获取或创建Session
func (m *Manager) GetOrCreateSession(ctx *app.RequestContext) Store {
	if !m.IsEnabled() {
		return nil
	}

	// 尝试从Cookie获取Session ID
	sessionID := string(ctx.Cookie(m.config.CookieName))
	if sessionID == "" {
		// 生成新的Session ID
		sessionID = m.generateSessionID()
		
		// 设置Session Cookie
		cookie := fmt.Sprintf("%s=%s; Max-Age=%d; Path=%s; HttpOnly=%t; SameSite=%s",
			m.config.CookieName,
			sessionID,
			m.config.MaxAge,
			m.config.CookiePath,
			m.config.HttpOnly,
			m.config.SameSite,
		)
		
		if m.config.CookieDomain != "" {
			cookie += "; Domain=" + m.config.CookieDomain
		}
		if m.config.Secure {
			cookie += "; Secure"
		}
		
		ctx.Header("Set-Cookie", cookie)
	}

	// 创建Session存储
	return NewMemoryStore(sessionID)
}

// DestroySession 销毁Session
func (m *Manager) DestroySession(ctx *app.RequestContext) {
	// 删除Session Cookie
	cookie := fmt.Sprintf("%s=; Max-Age=-1; Path=%s",
		m.config.CookieName, m.config.CookiePath)
	
	if m.config.CookieDomain != "" {
		cookie += "; Domain=" + m.config.CookieDomain
	}
	
	ctx.Header("Set-Cookie", cookie)
}

// Middleware Session中间件
func (m *Manager) Middleware() func(context.Context, *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		if !m.IsEnabled() {
			return
		}

		// 获取或创建Session
		session := m.GetOrCreateSession(c)
		if session != nil {
			// 将Session存储到上下文中
			c.Set("session", session)
			c.Set("session_id", session.GetID())
		}
	}
}

// StartCleanup 启动清理任务
func (m *Manager) StartCleanup() {
	if !m.IsEnabled() || m.config.CleanInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(m.config.CleanInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				m.cleanup()
			}
		}
	}()
}

// cleanup 清理过期Session
func (m *Manager) cleanup() {
	// 这里可以实现清理逻辑，比如从存储中删除过期的Session
	// 目前使用内存存储，暂时不需要特殊处理
}