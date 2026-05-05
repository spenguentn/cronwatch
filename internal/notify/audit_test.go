package notify

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

type stubNotifier struct {
	err error
}

func (s *stubNotifier) Notify(_ context.Context, _ Message) error { return s.err }

func TestAuditNotifier_RecordsSuccess(t *testing.T) {
	var buf bytes.Buffer
	an := NewAuditNotifier(&stubNotifier{}, &buf)

	msg := Message{Subject: "job.backup", Body: "ok", Severity: "info"}
	if err := an.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := an.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if !e.Success {
		t.Error("expected Success=true")
	}
	if e.Subject != "job.backup" {
		t.Errorf("unexpected subject: %s", e.Subject)
	}
	if !strings.Contains(buf.String(), "OK") {
		t.Errorf("log line missing OK status: %s", buf.String())
	}
}

func TestAuditNotifier_RecordsFailure(t *testing.T) {
	sentinel := errors.New("delivery failed")
	an := NewAuditNotifier(&stubNotifier{err: sentinel}, nil)

	_ = an.Notify(context.Background(), Message{Subject: "job.sync", Severity: "error"})

	entries := an.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Success {
		t.Error("expected Success=false")
	}
	if !errors.Is(entries[0].Err, sentinel) {
		t.Errorf("unexpected error: %v", entries[0].Err)
	}
}

func TestAuditNotifier_Reset(t *testing.T) {
	an := NewAuditNotifier(&stubNotifier{}, nil)
	_ = an.Notify(context.Background(), Message{Subject: "a"})
	_ = an.Notify(context.Background(), Message{Subject: "b"})

	if len(an.Entries()) != 2 {
		t.Fatal("expected 2 entries before reset")
	}
	an.Reset()
	if len(an.Entries()) != 0 {
		t.Fatal("expected 0 entries after reset")
	}
}

func TestAuditNotifier_NilWriter(t *testing.T) {
	an := NewAuditNotifier(&stubNotifier{}, nil)
	// must not panic with nil writer
	if err := an.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditNotifier_MultipleEntries(t *testing.T) {
	an := NewAuditNotifier(&stubNotifier{}, nil)
	for i := 0; i < 5; i++ {
		_ = an.Notify(context.Background(), Message{Subject: "job"})
	}
	if got := len(an.Entries()); got != 5 {
		t.Errorf("expected 5 entries, got %d", got)
	}
}
