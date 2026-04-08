package domain

import (
	"context"
	"net"
	"time"
)

// Transport represents the DNS transport protocol.
type Transport int

const (
	TransportUDP Transport = iota
	TransportTCP
	TransportDoH
	TransportDoT
)

func (t Transport) String() string {
	switch t {
	case TransportUDP:
		return "udp"
	case TransportTCP:
		return "tcp"
	case TransportDoH:
		return "doh"
	case TransportDoT:
		return "dot"
	default:
		return "unknown"
	}
}

// Common DNS query type constants.
const (
	QTypeA     uint16 = 1
	QTypeNS    uint16 = 2
	QTypeCNAME uint16 = 5
	QTypeSOA   uint16 = 6
	QTypeMX    uint16 = 15
	QTypeTXT   uint16 = 16
	QTypeAAAA  uint16 = 28
	QTypeSRV   uint16 = 33
	QTypeANY   uint16 = 255
)

// Query represents an incoming DNS query.
type Query struct {
	Domain    string    `json:"domain"`
	Type      uint16    `json:"type"`
	ClientIP  net.IP    `json:"clientIp"`
	Transport Transport `json:"transport"`
}

// Response represents the result of DNS resolution.
type Response struct {
	Answers  []RR          `json:"-"`
	Blocked  bool          `json:"blocked"`
	Cached   bool          `json:"cached"`
	Upstream string        `json:"upstream,omitempty"`
	Latency  time.Duration `json:"latency"`
}

// RR is a simplified DNS resource record for domain-layer use.
type RR struct {
	Name  string `json:"name"`
	Type  uint16 `json:"type"`
	TTL   uint32 `json:"ttl"`
	Value string `json:"value"`
}

// QueryLogEntry represents a persisted DNS query record.
type QueryLogEntry struct {
	ID        uint64    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	ClientIP  string    `json:"clientIp"`
	Domain    string    `json:"domain"`
	QueryType uint16   `json:"queryType"`
	Result    string    `json:"result"`
	Upstream  string    `json:"upstream,omitempty"`
	LatencyUs int64    `json:"latencyUs"`
	Answer    string    `json:"answer,omitempty"`
}

// QueryLogFilter defines filter criteria for query log queries.
type QueryLogFilter struct {
	Domain   string    `json:"domain,omitempty"`
	ClientIP string    `json:"clientIp,omitempty"`
	Result   string    `json:"result,omitempty"`
	From     time.Time `json:"from,omitempty"`
	To       time.Time `json:"to,omitempty"`
	Page     int       `json:"page"`
	PerPage  int       `json:"perPage"`
}

// Resolver resolves DNS queries.
type Resolver interface {
	Resolve(ctx context.Context, query *Query) (*Response, error)
}
