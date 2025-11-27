package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/spf13/viper"
)

// TLSManager TLS证书管理器
type TLSManager struct {
	config      *TLSServerConfig
	tlsConfig   *tls.Config
	certWatcher *CertWatcher
}

// TLSServerConfig TLS服务器配置
type TLSServerConfig struct {
	// 基本配置
	Basic struct {
		Enable      bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
		Environment string `mapstructure:"environment" yaml:"environment" json:"environment"` // dev, test, prod
		Debug       bool   `mapstructure:"debug" yaml:"debug" json:"debug"`
		LogLevel    string `mapstructure:"log_level" yaml:"log_level" json:"log_level"`
	} `mapstructure:"basic" yaml:"basic" json:"basic"`

	// 证书配置
	Certificate struct {
		CertFile       string `mapstructure:"cert_file" yaml:"cert_file" json:"cert_file"`
		KeyFile        string `mapstructure:"key_file" yaml:"key_file" json:"key_file"`
		CAFile         string `mapstructure:"ca_file" yaml:"ca_file" json:"ca_file"`
		CertData       string `mapstructure:"cert_data" yaml:"cert_data" json:"cert_data"`                   // PEM编码的证书内容
		KeyData        string `mapstructure:"key_data" yaml:"key_data" json:"key_data"`                      // PEM编码的私钥内容
		PassPhrase     string `mapstructure:"pass_phrase" yaml:"pass_phrase" json:"pass_phrase"`             // 私钥密码
		ValidityPeriod int    `mapstructure:"validity_period" yaml:"validity_period" json:"validity_period"` // 证书有效期检查(天)
	} `mapstructure:"certificate" yaml:"certificate" json:"certificate"`

	// TLS版本配置
	Version struct {
		MinVersion        string   `mapstructure:"min_version" yaml:"min_version" json:"min_version"` // "1.0", "1.1", "1.2", "1.3"
		MaxVersion        string   `mapstructure:"max_version" yaml:"max_version" json:"max_version"` // "1.0", "1.1", "1.2", "1.3"
		SupportedVersions []string `mapstructure:"supported_versions" yaml:"supported_versions" json:"supported_versions"`
	} `mapstructure:"version" yaml:"version" json:"version"`

	// 密码套件配置
	Cipher struct {
		Suites        []string `mapstructure:"suites" yaml:"suites" json:"suites"`
		PreferServer  bool     `mapstructure:"prefer_server" yaml:"prefer_server" json:"prefer_server"`
		Curves        []string `mapstructure:"curves" yaml:"curves" json:"curves"`                         // 支持的椭圆曲线
		SignatureAlgs []string `mapstructure:"signature_algs" yaml:"signature_algs" json:"signature_algs"` // 签名算法
	} `mapstructure:"cipher" yaml:"cipher" json:"cipher"`

	// 客户端认证配置
	ClientAuth struct {
		Mode        string   `mapstructure:"mode" yaml:"mode" json:"mode"` // "NoClientCert", "RequestClientCert", etc.
		CAFile      string   `mapstructure:"ca_file" yaml:"ca_file" json:"ca_file"`
		CAData      string   `mapstructure:"ca_data" yaml:"ca_data" json:"ca_data"`                // PEM编码的CA证书内容
		CRLFile     string   `mapstructure:"crl_file" yaml:"crl_file" json:"crl_file"`             // 证书撤销列表
		VerifyDepth int      `mapstructure:"verify_depth" yaml:"verify_depth" json:"verify_depth"` // 验证深度
		AllowedCNs  []string `mapstructure:"allowed_cns" yaml:"allowed_cns" json:"allowed_cns"`    // 允许的CN列表
	} `mapstructure:"client_auth" yaml:"client_auth" json:"client_auth"`

	// 证书自动管理
	AutoManagement struct {
		Enable         bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
		ReloadInterval int    `mapstructure:"reload_interval" yaml:"reload_interval" json:"reload_interval"` // 秒
		WatchFiles     bool   `mapstructure:"watch_files" yaml:"watch_files" json:"watch_files"`
		BackupEnabled  bool   `mapstructure:"backup_enabled" yaml:"backup_enabled" json:"backup_enabled"`
		BackupDir      string `mapstructure:"backup_dir" yaml:"backup_dir" json:"backup_dir"`
		HealthCheck    struct {
			Enable   bool `mapstructure:"enable" yaml:"enable" json:"enable"`
			Interval int  `mapstructure:"interval" yaml:"interval" json:"interval"` // 秒
			Timeout  int  `mapstructure:"timeout" yaml:"timeout" json:"timeout"`    // 秒
		} `mapstructure:"health_check" yaml:"health_check" json:"health_check"`
	} `mapstructure:"auto_management" yaml:"auto_management" json:"auto_management"`

	// ALPN协议配置
	ALPN struct {
		NextProtos   []string `mapstructure:"next_protos" yaml:"next_protos" json:"next_protos"`
		H2Enabled    bool     `mapstructure:"h2_enabled" yaml:"h2_enabled" json:"h2_enabled"`
		H2CEnabled   bool     `mapstructure:"h2c_enabled" yaml:"h2c_enabled" json:"h2c_enabled"`
		HTTP1Enabled bool     `mapstructure:"http1_enabled" yaml:"http1_enabled" json:"http1_enabled"`
	} `mapstructure:"alpn" yaml:"alpn" json:"alpn"`

	// 会话管理
	Session struct {
		TicketsEnabled    bool   `mapstructure:"tickets_enabled" yaml:"tickets_enabled" json:"tickets_enabled"`
		TicketKey         string `mapstructure:"ticket_key" yaml:"ticket_key" json:"ticket_key"`
		TicketKeyRotation bool   `mapstructure:"ticket_key_rotation" yaml:"ticket_key_rotation" json:"ticket_key_rotation"`
		TicketLifetime    int    `mapstructure:"ticket_lifetime" yaml:"ticket_lifetime" json:"ticket_lifetime"` // 秒
		CacheSize         int    `mapstructure:"cache_size" yaml:"cache_size" json:"cache_size"`
		CacheTTL          int    `mapstructure:"cache_ttl" yaml:"cache_ttl" json:"cache_ttl"` // 秒
	} `mapstructure:"session" yaml:"session" json:"session"`

	// OCSP配置
	OCSP struct {
		Enable       bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
		StaplingFile string `mapstructure:"stapling_file" yaml:"stapling_file" json:"stapling_file"`
		ResponderURL string `mapstructure:"responder_url" yaml:"responder_url" json:"responder_url"`
		CacheTime    int    `mapstructure:"cache_time" yaml:"cache_time" json:"cache_time"` // 秒
	} `mapstructure:"ocsp" yaml:"ocsp" json:"ocsp"`

	// HSTS配置
	HSTS struct {
		Enable            bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
		MaxAge            int    `mapstructure:"max_age" yaml:"max_age" json:"max_age"` // 秒
		IncludeSubDomains bool   `mapstructure:"include_subdomains" yaml:"include_subdomains" json:"include_subdomains"`
		Preload           bool   `mapstructure:"preload" yaml:"preload" json:"preload"`
		Header            string `mapstructure:"header" yaml:"header" json:"header"`
	} `mapstructure:"hsts" yaml:"hsts" json:"hsts"`

	// 性能优化
	Performance struct {
		ReadTimeout     int  `mapstructure:"read_timeout" yaml:"read_timeout" json:"read_timeout"`    // 秒
		WriteTimeout    int  `mapstructure:"write_timeout" yaml:"write_timeout" json:"write_timeout"` // 秒
		IdleTimeout     int  `mapstructure:"idle_timeout" yaml:"idle_timeout" json:"idle_timeout"`    // 秒
		MaxHeaderBytes  int  `mapstructure:"max_header_bytes" yaml:"max_header_bytes" json:"max_header_bytes"`
		KeepAlive       bool `mapstructure:"keep_alive" yaml:"keep_alive" json:"keep_alive"`
		TCP_NODELAY     bool `mapstructure:"tcp_nodelay" yaml:"tcp_nodelay" json:"tcp_nodelay"`
		ReusePort       bool `mapstructure:"reuse_port" yaml:"reuse_port" json:"reuse_port"`
		ReadBufferSize  int  `mapstructure:"read_buffer_size" yaml:"read_buffer_size" json:"read_buffer_size"`
		WriteBufferSize int  `mapstructure:"write_buffer_size" yaml:"write_buffer_size" json:"write_buffer_size"`
	} `mapstructure:"performance" yaml:"performance" json:"performance"`

	// 监控配置
	Monitoring struct {
		Enable         bool   `mapstructure:"enable" yaml:"enable" json:"enable"`
		MetricsPath    string `mapstructure:"metrics_path" yaml:"metrics_path" json:"metrics_path"`
		LogConnections bool   `mapstructure:"log_connections" yaml:"log_connections" json:"log_connections"`
		LogHandshakes  bool   `mapstructure:"log_handshakes" yaml:"log_handshakes" json:"log_handshakes"`
		LogErrors      bool   `mapstructure:"log_errors" yaml:"log_errors" json:"log_errors"`
		StatsInterval  int    `mapstructure:"stats_interval" yaml:"stats_interval" json:"stats_interval"` // 秒
	} `mapstructure:"monitoring" yaml:"monitoring" json:"monitoring"`
}

