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

// TestIntegration_DedupeWithWebhook verifies that the DedupeNotifier prevents
// duplicate HTTP calls to a real webhook endpoint within the window.
func TestIntegration_DedupeWithWebhook(t *testing.T) {
	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	d := notify.NewDedupeNotifier(wh, time.Minute)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := d.Notify(ctx, "cron missed", "job backup"); err != nil {
			t.Fatalf("Notify error: %v", err)
		}
	}

	if got := hits.Load(); got != 1 {
		t.Fatalf("expected 1 HTTP hit, got %d", got)
	}
}

// TestIntegration_DedupeFlushResends ensures that after Flush a new HTTP
// request is issued even for a previously seen message.
func TestIntegration_DedupeFlushResends(t *testing.T) {
	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	d := notify.NewDedupeNotifier(wh, time.Minute)

	ctx := context.Background()
	_ = d.Notify(ctx, "drift detected", "job sync")
	d.Flush()
	_ = d.Notify(ctx, "drift detected", "job sync")

	if got := hits.Load(); got != 2 {
		t.Fatalf("expected 2 HTTP hits after flush, got %d", got)
	}
}
