package notify

import (
	"context"
	"strings"
)

// TransformFunc is a function that transforms a Message before delivery.
type TransformFunc func(msg Message) Message

// TransformNotifier wraps a Notifier and applies a TransformFunc to each
// message before forwarding it to the inner notifier.
type TransformNotifier struct {
	inner     Notifier
	transform TransformFunc
}

// NewTransformNotifier returns a TransformNotifier that applies fn to every
// message before passing it to inner. If fn is nil the message is forwarded
// unchanged.
func NewTransformNotifier(inner Notifier, fn TransformFunc) *TransformNotifier {
	return &TransformNotifier{inner: inner, transform: fn}
}

// Notify applies the transform function to msg and forwards the result.
func (t *TransformNotifier) Notify(ctx context.Context, msg Message) error {
	if t.inner == nil {
		return nil
	}
	out := msg
	if t.transform != nil {
		out = t.transform(msg)
	}
	return t.inner.Notify(ctx, out)
}

// PrefixSubject returns a TransformFunc that prepends prefix to msg.Subject.
func PrefixSubject(prefix string) TransformFunc {
	return func(msg Message) Message {
		msg.Subject = prefix + msg.Subject
		return msg
	}
}

// UpperCaseSubject returns a TransformFunc that upper-cases msg.Subject.
func UpperCaseSubject() TransformFunc {
	return func(msg Message) Message {
		msg.Subject = strings.ToUpper(msg.Subject)
		return msg
	}
}

// AddMeta returns a TransformFunc that merges the supplied key/value pairs
// into msg.Meta, creating the map if necessary. Existing keys are preserved.
func AddMeta(kvs map[string]string) TransformFunc {
	return func(msg Message) Message {
		if msg.Meta == nil {
			msg.Meta = make(map[string]string, len(kvs))
		}
		for k, v := range kvs {
			if _, exists := msg.Meta[k]; !exists {
				msg.Meta[k] = v
			}
		}
		return msg
	}
}
