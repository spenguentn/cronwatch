package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTimeoutNotifier_DeliversBeforeDeadline(t *testing.T) {
	var received Message
	inner := NotifierFunc(func(_ context.Context, msg Message) error {
		received = msg
		return nil
	})

	n := NewTimeoutNotifier(inner, 100*time.Millisecond)
	msg := Message{Subject: "ok", Body: "body"}

	if err := n.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Subject != msg.Subject {
		t.Errorf("got subject %q, want %q", received.Subject, msg.Subject)
	}
}

func TestTimeoutNotifier_TimesOut(t *testing.T) {
	slow := NotifierFunc(func(ctx context.Context, _ Message) error {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	n := NewTimeoutNotifier(slow, 20*time.Millisecond)
	err := n.Notify(context.Background(), Message{Subject: "slow"})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestTimeoutNotifier_RespectsParentCancel(t *testing.T) {
	slow := NotifierFunc(func(ctx context.Context, _ Message) error {
		<-ctx.Done()
		return ctx.Err()
	})

	n := NewTimeoutNotifier(slow, 2*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := n.Notify(ctx, Message{Subject: "cancelled"})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestTimeoutNotifier_NilInnerIsNoop(t *testing.T) {
	n := NewTimeoutNotifier(nil, 50*time.Millisecond)
	if err := n.Notify(context.Background(), Message{Subject: "noop"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTimeoutNotifier_ZeroDurationDefaulted(t *testing.T) {
	n := NewTimeoutNotifier(nil, 0)
	if n.timeout != 5*time.Second {
		t.Errorf("expected default 5s, got %v", n.timeout)
	}
}

func TestTimeoutNotifier_PropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return sentinel
	})

	n := NewTimeoutNotifier(inner, 100*time.Millisecond)
	err := n.Notify(context.Background(), Message{Subject: "err"})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
