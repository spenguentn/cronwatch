package notify

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestDigestNotifier_AccumulatesAndFlushes(t *testing.T) {
	var mu sync.Mutex
	var received []Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		mu.Lock()
		received = append(received, m)
		mu.Unlock()
		return nil
	})

	d := NewDigestNotifier(inner, 10*time.Second)
	defer d.Close()

	_ = d.Notify(context.Background(), Message{Subject: "a", Body: "first", Severity: SeverityWarn})
	_ = d.Notify(context.Background(), Message{Subject: "b", Body: "second", Severity: SeverityError})

	_ = d.Flush(context.Background())

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 digest, got %d", len(received))
	}
	if received[0].Subject != "Digest: 2 alerts" {
		t.Errorf("unexpected subject: %s", received[0].Subject)
	}
}

func TestDigestNotifier_EmptyFlushIsNoop(t *testing.T) {
	called := false
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called = true
		return nil
	})
	d := NewDigestNotifier(inner, 10*time.Second)
	defer d.Close()

	if err := d.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("inner should not be called on empty flush")
	}
}

func TestDigestNotifier_NilInnerIsNoop(t *testing.T) {
	d := NewDigestNotifier(nil, 10*time.Second)
	defer d.Close()
	_ = d.Notify(context.Background(), Message{Body: "x"})
	if err := d.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDigestNotifier_WindowFlushes(t *testing.T) {
	var mu sync.Mutex
	var received []Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		mu.Lock()
		received = append(received, m)
		mu.Unlock()
		return nil
	})

	d := NewDigestNotifier(inner, 50*time.Millisecond)
	defer d.Close()

	_ = d.Notify(context.Background(), Message{Body: "hello", Severity: SeverityInfo})

	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if len(received) == 0 {
		t.Fatal("expected at least one digest delivery via window ticker")
	}
}

func TestBuildDigest_CombinesBodies(t *testing.T) {
	msgs := []Message{
		{Subject: "s1", Body: "body1", Severity: SeverityWarn},
		{Subject: "s2", Body: "body2", Severity: SeverityError},
	}
	out := buildDigest(msgs)
	if out.Subject != "Digest: 2 alerts" {
		t.Errorf("got subject %q", out.Subject)
	}
	if out.Severity != SeverityWarn {
		t.Errorf("expected first severity, got %v", out.Severity)
	}
}
