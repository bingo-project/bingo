# gRPC-Gateway 渐进式学习指南

本文档基于 Bingo 项目，从零开始介绍如何使用 gRPC-Gateway 让一份代码同时支持 HTTP 和 gRPC 两种协议。

## 概述

### 什么是 gRPC-Gateway？

gRPC-Gateway 是一个 protoc 插件，它读取 gRPC 服务定义，生成一个反向代理服务器，将 RESTful HTTP API 转换为 gRPC 调用。

```
HTTP 请求 ──→ gRPC-Gateway (反向代理) ──→ gRPC 服务器
    ↑                                         │
    └─────────── JSON 响应 ←──────────────────┘
```

### 为什么使用 gRPC-Gateway？

| 传统方式 | gRPC-Gateway |
|---------|--------------|
| HTTP 路由手写 | Proto 注解自动生成 |
| Controller 层手写 | 只写 gRPC Handler |
| HTTP/gRPC 分别维护 | 改 Proto 自动同步 |
| Swagger 手写注解 | OpenAPI 自动生成 |

## 第一阶段：环境准备

### 1.1 安装工具

```bash
# protoc 编译器（如已安装可跳过）
brew install protobuf

# Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# gRPC-Gateway 插件
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# OpenAPI/Swagger 生成插件
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### 1.2 下载依赖的 proto 文件

gRPC-Gateway 需要 Google API 的 proto 定义：

```bash
# 在项目根目录创建 third_party 目录
mkdir -p third_party/google/api
mkdir -p third_party/protoc-gen-openapiv2/options

# 下载必要的 proto 文件
curl -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto \
  -o third_party/google/api/annotations.proto

curl -L https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto \
  -o third_party/google/api/http.proto

curl -L https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/annotations.proto \
  -o third_party/protoc-gen-openapiv2/options/annotations.proto

curl -L https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/openapiv2.proto \
  -o third_party/protoc-gen-openapiv2/options/openapiv2.proto
```

## 第二阶段：编写 Proto 文件

### 2.1 基本结构

以 Login API 为例，创建 `pkg/proto/apiserver/v1/apiserver.proto`：

```protobuf
syntax = "proto3";

package apiserver.v1;

option go_package = "bingo/pkg/proto/apiserver/v1;v1";

// 导入 HTTP 注解
import "google/api/annotations.proto";

// 导入 OpenAPI 注解（可选，用于生成 Swagger）
import "protoc-gen-openapiv2/options/annotations.proto";

// 定义服务
service ApiServer {
  // 用户登录
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (google.api.http) = {
      post: "/v1/auth/login"
      body: "*"
    };
  }

  // 用户注册
  rpc Register(RegisterRequest) returns (RegisterResponse) {
    option (google.api.http) = {
      post: "/v1/auth/register"
      body: "*"
    };
  }

  // 获取用户信息（需要认证）
  rpc GetUserInfo(GetUserInfoRequest) returns (UserInfoResponse) {
    option (google.api.http) = {
      get: "/v1/auth/user-info"
    };
  }
}

// 请求消息
message LoginRequest {
  string username = 1;
  string password = 2;
}

// 响应消息
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

### 2.2 HTTP 映射规则

`google.api.http` 注解支持以下映射方式：

```protobuf
// GET 请求，路径参数
rpc GetPost(GetPostRequest) returns (Post) {
  option (google.api.http) = {
    get: "/v1/posts/{post_id}"
  };
}
message GetPostRequest {
  string post_id = 1;  // 自动从 URL 路径提取
}

// POST 请求，请求体
rpc CreatePost(CreatePostRequest) returns (Post) {
  option (google.api.http) = {
    post: "/v1/posts"
    body: "*"  // 整个请求体映射到消息
  };
}

// PUT 请求，路径参数 + 请求体
rpc UpdatePost(UpdatePostRequest) returns (Post) {
  option (google.api.http) = {
    put: "/v1/posts/{post_id}"
    body: "post"  // 只有 post 字段从请求体获取
  };
}
message UpdatePostRequest {
  string post_id = 1;  // 从 URL 路径
  Post post = 2;       // 从请求体
}

// DELETE 请求
rpc DeletePost(DeletePostRequest) returns (Empty) {
  option (google.api.http) = {
    delete: "/v1/posts/{post_id}"
  };
}

// Query 参数（GET 请求的非路径字段自动变成 query 参数）
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

### 2.3 添加 OpenAPI 注解（可选）

为生成更完善的 Swagger 文档：

```protobuf
import "protoc-gen-openapiv2/options/annotations.proto";

// 文件级别的 OpenAPI 配置
option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_swagger) = {
  info: {
    title: "Bingo API";
    version: "1.0";
    description: "Bingo 脚手架 API 文档";
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
        description: "Bearer token 认证";
      };
    };
  };
};

// 方法级别注解
rpc Login(LoginRequest) returns (LoginResponse) {
  option (google.api.http) = {
    post: "/v1/auth/login"
    body: "*"
  };
  option (grpc.gateway.protoc_gen_openapiv2.options.openapiv2_operation) = {
    summary: "用户登录";
    description: "使用用户名密码登录，返回 JWT Token";
    tags: "认证";
  };
}
```

## 第三阶段：生成代码

### 3.1 更新 Makefile

修改 `scripts/make-rules/generate.mk`：

```makefile
APIROOT := $(ROOT_DIR)/pkg/proto
THIRD_PARTY := $(ROOT_DIR)/third_party
OPENAPI_DIR := $(ROOT_DIR)/api/openapi

