package resolver

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	mdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// Forwarder resolves DNS queries by forwarding to upstream servers in parallel.
type Forwarder struct {
	upstreams []upstream
	logger    zerolog.Logger
}

type upstream struct {
	address string
	circuit *CircuitBreaker
}

type forwardResult struct {
	response *domain.Response
	upstream string
	err      error
}

// NewForwarder creates a forwarder with the given upstream addresses.
func NewForwarder(addrs []string, logger zerolog.Logger) *Forwarder {
	ups := make([]upstream, len(addrs))
	for i, addr := range addrs {
		ups[i] = upstream{
			address: addr,
			circuit: NewCircuitBreaker(5, 30*time.Second),
		}
	}
	return &Forwarder{
		upstreams: ups,
		logger:    logger.With().Str("component", "forwarder").Logger(),
	}
}

// Resolve queries all available upstreams in parallel and returns the fastest response.
func (f *Forwarder) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	results := make(chan forwardResult, len(f.upstreams))
	sent := 0

	for i := range f.upstreams {
		u := &f.upstreams[i]
		if !u.circuit.Allow() {
			continue
		}
		sent++
		go func(u *upstream) {
			resp, err := f.queryUpstream(ctx, u.address, q)
			if err != nil {
				u.circuit.RecordFailure()
				results <- forwardResult{err: err, upstream: u.address}
			} else {
				u.circuit.RecordSuccess()
				results <- forwardResult{response: resp, upstream: u.address}
			}
		}(u)
	}

	if sent == 0 {
		return nil, fmt.Errorf("all upstreams unavailable")
	}

	var lastErr error
	for range sent {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("upstream timeout: %w", ctx.Err())
		case r := <-results:
			if r.err != nil {
				lastErr = r.err
				continue
			}
			r.response.Upstream = r.upstream
			return r.response, nil
		}
	}

	return nil, fmt.Errorf("all upstreams failed: %w", lastErr)
}

func (f *Forwarder) queryUpstream(ctx context.Context, addr string, q *domain.Query) (*domain.Response, error) {
	msg := dnsutil.SetQuestion(new(mdns.Msg), dnsutil.Fqdn(q.Domain), q.Type)
	if msg == nil {
		return nil, fmt.Errorf("unsupported query type %d", q.Type)
	}

	c := new(mdns.Client)

	start := time.Now()
	resp, _, err := c.Exchange(ctx, msg, "udp", addr+":53")
	latency := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("exchange with %s: %w", addr, err)
	}

	return msgToResponse(resp, latency), nil
}

// msgToResponse converts a dns.Msg response to a domain.Response.
func msgToResponse(msg *mdns.Msg, latency time.Duration) *domain.Response {
	var answers []domain.RR
	for _, rr := range msg.Answer {
		hdr := rr.Header()
		dr := domain.RR{
			Name: hdr.Name,
			Type: mdns.RRToType(rr),
			TTL:  hdr.TTL,
		}

		switch v := rr.(type) {
		case *mdns.A:
			dr.Value = v.Addr.String()
		case *mdns.AAAA:
			addr := v.Addr
			if addr == (netip.Addr{}) {
				dr.Value = "::"
			} else {
				dr.Value = addr.String()
			}
		case *mdns.CNAME:
			dr.Value = v.Target
		case *mdns.MX:
			dr.Value = v.Mx
		case *mdns.TXT:
			if len(v.Txt) > 0 {
				dr.Value = v.Txt[0]
			}
		case *mdns.NS:
			dr.Value = v.Ns
		default:
			dr.Value = rr.String()
		}
		answers = append(answers, dr)
	}

	return &domain.Response{
		Answers: answers,
		Latency: latency,
	}
}
