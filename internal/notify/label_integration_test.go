package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/cronwatch/internal/notify"
)

func TestIntegration_LabelEnrichesWebhookPayload(t *testing.T) {
	var received notify.Message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 0)
	n := notify.NewLabelNotifier(wh, map[string]string{
		"service": "cronwatch",
		"tier":    "backend",
	})

	msg := notify.Message{Subject: "job.missed", Body: "backup did not run"}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Meta["service"] != "cronwatch" {
		t.Errorf("expected service=cronwatch in payload, got %q", received.Meta["service"])
	}
	if received.Meta["tier"] != "backend" {
		t.Errorf("expected tier=backend in payload, got %q", received.Meta["tier"])
	}
}

func TestIntegration_LabelPreservesUserMeta(t *testing.T) {
	var received notify.Message
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, 0)
	n := notify.NewLabelNotifier(wh, map[string]string{"env": "prod"})

	msg := notify.Message{
		Subject: "job.drift",
		Meta:    map[string]string{"env": "staging", "job": "cleanup"},
	}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Meta["env"] != "staging" {
		t.Errorf("user-supplied env should not be overwritten, got %q", received.Meta["env"])
	}
	if received.Meta["job"] != "cleanup" {
		t.Errorf("expected job=cleanup, got %q", received.Meta["job"])
	}
}
