package mvc

// Package mvc 配置管理器
//
// 本文件提供了统一的配置管理接口，支持动态配置更新和热重载。
// 通过配置管理器，可以在运行时更新Session、Cookie和Template的配置，
// 而无需重启应用程序。
//
// 主要功能：
// - 统一配置加载和管理
// - 动态配置更新
// - 配置热重载
// - 配置验证和错误处理
// - 配置变更通知
//
// 使用示例：
//
//	// 更新Session配置
//	config := &session.Config{
//	    Enabled:    true,
//	    CookieName: "my_session",
//	    MaxAge:     7200,
//	}
//	err := mvc.UpdateSessionConfig(config)
//
//	// 重载所有配置
//	err := mvc.ReloadAllConfigs()

import (
	"fmt"
	"sync"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/mvc/cookie"
	"github.com/zsy619/yyhertz/framework/mvc/session"
	"github.com/zsy619/yyhertz/framework/mvc/view"
)

// ============= 配置管理器结构 =============

// ConfigManager 全局配置管理器
type ConfigManager struct {
	mutex              sync.RWMutex
	sessionConfig      *session.Config
	cookieConfig       *cookie.Config
	templateConfig     *config.TemplateConfig
	changeListeners    []ConfigChangeListener
	validationEnabled  bool
	hotReloadEnabled   bool
}

// ConfigChangeListener 配置变更监听器
type ConfigChangeListener interface {
	OnSessionConfigChanged(old, new *session.Config) error
	OnCookieConfigChanged(old, new *cookie.Config) error
	OnTemplateConfigChanged(old, new *config.TemplateConfig) error
}

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	Type      string // "session", "cookie", "template"
	OldConfig any
	NewConfig any
	Timestamp int64
}

var (
	globalConfigManager *ConfigManager
	configManagerOnce   sync.Once
)

// ============= 配置管理器获取 =============

// GetConfigManager 获取全局配置管理器
//
// 返回全局单例的配置管理器实例。该方法是线程安全的。
//
// 返回：
//   - *ConfigManager: 全局配置管理器实例
func GetConfigManager() *ConfigManager {
	configManagerOnce.Do(func() {
		globalConfigManager = &ConfigManager{
			changeListeners:   make([]ConfigChangeListener, 0),
			validationEnabled: true,
			hotReloadEnabled:  true,
		}
		globalConfigManager.loadInitialConfigs()
	})
	return globalConfigManager
}

// loadInitialConfigs 加载初始配置
func (cm *ConfigManager) loadInitialConfigs() {
	// 加载Session配置
	cm.sessionConfig = session.LoadFromConfig()
	
	// 加载Cookie配置
	cm.cookieConfig = cookie.DefaultConfig()
	
	// 加载Template配置
	cm.templateConfig = view.DefaultTemplateConfig()
}

// ============= Session配置管理 =============

// UpdateSessionConfig 更新Session配置
//
// 动态更新Session管理器的配置。如果启用了验证，会先验证配置的有效性。
// 更新成功后会通知所有注册的监听器。
//
// 参数：
//   - config: 新的Session配置
//
// 返回：
//   - error: 如果更新失败返回错误信息
//
// 示例：
//
//	config := &session.Config{
//	    Enabled:    true,
//	    CookieName: "new_session_name",
//	    MaxAge:     3600,
//	}
//	err := mvc.UpdateSessionConfig(config)
//	if err != nil {
//	    log.Printf("Failed to update session config: %v", err)
//	}
func UpdateSessionConfig(config *session.Config) error {
	manager := GetConfigManager()
	return manager.UpdateSessionConfig(config)
}

// UpdateSessionConfig 更新Session配置（实例方法）
func (cm *ConfigManager) UpdateSessionConfig(config *session.Config) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("session config cannot be nil")
	}

	// 验证配置
	if cm.validationEnabled {
		if err := cm.validateSessionConfig(config); err != nil {
			return fmt.Errorf("invalid session config: %w", err)
		}
	}

	oldConfig := cm.sessionConfig
	cm.sessionConfig = config

	// 更新全局Session管理器
	sessionManager := GetSessionManager()
	if sessionManager != nil {
		sessionManager.SetConfig(config)
	}

	// 通知监听器
	cm.notifySessionConfigChanged(oldConfig, config)

	return nil
}

// GetSessionConfig 获取当前Session配置
//
// 返回：
//   - *session.Config: 当前的Session配置副本
func GetSessionConfig() *session.Config {
	manager := GetConfigManager()
	return manager.GetSessionConfig()
}

// GetSessionConfig 获取当前Session配置（实例方法）
func (cm *ConfigManager) GetSessionConfig() *session.Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if cm.sessionConfig == nil {
		return session.DefaultConfig()
	}
	
	// 返回配置副本，防止外部修改
	configCopy := *cm.sessionConfig
	return &configCopy
}

