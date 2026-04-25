package watcher

import (
	"fmt"
	"time"

	"github.com/cronwatch/cronwatch/internal/alert"
	"github.com/cronwatch/cronwatch/internal/schedule"
	"github.com/cronwatch/cronwatch/internal/store"
)

// JobStatus represents the result of checking a single job.
type JobStatus struct {
	Name    string
	Missed  bool
	Drifted bool
	Delta   time.Duration
	Err     error
}

// Checker evaluates each configured job against its last recorded run.
type Checker struct {
	updater   *store.Updater
	dispatcher *alert.Dispatcher
}

// NewChecker creates a Checker backed by the given updater and dispatcher.
func NewChecker(u *store.Updater, d *alert.Dispatcher) *Checker {
	return &Checker{updater: u, dispatcher: d}
}

// CheckJob inspects a single job by name, schedule expression, and allowed
// drift tolerance. It dispatches alerts and returns the resulting JobStatus.
func (c *Checker) CheckJob(name, expr string, tolerance time.Duration) JobStatus {
	status := JobStatus{Name: name}

	lastRun, ok := c.updater.LastRun(name)
	if !ok {
		// No recorded run yet — treat as a warning, not an error.
		c.dispatcher.Warn(fmt.Sprintf("job %q has no recorded run", name))
		return status
	}

	result, err := schedule.CheckDrift(expr, lastRun, tolerance)
	if err != nil {
		status.Err = err
		c.dispatcher.Error(fmt.Sprintf("job %q drift check failed: %v", name, err))
		return status
	}

	status.Delta = result.Delta

	if result.Missed {
		status.Missed = true
		c.dispatcher.Error(fmt.Sprintf("job %q missed its scheduled run (delta: %s)", name, result.Delta))
		return status
	}

	if result.Drifted {
		status.Drifted = true
		c.dispatcher.Warn(fmt.Sprintf("job %q drifted by %s (tolerance: %s)", name, result.Delta, tolerance))
	}

	return status
}
