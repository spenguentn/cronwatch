package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ExpireNotifier wraps an inner Notifier and silently drops messages
// whose metadata "expires_at" timestamp (set via WithExpiry) has passed.
// Messages without an expiry are always forwarded.
type ExpireNotifier struct {
	mu    sync.Mutex
	inner Notifier
	now   func() time.Time
}

// NewExpireNotifier returns an ExpireNotifier that forwards to inner.
func NewExpireNotifier(inner Notifier) *ExpireNotifier {
	return &ExpireNotifier{
		inner: inner,
		now:   time.Now,
	}
}

// WithExpiry returns a copy of msg with its expiry timestamp stored in metadata.
func WithExpiry(msg Message, t time.Time) Message {
	meta := make(map[string]string, len(msg.Meta)+1)
	for k, v := range msg.Meta {
		meta[k] = v
	}
	meta["expires_at"] = t.UTC().Format(time.RFC3339)
	msg.Meta = meta
	return msg
}

// Notify forwards msg to the inner notifier unless the message has expired.
func (e *ExpireNotifier) Notify(ctx context.Context, msg Message) error {
	if e.inner == nil {
		return nil
	}

	e.mu.Lock()
	now := e.now()
	e.mu.Unlock()

	if raw, ok := msg.Meta["expires_at"]; ok {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return fmt.Errorf("expire: invalid expires_at %q: %w", raw, err)
		}
		if now.After(t) {
			return nil // silently drop
		}
	}

	return e.inner.Notify(ctx, msg)
}
