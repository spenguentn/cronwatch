package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"cronwatch/internal/notify"
)

func TestIntegration_EscalateDeliversToPrimary(t *testing.T) {
	var primaryHits, secondaryHits atomic.Int32
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primaryHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer primary.Close()
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondaryHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer secondary.Close()

	p := notify.NewWebhookNotifier(primary.URL, 2*time.Second)
	s := notify.NewWebhookNotifier(secondary.URL, 2*time.Second)
	e := notify.NewEscalateNotifier(p, s, 0)

	if err := e.Notify(context.Background(), notify.Message{Subject: "ok"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if primaryHits.Load() != 1 {
		t.Fatalf("expected 1 primary hit, got %d", primaryHits.Load())
	}
	if secondaryHits.Load() != 0 {
		t.Fatalf("expected 0 secondary hits, got %d", secondaryHits.Load())
	}
}

func TestIntegration_EscalateOnPrimaryFailure(t *testing.T) {
	var secondaryHits atomic.Int32
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondaryHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer secondary.Close()

	p := notify.NewWebhookNotifier("http://127.0.0.1:0", 500*time.Millisecond)
	s := notify.NewWebhookNotifier(secondary.URL, 2*time.Second)
	e := notify.NewEscalateNotifier(p, s, 0)

	// primary will fail (nothing listening), secondary should receive
	_ = e.Notify(context.Background(), notify.Message{Subject: "escalate-me"})
	if secondaryHits.Load() != 1 {
		t.Fatalf("expected 1 secondary hit after escalation, got %d", secondaryHits.Load())
	}
}
