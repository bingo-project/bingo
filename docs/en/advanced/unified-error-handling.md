# Unified HTTP/gRPC/WebSocket Error Handling

This document describes how to design a unified error handling mechanism that allows HTTP, gRPC, and WebSocket protocols to share the same error definitions and response format.

## Design Goals

1. **Single Error Definition** - Business errors defined once, shared across all protocols
2. **Consistent Format** - Clients receive uniform error format
3. **Clear Semantics** - Error codes have hierarchical structure for classification
4. **Extensible** - Support dynamic messages, metadata, and advanced features

## Core Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  Error Definition Layer (errno)              │
│  ErrUserNotFound, ErrTokenInvalid, ErrPermissionDenied...   │
└─────────────────────────────────┬───────────────────────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    ▼             ▼             ▼
             ┌──────────┐  ┌──────────┐  ┌──────────┐
             │   HTTP   │  │   gRPC   │  │WebSocket │
             │ Response │  │  Status  │  │ Message  │
             └──────────┘  └──────────┘  └──────────┘
                    │             │             │
                    ▼             ▼             ▼
             ┌───────────────────────────────────────────────────────────┐
             │                    Unified Response Format                 │
             │  {"code": 401, "reason": "xxx", "message": "xxx"}         │
             └───────────────────────────────────────────────────────────┘
```

## Phase 1: Using bingo/pkg/errorsx

We use the `bingo/pkg/errorsx` package with its ErrorX type for a clean design:

### 1.1 ErrorX Struct

```go
// bingo/pkg/errorsx

// ErrorX defines the unified error type
type ErrorX struct {
	// Code represents HTTP status code, automatically converted to gRPC status code
	Code int `json:"code,omitempty"`

	// Reason represents business error code, hierarchical naming: "Category.SubCategory.ErrorName"
	Reason string `json:"reason,omitempty"`

	// Message represents a short error message
	Message string `json:"message,omitempty"`

	// Metadata stores additional context information
	Metadata map[string]string `json:"metadata,omitempty"`
}
```

**Design Advantages:**
- Only specify HTTP status code, auto-convert to gRPC status code
- Use `GRPCStatus()` method to get gRPC status (with ErrorInfo details)
- Support chained calls like `WithMessage()`, `KV()`

### 1.2 Core Methods

```go
// New creates a new error
func New(code int, reason string, format string, args ...any) *ErrorX

// WithMessage sets error message
func (err *ErrorX) WithMessage(format string, args ...any) *ErrorX

// KV adds metadata
func (err *ErrorX) KV(kvs ...string) *ErrorX

// GRPCStatus returns gRPC status (with ErrorInfo details)
func (err *ErrorX) GRPCStatus() *status.Status

// FromError converts any error to ErrorX
func FromError(err error) *ErrorX

// Is checks error type
func (err *ErrorX) Is(target error) bool
```

### 1.3 HTTP ↔ gRPC Status Code Mapping

`errorsx` uses kratos's `httpstatus` package for automatic mapping:

| HTTP Status Code | gRPC Status Code |
|-----------------|------------------|
| 200 OK | OK |
| 400 Bad Request | InvalidArgument |
| 401 Unauthorized | Unauthenticated |
| 403 Forbidden | PermissionDenied |
| 404 Not Found | NotFound |
| 409 Conflict | AlreadyExists |
| 429 Too Many Requests | ResourceExhausted |
| 500 Internal Server Error | Internal |
| 503 Service Unavailable | Unavailable |

## Phase 2: Define Business Error Codes

### 2.1 Common Error Codes

Modify `internal/pkg/errno/code.go`, reusing `errorsx` predefined errors:

```go
package errno

import (
	"net/http"

	"bingo/pkg/errorsx"
)

var (
	// OK represents successful request
	OK = &errorsx.ErrorX{Code: http.StatusOK, Message: ""}

	// Common errors (reuse errorsx predefined)
	ErrInternal         = errorsx.ErrInternal
	ErrNotFound         = errorsx.ErrNotFound
	ErrBind             = errorsx.ErrBind
	ErrInvalidArgument  = errorsx.ErrInvalidArgument
	ErrUnauthenticated  = errorsx.ErrUnauthenticated
	ErrPermissionDenied = errorsx.ErrPermissionDenied
	ErrOperationFailed  = errorsx.ErrOperationFailed

	// Authentication errors
	ErrSignToken    = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.SignToken", Message: "Error occurred while signing the JSON web token."}
	ErrTokenInvalid = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.TokenInvalid", Message: "Token was invalid."}
	ErrTokenExpired = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.TokenExpired", Message: "Token has expired."}

	// Service errors
	ErrServiceUnavailable = &errorsx.ErrorX{Code: http.StatusServiceUnavailable, Reason: "ServiceUnavailable", Message: "Service unavailable."}
	ErrTooManyRequests    = &errorsx.ErrorX{Code: http.StatusTooManyRequests, Reason: "TooManyRequests", Message: "Too many requests."}
)
```

### 2.2 Domain Error Codes

Create `internal/pkg/errno/user.go`:

```go
package errno

