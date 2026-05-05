package notify

import (
	"context"
	"math/rand"
	"sync"
)

// SamplingNotifier wraps a Notifier and forwards only a statistical sample
// of messages. The sample rate is expressed as a value in [0.0, 1.0] where
// 1.0 means forward every message and 0.0 means drop all messages.
type SamplingNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	rate     float64
	randFunc func() float64 // injectable for testing
}

// NewSamplingNotifier creates a SamplingNotifier that forwards messages to
// inner at the given sample rate. Rate must be in [0.0, 1.0]; values outside
// this range are clamped.
func NewSamplingNotifier(inner Notifier, rate float64) *SamplingNotifier {
	if rate < 0.0 {
		rate = 0.0
	}
	if rate > 1.0 {
		rate = 1.0
	}
	return &SamplingNotifier{
		inner:    inner,
		rate:     rate,
		randFunc: rand.Float64,
	}
}

// Notify forwards the message to the inner notifier with probability equal to
// the configured sample rate. Dropped messages return a nil error.
func (s *SamplingNotifier) Notify(ctx context.Context, msg Message) error {
	if s.inner == nil {
		return nil
	}

	s.mu.Lock()
	v := s.randFunc()
	s.mu.Unlock()

	if v >= s.rate {
		return nil
	}

	return s.inner.Notify(ctx, msg)
}

// SetRate updates the sample rate at runtime. The new rate is clamped to
// [0.0, 1.0].
func (s *SamplingNotifier) SetRate(rate float64) {
	if rate < 0.0 {
		rate = 0.0
	}
	if rate > 1.0 {
		rate = 1.0
	}
	s.mu.Lock()
	s.rate = rate
	s.mu.Unlock()
}
