// ABOUTME: Tests for JSON-RPC adapter that routes methods to handlers.
// ABOUTME: Validates method registration, dispatching, and error handling.

package jsonrpc_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

// Test request/response types
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func TestAdapter_Handle_Success(t *testing.T) {
	adapter := jsonrpc.NewAdapter()

	// Register handler
	jsonrpc.Register(adapter, "user.login", func(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
		return &LoginResponse{Token: "token-for-" + req.Username}, nil
	})

	// Create request
	params, _ := json.Marshal(LoginRequest{Username: "alice", Password: "secret"})
	req := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "user.login",
		Params:  params,
		ID:      1,
	}

	// Handle request
	resp := adapter.Handle(context.Background(), req)

	// Verify response
	assert.Equal(t, "2.0", resp.JSONRPC)
	assert.Equal(t, 1, resp.ID)
	assert.Nil(t, resp.Error)

	// Unmarshal result
	resultJSON, _ := json.Marshal(resp.Result)
	var result LoginResponse
	err := json.Unmarshal(resultJSON, &result)
	require.NoError(t, err)
	assert.Equal(t, "token-for-alice", result.Token)
}

func TestAdapter_Handle_MethodNotFound(t *testing.T) {
	adapter := jsonrpc.NewAdapter()

	req := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "unknown.method",
		ID:      1,
	}

	resp := adapter.Handle(context.Background(), req)

	assert.Equal(t, 1, resp.ID)
	require.NotNil(t, resp.Error)
	assert.Equal(t, -32004, resp.Error.Code) // NotFound -> -32004
	assert.Equal(t, "MethodNotFound", resp.Error.Reason)
}

func TestAdapter_Handle_InvalidParams(t *testing.T) {
	adapter := jsonrpc.NewAdapter()

	jsonrpc.Register(adapter, "user.login", func(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
		return &LoginResponse{Token: "test"}, nil
	})

	req := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "user.login",
		Params:  json.RawMessage(`{"invalid json`), // Malformed JSON
		ID:      1,
	}

	resp := adapter.Handle(context.Background(), req)

	require.NotNil(t, resp.Error)
	assert.Equal(t, -32602, resp.Error.Code) // InvalidParams -> -32602
}

func TestAdapter_Handle_HandlerError(t *testing.T) {
	adapter := jsonrpc.NewAdapter()

	jsonrpc.Register(adapter, "user.login", func(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
		return nil, errorsx.New(401, "Unauthenticated", "Invalid credentials")
	})

	params, _ := json.Marshal(LoginRequest{Username: "alice", Password: "wrong"})
	req := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "user.login",
		Params:  params,
		ID:      1,
	}

	resp := adapter.Handle(context.Background(), req)

	require.NotNil(t, resp.Error)
	assert.Equal(t, -32001, resp.Error.Code) // Unauthenticated -> -32001
	assert.Equal(t, "Unauthenticated", resp.Error.Reason)
	assert.Equal(t, "Invalid credentials", resp.Error.Message)
}

func TestAdapter_Handle_EmptyParams(t *testing.T) {
	adapter := jsonrpc.NewAdapter()

	type EmptyRequest struct{}
	type StatusResponse struct {
		Status string `json:"status"`
	}

	jsonrpc.Register(adapter, "system.status", func(ctx context.Context, req *EmptyRequest) (*StatusResponse, error) {
		return &StatusResponse{Status: "ok"}, nil
	})

	req := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "system.status",
		ID:      1,
	}

	resp := adapter.Handle(context.Background(), req)

	assert.Nil(t, resp.Error)
	resultJSON, _ := json.Marshal(resp.Result)
	var result StatusResponse
	json.Unmarshal(resultJSON, &result)
	assert.Equal(t, "ok", result.Status)
}

func TestAdapter_Handle_ContextPropagation(t *testing.T) {
	adapter := jsonrpc.NewAdapter()

	type ctxKey string
	const userIDKey ctxKey = "userID"

	jsonrpc.Register(adapter, "user.profile", func(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
		userID := ctx.Value(userIDKey).(string)
		return &LoginResponse{Token: "token-for-" + userID}, nil
	})

	ctx := context.WithValue(context.Background(), userIDKey, "user-123")
	req := &jsonrpc.Request{
		JSONRPC: "2.0",
		Method:  "user.profile",
		Params:  json.RawMessage(`{}`),
		ID:      1,
	}

	resp := adapter.Handle(ctx, req)

	assert.Nil(t, resp.Error)
	resultJSON, _ := json.Marshal(resp.Result)
	var result LoginResponse
	json.Unmarshal(resultJSON, &result)
	assert.Equal(t, "token-for-user-123", result.Token)
}
