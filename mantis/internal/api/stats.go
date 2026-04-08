package api

import "net/http"

// StatsHandler handles statistics endpoints.
type StatsHandler struct{}

// Summary returns dashboard summary stats.
func (h *StatsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	// TODO: wire to stats aggregator (Task 22).
	Success(w, map[string]interface{}{
		"totalQueries":  0,
		"blockedCount":  0,
		"cachedCount":   0,
		"blockedPercent": 0.0,
		"cacheHitRatio": 0.0,
	})
}

// Overtime returns time-series query data.
func (h *StatsHandler) Overtime(w http.ResponseWriter, r *http.Request) {
	Success(w, []interface{}{})
}

// TopDomains returns top queried/blocked domains.
func (h *StatsHandler) TopDomains(w http.ResponseWriter, r *http.Request) {
	Success(w, map[string]interface{}{
		"allowed": []interface{}{},
		"blocked": []interface{}{},
	})
}

// TopClients returns top clients by volume.
func (h *StatsHandler) TopClients(w http.ResponseWriter, r *http.Request) {
	Success(w, []interface{}{})
}
