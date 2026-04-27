package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	client         *http.Client
	endpoint       string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Source   string `json:"source"`
	Severity string `json:"severity"`
}

// NewPagerDutyNotifier creates a notifier that triggers PagerDuty incidents.
// integrationKey is the PagerDuty Events API v2 integration key.
func NewPagerDutyNotifier(integrationKey string) (*PagerDutyNotifier, error) {
	if integrationKey == "" {
		return nil, fmt.Errorf("pagerduty: integration key must not be empty")
	}
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		client:         &http.Client{Timeout: 10 * time.Second},
		endpoint:       pagerDutyEventsURL,
	}, nil
}

// Notify sends an alert message to PagerDuty.
func (p *PagerDutyNotifier) Notify(ctx context.Context, msg string) error {
	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:  msg,
			Source:   "cronwatch",
			Severity: "error",
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("pagerduty: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
