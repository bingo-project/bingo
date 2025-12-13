---
title: Testing Guide - Bingo Go Project Testing Standards
description: Learn about Bingo Go microservices project testing standards, including layered testing strategy, mock organization, and test styles.
---

# Testing Guide

## Quick Start

### Three-Layer Testing Strategy

| Layer | Test Method | What to Mock | Use Real Database? |
|-------|-------------|--------------|-------------------|
| **Store Layer** | SQLite in-memory database | No mock | ✅ Use SQLite |
| **Biz Layer** | Mock Store | Mock `store.IStore` | ❌ No |
| **Handler Layer** | Mock Biz | Mock `biz.IBiz` | ❌ No |

```
┌─────────────────────────────────────────────────────────────┐
│ Handler Layer Tests                                          │
│   - Use httptest for requests                                │
│   - Mock biz.IBiz                                           │
│   - Verify: parameter parsing, response format, error handling│
└───────────────────────────┬─────────────────────────────────┘
                            │ Mock
┌───────────────────────────▼─────────────────────────────────┐
│ Biz Layer Tests                                              │
│   - Mock store.IStore (in-memory map implementation)         │
│   - Verify: business logic, flow orchestration, transactions │
└───────────────────────────┬─────────────────────────────────┘
                            │ Mock
┌───────────────────────────▼─────────────────────────────────┐
│ Store Layer Tests                                            │
│   - SQLite in-memory database                                │
│   - Verify: SQL query logic, GORM behavior, edge cases       │
└─────────────────────────────────────────────────────────────┘
```

### Core Principles

- **Store Layer**: Test real SQL logic, don't mock the database
- **Biz Layer**: Test business logic only, mock Store dependencies
- **Handler Layer**: Test parameter parsing and response format only, mock Biz dependencies
- **Each layer tests only its own responsibilities**, no cross-layer testing

## Store Layer Testing

Store layer uses SQLite in-memory database for testing, no mocks.

**Why SQLite instead of Mock?**
- Store layer's core responsibility is SQL query logic
- Mocks lead to "testing mock behavior instead of real logic"
- SQLite can verify real SQL syntax and GORM behavior

### Test Template

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)

    err = db.AutoMigrate(&model.User{})
    require.NoError(t, err)

    return db
}

func TestUserStore_Create(t *testing.T) {
    db := setupTestDB(t)
    s := NewUserStore(&datastore{db: db})
    ctx := context.Background()

    t.Run("success", func(t *testing.T) {
        user := &model.User{Username: "test", Email: "test@example.com"}
        err := s.Create(ctx, user)

        require.NoError(t, err)
        assert.NotZero(t, user.ID)
    })

    t.Run("duplicate_username", func(t *testing.T) {
        user := &model.User{Username: "test", Email: "other@example.com"}
        err := s.Create(ctx, user)

        assert.Error(t, err)
    })
}
```

### Notes

1. **Prefer AutoMigrate**: GORM AutoMigrate works normally on SQLite
2. **Each test is independent**: Use `:memory:` to ensure test isolation
3. **MySQL-specific features**: If using MySQL-specific functions like `JSON_EXTRACT`, mark as integration test

## Biz Layer Testing

Biz layer uses Mock Store to test business logic, no real database dependency.

**Test Focus**:
- Business rule validation
- Flow orchestration logic
- Exception path handling

### Test Template

```go
import mockstore "your-project/internal/pkg/testing/mock/store"

func TestUserBiz_Create(t *testing.T) {
    store := mockstore.NewStore()
    biz := user.New(store)
    ctx := context.Background()

    t.Run("success", func(t *testing.T) {
        req := &CreateUserRequest{
            Username: "test",
            Age:      20,
        }

        user, err := biz.Create(ctx, req)

        require.NoError(t, err)
        assert.Equal(t, "test", user.Username)
    })

    t.Run("age_too_young", func(t *testing.T) {
        req := &CreateUserRequest{
            Username: "test",
            Age:      16,
        }

        _, err := biz.Create(ctx, req)

        assert.ErrorIs(t, err, errno.ErrUserAgeTooYoung)
    })

    t.Run("store_error", func(t *testing.T) {
        store.UserStore().CreateErr = errors.New("db error")
        defer func() { store.UserStore().CreateErr = nil }()

        req := &CreateUserRequest{Username: "test", Age: 20}
        _, err := biz.Create(ctx, req)

        assert.Error(t, err)
    })
}
```

## Handler Layer Testing

Handler layer uses Mock Biz for testing, verifying parameter parsing, response format, and error handling.

**Test Focus**:
- Request parameter binding and validation
- Response format correctness
- Error code mapping

### Test Template

```go
import mockbiz "your-project/internal/pkg/testing/mock/biz"

