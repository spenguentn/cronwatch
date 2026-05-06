package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_OnceWithWebhook(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 0)
	on := notify.NewOnceNotifier(wh)

	msg := notify.Message{Subject: "backup.missed", Body: "backup job did not run"}
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if err := on.Notify(ctx, msg); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}

	if hits.Load() != 1 {
		t.Fatalf("expected webhook hit exactly once, got %d", hits.Load())
	}
}

func TestIntegration_OnceResetResends(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 0)
	on := notify.NewOnceNotifier(wh)

	msg := notify.Message{Subject: "deploy.drift"}
	ctx := context.Background()

	_ = on.Notify(ctx, msg)
	_ = on.Notify(ctx, msg) // dropped
	on.Reset()
	_ = on.Notify(ctx, msg) // fires again after reset

	if hits.Load() != 2 {
		t.Fatalf("expected 2 webhook hits, got %d", hits.Load())
	}
}
