package domain

import "time"

// Client represents a network device seen by Mantis.
type Client struct {
	IP          string    `json:"ip"`
	MAC         string    `json:"mac,omitempty"`
	Hostname    string    `json:"hostname,omitempty"`
	QueryCount  int64     `json:"queryCount"`
	BlockedCount int64   `json:"blockedCount"`
	LastSeen    time.Time `json:"lastSeen"`
	BlockingEnabled bool `json:"blockingEnabled"`
}
