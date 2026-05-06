package notify

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
)

func TestTraceNotifier_RecordsSuccess(t *testing.T) {
	var buf bytes.Buffer
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tn := NewTraceNotifier(inner, &buf)

	msg := Message{Subject: "heartbeat", Body: "ok"}
	if err := tn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := tn.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Err != nil {
		t.Errorf("expected nil error, got %v", entries[0].Err)
	}
	if entries[0].Subject != "heartbeat" {
		t.Errorf("unexpected subject: %s", entries[0].Subject)
	}
	if !strings.Contains(buf.String(), "heartbeat") {
		t.Errorf("trace output missing subject: %s", buf.String())
	}
}

func TestTraceNotifier_RecordsError(t *testing.T) {
	var buf bytes.Buffer
	sentinel := errors.New("boom")
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return sentinel })
	tn := NewTraceNotifier(inner, &buf)

	_ = tn.Notify(context.Background(), Message{Subject: "job"})

	entries := tn.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !errors.Is(entries[0].Err, sentinel) {
		t.Errorf("expected sentinel error, got %v", entries[0].Err)
	}
	if !strings.Contains(buf.String(), "err=boom") {
		t.Errorf("trace output missing error: %s", buf.String())
	}
}

func TestTraceNotifier_Reset(t *testing.T) {
	var buf bytes.Buffer
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tn := NewTraceNotifier(inner, &buf)

	_ = tn.Notify(context.Background(), Message{Subject: "a"})
	_ = tn.Notify(context.Background(), Message{Subject: "b"})
	tn.Reset()

	if len(tn.Entries()) != 0 {
		t.Errorf("expected empty entries after reset")
	}
}

func TestTraceNotifier_NilInnerIsNoop(t *testing.T) {
	var buf bytes.Buffer
	tn := NewTraceNotifier(nil, &buf)
	if err := tn.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTraceNotifier_NilWriterDefaultsToStderr(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tn := NewTraceNotifier(inner, nil)
	if tn.out == nil {
		t.Error("expected non-nil writer")
	}
}

func TestTraceNotifier_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tn := NewTraceNotifier(inner, &buf)

	for i := 0; i < 5; i++ {
		_ = tn.Notify(context.Background(), Message{Subject: "ping"})
	}
	if len(tn.Entries()) != 5 {
		t.Errorf("expected 5 entries, got %d", len(tn.Entries()))
	}
}
