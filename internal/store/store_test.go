package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "store.json")
}

func TestNew_CreatesEmpty(t *testing.T) {
	s, err := New(tempPath(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := s.All(); len(got) != 0 {
		t.Fatalf("expected empty store, got %d records", len(got))
	}
}

func TestSetAndGet(t *testing.T) {
	s, _ := New(tempPath(t))
	r := JobRecord{Name: "backup", LastRun: time.Now().UTC(), LastExit: 0}
	if err := s.Set(r); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected record to exist")
	}
	if got.Name != "backup" {
		t.Errorf("name mismatch: %s", got.Name)
	}
}

func TestGet_Missing(t *testing.T) {
	s, _ := New(tempPath(t))
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected missing record")
	}
}

func TestPersistence(t *testing.T) {
	p := tempPath(t)
	s1, _ := New(p)
	_ = s1.Set(JobRecord{Name: "job1", LastExit: 1})

	s2, err := New(p)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	r, ok := s2.Get("job1")
	if !ok {
		t.Fatal("record not persisted")
	}
	if r.LastExit != 1 {
		t.Errorf("LastExit: want 1 got %d", r.LastExit)
	}
}

func TestNew_CorruptFile(t *testing.T) {
	p := tempPath(t)
	_ = os.WriteFile(p, []byte("not json"), 0o644)
	_, err := New(p)
	if err == nil {
		t.Fatal("expected error on corrupt file")
	}
}

func TestAll(t *testing.T) {
	s, _ := New(tempPath(t))
	_ = s.Set(JobRecord{Name: "a"})
	_ = s.Set(JobRecord{Name: "b"})
	if len(s.All()) != 2 {
		t.Errorf("expected 2 records")
	}
}
