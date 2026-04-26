package store

import (
	"time"
)

// Cleaner removes stale job records from the store that have not been
// updated within the given retention window.
type Cleaner struct {
	store     *Store
	retention time.Duration
}

// NewCleaner returns a Cleaner that will remove records older than retention.
func NewCleaner(s *Store, retention time.Duration) *Cleaner {
	return &Cleaner{
		store:     s,
		retention: retention,
	}
}

// Prune deletes all job entries whose last-run timestamp is older than
// now minus the retention window. It returns the names of removed jobs.
func (c *Cleaner) Prune(now time.Time) []string {
	cutoff := now.Add(-c.retention)

	c.store.mu.Lock()
	defer c.store.mu.Unlock()

	var removed []string
	for name, ts := range c.store.data {
		if ts.Before(cutoff) {
			delete(c.store.data, name)
			removed = append(removed, name)
		}
	}

	if len(removed) > 0 {
		_ = c.store.save()
	}

	return removed
}
