package gravity

import (
	"sync"
	"testing"
)

func TestReverseDomain(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"ads.example.com", "com.example.ads."},
		{"example.com.", "com.example."},
		{"com", "com."},
		{"a.b.c.d.e", "e.d.c.b.a."},
	}
	for _, tt := range tests {
		got := ReverseDomain(tt.input)
		if got != tt.want {
			t.Errorf("ReverseDomain(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestExactMatch(t *testing.T) {
	e := NewEngine()
	e.RebuildFromDomains([]string{"ads.example.com", "tracker.net"})

	if !e.IsBlocked("ads.example.com") {
		t.Error("ads.example.com should be blocked")
	}
	if !e.IsBlocked("tracker.net") {
		t.Error("tracker.net should be blocked")
	}
	if e.IsBlocked("example.com") {
		t.Error("example.com should not be blocked")
	}
	if e.IsBlocked("google.com") {
		t.Error("google.com should not be blocked")
	}
}

func TestWildcardMatch(t *testing.T) {
	e := NewEngine()
	// Blocking "ads.example.com" should also block subdomains.
	e.RebuildFromDomains([]string{"ads.example.com"})

	if !e.IsBlocked("sub.ads.example.com") {
		t.Error("sub.ads.example.com should match wildcard")
	}
	if !e.IsBlocked("deep.sub.ads.example.com") {
		t.Error("deep.sub.ads.example.com should match wildcard")
	}
}

func TestAllowlistOverride(t *testing.T) {
	e := NewEngine()
	e.RebuildFromDomains([]string{"example.com", "ads.example.com"})
	e.SetAllowRules([]string{"ads.example.com"})

	if !e.IsBlocked("example.com") {
		t.Error("example.com should be blocked")
	}
	if e.IsBlocked("ads.example.com") {
		t.Error("ads.example.com should be allowed (allowlist override)")
	}
}

func TestAddRemoveRules(t *testing.T) {
	e := NewEngine()

	e.AddBlockRule("test.com")
	if !e.IsBlocked("test.com") {
		t.Error("test.com should be blocked after AddBlockRule")
	}

	e.RemoveBlockRule("test.com")
	if e.IsBlocked("test.com") {
		t.Error("test.com should not be blocked after RemoveBlockRule")
	}

	e.AddBlockRule("blocked.com")
	e.AddAllowRule("blocked.com")
	if e.IsBlocked("blocked.com") {
		t.Error("blocked.com should be allowed (allowlist takes precedence)")
	}
}

func TestConcurrentReads(t *testing.T) {
	e := NewEngine()
	domains := make([]string, 10000)
	for i := range domains {
		domains[i] = "domain" + string(rune('a'+i%26)) + ".com"
	}
	e.RebuildFromDomains(domains)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range 100 {
				e.IsBlocked("domain" + string(rune('a'+i%26)) + ".com")
			}
		}()
	}

	// Rebuild while reads are happening.
	wg.Add(1)
	go func() {
		defer wg.Done()
		e.RebuildFromDomains(domains[:5000])
	}()

	wg.Wait()
}

func TestBlockedCount(t *testing.T) {
	e := NewEngine()
	e.RebuildFromDomains([]string{"a.com", "b.com", "c.com"})
	if e.BlockedCount() != 3 {
		t.Errorf("expected 3, got %d", e.BlockedCount())
	}
}
