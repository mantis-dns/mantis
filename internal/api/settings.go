package api

import (
	"encoding/json"
	"net/http"

	"github.com/mantis-dns/mantis/internal/domain"
)

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

// Update sets settings.
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var updates map[string]string
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}

	for key, value := range updates {
		if err := h.repo.Set(r.Context(), key, value); err != nil {
			Error(w, "INTERNAL_ERROR", "failed to set "+key, http.StatusInternalServerError)
			return
		}
	}

	Success(w, map[string]string{"status": "updated"})
}
