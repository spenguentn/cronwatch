package schedule

import (
	"fmt"
	"time"
)

// DriftResult holds the result of a drift check for a single job.
type DriftResult struct {
	JobName      string
	Expected     time.Time
	Actual       time.Time
	Drift        time.Duration
	Missed       bool
	DriftExceeded bool
}

// CheckDrift compares the actual last-run time against the most recently
// expected run time derived from the cron expression.
// maxDrift is the threshold beyond which drift is flagged.
func CheckDrift(jobName, cronExpr string, lastRun time.Time, now time.Time, maxDrift time.Duration) (DriftResult, error) {
	if err := Validate(cronExpr); err != nil {
		return DriftResult{}, fmt.Errorf("job %q: %w", jobName, err)
	}

	expected, err := PrevRun(cronExpr, now)
	if err != nil {
		return DriftResult{}, fmt.Errorf("job %q: could not determine previous run: %w", jobName, err)
	}

	result := DriftResult{
		JobName:  jobName,
		Expected: expected,
		Actual:   lastRun,
	}

	if lastRun.IsZero() || lastRun.Before(expected) {
		result.Missed = true
		return result, nil
	}

	result.Drift = lastRun.Sub(expected)
	if result.Drift < 0 {
		result.Drift = -result.Drift
	}
	if maxDrift > 0 && result.Drift > maxDrift {
		result.DriftExceeded = true
	}
	return result, nil
}
