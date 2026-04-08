package api

import (
	"net/http"
	"time"

	"github.com/mantis-dns/mantis/internal/stats"
)

// StatsHandler handles statistics endpoints.
type StatsHandler struct {
	aggregator *stats.Aggregator
}

// Summary returns dashboard summary stats.
func (h *StatsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary := h.aggregator.Summary(24 * time.Hour)
	Success(w, summary)
}

// Overtime returns time-series query data.
func (h *StatsHandler) Overtime(w http.ResponseWriter, r *http.Request) {
	points := h.aggregator.Overtime(24 * time.Hour)
	Success(w, points)
}

// TopDomains returns top queried/blocked domains.
func (h *StatsHandler) TopDomains(w http.ResponseWriter, r *http.Request) {
	allowed, blocked := h.aggregator.TopDomains(24*time.Hour, 10)
	Success(w, map[string]interface{}{
		"allowed": allowed,
		"blocked": blocked,
	})
}

// TopClients returns top clients by volume.
func (h *StatsHandler) TopClients(w http.ResponseWriter, r *http.Request) {
	clients := h.aggregator.TopClients(24*time.Hour, 10)
	Success(w, clients)
}
