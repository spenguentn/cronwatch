// Package metrics provides a thread-safe collector for runtime statistics
// gathered while cronwatch monitors scheduled jobs.
//
// Usage:
//
//	col := metrics.NewCollector()
//
//	// During each watcher tick:
//	col.RecordCheck(jobName)
//
//	// When drift is detected:
//	col.RecordDrift(jobName, driftDuration)
//
//	// When a run is missed entirely:
//	col.RecordMiss(jobName)
//
//	// Inspect current state:
//	stats := col.Snapshot()
//	for _, s := range stats {
//		fmt.Printf("%s: checks=%d missed=%d drifts=%d\n",
//			s.Name, s.TotalChecks, s.MissedRuns, s.DriftEvents)
//	}
//
// All methods are safe for concurrent use.
package metrics
