package notify

import (
	"context"
	"errors"
	"testing"
)

func TestLabelNotifier_AddsLabels(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})
	n := NewLabelNotifier(inner, map[string]string{"env": "prod", "region": "us-east-1"})
	msg := Message{Subject: "test"}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got.Meta["env"])
	}
	if got.Meta["region"] != "us-east-1" {
		t.Errorf("expected region=us-east-1, got %q", got.Meta["region"])
	}
}

func TestLabelNotifier_DoesNotOverwriteExistingMeta(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})
	n := NewLabelNotifier(inner, map[string]string{"env": "prod"})
	msg := Message{
		Subject: "test",
		Meta:    map[string]string{"env": "staging"},
	}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["env"] != "staging" {
		t.Errorf("expected existing value staging to be preserved, got %q", got.Meta["env"])
	}
}

func TestLabelNotifier_NilInnerReturnsError(t *testing.T) {
	n := NewLabelNotifier(nil, map[string]string{"k": "v"})
	err := n.Notify(context.Background(), Message{Subject: "x"})
	if err == nil {
		t.Fatal("expected error for nil inner, got nil")
	}
}

func TestLabelNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("inner error")
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return sentinel })
	n := NewLabelNotifier(inner, map[string]string{"k": "v"})
	if err := n.Notify(context.Background(), Message{Subject: "x"}); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestLabelNotifier_EmptyLabelsPassesThrough(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})
	n := NewLabelNotifier(inner, map[string]string{})
	msg := Message{Subject: "hello", Body: "world"}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != "hello" {
		t.Errorf("expected subject hello, got %q", got.Subject)
	}
}

// TestLabelNotifier_DoesNotMutateOriginalMeta verifies that the notifier does
// not modify the Meta map of the original message passed by the caller.
func TestLabelNotifier_DoesNotMutateOriginalMeta(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	n := NewLabelNotifier(inner, map[string]string{"env": "prod"})
	origMeta := map[string]string{"service": "worker"}
	msg := Message{Subject: "test", Meta: origMeta}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := origMeta["env"]; ok {
		t.Error("original Meta map was mutated: 'env' key should not have been added")
	}
}
