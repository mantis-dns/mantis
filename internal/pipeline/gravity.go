package pipeline

import (
	"context"

	mdns "codeberg.org/miekg/dns"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/gravity"
)

// GravityHandler blocks domains found in the Gravity engine.
type GravityHandler struct {
	engine *gravity.Engine
}

// NewGravityHandler creates a gravity handler.
func NewGravityHandler(engine *gravity.Engine) *GravityHandler {
	return &GravityHandler{engine: engine}
}

func (h *GravityHandler) Handle(ctx context.Context, q *domain.Query, next QueryHandler) (*domain.Response, error) {
	if h.engine.IsBlocked(q.Domain) {
		return blockedResponse(q), nil
	}
	return next.Handle(ctx, q, nil)
}

func blockedResponse(q *domain.Query) *domain.Response {
	resp := &domain.Response{Blocked: true}

	switch q.Type {
	case mdns.TypeA:
		resp.Answers = []domain.RR{{
			Name:  q.Domain,
			Type:  mdns.TypeA,
			TTL:   300,
			Value: "0.0.0.0",
		}}
	case mdns.TypeAAAA:
		resp.Answers = []domain.RR{{
			Name:  q.Domain,
			Type:  mdns.TypeAAAA,
			TTL:   300,
			Value: "::",
		}}
	}

	return resp
}
