// Package notify provides webhook-based alerting for cronwatch.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Payload is the JSON body sent to a webhook endpoint.
type Payload struct {
	Level   string `json:"level"`
	Job     string `json:"job"`
	Message string `json:"message"`
	TS      int64  `json:"ts"`
}

// WebhookNotifier sends alert payloads to an HTTP endpoint.
type WebhookNotifier struct {
	URL    string
	client *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with a sensible timeout.
func NewWebhookNotifier(url string) *WebhookNotifier {
	return &WebhookNotifier{
		URL: url,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a JSON payload to the configured webhook URL.
func (w *WebhookNotifier) Notify(ctx context.Context, level, job, message string) error {
	p := Payload{
		Level:   level,
		Job:     job,
		Message: message,
		TS:      time.Now().Unix(),
	}
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.URL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("notify: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notify: unexpected status %d from webhook", resp.StatusCode)
	}
	return nil
}
