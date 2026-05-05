package notify

import "context"

// NotifierFunc is a function that implements [Notifier].
// It allows inline notifiers without defining a named type.
type NotifierFunc func(ctx context.Context, msg Message) error

// Notify calls the underlying function.
func (f NotifierFunc) Notify(ctx context.Context, msg Message) error {
	if f == nil {
		return nil
	}
	return f(ctx, msg)
}
