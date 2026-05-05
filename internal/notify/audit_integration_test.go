package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_AuditWithWebhook(t *testing.T) {
	var received int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received++
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, nil)
	var buf strings.Builder
	an := notify.NewAuditNotifier(wh, &buf)

	msgs := []notify.Message{
		{Subject: "job.etl", Body: "drift detected", Severity: "warn"},
		{Subject: "job.backup", Body: "missed run", Severity: "error"},
	}
	for _, m := range msgs {
		if err := an.Notify(context.Background(), m); err != nil {
			t.Fatalf("Notify: %v", err)
		}
	}

	if received != 2 {
		t.Errorf("expected 2 webhook calls, got %d", received)
	}

	entries := an.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 audit entries, got %d", len(entries))
	}
	for _, e := range entries {
		if !e.Success {
			t.Errorf("entry %q should be successful", e.Subject)
		}
	}

	log := buf.String()
	if !strings.Contains(log, "job.etl") || !strings.Contains(log, "job.backup") {
		t.Errorf("log missing expected subjects: %s", log)
	}
}

func TestIntegration_AuditRecordsWebhookFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL, nil)
	an := notify.NewAuditNotifier(wh, nil)

	err := an.Notify(context.Background(), notify.Message{Subject: "job.fail", Severity: "error"})
	if err == nil {
		t.Fatal("expected error from non-OK status")
	}

	entries := an.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(entries))
	}
	if entries[0].Success {
		t.Error("expected entry to record failure")
	}
}
