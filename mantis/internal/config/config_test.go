package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.DNS.ListenAddress != "0.0.0.0:53" {
		t.Errorf("expected 0.0.0.0:53, got %s", cfg.DNS.ListenAddress)
	}
	if cfg.DNS.CacheSize != 10000 {
		t.Errorf("expected cache size 10000, got %d", cfg.DNS.CacheSize)
	}
	if cfg.Logging.RetentionDays != 30 {
		t.Errorf("expected 30 retention days, got %d", cfg.Logging.RetentionDays)
	}
}

func TestLoadFromTOML(t *testing.T) {
	tomlContent := `
[dns]
listen_address = "0.0.0.0:5353"
upstreams = ["9.9.9.9"]
resolver_mode = "forward"
cache_size = 5000
blocking_mode = "nxdomain"

[api]
listen_address = "127.0.0.1:9090"
rate_limit = 120

[storage]
data_dir = "/tmp/mantis-test"

[logging]
level = "debug"
query_log = "blocked"
retention_days = 7
`
	path := writeTempFile(t, tomlContent)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DNS.ListenAddress != "0.0.0.0:5353" {
		t.Errorf("expected 0.0.0.0:5353, got %s", cfg.DNS.ListenAddress)
	}
	if cfg.DNS.BlockingMode != "nxdomain" {
		t.Errorf("expected nxdomain, got %s", cfg.DNS.BlockingMode)
	}
	if cfg.API.RateLimit != 120 {
		t.Errorf("expected rate limit 120, got %d", cfg.API.RateLimit)
	}
	if cfg.Storage.DataDir != "/tmp/mantis-test" {
		t.Errorf("expected /tmp/mantis-test, got %s", cfg.Storage.DataDir)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("expected debug, got %s", cfg.Logging.Level)
	}
}

func TestEnvOverride(t *testing.T) {
	tomlContent := `
[dns]
listen_address = "0.0.0.0:53"
upstreams = ["1.1.1.1"]
resolver_mode = "forward"
cache_size = 10000
blocking_mode = "null"

[storage]
data_dir = "/var/lib/mantis"

[logging]
level = "info"
query_log = "all"
retention_days = 30

[api]
listen_address = "0.0.0.0:8080"
`
	path := writeTempFile(t, tomlContent)

	t.Setenv("MANTIS_DNS_LISTEN_ADDRESS", "0.0.0.0:5353")
	t.Setenv("MANTIS_DNS_CACHE_SIZE", "20000")
	t.Setenv("MANTIS_DNS_UPSTREAMS", "9.9.9.9,8.8.4.4")
	t.Setenv("MANTIS_STORAGE_DATA_DIR", "/tmp/override")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.DNS.ListenAddress != "0.0.0.0:5353" {
		t.Errorf("env override failed: expected 0.0.0.0:5353, got %s", cfg.DNS.ListenAddress)
	}
	if cfg.DNS.CacheSize != 20000 {
		t.Errorf("env override failed: expected 20000, got %d", cfg.DNS.CacheSize)
	}
	if len(cfg.DNS.Upstreams) != 2 || cfg.DNS.Upstreams[0] != "9.9.9.9" {
		t.Errorf("env override failed for upstreams: got %v", cfg.DNS.Upstreams)
	}
	if cfg.Storage.DataDir != "/tmp/override" {
		t.Errorf("env override failed: expected /tmp/override, got %s", cfg.Storage.DataDir)
	}
}

func TestMissingFileUsesDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load with empty path failed: %v", err)
	}
	if cfg.DNS.ListenAddress != "0.0.0.0:53" {
		t.Errorf("expected default, got %s", cfg.DNS.ListenAddress)
	}
}

func TestInvalidConfigReturnsError(t *testing.T) {
	tests := []struct {
		name string
		toml string
	}{
		{
			name: "bad resolver mode",
			toml: `
[dns]
listen_address = "0.0.0.0:53"
upstreams = ["1.1.1.1"]
resolver_mode = "invalid"
blocking_mode = "null"
[storage]
data_dir = "/tmp"
[logging]
level = "info"
query_log = "all"
retention_days = 30
[api]
listen_address = "0.0.0.0:8080"
`,
		},
		{
			name: "bad listen address",
			toml: `
[dns]
listen_address = "not-an-address"
upstreams = ["1.1.1.1"]
resolver_mode = "forward"
blocking_mode = "null"
[storage]
data_dir = "/tmp"
[logging]
level = "info"
query_log = "all"
retention_days = 30
[api]
listen_address = "0.0.0.0:8080"
`,
		},
		{
			name: "negative cache size",
			toml: `
[dns]
listen_address = "0.0.0.0:53"
upstreams = ["1.1.1.1"]
resolver_mode = "forward"
cache_size = -1
blocking_mode = "null"
[storage]
data_dir = "/tmp"
[logging]
level = "info"
query_log = "all"
retention_days = 30
[api]
listen_address = "0.0.0.0:8080"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.toml)
			_, err := Load(path)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestDHCPValidation(t *testing.T) {
	tomlContent := `
[dns]
listen_address = "0.0.0.0:53"
upstreams = ["1.1.1.1"]
resolver_mode = "forward"
blocking_mode = "null"

[dhcp]
enabled = true
interface = ""
range_start = ""
range_end = ""

[storage]
data_dir = "/tmp"

[logging]
level = "info"
query_log = "all"
retention_days = 30

[api]
listen_address = "0.0.0.0:8080"
`
	path := writeTempFile(t, tomlContent)
	_, err := Load(path)
	if err == nil {
		t.Error("expected validation error for DHCP without interface, got nil")
	}
}

func TestDurationFields(t *testing.T) {
	tomlContent := `
[dns]
listen_address = "0.0.0.0:53"
upstreams = ["1.1.1.1"]
resolver_mode = "forward"
blocking_mode = "null"

[dhcp]
lease_duration = "12h"

[storage]
data_dir = "/tmp"

[logging]
level = "info"
query_log = "all"
retention_days = 30

[api]
listen_address = "0.0.0.0:8080"

[gravity]
update_interval = "6h"
`
	path := writeTempFile(t, tomlContent)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.DHCP.LeaseDuration.Duration != 12*time.Hour {
		t.Errorf("expected 12h, got %v", cfg.DHCP.LeaseDuration)
	}
	if cfg.Gravity.UpdateInterval.Duration != 6*time.Hour {
		t.Errorf("expected 6h, got %v", cfg.Gravity.UpdateInterval)
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