// CertWatcher 证书监视器
type CertWatcher struct {
	certFile   string
	keyFile    string
	reloadChan chan struct{}
	stopChan   chan struct{}
	interval   time.Duration
}

// DefaultTLSServerConfig 默认TLS服务器配置
func DefaultTLSServerConfig() *TLSServerConfig {
	config := &TLSServerConfig{}

	// 基本配置
	config.Basic.Enable = false
	config.Basic.Environment = "development"
	config.Basic.Debug = true
	config.Basic.LogLevel = "info"

	// 证书配置
	config.Certificate.CertFile = ""
	config.Certificate.KeyFile = ""
	config.Certificate.ValidityPeriod = 30

	// TLS版本配置
	config.Version.MinVersion = "1.2"
	config.Version.MaxVersion = "1.3"
	config.Version.SupportedVersions = []string{"1.2", "1.3"}

	// 密码套件配置
	config.Cipher.PreferServer = true
	config.Cipher.Suites = []string{
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_RSA_WITH_AES_128_GCM_SHA256",
	}

	// 客户端认证配置
	config.ClientAuth.Mode = "NoClientCert"
	config.ClientAuth.VerifyDepth = 1

	// 自动管理配置
	config.AutoManagement.Enable = false
	config.AutoManagement.ReloadInterval = 300
	config.AutoManagement.WatchFiles = false
	config.AutoManagement.BackupEnabled = false
	config.AutoManagement.HealthCheck.Enable = false
	config.AutoManagement.HealthCheck.Interval = 60
	config.AutoManagement.HealthCheck.Timeout = 10

	// ALPN配置
	config.ALPN.NextProtos = []string{"h2", "http/1.1"}
	config.ALPN.H2Enabled = true
	config.ALPN.H2CEnabled = false
	config.ALPN.HTTP1Enabled = true

	// 会话配置
	config.Session.TicketsEnabled = true
	config.Session.TicketKeyRotation = true
	config.Session.TicketLifetime = 3600
	config.Session.CacheSize = 1000
	config.Session.CacheTTL = 300

	return config
}

