package unified

import "errors"

// 统一管理器相关错误定义
var (
	// ErrManagerNotInitialized 管理器未初始化错误
	ErrManagerNotInitialized = errors.New("unified manager is not initialized")

	// ErrCookieHelperNotInitialized Cookie辅助器未初始化错误
	ErrCookieHelperNotInitialized = errors.New("cookie helper is not initialized")

	// ErrSessionManagerNotInitialized Session管理器未初始化错误
	ErrSessionManagerNotInitialized = errors.New("session manager is not initialized")

	// ErrTemplateEngineNotInitialized 模板引擎未初始化错误
	ErrTemplateEngineNotInitialized = errors.New("template engine is not initialized")

	// ErrCSRFManagerNotInitialized CSRF管理器未初始化错误
	ErrCSRFManagerNotInitialized = errors.New("csrf manager is not initialized")

	// ErrContextProviderNotInitialized 上下文提供者未初始化错误
	ErrContextProviderNotInitialized = errors.New("context provider is not initialized")

	// ErrInvalidContext 无效上下文错误
	ErrInvalidContext = errors.New("invalid context provided")

	// ErrFilterNotFound 过滤器未找到错误
	ErrFilterNotFound = errors.New("filter not found")

	// ErrInvalidFilterPosition 无效过滤器位置错误
	ErrInvalidFilterPosition = errors.New("invalid filter position")

	// ErrFilterAlreadyExists 过滤器已存在错误
	ErrFilterAlreadyExists = errors.New("filter already exists")
)