package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type rateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	rateLimits   sync.Map
	cleanupTimer *time.Timer
)

func init() {
	cleanupTimer = time.AfterFunc(5*time.Minute, cleanupRateLimits)
}

func cleanupRateLimits() {
	now := time.Now()
	rateLimits.Range(func(key, value interface{}) bool {
		entry := value.(*rateLimitEntry)
		if now.Sub(entry.lastSeen) > 10*time.Minute {
			rateLimits.Delete(key)
		}
		return true
	})
	cleanupTimer.Reset(5 * time.Minute)
}

func getLimiter(ip string) *rate.Limiter {
	raw, _ := rateLimits.LoadOrStore(ip, &rateLimitEntry{
		limiter:  rate.NewLimiter(rate.Every(time.Minute), 10),
		lastSeen: time.Now(),
	})
	entry := raw.(*rateLimitEntry)
	entry.lastSeen = time.Now()
	return entry.limiter
}

// clientIP resolves the effective client address for rate limiting. It prefers
// the first X-Forwarded-For entry (set by Render/edge proxies) and falls back
// to RemoteAddr with the port stripped. Keying the full RemoteAddr (which
// includes an ephemeral port) would let clients bypass the limit by opening
// new connections.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for _, p := range parts {
			if ip := strings.TrimSpace(p); ip != "" {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimitMiddleware applies token-bucket rate limiting per IP.
// Guest: 10 requests/minute. Authenticated users are exempt.
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			next.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("geopulse_session")
		if err != nil || cookie.Value == "" {
			ip := clientIP(r)
			limiter := getLimiter(ip)
			if !limiter.Allow() {
				w.Header().Set("X-RateLimit-Reset", time.Now().Add(time.Minute).Format(time.RFC3339))
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
