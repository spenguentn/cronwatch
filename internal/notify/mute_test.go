package notify

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestMuteNotifier_ForwardsWhenNotMuted(t *testing.T) {
	var called atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called.Add(1)
		return nil
	})
	m := NewMuteNotifier(inner)
	if err := m.Notify(context.Background(), Message{Subject: "hello"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() != 1 {
		t.Fatalf("expected inner called once, got %d", called.Load())
	}
}

func TestMuteNotifier_SuppressesDuringMute(t *testing.T) {
	var called atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called.Add(1)
		return nil
	})
	m := NewMuteNotifier(inner)
	m.Mute(10 * time.Minute)
	if err := m.Notify(context.Background(), Message{Subject: "muted"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() != 0 {
		t.Fatalf("expected inner not called, got %d", called.Load())
	}
}

func TestMuteNotifier_UnmuteAllowsDelivery(t *testing.T) {
	var called atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called.Add(1)
		return nil
	})
	m := NewMuteNotifier(inner)
	m.Mute(10 * time.Minute)
	m.Unmute()
	if err := m.Notify(context.Background(), Message{Subject: "unmuted"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() != 1 {
		t.Fatalf("expected inner called once after unmute, got %d", called.Load())
	}
}

func TestMuteNotifier_IsMuted(t *testing.T) {
	m := NewMuteNotifier(nil)
	if m.IsMuted() {
		t.Fatal("expected not muted initially")
	}
	m.Mute(5 * time.Minute)
	if !m.IsMuted() {
		t.Fatal("expected muted after Mute()")
	}
	m.Unmute()
	if m.IsMuted() {
		t.Fatal("expected not muted after Unmute()")
	}
}

func TestMuteNotifier_NilInnerIsNoop(t *testing.T) {
	m := NewMuteNotifier(nil)
	if err := m.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Fatalf("unexpected error with nil inner: %v", err)
	}
}

func TestMuteNotifier_PropagatesError(t *testing.T) {
	sentinel := errors.New("inner error")
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		return sentinel
	})
	m := NewMuteNotifier(inner)
	if err := m.Notify(context.Background(), Message{Subject: "err"}); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestMuteNotifier_MuteExpiry(t *testing.T) {
	var called atomic.Int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		called.Add(1)
		return nil
	})
	now := time.Now()
	m := NewMuteNotifier(inner)
	m.nowFn = func() time.Time { return now }
	m.Mute(1 * time.Second)
	// advance time past mute window
	m.nowFn = func() time.Time { return now.Add(2 * time.Second) }
	if err := m.Notify(context.Background(), Message{Subject: "after expiry"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() != 1 {
		t.Fatalf("expected delivery after mute expiry, got %d calls", called.Load())
	}
}
