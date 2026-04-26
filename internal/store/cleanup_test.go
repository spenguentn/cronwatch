package store

import (
	"testing"
	"time"
)

func TestPrune_RemovesStaleEntries(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	now := time.Now()
	old := now.Add(-48 * time.Hour)
	recent := now.Add(-1 * time.Hour)

	s.mu.Lock()
	s.data["stale-job"] = old
	s.data["fresh-job"] = recent
	s.mu.Unlock()

	cleaner := NewCleaner(s, 24*time.Hour)
	removed := cleaner.Prune(now)

	if len(removed) != 1 || removed[0] != "stale-job" {
		t.Errorf("expected [stale-job] removed, got %v", removed)
	}

	if _, err := s.Get("fresh-job"); err != nil {
		t.Errorf("fresh-job should still exist: %v", err)
	}

	if _, err := s.Get("stale-job"); err == nil {
		t.Error("stale-job should have been removed")
	}
}

func TestPrune_NothingToRemove(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	now := time.Now()
	s.mu.Lock()
	s.data["fresh-job"] = now.Add(-1 * time.Hour)
	s.mu.Unlock()

	cleaner := NewCleaner(s, 24*time.Hour)
	removed := cleaner.Prune(now)

	if len(removed) != 0 {
		t.Errorf("expected no removals, got %v", removed)
	}
}

func TestPrune_EmptyStore(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	cleaner := NewCleaner(s, 24*time.Hour)
	removed := cleaner.Prune(time.Now())

	if len(removed) != 0 {
		t.Errorf("expected no removals on empty store, got %v", removed)
	}
}

func TestPrune_PersistsAfterRemoval(t *testing.T) {
	path := tempPath(t)
	s, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	now := time.Now()
	s.mu.Lock()
	s.data["old"] = now.Add(-72 * time.Hour)
	s.mu.Unlock()

	cleaner := NewCleaner(s, 24*time.Hour)
	cleaner.Prune(now)

	s2, err := New(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if _, err := s2.Get("old"); err == nil {
		t.Error("expected old entry to be absent after reload")
	}
}
