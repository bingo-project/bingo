# 统一 HTTP/gRPC/WebSocket 错误处理

本文档介绍如何设计一套统一的错误处理机制，让 HTTP、gRPC 和 WebSocket 三种协议共享相同的错误定义和响应格式。

## 设计目标

1. **单一错误定义** - 业务错误只定义一次，三种协议共享
2. **格式一致** - 客户端收到的错误格式统一
3. **语义清晰** - 错误码有层次结构，便于分类和处理
4. **可扩展** - 支持动态消息、元数据等高级特性

## 核心架构

```
┌─────────────────────────────────────────────────────────────┐
│                     错误定义层 (errno)                        │
│  ErrUserNotFound, ErrTokenInvalid, ErrPermissionDenied...  │
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
             │                    统一响应格式                              │
             │  {"code": 401, "reason": "xxx", "message": "xxx"}         │
             └───────────────────────────────────────────────────────────┘
```

## 第一阶段：使用 bingo/pkg/errorsx

我们直接使用 `bingo/pkg/errorsx` 提供的 ErrorX 类型，设计简洁：

### 1.1 ErrorX 结构体

```go
// bingo/pkg/errorsx

// ErrorX 定义了统一的错误类型
type ErrorX struct {
	// Code 表示 HTTP 状态码，自动转换为对应的 gRPC 状态码
	Code int `json:"code,omitempty"`

	// Reason 表示业务错误码，层次化命名: "Category.SubCategory.ErrorName"
	Reason string `json:"reason,omitempty"`

	// Message 表示简短的错误信息
	Message string `json:"message,omitempty"`

	// Metadata 用于存储额外的上下文信息
	Metadata map[string]string `json:"metadata,omitempty"`
}
```

**设计优势：**
- 只需指定 HTTP 状态码，自动转换为 gRPC 状态码
- 通过 `GRPCStatus()` 方法获取 gRPC status（带 ErrorInfo details）
- 支持 `WithMessage()`、`KV()` 等链式调用

### 1.2 核心方法

```go
// New 创建新错误
func New(code int, reason string, format string, args ...any) *ErrorX

// WithMessage 设置错误消息
func (err *ErrorX) WithMessage(format string, args ...any) *ErrorX

// KV 添加元数据
func (err *ErrorX) KV(kvs ...string) *ErrorX

// GRPCStatus 返回 gRPC status（带 ErrorInfo details）
func (err *ErrorX) GRPCStatus() *status.Status

// FromError 从任意 error 转换为 ErrorX
func FromError(err error) *ErrorX

// Is 判断错误类型
func (err *ErrorX) Is(target error) bool
```

### 1.3 HTTP ↔ gRPC 状态码映射

`errorsx` 使用 kratos 的 `httpstatus` 包自动处理映射：

| HTTP 状态码 | gRPC 状态码 |
|------------|------------|
| 200 OK | OK |
| 400 Bad Request | InvalidArgument |
| 401 Unauthorized | Unauthenticated |
| 403 Forbidden | PermissionDenied |
| 404 Not Found | NotFound |
| 409 Conflict | AlreadyExists |
| 429 Too Many Requests | ResourceExhausted |
| 500 Internal Server Error | Internal |
| 503 Service Unavailable | Unavailable |

## 第二阶段：定义业务错误码

### 2.1 通用错误码

修改 `internal/pkg/errno/code.go`，复用 `errorsx` 预定义错误：

```go
package errno

import (
	"net/http"

	"bingo/pkg/errorsx"
)

var (
	// OK 代表请求成功
	OK = &errorsx.ErrorX{Code: http.StatusOK, Message: ""}

	// 通用错误（复用 errorsx 预定义）
	ErrInternal         = errorsx.ErrInternal
	ErrNotFound         = errorsx.ErrNotFound
	ErrBind             = errorsx.ErrBind
	ErrInvalidArgument  = errorsx.ErrInvalidArgument
	ErrUnauthenticated  = errorsx.ErrUnauthenticated
	ErrPermissionDenied = errorsx.ErrPermissionDenied
	ErrOperationFailed  = errorsx.ErrOperationFailed

	// 认证错误
	ErrSignToken    = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.SignToken", Message: "Error occurred while signing the JSON web token."}
	ErrTokenInvalid = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.TokenInvalid", Message: "Token was invalid."}
	ErrTokenExpired = &errorsx.ErrorX{Code: http.StatusUnauthorized, Reason: "Unauthenticated.TokenExpired", Message: "Token has expired."}

	// 服务错误
	ErrServiceUnavailable = &errorsx.ErrorX{Code: http.StatusServiceUnavailable, Reason: "ServiceUnavailable", Message: "Service unavailable."}
	ErrTooManyRequests    = &errorsx.ErrorX{Code: http.StatusTooManyRequests, Reason: "TooManyRequests", Message: "Too many requests."}
)
```