// NewTLSManager 创建TLS管理器
func NewTLSManager(config *TLSServerConfig) (*TLSManager, error) {
	if config == nil {
		config = DefaultTLSServerConfig()
	}

	manager := &TLSManager{
		config: config,
	}

	if config.Basic.Enable {
		if err := manager.loadTLSConfig(); err != nil {
			return nil, fmt.Errorf("加载TLS配置失败: %w", err)
		}

		if config.AutoManagement.Enable {
			manager.startCertWatcher()
		}
	}

	log.Printf("TLS管理器初始化完成: enabled=%v, auto_reload=%v, min_version=%s, max_version=%s",
		config.Basic.Enable,
		config.AutoManagement.Enable,
		config.Version.MinVersion,
		config.Version.MaxVersion)

	return manager, nil
}

// loadTLSConfig 加载TLS配置
func (m *TLSManager) loadTLSConfig() error {
	// 加载证书
	cert, err := tls.LoadX509KeyPair(m.config.Certificate.CertFile, m.config.Certificate.KeyFile)
	if err != nil {
		return fmt.Errorf("加载证书失败: %w", err)
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: m.config.Cipher.PreferServer,
		NextProtos:               m.config.ALPN.NextProtos,
		SessionTicketsDisabled:   !m.config.Session.TicketsEnabled,
	}

	// 设置TLS版本
	if minVer, err := parseTLSVersion(m.config.Version.MinVersion); err != nil {
		return fmt.Errorf("解析最小TLS版本失败: %w", err)
	} else {
		tlsConfig.MinVersion = minVer
	}

	if maxVer, err := parseTLSVersion(m.config.Version.MaxVersion); err != nil {
		return fmt.Errorf("解析最大TLS版本失败: %w", err)
	} else {
		tlsConfig.MaxVersion = maxVer
	}

	// 设置密码套件
	if cipherSuites, err := parseCipherSuites(m.config.Cipher.Suites); err != nil {
		return fmt.Errorf("解析密码套件失败: %w", err)
	} else {
		tlsConfig.CipherSuites = cipherSuites
	}

	// 设置客户端认证
	if clientAuth, err := parseClientAuth(m.config.ClientAuth.Mode); err != nil {
		return fmt.Errorf("解析客户端认证模式失败: %w", err)
	} else {
		tlsConfig.ClientAuth = clientAuth
	}

	// 加载客户端CA证书
	if m.config.ClientAuth.CAFile != "" {
		clientCAs, err := loadCACerts(m.config.ClientAuth.CAFile)
		if err != nil {
			return fmt.Errorf("加载客户端CA证书失败: %w", err)
		}
		tlsConfig.ClientCAs = clientCAs
	}

	// 设置会话票据密钥
	if m.config.Session.TicketKey != "" {
		key := []byte(m.config.Session.TicketKey)
		if len(key) != 32 {
			return fmt.Errorf("会话票据密钥长度必须为32字节")
		}
		tlsConfig.SetSessionTicketKeys([][32]byte{[32]byte(key)})
	}

	m.tlsConfig = tlsConfig

	log.Printf("TLS配置加载成功: cert_file=%s, key_file=%s, client_auth=%s, cipher_count=%d",
		m.config.Certificate.CertFile,
		m.config.Certificate.KeyFile,
		m.config.ClientAuth.Mode,
		len(tlsConfig.CipherSuites))

	return nil
}

