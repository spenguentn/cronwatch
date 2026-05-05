package notify

import (
	"context"
	"errors"
	"testing"
)

func TestTemplateNotifier_RendersSubjectAndBody(t *testing.T) {
	cap := &captureNotifier{}
	tn, err := NewTemplateNotifier(cap, "[{{.Severity}}] {{.Subject}}", "job: {{.Subject}}\n{{.Body}}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := Message{Subject: "backup", Body: "details", Severity: "warn"}
	if err := tn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("Notify error: %v", err)
	}

	got := cap.last
	if want := "[warn] backup"; got.Subject != want {
		t.Errorf("subject: got %q, want %q", got.Subject, want)
	}
	if want := "job: backup\ndetails"; got.Body != want {
		t.Errorf("body: got %q, want %q", got.Body, want)
	}
}

func TestTemplateNotifier_EmptyTemplatesPassThrough(t *testing.T) {
	cap := &captureNotifier{}
	tn, err := NewTemplateNotifier(cap, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := Message{Subject: "original", Body: "body"}
	_ = tn.Notify(context.Background(), msg)

	if cap.last.Subject != "original" {
		t.Errorf("subject should be unchanged, got %q", cap.last.Subject)
	}
	if cap.last.Body != "body" {
		t.Errorf("body should be unchanged, got %q", cap.last.Body)
	}
}

func TestTemplateNotifier_NilInnerReturnsError(t *testing.T) {
	_, err := NewTemplateNotifier(nil, "{{.Subject}}", "")
	if err == nil {
		t.Fatal("expected error for nil inner notifier")
	}
}

func TestTemplateNotifier_InvalidSubjectTemplate(t *testing.T) {
	cap := &captureNotifier{}
	_, err := NewTemplateNotifier(cap, "{{.Unclosed", "")
	if err == nil {
		t.Fatal("expected error for invalid subject template")
	}
}

func TestTemplateNotifier_InvalidBodyTemplate(t *testing.T) {
	cap := &captureNotifier{}
	_, err := NewTemplateNotifier(cap, "", "{{.Unclosed")
	if err == nil {
		t.Fatal("expected error for invalid body template")
	}
}

func TestTemplateNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	tn, _ := NewTemplateNotifier(&errorNotifier{err: sentinel}, "{{.Subject}}", "")

	err := tn.Notify(context.Background(), Message{Subject: "x"})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestTemplateNotifier_AccessMetaInTemplate(t *testing.T) {
	cap := &captureNotifier{}
	tn, err := NewTemplateNotifier(cap, "{{index .Meta \"job\"}}", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msg := Message{Subject: "ignored", Meta: map[string]string{"job": "nightly-backup"}}
	_ = tn.Notify(context.Background(), msg)

	if cap.last.Subject != "nightly-backup" {
		t.Errorf("subject: got %q, want %q", cap.last.Subject, "nightly-backup")
	}
}

// ---------------------------------------------------------------------------
// helpers shared across notify package tests
// ---------------------------------------------------------------------------

type captureNotifier struct{ last Message }

func (c *captureNotifier) Notify(_ context.Context, m Message) error {
	c.last = m
	return nil
}

type errorNotifier struct{ err error }

func (e *errorNotifier) Notify(_ context.Context, _ Message) error { return e.err }
