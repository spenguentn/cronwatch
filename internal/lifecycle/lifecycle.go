// Package lifecycle manages graceful startup and shutdown of cronwatch services.
// It coordinates signal handling, service registration, and ordered teardown
// so that all components (watcher, metrics server, health server) stop cleanly.
package lifecycle

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Service is any component that can be started and stopped.
type Service interface {
	// Start begins the service. It should block until the service is ready.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the service within the given timeout.
	Stop(ctx context.Context) error
	// Name returns a human-readable identifier used in logs.
	Name() string
}

// Manager coordinates the startup and shutdown of registered services.
type Manager struct {
	services []Service
	logger   *log.Logger
	timeout  time.Duration
	mu       sync.Mutex
}

// New creates a Manager with the given shutdown timeout.
// If logger is nil, output goes to stderr.
func New(timeout time.Duration, logger *log.Logger) *Manager {
	if logger == nil {
		logger = log.New(os.Stderr, "[lifecycle] ", log.LstdFlags)
	}
	return &Manager{
		timeout: timeout,
		logger:  logger,
	}
}

// Register adds a service to the manager. Services are started in registration
// order and stopped in reverse order.
func (m *Manager) Register(svc Service) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services = append(m.services, svc)
}

// Run starts all registered services and blocks until an OS signal (SIGINT or
// SIGTERM) is received or the parent context is cancelled. It then shuts down
// services in reverse order.
func (m *Manager) Run(ctx context.Context) error {
	m.mu.Lock()
	svcs := make([]Service, len(m.services))
	copy(svcs, m.services)
	m.mu.Unlock()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start each service.
	for _, svc := range svcs {
		m.logger.Printf("starting %s", svc.Name())
		if err := svc.Start(runCtx); err != nil {
			// Stop already-started services before returning.
			m.stopAll(svcs, 0)
			return fmt.Errorf("start %s: %w", svc.Name(), err)
		}
	}

	// Wait for termination signal or context cancellation.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case sig := <-sigCh:
		m.logger.Printf("received signal %s, shutting down", sig)
	case <-ctx.Done():
		m.logger.Printf("context cancelled, shutting down")
	}

	cancel()
	m.stopAll(svcs, len(svcs)-1)
	return nil
}

// stopAll stops services from index i down to 0 (reverse order).
func (m *Manager) stopAll(svcs []Service, from int) {
	for i := from; i >= 0; i-- {
		svc := svcs[i]
		m.logger.Printf("stopping %s", svc.Name())
		ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
		if err := svc.Stop(ctx); err != nil {
			m.logger.Printf("error stopping %s: %v", svc.Name(), err)
		}
		cancel()
	}
}
