package healthcheck_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/healthcheck"
)

func TestNew_IsHealthy(t *testing.T) {
	c := healthcheck.New()
	s := c.Status()
	if !s.OK {
		t.Fatal("expected new checker to be healthy")
	}
	if s.StartedAt.IsZero() {
		t.Fatal("expected non-zero StartedAt")
	}
}

func TestSetHealthy_Unhealthy(t *testing.T) {
	c := healthcheck.New()
	c.SetHealthy(false, "watcher stalled")
	s := c.Status()
	if s.OK {
		t.Fatal("expected unhealthy status")
	}
	if s.Message != "watcher stalled" {
		t.Fatalf("unexpected message: %q", s.Message)
	}
	if s.CheckedAt.IsZero() {
		t.Fatal("expected CheckedAt to be set")
	}
}

func TestSetHealthy_Recovery(t *testing.T) {
	c := healthcheck.New()
	c.SetHealthy(false, "down")
	c.SetHealthy(true, "")
	if !c.Status().OK {
		t.Fatal("expected recovery to healthy")
	}
}

func TestHandler_HealthyReturns200(t *testing.T) {
	c := healthcheck.New()
	rr := httptest.NewRecorder()
	c.Handler()(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var s healthcheck.Status
	if err := json.NewDecoder(rr.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !s.OK {
		t.Fatal("expected ok=true in response")
	}
}

func TestHandler_UnhealthyReturns503(t *testing.T) {
	c := healthcheck.New()
	c.SetHealthy(false, "no checks run")
	rr := httptest.NewRecorder()
	c.Handler()(rr, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rr.Code)
	}
}

func TestStatus_CheckedAtUpdates(t *testing.T) {
	c := healthcheck.New()
	before := time.Now()
	c.SetHealthy(true, "all good")
	after := time.Now()
	s := c.Status()
	if s.CheckedAt.Before(before) || s.CheckedAt.After(after) {
		t.Fatalf("CheckedAt %v not in expected range", s.CheckedAt)
	}
}
