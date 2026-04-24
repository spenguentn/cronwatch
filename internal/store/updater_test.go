package store

import (
	"testing"
	"time"
)

func newUpdater(t *testing.T) *Updater {
	t.Helper()
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return NewUpdater(s)
}

func TestRecord_Success(t *testing.T) {
	u := newUpdater(t)
	now := time.Now().UTC()
	err := u.Record(RunResult{JobName: "myjob", RunAt: now, ExitCode: 0})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	last, ok := u.LastRun("myjob")
	if !ok {
		t.Fatal("expected record")
	}
	if !last.Equal(now) {
		t.Errorf("time mismatch: want %v got %v", now, last)
	}
}

func TestRecord_EmptyName(t *testing.T) {
	u := newUpdater(t)
	err := u.Record(RunResult{JobName: ""})
	if err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestRecord_ZeroTime(t *testing.T) {
	u := newUpdater(t)
	err := u.Record(RunResult{JobName: "zerojob"})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	last, ok := u.LastRun("zerojob")
	if !ok {
		t.Fatal("expected record")
	}
	if last.IsZero() {
		t.Error("expected non-zero time when RunAt was zero")
	}
}

func TestLastRun_Missing(t *testing.T) {
	u := newUpdater(t)
	_, ok := u.LastRun("ghost")
	if ok {
		t.Fatal("expected missing")
	}
}

func TestRecord_NonZeroExit(t *testing.T) {
	u := newUpdater(t)
	_ = u.Record(RunResult{JobName: "failing", ExitCode: 2})
	r, _ := u.store.Get("failing")
	if r.LastExit != 2 {
		t.Errorf("LastExit: want 2 got %d", r.LastExit)
	}
}
