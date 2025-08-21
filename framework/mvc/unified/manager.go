// Package unified 提供Cookie、Session、CSRF、模板功能的统一管理
//
// 本包通过统一管理器模式，将分散的功能组件集中管理，
// 提供一致的API接口，减少重复代码，提高系统性能和可维护性。
//
// 主要功能：
// - 统一的Cookie、Session、CSRF、模板管理
// - 高性能的上下文数据存取
// - 全局过滤器支持
// - 类型安全的数据操作
//
// 使用示例：
//
//	// 初始化统一管理器
//	manager := unified.GetManager()
//
//	// 使用统一接口操作数据
//	session := manager.GetSession(ctx)
//	session.Set("user", userInfo)
//
//	// 使用过滤器
//	manager.AddFilter("/*", unified.FilterSSO)
package unified

import (
	"sync"

	context "github.com/zsy619/yyhertz/framework/mvc/context"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/security"
	"github.com/zsy619/yyhertz/framework/mvc/session"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// Manager 统一管理器，集成所有核心功能
//
// 设计原则：
// - 单例模式：全局唯一实例，减少内存占用
// - 线程安全：支持高并发访问
// - 懒加载：按需初始化组件
// - 插件化：支持功能扩展
//
// 实现了context.UnifiedManagerInterface接口，提供统一管理功能
type Manager struct {
	// ============= 核心组件 =============
	cookieHelper   *cookie.Helper        // Cookie管理器
	sessionManager *session.Manager      // Session管理器
	templateEngine *view.TemplateEngine  // 模板引擎
	csrfManager    *security.CSRFManager // CSRF管理器

	// ============= 统一数据管理 =============
	contextProvider *ContextProvider // 上下文数据提供者

	// ============= 过滤器管理 =============
	filters     []Filter     // 全局过滤器列表
	filterMutex sync.RWMutex // 过滤器操作锁

	// ============= 初始化控制 =============
	initialized bool         // 是否已初始化
	initOnce    sync.Once    // 初始化控制
	mutex       sync.RWMutex // 并发控制锁
}

// 全局管理器实例
var (
	globalManager *Manager
	managerOnce   sync.Once
)

// GetManager 获取全局统一管理器实例
//
// 使用单例模式确保全局唯一性，线程安全。
// 如果尚未初始化，会自动调用初始化流程。
//
// 返回：
//   - *Manager: 全局统一管理器实例
//
// 示例：
//
//	manager := unified.GetManager()
//	session := manager.GetSession(ctx)
func GetManager() *Manager {
	managerOnce.Do(func() {
		globalManager = &Manager{
			filters:         make([]Filter, 0),
			contextProvider: NewContextProvider(),
		}
		globalManager.initialize()
	})
	return globalManager
}

// initialize 初始化管理器
//
// 按需加载各个组件，如果外部已经设置了全局管理器，
// 则优先使用外部实例，否则创建默认实例。
func (m *Manager) initialize() {
	m.initOnce.Do(func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		// 初始化Cookie管理器
		if m.cookieHelper == nil {
			// 尝试使用全局实例
			if globalHelper := getGlobalCookieHelper(); globalHelper != nil {
				m.cookieHelper = globalHelper
			} else {
				m.cookieHelper = cookie.NewHelper(cookie.DefaultConfig())
			}
		}

		// 初始化Session管理器
		if m.sessionManager == nil {
			// 尝试使用全局实例
			if globalManager := getGlobalSessionManager(); globalManager != nil {
				m.sessionManager = globalManager
			} else {
				m.sessionManager = session.NewManager(session.DefaultConfig())
			}
		}

		// 初始化模板引擎
		if m.templateEngine == nil {
			// 尝试使用全局实例
			if globalEngine := getGlobalTemplateEngine(); globalEngine != nil {
				m.templateEngine = globalEngine
			} else {
				if engine, err := view.NewTemplateEngine(view.DefaultTemplateConfig()); err == nil {
					m.templateEngine = engine
				}
			}
		}

		// 初始化CSRF管理器
		if m.csrfManager == nil {
			// 尝试使用全局实例
			if globalCSRF := getGlobalCSRFManager(); globalCSRF != nil {
				m.csrfManager = globalCSRF
			} else {
				m.csrfManager = security.NewCSRFManager(security.DefaultCSRFConfig())
			}
		}

		m.initialized = true

		// 将自己注册到context包的全局管理器访问器中
		context.SetGlobalUnifiedManager(m)
	})
}

