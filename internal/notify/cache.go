package notify

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// CacheEntry holds a cached notification result.
type CacheEntry struct {
	NotifiedAt time.Time
	Err        error
}

// CacheNotifier skips delivery when an identical message was successfully
// delivered within the configured TTL, avoiding redundant notifications.
type CacheNotifier struct {
	inner Notifier
	ttl   time.Duration
	now   func() time.Time

	mu    sync.Mutex
	cache map[string]CacheEntry
}

// NewCacheNotifier wraps inner, suppressing re-delivery of messages whose
// content hash has been successfully sent within ttl. A zero or negative ttl
// defaults to 5 minutes.
func NewCacheNotifier(inner Notifier, ttl time.Duration) *CacheNotifier {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &CacheNotifier{
		inner: inner,
		ttl:   ttl,
		now:   time.Now,
		cache: make(map[string]CacheEntry),
	}
}

// Notify delivers msg via inner only when no successful delivery for the same
// content hash exists within the TTL window.
func (c *CacheNotifier) Notify(ctx context.Context, msg Message) error {
	if c.inner == nil {
		return nil
	}

	key := c.hashMessage(msg)
	now := c.now()

	c.mu.Lock()
	entry, found := c.cache[key]
	c.mu.Unlock()

	if found && entry.Err == nil && now.Sub(entry.NotifiedAt) < c.ttl {
		return nil
	}

	err := c.inner.Notify(ctx, msg)

	c.mu.Lock()
	c.cache[key] = CacheEntry{NotifiedAt: now, Err: err}
	c.mu.Unlock()

	return err
}

// Invalidate removes the cache entry for the given message, forcing the next
// call to Notify to deliver regardless of TTL.
func (c *CacheNotifier) Invalidate(msg Message) {
	key := c.hashMessage(msg)
	c.mu.Lock()
	delete(c.cache, key)
	c.mu.Unlock()
}

// Flush clears all cached entries.
func (c *CacheNotifier) Flush() {
	c.mu.Lock()
	c.cache = make(map[string]CacheEntry)
	c.mu.Unlock()
}

func (c *CacheNotifier) hashMessage(msg Message) string {
	h := sha256.Sum256([]byte(msg.Subject + "\x00" + msg.Body))
	return fmt.Sprintf("%x", h)
}
