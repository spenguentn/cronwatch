package notify

import (
	"context"
	"fmt"
	"sync"
)

// WatermarkNotifier suppresses notifications whose severity is below a
// configurable high-water mark. The mark can be raised or lowered at runtime.
// Notifications at or above the mark are forwarded to the inner Notifier.
type WatermarkNotifier struct {
	mu    sync.RWMutex
	inner Notifier
	mark  Severity
}

// NewWatermarkNotifier returns a WatermarkNotifier that forwards messages
// whose severity is >= mark to inner.
func NewWatermarkNotifier(inner Notifier, mark Severity) *WatermarkNotifier {
	return &WatermarkNotifier{inner: inner, mark: mark}
}

// SetMark updates the high-water mark. Messages below the new mark will be
// suppressed on subsequent calls to Notify.
func (w *WatermarkNotifier) SetMark(mark Severity) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.mark = mark
}

// Mark returns the current high-water mark.
func (w *WatermarkNotifier) Mark() Severity {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.mark
}

// Notify forwards msg to the inner Notifier only when msg.Severity >= mark.
// Messages below the mark are silently dropped (nil error returned).
func (w *WatermarkNotifier) Notify(ctx context.Context, msg Message) error {
	if w.inner == nil {
		return nil
	}
	w.mu.RLock()
	mark := w.mark
	w.mu.RUnlock()

	if msg.Severity < mark {
		return nil
	}
	if err := w.inner.Notify(ctx, msg); err != nil {
		return fmt.Errorf("watermark notifier: %w", err)
	}
	return nil
}
