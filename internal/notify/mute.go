package notify

import (
	"context"
	"sync"
	"time"
)

// MuteNotifier suppresses all notifications during a configured mute window.
// Calls to Notify during an active mute period are silently dropped.
// The mute can be set programmatically or scheduled for a fixed duration.
type MuteNotifier struct {
	mu      sync.RWMutex
	inner   Notifier
	muteEnd time.Time
	nowFn   func() time.Time
}

// NewMuteNotifier wraps inner with a mute window capability.
func NewMuteNotifier(inner Notifier) *MuteNotifier {
	return &MuteNotifier{
		inner: inner,
		nowFn: time.Now,
	}
}

// Mute suppresses notifications for the given duration.
func (m *MuteNotifier) Mute(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.muteEnd = m.nowFn().Add(d)
}

// Unmute cancels any active mute immediately.
func (m *MuteNotifier) Unmute() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.muteEnd = time.Time{}
}

// IsMuted reports whether notifications are currently suppressed.
func (m *MuteNotifier) IsMuted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nowFn().Before(m.muteEnd)
}

// Notify forwards msg to the inner notifier unless a mute is active.
func (m *MuteNotifier) Notify(ctx context.Context, msg Message) error {
	if m.inner == nil {
		return nil
	}
	if m.IsMuted() {
		return nil
	}
	return m.inner.Notify(ctx, msg)
}
