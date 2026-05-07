package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/internal/notify"
)

func TestIntegration_SeenSuppressesDuplicateWebhook(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	sn := notify.NewSeenNotifier(wh, time.Minute)

	msg := notify.Message{Subject: "backup.missed", Body: "nightly backup did not run"}
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := sn.Notify(ctx, msg); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}

	if n := calls.Load(); n != 1 {
		t.Fatalf("expected 1 webhook call, got %d", n)
	}
}

func TestIntegration_SeenAllowsAfterWindowExpiry(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	sn := notify.NewSeenNotifier(wh, 50*time.Millisecond)

	msg := notify.Message{Subject: "sync.drift", Body: "drift exceeded threshold"}
	ctx := context.Background()

	_ = sn.Notify(ctx, msg)
	time.Sleep(100 * time.Millisecond)
	_ = sn.Notify(ctx, msg)

	if n := calls.Load(); n != 2 {
		t.Fatalf("expected 2 webhook calls after expiry, got %d", n)
	}
}
