package notify

import (
	"bytes"
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"
)

// syncNotifier records calls and can be made to fail.
type syncNotifier struct {
	mu   sync.Mutex
	msgs []Message
	err  error
}

func (n *syncNotifier) Notify(_ context.Context, msg Message) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.msgs = append(n.msgs, msg)
	return n.err
}

func (n *syncNotifier) count() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.msgs)
}

func TestShadowNotifier_BothReceiveMessage(t *testing.T) {
	primary := &syncNotifier{}
	shadow := &syncNotifier{}
	sn := NewShadowNotifier(primary, shadow, nil)

	msg := Message{Subject: "hello", Body: "world"}
	if err := sn.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Allow the goroutine to complete.
	time.Sleep(20 * time.Millisecond)

	if primary.count() != 1 {
		t.Errorf("primary: want 1 message, got %d", primary.count())
	}
	if shadow.count() != 1 {
		t.Errorf("shadow: want 1 message, got %d", shadow.count())
	}
}

func TestShadowNotifier_PrimaryErrorReturned(t *testing.T) {
	primary := &syncNotifier{err: errors.New("primary down")}
	shadow := &syncNotifier{}
	sn := NewShadowNotifier(primary, shadow, nil)

	err := sn.Notify(context.Background(), Message{Subject: "x"})
	if err == nil {
		t.Fatal("expected error from primary")
	}
}

func TestShadowNotifier_ShadowErrorLogged(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	primary := &syncNotifier{}
	shadow := &syncNotifier{err: errors.New("shadow down")}
	sn := NewShadowNotifier(primary, shadow, logger)

	if err := sn.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if buf.Len() == 0 {
		t.Error("expected shadow error to be logged")
	}
}

func TestShadowNotifier_NilPrimaryIsNoop(t *testing.T) {
	sn := NewShadowNotifier(nil, nil, nil)
	if err := sn.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShadowNotifier_NilShadowSafe(t *testing.T) {
	primary := &syncNotifier{}
	sn := NewShadowNotifier(primary, nil, nil)
	if err := sn.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if primary.count() != 1 {
		t.Errorf("primary: want 1 message, got %d", primary.count())
	}
}
