package middleware

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/mvc/response"
)

// TLSConfig TLS配置结构
type TLSConfig struct {
	// 基础配置
	Enable     bool   `json:"enable" yaml:"enable"`           // 是否启用TLS
	CertFile   string `json:"cert_file" yaml:"cert_file"`     // 证书文件路径
	KeyFile    string `json:"key_file" yaml:"key_file"`       // 私钥文件路径
	MinVersion uint16 `json:"min_version" yaml:"min_version"` // 最小TLS版本
	MaxVersion uint16 `json:"max_version" yaml:"max_version"` // 最大TLS版本

	// 安全配置
	RequireHTTPS   bool `json:"require_https" yaml:"require_https"`     // 是否强制HTTPS
	HSTSEnabled    bool `json:"hsts_enabled" yaml:"hsts_enabled"`       // 是否启用HSTS
	HSTSMaxAge     int  `json:"hsts_max_age" yaml:"hsts_max_age"`       // HSTS最大年龄（秒）
	HSTSSubdomains bool `json:"hsts_subdomains" yaml:"hsts_subdomains"` // HSTS是否包含子域名

	// 密码套件配置
	CipherSuites []uint16 `json:"cipher_suites" yaml:"cipher_suites"` // 支持的密码套件
	PreferServer bool     `json:"prefer_server" yaml:"prefer_server"` // 是否优先服务器密码套件

	// 客户端证书配置
	ClientAuth   tls.ClientAuthType `json:"client_auth" yaml:"client_auth"`       // 客户端认证模式
	ClientCAFile string             `json:"client_ca_file" yaml:"client_ca_file"` // 客户端CA证书文件

	// 重定向配置
	HTTPSRedirect bool `json:"https_redirect" yaml:"https_redirect"` // HTTP是否重定向到HTTPS
	RedirectPort  int  `json:"redirect_port" yaml:"redirect_port"`   // HTTPS重定向端口
}

// DefaultTLSConfig 返回一个默认的 TLS 配置。
// 该配置包含以下默认值：
//   - Enable: false，默认禁用 TLS。
//   - MinVersion: TLS 1.2，最低支持的 TLS 版本。
//   - MaxVersion: TLS 1.3，最高支持的 TLS 版本。
//   - RequireHTTPS: false，不强制要求 HTTPS。
//   - HSTSEnabled: true，启用 HSTS（HTTP Strict Transport Security）。
//   - HSTSMaxAge: 31536000（1年），HSTS 的最大有效期。
//   - HSTSSubdomains: true，HSTS 包含子域名。
//   - CipherSuites: 支持的加密套件列表，优先使用 ECDHE 和 AES-GCM 算法。
//   - PreferServer: true，优先使用服务器的加密套件偏好。
//   - ClientAuth: NoClientCert，不要求客户端证书。
//   - HTTPSRedirect: true，启用 HTTPS 重定向。
//   - RedirectPort: 443，重定向的目标端口。
//
// 该函数通常用于初始化 TLS 配置，适用于需要安全通信的场景。
func DefaultTLSConfig() *TLSConfig {
	return &TLSConfig{
		Enable:         false,
		MinVersion:     tls.VersionTLS12,
		MaxVersion:     tls.VersionTLS13,
		RequireHTTPS:   false,
		HSTSEnabled:    true,
		HSTSMaxAge:     31536000, // 1年
		HSTSSubdomains: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServer:  true,
		ClientAuth:    tls.NoClientCert,
		HTTPSRedirect: true,
		RedirectPort:  443,
	}
}

// TLSSupportMiddleware 是一个中间件函数，用于处理与TLS（传输层安全）相关的逻辑。
// 它根据提供的TLS配置（cfg）来决定是否强制使用HTTPS、重定向HTTP请求到HTTPS，以及设置安全头。
//
// 参数:
//
//	cfg *TLSConfig: TLS配置对象。如果为nil，将使用默认配置。
//
// 返回值:
//
//	app.HandlerFunc: 返回一个中间件处理函数，该函数会在每个请求中执行TLS相关的逻辑。
//
// 功能说明:
//   - 记录TLS中间件的处理开始和结束。
//   - 检查请求是否为HTTPS。
//   - 如果启用了强制HTTPS且请求为HTTP，根据配置决定是否重定向到HTTPS或返回错误。
//   - 如果是HTTPS请求，设置安全头。
//   - 继续处理后续中间件或请求。
//
// 注意:
//   - 如果启用了强制HTTPS但未启用重定向，会返回400错误。
//   - 安全头的设置仅在HTTPS请求中生效。
func TLSSupportMiddleware(cfg *TLSConfig) app.HandlerFunc {
	if cfg == nil {
		cfg = DefaultTLSConfig()
	}

	return func(ctx context.Context, c *app.RequestContext) {
		// 记录TLS中间件处理开始
		config.WithFields(map[string]any{
			"path":        string(c.Path()),
			"method":      string(c.Method()),
			"tls_enabled": cfg.Enable,
		}).Debug("TLS中间件处理开始")

		// 检查是否为HTTPS请求
		isHTTPS := isHTTPSRequest(c)

		// 如果启用了强制HTTPS且当前是HTTP请求
		if cfg.RequireHTTPS && !isHTTPS {
			if cfg.HTTPSRedirect {
				// 重定向到HTTPS
				redirectToHTTPS(c, cfg.RedirectPort)
				return
			} else {
				go func() {
					// 返回错误
					config.Warn("HTTP请求被拒绝，要求HTTPS")
				}()
				c.JSON(consts.StatusBadRequest, response.ErrorResponse{
					Code:    constant.CodeHTTPSRequired,
					Message: "HTTPS连接是必需的",
					Error:   "此服务要求使用安全连接",
				})
				c.Abort()
				return
			}
		}

		// 如果是HTTPS请求，设置安全头
		if isHTTPS {
			setSecurityHeaders(c, cfg)
		}

		// 继续处理请求
		c.Next(ctx)

		config.Debug("TLS中间件处理完成")
	}
}

