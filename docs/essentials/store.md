# Store 包设计

## 概述

Store 包是 Bingo 数据访问层的核心，通过**泛型**和**组合模式**实现了一个灵活、可扩展的数据访问框架，减少重复代码并提高代码复用率。

> **注意**: 本文档说明的是 `pkg/store` 中的通用 Store[T] 设计，它提供了可复用的数据访问基础。具体业务 Store 实现在 `internal/pkg/store` 中，通过组合 Store[T] 并实现业务特定的扩展接口来实现功能（详见[项目结构](../guide/project-structure.md)）。

## 包结构

```
pkg/store/
├── store.go          # 通用 Store[T] 实现
├── logger.go         # Logger 接口定义
└── where/
    └── where.go      # 查询条件构建器

internal/pkg/store/
├── store.go          # IStore 接口和应用级实现 (datastore)
├── logger.go         # 业务日志实现
└── <model>.go        # 各个业务 Store 实现
```

> **注意**: internal/pkg/store 中所有文件必须平铺在同一目录中，避免循环引用。使用命名规范而非目录结构来组织模块。

## 命名规范

`internal/pkg/store` 中所有文件必须平铺（避免循环引用），使用命名规范来组织代码。

### 规范

- **文件名**: `<prefix>_<model>.go` (如 `sys_admin.go`, `bot_channel.go`, `user.go`)
  - 系统模块: `sys_` 前缀
  - 其他模块: 模块名前缀（如 `bot_`, `api_`）
  - 无冲突时可省略前缀（如 `user.go`）

- **Store 接口**: `<Prefix><Model>Store` (如 `BotChannelStore`, `AdminStore`)
  - 需要与同名模型区分时才加模块前缀

- **实现结构体**: 小写 `<prefix><model>Store` (如 `botChannelStore`)

- **扩展接口**: `<Prefix><Model>Expansion` (如 `BotChannelExpansion`)

- **创建函数**: `New<Prefix><Model>Store()`

- **IStore 方法**: 复数或单数，保持简洁（如 `Users()`, `Bot()`, `BotChannel()`）

## 核心设计

### 1. 泛型 + 组合模式

通用的 `Store[T]` 实现所有 CRUD 操作，业务特定的 Store 通过组合来扩展：

```go
// pkg/store - 通用实现
type Store[T any] struct {
    logger  Logger
    storage DBProvider
}

// internal/pkg/store - 业务扩展
type userStore struct {
    *genericstore.Store[User]
}

type UserStore interface {
    Create(ctx context.Context, obj *User) error
    Update(ctx context.Context, obj *User, fields ...string) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*User, error)
    List(ctx context.Context, opts *where.Options) (int64, []*User, error)

    UserExpansion  // 业务特定的扩展接口
}
```

### 2. 条件构建器

使用 `where` 包提供链式 API 构建查询条件：

```go
// 分页查询
opts := where.F("status", "active").P(1, 10)

// 复杂条件
opts := where.NewWhere().
    F("status", "active").
    Q("created_at > ?", time.Now().AddDate(0, 0, -7)).
    Load("User", "Tags").
    P(1, 20)

// 便捷函数
opts := where.P(1, 10)              // 分页
opts := where.F("field", value)      // 过滤
opts := where.Load("Association")    // 预加载
```

支持的操作：
- `F(kvs...)` - 过滤条件
- `Q(query, args...)` - 自定义 SQL
- `P(page, pageSize)` - 分页
- `O(offset)` / `L(limit)` - 偏移和限制
- `C(clauses...)` - GORM 子句
- `Load(associations...)` - 预加载关联

### 3. 事务上下文

通过 context 自动处理事务，Store 层透明支持：

```go
// internal/pkg/store/store.go
func (ds *datastore) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
    db := ds.core

    // 自动从上下文提取事务
    if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
        db = tx
    }

    // 应用查询条件
    for _, whr := range wheres {
        db = whr.Where(db)
    }
    return db
}

// 事务 API
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
    return ds.core.WithContext(ctx).Transaction(
        func(tx *gorm.DB) error {
            ctx = context.WithValue(ctx, transactionKey{}, tx)
            return fn(ctx)
        },
    )
}
```

## API 说明

### Store[T] 方法

