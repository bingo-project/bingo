# 认证中间件迁移指南

本文档介绍如何将认证逻辑从 HTTP Middleware 迁移到 gRPC Interceptor，实现 HTTP/gRPC/WebSocket 三种协议共享同一套认证机制。

## 设计目标

1. **单一认证逻辑** - 认证代码只写一次
2. **协议无关** - Biz 层不感知是 HTTP、gRPC 还是 WebSocket
3. **灵活的白名单** - 支持按方法配置免认证接口
4. **用户信息传递** - 通过 Context 传递认证信息

## 架构对比

### 当前架构（HTTP 独立认证）

```
HTTP 请求
    ↓
Gin Middleware (authn.go)  ← HTTP 专用
    ↓
Controller
    ↓
Biz 层

gRPC 请求
    ↓
gRPC Interceptor (另一套) ← gRPC 专用
    ↓
gRPC Handler
    ↓
Biz 层
```

### 目标架构（gRPC 统一认证）

```
HTTP 请求                    gRPC 请求              WebSocket 请求
    ↓                           ↓                       ↓
gRPC-Gateway                 gRPC Server             WS Handler
    ↓                           ↓                       ↓
    └───────────────────────────┼───────────────────────┘
                                ↓
                    gRPC Interceptor (统一认证)
                                ↓
                          gRPC Handler
                                ↓
                            Biz 层
```

## 第一阶段：创建统一的 Context 工具

### 1.1 Context Key 定义

创建 `internal/pkg/contextx/contextx.go`：

```go
package contextx

import (
	"context"
)

// 定义 context key 类型，避免冲突
type contextKey int

const (
	userIDKey contextKey = iota
	usernameKey
	requestIDKey
	clientIPKey
)

// WithUserID 设置用户 ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserID 获取用户 ID
func UserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

// WithUsername 设置用户名
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

// Username 获取用户名
func Username(ctx context.Context) string {
	if v, ok := ctx.Value(usernameKey).(string); ok {
		return v
	}
	return ""
}

// WithRequestID 设置请求 ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestID 获取请求 ID
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// WithClientIP 设置客户端 IP
func WithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey, ip)
}

// ClientIP 获取客户端 IP
func ClientIP(ctx context.Context) string {
	if v, ok := ctx.Value(clientIPKey).(string); ok {
		return v
	}
	return ""
}
```

### 1.2 用户信息模型

创建 `internal/pkg/contextx/user.go`：

```go
package contextx

import "context"

// UserInfo 认证后的用户信息
type UserInfo struct {
	UID      string
	Username string
	Nickname string
	Email    string
	Status   int
}

const userInfoKey contextKey = 100

// WithUserInfo 设置用户信息
func WithUserInfo(ctx context.Context, info *UserInfo) context.Context {
	return context.WithValue(ctx, userInfoKey, info)
}

// GetUserInfo 获取用户信息
func GetUserInfo(ctx context.Context) *UserInfo {
	if v, ok := ctx.Value(userInfoKey).(*UserInfo); ok {
		return v
	}
	return nil
}
```

## 第二阶段：实现 gRPC 认证拦截器

### 2.1 UserRetriever 接口

定义获取用户信息的接口，解耦认证逻辑和存储层：

```go
// internal/pkg/middleware/grpc/authn.go

package grpc

import (
	"context"

	"bingo/internal/pkg/contextx"
)

// UserRetriever 用户信息获取接口
type UserRetriever interface {
	GetUser(ctx context.Context, userID string) (*contextx.UserInfo, error)
}
```

### 2.2 认证拦截器

```go
// internal/pkg/middleware/grpc/authn.go

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"bingo/internal/pkg/contextx"
	"bingo/internal/pkg/errno"
	"bingo/pkg/token"
)

// AuthnInterceptor 创建认证拦截器
func AuthnInterceptor(retriever UserRetriever) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 1. 从 metadata 中提取 token
		tokenStr, err := extractToken(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// 2. 解析 token
		claims, err := token.Parse(tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, errno.ErrTokenInvalid.Message)
		}

		// 3. 获取用户信息
		userInfo, err := retriever.GetUser(ctx, claims.Subject)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, errno.ErrUserNotFound.Message)
		}

		// 4. 将用户信息注入 context
		ctx = contextx.WithUserID(ctx, userInfo.UID)
		ctx = contextx.WithUsername(ctx, userInfo.Username)
		ctx = contextx.WithUserInfo(ctx, userInfo)

		// 5. 继续处理请求
		return handler(ctx, req)
	}
}

// extractToken 从 gRPC metadata 中提取 token
func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errno.ErrTokenInvalid
	}

	// 支持 "authorization" 和 "Authorization" 两种 header
	values := md.Get("authorization")
	if len(values) == 0 {
		values = md.Get("Authorization")
	}
	if len(values) == 0 {
		return "", errno.ErrTokenInvalid
	}

	// 提取 Bearer token
	auth := values[0]
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", errno.ErrTokenInvalid
	}

	return strings.TrimPrefix(auth, prefix), nil
}
```

