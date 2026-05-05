package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourorg/cronwatch/internal/notify"
)

func TestIntegration_TemplateWithWebhook(t *testing.T) {
	var mu sync.Mutex
	var received []map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		received = append(received, payload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL)
	tn, err := notify.NewTemplateNotifier(wh, "ALERT: {{.Subject}}", "severity={{.Severity}} body={{.Body}}")
	if err != nil {
		t.Fatalf("NewTemplateNotifier: %v", err)
	}

	msg := notify.Message{
		Subject:  "disk-cleanup",
		Body:     "missed run",
		Severity: "error",
	}
	if err := tn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Fatalf("expected 1 webhook call, got %d", len(received))
	}

	if got, want := received[0]["subject"], "ALERT: disk-cleanup"; got != want {
		t.Errorf("subject: got %q, want %q", got, want)
	}
	if got, want := received[0]["body"], "severity=error body=missed run"; got != want {
		t.Errorf("body: got %q, want %q", got, want)
	}
}

func TestIntegration_TemplateChainedWithTransform(t *testing.T) {
	var mu sync.Mutex
	var received []map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		_ = json.NewDecoder(r.Body).Decode(&payload)
		mu.Lock()
		received = append(received, payload)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL)
	// First upper-case the subject via TransformNotifier, then wrap with template.
	tr := notify.NewTransformNotifier(wh, notify.UpperCaseSubject)
	tn, err := notify.NewTemplateNotifier(tr, "[cronwatch] {{.Subject}}", "")
	if err != nil {
		t.Fatalf("NewTemplateNotifier: %v", err)
	}

	_ = tn.Notify(context.Background(), notify.Message{Subject: "heartbeat", Body: "ok"})

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Fatalf("expected 1 call, got %d", len(received))
	}
	// template runs first → "[cronwatch] heartbeat", then UpperCase → "[CRONWATCH] HEARTBEAT"
	if got := received[0]["subject"]; got == "" {
		t.Error("subject should not be empty")
	}
}
