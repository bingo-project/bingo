# Store Package Design

## Overview

The Store package is the core of Bingo's data access layer. It implements a flexible and extensible data access framework through **generics** and **composition pattern**, reducing code duplication and improving code reusability.

> **Note**: This document describes the universal `Store[T]` design in `pkg/store`, which provides reusable data access foundation. Business-specific Store implementations are in `internal/pkg/store`, extending Store[T] through composition and implementing business-specific extension interfaces (see [Project Structure](../guide/project-structure.md)).

## Package Structure

```
pkg/store/
├── store.go          # Universal Store[T] implementation
├── logger.go         # Logger interface definition
└── where/
    └── where.go      # Query condition builder

internal/pkg/store/
├── store.go          # IStore interface and application-level implementation (datastore)
├── logger.go         # Business logger implementation
└── <model>.go        # Business Store implementations
```

> **Note**: All files in internal/pkg/store must be flat in the same directory to avoid circular imports. Use naming conventions rather than directory structure to organize modules.

## Naming Convention

All files in `internal/pkg/store` must be flat (avoiding circular imports), using naming conventions to organize code.

### Convention

- **File names**: `<prefix>_<model>.go` (e.g., `sys_admin.go`, `bot_channel.go`, `user.go`)
  - System modules: `sys_` prefix
  - Other modules: module name prefix (e.g., `bot_`, `api_`)
  - Can omit prefix when no conflict (e.g., `user.go`)

- **Store interface**: `<Prefix><Model>Store` (e.g., `BotChannelStore`, `AdminStore`)
  - Add module prefix only when needed to distinguish from same-name models

- **Implementation struct**: lowercase `<prefix><model>Store` (e.g., `botChannelStore`)

- **Extension interface**: `<Prefix><Model>Expansion` (e.g., `BotChannelExpansion`)

- **Factory function**: `New<Prefix><Model>Store()`

- **IStore methods**: Plural or singular, keep it simple (e.g., `Users()`, `Bot()`, `BotChannel()`)

## Core Design

### 1. Generics + Composition Pattern

Universal `Store[T]` implements all CRUD operations. Business-specific Stores extend through composition:

```go
// pkg/store - Universal implementation
type Store[T any] struct {
    logger  Logger
    storage DBProvider
}

// internal/pkg/store - Business extension
type userStore struct {
    *genericstore.Store[User]
}

type UserStore interface {
    Create(ctx context.Context, obj *User) error
    Update(ctx context.Context, obj *User, fields ...string) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*User, error)
    List(ctx context.Context, opts *where.Options) (int64, []*User, error)

    UserExpansion  // Business-specific extension interface
}
```

### 2. Condition Builder

Use `where` package for fluent API to build query conditions:

```go
// Pagination query
opts := where.F("status", "active").P(1, 10)

// Complex conditions
opts := where.NewWhere().
    F("status", "active").
    Q("created_at > ?", time.Now().AddDate(0, 0, -7)).
    Load("User", "Tags").
    P(1, 20)

// Convenience functions
opts := where.P(1, 10)              // Pagination
opts := where.F("field", value)      // Filter
opts := where.Load("Association")    // Preload
```

Supported operations:
- `F(kvs...)` - Filter conditions
- `Q(query, args...)` - Custom SQL
- `P(page, pageSize)` - Pagination
- `O(offset)` / `L(limit)` - Offset and limit
- `C(clauses...)` - GORM clauses
- `Load(associations...)` - Preload associations

### 3. Transaction Context

Automatically handle transactions through context. Store layer transparently supports transactions:

```go
// internal/pkg/store/store.go
func (ds *datastore) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
    db := ds.core

    // Automatically extract transaction from context
    if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
        db = tx
    }

    // Apply query conditions
    for _, whr := range wheres {
        db = whr.Where(db)
    }
    return db
}

// Transaction API
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
    return ds.core.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            ctx = context.WithValue(ctx, transactionKey{}, tx)
            return fn(ctx)
        },
    )
}
```

## API Reference

### Store[T] Methods

```go
// CRUD operations
Create(ctx context.Context, obj *T) error
Update(ctx context.Context, obj *T, fields ...string) error
Delete(ctx context.Context, opts *where.Options) error
Get(ctx context.Context, opts *where.Options) (*T, error)

// Query operations
List(ctx context.Context, opts *where.Options) (count int64, ret []*T, err error)
Find(ctx context.Context, opts *where.Options) (ret []*T, err error)
Last(ctx context.Context, opts *where.Options) (*T, error)

// Batch and conditional operations
CreateInBatch(ctx context.Context, objs []*T, batchSize int) error
CreateIfNotExist(ctx context.Context, obj *T) error
FirstOrCreate(ctx context.Context, where any, obj *T) error
UpdateOrCreate(ctx context.Context, where any, obj *T) error
Upsert(ctx context.Context, obj *T, fields ...string) error
DeleteInBatch(ctx context.Context, ids []uint) error

// Raw database access
DB(ctx context.Context, wheres ...where.Where) *gorm.DB
```

