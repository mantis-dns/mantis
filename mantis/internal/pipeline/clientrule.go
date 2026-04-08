package pipeline

import (
	"context"

	"github.com/mantis-dns/mantis/internal/domain"
)

// ClientRuleHandler applies per-client blocking rules.
// For now, passes through to the next handler. Per-client toggle implemented in API layer.
type ClientRuleHandler struct{}

// NewClientRuleHandler creates a client rule handler.
func NewClientRuleHandler() *ClientRuleHandler {
	return &ClientRuleHandler{}
}

func (h *ClientRuleHandler) Handle(ctx context.Context, q *domain.Query, next QueryHandler) (*domain.Response, error) {
	return next.Handle(ctx, q, nil)
}
