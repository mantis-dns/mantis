package mantistls

import (
	"crypto/tls"

	"golang.org/x/crypto/acme/autocert"
	"github.com/rs/zerolog"
)

// NewACMEManager creates a TLS manager using Let's Encrypt auto-provisioning.
func NewACMEManager(domain, email, cacheDir string, logger zerolog.Logger) *Manager {
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(cacheDir),
		Email:      email,
	}

	tlsCfg := &tls.Config{
		GetCertificate: m.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	}

	logger.Info().Str("domain", domain).Msg("ACME auto-cert enabled")

	return &Manager{config: tlsCfg, logger: logger}
}
