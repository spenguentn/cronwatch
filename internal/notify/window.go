package notify

import (
	"context"
	"sync"
	"time"
)

// WindowNotifier forwards messages only when the current time falls within
// one of the configured daily time windows (e.g. business hours).
// Messages outside all windows are silently dropped.
type WindowNotifier struct {
	inner   Notifier
	windows []TimeWindow
	mu      sync.RWMutex
	now     func() time.Time
}

// TimeWindow represents a half-open daily interval [Start, End).
type TimeWindow struct {
	Start time.Duration // offset from midnight
	End   time.Duration // offset from midnight
}

// NewWindowNotifier creates a WindowNotifier that forwards to inner only
// during the provided windows. If no windows are provided all messages pass.
func NewWindowNotifier(inner Notifier, windows []TimeWindow) *WindowNotifier {
	return &WindowNotifier{
		inner:   inner,
		windows: windows,
		now:     time.Now,
	}
}

// Notify sends the message if the current time is inside any configured window.
func (w *WindowNotifier) Notify(ctx context.Context, msg Message) error {
	if w.inner == nil {
		return nil
	}

	w.mu.RLock()
	windows := w.windows
	w.mu.RUnlock()

	if len(windows) == 0 || w.inWindow(windows) {
		return w.inner.Notify(ctx, msg)
	}
	return nil
}

// SetWindows replaces the active windows at runtime.
func (w *WindowNotifier) SetWindows(windows []TimeWindow) {
	w.mu.Lock()
	w.windows = windows
	w.mu.Unlock()
}

func (w *WindowNotifier) inWindow(windows []TimeWindow) bool {
	now := w.now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	offset := now.Sub(midnight)
	for _, win := range windows {
		if offset >= win.Start && offset < win.End {
			return true
		}
	}
	return false
}
