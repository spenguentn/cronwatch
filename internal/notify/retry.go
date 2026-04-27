package notify

import (
	"context"
	"fmt"
	"time"
)

// Notifier is the interface implemented by any alert sender.
type Notifier interface {
	Notify(ctx context.Context, level, job, message string) error
}

// RetryNotifier wraps a Notifier and retries on transient failures.
type RetryNotifier struct {
	inner   Notifier
	attempts int
	delay    time.Duration
}

// NewRetryNotifier wraps n, retrying up to attempts times with delay between tries.
func NewRetryNotifier(n Notifier, attempts int, delay time.Duration) *RetryNotifier {
	if attempts < 1 {
		attempts = 1
	}
	return &RetryNotifier{inner: n, attempts: attempts, delay: delay}
}

// Notify calls the wrapped Notifier, retrying on error up to the configured limit.
func (r *RetryNotifier) Notify(ctx context.Context, level, job, message string) error {
	var last error
	for i := 0; i < r.attempts; i++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry: context done: %w", err)
		}
		last = r.inner.Notify(ctx, level, job, message)
		if last == nil {
			return nil
		}
		if i < r.attempts-1 {
			select {
			case <-time.After(r.delay):
			case <-ctx.Done():
				return fmt.Errorf("retry: context done while waiting: %w", ctx.Err())
			}
		}
	}
	return fmt.Errorf("retry: all %d attempts failed: %w", r.attempts, last)
}