// GetTLSConfig 获取TLS配置
func (m *TLSManager) GetTLSConfig() *tls.Config {
	return m.tlsConfig
}

// startCertWatcher 启动证书监视器
func (m *TLSManager) startCertWatcher() {
	if m.certWatcher != nil {
		return
	}

	m.certWatcher = &CertWatcher{
		certFile:   m.config.Certificate.CertFile,
		keyFile:    m.config.Certificate.KeyFile,
		reloadChan: make(chan struct{}, 1),
		stopChan:   make(chan struct{}),
		interval:   time.Duration(m.config.AutoManagement.ReloadInterval) * time.Second,
	}

	go m.certWatcher.watch(m)

	log.Printf("证书监视器启动: interval=%d秒", m.config.AutoManagement.ReloadInterval)
}

// watch 监视证书文件变化
func (w *CertWatcher) watch(manager *TLSManager) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	var lastModTime time.Time

	for {
		select {
		case <-ticker.C:
			// 检查证书文件是否存在
			if _, err := ioutil.ReadFile(w.certFile); err != nil {
				GetGlobalLogger().WithFields(map[string]any{
					"cert_file": w.certFile,
					"error":     err.Error(),
				}).Error("读取证书文件失败")
				continue
			}

			if _, err := ioutil.ReadFile(w.keyFile); err != nil {
				GetGlobalLogger().WithFields(map[string]any{
					"key_file": w.keyFile,
					"error":    err.Error(),
				}).Error("读取密钥文件失败")
				continue
			}

			// 简单的时间间隔检测（实际应该检查文件修改时间）
			currentTime := time.Now()

			if !lastModTime.IsZero() && currentTime.Sub(lastModTime) > time.Duration(w.interval) {
				// 重新加载证书
				if err := manager.loadTLSConfig(); err != nil {
					GetGlobalLogger().WithFields(map[string]any{
						"error": err.Error(),
					}).Error("重新加载TLS配置失败")
				} else {
					GetGlobalLogger().Info("证书自动重载成功")
					select {
					case w.reloadChan <- struct{}{}:
					default:
					}
				}
			}

			lastModTime = currentTime

		case <-w.stopChan:
			return
		}
	}
}

