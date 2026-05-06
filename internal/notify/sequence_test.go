package notify

import (
	"context"
	"errors"
	"testing"
)

func TestSequenceNotifier_AllSucceed(t *testing.T) {
	var calls []string
	make := func(name string) NotifierFunc {
		return func(_ context.Context, _ Message) error {
			calls = append(calls, name)
			return nil
		}
	}
	s := NewSequenceNotifier([]Notifier{make("a"), make("b"), make("c")})
	if err := s.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(calls))
	}
}

func TestSequenceNotifier_StopsOnFirstSuccess(t *testing.T) {
	var calls int
	make := func(fail bool) NotifierFunc {
		return func(_ context.Context, _ Message) error {
			calls++
			if fail {
				return errors.New("fail")
			}
			return nil
		}
	}
	s := NewSequenceNotifier(
		[]Notifier{make(true), make(false), make(false)},
		StopOnFirstSuccess(),
	)
	if err := s.Notify(context.Background(), Message{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// first fails, second succeeds → should stop, third never called
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestSequenceNotifier_ReturnsLastError(t *testing.T) {
	errA := errors.New("err-a")
	errB := errors.New("err-b")
	steps := []Notifier{
		NotifierFunc(func(_ context.Context, _ Message) error { return errA }),
		NotifierFunc(func(_ context.Context, _ Message) error { return errB }),
	}
	s := NewSequenceNotifier(steps)
	err := s.Notify(context.Background(), Message{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errB) {
		t.Fatalf("expected last error to wrap errB, got: %v", err)
	}
}

func TestSequenceNotifier_NilStepsSkipped(t *testing.T) {
	var called bool
	steps := []Notifier{
		nil,
		NotifierFunc(func(_ context.Context, _ Message) error {
			called = true
			return nil
		}),
		nil,
	}
	s := NewSequenceNotifier(steps)
	if err := s.Notify(context.Background(), Message{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected non-nil step to be called")
	}
}

func TestSequenceNotifier_EmptySteps(t *testing.T) {
	s := NewSequenceNotifier(nil)
	if err := s.Notify(context.Background(), Message{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSequenceNotifier_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := NewSequenceNotifier([]Notifier{
		NotifierFunc(func(_ context.Context, _ Message) error { return nil }),
	})
	if err := s.Notify(ctx, Message{}); err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
