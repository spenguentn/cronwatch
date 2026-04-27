package notify

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// EmailConfig holds SMTP configuration for sending alert emails.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
	Timeout  time.Duration
}

// emailNotifier sends alert notifications via SMTP email.
type emailNotifier struct {
	cfg EmailConfig
}

// NewEmailNotifier creates a Notifier that sends alerts via email.
// To must contain at least one recipient.
func NewEmailNotifier(cfg EmailConfig) (Notifier, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("email notifier: host is required")
	}
	if len(cfg.To) == 0 {
		return nil, fmt.Errorf("email notifier: at least one recipient required")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &emailNotifier{cfg: cfg}, nil
}

// Notify sends an email alert with the given level and message.
func (e *emailNotifier) Notify(ctx context.Context, level, message string) error {
	addr := fmt.Sprintf("%s:%d", e.cfg.Host, e.cfg.Port)

	subject := fmt.Sprintf("[cronwatch] %s alert", strings.ToUpper(level))
	body := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\n\r\n%s",
		strings.Join(e.cfg.To, ", "),
		e.cfg.From,
		subject,
		message,
	)

	var auth smtp.Auth
	if e.cfg.Username != "" {
		auth = smtp.PlainAuth("", e.cfg.Username, e.cfg.Password, e.cfg.Host)
	}

	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		err := smtp.SendMail(addr, auth, e.cfg.From, e.cfg.To, []byte(body))
		ch <- result{err: err}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email notifier: context cancelled: %w", ctx.Err())
	case res := <-ch:
		if res.err != nil {
			return fmt.Errorf("email notifier: send failed: %w", res.err)
		}
	}
	return nil
}
