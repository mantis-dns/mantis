package api

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		return isSameHost(origin, r.Host)
	},
}

// Stream upgrades to WebSocket for live query streaming.
func (h *QueryHandler) Stream(w http.ResponseWriter, r *http.Request) {
	if h.activeConns.Load() >= maxWebSocketConns {
		Error(w, "RATE_LIMITED", "too many WebSocket connections", http.StatusTooManyRequests)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.activeConns.Add(1)
	defer h.activeConns.Add(-1)
	defer conn.Close()

	ch := h.bus.Subscribe(1000)
	defer h.bus.Unsubscribe(ch)

	// Ping every 30s.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if conn.WriteMessage(websocket.PingMessage, nil) != nil {
				return
			}
		}
	}()

	for ev := range ch {
		if err := conn.WriteJSON(ev); err != nil {
			break
		}
	}
}
