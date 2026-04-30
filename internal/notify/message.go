package notify

import "context"

// Severity represents the urgency level of a notification.
type Severity int

const (
	// SeverityInfo is an informational alert.
	SeverityInfo Severity = iota
	// SeverityWarn indicates a non-critical anomaly such as drift.
	SeverityWarn
	// SeverityError indicates a critical failure such as a missed run.
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "INFO"
	case SeverityWarn:
		return "WARN"
	case SeverityError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Message is the notification payload sent by every Notifier.
type Message struct {
	// Subject is a short one-line summary of the alert.
	Subject string
	// Body contains the full human-readable detail.
	Body string
	// Severity classifies the urgency of the alert.
	Severity Severity
	// Job is the name of the cron job that triggered the alert.
	Job string
}

// Notifier is the common interface implemented by all notification backends.
type Notifier interface {
	// Notify sends msg using the underlying transport.
	// Implementations must respect ctx cancellation.
	Notify(ctx context.Context, msg Message) error
}
