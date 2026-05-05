package notify

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"
)

// AuditEntry records a single notification attempt.
type AuditEntry struct {
	Timestamp time.Time
	Subject   string
	Severity  string
	Success   bool
	Err       error
}

// AuditNotifier wraps a Notifier and records every dispatch attempt.
type AuditNotifier struct {
	mu      sync.Mutex
	inner   Notifier
	log     io.Writer
	entries []AuditEntry
}

// NewAuditNotifier returns an AuditNotifier that delegates to inner and writes
// a human-readable line to w on every call. w may be nil to suppress output.
func NewAuditNotifier(inner Notifier, w io.Writer) *AuditNotifier {
	return &AuditNotifier{inner: inner, log: w}
}

// Notify dispatches the message via the inner notifier and records the result.
func (a *AuditNotifier) Notify(ctx context.Context, msg Message) error {
	err := a.inner.Notify(ctx, msg)

	entry := AuditEntry{
		Timestamp: time.Now().UTC(),
		Subject:   msg.Subject,
		Severity:  msg.Severity,
		Success:   err == nil,
		Err:       err,
	}

	a.mu.Lock()
	a.entries = append(a.entries, entry)
	a.mu.Unlock()

	if a.log != nil {
		status := "OK"
		if err != nil {
			status = fmt.Sprintf("ERR: %v", err)
		}
		fmt.Fprintf(a.log, "%s [audit] subject=%q severity=%s status=%s\n",
			entry.Timestamp.Format(time.RFC3339), entry.Subject, entry.Severity, status)
	}

	return err
}

// Entries returns a snapshot of all recorded audit entries.
func (a *AuditNotifier) Entries() []AuditEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AuditEntry, len(a.entries))
	copy(out, a.entries)
	return out
}

// Reset clears all recorded entries.
func (a *AuditNotifier) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.entries = nil
}
