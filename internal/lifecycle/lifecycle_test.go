package lifecycle_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/lifecycle"
)

// stubService is a minimal Service implementation for testing.
type stubService struct {
	startErr error
	stopErr  error
	started atomic.Bool
	stopped atomic.Bool
}

func (s *stubService) Start(_ context.Context) error {
	if s.startErr != nil {
		return s.startErr
	}
	s.started.Store(true)
	return nil
}

func (s *stubService) Stop(_ context.Context) error {
	if s.stopErr != nil {
		return s.stopErr
	}
	s.stopped.Store(true)
	return nil
}

func TestNew_RegisterAndStart(t *testing.T) {
	svc := &stubService{}
	lc := lifecycle.New()
	lc.Register(svc)

	ctx := context.Background()
	if err := lc.Start(ctx); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}
	defer lc.Stop(ctx) //nolint:errcheck

	if !svc.started.Load() {
		t.Error("expected service to be started")
	}
}

func TestStart_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	svc := &stubService{startErr: want}
	lc := lifecycle.New()
	lc.Register(svc)

	err := lc.Start(context.Background())
	if !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}

func TestStop_CallsAllServices(t *testing.T) {
	a, b := &stubService{}, &stubService{}
	lc := lifecycle.New()
	lc.Register(a)
	lc.Register(b)

	ctx := context.Background()
	_ = lc.Start(ctx)
	if err := lc.Stop(ctx); err != nil {
		t.Fatalf("unexpected stop error: %v", err)
	}

	if !a.stopped.Load() || !b.stopped.Load() {
		t.Error("expected both services to be stopped")
	}
}

func TestWait_ReturnsAfterStop(t *testing.T) {
	lc := lifecycle.New()
	ctx := context.Background()
	_ = lc.Start(ctx)

	done := make(chan struct{})
	go func() {
		lc.Wait()
		close(done)
	}()

	_ = lc.Stop(ctx)

	select {
	case <-done:
		// ok
	case <-time.After(2 * time.Second):
		t.Fatal("Wait did not return after Stop")
	}
}
