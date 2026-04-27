// Package ratelimit provides a token-bucket rate limiter for alert notifications,
// preventing alert storms when many cron jobs drift or miss simultaneously.
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter controls how frequently notifications may be sent per job.
type Limiter struct {
	mu       sync.Mutex
	buckets  map[string]bucket
	rate     time.Duration // minimum interval between alerts for a given key
	burst    int          // max alerts allowed before rate limiting kicks in
}

type bucket struct {
	tokens    int
	lastRefil time.Time
}

// New creates a Limiter that allows at most burst notifications per key,
// refilling one token every rate duration.
func New(rate time.Duration, burst int) *Limiter {
	if burst < 1 {
		burst = 1
	}
	return &Limiter{
		buckets: make(map[string]bucket),
		rate:    rate,
		burst:   burst,
	}
}

// Allow reports whether a notification for key is permitted at time now.
// It consumes one token if allowed.
func (l *Limiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		b = bucket{tokens: l.burst, lastRefil: now}
	}

	// Refill tokens based on elapsed time.
	if l.rate > 0 {
		refills := int(now.Sub(b.lastRefil) / l.rate)
		if refills > 0 {
			b.tokens += refills
			if b.tokens > l.burst {
				b.tokens = l.burst
			}
			b.lastRefil = b.lastRefil.Add(time.Duration(refills) * l.rate)
		}
	}

	if b.tokens <= 0 {
		l.buckets[key] = b
		return false
	}

	b.tokens--
	l.buckets[key] = b
	return true
}

// Reset clears the bucket for key, restoring full burst capacity.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.buckets, key)
}

// Notifier wraps another notify.Notifier and applies rate limiting.
type Notifier struct {
	inner   Notify
	limiter *Limiter
	key     string
}

// Notify is the minimal interface expected from a downstream notifier.
type Notify interface {
	Notify(ctx context.Context, subject, body string) error
}

// NewNotifier wraps inner with rate limiting keyed by key.
func NewNotifier(inner Notify, limiter *Limiter, key string) *Notifier {
	return &Notifier{inner: inner, limiter: limiter, key: key}
}

// Notify sends the notification only if the rate limiter permits it.
// Returns nil (silently dropped) when rate-limited.
func (n *Notifier) Notify(ctx context.Context, subject, body string) error {
	if !n.limiter.Allow(n.key, time.Now()) {
		return nil
	}
	return n.inner.Notify(ctx, subject, body)
}
