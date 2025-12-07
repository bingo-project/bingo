---
title: API Server Guide - Bingo Go Core Microservice
description: Deep dive into Bingo API Server architecture, including HTTP RESTful API, gRPC inter-service communication, authentication, and performance optimization.
---

# API Server Guide

The API Server is the core service in Bingo's microservice architecture. It provides user-facing APIs through three communication protocols: HTTP, gRPC, and WebSocket, supporting diverse business scenarios.

## Service Overview

### Port Configuration

| Protocol | Port | Purpose | Status |
|----------|------|---------|--------|
| HTTP | 8080 | RESTful API endpoints | Stable |
| gRPC | 9090 | Inter-service communication | Stable |
| WebSocket | 8081 | Real-time communication | In Development |

### Core Features

- **High Concurrency**: Leverages Go's goroutines to handle tens of thousands of concurrent connections
- **Stateless Design**: Enables horizontal scaling with multiple instances
- **Distributed Caching**: Uses Redis for improved performance
- **Complete Permission Management**: Implements RBAC with Casbin
- **Request Monitoring**: Detailed logging and performance metrics

## HTTP API

The HTTP API follows RESTful design principles and serves as the primary entry point for client applications.

### Endpoint Design

```
/api/v1/{resource}         - Resource operations
/api/v1/{resource}/{id}    - Single resource operations
/api/v1/{resource}/search  - Resource search
```

### Basic Operations

| Method | Operation | Example |
|--------|-----------|---------|
| GET | Retrieve resource | `GET /api/v1/users/123` |
| POST | Create resource | `POST /api/v1/users` |
| PUT | Full update | `PUT /api/v1/users/123` |
| PATCH | Partial update | `PATCH /api/v1/users/123` |
| DELETE | Delete resource | `DELETE /api/v1/users/123` |

### Authentication

The API Server supports multiple authentication methods:

#### 1. JWT Token Authentication

```bash
# Login to get token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'

# Response
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}

# Use token to access API
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

#### 2. Session Authentication

```bash
# Cookie is automatically included
curl -X GET http://localhost:8080/api/v1/users \
  -b "session=xxx"
```

### Response Format

```json
{
  "code": 0,           // Business status code
  "message": "success",
  "data": {
    "id": 1,
    "name": "John",
    "email": "john@example.com"
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### Error Responses

```json
{
  "code": 400,
  "message": "Invalid request parameters",
  "errors": {
    "email": "Invalid email format"
  }
}
```

### Pagination

```bash
# Query page 2 with 20 items per page
GET /api/v1/users?page=2&page_size=20&sort=-created_at

# Response includes pagination info
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

## gRPC Service

**gRPC** enables high-performance inter-service communication using HTTP/2 and Protocol Buffers.

### Service Registry

| Service | Purpose | Typical Caller |
|---------|---------|----------------|
| `UserService` | User operations | Admin Server, Bot Service |
| `AuthService` | Authentication | All services |
| `PostService` | Content operations | Admin Server, Scheduler |
| `CommentService` | Comment operations | Admin Server |

### Definition Example

Protocol Buffer definitions are located in `api/pb/apiserver/v1/` directory.

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

### Go Client Usage

```go
package main

import (
  "context"
  "log"

  pb "yourmodule/api/pb/apiserver/v1"
  "google.golang.org/grpc"
)

func main() {
  // Establish connection
  conn, err := grpc.Dial(
    "localhost:9090",
    grpc.WithInsecure(),
  )
  if err != nil {
    log.Fatal(err)
  }
  defer conn.Close()

  // Create client
  client := pb.NewUserServiceClient(conn)

  // Call method
  resp, err := client.GetUser(context.Background(), &pb.GetUserRequest{
    Id: 123,
  })
  if err != nil {
    log.Fatal(err)
  }

  log.Printf("User: %v", resp)
}
```

### Inter-Service Communication

```
┌─────────────┐                  ┌──────────────┐
│ Admin       │ ──gRPC──────────▶ │ API Server   │
│ Server      │ ◀───────gRPC────── │ (9090)       │
└─────────────┘                  └──────────────┘

┌─────────────┐
│ Scheduler   │ ──gRPC──────────▶ │ API Server   │
│             │ ◀───────gRPC────── │ (9090)       │
└─────────────┘                  └──────────────┘
```

## WebSocket Real-Time Communication

⚠️ **Status**: This feature is currently in development and expected in the next release.

### Design Goals

- Support real-time message push
- Support multiple client connections
- Support event subscription and broadcast
- Automatic reconnection mechanism
- Connection lifecycle management

### Planned Endpoints

```
ws://localhost:8081/ws/chat      # Chat messages
ws://localhost:8081/ws/notify    # System notifications
ws://localhost:8081/ws/stream    # Data streams
```

### Planned Features

#### 1. Connection Authentication

```javascript
// Include token when establishing connection
const ws = new WebSocket(
  'ws://localhost:8081/ws/chat?token=' + token
);
```

#### 2. Message Format

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

#### 3. Event Subscription

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

#### 4. Heartbeat

The server sends a ping message every 30 seconds; clients must respond with pong:

```javascript
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);

  if (msg.type === 'ping') {
    ws.send(JSON.stringify({ type: 'pong' }));
  }
};
```

### Development Roadmap

- [ ] Connection management and authentication
- [ ] Message routing and distribution
- [ ] Event subscription system
- [ ] Connection pool management
- [ ] Client library (JavaScript/TypeScript)
- [ ] Unit and integration tests
- [ ] Performance optimization

## Configuration

### Environment Variables

```bash
# HTTP service
API_SERVER_HTTP_PORT=8080
API_SERVER_HTTP_HOST=0.0.0.0

