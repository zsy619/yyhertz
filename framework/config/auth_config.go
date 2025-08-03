package config

import (
	"github.com/spf13/viper"
)

// AuthConfig 认证配置结构
type AuthConfig struct {
	// CAS相关配置
	CAS struct {
		Enabled             bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
		Host                string `mapstructure:"host" yaml:"host" json:"host"`
		LoginPath           string `mapstructure:"login_path" yaml:"login_path" json:"login_path"`
		LogoutPath          string `mapstructure:"logout_path" yaml:"logout_path" json:"logout_path"`
		ValidatePath        string `mapstructure:"validate_path" yaml:"validate_path" json:"validate_path"`
		ServiceValidatePath string `mapstructure:"service_validate_path" yaml:"service_validate_path" json:"service_validate_path"`
		TicketValidationURL string `mapstructure:"ticket_validation_url" yaml:"ticket_validation_url" json:"ticket_validation_url"`
		ServiceURL          string `mapstructure:"service_url" yaml:"service_url" json:"service_url"`
		Version             string `mapstructure:"version" yaml:"version" json:"version"` // CAS版本: 1.0, 2.0, 3.0
		SSOEnabled          bool   `mapstructure:"sso_enabled" yaml:"sso_enabled" json:"sso_enabled"`
		ProxyReceptorURL    string `mapstructure:"proxy_receptor_url" yaml:"proxy_receptor_url" json:"proxy_receptor_url"`
		RenewEnabled        bool   `mapstructure:"renew_enabled" yaml:"renew_enabled" json:"renew_enabled"`
		GatewayEnabled      bool   `mapstructure:"gateway_enabled" yaml:"gateway_enabled" json:"gateway_enabled"`
	} `mapstructure:"cas" yaml:"cas" json:"cas"`

	// 登录路径配置
	LoginPaths struct {
		School struct {
			Path    string `mapstructure:"path" yaml:"path" json:"path"`
			CASPath string `mapstructure:"cas_path" yaml:"cas_path" json:"cas_path"`
		} `mapstructure:"school" yaml:"school" json:"school"`
		Admin struct {
			Path    string `mapstructure:"path" yaml:"path" json:"path"`
			CASPath string `mapstructure:"cas_path" yaml:"cas_path" json:"cas_path"`
		} `mapstructure:"admin" yaml:"admin" json:"admin"`
		Common struct {
			LoginURI       string `mapstructure:"login_uri" yaml:"login_uri" json:"login_uri"`
			LoginMobileURI string `mapstructure:"login_mobile_uri" yaml:"login_mobile_uri" json:"login_mobile_uri"`
			LogoutURI      string `mapstructure:"logout_uri" yaml:"logout_uri" json:"logout_uri"`
		} `mapstructure:"common" yaml:"common" json:"common"`
	} `mapstructure:"login_paths" yaml:"login_paths" json:"login_paths"`

	// JWT配置
	JWT struct {
		Secret         string `mapstructure:"secret" yaml:"secret" json:"secret"`
		TokenTTL       int    `mapstructure:"token_ttl" yaml:"token_ttl" json:"token_ttl"`       // 小时
		RefreshTTL     int    `mapstructure:"refresh_ttl" yaml:"refresh_ttl" json:"refresh_ttl"` // 小时
		Issuer         string `mapstructure:"issuer" yaml:"issuer" json:"issuer"`
		SigningMethod  string `mapstructure:"signing_method" yaml:"signing_method" json:"signing_method"` // HS256, RS256, etc.
		PublicKeyPath  string `mapstructure:"public_key_path" yaml:"public_key_path" json:"public_key_path"`
		PrivateKeyPath string `mapstructure:"private_key_path" yaml:"private_key_path" json:"private_key_path"`
	} `mapstructure:"jwt" yaml:"jwt" json:"jwt"`

	// Session配置
	Session struct {
		Name     string   `mapstructure:"name" yaml:"name" json:"name"`
		Secret   string   `mapstructure:"secret" yaml:"secret" json:"secret"`
		MaxAge   int      `mapstructure:"max_age" yaml:"max_age" json:"max_age"` // 秒
		Path     string   `mapstructure:"path" yaml:"path" json:"path"`
		Domain   string   `mapstructure:"domain" yaml:"domain" json:"domain"`
		Secure   bool     `mapstructure:"secure" yaml:"secure" json:"secure"`
		HttpOnly bool     `mapstructure:"http_only" yaml:"http_only" json:"http_only"`
		SameSite string   `mapstructure:"same_site" yaml:"same_site" json:"same_site"` // Strict, Lax, None
		Store    string   `mapstructure:"store" yaml:"store" json:"store"`             // cookie, memory, redis
		KeyPairs []string `mapstructure:"key_pairs" yaml:"key_pairs" json:"key_pairs"`
	} `mapstructure:"session" yaml:"session" json:"session"`

	// OAuth配置
	OAuth struct {
		Providers map[string]OAuthProvider `mapstructure:"providers" yaml:"providers" json:"providers"`
	} `mapstructure:"oauth" yaml:"oauth" json:"oauth"`

	// 权限配置
	Authorization struct {
		Enable          bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
		DefaultRole     string   `mapstructure:"default_role" yaml:"default_role" json:"default_role"`
		AdminRoles      []string `mapstructure:"admin_roles" yaml:"admin_roles" json:"admin_roles"`
		GuestPaths      []string `mapstructure:"guest_paths" yaml:"guest_paths" json:"guest_paths"`             // 不需要认证的路径
		PublicPaths     []string `mapstructure:"public_paths" yaml:"public_paths" json:"public_paths"`          // 公开路径
		ProtectedPaths  []string `mapstructure:"protected_paths" yaml:"protected_paths" json:"protected_paths"` // 需要认证的路径
		RoleBasedAccess bool     `mapstructure:"role_based_access" yaml:"role_based_access" json:"role_based_access"`
	} `mapstructure:"authorization" yaml:"authorization" json:"authorization"`

	// 安全配置
	Security struct {
		PasswordPolicy struct {
			MinLength        int  `mapstructure:"min_length" yaml:"min_length" json:"min_length"`
			RequireUppercase bool `mapstructure:"require_uppercase" yaml:"require_uppercase" json:"require_uppercase"`
			RequireLowercase bool `mapstructure:"require_lowercase" yaml:"require_lowercase" json:"require_lowercase"`
			RequireNumbers   bool `mapstructure:"require_numbers" yaml:"require_numbers" json:"require_numbers"`
			RequireSymbols   bool `mapstructure:"require_symbols" yaml:"require_symbols" json:"require_symbols"`
			MaxAge           int  `mapstructure:"max_age" yaml:"max_age" json:"max_age"` // 天
		} `mapstructure:"password_policy" yaml:"password_policy" json:"password_policy"`

		LoginAttempts struct {
			MaxAttempts int `mapstructure:"max_attempts" yaml:"max_attempts" json:"max_attempts"`
			LockoutTime int `mapstructure:"lockout_time" yaml:"lockout_time" json:"lockout_time"` // 分钟
			ResetTime   int `mapstructure:"reset_time" yaml:"reset_time" json:"reset_time"`       // 分钟
		} `mapstructure:"login_attempts" yaml:"login_attempts" json:"login_attempts"`

		TwoFactorAuth struct {
			Enable        bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
			Issuer        string `mapstructure:"issuer" yaml:"issuer" json:"issuer"`
			SecretLength  int    `mapstructure:"secret_length" yaml:"secret_length" json:"secret_length"`
			QRCodeService string `mapstructure:"qr_code_service" yaml:"qr_code_service" json:"qr_code_service"`
		} `mapstructure:"two_factor_auth" yaml:"two_factor_auth" json:"two_factor_auth"`
	} `mapstructure:"security" yaml:"security" json:"security"`

	// 应用配置
	Application struct {
		Name        string `mapstructure:"name" yaml:"name" json:"name"`
		LocalDomain string `mapstructure:"local_domain" yaml:"local_domain" json:"local_domain"`
		BaseURL     string `mapstructure:"base_url" yaml:"base_url" json:"base_url"`
		Environment string `mapstructure:"environment" yaml:"environment" json:"environment"` // dev, test, prod
		Debug       bool   `mapstructure:"debug" yaml:"debug" json:"debug"`
		Version     string `mapstructure:"version" yaml:"version" json:"version"`
	} `mapstructure:"application" yaml:"application" json:"application"`

	// 日志配置
	Logging struct {
		Enable         bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
		Level          string `mapstructure:"level" yaml:"level" json:"level"`
		Format         string `mapstructure:"format" yaml:"format" json:"format"`
		LogFile        string `mapstructure:"log_file" yaml:"log_file" json:"log_file"`
		LogLoginEvents bool   `mapstructure:"log_login_events" yaml:"log_login_events" json:"log_login_events"`
		LogAuthEvents  bool   `mapstructure:"log_auth_events" yaml:"log_auth_events" json:"log_auth_events"`
		LogErrorEvents bool   `mapstructure:"log_error_events" yaml:"log_error_events" json:"log_error_events"`
	} `mapstructure:"logging" yaml:"logging" json:"logging"`
}

