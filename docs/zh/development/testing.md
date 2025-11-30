---
title: 测试指南 - Bingo Go 项目测试规范
description: 了解 Bingo Go 微服务项目的测试规范，包括测试分类、测试工具、测试风格、Mock 策略和 CI/CD 集成。
---

# 测试指南

## 测试分类

| 类型 | 测试范围 | 依赖处理 | 速度 | 目的 |
|------|---------|---------|------|------|
| **单元测试** | 单个函数/方法 | Mock 隔离 | 毫秒级 | 验证逻辑正确 |
| **集成测试** | 多组件协作 | 真实依赖 | 秒级 | 验证交互正确 |
| **E2E 测试** | 完整流程 | 全真实环境 | 分钟级 | 验证业务流程 |

### 测试金字塔

```
        /\
       /  \      E2E（少量，最慢）
      /────\
     /      \    集成测试（适量）
    /────────\
   /          \  单元测试（大量，最快）
  /────────────\
```

**原则**：底层测试多，上层测试少。单元测试覆盖逻辑分支，集成测试覆盖关键路径。

## 测试工具

| 工具 | 用途 | 说明 |
|------|------|------|
| **testify** | 断言 | `assert`（继续执行）/`require`（立即终止）|
| **goconvey** | BDD 风格测试 | 嵌套 `Convey` 描述场景，`So` 断言 |
| **gomonkey** | 运行时 Mock | Monkey patching，支持私有方法 |
| **sqlmock** | Store 单元测试 | 验证 SQL 结构正确性，不需真实 DB |
| **testcontainers-go** | 集成测试 | 动态启动 Docker 容器（MySQL 等） |

## 测试规范

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

### GoConvey BDD 风格（推荐用于复杂场景）

适用于需要共享 setup、多分支条件、HTTP 中间件等复杂场景：

```go
func TestMiddleware(t *testing.T) {
    Convey("TestMiddleware", t, func() {
        // 共享 setup
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
            // 使用 gomonkey mock 依赖
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

**GoConvey 规范**：
- 外层 `Convey` 第一个参数与函数名一致
- 内层 `Convey` 描述具体场景
- 使用 `So(actual, ShouldXxx, expected)` 断言

### 集成测试标记

使用 `-short` 标记区分单元测试和集成测试：

```go
func TestStore_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    // 集成测试代码
}
```

运行命令：
```bash
go test ./...              # 运行所有测试
go test -short ./...       # 只运行单元测试
go test -run Integration   # 只运行集成测试
```

## Mock 策略

### 方式一：gomonkey（运行时 Mock）

适用于 mock 私有方法、第三方库、全局函数：

```go
import "github.com/agiledragon/gomonkey/v2"

// Mock 结构体方法（包括私有方法）
patches := gomonkey.ApplyPrivateMethod(store.S.Users(), "Get",
    func(ctx context.Context, id uint64) (*model.User, error) {
        return &model.User{ID: id, Name: "test"}, nil
    })
defer patches.Reset()

// Mock 全局函数
patches := gomonkey.ApplyFunc(time.Now, func() time.Time {
    return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
})
defer patches.Reset()
```

**注意**：gomonkey 在 macOS ARM64 需要禁用内联优化：
```bash
go test -gcflags="all=-N -l" ./...
```

### 方式二：接口 Mock（依赖注入）

适用于通过接口定义的依赖，参考 k8s client-go/testing 设计。

#### 分层规范

| 层级 | 位置 | 适用场景 |
|------|------|----------|
| 共享 mock | `internal/pkg/testing/mock/` | 多包需要（IStore, BlockchainClient） |
| 包私有 mock | `xxx_test.go` | 单包专用（mockAddressCache 等） |

#### 命名规范

- 共享 mock：`mock.Xxx`（导出，供其他包 `_test.go` 使用）
- 包私有 mock：`mockXxx`（不导出，仅本包测试使用）

#### 目录结构

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

## Store 层测试策略

| 层级 | 工具 | 目的 |
|------|------|------|
| 单元测试 | sqlmock | 验证 GORM 生成的 SQL 结构 |
| 集成测试 | testcontainers-go | 验证真实 MySQL 行为 |

### sqlmock 示例

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