**Method Description**:

- **CreateIfNotExist**: Create object if not exists (using OnConflict DoNothing)
- **FirstOrCreate**: Find object by condition, create if not exists
- **UpdateOrCreate**: Update or create object in transaction, supports optimistic locking
- **DeleteInBatch**: Batch delete objects by IDs

### IStore Interface

IStore is the application-level unified data access interface responsible for returning each Store implementation. The interface adopts a modular design, returning corresponding Stores through methods:

```go
type IStore interface {
    // Transaction and database
    DB(ctx context.Context, wheres ...where.Where) *gorm.DB
    TX(ctx context.Context, fn func(ctx context.Context) error) error

    // Business Store methods (organized by module)
    // Example: Users() UserStore, Admin() AdminStore, etc.
}
```

## Usage Examples

This chapter demonstrates basic usage of Store through examples.

### Basic CRUD Operations

Assume we have a `User` model with basic operations through Store:

```go
// Create
user := &User{Name: "John", Email: "john@example.com"}
err := store.Users().Create(ctx, user)

// Read
user, err := store.Users().Get(ctx, where.F("id", 1))

// Update (update only specified fields)
user.Email = "newemail@example.com"
err := store.Users().Update(ctx, user, "email")

// Delete
err := store.Users().Delete(ctx, where.F("id", 1))
```

### Query and Pagination

```go
// Build query conditions
opts := where.F("status", "active").
    P(1, 10)  // Page 1, 10 items per page

// Execute query
count, users, err := store.Users().List(ctx, opts)
```

### Transaction Processing

Use `TX()` method when multiple operations need atomic guarantees:

```go
err := store.TX(ctx, func(ctx context.Context) error {
    // Store automatically uses transaction
    if err := store.Users().Create(ctx, user1); err != nil {
        return err  // Auto rollback
    }
    if err := store.Users().Create(ctx, user2); err != nil {
        return err  // Auto rollback
    }
    return nil  // Auto commit
})
```

### Extension Operations

Business-specific Stores can add custom operations through extension interfaces:

```go
// internal/pkg/store/user.go
type UserExpansion interface {
    FindByEmail(ctx context.Context, email string) (*User, error)
}

func (s *userStore) FindByEmail(ctx context.Context, email string) (*User, error) {
    return s.Get(ctx, where.F("email", email))
}
```

## Adding New Business Store

Adding a new Store requires following these steps and naming conventions (see "Naming Convention" section):

### 1. Create Store Interface and Implementation

```go
// internal/pkg/store/user.go
package store

// Store interface defines CRUD operations
type UserStore interface {
    Create(ctx context.Context, obj *User) error
    Update(ctx context.Context, obj *User, fields ...string) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*User, error)
    List(ctx context.Context, opts *where.Options) (int64, []*User, error)

    UserExpansion  // Extension interface
}

// Extension interface defines business-specific operations
type UserExpansion interface {
    FindByEmail(ctx context.Context, email string) (*User, error)
}

// Implementation class
type userStore struct {
    *genericstore.Store[User]
}

// Factory function
func NewUserStore(store *datastore) *userStore {
    return &userStore{
        Store: genericstore.NewStore[User](store, NewLogger()),
    }
}

// Implement extension method
func (s *userStore) FindByEmail(ctx context.Context, email string) (*User, error) {
    return s.Get(ctx, where.F("email", email))
}
```

### 2. Register to IStore

Add method in `internal/pkg/store/store.go`:

```go
type IStore interface {
    Users() UserStore  // New addition
    // ...
}

func (ds *datastore) Users() UserStore {
    return NewUserStore(ds)
}
```

## Related Content

- [Layered Architecture in Detail](./layered-design.md) - Understand Store layer's role in three-layer architecture
- [Overall Architecture](./architecture.md) - Data access design in microservice architecture
- [Store Naming Convention and Best Practices](../development/standards.md#store-naming-convention) - Development standards
- [Develop Your First Feature](../guide/first-feature.md) - Practical application example
- [Database Layer](../components/database.md) - GORM usage guide (coming soon)
