// ABOUTME: Signal handling utilities for graceful shutdown.
// ABOUTME: SetupSignalHandler returns context that cancels on SIGTERM/SIGINT.

package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SetupSignalHandler returns a context that is canceled when SIGTERM or SIGINT is received.
// On second signal, os.Exit(1) is called immediately.
func SetupSignalHandler() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c
		cancel() // First signal: graceful shutdown
		<-c
		os.Exit(1) // Second signal: force exit
	}()

	return ctx
}
