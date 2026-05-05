package notify

import (
	"context"
)

// notifierFunc is a test helper that adapts a plain function to the Notifier
// interface. Defined in the internal test package so it is available to all
// *_test.go files within the package without being exported.
type notifierFunc func(ctx context.Context, msg Message) error

func (f notifierFunc) Notify(ctx context.Context, msg Message) error {
	return f(ctx, msg)
}
