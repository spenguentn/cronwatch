package notify_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_FallbackDeliversToPrimary(t *testing.T) {
	var hits int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	primary := notify.NewWebhookNotifier(ts.URL)
	fallback := notify.NotifierFunc(func(_ context.Context, _ notify.Message) error {
		t.Fatal("fallback should not be called when primary succeeds")
		return nil
	})

	fn := notify.NewFallbackNotifier(primary, fallback)
	if err := fn.Notify(context.Background(), notify.Message{Subject: "ping", Body: "ok"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 webhook hit, got %d", hits)
	}
}

func TestIntegration_FallbackUsesSecondaryOnPrimaryFailure(t *testing.T) {
	var secondaryHits int32
	secondary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&secondaryHits, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer secondary.Close()

	primary := notify.NotifierFunc(func(_ context.Context, _ notify.Message) error {
		return errors.New("primary unavailable")
	})
	fallback := notify.NewWebhookNotifier(secondary.URL)

	fn := notify.NewFallbackNotifier(primary, fallback)
	if err := fn.Notify(context.Background(), notify.Message{Subject: "alert", Body: "drift"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&secondaryHits) != 1 {
		t.Fatalf("expected 1 secondary hit, got %d", secondaryHits)
	}
}
