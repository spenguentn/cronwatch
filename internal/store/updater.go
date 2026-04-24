package store

import (
	"fmt"
	"time"
)

// RunResult carries the outcome of a single cron job execution.
type RunResult struct {
	JobName  string
	RunAt    time.Time
	ExitCode int
}

// Updater wraps a Store and applies RunResults to it.
type Updater struct {
	store *Store
}

// NewUpdater creates an Updater backed by the given Store.
func NewUpdater(s *Store) *Updater {
	return &Updater{store: s}
}

// Record persists a RunResult, returning an error if the job name is empty.
func (u *Updater) Record(r RunResult) error {
	if r.JobName == "" {
		return fmt.Errorf("store: job name must not be empty")
	}
	if r.RunAt.IsZero() {
		r.RunAt = time.Now().UTC()
	}
	return u.store.Set(JobRecord{
		Name:     r.JobName,
		LastRun:  r.RunAt,
		LastExit: r.ExitCode,
	})
}

// LastRun returns the last recorded run time for a job, and whether it exists.
func (u *Updater) LastRun(jobName string) (time.Time, bool) {
	r, ok := u.store.Get(jobName)
	if !ok {
		return time.Time{}, false
	}
	return r.LastRun, true
}
