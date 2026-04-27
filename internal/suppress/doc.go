// Package suppress implements a per-key suppression window that prevents
// the same alert from being dispatched multiple times within a configurable
// cooldown period.
//
// # Overview
//
// When a cron job repeatedly misses its schedule or drifts, the monitoring
// loop can fire alerts on every check cycle. The suppress package sits in
// front of any Notify implementation and silently drops duplicate calls for
// the same subject key until the cooldown duration has elapsed.
//
// # Usage
//
//	s := suppress.New(15 * time.Minute)
//	n := suppress.NewNotifier(inner, s)
//	// Only the first Notify call per subject within 15 minutes is forwarded.
//	_ = n.Notify(ctx, "job.backup", "missed run")
//
// # Resetting
//
// Call Reset on the underlying Suppressor to clear a key's cooldown early,
// for example after a job recovers and you want the next failure to alert
// immediately.
package suppress
