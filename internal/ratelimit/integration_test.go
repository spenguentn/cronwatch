package ratelimit_test

import (
	"context"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/ratelimit"
)

// TestIntegration_MultiJob verifies that independent jobs each receive their
// full burst allocation and refill independently.
func TestIntegration_MultiJob(t *testing.T) {
	l := ratelimit.New(100*time.Millisecond, 2)
	stubs := map[string]*stubNotifier{
		"alpha": {},
		"beta":  {},
	}

	ctx := context.Background()
	for name, stub := range stubs {
		n := ratelimit.NewNotifier(stub, l, name)
		for i := 0; i < 5; i++ {
			_ = n.Notify(ctx, "alert", "body")
		}
	}

	for name, stub := range stubs {
		if got := stub.calls.Load(); got != 2 {
			t.Errorf("job %q: expected 2 calls, got %d", name, got)
		}
	}
}

// TestIntegration_RefillAllowsNewAlerts verifies that after the refill window
// passes, new notifications are accepted again.
func TestIntegration_RefillAllowsNewAlerts(t *testing.T) {
	l := ratelimit.New(50*time.Millisecond, 1)
	stub := &stubNotifier{}
	n := ratelimit.NewNotifier(stub, l, "job1")
	ctx := context.Background()

	_ = n.Notify(ctx, "first", "")
	_ = n.Notify(ctx, "blocked", "") // should be rate-limited

	time.Sleep(60 * time.Millisecond)

	_ = n.Notify(ctx, "after refill", "")

	if got := stub.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls (first + after refill), got %d", got)
	}
}

// TestIntegration_ResetUnblocksJob verifies that Reset immediately restores
// capacity for a throttled job without waiting for the refill interval.
func TestIntegration_ResetUnblocksJob(t *testing.T) {
	l := ratelimit.New(time.Hour, 1) // very slow refill
	stub := &stubNotifier{}
	n := ratelimit.NewNotifier(stub, l, "job1")
	ctx := context.Background()

	_ = n.Notify(ctx, "first", "")
	_ = n.Notify(ctx, "blocked", "")

	if stub.calls.Load() != 1 {
		t.Fatal("expected exactly 1 call before reset")
	}

	l.Reset("job1")
	_ = n.Notify(ctx, "after reset", "")

	if got := stub.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", got)
	}
}
