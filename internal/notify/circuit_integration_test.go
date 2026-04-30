package notify_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_CircuitWithWebhook(t *testing.T) {
	var calls atomic.Int32
	var fail atomic.Bool
	fail.Store(true)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if fail.Load() {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 2*time.Second)
	cb := notify.NewCircuitBreaker(wh, 3, 50*time.Millisecond)
	ctx := context.Background()
	msg := notify.Message{Subject: "cron drift", Body: "job late by 5m"}

	// exhaust failures to open the circuit
	for i := 0; i < 3; i++ {
		_ = cb.Notify(ctx, msg)
	}

	if !cb.IsOpen() {
		t.Fatal("circuit should be open after 3 failures")
	}

	before := calls.Load()
	err := cb.Notify(ctx, msg)
	if !errors.Is(err, notify.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
	if calls.Load() != before {
		t.Error("webhook should not be called while circuit is open")
	}

	// wait for reset timeout then let the probe succeed
	time.Sleep(60 * time.Millisecond)
	fail.Store(false)

	if err := cb.Notify(ctx, msg); err != nil {
		t.Fatalf("expected success after reset timeout, got: %v", err)
	}
	if cb.IsOpen() {
		t.Error("circuit should be closed after successful probe")
	}
}

func TestIntegration_CircuitReopenOnHalfOpenFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 2*time.Second)
	cb := notify.NewCircuitBreaker(wh, 2, 30*time.Millisecond)
	ctx := context.Background()
	msg := notify.Message{Subject: "missed run", Body: "job did not execute"}

	_ = cb.Notify(ctx, msg)
	_ = cb.Notify(ctx, msg)

	time.Sleep(40 * time.Millisecond)

	// half-open probe still fails
	_ = cb.Notify(ctx, msg)

	if !cb.IsOpen() {
		t.Error("circuit should re-open when half-open probe fails")
	}
}
