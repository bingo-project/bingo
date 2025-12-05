// ABOUTME: Tests for pluggable server runner.
// ABOUTME: Validates server lifecycle management and graceful shutdown.

package server_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"bingo/internal/pkg/server"
)

type mockServer struct {
	name           string
	runCalled      atomic.Bool
	stopCalled     atomic.Bool
	runErr         error
	shutdownOrder  *[]string
}

func (m *mockServer) Run(ctx context.Context) error {
	m.runCalled.Store(true)
	if m.runErr != nil {
		return m.runErr
	}
	<-ctx.Done()
	return nil
}

func (m *mockServer) Shutdown(ctx context.Context) error {
	m.stopCalled.Store(true)
	if m.shutdownOrder != nil {
		*m.shutdownOrder = append(*m.shutdownOrder, m.name)
	}
	return nil
}

func (m *mockServer) Name() string {
	return m.name
}

func TestRunner_Run_StartsAllServers(t *testing.T) {
	srv1 := &mockServer{name: "server1"}
	srv2 := &mockServer{name: "server2"}

	runner := server.NewRunner(srv1, srv2)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = runner.Run(ctx)

	assert.True(t, srv1.runCalled.Load())
	assert.True(t, srv2.runCalled.Load())
}

func TestRunner_Shutdown_ReverseOrder(t *testing.T) {
	shutdownOrder := make([]string, 0)

	srv1 := &mockServer{name: "server1", shutdownOrder: &shutdownOrder}
	srv2 := &mockServer{name: "server2", shutdownOrder: &shutdownOrder}

	runner := server.NewRunner(srv1, srv2)
	runner.Shutdown(context.Background())

	// Should be reverse order: server2 then server1
	assert.Equal(t, []string{"server2", "server1"}, shutdownOrder)
}

func TestRunner_EmptyServers(t *testing.T) {
	runner := server.NewRunner()
	err := runner.Run(context.Background())
	assert.NoError(t, err)
}
