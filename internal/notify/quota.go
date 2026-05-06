package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// QuotaNotifier limits the total number of notifications delivered within a
// rolling time window. Once the quota is exhausted, further calls are silently
// dropped until the window resets.
type QuotaNotifier struct {
	inner  Notifier
	max    int
	window time.Duration

	mu      sync.Mutex
	count   int
	windowStart time.Time
	now     func() time.Time
}

// NewQuotaNotifier wraps inner and allows at most max notifications per window
// duration. A zero or negative max is treated as unlimited.
func NewQuotaNotifier(inner Notifier, max int, window time.Duration) *QuotaNotifier {
	return &QuotaNotifier{
		inner:       inner,
		max:         max,
		window:      window,
		windowStart: time.Now(),
		now:         time.Now,
	}
}

// Notify forwards msg to the inner notifier if the quota has not been exceeded
// for the current window. Drops the message silently otherwise.
func (q *QuotaNotifier) Notify(ctx context.Context, msg Message) error {
	if q.inner == nil {
		return nil
	}
	if q.max <= 0 {
		return q.inner.Notify(ctx, msg)
	}

	q.mu.Lock()
	now := q.now()
	if now.Sub(q.windowStart) >= q.window {
		q.count = 0
		q.windowStart = now
	}
	if q.count >= q.max {
		q.mu.Unlock()
		return nil
	}
	q.count++
	q.mu.Unlock()

	return q.inner.Notify(ctx, msg)
}

// Reset clears the current window count, allowing notifications immediately.
func (q *QuotaNotifier) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.count = 0
	q.windowStart = q.now()
}

// Remaining returns the number of notifications still allowed in the current
// window, or -1 if unlimited.
func (q *QuotaNotifier) Remaining() int {
	if q.max <= 0 {
		return -1
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	now := q.now()
	if now.Sub(q.windowStart) >= q.window {
		return q.max
	}
	remaining := q.max - q.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// String returns a human-readable description of the quota configuration.
func (q *QuotaNotifier) String() string {
	return fmt.Sprintf("QuotaNotifier(max=%d, window=%s)", q.max, q.window)
}
