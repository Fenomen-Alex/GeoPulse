package quota

import (
	"log"
	"sync"
)

// UsageStore abstracts the DB operations the persistent tracker needs so the
// quota package never depends on the concrete SQL layer (and tests can stub it).
type UsageStore interface {
	EnsureUser(userID string) error
	GetUserDailyUsage(userID string) (used int, dailyQuota int, err error)
	IncrementUserUsage(userID string) error
}

// Persistent is a Tracker whose counters live in user_daily_usage (embedded
// SQLite) instead of process memory, so usage survives restarts. It wraps a
// UsageStore and serializes access per process.
type Persistent struct {
	mu    sync.Mutex
	store UsageStore
}

// NewPersistent builds a DB-backed tracker. If the store is nil it degrades to
// an in-memory tracker so startup can survive a broken migration environment.
func NewPersistent(store UsageStore) Tracker {
	if store == nil {
		return New()
	}
	return &Persistent{store: store}
}

// Consume atomically increments the user's usage for today via the store. On a
// store failure it fails open (allows the request) and logs, so a transient DB
// error never intermittently blocks users; persistence resumes on the next run.
func (p *Persistent) Consume(userID string) (remaining int, allowed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.store.EnsureUser(userID); err != nil {
		log.Printf("quota: EnsureUser failed for %s: %v", userID, err)
		return DefaultDailyQuota, true
	}

	used, quota, err := p.store.GetUserDailyUsage(userID)
	if err != nil {
		log.Printf("quota: GetUserDailyUsage failed for %s: %v", userID, err)
		return DefaultDailyQuota, true
	}

	if used >= quota {
		return 0, false
	}

	if err := p.store.IncrementUserUsage(userID); err != nil {
		log.Printf("quota: IncrementUserUsage failed for %s: %v", userID, err)
		return quota - used - 1, true
	}

	return quota - used - 1, true
}

// Remaining reports the current allowance without consuming a run.
func (p *Persistent) Remaining(userID string) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	used, quota, err := p.store.GetUserDailyUsage(userID)
	if err != nil {
		log.Printf("quota: GetUserDailyUsage failed for %s: %v", userID, err)
		return DefaultDailyQuota
	}
	remaining := quota - used
	if remaining < 0 {
		return 0
	}
	return remaining
}

var _ Tracker = (*Persistent)(nil)