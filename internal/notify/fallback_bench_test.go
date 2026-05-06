package notify

import (
	"context"
	"errors"
	"testing"
)

func BenchmarkFallbackNotifier_PrimarySucceeds(b *testing.B) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		return nil
	})
	fallback := NotifierFunc(func(_ context.Context, _ Message) error {
		return nil
	})
	fn := NewFallbackNotifier(primary, fallback)
	msg := Message{Subject: "bench", Body: "body"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn.Notify(context.Background(), msg)
	}
}

func BenchmarkFallbackNotifier_FallsBack(b *testing.B) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error {
		return errors.New("down")
	})
	fallback := NotifierFunc(func(_ context.Context, _ Message) error {
		return nil
	})
	fn := NewFallbackNotifier(primary, fallback)
	msg := Message{Subject: "bench", Body: "body"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn.Notify(context.Background(), msg)
	}
}
