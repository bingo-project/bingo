// ABOUTME: Tests for JSON-RPC 2.0 message types and response constructors.
// ABOUTME: Validates request/response serialization and error handling.

package jsonrpc_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

func TestRequest_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected jsonrpc.Request
	}{
		{
			name:  "basic request with params",
			input: `{"jsonrpc":"2.0","method":"user.login","params":{"username":"test"},"id":1}`,
			expected: jsonrpc.Request{
				JSONRPC: "2.0",
				Method:  "user.login",
				Params:  json.RawMessage(`{"username":"test"}`),
				ID:      float64(1), // JSON numbers unmarshal to float64
			},
		},
		{
			name:  "notification without id",
			input: `{"jsonrpc":"2.0","method":"heartbeat"}`,
			expected: jsonrpc.Request{
				JSONRPC: "2.0",
				Method:  "heartbeat",
			},
		},
		{
			name:  "request with string id",
			input: `{"jsonrpc":"2.0","method":"test","id":"abc-123"}`,
			expected: jsonrpc.Request{
				JSONRPC: "2.0",
				Method:  "test",
				ID:      "abc-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req jsonrpc.Request
			err := json.Unmarshal([]byte(tt.input), &req)
			require.NoError(t, err)
			assert.Equal(t, tt.expected.JSONRPC, req.JSONRPC)
			assert.Equal(t, tt.expected.Method, req.Method)
			assert.Equal(t, tt.expected.ID, req.ID)
			if tt.expected.Params != nil {
				assert.JSONEq(t, string(tt.expected.Params), string(req.Params))
			}
		})
	}
}

func TestNewResponse_Success(t *testing.T) {
	result := map[string]string{"token": "abc123"}
	resp := jsonrpc.NewResponse(1, result)

	assert.Equal(t, jsonrpc.Version, resp.JSONRPC)
	assert.Equal(t, 1, resp.ID)
	assert.Equal(t, result, resp.Result)
	assert.Nil(t, resp.Error)
}

func TestNewErrorResponse(t *testing.T) {
	err := errorsx.New(404, "NotFound", "User not found")
	resp := jsonrpc.NewErrorResponse(1, err)

	assert.Equal(t, jsonrpc.Version, resp.JSONRPC)
	assert.Equal(t, 1, resp.ID)
	assert.Nil(t, resp.Result)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32004, resp.Error.Code) // 404 -> -32004
	assert.Equal(t, "NotFound", resp.Error.Reason)
	assert.Equal(t, "User not found", resp.Error.Message)
}

func TestNewErrorResponse_WithMetadata(t *testing.T) {
	err := errorsx.New(400, "InvalidParams", "Validation failed").
		WithMetadata(map[string]string{"field": "email"})
	resp := jsonrpc.NewErrorResponse("req-1", err)

	require.NotNil(t, resp.Error)
	assert.Equal(t, "email", resp.Error.Data["field"])
}

func TestNewNotification(t *testing.T) {
	params := map[string]int{"count": 5}
	resp := jsonrpc.NewNotification("update", params)

	assert.Equal(t, jsonrpc.Version, resp.JSONRPC)
	assert.Equal(t, "update", resp.Method)
	assert.Equal(t, params, resp.Result)
	assert.Nil(t, resp.ID)
}

func TestResponse_Marshal(t *testing.T) {
	resp := jsonrpc.NewResponse(1, map[string]string{"status": "ok"})
	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var parsed map[string]any
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "2.0", parsed["jsonrpc"])
	assert.Equal(t, float64(1), parsed["id"])
	assert.NotNil(t, parsed["result"])
	assert.Nil(t, parsed["error"])
}
