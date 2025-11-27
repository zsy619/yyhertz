package security

// CSRFTokenAdapter CSRF token适配器
//
// 此适配器实现了view包中的CSRFTokenProvider接口，
// 用于解决循环导入问题
type CSRFTokenAdapter struct {
	manager *CSRFManager
}

// NewCSRFTokenAdapter 创建新的CSRF token适配器
func NewCSRFTokenAdapter(manager *CSRFManager) *CSRFTokenAdapter {
	return &CSRFTokenAdapter{
		manager: manager,
	}
}

// GenerateSimpleToken 实现CSRFTokenProvider接口
func (a *CSRFTokenAdapter) GenerateSimpleToken() string {
	if a.manager == nil {
		return "csrf-adapter-manager-nil"
	}
	return a.manager.GenerateSimpleToken()
}