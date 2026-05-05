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
	c.msgs = append(c.msgs, msg)
	return nil
}

func (c *captureNotifier) received() []Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]Message, len(c.msgs))
	copy(out, c.msgs)
	return out
}

func TestBatchNotifier_FlushOnMaxSize(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBatchNotifier(cap, 10*time.Second, 3)
	ctx := context.Background()

	_ = b.Notify(ctx, Message{Subject: "a", Body: "1"})
	_ = b.Notify(ctx, Message{Subject: "b", Body: "2"})
	// third message triggers flush
	_ = b.Notify(ctx, Message{Subject: "c", Body: "3"})

	time.Sleep(20 * time.Millisecond)
	msgs := cap.received()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 batched delivery, got %d", len(msgs))
	}
	if msgs[0].Subject == "" {
		t.Error("expected non-empty combined subject")
	}
}

func TestBatchNotifier_FlushOnWindow(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBatchNotifier(cap, 50*time.Millisecond, 100)
	ctx := context.Background()

	_ = b.Notify(ctx, Message{Subject: "x", Body: "body"})

	time.Sleep(120 * time.Millisecond)
	msgs := cap.received()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 delivery after window, got %d", len(msgs))
	}
}

func TestBatchNotifier_ManualFlush(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBatchNotifier(cap, 10*time.Second, 100)
	ctx := context.Background()

	_ = b.Notify(ctx, Message{Subject: "m1", Body: "b1"})
	_ = b.Notify(ctx, Message{Subject: "m2", Body: "b2"})
	_ = b.Flush(ctx)

	msgs := cap.received()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 batched delivery, got %d", len(msgs))
	}
}

func TestBatchNotifier_NilInner(t *testing.T) {
	b := NewBatchNotifier(nil, 10*time.Second, 5)
	if err := b.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Errorf("expected nil error with nil inner, got %v", err)
	}
}

func TestBatchNotifier_SingleMessageNotBatched(t *testing.T) {
	cap := &captureNotifier{}
	b := NewBatchNotifier(cap, 50*time.Millisecond, 5)
	ctx := context.Background()

	_ = b.Notify(ctx, Message{Subject: "only", Body: "one"})
	_ = b.Flush(ctx)

	msgs := cap.received()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Subject != "only" {
		t.Errorf("expected subject 'only', got %q", msgs[0].Subject)
	}
}
