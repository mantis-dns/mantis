package resolver

import (
	"context"
	"fmt"
	"net/netip"
	"strings"
	"time"

	mdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

const maxRecursionDepth = 30

// Root hints (IANA root servers).
var rootServers = []string{
	"198.41.0.4",     // a.root-servers.net
	"170.247.170.2",  // b.root-servers.net
	"192.33.4.12",    // c.root-servers.net
	"199.7.91.13",    // d.root-servers.net
	"192.203.230.10", // e.root-servers.net
	"192.5.5.241",    // f.root-servers.net
	"192.112.36.4",   // g.root-servers.net
	"198.97.190.53",  // h.root-servers.net
	"192.36.148.17",  // i.root-servers.net
	"192.58.128.30",  // j.root-servers.net
	"193.0.14.129",   // k.root-servers.net
	"199.7.83.42",    // l.root-servers.net
	"202.12.27.33",   // m.root-servers.net
}

// Recursive implements iterative DNS resolution from root hints.
type Recursive struct {
	cache  *DNSCache
	client *mdns.Client
	logger zerolog.Logger
}

// NewRecursive creates a recursive resolver.
func NewRecursive(cache *DNSCache, logger zerolog.Logger) *Recursive {
	return &Recursive{
		cache:  cache,
		client: &mdns.Client{},
		logger: logger.With().Str("component", "recursive").Logger(),
	}
}

// Resolve performs iterative resolution starting from root servers.
func (r *Recursive) Resolve(ctx context.Context, q *domain.Query) (*domain.Response, error) {
	start := time.Now()
	resp, err := r.resolve(ctx, q.Domain, q.Type, rootServers, 0)
	if err != nil {
		return nil, err
	}
	resp.Latency = time.Since(start)
	resp.Upstream = "recursive"
	return resp, nil
}

func (r *Recursive) resolve(ctx context.Context, name string, qtype uint16, servers []string, depth int) (*domain.Response, error) {
	if depth > maxRecursionDepth {
		return nil, fmt.Errorf("max recursion depth exceeded for %s", name)
	}

	fqdn := dnsutil.Fqdn(name)

	// Check cache first.
	if cached := r.cache.Get(fqdn, qtype); cached != nil {
		return cached, nil
	}

	for _, server := range servers {
		resp, referral, err := r.query(ctx, fqdn, qtype, server)
		if err != nil {
			r.logger.Debug().Err(err).Str("server", server).Str("domain", fqdn).Msg("query failed")
			continue
		}

		// Got an answer.
		if resp != nil && len(resp.Answers) > 0 {
			// Follow CNAME if needed.
			if qtype != mdns.TypeCNAME {
				for _, ans := range resp.Answers {
					if ans.Type == mdns.TypeCNAME && ans.Value != "" {
						cnameResp, err := r.resolve(ctx, ans.Value, qtype, rootServers, depth+1)
						if err == nil && len(cnameResp.Answers) > 0 {
							// Prepend CNAME to answers.
							combined := append(resp.Answers, cnameResp.Answers...)
							resp.Answers = combined
						}
						break
					}
				}
			}
			// Cache the result.
			if ttl := minTTL(resp); ttl > 0 {
				r.cache.Set(fqdn, qtype, resp, time.Duration(ttl)*time.Second)
			}
			return resp, nil
		}

		// Got a referral -- follow NS.
		if len(referral) > 0 {
			result, err := r.resolve(ctx, name, qtype, referral, depth+1)
			if err == nil {
				return result, nil
			}
		}
	}

	return nil, fmt.Errorf("resolution failed for %s after trying all servers", name)
}

func (r *Recursive) query(ctx context.Context, fqdn string, qtype uint16, server string) (*domain.Response, []string, error) {
	msg := dnsutil.SetQuestion(new(mdns.Msg), fqdn, qtype)
	if msg == nil {
		return nil, nil, fmt.Errorf("unsupported query type %d", qtype)
	}

	addr := server + ":53"
	resp, _, err := r.client.Exchange(ctx, msg, "udp", addr)
	if err != nil {
		return nil, nil, err
	}

	// Check for answers.
	if len(resp.Answer) > 0 {
		return msgToResponse(resp, 0), nil, nil
	}

	// Check for NS referrals in authority section.
	var referrals []string
	glue := make(map[string]string) // NS name -> IP from additional section

	for _, rr := range resp.Extra {
		if a, ok := rr.(*mdns.A); ok {
			glue[strings.ToLower(rr.Header().Name)] = a.Addr.String()
		}
	}

	for _, rr := range resp.Ns {
		if ns, ok := rr.(*mdns.NS); ok {
			nsName := strings.ToLower(ns.Ns)
			if ip, ok := glue[nsName]; ok {
				referrals = append(referrals, ip)
			}
		}
	}

	// If no glue records, try to resolve NS names.
	if len(referrals) == 0 {
		for _, rr := range resp.Ns {
			if ns, ok := rr.(*mdns.NS); ok {
				nsResp, err := r.resolve(context.Background(), ns.Ns, mdns.TypeA, rootServers, 0)
				if err == nil && len(nsResp.Answers) > 0 {
					for _, ans := range nsResp.Answers {
						if ans.Type == mdns.TypeA {
							addr, err := netip.ParseAddr(ans.Value)
							if err == nil {
								referrals = append(referrals, addr.String())
							}
						}
					}
				}
				if len(referrals) > 0 {
					break
				}
			}
		}
	}

	return nil, referrals, nil
}

func minTTL(resp *domain.Response) uint32 {
	if len(resp.Answers) == 0 {
		return 0
	}
	min := resp.Answers[0].TTL
	for _, rr := range resp.Answers[1:] {
		if rr.TTL < min {
			min = rr.TTL
		}
	}
	return min
}

// msgToResponse is defined in forwarder.go -- reuse it.
// This file only adds the Recursive resolver type.
