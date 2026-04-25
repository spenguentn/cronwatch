// Package store provides a lightweight, file-backed persistence layer for
// cronwatch job execution records.
//
// A Store holds one JobRecord per monitored cron job and flushes changes to
// a JSON file on every write, using an atomic rename to avoid partial writes.
//
// # File Format
//
// State is persisted as a JSON object mapping job names to their most recent
// JobRecord. The file is written atomically: changes are first written to a
// temporary file in the same directory, then renamed into place, ensuring the
// state file is never left in a partially-written state.
//
// # Concurrency
//
// Store and Updater are safe for concurrent use by multiple goroutines.
// Internal writes are serialised with a mutex.
//
// # Typical usage
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
