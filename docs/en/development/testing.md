---
title: Testing Guide - Bingo Go Project Testing Standards
description: Learn about Bingo Go microservices project testing standards, including test categories, testing tools, test styles, mock strategies, and CI/CD integration.
---

# Testing Guide

## Test Categories

| Type | Scope | Dependencies | Speed | Purpose |
|------|-------|--------------|-------|---------|
| **Unit Test** | Single function/method | Mock isolation | Milliseconds | Verify logic |
| **Integration Test** | Multi-component | Real dependencies | Seconds | Verify interaction |
| **E2E Test** | Complete flow | Full real environment | Minutes | Verify business flow |

### Test Pyramid

```
        /\
       /  \      E2E (few, slowest)
      /────\
     /      \    Integration (moderate)
    /────────\
   /          \  Unit tests (many, fastest)
  /────────────\
```

**Principle**: More tests at the bottom, fewer at the top. Unit tests cover logic branches, integration tests cover critical paths.

## Testing Tools

| Tool | Purpose | Description |
|------|---------|-------------|
| **testify** | Assertions | `assert` (continues) / `require` (stops immediately) |
| **goconvey** | BDD style testing | Nested `Convey` for scenarios, `So` for assertions |
| **gomonkey** | Runtime Mock | Monkey patching, supports private methods |
| **sqlmock** | Store unit tests | Verify SQL structure without real DB |
| **testcontainers-go** | Integration tests | Dynamically start Docker containers (MySQL, etc.) |

## Testing Standards

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

### GoConvey BDD Style (Recommended for Complex Scenarios)

Suitable for scenarios requiring shared setup, multiple branches, HTTP middleware, etc.:

```go
func TestMiddleware(t *testing.T) {
    Convey("TestMiddleware", t, func() {
        // Shared setup
        w := httptest.NewRecorder()
        ctx, _ := gin.CreateTestContext(w)

        Convey("disabled", func() {
            config.Enabled = false
            Middleware()(ctx)
            So(ctx.IsAborted(), ShouldBeFalse)
        })

        Convey("invalid request", func() {
            ctx.Request, _ = http.NewRequest("GET", "/path", nil)
            Middleware()(ctx)
            So(ctx.IsAborted(), ShouldBeTrue)
        })

        Convey("valid request", func() {
            // Use gomonkey to mock dependencies
            patches := gomonkey.ApplyPrivateMethod(store.S.Users(), "Get",
                func(ctx context.Context, id uint64) (*model.User, error) {
                    return &model.User{ID: 1}, nil
                })
            defer patches.Reset()

            Middleware()(ctx)
            So(ctx.IsAborted(), ShouldBeFalse)
        })
    })
}
```

**GoConvey Conventions**:
- Outer `Convey` first parameter matches function name
- Inner `Convey` describes specific scenarios
- Use `So(actual, ShouldXxx, expected)` for assertions

### Integration Test Markers

Use `-short` flag to distinguish unit tests from integration tests:

```go
func TestStore_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // Integration test code
}
```

Run commands:
```bash
go test ./...              # Run all tests
go test -short ./...       # Run unit tests only
go test -run Integration   # Run integration tests only
```

## Mock Strategies

### Option 1: gomonkey (Runtime Mock)

Suitable for mocking private methods, third-party libraries, global functions:

```go
import "github.com/agiledragon/gomonkey/v2"

// Mock struct method (including private methods)
patches := gomonkey.ApplyPrivateMethod(store.S.Users(), "Get",
    func(ctx context.Context, id uint64) (*model.User, error) {
        return &model.User{ID: id, Name: "test"}, nil
    })
defer patches.Reset()

// Mock global function
patches := gomonkey.ApplyFunc(time.Now, func() time.Time {
    return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
})
defer patches.Reset()
```

**Note**: gomonkey on macOS ARM64 requires disabling inline optimization:
```bash
go test -gcflags="all=-N -l" ./...
```

### Option 2: Interface Mock (Dependency Injection)

Suitable for dependencies defined through interfaces, following k8s client-go/testing design.

#### Layering Standards

| Level | Location | Use Case |
|-------|----------|----------|
| Shared mock | `internal/pkg/testing/mock/` | Multi-package needs (IStore, BlockchainClient) |
| Package-private mock | `xxx_test.go` | Single package only (mockAddressCache, etc.) |

#### Naming Conventions

- Shared mock: `mock.Xxx` (exported, for other packages' `_test.go` files)
- Package-private mock: `mockXxx` (unexported, only for current package tests)

#### Directory Structure

```
internal/pkg/testing/
└── mock/
    ├── store.go              # mock.Store (IStore)
    ├── blockchain.go         # mock.BlockchainClient
    ├── account.go            # mock.AccountService
    └── scanner/
        ├── loader.go         # mock.AddressLoader
        └── cache.go          # mock.AddressCache, mock.TokenCache
```

## Store Layer Testing Strategy

| Level | Tool | Purpose |
|-------|------|---------|
| Unit Test | sqlmock | Verify GORM-generated SQL structure |
| Integration Test | testcontainers-go | Verify real MySQL behavior |

### sqlmock Example

```go
func TestUserStore_GetByID(t *testing.T) {
    db, mock, _ := sqlmock.New()
    gormDB, _ := gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{})

    mock.ExpectQuery("SELECT .* FROM `users` WHERE `id` = ?").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "test"))

    store := NewUserStore(gormDB)
    user, err := store.GetByID(context.Background(), 1)

    require.NoError(t, err)
    assert.Equal(t, "test", user.Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
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

## Next Step

- [Docker Deployment](../deployment/docker.md) - Deploy Bingo projects using Docker
- [Microservice Decomposition](../advanced/microservices.md) - Learn how to decompose monolith into microservices
