// Package notify provides composable notification primitives for cronwatch.
//
// # BatchNotifier
//
// BatchNotifier accumulates outgoing messages and delivers them as a single
// combined notification, reducing alert noise when many jobs fire at once.
//
// Messages are flushed when either condition is met:
//   - The configured time window elapses since the first queued message.
//   - The number of pending messages reaches maxSize.
//
// Callers may also trigger an immediate flush via Flush.
//
// Example:
//
//	wh := notify.NewWebhookNotifier(webhookURL, 5*time.Second)
//	b := notify.NewBatchNotifier(wh, 30*time.Second, 20)
//	_ = b.Notify(ctx, msg)
//	// ... more Notify calls ...
//	_ = b.Flush(ctx) // send whatever is pending right now
package notify
