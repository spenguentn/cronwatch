package notify

import (
	"context"
	"errors"
	"testing"
)

func TestPriority_String(t *testing.T) {
	cases := []struct {
		p    Priority
		want string
	}{
		{PriorityLow, "low"},
		{PriorityNormal, "normal"},
		{PriorityHigh, "high"},
		{Priority(99), "unknown(99)"},
	}
	for _, tc := range cases {
		if got := tc.p.String(); got != tc.want {
			t.Errorf("Priority(%d).String() = %q, want %q", int(tc.p), got, tc.want)
		}
	}
}

func TestWithPriority_SetsMeta(t *testing.T) {
	msg := Message{Subject: "test"}
	out := WithPriority(msg, PriorityHigh)
	if got := out.Meta["priority"]; got != "high" {
		t.Errorf("meta[priority] = %q, want %q", got, "high")
	}
}

func TestWithPriority_PreservesExistingMeta(t *testing.T) {
	msg := Message{Subject: "test", Meta: map[string]string{"source": "cronwatch"}}
	out := WithPriority(msg, PriorityLow)
	if got := out.Meta["source"]; got != "cronwatch" {
		t.Errorf("meta[source] = %q, want %q", got, "cronwatch")
	}
}

func TestPriorityRouter_RoutesLow(t *testing.T) {
	var routed string
	make := func(label string) Notifier {
		return notifierFunc(func(_ context.Context, _ Message) error {
			routed = label
			return nil
		})
	}
	r := NewPriorityRouter(make("low"), make("normal"), make("high"))
	msg := WithPriority(Message{Subject: "drift"}, PriorityLow)
	if err := r.Notify(context.Background(), msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if routed != "low" {
		t.Errorf("routed to %q, want %q", routed, "low")
	}
}

func TestPriorityRouter_RoutesHigh(t *testing.T) {
	var routed string
	make := func(label string) Notifier {
		return notifierFunc(func(_ context.Context, _ Message) error {
			routed = label
			return nil
		})
	}
	r := NewPriorityRouter(make("low"), make("normal"), make("high"))
	msg := WithPriority(Message{Subject: "missed"}, PriorityHigh)
	_ = r.Notify(context.Background(), msg)
	if routed != "high" {
		t.Errorf("routed to %q, want %q", routed, "high")
	}
}

func TestPriorityRouter_FallsBackToNormal(t *testing.T) {
	var routed string
	make := func(label string) Notifier {
		return notifierFunc(func(_ context.Context, _ Message) error {
			routed = label
			return nil
		})
	}
	r := NewPriorityRouter(make("low"), make("normal"), make("high"))
	// No priority metadata set.
	_ = r.Notify(context.Background(), Message{Subject: "no-priority"})
	if routed != "normal" {
		t.Errorf("routed to %q, want %q", routed, "normal")
	}
}

func TestPriorityRouter_NilNotifierDrops(t *testing.T) {
	r := NewPriorityRouter(nil, nil, nil)
	msg := WithPriority(Message{Subject: "x"}, PriorityHigh)
	if err := r.Notify(context.Background(), msg); err != nil {
		t.Errorf("expected nil error for nil notifier, got %v", err)
	}
}

func TestPriorityRouter_PropagatesError(t *testing.T) {
	sentinel := errors.New("notify failed")
	fail := notifierFunc(func(_ context.Context, _ Message) error { return sentinel })
	r := NewPriorityRouter(nil, fail, nil)
	msg := WithPriority(Message{Subject: "x"}, PriorityNormal)
	if err := r.Notify(context.Background(), msg); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}
