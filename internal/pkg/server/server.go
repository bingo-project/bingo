// ABOUTME: Pluggable server interface and runner.
// ABOUTME: Enables configuration-driven protocol selection with graceful shutdown.

package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"
)

// Server is the interface for pluggable servers.
type Server interface {
	// Run starts the server (blocks until context is cancelled or error).
	Run(ctx context.Context) error
	// Shutdown gracefully shuts down the server.
	Shutdown(ctx context.Context) error
	// Name returns the server name (for logging).
	Name() string
}

// Runner manages multiple servers lifecycle.
type Runner struct {
	servers []Server
}

// NewRunner creates a new server runner.
func NewRunner(servers ...Server) *Runner {
	return &Runner{servers: servers}
}

// Run starts all servers, any failure triggers shutdown of all.
func (r *Runner) Run(ctx context.Context) error {
	if len(r.servers) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	// Start all servers
	for _, srv := range r.servers {
		srv := srv
		g.Go(func() error {
			return srv.Run(ctx)
		})
	}

	// Wait for context cancellation, then trigger graceful shutdown
	g.Go(func() error {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		return r.Shutdown(shutdownCtx)
	})

	return g.Wait()
}

// Shutdown gracefully shuts down all servers in reverse order.
func (r *Runner) Shutdown(ctx context.Context) error {
	var errs []error

	// Shutdown in reverse order: last started, first stopped
	for i := len(r.servers) - 1; i >= 0; i-- {
		srv := r.servers[i]
		if err := srv.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", srv.Name(), err))
		}
	}

	return errors.Join(errs...)
}
