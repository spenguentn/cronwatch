package notify

import (
	"context"
	"sync"
	"time"
)

// BufferNotifier batches notifications and flushes them after a window or
// when the buffer reaches a maximum size. All buffered messages are joined
// and forwarded as a single notification to the inner Notifier.
type BufferNotifier struct {
	inner   Notifier
	window  time.Duration
	maxSize int

	mu      sync.Mutex
	buf     []Message
	timer   *time.Timer
}

// NewBufferNotifier creates a BufferNotifier that accumulates messages for up
// to window duration or until maxSize messages are queued, then flushes them
// as a single batched notification via inner.
func NewBufferNotifier(inner Notifier, window time.Duration, maxSize int) *BufferNotifier {
	if maxSize <= 0 {
		maxSize = 10
	}
	return &BufferNotifier{
		inner:   inner,
		window:  window,
		maxSize: maxSize,
	}
}

// Notify adds msg to the buffer. If the buffer reaches maxSize it is flushed
// immediately; otherwise a timer is started to flush after the window elapses.
func (b *BufferNotifier) Notify(ctx context.Context, msg Message) error {
	b.mu.Lock()
	b.buf = append(b.buf, msg)
	size := len(b.buf)

	if size >= b.maxSize {
		msgs := b.drain()
		b.mu.Unlock()
		return b.flush(ctx, msgs)
	}

	if b.timer == nil {
		b.timer = time.AfterFunc(b.window, func() {
			b.mu.Lock()
			msgs := b.drain()
			b.mu.Unlock()
			if len(msgs) > 0 {
				//nolint:errcheck
				b.flush(context.Background(), msgs)
			}
		})
	}
	b.mu.Unlock()
	return nil
}

// Flush forces an immediate flush of any buffered messages.
func (b *BufferNotifier) Flush(ctx context.Context) error {
	b.mu.Lock()
	msgs := b.drain()
	b.mu.Unlock()
	if len(msgs) == 0 {
		return nil
	}
	return b.flush(ctx, msgs)
}

// drain stops the timer and returns the buffered messages, resetting the buffer.
func (b *BufferNotifier) drain() []Message {
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	msgs := b.buf
	b.buf = nil
	return msgs
}

// flush sends a single batched message derived from msgs to the inner notifier.
func (b *BufferNotifier) flush(ctx context.Context, msgs []Message) error {
	batched := batch(msgs)
	return b.inner.Notify(ctx, batched)
}

// batch merges multiple messages into one, using the highest severity and
// concatenating subjects/bodies.
func batch(msgs []Message) Message {
	if len(msgs) == 1 {
		return msgs[0]
	}
	out := msgs[0]
	for _, m := range msgs[1:] {
		if m.Severity > out.Severity {
			out.Severity = m.Severity
		}
		out.Subject += "; " + m.Subject
		out.Body += "\n---\n" + m.Body
	}
	return out
}
