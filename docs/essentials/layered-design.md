# 分层架构详解

Bingo 采用经典的三层架构设计,本文详细介绍每一层的职责和设计原则。

## 三层架构

```
┌─────────────────────────────────────────┐
│         Controller Layer                │  HTTP/gRPC 处理层
│  - 参数验证                              │
│  - 请求响应转换                          │
│  - 错误处理                              │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Business Layer (Biz)           │  业务逻辑层
│  - 业务规则                              │
│  - 业务流程编排                          │
│  - 事务控制                              │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Store Layer                    │  数据访问层
│  - 数据库操作                            │
│  - 缓存操作                              │
│  - 第三方服务调用                        │
└─────────────────────────────────────────┘
```

## Controller 层(控制器层)

### 职责

1. **接收请求**: 处理 HTTP/gRPC 请求
2. **参数验证**: 绑定和验证请求参数
3. **调用业务**: 调用 Biz 层处理业务
4. **返回响应**: 构造并返回响应

### 代码示例

```go
// internal/apiserver/controller/v1/user/user.go
type UserController struct {
    biz biz.IBiz
}

func (ctrl *UserController) Get(c *gin.Context) {
    // 1. 参数验证
    var req GetUserRequest
    if err := c.ShouldBindUri(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // 2. 调用业务层
    user, err := ctrl.biz.Users().Get(c.Context(), req.UserID)
    if err != nil {
        core.WriteResponse(c, err, nil)
        return
    }

    // 3. 返回响应
    core.WriteResponse(c, nil, user)
}
```

### 设计原则

- **薄控制器**: 只做参数处理和响应,不包含业务逻辑
- **统一响应**: 使用统一的响应格式
- **错误处理**: 统一的错误处理机制
- **版本隔离**: 不同 API 版本独立目录(`v1/`, `v2/`)

### 不应该做的事

❌ **在 Controller 中写业务逻辑**
```go
// 错误示例
func (ctrl *UserController) Create(c *gin.Context) {
    // ❌ 业务规则不应该在这里
    if user.Age < 18 {
        return errors.New("年龄不足")
    }

    // ❌ 密码加密不应该在这里
    hashedPassword := encrypt(user.Password)
}
```

✅ **应该调用 Biz 层**
```go
// 正确示例
func (ctrl *UserController) Create(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // ✅ 业务逻辑交给 Biz 层
    user, err := ctrl.biz.Users().Create(c.Context(), &req)
    core.WriteResponse(c, err, user)
}
```

## Biz 层(业务逻辑层)

### 职责

1. **业务规则**: 实现核心业务逻辑和规则
2. **流程编排**: 编排多个 Store 操作
3. **事务控制**: 处理数据库事务
4. **业务验证**: 业务级别的验证

### 代码示例

