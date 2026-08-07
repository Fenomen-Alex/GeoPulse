package quota

import (
	"fmt"
	"sync"
	"time"
)

// DefaultDailyQuota is the number of spatial runs allowed per authenticated
// user per 24-hour rolling window (bucket-reset at midnight UTC).
const DefaultDailyQuota = 15

// Quota tracks per-user daily spatial-run consumption in a concurrent-safe
// in-memory store. Production can swap this for the Turso-backed counters in
// internal/db without changing handler call sites.
type Quota struct {
	mu   sync.Mutex
	used map[string]int
}

func New() *Quota {
	return &Quota{used: make(map[string]int)}
}

// Consume atomically increments the user's usage for today. It returns the
// remaining quota for the window and whether the request is allowed. When the
// user has exhausted their quota, allowed is false and remaining is 0.
func (q *Quota) Consume(userID string) (remaining int, allowed bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, time.Now().UTC().Format("2006-01-02"))
	used := q.used[key] + 1
	if used > DefaultDailyQuota {
		return 0, false
	}
	q.used[key] = used
	return DefaultDailyQuota - used, true
}
