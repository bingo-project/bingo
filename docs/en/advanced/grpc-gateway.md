# gRPC-Gateway Progressive Learning Guide

This document provides a step-by-step guide on using gRPC-Gateway in Bingo to support both HTTP and gRPC with a single codebase.

## Overview

### What is gRPC-Gateway?

gRPC-Gateway is a protoc plugin that reads gRPC service definitions and generates a reverse proxy server that translates RESTful HTTP APIs into gRPC calls.

```
HTTP Request ──→ gRPC-Gateway (Reverse Proxy) ──→ gRPC Server
    ↑                                               │
    └─────────── JSON Response ←────────────────────┘
```

### Why Use gRPC-Gateway?

| Traditional Approach | gRPC-Gateway |
|---------------------|--------------|
| Manual HTTP routing | Auto-generated from Proto annotations |
| Manual Handler layer | Only write gRPC Handler |
| Maintain HTTP/gRPC separately | Change Proto, sync automatically |
| Manual Swagger annotations | Auto-generated OpenAPI |

## Phase 1: Environment Setup

### 1.1 Install Tools

```bash
# protoc compiler (skip if already installed)
brew install protobuf

# Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# gRPC-Gateway plugins
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# OpenAPI/Swagger generation plugin
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### 1.2 Download Required Proto Files

gRPC-Gateway requires Google API proto definitions:

```bash
# Create third_party directory at project root
mkdir -p third_party/google/api
mkdir -p third_party/protoc-gen-openapiv2/options

# Download necessary proto files
curl -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto \
  -o third_party/google/api/annotations.proto

curl -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto \
  -o third_party/google/api/http.proto

curl -L https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/annotations.proto \
  -o third_party/protoc-gen-openapiv2/options/annotations.proto

curl -L https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/openapiv2.proto \
  -o third_party/protoc-gen-openapiv2/options/openapiv2.proto
```

## Phase 2: Write Proto Files

### 2.1 Basic Structure

Using Login API as an example, create `pkg/proto/apiserver/v1/apiserver.proto`:

```protobuf
syntax = "proto3";

package apiserver.v1;

option go_package = "bingo/pkg/proto/apiserver/v1;v1";

// Import HTTP annotations
import "google/api/annotations.proto";

// Import OpenAPI annotations (optional, for Swagger generation)
import "protoc-gen-openapiv2/options/annotations.proto";

// Define service
service ApiServer {
  // User login
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "/v1/auth/login"
      body: "*"
    };
  }

  // User registration
  rpc Register(RegisterRequest) returns (RegisterResponse) {
    option (google.api.http) = {
      post: "/v1/auth/register"
      body: "*"
    };
  }

  // Get user info (requires authentication)
  rpc GetUserInfo(GetUserInfoRequest) returns (UserInfoResponse) {
    option (google.api.http) = {
      get: "/v1/auth/user-info"
    };
  }
}

// Request messages
message LoginRequest {
  string username = 1;
  string password = 2;
}

// Response messages
message LoginResponse {
  string access_token = 1;
  int64 expires_at = 2;
}

message RegisterRequest {
  string username = 1;
  string password = 2;
  string nickname = 3;
}

message RegisterResponse {
  string access_token = 1;
  int64 expires_at = 2;
}

message GetUserInfoRequest {}

message UserInfoResponse {
  string uid = 1;
  string username = 2;
  string nickname = 3;
  string email = 4;
  string avatar = 5;
}
```

### 2.2 HTTP Mapping Rules

The `google.api.http` annotation supports these mapping methods:

```protobuf
// GET request, path parameter
rpc GetPost(GetPostRequest) returns (Post) {
  option (google.api.http) = {
    get: "/v1/posts/{post_id}"
  };
}
message GetPostRequest {
  string post_id = 1;  // Automatically extracted from URL path
}

// POST request, request body
rpc CreatePost(CreatePostRequest) returns (Post) {
  option (google.api.http) = {
    post: "/v1/posts"
    body: "*"  // Entire request body maps to message
  };
}

// PUT request, path parameter + request body
rpc UpdatePost(UpdatePostRequest) returns (Post) {
  option (google.api.http) = {
    put: "/v1/posts/{post_id}"
    body: "post"  // Only post field from request body
  };
}
message UpdatePostRequest {
  string post_id = 1;  // From URL path
  Post post = 2;       // From request body
}

// DELETE request
rpc DeletePost(DeletePostRequest) returns (Empty) {
  option (google.api.http) = {
    delete: "/v1/posts/{post_id}"
  };
}

