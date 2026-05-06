package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_CheckpointWithWebhook(t *testing.T) {
	received := make(chan struct{}, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- struct{}{}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	cp := notify.NewCheckpointNotifier(wh)

	msg := notify.Message{Subject: "integration.job", Body: "ok"}
	if err := cp.Notify(context.Background(), msg); err != nil {
		t.Fatalf("Notify failed: %v", err)
	}

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook did not receive request")
	}

	_, ok := cp.LastDelivery("integration.job")
	if !ok {
		t.Fatal("expected checkpoint to be recorded after successful webhook delivery")
	}
}

func TestIntegration_CheckpointPayloadIntact(t *testing.T) {
	var got notify.Message
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	cp := notify.NewCheckpointNotifier(wh)

	sent := notify.Message{Subject: "payload.check", Body: "body text"}
	if err := cp.Notify(context.Background(), sent); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if got.Subject != sent.Subject {
		t.Errorf("subject: got %q, want %q", got.Subject, sent.Subject)
	}
}