// isHTTPSRequest 检查当前请求是否为HTTPS请求。
// 该函数通过以下方式验证：
// 1. 检查URI的scheme是否为"https"。
// 2. 检查请求头中的"X-Forwarded-Proto"是否为"https"（适用于代理场景）。
// 3. 检查请求头中的"X-Forwarded-SSL"是否为"on"。
// 4. 检查请求头中的"Front-End-Https"是否为"on"。
// 如果以上任一条件满足，则返回true，否则返回false。
// 参数:
//
//	c *app.RequestContext: 请求上下文对象。
//
// 返回值:
//
//	bool: 如果请求为HTTPS则返回true，否则返回false。
func isHTTPSRequest(c *app.RequestContext) bool {
	// 检查URI scheme
	if strings.ToLower(string(c.URI().Scheme())) == "https" {
		return true
	}

	// 检查X-Forwarded-Proto头（用于代理场景）
	proto := string(c.GetHeader("X-Forwarded-Proto"))
	if strings.ToLower(proto) == "https" {
		return true
	}

	// 检查X-Forwarded-SSL头
	ssl := string(c.GetHeader("X-Forwarded-SSL"))
	if strings.ToLower(ssl) == "on" {
		return true
	}

	// 检查Front-End-Https头
	frontEndHttps := string(c.GetHeader("Front-End-Https"))
	if strings.ToLower(frontEndHttps) == "on" {
		return true
	}

	return false
}

// redirectToHTTPS 将HTTP请求重定向到HTTPS。
//
// 参数:
//   - c: 请求上下文，包含当前请求的详细信息。
//   - httpsPort: HTTPS服务的端口号。
//
// 功能描述:
//  1. 从请求中提取主机名（移除端口号）。
//  2. 根据是否使用默认HTTPS端口（443）构建完整的HTTPS URL。
//  3. 记录重定向日志（原始URL和目标URL）。
//  4. 返回301状态码（永久重定向）到目标HTTPS URL。
//
// 注意:
//   - 此函数适用于需要强制HTTPS访问的场景。
//   - 如果httpsPort为443，URL中不会包含端口号。
func redirectToHTTPS(c *app.RequestContext, httpsPort int) {
	host := string(c.Host())

	// 移除端口号
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// 构建HTTPS URL
	var httpsURL string
	if httpsPort == 443 {
		httpsURL = "https://" + host + string(c.URI().RequestURI())
	} else {
		httpsURL = "https://" + host + ":" + string(rune(httpsPort)) + string(c.URI().RequestURI())
	}

	config.WithFields(map[string]any{
		"from": string(c.URI().String()),
		"to":   httpsURL,
	}).Info("重定向到HTTPS")

	c.Redirect(consts.StatusMovedPermanently, []byte(httpsURL))
}

