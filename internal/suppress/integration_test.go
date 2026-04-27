package suppress_test

import (
	"context"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/suppress"
)

// TestIntegration_CooldownExpiry verifies that once the cooldown window
// elapses a suppressed key becomes allowed again.
func TestIntegration_CooldownExpiry(t *testing.T) {
	const cooldown = 50 * time.Millisecond

	stub := &stubNotify{}
	s := suppress.New(cooldown)
	n := suppress.NewNotifier(stub, s)
	ctx := context.Background()

	// First alert — should pass.
	if err := n.Notify(ctx, "cron.daily", "missed"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Immediate retry — should be suppressed.
	_ = n.Notify(ctx, "cron.daily", "missed")

	if got := stub.calls.Load(); got != 1 {
		t.Fatalf("after immediate retry: expected 1 call, got %d", got)
	}

	// Wait for cooldown to expire.
	time.Sleep(cooldown + 10*time.Millisecond)

	// Should be allowed again.
	if err := n.Notify(ctx, "cron.daily", "missed"); err != nil {
		t.Fatalf("unexpected error after cooldown: %v", err)
	}

	if got := stub.calls.Load(); got != 2 {
		t.Fatalf("after cooldown expiry: expected 2 calls, got %d", got)
	}
}

// TestIntegration_MultipleJobsIndependent confirms that suppression state
// for one job does not bleed into another.
func TestIntegration_MultipleJobsIndependent(t *testing.T) {
	stub := &stubNotify{}
	s := suppress.New(5 * time.Minute)
	n := suppress.NewNotifier(stub, s)
	ctx := context.Background()

	jobs := []string{"backup", "report", "cleanup"}
	for _, job := range jobs {
		_ = n.Notify(ctx, job, "alert")
	}
	// Fire each a second time — all should be suppressed.
	for _, job := range jobs {
		_ = n.Notify(ctx, job, "alert")
	}

	if got := stub.calls.Load(); int(got) != len(jobs) {
		t.Fatalf("expected %d calls (one per job), got %d", len(jobs), got)
	}
}
