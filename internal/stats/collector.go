package stats

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/pebble/v2"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/event"
)

// Collector listens to the event bus and maintains statistics.
type Collector struct {
	db      *pebble.DB
	bus     *event.Bus
	current *domain.StatBucket
	mu      sync.Mutex
	done    chan struct{}
}

// NewCollector creates a stats collector subscribed to the event bus.
func NewCollector(db *pebble.DB, bus *event.Bus) *Collector {
	return &Collector{
		db:   db,
		bus:  bus,
		current: &domain.StatBucket{
			Timestamp:  truncateHour(time.Now()),
			TopDomains: make(map[string]int64),
			TopClients: make(map[string]int64),
		},
		done: make(chan struct{}),
	}
}

// Start begins collecting stats from the event bus.
func (c *Collector) Start() {
	ch := c.bus.Subscribe(10000)
	go c.collectLoop(ch)
}

// Stop flushes current bucket and stops collecting.
func (c *Collector) Stop() {
	close(c.done)
	c.mu.Lock()
	c.flushBucket()
	c.mu.Unlock()
}

func (c *Collector) collectLoop(ch <-chan event.QueryEvent) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case ev, ok := <-ch:
			if !ok {
				return
			}
			c.record(ev)
		case <-ticker.C:
			c.mu.Lock()
			c.checkHourRollover()
			c.mu.Unlock()
		case <-c.done:
			return
		}
	}
}

func (c *Collector) record(ev event.QueryEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.checkHourRollover()

	c.current.TotalQueries++
	if ev.Result == "blocked" {
		c.current.BlockedCount++
	}
	if ev.Result == "cached" {
		c.current.CachedCount++
	}

	c.current.TopDomains[ev.Domain]++
	c.current.TopClients[ev.ClientIP]++

	// Cap top maps at 200 entries to prevent unbounded growth.
	if len(c.current.TopDomains) > 200 {
		c.pruneMap(c.current.TopDomains, 100)
	}
	if len(c.current.TopClients) > 200 {
		c.pruneMap(c.current.TopClients, 100)
	}
}

func (c *Collector) checkHourRollover() {
	now := truncateHour(time.Now())
	if !now.Equal(c.current.Timestamp) {
		c.flushBucket()
		c.current = &domain.StatBucket{
			Timestamp:  now,
			TopDomains: make(map[string]int64),
			TopClients: make(map[string]int64),
		}
	}
}

func (c *Collector) flushBucket() {
	if c.current.TotalQueries == 0 {
		return
	}
	key := []byte(fmt.Sprintf("stat:hour:%020d", c.current.Timestamp.Unix()))
	data, err := json.Marshal(c.current)
	if err != nil {
		return
	}
	c.db.Set(key, data, pebble.Sync)
}

func (c *Collector) pruneMap(m map[string]int64, keep int) {
	if len(m) <= keep {
		return
	}
	// Simple pruning: remove entries with lowest counts.
	type entry struct {
		key   string
		count int64
	}
	entries := make([]entry, 0, len(m))
	for k, v := range m {
		entries = append(entries, entry{k, v})
	}
	// Sort by count descending (simple bubble for small sets).
	for i := range entries {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].count > entries[i].count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	for k := range m {
		delete(m, k)
	}
	for i := 0; i < keep && i < len(entries); i++ {
		m[entries[i].key] = entries[i].count
	}
}

func truncateHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}