// setSecurityHeaders 设置HTTP响应中的安全头信息，包括HSTS、CSP等。
//
// 参数:
//   - c: 请求上下文，用于设置响应头。
//   - cfg: TLS配置，包含HSTS相关的配置选项。
//
// 功能:
//   - 如果启用了HSTS，设置Strict-Transport-Security头，包括max-age、includeSubDomains和preload选项。
//   - 设置其他安全头，如X-Content-Type-Options、X-Frame-Options、X-XSS-Protection和Referrer-Policy。
//   - 设置内容安全策略（CSP）头，限制资源的加载来源和行为。
//
// 日志:
//   - 记录HSTS头的设置值。
//   - 记录安全头设置完成的信息。
func setSecurityHeaders(c *app.RequestContext, cfg *TLSConfig) {
	// 设置HSTS头
	if cfg.HSTSEnabled {
		hstsValue := "max-age=" + string(rune(cfg.HSTSMaxAge))
		if cfg.HSTSSubdomains {
			hstsValue += "; includeSubDomains"
		}
		hstsValue += "; preload"

		c.Response.Header.Set("Strict-Transport-Security", hstsValue)

		config.WithFields(map[string]any{
			"hsts_value": hstsValue,
		}).Debug("设置HSTS头")
	}

	// 设置其他安全头
	c.Response.Header.Set("X-Content-Type-Options", "nosniff")
	c.Response.Header.Set("X-Frame-Options", "DENY")
	c.Response.Header.Set("X-XSS-Protection", "1; mode=block")
	c.Response.Header.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	// 设置CSP头（内容安全策略）
	csp := "default-src 'self'; " +
		"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: https:; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"media-src 'self'; " +
		"object-src 'none'; " +
		"child-src 'self'; " +
		"frame-ancestors 'none'; " +
		"form-action 'self'; " +
		"upgrade-insecure-requests"

	c.Response.Header.Set("Content-Security-Policy", csp)

	config.Debug("设置安全响应头完成")
}

// GetTLSConfigFromCertFiles 从证书文件和密钥文件加载TLS配置。
//
// 参数:
//   - certFile: 证书文件路径。
//   - keyFile: 密钥文件路径。
//   - cfg: TLS配置参数，包括最小/最大TLS版本、密码套件等。
//
// 返回值:
//   - *tls.Config: 成功时返回TLS配置对象。
//   - error: 加载证书或配置失败时返回错误。
//
// 功能说明:
//  1. 加载X.509证书和密钥对。
//  2. 根据提供的TLS配置参数（如版本、密码套件等）生成TLS配置。
//  3. 记录加载成功或失败的日志信息。
//
// 注意:
//   - 如果证书或密钥文件加载失败，会记录错误日志并返回错误。
//   - 成功时，会记录TLS配置的详细信息。
func GetTLSConfigFromCertFiles(certFile, keyFile string, cfg *TLSConfig) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		config.WithFields(map[string]any{
			"cert_file": certFile,
			"key_file":  keyFile,
			"error":     err.Error(),
		}).Error("加载TLS证书失败")
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		MinVersion:               cfg.MinVersion,
		MaxVersion:               cfg.MaxVersion,
		CipherSuites:             cfg.CipherSuites,
		PreferServerCipherSuites: cfg.PreferServer,
		ClientAuth:               cfg.ClientAuth,
	}

	config.WithFields(map[string]any{
		"cert_file":   certFile,
		"min_version": cfg.MinVersion,
		"max_version": cfg.MaxVersion,
		"client_auth": cfg.ClientAuth,
	}).Info("TLS配置创建成功")

	return tlsConfig, nil
}

// ValidateTLSConfig 验证 TLS 配置的有效性。
//
// 参数:
//
//	cfg: 指向 TLSConfig 结构体的指针，包含 TLS 相关配置。
//
// 返回值:
//
//	error: 如果配置无效，返回包含具体错误信息的 ValidationError；
//	       如果配置有效或 cfg 为 nil，返回 nil。
//
// 验证规则:
//   - 如果 cfg 为 nil，直接返回 nil。
//   - 如果启用了 TLS (cfg.Enable 为 true)，则检查以下内容:
//   - CertFile 和 KeyFile 不能为空。
//   - MinVersion 不能大于 MaxVersion。
//   - HSTSMaxAge 不能为负数。
//   - RedirectPort 必须在 1-65535 范围内。
//   - 如果所有验证通过，记录日志并返回 nil。
//
// 示例:
//
//	err := ValidateTLSConfig(&TLSConfig{Enable: true, CertFile: "cert.pem", KeyFile: "key.pem"})
//	if err != nil {
//	    log.Fatal(err)
//	}
func ValidateTLSConfig(cfg *TLSConfig) error {
	if cfg == nil {
		return nil
	}

	if cfg.Enable {
		if cfg.CertFile == "" {
			return &response.ValidationError{
				Field:   "cert_file",
				Message: "TLS证书文件路径不能为空",
			}
		}

		if cfg.KeyFile == "" {
			return &response.ValidationError{
				Field:   "key_file",
				Message: "TLS私钥文件路径不能为空",
			}
		}

		if cfg.MinVersion > cfg.MaxVersion {
			return &response.ValidationError{
				Field:   "tls_version",
				Message: "最小TLS版本不能大于最大TLS版本",
			}
		}

		if cfg.HSTSMaxAge < 0 {
			return &response.ValidationError{
				Field:   "hsts_max_age",
				Message: "HSTS最大年龄不能为负数",
			}
		}

		if cfg.RedirectPort <= 0 || cfg.RedirectPort > 65535 {
			return &response.ValidationError{
				Field:   "redirect_port",
				Message: "重定向端口必须在1-65535范围内",
			}
		}
	}

	config.Info("TLS配置验证通过")
	return nil
}