// Query parameters (non-path fields in GET become query params)
rpc ListPosts(ListPostsRequest) returns (ListPostsResponse) {
  option (google.api.http) = {
    get: "/v1/posts"
  };
}
message ListPostsRequest {
  int32 page = 1;      // ?page=1
  int32 page_size = 2; // ?page_size=10
}
```

### 2.3 Add OpenAPI Annotations (Optional)

For more complete Swagger documentation:

```protobuf
import "protoc-gen-openapiv2/options/annotations.proto";

// File-level OpenAPI configuration
option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: {
    title: "Bingo API";
    version: "1.0";
    description: "Bingo Scaffold API Documentation";
    contact: {
      name: "Bingo Project";
      url: "https://github.com/bingo-project/bingo";
    };
  };
  schemes: HTTP;
  schemes: HTTPS;
  consumes: "application/json";
  produces: "application/json";
  security_definitions: {
    security: {
      key: "Bearer";
      value: {
        type: TYPE_API_KEY;
        in: IN_HEADER;
        name: "Authorization";
        description: "Bearer token authentication";
      };
    };
  };
};

// Method-level annotations
rpc Login(LoginRequest) returns (LoginResponse) {
  option (google.api.http) = {
    post: "/v1/auth/login"
    body: "*"
  };
  option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
    summary: "User Login";
    description: "Login with username and password, returns JWT Token";
    tags: "Authentication";
  };
}
```

## Phase 3: Generate Code

### 3.1 Update Makefile

Modify `scripts/make-rules/generate.mk`:

```makefile
APIROOT := $(ROOT_DIR)/pkg/proto
THIRD_PARTY := $(ROOT_DIR)/third_party
OPENAPI_DIR := $(ROOT_DIR)/api/openapi

.PHONY: gen.protoc
gen.protoc: ## Compile protobuf files (including gRPC-Gateway)
	@echo "===========> Generate protobuf files"
	@protoc \
		--proto_path=$(APIROOT) \
		--proto_path=$(THIRD_PARTY) \
		--go_out=paths=source_relative:$(APIROOT) \
		--go-grpc_out=paths=source_relative:$(APIROOT) \
		--grpc-gateway_out=paths=source_relative:$(APIROOT) \
		--openapiv2_out=$(OPENAPI_DIR) \
		$(shell find $(APIROOT) -name "*.proto")

.PHONY: gen.gateway
gen.gateway: ## Generate only gRPC-Gateway code
	@echo "===========> Generate gRPC-Gateway files"
	@protoc \
		--proto_path=$(APIROOT) \
		--proto_path=$(THIRD_PARTY) \
		--grpc-gateway_out=paths=source_relative:$(APIROOT) \
		$(shell find $(APIROOT) -name "*.proto")

.PHONY: gen.openapi
gen.openapi: ## Generate only OpenAPI documentation
	@echo "===========> Generate OpenAPI files"
	@mkdir -p $(OPENAPI_DIR)
	@protoc \
		--proto_path=$(APIROOT) \
		--proto_path=$(THIRD_PARTY) \
		--openapiv2_out=$(OPENAPI_DIR) \
		--openapiv2_opt=logtostderr=true \
		$(shell find $(APIROOT) -name "*.proto")
```

### 3.2 Update tools.mk

Add installation commands for new tools:

```makefile
.PHONY: install.grpc-gateway
install.grpc-gateway:
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### 3.3 Execute Generation

```bash
make gen.protoc
```

Generated files:
- `apiserver.pb.go` - protobuf message definitions
- `apiserver_grpc.pb.go` - gRPC server/client code
- `apiserver.pb.gw.go` - **gRPC-Gateway reverse proxy code**
- `api/openapi/apiserver.swagger.json` - OpenAPI documentation

## Phase 4: Implement Service

### 4.1 Implement gRPC Handler

Create `internal/apiserver/handler/grpc/auth.go`:

```go
package grpc

import (
	"context"

	"bingo/internal/apiserver/biz"
	pb "bingo/pkg/proto/apiserver/v1"
)

type Handler struct {
	pb.UnimplementedApiServerServer
	b biz.IBiz
}

func NewHandler(b biz.IBiz) *Handler {
	return &Handler{b: b}
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Call existing Biz layer
	resp, err := h.b.Auth().Login(ctx, &v1.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err // Will be converted to gRPC status error
	}

	return &pb.LoginResponse{
		AccessToken: resp.AccessToken,
		ExpiresAt:   resp.ExpiresAt,
	}, nil
}

func (h *Handler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	resp, err := h.b.Auth().Register(ctx, &v1.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
	})
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		AccessToken: resp.AccessToken,
		ExpiresAt:   resp.ExpiresAt,
	}, nil
}
```

