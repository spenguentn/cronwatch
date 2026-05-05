package notify

import (
	"context"
	"fmt"
)

// TeeNotifier sends every message to two notifiers: primary and secondary.
// The primary result is always returned. Secondary errors are collected but
// do not affect the primary outcome; they are returned as a combined error
// only when the primary succeeds and the secondary fails.
type TeeNotifier struct {
	primary   Notifier
	secondary Notifier
}

// NewTeeNotifier creates a TeeNotifier that delivers to both primary and
// secondary. If either is nil it is silently skipped.
func NewTeeNotifier(primary, secondary Notifier) *TeeNotifier {
	return &TeeNotifier{
		primary:   primary,
		secondary: secondary,
	}
}

// Notify sends msg to both notifiers. The primary error takes precedence.
// If the primary succeeds but the secondary fails, the secondary error is
// returned so callers can log or handle it without losing the primary result.
func (t *TeeNotifier) Notify(ctx context.Context, msg Message) error {
	var primaryErr error
	if t.primary != nil {
		primaryErr = t.primary.Notify(ctx, msg)
	}

	var secondaryErr error
	if t.secondary != nil {
		secondaryErr = t.secondary.Notify(ctx, msg)
	}

	if primaryErr != nil {
		return primaryErr
	}
	if secondaryErr != nil {
		return fmt.Errorf("tee secondary: %w", secondaryErr)
	}
	return nil
}
