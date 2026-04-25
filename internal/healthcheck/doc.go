// Package healthcheck provides a lightweight liveness and readiness probe
// for the cronwatch daemon.
//
// Usage:
//
//	checker := healthcheck.New()
//	srv := healthcheck.NewServer(":9091", checker)
//	addr, err := srv.Start()
//
// The Checker can be updated from the watcher loop to reflect real-time
// daemon state:
//
//	checker.SetHealthy(false, "no cron checks completed in 5 minutes")
//
// The HTTP server exposes a single endpoint:
//
//	GET /health  → 200 {"ok":true,...} or 503 {"ok":false,"message":"..."}
//
// This allows external tools (systemd, Kubernetes, load balancers) to
// determine whether cronwatch is operating normally.
package healthcheck
