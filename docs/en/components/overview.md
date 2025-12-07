---
title: Core Components Overview - Bingo Go Microservices Framework Components
description: Learn about Bingo Go microservices framework core components, including Bootstrap initialization, Redis caching, Swagger documentation, parameter validation, and error handling.
---

# Core Components Overview

Bingo includes a set of out-of-the-box core components. This document introduces the functionality and usage of each component.

## Component List

### Basic Components

| Component | Description | Original Package |
|-----------|-------------|------------------|
| **Bootstrap** | Application initialization, unified startup flow | `internal/pkg/bootstrap` |
| **Facade** | Global component access, singleton pattern | `internal/pkg/facade` |
| **Logger** | Zap-based structured logging | [component-base/log](https://github.com/bingo-project/component-base) |
| **Database** | GORM database wrapper | [gorm.io/gorm](https://gorm.io) |
| **Redis** | Redis cache wrapper | [go-redis/redis](https://github.com/redis/go-redis) |

### Authentication

| Component | Description | Original Package |
|-----------|-------------|------------------|
| **JWT** | JWT authentication | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) |
| **Casbin** | RBAC permission control | [casbin/casbin](https://github.com/casbin/casbin) |

### Async Tasks

| Component | Description | Original Package |
|-----------|-------------|------------------|
| **Asynq** | Task queue and scheduled tasks | [hibiken/asynq](https://github.com/hibiken/asynq) |

### Other Components

| Component | Description | Original Package |
|-----------|-------------|------------------|
| **Swagger** | API documentation generation | [swaggo/swag](https://github.com/swaggo/swag) |
| **Validator** | Parameter validation | [go-playground/validator](https://github.com/go-playground/validator) |
| **Snowflake** | Distributed ID generation | [bwmarrin/snowflake](https://github.com/bwmarrin/snowflake) |

## Bootstrap and Facade

Bingo uses Bootstrap for application initialization and Facade for global component access.

### Bootstrap Initialization

`internal/pkg/bootstrap/app.go` defines the unified startup flow:

```go
// Boot initializes all core components
func Boot() {
    InitLog()        // Logging system
    InitTimezone()   // Timezone settings
    InitSnowflake()  // Distributed ID
    InitMail()       // Email service
    InitCache()      // Cache service
    InitAES()        // Encryption component
    InitQueue()      // Task queue
}
```

### Facade Global Access

`internal/pkg/facade/facade.go` provides global component instances:

```go
import "bingo/internal/pkg/facade"

// Access configuration
cfg := facade.Config

// Access Redis
facade.Redis.Set(ctx, "key", "value", time.Hour)

// Access cache service
facade.Cache.Set(ctx, "key", value, time.Hour)

// Generate distributed ID
id := facade.Snowflake.Generate()

// Send email
facade.Mail.Send(to, subject, body)
```

### Initialization Order

```
1. Load config (InitConfig)
    ↓
2. Initialize logging (InitLog)
    ↓
3. Initialize timezone (InitTimezone)
    ↓
4. Initialize distributed ID (InitSnowflake)
    ↓
5. Initialize cache (InitCache)
    ↓
6. Initialize database (InitDB)
    ↓
7. Initialize Store (NewStore)
```

## Redis

Based on [go-redis](https://github.com/redis/go-redis).

### Usage

```go
import "bingo/internal/pkg/facade"

// Access Redis client through Facade
rdb := facade.Redis

// Basic operations
rdb.Set(ctx, "key", "value", time.Hour)
val, err := rdb.Get(ctx, "key").Result()

// Cache objects
type User struct {
    ID   uint64
    Name string
}

// Write to cache
user := &User{ID: 1, Name: "test"}
data, _ := json.Marshal(user)
rdb.Set(ctx, "user:1", data, time.Hour)

// Read from cache
data, _ := rdb.Get(ctx, "user:1").Bytes()
var user User
json.Unmarshal(data, &user)
```

> For detailed usage, refer to the [go-redis documentation](https://redis.uptrace.dev/)

## Swagger

API documentation auto-generation based on [swaggo/swag](https://github.com/swaggo/swag).

### Annotation Example

```go
// @Summary      Create user
// @Description  Create a new user account
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUserRequest  true  "User information"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
    // ...
}
```

### Generate Documentation

```bash
# Generate Swagger docs
make swagger

# Access documentation
open http://localhost:8080/swagger/index.html
```

> For detailed annotation syntax, refer to the [swag documentation](https://github.com/swaggo/swag#declarative-comments-format)

## Validator

Parameter validation based on [go-playground/validator](https://github.com/go-playground/validator), integrated with the Gin framework.

### Usage

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"gte=18,lte=100"`
    Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest

    // Gin auto-validation
    if err := c.ShouldBindJSON(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // Validation passed, continue processing
}
```

### Common Validation Tags

| Tag | Description |
|-----|-------------|
| `required` | Required field |
| `min=3` | Minimum length/value |
| `max=20` | Maximum length/value |
| `email` | Email format |
| `url` | URL format |
| `oneof=a b` | Enum values |
| `gte=18` | Greater than or equal |
| `lte=100` | Less than or equal |

> For a complete list of validation tags, refer to the [validator documentation](https://pkg.go.dev/github.com/go-playground/validator/v10)

## Error Handling

Unified error handling component located in `internal/pkg/errno`.

### Define Error Codes

```go
// internal/pkg/errno/code.go
var (
    ErrUserNotFound = errno.New(10001, "User not found")
    ErrInvalidToken = errno.New(10002, "Invalid token")
)
```

### Use Error Codes

```go
// Return error in Biz layer
if user == nil {
    return nil, errno.ErrUserNotFound
}

// Handle uniformly in Handler layer
func (h *UserHandler) Get(c *gin.Context) {
    user, err := h.biz.Users().Get(c.Request.Context(), id)
    // Unified error response
    core.WriteResponse(c, err, user)
}
```

### Error Response Format

```json
{
  "code": 10001,
  "message": "User not found"
}
```

## Extending Components

### Adding New Global Components

1. Add variable definition in `internal/pkg/facade/facade.go`
2. Add initialization function in `internal/pkg/bootstrap/`
3. Call the initialization function in `Boot()`

```go
// 1. Add variable in facade.go
var MyComponent *mypackage.Client

// 2. Add initialization in bootstrap/
func InitMyComponent() {
    facade.MyComponent = mypackage.NewClient(facade.Config.MyComponent)
}

// 3. Call in app.go
func Boot() {
    // ... other initializations
    InitMyComponent()
}
```

## Next Step

- [Development Standards](../development/standards.md) - Code style and best practices
