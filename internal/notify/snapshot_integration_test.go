package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_SnapshotWithWebhook(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 0)
	sn := notify.NewSnapshotNotifier(wh, 20)

	for i := 0; i < 3; i++ {
		if err := sn.Notify(context.Background(), notify.Message{
			Subject:  "cron.heartbeat",
			Body:     "tick",
			Severity: "info",
		}); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}

	if calls != 3 {
		t.Errorf("expected 3 webhook calls, got %d", calls)
	}

	entries := sn.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 snapshot entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Err != nil {
			t.Errorf("expected no error in entry, got %v", e.Err)
		}
		if e.Subject != "cron.heartbeat" {
			t.Errorf("unexpected subject: %s", e.Subject)
		}
	}
}

func TestIntegration_SnapshotRecordsWebhookFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 0)
	sn := notify.NewSnapshotNotifier(wh, 5)

	_ = sn.Notify(context.Background(), notify.Message{Subject: "cron.fail", Severity: "error"})

	entries := sn.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Err == nil {
		t.Error("expected error recorded for non-OK webhook response")
	}
}
