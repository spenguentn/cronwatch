package alert_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/alert"
)

func TestLogNotifier_Notify(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewLogNotifier(&buf)

	a := alert.Alert{
		JobName:   "backup",
		Level:     alert.LevelWarn,
		Message:   "drift exceeded threshold",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	if err := n.Notify(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN in output, got: %s", out)
	}
	if !strings.Contains(out, "backup") {
		t.Errorf("expected job name in output, got: %s", out)
	}
	if !strings.Contains(out, "drift exceeded threshold") {
		t.Errorf("expected message in output, got: %s", out)
	}
}

func TestLogNotifier_NilWriter(t *testing.T) {
	// Should not panic when w is nil (falls back to stderr).
	n := alert.NewLogNotifier(nil)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
}

func TestDispatcher_Warn(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewLogNotifier(&buf)
	d := alert.NewDispatcher(n)

	if err := d.Warn("sync", "minor drift detected"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN level, got: %s", out)
	}
}

func TestDispatcher_Error(t *testing.T) {
	var buf bytes.Buffer
	n := alert.NewLogNotifier(&buf)
	d := alert.NewDispatcher(n)

	if err := d.Error("cleanup", "missed run"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR level, got: %s", out)
	}
}

// failNotifier always returns an error.
type failNotifier struct{}

func (f *failNotifier) Notify(_ alert.Alert) error {
	return errors.New("notify failed")
}

func TestDispatcher_PropagatesError(t *testing.T) {
	d := alert.NewDispatcher(&failNotifier{})
	if err := d.Warn("job", "msg"); err == nil {
		t.Fatal("expected error from failing notifier")
	}
}
