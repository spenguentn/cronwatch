package notify

import (
	"context"
	"fmt"
	"time"
)

// EnvelopeNotifier wraps a Notifier and enriches every outbound Message
// with a consistent set of metadata fields before forwarding it.
//
// Fields added (only when not already present):
//   - source:    the configured source label (e.g. "cronwatch")
//   - host:      the value returned by the supplied hostname func
//   - sent_at:   RFC3339 timestamp of when Notify was called
type EnvelopeNotifier struct {
	inner    Notifier
	source   string
	hostname func() string
	now      func() time.Time
}

// EnvelopeOption configures an EnvelopeNotifier.
type EnvelopeOption func(*EnvelopeNotifier)

// WithSource sets the "source" metadata field.
func WithSource(s string) EnvelopeOption {
	return func(e *EnvelopeNotifier) { e.source = s }
}

// WithHostnameFunc overrides the function used to resolve the hostname.
func WithHostnameFunc(fn func() string) EnvelopeOption {
	return func(e *EnvelopeNotifier) { e.hostname = fn }
}

// NewEnvelopeNotifier wraps inner with metadata enrichment.
func NewEnvelopeNotifier(inner Notifier, opts ...EnvelopeOption) *EnvelopeNotifier {
	en := &EnvelopeNotifier{
		inner:  inner,
		source: "cronwatch",
		now:    time.Now,
		hostname: func() string {
			return "unknown"
		},
	}
	for _, o := range opts {
		o(en)
	}
	return en
}

// Notify enriches msg with envelope metadata and forwards to the inner Notifier.
func (e *EnvelopeNotifier) Notify(ctx context.Context, msg Message) error {
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}
	if _, ok := msg.Meta["source"]; !ok {
		msg.Meta["source"] = e.source
	}
	if _, ok := msg.Meta["host"]; !ok {
		msg.Meta["host"] = e.hostname()
	}
	if _, ok := msg.Meta["sent_at"]; !ok {
		msg.Meta["sent_at"] = e.now().UTC().Format(time.RFC3339)
	}
	if e.inner == nil {
		return fmt.Errorf("envelope: inner notifier is nil")
	}
	return e.inner.Notify(ctx, msg)
}
