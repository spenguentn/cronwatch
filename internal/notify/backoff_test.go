package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errTransient = errors.New("transient error")

func TestBackoffNotifier_SucceedsFirstTry(t *testing.T) {
	calls := 0
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		calls++
		return nil
	})
	b := NewBackoffNotifier(inner, 3, time.Millisecond, time.Second)
	if err := b.Notify(context.Background(), Message{Subject: "ok"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestBackoffNotifier_RetriesAndSucceeds(t *testing.T) {
	calls := 0
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		calls++
		if calls < 3 {
			return errTransient
		}
		return nil
	})
	b := NewBackoffNotifier(inner, 5, time.Millisecond, 10*time.Millisecond)
	if err := b.Notify(context.Background(), Message{Subject: "retry"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestBackoffNotifier_AllAttemptsFail(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return errTransient
	})
	b := NewBackoffNotifier(inner, 3, time.Millisecond, 10*time.Millisecond)
	err := b.Notify(context.Background(), Message{Subject: "fail"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errTransient) {
		t.Fatalf("expected wrapped errTransient, got: %v", err)
	}
}

func TestBackoffNotifier_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return errTransient
	})
	b := NewBackoffNotifier(inner, 3, time.Millisecond, 10*time.Millisecond)
	err := b.Notify(ctx, Message{Subject: "cancelled"})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestBackoffNotifier_NilInnerIsNoop(t *testing.T) {
	b := NewBackoffNotifier(nil, 3, time.Millisecond, time.Second)
	if err := b.Notify(context.Background(), Message{Subject: "noop"}); err != nil {
		t.Fatalf("expected nil for nil inner, got: %v", err)
	}
}

func TestBackoffNotifier_DelayCapApplied(t *testing.T) {
	b := NewBackoffNotifier(nil, 1, 100*time.Millisecond, 50*time.Millisecond)
	delay := b.delayFor(10)
	if delay != 50*time.Millisecond {
		t.Fatalf("expected delay capped at 50ms, got %v", delay)
	}
}

func TestBackoffNotifier_SleepCancelledAborts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		calls++
		return errTransient
	})
	b := NewBackoffNotifier(inner, 5, time.Millisecond, time.Second)
	b.sleep = func(c context.Context, d time.Duration) error {
		cancel()
		return c.Err()
	}
	err := b.Notify(ctx, Message{Subject: "abort"})
	if err == nil {
		t.Fatal("expected error after sleep cancel")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call before abort, got %d", calls)
	}
}
