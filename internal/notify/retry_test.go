package notify_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/notify"
)

type stubNotifier struct {
	calls  atomic.Int32
	failN  int
	err    error
}

func (s *stubNotifier) Notify(_ context.Context, _, _, _ string) error {
	n := int(s.calls.Add(1))
	if n <= s.failN {
		return s.err
	}
	return nil
}

func TestRetryNotifier_SucceedsFirstTry(t *testing.T) {
	stub := &stubNotifier{}
	r := notify.NewRetryNotifier(stub, 3, time.Millisecond)
	if err := r.Notify(context.Background(), "warn", "j", "m"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls.Load() != 1 {
		t.Errorf("want 1 call, got %d", stub.calls.Load())
	}
}

func TestRetryNotifier_RetriesAndSucceeds(t *testing.T) {
	stub := &stubNotifier{failN: 2, err: errors.New("transient")}
	r := notify.NewRetryNotifier(stub, 3, time.Millisecond)
	if err := r.Notify(context.Background(), "warn", "j", "m"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls.Load() != 3 {
		t.Errorf("want 3 calls, got %d", stub.calls.Load())
	}
}

func TestRetryNotifier_AllAttemptsFail(t *testing.T) {
	stub := &stubNotifier{failN: 5, err: errors.New("permanent")}
	r := notify.NewRetryNotifier(stub, 3, time.Millisecond)
	err := r.Notify(context.Background(), "error", "j", "m")
	if err == nil {
		t.Fatal("expected error after all retries")
	}
	if stub.calls.Load() != 3 {
		t.Errorf("want 3 calls, got %d", stub.calls.Load())
	}
}

func TestRetryNotifier_CancelledContext(t *testing.T) {
	stub := &stubNotifier{failN: 5, err: errors.New("fail")}
	r := notify.NewRetryNotifier(stub, 5, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after first failure triggers the delay path.
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := r.Notify(ctx, "warn", "j", "m")
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}
