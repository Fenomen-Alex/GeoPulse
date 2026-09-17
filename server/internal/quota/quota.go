package quota

import (
	"fmt"
	"sync"
	"time"
)

// DefaultDailyQuota is the number of spatial runs allowed per authenticated
// user per 24-hour rolling window (bucket-reset at midnight UTC).
const DefaultDailyQuota = 15

// Tracker is implemented by both the in-memory Quota and the DB-backed
// Persistent tracker so handlers and status handlers work with either.
type Tracker interface {
	// Consume increments the user's usage and reports the remaining allowance
	// for the window. If the user is over quota it returns (0, false).
	Consume(userID string) (remaining int, allowed bool)
	// Remaining reports the current allowance without consuming a run.
	Remaining(userID string) int
}

// Quota tracks per-user daily spatial-run consumption in a concurrent-safe
// in-memory store.
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

// Remaining returns the user's current allowance without consuming a run.
func (q *Quota) Remaining(userID string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	key := fmt.Sprintf("%s:%s", userID, time.Now().UTC().Format("2006-01-02"))
	remaining := DefaultDailyQuota - q.used[key]
	if remaining < 0 {
		return 0
	}
	return remaining
}

var _ Tracker = (*Quota)(nil)
