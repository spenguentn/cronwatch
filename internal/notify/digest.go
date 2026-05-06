package notify

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// DigestNotifier accumulates messages over a fixed window and delivers a single
// digest summary to the inner Notifier when the window elapses or Flush is
// called explicitly.
type DigestNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	window   time.Duration
	pending  []Message
	stopCh   chan struct{}
	wg       sync.WaitGroup
	nowFn    func() time.Time
}

// NewDigestNotifier creates a DigestNotifier that flushes to inner every window
// duration. A background goroutine is started; call Close to stop it.
func NewDigestNotifier(inner Notifier, window time.Duration) *DigestNotifier {
	d := &DigestNotifier{
		inner:  inner,
		window: window,
		stopCh: make(chan struct{}),
		nowFn:  time.Now,
	}
	d.wg.Add(1)
	go d.loop()
	return d
}

// Notify queues msg for the next digest flush.
func (d *DigestNotifier) Notify(ctx context.Context, msg Message) error {
	if d.inner == nil {
		return nil
	}
	d.mu.Lock()
	d.pending = append(d.pending, msg)
	d.mu.Unlock()
	return nil
}

// Flush immediately delivers a digest of all pending messages.
func (d *DigestNotifier) Flush(ctx context.Context) error {
	d.mu.Lock()
	msgs := d.pending
	d.pending = nil
	d.mu.Unlock()
	if len(msgs) == 0 || d.inner == nil {
		return nil
	}
	return d.inner.Notify(ctx, buildDigest(msgs))
}

// Close stops the background flush loop.
func (d *DigestNotifier) Close() {
	close(d.stopCh)
	d.wg.Wait()
}

func (d *DigestNotifier) loop() {
	defer d.wg.Done()
	ticker := time.NewTicker(d.window)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = d.Flush(context.Background())
		case <-d.stopCh:
			return
		}
	}
}

func buildDigest(msgs []Message) Message {
	var sb strings.Builder
	for i, m := range msgs {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		sb.WriteString(m.Body)
	}
	return Message{
		Subject:  fmt.Sprintf("Digest: %d alerts", len(msgs)),
		Body:     sb.String(),
		Severity: msgs[0].Severity,
		Meta:     msgs[0].Meta,
	}
}
