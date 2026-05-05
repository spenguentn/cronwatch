// Package notify provides notification primitives for cronwatch alerts.
//
// # AuditNotifier
//
// AuditNotifier is a decorator that wraps any Notifier and records every
// dispatch attempt — its subject, severity, timestamp, and whether the
// underlying delivery succeeded or failed.
//
// It is useful for post-mortem analysis, testing pipelines, and surfacing
// delivery statistics via the /metrics endpoint.
//
// Usage:
//
//	inner := notify.NewWebhookNotifier(webhookURL, nil)
//	an := notify.NewAuditNotifier(inner, os.Stderr)
//
//	// Use an wherever a Notifier is expected.
//	an.Notify(ctx, msg)
//
//	// Inspect recorded entries.
//	for _, e := range an.Entries() {
//		if !e.Success {
//			log.Printf("delivery failed for %s: %v", e.Subject, e.Err)
//		}
//	}
//
//	// Clear history when no longer needed.
//	an.Reset()
package notify
