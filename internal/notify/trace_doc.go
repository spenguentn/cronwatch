// Package notify provides composable notification primitives for cronwatch.
//
// # TraceNotifier
//
// TraceNotifier wraps any Notifier and records every Notify call as a
// TraceEntry, capturing the wall-clock timestamp, message subject, result
// error, and round-trip latency.
//
// A human-readable trace line is written to the configured io.Writer (defaults
// to os.Stderr) on every call:
//
//	[trace] 2024-01-15T10:00:00Z subject="backup" latency=1.234ms status=ok
//
// Entries can be inspected programmatically via Entries() for testing or
// observability pipelines, and cleared with Reset().
//
// # Usage
//
//	var buf bytes.Buffer
//	tn := notify.NewTraceNotifier(inner, &buf)
//	_ = tn.Notify(ctx, msg)
//	for _, e := range tn.Entries() {
//		fmt.Println(e.Subject, e.Latency, e.Err)
//	}
package notify
