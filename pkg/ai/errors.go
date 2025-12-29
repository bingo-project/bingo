// ABOUTME: AI package error definitions.
// ABOUTME: Contains sentinel errors for AI operations.

package ai

import "errors"

var (
	ErrModelNotFound    = errors.New("model not found")
	ErrProviderNotFound = errors.New("provider not found")
	ErrStreamClosed     = errors.New("stream closed")
	ErrEmptyMessages    = errors.New("messages cannot be empty")
	ErrContextTooLong   = errors.New("context too long")
)
