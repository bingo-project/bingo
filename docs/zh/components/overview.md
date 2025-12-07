---
title: 核心组件概览 - Bingo Go 微服务框架组件
description: 了解 Bingo Go 微服务框架的核心组件，包括 Bootstrap 启动引导、Redis 缓存、Swagger 文档、参数验证和错误处理等组件的使用方法。
---

# 核心组件概览

Bingo 内置了一系列开箱即用的核心组件,本文介绍各个组件的功能和使用场景。

## 组件列表

### 基础组件

| 组件 | 说明 | 原始包 |
|------|------|------|
| **Bootstrap** | 应用启动引导,统一初始化流程 | `internal/pkg/bootstrap` |
| **Facade** | 全局组件访问,单例模式 | `internal/pkg/facade` |
| **Logger** | 基于 Zap 的结构化日志 | [component-base/log](https://github.com/bingo-project/component-base) |
| **Database** | GORM 数据库封装 | [gorm.io/gorm](https://gorm.io) |
| **Redis** | Redis 缓存封装 | [go-redis/redis](https://github.com/redis/go-redis) |

### 认证授权

| 组件 | 说明 | 原始包 |
|------|------|------|
| **JWT** | JWT 认证 | [golang-jwt/jwt](https://github.com/golang-jwt/jwt) |
| **Casbin** | RBAC 权限控制 | [casbin/casbin](https://github.com/casbin/casbin) |

### 异步任务

| 组件 | 说明 | 原始包 |
|------|------|------|
| **Asynq** | 任务队列和定时任务 | [hibiken/asynq](https://github.com/hibiken/asynq) |

### 其他组件

| 组件 | 说明 | 原始包 |
|------|------|------|
| **Swagger** | API 文档生成 | [swaggo/swag](https://github.com/swaggo/swag) |
| **Validator** | 参数验证 | [go-playground/validator](https://github.com/go-playground/validator) |
| **Snowflake** | 分布式 ID 生成 | [bwmarrin/snowflake](https://github.com/bwmarrin/snowflake) |

## Bootstrap 与 Facade

Bingo 使用 Bootstrap 进行应用初始化,通过 Facade 提供全局组件访问。

### Bootstrap 初始化

`internal/pkg/bootstrap/app.go` 定义了统一的启动流程:

```go
// Boot 初始化所有核心组件
func Boot() {
    InitLog()        // 日志系统
    InitTimezone()   // 时区设置
    InitSnowflake()  // 分布式 ID
    InitMail()       // 邮件服务
    InitCache()      // 缓存服务
    InitAES()        // 加密组件
    InitQueue()      // 任务队列
}
```

### Facade 全局访问

`internal/pkg/facade/facade.go` 提供全局组件实例:

```go
import "bingo/internal/pkg/facade"

// 访问配置
cfg := facade.Config

// 访问 Redis
facade.Redis.Set(ctx, "key", "value", time.Hour)

// 访问缓存服务
facade.Cache.Set(ctx, "key", value, time.Hour)

// 生成分布式 ID
id := facade.Snowflake.Generate()

// 发送邮件
facade.Mail.Send(to, subject, body)
```

### 初始化顺序

```
1. 加载配置 (InitConfig)
    ↓
2. 初始化日志 (InitLog)
    ↓
3. 初始化时区 (InitTimezone)
    ↓
4. 初始化分布式 ID (InitSnowflake)
    ↓
5. 初始化缓存 (InitCache)
    ↓
6. 初始化数据库 (InitDB)
    ↓
7. 初始化 Store (NewStore)
```

## Redis

基于 [go-redis](https://github.com/redis/go-redis) 封装。

### 使用方式

```go
import "bingo/internal/pkg/facade"

// 通过 Facade 访问 Redis 客户端
rdb := facade.Redis

// 基本操作
rdb.Set(ctx, "key", "value", time.Hour)
val, err := rdb.Get(ctx, "key").Result()

// 缓存对象
type User struct {
    ID   uint64
    Name string
}

// 写入缓存
user := &User{ID: 1, Name: "test"}
data, _ := json.Marshal(user)
rdb.Set(ctx, "user:1", data, time.Hour)

// 读取缓存
data, _ := rdb.Get(ctx, "user:1").Bytes()
var user User
json.Unmarshal(data, &user)
```

> 详细用法请参考 [go-redis 文档](https://redis.uptrace.dev/)

## Swagger

基于 [swaggo/swag](https://github.com/swaggo/swag) 的 API 文档自动生成。

### 注解示例

```go
// @Summary      创建用户
// @Description  创建新用户账号
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        body  body      CreateUserRequest  true  "用户信息"
// @Success      200   {object}  UserResponse
// @Failure      400   {object}  ErrorResponse
// @Router       /v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
    // ...
}
```

### 生成文档

```bash
# 生成 Swagger 文档
make swagger

# 访问文档
open http://localhost:8080/swagger/index.html
```

> 详细注解语法请参考 [swag 文档](https://github.com/swaggo/swag#declarative-comments-format)

## Validator

基于 [go-playground/validator](https://github.com/go-playground/validator) 的参数验证，与 Gin 框架集成。

### 使用方式

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"gte=18,lte=100"`
    Password string `json:"password" binding:"required,min=6"`
}

func (h *UserHandler) Create(c *gin.Context) {
    var req CreateUserRequest

    // Gin 自动验证
    if err := c.ShouldBindJSON(&req); err != nil {
        core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
        return
    }

    // 验证通过,继续处理
}
```

### 常用验证标签

| 标签 | 说明 |
|------|------|
| `required` | 必填字段 |
| `min=3` | 最小长度/值 |
| `max=20` | 最大长度/值 |
| `email` | 邮箱格式 |
| `url` | URL 格式 |
| `oneof=a b` | 枚举值 |
| `gte=18` | 大于等于 |
| `lte=100` | 小于等于 |

> 完整验证标签列表请参考 [validator 文档](https://pkg.go.dev/github.com/go-playground/validator/v10)

## Error

统一错误处理组件，位于 `internal/pkg/errno`。

### 定义错误码

```go
// internal/pkg/errno/code.go
var (
    ErrUserNotFound = errno.New(10001, "用户不存在")
    ErrInvalidToken = errno.New(10002, "无效的令牌")
)
```

### 使用错误码

```go
// Biz 层返回错误
if user == nil {
    return nil, errno.ErrUserNotFound
}

// Handler 层统一处理
func (h *UserHandler) Get(c *gin.Context) {
    user, err := h.biz.Users().Get(c.Request.Context(), id)
    // 统一错误响应
    core.Response(c, user, err)
}
```

### 错误响应格式

```json
{
  "code": 10001,
  "message": "用户不存在"
}
```

## 扩展组件

### 添加新的全局组件

1. 在 `internal/pkg/facade/facade.go` 中添加变量定义
2. 在 `internal/pkg/bootstrap/` 中添加初始化函数
3. 在 `Boot()` 中调用初始化函数

```go
// 1. facade.go 添加变量
var MyComponent *mypackage.Client

// 2. bootstrap/ 添加初始化
func InitMyComponent() {
    facade.MyComponent = mypackage.NewClient(facade.Config.MyComponent)
}

// 3. app.go 中调用
func Boot() {
    // ... 其他初始化
    InitMyComponent()
}
```

## 下一步

- [开发规范](../development/standards.md) - 代码风格和最佳实践
