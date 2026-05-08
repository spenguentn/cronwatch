package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_StickyDeliversToWebhook(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 0)
	sn := notify.NewStickyNotifier(wh)

	msg := notify.Message{Subject: "backup.daily", Body: "missed run detected"}
	if err := sn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if err := sn.Resend(context.Background(), "backup.daily"); err != nil {
		t.Fatalf("Resend: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 webhook calls, got %d", got)
	}
}

func TestIntegration_StickyPayloadIntact(t *testing.T) {
	type payload struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	var received payload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 0)
	sn := notify.NewStickyNotifier(wh)

	_ = sn.Notify(context.Background(), notify.Message{Subject: "deploy.check", Body: "drift exceeded"})
	_ = sn.Resend(context.Background(), "deploy.check")

	if received.Subject != "deploy.check" {
		t.Errorf("unexpected subject: %s", received.Subject)
	}
	if received.Body != "drift exceeded" {
		t.Errorf("unexpected body: %s", received.Body)
	}
}
