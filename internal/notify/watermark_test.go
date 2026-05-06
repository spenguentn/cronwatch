package notify

import (
	"context"
	"errors"
	"testing"
)

func TestWatermarkNotifier_AboveMark(t *testing.T) {
	var got Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = m
		return nil
	})
	w := NewWatermarkNotifier(inner, SeverityWarn)
	msg := Message{Subject: "disk full", Severity: SeverityError}
	if err := w.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Subject != msg.Subject {
		t.Errorf("expected message forwarded, got %q", got.Subject)
	}
}

func TestWatermarkNotifier_BelowMark(t *testing.T) {
	called := false
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called = true
		return nil
	})
	w := NewWatermarkNotifier(inner, SeverityError)
	if err := w.Notify(context.Background(), Message{Severity: SeverityWarn}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("inner should not have been called for message below mark")
	}
}

func TestWatermarkNotifier_AtMark(t *testing.T) {
	called := false
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called = true
		return nil
	})
	w := NewWatermarkNotifier(inner, SeverityWarn)
	if err := w.Notify(context.Background(), Message{Severity: SeverityWarn}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("inner should have been called for message at mark")
	}
}

func TestWatermarkNotifier_SetMark(t *testing.T) {
	called := false
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called = true
		return nil
	})
	w := NewWatermarkNotifier(inner, SeverityError)
	// lower the mark so warn passes
	w.SetMark(SeverityWarn)
	if w.Mark() != SeverityWarn {
		t.Errorf("expected mark SeverityWarn, got %v", w.Mark())
	}
	if err := w.Notify(context.Background(), Message{Severity: SeverityWarn}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("inner should have been called after lowering mark")
	}
}

func TestWatermarkNotifier_NilInner(t *testing.T) {
	w := NewWatermarkNotifier(nil, SeverityWarn)
	if err := w.Notify(context.Background(), Message{Severity: SeverityError}); err != nil {
		t.Fatalf("nil inner should be a noop, got error: %v", err)
	}
}

func TestWatermarkNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return sentinel
	})
	w := NewWatermarkNotifier(inner, SeverityWarn)
	err := w.Notify(context.Background(), Message{Severity: SeverityError})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
