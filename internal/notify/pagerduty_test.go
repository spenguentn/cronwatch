package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewPagerDutyNotifier_EmptyKey(t *testing.T) {
	_, err := NewPagerDutyNotifier("")
	if err == nil {
		t.Fatal("expected error for empty integration key")
	}
}

func TestPagerDutyNotifier_Success(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n, err := NewPagerDutyNotifier("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.endpoint = ts.URL

	if err := n.Notify(context.Background(), "job missed"); err != nil {
		t.Fatalf("Notify returned error: %v", err)
	}

	if received.RoutingKey != "test-key" {
		t.Errorf("routing key: got %q, want %q", received.RoutingKey, "test-key")
	}
	if received.Payload.Summary != "job missed" {
		t.Errorf("summary: got %q, want %q", received.Payload.Summary, "job missed")
	}
	if received.Payload.Source != "cronwatch" {
		t.Errorf("source: got %q, want %q", received.Payload.Source, "cronwatch")
	}
	if received.EventAction != "trigger" {
		t.Errorf("event_action: got %q, want %q", received.EventAction, "trigger")
	}
}

func TestPagerDutyNotifier_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	n, _ := NewPagerDutyNotifier("bad-key")
	n.endpoint = ts.URL

	if err := n.Notify(context.Background(), "test"); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestPagerDutyNotifier_BadURL(t *testing.T) {
	n, _ := NewPagerDutyNotifier("key")
	n.endpoint = "://invalid"

	if err := n.Notify(context.Background(), "test"); err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestPagerDutyNotifier_CancelledContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n, _ := NewPagerDutyNotifier("key")
	n.endpoint = ts.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := n.Notify(ctx, "test"); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