// ============= Cookie 操作 =============

// SetCookie 设置Cookie
func (m *Manager) SetCookie(ctx *context.Context, name, value string, options ...interface{}) {
	if ctx == nil || m.cookieHelper == nil {
		return
	}
	// 转换interface{}参数为*cookie.Options
	var cookieOpts []*cookie.Options
	for _, opt := range options {
		if cookieOpt, ok := opt.(*cookie.Options); ok {
			cookieOpts = append(cookieOpts, cookieOpt)
		}
	}
	m.cookieHelper.Set(ctx.Request(), name, value, cookieOpts...)
}

// GetCookie 获取Cookie值
func (m *Manager) GetCookie(ctx *context.Context, name string) string {
	if ctx == nil || m.cookieHelper == nil {
		return ""
	}
	return m.cookieHelper.Get(ctx.Request(), name)
}

// DeleteCookie 删除Cookie
func (m *Manager) DeleteCookie(ctx *context.Context, name string, path ...string) {
	if ctx == nil || m.cookieHelper == nil {
		return
	}
	m.cookieHelper.Delete(ctx.Request(), name, path...)
}

// HasCookie 检查Cookie是否存在
func (m *Manager) HasCookie(ctx *context.Context, name string) bool {
	if ctx == nil || m.cookieHelper == nil {
		return false
	}
	return m.cookieHelper.Has(ctx.Request(), name)
}

// ============= Session 操作 =============

// GetSession 获取Session存储
func (m *Manager) GetSession(ctx *context.Context) session.Store {
	if ctx == nil || m.sessionManager == nil {
		return nil
	}

	// 1. 先检查是否已经在中间件或前一次调用中缓存了 session
	if s, exists := ctx.Request().Get("unified_session"); exists {
		if store, ok := s.(session.Store); ok {
			return store
		}
	}

	// 2. 创建Session并缓存到Context中
	store := m.sessionManager.GetOrCreateSession(ctx.Request())
	if store != nil {
		// 将 session store 保存到 context 中，确保后续调用使用相同的 store
		// 使用不同的key避免与传统方法冲突
		ctx.Request().Set("unified_session", store)
	}

	return store
}

// SetSessionData 设置Session数据（便捷方法）
func (m *Manager) SetSessionData(ctx *context.Context, key string, value interface{}) {
	if store := m.GetSession(ctx); store != nil {
		store.Set(key, value)
	}
}

// GetSessionData 获取Session数据（便捷方法）
func (m *Manager) GetSessionData(ctx *context.Context, key string) interface{} {
	if store := m.GetSession(ctx); store != nil {
		return store.Get(key)
	}
	return nil
}

// DeleteSessionData 删除Session数据（便捷方法）
func (m *Manager) DeleteSessionData(ctx *context.Context, key string) {
	if store := m.GetSession(ctx); store != nil {
		store.Delete(key)
	}
}

// ============= CSRF 操作 =============

// GenerateCSRFToken 生成CSRF令牌
func (m *Manager) GenerateCSRFToken(userID, clientIP string) (*security.CSRFToken, error) {
	if m.csrfManager == nil {
		return nil, ErrCSRFManagerNotInitialized
	}
	return m.csrfManager.GenerateToken(userID, clientIP)
}

