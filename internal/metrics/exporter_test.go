package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExporter_EmptySnapshot(t *testing.T) {
	c := NewCollector()
	e := NewExporter(c)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	e.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var snap map[string]JobMetrics
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(snap) != 0 {
		t.Errorf("expected empty snapshot, got %d entries", len(snap))
	}
}

func TestExporter_WithData(t *testing.T) {
	c := NewCollector()
	c.RecordCheck("backup", time.Now())
	c.RecordMiss("backup")
	c.RecordDrift("backup", 90*time.Second)

	e := NewExporter(c)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	e.Handler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var snap map[string]JobMetrics
	if err := json.NewDecoder(rec.Body).Decode(&snap); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	m, ok := snap["backup"]
	if !ok {
		t.Fatal("expected 'backup' key in snapshot")
	}
	if m.Checks != 1 {
		t.Errorf("expected 1 check, got %d", m.Checks)
	}
	if m.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", m.Misses)
	}
	if m.Drifts != 1 {
		t.Errorf("expected 1 drift, got %d", m.Drifts)
	}
}

func TestExporter_RegisterRoutes(t *testing.T) {
	c := NewCollector()
	e := NewExporter(c)

	mux := http.NewServeMux()
	e.RegisterRoutes(mux, "/metrics")

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from registered route, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json content-type, got %q", ct)
	}
}
