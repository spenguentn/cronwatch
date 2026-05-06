package notify

import (
	"context"
	"fmt"
)

// HeaderNotifier wraps an inner Notifier and injects fixed key-value pairs
// into every message's Meta map before forwarding. Existing meta keys are
// never overwritten, preserving values set by earlier pipeline stages.
type HeaderNotifier struct {
	inner   Notifier
	headers map[string]string
}

// NewHeaderNotifier returns a HeaderNotifier that merges headers into each
// message's Meta before delegating to inner. If inner is nil the notifier
// is a no-op. headers may be nil or empty.
func NewHeaderNotifier(inner Notifier, headers map[string]string) *HeaderNotifier {
	h := make(map[string]string, len(headers))
	for k, v := range headers {
		h[k] = v
	}
	return &HeaderNotifier{inner: inner, headers: h}
}

// Notify injects the configured headers into msg.Meta (without overwriting
// existing keys) and forwards the enriched message to the inner Notifier.
func (n *HeaderNotifier) Notify(ctx context.Context, msg Message) error {
	if n.inner == nil {
		return nil
	}
	if len(n.headers) == 0 {
		return n.inner.Notify(ctx, msg)
	}

	merged := make(map[string]string, len(msg.Meta)+len(n.headers))
	for k, v := range n.headers {
		merged[k] = v
	}
	// Existing meta wins over headers.
	for k, v := range msg.Meta {
		merged[k] = v
	}
	msg.Meta = merged

	if err := n.inner.Notify(ctx, msg); err != nil {
		return fmt.Errorf("header notifier: %w", err)
	}
	return nil
}