.PHONY: gen.protoc
gen.protoc: ## 编译 protobuf 文件（包含 gRPC-Gateway）
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
gen.gateway: ## 仅生成 gRPC-Gateway 代码
	@echo "===========> Generate gRPC-Gateway files"
	@protoc \
		--proto_path=$(APIROOT) \
		--proto_path=$(THIRD_PARTY) \
		--grpc-gateway_out=paths=source_relative:$(APIROOT) \
		$(shell find $(APIROOT) -name "*.proto")

.PHONY: gen.openapi
gen.openapi: ## 仅生成 OpenAPI 文档
	@echo "===========> Generate OpenAPI files"
	@mkdir -p $(OPENAPI_DIR)
	@protoc \
		--proto_path=$(APIROOT) \
		--proto_path=$(THIRD_PARTY) \
		--openapiv2_out=$(OPENAPI_DIR) \
		--openapiv2_opt=logtostderr=true \
		$(shell find $(APIROOT) -name "*.proto")
```

### 3.2 更新 tools.mk

添加新工具的安装命令：

```makefile
.PHONY: install.grpc-gateway
install.grpc-gateway:
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	@$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### 3.3 执行生成

```bash
make gen.protoc
```

生成的文件：
- `apiserver.pb.go` - protobuf 消息定义
- `apiserver_grpc.pb.go` - gRPC 服务端/客户端代码
- `apiserver.pb.gw.go` - **gRPC-Gateway 反向代理代码**
- `api/openapi/apiserver.swagger.json` - OpenAPI 文档

## 第四阶段：实现服务

### 4.1 实现 gRPC Handler

创建 `internal/apiserver/handler/grpc/auth.go`：

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
	// 调用已有的 Biz 层
	resp, err := h.b.Auth().Login(ctx, &v1.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err // 会被转换为 gRPC status error
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

### 4.2 创建 gRPC-Gateway 服务器

创建 `internal/pkg/server/grpc_gateway.go`：

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
	// 连接到 gRPC 服务器
	conn, err := grpc.NewClient(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	// 创建 Gateway 多路复用器
	gwmux := runtime.NewServeMux(
		// JSON 序列化选项
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseEnumNumbers:  true,  // 枚举用数字
				EmitUnpopulated: false, // 不输出零值字段
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true, // 忽略未知字段
			},
		}),
		// 自定义错误处理（见下一节）
		runtime.WithErrorHandler(customErrorHandler),
	)

	// 注册处理器
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

### 4.3 自定义错误处理

```go
package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"bingo/pkg/errorsx"
)

// ErrResponse 统一错误响应格式（与 errorsx.ErrorX 对齐）
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
	// 使用 errorsx.FromError 从 gRPC error 解析 ErrorX
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

### 4.4 启动服务

修改 `internal/apiserver/run.go`：

```go
package apiserver

import (
	"context"

	pb "bingo/pkg/proto/apiserver/v1"
)

func run() error {
	// ... 初始化 store, biz 等

	// 1. 启动 gRPC 服务器
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			// 拦截器...
		),
	)
	pb.RegisterApiServerServer(grpcServer, handler.NewHandler(bizInstance))

	go func() {
		lis, _ := net.Listen("tcp", ":9090")
		grpcServer.Serve(lis)
	}()

	// 2. 启动 gRPC-Gateway（HTTP 反向代理）
	gwServer, err := server.NewGRPCGatewayServer(
		":8080",  // HTTP 端口
		":9090",  // gRPC 端口
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

## 第五阶段：验证

### 5.1 测试 gRPC 接口

```bash
# 使用 grpcurl
grpcurl -plaintext -d '{"username":"test","password":"123456"}' \
  localhost:9090 apiserver.v1.ApiServer/Login
```

### 5.2 测试 HTTP 接口

```bash
# 使用 curl
curl -X POST http://localhost:8080/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

### 5.3 查看 Swagger 文档

生成的 `api/openapi/apiserver.swagger.json` 可以：
- 导入 Swagger UI
- 导入 Postman
- 用 Swagger Editor 在线查看

## 常见问题

### Q1: 如何处理文件上传？

gRPC-Gateway 不适合处理 multipart/form-data。建议混合使用：

```go
mux := http.NewServeMux()
mux.Handle("/", gwmux)                           // 大部分 API
mux.Handle("/v1/file/upload", ginFileHandler)    // 文件上传走 Gin
```

### Q2: 如何传递自定义 Header？

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

### Q3: 如何获取客户端 IP？

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

## 相关文档

- [可插拔协议层](protocol-layer.md) - HTTP/gRPC/WebSocket 统一架构
- [WebSocket 设计与实现](websocket.md) - JSON-RPC 2.0 消息格式、中间件架构

---

**下一步**：了解 [统一错误处理](unified-error-handling.md)，学习如何让 HTTP/gRPC/WebSocket 三协议共享相同的错误格式。