### 2.2 领域错误码

创建 `internal/pkg/errno/user.go`：

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

创建 `internal/pkg/errno/app.go`：

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

## 第三阶段：协议适配层

### 3.1 HTTP 响应处理

修改 `internal/pkg/core/response.go`：

```go
package core

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"bingo/pkg/errorsx"

	"bingo/internal/pkg/errno"
)

// ErrResponse 统一错误响应格式（与 errorsx.ErrorX 对齐）
type ErrResponse struct {
	Code     int               `json:"code"`
	Reason   string            `json:"reason"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// WriteResponse 写入 HTTP 响应
func WriteResponse(c *gin.Context, err error, data interface{}) {
	if err == nil {
		c.JSON(http.StatusOK, data)
		return
	}

	// 转换为 ErrorX
	e := errorsx.FromError(err)

	c.JSON(e.Code, ErrResponse{
		Code:     e.Code,
		Reason:   e.Reason,
		Message:  e.Message,
		Metadata: e.Metadata,
	})
}

// HandleJSONRequest 通用 JSON 请求处理器
func HandleJSONRequest[Req, Resp any](
	c *gin.Context,
	handler func(ctx context.Context, req *Req) (*Resp, error),
	validators ...func(*Req) error,
) {
	var req Req

	// 绑定请求体
	if err := c.ShouldBindJSON(&req); err != nil {
		WriteResponse(c, errno.ErrBind.WithMessage(err.Error()), nil)
		return
	}

	// 运行验证器
	for _, validate := range validators {
		if err := validate(&req); err != nil {
			WriteResponse(c, err, nil)
			return
		}
	}

	// 调用处理函数
	resp, err := handler(c.Request.Context(), &req)
	WriteResponse(c, err, resp)
}
```

### 3.2 gRPC 错误处理

在 gRPC Handler 中，使用 `GRPCStatus().Err()` 返回错误：

```go
// internal/apiserver/handler/grpc/auth.go

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	resp, err := h.biz.Auth().Login(ctx, bizReq)
	if err != nil {
		// ErrorX 自动转换为 gRPC status（带 ErrorInfo details）
		return nil, errorsx.FromError(err).GRPCStatus().Err()
	}
	return resp, nil
}
```

gRPC-Gateway 自定义错误处理器：

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
	// 从 gRPC error 解析 ErrorX
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

### 3.3 WebSocket 错误处理

参考 [WebSocket 统一方案](./websocket-unification.md) 中的消息格式：

```go
// pkg/ws/message/message.go

package message

import (
	"bingo/pkg/errorsx"
)

// Response WebSocket 响应消息
type Response struct {
	Seq   string      `json:"seq"`
	Cmd   string      `json:"cmd"`
	Data  interface{} `json:"data,omitempty"`
	Error *Error      `json:"error,omitempty"`
}

