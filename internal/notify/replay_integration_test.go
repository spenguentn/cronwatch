package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_ReplayDeliversToWebhook(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	webhook := notify.NewWebhookNotifier(srv.URL, 0)
	replay := notify.NewReplayNotifier(20)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_ = replay.Notify(ctx, notify.Message{Subject: "job.missed", Body: "check failed"})
	}

	if err := replay.Replay(ctx, webhook); err != nil {
		t.Fatalf("replay error: %v", err)
	}
	if got := hits.Load(); got != 3 {
		t.Fatalf("expected 3 webhook calls, got %d", got)
	}
	if replay.Len() != 0 {
		t.Fatal("buffer should be empty after successful replay")
	}
}

func TestIntegration_ReplayRetainsOnWebhookFailure(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	webhook := notify.NewWebhookNotifier(srv.URL, 0)
	replay := notify.NewReplayNotifier(10)
	ctx := context.Background()

	_ = replay.Notify(ctx, notify.Message{Subject: "drift"})
	_ = replay.Notify(ctx, notify.Message{Subject: "miss"})

	err := replay.Replay(ctx, webhook)
	if err == nil {
		t.Fatal("expected error from failing webhook")
	}
	if replay.Len() != 2 {
		t.Fatalf("expected both messages retained, got %d", replay.Len())
	}
}
