package notify

import (
	"context"
	"fmt"
	"log"
	"io"
)

// ShadowNotifier forwards every message to a primary notifier and, in the
// background, also sends it to a shadow notifier. Errors from the shadow are
// logged but never returned to the caller, making it safe to use for dark
// launches or canary comparisons.
type ShadowNotifier struct {
	primary Notifier
	shadow  Notifier
	logger  *log.Logger
}

// NewShadowNotifier creates a ShadowNotifier. If logger is nil a discard
// logger is used.
func NewShadowNotifier(primary, shadow Notifier, logger *log.Logger) *ShadowNotifier {
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}
	return &ShadowNotifier{
		primary: primary,
		shadow:  shadow,
		logger:  logger,
	}
}

// Notify sends msg to the primary notifier and asynchronously to the shadow.
// Only the primary's error is returned.
func (s *ShadowNotifier) Notify(ctx context.Context, msg Message) error {
	if s.primary == nil {
		return nil
	}

	// Fire shadow in a goroutine; use a detached context so a cancelled caller
	// context does not abort the shadow delivery.
	if s.shadow != nil {
		go func() {
			if err := s.shadow.Notify(context.Background(), msg); err != nil {
				s.logger.Printf("shadow notifier error: %v", err)
			}
		}()
	}

	if err := s.primary.Notify(ctx, msg); err != nil {
		return fmt.Errorf("shadow primary: %w", err)
	}
	return nil
}
