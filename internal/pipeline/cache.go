package pipeline

import (
	"context"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/resolver"
)

// CacheHandler checks the DNS cache before delegating to the next handler.
type CacheHandler struct {
	cache *resolver.DNSCache
}

// NewCacheHandler creates a cache handler.
func NewCacheHandler(cache *resolver.DNSCache) *CacheHandler {
	return &CacheHandler{cache: cache}
}

func (h *CacheHandler) Handle(ctx context.Context, q *domain.Query, next QueryHandler) (*domain.Response, error) {
	if resp := h.cache.Get(q.Domain, q.Type); resp != nil {
		return resp, nil
	}

	resp, err := next.Handle(ctx, q, nil)
	if err != nil {
		return nil, err
	}

	ttl := extractTTL(resp)
	if ttl > 0 {
		h.cache.Set(q.Domain, q.Type, resp, ttl)
	}

	return resp, nil
}

func extractTTL(resp *domain.Response) time.Duration {
	if len(resp.Answers) == 0 {
		return 0
	}
	minTTL := resp.Answers[0].TTL
	for _, rr := range resp.Answers[1:] {
		if rr.TTL < minTTL {
			minTTL = rr.TTL
		}
	}
	return time.Duration(minTTL) * time.Second
}
