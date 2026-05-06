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

func TestIntegration_ExpireForwardsActiveMessage(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	en := notify.NewExpireNotifier(wh)

	msg := notify.WithExpiry(
		notify.Message{Subject: "active", Body: "still valid"},
		time.Now().Add(time.Hour),
	)

	if err := en.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits.Load() != 1 {
		t.Errorf("expected 1 webhook hit, got %d", hits.Load())
	}
}

func TestIntegration_ExpireDropsStaleMessage(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	en := notify.NewExpireNotifier(wh)

	msg := notify.WithExpiry(
		notify.Message{Subject: "stale", Body: "too late"},
		time.Now().Add(-time.Minute),
	)

	if err := en.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits.Load() != 0 {
		t.Errorf("expected 0 webhook hits for expired message, got %d", hits.Load())
	}
}
