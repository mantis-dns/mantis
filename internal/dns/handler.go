package dns

import (
	"context"
	"net"

	mdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// Handler implements dns.Handler by converting DNS messages to domain types.
type Handler struct {
	resolver domain.Resolver
	logger   zerolog.Logger
}

// ServeDNS handles an incoming DNS query.
func (h *Handler) ServeDNS(ctx context.Context, w mdns.ResponseWriter, r *mdns.Msg) {
	if len(r.Question) == 0 {
		writeServfail(w, r)
		return
	}

	q := queryFromMsg(r, w.RemoteAddr())

	resp, err := h.resolver.Resolve(ctx, q)
	if err != nil {
		h.logger.Warn().Err(err).Str("domain", q.Domain).Msg("resolve failed")
		writeServfail(w, r)
		return
	}

	writeResponse(w, r, resp)
}

// queryFromMsg converts a dns.Msg into a domain.Query.
func queryFromMsg(r *mdns.Msg, remoteAddr net.Addr) *domain.Query {
	qName, qType := dnsutil.Question(r)

	var clientIP net.IP
	switch addr := remoteAddr.(type) {
	case *net.UDPAddr:
		clientIP = addr.IP
	case *net.TCPAddr:
		clientIP = addr.IP
	}

	transport := domain.TransportUDP
	if _, ok := remoteAddr.(*net.TCPAddr); ok {
		transport = domain.TransportTCP
	}

	return &domain.Query{
		Domain:    dnsutil.Fqdn(qName),
		Type:      qType,
		ClientIP:  clientIP,
		Transport: transport,
	}
}