### 2.3 白名单机制

使用 `selector` 包实现按方法过滤：

```go
// internal/apiserver/grpcserver.go

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"

	mw "bingo/internal/pkg/middleware/grpc"
	pb "bingo/pkg/proto/apiserver/v1"
)

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(retriever mw.UserRetriever) *grpc.Server {
	// 配置拦截器链
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			// 请求 ID 拦截器（所有请求都执行）
			mw.RequestIDInterceptor(),

			// 认证拦截器（带白名单过滤）
			selector.UnaryServerInterceptor(
				mw.AuthnInterceptor(retriever),
				NewAuthnWhiteListMatcher(),
			),

			// 授权拦截器（带白名单过滤）
			selector.UnaryServerInterceptor(
				mw.AuthzInterceptor(authz),
				NewAuthzWhiteListMatcher(),
			),

			// 请求验证拦截器
			mw.ValidatorInterceptor(),
		),
	}

	srv := grpc.NewServer(opts...)
	pb.RegisterApiServerServer(srv, handler.NewHandler(bizInstance))
	reflection.Register(srv)

	return srv
}

// NewAuthnWhiteListMatcher 创建认证白名单
func NewAuthnWhiteListMatcher() selector.Matcher {
	// 不需要认证的方法
	whitelist := map[string]struct{}{
		pb.ApiServer_Healthz_FullMethodName:  {},
		pb.ApiServer_Version_FullMethodName:  {},
		pb.ApiServer_Login_FullMethodName:    {},
		pb.ApiServer_Register_FullMethodName: {},
	}

	return selector.MatchFunc(func(ctx context.Context, call interceptors.CallMeta) bool {
		// 返回 true 表示需要认证，false 表示跳过
		_, skip := whitelist[call.FullMethod()]
		return !skip
	})
}

// NewAuthzWhiteListMatcher 创建授权白名单
func NewAuthzWhiteListMatcher() selector.Matcher {
	whitelist := map[string]struct{}{
		pb.ApiServer_Healthz_FullMethodName:  {},
		pb.ApiServer_Version_FullMethodName:  {},
		pb.ApiServer_Login_FullMethodName:    {},
		pb.ApiServer_Register_FullMethodName: {},
		pb.ApiServer_GetUserInfo_FullMethodName: {}, // 登录后可直接访问
	}

	return selector.MatchFunc(func(ctx context.Context, call interceptors.CallMeta) bool {
		_, skip := whitelist[call.FullMethod()]
		return !skip
	})
}
```

## 第三阶段：其他拦截器

### 3.1 请求 ID 拦截器

```go
// internal/pkg/middleware/grpc/requestid.go

package grpc

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"bingo/internal/pkg/contextx"
)

// RequestIDInterceptor 请求 ID 拦截器
func RequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// 尝试从 metadata 获取请求 ID
		requestID := ""
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-request-id"); len(values) > 0 {
				requestID = values[0]
			}
		}

		// 如果没有则生成新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 注入 context
		ctx = contextx.WithRequestID(ctx, requestID)

		return handler(ctx, req)
	}
}
```

### 3.2 客户端 IP 拦截器

```go
// internal/pkg/middleware/grpc/clientip.go

package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"bingo/internal/pkg/contextx"
)

// ClientIPInterceptor 客户端 IP 拦截器
func ClientIPInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		clientIP := extractClientIP(ctx)
		ctx = contextx.WithClientIP(ctx, clientIP)
		return handler(ctx, req)
	}
}

func extractClientIP(ctx context.Context) string {
	// 优先从 metadata 获取（经过反向代理时）
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// X-Real-IP（nginx）
		if values := md.Get("x-real-ip"); len(values) > 0 {
			return values[0]
		}
		// X-Forwarded-For（标准）
		if values := md.Get("x-forwarded-for"); len(values) > 0 {
			// 取第一个 IP
			ips := strings.Split(values[0], ",")
			return strings.TrimSpace(ips[0])
		}
	}

	// 从 peer 获取直连 IP
	if p, ok := peer.FromContext(ctx); ok {
		addr := p.Addr.String()
		// 移除端口号
		if idx := strings.LastIndex(addr, ":"); idx != -1 {
			return addr[:idx]
		}
		return addr
	}

	return ""
}
```

