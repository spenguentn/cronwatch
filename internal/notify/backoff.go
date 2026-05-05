package notify

import (
	"context"
	"fmt"
	"math"
	"time"
)

// BackoffNotifier wraps a Notifier and retries with exponential backoff on failure.
// Each successive attempt waits base * 2^attempt before retrying, capped at maxDelay.
type BackoffNotifier struct {
	inner    Notifier
	attempts int
	base     time.Duration
	maxDelay time.Duration
	sleep    func(context.Context, time.Duration) error
}

// NewBackoffNotifier creates a BackoffNotifier that retries up to attempts times
// with exponential backoff starting at base, capped at maxDelay.
func NewBackoffNotifier(inner Notifier, attempts int, base, maxDelay time.Duration) *BackoffNotifier {
	if attempts < 1 {
		attempts = 1
	}
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	if maxDelay <= 0 {
		maxDelay = 30 * time.Second
	}
	return &BackoffNotifier{
		inner:    inner,
		attempts: attempts,
		base:     base,
		maxDelay: maxDelay,
		sleep:    contextSleep,
	}
}

// Notify delivers msg, retrying with exponential backoff on transient errors.
func (b *BackoffNotifier) Notify(ctx context.Context, msg Message) error {
	if b.inner == nil {
		return nil
	}
	var last error
	for i := 0; i < b.attempts; i++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("backoff: context cancelled: %w", err)
		}
		if last = b.inner.Notify(ctx, msg); last == nil {
			return nil
		}
		if i < b.attempts-1 {
			delay := b.delayFor(i)
			if err := b.sleep(ctx, delay); err != nil {
				return fmt.Errorf("backoff: context cancelled during wait: %w", err)
			}
		}
	}
	return fmt.Errorf("backoff: all %d attempts failed: %w", b.attempts, last)
}

func (b *BackoffNotifier) delayFor(attempt int) time.Duration {
	exp := math.Pow(2, float64(attempt))
	d := time.Duration(float64(b.base) * exp)
	if d > b.maxDelay {
		d = b.maxDelay
	}
	return d
}

func contextSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
