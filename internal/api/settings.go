package api

import (
	"encoding/json"
	"net/http"

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
		if err := h.repo.Set(r.Context(), key, value); err != nil {
			Error(w, "INTERNAL_ERROR", "failed to set "+key, http.StatusInternalServerError)
			return
		}
	}

	Success(w, map[string]string{"status": "updated"})
}