### 3.3 日志拦截器

```go
// internal/pkg/middleware/grpc/logger.go

package grpc

import (
	"context"
	"time"

	"github.com/bingo-project/component-base/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"bingo/internal/pkg/contextx"
)

// LoggerInterceptor 日志拦截器
func LoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// 执行请求
		resp, err := handler(ctx, req)

		// 记录日志
		duration := time.Since(start)
		code := status.Code(err)

		log.C(ctx).Infow("gRPC request",
			"method", info.FullMethod,
			"duration", duration,
			"code", code,
			"request_id", contextx.RequestID(ctx),
			"user_id", contextx.UserID(ctx),
			"client_ip", contextx.ClientIP(ctx),
		)

		return resp, err
	}
}
```

### 3.4 Recovery 拦截器

```go
// internal/pkg/middleware/grpc/recovery.go

package grpc

import (
	"context"
	"runtime/debug"

	"github.com/bingo-project/component-base/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor panic 恢复拦截器
func RecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				log.C(ctx).Errorw("gRPC panic recovered",
					"panic", r,
					"stack", string(debug.Stack()),
					"method", info.FullMethod,
				)
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()

		return handler(ctx, req)
	}
}
```

## 第四阶段：gRPC-Gateway 集成

### 4.1 HTTP Header 转发到 gRPC Metadata

```go
// internal/pkg/server/grpc_gateway.go

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
)

func NewGRPCGatewayServer(...) (*GRPCGatewayServer, error) {
	gwmux := runtime.NewServeMux(
		// 将 HTTP Header 转发到 gRPC Metadata
		runtime.WithIncomingHeaderMatcher(customHeaderMatcher),

		// 从 HTTP 请求提取 metadata（如 Client IP）
		runtime.WithMetadata(extractMetadataFromHTTP),

		// 自定义错误处理
		runtime.WithErrorHandler(customErrorHandler),
	)

	// ...
}

// customHeaderMatcher 选择需要转发的 HTTP Header
func customHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case "authorization":
		return "authorization", true
	case "x-request-id":
		return "x-request-id", true
	case "x-real-ip":
		return "x-real-ip", true
	case "x-forwarded-for":
		return "x-forwarded-for", true
	default:
		return "", false
	}
}

// extractMetadataFromHTTP 从 HTTP 请求提取元数据
func extractMetadataFromHTTP(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}

	// 提取真实 IP
	if ip := getRealIP(req); ip != "" {
		md.Set("x-real-ip", ip)
	}

	// 生成或转发 Request ID
	requestID := req.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	md.Set("x-request-id", requestID)

	return md
}

func getRealIP(req *http.Request) string {
	// X-Real-IP
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// X-Forwarded-For
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	// RemoteAddr
	if idx := strings.LastIndex(req.RemoteAddr, ":"); idx != -1 {
		return req.RemoteAddr[:idx]
	}
	return req.RemoteAddr
}
```

## 第五阶段：WebSocket 认证

WebSocket 需要在连接建立时认证，之后复用认证信息。

### 5.1 连接时认证

```go
// pkg/ws/server/client.go

func ServeWs(c *gin.Context) {
	// 1. 从 query 或 header 获取 token
	tokenStr := c.Query("token")
	if tokenStr == "" {
		tokenStr = extractBearerToken(c.GetHeader("Authorization"))
	}

	if tokenStr == "" {
		c.JSON(401, gin.H{"code": "AuthFailure.TokenInvalid", "message": "Token required"})
		return
	}

	// 2. 验证 token
	claims, err := token.Parse(tokenStr)
	if err != nil {
		c.JSON(401, gin.H{"code": "AuthFailure.TokenInvalid", "message": "Invalid token"})
		return
	}

	// 3. 获取用户信息
	userInfo, err := retriever.GetUser(c, claims.Subject)
	if err != nil {
		c.JSON(401, gin.H{"code": "AuthFailure.UserNotFound", "message": "User not found"})
		return
	}

	// 4. 升级 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 5. 创建带认证信息的 client
	ctx := contextx.WithUserID(context.Background(), userInfo.UID)
	ctx = contextx.WithUsername(ctx, userInfo.Username)
	ctx = contextx.WithUserInfo(ctx, userInfo)

	client := NewClient(conn, ctx)
	ClientManager.register <- client

	go client.writePump()
	go client.readPump()
}
```