import (
	"net/http"

	"bingo/pkg/errorsx"
)

var (
	ErrUserNotFound      = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.UserNotFound", Message: "User not found."}
	ErrUserAlreadyExists = &errorsx.ErrorX{Code: http.StatusConflict, Reason: "OperationFailed.UserAlreadyExists", Message: "User already exists."}
	ErrPasswordIncorrect = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.PasswordIncorrect", Message: "Password is incorrect."}
	ErrUsernameInvalid   = &errorsx.ErrorX{Code: http.StatusBadRequest, Reason: "InvalidArgument.UsernameInvalid", Message: "Invalid username format."}
)
```

Create `internal/pkg/errno/app.go`:

```go
package errno

import (
	"net/http"

	"bingo/pkg/errorsx"
)

var (
	ErrAppNotFound = &errorsx.ErrorX{Code: http.StatusNotFound, Reason: "NotFound.AppNotFound", Message: "Application not found."}
	ErrAppDisabled = &errorsx.ErrorX{Code: http.StatusForbidden, Reason: "PermissionDenied.AppDisabled", Message: "Application is disabled."}
)
```

## Phase 3: Protocol Adaptation Layer

### 3.1 HTTP Response Handling

Modify `internal/pkg/core/response.go`:

```go
package core

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"bingo/pkg/errorsx"

	"bingo/internal/pkg/errno"
)

