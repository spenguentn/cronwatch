package notify

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestQuotaNotifier_AllowsUpToMax(t *testing.T) {
	var calls atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		calls.Add(1)
		return nil
	})
	q := NewQuotaNotifier(inner, 3, time.Hour)
	ctx := context.Background()
	msg := Message{Subject: "test"}

	for i := 0; i < 5; i++ {
		_ = q.Notify(ctx, msg)
	}

	if got := calls.Load(); got != 3 {
		t.Errorf("expected 3 calls, got %d", got)
	}
}

func TestQuotaNotifier_WindowReset(t *testing.T) {
	var calls atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		calls.Add(1)
		return nil
	})
	q := NewQuotaNotifier(inner, 2, 50*time.Millisecond)
	fakeNow := time.Now()
	q.now = func() time.Time { return fakeNow }
	q.windowStart = fakeNow
	ctx := context.Background()
	msg := Message{Subject: "job"}

	_ = q.Notify(ctx, msg)
	_ = q.Notify(ctx, msg)
	_ = q.Notify(ctx, msg) // should be dropped

	// advance past the window
	fakeNow = fakeNow.Add(100 * time.Millisecond)
	_ = q.Notify(ctx, msg)

	if got := calls.Load(); got != 3 {
		t.Errorf("expected 3 calls after reset, got %d", got)
	}
}

func TestQuotaNotifier_Reset(t *testing.T) {
	var calls atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		calls.Add(1)
		return nil
	})
	q := NewQuotaNotifier(inner, 1, time.Hour)
	ctx := context.Background()
	msg := Message{Subject: "job"}

	_ = q.Notify(ctx, msg)
	_ = q.Notify(ctx, msg) // dropped
	q.Reset()
	_ = q.Notify(ctx, msg) // allowed again

	if got := calls.Load(); got != 2 {
		t.Errorf("expected 2 calls, got %d", got)
	}
}

func TestQuotaNotifier_Remaining(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	q := NewQuotaNotifier(inner, 5, time.Hour)
	ctx := context.Background()
	msg := Message{Subject: "x"}

	if r := q.Remaining(); r != 5 {
		t.Errorf("expected 5 remaining, got %d", r)
	}
	_ = q.Notify(ctx, msg)
	_ = q.Notify(ctx, msg)
	if r := q.Remaining(); r != 3 {
		t.Errorf("expected 3 remaining, got %d", r)
	}
}

func TestQuotaNotifier_UnlimitedWhenZeroMax(t *testing.T) {
	var calls atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		calls.Add(1)
		return nil
	})
	q := NewQuotaNotifier(inner, 0, time.Hour)
	ctx := context.Background()
	msg := Message{Subject: "x"}

	for i := 0; i < 10; i++ {
		_ = q.Notify(ctx, msg)
	}
	if got := calls.Load(); got != 10 {
		t.Errorf("expected 10 calls, got %d", got)
	}
	if r := q.Remaining(); r != -1 {
		t.Errorf("expected -1 for unlimited, got %d", r)
	}
}

func TestQuotaNotifier_NilInnerIsNoop(t *testing.T) {
	q := NewQuotaNotifier(nil, 5, time.Hour)
	err := q.Notify(context.Background(), Message{Subject: "x"})
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestQuotaNotifier_PropagatesError(t *testing.T) {
	want := errors.New("downstream failure")
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return want })
	q := NewQuotaNotifier(inner, 3, time.Hour)
	err := q.Notify(context.Background(), Message{Subject: "x"})
	if !errors.Is(err, want) {
		t.Errorf("expected %v, got %v", want, err)
	}
}
