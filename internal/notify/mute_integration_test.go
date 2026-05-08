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

func TestIntegration_MuteBlocksWebhookDuringWindow(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := notify.NewWebhookNotifier(server.URL)
	mn := notify.NewMuteNotifier(wh)
	mn.Mute(10 * time.Minute)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := mn.Notify(ctx, notify.Message{Subject: "cron-job", Body: "missed"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if hits.Load() != 0 {
		t.Fatalf("expected 0 webhook hits during mute, got %d", hits.Load())
	}
}

func TestIntegration_MuteUnmuteResumesDelivery(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := notify.NewWebhookNotifier(server.URL)
	mn := notify.NewMuteNotifier(wh)
	mn.Mute(10 * time.Minute)

	ctx := context.Background()
	_ = mn.Notify(ctx, notify.Message{Subject: "suppressed"})

	mn.Unmute()
	if err := mn.Notify(ctx, notify.Message{Subject: "delivered"}); err != nil {
		t.Fatalf("unexpected error after unmute: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("expected 1 webhook hit after unmute, got %d", hits.Load())
	}
}
