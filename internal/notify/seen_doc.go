// Package notify provides composable notification primitives for cronwatch.
//
// # SeenNotifier
//
// SeenNotifier wraps any Notifier and suppresses messages whose full content
// (subject + body) has already been forwarded within a configurable rolling
// window. It differs from DedupeNotifier, which keys only on the message
// subject.
//
// Use SeenNotifier when the same alert type may legitimately fire with
// different bodies (e.g. varying drift values) and you want each unique
// payload to reach the downstream notifier at most once per window.
//
// Example:
//
//	wh := notify.NewWebhookNotifier(webhookURL, 5*time.Second)
//	sn := notify.NewSeenNotifier(wh, 10*time.Minute)
//
//	// Only the first message with this exact subject+body is forwarded;
//	// subsequent identical messages within 10 minutes are dropped.
//	_ = sn.Notify(ctx, msg)
//
// Calling Reset clears all tracked hashes immediately, allowing the next
// matching message to pass through regardless of the window.
package notify
