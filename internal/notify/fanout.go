package notify

import (
	"context"
	"errors"
	"fmt"
)

// FanoutNotifier sends a notification to all registered notifiers concurrently
// and collects any errors that occur. Unlike MultiNotifier which is sequential,
// FanoutNotifier dispatches in parallel and waits for all to complete.
type FanoutNotifier struct {
	notifiers []Notifier
}

// NewFanoutNotifier creates a FanoutNotifier that dispatches to all provided notifiers
// concurrently. At least one notifier must be provided.
func NewFanoutNotifier(notifiers ...Notifier) (*FanoutNotifier, error) {
	if len(notifiers) == 0 {
		return nil, errors.New("fanout: at least one notifier required")
	}
	filtered := make([]Notifier, 0, len(notifiers))
	for _, n := range notifiers {
		if n != nil {
			filtered = append(filtered, n)
		}
	}
	if len(filtered) == 0 {
		return nil, errors.New("fanout: all provided notifiers are nil")
	}
	return &FanoutNotifier{notifiers: filtered}, nil
}

// Notify sends msg to all notifiers concurrently and returns a combined error
// if any notifier fails. The context is forwarded to each notifier.
func (f *FanoutNotifier) Notify(ctx context.Context, msg Message) error {
	type result struct {
		idx int
		err error
	}

	results := make(chan result, len(f.notifiers))

	for i, n := range f.notifiers {
		go func(idx int, notifier Notifier) {
			err := notifier.Notify(ctx, msg)
			results <- result{idx: idx, err: err}
		}(i, n)
	}

	var errs []error
	for range f.notifiers {
		r := <-results
		if r.err != nil {
			errs = append(errs, fmt.Errorf("notifier[%d]: %w", r.idx, r.err))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
