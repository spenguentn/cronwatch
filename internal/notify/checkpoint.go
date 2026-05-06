package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CheckpointEntry records the last successful delivery for a given subject.
type CheckpointEntry struct {
	Subject   string
	Delivered time.Time
}

// CheckpointNotifier wraps an inner Notifier and records the timestamp of
// each successful delivery, keyed by message subject. Callers can query
// the last successful delivery time or reset checkpoints.
type CheckpointNotifier struct {
	mu    sync.RWMutex
	inner Notifier
	marks map[string]time.Time
	now   func() time.Time
}

// NewCheckpointNotifier returns a CheckpointNotifier that delegates to inner.
func NewCheckpointNotifier(inner Notifier) *CheckpointNotifier {
	return &CheckpointNotifier{
		inner: inner,
		marks: make(map[string]time.Time),
		now:   time.Now,
	}
}

// Notify forwards the message to the inner notifier. On success the current
// time is stored as the checkpoint for msg.Subject.
func (c *CheckpointNotifier) Notify(ctx context.Context, msg Message) error {
	if c.inner == nil {
		return nil
	}
	if err := c.inner.Notify(ctx, msg); err != nil {
		return fmt.Errorf("checkpoint: inner notify: %w", err)
	}
	c.mu.Lock()
	c.marks[msg.Subject] = c.now()
	c.mu.Unlock()
	return nil
}

// LastDelivery returns the time of the last successful delivery for the given
// subject and true, or the zero time and false if no checkpoint exists.
func (c *CheckpointNotifier) LastDelivery(subject string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.marks[subject]
	return t, ok
}

// Reset clears the checkpoint for the given subject.
func (c *CheckpointNotifier) Reset(subject string) {
	c.mu.Lock()
	delete(c.marks, subject)
	c.mu.Unlock()
}

// Snapshot returns a copy of all current checkpoint entries.
func (c *CheckpointNotifier) Snapshot() []CheckpointEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]CheckpointEntry, 0, len(c.marks))
	for subj, t := range c.marks {
		out = append(out, CheckpointEntry{Subject: subj, Delivered: t})
	}
	return out
}
