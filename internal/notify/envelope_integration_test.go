package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/notify"
)

func TestIntegration_EnvelopeWithWebhook(t *testing.T) {
	var mu sync.Mutex
	var received map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		received = payload
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL)
	en := notify.NewEnvelopeNotifier(wh,
		notify.WithSource("integration-test"),
		notify.WithHostnameFunc(func() string { return "ci-runner" }),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msg := notify.Message{
		Subject:  "backup: missed run",
		Body:     "expected at 02:00, not seen",
		Severity: notify.SeverityWarn,
	}
	if err := en.Notify(ctx, msg); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if received == nil {
		t.Fatal("webhook did not receive payload")
	}
	meta, ok := received["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta field missing or wrong type in payload: %v", received)
	}
	if got := meta["source"]; got != "integration-test" {
		t.Errorf("source: want %q, got %v", "integration-test", got)
	}
	if got := meta["host"]; got != "ci-runner" {
		t.Errorf("host: want %q, got %v", "ci-runner", got)
	}
	if _, ok := meta["sent_at"]; !ok {
		t.Error("sent_at missing from envelope metadata")
	}
}

func TestIntegration_EnvelopePreservesUserMeta(t *testing.T) {
	var mu sync.Mutex
	var received map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		received = payload
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL)
	en := notify.NewEnvelopeNotifier(wh, notify.WithSource("should-be-ignored"))

	ctx := context.Background()
	msg := notify.Message{
		Subject: "test",
		Meta:    map[string]string{"source": "user-provided"},
	}
	_ = en.Notify(ctx, msg)

	mu.Lock()
	defer mu.Unlock()

	meta, _ := received["meta"].(map[string]any)
	if got := meta["source"]; got != "user-provided" {
		t.Errorf("user meta should be preserved: got %v", got)
	}
}
