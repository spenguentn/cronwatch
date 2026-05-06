package notify

import (
	"context"
	"sync"
	"time"
)

// HedgeNotifier sends the message to the primary notifier and, if it does not
// complete within the hedge delay, concurrently fires the same message to the
// secondary notifier. The first successful response wins; the other goroutine
// is abandoned (its context is cancelled).
type HedgeNotifier struct {
	primary   Notifier
	secondary Notifier
	delay     time.Duration
}

// NewHedgeNotifier returns a HedgeNotifier. delay is the duration to wait for
// primary before also sending to secondary. A zero or negative delay defaults
// to 200ms.
func NewHedgeNotifier(primary, secondary Notifier, delay time.Duration) *HedgeNotifier {
	if delay <= 0 {
		delay = 200 * time.Millisecond
	}
	return &HedgeNotifier{primary: primary, secondary: secondary, delay: delay}
}

// Notify sends msg to primary immediately. If primary has not returned within
// the hedge delay, secondary is also invoked. The first nil-error result is
// returned; if both fail the primary error is returned.
func (h *HedgeNotifier) Notify(ctx context.Context, msg Message) error {
	type result struct {
		err error
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch := make(chan result, 2)

	send := func(n Notifier) {
		if n == nil {
			ch <- result{err: ErrNilNotifier}
			return
		}
		ch <- result{err: n.Notify(ctx, msg)}
	}

	go send(h.primary)

	var once sync.Once
	timer := time.NewTimer(h.delay)
	defer timer.Stop()

	var primaryErr error
	received := 0

	for received < 2 {
		select {
		case <-timer.C:
			once.Do(func() { go send(h.secondary) })
		case r := <-ch:
			received++
			if r.err == nil {
				cancel()
				return nil
			}
			if received == 1 {
				primaryErr = r.err
				// ensure secondary is started if timer hasn't fired yet
				once.Do(func() { go send(h.secondary) })
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return primaryErr
}
