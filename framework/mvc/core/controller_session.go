package core

import (
	"github.com/zsy619/yyhertz/framework/mvc/session"
)

// ============= Session操作方法（委托给Manager） =============

// getSession 获取Session存储
func (c *BaseController) getSession() session.Store {
	if c.Ctx == nil {
		return nil
	}
	
	// 先检查是否已经在中间件中设置了 session
	if s, exists := c.Ctx.Request().Get("session"); exists {
		if store, ok := s.(session.Store); ok {
			return store
		}
	}
	
	// 获取Session管理器（优先使用全局管理器）
	var sessionManager *session.Manager
	
	// 尝试获取全局Session管理器
	if globalManager := getGlobalSessionManagerIfAvailable(); globalManager != nil {
		sessionManager = globalManager
	} else {
		// 如果全局管理器不可用，确保控制器有本地管理器
		if c.sessionHelper == nil {
			c.sessionHelper = session.NewManager(session.DefaultConfig())
		}
		sessionManager = c.sessionHelper
	}
	
	// 如果没有从中间件获取到Session，创建一个新的并保存到 context 中
	store := sessionManager.GetOrCreateSession(c.Ctx.Request())
	if store != nil {
		// 将 session store 保存到 context 中，确保后续调用使用相同的 store
		c.Ctx.Request().Set("session", store)
	}
	
	return store
}

// SetSession 设置Session数据
func (c *BaseController) SetSession(key string, value any) {
	if store := c.getSession(); store != nil {
		store.Set(key, value)
	}
}

// GetSession 获取Session数据
func (c *BaseController) GetSession(key string) any {
	if store := c.getSession(); store != nil {
		return store.Get(key)
	}
	return nil
}

// DeleteSession 删除Session数据
func (c *BaseController) DeleteSession(key string) {
	if store := c.getSession(); store != nil {
		store.Delete(key)
	}
}

// HasSession 检查Session数据是否存在
func (c *BaseController) HasSession(key string) bool {
	if store := c.getSession(); store != nil {
		return store.Exists(key)
	}
	return false
}

// GetSessionID 获取Session ID
func (c *BaseController) GetSessionID() string {
	if store := c.getSession(); store != nil {
		return store.GetID()
	}
	return ""
}

// DestroySession 删除全部Session数据
func (c *BaseController) DestroySession() {
	if store := c.getSession(); store != nil {
		store.Clear()
	}
}
