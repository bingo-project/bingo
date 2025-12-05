// ABOUTME: JSON-RPC 2.0 response constructors.
// ABOUTME: Creates success responses, error responses, and notifications.

package jsonrpc

import "bingo/pkg/errorsx"

// NewResponse creates a success response.
func NewResponse(id any, result any) *Response {
	return &Response{
		JSONRPC: Version,
		Result:  result,
		ID:      id,
	}
}

// NewErrorResponse creates an error response from an error.
func NewErrorResponse(id any, err error) *Response {
	e := errorsx.FromError(err)
	return &Response{
		JSONRPC: Version,
		Error: &Error{
			Code:    e.JSONRPCCode(),
			Reason:  e.Reason,
			Message: e.Message,
			Data:    e.Metadata,
		},
		ID: id,
	}
}

// NewNotification creates a server push notification (no id).
func NewNotification(method string, params any) *Response {
	return &Response{
		JSONRPC: Version,
		Method:  method,
		Result:  params,
	}
}
