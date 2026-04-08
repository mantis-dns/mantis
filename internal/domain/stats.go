package domain

import "time"

// StatsSummary holds aggregated statistics for a time period.
type StatsSummary struct {
	TotalQueries  int64   `json:"totalQueries"`
	BlockedCount  int64   `json:"blockedCount"`
	CachedCount   int64   `json:"cachedCount"`
	BlockedPercent float64 `json:"blockedPercent"`
	CacheHitRatio float64 `json:"cacheHitRatio"`
}

// StatsPoint represents a single data point in a time series.
type StatsPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	TotalQueries int64     `json:"totalQueries"`
	BlockedCount int64     `json:"blockedCount"`
	CachedCount  int64     `json:"cachedCount"`
}

// TopDomain represents a domain ranked by query count.
type TopDomain struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

// TopClient represents a client ranked by query volume.
type TopClient struct {
	IP       string `json:"ip"`
	Hostname string `json:"hostname,omitempty"`
	Count    int64  `json:"count"`
}

// StatBucket holds aggregated stats for a time bucket (hour/day).
type StatBucket struct {
	Timestamp    time.Time          `json:"timestamp"`
	TotalQueries int64              `json:"totalQueries"`
	BlockedCount int64              `json:"blockedCount"`
	CachedCount  int64              `json:"cachedCount"`
	TopDomains   map[string]int64   `json:"topDomains,omitempty"`
	TopClients   map[string]int64   `json:"topClients,omitempty"`
}