// Stop 停止证书监视器
func (m *TLSManager) Stop() {
	if m.certWatcher != nil {
		close(m.certWatcher.stopChan)
		m.certWatcher = nil
		GetGlobalLogger().Info("证书监视器已停止")
	}
}

// parseTLSVersion 解析TLS版本
func parseTLSVersion(version string) (uint16, error) {
	switch version {
	case "1.0":
		return tls.VersionTLS10, nil
	case "1.1":
		return tls.VersionTLS11, nil
	case "1.2":
		return tls.VersionTLS12, nil
	case "1.3":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("不支持的TLS版本: %s", version)
	}
}

// parseClientAuth 解析客户端认证模式
func parseClientAuth(auth string) (tls.ClientAuthType, error) {
	switch auth {
	case "NoClientCert":
		return tls.NoClientCert, nil
	case "RequestClientCert":
		return tls.RequestClientCert, nil
	case "RequireAnyClientCert":
		return tls.RequireAnyClientCert, nil
	case "VerifyClientCertIfGiven":
		return tls.VerifyClientCertIfGiven, nil
	case "RequireAndVerifyClientCert":
		return tls.RequireAndVerifyClientCert, nil
	default:
		return tls.NoClientCert, fmt.Errorf("不支持的客户端认证模式: %s", auth)
	}
}

// parseCipherSuites 解析密码套件
func parseCipherSuites(suites []string) ([]uint16, error) {
	var result []uint16

	cipherMap := map[string]uint16{
		"TLS_RSA_WITH_RC4_128_SHA":                      tls.TLS_RSA_WITH_RC4_128_SHA,
		"TLS_RSA_WITH_3DES_EDE_CBC_SHA":                 tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
		"TLS_RSA_WITH_AES_128_CBC_SHA":                  tls.TLS_RSA_WITH_AES_128_CBC_SHA,
		"TLS_RSA_WITH_AES_256_CBC_SHA":                  tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		"TLS_RSA_WITH_AES_128_CBC_SHA256":               tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		"TLS_RSA_WITH_AES_128_GCM_SHA256":               tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_RSA_WITH_AES_256_GCM_SHA384":               tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_RC4_128_SHA":              tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA":          tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA":          tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_RC4_128_SHA":                tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
		"TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA":           tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA":            tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
		"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA":            tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
		"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256":       tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256":         tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":         tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":       tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":         tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":       tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	}

	for _, suite := range suites {
		if cipherID, ok := cipherMap[suite]; ok {
			result = append(result, cipherID)
		} else {
			return nil, fmt.Errorf("不支持的密码套件: %s", suite)
		}
	}

	return result, nil
}

// loadCACerts 加载CA证书
func loadCACerts(caFile string) (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("解析CA证书失败")
	}

	return caCertPool, nil
}

// GetConfigName 实现 ConfigInterface 接口
func (c TLSServerConfig) GetConfigName() string {
	return TLSConfigName
}

