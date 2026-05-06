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

func TestIntegration_WindowAllowsDuringBusinessHours(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhook := notify.NewWebhookNotifier(server.URL, 5*time.Second)

	// Window covers the current hour — force time into window.
	now := time.Now()
	hour := now.Hour()
	start := time.Duration(hour) * time.Hour
	end := start + time.Hour

	win := notify.NewWindowNotifier(webhook, []notify.TimeWindow{{Start: start, End: end}})

	msg := notify.Message{Subject: "cron.daily missed", Body: "drift detected", Severity: notify.SeverityWarn}
	if err := win.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Errorf("expected 1 webhook hit, got %d", hits)
	}
}

func TestIntegration_WindowDropsOutsideHours(t *testing.T) {
	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhook := notify.NewWebhookNotifier(server.URL, 5*time.Second)

	// Window in the past — will never match current time.
	win := notify.NewWindowNotifier(webhook, []notify.TimeWindow{
		{Start: 0, End: time.Nanosecond},
	})

	msg := notify.Message{Subject: "cron.hourly missed", Severity: notify.SeverityError}
	if err := win.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&hits) != 0 {
		t.Errorf("expected 0 webhook hits, got %d", hits)
	}
}
