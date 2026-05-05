package notify

import (
	"context"
	"fmt"
	"time"
)

// TimeoutNotifier wraps a Notifier and enforces a per-notification deadline.
// If the inner Notifier does not return within the configured timeout, the
// context passed to it is cancelled and an error is returned to the caller.
type TimeoutNotifier struct {
	inner   Notifier
	timeout time.Duration
}

// NewTimeoutNotifier returns a TimeoutNotifier that cancels delivery after d.
// A zero or negative duration is replaced with 5 seconds.
func NewTimeoutNotifier(inner Notifier, d time.Duration) *TimeoutNotifier {
	if d <= 0 {
		d = 5 * time.Second
	}
	return &TimeoutNotifier{inner: inner, timeout: d}
}

// Notify delivers msg through the inner Notifier, aborting after the
// configured timeout. The parent context is still respected: if it is already
// cancelled the call returns immediately.
func (t *TimeoutNotifier) Notify(ctx context.Context, msg Message) error {
	if t.inner == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	type result struct{ err error }
	ch := make(chan result, 1)

	go func() {
		ch <- result{err: t.inner.Notify(ctx, msg)}
	}()

	select {
	case res := <-ch:
		return res.err
	case <-ctx.Done():
		return fmt.Errorf("timeout notifier: %w", ctx.Err())
	}
}
