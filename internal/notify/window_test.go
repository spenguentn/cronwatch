package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

func makeWindow(start, end string) TimeWindow {
	parse := func(s string) time.Duration {
		t, _ := time.Parse("15:04", s)
		return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute
	}
	return TimeWindow{Start: parse(start), End: parse(end)}
}

func fixedNow(hour, minute int) func() time.Time {
	return func() time.Time {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	}
}

func TestWindowNotifier_InsideWindow(t *testing.T) {
	var received bool
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		received = true
		return nil
	})
	w := NewWindowNotifier(inner, []TimeWindow{makeWindow("09:00", "17:00")})
	w.now = fixedNow(10, 30)

	if err := w.Notify(context.Background(), Message{Subject: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !received {
		t.Error("expected message to be forwarded inside window")
	}
}

func TestWindowNotifier_OutsideWindow(t *testing.T) {
	var received bool
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		received = true
		return nil
	})
	w := NewWindowNotifier(inner, []TimeWindow{makeWindow("09:00", "17:00")})
	w.now = fixedNow(20, 0)

	if err := w.Notify(context.Background(), Message{Subject: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received {
		t.Error("expected message to be dropped outside window")
	}
}

func TestWindowNotifier_NoWindows_PassesAll(t *testing.T) {
	var received bool
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		received = true
		return nil
	})
	w := NewWindowNotifier(inner, nil)
	w.now = fixedNow(3, 0)

	if err := w.Notify(context.Background(), Message{Subject: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !received {
		t.Error("expected message to pass when no windows configured")
	}
}

func TestWindowNotifier_NilInner_IsNoop(t *testing.T) {
	w := NewWindowNotifier(nil, []TimeWindow{makeWindow("00:00", "23:59")})
	w.now = fixedNow(12, 0)
	if err := w.Notify(context.Background(), Message{}); err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func TestWindowNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("inner error")
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return sentinel
	})
	w := NewWindowNotifier(inner, []TimeWindow{makeWindow("08:00", "18:00")})
	w.now = fixedNow(12, 0)

	if err := w.Notify(context.Background(), Message{}); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}

func TestWindowNotifier_SetWindows_UpdatesAtRuntime(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})
	w := NewWindowNotifier(inner, []TimeWindow{makeWindow("09:00", "17:00")})
	w.now = fixedNow(20, 0) // outside original window

	_ = w.Notify(context.Background(), Message{})
	if count != 0 {
		t.Fatal("expected drop before window update")
	}

	w.SetWindows([]TimeWindow{makeWindow("19:00", "21:00")})
	_ = w.Notify(context.Background(), Message{})
	if count != 1 {
		t.Error("expected delivery after window update")
	}
}
