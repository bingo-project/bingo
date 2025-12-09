// ABOUTME: Tests for signal handling utilities.
// ABOUTME: Verifies SetupSignalHandler cancels context on SIGINT/SIGTERM.

package app

import (
	"syscall"
	"testing"
	"time"
)

func TestSetupSignalHandler(t *testing.T) {
	ctx := SetupSignalHandler()

	// Context should not be done initially
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done initially")
	default:
		// expected
	}

	// Send SIGINT
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	// Context should be done after signal
	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context should be done after SIGINT")
	}
}
