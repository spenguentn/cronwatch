package notify

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

type countingNotifier struct {
	calls atomic.Int64
	err   error
}

func (c *countingNotifier) Notify(_ context.Context, _ Message) error {
	c.calls.Add(1)
	return c.err
}

func TestFanoutNotifier_NoNotifiers(t *testing.T) {
	_, err := NewFanoutNotifier()
	if err == nil {
		t.Fatal("expected error for empty notifiers")
	}
}

func TestFanoutNotifier_NilNotifiers(t *testing.T) {
	_, err := NewFanoutNotifier(nil, nil)
	if err == nil {
		t.Fatal("expected error when all notifiers are nil")
	}
}

func TestFanoutNotifier_AllSucceed(t *testing.T) {
	a := &countingNotifier{}
	b := &countingNotifier{}
	c := &countingNotifier{}

	f, err := NewFanoutNotifier(a, b, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := Message{Subject: "test", Body: "hello"}
	if err := f.Notify(context.Background(), msg); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	for i, n := range []*countingNotifier{a, b, c} {
		if n.calls.Load() != 1 {
			t.Errorf("notifier[%d]: expected 1 call, got %d", i, n.calls.Load())
		}
	}
}

func TestFanoutNotifier_PartialFailure(t *testing.T) {
	a := &countingNotifier{}
	b := &countingNotifier{err: errors.New("b failed")}
	c := &countingNotifier{}

	f, err := NewFanoutNotifier(a, b, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = f.Notify(context.Background(), Message{Subject: "x", Body: "y"})
	if err == nil {
		t.Fatal("expected error from partial failure")
	}
	if !errors.Is(err, b.err) {
		t.Errorf("expected wrapped b.err in combined error, got: %v", err)
	}
	// a and c should still have been called
	if a.calls.Load() != 1 {
		t.Errorf("expected a to be called once")
	}
	if c.calls.Load() != 1 {
		t.Errorf("expected c to be called once")
	}
}

func TestFanoutNotifier_CancelledContext(t *testing.T) {
	blocking := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
	}))
	defer blocking.Close()

	wh, err := NewWebhookNotifier(blocking.URL, 500*time.Millisecond)
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}

	f, err := NewFanoutNotifier(wh)
	if err != nil {
		t.Fatalf("fanout: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if err := f.Notify(ctx, Message{Subject: "s", Body: "b"}); err == nil {
		t.Fatal("expected error on cancelled context")
	}
}
