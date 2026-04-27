package ratelimit_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/ratelimit"
)

func TestAllow_BurstConsumed(t *testing.T) {
	l := ratelimit.New(time.Minute, 3)
	now := time.Now()

	for i := 0; i < 3; i++ {
		if !l.Allow("job1", now) {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
	}
	if l.Allow("job1", now) {
		t.Fatal("expected deny after burst exhausted")
	}
}

func TestAllow_RefillOverTime(t *testing.T) {
	l := ratelimit.New(time.Second, 1)
	now := time.Now()

	if !l.Allow("job1", now) {
		t.Fatal("expected first allow")
	}
	if l.Allow("job1", now) {
		t.Fatal("expected deny immediately after")
	}

	// Advance time by 2 seconds — should refill 2 tokens, capped at burst=1.
	later := now.Add(2 * time.Second)
	if !l.Allow("job1", later) {
		t.Fatal("expected allow after refill")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	l := ratelimit.New(time.Minute, 1)
	now := time.Now()

	l.Allow("jobA", now)
	if !l.Allow("jobB", now) {
		t.Fatal("jobB should be independent of jobA")
	}
}

func TestReset_RestoresBurst(t *testing.T) {
	l := ratelimit.New(time.Minute, 2)
	now := time.Now()

	l.Allow("job1", now)
	l.Allow("job1", now)
	if l.Allow("job1", now) {
		t.Fatal("expected deny after burst")
	}

	l.Reset("job1")
	if !l.Allow("job1", now) {
		t.Fatal("expected allow after reset")
	}
}

// stubNotifier counts calls and optionally returns an error.
type stubNotifier struct {
	calls atomic.Int32
	err   error
}

func (s *stubNotifier) Notify(_ context.Context, _, _ string) error {
	s.calls.Add(1)
	return s.err
}

func TestNotifier_RateLimitsDownstream(t *testing.T) {
	l := ratelimit.New(time.Minute, 2)
	stub := &stubNotifier{}
	n := ratelimit.NewNotifier(stub, l, "job1")
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_ = n.Notify(ctx, "subject", "body")
	}

	if got := stub.calls.Load(); got != 2 {
		t.Fatalf("expected 2 downstream calls, got %d", got)
	}
}

func TestNotifier_PropagatesError(t *testing.T) {
	l := ratelimit.New(time.Minute, 5)
	want := errors.New("send failed")
	stub := &stubNotifier{err: want}
	n := ratelimit.NewNotifier(stub, l, "job1")

	got := n.Notify(context.Background(), "s", "b")
	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