// SetDefaults 实现 ConfigInterface 接口 - 设置默认值
func (c TLSServerConfig) SetDefaults(v *viper.Viper) {
	// 基本配置默认值
	v.SetDefault("basic.enable", false)
	v.SetDefault("basic.environment", "development")
	v.SetDefault("basic.debug", true)
	v.SetDefault("basic.log_level", "info")

	// 证书配置默认值
	v.SetDefault("certificate.cert_file", "./certs/server.crt")
	v.SetDefault("certificate.key_file", "./certs/server.key")
	v.SetDefault("certificate.ca_file", "")
	v.SetDefault("certificate.cert_data", "")
	v.SetDefault("certificate.key_data", "")
	v.SetDefault("certificate.pass_phrase", "")
	v.SetDefault("certificate.validity_period", 30)

	// TLS版本配置默认值
	v.SetDefault("version.min_version", "1.2")
	v.SetDefault("version.max_version", "1.3")
	v.SetDefault("version.supported_versions", []string{"1.2", "1.3"})

	// 密码套件配置默认值
	v.SetDefault("cipher.suites", []string{
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		"TLS_RSA_WITH_AES_256_GCM_SHA384",
		"TLS_RSA_WITH_AES_128_GCM_SHA256",
	})
	v.SetDefault("cipher.prefer_server", true)
	v.SetDefault("cipher.curves", []string{"P-256", "P-384", "P-521"})
	v.SetDefault("cipher.signature_algs", []string{"RSA-PSS-SHA256", "ECDSA-SHA256"})

	// 客户端认证配置默认值
	v.SetDefault("client_auth.mode", "NoClientCert")
	v.SetDefault("client_auth.ca_file", "")
	v.SetDefault("client_auth.ca_data", "")
	v.SetDefault("client_auth.crl_file", "")
	v.SetDefault("client_auth.verify_depth", 5)
	v.SetDefault("client_auth.allowed_cns", []string{})

	// 证书自动管理默认值
	v.SetDefault("auto_management.enable", false)
	v.SetDefault("auto_management.reload_interval", 300)
	v.SetDefault("auto_management.watch_files", true)
	v.SetDefault("auto_management.backup_enabled", false)
	v.SetDefault("auto_management.backup_dir", "./certs/backup")
	v.SetDefault("auto_management.health_check.enable", true)
	v.SetDefault("auto_management.health_check.interval", 60)
	v.SetDefault("auto_management.health_check.timeout", 10)

	// ALPN协议配置默认值
	v.SetDefault("alpn.next_protos", []string{"h2", "http/1.1"})
	v.SetDefault("alpn.h2_enabled", true)
	v.SetDefault("alpn.h2c_enabled", false)
	v.SetDefault("alpn.http1_enabled", true)

	// 会话管理默认值
	v.SetDefault("session.tickets_enabled", true)
	v.SetDefault("session.ticket_key", "")
	v.SetDefault("session.ticket_key_rotation", false)
	v.SetDefault("session.ticket_lifetime", 86400)
	v.SetDefault("session.cache_size", 1000)
	v.SetDefault("session.cache_ttl", 3600)

	// OCSP配置默认值
	v.SetDefault("ocsp.enable", false)
	v.SetDefault("ocsp.stapling_file", "")
	v.SetDefault("ocsp.responder_url", "")
	v.SetDefault("ocsp.cache_time", 3600)

	// HSTS配置默认值
	v.SetDefault("hsts.enable", false)
	v.SetDefault("hsts.max_age", 31536000)
	v.SetDefault("hsts.include_subdomains", false)
	v.SetDefault("hsts.preload", false)
	v.SetDefault("hsts.header", "Strict-Transport-Security")

	// 性能优化默认值
	v.SetDefault("performance.read_timeout", 60)
	v.SetDefault("performance.write_timeout", 60)
	v.SetDefault("performance.idle_timeout", 120)
	v.SetDefault("performance.max_header_bytes", 1048576)
	v.SetDefault("performance.keep_alive", true)
	v.SetDefault("performance.tcp_nodelay", true)
	v.SetDefault("performance.reuse_port", false)
	v.SetDefault("performance.read_buffer_size", 4096)
	v.SetDefault("performance.write_buffer_size", 4096)

	// 监控配置默认值
	v.SetDefault("monitoring.enable", false)
	v.SetDefault("monitoring.metrics_path", "/metrics")
	v.SetDefault("monitoring.log_connections", false)
	v.SetDefault("monitoring.log_handshakes", true)
	v.SetDefault("monitoring.log_errors", true)
	v.SetDefault("monitoring.stats_interval", 60)
}

