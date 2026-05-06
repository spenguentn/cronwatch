package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_TagEnrichesWebhookPayload(t *testing.T) {
	var received notify.Message

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	webhook := notify.NewWebhookNotifier(srv.URL, 0)
	tagger := notify.NewTagNotifier(webhook, map[string]string{
		"source": "cronwatch",
		"env":    "production",
	})

	msg := notify.Message{Subject: "job missed", Body: "backup-daily did not run"}
	if err := tagger.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Meta["source"] != "cronwatch" {
		t.Errorf("expected source=cronwatch, got %q", received.Meta["source"])
	}
	if received.Meta["env"] != "production" {
		t.Errorf("expected env=production, got %q", received.Meta["env"])
	}
	if received.Subject != "job missed" {
		t.Errorf("subject should be unchanged, got %q", received.Subject)
	}
}

func TestIntegration_TagChainedWithTransform(t *testing.T) {
	var received notify.Message

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	webhook := notify.NewWebhookNotifier(srv.URL, 0)
	transformed := notify.NewTransformNotifier(webhook, notify.PrefixSubject("[ALERT] "))
	tagged := notify.NewTagNotifier(transformed, map[string]string{"tier": "critical"})

	msg := notify.Message{Subject: "disk full", Body: "root partition at 99%"}
	if err := tagged.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Subject != "[ALERT] disk full" {
		t.Errorf("transform should apply after tag, got %q", received.Subject)
	}
	if received.Meta["tier"] != "critical" {
		t.Errorf("tag should be present, got %q", received.Meta["tier"])
	}
}
