package suppress_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/suppress"
)

func TestAllow_FirstCallPermitted(t *testing.T) {
	s := suppress.New(5 * time.Minute)
	if !s.Allow("job-a") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallSuppressed(t *testing.T) {
	s := suppress.New(5 * time.Minute)
	s.Allow("job-a")
	if s.Allow("job-a") {
		t.Fatal("expected second call within cooldown to be suppressed")
	}
}

func TestAllow_AfterCooldown(t *testing.T) {
	now := time.Now()
	s := suppress.New(1 * time.Second)
	// Manually inject a clock by using a small cooldown and sleeping.
	s.Allow("job-b")
	time.Sleep(10 * time.Millisecond)
	// Still within 1s cooldown — should be suppressed.
	if s.Allow("job-b") {
		t.Fatalf("expected suppression within cooldown (elapsed %v)", time.Since(now))
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	s := suppress.New(5 * time.Minute)
	s.Allow("job-a")
	if !s.Allow("job-b") {
		t.Fatal("expected independent key to be allowed")
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	s := suppress.New(5 * time.Minute)
	s.Allow("job-a")
	s.Reset("job-a")
	if !s.Allow("job-a") {
		t.Fatal("expected allow after reset")
	}
}

// stubNotify counts Notify invocations.
type stubNotify struct {
	calls atomic.Int32
	err   error
}

func (s *stubNotify) Notify(_ context.Context, _, _ string) error {
	s.calls.Add(1)
	return s.err
}

func TestNotifier_SuppressDuplicates(t *testing.T) {
	stub := &stubNotify{}
	n := suppress.NewNotifier(stub, suppress.New(5*time.Minute))

	_ = n.Notify(context.Background(), "job-a", "body")
	_ = n.Notify(context.Background(), "job-a", "body")
	_ = n.Notify(context.Background(), "job-a", "body")

	if got := stub.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestNotifier_PropagatesError(t *testing.T) {
	want := errors.New("notify failed")
	stub := &stubNotify{err: want}
	n := suppress.NewNotifier(stub, suppress.New(5*time.Minute))

	if got := n.Notify(context.Background(), "job-a", "body"); !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
