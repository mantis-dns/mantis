package pipeline

import (
	"context"

	"github.com/mantis-dns/mantis/internal/domain"
)

// Chain links QueryHandlers and implements domain.Resolver.
type Chain struct {
	handlers []QueryHandler
}

// NewChain creates a pipeline from the given handlers.
func NewChain(handlers ...QueryHandler) *Chain {
	return &Chain{handlers: handlers}
}

// Resolve starts the chain.
func (c *Chain) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
	if len(c.handlers) == 0 {
		return nil, domain.ErrInternal
	}
	return c.handlers[0].Handle(ctx, q, &chainLink{handlers: c.handlers, index: 1})
}

type chainLink struct {
	handlers []QueryHandler
	index    int
}

func (cl *chainLink) Handle(ctx context.Context, q *domain.Query, _ QueryHandler) (*domain.Response, error) {
	if cl.index >= len(cl.handlers) {
		return (&terminalHandler{}).Handle(ctx, q, nil)
	}
	next := &chainLink{handlers: cl.handlers, index: cl.index + 1}
	return cl.handlers[cl.index].Handle(ctx, q, next)
}
