package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestIntegration_FanoutDelivery(t *testing.T) {
	var hits atomic.Int64

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	})

	server1 := httptest.NewServer(handler)
	defer server1.Close()
	server2 := httptest.NewServer(handler)
	defer server2.Close()
	server3 := httptest.NewServer(handler)
	defer server3.Close()

	wh1, _ := NewWebhookNotifier(server1.URL, 2*time.Second)
	wh2, _ := NewWebhookNotifier(server2.URL, 2*time.Second)
	wh3, _ := NewWebhookNotifier(server3.URL, 2*time.Second)

	f, err := NewFanoutNotifier(wh1, wh2, wh3)
	if err != nil {
		t.Fatalf("NewFanoutNotifier: %v", err)
	}

	msg := Message{Subject: "cron drift", Body: "job backup exceeded threshold"}
	if err := f.Notify(context.Background(), msg); err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if got := hits.Load(); got != 3 {
		t.Errorf("expected 3 webhook hits, got %d", got)
	}
}

func TestIntegration_FanoutWithRetry(t *testing.T) {
	var attempts atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	wh, _ := NewWebhookNotifier(server.URL, 2*time.Second)
	retrying := NewRetryNotifier(wh, 3, 10*time.Millisecond)

	f, err := NewFanoutNotifier(retrying)
	if err != nil {
		t.Fatalf("NewFanoutNotifier: %v", err)
	}

	if err := f.Notify(context.Background(), Message{Subject: "s", Body: "b"}); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}

	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}
