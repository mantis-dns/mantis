package dns

import (
	"net/netip"

	mdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/mantis-dns/mantis/internal/domain"
)

// writeResponse constructs and sends a DNS response from a domain.Response.
func writeResponse(w mdns.ResponseWriter, r *mdns.Msg, resp *domain.Response) {
	msg := new(mdns.Msg)
	dnsutil.SetReply(msg, r)
	msg.RecursionAvailable = true

	qName := r.Question[0].Header().Name

	for _, rr := range resp.Answers {
		dnsRR := domainRRtoDNS(rr, qName)
		if dnsRR != nil {
			msg.Answer = append(msg.Answer, dnsRR)
		}
	}

	packAndWrite(w, msg)
}

// writeServfail sends a SERVFAIL response.
func writeServfail(w mdns.ResponseWriter, r *mdns.Msg) {
	msg := new(mdns.Msg)
	dnsutil.SetReply(msg, r)
	msg.Rcode = mdns.RcodeServerFailure
	msg.RecursionAvailable = true
	packAndWrite(w, msg)
}

func packAndWrite(w mdns.ResponseWriter, msg *mdns.Msg) {
	if err := msg.Pack(); err != nil {
		return
	}
	w.Write(msg.Data)
}

// domainRRtoDNS converts a domain.RR to a dns.RR.
func domainRRtoDNS(rr domain.RR, name string) mdns.RR {
	hdr := mdns.Header{
		Name:   name,
		Class:  mdns.ClassINET,
		TTL:    rr.TTL,
	}

	switch rr.Type {
	case mdns.TypeA:
		a := new(mdns.A)
		*a.Header() = hdr
		addr, _ := netip.ParseAddr(rr.Value)
		a.Addr = addr
		return a
	case mdns.TypeAAAA:
		aaaa := new(mdns.AAAA)
		*aaaa.Header() = hdr
		addr, _ := netip.ParseAddr(rr.Value)
		aaaa.Addr = addr
		return aaaa
	case mdns.TypeCNAME:
		cname := new(mdns.CNAME)
		*cname.Header() = hdr
		cname.Target = dnsutil.Fqdn(rr.Value)
		return cname
	case mdns.TypeMX:
		mx := new(mdns.MX)
		*mx.Header() = hdr
		mx.Mx = dnsutil.Fqdn(rr.Value)
		mx.Preference = 10
		return mx
	case mdns.TypeTXT:
		txt := new(mdns.TXT)
		*txt.Header() = hdr
		txt.Txt = []string{rr.Value}
		return txt
	case mdns.TypeNS:
		ns := new(mdns.NS)
		*ns.Header() = hdr
		ns.Ns = dnsutil.Fqdn(rr.Value)
		return ns
	default:
		return nil
	}
}
