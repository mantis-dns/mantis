package config

import (
	"fmt"
	"net"
	"strings"
)

// Validate checks the configuration for invalid values.
func Validate(cfg *Config) error {
	var errs []string

	if err := validateAddress(cfg.DNS.ListenAddress, "dns.listen_address"); err != nil {
		errs = append(errs, err.Error())
	}

	if len(cfg.DNS.Upstreams) == 0 && cfg.DNS.ResolverMode == "forward" {
		errs = append(errs, "dns.upstreams: at least one upstream required in forward mode")
	}

	if cfg.DNS.ResolverMode != "forward" && cfg.DNS.ResolverMode != "recursive" {
		errs = append(errs, fmt.Sprintf("dns.resolver_mode: must be 'forward' or 'recursive', got %q", cfg.DNS.ResolverMode))
	}

	if cfg.DNS.CacheSize < 0 {
		errs = append(errs, "dns.cache_size: must be >= 0")
	}

	if cfg.DNS.BlockingMode != "null" && cfg.DNS.BlockingMode != "nxdomain" {
		errs = append(errs, fmt.Sprintf("dns.blocking_mode: must be 'null' or 'nxdomain', got %q", cfg.DNS.BlockingMode))
	}

	if cfg.DoH.Enabled {
		if err := validateAddress(cfg.DoH.ListenAddress, "doh.listen_address"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if cfg.DoT.Enabled {
		if err := validateAddress(cfg.DoT.ListenAddress, "dot.listen_address"); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if err := validateAddress(cfg.API.ListenAddress, "api.listen_address"); err != nil {
		errs = append(errs, err.Error())
	}

	if cfg.API.RateLimit < 0 {
		errs = append(errs, "api.rate_limit: must be >= 0")
	}

	if cfg.Storage.DataDir == "" {
		errs = append(errs, "storage.data_dir: must not be empty")
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[cfg.Logging.Level] {
		errs = append(errs, fmt.Sprintf("logging.level: must be debug/info/warn/error, got %q", cfg.Logging.Level))
	}

	validQueryLog := map[string]bool{"all": true, "blocked": true, "none": true}
	if !validQueryLog[cfg.Logging.QueryLog] {
		errs = append(errs, fmt.Sprintf("logging.query_log: must be all/blocked/none, got %q", cfg.Logging.QueryLog))
	}

	if cfg.Logging.RetentionDays < 1 {
		errs = append(errs, "logging.retention_days: must be >= 1")
	}

	if cfg.DHCP.Enabled {
		if cfg.DHCP.Interface == "" {
			errs = append(errs, "dhcp.interface: required when DHCP is enabled")
		}
		if cfg.DHCP.RangeStart == "" {
			errs = append(errs, "dhcp.range_start: required when DHCP is enabled")
		}
		if cfg.DHCP.RangeEnd == "" {
			errs = append(errs, "dhcp.range_end: required when DHCP is enabled")
		}
		if cfg.DHCP.RangeStart != "" && net.ParseIP(cfg.DHCP.RangeStart) == nil {
			errs = append(errs, fmt.Sprintf("dhcp.range_start: invalid IP %q", cfg.DHCP.RangeStart))
		}
		if cfg.DHCP.RangeEnd != "" && net.ParseIP(cfg.DHCP.RangeEnd) == nil {
			errs = append(errs, fmt.Sprintf("dhcp.range_end: invalid IP %q", cfg.DHCP.RangeEnd))
		}
		if cfg.DHCP.Gateway != "" && net.ParseIP(cfg.DHCP.Gateway) == nil {
			errs = append(errs, fmt.Sprintf("dhcp.gateway: invalid IP %q", cfg.DHCP.Gateway))
		}
	}

	if cfg.Gravity.UpdateInterval.Duration < 0 {
		errs = append(errs, "gravity.update_interval: must be >= 0")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func validateAddress(addr, field string) error {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("%s: invalid address %q: %w", field, addr, err)
	}
	if host != "" && net.ParseIP(host) == nil {
		return fmt.Errorf("%s: invalid IP in %q", field, addr)
	}
	port, err := net.LookupPort("tcp", portStr)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("%s: invalid port in %q", field, addr)
	}
	return nil
}
