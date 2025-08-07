package middleware

import (
	"crypto/tls"
	"testing"
)

func TestDefaultTLSConfig(t *testing.T) {
	cfg := DefaultTLSConfig()
	
	if cfg == nil {
		t.Fatal("DefaultTLSConfig() returned nil")
	}
	
	if cfg.Enable {
		t.Error("Expected Enable to be false by default")
	}
	
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected MinVersion to be TLS 1.2, got %d", cfg.MinVersion)
	}
	
	if cfg.MaxVersion != tls.VersionTLS13 {
		t.Errorf("Expected MaxVersion to be TLS 1.3, got %d", cfg.MaxVersion)
	}
	
	if cfg.HSTSMaxAge != 31536000 {
		t.Errorf("Expected HSTSMaxAge to be 31536000, got %d", cfg.HSTSMaxAge)
	}
	
	if !cfg.HSTSEnabled {
		t.Error("Expected HSTSEnabled to be true by default")
	}
	
	if !cfg.HSTSSubdomains {
		t.Error("Expected HSTSSubdomains to be true by default")
	}
}

func TestValidateTLSConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *TLSConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name: "disabled TLS",
			config: &TLSConfig{
				Enable: false,
			},
			wantErr: false,
		},
		{
			name: "valid TLS config",
			config: &TLSConfig{
				Enable:       true,
				CertFile:     "/path/to/cert.pem",
				KeyFile:      "/path/to/key.pem",
				MinVersion:   tls.VersionTLS12,
				MaxVersion:   tls.VersionTLS13,
				HSTSMaxAge:   31536000,
				RedirectPort: 443,
			},
			wantErr: false,
		},
		{
			name: "missing cert file",
			config: &TLSConfig{
				Enable:  true,
				KeyFile: "/path/to/key.pem",
			},
			wantErr: true,
		},
		{
			name: "missing key file",
			config: &TLSConfig{
				Enable:   true,
				CertFile: "/path/to/cert.pem",
			},
			wantErr: true,
		},
		{
			name: "invalid TLS version range",
			config: &TLSConfig{
				Enable:     true,
				CertFile:   "/path/to/cert.pem",
				KeyFile:    "/path/to/key.pem",
				MinVersion: tls.VersionTLS13,
				MaxVersion: tls.VersionTLS12,
			},
			wantErr: true,
		},
		{
			name: "negative HSTS max age",
			config: &TLSConfig{
				Enable:     true,
				CertFile:   "/path/to/cert.pem",
				KeyFile:    "/path/to/key.pem",
				HSTSMaxAge: -1,
			},
			wantErr: true,
		},
		{
			name: "invalid redirect port",
			config: &TLSConfig{
				Enable:       true,
				CertFile:     "/path/to/cert.pem",
				KeyFile:      "/path/to/key.pem",
				RedirectPort: 70000,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTLSConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}