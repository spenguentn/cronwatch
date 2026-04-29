package notify_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/notify"
)

type countingNotifier struct {
	calls atomic.Int32
}

func (c *countingNotifier) Notify(_ context.Context, _, _ string) error {
	c.calls.Add(1)
	return nil
}

func TestDedupeNotifier_FirstCallPasses(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner, time.Minute)

	if err := d.Notify(context.Background(), "subject", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestDedupeNotifier_DuplicateSuppressed(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner, time.Minute)

	ctx := context.Background()
	_ = d.Notify(ctx, "subject", "body")
	_ = d.Notify(ctx, "subject", "body")
	_ = d.Notify(ctx, "subject", "body")

	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestDedupeNotifier_DifferentMessagesPass(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner, time.Minute)

	ctx := context.Background()
	_ = d.Notify(ctx, "subject-a", "body")
	_ = d.Notify(ctx, "subject-b", "body")
	_ = d.Notify(ctx, "subject-a", "body-different")

	if got := inner.calls.Load(); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestDedupeNotifier_FlushAllowsResend(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner, time.Minute)

	ctx := context.Background()
	_ = d.Notify(ctx, "subject", "body")
	d.Flush()
	_ = d.Notify(ctx, "subject", "body")

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after flush, got %d", got)
	}
}

func TestDedupeNotifier_WindowExpiry(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner, 50*time.Millisecond)

	ctx := context.Background()
	_ = d.Notify(ctx, "subject", "body")
	time.Sleep(80 * time.Millisecond)
	_ = d.Notify(ctx, "subject", "body")

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after window expiry, got %d", got)
	}
}

func TestDedupeNotifier_DefaultWindow(t *testing.T) {
	inner := &countingNotifier{}
	// zero window should fall back to default (5 minutes)
	d := notify.NewDedupeNotifier(inner, 0)

	ctx := context.Background()
	_ = d.Notify(ctx, "x", "y")
	_ = d.Notify(ctx, "x", "y")

	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call with default window, got %d", got)
	}
}
