package notify

import (
	"context"
	"fmt"
)

// SequenceNotifier delivers a message through a chain of notifiers in order.
// Unlike FanoutNotifier, it stops on the first error rather than attempting all.
// This is useful when notifiers represent fallback steps (e.g. try Slack, then
// email, then PagerDuty) and you want to stop as soon as one succeeds or fail
// fast on a hard error.
type SequenceNotifier struct {
	steps    []Notifier
	stopOnOK bool
}

// SequenceOption configures a SequenceNotifier.
type SequenceOption func(*SequenceNotifier)

// StopOnFirstSuccess causes the sequence to stop after the first notifier
// that returns a nil error, rather than continuing through all steps.
func StopOnFirstSuccess() SequenceOption {
	return func(s *SequenceNotifier) {
		s.stopOnOK = true
	}
}

// NewSequenceNotifier creates a notifier that runs each step in order.
// Nil entries in steps are silently skipped.
func NewSequenceNotifier(steps []Notifier, opts ...SequenceOption) *SequenceNotifier {
	s := &SequenceNotifier{}
	for _, n := range steps {
		if n != nil {
			s.steps = append(s.steps, n)
		}
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Notify sends msg through each step in order. If stopOnOK is set it returns
// nil as soon as one step succeeds. Otherwise it returns the last non-nil
// error encountered, or nil if all steps succeeded.
func (s *SequenceNotifier) Notify(ctx context.Context, msg Message) error {
	if len(s.steps) == 0 {
		return nil
	}
	var lastErr error
	for i, n := range s.steps {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("sequence step %d: context cancelled: %w", i, err)
		}
		if err := n.Notify(ctx, msg); err != nil {
			lastErr = fmt.Errorf("sequence step %d: %w", i, err)
			continue
		}
		if s.stopOnOK {
			return nil
		}
	}
	return lastErr
}
