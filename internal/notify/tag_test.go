package notify

import (
	"context"
	"errors"
	"testing"
)

func TestTagNotifier_AddsTags(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	n := NewTagNotifier(inner, map[string]string{"env": "prod", "team": "ops"})
	msg := Message{Subject: "hello", Body: "world"}

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["env"] != "prod" {
		t.Errorf("expected env=prod, got %q", got.Meta["env"])
	}
	if got.Meta["team"] != "ops" {
		t.Errorf("expected team=ops, got %q", got.Meta["team"])
	}
}

func TestTagNotifier_OverwritesExistingMeta(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	n := NewTagNotifier(inner, map[string]string{"env": "prod"})
	msg := Message{
		Subject: "s",
		Meta:    map[string]string{"env": "staging", "region": "us-east"},
	}

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["env"] != "prod" {
		t.Errorf("tag should overwrite existing meta: got %q", got.Meta["env"])
	}
	if got.Meta["region"] != "us-east" {
		t.Errorf("unrelated meta should be preserved: got %q", got.Meta["region"])
	}
}

func TestTagNotifier_NilInnerReturnsError(t *testing.T) {
	n := NewTagNotifier(nil, map[string]string{"k": "v"})
	err := n.Notify(context.Background(), Message{Subject: "x"})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestTagNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return sentinel
	})

	n := NewTagNotifier(inner, map[string]string{"k": "v"})
	err := n.Notify(context.Background(), Message{Subject: "x"})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestTagNotifier_EmptyTagsPassesThrough(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})

	n := NewTagNotifier(inner, nil)
	msg := Message{Subject: "s", Meta: map[string]string{"a": "b"}}

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Meta["a"] != "b" {
		t.Errorf("original meta should survive empty tag set")
	}
}

func TestTagNotifier_DoesNotMutateOriginalMeta(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	n := NewTagNotifier(inner, map[string]string{"injected": "yes"})

	orig := map[string]string{"a": "1"}
	msg := Message{Subject: "s", Meta: orig}

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := orig["injected"]; ok {
		t.Error("TagNotifier must not mutate the original Meta map")
	}
}
