package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/cronwatch/internal/notify"
)

func TestIntegration_ShardDistributesAcrossWebhooks(t *testing.T) {
	var counts [3]int64
	servers := make([]*httptest.Server, 3)
	notifiers := make([]notify.Notifier, 3)

	for i := range servers {
		i := i
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt64(&counts[i], 1)
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(servers[i].Close)

		var err error
		notifiers[i], err = notify.NewWebhookNotifier(servers[i].URL)
		if err != nil {
			t.Fatalf("webhook %d: %v", i, err)
		}
	}

	sn, err := notify.NewShardNotifier(notifiers...)
	if err != nil {
		t.Fatalf("NewShardNotifier: %v", err)
	}

	subjects := []string{"alpha", "beta", "gamma", "delta", "epsilon",
		"zeta", "eta", "theta", "iota", "kappa"}
	for _, subj := range subjects {
		if err := sn.Notify(context.Background(), notify.Message{Subject: subj, Body: "test"}); err != nil {
			t.Errorf("Notify(%q): %v", subj, err)
		}
	}

	total := atomic.LoadInt64(&counts[0]) + atomic.LoadInt64(&counts[1]) + atomic.LoadInt64(&counts[2])
	if total != int64(len(subjects)) {
		t.Errorf("expected %d total deliveries, got %d", len(subjects), total)
	}

	// Each shard should have received at least one message with 10 subjects.
	for i, c := range counts {
		if atomic.LoadInt64(&c) == 0 {
			t.Errorf("shard %d received no messages", i)
		}
	}
}

func TestIntegration_ShardConsistentRerouting(t *testing.T) {
	var hits [2]int64
	servers := make([]*httptest.Server, 2)
	notifiers := make([]notify.Notifier, 2)

	for i := range servers {
		i := i
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			atomic.AddInt64(&hits[i], 1)
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(servers[i].Close)
		notifiers[i], _ = notify.NewWebhookNotifier(servers[i].URL)
	}

	sn, _ := notify.NewShardNotifier(notifiers...)
	subject := "stable-job"

	for i := 0; i < 5; i++ {
		_ = sn.Notify(context.Background(), notify.Message{Subject: subject})
	}

	idx := sn.ShardIndex(subject)
	if atomic.LoadInt64(&hits[idx]) != 5 {
		t.Errorf("expected shard %d to receive all 5 messages, got %d", idx, hits[idx])
	}
}
