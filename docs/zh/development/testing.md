---
title: 测试指南 - Bingo Go 项目测试规范
description: 了解 Bingo Go 微服务项目的测试规范，包括分层测试策略、Mock 组织和测试风格。
---

# 测试指南

## 快速开始

### 三层测试策略

| 层级 | 测试方式 | Mock 什么 | 用真实数据库？ |
|------|----------|-----------|---------------|
| **Store 层** | SQLite 内存数据库 | 不 mock | ✅ 用 SQLite |
| **Biz 层** | Mock Store | Mock `store.IStore` | ❌ 不用 |
| **Handler 层** | Mock Biz | Mock `biz.IBiz` | ❌ 不用 |

```
┌─────────────────────────────────────────────────────────────┐
│ Handler 层测试                                               │
│   - httptest 发起请求                                        │
│   - Mock biz.IBiz                                           │
│   - 验证：参数解析、响应格式、错误处理                         │
└───────────────────────────┬─────────────────────────────────┘
                            │ Mock
┌───────────────────────────▼─────────────────────────────────┐
│ Biz 层测试                                                   │
│   - Mock store.IStore (内存 map 实现)                        │
│   - 验证：业务逻辑、流程编排、事务控制                         │
└───────────────────────────┬─────────────────────────────────┘
                            │ Mock
┌───────────────────────────▼─────────────────────────────────┐
│ Store 层测试                                                 │
│   - SQLite 内存数据库                                        │
│   - 验证：SQL 查询逻辑、GORM 行为、边界条件                    │
└─────────────────────────────────────────────────────────────┘
```

### 核心原则

- **Store 层**：测试真实 SQL 逻辑，不 mock 数据库
- **Biz 层**：只测业务逻辑，mock 掉 Store 依赖
- **Handler 层**：只测参数解析和响应格式，mock 掉 Biz 依赖
- **每层只测自己的职责**，不要跨层测试

## Store 层测试

Store 层使用 SQLite 内存数据库测试，不使用 Mock。

**为什么用 SQLite 而不是 Mock？**
- Store 层的核心职责是 SQL 查询逻辑
- Mock 会导致"测试 mock 行为而非真实逻辑"
- SQLite 能验证真实的 SQL 语法和 GORM 行为

### 测试模板

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

### 注意事项

1. **优先用 AutoMigrate**：GORM AutoMigrate 在 SQLite 上正常工作
2. **每个测试独立**：使用 `:memory:` 确保测试隔离
3. **MySQL 特有功能**：如用到 `JSON_EXTRACT` 等 MySQL 特有函数，标记为集成测试

## Biz 层测试

Biz 层使用 Mock Store 测试业务逻辑，不依赖真实数据库。

**测试重点**：
- 业务规则验证
- 流程编排逻辑
- 异常路径处理

### 测试模板

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

## Handler 层测试

Handler 层使用 Mock Biz 测试，验证参数解析、响应格式和错误处理。

**测试重点**：
- 请求参数绑定和验证
- 响应格式正确性
- 错误码映射

### 测试模板

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
        // 验证响应 JSON 结构
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

## Mock 代码组织

所有 mock 代码统一放在 `internal/pkg/testing/mock/` 目录。

### 目录结构

```
internal/pkg/testing/mock/
├── store/           # Store 层 mock
│   └── store.go     # 实现 store.IStore 接口
├── biz/             # Biz 层 mock
│   └── biz.go       # 实现 biz.IBiz 接口
└── ...              # 其他模块 mock
```

### 实现规范

1. **ABOUTME 注释**：每个 mock 文件必须有 ABOUTME 注释说明用途

2. **接口验证**：用编译时检查确保实现完整
   ```go
   var _ store.IStore = (*Store)(nil)
   ```

3. **可配置错误**：支持注入错误用于测试异常路径
   ```go
   type UserStore struct {
       CreateErr error
       GetErr    error
       GetResult *model.User
   }
   ```

4. **状态暴露**：提供 helper 方法暴露内部状态用于断言
   ```go
   func (m *UserStore) Users() map[uint64]*model.User
   ```

### 新增 Mock 步骤

1. 在 `internal/pkg/testing/mock/` 下找到或创建对应模块目录
2. 创建 mock 文件，添加 ABOUTME 注释
3. 实现目标接口，添加编译时检查
4. 添加可配置错误字段和状态暴露方法

## 测试工具

| 工具 | 用途 | 说明 |
|------|------|------|
| **testify** | 断言 | `assert`（继续执行）/`require`（立即终止）|
| **SQLite** | Store 层测试 | 内存数据库，验证真实 SQL 行为 |
| **testcontainers-go** | 集成测试 | 动态启动 Docker 容器（MySQL 等） |

## 测试风格

项目支持两种测试风格，根据场景选择：

### Table-Driven Tests（推荐用于纯函数）

适用于输入输出明确、多组数据验证的场景：

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

**命名规范**：
- 测试用例 name 使用 snake_case（如 `valid_input`、`missing_chain_id`）
- 变量使用 `tt` 表示当前测试用例
- 使用 `want`/`wantErr` 表示期望结果

### 子测试风格（推荐用于复杂场景）

适用于需要共享 setup、多分支条件的场景：

```go
func TestMiddleware(t *testing.T) {
    // 共享 setup
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

## 集成测试

集成测试验证多组件协作，使用真实依赖（如 MySQL、Redis）。

### 标记方式

使用 `-short` 标记区分单元测试和集成测试：

```go
func TestUserStore_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // 使用真实数据库的测试代码
}
```

### 运行命令

```bash
go test ./...              # 运行所有测试
go test -short ./...       # 只运行单元测试（跳过集成测试）
go test -run Integration   # 只运行集成测试
```

## CI/CD 集成

推荐 GitHub Actions，分离单元测试和集成测试：

```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.24'

    # 单元测试（快速，每次都跑）
    - name: Unit Tests
      run: go test -short -race ./...

    # 集成测试（需要 Docker）
    - name: Integration Tests
      run: go test -run Integration ./...
```

## 下一步

- [Docker 部署](../deployment/docker.md) - 使用 Docker 部署 Bingo 项目
- [微服务拆分](../advanced/microservices.md) - 了解如何将单体应用拆分为微服务
