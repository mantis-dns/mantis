package resolver

import (
	"sync"
	"testing"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
)

func makeResp(value string) *domain.Response {
	return &domain.Response{
		Answers: []domain.RR{{Name: "test", Type: 1, TTL: 60, Value: value}},
	}
}

func TestCacheHitMiss(t *testing.T) {
	c := NewDNSCache(100)
	defer c.Close()

	c.Set("example.com.", 1, makeResp("1.2.3.4"), 5*time.Minute)

	got := c.Get("example.com.", 1)
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if !got.Cached {
		t.Error("expected Cached=true")
	}

	miss := c.Get("other.com.", 1)
	if miss != nil {
		t.Error("expected cache miss")
	}
}

func TestCasInsensitive(t *testing.T) {
	c := NewDNSCache(100)
	defer c.Close()

	c.Set("Example.COM.", 1, makeResp("1.2.3.4"), time.Minute)
	if c.Get("example.com.", 1) == nil {
		t.Error("expected case-insensitive hit")
	}
}

func TestCacheTTLExpiry(t *testing.T) {
	c := NewDNSCache(100)
	defer c.Close()

	c.Set("expire.com.", 1, makeResp("1.2.3.4"), 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	if c.Get("expire.com.", 1) != nil {
		t.Error("expected expired entry to miss")
	}
	if c.Size() != 0 {
		t.Error("expired entry should be removed")
	}
}

func TestCacheLRUEviction(t *testing.T) {
	c := NewDNSCache(3)
	defer c.Close()

	c.Set("a.com.", 1, makeResp("1"), time.Hour)
	c.Set("b.com.", 1, makeResp("2"), time.Hour)
	c.Set("c.com.", 1, makeResp("3"), time.Hour)

	// Access a.com to make it recently used.
	c.Get("a.com.", 1)

	// Add d.com -- should evict b.com (LRU).
	c.Set("d.com.", 1, makeResp("4"), time.Hour)

	if c.Get("b.com.", 1) != nil {
		t.Error("b.com should have been evicted (LRU)")
	}
	if c.Get("a.com.", 1) == nil {
		t.Error("a.com should still be cached (recently accessed)")
	}
	if c.Get("d.com.", 1) == nil {
		t.Error("d.com should be cached")
	}
}

func TestCacheConcurrent(t *testing.T) {
	c := NewDNSCache(1000)
	defer c.Close()

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "domain" + qtypeStr(uint16(i)) + ".com."
			c.Set(key, 1, makeResp("1.2.3.4"), time.Hour)
			c.Get(key, 1)
		}(i)
	}
	wg.Wait()

	if c.Size() != 100 {
		t.Errorf("expected 100 entries, got %d", c.Size())
	}
}

func TestCacheZeroSize(t *testing.T) {
	c := NewDNSCache(0)
	defer c.Close()

	c.Set("test.com.", 1, makeResp("1.2.3.4"), time.Hour)
	if c.Get("test.com.", 1) != nil {
		t.Error("zero-size cache should not store")
	}
}
