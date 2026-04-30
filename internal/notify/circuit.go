package notify

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the circuit breaker is open and calls are rejected.
var ErrCircuitOpen = errors.New("circuit breaker is open")

type circuitState int

const (
	stateClosed circuitState = iota
	stateOpen
	stateHalfOpen
)

// CircuitBreaker wraps a Notifier and stops forwarding notifications when
// the underlying notifier fails repeatedly, allowing it time to recover.
type CircuitBreaker struct {
	mu           sync.Mutex
	inner        Notifier
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	state        circuitState
	openedAt     time.Time
}

// NewCircuitBreaker returns a Notifier that opens after maxFailures consecutive
// errors and retries after resetTimeout.
func NewCircuitBreaker(inner Notifier, maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 3
	}
	if resetTimeout <= 0 {
		resetTimeout = 30 * time.Second
	}
	return &CircuitBreaker{
		inner:        inner,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Notify forwards the message to the inner notifier unless the circuit is open.
func (cb *CircuitBreaker) Notify(ctx context.Context, msg Message) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateOpen:
		if time.Since(cb.openedAt) >= cb.resetTimeout {
			cb.state = stateHalfOpen
		} else {
			return fmt.Errorf("%w: retry after %s", ErrCircuitOpen,
				cb.resetTimeout-time.Since(cb.openedAt).Round(time.Second))
		}
	case stateClosed, stateHalfOpen:
		// proceed
	}

	err := cb.inner.Notify(ctx, msg)
	if err != nil {
		cb.failures++
		if cb.failures >= cb.maxFailures || cb.state == stateHalfOpen {
			cb.state = stateOpen
			cb.openedAt = time.Now()
		}
		return err
	}

	// success — reset
	cb.failures = 0
	cb.state = stateClosed
	return nil
}

// Reset forces the circuit back to closed state, clearing failure counts.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
}

// IsOpen reports whether the circuit is currently open.
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state == stateOpen
}
