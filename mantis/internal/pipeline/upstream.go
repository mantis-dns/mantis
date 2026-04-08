package pipeline

import (
	"context"

	"github.com/mantis-dns/mantis/internal/domain"
)

// UpstreamHandler delegates to the configured resolver (forwarder or recursive).
type UpstreamHandler struct {
	resolver domain.Resolver
}

// NewUpstreamHandler creates an upstream handler.
func NewUpstreamHandler(resolver domain.Resolver) *UpstreamHandler {
	return &UpstreamHandler{resolver: resolver}
}

func (h *UpstreamHandler) Handle(ctx context.Context, q *domain.Query, _ QueryHandler) (*domain.Response, error) {
	return h.resolver.Resolve(ctx, q)
}