// ErrResponse unified error response format (aligned with errorsx.ErrorX)
type ErrResponse struct {
	Code     int               `json:"code"`
	Reason   string            `json:"reason"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Response writes HTTP response
func Response(c *gin.Context, data any, err error) {
	if err != nil {
		e := errorsx.FromError(err)
		c.JSON(e.Code, ErrResponse{
			Reason:   e.Reason,
			Message:  e.Message,
			Metadata: e.Metadata,
		})
		return
	}

	c.JSON(http.StatusOK, data)
}

// HandleJSONRequest generic JSON request handler
func HandleJSONRequest[Req, Resp any](
	c *gin.Context,
	handler func(ctx context.Context, req *Req) (*Resp, error),
	validators ...func(*Req) error,
) {
	var req Req

	// Bind request body
	if err := c.ShouldBindJSON(&req); err != nil {
		Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
		return
	}

	// Run validators
	for _, validate := range validators {
		if err := validate(&req); err != nil {
			Response(c, nil, err)
			return
		}
	}

	// Call handler function
	resp, err := handler(c.Request.Context(), &req)
	Response(c, resp, err)
}
```

### 3.2 gRPC Error Handling

In gRPC Handler, use `GRPCStatus().Err()` to return errors:

```go
// internal/apiserver/handler/grpc/auth.go

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp, err := h.biz.Auth().Login(ctx, bizReq)
	if err != nil {
		// ErrorX automatically converts to gRPC status (with ErrorInfo details)
		return nil, errorsx.FromError(err).GRPCStatus().Err()
	}
	return resp, nil
}
```

gRPC-Gateway custom error handler:

```go
// internal/apiserver/gateway/error.go

func CustomErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	// Parse ErrorX from gRPC error
	e := errorsx.FromError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Code)

	resp := ErrResponse{
		Code:     e.Code,
		Reason:   e.Reason,
		Message:  e.Message,
		Metadata: e.Metadata,
	}
	json.NewEncoder(w).Encode(resp)
}
```

### 3.3 WebSocket Error Handling (JSON-RPC 2.0)

WebSocket layer uses JSON-RPC 2.0 protocol, error codes follow JSON-RPC specification:

```go
// pkg/jsonrpc/message.go

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
	Code    int               `json:"code"`    // JSON-RPC error code
	Reason  string            `json:"reason"`  // Business error code
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
}
```

```go
// pkg/jsonrpc/response.go

// NewErrorResponse creates an error response from an error.
func NewErrorResponse(id any, err error) *Response {
	e := errorsx.FromError(err)

	return &Response{
		JSONRPC: Version,
		Error: &Error{
			Code:    e.JSONRPCCode(),  // Convert to JSON-RPC error code
			Reason:  e.Reason,
			Message: e.Message,
			Data:    e.Metadata,
		},
		ID: id,
	}
}
```

#### HTTP → JSON-RPC Error Code Mapping

`errorsx.JSONRPCCode()` method automatically converts HTTP status codes to JSON-RPC error codes:

| HTTP Status Code | JSON-RPC Error Code | Description |
|------------------|---------------------|-------------|
| 400 Bad Request | -32602 | Invalid params |
| 401 Unauthorized | -32001 | Unauthenticated |
| 403 Forbidden | -32003 | Permission denied |
| 404 Not Found | -32004 | Not found |
| 409 Conflict | -32009 | Conflict |
| 429 Too Many Requests | -32029 | Too many requests |
| 500 Internal Server Error | -32603 | Internal error |
| 503 Service Unavailable | -32053 | Service unavailable |

## Phase 4: Biz Layer Error Usage

### 4.1 Return Predefined Errors

```go
func (b *userBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	user, err := b.store.User().GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errno.ErrUserNotFound  // Return predefined error directly
	}

	if err := auth.Compare(user.Password, req.Password); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// ...
}
```

### 4.2 Dynamic Messages

```go
func (b *userBiz) Create(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	// Use WithMessage to set dynamic error message
	if len(req.Username) < 3 {
		return nil, errno.ErrUsernameInvalid.WithMessage("Username must be at least 3 characters, got %d", len(req.Username))
	}

	// ...
}
```

### 4.3 Add Metadata

```go
func (b *userBiz) Get(ctx context.Context, uid string) (*v1.User, error) {
	user, err := b.store.User().GetByUID(ctx, uid)
	if err != nil {
		// Use KV to add debug information
		return nil, errno.ErrUserNotFound.KV("uid", uid, "source", "biz.user.Get")
	}
	return user, nil
}
```

### 4.4 Error Comparison

```go
import "errors"

func handleError(err error) {
	// Use errors.Is to check error type
	if errors.Is(err, errno.ErrUserNotFound) {
		// Special handling for user not found
	}

	// Or use ErrorX.Is method
	if errno.ErrUnauthenticated.Is(err) {
		// Special handling for unauthenticated
	}
}
```

## Error Code Naming Convention

### Hierarchical Naming

```
Category.SubCategory.ErrorName
```

**Category (First Level):**
- `InvalidParameter` - Parameter error (400)
- `AuthFailure` - Authentication failure (401)
- `Forbidden` - Authorization failure (403)
- `ResourceNotFound` - Resource not found (404)
- `ResourceAlreadyExists` - Resource already exists (409)
- `TooManyRequests` - Too many requests (429)
- `InternalError` - Internal error (500)
- `ServiceUnavailable` - Service unavailable (503)

**Examples:**
```
AuthFailure.TokenInvalid       - Token invalid
AuthFailure.TokenExpired       - Token expired
AuthFailure.PasswordIncorrect  - Password incorrect

ResourceNotFound.UserNotFound  - User not found
ResourceNotFound.AppNotFound   - Application not found

InvalidParameter.BindError     - Parameter binding error
InvalidParameter.UsernameInvalid - Username format invalid
```

## Complete Flow Examples

### HTTP Request (Gin)

```
POST /v1/auth/login
{"username": "test", "password": "wrong"}

↓ Gin Handler
↓ Biz.Login() returns errno.ErrPasswordIncorrect
↓ core.Response()

HTTP 401
{"code": 401, "reason": "Unauthenticated.PasswordIncorrect", "message": "Password is incorrect."}
```

### gRPC Request

```
grpcurl -d '{"username":"test","password":"wrong"}' :9090 apiserver.v1.AuthService/Login

↓ gRPC Handler
↓ Biz.Login() returns errno.ErrPasswordIncorrect
↓ return nil, errorsx.FromError(err).GRPCStatus().Err()

gRPC Status: UNAUTHENTICATED
Details: ErrorInfo{Reason: "Unauthenticated.PasswordIncorrect"}
Message: "Password is incorrect."
```

### gRPC-Gateway (HTTP → gRPC)

```
POST /v1/auth/login (via gRPC-Gateway)
{"username": "test", "password": "wrong"}

↓ gRPC-Gateway forwards to gRPC
↓ gRPC Handler returns status with ErrorInfo
↓ CustomErrorHandler parses ErrorInfo

HTTP 401
{"code": 401, "reason": "Unauthenticated.PasswordIncorrect", "message": "Password is incorrect."}
```

### WebSocket Request (JSON-RPC 2.0)

```json
{"jsonrpc": "2.0", "method": "Login", "params": {"username": "test", "password": "wrong"}, "id": 1}

↓ JSON-RPC Router
↓ Biz.Login() returns errno.ErrPasswordIncorrect
↓ jsonrpc.NewErrorResponse()

{"jsonrpc": "2.0", "error": {"code": -32001, "reason": "Unauthenticated.PasswordIncorrect", "message": "Password is incorrect."}, "id": 1}
```

Note: WebSocket layer's `error.code` is JSON-RPC error code (-32001), not HTTP status code (401).

## Summary

Advantages of using `bingo/pkg/errorsx`:

1. **Clean Design** - Only specify HTTP status code, auto-convert to gRPC/JSON-RPC error codes
2. **Production Ready** - Proven in practice
3. **gRPC Friendly** - `GRPCStatus()` method automatically adds ErrorInfo details
4. **JSON-RPC Friendly** - `JSONRPCCode()` method automatically converts error codes
5. **Chained Calls** - Support `WithMessage()`, `KV()` and other methods

## Related Documentation

- [Pluggable Protocol Layer](protocol-layer.md) - HTTP/gRPC/WebSocket unified architecture
- [WebSocket Design and Implementation](websocket.md) - JSON-RPC 2.0 message format, middleware architecture
- [Unified Authentication](unified-auth.md) - Plugin-based authentication architecture

---

**Next Step**: Learn about [Microservices](microservices.md) to understand how to evolve from a monolith to microservices architecture.
