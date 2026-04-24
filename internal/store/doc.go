// Package store provides a lightweight, file-backed persistence layer for
// cronwatch job execution records.
//
// A Store holds one JobRecord per monitored cron job and flushes changes to
// a JSON file on every write, using an atomic rename to avoid partial writes.
//
// Typical usage:
//
//	s, err := store.New("/var/lib/cronwatch/state.json")
//	if err != nil { ... }
//	u := store.NewUpdater(s)
//
//	// After a job runs:
//	_ = u.Record(store.RunResult{
//		JobName:  "daily-backup",
//		RunAt:    time.Now(),
//		ExitCode: 0,
//	})
//
//	// Query last run time for drift checking:
//	last, ok := u.LastRun("daily-backup")
package store
