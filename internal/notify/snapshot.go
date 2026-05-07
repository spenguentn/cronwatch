package notify

import (
	"context"
	"sync"
	"time"
)

// SnapshotEntry records a single notification attempt.
type SnapshotEntry struct {
	Subject   string
	Body      string
	Severity  string
	Timestamp time.Time
	Err       error
}

// SnapshotNotifier wraps an inner Notifier and keeps an in-memory
// snapshot of the most recent N notification attempts (successes and
// failures alike). Useful for debugging pipelines and integration tests.
type SnapshotNotifier struct {
	mu      sync.Mutex
	inner   Notifier
	entries []SnapshotEntry
	cap     int
}

// NewSnapshotNotifier returns a SnapshotNotifier that retains up to cap
// entries. If cap <= 0 it defaults to 64.
func NewSnapshotNotifier(inner Notifier, cap int) *SnapshotNotifier {
	if cap <= 0 {
		cap = 64
	}
	return &SnapshotNotifier{inner: inner, cap: cap}
}

// Notify forwards the message to the inner notifier and records the
// attempt regardless of outcome.
func (s *SnapshotNotifier) Notify(ctx context.Context, msg Message) error {
	var err error
	if s.inner != nil {
		err = s.inner.Notify(ctx, msg)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := SnapshotEntry{
		Subject:   msg.Subject,
		Body:      msg.Body,
		Severity:  msg.Severity,
		Timestamp: time.Now(),
		Err:       err,
	}
	if len(s.entries) >= s.cap {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, entry)
	return err
}

// Entries returns a copy of all recorded entries.
func (s *SnapshotNotifier) Entries() []SnapshotEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]SnapshotEntry, len(s.entries))
	copy(out, s.entries)
	return out
}

// Reset clears all recorded entries.
func (s *SnapshotNotifier) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = s.entries[:0]
}
