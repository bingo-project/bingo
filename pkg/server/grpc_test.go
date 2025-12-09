// ABOUTME: Tests for GRPCServer implementation.
// ABOUTME: Verifies gRPC server start, stop, and Runnable interface.

package server

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestGRPCServerStartStop(t *testing.T) {
	srv := NewGRPCServer(":0", grpc.NewServer())

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Verify server is running by attempting connection
	addr := srv.Addr()
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc.Dial failed: %v", err)
	}
	conn.Close()

	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start returned error: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start did not return after cancel")
	}
}

func TestGRPCServerName(t *testing.T) {
	srv := NewGRPCServer(":9090", nil)
	if srv.Name() != "grpc" {
		t.Fatalf("Name() = %q, want %q", srv.Name(), "grpc")
	}
}
