package metrics_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/metrics"
)

func TestRecordCheck(t *testing.T) {
	c := metrics.NewCollector()
	c.RecordCheck("backup")
	c.RecordCheck("backup")

	s, ok := c.Get("backup")
	if !ok {
		t.Fatal("expected stats for backup")
	}
	if s.TotalChecks != 2 {
		t.Errorf("TotalChecks = %d, want 2", s.TotalChecks)
	}
	if s.LastChecked.IsZero() {
		t.Error("LastChecked should not be zero")
	}
}

func TestRecordMiss(t *testing.T) {
	c := metrics.NewCollector()
	c.RecordMiss("cleanup")
	c.RecordMiss("cleanup")
	c.RecordMiss("cleanup")

	s, ok := c.Get("cleanup")
	if !ok {
		t.Fatal("expected stats for cleanup")
	}
	if s.MissedRuns != 3 {
		t.Errorf("MissedRuns = %d, want 3", s.MissedRuns)
	}
}

func TestRecordDrift(t *testing.T) {
	c := metrics.NewCollector()
	d := 90 * time.Second
	c.RecordDrift("sync", d)

	s, ok := c.Get("sync")
	if !ok {
		t.Fatal("expected stats for sync")
	}
	if s.DriftEvents != 1 {
		t.Errorf("DriftEvents = %d, want 1", s.DriftEvents)
	}
	if s.LastDrift != d {
		t.Errorf("LastDrift = %v, want %v", s.LastDrift, d)
	}
}

func TestGet_Missing(t *testing.T) {
	c := metrics.NewCollector()
	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}

func TestSnapshot(t *testing.T) {
	c := metrics.NewCollector()
	c.RecordCheck("job-a")
	c.RecordCheck("job-b")
	c.RecordMiss("job-b")

	snap := c.Snapshot()
	if len(snap) != 2 {
		t.Errorf("Snapshot len = %d, want 2", len(snap))
	}
}

func TestMultipleJobs_Independent(t *testing.T) {
	c := metrics.NewCollector()
	c.RecordMiss("alpha")
	c.RecordCheck("beta")

	a, _ := c.Get("alpha")
	b, _ := c.Get("beta")

	if a.TotalChecks != 0 {
		t.Errorf("alpha TotalChecks should be 0, got %d", a.TotalChecks)
	}
	if b.MissedRuns != 0 {
		t.Errorf("beta MissedRuns should be 0, got %d", b.MissedRuns)
	}
}
