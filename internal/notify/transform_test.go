package notify

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type captureNotifier struct {
	got Message
	err error
}

func (c *captureNotifier) Notify(_ context.Context, msg Message) error {
	c.got = msg
	return c.err
}

func TestTransformNotifier_AppliesTransform(t *testing.T) {
	cap := &captureNotifier{}
	n := NewTransformNotifier(cap, PrefixSubject("[CRON] "))

	msg := Message{Subject: "job missed", Body: "details"}
	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.got.Subject != "[CRON] job missed" {
		t.Errorf("expected prefixed subject, got %q", cap.got.Subject)
	}
}

func TestTransformNotifier_NilTransformPassesThrough(t *testing.T) {
	cap := &captureNotifier{}
	n := NewTransformNotifier(cap, nil)

	msg := Message{Subject: "original"}
	_ = n.Notify(context.Background(), msg)
	if cap.got.Subject != "original" {
		t.Errorf("expected unchanged subject, got %q", cap.got.Subject)
	}
}

func TestTransformNotifier_NilInnerIsNoop(t *testing.T) {
	n := NewTransformNotifier(nil, PrefixSubject("X"))
	if err := n.Notify(context.Background(), Message{Subject: "s"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransformNotifier_PropagatesError(t *testing.T) {
	want := errors.New("send failed")
	cap := &captureNotifier{err: want}
	n := NewTransformNotifier(cap, nil)
	got := n.Notify(context.Background(), Message{})
	if !errors.Is(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestUpperCaseSubject(t *testing.T) {
	cap := &captureNotifier{}
	n := NewTransformNotifier(cap, UpperCaseSubject())
	_ = n.Notify(context.Background(), Message{Subject: "hello world"})
	if cap.got.Subject != strings.ToUpper("hello world") {
		t.Errorf("expected upper-cased subject, got %q", cap.got.Subject)
	}
}

func TestAddMeta_NewKeys(t *testing.T) {
	cap := &captureNotifier{}
	n := NewTransformNotifier(cap, AddMeta(map[string]string{"env": "prod", "team": "ops"}))
	_ = n.Notify(context.Background(), Message{Subject: "s"})
	if cap.got.Meta["env"] != "prod" || cap.got.Meta["team"] != "ops" {
		t.Errorf("expected meta to contain env and team, got %v", cap.got.Meta)
	}
}

func TestAddMeta_DoesNotOverwriteExisting(t *testing.T) {
	cap := &captureNotifier{}
	n := NewTransformNotifier(cap, AddMeta(map[string]string{"env": "staging"}))
	msg := Message{Subject: "s", Meta: map[string]string{"env": "prod"}}
	_ = n.Notify(context.Background(), msg)
	if cap.got.Meta["env"] != "prod" {
		t.Errorf("existing meta key should not be overwritten, got %q", cap.got.Meta["env"])
	}
}