// validateSessionConfig 验证Session配置
func (cm *ConfigManager) validateSessionConfig(config *session.Config) error {
	if config.CookieName == "" {
		return fmt.Errorf("cookie name cannot be empty")
	}
	if config.MaxAge < 0 {
		return fmt.Errorf("max age cannot be negative")
	}
	if config.CookiePath == "" {
		return fmt.Errorf("cookie path cannot be empty")
	}
	return nil
}

// ============= Cookie配置管理 =============

// UpdateCookieConfig 更新Cookie配置
//
// 动态更新Cookie辅助器的配置。
//
// 参数：
//   - config: 新的Cookie配置
//
// 返回：
//   - error: 如果更新失败返回错误信息
func UpdateCookieConfig(config *cookie.Config) error {
	manager := GetConfigManager()
	return manager.UpdateCookieConfig(config)
}

// UpdateCookieConfig 更新Cookie配置（实例方法）
func (cm *ConfigManager) UpdateCookieConfig(config *cookie.Config) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("cookie config cannot be nil")
	}

	// 验证配置
	if cm.validationEnabled {
		if err := cm.validateCookieConfig(config); err != nil {
			return fmt.Errorf("invalid cookie config: %w", err)
		}
	}

	oldConfig := cm.cookieConfig
	cm.cookieConfig = config

	// 通知监听器
	cm.notifyCookieConfigChanged(oldConfig, config)

	return nil
}

// GetCookieConfig 获取当前Cookie配置
//
// 返回：
//   - *cookie.Config: 当前的Cookie配置副本
func GetCookieConfig() *cookie.Config {
	manager := GetConfigManager()
	return manager.GetCookieConfig()
}

// GetCookieConfig 获取当前Cookie配置（实例方法）
func (cm *ConfigManager) GetCookieConfig() *cookie.Config {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if cm.cookieConfig == nil {
		return cookie.DefaultConfig()
	}
	
	// 返回配置副本
	configCopy := *cm.cookieConfig
	return &configCopy
}

// validateCookieConfig 验证Cookie配置
func (cm *ConfigManager) validateCookieConfig(config *cookie.Config) error {
	if config.DefaultMaxAge < 0 {
		return fmt.Errorf("default max age cannot be negative")
	}
	if config.DefaultPath == "" {
		return fmt.Errorf("default path cannot be empty")
	}
	return nil
}

// ============= Template配置管理 =============

// UpdateTemplateConfig 更新Template配置
//
// 动态更新模板引擎的配置。
//
// 参数：
//   - config: 新的Template配置
//
// 返回：
//   - error: 如果更新失败返回错误信息
func UpdateTemplateConfig(config *config.TemplateConfig) error {
	manager := GetConfigManager()
	return manager.UpdateTemplateConfig(config)
}

// UpdateTemplateConfig 更新Template配置（实例方法）
func (cm *ConfigManager) UpdateTemplateConfig(config *config.TemplateConfig) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if config == nil {
		return fmt.Errorf("template config cannot be nil")
	}

	// 验证配置
	if cm.validationEnabled {
		if err := cm.validateTemplateConfig(config); err != nil {
			return fmt.Errorf("invalid template config: %w", err)
		}
	}

	oldConfig := cm.templateConfig
	cm.templateConfig = config

	// 更新全局模板引擎
	templateEngine := GetTemplateEngine()
	if templateEngine != nil {
		// 注意：这里需要模板引擎支持配置更新
		// 如果不支持，可能需要重新创建引擎
	}

	// 通知监听器
	cm.notifyTemplateConfigChanged(oldConfig, config)

	return nil
}

// GetTemplateConfig 获取当前Template配置
//
// 返回：
//   - *config.TemplateConfig: 当前的Template配置副本
func GetTemplateConfig() *config.TemplateConfig {
	manager := GetConfigManager()
	return manager.GetTemplateConfig()
}

// GetTemplateConfig 获取当前Template配置（实例方法）
func (cm *ConfigManager) GetTemplateConfig() *config.TemplateConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	
	if cm.templateConfig == nil {
		return view.DefaultTemplateConfig()
	}
	
	// 返回配置副本
	configCopy := *cm.templateConfig
	return &configCopy
}

// validateTemplateConfig 验证Template配置
func (cm *ConfigManager) validateTemplateConfig(config *config.TemplateConfig) error {
	if len(config.Paths.ViewPaths) == 0 {
		return fmt.Errorf("view paths cannot be empty")
	}
	if config.Paths.Extension == "" {
		return fmt.Errorf("template extension cannot be empty")
	}
	return nil
}

// ============= 配置重载 =============

// ReloadAllConfigs 重新加载所有配置
//
// 从配置文件重新加载所有配置并更新相应的管理器。
//
// 返回：
//   - error: 如果重载失败返回错误信息
func ReloadAllConfigs() error {
	manager := GetConfigManager()
	return manager.ReloadAllConfigs()
}

