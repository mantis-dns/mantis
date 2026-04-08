package pipeline

import (
	"context"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/event"
)

// Chain links QueryHandlers and implements domain.Resolver.
type Chain struct {
	handlers []QueryHandler
	bus      *event.Bus
}

// NewChain creates a pipeline from the given handlers.
func NewChain(bus *event.Bus, handlers ...QueryHandler) *Chain {
	return &Chain{handlers: handlers, bus: bus}
}

// Resolve starts the chain and publishes a query event.
func (c *Chain) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
	if len(c.handlers) == 0 {
		return nil, domain.ErrInternal
	}

	start := time.Now()
	resp, err := c.handlers[0].Handle(ctx, q, &chainLink{handlers: c.handlers, index: 1})
	latency := time.Since(start)

	if c.bus != nil {
		ev := event.QueryEvent{
			Timestamp: start,
			ClientIP:  q.ClientIP.String(),
			Domain:    q.Domain,
			QueryType: q.Type,
			LatencyUs: latency.Microseconds(),
		}

		if err != nil {
			ev.Result = "error"
		} else if resp.Blocked {
			ev.Result = "blocked"
		} else if resp.Cached {
			ev.Result = "cached"
		} else {
			ev.Result = "allowed"
		}

		if resp != nil {
			ev.Upstream = resp.Upstream
			if len(resp.Answers) > 0 {
				ev.Answer = resp.Answers[0].Value
			}
		}

		c.bus.Publish(ev)
	}

	return resp, err
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
