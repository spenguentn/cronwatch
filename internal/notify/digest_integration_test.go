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

func TestIntegration_DigestFlushesToWebhook(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	d := notify.NewDigestNotifier(wh, 10*time.Second)
	defer d.Close()

	for i := 0; i < 5; i++ {
		_ = d.Notify(context.Background(), notify.Message{
			Subject:  "alert",
			Body:     "job missed",
			Severity: notify.SeverityWarn,
		})
	}

	if err := d.Flush(context.Background()); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 webhook call (digest), got %d", got)
	}
}

func TestIntegration_DigestWindowDeliversAutomatically(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	d := notify.NewDigestNotifier(wh, 60*time.Millisecond)
	defer d.Close()

	_ = d.Notify(context.Background(), notify.Message{
		Subject:  "cron drift",
		Body:     "drift detected",
		Severity: notify.SeverityError,
	})

	time.Sleep(200 * time.Millisecond)

	if got := atomic.LoadInt32(&calls); got == 0 {
		t.Error("expected automatic digest delivery via window, got none")
	}
}
