package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CooldownNotifier suppresses repeated notifications for the same subject
// until a per-subject cooldown period has elapsed since the last delivery.
type CooldownNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	cooldown time.Duration
	last     map[string]time.Time
	now      func() time.Time
}

// NewCooldownNotifier wraps inner and enforces a minimum cooldown between
// successive notifications sharing the same subject.
func NewCooldownNotifier(inner Notifier, cooldown time.Duration) *CooldownNotifier {
	return &CooldownNotifier{
		inner:    inner,
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Notify forwards msg to the inner notifier only if the cooldown period has
// elapsed since the last successful delivery for msg.Subject.
func (c *CooldownNotifier) Notify(ctx context.Context, msg Message) error {
	if c.inner == nil {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if last, ok := c.last[msg.Subject]; ok {
		if now.Sub(last) < c.cooldown {
			return nil // still within cooldown window
		}
	}

	if err := c.inner.Notify(ctx, msg); err != nil {
		return fmt.Errorf("cooldown: %w", err)
	}

	c.last[msg.Subject] = now
	return nil
}

// Reset clears the cooldown state for the given subject, allowing the next
// notification to pass through immediately.
func (c *CooldownNotifier) Reset(subject string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.last, subject)
}

// ResetAll clears all tracked cooldown state.
func (c *CooldownNotifier) ResetAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.last = make(map[string]time.Time)
}
