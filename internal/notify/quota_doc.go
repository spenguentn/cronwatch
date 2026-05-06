// Package notify provides composable notification primitives for cronwatch.
//
// # QuotaNotifier
//
// QuotaNotifier enforces a hard cap on the total number of notifications
// delivered within a rolling time window. This is useful when alerting on
// frequently-firing cron jobs to prevent notification storms.
//
// Usage:
//
//	// Allow at most 10 alerts per hour.
//	 q := notify.NewQuotaNotifier(inner, 10, time.Hour)
//	 _ = q.Notify(ctx, msg)
//
// Once the quota is exhausted, subsequent calls to Notify return nil without
// forwarding the message. The quota resets automatically when the window
// elapses, or can be reset manually via Reset.
//
// Remaining returns the number of notifications still available in the current
// window, or -1 when the notifier is configured with an unlimited quota
// (max <= 0).
//
// QuotaNotifier is safe for concurrent use.
package notify
