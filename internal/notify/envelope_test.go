package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

type captureNotifier struct {
	got Message
	err error
}

func (c *captureNotifier) Notify(_ context.Context, msg Message) error {
	c.got = msg
	return c.err
}

func fixedTime() time.Time {
	return time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
}

func TestEnvelopeNotifier_AddsMetadata(t *testing.T) {
	cap := &captureNotifier{}
	en := NewEnvelopeNotifier(cap,
		WithSource("test-source"),
		WithHostnameFunc(func() string { return "host-1" }),
	)
	en.now = fixedTime

	msg := Message{Subject: "job missed", Body: "backup missed run"}
	if err := en.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := cap.got.Meta["source"]; got != "test-source" {
		t.Errorf("source: want %q, got %q", "test-source", got)
	}
	if got := cap.got.Meta["host"]; got != "host-1" {
		t.Errorf("host: want %q, got %q", "host-1", got)
	}
	want := fixedTime().UTC().Format(time.RFC3339)
	if got := cap.got.Meta["sent_at"]; got != want {
		t.Errorf("sent_at: want %q, got %q", want, got)
	}
}

func TestEnvelopeNotifier_DoesNotOverwriteExistingMeta(t *testing.T) {
	cap := &captureNotifier{}
	en := NewEnvelopeNotifier(cap, WithSource("cronwatch"))
	en.now = fixedTime

	msg := Message{
		Subject: "s",
		Meta:    map[string]string{"source": "custom", "host": "myhost"},
	}
	_ = en.Notify(context.Background(), msg)

	if got := cap.got.Meta["source"]; got != "custom" {
		t.Errorf("source should not be overwritten: got %q", got)
	}
	if got := cap.got.Meta["host"]; got != "myhost" {
		t.Errorf("host should not be overwritten: got %q", got)
	}
}

func TestEnvelopeNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	cap := &captureNotifier{err: sentinel}
	en := NewEnvelopeNotifier(cap)
	en.now = fixedTime

	err := en.Notify(context.Background(), Message{Subject: "x"})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestEnvelopeNotifier_NilInner(t *testing.T) {
	en := NewEnvelopeNotifier(nil)
	en.now = fixedTime

	err := en.Notify(context.Background(), Message{Subject: "x"})
	if err == nil {
		t.Error("expected error for nil inner notifier")
	}
}

func TestEnvelopeNotifier_DefaultSource(t *testing.T) {
	cap := &captureNotifier{}
	en := NewEnvelopeNotifier(cap)
	en.now = fixedTime

	_ = en.Notify(context.Background(), Message{Subject: "x"})
	if got := cap.got.Meta["source"]; got != "cronwatch" {
		t.Errorf("default source: want %q, got %q", "cronwatch", got)
	}
}
