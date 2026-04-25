package healthcheck_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/healthcheck"
)

func startTestServer(t *testing.T) (*healthcheck.Server, *healthcheck.Checker, string) {
	t.Helper()
	checker := healthcheck.New()
	srv := healthcheck.NewServer("127.0.0.1:0", checker)
	addr, err := srv.Start()
	if err != nil {
		t.Fatalf("failed to start server: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	})
	return srv, checker, addr
}

func TestServer_HealthEndpoint(t *testing.T) {
	_, _, addr := startTestServer(t)
	resp, err := http.Get(fmt.Sprintf("http://%s/health", addr))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var s healthcheck.Status
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !s.OK {
		t.Fatal("expected ok=true")
	}
}

func TestServer_UnhealthyEndpoint(t *testing.T) {
	_, checker, addr := startTestServer(t)
	checker.SetHealthy(false, "test failure")
	resp, err := http.Get(fmt.Sprintf("http://%s/health", addr))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}
}

func TestServer_Shutdown(t *testing.T) {
	srv, _, addr := startTestServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
	_, err := http.Get(fmt.Sprintf("http://%s/health", addr))
	if err == nil {
		t.Fatal("expected connection error after shutdown")
	}
}
