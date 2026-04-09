package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
	"golang.org/x/time/rate"
)

// AuthMiddleware checks session cookie or API key.
func AuthMiddleware(sessions domain.SessionRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check session cookie.
			cookie, err := r.Cookie("session")
			if err == nil && cookie.Value != "" {
				session, err := sessions.GetSession(r.Context(), cookie.Value)
				if err == nil && session != nil {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Check API key header.
			apiKey := r.Header.Get("X-Api-Key")
			if apiKey != "" {
				hash := sha256.Sum256([]byte(apiKey))
				hashStr := hex.EncodeToString(hash[:])
				key, err := sessions.GetAPIKey(r.Context(), hashStr)
				if err == nil && key != nil {
					next.ServeHTTP(w, r)
					return
				}
			}

			Error(w, "UNAUTHORIZED", "authentication required", http.StatusUnauthorized)
		})
	}
}

// rateLimiterEntry tracks a limiter and its last access time.
type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware limits requests per client IP with TTL eviction.
func RateLimitMiddleware(rps int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	limiters := make(map[string]*rateLimiterEntry)

	// Evict stale entries every 5 minutes.
	go func() {
		for range time.Tick(5 * time.Minute) {
			mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for k, v := range limiters {
				if v.lastSeen.Before(cutoff) {
					delete(limiters, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := clientIP(r)

			mu.Lock()
			entry, ok := limiters[key]
			if !ok {
				// Cap total tracked IPs to prevent memory exhaustion.
				if len(limiters) >= 10000 {
					evictOldestLocked(limiters)
				}
				entry = &rateLimiterEntry{
					limiter: rate.NewLimiter(rate.Limit(rps)/60, rps),
				}
				limiters[key] = entry
			}
			entry.lastSeen = time.Now()
			mu.Unlock()

			if !entry.limiter.Allow() {
				w.Header().Set("Retry-After", "60")
				Error(w, "RATE_LIMITED", "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// evictOldestLocked removes the least-recently-seen entry from the map.
// Caller must hold mu.
func evictOldestLocked(limiters map[string]*rateLimiterEntry) {
	var oldestKey string
	var oldestTime time.Time
	first := true
	for k, v := range limiters {
		if first || v.lastSeen.Before(oldestTime) {
			oldestKey = k
			oldestTime = v.lastSeen
			first = false
		}
	}
	if oldestKey != "" {
		delete(limiters, oldestKey)
	}
}

// SecurityHeaders adds security headers to responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS with same-origin enforcement.
func CORSMiddleware(trustedHost string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && isSameOrigin(origin, trustedHost) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isSameOrigin(origin, trustedHost string) bool {
	// Strip scheme from origin.
	o := origin
	o = strings.TrimPrefix(o, "http://")
	o = strings.TrimPrefix(o, "https://")
	o = strings.TrimSuffix(o, "/")
	// Compare host portions (ignoring port differences).
	oHost, _, _ := net.SplitHostPort(o)
	if oHost == "" {
		oHost = o
	}
	tHost, _, _ := net.SplitHostPort(trustedHost)
	if tHost == "" {
		tHost = trustedHost
	}
	return oHost == tHost
}

// clientIP extracts client IP from RemoteAddr only (no XFF trust without proxy config).
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
