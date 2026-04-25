// Package healthcheck provides a simple HTTP health endpoint
// that reports daemon liveness and readiness based on watcher state.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// Status represents the health state of the daemon.
type Status struct {
	OK        bool      `json:"ok"`
	StartedAt time.Time `json:"started_at"`
	CheckedAt time.Time `json:"checked_at,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// Checker tracks daemon health state.
type Checker struct {
	mu        sync.RWMutex
	startedAt time.Time
	lastCheck time.Time
	healthy   bool
	message   string
}

// New creates a Checker marked healthy at construction time.
func New() *Checker {
	return &Checker{
		startedAt: time.Now(),
		healthy:   true,
	}
}

// SetHealthy updates the health state and optional message.
func (c *Checker) SetHealthy(ok bool, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.healthy = ok
	c.message = msg
	c.lastCheck = time.Now()
}

// Status returns a snapshot of the current health state.
func (c *Checker) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return Status{
		OK:        c.healthy,
		StartedAt: c.startedAt,
		CheckedAt: c.lastCheck,
		Message:   c.message,
	}
}

// Handler returns an http.HandlerFunc that serves the health status as JSON.
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s := c.Status()
		w.Header().Set("Content-Type", "application/json")
		if !s.OK {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(s)
	}
}
