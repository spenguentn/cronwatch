package notify

import (
	"context"
	"testing"
)

func BenchmarkTagNotifier_Notify(b *testing.B) {
	noop := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tagger := NewTagNotifier(noop, map[string]string{
		"env":    "prod",
		"region": "us-east-1",
		"team":   "platform",
	})
	msg := Message{
		Subject: "job/backup-daily missed scheduled run",
		Body:    "expected at 02:00 UTC, last seen 26h ago",
		Meta:    map[string]string{"job": "backup-daily"},
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tagger.Notify(ctx, msg)
	}
}

func BenchmarkTagNotifier_NoExistingMeta(b *testing.B) {
	noop := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	tagger := NewTagNotifier(noop, map[string]string{"env": "prod"})
	msg := Message{Subject: "s", Body: "b"}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tagger.Notify(ctx, msg)
	}
}
