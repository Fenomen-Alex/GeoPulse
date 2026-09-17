package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// UserIDFromContext returns the authenticated user ID set by RequireAuth, or
// an empty string when no session is present (guests).
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

// userRateLimitEntry is a per-user token bucket plus last-seen marker so
// dormant buckets get reclaimed.
type userRateLimitEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	userRateLimits sync.Map
	userCleanup    *time.Timer
)

func init() {
	userCleanup = time.AfterFunc(5*time.Minute, cleanupUserRateLimits)
}

func cleanupUserRateLimits() {
	now := time.Now()
	userRateLimits.Range(func(key, value interface{}) bool {
		entry := value.(*userRateLimitEntry)
		if now.Sub(entry.lastSeen) > 10*time.Minute {
			userRateLimits.Delete(key)
		}
		return true
	})
	userCleanup.Reset(5 * time.Minute)
}

func getUserLimiter(id string, r rate.Limit, burst int) *rate.Limiter {
	raw, _ := userRateLimits.LoadOrStore(id, &userRateLimitEntry{
		limiter:  rate.NewLimiter(r, burst),
		lastSeen: time.Now(),
	})
	entry := raw.(*userRateLimitEntry)
	entry.lastSeen = time.Now()
	return entry.limiter
}

// UserRateLimitMiddleware applies a token bucket per authenticated user
// (falling back to the resolved client IP for guests). Endpoints protected by
// it are NOT quota-counted (e.g. geocoding), so this caps upstream abuse while
// still allowing real interactive use.
func UserRateLimitMiddleware(every time.Duration, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := UserIDFromContext(r.Context())
			if id == "" {
				id = clientIP(r)
			}
			lim := getUserLimiter(id, rate.Every(every), burst)
			if !lim.Allow() {
				w.Header().Set("X-RateLimit-Reset", time.Now().Add(every).Format(time.RFC3339))
				http.Error(w, `{"error":"Too many requests. Please slow down."}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// PerIPRateLimitMiddleware applies a token bucket keyed by the resolved client
// IP regardless of authentication/session. Public endpoints that are abusable
// (e.g. the contact form) use this so logged-in users cannot spam them either.
func PerIPRateLimitMiddleware(every time.Duration, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lim := getUserLimiter("ip:"+clientIP(r), rate.Every(every), burst)
			if !lim.Allow() {
				w.Header().Set("X-RateLimit-Reset", time.Now().Add(every).Format(time.RFC3339))
				http.Error(w, `{"error":"Too many requests. Please try again later."}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}