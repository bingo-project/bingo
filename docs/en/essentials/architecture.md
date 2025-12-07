---
title: Bingo Architecture - Go Microservices Architecture Design Guide
description: Deep dive into Bingo Go microservices framework architecture design, including service composition, layered architecture, and tech stack. Learn how to build scalable Golang backend systems with Gin, GORM, and Redis.
---

# Overall Architecture

Bingo adopts a **microservice architecture** with a clear **three-layer design**, enabling teams to build scalable and maintainable backend systems.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Client Application                          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ↓
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway / LB                           │
└────────────────────────────┬────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        ↓                    ↓                    ↓
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│  API Server      │ │  Admin Server    │ │ Scheduler        │
│  ┌────────────┐  │ │  ┌────────────┐  │ │  ┌────────────┐  │
│  │  Handler   │  │ │  │  Handler   │  │ │  │ Job Engine │  │
│  └──────┬─────┘  │ │  └──────┬─────┘  │ │  └──────┬─────┘  │
│         ↓        │ │         ↓        │ │         ↓        │
│  ┌────────────┐  │ │  ┌────────────┐  │ │  ┌────────────┐  │
│  │   Biz      │  │ │  │   Biz      │  │ │  │   Biz      │  │
│  └──────┬─────┘  │ │  └──────┬─────┘  │ │  └──────┬─────┘  │
│         ↓        │ │         ↓        │ │         ↓        │
│  ┌────────────┐  │ │  ┌────────────┐  │ │  ┌────────────┐  │
│  │   Store    │  │ │  │   Store    │  │ │  │   Store    │  │
│  └──────┬─────┘  │ │  └──────┬─────┘  │ │  └──────┬─────┘  │
└─────────┼────────┘ └────────┼────────┘ └────────┼────────┘
          │                    │                   │
          └────────────────────┼───────────────────┘
                               ↓
        ┌──────────────────────────────────────────┐
        │     Shared Data Layer (Database/Cache)   │
        │  ┌──────────────┐  ┌──────────────────┐ │
        │  │    MySQL     │  │     Redis        │ │
        │  └──────────────┘  └──────────────────┘ │
        └──────────────────────────────────────────┘
```

## Core Design Principles

### 1. Microservice Architecture

**Multiple Independent Services**:
- Each service is independently deployable
- Services communicate via HTTP/gRPC
- Enables horizontal scaling and independent team ownership

**Service Types**:
- **API Server**: Handles HTTP requests from clients
- **Admin Server**: Administrative backend operations
- **Scheduler**: Scheduled task processing
- **Bot Service**: Third-party integrations (Discord, Telegram, etc.)

### 2. Three-Layer Architecture

Every service follows the three-layer design:

#### Layer 1: Handler (HTTP Handler)
- **Responsibility**: Handle HTTP requests, parameter validation
- **Input**: HTTP requests
- **Output**: JSON responses
- **Key Features**:
  - Request parameter validation
  - Authorization checks
  - Response formatting
  - Error handling

#### Layer 2: Business Logic (Biz)
- **Responsibility**: Core business logic implementation
- **Input**: Validated parameters from Handler
- **Output**: Business results
- **Key Features**:
  - Core business rules
  - Data validation
  - Cross-service calls
  - Cache management

#### Layer 3: Data Access (Store)
- **Responsibility**: Database and external service interactions
- **Input**: Data queries and updates
- **Output**: Data objects
- **Key Features**:
  - Database operations (CRUD)
  - Query optimization
  - Transaction management
  - Cache integration

### 3. Dependency Injection

```go
// Interface-based dependency injection
type UserBiz interface {
    CreateUser(ctx context.Context, user *User) error
    GetUser(ctx context.Context, id string) (*User, error)
}

// Implementation can be injected
type userBiz struct {
    store UserStore
}

// Easy to test with mock implementations
```

Benefits:
- Easy to test with mocks
- Loose coupling between layers
- Easy to replace implementations

### 4. Generic Data Access Layer

Bingo uses Go generics to implement a reusable Store pattern:

```go
// Generic Store[T] reduces code duplication
type Store[T any] interface {
    Create(ctx context.Context, obj *T) error
    Get(ctx context.Context, id string) (*T, error)
    List(ctx context.Context, query Query) ([]*T, error)
    Update(ctx context.Context, id string, obj *T) error
    Delete(ctx context.Context, id string) error
}
```

## Data Flow

### Request Processing Flow

```
1. Client Request
   ↓
2. Handler
   - Parse parameters
   - Validate input
   ↓
3. Biz Layer
   - Execute business logic
   - Coordinate data access
   ↓
4. Store Layer
   - Database queries
   - Cache operations
   ↓
5. Response
   - Return result to client
```

### Example: User Creation

```
POST /api/v1/users

1. Handler.CreateUser()
   - Validate request body
   - Check authorization

2. Biz.CreateUser()
   - Check duplicate username
   - Encrypt password
   - Generate user ID

3. Store.CreateUser()
   - Insert into database
   - Update cache

4. Return Response
   - Return user object
```

## Service Communication

### Synchronous Communication (gRPC)

```go
// Service-to-service calls
client := userservice.NewClient(conn)
user, err := client.GetUser(ctx, &GetUserRequest{Id: "123"})
```

**Use Cases**:
- Real-time data queries
- Immediate response needed
- Small data transfers

### Asynchronous Communication (Message Queue)

```go
// Using Asynq for task queue
task := asynq.NewTask("send_email", payload)
info, err := client.Enqueue(task)
```

**Use Cases**:
- Heavy computations
- Delayed processing
- Decoupled services

## Scalability Patterns

### Horizontal Scaling

1. **Stateless Services**: No local state, easily deployable
2. **Load Balancing**: Distribute requests across instances
3. **Database Replication**: Master-slave setup for read scaling
4. **Caching Layer**: Redis for frequently accessed data

### Example Scaling Architecture

```
                    ┌─ API Server 1
                    ├─ API Server 2
Load Balancer ─────┤
                    └─ API Server N
                           ↓
                    ┌──────────────┐
                    │   Database   │
                    │ (Replicated) │
                    └──────────────┘
```

## Next Step

- [Layered Architecture in Detail](./layered-design.md) - Understand the internal architecture of each service
