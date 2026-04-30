package notify

import (
	"context"
	"sync"
	"testing"
	"time"
)

type captureNotifier struct {
	mu   sync.Mutex
	msgs []Message
}

func (c *captureNotifier) Notify(_ context.Context, msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.c.msgs = append(c.msgs, msg)
	return nil
}

func (c *captureNotifier) received() []Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Message, len(c.msgs))
	copy(out, c.msgs)
	return out
}

func TestBufferNotifier_FlushOnMaxSize(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBufferNotifier(cap, 10*time.Second, 3)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := b.Notify(ctx, Message{Subject: "msg", Body: "body"}); err != nil {
			t.Fatalf("Notify error: %v", err)
		}
	}

	msgs := cap.received()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 batched message, got %d", len(msgs))
	}
}

func TestBufferNotifier_FlushOnWindow(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBufferNotifier(cap, 50*time.Millisecond, 100)
	ctx := context.Background()

	if err := b.Notify(ctx, Message{Subject: "a", Body: "b"}); err != nil {
		t.Fatalf("Notify error: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	msgs := cap.received()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message after window, got %d", len(msgs))
	}
}

func TestBufferNotifier_ManualFlush(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBufferNotifier(cap, 10*time.Second, 100)
	ctx := context.Background()

	_ = b.Notify(ctx, Message{Subject: "x", Body: "y"})
	_ = b.Notify(ctx, Message{Subject: "p", Body: "q"})

	if err := b.Flush(ctx); err != nil {
		t.Fatalf("Flush error: %v", err)
	}

	msgs := cap.received()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 batched message, got %d", len(msgs))
	}
	if msgs[0].Subject != "x; p" {
		t.Errorf("unexpected subject: %q", msgs[0].Subject)
	}
}

func TestBufferNotifier_FlushEmpty(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBufferNotifier(cap, 10*time.Second, 10)
	if err := b.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	if len(cap.received()) != 0 {
		t.Error("expected no messages on empty flush")
	}
}

func TestBatch_SingleMessage(t *testing.T) {
	msgs := []Message{{Subject: "only", Body: "body", Severity: SeverityWarn}}
	out := batch(msgs)
	if out.Subject != "only" {
		t.Errorf("unexpected subject: %q", out.Subject)
	}
}

func TestBatch_PicksHighestSeverity(t *testing.T) {
	msgs := []Message{
		{Subject: "a", Severity: SeverityWarn},
		{Subject: "b", Severity: SeverityError},
	}
	out := batch(msgs)
	if out.Severity != SeverityError {
		t.Errorf("expected SeverityError, got %v", out.Severity)
	}
}
