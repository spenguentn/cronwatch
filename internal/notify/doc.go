// Package notify provides alert notification backends for cronwatch.
//
// Supported notifiers:
//
//   - WebhookNotifier: sends JSON payloads to an HTTP endpoint.
//   - SlackNotifier:   sends formatted messages to a Slack incoming webhook.
//   - EmailNotifier:   sends SMTP email alerts to one or more recipients.
//   - RetryNotifier:   wraps any Notifier with configurable retry logic.
//
// All notifiers implement the Notifier interface:
//
//	type Notifier interface {
//		Notify(ctx context.Context, level, message string) error
//	}
//
// Typical usage:
//
//	base, _ := notify.NewEmailNotifier(notify.EmailConfig{
//		Host: "smtp.example.com",
//		From: "cronwatch@example.com",
//		To:   []string{"ops@example.com"},
//	})
//	notifier := notify.NewRetryNotifier(base, 3, time.Second)
//	notifier.Notify(ctx, "error", "job backup missed its window")
package notify
