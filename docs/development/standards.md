# 开发规范

本文定义 Bingo 项目的代码规范和开发约定。

## 命名规范

### 包名
- **小写**,简短,有意义
- **单数**形式(除非特殊情况)
- 不使用下划线或驼峰

```go
// ✅ 正确
package user
package article
package auth

// ❌ 错误
package users          // 应该用单数
package articleMgr     // 不使用驼峰
package user_service   // 不使用下划线
```

### 文件名
- **蛇形命名**(snake_case)
- 与主要类型相关

```
user_controller.go
article_store.go
auth_middleware.go
```

### 接口名
- **`I` 前缀**(Interface)
- 大驼峰命名

```go
type IStore interface {}
type IBiz interface {}
type ICache interface {}
```

### 结构体和方法
- **大驼峰**(导出) 或 **小驼峰**(私有)

```go
// 导出
type UserController struct {}
func (c *UserController) Create() {}

// 私有
type userCache struct {}
func (c *userCache) get() {}
```

### 常量和变量
```go
// 常量:大驼峰
const MaxRetryCount = 3
const DefaultPageSize = 20

// 变量:小驼峰
var userCache *cache.Cache
var defaultTimeout = 30 * time.Second
```

## 错误处理

### 统一错误码

定义在 `internal/pkg/errno/code.go`:

```go
var (
    // 用户相关错误 (100xx)
    ErrUserNotFound      = errno.New(10001, "用户不存在")
    ErrUserAlreadyExists = errno.New(10002, "用户已存在")
    ErrInvalidPassword   = errno.New(10003, "密码错误")

    // 文章相关错误 (200xx)
    ErrArticleNotFound = errno.New(20001, "文章不存在")
)
```

### 错误返回

```go
// ✅ 正确:返回定义的错误码
if user == nil {
    return nil, errno.ErrUserNotFound
}

// ✅ 正确:包装错误
if err := db.Create(user).Error; err != nil {
    return nil, fmt.Errorf("failed to create user: %w", err)
}

// ❌ 错误:直接返回字符串错误
if user == nil {
    return nil, errors.New("用户不存在")
}
```

### Controller 层错误处理

```go
func (ctrl *UserController) Get(c *gin.Context) {
    user, err := ctrl.biz.Users().Get(c.Context(), id)
    if err != nil {
        // 统一错误响应
        core.WriteResponse(c, err, nil)
        return
    }

    core.WriteResponse(c, nil, user)
}
```

## 日志规范

### 日志级别

- **Debug**: 调试信息
- **Info**: 重要业务流程
- **Warn**: 警告信息,不影响主流程
- **Error**: 错误信息,需要关注

### 日志记录

```go
import "github.com/bingo-project/bingo/internal/pkg/logger"

// ✅ 结构化日志
logger.Info("user created",
    zap.String("username", username),
    zap.Uint64("user_id", userID),
)

// ✅ 错误日志(带上下文)
logger.Error("failed to create user",
    zap.Error(err),
    zap.String("username", username),
)

// ❌ 不推荐:非结构化日志
logger.Info("user created: " + username)
```

### 日志最佳实践

1. **记录关键业务操作**
```go
logger.Info("user login",
    zap.Uint64("user_id", userID),
    zap.String("ip", clientIP),
)
```

2. **记录错误时带上下文**
```go
logger.Error("database query failed",
    zap.Error(err),
    zap.String("sql", sql),
    zap.Any("params", params),
)
```

3. **不记录敏感信息**
```go
// ❌ 错误:记录密码
logger.Info("user login", zap.String("password", password))

// ✅ 正确:不记录敏感信息
logger.Info("user login", zap.String("username", username))
```

## 注释规范

### 文件注释

每个文件开头必须有 `ABOUTME` 注释:

```go
// ABOUTME: User business logic implementation
// ABOUTME: Handles user registration, login, and profile management
package user
```

### 函数注释

```go
// CreateUser 创建新用户
// 参数:
//   - ctx: 上下文
//   - req: 创建用户请求
// 返回:
//   - *model.User: 创建的用户
//   - error: 错误信息
func (b *UserBiz) CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // ...
}
```

### Swagger 注释

