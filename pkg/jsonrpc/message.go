// ABOUTME: JSON-RPC 2.0 message types for WebSocket communication.
// ABOUTME: Defines Request, Response, and Error structures per JSON-RPC 2.0 spec.

package jsonrpc

import "encoding/json"

// Version is the JSON-RPC protocol version.
const Version = "2.0"

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method,omitempty"` // For notifications
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
	ID      any    `json:"id,omitempty"`
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Code    int               `json:"code"`
	Reason  string            `json:"reason"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
}

// IsNotification returns true if the request is a notification (no ID).
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// Push represents a server-initiated push message (no ID, not tied to a request).
type Push struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Data    any    `json:"data,omitempty"`
}
