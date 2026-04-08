package dot

import (
	"net/netip"

	mdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/mantis-dns/mantis/internal/domain"
)

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
		if addr, err := netip.ParseAddr(rr.Value); err == nil {
			a.Addr = addr
		}
		return a
	case mdns.TypeAAAA:
		aaaa := new(mdns.AAAA)
		*aaaa.Header() = hdr
		if addr, err := netip.ParseAddr(rr.Value); err == nil {
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