```go
// CRUD 操作
Create(ctx context.Context, obj *T) error
Update(ctx context.Context, obj *T, fields ...string) error
Delete(ctx context.Context, opts *where.Options) error
Get(ctx context.Context, opts *where.Options) (*T, error)

// 查询操作
List(ctx context.Context, opts *where.Options) (count int64, ret []*T, err error)
Find(ctx context.Context, opts *where.Options) (ret []*T, err error)
Last(ctx context.Context, opts *where.Options) (*T, error)

// 批量和条件操作
CreateInBatch(ctx context.Context, objs []*T, batchSize int) error
CreateIfNotExist(ctx context.Context, obj *T) error
FirstOrCreate(ctx context.Context, where any, obj *T) error
UpdateOrCreate(ctx context.Context, where any, obj *T) error
Upsert(ctx context.Context, obj *T, fields ...string) error
DeleteInBatch(ctx context.Context, ids []uint) error

// 原始数据库访问
DB(ctx context.Context, wheres ...where.Where) *gorm.DB
```

**方法说明**：

- **CreateIfNotExist**: 创建对象，如果已存在则忽略（使用 OnConflict DoNothing）
- **FirstOrCreate**: 根据条件查找对象，不存在则创建
- **UpdateOrCreate**: 根据条件在事务中更新或创建对象，支持乐观锁
- **DeleteInBatch**: 按 ID 批量删除对象

### IStore 接口

IStore 是应用级的统一数据访问接口，负责返回各个 Store 实现。接口采用模块化设计，通过方法返回相应的 Store：

```go
type IStore interface {
    // 事务和数据库
    DB(ctx context.Context, wheres ...where.Where) *gorm.DB
    TX(ctx context.Context, fn func(ctx context.Context) error) error

    // 业务 Store 方法（按模块组织）
    // 例如: Users() UserStore, Admin() AdminStore 等
}
```

## 使用示例

本章通过一个简单示例展示 Store 的基本用法。

### 基本 CRUD 操作

假设有一个 `User` 模型，通过 Store 实现基本操作：

```go
// 创建
user := &User{Name: "John", Email: "john@example.com"}
err := store.Users().Create(ctx, user)

// 读取
user, err := store.Users().Get(ctx, where.F("id", 1))

// 更新（仅更新指定字段）
user.Email = "newemail@example.com"
err := store.Users().Update(ctx, user, "email")

// 删除
err := store.Users().Delete(ctx, where.F("id", 1))
```

### 查询和分页

```go
// 构建查询条件
opts := where.F("status", "active").
    P(1, 10)  // 第1页，每页10条

// 执行查询
count, users, err := store.Users().List(ctx, opts)
```

### 事务处理

多个操作需要原子性保证时，使用 `TX()` 方法：

```go
err := store.TX(ctx, func(ctx context.Context) error {
    // Store 会自动使用事务
    if err := store.Users().Create(ctx, user1); err != nil {
        return err  // 自动回滚
    }
    if err := store.Users().Create(ctx, user2); err != nil {
        return err  // 自动回滚
    }
    return nil  // 自动提交
})
```

### 扩展操作

特定业务 Store 可以通过扩展接口添加自定义操作：

```go
// internal/pkg/store/user.go
type UserExpansion interface {
    FindByEmail(ctx context.Context, email string) (*User, error)
}

func (s *userStore) FindByEmail(ctx context.Context, email string) (*User, error) {
    return s.Get(ctx, where.F("email", email))
}
```

## 添加新的业务 Store

添加新的 Store 需要遵循以下步骤和命名规范（参考"命名规范"一节）：

### 1. 创建 Store 接口和实现

```go
// internal/pkg/store/user.go
package store

// Store 接口定义 CRUD 操作
type UserStore interface {
    Create(ctx context.Context, obj *User) error
    Update(ctx context.Context, obj *User, fields ...string) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*User, error)
    List(ctx context.Context, opts *where.Options) (int64, []*User, error)

    UserExpansion  // 扩展接口
}

// 扩展接口定义业务特定操作
type UserExpansion interface {
    FindByEmail(ctx context.Context, email string) (*User, error)
}

// 实现类
type userStore struct {
    *genericstore.Store[User]
}

// 工厂函数
func NewUserStore(store *datastore) *userStore {
    return &userStore{
        Store: genericstore.NewStore[User](store, NewLogger()),
    }
}

// 实现扩展方法
func (s *userStore) FindByEmail(ctx context.Context, email string) (*User, error) {
    return s.Get(ctx, where.F("email", email))
}
```

### 2. 注册到 IStore

在 `internal/pkg/store/store.go` 中添加方法：

```go
type IStore interface {
    Users() UserStore  // 新增
    // ...
}

func (ds *datastore) Users() UserStore {
    return NewUserStore(ds)
}
```

## 相关内容

- [分层架构详解](./layered-design.md) - 了解 Store 层在三层架构中的角色
- [整体架构](./architecture.md) - 微服务架构中的数据访问设计
- [Store 命名规范和最佳实践](../development/standards.md#store-命名规范) - 开发规范
- [开发第一个功能](../guide/first-feature.md) - 实战应用示例
- [数据库层](../components/database.md) - GORM 使用指南（待实现）
