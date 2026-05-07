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

func TestIntegration_CacheSkipsWebhookOnRepeat(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := notify.NewWebhookNotifier(server.URL, 5*time.Second)
	c := notify.NewCacheNotifier(wh, time.Minute)

	msg := notify.Message{Subject: "job/backup missed", Body: "last run exceeded threshold"}
	for i := 0; i < 5; i++ {
		if err := c.Notify(context.Background(), msg); err != nil {
			t.Fatalf("Notify error on iteration %d: %v", i, err)
		}
	}

	if n := atomic.LoadInt32(&hits); n != 1 {
		t.Fatalf("expected 1 webhook hit, got %d", n)
	}
}

func TestIntegration_CacheResendsAfterTTL(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh := notify.NewWebhookNotifier(server.URL, 5*time.Second)
	c := notify.NewCacheNotifier(wh, 60*time.Millisecond)

	msg := notify.Message{Subject: "job/sync drift", Body: "drift exceeded 30s"}
	if err := c.Notify(context.Background(), msg); err != nil {
		t.Fatalf("first Notify error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if err := c.Notify(context.Background(), msg); err != nil {
		t.Fatalf("second Notify error: %v", err)
	}

	if n := atomic.LoadInt32(&hits); n != 2 {
		t.Fatalf("expected 2 webhook hits after TTL expiry, got %d", n)
	}
}