func TestUserHandler_Get(t *testing.T) {
    biz := mockbiz.NewBiz()
    handler := NewUserHandler(biz)

    t.Run("success", func(t *testing.T) {
        biz.UserBiz().GetResult = &model.User{ID: 1, Username: "test"}

        w := httptest.NewRecorder()
        ctx, _ := gin.CreateTestContext(w)
        ctx.Params = gin.Params{{Key: "id", Value: "1"}}
        ctx.Request, _ = http.NewRequest("GET", "/users/1", nil)

        handler.Get(ctx)

        assert.Equal(t, http.StatusOK, w.Code)
        // Verify response JSON structure
    })

    t.Run("invalid_id", func(t *testing.T) {
        w := httptest.NewRecorder()
        ctx, _ := gin.CreateTestContext(w)
        ctx.Params = gin.Params{{Key: "id", Value: "abc"}}
        ctx.Request, _ = http.NewRequest("GET", "/users/abc", nil)

        handler.Get(ctx)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

    t.Run("not_found", func(t *testing.T) {
        biz.UserBiz().GetErr = errno.ErrUserNotFound

        w := httptest.NewRecorder()
        ctx, _ := gin.CreateTestContext(w)
        ctx.Params = gin.Params{{Key: "id", Value: "999"}}
        ctx.Request, _ = http.NewRequest("GET", "/users/999", nil)

        handler.Get(ctx)

        assert.Equal(t, http.StatusNotFound, w.Code)
    })
}
```

## Mock Code Organization

All mock code is centralized in `internal/pkg/testing/mock/` directory.

### Directory Structure

```
internal/pkg/testing/mock/
├── store/           # Store layer mock
│   └── store.go     # Implements store.IStore interface
├── biz/             # Biz layer mock
│   └── biz.go       # Implements biz.IBiz interface
└── ...              # Other module mocks
```

### Implementation Standards

1. **ABOUTME comments**: Each mock file must have ABOUTME comments explaining its purpose

2. **Interface verification**: Use compile-time checks to ensure complete implementation
   ```go
   var _ store.IStore = (*Store)(nil)
   ```

3. **Configurable errors**: Support error injection for testing exception paths
   ```go
   type UserStore struct {
       CreateErr error
       GetErr    error
       GetResult *model.User
   }
   ```

4. **State exposure**: Provide helper methods to expose internal state for assertions
   ```go
   func (m *UserStore) Users() map[uint64]*model.User
   ```

### Steps to Add New Mock

1. Find or create the corresponding module directory under `internal/pkg/testing/mock/`
2. Create mock file with ABOUTME comments
3. Implement target interface with compile-time check
4. Add configurable error fields and state exposure methods

## Testing Tools

| Tool | Purpose | Description |
|------|---------|-------------|
| **testify** | Assertions | `assert` (continues) / `require` (stops immediately) |
| **SQLite** | Store layer tests | In-memory database, verify real SQL behavior |
| **testcontainers-go** | Integration tests | Dynamically start Docker containers (MySQL, etc.) |

## Test Styles

The project supports two testing styles, choose based on scenario:

### Table-Driven Tests (Recommended for Pure Functions)

Suitable for scenarios with clear input/output and multiple data validations:

```go
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name:  "valid_input",
            input: InputType{...},
            want:  OutputType{...},
        },
        {
            name:    "invalid_input",
            input:   InputType{...},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FuncUnderTest(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Naming Conventions**:
- Test case name uses snake_case (e.g., `valid_input`, `missing_chain_id`)
- Use `tt` for current test case variable
- Use `want`/`wantErr` for expected results

### Subtest Style (Recommended for Complex Scenarios)

Suitable for scenarios requiring shared setup, multiple branches:

```go
func TestMiddleware(t *testing.T) {
    // Shared setup
    setupTest := func() (*httptest.ResponseRecorder, *gin.Context) {
        w := httptest.NewRecorder()
        ctx, _ := gin.CreateTestContext(w)
        return w, ctx
    }

    t.Run("disabled", func(t *testing.T) {
        w, ctx := setupTest()
        config.Enabled = false

        Middleware()(ctx)

        assert.False(t, ctx.IsAborted())
    })

    t.Run("invalid_request", func(t *testing.T) {
        w, ctx := setupTest()
        ctx.Request, _ = http.NewRequest("GET", "/path", nil)

        Middleware()(ctx)

        assert.True(t, ctx.IsAborted())
    })
}
```

## Integration Tests

Integration tests verify multi-component collaboration using real dependencies (like MySQL, Redis).

### Marking Method

Use `-short` flag to distinguish unit tests from integration tests:

```go
func TestUserStore_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // Test code using real database
}
```

### Run Commands

```bash
go test ./...              # Run all tests
go test -short ./...       # Run unit tests only (skip integration tests)
go test -run Integration   # Run integration tests only
```

## CI/CD Integration

Recommended GitHub Actions with separate unit and integration tests:

```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'

    # Unit tests (fast, run every time)
    - name: Unit Tests
      run: go test -short -race ./...

    # Integration tests (requires Docker)
    - name: Integration Tests
      run: go test -run Integration ./...
```

## Next Steps

- [Docker Deployment](../deployment/docker.md) - Deploy Bingo projects using Docker
- [Microservice Decomposition](../advanced/microservices.md) - Learn how to decompose monolith into microservices
