package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

// ipBucket is a token-bucket state for a single IP address.
type ipBucket struct {
	mu        sync.Mutex
	tokens    float64
	lastCheck time.Time
}

// RateLimiter is a per-IP token-bucket rate limiter.
// It is safe for concurrent use.
type RateLimiter struct {
	mu      sync.RWMutex
	clients map[string]*ipBucket
	rps     float64 // refill rate in tokens/second
	burst   float64 // maximum token capacity
}

// NewRateLimiter creates a rate limiter allowing rpm requests per minute per IP,
// with a burst of burst simultaneous requests.
func NewRateLimiter(rpm, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ipBucket),
		rps:     float64(rpm) / 60.0,
		burst:   float64(burst),
	}
	go rl.evictLoop()
	return rl
}

// Allow returns true if the given IP may proceed, false if rate-limited.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	b, ok := rl.clients[ip]
	if !ok {
		b = &ipBucket{tokens: rl.burst, lastCheck: time.Now()}
		rl.clients[ip] = b
	}
	rl.mu.Unlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.lastCheck = now
	b.tokens += elapsed * rl.rps
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Middleware returns an http.Handler that rejects requests exceeding the rate.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := extractIP(r.RemoteAddr)
		if !rl.Allow(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message":"Too Many Requests"}`)) //nolint:errcheck
			return
		}
		next.ServeHTTP(w, r)
	})
}

// evictLoop removes stale IP buckets every 10 minutes to prevent unbounded growth.
func (rl *RateLimiter) evictLoop() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		rl.mu.Lock()
		for ip, b := range rl.clients {
			b.mu.Lock()
			stale := b.lastCheck.Before(cutoff)
			b.mu.Unlock()
			if stale {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// extractIP strips the port from a host:port address.
func extractIP(addr string) string {
	if i := strings.LastIndex(addr, ":"); i > 0 {
		return addr[:i]
	}
	return addr
}

// CSRFGuard rejects state-mutating requests that lack the X-Api-Key or
// X-Requested-With header. Browsers cannot send custom headers in cross-origin
// requests without an explicit CORS pre-flight, so this effectively prevents
// CSRF for API routes that are already protected by the API-key middleware.
func CSRFGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("X-Api-Key") == "" && r.Header.Get("X-Requested-With") == "" {
			// Allow requests that carry a Bearer / Basic auth token (login flow).
			if r.Header.Get("Authorization") != "" {
				next.ServeHTTP(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"message":"CSRF check failed: missing X-Api-Key or X-Requested-With"}`)) //nolint:errcheck
			return
		}
		next.ServeHTTP(w, r)
	})
}
