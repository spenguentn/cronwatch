package notify

import (
	"context"
	"strings"
)

// Notifier is the interface implemented by all notification backends.
type Notifier interface {
	Notify(ctx context.Context, subject, body string) error
}

// FilterFunc decides whether a notification should be delivered.
// It receives the subject and body and returns true if the notification
// should proceed.
type FilterFunc func(subject, body string) bool

// FilterNotifier wraps a Notifier and applies a FilterFunc before
// forwarding the notification. If the filter returns false the call
// is silently dropped and nil is returned.
type FilterNotifier struct {
	inner  Notifier
	filter FilterFunc
}

// NewFilterNotifier creates a FilterNotifier that only delivers
// notifications for which fn returns true.
func NewFilterNotifier(n Notifier, fn FilterFunc) *FilterNotifier {
	if fn == nil {
		fn = func(_, _ string) bool { return true }
	}
	return &FilterNotifier{inner: n, filter: fn}
}

// Notify forwards the notification only when the filter passes.
func (f *FilterNotifier) Notify(ctx context.Context, subject, body string) error {
	if !f.filter(subject, body) {
		return nil
	}
	return f.inner.Notify(ctx, subject, body)
}

// SubjectContainsFilter returns a FilterFunc that passes notifications
// whose subject contains any of the provided substrings.
func SubjectContainsFilter(substrings ...string) FilterFunc {
	return func(subject, _ string) bool {
		for _, s := range substrings {
			if strings.Contains(subject, s) {
				return true
			}
		}
		return false
	}
}

// SeverityFilter returns a FilterFunc that passes only notifications
// whose subject starts with one of the provided severity prefixes
// (e.g. "[ERROR]", "[WARN]").
func SeverityFilter(prefixes ...string) FilterFunc {
	return func(subject, _ string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(subject, p) {
				return true
			}
		}
		return false
	}
}
