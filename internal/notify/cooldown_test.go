package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCooldownNotifier_FirstCallPasses(t *testing.T) {
	var got []Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = append(got, m)
		return nil
	})

	n := NewCooldownNotifier(inner, 5*time.Minute)
	msg := Message{Subject: "job.a", Body: "fired"}

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(got))
	}
}

func TestCooldownNotifier_SecondCallSuppressed(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})

	n := NewCooldownNotifier(inner, 5*time.Minute)
	msg := Message{Subject: "job.a", Body: "fired"}

	_ = n.Notify(context.Background(), msg)
	_ = n.Notify(context.Background(), msg)

	if count != 1 {
		t.Fatalf("expected 1 delivery, got %d", count)
	}
}

func TestCooldownNotifier_AfterCooldown(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})

	now := time.Now()
	n := NewCooldownNotifier(inner, time.Minute)
	n.now = func() time.Time { return now }

	msg := Message{Subject: "job.a"}
	_ = n.Notify(context.Background(), msg)

	// advance past cooldown
	n.now = func() time.Time { return now.Add(2 * time.Minute) }
	_ = n.Notify(context.Background(), msg)

	if count != 2 {
		t.Fatalf("expected 2 deliveries, got %d", count)
	}
}

func TestCooldownNotifier_IndependentSubjects(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})

	n := NewCooldownNotifier(inner, 5*time.Minute)
	_ = n.Notify(context.Background(), Message{Subject: "job.a"})
	_ = n.Notify(context.Background(), Message{Subject: "job.b"})

	if count != 2 {
		t.Fatalf("expected 2 deliveries for independent subjects, got %d", count)
	}
}

func TestCooldownNotifier_Reset(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})

	n := NewCooldownNotifier(inner, 5*time.Minute)
	msg := Message{Subject: "job.a"}

	_ = n.Notify(context.Background(), msg)
	n.Reset("job.a")
	_ = n.Notify(context.Background(), msg)

	if count != 2 {
		t.Fatalf("expected 2 deliveries after reset, got %d", count)
	}
}

func TestCooldownNotifier_NilInnerIsNoop(t *testing.T) {
	n := NewCooldownNotifier(nil, time.Minute)
	if err := n.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCooldownNotifier_PropagatesError(t *testing.T) {
	expected := errors.New("downstream failure")
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return expected
	})

	n := NewCooldownNotifier(inner, time.Minute)
	err := n.Notify(context.Background(), Message{Subject: "job.a"})
	if !errors.Is(err, expected) {
		t.Fatalf("expected wrapped error, got %v", err)
	}
}