# gRPC service
API_SERVER_GRPC_PORT=9090
API_SERVER_GRPC_HOST=0.0.0.0

# WebSocket service (in development)
API_SERVER_WEBSOCKET_PORT=8081
API_SERVER_WEBSOCKET_HOST=0.0.0.0

# Authentication
JWT_SECRET=your_secret_key
JWT_EXPIRE=86400

# CORS
CORS_ORIGINS=http://localhost:3000,https://example.com
```

### Docker Startup

```bash
docker run -d \
  -p 8080:8080 \
  -p 9090:9090 \
  -p 8081:8081 \
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
      - "9090:9090"
      - "8081:8081"
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

## Performance Optimization

### 1. Caching Strategy

The API Server uses a multi-layer caching approach:

```
Client Cache (HTTP Cache-Control)
    ↓
CDN Cache (optional)
    ↓
Redis Cache (hot data)
    ↓
Database Query
```

### 2. Connection Pooling

```go
// Reuse gRPC connections
conn, _ := grpc.Dial("localhost:9090")
// Multiple RPC calls share the single connection

// Database connection pool
db.SetMaxOpenConns(25)  // Maximum connections
db.SetMaxIdleConns(5)   // Minimum connections
```

### 3. Rate Limiting

```bash
# Middleware rate limiting configuration
# 100 requests per minute per IP
rate_limit:
  enabled: true
  requests_per_minute: 100
  by_ip: true
```

## Monitoring Metrics

The API Server exposes the following Prometheus metrics:

```
# HTTP requests
api_server_http_requests_total{method,status,endpoint}
api_server_http_request_duration_seconds{endpoint}
api_server_http_request_size_bytes{endpoint}
api_server_http_response_size_bytes{endpoint}

# gRPC requests
api_server_grpc_requests_total{service,method,status}
api_server_grpc_request_duration_seconds{service,method}

# Connection status
api_server_active_connections{type}  # http, grpc, websocket
api_server_connection_errors_total{type,reason}
```

## FAQ

### Q: How does the API Server handle concurrent requests?

A: The API Server leverages Go's goroutines for automatic concurrency handling. Each request spawns a new goroutine. Control the number of processors using the GOMAXPROCS environment variable:

```bash
export GOMAXPROCS=8  # Use 8 processors
```

### Q: How do I scale the API Server?

A: The stateless design enables easy scaling:

```
┌─────────────┐
│   Nginx     │  Load Balancer
└──────┬──────┘
       │
   ┌───┴────┬──────┬──────┐
   ▼        ▼      ▼      ▼
API-1    API-2  API-3  API-N
   │        │      │      │
   └────────┴──────┴──────┘
      MySQL + Redis
```

All instances share the same database and cache, allowing you to add or remove instances independently.

### Q: When will WebSocket be released?

A: WebSocket is currently in development. We plan to release it in v1.2.0. Stay tuned!

## Next Step

- [Scheduler](./scheduler.md) - Learn about scheduled tasks and async queue service
