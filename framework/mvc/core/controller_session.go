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
	
	// 1. 先检查是否已经在中间件或前一次调用中缓存了 session
	if s, exists := c.Ctx.Request().Get("session"); exists {
		if store, ok := s.(session.Store); ok {
			return store
		}
	}
	
	// 2. 如果有统一管理器，优先使用它的session管理器
	// 注意：现在主要的Session操作已经在具体方法中优先使用统一管理器了
	// 这里主要为 GetSessionID 和 DestroySession 等方法提供支持
	
	// 3. 尝试获取全局Session管理器
	var sessionManager *session.Manager
	if globalManager := getGlobalSessionManagerIfAvailable(); globalManager != nil {
		sessionManager = globalManager
	} else {
		// 4. 如果全局管理器不可用，确保控制器有本地管理器
		if c.sessionHelper == nil {
			c.sessionHelper = session.NewManager(session.DefaultConfig())
		}
		sessionManager = c.sessionHelper
	}
	
	// 创建Session并缓存到Context中
	store := sessionManager.GetOrCreateSession(c.Ctx.Request())
	if store != nil {
		// 将 session store 保存到 context 中，确保后续调用使用相同的 store
		c.Ctx.Request().Set("session", store)
	}
	
	return store
}

// SetSession 设置Session数据
func (c *BaseController) SetSession(key string, value any) {
	// 优先使用统一管理器
	if c.unifiedManager != nil && c.unifiedManager.IsInitialized() {
		c.unifiedManager.SetSessionData(c.Ctx, key, value)
		return
	}
	
	// 回退到传统方法
	if store := c.getSession(); store != nil {
		store.Set(key, value)
	}
}

// GetSession 获取Session数据
func (c *BaseController) GetSession(key string) any {
	// 优先使用统一管理器
	if c.unifiedManager != nil && c.unifiedManager.IsInitialized() {
		return c.unifiedManager.GetSessionData(c.Ctx, key)
	}
	
	// 回退到传统方法
	if store := c.getSession(); store != nil {
		return store.Get(key)
	}
	return nil
}

// DeleteSession 删除Session数据
func (c *BaseController) DeleteSession(key string) {
	// 优先使用统一管理器
	if c.unifiedManager != nil && c.unifiedManager.IsInitialized() {
		c.unifiedManager.DeleteSessionData(c.Ctx, key)
		return
	}
	
	// 回退到传统方法
	if store := c.getSession(); store != nil {
		store.Delete(key)
	}
}

// HasSession 检查Session数据是否存在
func (c *BaseController) HasSession(key string) bool {
	// 优先使用统一管理器
	if c.unifiedManager != nil && c.unifiedManager.IsInitialized() {
		// 统一管理器没有直接的HasSession方法，使用Get来检查
		return c.unifiedManager.GetSessionData(c.Ctx, key) != nil
	}
	
	// 回退到传统方法
	if store := c.getSession(); store != nil {
		return store.Exists(key)
	}
	return false
}

// GetSessionID 获取Session ID
func (c *BaseController) GetSessionID() string {
	// 统一管理器没有直接的GetSessionID方法，使用传统方法
	if store := c.getSession(); store != nil {
		return store.GetID()
	}
	return ""
}

// DestroySession 删除全部Session数据
func (c *BaseController) DestroySession() {
	// 统一管理器没有直接的DestroySession方法，使用传统方法
	if store := c.getSession(); store != nil {
		store.Clear()
	}
}
