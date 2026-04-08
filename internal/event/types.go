package event

import "time"

// QueryEvent is emitted after every DNS resolution.
type QueryEvent struct {
	Timestamp time.Time `json:"timestamp"`
	ClientIP  string    `json:"clientIp"`
	Domain    string    `json:"domain"`
	QueryType uint16    `json:"queryType"`
	Result    string    `json:"result"` // allowed, blocked, cached, error
	Upstream  string    `json:"upstream,omitempty"`
	LatencyUs int64     `json:"latencyUs"`
	Answer    string    `json:"answer,omitempty"`
}
