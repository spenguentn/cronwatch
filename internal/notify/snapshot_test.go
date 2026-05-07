package notify

import (
	"context"
	"errors"
	"testing"
)

func TestSnapshotNotifier_RecordsSuccess(t *testing.T) {
	sn := NewSnapshotNotifier(NotifierFunc(func(_ context.Context, _ Message) error {
		return nil
	}), 10)

	msg := Message{Subject: "job.ok", Body: "all good", Severity: "info"}
	if err := sn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries := sn.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Subject != "job.ok" {
		t.Errorf("expected subject job.ok, got %s", entries[0].Subject)
	}
	if entries[0].Err != nil {
		t.Errorf("expected nil error, got %v", entries[0].Err)
	}
}

func TestSnapshotNotifier_RecordsFailure(t *testing.T) {
	sentinel := errors.New("send failed")
	sn := NewSnapshotNotifier(NotifierFunc(func(_ context.Context, _ Message) error {
		return sentinel
	}), 10)

	_ = sn.Notify(context.Background(), Message{Subject: "job.fail"})

	entries := sn.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if !errors.Is(entries[0].Err, sentinel) {
		t.Errorf("expected sentinel error, got %v", entries[0].Err)
	}
}

func TestSnapshotNotifier_CapDropsOldest(t *testing.T) {
	sn := NewSnapshotNotifier(nil, 3)

	for i := 0; i < 5; i++ {
		_ = sn.Notify(context.Background(), Message{Subject: string(rune('a'+i))})
	}

	entries := sn.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Subject != "c" {
		t.Errorf("expected oldest retained entry 'c', got %s", entries[0].Subject)
	}
}

func TestSnapshotNotifier_Reset(t *testing.T) {
	sn := NewSnapshotNotifier(nil, 10)
	_ = sn.Notify(context.Background(), Message{Subject: "x"})
	sn.Reset()

	if len(sn.Entries()) != 0 {
		t.Error("expected empty entries after reset")
	}
}

func TestSnapshotNotifier_NilInnerIsNoop(t *testing.T) {
	sn := NewSnapshotNotifier(nil, 5)
	if err := sn.Notify(context.Background(), Message{Subject: "noop"}); err != nil {
		t.Fatalf("expected nil error for nil inner, got %v", err)
	}
	if len(sn.Entries()) != 1 {
		t.Error("expected entry recorded even with nil inner")
	}
}

func TestSnapshotNotifier_DefaultCap(t *testing.T) {
	sn := NewSnapshotNotifier(nil, 0)
	if sn.cap != 64 {
		t.Errorf("expected default cap 64, got %d", sn.cap)
	}
}
