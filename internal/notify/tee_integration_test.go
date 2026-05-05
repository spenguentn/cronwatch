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

func TestIntegration_TeeDeliversToBothWebhooks(t *testing.T) {
	var hitsA, hitsB atomic.Int32

	serverA := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitsA.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer serverA.Close()

	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitsB.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer serverB.Close()

	a := notify.NewWebhookNotifier(serverA.URL, 0)
	b := notify.NewWebhookNotifier(serverB.URL, 0)
	tee := notify.NewTeeNotifier(a, b)

	msg := notify.Message{Subject: "cron missed", Body: "backup job", Severity: "warn"}
	if err := tee.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if hitsA.Load() != 1 {
		t.Errorf("serverA received %d requests, want 1", hitsA.Load())
	}
	if hitsB.Load() != 1 {
		t.Errorf("serverB received %d requests, want 1", hitsB.Load())
	}
}

func TestIntegration_TeePrimaryPayloadIntact(t *testing.T) {
	type payload struct {
		Subject string `json:"subject"`
	}

	var received payload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	primary := notify.NewWebhookNotifier(server.URL, 0)
	devNull := notify.NotifierFunc(func(_ context.Context, _ notify.Message) error { return nil })
	tee := notify.NewTeeNotifier(primary, devNull)

	msg := notify.Message{Subject: "drift detected", Body: "job: nightly-sync"}
	if err := tee.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Subject != msg.Subject {
		t.Errorf("got subject %q, want %q", received.Subject, msg.Subject)
	}
}
