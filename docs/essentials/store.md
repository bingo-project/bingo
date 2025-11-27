# Store 包设计

## 概述

Store 包是 Bingo 数据访问层的核心，通过**泛型**和**组合模式**实现了一个灵活、可扩展的数据访问框架，减少重复代码并提高代码复用率。

## 包结构

```
pkg/store/
├── store.go          # 通用 Store[T] 实现
├── logger.go         # Logger 接口定义
└── where/
    └── where.go      # 查询条件构建器

internal/pkg/store/
├── store.go          # IStore 接口和应用级实现
├── schedule.go       # Schedule 业务 Store
└── logger.go         # 业务日志实现
```

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
type scheduleStore struct {
    *genericstore.Store[syscfg.Schedule]
}

type ScheduleStore interface {
    Create(ctx context.Context, obj *syscfg.Schedule) error
    Update(ctx context.Context, obj *syscfg.Schedule, fields ...string) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*syscfg.Schedule, error)
    List(ctx context.Context, opts *where.Options) (int64, []*syscfg.Schedule, error)

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
// CRUD
Create(ctx context.Context, obj *T) error
Update(ctx context.Context, obj *T, fields ...string) error
Delete(ctx context.Context, opts *where.Options) error
Get(ctx context.Context, opts *where.Options) (*T, error)

// 查询
List(ctx context.Context, opts *where.Options) (count int64, ret []*T, err error)
Find(ctx context.Context, opts *where.Options) (ret []*T, err error)
Last(ctx context.Context, opts *where.Options) (*T, error)

// 批量操作
CreateInBatch(ctx context.Context, objs []*T, batchSize int) error
Upsert(ctx context.Context, obj *T, fields ...string) error

// 原始数据库访问
DB(ctx context.Context, wheres ...where.Where) *gorm.DB
```

### IStore 接口

```go
type IStore interface {
    DB(ctx context.Context, wheres ...where.Where) *gorm.DB
    TX(ctx context.Context, fn func(ctx context.Context) error) error

    Schedules() ScheduleStore
    Users() UserStore
    // ... 其他业务 Store
}
```

## 使用示例

### 简单 CRUD

```go
// 创建
schedule := &Schedule{Name: "daily-task", Status: "active"}
err := store.Schedules().Create(ctx, schedule)

// 读取
schedule, err := store.Schedules().Get(ctx, where.F("id", 1))

// 更新
schedule.Status = "inactive"
err := store.Schedules().Update(ctx, schedule, "status")

// 删除
err := store.Schedules().Delete(ctx, where.F("id", 1))
```

### 分页查询

```go
opts := where.F("status", "active").
    Load("User").
    P(1, 10)  // 第1页，每页10条

count, schedules, err := store.Schedules().List(ctx, opts)
```

### 事务处理

```go
err := store.TX(ctx, func(ctx context.Context) error {
    if err := store.Schedules().Create(ctx, schedule); err != nil {
        return err  // 自动回滚
    }
    if err := store.Users().Update(ctx, user, "updated_at"); err != nil {
        return err  // 自动回滚
    }
    return nil  // 自动提交
})
```

### 业务特定操作

```go
// internal/pkg/store/schedule.go
type UserExpansion interface {
    GetActiveSchedules(ctx context.Context) ([]*syscfg.Schedule, error)
}

func (s *scheduleStore) GetActiveSchedules(ctx context.Context) ([]*syscfg.Schedule, error) {
    opts := where.F("status", syscfg.ScheduleStatusEnabled)
    _, schedules, err := s.List(ctx, opts)
    return schedules, err
}
```

## Biz 层集成

```go
// internal/scheduler/biz/syscfg/schedule.go
type ScheduleBiz struct {
    store store.IStore
}

func NewSchedule(store store.IStore) *ScheduleBiz {
    return &ScheduleBiz{store: store}
}

func (b *ScheduleBiz) GetConfigs(ctx context.Context) ([]*asynq.PeriodicTaskConfig, error) {
    whr := where.F("status", syscfg.ScheduleStatusEnabled)
    _, configs, err := b.store.Schedules().List(ctx, whr)
    // ... 处理结果
    return ret, err
}
```

关键点：
- 通过**构造函数依赖注入** IStore 接口
- 使用 **where 包**构建灵活的查询条件
- 需要事务时使用 **store.TX()**

## 添加新的业务 Store

1. **创建模型文件** `internal/pkg/store/user.go`

```go
package store

type UserStore interface {
    Create(ctx context.Context, obj *model.User) error
    Update(ctx context.Context, obj *model.User, fields ...string) error
    Delete(ctx context.Context, opts *where.Options) error
    Get(ctx context.Context, opts *where.Options) (*model.User, error)
    List(ctx context.Context, opts *where.Options) (int64, []*model.User, error)

    UserExpansion
}

type UserExpansion interface {
    GetByUsername(ctx context.Context, username string) (*model.User, error)
}

type userStore struct {
    *genericstore.Store[model.User]
}

func NewUserStore(store *datastore) *userStore {
    return &userStore{
        Store: genericstore.NewStore[model.User](store, NewLogger()),
    }
}

func (s *userStore) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    return s.Get(ctx, where.F("username", username))
}
```

2. **更新 IStore** `internal/pkg/store/store.go`

```go
type IStore interface {
    Schedules() ScheduleStore
    Users() UserStore  // 新增
}

func (ds *datastore) Users() UserStore {
    return NewUserStore(ds)
}
```

3. **在 Biz 层使用**

```go
type UserBiz struct {
    store store.IStore
}

func (b *UserBiz) GetByUsername(ctx context.Context, username string) (*model.User, error) {
    return b.store.Users().GetByUsername(ctx, username)
}
```

## 最佳实践

| 原则 | 说明 |
|------|------|
| **依赖接口** | Biz 层依赖 IStore 接口，不依赖具体实现 |
| **预加载关联** | 使用 `Load()` 避免 N+1 查询 |
| **业务逻辑分离** | 业务规则在 Biz 层，Store 层仅做数据操作 |
| **事务隔离** | 多个操作需要原子性时使用 `TX()` |
| **错误处理** | 数据库错误会自动记录日志 |

## 下一步

- [分层架构详解](./layered-design.md) - 三层架构设计原则
- [开发第一个功能](../guide/first-feature.md) - 实践分层架构
