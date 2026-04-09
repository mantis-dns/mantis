package api

import (
	"encoding/json"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mantis-dns/mantis/internal/domain"
)

// DHCPHandler handles DHCP lease endpoints.
type DHCPHandler struct {
	leases domain.LeaseRepository
}

// ListLeases returns all DHCP leases.
func (h *DHCPHandler) ListLeases(w http.ResponseWriter, r *http.Request) {
	leases, err := h.leases.List(r.Context())
	if err != nil {
		Error(w, "INTERNAL_ERROR", "failed to list leases", http.StatusInternalServerError)
		return
	}
	if leases == nil {
		leases = []domain.DhcpLease{}
	}
	Success(w, leases)
}

// CreateStaticLease adds a static DHCP lease.
func (h *DHCPHandler) CreateStaticLease(w http.ResponseWriter, r *http.Request) {
	var lease domain.DhcpLease
	if err := json.NewDecoder(r.Body).Decode(&lease); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid request body", http.StatusBadRequest)
		return
	}
	if lease.MAC == "" || lease.IP == "" {
		Error(w, "VALIDATION_ERROR", "mac and ip required", http.StatusBadRequest)
		return
	}
	if _, err := net.ParseMAC(lease.MAC); err != nil {
		Error(w, "VALIDATION_ERROR", "invalid MAC address format", http.StatusBadRequest)
		return
	}
	if net.ParseIP(lease.IP) == nil {
		Error(w, "VALIDATION_ERROR", "invalid IP address format", http.StatusBadRequest)
		return
	}

	lease.IsStatic = true
	lease.LeaseEnd = time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

	if err := h.leases.Create(r.Context(), &lease); err != nil {
		Error(w, "INTERNAL_ERROR", "failed to create static lease", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, successResponse{
		Data: lease,
		Meta: meta{RequestID: requestID()},
	})
}

// DeleteStaticLease removes a static DHCP lease.
func (h *DHCPHandler) DeleteStaticLease(w http.ResponseWriter, r *http.Request) {
	mac := chi.URLParam(r, "mac")
	if err := h.leases.Delete(r.Context(), mac); err != nil {
		Error(w, "INTERNAL_ERROR", "failed to delete lease", http.StatusInternalServerError)
		return
	}
	Success(w, map[string]string{"status": "deleted"})
}
