// Package notify provides composable notifier primitives for cronwatch.
//
// # FallbackNotifier
//
// FallbackNotifier wraps a primary [Notifier] and one or more fallback
// notifiers. It attempts delivery to the primary first. If the primary
// returns an error (or is nil), it tries each fallback in order, returning
// nil as soon as one succeeds.
//
// This is useful for building resilient alerting pipelines where, for
// example, PagerDuty is the primary target but email or Slack act as
// backup channels if PagerDuty is unreachable.
//
// Example usage:
//
//	primary := notify.NewPagerDutyNotifier(pdKey)
//	backup := notify.NewSlackNotifier(webhookURL)
//
//	fn := notify.NewFallbackNotifier(primary, backup)
//	// fn.Notify delivers to primary; falls back to backup on error.
package notify
