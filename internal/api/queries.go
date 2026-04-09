package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/event"
)

const maxPerPage = 1000
const maxWebSocketConns = 100

// QueryHandler handles query log endpoints.
type QueryHandler struct {
	queryLog    domain.QueryLogRepository
	bus         *event.Bus
	activeConns atomic.Int64
}

// List returns paginated query log entries.
func (h *QueryHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("perPage"))
	if perPage < 1 {
		perPage = 50
	}
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	filter := domain.QueryLogFilter{
		Domain:   r.URL.Query().Get("domain"),
		ClientIP: r.URL.Query().Get("client"),
		Result:   r.URL.Query().Get("result"),
		Page:     page,
		PerPage:  perPage,
	}

	entries, total, err := h.queryLog.Query(r.Context(), filter)
	if err != nil {
		Error(w, "INTERNAL_ERROR", "failed to query log", http.StatusInternalServerError)
		return
	}

	Paginated(w, entries, page, perPage, total)
}

var trustedAPIHost string

// SetTrustedAPIHost sets the trusted host for WebSocket origin checks.
func SetTrustedAPIHost(host string) {
	trustedAPIHost = host
}

// Stream upgrades to WebSocket for live query streaming.
func (h *QueryHandler) Stream(w http.ResponseWriter, r *http.Request) {
	if h.activeConns.Load() >= maxWebSocketConns {
		Error(w, "RATE_LIMITED", "too many WebSocket connections", http.StatusTooManyRequests)
		return
	}

	origin := r.Header.Get("Origin")
	if origin != "" && !isSameOrigin(origin, trustedAPIHost) {
		Error(w, "FORBIDDEN", "origin not allowed", http.StatusForbidden)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		return
	}
	defer conn.CloseNow()

	h.activeConns.Add(1)
	defer h.activeConns.Add(-1)

	ch := h.bus.Subscribe(1000)
	defer h.bus.Unsubscribe(ch)

	ctx := r.Context()

	// Mutex protects all concurrent writes to conn.
	var mu sync.Mutex
	done := make(chan struct{})

	// Ping every 30s.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				mu.Lock()
				err := conn.Ping(ctx)
				mu.Unlock()
				if err != nil {
					return
				}
			}
		}
	}()

	for ev := range ch {
		data, err := json.Marshal(ev)
		if err != nil {
			continue
		}
		mu.Lock()
		err = conn.Write(ctx, websocket.MessageText, data)
		mu.Unlock()
		if err != nil {
			break
		}
	}

	close(done)
}
