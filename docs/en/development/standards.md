---
title: Development Standards - Bingo Go Project Code Style and Conventions
description: Learn about Bingo Go microservices project development standards, including naming conventions, error handling, logging, comments, database conventions, and Git workflow.
---

# Development Standards

This document defines coding standards and development conventions for the Bingo project.

## Naming Conventions

### Package Names
- **Lowercase**, short, meaningful
- **Singular** form (unless special cases)
- No underscores or camelCase

```go
// ✅ Correct
package user
package article
package auth

// ❌ Wrong
package users          // Should be singular
package articleMgr     // Don't use camelCase
package user_service   // Don't use underscore
```

### File Names
- **snake_case**
- Related to main type

```
user_handler.go
article_store.go
auth_middleware.go
```

### Interface Names
- **`I` prefix** (Interface)
- PascalCase

```go
type IStore interface {}
type IBiz interface {}
type ICache interface {}
```

### Structs and Methods
- **PascalCase** (exported) or **camelCase** (private)

```go
// Exported
type UserHandler struct {}
func (h *UserHandler) Create() {}

// Private
type userCache struct {}
func (c *userCache) get() {}
```

### Constants and Variables
```go
// Constants: PascalCase
const MaxRetryCount = 3
const DefaultPageSize = 20

// Variables: camelCase
var userCache *cache.Cache
var defaultTimeout = 30 * time.Second
```

## Error Handling

### Unified Error Codes

Define in `internal/pkg/errno/code.go`:

```go
var (
    // User-related errors (100xx)
    ErrUserNotFound      = errno.New(10001, "User not found")
    ErrUserAlreadyExists = errno.New(10002, "User already exists")
    ErrInvalidPassword   = errno.New(10003, "Password incorrect")

    // Article-related errors (200xx)
    ErrArticleNotFound = errno.New(20001, "Article not found")
)
```

### Error Return

```go
// ✅ Correct: Return defined error code
if user == nil {
    return nil, errno.ErrUserNotFound
}

// ✅ Correct: Wrap error
if err := db.Create(user).Error; err != nil {
    return nil, fmt.Errorf("failed to create user: %w", err)
}

// ❌ Wrong: Direct string error
if user == nil {
    return nil, errors.New("user not found")
}
```

### Handler Layer Error Handling

```go
func (h *UserHandler) Get(c *gin.Context) {
    user, err := h.biz.Users().Get(c.Context(), id)
    if err != nil {
        // Unified error response
        core.WriteResponse(c, err, nil)
        return
    }

    core.WriteResponse(c, nil, user)
}
```

## Logging Standards

### Log Levels

- **Debug**: Debug information
- **Info**: Important business processes
- **Warn**: Warning messages, not affecting main flow
- **Error**: Error messages requiring attention

### Logging

```go
import "github.com/bingo-project/bingo/internal/pkg/logger"

// ✅ Structured logging
logger.Info("user created",
    zap.String("username", username),
    zap.Uint64("user_id", userID),
)

// ✅ Error logging (with context)
logger.Error("failed to create user",
    zap.Error(err),
    zap.String("username", username),
)

// ❌ Not recommended: Unstructured logging
logger.Info("user created: " + username)
```

### Logging Best Practices

1. **Log critical business operations**
```go
logger.Info("user login",
    zap.Uint64("user_id", userID),
    zap.String("ip", clientIP),
)
```

2. **Log errors with context**
```go
logger.Error("database query failed",
    zap.Error(err),
    zap.String("sql", sql),
    zap.Any("params", params),
)
```

3. **Don't log sensitive information**
```go
// ❌ Wrong: Log password
logger.Info("user login", zap.String("password", password))

// ✅ Correct: Don't log sensitive info
logger.Info("user login", zap.String("username", username))
```

## Comment Standards

### File Comments

Every file must start with `ABOUTME` comment:

```go
// ABOUTME: User business logic implementation
// ABOUTME: Handles user registration, login, and profile management
package user
```

### Function Comments

```go
// CreateUser creates a new user
// Parameters:
//   - ctx: context
//   - req: create user request
// Returns:
//   - *model.User: created user
//   - error: error message
func (b *UserBiz) CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // ...
}
```

### Swagger Comments

```go
// @Summary      Get user info
// @Description  Get user details by user ID
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /v1/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
    // ...
}
```

## Code Organization

### Handler Layer

```go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/bingo-project/bingo/internal/apiserver/biz"
    "github.com/bingo-project/bingo/pkg/core"
)

// 1. Type definitions
type UserHandler struct {
    biz biz.IBiz
}

// 2. Constructor
func New(biz biz.IBiz) *UserHandler {
    return &UserHandler{biz: biz}
}

// 3. HTTP handlers (in CRUD order)
func (h *UserHandler) Create(c *gin.Context) {}
func (h *UserHandler) Get(c *gin.Context) {}
func (h *UserHandler) List(c *gin.Context) {}
func (h *UserHandler) Update(c *gin.Context) {}
func (h *UserHandler) Delete(c *gin.Context) {}

// 4. Private helper methods
func (h *UserHandler) validateRequest() {}
```

