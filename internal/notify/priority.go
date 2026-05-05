package notify

import (
	"context"
	"fmt"
)

// Priority represents the urgency level of a notification.
type Priority int

const (
	// PriorityLow is used for informational alerts.
	PriorityLow Priority = iota
	// PriorityNormal is the default priority.
	PriorityNormal
	// PriorityHigh is used for critical alerts requiring immediate attention.
	PriorityHigh
)

// String returns a human-readable priority label.
func (p Priority) String() string {
	switch p {
	case PriorityLow:
		return "low"
	case PriorityNormal:
		return "normal"
	case PriorityHigh:
		return "high"
	default:
		return fmt.Sprintf("unknown(%d)", int(p))
	}
}

// PriorityRouter routes a message to different inner notifiers based on
// the "priority" metadata key. If no priority metadata is present or the
// value is unrecognised, the normal-priority notifier is used as fallback.
type PriorityRouter struct {
	low    Notifier
	normal Notifier
	high   Notifier
}

// NewPriorityRouter constructs a PriorityRouter. Any of the notifiers may be
// nil, in which case messages routed to that tier are silently dropped.
func NewPriorityRouter(low, normal, high Notifier) *PriorityRouter {
	return &PriorityRouter{low: low, normal: normal, high: high}
}

// Notify dispatches msg to the notifier that matches the "priority" metadata
// value. Falls back to the normal notifier when the key is absent or unknown.
func (r *PriorityRouter) Notify(ctx context.Context, msg Message) error {
	var target Notifier

	switch msg.Meta["priority"] {
	case PriorityLow.String():
		target = r.low
	case PriorityHigh.String():
		target = r.high
	default:
		target = r.normal
	}

	if target == nil {
		return nil
	}
	return target.Notify(ctx, msg)
}

// WithPriority returns a copy of msg with the "priority" metadata key set.
func WithPriority(msg Message, p Priority) Message {
	if msg.Meta == nil {
		msg.Meta = make(map[string]string)
	}
	msg.Meta["priority"] = p.String()
	return msg
}
