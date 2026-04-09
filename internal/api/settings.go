package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
)

// Writable setting keys. Keys outside this set are rejected.
var allowedSettingKeys = map[string]bool{
	"dns.upstreams":       true,
	"dns.blocking_mode":   true,
	"dns.cache_size":      true,
	"dns.resolver_mode":   true,
	"logging.level":       true,
	"logging.query_log":   true,
	"logging.retention_days": true,
	"logging.privacy_mode":  true,
	"gravity.update_interval": true,
}

// SettingsHandler handles settings endpoints.
type SettingsHandler struct {
	repo domain.SettingsRepository
}

// GetAll returns all settings.
func (h *SettingsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.GetAll(r.Context())
	if err != nil {
		Error(w, "INTERNAL_ERROR", "failed to get settings", http.StatusInternalServerError)
		return
	}
	if settings == nil {
		settings = []domain.Setting{}
	}
	Success(w, settings)
}

// Update sets settings. Only keys in the allowlist are accepted.
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}

	for key := range updates {
		if !allowedSettingKeys[key] {
			Error(w, "VALIDATION_ERROR", "setting key not allowed: "+key, http.StatusBadRequest)
			return
		}
	}

	for key, value := range updates {
		if err := validateSettingValue(key, value); err != nil {
			Error(w, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
			return
		}
		if err := h.repo.Set(r.Context(), key, value); err != nil {
			Error(w, "INTERNAL_ERROR", "failed to set "+key, http.StatusInternalServerError)
			return
		}
	}

	Success(w, map[string]string{"status": "updated"})
}

// validateSettingValue validates the value for a given setting key.
func validateSettingValue(key, value string) error {
	switch key {
	case "dns.upstreams":
		// Comma-separated list of valid IPs or hostnames.
		for _, upstream := range strings.Split(value, ",") {
			upstream = strings.TrimSpace(upstream)
			if upstream == "" {
				continue
			}
			host, _, err := net.SplitHostPort(upstream)
			if err != nil {
				host = upstream
			}
			if ip := net.ParseIP(host); ip != nil {
				continue
			}
			// Allow hostnames as well.
			if _, err := net.LookupHost(host); err != nil {
				return fmt.Errorf("invalid upstream %q: not a valid IP or resolvable hostname", upstream)
			}
		}
	case "dns.cache_size":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("dns.cache_size must be a non-negative integer")
		}
	case "dns.resolver_mode":
		valid := map[string]bool{"recursive": true, "forwarding": true}
		if !valid[value] {
			return fmt.Errorf("dns.resolver_mode must be 'recursive' or 'forwarding'")
		}
	case "logging.retention_days":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return fmt.Errorf("logging.retention_days must be a non-negative integer")
		}
	case "gravity.update_interval":
		if _, err := time.ParseDuration(value); err != nil {
			return fmt.Errorf("gravity.update_interval must be a valid Go duration (e.g. 24h, 30m)")
		}
	}
	return nil
}
