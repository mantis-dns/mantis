package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"

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

// RateLimitMiddleware limits requests per session/IP.
func RateLimitMiddleware(rps int) func(http.Handler) http.Handler {
	limiters := &sync.Map{}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := clientIP(r)
			val, _ := limiters.LoadOrStore(key, rate.NewLimiter(rate.Limit(rps)/60, rps))
			limiter := val.(*rate.Limiter)
			if !limiter.Allow() {
				Error(w, "RATE_LIMITED", "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds security headers to responses.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'")
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware handles CORS.
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Api-Key")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	ip, _, _ := strings.Cut(r.RemoteAddr, ":")
	return ip
}