// Error 错误格式（与 HTTP 完全一致）
type Error struct {
	Code     int               `json:"code"`
	Reason   string            `json:"reason"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewResponse 创建成功响应
func NewResponse(seq, cmd string, data interface{}) *Response {
	return &Response{Seq: seq, Cmd: cmd, Data: data}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(seq, cmd string, err error) *Response {
	e := errorsx.FromError(err)
	return &Response{
		Seq: seq,
		Cmd: cmd,
		Error: &Error{
			Code:     e.Code,
			Reason:   e.Reason,
			Message:  e.Message,
			Metadata: e.Metadata,
		},
	}
}
```

## 第四阶段：Biz 层错误使用

### 4.1 返回预定义错误

```go
func (b *userBiz) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginResponse, error) {
	user, err := b.store.User().GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, errno.ErrUserNotFound  // 直接返回预定义错误
	}

	if err := auth.Compare(user.Password, req.Password); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// ...
}
```

### 4.2 动态消息

```go
func (b *userBiz) Create(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	// 使用 WithMessage 动态设置错误消息
	if len(req.Username) < 3 {
		return nil, errno.ErrUsernameInvalid.WithMessage("Username must be at least 3 characters, got %d", len(req.Username))
	}

	// ...
}
```

### 4.3 添加元数据

```go
func (b *userBiz) Get(ctx context.Context, uid string) (*v1.User, error) {
	user, err := b.store.User().GetByUID(ctx, uid)
	if err != nil {
		// 使用 KV 添加调试信息
		return nil, errno.ErrUserNotFound.KV("uid", uid, "source", "biz.user.Get")
	}
	return user, nil
}
```

### 4.4 错误判断

```go
import "errors"

func handleError(err error) {
	// 使用 errors.Is 判断错误类型
	if errors.Is(err, errno.ErrUserNotFound) {
		// 用户不存在的特殊处理
	}

	// 或使用 ErrorX.Is 方法
	if errno.ErrUnauthenticated.Is(err) {
		// 未授权的特殊处理
	}
}
```

## 错误码命名规范

### 层次化命名

```
Category.SubCategory.ErrorName
```

**Category（一级分类）：**
- `InvalidParameter` - 参数错误 (400)
- `AuthFailure` - 认证失败 (401)
- `Forbidden` - 授权失败 (403)
- `ResourceNotFound` - 资源不存在 (404)
- `ResourceAlreadyExists` - 资源已存在 (409)
- `TooManyRequests` - 请求过多 (429)
- `InternalError` - 内部错误 (500)
- `ServiceUnavailable` - 服务不可用 (503)

**示例：**
```
AuthFailure.TokenInvalid       - Token 无效
AuthFailure.TokenExpired       - Token 过期
AuthFailure.PasswordIncorrect  - 密码错误

ResourceNotFound.UserNotFound  - 用户不存在
ResourceNotFound.AppNotFound   - 应用不存在

InvalidParameter.BindError     - 参数绑定错误
InvalidParameter.UsernameInvalid - 用户名格式错误
```

## 完整流程示例

### HTTP 请求（Gin）

```
POST /v1/auth/login
{"username": "test", "password": "wrong"}

↓ Gin Handler
↓ Biz.Login() returns errno.ErrPasswordIncorrect
↓ core.WriteResponse()

HTTP 401
{"code": 401, "reason": "Unauthenticated.PasswordIncorrect", "message": "Password is incorrect."}
```

### gRPC 请求

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
POST /v1/auth/login (走 gRPC-Gateway)
{"username": "test", "password": "wrong"}

↓ gRPC-Gateway 转发到 gRPC
↓ gRPC Handler 返回 status with ErrorInfo
↓ CustomErrorHandler 解析 ErrorInfo

HTTP 401
{"code": 401, "reason": "Unauthenticated.PasswordIncorrect", "message": "Password is incorrect."}
```

### WebSocket 请求

```json
{"seq": "1", "cmd": "Login", "data": {"username": "test", "password": "wrong"}}

↓ WS Adapter
↓ Biz.Login() returns errno.ErrPasswordIncorrect
↓ message.NewErrorResponse()

{"seq": "1", "cmd": "Login", "error": {"code": 401, "reason": "Unauthenticated.PasswordIncorrect", "message": "Password is incorrect."}}
```

## 总结

使用 `bingo/pkg/errorsx` 的优势：

1. **简洁设计** - 只需指定 HTTP 状态码，自动转换为 gRPC 状态码
2. **生产就绪** - 经过实践验证
3. **gRPC 友好** - `GRPCStatus()` 方法自动添加 ErrorInfo details
4. **链式调用** - 支持 `WithMessage()`、`KV()` 等方法

## 相关文档

- [可插拔协议层](protocol-layer.md) - HTTP/gRPC/WebSocket 统一架构
- [WebSocket 设计与实现](websocket.md) - JSON-RPC 2.0 消息格式、中间件架构
- [认证中间件迁移](auth-middleware-migration.md) - 统一认证实现

---

**下一步**：了解 [微服务拆分](microservices.md)，学习如何将单体应用演进为微服务架构。
