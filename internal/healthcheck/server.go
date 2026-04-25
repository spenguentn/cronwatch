package healthcheck

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"
)

// Server wraps an HTTP server that exposes the health endpoint.
type Server struct {
	httpServer *http.Server
	checker    *Checker
}

// NewServer creates a Server listening on addr and using the provided Checker.
func NewServer(addr string, checker *Checker) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", checker.Handler())

	return &Server{
		checker: checker,
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	}
}

// Start begins serving in a background goroutine.
// It returns the actual listening address (useful when addr uses port :0).
func (s *Server) Start() (string, error) {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return "", err
	}
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("healthcheck server error: %v", err)
		}
	}()
	return ln.Addr().String(), nil
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
