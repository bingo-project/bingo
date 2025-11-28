# 核心组件概览

Bingo 内置了一系列开箱即用的核心组件,本文介绍各个组件的功能和使用场景。

## 组件列表

### 基础组件

| 组件 | 说明 | 文档 |
|------|------|------|
| **Bootstrap** | 应用启动引导,统一初始化流程 | [配置管理](#bootstrap) |
| **Logger** | 基于 Zap 的结构化日志 | 🚧 |
| **Database** | GORM 数据库封装 | 🚧 |
| **Redis** | Redis 缓存封装 | 本文 |

### 认证授权

| 组件 | 说明 | 文档 |
|------|------|------|
| **Authn** | JWT 认证中间件 | 🚧 |
| **Authz** | Casbin 权限控制 | 🚧 |

### 异步任务

| 组件 | 说明 | 文档 |
|------|------|------|
| **Task Queue** | Asynq 任务队列 | 🚧 |
| **Scheduler** | 定时任务调度 | 🚧 |

### 其他组件

| 组件 | 说明 | 文档 |
|------|------|------|
| **Swagger** | API 文档生成 | 本文 |
| **Validator** | 参数验证 | 本文 |
| **Error** | 统一错误处理 | 本文 |

## Bootstrap

应用启动引导组件,统一管理各个组件的初始化。

### 使用方式

```go
// internal/apiserver/app.go
bootstrap := bootstrap.NewBootstrap()
bootstrap.InitConfig("bingo-apiserver.yaml")
bootstrap.Boot()  // 初始化所有组件

// 获取组件实例
db := bootstrap.GetDB()
redis := bootstrap.GetRedis()
logger := bootstrap.GetLogger()
```

### 初始化顺序

```
1. 加载配置(Viper)
    ↓
2. 初始化日志(Zap)
    ↓
3. 连接数据库(GORM)
    ↓
4. 连接 Redis
    ↓
5. 初始化其他组件
```

## Redis

基于 go-redis 的 Redis 封装。

### 使用方式

```go
import "github.com/bingo-project/bingo/internal/pkg/db"

// 获取 Redis 客户端
rdb := bootstrap.GetRedis()

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

## Swagger

基于 swaggo/swag 的 API 文档自动生成。

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
func (ctrl *UserController) Create(c *gin.Context) {
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

## Validator

基于 go-playground/validator 的参数验证。

### 使用方式

```go
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"gte=18,lte=100"`
    Password string `json:"password" binding:"required,min=6"`
}

func (ctrl *UserController) Create(c *gin.Context) {
    var req CreateUserRequest

    // Gin 自动验证
    if err := c.ShouldBindJSON(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // 验证通过,继续处理
}
```

### 常用验证标签

```go
required        // 必填
min=3          // 最小长度/值
max=20         // 最大长度/值
email          // 邮箱格式
url            // URL 格式
oneof=red blue // 枚举值
gte=18         // 大于等于
lte=100        // 小于等于
```

## Error

统一错误处理组件。

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
// Biz 层返回
if user == nil {
    return nil, errno.ErrUserNotFound
}

// Controller 层统一处理
func (ctrl *UserController) Get(c *gin.Context) {
    user, err := ctrl.biz.Users().Get(c.Context(), id)
    // 统一错误响应
    core.WriteResponse(c, err, user)
}
```

### 错误响应格式

```json
{
  "code": 10001,
  "message": "用户不存在"
}
```

## 组件扩展

### 添加新组件

1. 在 `internal/pkg/` 创建组件目录
2. 实现组件初始化逻辑
3. 在 Bootstrap 中注册

```go
// internal/pkg/bootstrap/bootstrap.go
func (b *Bootstrap) Boot() error {
    // ... 其他组件

    // 初始化新组件
    if err := b.initMyComponent(); err != nil {
        return err
    }

    return nil
}
```

## 下一步

> 部分文档正在筹备中，敬请期待！有关认证系统、权限系统、任务队列、数据库层和日志系统的详细文档将陆续发布。
