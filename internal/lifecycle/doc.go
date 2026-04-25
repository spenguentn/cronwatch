// Package lifecycle manages the startup and graceful shutdown
// of all long-running cronwatch subsystems.
//
// A Lifecycle coordinates multiple services — watcher, metrics server,
// health-check server — ensuring they start in dependency order and
// shut down cleanly when an OS signal is received or the context is
// cancelled.
//
// Basic usage:
//
//	lc := lifecycle.New(cfg, logger)
//	if err := lc.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	lc.Wait() // blocks until shutdown
//
All registered services must implement the Service interface:
//
//	type Service interface {
//		Start(ctx context.Context) error
//		Stop(ctx context.Context) error
//	}
package lifecycle
