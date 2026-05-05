package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/yourorg/cronwatch/internal/notify"
)

func TestIntegration_PriorityRouterWithWebhook(t *testing.T) {
	var mu sync.Mutex
	received := map[string]int{}

	newWebhookFor := func(label string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			received[label]++
			mu.Unlock()
			w.WriteHeader(http.StatusOK)
		}))
	}

	lowSrv := newWebhookFor("low")
	defer lowSrv.Close()
	normalSrv := newWebhookFor("normal")
	defer normalSrv.Close()
	highSrv := newWebhookFor("high")
	defer highSrv.Close()

	router := notify.NewPriorityRouter(
		notify.NewWebhookNotifier(lowSrv.URL),
		notify.NewWebhookNotifier(normalSrv.URL),
		notify.NewWebhookNotifier(highSrv.URL),
	)

	ctx := context.Background()
	_ = router.Notify(ctx, notify.WithPriority(notify.Message{Subject: "drift"}, notify.PriorityLow))
	_ = router.Notify(ctx, notify.WithPriority(notify.Message{Subject: "missed"}, notify.PriorityHigh))
	_ = router.Notify(ctx, notify.Message{Subject: "info"}) // no priority → normal

	mu.Lock()
	defer mu.Unlock()
	if received["low"] != 1 {
		t.Errorf("low webhook called %d times, want 1", received["low"])
	}
	if received["high"] != 1 {
		t.Errorf("high webhook called %d times, want 1", received["high"])
	}
	if received["normal"] != 1 {
		t.Errorf("normal webhook called %d times, want 1", received["normal"])
	}
}

func TestIntegration_PriorityWithEnvelope(t *testing.T) {
	var body map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := notify.NewWebhookNotifier(srv.URL)
	env := notify.NewEnvelopeNotifier(wh, notify.WithSource("cronwatch"))
	router := notify.NewPriorityRouter(nil, env, nil)

	msg := notify.WithPriority(notify.Message{Subject: "check", Body: "ok"}, notify.PriorityNormal)
	if err := router.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
