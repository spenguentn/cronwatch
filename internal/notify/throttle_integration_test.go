package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestIntegration_ThrottleWithWebhook(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhook := NewWebhookNotifier(server.URL, 5*time.Second)
	tn, err := NewThrottleNotifier(webhook, 500*time.Millisecond, 2)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 4; i++ {
		_ = tn.Notify(ctx, "backup-job", "drift detected")
	}

	if got := hits.Load(); got != 2 {
		t.Errorf("expected 2 webhook hits within window, got %d", got)
	}

	// Wait for the window to expire, then verify new notifications go through.
	time.Sleep(600 * time.Millisecond)
	_ = tn.Notify(ctx, "backup-job", "drift detected")

	if got := hits.Load(); got != 3 {
		t.Errorf("expected 3 total webhook hits after window reset, got %d", got)
	}
}

func TestIntegration_ThrottleMultipleJobs(t *testing.T) {
	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhook := NewWebhookNotifier(server.URL, 5*time.Second)
	tn, _ := NewThrottleNotifier(webhook, time.Minute, 1)

	ctx := context.Background()
	jobs := []string{"job-a", "job-b", "job-c"}
	for _, job := range jobs {
		_ = tn.Notify(ctx, job, "missed run")
		_ = tn.Notify(ctx, job, "missed run") // throttled
	}

	if got := hits.Load(); got != 3 {
		t.Errorf("expected 1 hit per job (3 total), got %d", got)
	}
}
