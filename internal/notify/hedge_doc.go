// Package notify provides composable notification primitives for cronwatch.
//
// # HedgeNotifier
//
// HedgeNotifier implements the "hedged request" pattern: it sends a message to
// a primary Notifier immediately and, if the primary has not responded within a
// configurable delay, concurrently fires the same message to a secondary
// Notifier. The first successful delivery wins and the other in-flight request
// is abandoned via context cancellation.
//
// This is useful when latency predictability matters more than minimising
// duplicate deliveries — for example, paging an on-call engineer where a
// 1-second delay is unacceptable but an occasional double-delivery is fine.
//
// Usage:
//
//	primary := notify.NewWebhookNotifier(primaryURL, 5*time.Second)
//	secondary := notify.NewSlackNotifier(slackURL, 5*time.Second)
//	h := notify.NewHedgeNotifier(primary, secondary, 500*time.Millisecond)
//
//	// If primary responds in < 500 ms, secondary is never called.
//	// Otherwise both are tried concurrently.
//	err := h.Notify(ctx, msg)
package notify
