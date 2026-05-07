package notify

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// SeenNotifier suppresses messages whose content hash has been seen within a
// rolling time window. Unlike DedupeNotifier (which keys on subject alone),
// SeenNotifier hashes the full message body so identical subjects with
// different bodies are treated as distinct.
type SeenNotifier struct {
	mu      sync.Mutex
	inner   Notifier
	window  time.Duration
	seen    map[string]time.Time
	nowFunc func() time.Time
}

// NewSeenNotifier wraps inner, suppressing any message whose (subject+body)
// hash has already been forwarded within window.
func NewSeenNotifier(inner Notifier, window time.Duration) *SeenNotifier {
	return &SeenNotifier{
		inner:   inner,
		window:  window,
		seen:    make(map[string]time.Time),
		nowFunc: time.Now,
	}
}

// Notify forwards msg only if its content hash has not been seen within the
// configured window. Expired entries are pruned on each call.
func (s *SeenNotifier) Notify(ctx context.Context, msg Message) error {
	if s.inner == nil {
		return nil
	}

	hash := contentHash(msg.Subject, msg.Body)
	now := s.nowFunc()

	s.mu.Lock()
	s.prune(now)
	_, exists := s.seen[hash]
	if !exists {
		s.seen[hash] = now
	}
	s.mu.Unlock()

	if exists {
		return nil
	}
	return s.inner.Notify(ctx, msg)
}

// Reset clears all seen hashes, allowing all messages to pass through again.
func (s *SeenNotifier) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen = make(map[string]time.Time)
}

// prune removes entries older than the window. Caller must hold s.mu.
func (s *SeenNotifier) prune(now time.Time) {
	for h, t := range s.seen {
		if now.Sub(t) > s.window {
			delete(s.seen, h)
		}
	}
}

func contentHash(subject, body string) string {
	h := sha256.Sum256([]byte(subject + "\x00" + body))
	return fmt.Sprintf("%x", h)
}
