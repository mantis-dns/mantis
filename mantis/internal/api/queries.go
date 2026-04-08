package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/event"
)

// QueryHandler handles query log endpoints.
type QueryHandler struct {
	queryLog domain.QueryLogRepository
	bus      *event.Bus
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
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Stream upgrades to WebSocket for live query streaming.
func (h *QueryHandler) Stream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ch := h.bus.Subscribe(1000)

	for ev := range ch {
		if err := conn.WriteJSON(ev); err != nil {
			break
		}
	}
}
