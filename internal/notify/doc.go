// Package notify provides outbound notification mechanisms for cronwatch alerts.
//
// It includes a WebhookNotifier that delivers alert payloads to HTTP endpoints
// via POST requests, and a RetryNotifier that wraps any Notifier implementation
// with configurable retry logic and exponential back-off.
//
// # WebhookNotifier
//
// The WebhookNotifier sends a JSON-encoded alert body to a configured URL.
// It respects context cancellation and returns an error for non-2xx responses.
//
//	notifier := notify.NewWebhookNotifier("https://hooks.example.com/cronwatch", http.DefaultClient)
//	err := notifier.Notify(ctx, alert.Event{...})
//
// # RetryNotifier
//
// RetryNotifier wraps another Notifier and retries on transient failures.
// Attempts are spaced with a simple exponential back-off. The total number
// of attempts (including the first) is controlled by the MaxAttempts field.
//
//	base := notify.NewWebhookNotifier(url, client)
//	retrying := notify.NewRetryNotifier(base, notify.RetryConfig{
//		MaxAttempts: 3,
//		BaseDelay:   500 * time.Millisecond,
//	})
//
// Both notifiers satisfy the alert.Notifier interface, so they can be
// registered directly with an alert.Dispatcher.
package notify
