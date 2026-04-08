package pipeline

import (
	"context"

	"github.com/mantis-dns/mantis/internal/domain"
)

// QueryHandler processes a DNS query and optionally delegates to the next handler.
type QueryHandler interface {
	Handle(ctx context.Context, q *domain.Query, next QueryHandler) (*domain.Response, error)
}

// terminalHandler is the end of the chain — returns SERVFAIL.
type terminalHandler struct{}

func (t *terminalHandler) Handle(_ context.Context, _ *domain.Query, _ QueryHandler) (*domain.Response, error) {
	return nil, domain.ErrInternal
}
