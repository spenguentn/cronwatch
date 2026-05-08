package notify

import (
	"context"
	"errors"
	"testing"
)

func TestStickyNotifier_ForwardsAndRecords(t *testing.T) {
	var received []Message
	inner := NotifierFunc(func(_ context.Context, msg Message) error {
		received = append(received, msg)
		return nil
	})
	sn := NewStickyNotifier(inner)
	msg := Message{Subject: "job.a", Body: "missed run"}
	if err := sn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(received))
	}
	last, ok := sn.Last("job.a")
	if !ok {
		t.Fatal("expected last message to be recorded")
	}
	if last.Body != "missed run" {
		t.Errorf("unexpected body: %s", last.Body)
	}
}

func TestStickyNotifier_Resend(t *testing.T) {
	count := 0
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		count++
		return nil
	})
	sn := NewStickyNotifier(inner)
	_ = sn.Notify(context.Background(), Message{Subject: "job.b", Body: "drift detected"})
	if err := sn.Resend(context.Background(), "job.b"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 deliveries, got %d", count)
	}
}

func TestStickyNotifier_ResendMissing(t *testing.T) {
	sn := NewStickyNotifier(NotifierFunc(func(_ context.Context, _ Message) error { return nil }))
	err := sn.Resend(context.Background(), "unknown")
	if err == nil {
		t.Fatal("expected error for unknown subject")
	}
}

func TestStickyNotifier_Reset(t *testing.T) {
	sn := NewStickyNotifier(NotifierFunc(func(_ context.Context, _ Message) error { return nil }))
	_ = sn.Notify(context.Background(), Message{Subject: "job.c", Body: "x"})
	sn.Reset()
	_, ok := sn.Last("job.c")
	if ok {
		t.Error("expected no message after reset")
	}
}

func TestStickyNotifier_NilInnerIsNoop(t *testing.T) {
	sn := NewStickyNotifier(nil)
	if err := sn.Notify(context.Background(), Message{Subject: "x", Body: "y"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStickyNotifier_InnerErrorNotRecorded(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("delivery failed")
	})
	sn := NewStickyNotifier(inner)
	_ = sn.Notify(context.Background(), Message{Subject: "job.d", Body: "alert"})
	_, ok := sn.Last("job.d")
	if ok {
		t.Error("expected no record when inner returns error")
	}
}