```go
// @Summary      获取用户信息
// @Description  根据用户 ID 获取用户详细信息
// @Tags         用户管理
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "用户ID"
// @Success      200  {object}  UserResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /v1/users/{id} [get]
func (ctrl *UserController) Get(c *gin.Context) {
    // ...
}
```

## 代码组织

### Controller 层

```go
package user

import (
    "github.com/gin-gonic/gin"
    "github.com/bingo-project/bingo/internal/apiserver/biz"
    "github.com/bingo-project/bingo/pkg/core"
)

// 1. 类型定义
type UserController struct {
    biz biz.IBiz
}

// 2. 构造函数
func New(biz biz.IBiz) *UserController {
    return &UserController{biz: biz}
}

// 3. HTTP 处理方法(按 CRUD 顺序)
func (ctrl *UserController) Create(c *gin.Context) {}
func (ctrl *UserController) Get(c *gin.Context) {}
func (ctrl *UserController) List(c *gin.Context) {}
func (ctrl *UserController) Update(c *gin.Context) {}
func (ctrl *UserController) Delete(c *gin.Context) {}

// 4. 私有辅助方法
func (ctrl *UserController) validateRequest() {}
```

### Biz 层

```go
package user

// 1. 接口定义
type UserBiz interface {
    Create(ctx context.Context, req *CreateUserRequest) (*model.User, error)
    Get(ctx context.Context, id uint64) (*model.User, error)
}

// 2. 实现结构体
type userBiz struct {
    ds store.IStore
}

// 3. 构造函数
func New(ds store.IStore) UserBiz {
    return &userBiz{ds: ds}
}

// 4. 接口实现
func (b *userBiz) Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {}

// 5. 私有方法
func (b *userBiz) validateUser(req *CreateUserRequest) error {}
```

### Store 层

```go
package store

// 1. 接口定义
type UserStore interface {
    Create(ctx context.Context, user *model.User) error
    Get(ctx context.Context, id uint64) (*model.User, error)
}

// 2. 实现结构体
type userStore struct {
    db *gorm.DB
}

// 3. 构造函数
func newUserStore(db *gorm.DB) UserStore {
    return &userStore{db: db}
}

// 4. 接口实现
func (s *userStore) Create(ctx context.Context, user *model.User) error {}
```

## 数据库规范

### Model 定义

```go
type User struct {
    ID        uint64    `gorm:"primarykey" json:"id"`
    Username  string    `gorm:"size:50;not null;uniqueIndex" json:"username"`
    Email     string    `gorm:"size:100;not null;index" json:"email"`
    Password  string    `gorm:"size:255;not null" json:"-"` // json:"-" 不序列化
    Status    int       `gorm:"default:1" json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string {
    return "users"
}
```

### 查询优化

```go
// ✅ 使用索引字段查询
db.Where("username = ?", username).First(&user)

// ✅ 预加载关联数据(避免 N+1)
db.Preload("Articles").Find(&users)

// ✅ 只查询需要的字段
db.Select("id", "username", "email").Find(&users)

// ❌ 避免全表扫描
db.Where("username LIKE ?", "%"+keyword+"%").Find(&users)
```

## Git 规范

### Commit Message

```bash
# 格式
<type>: <subject>

# 类型
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式调整
refactor: 重构
test: 测试相关
chore: 构建/工具相关

# 示例
feat: add user login API
fix: resolve password encryption issue
docs: update API documentation
refactor: simplify user validation logic
```

### 分支管理

```
main/master     生产分支
develop         开发分支
feature/*       功能分支
bugfix/*        bug修复分支
hotfix/*        紧急修复分支
```

## 测试规范

### 测试文件命名

```
user.go       -> user_test.go
article.go    -> article_test.go
```

### 测试用例

```go
func TestUserBiz_Create(t *testing.T) {
    // 1. 准备测试数据
    req := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
    }

    // 2. 执行测试
    user, err := userBiz.Create(context.Background(), req)

    // 3. 断言
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "testuser", user.Username)
}
```

## Makefile 使用

```bash
# 编译
make build
make build BINS="bingo-apiserver"

# 测试
make test
make cover

# 代码检查
make lint
make format

# 生成文档
make swagger
```

## 下一步

- [业务开发指南](./business-guide.md) - 实践这些规范
- [测试指南](./testing.md) - 编写高质量测试
- [最佳实践](./best-practices.md) - 进阶技巧
