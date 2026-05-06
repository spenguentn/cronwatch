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

func TestIntegration_HedgeDeliversToPrimary(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	primary := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	secondary := notify.NewWebhookNotifier(srv.URL+"/secondary", 5*time.Second)
	h := notify.NewHedgeNotifier(primary, secondary, 200*time.Millisecond)

	err := h.Notify(context.Background(), notify.Message{Subject: "integration-hedge", Body: "ok"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only primary should have been called (fast path)
	if hits.Load() != 1 {
		t.Fatalf("expected 1 webhook call, got %d", hits.Load())
	}
}

func TestIntegration_HedgeUsesSecondaryOnSlowPrimary(t *testing.T) {
	var secondaryHits atomic.Int32

	// primary blocks indefinitely
	primarySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer primarySrv.Close()

	secondarySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondaryHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer secondarySrv.Close()

	primary := notify.NewWebhookNotifier(primarySrv.URL, 30*time.Second)
	secondary := notify.NewWebhookNotifier(secondarySrv.URL, 5*time.Second)
	h := notify.NewHedgeNotifier(primary, secondary, 30*time.Millisecond)

	err := h.Notify(context.Background(), notify.Message{Subject: "hedge-slow", Body: "fallthrough"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secondaryHits.Load() == 0 {
		t.Fatal("expected secondary to be called")
	}
}