### 4.2 Create gRPC-Gateway Server

Create `internal/pkg/server/grpc_gateway.go`:

```go
package server

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type GRPCGatewayServer struct {
	srv         *http.Server
	grpcAddress string
}

func NewGRPCGatewayServer(
	httpAddr string,
	grpcAddr string,
	registerHandler func(mux *runtime.ServeMux, conn *grpc.ClientConn) error,
) (*GRPCGatewayServer, error) {
	// Connect to gRPC server
	conn, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	// Create Gateway multiplexer
	gwmux := runtime.NewServeMux(
		// JSON serialization options
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseEnumNumbers:  true,  // Use numbers for enums
				EmitUnpopulated: false, // Don't output zero-value fields
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true, // Ignore unknown fields
			},
		}),
		// Custom error handler (see next section)
		runtime.WithErrorHandler(customErrorHandler),
	)

	// Register handlers
	if err := registerHandler(gwmux, conn); err != nil {
		return nil, err
	}

	return &GRPCGatewayServer{
		srv: &http.Server{
			Addr:    httpAddr,
			Handler: gwmux,
		},
		grpcAddress: grpcAddr,
	}, nil
}

func (s *GRPCGatewayServer) Run() error {
	return s.srv.ListenAndServe()
}

func (s *GRPCGatewayServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
```

### 4.3 Custom Error Handling

```go
package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"bingo/pkg/errorsx"
)

// ErrResponse unified error response format (aligned with errorsx.ErrorX)
type ErrResponse struct {
	Code     int               `json:"code"`
	Reason   string            `json:"reason"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func customErrorHandler(
	ctx context.Context,
	mux *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	// Use errorsx.FromError to parse ErrorX from gRPC error
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

### 4.4 Start Service

Modify `internal/apiserver/run.go`:

```go
package apiserver

import (
	"context"

	pb "bingo/pkg/proto/apiserver/v1"
)

func run() error {
	// ... initialize store, biz, etc.

	// 1. Start gRPC server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			// interceptors...
		),
	)
	pb.RegisterApiServerServer(grpcServer, handler.NewHandler(bizInstance))

	go func() {
		lis, _ := net.Listen("tcp", ":9090")
		grpcServer.Serve(lis)
	}()

	// 2. Start gRPC-Gateway (HTTP reverse proxy)
	gwServer, err := server.NewGRPCGatewayServer(
		":8080",  // HTTP port
		":9090",  // gRPC port
		func(mux *runtime.ServeMux, conn *grpc.ClientConn) error {
			return pb.RegisterApiServerHandler(context.Background(), mux, conn)
		},
	)
	if err != nil {
		return err
	}

	return gwServer.Run()
}
```

## Phase 5: Verification

### 5.1 Test gRPC Interface

```bash
# Using grpcurl
grpcurl -plaintext -d '{"username":"test","password":"123456"}' \
  localhost:9090 apiserver.v1.ApiServer/Login
```

### 5.2 Test HTTP Interface

```bash
# Using curl
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

### 5.3 View Swagger Documentation

Generated `api/openapi/apiserver.swagger.json` can be:
- Imported into Swagger UI
- Imported into Postman
- Viewed online with Swagger Editor

## FAQ

### Q1: How to Handle File Uploads?

gRPC-Gateway is not suitable for multipart/form-data. Use a mixed approach:

```go
mux := http.NewServeMux()
mux.Handle("/", gwmux)                           // Most APIs
mux.Handle("/v1/file/upload", ginFileHandler)    // File upload via Gin
```

### Q2: How to Pass Custom Headers?

```go
gwmux := runtime.NewServeMux(
	runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
		switch key {
		case "X-Request-Id", "Authorization":
			return key, true
		}
		return "", false
	}),
)
```

### Q3: How to Get Client IP?

```go
gwmux := runtime.NewServeMux(
	runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
		return metadata.Pairs(
			"x-real-ip", getRealIP(req),
			"x-request-id", req.Header.Get("X-Request-Id"),
		)
	}),
)
```

## Related Documentation

- [Pluggable Protocol Layer](protocol-layer.md) - HTTP/gRPC/WebSocket unified architecture
- [WebSocket Design and Implementation](websocket.md) - JSON-RPC 2.0 message format, middleware architecture

---

**Next Step**: Learn about [Unified Error Handling](unified-error-handling.md) to understand how HTTP/gRPC/WebSocket share the same error format.
