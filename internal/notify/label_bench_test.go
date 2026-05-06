package notify

import (
	"context"
	"testing"
)

func BenchmarkLabelNotifier_Notify(b *testing.B) {
	noop := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	n := NewLabelNotifier(noop, map[string]string{
		"env":    "prod",
		"region": "us-east-1",
		"tier":   "backend",
	})
	msg := Message{Subject: "benchmark", Body: "body"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = n.Notify(context.Background(), msg)
	}
}

func BenchmarkLabelNotifier_NoExistingMeta(b *testing.B) {
	noop := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	n := NewLabelNotifier(noop, map[string]string{"k": "v"})
	msg := Message{Subject: "bench", Meta: map[string]string{"existing": "value"}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = n.Notify(context.Background(), msg)
	}
}
