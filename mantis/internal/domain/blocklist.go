package domain

import "time"

// BlocklistFormat represents the format of a blocklist file.
type BlocklistFormat string

const (
	FormatHosts   BlocklistFormat = "hosts"
	FormatDomains BlocklistFormat = "domains"
	FormatAdblock BlocklistFormat = "adblock"
)

// BlocklistStatus represents the last download status.
type BlocklistStatus string

const (
	StatusSuccess BlocklistStatus = "success"
	StatusError   BlocklistStatus = "error"
	StatusPending BlocklistStatus = "pending"
)

// BlocklistSource represents a remote blocklist.
type BlocklistSource struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	URL         string          `json:"url"`
	Enabled     bool            `json:"enabled"`
	Format      BlocklistFormat `json:"format"`
	DomainCount int             `json:"domainCount"`
	LastUpdated time.Time       `json:"lastUpdated,omitempty"`
	LastStatus  BlocklistStatus `json:"lastStatus"`
}

// RuleType represents a custom rule type.
type RuleType string

const (
	RuleBlock RuleType = "block"
	RuleAllow RuleType = "allow"
)

// CustomRule represents a user-defined allow or block rule.
type CustomRule struct {
	ID      string   `json:"id"`
	Domain  string   `json:"domain"`
	Type    RuleType `json:"type"`
	Comment string   `json:"comment,omitempty"`
	Created time.Time `json:"created"`
}
