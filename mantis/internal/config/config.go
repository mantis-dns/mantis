package config

import (
	"fmt"
	"time"
)

// Duration wraps time.Duration with TOML string unmarshaling support.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", string(text), err)
	}
	return nil
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(d.Duration.String()), nil
}

// Config holds all Mantis configuration.
type Config struct {
	DNS     DNSConfig     `toml:"dns"`
	DoH     DoHConfig     `toml:"doh"`
	DoT     DoTConfig     `toml:"dot"`
	TLS     TLSConfig     `toml:"tls"`
	DHCP    DHCPConfig    `toml:"dhcp"`
	API     APIConfig     `toml:"api"`
	Storage StorageConfig `toml:"storage"`
	Logging LoggingConfig `toml:"logging"`
	Gravity GravityConfig `toml:"gravity"`
}

// DNSConfig configures the DNS server.
type DNSConfig struct {
	ListenAddress string   `toml:"listen_address"`
	Upstreams     []string `toml:"upstreams"`
	ResolverMode  string   `toml:"resolver_mode"`
	CacheSize     int      `toml:"cache_size"`
	BlockingMode  string   `toml:"blocking_mode"`
}

// DoHConfig configures DNS-over-HTTPS.
type DoHConfig struct {
	Enabled       bool   `toml:"enabled"`
	ListenAddress string `toml:"listen_address"`
}

// DoTConfig configures DNS-over-TLS.
type DoTConfig struct {
	Enabled       bool   `toml:"enabled"`
	ListenAddress string `toml:"listen_address"`
}

// TLSConfig configures TLS certificates.
type TLSConfig struct {
	CertFile    string `toml:"cert_file"`
	KeyFile     string `toml:"key_file"`
	ACMEEnabled bool   `toml:"acme_enabled"`
	ACMEDomain  string `toml:"acme_domain"`
	ACMEEmail   string `toml:"acme_email"`
}

// DHCPConfig configures the DHCP server.
type DHCPConfig struct {
	Enabled       bool          `toml:"enabled"`
	Interface     string        `toml:"interface"`
	RangeStart    string        `toml:"range_start"`
	RangeEnd      string        `toml:"range_end"`
	LeaseDuration Duration `toml:"lease_duration"`
	Gateway       string        `toml:"gateway"`
	SubnetMask    string        `toml:"subnet_mask"`
}

// APIConfig configures the admin API.
type APIConfig struct {
	ListenAddress string `toml:"listen_address"`
	RateLimit     int    `toml:"rate_limit"`
}

// StorageConfig configures data storage.
type StorageConfig struct {
	DataDir string `toml:"data_dir"`
}

// LoggingConfig configures logging behavior.
type LoggingConfig struct {
	Level         string `toml:"level"`
	QueryLog      string `toml:"query_log"`
	RetentionDays int    `toml:"retention_days"`
	PrivacyMode   bool   `toml:"privacy_mode"`
}

// GravityConfig configures blocklist updates.
type GravityConfig struct {
	UpdateInterval Duration `toml:"update_interval"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		DNS: DNSConfig{
			ListenAddress: "0.0.0.0:53",
			Upstreams:     []string{"1.1.1.1", "8.8.8.8"},
			ResolverMode:  "forward",
			CacheSize:     10000,
			BlockingMode:  "null",
		},
		DoH: DoHConfig{
			Enabled:       false,
			ListenAddress: "0.0.0.0:443",
		},
		DoT: DoTConfig{
			Enabled:       false,
			ListenAddress: "0.0.0.0:853",
		},
		TLS: TLSConfig{},
		DHCP: DHCPConfig{
			Enabled:       false,
			LeaseDuration: Duration{24 * time.Hour},
			SubnetMask:    "255.255.255.0",
		},
		API: APIConfig{
			ListenAddress: "0.0.0.0:8080",
			RateLimit:     60,
		},
		Storage: StorageConfig{
			DataDir: "/var/lib/mantis",
		},
		Logging: LoggingConfig{
			Level:         "info",
			QueryLog:      "all",
			RetentionDays: 30,
			PrivacyMode:   false,
		},
		Gravity: GravityConfig{
			UpdateInterval: Duration{24 * time.Hour},
		},
	}
}
