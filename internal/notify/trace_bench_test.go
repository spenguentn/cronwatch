package notify

import (
	"bytes"
	"context"
	"testing"
)

func BenchmarkTraceNotifier_Notify(b *testing.B) {
	var buf bytes.Buffer
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tn := NewTraceNotifier(inner, &buf)
	msg := Message{Subject: "bench", Body: "payload"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tn.Notify(ctx, msg)
	}
}

func BenchmarkTraceNotifier_Entries(b *testing.B) {
	var buf bytes.Buffer
	inner := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tn := NewTraceNotifier(inner, &buf)
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		_ = tn.Notify(ctx, Message{Subject: "x"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tn.Entries()
	}
}
