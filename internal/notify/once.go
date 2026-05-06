package notify

import (
	"context"
	"sync"
)

// OnceNotifier forwards a message to the inner Notifier exactly once per key.
// Subsequent calls with the same key are silently dropped until Reset is called.
// It is safe for concurrent use.
type OnceNotifier struct {
	inner   Notifier
	mu      sync.Mutex
	seen    map[string]struct{}
}

// NewOnceNotifier returns a OnceNotifier that wraps inner.
// If inner is nil, Notify is a no-op.
func NewOnceNotifier(inner Notifier) *OnceNotifier {
	return &OnceNotifier{
		inner: inner,
		seen:  make(map[string]struct{}),
	}
}

// Notify forwards msg to the inner Notifier the first time a given subject is
// seen. Duplicate subjects are dropped without error.
func (n *OnceNotifier) Notify(ctx context.Context, msg Message) error {
	if n.inner == nil {
		return nil
	}

	n.mu.Lock()
	_, already := n.seen[msg.Subject]
	if !already {
		n.seen[msg.Subject] = struct{}{}
	}
	n.mu.Unlock()

	if already {
		return nil
	}
	return n.inner.Notify(ctx, msg)
}

// Reset clears the set of seen subjects so that all keys may fire once more.
func (n *OnceNotifier) Reset() {
	n.mu.Lock()
	n.seen = make(map[string]struct{})
	n.mu.Unlock()
}

// Seen returns the number of distinct subjects that have already fired.
func (n *OnceNotifier) Seen() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.seen)
}
