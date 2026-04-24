package config

import (
	"os"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cronwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTempConfig(t, `
check_interval: 2m
log_level: debug
alert:
  email: ops@example.com
jobs:
  - name: backup
    schedule: "0 2 * * *"
    drift_tolerance: 10m
    timeout: 30m
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 2*time.Minute {
		t.Errorf("check_interval: got %v, want 2m", cfg.CheckInterval)
	}
	if len(cfg.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(cfg.Jobs))
	}
	if cfg.Jobs[0].Name != "backup" {
		t.Errorf("job name: got %q, want %q", cfg.Jobs[0].Name, "backup")
	}
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: cleanup
    schedule: "*/5 * * * *"
`)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 1*time.Minute {
		t.Errorf("default check_interval: got %v, want 1m", cfg.CheckInterval)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("default log_level: got %q, want %q", cfg.LogLevel, "info")
	}
	if cfg.Jobs[0].DriftTolerance != 5*time.Minute {
		t.Errorf("default drift_tolerance: got %v, want 5m", cfg.Jobs[0].DriftTolerance)
	}
}

func TestLoad_NoJobs(t *testing.T) {
	path := writeTempConfig(t, `log_level: info\n`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for config with no jobs, got nil")
	}
}

func TestLoad_MissingSchedule(t *testing.T) {
	path := writeTempConfig(t, `
jobs:
  - name: orphan
`)
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for job missing schedule, got nil")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
