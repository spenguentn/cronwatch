package watcher

import (
	"bytes"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/alert"
	"github.com/cronwatch/cronwatch/internal/store"
)

func makeChecker(t *testing.T) (*Checker, *store.Updater, *bytes.Buffer) {
	t.Helper()
	p := t.TempDir() + "/state.json"
	s, err := store.New(p)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	u := store.NewUpdater(s)
	var buf bytes.Buffer
	notifier := alert.NewLogNotifier(&buf)
	d := alert.NewDispatcher(notifier)
	return NewChecker(u, d), u, &buf
}

func TestCheckJob_NoLastRun(t *testing.T) {
	checker, _, buf := makeChecker(t)
	status := checker.CheckJob("backup", "0 * * * *", time.Minute*5)
	if status.Missed || status.Drifted || status.Err != nil {
		t.Errorf("unexpected status for missing run: %+v", status)
	}
	if buf.Len() == 0 {
		t.Error("expected a warning log for missing run, got none")
	}
}

func TestCheckJob_NoDrift(t *testing.T) {
	checker, u, buf := makeChecker(t)
	// Record a run close to the last expected tick.
	now := time.Now().Add(-30 * time.Second)
	if err := u.Record("sync", now); err != nil {
		t.Fatalf("Record: %v", err)
	}
	status := checker.CheckJob("sync", "* * * * *", time.Minute*2)
	if status.Err != nil {
		t.Fatalf("unexpected error: %v", status.Err)
	}
	if status.Missed || status.Drifted {
		t.Errorf("expected no drift/missed, got %+v", status)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no alerts, got: %s", buf.String())
	}
}

func TestCheckJob_InvalidExpr(t *testing.T) {
	checker, u, _ := makeChecker(t)
	if err := u.Record("bad", time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("Record: %v", err)
	}
	status := checker.CheckJob("bad", "not-a-cron", time.Minute)
	if status.Err == nil {
		t.Error("expected error for invalid expression, got nil")
	}
}
