package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSeenNotifier_FirstCallPasses(t *testing.T) {
	var got []Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		got = append(got, m)
		return nil
	})
	sn := NewSeenNotifier(inner, time.Minute)
	msg := Message{Subject: "job.missed", Body: "backup missed"}

	if err := sn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(got))
	}
}

func TestSeenNotifier_DuplicateSuppressed(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})
	sn := NewSeenNotifier(inner, time.Minute)
	msg := Message{Subject: "job.missed", Body: "backup missed"}

	_ = sn.Notify(context.Background(), msg)
	_ = sn.Notify(context.Background(), msg)

	if count != 1 {
		t.Fatalf("expected 1 delivery, got %d", count)
	}
}

func TestSeenNotifier_DifferentBodyPasses(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})
	sn := NewSeenNotifier(inner, time.Minute)

	_ = sn.Notify(context.Background(), Message{Subject: "job", Body: "run 1"})
	_ = sn.Notify(context.Background(), Message{Subject: "job", Body: "run 2"})

	if count != 2 {
		t.Fatalf("expected 2 deliveries, got %d", count)
	}
}

func TestSeenNotifier_WindowExpiry(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})
	sn := NewSeenNotifier(inner, 100*time.Millisecond)
	now := time.Now()
	sn.nowFunc = func() time.Time { return now }

	msg := Message{Subject: "s", Body: "b"}
	_ = sn.Notify(context.Background(), msg)

	// advance past window
	sn.nowFunc = func() time.Time { return now.Add(200 * time.Millisecond) }
	_ = sn.Notify(context.Background(), msg)

	if count != 2 {
		t.Fatalf("expected 2 deliveries after expiry, got %d", count)
	}
}

func TestSeenNotifier_Reset(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})
	sn := NewSeenNotifier(inner, time.Minute)
	msg := Message{Subject: "x", Body: "y"}

	_ = sn.Notify(context.Background(), msg)
	sn.Reset()
	_ = sn.Notify(context.Background(), msg)

	if count != 2 {
		t.Fatalf("expected 2 deliveries after reset, got %d", count)
	}
}

func TestSeenNotifier_NilInnerIsNoop(t *testing.T) {
	sn := NewSeenNotifier(nil, time.Minute)
	if err := sn.Notify(context.Background(), Message{Subject: "x", Body: "y"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSeenNotifier_PropagatesError(t *testing.T) {
	want := errors.New("send failed")
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return want })
	sn := NewSeenNotifier(inner, time.Minute)

	if err := sn.Notify(context.Background(), Message{Subject: "a", Body: "b"}); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
