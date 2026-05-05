package notify

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestReplayNotifier_RecordsMessages(t *testing.T) {
	r := NewReplayNotifier(10)
	ctx := context.Background()

	_ = r.Notify(ctx, Message{Subject: "a"})
	_ = r.Notify(ctx, Message{Subject: "b"})

	if got := r.Len(); got != 2 {
		t.Fatalf("expected 2 buffered, got %d", got)
	}
}

func TestReplayNotifier_CapDropsOldest(t *testing.T) {
	r := NewReplayNotifier(2)
	ctx := context.Background()

	_ = r.Notify(ctx, Message{Subject: "first"})
	_ = r.Notify(ctx, Message{Subject: "second"})
	_ = r.Notify(ctx, Message{Subject: "third"})

	if got := r.Len(); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestReplayNotifier_ReplayDelivers(t *testing.T) {
	r := NewReplayNotifier(10)
	ctx := context.Background()

	_ = r.Notify(ctx, Message{Subject: "x"})
	_ = r.Notify(ctx, Message{Subject: "y"})

	var received []Message
	dst := NotifierFunc(func(_ context.Context, m Message) error {
		received = append(received, m)
		return nil
	})

	if err := r.Replay(ctx, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(received) != 2 {
		t.Fatalf("expected 2 delivered, got %d", len(received))
	}
	if r.Len() != 0 {
		t.Fatalf("buffer should be empty after successful replay")
	}
}

func TestReplayNotifier_ReplayKeepsFailedMessages(t *testing.T) {
	r := NewReplayNotifier(10)
	ctx := context.Background()

	_ = r.Notify(ctx, Message{Subject: "fail"})
	_ = r.Notify(ctx, Message{Subject: "ok"})

	calls := 0
	dst := NotifierFunc(func(_ context.Context, m Message) error {
		calls++
		if m.Subject == "fail" {
			return errors.New("send error")
		}
		return nil
	})

	err := r.Replay(ctx, dst)
	if err == nil {
		t.Fatal("expected error")
	}
	if r.Len() != 1 {
		t.Fatalf("expected 1 failed message retained, got %d", r.Len())
	}
}

func TestReplayNotifier_Reset(t *testing.T) {
	r := NewReplayNotifier(10)
	_ = r.Notify(context.Background(), Message{Subject: "z"})
	r.Reset()
	if r.Len() != 0 {
		t.Fatal("expected empty buffer after Reset")
	}
}

func TestReplayNotifier_DefaultCapacity(t *testing.T) {
	r := NewReplayNotifier(0)
	ctx := context.Background()
	for i := 0; i < 105; i++ {
		_ = r.Notify(ctx, Message{Subject: fmt.Sprintf("msg-%d", i)})
	}
	if r.Len() != 100 {
		t.Fatalf("expected 100 (default cap), got %d", r.Len())
	}
}

func TestReplayNotifier_OldestAge(t *testing.T) {
	r := NewReplayNotifier(10)
	if r.OldestAge() != 0 {
		t.Fatal("expected zero age for empty buffer")
	}

	msg := Message{
		Subject: "aged",
		Meta:    map[string]string{"timestamp": time.Now().Add(-5 * time.Second).Format(time.RFC3339Nano)},
	}
	_ = r.Notify(context.Background(), msg)

	age := r.OldestAge()
	if age < 4*time.Second {
		t.Fatalf("expected age >= 4s, got %v", age)
	}
}
