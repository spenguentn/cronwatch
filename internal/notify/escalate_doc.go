// Package notify provides composable notification primitives for cronwatch.
//
// # EscalateNotifier
//
// EscalateNotifier wraps a primary and secondary [Notifier]. It first attempts
// delivery via the primary. If the primary returns an error the message is
// immediately forwarded to the secondary.
//
// Optionally, a deadline can be configured. When non-zero, each successfully
// delivered message is tracked internally. If [EscalateNotifier.Acknowledge]
// is not called within the deadline, the next call to [EscalateNotifier.Tick]
// will escalate the message to the secondary notifier with an "[ESCALATED]"
// prefix on the subject.
//
// Typical usage:
//
//	primary := notify.NewSlackNotifier(slackURL, 5*time.Second)
//	secondary := notify.NewPagerDutyNotifier(pdKey)
//	esc := notify.NewEscalateNotifier(primary, secondary, 30*time.Minute)
//
//	// In your alert loop:
//	esc.Notify(ctx, msg)
//
//	// When an operator acknowledges via your UI:
//	esc.Acknowledge(msg.Subject)
//
//	// Drive deadline checks from a ticker:
//	esc.Tick(ctx, time.Now())
package notify
