package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSlackNotifier_Success(t *testing.T) {
	var received slackPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL, 5*time.Second)
	if err := n.Notify(context.Background(), "hello slack"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Text != "hello slack" {
		t.Errorf("expected 'hello slack', got %q", received.Text)
	}
}

func TestSlackNotifier_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewSlackNotifier(ts.URL, 5*time.Second)
	err := n.Notify(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestSlackNotifier_BadURL(t *testing.T) {
	n := NewSlackNotifier("http://127.0.0.1:0/no-server", time.Second)
	err := n.Notify(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
}

func TestSlackNotifier_CancelledContext(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	n := NewSlackNotifier(ts.URL, 5*time.Second)
	err := n.Notify(ctx, "test")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestSlackNotifier_DefaultTimeout(t *testing.T) {
	n := NewSlackNotifier("http://example.com", 0)
	if n.client.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", n.client.Timeout)
	}
}
