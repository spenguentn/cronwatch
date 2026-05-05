// Package notify provides composable notification primitives for cronwatch.
//
// # RoutingNotifier
//
// RoutingNotifier dispatches each [Message] to a specific [Notifier] chosen
// by a user-supplied key function. This allows different cron jobs (or job
// tiers) to be routed to distinct alerting channels without duplicating
// pipeline configuration.
//
// Basic usage:
//
//	// Route by the "job" metadata key.
//	// Critical jobs go to PagerDuty; everything else falls back to Slack.
//	 rn := notify.NewRoutingNotifier(
//	     func(m notify.Message) string { return m.Meta["job"] },
//	     slackNotifier,                          // fallback
//	     notify.Route{Name: "db-backup",  Notifier: pagerDutyNotifier},
//	     notify.Route{Name: "log-rotate", Notifier: slackNotifier},
//	 )
//
// Routes can be added or removed at runtime via [RoutingNotifier.AddRoute] and
// [RoutingNotifier.RemoveRoute]; both methods are safe for concurrent use.
//
// If no route matches and no fallback is configured, the message is silently
// dropped — consistent with the opt-in alerting philosophy used elsewhere in
// this package.
package notify
