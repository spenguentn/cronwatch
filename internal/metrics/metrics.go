// Package metrics tracks runtime statistics for monitored cron jobs.
package metrics

import (
	"sync"
	"time"
)

// JobStats holds execution statistics for a single cron job.
type JobStats struct {
	Name        string
	TotalChecks int64
	MissedRuns  int64
	DriftEvents int64
	LastDrift   time.Duration
	LastChecked time.Time
}

// Collector accumulates metrics for all monitored jobs.
type Collector struct {
	mu   sync.RWMutex
	jobs map[string]*JobStats
}

// NewCollector returns an initialised Collector.
func NewCollector() *Collector {
	return &Collector{
		jobs: make(map[string]*JobStats),
	}
}

// RecordCheck increments the total check counter for the named job.
func (c *Collector) RecordCheck(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.getOrCreate(name)
	s.TotalChecks++
	s.LastChecked = time.Now()
}

// RecordMiss increments the missed-run counter for the named job.
func (c *Collector) RecordMiss(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.getOrCreate(name).MissedRuns++
}

// RecordDrift increments the drift counter and stores the latest drift value.
func (c *Collector) RecordDrift(name string, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s := c.getOrCreate(name)
	s.DriftEvents++
	s.LastDrift = d
}

// Snapshot returns a copy of stats for all jobs.
func (c *Collector) Snapshot() []JobStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]JobStats, 0, len(c.jobs))
	for _, s := range c.jobs {
		out = append(out, *s)
	}
	return out
}

// Get returns stats for a single job and whether it was found.
func (c *Collector) Get(name string) (JobStats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.jobs[name]
	if !ok {
		return JobStats{}, false
	}
	return *s, true
}

// Reset clears all recorded statistics for the named job. If the job does not
// exist, Reset is a no-op.
func (c *Collector) Reset(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.jobs[name]; ok {
		c.jobs[name] = &JobStats{Name: name}
	}
}

// getOrCreate must be called with the write lock held.
func (c *Collector) getOrCreate(name string) *JobStats {
	if _, ok := c.jobs[name]; !ok {
		c.jobs[name] = &JobStats{Name: name}
	}
	return c.jobs[name]
}
