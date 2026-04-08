package domain

import "time"

// DhcpLease represents a DHCP lease assignment.
type DhcpLease struct {
	MAC      string    `json:"mac"`
	IP       string    `json:"ip"`
	Hostname string    `json:"hostname,omitempty"`
	LeaseEnd time.Time `json:"leaseEnd"`
	IsStatic bool      `json:"isStatic"`
}
