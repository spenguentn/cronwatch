package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/cronwatch/internal/notify"
)

func TestIntegration_WatermarkAllowsHighSeverity(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	webhook := notify.NewWebhookNotifier(srv.URL, nil)
	wm := notify.NewWatermarkNotifier(webhook, notify.SeverityError)

	ctx := context.Background()
	// below mark — should be dropped
	_ = wm.Notify(ctx, notify.Message{Subject: "info", Severity: notify.SeverityInfo})
	_ = wm.Notify(ctx, notify.Message{Subject: "warn", Severity: notify.SeverityWarn})
	// at/above mark — should be forwarded
	if err := wm.Notify(ctx, notify.Message{Subject: "error", Severity: notify.SeverityError}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hits.Load() != 1 {
		t.Errorf("expected 1 webhook hit, got %d", hits.Load())
	}
}

func TestIntegration_WatermarkDynamicLowering(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	webhook := notify.NewWebhookNotifier(srv.URL, nil)
	wm := notify.NewWatermarkNotifier(webhook, notify.SeverityError)

	ctx := context.Background()
	// warn is below mark, dropped
	_ = wm.Notify(ctx, notify.Message{Severity: notify.SeverityWarn})
	if hits.Load() != 0 {
		t.Fatalf("expected 0 hits before lowering mark, got %d", hits.Load())
	}

	// lower the mark to warn
	wm.SetMark(notify.SeverityWarn)
	if err := wm.Notify(ctx, notify.Message{Severity: notify.SeverityWarn}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hits.Load() != 1 {
		t.Errorf("expected 1 hit after lowering mark, got %d", hits.Load())
	}
}
