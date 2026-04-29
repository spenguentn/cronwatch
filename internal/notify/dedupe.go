package notify

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// DedupeNotifier wraps a Notifier and suppresses duplicate messages within a
// configurable window. Two messages are considered duplicates when their
// subject and body hash to the same value.
type DedupeNotifier struct {
	inner  Notifier
	window time.Duration

	mu   sync.Mutex
	seen map[string]time.Time
}

// NewDedupeNotifier returns a DedupeNotifier that forwards each unique
// (subject, body) pair at most once per window duration.
func NewDedupeNotifier(inner Notifier, window time.Duration) *DedupeNotifier {
	if window <= 0 {
		window = 5 * time.Minute
	}
	return &DedupeNotifier{
		inner:  inner,
		window: window,
		seen:   make(map[string]time.Time),
	}
}

// Notify forwards the message only if an identical message has not been sent
// within the configured deduplication window.
func (d *DedupeNotifier) Notify(ctx context.Context, subject, body string) error {
	key := d.hashKey(subject, body)

	d.mu.Lock()
	if last, ok := d.seen[key]; ok && time.Since(last) < d.window {
		d.mu.Unlock()
		return nil
	}
	d.seen[key] = time.Now()
	d.mu.Unlock()

	return d.inner.Notify(ctx, subject, body)
}

// Flush removes all cached entries, allowing all messages to pass through again.
func (d *DedupeNotifier) Flush() {
	d.mu.Lock()
	d.seen = make(map[string]time.Time)
	d.mu.Unlock()
}

func (d *DedupeNotifier) hashKey(subject, body string) string {
	h := sha256.Sum256([]byte(subject + "\x00" + body))
	return fmt.Sprintf("%x", h)
}
