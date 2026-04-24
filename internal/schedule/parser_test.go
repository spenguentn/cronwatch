package schedule

import (
	"testing"
	"time"
)

var fixedTime = time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

func TestNextRun(t *testing.T) {
	next, err := NextRun("*/5 * * * *", fixedTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2024, 1, 15, 12, 5, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestNextRun_InvalidExpr(t *testing.T) {
	_, err := NextRun("not-a-cron", fixedTime)
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestPrevRun(t *testing.T) {
	// At 12:00, the previous 5-min tick was 11:55
	prev, err := PrevRun("*/5 * * * *", fixedTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := time.Date(2024, 1, 15, 11, 55, 0, 0, time.UTC)
	if !prev.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, prev)
	}
}

func TestPrevRun_InvalidExpr(t *testing.T) {
	_, err := PrevRun("bad expr", fixedTime)
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		expr  string
		wantErr bool
	}{
		{"0 * * * *", false},
		{"*/15 9-17 * * 1-5", false},
		{"60 * * * *", true},
		{"not valid", true},
	}
	for _, tt := range tests {
		err := Validate(tt.expr)
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate(%q): wantErr=%v, got %v", tt.expr, tt.wantErr, err)
		}
	}
}
