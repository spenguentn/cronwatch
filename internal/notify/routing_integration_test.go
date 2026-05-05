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

func TestIntegration_RoutingToWebhooks(t *testing.T) {
	var criticalHits, warnHits int32

	criticalSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&criticalHits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer criticalSrv.Close()

	warnSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&warnHits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer warnSrv.Close()

	criticalN := notify.NewWebhookNotifier(criticalSrv.URL, 0)
	warnN := notify.NewWebhookNotifier(warnSrv.URL, 0)

	keyFn := func(m notify.Message) string {
		if m.Meta != nil {
			return m.Meta["tier"]
		}
		return ""
	}

	rn := notify.NewRoutingNotifier(
		keyFn,
		warnN,
		notify.Route{Name: "critical", Notifier: criticalN},
		notify.Route{Name: "warn", Notifier: warnN},
	)

	ctx := context.Background()

	_ = rn.Notify(ctx, notify.Message{Subject: "db missed", Meta: map[string]string{"tier": "critical"}})
	_ = rn.Notify(ctx, notify.Message{Subject: "log drift", Meta: map[string]string{"tier": "warn"}})
	// unmatched — goes to fallback (warnSrv)
	_ = rn.Notify(ctx, notify.Message{Subject: "unknown", Meta: map[string]string{"tier": "other"}})

	if atomic.LoadInt32(&criticalHits) != 1 {
		t.Errorf("critical webhook: expected 1 hit, got %d", criticalHits)
	}
	if atomic.LoadInt32(&warnHits) != 2 {
		t.Errorf("warn webhook: expected 2 hits, got %d", warnHits)
	}
}

func TestIntegration_RoutingPreservesPayload(t *testing.T) {
	var received notify.Message

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL, 0)
	rn := notify.NewRoutingNotifier(
		func(m notify.Message) string { return "target" },
		nil,
		notify.Route{Name: "target", Notifier: n},
	)

	msg := notify.Message{
		Subject:  "payload check",
		Body:     "details here",
		Severity: notify.SeverityError,
		Meta:     map[string]string{"job": "archiver"},
	}
	if err := rn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Subject != msg.Subject {
		t.Errorf("subject mismatch: got %q", received.Subject)
	}
}
