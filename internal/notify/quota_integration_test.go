package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/notify"
)

func TestIntegration_QuotaLimitsWebhookCalls(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	q := notify.NewQuotaNotifier(wh, 3, time.Hour)

	ctx := context.Background()
	msg := notify.Message{Subject: "cron-job", Body: "missed run"}

	for i := 0; i < 7; i++ {
		_ = q.Notify(ctx, msg)
	}

	if got := hits.Load(); got != 3 {
		t.Errorf("expected exactly 3 webhook hits, got %d", got)
	}
}

func TestIntegration_QuotaResetResends(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	q := notify.NewQuotaNotifier(wh, 2, time.Hour)

	ctx := context.Background()
	msg := notify.Message{Subject: "cron-job", Body: "drift detected"}

	_ = q.Notify(ctx, msg)
	_ = q.Notify(ctx, msg)
	_ = q.Notify(ctx, msg) // dropped

	q.Reset()

	_ = q.Notify(ctx, msg) // allowed after reset

	if got := hits.Load(); got != 3 {
		t.Errorf("expected 3 webhook hits after reset, got %d", got)
	}
}
