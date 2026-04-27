// Package suppress provides a suppression window mechanism that prevents
// duplicate alerts from firing within a configurable cooldown period.
package suppress

import (
	"context"
	"sync"
	"time"
)

// Suppressor tracks per-key suppression windows.
type Suppressor struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
	now      func() time.Time
}

// New creates a Suppressor with the given cooldown duration.
func New(cooldown time.Duration) *Suppressor {
	return &Suppressor{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if the key is outside its suppression window and
// records the current time as the last alert time for that key.
func (s *Suppressor) Allow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if t, ok := s.last[key]; ok && now.Sub(t) < s.cooldown {
		return false
	}
	s.last[key] = now
	return true
}

// Reset clears the suppression record for a key, allowing the next
// alert to pass through immediately.
func (s *Suppressor) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.last, key)
}

// Notifier wraps an alert.Notifier and suppresses duplicate notifications.
type Notifier struct {
	inner      Notify
	suppressor *Suppressor
}

// Notify is the minimal interface required by Notifier.
type Notify interface {
	Notify(ctx context.Context, subject, body string) error
}

// NewNotifier returns a Notifier that suppresses repeated calls for the
// same subject within the suppressor's cooldown window.
func NewNotifier(inner Notify, s *Suppressor) *Notifier {
	return &Notifier{inner: inner, suppressor: s}
}

// Notify forwards the notification only if the subject is not suppressed.
func (n *Notifier) Notify(ctx context.Context, subject, body string) error {
	if !n.suppressor.Allow(subject) {
		return nil
	}
	return n.inner.Notify(ctx, subject, body)
}
