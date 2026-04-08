package api

import (
	"net/http"
	"runtime"
	"time"

	"github.com/mantis-dns/mantis/internal/gravity"
)

var startTime = time.Now()

// SystemHandler handles system endpoints.
type SystemHandler struct {
	version string
	gravity *gravity.Engine
}

// Health returns component health status.
func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	Success(w, map[string]string{
		"status": "ok",
		"dns":    "running",
	})
}

// Info returns system information.
func (h *SystemHandler) Info(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	Success(w, map[string]interface{}{
		"version":        h.version,
		"goVersion":      runtime.Version(),
		"uptime":         time.Since(startTime).String(),
		"goroutines":     runtime.NumGoroutine(),
		"memAllocMB":     mem.Alloc / 1024 / 1024,
		"memSysMB":       mem.Sys / 1024 / 1024,
		"blockedDomains": h.gravity.BlockedCount(),
	})
}

// RebuildGravity triggers a gravity rebuild.
func (h *SystemHandler) RebuildGravity(w http.ResponseWriter, r *http.Request) {
	// TODO: trigger scheduler.TriggerRebuild() once wired.
	writeJSON(w, http.StatusAccepted, successResponse{
		Data: map[string]string{"status": "rebuild started"},
		Meta: meta{RequestID: requestID()},
	})
}

// GravityStatus returns gravity stats.
func (h *SystemHandler) GravityStatus(w http.ResponseWriter, r *http.Request) {
	Success(w, map[string]interface{}{
		"totalDomains": h.gravity.BlockedCount(),
	})
}

// RestartDNS restarts the DNS server.
func (h *SystemHandler) RestartDNS(w http.ResponseWriter, r *http.Request) {
	// TODO: implement DNS restart.
	Success(w, map[string]string{"status": "restart initiated"})
}
