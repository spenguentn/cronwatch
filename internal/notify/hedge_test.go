package notify

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestHedgeNotifier_PrimarySucceedsBeforeDelay(t *testing.T) {
	var calls atomic.Int32
	fast := NotifierFunc(func(_ context.Context, _ Message) error {
		calls.Add(1)
		return nil
	})
	slow := NotifierFunc(func(_ context.Context, _ Message) error {
		t.Error("secondary should not be called")
		return nil
	})

	h := NewHedgeNotifier(fast, slow, 100*time.Millisecond)
	if err := h.Notify(context.Background(), Message{Subject: "ok"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
}

func TestHedgeNotifier_SecondaryHedgesOnSlowPrimary(t *testing.T) {
	var secondaryCalled atomic.Bool
	blocked := make(chan struct{})

	primary := NotifierFunc(func(ctx context.Context, _ Message) error {
		<-ctx.Done()
		return ctx.Err()
	})
	secondary := NotifierFunc(func(_ context.Context, _ Message) error {
		secondary Called.Store(true)
		close(blocked)
		return nil
	})

	h := NewHedgeNotifier(primary, secondary, 20*time.Millisecond)
	if err := h.Notify(context.Background(), Message{Subject: "hedge"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !secondaryCalled.Load() {
		t.Fatal("secondary was not called")
	}
}

func TestHedgeNotifier_BothFail_ReturnsPrimaryError(t *testing.T) {
	primErr := errors.New("primary failed")
	primary := NotifierFunc(func(_ context.Context, _ Message) error { return primErr })
	secondary := NotifierFunc(func(_ context.Context, _ Message) error { return errors.New("secondary failed") })

	h := NewHedgeNotifier(primary, secondary, 10*time.Millisecond)
	err := h.Notify(context.Background(), Message{Subject: "fail"})
	if !errors.Is(err, primErr) {
		t.Fatalf("expected primary error, got %v", err)
	}
}

func TestHedgeNotifier_DefaultDelay(t *testing.T) {
	h := NewHedgeNotifier(nil, nil, 0)
	if h.delay != 200*time.Millisecond {
		t.Fatalf("expected 200ms default, got %v", h.delay)
	}
}

func TestHedgeNotifier_CancelledContext(t *testing.T) {
	primary := NotifierFunc(func(ctx context.Context, _ Message) error {
		<-ctx.Done()
		return ctx.Err()
	})
	secondary := NotifierFunc(func(_ context.Context, _ Message) error { return nil })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	h := NewHedgeNotifier(primary, secondary, 50*time.Millisecond)
	err := h.Notify(ctx, Message{Subject: "cancelled"})
	if err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
