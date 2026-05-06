package notify

import (
	"context"
	"fmt"
)

// LabelNotifier attaches a fixed set of key/value label pairs to every
// message's Meta map before forwarding to the inner Notifier. Unlike
// TagNotifier (which appends to a "tags" slice), LabelNotifier writes
// arbitrary key/value pairs directly into Meta, making them addressable
// by downstream routers or filters.
type LabelNotifier struct {
	inner  Notifier
	labels map[string]string
}

// NewLabelNotifier returns a LabelNotifier that merges labels into each
// message before delegating to inner. Existing Meta keys are NOT
// overwritten; labels are only applied when the key is absent.
func NewLabelNotifier(inner Notifier, labels map[string]string) *LabelNotifier {
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return &LabelNotifier{inner: inner, labels: copy}
}

// Notify merges configured labels into msg.Meta (without overwriting
// existing keys) and forwards the enriched message to the inner Notifier.
func (l *LabelNotifier) Notify(ctx context.Context, msg Message) error {
	if l.inner == nil {
		return fmt.Errorf("label: inner notifier is nil")
	}
	if msg.Meta == nil {
		msg.Meta = make(map[string]string, len(l.labels))
	}
	for k, v := range l.labels {
		if _, exists := msg.Meta[k]; !exists {
			msg.Meta[k] = v
		}
	}
	return l.inner.Notify(ctx, msg)
}
