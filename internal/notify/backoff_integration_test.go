package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_BackoffRetriesWebhook(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	wh := notify.NewWebhookNotifier(server.URL, 5*time.Second)
	b := notify.NewBackoffNotifier(wh, 5, 5*time.Millisecond, 50*time.Millisecond)

	err := b.Notify(context.Background(), notify.Message{
		Subject:  "integration test",
		Body:     "retrying webhook",
		Severity: notify.SeverityWarn,
	})
	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls to server, got %d", calls.Load())
	}
}

func TestIntegration_BackoffExhaustedReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	wh := notify.NewWebhookNotifier(server.URL, 5*time.Second)
	b := notify.NewBackoffNotifier(wh, 3, 5*time.Millisecond, 20*time.Millisecond)

	err := b.Notify(context.Background(), notify.Message{
		Subject:  "always fails",
		Severity: notify.SeverityError,
	})
	if err == nil {
		t.Fatal("expected error when all attempts fail")
	}
}
