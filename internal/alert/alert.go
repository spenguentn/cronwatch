// Package alert provides alerting mechanisms for cron job drift and missed runs.
package alert

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert represents a single alert event.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	Timestamp time.Time
}

// Notifier sends alerts to a destination.
type Notifier interface {
	Notify(a Alert) error
}

// LogNotifier writes alerts to an io.Writer (default: stderr).
type LogNotifier struct {
	logger *log.Logger
}

// NewLogNotifier creates a LogNotifier writing to w.
// If w is nil, os.Stderr is used.
func NewLogNotifier(w io.Writer) *LogNotifier {
	if w == nil {
		w = os.Stderr
	}
	return &LogNotifier{
		logger: log.New(w, "", 0),
	}
}

// Notify writes the alert as a formatted log line.
func (n *LogNotifier) Notify(a Alert) error {
	n.logger.Println(format(a))
	return nil
}

func format(a Alert) string {
	return fmt.Sprintf("%s [%s] job=%q %s",
		a.Timestamp.UTC().Format(time.RFC3339),
		string(a.Level),
		a.JobName,
		a.Message,
	)
}

// Dispatcher holds a Notifier and dispatches alerts.
type Dispatcher struct {
	notifier Notifier
}

// NewDispatcher creates a Dispatcher backed by n.
func NewDispatcher(n Notifier) *Dispatcher {
	return &Dispatcher{notifier: n}
}

// Warn sends a warning-level alert.
func (d *Dispatcher) Warn(jobName, message string) error {
	return d.notifier.Notify(Alert{
		JobName:   jobName,
		Level:     LevelWarn,
		Message:   message,
		Timestamp: time.Now(),
	})
}

// Error sends an error-level alert.
func (d *Dispatcher) Error(jobName, message string) error {
	return d.notifier.Notify(Alert{
		JobName:   jobName,
		Level:     LevelError,
		Message:   message,
		Timestamp: time.Now(),
	})
}
