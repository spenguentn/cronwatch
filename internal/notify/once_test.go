package notify

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestOnceNotifier_ForwardsFirstCall(t *testing.T) {
	var count atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count.Add(1)
		return nil
	})

	on := NewOnceNotifier(inner)
	msg := Message{Subject: "job.missed"}

	if err := on.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.Load() != 1 {
		t.Fatalf("expected 1 delivery, got %d", count.Load())
	}
}

func TestOnceNotifier_DropsDuplicate(t *testing.T) {
	var count atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count.Add(1)
		return nil
	})

	on := NewOnceNotifier(inner)
	msg := Message{Subject: "job.missed"}

	_ = on.Notify(context.Background(), msg)
	_ = on.Notify(context.Background(), msg)
	_ = on.Notify(context.Background(), msg)

	if count.Load() != 1 {
		t.Fatalf("expected 1 delivery, got %d", count.Load())
	}
	if on.Seen() != 1 {
		t.Fatalf("expected Seen()=1, got %d", on.Seen())
	}
}

func TestOnceNotifier_IndependentSubjects(t *testing.T) {
	var count atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count.Add(1)
		return nil
	})

	on := NewOnceNotifier(inner)
	_ = on.Notify(context.Background(), Message{Subject: "a"})
	_ = on.Notify(context.Background(), Message{Subject: "b"})
	_ = on.Notify(context.Background(), Message{Subject: "a"})

	if count.Load() != 2 {
		t.Fatalf("expected 2 deliveries, got %d", count.Load())
	}
	if on.Seen() != 2 {
		t.Fatalf("expected Seen()=2, got %d", on.Seen())
	}
}

func TestOnceNotifier_ResetAllowsResend(t *testing.T) {
	var count atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count.Add(1)
		return nil
	})

	on := NewOnceNotifier(inner)
	msg := Message{Subject: "job.drift"}

	_ = on.Notify(context.Background(), msg)
	on.Reset()
	_ = on.Notify(context.Background(), msg)

	if count.Load() != 2 {
		t.Fatalf("expected 2 deliveries after reset, got %d", count.Load())
	}
	if on.Seen() != 1 {
		t.Fatalf("expected Seen()=1 after reset+one call, got %d", on.Seen())
	}
}

func TestOnceNotifier_NilInnerIsNoop(t *testing.T) {
	on := NewOnceNotifier(nil)
	if err := on.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("expected no error with nil inner, got %v", err)
	}
}