// ValidateCSRFToken 验证CSRF令牌
func (m *Manager) ValidateCSRFToken(tokenValue, userID, clientIP string) (bool, error) {
	if m.csrfManager == nil {
		return false, ErrCSRFManagerNotInitialized
	}
	return m.csrfManager.ValidateToken(tokenValue, userID, clientIP)
}

// ============= 模板 操作 =============

// RenderTemplate 渲染模板
func (m *Manager) RenderTemplate(templateName string, data interface{}) (string, error) {
	if m.templateEngine == nil {
		return "", ErrTemplateEngineNotInitialized
	}
	return m.templateEngine.Render(templateName, data)
}

// RenderTemplateWithLayout 使用布局渲染模板
func (m *Manager) RenderTemplateWithLayout(templateName, layoutName string, data interface{}) (string, error) {
	if m.templateEngine == nil {
		return "", ErrTemplateEngineNotInitialized
	}
	return m.templateEngine.RenderWithLayout(templateName, layoutName, data)
}

// ============= 上下文数据操作 =============

// SetContextData 设置上下文数据
func (m *Manager) SetContextData(ctx *context.Context, key string, value interface{}) {
	if m.contextProvider == nil {
		return
	}
	m.contextProvider.Set(ctx, key, value)
}

// GetContextData 获取上下文数据
func (m *Manager) GetContextData(ctx *context.Context, key string) interface{} {
	if m.contextProvider == nil {
		return nil
	}
	return m.contextProvider.Get(ctx, key)
}

// GetTypedContextData 获取类型安全的上下文数据
func (m *Manager) GetTypedContextData(ctx *context.Context, key string, target interface{}) (interface{}, bool) {
	if m.contextProvider == nil {
		return nil, false
	}
	return m.contextProvider.GetTypedValue(ctx, key, target)
}

// ============= 组件访问器 =============

// GetCookieHelper 获取Cookie辅助器
func (m *Manager) GetCookieHelper() *cookie.Helper {
	return m.cookieHelper
}

// GetSessionManager 获取Session管理器
func (m *Manager) GetSessionManager() *session.Manager {
	return m.sessionManager
}

// GetTemplateEngine 获取模板引擎
func (m *Manager) GetTemplateEngine() *view.TemplateEngine {
	return m.templateEngine
}

// GetCSRFManager 获取CSRF管理器
func (m *Manager) GetCSRFManager() *security.CSRFManager {
	return m.csrfManager
}

// ============= 状态查询 =============

// IsInitialized 检查是否已初始化
func (m *Manager) IsInitialized() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.initialized
}

// GetComponentStatus 获取各组件状态
func (m *Manager) GetComponentStatus() map[string]bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]bool{
		"manager":  m.initialized,
		"cookie":   m.cookieHelper != nil,
		"session":  m.sessionManager != nil,
		"template": m.templateEngine != nil,
		"csrf":     m.csrfManager != nil,
		"context":  m.contextProvider != nil,
	}
}

// ============= 辅助函数（外部全局管理器访问） =============

// 这些函数尝试从外部全局管理器获取实例，如果不可用则返回nil
// 这样可以优先使用已经初始化的全局实例，避免重复创建

// getGlobalCookieHelper 尝试获取全局Cookie辅助器
func getGlobalCookieHelper() *cookie.Helper {
	// 尝试从mvc包的全局管理器获取
	// 这里先返回nil，实际实现时会调用mvc.GetCookieHelper()
	return nil
}

// getGlobalSessionManager 尝试获取全局Session管理器
func getGlobalSessionManager() *session.Manager {
	// 尝试从mvc包的全局管理器获取
	return nil
}

// getGlobalTemplateEngine 尝试获取全局模板引擎
func getGlobalTemplateEngine() *view.TemplateEngine {
	// 尝试从mvc包的全局管理器获取
	return nil
}

// getGlobalCSRFManager 尝试获取全局CSRF管理器
func getGlobalCSRFManager() *security.CSRFManager {
	// 尝试从mvc包的全局管理器获取
	return nil
}