// GenerateDefaultContent 实现 ConfigInterface 接口 - 生成默认配置文件内容
func (c TLSServerConfig) GenerateDefaultContent() string {
	return `# YYHertz TLS Configuration
# TLS/SSL 安全传输层配置文件

# 基本配置
basic:
  enable: false                              # 是否启用TLS
  environment: "development"                 # 环境: development, testing, production
  debug: true                                # 调试模式
  log_level: "info"                          # 日志级别: debug, info, warn, error

# 证书配置
certificate:
  cert_file: "./certs/server.crt"            # 服务器证书文件路径
  key_file: "./certs/server.key"             # 服务器私钥文件路径
  ca_file: ""                                # CA证书文件路径
  cert_data: ""                              # PEM编码的证书内容（直接配置）
  key_data: ""                               # PEM编码的私钥内容（直接配置）
  pass_phrase: ""                            # 私钥密码
  validity_period: 30                        # 证书有效期检查阈值(天)

# TLS版本配置
version:
  min_version: "1.2"                         # 最小TLS版本: 1.0, 1.1, 1.2, 1.3
  max_version: "1.3"                         # 最大TLS版本: 1.0, 1.1, 1.2, 1.3
  supported_versions: ["1.2", "1.3"]        # 支持的TLS版本列表

# 密码套件配置
cipher:
  suites:                                    # 支持的密码套件
    - "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"
    - "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305"
    - "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
    - "TLS_RSA_WITH_AES_256_GCM_SHA384"
    - "TLS_RSA_WITH_AES_128_GCM_SHA256"
  prefer_server: true                        # 优先使用服务器密码套件
  curves: ["P-256", "P-384", "P-521"]        # 支持的椭圆曲线
  signature_algs: ["RSA-PSS-SHA256", "ECDSA-SHA256"] # 签名算法

# 客户端认证配置
client_auth:
  mode: "NoClientCert"                       # 客户端认证模式: NoClientCert, RequestClientCert, RequireAnyClientCert, VerifyClientCertIfGiven, RequireAndVerifyClientCert
  ca_file: ""                                # 客户端CA证书文件
  ca_data: ""                                # 客户端CA证书内容
  crl_file: ""                               # 证书撤销列表文件
  verify_depth: 5                            # 证书链验证深度
  allowed_cns: []                            # 允许的客户端证书CN列表

# 证书自动管理
auto_management:
  enable: false                              # 启用证书自动管理
  reload_interval: 300                       # 证书重载检查间隔(秒)
  watch_files: true                          # 监控证书文件变化
  backup_enabled: false                      # 启用证书备份
  backup_dir: "./certs/backup"               # 证书备份目录
  
  # 健康检查
  health_check:
    enable: true                             # 启用证书健康检查
    interval: 60                             # 检查间隔(秒)
    timeout: 10                              # 检查超时(秒)

# ALPN协议配置
alpn:
  next_protos: ["h2", "http/1.1"]            # 支持的应用层协议
  h2_enabled: true                           # 启用HTTP/2
  h2c_enabled: false                         # 启用HTTP/2 over cleartext
  http1_enabled: true                        # 启用HTTP/1.1

# 会话管理
session:
  tickets_enabled: true                      # 启用会话票据
  ticket_key: ""                             # 会话票据密钥
  ticket_key_rotation: false                 # 启用票据密钥轮换
  ticket_lifetime: 86400                     # 会话票据生存时间(秒)
  cache_size: 1000                           # 会话缓存大小
  cache_ttl: 3600                            # 会话缓存TTL(秒)

# OCSP配置
ocsp:
  enable: false                              # 启用OCSP装订
  stapling_file: ""                          # OCSP响应文件
  responder_url: ""                          # OCSP响应服务器URL
  cache_time: 3600                           # OCSP响应缓存时间(秒)

# HSTS配置
hsts:
  enable: false                              # 启用HSTS
  max_age: 31536000                          # HSTS最大生存时间(秒)
  include_subdomains: false                  # 包含子域名
  preload: false                             # 启用HSTS预加载
  header: "Strict-Transport-Security"        # HSTS头名称

# 性能优化
performance:
  read_timeout: 60                           # 读取超时(秒)
  write_timeout: 60                          # 写入超时(秒)
  idle_timeout: 120                          # 空闲超时(秒)
  max_header_bytes: 1048576                  # 最大请求头大小(字节)
  keep_alive: true                           # 启用Keep-Alive
  tcp_nodelay: true                          # 启用TCP_NODELAY
  reuse_port: false                          # 启用端口复用
  read_buffer_size: 4096                     # 读缓冲区大小(字节)
  write_buffer_size: 4096                    # 写缓冲区大小(字节)

# 监控配置
monitoring:
  enable: false                              # 启用TLS监控
  metrics_path: "/metrics"                   # 监控指标路径
  log_connections: false                     # 记录连接日志
  log_handshakes: true                       # 记录握手日志
  log_errors: true                           # 记录错误日志
  stats_interval: 60                         # 统计信息更新间隔(秒)
`
}
