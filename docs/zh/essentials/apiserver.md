---
title: API Server 详解 - Bingo Go 微服务核心服务
description: 深入了解 Bingo API Server 的架构设计，包括 HTTP RESTful API、gRPC 服务间通信、认证授权、性能优化等核心功能。
---

# API Server 详解

API Server 是 Bingo 微服务架构的核心服务，负责对外提供用户级别的 API 接口。它支持 HTTP、gRPC 和 WebSocket 三种通信协议，可以满足不同场景的业务需求。

## 服务概览

### 端口配置

| 协议 | 端口 | 用途 | 状态 |
|------|------|------|------|
| HTTP | 8080 | RESTful API 接口 | 稳定 |
| gRPC | 8081 | 服务间通信 | 稳定 |
| WebSocket | 8082 | 实时通信 | 开发中 |

### 核心特性

- **高并发**：采用 Go 的协程特性，支持数万级并发
- **无状态设计**：便于水平扩展，支持多实例部署
- **分布式缓存**：利用 Redis 进行缓存，提升性能
- **完整权限管理**：使用 Casbin 实现 RBAC 权限模型
- **请求监控**：详细的日志记录和性能指标

## HTTP API

HTTP API 采用 RESTful 设计，是客户端访问的主要入口。

### API 端点设计

```
/api/v1/{resource}         - 资源操作
/api/v1/{resource}/{id}    - 单个资源操作
/api/v1/{resource}/search  - 资源搜索
```

### 基本操作

| 方法 | 操作 | 示例 |
|------|------|------|
| GET | 查询资源 | `GET /api/v1/users/123` |
| POST | 创建资源 | `POST /api/v1/users` |
| PUT | 完全更新 | `PUT /api/v1/users/123` |
| PATCH | 部分更新 | `PATCH /api/v1/users/123` |
| DELETE | 删除资源 | `DELETE /api/v1/users/123` |

### 认证方式

API Server 支持多种认证方式：

#### 1. JWT 令牌认证

```bash
# 登录获取令牌
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'

# 响应
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}

# 使用令牌访问 API
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

#### 2. Session 认证

```bash
# Cookie 自动包含，无需手动设置
curl -X GET http://localhost:8080/api/v1/users \
  -b "session=xxx"
