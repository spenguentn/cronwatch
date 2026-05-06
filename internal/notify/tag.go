package notify

import (
	"context"
	"fmt"
)

// TagNotifier wraps an inner Notifier and enriches every outgoing Message
// with a fixed set of metadata tags before forwarding it.
//
// Tags are applied in the order they are provided. If a key already exists in
// the message metadata it is overwritten.
type TagNotifier struct {
	inner Notifier
	tags  map[string]string
}

// NewTagNotifier returns a TagNotifier that merges tags into every message
// before delegating to inner. A nil inner causes Notify to return an error.
func NewTagNotifier(inner Notifier, tags map[string]string) *TagNotifier {
	copy := make(map[string]string, len(tags))
	for k, v := range tags {
		copy[k] = v
	}
	return &TagNotifier{inner: inner, tags: copy}
}

// Notify merges the configured tags into a shallow copy of msg, then forwards
// the enriched message to the inner Notifier.
func (t *TagNotifier) Notify(ctx context.Context, msg Message) error {
	if t.inner == nil {
		return fmt.Errorf("tag notifier: inner notifier is nil")
	}

	enriched := msg
	merged := make(map[string]string, len(msg.Meta)+len(t.tags))
	for k, v := range msg.Meta {
		merged[k] = v
	}
	for k, v := range t.tags {
		merged[k] = v
	}
	enriched.Meta = merged

	return t.inner.Notify(ctx, enriched)
}
