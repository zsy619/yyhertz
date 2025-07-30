package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"
)

// TLSManager TLS证书管理器
type TLSManager struct {
	config     *TLSServerConfig
	tlsConfig  *tls.Config
	certWatcher *CertWatcher
}

// TLSServerConfig TLS服务器配置
type TLSServerConfig struct {
	// 基本配置
	Enable         bool     `json:"enable" yaml:"enable"`
	CertFile       string   `json:"cert_file" yaml:"cert_file"`
	KeyFile        string   `json:"key_file" yaml:"key_file"`
	CAFile         string   `json:"ca_file" yaml:"ca_file"`
	
	// TLS版本配置
	MinVersion     string   `json:"min_version" yaml:"min_version"`     // "1.0", "1.1", "1.2", "1.3"
	MaxVersion     string   `json:"max_version" yaml:"max_version"`
	
	// 密码套件配置
	CipherSuites   []string `json:"cipher_suites" yaml:"cipher_suites"`
	PreferServer   bool     `json:"prefer_server" yaml:"prefer_server"`
	
	// 客户端认证配置
	ClientAuth     string   `json:"client_auth" yaml:"client_auth"`     // "NoClientCert", "RequestClientCert", "RequireAnyClientCert", "VerifyClientCertIfGiven", "RequireAndVerifyClientCert"
	ClientCAFile   string   `json:"client_ca_file" yaml:"client_ca_file"`
	
	// 证书自动重载
	AutoReload     bool     `json:"auto_reload" yaml:"auto_reload"`
	ReloadInterval int      `json:"reload_interval" yaml:"reload_interval"` // 秒
	
	// ALPN配置
	NextProtos     []string `json:"next_protos" yaml:"next_protos"`
	
	// 会话恢复
	SessionTicketsEnabled bool `json:"session_tickets_enabled" yaml:"session_tickets_enabled"`
	SessionTicketKey      string `json:"session_ticket_key" yaml:"session_ticket_key"`
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
	return &TLSServerConfig{
		Enable:                false,
		MinVersion:            "1.2",
		MaxVersion:            "1.3",
		PreferServer:          true,
		ClientAuth:            "NoClientCert",
		AutoReload:            false,
		ReloadInterval:        300, // 5分钟
		NextProtos:            []string{"h2", "http/1.1"},
		SessionTicketsEnabled: true,
		CipherSuites: []string{
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_RSA_WITH_AES_256_GCM_SHA384",
			"TLS_RSA_WITH_AES_128_GCM_SHA256",
		},
	}
}

// NewTLSManager 创建TLS管理器
func NewTLSManager(config *TLSServerConfig) (*TLSManager, error) {
	if config == nil {
		config = DefaultTLSServerConfig()
	}
	
	manager := &TLSManager{
		config: config,
	}
	
	if config.Enable {
		if err := manager.loadTLSConfig(); err != nil {
			return nil, fmt.Errorf("加载TLS配置失败: %w", err)
		}
		
		if config.AutoReload {
			manager.startCertWatcher()
		}
	}
	
	GetGlobalLogger().WithFields(map[string]any{
		"enabled":     config.Enable,
		"auto_reload": config.AutoReload,
		"min_version": config.MinVersion,
		"max_version": config.MaxVersion,
	}).Info("TLS管理器初始化完成")
	
	return manager, nil
}

// loadTLSConfig 加载TLS配置
func (m *TLSManager) loadTLSConfig() error {
	// 加载证书
	cert, err := tls.LoadX509KeyPair(m.config.CertFile, m.config.KeyFile)
	if err != nil {
		return fmt.Errorf("加载证书失败: %w", err)
	}
	
	// 创建TLS配置
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: m.config.PreferServer,
		NextProtos:               m.config.NextProtos,
		SessionTicketsDisabled:   !m.config.SessionTicketsEnabled,
	}
	
	// 设置TLS版本
	if minVer, err := parseTLSVersion(m.config.MinVersion); err != nil {
		return fmt.Errorf("解析最小TLS版本失败: %w", err)
	} else {
		tlsConfig.MinVersion = minVer
	}
	
	if maxVer, err := parseTLSVersion(m.config.MaxVersion); err != nil {
		return fmt.Errorf("解析最大TLS版本失败: %w", err)
	} else {
		tlsConfig.MaxVersion = maxVer
	}
	
	// 设置密码套件
	if cipherSuites, err := parseCipherSuites(m.config.CipherSuites); err != nil {
		return fmt.Errorf("解析密码套件失败: %w", err)
	} else {
		tlsConfig.CipherSuites = cipherSuites
	}
	
	// 设置客户端认证
	if clientAuth, err := parseClientAuth(m.config.ClientAuth); err != nil {
		return fmt.Errorf("解析客户端认证模式失败: %w", err)
	} else {
		tlsConfig.ClientAuth = clientAuth
	}
	
	// 加载客户端CA证书
	if m.config.ClientCAFile != "" {
		clientCAs, err := loadCACerts(m.config.ClientCAFile)
		if err != nil {
			return fmt.Errorf("加载客户端CA证书失败: %w", err)
		}
		tlsConfig.ClientCAs = clientCAs
	}
	
	// 设置会话票据密钥
	if m.config.SessionTicketKey != "" {
		key := []byte(m.config.SessionTicketKey)
		if len(key) != 32 {
			return fmt.Errorf("会话票据密钥长度必须为32字节")
		}
		tlsConfig.SetSessionTicketKeys([][32]byte{[32]byte(key)})
	}
	
	m.tlsConfig = tlsConfig
	
	GetGlobalLogger().WithFields(map[string]any{
		"cert_file":    m.config.CertFile,
		"key_file":     m.config.KeyFile,
		"client_auth":  m.config.ClientAuth,
		"cipher_count": len(tlsConfig.CipherSuites),
	}).Info("TLS配置加载成功")
	
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
		certFile:   m.config.CertFile,
		keyFile:    m.config.KeyFile,
		reloadChan: make(chan struct{}, 1),
		stopChan:   make(chan struct{}),
		interval:   time.Duration(m.config.ReloadInterval) * time.Second,
	}
	
	go m.certWatcher.watch(m)
	
	GetGlobalLogger().WithFields(map[string]any{
		"interval": m.config.ReloadInterval,
	}).Info("证书监视器启动")
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