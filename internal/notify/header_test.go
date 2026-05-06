package notify

import (
	"context"
	"errors"
	"testing"
)

func TestHeaderNotifier_InjectsHeaders(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	n := NewHeaderNotifier(inner, map[string]string{"env": "prod", "region": "us-east"})
	msg := Message{Subject: "test", Body: "body"}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got.Meta["env"])
	}
	if got.Meta["region"] != "us-east" {
		t.Errorf("expected region=us-east, got %q", got.Meta["region"])
	}
}

func TestHeaderNotifier_DoesNotOverwriteExistingMeta(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	n := NewHeaderNotifier(inner, map[string]string{"env": "prod"})
	msg := Message{
		Subject: "test",
		Meta:    map[string]string{"env": "staging"},
	}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["env"] != "staging" {
		t.Errorf("existing meta should win: expected staging, got %q", got.Meta["env"])
	}
}

func TestHeaderNotifier_NilInnerIsNoop(t *testing.T) {
	n := NewHeaderNotifier(nil, map[string]string{"k": "v"})
	if err := n.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("expected nil error from nil inner, got %v", err)
	}
}

func TestHeaderNotifier_EmptyHeadersPassesThrough(t *testing.T) {
	called := false
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		called = true
		return nil
	})
	n := NewHeaderNotifier(inner, nil)
	if err := n.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected inner to be called")
	}
}

func TestHeaderNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return sentinel
	})
	n := NewHeaderNotifier(inner, map[string]string{"x": "1"})
	err := n.Notify(context.Background(), Message{Subject: "boom"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel in error chain, got %v", err)
	}
}
