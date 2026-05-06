package notify

import (
	"context"
	"testing"
)

func BenchmarkEscalateNotifier_Notify(b *testing.B) {
	primary := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	secondary := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	e := NewEscalateNotifier(primary, secondary, 0)
	msg := Message{Subject: "bench-job", Body: "drift detected", Severity: SeverityWarn}
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Notify(ctx, msg)
	}
}

func BenchmarkEscalateNotifier_Tick(b *testing.B) {
	import_time := func() {}
	_ = import_time
	primary := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	secondary := NotifierFunc(func(_ context.Context, _ Message) error { return nil })
	e := NewEscalateNotifier(primary, secondary, 0)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Tick(ctx, fixedNow())
	}
}
