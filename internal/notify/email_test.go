package notify

import (
	"context"
	"io"
	"net"
	"testing"
	"time"
)

// startSMTPStub starts a minimal TCP server that accepts one connection,
// performs a bare SMTP handshake and returns the received data.
func startSMTPStub(t *testing.T) (addr string, received chan string) {
	t.Helper()
	received = make(chan string, 1)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		// Minimal SMTP: greet, then read everything
		_, _ = conn.Write([]byte("220 stub ESMTP\r\n"))
		buf, _ := io.ReadAll(conn)
		received <- string(buf)
	}()

	return ln.Addr().String(), received
}

func TestNewEmailNotifier_MissingHost(t *testing.T) {
	_, err := NewEmailNotifier(EmailConfig{To: []string{"a@b.com"}})
	if err == nil {
		t.Fatal("expected error for missing host")
	}
}

func TestNewEmailNotifier_NoRecipients(t *testing.T) {
	_, err := NewEmailNotifier(EmailConfig{Host: "localhost"})
	if err == nil {
		t.Fatal("expected error for empty recipients")
	}
}

func TestNewEmailNotifier_DefaultPort(t *testing.T) {
	n, err := NewEmailNotifier(EmailConfig{
		Host: "localhost",
		To:   []string{"ops@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	en := n.(*emailNotifier)
	if en.cfg.Port != 587 {
		t.Errorf("expected default port 587, got %d", en.cfg.Port)
	}
}

func TestEmailNotifier_CancelledContext(t *testing.T) {
	n, err := NewEmailNotifier(EmailConfig{
		Host:    "192.0.2.1", // TEST-NET, unreachable
		Port:    25,
		From:    "from@example.com",
		To:      []string{"to@example.com"},
		Timeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err = n.Notify(ctx, "warn", "test message")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestNewEmailNotifier_DefaultTimeout(t *testing.T) {
	n, err := NewEmailNotifier(EmailConfig{
		Host: "localhost",
		To:   []string{"ops@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	en := n.(*emailNotifier)
	if en.cfg.Timeout != 10*time.Second {
		t.Errorf("expected default timeout 10s, got %v", en.cfg.Timeout)
	}
}
