package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/example/cronwatch/internal/notify"
)

func TestIntegration_SamplingWithWebhook(t *testing.T) {
	var hits atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	wh := notify.NewWebhookNotifier(server.URL)
	// rate=1.0 — every message must reach the webhook
	sn := notify.NewSamplingNotifier(wh, 1.0)

	const total = 10
	for i := 0; i < total; i++ {
		err := sn.Notify(context.Background(), notify.Message{
			Subject: "backup.daily",
			Body:    "drift detected",
		})
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
	}

	if got := hits.Load(); got != total {
		t.Errorf("expected %d webhook hits, got %d", total, got)
	}
}

func TestIntegration_SamplingDropsAllAtZeroRate(t *testing.T) {
	var hits atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	wh := notify.NewWebhookNotifier(server.URL)
	sn := notify.NewSamplingNotifier(wh, 0.0)

	for i := 0; i < 20; i++ {
		err := sn.Notify(context.Background(), notify.Message{
			Subject: "backup.daily",
			Body:    "missed run",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if got := hits.Load(); got != 0 {
		t.Errorf("expected 0 webhook hits at rate 0.0, got %d", got)
	}
}
