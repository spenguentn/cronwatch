package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// NextRun returns the next scheduled run time after the given time.
func NextRun(cronExpr string, after time.Time) (time.Time, error) {
	sched, err := parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}
	return sched.Next(after), nil
}

// PrevRun returns the most recent scheduled run time before the given time.
// It approximates by stepping backwards using the schedule interval.
func PrevRun(cronExpr string, before time.Time) (time.Time, error) {
	sched, err := parse(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	// Walk backwards: find next after (before - 2*interval) and verify
	// Use a lookback window of 1 year, stepping via Next
	candidate := before.Add(-365 * 24 * time.Hour)
	var prev time.Time
	for {
		n := sched.Next(candidate)
		if n.IsZero() || n.After(before) || n.Equal(before) {
			break
		}
		prev = n
		candidate = n
	}
	if prev.IsZero() {
		return time.Time{}, fmt.Errorf("no previous run found within lookback window")
	}
	return prev, nil
}

// Validate checks whether a cron expression is parseable.
func Validate(cronExpr string) error {
	_, err := parse(cronExpr)
	return err
}

func parse(cronExpr string) (cron.Schedule, error) {
	p := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)
	sched, err := p.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression %q: %w", cronExpr, err)
	}
	return sched, nil
}
