# Layered Architecture in Detail

Bingo adopts a classic three-layer architecture design. This document explains the responsibility and design principles of each layer.

## Three-Layer Architecture

```
┌─────────────────────────────────────────┐
│         Controller Layer                │  HTTP/gRPC Handler Layer
│  - Parameter validation                 │
│  - Request/response conversion          │
│  - Error handling                       │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Business Layer (Biz)           │  Business Logic Layer
│  - Business rules                       │
│  - Process orchestration                │
│  - Transaction control                  │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Store Layer                    │  Data Access Layer
│  - Database operations                  │
│  - Cache operations                     │
│  - Third-party service calls            │
└─────────────────────────────────────────┘
```

## Controller Layer (HTTP/gRPC Handler)

### Responsibilities

1. **Receive Requests**: Handle HTTP/gRPC requests
2. **Parameter Validation**: Bind and validate request parameters
3. **Call Business Logic**: Invoke Biz layer to process business logic
4. **Return Responses**: Construct and return responses

### Code Example

```go
// internal/apiserver/controller/v1/user/user.go
type UserController struct {
    biz biz.IBiz
}

func (ctrl *UserController) Get(c *gin.Context) {
    // 1. Parameter validation
    var req GetUserRequest
    if err := c.ShouldBindUri(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // 2. Call business layer
    user, err := ctrl.biz.Users().Get(c.Context(), req.UserID)
    if err != nil {
        core.WriteResponse(c, err, nil)
        return
    }

    // 3. Return response
    core.WriteResponse(c, nil, user)
}
```

### Design Principles

- **Thin Controllers**: Only handle parameter processing and responses, no business logic
- **Unified Responses**: Use consistent response format
- **Error Handling**: Unified error handling mechanism
- **Version Isolation**: Different API versions in separate directories (`v1/`, `v2/`)

### What NOT to Do

❌ **Don't write business logic in Controller**
```go
// Wrong example
func (ctrl *UserController) Create(c *gin.Context) {
    // ❌ Business rules shouldn't be here
    if user.Age < 18 {
        return errors.New("Age too young")
    }

    // ❌ Password encryption shouldn't be here
    hashedPassword := encrypt(user.Password)
}
```

✅ **Call Biz layer instead**
```go
// Correct example
func (ctrl *UserController) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // ✅ Business logic goes to Biz layer
    user, err := ctrl.biz.Users().Create(c.Context(), &req)
    core.WriteResponse(c, err, user)
}
```

## Biz Layer (Business Logic)

### Responsibilities

1. **Business Rules**: Implement core business logic and rules
2. **Process Orchestration**: Coordinate multiple Store operations
3. **Transaction Control**: Handle database transactions
4. **Business Validation**: Business-level validation

### Code Example

```go
// internal/apiserver/biz/user/user.go
type UserBiz struct {
    ds store.IStore
}

func (b *UserBiz) Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // 1. Business rule validation
    if err := b.validateUser(req); err != nil {
        return nil, err
    }

    // 2. Process business logic
    req.Password = encryptPassword(req.Password)

    // 3. Build model
    user := &model.User{
        Username: req.Username,
        Password: req.Password,
        Email:    req.Email,
    }

    // 4. Persist data
    if err := b.ds.Users().Create(ctx, user); err != nil {
        return nil, err
    }

    // 5. Business process orchestration (e.g., send welcome email)
    go b.sendWelcomeEmail(user.Email)

    return user, nil
}

func (b *UserBiz) validateUser(req *CreateUserRequest) error {
    // Business rule validation
    if req.Age < 18 {
        return errno.ErrUserAgeTooYoung
    }

    // Check if username already exists
    exists, err := b.ds.Users().ExistsByUsername(ctx, req.Username)
    if err != nil {
        return err
    }
    if exists {
        return errno.ErrUserAlreadyExists
    }

    return nil
}
```

### Design Principles

- **Core Business**: All business logic lives here
- **Interface Programming**: Depends on Store interface, not concrete implementation
- **Testability**: Unit test via Mock Store
- **Transaction Control**: Use Store's transaction methods when needed

### Typical Scenarios

#### Scenario 1: Single Table Operation
```go
func (b *UserBiz) Get(ctx context.Context, id uint64) (*model.User, error) {
    return b.ds.Users().Get(ctx, id)
}
```

#### Scenario 2: Multi-Table Operation Orchestration
```go
func (b *OrderBiz) Create(ctx context.Context, req *CreateOrderRequest) error {
    // 1. Check inventory
    stock, err := b.ds.Products().GetStock(ctx, req.ProductID)
    if err != nil {
        return err
    }
    if stock < req.Quantity {
        return errno.ErrInsufficientStock
    }

    // 2. Create order
    order := &model.Order{...}
    if err := b.ds.Orders().Create(ctx, order); err != nil {
        return err
    }

    // 3. Decrease stock
    if err := b.ds.Products().DecreaseStock(ctx, req.ProductID, req.Quantity); err != nil {
        return err
    }

    return nil
}
```

