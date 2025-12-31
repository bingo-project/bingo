// ABOUTME: AI streaming response handler.
// ABOUTME: Provides ChatStream for handling SSE responses.

package ai

import (
	"sync"
)

const (
	// DefaultStreamBufferSize is the default buffer size for chat streams.
	DefaultStreamBufferSize = 100
)

// ChatStream represents a streaming chat response
type ChatStream struct {
	chunks chan *StreamChunk
	err    error
	done   bool
	mu     sync.Mutex
}

// NewChatStream creates a new ChatStream
func NewChatStream(bufferSize int) *ChatStream {
	return &ChatStream{
		chunks: make(chan *StreamChunk, bufferSize),
	}
}

// Recv receives the next chunk from the stream
func (s *ChatStream) Recv() (*StreamChunk, error) {
	chunk, ok := <-s.chunks
	if !ok {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.err != nil {
			return nil, s.err
		}

		return nil, ErrStreamClosed
	}

	return chunk, nil
}

// Send sends a chunk to the stream (used by providers)
func (s *ChatStream) Send(chunk *StreamChunk) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.done {
		s.chunks <- chunk
	}
}

// Close closes the stream
func (s *ChatStream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.done {
		s.done = true
		close(s.chunks)
	}
}

// CloseWithError closes the stream with an error
func (s *ChatStream) CloseWithError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.done {
		s.done = true
		s.err = err
		close(s.chunks)
	}
}
