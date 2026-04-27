// Package notify provides notification backends for cronwatch alerts.
//
// Supported notifiers:
//
//   - WebhookNotifier: sends a JSON POST to an arbitrary HTTP endpoint.
//   - SlackNotifier: posts a message to a Slack incoming webhook URL.
//   - EmailNotifier: delivers alerts via SMTP.
//   - PagerDutyNotifier: triggers incidents via the PagerDuty Events API v2.
//
// All notifiers implement a common Notify(ctx, msg) error signature and can
// be composed with RetryNotifier to add automatic retry logic with
// configurable attempts and back-off delay.
//
// Example usage:
//
//	base, err := notify.NewPagerDutyNotifier(integrationKey)
//	if err != nil { ... }
//	n := notify.NewRetryNotifier(base, 3, 2*time.Second)
//	if err := n.Notify(ctx, "cron job missed"); err != nil { ... }
package notify
