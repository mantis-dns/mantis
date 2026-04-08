package dhcp

import (
	"fmt"
	"net"
	"sync"

	"github.com/mantis-dns/mantis/internal/domain"
)

// Pool manages a range of IPv4 addresses for DHCP assignment.
type Pool struct {
	start    net.IP
	end      net.IP
	assigned map[string]string // IP -> MAC
	mu       sync.Mutex
}

// NewPool creates an IP address pool from start to end.
func NewPool(start, end string) (*Pool, error) {
	s := net.ParseIP(start).To4()
	e := net.ParseIP(end).To4()
	if s == nil || e == nil {
		return nil, fmt.Errorf("invalid IP range: %s - %s", start, end)
	}
	return &Pool{
		start:    s,
		end:      e,
		assigned: make(map[string]string),
	}, nil
}

// Allocate finds a free IP for the given MAC, or returns existing assignment.
func (p *Pool) Allocate(mac string) (net.IP, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if MAC already has an IP.
	for ip, m := range p.assigned {
		if m == mac {
			return net.ParseIP(ip).To4(), nil
		}
	}

	// Find first free IP.
	ip := make(net.IP, 4)
	copy(ip, p.start)
	for ; ipLessOrEqual(ip, p.end); incIP(ip) {
		ipStr := ip.String()
		if _, taken := p.assigned[ipStr]; !taken {
			p.assigned[ipStr] = mac
			result := make(net.IP, 4)
			copy(result, ip)
			return result, nil
		}
	}
	return nil, domain.ErrPoolExhausted
}

// Release frees an IP assignment.
func (p *Pool) Release(ip net.IP) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.assigned, ip.To4().String())
}

// Reserve marks an IP as assigned to a MAC (for loading existing leases).
func (p *Pool) Reserve(ip net.IP, mac string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.assigned[ip.To4().String()] = mac
}

func ipLessOrEqual(a, b net.IP) bool {
	for i := range 4 {
		if a[i] < b[i] {
			return true
		}
		if a[i] > b[i] {
			return false
		}
	}
	return true // equal
}

func incIP(ip net.IP) {
	for i := 3; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}
