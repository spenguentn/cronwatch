package notify

import (
	"context"
	"errors"
	"testing"
)

func TestTeeNotifier_BothReceiveMessage(t *testing.T) {
	var gotA, gotB Message
	a := NotifierFunc(func(_ context.Context, m Message) error { gotA = m; return nil })
	b := NotifierFunc(func(_ context.Context, m Message) error { gotB = m; return nil })

	tee := NewTeeNotifier(a, b)
	msg := Message{Subject: "hello", Body: "world"}
	if err := tee.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotA.Subject != msg.Subject {
		t.Errorf("primary got %q, want %q", gotA.Subject, msg.Subject)
	}
	if gotB.Subject != msg.Subject {
		t.Errorf("secondary got %q, want %q", gotB.Subject, msg.Subject)
	}
}

func TestTeeNotifier_PrimaryErrorReturned(t *testing.T) {
	primaryErr := errors.New("primary down")
	a := NotifierFunc(func(_ context.Context, _ Message) error { return primaryErr })
	b := NotifierFunc(func(_ context.Context, _ Message) error { return nil })

	tee := NewTeeNotifier(a, b)
	err := tee.Notify(context.Background(), Message{Subject: "x"})
	if !errors.Is(err, primaryErr) {
		t.Errorf("expected primary error, got %v", err)
	}
}

func TestTeeNotifier_SecondaryErrorWrapped(t *testing.T) {
	secondaryErr := errors.New("secondary down")
	a := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	b := NotifierFunc(func(_ context.Context, _ Message) error { return secondaryErr })

	tee := NewTeeNotifier(a, b)
	err := tee.Notify(context.Background(), Message{Subject: "x"})
	if err == nil {
		t.Fatal("expected secondary error, got nil")
	}
	if !errors.Is(err, secondaryErr) {
		t.Errorf("expected wrapped secondary error, got %v", err)
	}
}

func TestTeeNotifier_NilPrimarySkipped(t *testing.T) {
	var called bool
	b := NotifierFunc(func(_ context.Context, _ Message) error { called = true; return nil })

	tee := NewTeeNotifier(nil, b)
	if err := tee.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("secondary should have been called")
	}
}

func TestTeeNotifier_NilSecondarySkipped(t *testing.T) {
	var called bool
	a := NotifierFunc(func(_ context.Context, _ Message) error { called = true; return nil })

	tee := NewTeeNotifier(a, nil)
	if err := tee.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("primary should have been called")
	}
}

func TestTeeNotifier_BothNilNoError(t *testing.T) {
	tee := NewTeeNotifier(nil, nil)
	if err := tee.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
