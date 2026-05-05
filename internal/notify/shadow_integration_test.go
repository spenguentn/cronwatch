package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/notify"
)

func TestIntegration_ShadowDeliversToBothWebhooks(t *testing.T) {
	var primaryHits, shadowHits atomic.Int32

	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primaryHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer primaryServer.Close()

	shadowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shadowHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer shadowServer.Close()

	primary := notify.NewWebhookNotifier(primaryServer.URL, 5*time.Second)
	shadow := notify.NewWebhookNotifier(shadowServer.URL, 5*time.Second)
	sn := notify.NewShadowNotifier(primary, shadow, nil)

	msg := notify.Message{Subject: "cron drift", Body: "job late by 5m", Severity: notify.SeverityWarn}
	if err := sn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Allow shadow goroutine to complete.
	time.Sleep(50 * time.Millisecond)

	if primaryHits.Load() != 1 {
		t.Errorf("primary: want 1 hit, got %d", primaryHits.Load())
	}
	if shadowHits.Load() != 1 {
		t.Errorf("shadow: want 1 hit, got %d", shadowHits.Load())
	}
}

func TestIntegration_ShadowPrimaryPayloadIntact(t *testing.T) {
	received := make(chan string, 1)

	primaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer primaryServer.Close()

	primary := notify.NewWebhookNotifier(primaryServer.URL, 5*time.Second)
	sn := notify.NewShadowNotifier(primary, nil, nil)

	if err := sn.Notify(context.Background(), notify.Message{Subject: "check"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case ct := <-received:
		if ct != "application/json" {
			t.Errorf("want application/json, got %q", ct)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for primary request")
	}
}
