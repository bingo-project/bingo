// ABOUTME: Health check server for K8s liveness and readiness probes.
// ABOUTME: Runs on independent port, provides /healthz and /readyz endpoints.

package server

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"sync/atomic"
)

// HealthServer provides health check endpoints.
type HealthServer struct {
	addr     string
	server   *http.Server
	listener net.Listener
	ready    atomic.Bool
}

// NewHealthServer creates a new health check server.
func NewHealthServer(addr string) *HealthServer {
	return &HealthServer{
		addr: addr,
	}
}

// SetReady sets the readiness state.
func (s *HealthServer) SetReady(ready bool) {
	s.ready.Store(ready)
}

// Start starts the health server and blocks until ctx is canceled.
func (s *HealthServer) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.listener = ln

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)

	s.server = &http.Server{
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return s.server.Shutdown(context.Background())
	}
}

// Name returns the server name for logging.
func (s *HealthServer) Name() string {
	return "health"
}

// Addr returns the actual listen address.
func (s *HealthServer) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.addr
}

func (s *HealthServer) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *HealthServer) handleReadyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.ready.Load() {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "shutting_down"})
	}
}