// OAuthProvider OAuth提供商配置
type OAuthProvider struct {
	ClientID     string   `mapstructure:"client_id" yaml:"client_id" json:"client_id"`
	ClientSecret string   `mapstructure:"client_secret" yaml:"client_secret" json:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url" yaml:"redirect_url" json:"redirect_url"`
	Scopes       []string `mapstructure:"scopes" yaml:"scopes" json:"scopes"`
	AuthURL      string   `mapstructure:"auth_url" yaml:"auth_url" json:"auth_url"`
	TokenURL     string   `mapstructure:"token_url" yaml:"token_url" json:"token_url"`
	UserInfoURL  string   `mapstructure:"user_info_url" yaml:"user_info_url" json:"user_info_url"`
	Enable       bool     `mapstructure:"enable" yaml:"enable" json:"enable"`
}

// GetConfigName 实现 ConfigInterface 接口
func (c AuthConfig) GetConfigName() string {
	return AuthConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c AuthConfig) SetDefaults(v *viper.Viper) {
	// CAS默认配置
	v.SetDefault("cas.enabled", false)
	v.SetDefault("cas.host", "https://cas.example.com")
	v.SetDefault("cas.login_path", "/cas/login")
	v.SetDefault("cas.logout_path", "/cas/logout")
	v.SetDefault("cas.validate_path", "/cas/validate")
	v.SetDefault("cas.service_validate_path", "/cas/serviceValidate")
	v.SetDefault("cas.version", "3.0")
	v.SetDefault("cas.sso_enabled", true)
	v.SetDefault("cas.renew_enabled", false)
	v.SetDefault("cas.gateway_enabled", false)

	// 登录路径默认配置
	v.SetDefault("login_paths.school.path", "/login/school")
	v.SetDefault("login_paths.school.cas_path", "/cas/login/school")
	v.SetDefault("login_paths.admin.path", "/login/admin")
	v.SetDefault("login_paths.admin.cas_path", "/cas/login/admin")
	v.SetDefault("login_paths.common.login_uri", "/login")
	v.SetDefault("login_paths.common.login_mobile_uri", "/mobile/login")
	v.SetDefault("login_paths.common.logout_uri", "/logout")

	// JWT默认配置
	v.SetDefault("jwt.secret", "your-jwt-secret-key-change-me")
	v.SetDefault("jwt.token_ttl", 24)
	v.SetDefault("jwt.refresh_ttl", 168)
	v.SetDefault("jwt.issuer", "YYHertz")
	v.SetDefault("jwt.signing_method", "HS256")

	// Session默认配置
	v.SetDefault("session.name", "YYHERTZ_SESSION")
	v.SetDefault("session.secret", "your-session-secret-change-me")
	v.SetDefault("session.max_age", 3600)
	v.SetDefault("session.path", "/")
	v.SetDefault("session.domain", "")
	v.SetDefault("session.secure", false)
	v.SetDefault("session.http_only", true)
	v.SetDefault("session.same_site", "Lax")
	v.SetDefault("session.store", "cookie")

	// OAuth默认配置
	v.SetDefault("oauth.providers", map[string]any{})

	// 权限默认配置
	v.SetDefault("authorization.enable", true)
	v.SetDefault("authorization.default_role", "user")
	v.SetDefault("authorization.admin_roles", []string{"admin", "superadmin"})
	v.SetDefault("authorization.guest_paths", []string{"/", "/login", "/register", "/forgot-password"})
	v.SetDefault("authorization.public_paths", []string{"/public", "/assets", "/static"})
	v.SetDefault("authorization.protected_paths", []string{"/admin", "/dashboard", "/profile"})
	v.SetDefault("authorization.role_based_access", true)

	// 安全默认配置
	v.SetDefault("security.password_policy.min_length", 8)
	v.SetDefault("security.password_policy.require_uppercase", true)
	v.SetDefault("security.password_policy.require_lowercase", true)
	v.SetDefault("security.password_policy.require_numbers", true)
	v.SetDefault("security.password_policy.require_symbols", false)
	v.SetDefault("security.password_policy.max_age", 90)

	v.SetDefault("security.login_attempts.max_attempts", 5)
	v.SetDefault("security.login_attempts.lockout_time", 15)
	v.SetDefault("security.login_attempts.reset_time", 60)

	v.SetDefault("security.two_factor_auth.enable", false)
	v.SetDefault("security.two_factor_auth.issuer", "YYHertz")
	v.SetDefault("security.two_factor_auth.secret_length", 32)
	v.SetDefault("security.two_factor_auth.qr_code_service", "google-charts")

	// 应用默认配置
	v.SetDefault("application.name", "YYHertz Auth")
	v.SetDefault("application.local_domain", "localhost")
	v.SetDefault("application.base_url", "http://localhost:8888")
	v.SetDefault("application.environment", "development")
	v.SetDefault("application.debug", true)
	v.SetDefault("application.version", "1.0.0")

	// 日志默认配置
	v.SetDefault("logging.enable", true)
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.log_file", "./logs/auth.log")
	v.SetDefault("logging.log_login_events", true)
	v.SetDefault("logging.log_auth_events", true)
	v.SetDefault("logging.log_error_events", true)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c AuthConfig) GenerateDefaultContent() string {
	return `# YYHertz Authentication Configuration
# 认证系统配置文件

# CAS认证配置
cas:
  enabled: false                              # 是否启用CAS认证
  host: "https://cas.example.com"             # CAS服务器地址
  login_path: "/cas/login"                    # CAS登录路径
  logout_path: "/cas/logout"                  # CAS登出路径
  validate_path: "/cas/validate"              # CAS验证路径
  service_validate_path: "/cas/serviceValidate" # CAS服务验证路径
  ticket_validation_url: ""                   # Ticket验证URL
  service_url: ""                            # 服务URL
  version: "3.0"                             # CAS版本: 1.0, 2.0, 3.0
  sso_enabled: true                          # 启用单点登录
  proxy_receptor_url: ""                     # 代理接收URL
  renew_enabled: false                       # 强制重新认证
  gateway_enabled: false                     # 网关模式

# 登录路径配置
login_paths:
  school:
    path: "/login/school"                    # 学校登录路径
    cas_path: "/cas/login/school"            # 学校CAS登录路径
  admin:
    path: "/login/admin"                     # 管理员登录路径
    cas_path: "/cas/login/admin"             # 管理员CAS登录路径
  common:
    login_uri: "/login"                      # 通用登录URI
    login_mobile_uri: "/mobile/login"        # 移动端登录URI
    logout_uri: "/logout"                    # 登出URI

# JWT配置
jwt:
  secret: "your-jwt-secret-key-change-me"    # JWT密钥
  token_ttl: 24                              # Token生存时间(小时)
  refresh_ttl: 168                           # 刷新Token生存时间(小时)
  issuer: "YYHertz"                          # 签发者
  signing_method: "HS256"                    # 签名方法: HS256, RS256
  public_key_path: ""                        # 公钥文件路径(RS256使用)
  private_key_path: ""                       # 私钥文件路径(RS256使用)

# Session配置
session:
  name: "YYHERTZ_SESSION"                    # Session名称
  secret: "your-session-secret-change-me"    # Session密钥
  max_age: 3600                              # Session最大生存时间(秒)
  path: "/"                                  # Session路径
  domain: ""                                 # Session域名
  secure: false                              # 是否只能通过HTTPS传输
  http_only: true                            # 是否HttpOnly
  same_site: "Lax"                           # SameSite策略: Strict, Lax, None
  store: "cookie"                            # 存储方式: cookie, memory, redis
  key_pairs: []                              # 加密密钥对

# OAuth配置
oauth:
  providers:
    github:
      client_id: ""
      client_secret: ""
      redirect_url: "/auth/github/callback"
      scopes: ["user:email"]
      auth_url: "https://github.com/login/oauth/authorize"
      token_url: "https://github.com/login/oauth/access_token"
      user_info_url: "https://api.github.com/user"
      enable: false
    google:
      client_id: ""
      client_secret: ""
      redirect_url: "/auth/google/callback"
      scopes: ["openid", "profile", "email"]
      auth_url: "https://accounts.google.com/o/oauth2/auth"
      token_url: "https://oauth2.googleapis.com/token"
      user_info_url: "https://www.googleapis.com/oauth2/v2/userinfo"
      enable: false

# 权限配置
authorization:
  enable: true                               # 启用权限控制
  default_role: "user"                       # 默认角色
  admin_roles: ["admin", "superadmin"]       # 管理员角色列表
  guest_paths: ["/", "/login", "/register", "/forgot-password"] # 游客可访问路径
  public_paths: ["/public", "/assets", "/static"] # 公开路径
  protected_paths: ["/admin", "/dashboard", "/profile"] # 受保护路径
  role_based_access: true                    # 基于角色的访问控制

# 安全配置
security:
  # 密码策略
  password_policy:
    min_length: 8                            # 最小长度
    require_uppercase: true                  # 需要大写字母
    require_lowercase: true                  # 需要小写字母
    require_numbers: true                    # 需要数字
    require_symbols: false                   # 需要特殊符号
    max_age: 90                              # 密码最大使用天数
  
  # 登录尝试限制
  login_attempts:
    max_attempts: 5                          # 最大尝试次数
    lockout_time: 15                         # 锁定时间(分钟)
    reset_time: 60                           # 重置时间(分钟)
  
  # 双因子认证
  two_factor_auth:
    enable: false                            # 启用2FA
    issuer: "YYHertz"                        # 签发者名称
    secret_length: 32                        # 密钥长度
    qr_code_service: "google-charts"         # 二维码服务

# 应用配置
application:
  name: "YYHertz Auth"                       # 应用名称
  local_domain: "localhost"                  # 本地域名
  base_url: "http://localhost:8888"          # 基础URL
  environment: "development"                 # 环境: development, testing, production
  debug: true                                # 调试模式
  version: "1.0.0"                           # 版本号

# 日志配置
logging:
  enable: true                               # 启用日志
  level: "info"                              # 日志级别: debug, info, warn, error
  format: "json"                             # 日志格式: json, text
  log_file: "./logs/auth.log"                # 日志文件路径
  log_login_events: true                     # 记录登录事件
  log_auth_events: true                      # 记录认证事件
  log_error_events: true                     # 记录错误事件
`
}
