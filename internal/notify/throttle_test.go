package notify

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type countingNotifier struct {
	calls atomic.Int32
	err   error
}

func (c *countingNotifier) Notify(_ context.Context, _, _ string) error {
	c.calls.Add(1)
	return c.err
}

func TestThrottleNotifier_AllowsUpToMax(t *testing.T) {
	inner := &countingNotifier{}
	tn, err := NewThrottleNotifier(inner, time.Minute, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_ = tn.Notify(ctx, "job-a", "msg")
	}

	if got := inner.calls.Load(); got != 3 {
		t.Errorf("expected 3 calls, got %d", got)
	}
}

func TestThrottleNotifier_IndependentSubjects(t *testing.T) {
	inner := &countingNotifier{}
	tn, _ := NewThrottleNotifier(inner, time.Minute, 2)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		_ = tn.Notify(ctx, "job-a", "msg")
		_ = tn.Notify(ctx, "job-b", "msg")
	}

	if got := inner.calls.Load(); got != 4 {
		t.Errorf("expected 4 calls (2 per subject), got %d", got)
	}
}

func TestThrottleNotifier_Reset(t *testing.T) {
	inner := &countingNotifier{}
	tn, _ := NewThrottleNotifier(inner, time.Minute, 1)

	ctx := context.Background()
	_ = tn.Notify(ctx, "job-a", "msg") // allowed
	_ = tn.Notify(ctx, "job-a", "msg") // throttled
	tn.Reset("job-a")
	_ = tn.Notify(ctx, "job-a", "msg") // allowed after reset

	if got := inner.calls.Load(); got != 2 {
		t.Errorf("expected 2 calls after reset, got %d", got)
	}
}

func TestThrottleNotifier_NilInner(t *testing.T) {
	_, err := NewThrottleNotifier(nil, time.Minute, 1)
	if err == nil {
		t.Error("expected error for nil inner notifier")
	}
}

func TestThrottleNotifier_InvalidMax(t *testing.T) {
	inner := &countingNotifier{}
	_, err := NewThrottleNotifier(inner, time.Minute, 0)
	if err == nil {
		t.Error("expected error for max=0")
	}
}

func TestThrottleNotifier_PropagatesError(t *testing.T) {
	expected := errors.New("send failed")
	inner := &countingNotifier{err: expected}
	tn, _ := NewThrottleNotifier(inner, time.Minute, 5)

	err := tn.Notify(context.Background(), "job-a", "msg")
	if !errors.Is(err, expected) {
		t.Errorf("expected propagated error, got %v", err)
	}
}