#### Scenario 3: Transaction Control
```go
func (b *OrderBiz) Create(ctx context.Context, req *CreateOrderRequest) error {
    // Use transaction
    return b.ds.TX(ctx, func(ctx context.Context) error {
        // Execute multiple operations in transaction
        if err := b.ds.Orders().Create(ctx, order); err != nil {
            return err
        }

        if err := b.ds.Products().DecreaseStock(ctx, productID, quantity); err != nil {
            return err
        }

        return nil
    })
}
```

## Store Layer (Data Access)

> 💡 For detailed design documentation, refer to [Store Package Design](./store.md)

### Responsibilities

1. **Database Operations**: Encapsulate GORM operations
2. **Cache Operations**: Redis cache read/write
3. **Data Transformation**: Convert data formats
4. **Query Optimization**: SQL optimization and index usage

### Code Example

```go
// internal/apiserver/store/user.go
type UserStore interface {
    Create(ctx context.Context, user *model.User) error
    Get(ctx context.Context, id uint64) (*model.User, error)
    List(ctx context.Context, opts ListOptions) ([]*model.User, int64, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id uint64) error
}

type userStore struct {
    db *gorm.DB
}

func (s *userStore) Create(ctx context.Context, user *model.User) error {
    return s.db.WithContext(ctx).Create(user).Error
}

func (s *userStore) Get(ctx context.Context, id uint64) (*model.User, error) {
    var user model.User
    if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, errno.ErrUserNotFound
        }
        return nil, err
    }
    return &user, nil
}

func (s *userStore) List(ctx context.Context, opts ListOptions) ([]*model.User, int64, error) {
    var users []*model.User
    var count int64

    db := s.db.WithContext(ctx).Model(&model.User{})

    // Conditional query
    if opts.Username != "" {
        db = db.Where("username LIKE ?", "%"+opts.Username+"%")
    }

    // Count
    if err := db.Count(&count).Error; err != nil {
        return nil, 0, err
    }

    // Pagination
    if err := db.Offset(opts.Offset).Limit(opts.Limit).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, count, nil
}
```

### Design Principles

- **Pure Data Operations**: Only database/cache operations, no business logic
- **Interface Definition**: Each Store defines an interface
- **Error Conversion**: Convert database errors to business errors
- **Query Optimization**: Pay attention to N+1 problems, use Preload wisely

### Cache Usage Example

```go
func (s *userStore) Get(ctx context.Context, id uint64) (*model.User, error) {
    // 1. Try to get from cache
    cacheKey := fmt.Sprintf("user:%d", id)
    var user model.User

    if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil
    }

    // 2. Cache miss, query from database
    if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }

    // 3. Write to cache
    _ = s.cache.Set(ctx, cacheKey, &user, time.Hour)

    return &user, nil
}
```

## Why Layered Architecture?

### 1. Separation of Concerns
Each layer focuses on its own responsibility:
- Controller focuses on HTTP protocol
- Biz focuses on business rules
- Store focuses on data access

### 2. Easy to Test
```go
// Test Biz layer by mocking Store layer
func TestUserBiz_Create(t *testing.T) {
    mockStore := &MockStore{}
    biz := user.New(mockStore)

    // Test business logic
    err := biz.Create(ctx, req)
    assert.NoError(t, err)
}
```

### 3. Code Reuse
Biz layer can be reused by multiple Controllers:
```
HTTP Controller  ──┐
                   ├──→  User Biz  ──→  User Store
gRPC Service    ──┘
```

### 4. Easy to Maintain
- Modify database operations: only change Store layer
- Modify business rules: only change Biz layer
- Modify API format: only change Controller layer

### 5. Team Collaboration
Different layers can be developed in parallel:
- Frontend developers: Mock Controller and develop in parallel
- Backend developers: Define interface first, develop in layers

## Common Mistakes

### Mistake 1: Cross-Layer Calls

❌ **Controller directly calls Store**
```go
// Wrong
func (ctrl *UserController) Get(c *gin.Context) {
    // ❌ Controller shouldn't directly call Store
    user, err := ctrl.store.Users().Get(ctx, id)
}
```

✅ **Use Biz layer**
```go
// Correct
func (ctrl *UserController) Get(c *gin.Context) {
    user, err := ctrl.biz.Users().Get(ctx, id)
}
```

### Mistake 2: Business Logic Leak

❌ **Store layer contains business logic**
```go
// Wrong
func (s *userStore) Create(ctx context.Context, user *model.User) error {
    // ❌ Business validation shouldn't be in Store layer
    if user.Age < 18 {
        return errors.New("Age too young")
    }
    return s.db.Create(user).Error
}
```

✅ **Business logic in Biz layer**
```go
// Correct: Biz layer validates
func (b *userBiz) Create(ctx context.Context, req *CreateUserRequest) error {
    if req.Age < 18 {
        return errno.ErrUserAgeTooYoung
    }
    return b.ds.Users().Create(ctx, user)
}

// Store layer only does data operations
func (s *userStore) Create(ctx context.Context, user *model.User) error {
    return s.db.Create(user).Error
}
```

## Next Steps

- [Develop Your First Feature](../guide/first-feature.md) - Practice layered architecture
- [Development Standards](../development/standards.md) - Code style and conventions
- [Testing Guide](../development/testing.md) - How to test each layer (coming soon)
