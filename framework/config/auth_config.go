package config

import (
	"strings"
)

// AuthConfig 认证配置结构
type AuthConfig struct {
	// CAS相关配置
	CasEnabled              bool   `json:"casEnabled"`
	CasHost                 string `json:"casHost"`
	CasLoginPath            string `json:"casLoginPath"`
	CasLogoutPath           string `json:"casLogoutPath"`
	CasValidatePath         string `json:"casValidatePath"`
	CaseServiceValidatePath string `json:"caseServiceValidatePath"`

	// 登录路径配置
	LoginPathOfSchool string `json:"loginPathOfSchool"`
	LoginPathOfAdmin  string `json:"loginPathOfAdmin"`

	// CAS登录路径配置
	LoginCasPathOfSchool string `json:"loginCasPathOfSchool"`
	LoginCasPathOfAdmin  string `json:"loginCasPathOfAdmin"`

	// 通用登录配置
	LoginURI       string `json:"loginURI"`
	LoginMobileURI string `json:"loginMobileURI"`
	LogoutURI      string `json:"logoutURI"`

	// 应用配置
	AppName     string `json:"appName"`
	LocalDomain string `json:"localDomain"`
}

// NewAuthConfig 创建认证配置
func NewAuthConfig() *AuthConfig {
	return &AuthConfig{
		LoginURI:       "/admin/login",
		LoginMobileURI: "/mobile/login",
		LogoutURI:      "/admin/logout",
	}
}

// InitAuthConfig 初始化认证配置
func (ac *AuthConfig) InitAuthConfig() {
	// 从配置管理器获取配置
	ac.CasEnabled = GetBool("cas.enabled", false)
	ac.CasHost = strings.TrimRight(Get("cas.url", ""), "/")
	ac.AppName = Get("appname", "Hertz MVC")
	ac.LocalDomain = Get("site.domain", "")

	// 构建CAS相关路径
	if ac.CasHost != "" {
		ac.CasLoginPath = ac.CasHost + "/cas/login"
		ac.CasLogoutPath = ac.CasHost + "/cas/logout"
		ac.CasValidatePath = ac.CasHost + "/cas/validate"
		ac.CaseServiceValidatePath = ac.CasHost + "/cas/serviceValidate"
	}

	// 构建登录路径
	if ac.LocalDomain != "" {
		ac.LoginPathOfSchool = ac.GetLocalDomain() + "/school/login"
		ac.LoginPathOfAdmin = ac.GetLocalDomain() + "/admin/login"

		if ac.CasLoginPath != "" {
			ac.LoginCasPathOfSchool = ac.CasLoginPath + "?service=" + ac.LoginPathOfSchool
			ac.LoginCasPathOfAdmin = ac.CasLoginPath + "?admin=1&service=" + ac.LoginPathOfAdmin
		}
	}
}

// GetLocalDomain 获取本地域名
func (ac *AuthConfig) GetLocalDomain() string {
	return ac.LocalDomain
}

// GetLocalDomainWithSlash 获取带斜杠的本地域名
func (ac *AuthConfig) GetLocalDomainWithSlash() string {
	if ac.LocalDomain == "" {
		return ""
	}
	if strings.HasSuffix(ac.LocalDomain, "/") {
		return ac.LocalDomain
	}
	return ac.LocalDomain + "/"
}

// IsCasEnabled 检查CAS是否启用
func (ac *AuthConfig) IsCasEnabled() bool {
	return ac.CasEnabled
}

// GetCasLoginURL 获取CAS登录URL
func (ac *AuthConfig) GetCasLoginURL(service string) string {
	if !ac.CasEnabled || ac.CasLoginPath == "" {
		return ""
	}
	return ac.CasLoginPath + "?service=" + service
}

// GetCasLogoutURL 获取CAS登出URL
func (ac *AuthConfig) GetCasLogoutURL(service string) string {
	if !ac.CasEnabled || ac.CasLogoutPath == "" {
		return ""
	}
	if service != "" {
		return ac.CasLogoutPath + "?service=" + service
	}
	return ac.CasLogoutPath
}

// GetCasValidateURL 获取CAS验证URL
func (ac *AuthConfig) GetCasValidateURL(service, ticket string) string {
	if !ac.CasEnabled || ac.CasValidatePath == "" {
		return ""
	}
	return ac.CasValidatePath + "?service=" + service + "&ticket=" + ticket
}

// 全局认证配置实例
var DefaultAuthConfig = NewAuthConfig()

// 便捷函数

// InitAuth 初始化认证配置
func InitAuth() {
	DefaultAuthConfig.InitAuthConfig()
}

// GetAuthConfig 获取认证配置
func GetAuthConfig() *AuthConfig {
	return DefaultAuthConfig
}

// SetAuthConfig 设置认证配置
func SetAuthConfig(config *AuthConfig) {
	DefaultAuthConfig = config
}

// 包级别的便捷函数

// GetAppName 获取应用名称
func GetAppName() string {
	return DefaultAuthConfig.AppName
}

// GetLoginURI 获取登录URI
func GetLoginURI() string {
	return DefaultAuthConfig.LoginURI
}

// GetLogoutURI 获取登出URI
func GetLogoutURI() string {
	return DefaultAuthConfig.LogoutURI
}

// GetLocalDomain 获取本地域名
func GetLocalDomain() string {
	return DefaultAuthConfig.GetLocalDomain()
}

// GetLocalDomainWithSlash 获取带斜杠的本地域名
func GetLocalDomainWithSlash() string {
	return DefaultAuthConfig.GetLocalDomainWithSlash()
}

// IsCasEnabled 检查CAS是否启用
func IsCasEnabled() bool {
	return DefaultAuthConfig.IsCasEnabled()
}

// GetCasLoginURL 获取CAS登录URL
func GetCasLoginURL(service string) string {
	return DefaultAuthConfig.GetCasLoginURL(service)
}

// GetCasLogoutURL 获取CAS登出URL
func GetCasLogoutURL(service string) string {
	return DefaultAuthConfig.GetCasLogoutURL(service)
}

// GetCasValidateURL 获取CAS验证URL
func GetCasValidateURL(service, ticket string) string {
	return DefaultAuthConfig.GetCasValidateURL(service, ticket)
}

// C 配置获取的简写函数(兼容原代码)
func C(name, defaultValue string) string {
	return Get(name, defaultValue)
}

// C_LOCAL_DOMAIN 获取本地域名(兼容原代码)
func C_LOCAL_DOMAIN() string {
	return GetLocalDomain()
}

// C_LOCAL_DOMAIN_Backslash 获取带斜杠的本地域名(兼容原代码)
func C_LOCAL_DOMAIN_Backslash() string {
	return GetLocalDomainWithSlash()
}

// 初始化认证配置
func init() {
	// 延迟初始化，确保配置管理器已经加载完成
	DefaultAuthConfig.InitAuthConfig()
}