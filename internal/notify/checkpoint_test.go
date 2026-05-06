package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCheckpointNotifier_RecordsOnSuccess(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	cp := NewCheckpointNotifier(inner)
	fixed := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	cp.now = func() time.Time { return fixed }

	msg := Message{Subject: "job.backup"}
	if err := cp.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := cp.LastDelivery("job.backup")
	if !ok {
		t.Fatal("expected checkpoint to be set")
	}
	if !got.Equal(fixed) {
		t.Errorf("got %v, want %v", got, fixed)
	}
}

func TestCheckpointNotifier_NoRecordOnFailure(t *testing.T) {
	sentinel := errors.New("delivery failed")
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return sentinel })
	cp := NewCheckpointNotifier(inner)

	_ = cp.Notify(context.Background(), Message{Subject: "job.sync"})

	_, ok := cp.LastDelivery("job.sync")
	if ok {
		t.Fatal("checkpoint should not be set after failure")
	}
}

func TestCheckpointNotifier_Reset(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	cp := NewCheckpointNotifier(inner)
	cp.now = func() time.Time { return time.Now() }

	_ = cp.Notify(context.Background(), Message{Subject: "job.prune"})
	cp.Reset("job.prune")

	_, ok := cp.LastDelivery("job.prune")
	if ok {
		t.Fatal("expected checkpoint to be cleared after reset")
	}
}

func TestCheckpointNotifier_Snapshot(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	cp := NewCheckpointNotifier(inner)
	cp.now = func() time.Time { return time.Now() }

	subjects := []string{"job.a", "job.b", "job.c"}
	for _, s := range subjects {
		_ = cp.Notify(context.Background(), Message{Subject: s})
	}

	snap := cp.Snapshot()
	if len(snap) != len(subjects) {
		t.Errorf("snapshot length %d, want %d", len(snap), len(subjects))
	}
}

func TestCheckpointNotifier_NilInnerIsNoop(t *testing.T) {
	cp := NewCheckpointNotifier(nil)
	if err := cp.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("nil inner should be noop, got: %v", err)
	}
}

func TestCheckpointNotifier_MissingSubject(t *testing.T) {
	cp := NewCheckpointNotifier(nil)
	_, ok := cp.LastDelivery("nonexistent")
	if ok {
		t.Fatal("expected false for unknown subject")
	}
}
