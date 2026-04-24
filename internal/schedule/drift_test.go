package schedule

import (
	"testing"
	"time"
)

func TestCheckDrift_NoDrift(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 6, 0, 0, time.UTC)
	// Last run was exactly at the expected 12:05 tick
	lastRun := time.Date(2024, 1, 15, 12, 5, 0, 0, time.UTC)

	res, err := CheckDrift("myjob", "*/5 * * * *", lastRun, now, 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Missed {
		t.Error("expected job not to be missed")
	}
	if res.DriftExceeded {
		t.Error("expected drift not to be exceeded")
	}
	if res.Drift != 0 {
		t.Errorf("expected zero drift, got %v", res.Drift)
	}
}

func TestCheckDrift_DriftExceeded(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 6, 0, 0, time.UTC)
	// Last run was 45 seconds late
	lastRun := time.Date(2024, 1, 15, 12, 5, 45, 0, time.UTC)

	res, err := CheckDrift("latejob", "*/5 * * * *", lastRun, now, 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Missed {
		t.Error("job should not be marked missed")
	}
	if !res.DriftExceeded {
		t.Error("expected DriftExceeded to be true")
	}
}

func TestCheckDrift_Missed(t *testing.T) {
	now := time.Date(2024, 1, 15, 12, 6, 0, 0, time.UTC)
	// Last run was before the expected 12:05 tick
	lastRun := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	res, err := CheckDrift("missedjob", "*/5 * * * *", lastRun, now, 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Missed {
		t.Error("expected job to be marked as missed")
	}
}

func TestCheckDrift_InvalidExpr(t *testing.T) {
	_, err := CheckDrift("badjob", "not-valid", time.Now(), time.Now(), time.Minute)
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}
