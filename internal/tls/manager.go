package mantistls

import (
	"crypto/tls"
	"fmt"

	"github.com/rs/zerolog"
)

// Manager handles TLS certificate loading and configuration.
type Manager struct {
	config *tls.Config
	logger zerolog.Logger
}

// NewManager creates a TLS manager from certificate files.
func NewManager(certFile, keyFile string, logger zerolog.Logger) (*Manager, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load TLS cert: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		},
	}

	logger.Info().Str("cert", certFile).Msg("TLS certificate loaded")

	return &Manager{config: tlsCfg, logger: logger}, nil
}

// TLSConfig returns the configured tls.Config for use by DoH and DoT servers.
func (m *Manager) TLSConfig() *tls.Config {
	return m.config
}