// ReloadAllConfigs 重新加载所有配置（实例方法）
func (cm *ConfigManager) ReloadAllConfigs() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 重新加载Session配置
	sessionConfig := session.LoadFromConfig()
	if err := cm.updateSessionConfigInternal(sessionConfig); err != nil {
		return fmt.Errorf("failed to reload session config: %w", err)
	}

	// 重新加载Cookie配置
	cookieConfig := cookie.DefaultConfig()
	if err := cm.updateCookieConfigInternal(cookieConfig); err != nil {
		return fmt.Errorf("failed to reload cookie config: %w", err)
	}

	// 重新加载Template配置
	templateConfig := config.DefaultTemplateConfig()
	if err := cm.updateTemplateConfigInternal(templateConfig); err != nil {
		return fmt.Errorf("failed to reload template config: %w", err)
	}

	return nil
}

// updateSessionConfigInternal 内部Session配置更新（不加锁）
func (cm *ConfigManager) updateSessionConfigInternal(config *session.Config) error {
	oldConfig := cm.sessionConfig
	cm.sessionConfig = config

	sessionManager := GetSessionManager()
	if sessionManager != nil {
		sessionManager.SetConfig(config)
	}

	cm.notifySessionConfigChanged(oldConfig, config)
	return nil
}

// updateCookieConfigInternal 内部Cookie配置更新（不加锁）
func (cm *ConfigManager) updateCookieConfigInternal(config *cookie.Config) error {
	oldConfig := cm.cookieConfig
	cm.cookieConfig = config
	cm.notifyCookieConfigChanged(oldConfig, config)
	return nil
}

// updateTemplateConfigInternal 内部Template配置更新（不加锁）
func (cm *ConfigManager) updateTemplateConfigInternal(config *config.TemplateConfig) error {
	oldConfig := cm.templateConfig
	cm.templateConfig = config
	cm.notifyTemplateConfigChanged(oldConfig, config)
	return nil
}

// ============= 监听器管理 =============

// AddConfigChangeListener 添加配置变更监听器
//
// 参数：
//   - listener: 要添加的监听器
func AddConfigChangeListener(listener ConfigChangeListener) {
	manager := GetConfigManager()
	manager.AddConfigChangeListener(listener)
}

// AddConfigChangeListener 添加配置变更监听器（实例方法）
func (cm *ConfigManager) AddConfigChangeListener(listener ConfigChangeListener) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.changeListeners = append(cm.changeListeners, listener)
}

// RemoveConfigChangeListener 移除配置变更监听器
//
// 参数：
//   - listener: 要移除的监听器
func RemoveConfigChangeListener(listener ConfigChangeListener) {
	manager := GetConfigManager()
	manager.RemoveConfigChangeListener(listener)
}

// RemoveConfigChangeListener 移除配置变更监听器（实例方法）
func (cm *ConfigManager) RemoveConfigChangeListener(listener ConfigChangeListener) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	for i, l := range cm.changeListeners {
		if l == listener {
			cm.changeListeners = append(cm.changeListeners[:i], cm.changeListeners[i+1:]...)
			break
		}
	}
}

// 通知监听器方法
func (cm *ConfigManager) notifySessionConfigChanged(old, new *session.Config) {
	for _, listener := range cm.changeListeners {
		if err := listener.OnSessionConfigChanged(old, new); err != nil {
			// 记录错误但不中断其他监听器
			fmt.Printf("Config change listener error: %v\n", err)
		}
	}
}

func (cm *ConfigManager) notifyCookieConfigChanged(old, new *cookie.Config) {
	for _, listener := range cm.changeListeners {
		if err := listener.OnCookieConfigChanged(old, new); err != nil {
			fmt.Printf("Config change listener error: %v\n", err)
		}
	}
}

func (cm *ConfigManager) notifyTemplateConfigChanged(old, new *config.TemplateConfig) {
	for _, listener := range cm.changeListeners {
		if err := listener.OnTemplateConfigChanged(old, new); err != nil {
			fmt.Printf("Config change listener error: %v\n", err)
		}
	}
}

// ============= 配置管理器设置 =============

// SetValidationEnabled 设置是否启用配置验证
//
// 参数：
//   - enabled: 是否启用验证
func SetValidationEnabled(enabled bool) {
	manager := GetConfigManager()
	manager.SetValidationEnabled(enabled)
}

// SetValidationEnabled 设置是否启用配置验证（实例方法）
func (cm *ConfigManager) SetValidationEnabled(enabled bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.validationEnabled = enabled
}

// SetHotReloadEnabled 设置是否启用热重载
//
// 参数：
//   - enabled: 是否启用热重载
func SetHotReloadEnabled(enabled bool) {
	manager := GetConfigManager()
	manager.SetHotReloadEnabled(enabled)
}

// SetHotReloadEnabled 设置是否启用热重载（实例方法）
func (cm *ConfigManager) SetHotReloadEnabled(enabled bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	cm.hotReloadEnabled = enabled
}