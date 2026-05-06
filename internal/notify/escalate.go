package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EscalateNotifier forwards a message to a primary notifier and, if the
// primary fails or the message is not acknowledged within a deadline,
// escalates to a secondary notifier.
type EscalateNotifier struct {
	primary   Notifier
	secondary Notifier
	deadline  time.Duration
	mu        sync.Mutex
	pending   map[string]time.Time
}

// NewEscalateNotifier creates an EscalateNotifier. If the primary notifier
// returns an error, or Acknowledge is not called within deadline, the message
// is forwarded to secondary. A zero deadline disables time-based escalation.
func NewEscalateNotifier(primary, secondary Notifier, deadline time.Duration) *EscalateNotifier {
	return &EscalateNotifier{
		primary:   primary,
		secondary: secondary,
		deadline:  deadline,
		pending:   make(map[string]time.Time),
	}
}

// Notify sends msg to primary. On failure it immediately escalates to secondary.
func (e *EscalateNotifier) Notify(ctx context.Context, msg Message) error {
	if e.primary == nil {
		return e.notifySecondary(ctx, msg)
	}
	err := e.primary.Notify(ctx, msg)
	if err != nil {
		if e.secondary == nil {
			return fmt.Errorf("escalate: primary failed and no secondary: %w", err)
		}
		return e.notifySecondary(ctx, msg)
	}
	if e.deadline > 0 {
		e.mu.Lock()
		e.pending[msg.Subject] = time.Now().Add(e.deadline)
		e.mu.Unlock()
	}
	return nil
}

// Acknowledge marks a message subject as acknowledged, preventing escalation.
func (e *EscalateNotifier) Acknowledge(subject string) {
	e.mu.Lock()
	delete(e.pending, subject)
	e.mu.Unlock()
}

// Tick checks for unacknowledged messages past their deadline and escalates them.
func (e *EscalateNotifier) Tick(ctx context.Context, now time.Time) {
	e.mu.Lock()
	var overdue []string
	for subject, deadline := range e.pending {
		if now.After(deadline) {
			overdue = append(overdue, subject)
		}
	}
	for _, s := range overdue {
		delete(e.pending, s)
	}
	e.mu.Unlock()

	for _, subject := range overdue {
		_ = e.notifySecondary(ctx, Message{
			Subject:  "[ESCALATED] " + subject,
			Body:     "No acknowledgement received within deadline.",
			Severity: SeverityError,
		})
	}
}

func (e *EscalateNotifier) notifySecondary(ctx context.Context, msg Message) error {
	if e.secondary == nil {
		return nil
	}
	return e.secondary.Notify(ctx, msg)
}
