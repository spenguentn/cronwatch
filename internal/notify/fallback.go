package notify

import (
	"context"
	"fmt"
)

// FallbackNotifier delivers a message to a primary notifier and, if that
// fails, falls back to one or more secondary notifiers in order until one
// succeeds. The error from the last attempted notifier is returned only if
// all notifiers fail.
type FallbackNotifier struct {
	primary   Notifier
	fallbacks []Notifier
}

// NewFallbackNotifier returns a FallbackNotifier that tries primary first,
// then each fallback in order on failure.
func NewFallbackNotifier(primary Notifier, fallbacks ...Notifier) *FallbackNotifier {
	return &FallbackNotifier{
		primary:   primary,
		fallbacks: fallbacks,
	}
}

// Notify sends the message to the primary notifier. On failure it tries each
// fallback in order, returning nil as soon as one succeeds. If all notifiers
// fail, a combined error is returned.
func (f *FallbackNotifier) Notify(ctx context.Context, msg Message) error {
	if f.primary != nil {
		if err := f.primary.Notify(ctx, msg); err == nil {
			return nil
		}
	}

	var lastErr error
	for i, fb := range f.fallbacks {
		if fb == nil {
			continue
		}
		if err := fb.Notify(ctx, msg); err == nil {
			return nil
		} else {
			lastErr = fmt.Errorf("fallback[%d]: %w", i, err)
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("fallback: all notifiers nil or failed")
}
