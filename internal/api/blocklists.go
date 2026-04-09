package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mantis-dns/mantis/internal/domain"
)

// BlocklistHandler handles blocklist CRUD.
type BlocklistHandler struct {
	repo domain.BlocklistRepository
}

// List returns all blocklist sources.
func (h *BlocklistHandler) List(w http.ResponseWriter, r *http.Request) {
	sources, err := h.repo.List(r.Context())
	if err != nil {
		Error(w, "INTERNAL_ERROR", "failed to list blocklists", http.StatusInternalServerError)
		return
	}
	if sources == nil {
		sources = []domain.BlocklistSource{}
	}
	Success(w, sources)
}

// Create adds a new blocklist source.
func (h *BlocklistHandler) Create(w http.ResponseWriter, r *http.Request) {
	var src domain.BlocklistSource
	if err := json.NewDecoder(r.Body).Decode(&src); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}
	if src.Name == "" || src.URL == "" {
		Error(w, "VALIDATION_ERROR", "name and url required", http.StatusBadRequest)
		return
	}
	if err := validateBlocklistURL(src.URL); err != nil {
		Error(w, "VALIDATION_ERROR", err.Error(), http.StatusBadRequest)
		return
	}

	src.ID = uuid.New().String()
	src.Enabled = true
	src.LastStatus = domain.StatusPending

	if err := h.repo.Create(r.Context(), &src); err != nil {
		Error(w, "INTERNAL_ERROR", "failed to create blocklist", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, successResponse{
		Data: src,
		Meta: meta{RequestID: requestID()},
	})
}

// Update modifies a blocklist source.
func (h *BlocklistHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	existing, err := h.repo.Get(r.Context(), id)
	if err != nil {
		Error(w, "NOT_FOUND", "blocklist not found", http.StatusNotFound)
		return
	}

	var updates domain.BlocklistSource
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}

	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.URL != "" {
		existing.URL = updates.URL
	}
	existing.Enabled = updates.Enabled
	if updates.Format != "" {
		existing.Format = updates.Format
	}

	if err := h.repo.Update(r.Context(), existing); err != nil {
		Error(w, "INTERNAL_ERROR", "failed to update blocklist", http.StatusInternalServerError)
		return
	}

	Success(w, existing)
}

// Delete removes a blocklist source.
func (h *BlocklistHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.repo.Delete(r.Context(), id); err != nil {
		Error(w, "INTERNAL_ERROR", "failed to delete blocklist", http.StatusInternalServerError)
		return
	}
	Success(w, map[string]string{"status": "deleted"})
}

// validateBlocklistURL ensures the URL is HTTP(S) and not targeting private/loopback addresses.
func validateBlocklistURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https")
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("URL must not target private/loopback addresses")
		}
	}
	return nil
}