```

### 响应格式

```json
{
  "code": 0,           // 业务状态码
  "message": "success",
  "data": {
    "id": 1,
    "name": "John",
    "email": "john@example.com"
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### 错误响应

```json
{
  "code": 400,
  "message": "Invalid request parameters",
  "errors": {
    "email": "Invalid email format"
  }
}
```

### 分页

```bash
# 查询第 2 页，每页 20 条
GET /api/v1/users?page=2&page_size=20&sort=-created_at

# 响应包含分页信息
{
  "code": 0,
  "data": [...],
  "pagination": {
    "page": 2,
    "page_size": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## gRPC 服务

gRPC 用于服务间的高性能通信，基于 HTTP/2 和 Protocol Buffers。

### 服务清单

| 服务 | 用途 | 典型调用方 |
|------|------|----------|
| `UserService` | 用户相关操作 | Admin Server、Bot Service |
| `AuthService` | 认证服务 | 所有服务 |
| `PostService` | 内容相关操作 | Admin Server、Scheduler |
| `CommentService` | 评论相关操作 | Admin Server |

### 定义示例

Protocol Buffer 定义位于 `api/pb/apiserver/v1/` 目录。

```protobuf
syntax = "proto3";

package apiserver.v1;

service UserService {
  rpc GetUser(GetUserRequest) returns (UserResponse);
  rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
  rpc CreateUser(CreateUserRequest) returns (UserResponse);
  rpc UpdateUser(UpdateUserRequest) returns (UserResponse);
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse);
}

message UserResponse {
  int64 id = 1;
  string username = 2;
  string email = 3;
  int64 created_at = 4;
}
```

### Go 客户端使用

```go
package main

import (
  "context"
  "log"

  pb "yourmodule/api/pb/apiserver/v1"
  "google.golang.org/grpc"
)

func main() {
  // 建立连接
  conn, err := grpc.Dial(
    "localhost:8081",
    grpc.WithInsecure(),
  )
  if err != nil {
    log.Fatal(err)
  }
  defer conn.Close()

  // 创建客户端
  client := pb.NewUserServiceClient(conn)

  // 调用方法
  resp, err := client.GetUser(context.Background(), &pb.GetUserRequest{
    Id: 123,
  })
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("User: %v", resp)
}
```

### 服务间通信架构

```
┌─────────────┐                  ┌──────────────┐
│ Admin       │ ──gRPC──────────▶ │ API Server   │
│ Server      │ ◀───────gRPC────── │ (8081)       │
└─────────────┘                  └──────────────┘

┌─────────────┐
│ Scheduler   │ ──gRPC──────────▶ │ API Server   │
│             │ ◀───────gRPC────── │ (8081)       │
└─────────────┘                  └──────────────┘
```

## WebSocket 实时通信

⚠️ **状态**：该功能仍在开发阶段，预计在下一个版本发布。

### 设计目标

- 支持实时消息推送
- 支持多客户端连接
- 支持事件订阅和广播
- 自动重连机制
- 连接生命周期管理

### 预期端点

```
ws://localhost:8082/ws/chat      # 聊天消息
ws://localhost:8082/ws/notify    # 系统通知
ws://localhost:8082/ws/stream    # 数据流
```

### 计划特性

#### 1. 连接认证

```javascript
// 客户端连接时携带 token
const ws = new WebSocket(
  'ws://localhost:8082/ws/chat?token=' + token
);
```

#### 2. 消息格式

```json
{
  "type": "message",
  "event": "user.message",
  "data": {
    "user_id": 123,
    "content": "Hello, World!",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

#### 3. 事件订阅

```javascript
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  switch(msg.type) {
    case 'subscribe_success':
      console.log('Subscribed to:', msg.event);
      break;
    case 'message':
      console.log('Received:', msg.data);
      break;
  }
};
```

#### 4. 心跳保活

服务器每 30 秒发送 ping 消息，客户端需回复 pong：

```javascript
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  if (msg.type === 'ping') {
    ws.send(JSON.stringify({ type: 'pong' }));
  }
};
```

### 开发进度

- [ ] 连接管理和认证
- [ ] 消息路由和分发
- [ ] 事件订阅系统
- [ ] 连接池管理
- [ ] 客户端库（JavaScript/TypeScript）
- [ ] 单元测试和集成测试
- [ ] 性能优化

## 配置说明

### 环境变量

```bash
# HTTP 服务
API_SERVER_HTTP_PORT=8080
API_SERVER_HTTP_HOST=0.0.0.0

# gRPC 服务
API_SERVER_GRPC_PORT=8081
API_SERVER_GRPC_HOST=0.0.0.0

# WebSocket 服务（开发中）
API_SERVER_WEBSOCKET_PORT=8082
API_SERVER_WEBSOCKET_HOST=0.0.0.0

# 认证
JWT_SECRET=your_secret_key
JWT_EXPIRE=86400

# CORS 跨域
CORS_ORIGINS=http://localhost:3000,https://example.com
```

### Docker 启动

```bash
docker run -d \
  -p 8080:8080 \
  -p 8081:8081 \
  -p 8082:8082 \
  -e JWT_SECRET=secret \
  -e DATABASE_URL=postgres://user:pass@db:5432/bingo \
  -e REDIS_URL=redis://cache:6379 \
  bingo-apiserver:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  apiserver:
    image: bingo-apiserver:latest
    ports:
      - "8080:8080"
      - "8081:8081"
      - "8082:8082"
    environment:
      - JWT_SECRET=secret
      - DATABASE_URL=postgres://db:5432/bingo
      - REDIS_URL=redis://cache:6379
    depends_on:
      - db
      - cache

  db:
    image: postgres:15
    environment:
      POSTGRES_DB: bingo
      POSTGRES_PASSWORD: pass

  cache:
    image: redis:7-alpine
```

## 性能优化

### 1. 缓存策略

API Server 使用多层缓存：

```
客户端缓存 (HTTP Cache-Control)
    ↓
CDN 缓存 (可选)
    ↓
Redis 缓存 (热数据)
    ↓
数据库查询
```

### 2. 连接池

```go
// gRPC 连接复用
conn, _ := grpc.Dial("localhost:8081")
// 复用单个连接，多个 RPC 调用共享

// 数据库连接池
db.SetMaxOpenConns(25)  // 最大连接数
db.SetMaxIdleConns(5)   // 最小连接数
```

### 3. 请求限流

```bash
# 使用中间件限流
# 每个 IP 每分钟限制 100 请求
rate_limit:
  enabled: true
  requests_per_minute: 100
  by_ip: true
```

## 监控指标

API Server 暴露以下 Prometheus 指标：

```
# HTTP 请求
api_server_http_requests_total{method,status,endpoint}
api_server_http_request_duration_seconds{endpoint}
api_server_http_request_size_bytes{endpoint}
api_server_http_response_size_bytes{endpoint}

# gRPC 请求
api_server_grpc_requests_total{service,method,status}
api_server_grpc_request_duration_seconds{service,method}

# 连接状态
api_server_active_connections{type}  # http, grpc, websocket
api_server_connection_errors_total{type,reason}
```

## 常见问题

### Q: 如何处理并发请求？

A: API Server 基于 Go 的协程，自动处理并发。每个请求创建一个新的协程，通过 GOMAXPROCS 环境变量控制线程数。

```bash
export GOMAXPROCS=8  # 使用 8 个处理器
```

### Q: 如何扩展 API Server？

A: 采用无状态设计，可通过以下方式扩展：

```
┌─────────────┐
│   Nginx     │  负载均衡
└──────┬──────┘
       │
   ┌───┴────┬──────┬──────┐
   ▼        ▼      ▼      ▼
API-1    API-2  API-3  API-N
   │        │      │      │
   └────────┴──────┴──────┘
         MySQL + Redis
```

每个实例共享同一个数据库和缓存，可以独立添加或移除。

### Q: WebSocket 何时发布？

A: WebSocket 功能仍在开发中。我们计划在 v1.2.0 版本发布，敬请期待。

## 下一步

- [Scheduler 调度器](./scheduler.md) - 了解定时任务和异步队列服务
