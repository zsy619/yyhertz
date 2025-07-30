package middleware

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/constant"
	"github.com/zsy619/yyhertz/framework/response"
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

// DefaultTLSConfig 默认TLS配置
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

// TLSSupportMiddleware TLS支持中间件
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
				// 返回错误
				config.Warn("HTTP请求被拒绝，要求HTTPS")
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

// isHTTPSRequest 检查是否为HTTPS请求
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

// redirectToHTTPS 重定向到HTTPS
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

// setSecurityHeaders 设置安全头
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

// GetTLSConfigFromCertFiles 从证书文件创建TLS配置
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

// ValidateTLSConfig 验证TLS配置
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
