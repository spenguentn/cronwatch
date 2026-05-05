package notify

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// BatchNotifier collects messages over a time window or until a max count is
// reached, then delivers them as a single combined notification.
type BatchNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	window   time.Duration
	maxSize  int
	pending  []Message
	timer    *time.Timer
	stopOnce sync.Once
	stopCh   chan struct{}
}

// NewBatchNotifier returns a BatchNotifier that accumulates messages and flushes
// them to inner either when maxSize messages are queued or window elapses.
func NewBatchNotifier(inner Notifier, window time.Duration, maxSize int) *BatchNotifier {
	if maxSize <= 0 {
		maxSize = 10
	}
	b := &BatchNotifier{
		inner:   inner,
		window:  window,
		maxSize: maxSize,
		stopCh:  make(chan struct{}),
	}
	return b
}

// Notify enqueues msg. If the batch reaches maxSize it is flushed immediately.
func (b *BatchNotifier) Notify(ctx context.Context, msg Message) error {
	if b.inner == nil {
		return nil
	}
	b.mu.Lock()
	b.pending = append(b.pending, msg)
	count := len(b.pending)
	if b.timer == nil {
		b.timer = time.AfterFunc(b.window, func() { b.flush(context.Background()) })
	}
	b.mu.Unlock()

	if count >= b.maxSize {
		b.flush(ctx)
	}
	return nil
}

// Flush forces immediate delivery of any queued messages.
func (b *BatchNotifier) Flush(ctx context.Context) error {
	b.flush(ctx)
	return nil
}

func (b *BatchNotifier) flush(ctx context.Context) {
	b.mu.Lock()
	if len(b.pending) == 0 {
		b.mu.Unlock()
		return
	}
	batch := make([]Message, len(b.pending))
	copy(batch, b.pending)
	b.pending = b.pending[:0]
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	b.mu.Unlock()

	combined := combineBatch(batch)
	_ = b.inner.Notify(ctx, combined)
}

func combineBatch(msgs []Message) Message {
	if len(msgs) == 1 {
		return msgs[0]
	}
	subjects := make([]string, 0, len(msgs))
	bodies := make([]string, 0, len(msgs))
	for _, m := range msgs {
		subjects = append(subjects, m.Subject)
		bodies = append(bodies, fmt.Sprintf("[%s] %s", m.Subject, m.Body))
	}
	return Message{
		Subject:  fmt.Sprintf("Batch alert (%d): %s", len(msgs), strings.Join(subjects, ", ")),
		Body:     strings.Join(bodies, "\n---\n"),
		Severity: msgs[0].Severity,
		Meta:     msgs[0].Meta,
	}
}
