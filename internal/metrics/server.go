package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const defaultShutdownTimeout = 5 * time.Second

// Server wraps an HTTP server that exposes metrics via an Exporter.
type Server struct {
	httpServer *http.Server
}

// NewServer creates a Server that listens on addr and serves metrics
// from the given Collector at the "/metrics" path.
func NewServer(addr string, c *Collector) *Server {
	mux := http.NewServeMux()
	exporter := NewExporter(c)
	exporter.RegisterRoutes(mux, "/metrics")

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

// Start begins listening in a background goroutine. It returns an error
// if the server fails to bind.
func (s *Server) Start() error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("metrics server: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

// Stop gracefully shuts down the HTTP server.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
