package resolver

import (
	"container/list"
	"strings"
	"sync"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
)

// DNSCache is a thread-safe LRU cache with TTL-aware eviction.
type DNSCache struct {
	entries map[string]*cacheEntry
	order   *list.List
	maxSize int
	mu      sync.RWMutex
	done    chan struct{}
}

type cacheEntry struct {
	key       string
	response  *domain.Response
	expiresAt time.Time
	element   *list.Element
}

// NewDNSCache creates a cache with the given max entries and starts a periodic sweep.
func NewDNSCache(maxSize int) *DNSCache {
	c := &DNSCache{
		entries: make(map[string]*cacheEntry),
		order:   list.New(),
		maxSize: maxSize,
		done:    make(chan struct{}),
	}
	go c.sweepLoop()
	return c
}

// Close stops the periodic sweep.
func (c *DNSCache) Close() {
	close(c.done)
}

func cacheKey(domainName string, qtype uint16) string {
	return strings.ToLower(domainName) + ":" + qtypeStr(qtype)
}

func qtypeStr(t uint16) string {
	// Simple int to string without fmt to avoid allocation.
	buf := [5]byte{}
	i := len(buf)
	for t > 0 {
		i--
		buf[i] = byte('0' + t%10)
		t /= 10
	}
	if i == len(buf) {
		return "0"
	}
	return string(buf[i:])
}

// Get retrieves a cached response. Returns nil on miss or expired entry.
func (c *DNSCache) Get(domainName string, qtype uint16) *domain.Response {
	key := cacheKey(domainName, qtype)

	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok {
		return nil
	}

	if time.Now().After(entry.expiresAt) {
		c.mu.Lock()
		c.removeLocked(key)
		c.mu.Unlock()
		return nil
	}

	c.mu.Lock()
	c.order.MoveToFront(entry.element)
	c.mu.Unlock()

	resp := *entry.response
	resp.Cached = true
	return &resp
}

// Set stores a response in the cache with the given TTL.
func (c *DNSCache) Set(domainName string, qtype uint16, resp *domain.Response, ttl time.Duration) {
	if ttl <= 0 || c.maxSize <= 0 {
		return
	}

	key := cacheKey(domainName, qtype)

	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.entries[key]; ok {
		c.order.Remove(existing.element)
		delete(c.entries, key)
	}

	for len(c.entries) >= c.maxSize {
		c.evictLRU()
	}

	entry := &cacheEntry{
		key:       key,
		response:  resp,
		expiresAt: time.Now().Add(ttl),
	}
	entry.element = c.order.PushFront(entry)
	c.entries[key] = entry
}

// Size returns the number of entries in the cache.
func (c *DNSCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *DNSCache) removeLocked(key string) {
	if entry, ok := c.entries[key]; ok {
		c.order.Remove(entry.element)
		delete(c.entries, key)
	}
}

func (c *DNSCache) evictLRU() {
	back := c.order.Back()
	if back == nil {
		return
	}
	entry := back.Value.(*cacheEntry)
	c.order.Remove(back)
	delete(c.entries, entry.key)
}

func (c *DNSCache) sweepLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.sweep()
		case <-c.done:
			return
		}
	}
}

func (c *DNSCache) sweep() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			c.order.Remove(entry.element)
			delete(c.entries, key)
		}
	}
}
