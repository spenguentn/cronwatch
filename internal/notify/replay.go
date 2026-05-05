package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ReplayNotifier records all messages and can replay them to a target notifier.
// Useful for recovering from transient failures by resending missed alerts.
type ReplayNotifier struct {
	mu       sync.Mutex
	recorded []Message
	cap      int
}

// NewReplayNotifier creates a ReplayNotifier that retains up to capacity messages.
// If capacity is <= 0 it defaults to 100.
func NewReplayNotifier(capacity int) *ReplayNotifier {
	if capacity <= 0 {
		capacity = 100
	}
	return &ReplayNotifier{cap: capacity}
}

// Notify records the message. It never forwards on its own; use Replay to drain.
func (r *ReplayNotifier) Notify(_ context.Context, msg Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.recorded) >= r.cap {
		// drop oldest to make room
		r.recorded = r.recorded[1:]
	}
	r.recorded = append(r.recorded, msg)
	return nil
}

// Replay sends all recorded messages to dst in order.
// Successfully sent messages are removed from the buffer.
// Returns the first error encountered; remaining unsent messages are kept.
func (r *ReplayNotifier) Replay(ctx context.Context, dst Notifier) error {
	r.mu.Lock()
	pending := make([]Message, len(r.recorded))
	copy(pending, r.recorded)
	r.mu.Unlock()

	var failed []Message
	var firstErr error
	for _, msg := range pending {
		if err := dst.Notify(ctx, msg); err != nil {
			failed = append(failed, msg)
			if firstErr == nil {
				firstErr = fmt.Errorf("replay: %w", err)
			}
		}
	}

	r.mu.Lock()
	r.recorded = failed
	r.mu.Unlock()
	return firstErr
}

// Len returns the number of buffered messages.
func (r *ReplayNotifier) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.recorded)
}

// Reset discards all buffered messages.
func (r *ReplayNotifier) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recorded = nil
}

// OldestAge returns how long ago the oldest buffered message was recorded,
// based on its Timestamp metadata field. Returns zero if the buffer is empty.
func (r *ReplayNotifier) OldestAge() time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.recorded) == 0 {
		return 0
	}
	if ts, ok := r.recorded[0].Meta["timestamp"]; ok {
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			return time.Since(t)
		}
	}
	return 0
}
