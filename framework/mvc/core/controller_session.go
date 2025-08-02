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
	if s, exists := c.Ctx.RequestContext.Get("session"); exists {
		if store, ok := s.(session.Store); ok {
			return store
		}
	}
	// 如果没有从中间件获取到Session，创建一个新的
	return c.sessionHelper.GetOrCreateSession(c.Ctx.RequestContext)
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