package notify

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestEscalateNotifier_PrimarySuccess(t *testing.T) {
	var got []Message
	primary := NotifierFunc(func(_ context.Context, m Message) error {
		got = append(got, m)
		return nil
	})
	secondary := NotifierFunc(func(_ context.Context, m Message) error {
		t.Fatal("secondary should not be called")
		return nil
	})
	e := NewEscalateNotifier(primary, secondary, 0)
	if err := e.Notify(context.Background(), Message{Subject: "job.ok"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 primary delivery, got %d", len(got))
	}
}

func TestEscalateNotifier_PrimaryFailEscalates(t *testing.T) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("primary down")
	})
	var mu sync.Mutex
	var escalated []Message
	secondary := NotifierFunc(func(_ context.Context, m Message) error {
		mu.Lock()
		escalated = append(escalated, m)
		mu.Unlock()
		return nil
	})
	e := NewEscalateNotifier(primary, secondary, 0)
	if err := e.Notify(context.Background(), Message{Subject: "job.fail"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(escalated) != 1 {
		t.Fatalf("expected 1 escalation, got %d", len(escalated))
	}
}

func TestEscalateNotifier_Acknowledge_PreventsEscalation(t *testing.T) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	secondary := NotifierFunc(func(_ context.Context, _ Message) error {
		t.Fatal("secondary should not be called after ack")
		return nil
	})
	e := NewEscalateNotifier(primary, secondary, 50*time.Millisecond)
	_ = e.Notify(context.Background(), Message{Subject: "acked-job"})
	e.Acknowledge("acked-job")
	time.Sleep(80 * time.Millisecond)
	e.Tick(context.Background(), time.Now())
}

func TestEscalateNotifier_Tick_EscalatesOverdue(t *testing.T) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	var escalated []Message
	secondary := NotifierFunc(func(_ context.Context, m Message) error {
		escalated = append(escalated, m)
		return nil
	})
	e := NewEscalateNotifier(primary, secondary, 10*time.Millisecond)
	_ = e.Notify(context.Background(), Message{Subject: "slow-job"})
	time.Sleep(30 * time.Millisecond)
	e.Tick(context.Background(), time.Now())
	if len(escalated) != 1 {
		t.Fatalf("expected 1 escalation via Tick, got %d", len(escalated))
	}
}

func TestEscalateNotifier_NilPrimaryEscalates(t *testing.T) {
	var got []Message
	secondary := NotifierFunc(func(_ context.Context, m Message) error {
		got = append(got, m)
		return nil
	})
	e := NewEscalateNotifier(nil, secondary, 0)
	if err := e.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected secondary to receive message")
	}
}
