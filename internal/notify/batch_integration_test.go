package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/notify"
)

func TestIntegration_BatchFlushesToWebhook(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	b := notify.NewBatchNotifier(wh, 10*time.Second, 3)
	ctx := context.Background()

	_ = b.Notify(ctx, notify.Message{Subject: "job-a missed", Body: "details"})
	_ = b.Notify(ctx, notify.Message{Subject: "job-b drift", Body: "details"})
	_ = b.Notify(ctx, notify.Message{Subject: "job-c missed", Body: "details"})

	time.Sleep(50 * time.Millisecond)
	if n := calls.Load(); n != 1 {
		t.Errorf("expected 1 webhook call for batched delivery, got %d", n)
	}
}

func TestIntegration_BatchWindowFlushesToWebhook(t *testing.T) {
	var received []map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		received = append(received, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	b := notify.NewBatchNotifier(wh, 60*time.Millisecond, 100)
	ctx := context.Background()

	_ = b.Notify(ctx, notify.Message{Subject: "drift detected", Body: "job foo"})

	time.Sleep(150 * time.Millisecond)
	if len(received) != 1 {
		t.Fatalf("expected 1 webhook call after window, got %d", len(received))
	}
	if subject, _ := received[0]["subject"].(string); subject == "" {
		t.Error("expected non-empty subject in webhook payload")
	}
}
