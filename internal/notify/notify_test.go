package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/cronwatch/internal/notify"
)

func newTestServer(t *testing.T, statusCode int, got *notify.Payload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got != nil {
			if err := json.NewDecoder(r.Body).Decode(got); err != nil {
				t.Errorf("decode body: %v", err)
			}
		}
		w.WriteHeader(statusCode)
	}))
}

func TestWebhookNotifier_Success(t *testing.T) {
	var got notify.Payload
	srv := newTestServer(t, http.StatusOK, &got)
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL)
	if err := n.Notify(context.Background(), "warn", "backup", "drift detected"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Level != "warn" {
		t.Errorf("level: want warn, got %s", got.Level)
	}
	if got.Job != "backup" {
		t.Errorf("job: want backup, got %s", got.Job)
	}
	if got.Message != "drift detected" {
		t.Errorf("message mismatch: %s", got.Message)
	}
	if got.TS == 0 {
		t.Error("expected non-zero timestamp")
	}
}

func TestWebhookNotifier_NonOKStatus(t *testing.T) {
	srv := newTestServer(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL)
	err := n.Notify(context.Background(), "error", "sync", "missed run")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhookNotifier_BadURL(t *testing.T) {
	n := notify.NewWebhookNotifier("http://127.0.0.1:0/nonexistent")
	err := n.Notify(context.Background(), "warn", "job", "msg")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}

func TestWebhookNotifier_CancelledContext(t *testing.T) {
	srv := newTestServer(t, http.StatusOK, nil)
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	n := notify.NewWebhookNotifier(srv.URL)
	err := n.Notify(ctx, "warn", "job", "msg")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
