package notify

import (
	"context"
	"fmt"
	"sync"
)

// StickyNotifier remembers the last message sent per subject and re-sends it
// automatically whenever Resend is called. This is useful for ensuring that
// on-call recipients who join late still receive the most recent alert state.
type StickyNotifier struct {
	mu    sync.Mutex
	inner Notifier
	last  map[string]Message
}

// NewStickyNotifier wraps inner, retaining the most recent message per subject.
func NewStickyNotifier(inner Notifier) *StickyNotifier {
	return &StickyNotifier{
		inner: inner,
		last:  make(map[string]Message),
	}
}

// Notify forwards the message to inner and stores it as the latest for the subject.
func (s *StickyNotifier) Notify(ctx context.Context, msg Message) error {
	if s.inner == nil {
		return nil
	}
	if err := s.inner.Notify(ctx, msg); err != nil {
		return err
	}
	s.mu.Lock()
	s.last[msg.Subject] = msg
	s.mu.Unlock()
	return nil
}

// Resend re-delivers the last known message for the given subject.
// Returns an error if no message has been recorded for that subject.
func (s *StickyNotifier) Resend(ctx context.Context, subject string) error {
	s.mu.Lock()
	msg, ok := s.last[subject]
	s.mu.Unlock()
	if !ok {
		return fmt.Errorf("sticky: no recorded message for subject %q", subject)
	}
	if s.inner == nil {
		return nil
	}
	return s.inner.Notify(ctx, msg)
}

// Last returns the most recently delivered message for the given subject and
// whether one exists.
func (s *StickyNotifier) Last(subject string) (Message, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg, ok := s.last[subject]
	return msg, ok
}

// Reset clears all stored messages.
func (s *StickyNotifier) Reset() {
	s.mu.Lock()
	s.last = make(map[string]Message)
	s.mu.Unlock()
}
