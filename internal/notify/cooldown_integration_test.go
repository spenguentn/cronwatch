package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/notify"
)

func TestIntegration_CooldownWithWebhook(t *testing.T) {
	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	n := notify.NewCooldownNotifier(wh, 10*time.Minute)

	msg := notify.Message{Subject: "backup.daily", Body: "missed run"}
	ctx := context.Background()

	// First call should reach the webhook.
	if err := n.Notify(ctx, msg); err != nil {
		t.Fatalf("first notify: %v", err)
	}
	// Second call within cooldown should be suppressed.
	if err := n.Notify(ctx, msg); err != nil {
		t.Fatalf("second notify: %v", err)
	}

	if got := hits.Load(); got != 1 {
		t.Fatalf("expected 1 webhook hit, got %d", got)
	}
}

func TestIntegration_CooldownResetResends(t *testing.T) {
	var payloads []notify.Message
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m notify.Message
		_ = json.NewDecoder(r.Body).Decode(&m)
		payloads = append(payloads, m)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	n := notify.NewCooldownNotifier(wh, 10*time.Minute)

	msg := notify.Message{Subject: "db.backup", Body: "alert"}
	ctx := context.Background()

	_ = n.Notify(ctx, msg)
	n.Reset("db.backup")
	_ = n.Notify(ctx, msg)

	if len(payloads) != 2 {
		t.Fatalf("expected 2 webhook deliveries after reset, got %d", len(payloads))
	}
}