### Biz Layer

```go
package user

// 1. Interface definition
type UserBiz interface {
    Create(ctx context.Context, req *CreateUserRequest) (*model.User, error)
    Get(ctx context.Context, id uint64) (*model.User, error)
}

// 2. Implementation struct
type userBiz struct {
    ds store.IStore
}

// 3. Constructor
func New(ds store.IStore) UserBiz {
    return &userBiz{ds: ds}
}

// 4. Interface implementation
func (b *userBiz) Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {}

// 5. Private methods
func (b *userBiz) validateUser(req *CreateUserRequest) error {}
```

### Store Layer

```go
package store

// 1. Interface definition
type UserStore interface {
    Create(ctx context.Context, user *model.User) error
    Get(ctx context.Context, id uint64) (*model.User, error)
}

// 2. Implementation struct
type userStore struct {
    db *gorm.DB
}

// 3. Constructor
func newUserStore(db *gorm.DB) UserStore {
    return &userStore{db: db}
}

// 4. Interface implementation
func (s *userStore) Create(ctx context.Context, user *model.User) error {}
```

### Store Naming Convention

Store layer uses the following naming convention (compatible with `pkg/store` universal Store[T]):

| Element | Convention | Example |
|---------|-----------|---------|
| File name | `<prefix>_<model>.go` | `user.go`, `sys_config.go` |
| Store interface | `<Prefix><Model>Store` | `UserStore`, `SysConfigStore` |
| Implementation struct | `<prefix><model>Store` (lowercase) | `userStore`, `sysConfigStore` |
| Extension interface | `<Prefix><Model>Expansion` | `UserExpansion`, `SysConfigExpansion` |
| Factory function | `New<Prefix><Model>Store()` | `NewUserStore()`, `NewSysConfigStore()` |
| IStore method | `<Model>s()` or `<Model>()` | `Users()`, `SysConfig()` |

**Prefix Convention**:
- **System modules**: `sys_` prefix (e.g., `sys_config.go`, `sys_admin.go`)
- **Specific modules**: module name prefix (e.g., `bot_admin.go`, `bot_channel.go`)
- **Independent features**: no prefix (e.g., `user.go`)

### Store Best Practices

| Principle | Description |
|-----------|-------------|
| **Flat files** | All files in `internal/pkg/store` are flat in one directory, avoid circular imports |
| **Depend on interface** | Upper layers depend on IStore interface, not concrete implementation |
| **Preload associations** | Use `Load()` to avoid N+1 queries |
| **Transaction consistency** | Use `TX()` for atomic operations |
| **Error handling** | Database errors are automatically logged |

## Database Standards

### Model Definition

```go
type User struct {
    ID        uint64    `gorm:"primarykey" json:"id"`
    Username  string    `gorm:"size:50;not null;uniqueIndex" json:"username"`
    Email     string    `gorm:"size:100;not null;index" json:"email"`
    Password  string    `gorm:"size:255;not null" json:"-"` // json:"-" don't serialize
    Status    int       `gorm:"default:1" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}
```

### Query Optimization

```go
// ✅ Use indexed fields
db.Where("username = ?", username).First(&user)

// ✅ Preload associations (avoid N+1)
db.Preload("Articles").Find(&users)

// ✅ Only query needed fields
db.Select("id", "username", "email").Find(&users)

// ❌ Avoid full table scan
db.Where("username LIKE ?", "%"+keyword+"%").Find(&users)
```

## Git Standards

### Commit Message

```bash
# Format
<type>: <subject>

# Types
feat: new feature
fix: bug fix
docs: documentation update
style: code format adjustment
refactor: refactoring
test: test related
chore: build/tool related

# Examples
feat: add user login API
fix: resolve password encryption issue
docs: update API documentation
refactor: simplify user validation logic
```

### Branch Management

```
main/master     Production branch
develop         Development branch
feature/*       Feature branch
bugfix/*        Bug fix branch
hotfix/*        Emergency fix branch
```

## Testing Standards

### Test File Naming

```
user.go       -> user_test.go
article.go    -> article_test.go
```

### Test Cases

```go
func TestUserBiz_Create(t *testing.T) {
    // 1. Prepare test data
    req := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
    }

    // 2. Execute test
    user, err := userBiz.Create(context.Background(), req)

    // 3. Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "testuser", user.Username)
}
```

## Makefile Usage

```bash
# Build
make build
make build BINS="bingo-apiserver"

# Test
make test
make cover

# Code check
make lint
make format

# Generate documentation
make swagger
```

## Next Step

- [Testing Guide](./testing.md) - Learn about project testing standards and best practices
- [Docker Deployment](../deployment/docker.md) - Deploy Bingo projects using Docker
