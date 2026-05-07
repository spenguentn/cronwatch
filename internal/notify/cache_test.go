package notify

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheNotifier_FirstCallDelivers(t *testing.T) {
	var calls int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	c := NewCacheNotifier(inner, time.Minute)
	if err := c.Notify(context.Background(), Message{Subject: "s", Body: "b"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestCacheNotifier_DuplicateSuppressed(t *testing.T) {
	var calls int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	c := NewCacheNotifier(inner, time.Minute)
	msg := Message{Subject: "alert", Body: "disk full"}
	_ = c.Notify(context.Background(), msg)
	_ = c.Notify(context.Background(), msg)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestCacheNotifier_DifferentMessagesDeliver(t *testing.T) {
	var calls int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	c := NewCacheNotifier(inner, time.Minute)
	_ = c.Notify(context.Background(), Message{Subject: "a", Body: "x"})
	_ = c.Notify(context.Background(), Message{Subject: "b", Body: "y"})
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}
}

func TestCacheNotifier_TTLExpiry_Redelivers(t *testing.T) {
	var calls int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	c := NewCacheNotifier(inner, 50*time.Millisecond)
	msg := Message{Subject: "s", Body: "b"}
	_ = c.Notify(context.Background(), msg)
	time.Sleep(80 * time.Millisecond)
	_ = c.Notify(context.Background(), msg)
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls after TTL expiry, got %d", calls)
	}
}

func TestCacheNotifier_ErrorNotCached(t *testing.T) {
	var calls int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("fail")
	})
	c := NewCacheNotifier(inner, time.Minute)
	msg := Message{Subject: "s", Body: "b"}
	_ = c.Notify(context.Background(), msg)
	_ = c.Notify(context.Background(), msg)
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls on repeated errors, got %d", calls)
	}
}

func TestCacheNotifier_Invalidate_ForcesRedeliver(t *testing.T) {
	var calls int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	c := NewCacheNotifier(inner, time.Minute)
	msg := Message{Subject: "s", Body: "b"}
	_ = c.Notify(context.Background(), msg)
	c.Invalidate(msg)
	_ = c.Notify(context.Background(), msg)
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 calls after invalidation, got %d", calls)
	}
}

func TestCacheNotifier_Flush_ClearsAll(t *testing.T) {
	var calls int32
	inner := NotifierFunc(func(_ context.Context, _ Message) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	c := NewCacheNotifier(inner, time.Minute)
	for _, s := range []string{"a", "b", "c"} {
		_ = c.Notify(context.Background(), Message{Subject: s})
	}
	c.Flush()
	for _, s := range []string{"a", "b", "c"} {
		_ = c.Notify(context.Background(), Message{Subject: s})
	}
	if atomic.LoadInt32(&calls) != 6 {
		t.Fatalf("expected 6 calls after flush, got %d", calls)
	}
}

func TestCacheNotifier_NilInnerIsNoop(t *testing.T) {
	c := NewCacheNotifier(nil, time.Minute)
	if err := c.Notify(context.Background(), Message{Subject: "s"}); err != nil {
		t.Fatalf("expected nil error for nil inner, got %v", err)
	}
}

func TestCacheNotifier_DefaultTTL(t *testing.T) {
	c := NewCacheNotifier(NotifierFunc(func(_ context.Context, _ Message) error { return nil }), 0)
	if c.ttl != 5*time.Minute {
		t.Fatalf("expected default TTL of 5m, got %v", c.ttl)
	}
}
