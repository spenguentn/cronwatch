package notify

import (
	"context"
	"errors"
	"testing"
)

func jobKeyFn(msg Message) string {
	if msg.Meta == nil {
		return ""
	}
	return msg.Meta["job"]
}

func TestRoutingNotifier_MatchedRoute(t *testing.T) {
	var got Message
	n := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	rn := NewRoutingNotifier(jobKeyFn, nil, Route{Name: "backup", Notifier: n})
	msg := Message{Subject: "drift", Meta: map[string]string{"job": "backup"}}
	if err := rn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != "drift" {
		t.Errorf("expected subject 'drift', got %q", got.Subject)
	}
}

func TestRoutingNotifier_FallbackOnMiss(t *testing.T) {
	var called bool
	fb := NotifierFunc(func(_ context.Context, _ Message) error {
		called = true
		return nil
	})

	rn := NewRoutingNotifier(jobKeyFn, fb)
	msg := Message{Subject: "miss", Meta: map[string]string{"job": "unknown"}}
	if err := rn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected fallback to be called")
	}
}

func TestRoutingNotifier_SilentDropWithNoFallback(t *testing.T) {
	rn := NewRoutingNotifier(jobKeyFn, nil)
	msg := Message{Subject: "drop", Meta: map[string]string{"job": "ghost"}}
	if err := rn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("expected silent drop, got error: %v", err)
	}
}

func TestRoutingNotifier_NilKeyFnReturnsError(t *testing.T) {
	rn := &RoutingNotifier{routes: map[string]Notifier{}}
	err := rn.Notify(context.Background(), Message{Subject: "x"})
	if err == nil {
		t.Fatal("expected error for nil keyFn")
	}
}

func TestRoutingNotifier_AddAndRemoveRoute(t *testing.T) {
	var calls int
	n := NotifierFunc(func(_ context.Context, _ Message) error {
		calls++
		return nil
	})

	rn := NewRoutingNotifier(jobKeyFn, nil)
	rn.AddRoute("dynamic", n)

	msg := Message{Meta: map[string]string{"job": "dynamic"}}
	_ = rn.Notify(context.Background(), msg)
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	rn.RemoveRoute("dynamic")
	_ = rn.Notify(context.Background(), msg)
	if calls != 1 {
		t.Error("expected no additional call after removal")
	}
}

func TestRoutingNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("route error")
	n := NotifierFunc(func(_ context.Context, _ Message) error { return sentinel })

	rn := NewRoutingNotifier(jobKeyFn, nil, Route{Name: "job1", Notifier: n})
	msg := Message{Meta: map[string]string{"job": "job1"}}
	if err := rn.Notify(context.Background(), msg); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
