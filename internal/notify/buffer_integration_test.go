package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/notify"
)

func TestIntegration_BufferFlushesToWebhook(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	buf := notify.NewBufferNotifier(wh, 10*time.Second, 3)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if err := buf.Notify(ctx, notify.Message{Subject: "job missed", Body: "details"}); err != nil {
			t.Fatalf("Notify[%d] error: %v", i, err)
		}
	}

	// Only 1 HTTP call should have been made (the batched flush)
	if got := calls.Load(); got != 1 {
		t.Errorf("expected 1 webhook call, got %d", got)
	}
}

func TestIntegration_BufferWindowFlushesToWebhook(t *testing.T) {
	var bodies []map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		bodies = append(bodies, payload)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL, 5*time.Second)
	buf := notify.NewBufferNotifier(wh, 60*time.Millisecond, 100)
	ctx := context.Background()

	_ = buf.Notify(ctx, notify.Message{Subject: "alpha", Body: "b1"})
	_ = buf.Notify(ctx, notify.Message{Subject: "beta", Body: "b2"})

	time.Sleep(200 * time.Millisecond)

	if len(bodies) != 1 {
		t.Fatalf("expected 1 batched request, got %d", len(bodies))
	}
}
