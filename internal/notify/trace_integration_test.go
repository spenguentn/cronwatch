package notify_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cronwatch/internal/notify"
)

func TestIntegration_TraceWithWebhook(t *testing.T) {
	var received int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received++
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	var buf bytes.Buffer
	wh := notify.NewWebhookNotifier(ts.URL)
	tn := notify.NewTraceNotifier(wh, &buf)

	ctx := context.Background()
	if err := tn.Notify(ctx, notify.Message{Subject: "deploy", Body: "done"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received != 1 {
		t.Errorf("expected 1 webhook call, got %d", received)
	}
	entries := tn.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 trace entry, got %d", len(entries))
	}
	if entries[0].Err != nil {
		t.Errorf("unexpected trace error: %v", entries[0].Err)
	}
	if !strings.Contains(buf.String(), "deploy") {
		t.Errorf("trace output missing subject")
	}
}

func TestIntegration_TraceRecordsWebhookFailure(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	var buf bytes.Buffer
	wh := notify.NewWebhookNotifier(ts.URL)
	tn := notify.NewTraceNotifier(wh, &buf)

	err := tn.Notify(context.Background(), notify.Message{Subject: "check"})
	if err == nil {
		t.Fatal("expected error from 500 response")
	}
	entries := tn.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 trace entry, got %d", len(entries))
	}
	if entries[0].Err == nil {
		t.Error("expected non-nil error in trace entry")
	}
	if !strings.Contains(buf.String(), "err=") {
		t.Errorf("trace output missing error indicator: %s", buf.String())
	}
}
