package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

type countingNotifier struct {
	calls int
	err   error
}

func (c *countingNotifier) Notify(_ context.Context, _ Message) error {
	c.calls++
	return c.err
}

func TestCircuit_ClosedByDefault(t *testing.T) {
	inner := &countingNotifier{}
	cb := NewCircuitBreaker(inner, 3, time.Minute)

	if cb.IsOpen() {
		t.Fatal("expected circuit to be closed initially")
	}
}

func TestCircuit_OpensAfterMaxFailures(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(inner, 3, time.Minute)
	ctx := context.Background()
	msg := Message{Subject: "test", Body: "body"}

	for i := 0; i < 3; i++ {
		_ = cb.Notify(ctx, msg)
	}

	if !cb.IsOpen() {
		t.Fatal("expected circuit to be open after max failures")
	}
}

func TestCircuit_RejectsWhenOpen(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(inner, 2, time.Minute)
	ctx := context.Background()
	msg := Message{Subject: "test", Body: "body"}

	_ = cb.Notify(ctx, msg)
	_ = cb.Notify(ctx, msg)

	prevCalls := inner.calls
	err := cb.Notify(ctx, msg)

	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if inner.calls != prevCalls {
		t.Error("inner notifier should not be called when circuit is open")
	}
}

func TestCircuit_ResetsOnSuccess(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(inner, 2, time.Minute)
	ctx := context.Background()
	msg := Message{Subject: "test", Body: "body"}

	_ = cb.Notify(ctx, msg)
	inner.err = nil
	_ = cb.Notify(ctx, msg)

	if cb.IsOpen() {
		t.Fatal("expected circuit to close after success")
	}
}

func TestCircuit_HalfOpenTransition(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(inner, 1, 10*time.Millisecond)
	ctx := context.Background()
	msg := Message{Subject: "test", Body: "body"}

	_ = cb.Notify(ctx, msg) // opens circuit

	time.Sleep(20 * time.Millisecond)

	inner.err = nil
	err := cb.Notify(ctx, msg) // half-open probe succeeds
	if err != nil {
		t.Fatalf("unexpected error after reset timeout: %v", err)
	}
	if cb.IsOpen() {
		t.Fatal("circuit should be closed after successful half-open probe")
	}
}

func TestCircuit_ManualReset(t *testing.T) {
	inner := &countingNotifier{err: errors.New("boom")}
	cb := NewCircuitBreaker(inner, 2, time.Hour)
	ctx := context.Background()
	msg := Message{Subject: "test", Body: "body"}

	_ = cb.Notify(ctx, msg)
	_ = cb.Notify(ctx, msg)

	cb.Reset()

	if cb.IsOpen() {
		t.Fatal("expected circuit closed after manual reset")
	}
}
