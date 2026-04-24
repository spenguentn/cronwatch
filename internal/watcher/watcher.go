// Package watcher polls configured cron jobs and checks for drift or missed runs.
package watcher

import (
	"context"
	"log"
	"time"

	"cronwatch/internal/alert"
	"cronwatch/internal/config"
	"cronwatch/internal/schedule"
	"cronwatch/internal/store"
)

// Watcher periodically evaluates each configured job against its last recorded run.
type Watcher struct {
	cfg      *config.Config
	updater  *store.Updater
	disp     *alert.Dispatcher
	interval time.Duration
}

// New creates a Watcher with the given dependencies.
func New(cfg *config.Config, updater *store.Updater, disp *alert.Dispatcher, interval time.Duration) *Watcher {
	return &Watcher{
		cfg:      cfg,
		updater:  updater,
		disp:     disp,
		interval: interval,
	}
}

// Run starts the watch loop, blocking until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Printf("watcher: starting, poll interval=%s", w.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("watcher: shutting down")
			return
		case t := <-ticker.C:
			w.check(t)
		}
	}
}

// check evaluates all jobs at the given instant.
func (w *Watcher) check(now time.Time) {
	for _, job := range w.cfg.Jobs {
		last, err := w.updater.LastRun(job.Name)
		if err != nil {
			// No recorded run yet — skip drift check but warn.
			w.disp.Warn(job.Name, "no last-run recorded")
			continue
		}

		result, err := schedule.CheckDrift(job.Schedule, last, now, job.DriftThreshold)
		if err != nil {
			log.Printf("watcher: drift check error for %s: %v", job.Name, err)
			continue
		}

		switch {
		case result.Missed:
			w.disp.Error(job.Name, "missed scheduled run")
		case result.DriftExceeded:
			w.disp.Warn(job.Name, "execution drift exceeded threshold")
		}
	}
}