### 5.2 消息处理时使用认证信息

```go
// pkg/ws/server/client.go

type Client struct {
	conn *websocket.Conn
	ctx  context.Context // 包含认证信息
	send chan []byte
}

func (c *Client) handleMessage(msg []byte) {
	// 从 client.ctx 获取用户信息
	userID := contextx.UserID(c.ctx)
	username := contextx.Username(c.ctx)

	log.Infow("WebSocket message received",
		"user_id", userID,
		"username", username,
		"message", string(msg),
	)

	// 调用 handler 时传递带认证信息的 context
	resp, err := handler(c.ctx, req.Data)
	// ...
}
```

## 第六阶段：Biz 层使用认证信息

Biz 层通过 `contextx` 包获取用户信息，完全不感知具体协议：

```go
// internal/apiserver/biz/user/user.go

func (b *userBiz) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	// 从 context 获取当前用户 ID
	userID := contextx.UserID(ctx)
	if userID == "" {
		return nil, errno.ErrUnauthorized
	}

	// 获取用户信息
	user, err := b.store.User().GetByUID(ctx, userID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	// 验证旧密码
	if err := auth.Compare(user.Password, req.OldPassword); err != nil {
		return nil, errno.ErrPasswordIncorrect
	}

	// 更新密码
	user.Password, _ = auth.Encrypt(req.NewPassword)
	if err := b.store.User().Update(ctx, user); err != nil {
		return nil, err
	}

	return &pb.ChangePasswordResponse{}, nil
}
```

## 完整拦截器链

```go
// internal/apiserver/grpcserver.go

func NewGRPCServer(retriever mw.UserRetriever, authz *auth.Authz) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			// 1. Recovery（最外层，捕获所有 panic）
			mw.RecoveryInterceptor(),

			// 2. 请求 ID（为每个请求生成唯一标识）
			mw.RequestIDInterceptor(),

			// 3. 客户端 IP（提取真实 IP）
			mw.ClientIPInterceptor(),

			// 4. 认证（带白名单）
			selector.UnaryServerInterceptor(
				mw.AuthnInterceptor(retriever),
				NewAuthnWhiteListMatcher(),
			),

			// 5. 授权（带白名单）
			selector.UnaryServerInterceptor(
				mw.AuthzInterceptor(authz),
				NewAuthzWhiteListMatcher(),
			),

			// 6. 参数验证
			mw.ValidatorInterceptor(),

			// 7. 日志（记录请求详情）
			mw.LoggerInterceptor(),
		),
	}

	srv := grpc.NewServer(opts...)
	// ...
	return srv
}
```

## 迁移清单

### 需要删除的文件

- `internal/apiserver/middleware/authn.go` (HTTP 认证中间件)
- `internal/admserver/middleware/authn.go` (重复代码)

### 需要创建的文件

- `internal/pkg/contextx/contextx.go`
- `internal/pkg/contextx/user.go`
- `internal/pkg/middleware/grpc/authn.go`
- `internal/pkg/middleware/grpc/requestid.go`
- `internal/pkg/middleware/grpc/clientip.go`
- `internal/pkg/middleware/grpc/logger.go`
- `internal/pkg/middleware/grpc/recovery.go`

### 需要修改的文件

- `internal/apiserver/grpcserver.go` - 添加拦截器链
- `internal/pkg/server/grpc_gateway.go` - 添加 Header 转发
- Biz 层 - 将 `gin.Context` 改为 `context.Context`

## 总结

迁移后的优势：

1. **代码复用** - 认证逻辑只写一次，HTTP/gRPC/WebSocket 共享
2. **一致性** - 三种协议的认证行为完全一致
3. **可维护性** - 修改认证逻辑只需改一处
4. **可测试性** - Biz 层只依赖 `context.Context`，易于单元测试
5. **灵活性** - 通过 selector 轻松配置白名单

## 相关文档

- [WebSocket 设计与实现](websocket.md) - WebSocket 完整设计方案
