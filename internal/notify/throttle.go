package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ThrottleNotifier wraps a Notifier and limits how frequently notifications
// can be sent per subject within a rolling time window.
type ThrottleNotifier struct {
	inner    Notifier
	window   time.Duration
	max      int
	mu       sync.Mutex
	counts   map[string][]time.Time
}

// NewThrottleNotifier creates a ThrottleNotifier that allows at most max
// notifications per subject within the given window duration.
func NewThrottleNotifier(inner Notifier, window time.Duration, max int) (*ThrottleNotifier, error) {
	if inner == nil {
		return nil, fmt.Errorf("throttle: inner notifier must not be nil")
	}
	if max <= 0 {
		return nil, fmt.Errorf("throttle: max must be greater than zero")
	}
	if window <= 0 {
		return nil, fmt.Errorf("throttle: window must be greater than zero")
	}
	return &ThrottleNotifier{
		inner:  inner,
		window: window,
		max:    max,
		counts: make(map[string][]time.Time),
	}, nil
}

// Notify sends the notification only if the subject has not exceeded the
// allowed rate within the rolling window. Excess calls are silently dropped.
func (t *ThrottleNotifier) Notify(ctx context.Context, subject, message string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-t.window)

	times := t.counts[subject]
	filtered := times[:0]
	for _, ts := range times {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}

	if len(filtered) >= t.max {
		t.counts[subject] = filtered
		return nil
	}

	t.counts[subject] = append(filtered, now)
	return t.inner.Notify(ctx, subject, message)
}

// Reset clears the throttle state for the given subject key.
func (t *ThrottleNotifier) Reset(subject string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.counts, subject)
}
