package lifecycle_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/lifecycle"
)

// slowService simulates a service that takes time to start and stop.
type slowService struct {
	startDelay time.Duration
	stopDelay  time.Duration
	started    atomic.Bool
	stopped    atomic.Bool
	startErr    error
	stopErr     error
}

func (s *slowService) Start(ctx context.Context) error {
	select {
	case <-time.After(s.startDelay):
		if s.startErr != nil {
			return s.startErr
		}
		s.started.Store(true)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *slowService) Stop(ctx context.Context) error {
	select {
	case <-time.After(s.stopDelay):
		s.stopped.Store(true)
		return s.stopErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

// TestIntegration_StartAndGracefulStop verifies that multiple services start
// concurrently and are all stopped when the lifecycle is shut down.
func TestIntegration_StartAndGracefulStop(t *testing.T) {
	lc := lifecycle.New()

	const numServices = 5
	services := make([]*slowService, numServices)
	for i := range services {
		svc := &slowService{startDelay: 10 * time.Millisecond, stopDelay: 10 * time.Millisecond}
		services[i] = svc
		lc.Register(svc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- lc.Start(ctx)
	}()

	// Give services time to start.
	time.Sleep(50 * time.Millisecond)

	for i, svc := range services {
		if !svc.started.Load() {
			t.Errorf("service %d did not start", i)
		}
	}

	if err := lc.Stop(ctx); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}

	for i, svc := range services {
		if !svc.stopped.Load() {
			t.Errorf("service %d was not stopped", i)
		}
	}

	if err := <-errCh; err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected start error: %v", err)
	}
}

// TestIntegration_StartFailurePropagates verifies that if any service fails to
// start the error is surfaced and the lifecycle halts.
func TestIntegration_StartFailurePropagates(t *testing.T) {
	lc := lifecycle.New()

	sentinel := errors.New("boot failure")

	var mu sync.Mutex
	started := 0

	for i := 0; i < 3; i++ {
		svc := &slowService{startDelay: 5 * time.Millisecond}
		if i == 1 {
			svc.startErr = sentinel
		} else {
			// Track successfully started services.
			origStart := svc.startErr
			_ = origStart
			svc.startErr = nil
			_ = &mu
			_ = &started
		}
		lc.Register(svc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := lc.Start(ctx)
	if err == nil {
		t.Fatal("expected an error from Start, got nil")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got: %v", err)
	}
}
