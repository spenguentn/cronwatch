package notify

import (
	"context"
	"errors"
	"testing"
)

func TestFallbackNotifier_PrimarySucceeds(t *testing.T) {
	calls := 0
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		calls++
		return nil
	})
	secondary := NotifierFunc(func(_ context.Context, _ Message) error {
		t.Fatal("secondary should not be called")
		return nil
	})

	fn := NewFallbackNotifier(primary, secondary)
	if err := fn.Notify(context.Background(), Message{Subject: "ok"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 primary call, got %d", calls)
	}
}

func TestFallbackNotifier_FallsBackOnPrimaryFailure(t *testing.T) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("primary down")
	})
	fbCalls := 0
	fallback := NotifierFunc(func(_ context.Context, _ Message) error {
		fbCalls++
		return nil
	})

	fn := NewFallbackNotifier(primary, fallback)
	if err := fn.Notify(context.Background(), Message{Subject: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fbCalls != 1 {
		t.Fatalf("expected 1 fallback call, got %d", fbCalls)
	}
}

func TestFallbackNotifier_AllFail(t *testing.T) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("primary err")
	})
	fallback := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("fallback err")
	})

	fn := NewFallbackNotifier(primary, fallback)
	if err := fn.Notify(context.Background(), Message{Subject: "test"}); err == nil {
		t.Fatal("expected error when all notifiers fail")
	}
}

func TestFallbackNotifier_NilPrimaryUsesFallback(t *testing.T) {
	fbCalls := 0
	fallback := NotifierFunc(func(_ context.Context, _ Message) error {
		fbCalls++
		return nil
	})

	fn := NewFallbackNotifier(nil, fallback)
	if err := fn.Notify(context.Background(), Message{Subject: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fbCalls != 1 {
		t.Fatalf("expected 1 fallback call, got %d", fbCalls)
	}
}

func TestFallbackNotifier_SkipsNilFallbacks(t *testing.T) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("primary err")
	})
	calls := 0
	real := NotifierFunc(func(_ context.Context, _ Message) error {
		calls++
		return nil
	})

	fn := NewFallbackNotifier(primary, nil, real)
	if err := fn.Notify(context.Background(), Message{Subject: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected real fallback to be called once, got %d", calls)
	}
}

func TestFallbackNotifier_SecondFallbackUsedWhenFirstFails(t *testing.T) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("primary err")
	})
	fb1 := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("fb1 err")
	})
	fb2Calls := 0
	fb2 := NotifierFunc(func(_ context.Context, _ Message) error {
		fb2Calls++
		return nil
	})

	fn := NewFallbackNotifier(primary, fb1, fb2)
	if err := fn.Notify(context.Background(), Message{Subject: "test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb2Calls != 1 {
		t.Fatalf("expected fb2 to be called once, got %d", fb2Calls)
	}
}
