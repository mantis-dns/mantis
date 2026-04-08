package doh

import (
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/netip"

	mdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

const maxBodySize = 65535

// Handler processes DNS-over-HTTPS requests per RFC 8484.
type Handler struct {
	resolver domain.Resolver
	logger   zerolog.Logger
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var wireMsg []byte
	var err error

	switch r.Method {
	case http.MethodPost:
		if r.Header.Get("Content-Type") != "application/dns-message" {
			http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
			return
		}
		wireMsg, err = io.ReadAll(io.LimitReader(r.Body, maxBodySize))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

	case http.MethodGet:
		dnsParam := r.URL.Query().Get("dns")
		if dnsParam == "" {
			http.Error(w, "missing dns parameter", http.StatusBadRequest)
			return
		}
		wireMsg, err = base64.RawURLEncoding.DecodeString(dnsParam)
		if err != nil {
			http.Error(w, "invalid base64", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	msg := &mdns.Msg{Data: wireMsg}
	if err := msg.Unpack(); err != nil {
		http.Error(w, "invalid DNS message", http.StatusBadRequest)
		return
	}

	if len(msg.Question) == 0 {
		http.Error(w, "no question", http.StatusBadRequest)
		return
	}

	qName, qType := dnsutil.Question(msg)
	clientIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	q := &domain.Query{
		Domain:    dnsutil.Fqdn(qName),
		Type:      qType,
		ClientIP:  net.ParseIP(clientIP),
		Transport: domain.TransportDoH,
	}

	resp, err := h.resolver.Resolve(r.Context(), q)
	if err != nil {
		h.logger.Warn().Err(err).Str("domain", q.Domain).Msg("DoH resolve failed")
		servfail := new(mdns.Msg)
		dnsutil.SetReply(servfail, msg)
		servfail.Rcode = mdns.RcodeServerFailure
		servfail.Pack()
		w.Header().Set("Content-Type", "application/dns-message")
		w.Write(servfail.Data)
		return
	}

	reply := new(mdns.Msg)
	dnsutil.SetReply(reply, msg)
	reply.RecursionAvailable = true

	for _, rr := range resp.Answers {
		dnsRR := domainRRtoDNS(rr, msg.Question[0].Header().Name)
		if dnsRR != nil {
			reply.Answer = append(reply.Answer, dnsRR)
		}
	}

	reply.Pack()
	w.Header().Set("Content-Type", "application/dns-message")
	w.Header().Set("Cache-Control", "max-age=300")
	w.Write(reply.Data)
}

// domainRRtoDNS converts domain.RR to dns.RR (same logic as internal/dns/writer.go).
func domainRRtoDNS(rr domain.RR, name string) mdns.RR {
	hdr := mdns.Header{
		Name:  name,
		Class: mdns.ClassINET,
		TTL:   rr.TTL,
	}

	switch rr.Type {
	case mdns.TypeA:
		a := new(mdns.A)
		*a.Header() = hdr
		if addr, err := parseAddr(rr.Value); err == nil {
			a.Addr = addr
		}
		return a
	case mdns.TypeAAAA:
		aaaa := new(mdns.AAAA)
		*aaaa.Header() = hdr
		if addr, err := parseAddr(rr.Value); err == nil {
			aaaa.Addr = addr
		}
		return aaaa
	case mdns.TypeCNAME:
		cname := new(mdns.CNAME)
		*cname.Header() = hdr
		cname.Target = dnsutil.Fqdn(rr.Value)
		return cname
	default:
		return nil
	}
}

func parseAddr(s string) (netip.Addr, error) {
	return netip.ParseAddr(s)
}
