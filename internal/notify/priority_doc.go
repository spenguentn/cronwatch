// Package notify provides composable notification primitives for cronwatch.
//
// # Priority Routing
//
// PriorityRouter dispatches a [Message] to one of three inner [Notifier]
// instances — low, normal, or high — based on the "priority" metadata key.
//
// Use [WithPriority] to attach a [Priority] value to a message before
// passing it to the router:
//
//	router := notify.NewPriorityRouter(lowNotifier, normalNotifier, highNotifier)
//	msg := notify.WithPriority(notify.Message{Subject: "job missed"}, notify.PriorityHigh)
//	err := router.Notify(ctx, msg)
//
// If the "priority" metadata key is absent or unrecognised, the message is
// forwarded to the normal-priority notifier as a safe default.
//
// Any tier notifier may be nil; messages routed to a nil notifier are
// silently dropped without error.
package notify
