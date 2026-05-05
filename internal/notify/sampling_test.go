package notify

import (
	"context"
	"testing"
)

func TestSamplingNotifier_AlwaysForwards(t *testing.T) {
	var received []Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		received = append(received, m)
		return nil
	})

	sn := NewSamplingNotifier(inner, 1.0)
	msg := Message{Subject: "job.check", Body: "ok"}

	for i := 0; i < 5; i++ {
		if err := sn.Notify(context.Background(), msg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if len(received) != 5 {
		t.Errorf("expected 5 forwarded, got %d", len(received))
	}
}

func TestSamplingNotifier_NeverForwards(t *testing.T) {
	var received []Message
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		received = append(received, m)
		return nil
	})

	sn := NewSamplingNotifier(inner, 0.0)
	msg := Message{Subject: "job.check", Body: "ok"}

	for i := 0; i < 10; i++ {
		if err := sn.Notify(context.Background(), msg); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if len(received) != 0 {
		t.Errorf("expected 0 forwarded, got %d", len(received))
	}
}

func TestSamplingNotifier_RateClamp(t *testing.T) {
	inner := NotifierFunc(func(_ context.Context, m Message) error { return nil })

	sn := NewSamplingNotifier(inner, 1.5)
	if sn.rate != 1.0 {
		t.Errorf("expected rate clamped to 1.0, got %f", sn.rate)
	}

	sn2 := NewSamplingNotifier(inner, -0.5)
	if sn2.rate != 0.0 {
		t.Errorf("expected rate clamped to 0.0, got %f", sn2.rate)
	}
}

func TestSamplingNotifier_SetRate(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		count++
		return nil
	})

	sn := NewSamplingNotifier(inner, 0.0)
	sn.randFunc = func() float64 { return 0.5 }

	_ = sn.Notify(context.Background(), Message{})
	if count != 0 {
		t.Fatal("expected no forward at rate 0.0")
	}

	sn.SetRate(1.0)
	_ = sn.Notify(context.Background(), Message{})
	if count != 1 {
		t.Fatalf("expected 1 forward after SetRate(1.0), got %d", count)
	}
}

func TestSamplingNotifier_NilInner(t *testing.T) {
	sn := NewSamplingNotifier(nil, 1.0)
	if err := sn.Notify(context.Background(), Message{Subject: "x"}); err != nil {
		t.Errorf("expected nil error for nil inner, got %v", err)
	}
}

func TestSamplingNotifier_DeterministicWithInjectedRand(t *testing.T) {
	var count int
	inner := NotifierFunc(func(_ context.Context, m Message) error {
		count++
		return nil
	})

	sn := NewSamplingNotifier(inner, 0.5)
	// rand returns 0.3 → below 0.5 → forward
	sn.randFunc = func() float64 { return 0.3 }
	_ = sn.Notify(context.Background(), Message{})
	if count != 1 {
		t.Errorf("expected 1 forward, got %d", count)
	}

	// rand returns 0.7 → above 0.5 → drop
	sn.randFunc = func() float64 { return 0.7 }
	_ = sn.Notify(context.Background(), Message{})
	if count != 1 {
		t.Errorf("expected still 1 forward after drop, got %d", count)
	}
}
