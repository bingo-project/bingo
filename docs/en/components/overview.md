# Core Components Overview

Bingo comes with a series of out-of-the-box core components. This document introduces the functionality and use cases of each component.

## Component List

### Basic Components

#### Web Framework - Gin

**Purpose**: High-performance HTTP server framework

**Key Features**:
- Fast routing
- Middleware support
- Parameter binding
- JSON validation

**Usage Example**:
```go
func getUserHandler(c *gin.Context) {
    userID := c.Param("id")
    user, err := biz.GetUser(c, userID)
    c.JSON(200, user)
}
```

**Documentation**: [Gin](https://github.com/gin-gonic/gin)

#### ORM - GORM

**Purpose**: Object-Relational Mapping for database operations

**Key Features**:
- Support for MySQL, PostgreSQL, SQLite
- Hooks and callbacks
- Association management
- Query builder

**Usage Example**:
```go
var user User
db.First(&user, "id = ?", userID)
```

**Documentation**: [GORM](https://gorm.io)

#### Caching - Redis

**Purpose**: Distributed caching and session management

**Key Features**:
- Key-value storage
- Expiration support
- Pub/Sub messaging
- Data persistence

**Usage Example**:
```go
val, err := redisClient.Get(ctx, "user:123").Result()
redisClient.Set(ctx, "user:123", userData, 1*time.Hour)
```

**Documentation**: [Redis](https://redis.io)

### Advanced Components

#### Task Queue - Asynq

**Purpose**: Reliable asynchronous task processing

**Key Features**:
- Task persistence
- Retry mechanism
- Priority queues
- Result tracking

**Use Cases**:
- Email sending
- Image processing
- Data export
- Batch operations

**Usage Example**:
```go
task := asynq.NewTask("send_email", payload)
info, err := client.Enqueue(task)
```

**Documentation**: [Asynq](https://github.com/hibiken/asynq)

#### Permission Control - Casbin

**Purpose**: Flexible RBAC (Role-Based Access Control) engine

**Key Features**:
- Role-based access control
- Resource-based access control
- ABAC (Attribute-Based Access Control)
- Policy management

**Use Cases**:
- User permissions
- API access control
- Resource authorization

**Example Policy**:
```
p, admin, /api/users, POST
p, user, /api/profile, GET
p, user, /api/profile, PUT
```

**Documentation**: [Casbin](https://casbin.org)

#### Logging - Zap

**Purpose**: High-performance structured logging

**Key Features**:
- Structured logging
- Multiple log levels
- Log rotation
- Custom fields

**Usage Example**:
```go
logger.Info("user created", zap.String("id", userID), zap.String("email", email))
```

**Documentation**: [Zap](https://github.com/uber-go/zap)

#### API Documentation - Swagger

**Purpose**: Auto-generate API documentation

**Key Features**:
- API endpoint documentation
- Request/response schemas
- Interactive testing
- OpenAPI specification

**Usage**:
```bash
# Generate Swagger docs
make swagger

# Access at http://localhost:8080/swagger/index.html
```

**Documentation**: [Swag](https://github.com/swaggo/swag)

#### Configuration Management - Viper

**Purpose**: Configuration file management

**Supported Formats**:
- YAML
- JSON
- TOML
- INI

**Usage Example**:
```go
viper.SetConfigName("config")
viper.ReadInConfig()
dbHost := viper.GetString("database.host")
```

**Documentation**: [Viper](https://github.com/spf13/viper)

#### Authentication - JWT

**Purpose**: Stateless authentication

**Key Features**:
- Token generation
- Token validation
- Claims management
- Token refresh

**Usage Example**:
```go
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString(secretKey)
```

**Documentation**: [golang-jwt](https://github.com/golang-jwt/jwt)

#### Validation - Validator

**Purpose**: Data structure validation

**Key Features**:
- Tag-based validation
- Custom validators
- Cross-field validation
- Localized error messages

**Usage Example**:
```go
type User struct {
    Email string `validate:"required,email"`
    Age   int    `validate:"min=18,max=100"`
}
```

**Documentation**: [Validator](https://github.com/go-playground/validator)

### Development Tools

#### CLI Tool - bingoctl

**Purpose**: Code generation and project management

**Core Features**:
- Project creation
- Code generation (CRUD, Models, etc.)
- Database migration
- Configuration management

**Usage**:
```bash
bingoctl create github.com/myorg/myapp
bingoctl make crud user
```

**Documentation**: [bingoctl](https://github.com/bingo-project/bingoctl)

#### Hot Reload - Air

**Purpose**: Development-time auto-reload

**Features**:
- Watch file changes
- Auto-rebuild
- Auto-restart
- Preserve state

**Configuration**:
```bash
cp .air.example.toml .air.toml
air
```

**Documentation**: [Air](https://github.com/cosmtrek/air)

## Integration Patterns

### Common Integration Scenarios

#### Database + Cache Pattern
```
Request
  ↓
Query Cache
  ↓ (miss)
Query Database
  ↓
Update Cache
  ↓
Response
```

#### Task Queue Pattern
```
HTTP Request
  ↓
Create Task (Asynq)
  ↓
Return Immediately
  ↓
Background Worker Processes Task
```

#### Authentication Pattern
```
Login Request
  ↓
Validate Credentials
  ↓
Generate JWT Token
  ↓
Return Token
  ↓
Client Includes Token in Requests
  ↓
Validate Token with Casbin
```

## Choosing the Right Components

| Scenario | Recommended Component |
|----------|----------------------|
| User authentication | JWT + Casbin |
| User data storage | GORM + MySQL |
| Session caching | Redis |
| Heavy tasks | Asynq |
| API documentation | Swagger |
| Configuration management | Viper |
| Data validation | Validator |
| Logging | Zap |

## Performance Considerations

1. **Database**: Use connection pooling, optimize queries
2. **Cache**: Set appropriate TTL, implement cache invalidation
3. **Task Queue**: Monitor queue depth, adjust worker count
4. **Logging**: Use appropriate log levels in production

## Next Steps

- [Core Architecture](../essentials/architecture.md) - Understand system design
- [Layered Architecture](../essentials/layered-design.md) - Learn layer responsibilities
- [Development Standards](../development/standards.md) - Follow best practices
