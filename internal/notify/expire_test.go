package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestExpireNotifier_NotExpired_Forwards(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	n := NewExpireNotifier(inner)
	msg := WithExpiry(Message{Subject: "ok"}, time.Now().Add(time.Hour))

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != "ok" {
		t.Errorf("expected message forwarded, got %q", got.Subject)
	}
}

func TestExpireNotifier_Expired_Drops(t *testing.T) {
	called := false
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called = true
		return nil
	})

	n := NewExpireNotifier(inner)
	msg := WithExpiry(Message{Subject: "stale"}, time.Now().Add(-time.Minute))

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected message to be dropped, but inner was called")
	}
}

func TestExpireNotifier_NoExpiry_Forwards(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	n := NewExpireNotifier(inner)
	msg := Message{Subject: "no-expiry"}

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != "no-expiry" {
		t.Errorf("expected forwarded, got %q", got.Subject)
	}
}

func TestExpireNotifier_InvalidExpiry_ReturnsError(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	n := NewExpireNotifier(inner)

	msg := Message{
		Subject: "bad",
		Meta:    map[string]string{"expires_at": "not-a-time"},
	}

	if err := n.Notify(context.Background(), msg); err == nil {
		t.Error("expected error for invalid expires_at")
	}
}

func TestExpireNotifier_NilInner_IsNoop(t *testing.T) {
	n := NewExpireNotifier(nil)
	msg := WithExpiry(Message{Subject: "x"}, time.Now().Add(time.Hour))
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestExpireNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("inner error")
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return sentinel })
	n := NewExpireNotifier(inner)

	msg := WithExpiry(Message{Subject: "e"}, time.Now().Add(time.Hour))
	if err := n.Notify(context.Background(), msg); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestWithExpiry_PreservesExistingMeta(t *testing.T) {
	msg := Message{
		Subject: "m",
		Meta:    map[string]string{"job": "backup"},
	}
	out := WithExpiry(msg, time.Now().Add(time.Hour))
	if out.Meta["job"] != "backup" {
		t.Errorf("existing meta lost: %v", out.Meta)
	}
	if _, ok := out.Meta["expires_at"]; !ok {
		t.Error("expires_at not set")
	}
}
