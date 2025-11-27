package errors

import "embed"

// =============================================================================
// 模块：常量定义
// 职责：定义所有错误处理相关的常量和嵌入资源
// =============================================================================

//go:embed templates/*
var TplFS embed.FS

const (
	// 默认配置常量
	DefaultLanguage = "zh-CN"
	DefaultTitle    = "YYHertz Framework"
	DefaultPriority = 1000

	// 统计常量
	MaxLastErrors = 50

	// 钩子阶段常量
	HookPhaseBefore = "before"
	HookPhaseAfter  = "after"

	// 支持的语言代码
	LanguageZhCN = "zh-CN"
	LanguageZh   = "zh"
	LanguageEnUS = "en-US"
	LanguageEn   = "en"
)

// initRetryableErrors 初始化可重试的错误类型
func InitRetryableErrors() map[int]bool {
	return map[int]bool{
		408: true, // Request Timeout
		429: true, // Too Many Requests
		500: true, // Internal Server Error
		502: true, // Bad Gateway
		503: true, // Service Unavailable
		504: true, // Gateway Timeout
	}
}