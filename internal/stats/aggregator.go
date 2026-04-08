package stats

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
)

// Aggregator reads persisted stat buckets and computes summaries.
type Aggregator struct {
	db        *pebble.DB
	collector *Collector
}

// NewAggregator creates a stats aggregator.
func NewAggregator(db *pebble.DB, collector *Collector) *Aggregator {
	return &Aggregator{db: db, collector: collector}
}

// Summary returns aggregated stats for the given duration.
func (a *Aggregator) Summary(period time.Duration) domain.StatsSummary {
	buckets := a.loadBuckets(time.Now().Add(-period), time.Now())

	// Include current in-memory bucket.
	a.collector.mu.Lock()
	current := *a.collector.current
	a.collector.mu.Unlock()
	buckets = append(buckets, current)

	var summary domain.StatsSummary
	for _, b := range buckets {
		summary.TotalQueries += b.TotalQueries
		summary.BlockedCount += b.BlockedCount
		summary.CachedCount += b.CachedCount
	}

	if summary.TotalQueries > 0 {
		summary.BlockedPercent = float64(summary.BlockedCount) / float64(summary.TotalQueries) * 100
		summary.CacheHitRatio = float64(summary.CachedCount) / float64(summary.TotalQueries) * 100
	}

	return summary
}

// Overtime returns time-series data points for the given duration.
func (a *Aggregator) Overtime(period time.Duration) []domain.StatsPoint {
	buckets := a.loadBuckets(time.Now().Add(-period), time.Now())

	points := make([]domain.StatsPoint, 0, len(buckets))
	for _, b := range buckets {
		points = append(points, domain.StatsPoint{
			Timestamp:    b.Timestamp,
			TotalQueries: b.TotalQueries,
			BlockedCount: b.BlockedCount,
			CachedCount:  b.CachedCount,
		})
	}
	return points
}

// TopDomains returns the top N domains from recent buckets.
func (a *Aggregator) TopDomains(period time.Duration, n int) (allowed, blocked []domain.TopDomain) {
	buckets := a.loadBuckets(time.Now().Add(-period), time.Now())

	a.collector.mu.Lock()
	current := *a.collector.current
	a.collector.mu.Unlock()
	buckets = append(buckets, current)

	merged := make(map[string]int64)
	for _, b := range buckets {
		for d, c := range b.TopDomains {
			merged[d] += c
		}
	}

	type entry struct {
		domain string
		count  int64
	}
	var entries []entry
	for d, c := range merged {
		entries = append(entries, entry{d, c})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].count > entries[j].count })

	for i := 0; i < n && i < len(entries); i++ {
		allowed = append(allowed, domain.TopDomain{Domain: entries[i].domain, Count: entries[i].count})
	}
	return allowed, nil // blocked breakdown requires per-result tracking (future)
}

// TopClients returns the top N clients from recent buckets.
func (a *Aggregator) TopClients(period time.Duration, n int) []domain.TopClient {
	buckets := a.loadBuckets(time.Now().Add(-period), time.Now())

	a.collector.mu.Lock()
	current := *a.collector.current
	a.collector.mu.Unlock()
	buckets = append(buckets, current)

	merged := make(map[string]int64)
	for _, b := range buckets {
		for c, cnt := range b.TopClients {
			merged[c] += cnt
		}
	}

	type entry struct {
		ip    string
		count int64
	}
	var entries []entry
	for ip, c := range merged {
		entries = append(entries, entry{ip, c})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].count > entries[j].count })

	var result []domain.TopClient
	for i := 0; i < n && i < len(entries); i++ {
		result = append(result, domain.TopClient{IP: entries[i].ip, Count: entries[i].count})
	}
	return result
}

func (a *Aggregator) loadBuckets(from, to time.Time) []domain.StatBucket {
	lower := []byte(fmt.Sprintf("stat:hour:%020d", from.Unix()))
	upper := []byte(fmt.Sprintf("stat:hour:%020d", to.Unix()+1))

	iter, err := a.db.NewIter(&pebble.IterOptions{
		LowerBound: lower,
		UpperBound: upper,
	})
	if err != nil {
		return nil
	}
	defer iter.Close()

	var buckets []domain.StatBucket
	for iter.First(); iter.Valid(); iter.Next() {
		var b domain.StatBucket
		if err := json.Unmarshal(iter.Value(), &b); err != nil {
			continue
		}
		buckets = append(buckets, b)
	}
	return buckets
}
