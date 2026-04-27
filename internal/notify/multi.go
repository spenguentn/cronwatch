// Package notify provides notification backends for cronwatch alerts.
package notify

import (
	"context"
	"errors"
	"fmt"
)

// MultiNotifier fans out a notification to multiple Notifier implementations.
// All notifiers are attempted regardless of individual failures; a combined
// error is returned if any notifier fails.
type MultiNotifier struct {
	notifiers []Notifier
}

// Notifier is the common interface implemented by all notification backends.
type Notifier interface {
	Notify(ctx context.Context, subject, message string) error
}

// NewMultiNotifier returns a MultiNotifier that dispatches to each of the
// provided notifiers in order. At least one notifier must be supplied.
func NewMultiNotifier(notifiers ...Notifier) (*MultiNotifier, error) {
	if len(notifiers) == 0 {
		return nil, errors.New("multi notifier: at least one notifier is required")
	}
	return &MultiNotifier{notifiers: notifiers}, nil
}

// Notify sends subject and message to every registered notifier.
// Execution continues even if a notifier returns an error so that all
// backends are attempted. Any errors are joined and returned together.
func (m *MultiNotifier) Notify(ctx context.Context, subject, message string) error {
	var errs []error
	for i, n := range m.notifiers {
		if err := n.Notify(ctx, subject, message); err != nil {
			errs = append(errs, fmt.Errorf("notifier[%d]: %w", i, err))
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
