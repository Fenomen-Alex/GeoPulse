package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// limiterEntry holds a rate limiter and its creation time for expiry tracking.
type limiterEntry struct {
	limiter   *rate.Limiter
	createdAt time.Time
}

// RateLimiterStore maintains concurrent-safe rate limiters keyed by identifier.
type RateLimiterStore struct {
	mu       sync.Map
	rate     rate.Limit
	burst    int
	lifetime time.Duration
}

// NewGuestRateLimiter creates a store enforcing 10 requests/hour per IP.
func NewGuestRateLimiter() *RateLimiterStore {
	return &RateLimiterStore{
		rate:     rate.Every(time.Hour / 10), // 10 requests per hour
		burst:    10,
		lifetime: time.Hour,
	}
}

// NewAuthRateLimiter creates a store enforcing 200 requests/day per user ID.
func NewAuthRateLimiter() *RateLimiterStore {
	return &RateLimiterStore{
		rate:     rate.Every(24 * time.Hour / 200), // 200 requests per day
		burst:    200,
		lifetime: 24 * time.Hour,
	}
}

// getLimiter retrieves or creates a rate limiter for the given key.
func (s *RateLimiterStore) getLimiter(key string) *rate.Limiter {
	now := time.Now()

	if val, ok := s.mu.Load(key); ok {
		entry := val.(*limiterEntry)
		// Reuse limiter if lifetime has not expired
		if now.Sub(entry.createdAt) < s.lifetime {
			return entry.limiter
		}
	}

	limiter := rate.NewLimiter(s.rate, s.burst)
	s.mu.Store(key, &limiterEntry{
		limiter:   limiter,
		createdAt: now,
	})
	return limiter
}

// GuestRateLimit returns middleware that rate-limits by client IP.
// Guest Quota: 10 requests/hour per IP.
func GuestRateLimit(store *RateLimiterStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			limiter := store.getLimiter(ip)

			if !limiter.Allow() {
				resetTime := time.Now().Add(store.lifetime)
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded","message":"guest quota exhausted, please sign in for higher limits"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractIP retrieves the client IP address from the request.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (for proxied requests)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	// Check X-Real-IP
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