```go
// internal/apiserver/biz/user/user.go
type UserBiz struct {
    ds store.IStore
}

func (b *UserBiz) Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // 1. 业务规则验证
    if err := b.validateUser(req); err != nil {
        return nil, err
    }

    // 2. 业务逻辑处理
    req.Password = encryptPassword(req.Password)

    // 3. 构建模型
    user := &model.User{
        Username: req.Username,
        Password: req.Password,
        Email:    req.Email,
    }

    // 4. 数据持久化
    if err := b.ds.Users().Create(ctx, user); err != nil {
        return nil, err
    }

    // 5. 业务流程编排(如发送欢迎邮件)
    go b.sendWelcomeEmail(user.Email)

    return user, nil
}

func (b *UserBiz) validateUser(req *CreateUserRequest) error {
    // 业务规则验证
    if req.Age < 18 {
        return errno.ErrUserAgeTooYoung
    }

    // 检查用户名是否已存在
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

### 设计原则

- **核心业务**: 所有业务逻辑都在这一层
- **接口编程**: 依赖 Store 接口,不依赖具体实现
- **可测试性**: 通过 Mock Store 进行单元测试
- **事务控制**: 需要事务时使用 Store 的事务方法

### 典型场景

#### 场景1:单表操作
```go
func (b *UserBiz) Get(ctx context.Context, id uint64) (*model.User, error) {
    return b.ds.Users().Get(ctx, id)
}
```

#### 场景2:多表操作编排
```go
func (b *OrderBiz) Create(ctx context.Context, req *CreateOrderRequest) error {
    // 1. 检查库存
    stock, err := b.ds.Products().GetStock(ctx, req.ProductID)
    if err != nil {
        return err
    }
    if stock < req.Quantity {
        return errno.ErrInsufficientStock
    }

    // 2. 创建订单
    order := &model.Order{...}
    if err := b.ds.Orders().Create(ctx, order); err != nil {
        return err
    }

    // 3. 减库存
    if err := b.ds.Products().DecreaseStock(ctx, req.ProductID, req.Quantity); err != nil {
        return err
    }

    return nil
}
```

#### 场景3:事务控制
```go
func (b *OrderBiz) Create(ctx context.Context, req *CreateOrderRequest) error {
    // 使用事务
    return b.ds.TX(ctx, func(ctx context.Context) error {
        // 在事务中执行多个操作
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

## Store 层(数据访问层)

> 💡 详细设计说明请参考 [Store 包设计](./store.md)

### 职责

1. **数据库操作**: 封装 GORM 操作
2. **缓存操作**: Redis 缓存读写
3. **数据转换**: 数据格式转换
4. **查询优化**: SQL 优化和索引使用

### 代码示例

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

    // 条件查询
    if opts.Username != "" {
        db = db.Where("username LIKE ?", "%"+opts.Username+"%")
    }

    // 计数
    if err := db.Count(&count).Error; err != nil {
        return nil, 0, err
    }

    // 分页
    if err := db.Offset(opts.Offset).Limit(opts.Limit).Find(&users).Error; err != nil {
        return nil, 0, err
    }

    return users, count, nil
}
```

### 设计原则

- **纯数据操作**: 只做数据库/缓存操作,不包含业务逻辑
- **接口定义**: 每个 Store 都定义接口
- **错误转换**: 将数据库错误转换为业务错误
- **查询优化**: 注意 N+1 问题,合理使用 Preload

### 缓存使用示例

```go
func (s *userStore) Get(ctx context.Context, id uint64) (*model.User, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("user:%d", id)
    var user model.User

    if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil
    }

    // 2. 缓存未命中,从数据库查询
    if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存
    _ = s.cache.Set(ctx, cacheKey, &user, time.Hour)

    return &user, nil
}
```

## 为什么要分层?

### 1. 关注点分离
每层只关注自己的职责:
- Controller 关注 HTTP 协议
- Biz 关注业务规则
- Store 关注数据访问

### 2. 易于测试
```go
// 测试 Biz 层时,Mock Store 层
func TestUserBiz_Create(t *testing.T) {
    mockStore := &MockStore{}
    biz := user.New(mockStore)

    // 测试业务逻辑
    err := biz.Create(ctx, req)
    assert.NoError(t, err)
}
```

### 3. 代码复用
Biz 层可以被多个 Controller 复用:
```
HTTP Controller  ──┐
                   ├──→  User Biz  ──→  User Store
gRPC Service    ──┘
```

### 4. 易于维护
- 修改数据库操作:只改 Store 层
- 修改业务规则:只改 Biz 层
- 修改 API 格式:只改 Controller 层

### 5. 团队协作
不同层可以并行开发:
- 前端开发者:先 Mock Controller,并行开发
- 后端开发者:先定义接口,分层开发

## 常见错误

### 错误1:跨层调用

❌ **Controller 直接调用 Store**
```go
// 错误
func (ctrl *UserController) Get(c *gin.Context) {
    // ❌ Controller 不应该直接调用 Store
    user, err := ctrl.store.Users().Get(ctx, id)
}
```

✅ **应该通过 Biz 层**
```go
// 正确
func (ctrl *UserController) Get(c *gin.Context) {
    user, err := ctrl.biz.Users().Get(ctx, id)
}
```

### 错误2:业务逻辑泄漏

❌ **Store 层包含业务逻辑**
```go
// 错误
func (s *userStore) Create(ctx context.Context, user *model.User) error {
    // ❌ 业务验证不应该在 Store 层
    if user.Age < 18 {
        return errors.New("年龄不足")
    }
    return s.db.Create(user).Error
}
```

✅ **业务逻辑在 Biz 层**
```go
// 正确:Biz 层验证
func (b *userBiz) Create(ctx context.Context, req *CreateUserRequest) error {
    if req.Age < 18 {
        return errno.ErrUserAgeTooYoung
    }
    return b.ds.Users().Create(ctx, user)
}

// Store 层只做数据操作
func (s *userStore) Create(ctx context.Context, user *model.User) error {
    return s.db.Create(user).Error
}
```

## 下一步

- [开发第一个功能](../guide/first-feature.md) - 实践分层架构
- [业务开发指南](../development/business-guide.md) - 复杂场景的分层实践（待实现）
- [测试指南](../development/testing.md) - 如何测试各层（待实现）
