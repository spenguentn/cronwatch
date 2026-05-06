package notify

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// TraceEntry records a single notification attempt.
type TraceEntry struct {
	At      time.Time
	Subject string
	Err     error
	Latency time.Duration
}

// TraceNotifier wraps an inner Notifier and writes trace entries to a writer
// for every Notify call, recording latency and outcome.
type TraceNotifier struct {
	inner  Notifier
	out    io.Writer
	entries []TraceEntry
}

// NewTraceNotifier returns a TraceNotifier that writes human-readable trace
// lines to out (defaults to os.Stderr when nil).
func NewTraceNotifier(inner Notifier, out io.Writer) *TraceNotifier {
	if out == nil {
		out = os.Stderr
	}
	return &TraceNotifier{inner: inner, out: out}
}

// Notify records latency and outcome, writes a trace line, then returns the
// inner error unchanged.
func (t *TraceNotifier) Notify(ctx context.Context, msg Message) error {
	if t.inner == nil {
		return nil
	}
	start := time.Now()
	err := t.inner.Notify(ctx, msg)
	latency := time.Since(start)

	entry := TraceEntry{
		At:      start,
		Subject: msg.Subject,
		Err:     err,
		Latency: latency,
	}
	t.entries = append(t.entries, entry)

	status := "ok"
	if err != nil {
		status = fmt.Sprintf("err=%v", err)
	}
	fmt.Fprintf(t.out, "[trace] %s subject=%q latency=%s status=%s\n",
		start.Format(time.RFC3339), msg.Subject, latency.Round(time.Microsecond), status)

	return err
}

// Entries returns a snapshot of all recorded trace entries.
func (t *TraceNotifier) Entries() []TraceEntry {
	out := make([]TraceEntry, len(t.entries))
	copy(out, t.entries)
	return out
}

// Reset clears all recorded entries.
func (t *TraceNotifier) Reset() {
	t.entries = nil
}
