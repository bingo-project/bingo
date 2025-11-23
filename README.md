# Bingo

一个开箱即用的 Go 语言中后台脚手架，基于微服务架构设计，让开发者只需关注业务开发。

## 项目定位

Bingo 是一个**生产级的 Go 中后台脚手架**，提供了完整的微服务架构、核心组件和最佳实践，帮助团队快速搭建可扩展的后端服务。

### 设计理念

- **开箱即用**：内置完整的技术栈和核心组件，快速启动项目
- **业务聚焦**：脚手架处理技术细节，开发者专注业务逻辑
- **灵活可扩展**：模块化设计，可根据需求自由组合或移除
- **生产就绪**：包含监控、日志、链路追踪等生产环境必备功能
- **最佳实践**：遵循 Go 社区最佳实践和设计模式

## 核心特性

### 架构层面
- **微服务架构**：多服务独立部署，支持水平扩展
- **分层设计**：Controller → Biz → Store 清晰的三层架构
- **依赖注入**：基于接口编程，易于测试和扩展
- **服务发现**：支持 gRPC 服务间通信

### 技术组件
- **Web 框架**：Gin，高性能 HTTP 框架
- **ORM**：GORM，支持多种数据库
- **缓存**：Redis 集成，支持分布式缓存
- **任务队列**：Asynq，可靠的异步任务处理
- **权限控制**：Casbin，灵活的 RBAC 权限引擎
- **配置管理**：Viper，支持多种配置格式
- **日志系统**：Zap，结构化高性能日志
- **API 文档**：Swagger，自动生成 API 文档

### 工程能力
- **热重启**：Air 支持开发时热重启
- **代码生成**：自动生成 CRUD 代码和 API 文档
- **Docker 支持**：一键容器化部署
- **监控指标**：Prometheus + pprof 性能监控
- **单元测试**：完整的测试框架和示例

## 内置示例功能

脚手架内置了一些基础功能作为开发参考，这些功能是**可选的**，可以根据实际需求保留或移除：

- **用户认证**：JWT、OAuth、Web3 等多种认证方式示例
- **权限管理**：基于 RBAC 的权限控制示例
- **应用管理**：多应用和 API Key 管理示例
- **机器人服务**：Discord/Telegram Bot 集成示例
- **定时任务**：基于 Asynq 的任务调度示例

这些内置功能主要用于：
1. 展示脚手架的使用方式和最佳实践
2. 提供可复用的代码模板
3. 作为业务开发的起点

> **提示**：你可以参考这些示例快速开发自己的业务功能，也可以直接删除不需要的模块。

## 技术栈

### 核心框架
- **Go**: 1.23.1+
- **Web 框架**: Gin v1.10.0
- **ORM**: GORM v1.25.10
- **数据库**: MySQL 5.7+ / PostgreSQL（可选）
- **缓存**: Redis 6.0+
- **gRPC**: google.golang.org/grpc v1.64.0
- **任务队列**: Asynq v0.24.1

### 工具库
- **日志**: Zap v1.27.0
- **权限**: Casbin v2.89.0
- **JWT**: golang-jwt/jwt v4.5.0
- **配置**: Viper v1.18.2
- **CLI**: Cobra v1.8.0
- **验证**: validator v10+
- **工具集**: Lancet v2.3.2

## 系统要求

- Go 1.23.1+
- MySQL 5.7+ 或 PostgreSQL
- Redis 6.0+
- Docker & Docker Compose（可选）

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd bingo
```

### 2. 配置环境

```bash
# 复制配置文件
cp configs/bingo-apiserver.example.yaml bingo-apiserver.yaml

# 根据实际环境修改配置
vim bingo-apiserver.yaml
```

### 3. 启动依赖服务

```bash
# 使用 Docker Compose 启动 MySQL 和 Redis
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis
```

### 4. 数据库迁移

```bash
# 编译项目（输出路径：./_output/platforms/<os>/<arch>/）
make build

# 复制配置文件，并修改数据库配置
cp configs/{app}-admserver.example.yaml {app}-admserver.yaml

# Build your app ctl
make build BINS="{app}ctl"

# 执行数据库迁移
./_output/platforms/{os}/{arch}/{app}ctl migrate up
```

> **说明**：`make build` 会将二进制文件输出到 `./_output/platforms/<os>/<arch>/` 目录（如 `./_output/platforms/darwin/arm64/`）

### 5. 启动服务

```bash
# 方式一：直接运行
make build
bingo-apiserver -c bingo-apiserver.yaml

# 方式二：开发模式（热重启）
cp .air.example.toml .air.toml
air
```

### 6. 验证服务

```bash
# 检查服务状态
curl http://localhost:8080/health

# 访问 Swagger 文档
open http://localhost:8080/swagger/index.html
```

## 项目架构

### 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端层                              │
│  Web Browser / Mobile App / Third-party Services            │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                      API 网关层（可选）                       │
│                   Nginx / Traefik / Kong                     │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┼───────────┬────────────┐
         ▼           ▼           ▼            ▼
    ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
    │   API   │ │  Admin  │ │Scheduler│ │   Bot   │
    │ Server  │ │ Server  │ │ Service │ │ Service │
    └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
         │           │           │            │
         └───────────┼───────────┴────────────┘
                     ▼
         ┌───────────────────────┐
         │   基础设施层            │
         ├───────────────────────┤
         │  MySQL / PostgreSQL   │
         │  Redis                │
         │  Message Queue        │
         └───────────────────────┘
```

### 服务说明

#### 1. API Server (bingo-apiserver)
- **端口**: 8080 (HTTP), 8081 (gRPC), 8082 (WebSocket)
- **职责**: 对外提供 RESTful API 和 WebSocket 服务
- **特点**: 面向 C 端用户，高并发，可水平扩展

#### 2. Admin Server (bingo-admserver)
- **端口**: 18080 (HTTP), 18081 (gRPC)
- **职责**: 管理后台 API 服务
- **特点**: 面向内部管理，权限控制严格

#### 3. Scheduler (bingo-scheduler)
- **端口**: 8080 (Web UI)
- **职责**: 定时任务调度和异步任务处理
- **特点**: 基于 Asynq，支持任务重试和监控

#### 4. Bot Service (bingo-bot)
- **职责**: 第三方平台集成（Discord、Telegram 等）
- **特点**: 事件驱动，异步处理

#### 5. CLI Tool (bingoctl)
- **职责**: 命令行工具，数据库迁移、代码生成等
- **特点**: 提升开发效率

## 项目结构

```
bingo/
├── cmd/                            # 可执行程序入口
│   ├── bingo-apiserver/            # API 服务入口
│   │   └── main.go                 # 主函数
│   ├── bingo-admserver/            # 管理服务入口
│   ├── bingo-scheduler/            # 调度服务入口
│   ├── bingo-bot/                  # 机器人服务入口
│   └── bingoctl/                   # CLI 工具入口
│
├── internal/                       # 内部应用代码（不可被外部导入）
│   ├── apiserver/                  # API 服务实现
│   │   ├── app.go                  # 应用初始化
│   │   ├── run.go                  # 服务启动逻辑
│   │   ├── biz/                    # 业务逻辑层
│   │   │   ├── auth/               # 认证业务
│   │   │   ├── user/               # 用户业务
│   │   │   └── ...                 # 其他业务模块
│   │   ├── controller/             # 控制器层（HTTP Handler）
│   │   │   └── v1/                 # API v1 版本
│   │   │       ├── auth/           # 认证相关接口
│   │   │       ├── user/           # 用户相关接口
│   │   │       └── ...
│   │   ├── store/                  # 数据访问层
│   │   │   ├── user.go             # 用户数据访问
│   │   │   └── ...
│   │   ├── router/                 # 路由定义
│   │   │   └── router.go
│   │   ├── middleware/             # 中间件
│   │   │   ├── authn.go            # 认证中间件
│   │   │   ├── authz.go            # 授权中间件
│   │   │   └── ...
│   │   └── grpc/                   # gRPC 服务实现
│   │
│   ├── admserver/                  # 管理服务（结构同 apiserver）
│   ├── scheduler/                  # 调度服务
│   │   ├── job/                    # 任务定义
│   │   └── scheduler/              # 调度器
│   ├── bot/                        # 机器人服务
│   └── pkg/                        # 内部共享包
│       ├── bootstrap/              # 应用启动引导
│       ├── config/                 # 配置定义
│       ├── model/                  # 数据模型
│       ├── logger/                 # 日志组件
│       ├── db/                     # 数据库组件
│       ├── auth/                   # 认证组件
│       ├── util/                   # 工具函数
│       └── ...
│
├── pkg/                            # 可被外部导入的公共包
│   ├── api/                        # API 定义
│   ├── proto/                      # Protocol Buffer 定义
│   └── ...
│
├── api/                            # API 文档
│   ├── swagger/                    # Swagger 文档
│   └── openapi/                    # OpenAPI 规范
│
├── configs/                        # 配置文件
│   ├── bingo-apiserver.example.yaml
│   └── ...
│
├── deployments/                    # 部署配置
│   └── docker/
│       └── docker-compose.yaml
│
├── build/                          # 构建脚本
│   ├── docker/                     # Dockerfile
│   └── scripts/                    # 构建脚本
│
├── scripts/                        # 开发脚本
│   └── make-rules/                 # Makefile 规则
│
├── storage/                        # 运行时数据
│   ├── log/                        # 日志文件
│   └── public/                     # 静态资源
│
├── Makefile                        # 构建配置
├── go.mod                          # Go 模块定义
└── README.md                       # 项目文档
```

### 目录说明

#### cmd/
存放所有可执行程序的入口文件，每个服务一个目录。遵循 Go 标准项目布局。

#### internal/
内部代码，不会被外部项目导入。这是 Go 的包可见性特性，确保内部实现不被外部依赖。

**每个服务的标准结构**：
- `app.go` / `run.go`: 应用初始化和启动逻辑
- `biz/`: 业务逻辑层，处理业务规则
- `controller/`: HTTP 处理器，负责请求响应
- `store/`: 数据访问层，封装数据库操作
- `router/`: 路由配置
- `middleware/`: 中间件
- `grpc/`: gRPC 服务实现

#### pkg/
可以被外部项目导入的公共包。如果你的项目需要提供 SDK，可以放在这里。

#### internal/pkg/
内部共享包，被多个内部服务使用，但不对外暴露。

## 分层架构详解

### 三层架构设计

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

### Controller 层（控制器层）

**职责**：
- 接收 HTTP/gRPC 请求
- 参数验证和绑定
- 调用 Biz 层处理业务
- 返回响应

**示例**：
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

### Biz 层（业务逻辑层）

**职责**：
- 实现核心业务逻辑
- 编排多个 Store 操作
- 处理事务
- 业务规则验证

**示例**：
```go
// internal/apiserver/biz/user/user.go
type UserBiz struct {
    ds store.IStore
}

func (b *UserBiz) Create(ctx context.Context, req *CreateUserRequest) error {
    // 1. 业务规则验证
    if err := b.validateUser(req); err != nil {
        return err
    }

    // 2. 密码加密（业务逻辑）
    req.Password = encryptPassword(req.Password)

    // 3. 数据持久化
    user := &model.User{
        Username: req.Username,
        Password: req.Password,
    }

    return b.ds.Users().Create(ctx, user)
}
```

### Store 层（数据访问层）

**职责**：
- 封装数据库操作
- 缓存操作
- 数据转换

**示例**：
```go
// internal/apiserver/store/user.go
type UserStore struct {
    db *gorm.DB
}

func (s *UserStore) Create(ctx context.Context, user *model.User) error {
    return s.db.WithContext(ctx).Create(user).Error
}

func (s *UserStore) Get(ctx context.Context, userID uint64) (*model.User, error) {
    var user model.User
    if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
        return nil, err
    }
    return &user, nil
}
```

### 为什么要分层？

1. **关注点分离**：每层只关注自己的职责
2. **易于测试**：可以针对每层编写单元测试
3. **代码复用**：Biz 层可以被多个 Controller 复用
4. **易于维护**：修改某一层不影响其他层
5. **团队协作**：不同层可以并行开发

## 核心组件详解

### 1. 配置管理（Bootstrap）

基于 Viper 实现的配置管理系统，支持多种配置源。

**使用方式**：
```go
// internal/pkg/bootstrap/bootstrap.go
bootstrap := NewBootstrap()
bootstrap.InitConfig("bingo-apiserver.yaml")
bootstrap.Boot()  // 初始化所有组件
```

**配置文件结构**：
```yaml
server:
  name: bingo
  mode: release
  addr: 0.0.0.0:8080

mysql:
  host: 127.0.0.1:3306
  database: bingo

redis:
  host: 127.0.0.1:6379
```

### 2. 数据库层（Store）

封装了 GORM 操作，提供统一的数据访问接口。

**接口设计**：
```go
// internal/apiserver/store/store.go
type IStore interface {
    Users() UserStore
    Apps() AppStore
    // ... 更多 Store
}
```

**使用示例**：
```go
// 在 Biz 层使用
user, err := biz.ds.Users().Get(ctx, userID)
```

### 3. 认证中间件（Authn）

基于 JWT 的认证中间件。

**使用方式**：
```go
// internal/apiserver/router/router.go
v1 := g.Group("/v1")
{
    // 公开接口
    v1.POST("/auth/login", authController.Login)

    // 需要认证的接口
    auth := v1.Group("")
    auth.Use(middleware.Authn())
    {
        auth.GET("/users/:id", userController.Get)
    }
}
```

### 4. 权限中间件（Authz）

基于 Casbin 的权限控制。

**权限模型**：
```ini
# RBAC 模型
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
```

### 5. 日志系统（Logger）

基于 Zap 的结构化日志。

**使用方式**：
```go
import "github.com/bingo-project/bingo/internal/pkg/logger"

// 结构化日志
logger.Info("user created",
    zap.String("username", username),
    zap.Uint64("user_id", userID),
)

// 错误日志
logger.Error("failed to create user", zap.Error(err))
```

### 6. 异步任务（Task Queue）

基于 Asynq 的任务队列。

**定义任务**：
```go
// internal/scheduler/job/email.go
type EmailTask struct {
    To      string
    Subject string
    Body    string
}

func (t *EmailTask) Handle(ctx context.Context) error {
    // 发送邮件逻辑
    return sendEmail(t.To, t.Subject, t.Body)
}
```

**提交任务**：
```go
task := &EmailTask{
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "Welcome to our platform!",
}
queue.Enqueue(task)
```

### 7. API 文档（Swagger）

自动生成 API 文档。

**注解示例**：
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

生成文档：
```bash
make swagger
```

## 业务开发指南

### 开发新功能的标准流程

假设我们要开发一个"文章管理"功能。

#### 1. 定义数据模型

```go
// internal/pkg/model/article.go
package model

type Article struct {
    ID        uint64    `gorm:"primarykey"`
    Title     string    `gorm:"size:200;not null"`
    Content   string    `gorm:"type:text"`
    AuthorID  uint64    `gorm:"not null"`
    Status    int       `gorm:"default:0"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (Article) TableName() string {
    return "articles"
}
```

#### 2. 创建数据库迁移

```bash
# 使用 bingoctl 生成迁移文件
bingoctl make migration create create_articles_table
```

编辑迁移文件：
```go
// internal/bingoctl/database/migration/xxxx_create_articles_table.go
func up(db *gorm.DB) error {
    return db.AutoMigrate(&model.Article{})
}
```

执行迁移：
```bash
{app}ctl migrate up
```

#### 3. 创建 Store 层

```go
// internal/apiserver/store/article.go
package store

type ArticleStore interface {
    Create(ctx context.Context, article *model.Article) error
    Get(ctx context.Context, id uint64) (*model.Article, error)
    List(ctx context.Context, opts ListOptions) ([]*model.Article, error)
    Update(ctx context.Context, article *model.Article) error
    Delete(ctx context.Context, id uint64) error
}

type articleStore struct {
    db *gorm.DB
}

func newArticleStore(db *gorm.DB) ArticleStore {
    return &articleStore{db: db}
}

func (s *articleStore) Create(ctx context.Context, article *model.Article) error {
    return s.db.WithContext(ctx).Create(article).Error
}

// ... 实现其他方法
```

#### 4. 创建 Biz 层

```go
// internal/apiserver/biz/article/article.go
package article

type ArticleBiz interface {
    Create(ctx context.Context, req *CreateArticleRequest) (*model.Article, error)
    Get(ctx context.Context, id uint64) (*model.Article, error)
    List(ctx context.Context, opts ListOptions) ([]*model.Article, error)
    Update(ctx context.Context, id uint64, req *UpdateArticleRequest) error
    Delete(ctx context.Context, id uint64) error
}

type articleBiz struct {
    ds store.IStore
}

func New(ds store.IStore) ArticleBiz {
    return &articleBiz{ds: ds}
}

func (b *articleBiz) Create(ctx context.Context, req *CreateArticleRequest) (*model.Article, error) {
    // 1. 业务验证
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // 2. 构建模型
    article := &model.Article{
        Title:    req.Title,
        Content:  req.Content,
        AuthorID: req.AuthorID,
        Status:   0,
    }

    // 3. 持久化
    if err := b.ds.Articles().Create(ctx, article); err != nil {
        return nil, err
    }

    return article, nil
}

// ... 实现其他方法
```

#### 5. 创建 Controller 层

```go
// internal/apiserver/controller/v1/article/article.go
package article

type ArticleController struct {
    biz biz.IBiz
}

func New(biz biz.IBiz) *ArticleController {
    return &ArticleController{biz: biz}
}

// @Summary      创建文章
// @Description  创建新文章
// @Tags         文章管理
// @Accept       json
// @Produce      json
// @Param        body  body      CreateArticleRequest  true  "文章信息"
// @Success      200   {object}  Article
// @Router       /v1/articles [post]
func (ctrl *ArticleController) Create(c *gin.Context) {
    var req CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // 从上下文获取当前用户 ID
    req.AuthorID = c.GetUint64("user_id")

    article, err := ctrl.biz.Articles().Create(c.Request.Context(), &req)
    if err != nil {
        core.WriteResponse(c, err, nil)
        return
    }

    core.WriteResponse(c, nil, article)
}

// ... 实现其他方法
```

#### 6. 注册路由

```go
// internal/apiserver/router/router.go
func InstallRouters(g *gin.Engine) {
    // ... 其他路由

    // 文章路由
    articleController := article.New(biz)
    v1auth := v1.Group("")
    v1auth.Use(middleware.Authn())
    {
        articles := v1auth.Group("/articles")
        {
            articles.POST("", articleController.Create)
            articles.GET("/:id", articleController.Get)
            articles.GET("", articleController.List)
            articles.PUT("/:id", articleController.Update)
            articles.DELETE("/:id", articleController.Delete)
        }
    }
}
```

#### 7. 生成 API 文档

```bash
make swagger
```

#### 8. 编写测试

```go
// internal/apiserver/biz/article/article_test.go
func TestArticleBiz_Create(t *testing.T) {
    // 准备测试数据
    req := &CreateArticleRequest{
        Title:    "Test Article",
        Content:  "Test Content",
        AuthorID: 1,
    }

    // 执行测试
    article, err := articleBiz.Create(context.Background(), req)

    // 断言
    assert.NoError(t, err)
    assert.NotNil(t, article)
    assert.Equal(t, "Test Article", article.Title)
}
```

运行测试：
```bash
make test
```

### 开发规范

#### 命名规范

- **包名**：小写，简短，有意义（如 `user`, `article`）
- **文件名**：蛇形命名（如 `user_controller.go`）
- **接口名**：`I` 前缀（如 `IStore`, `IBiz`）
- **结构体**：大驼峰（如 `UserController`）
- **函数/方法**：大驼峰（导出）或小驼峰（私有）

#### 错误处理

使用统一的错误码：
```go
// internal/pkg/errno/code.go
var (
    ErrUserNotFound = errno.New(10001, "用户不存在")
    ErrInvalidPassword = errno.New(10002, "密码错误")
)
```

返回错误：
```go
if user == nil {
    return nil, errno.ErrUserNotFound
}
```

#### 日志规范

```go
// 业务日志
logger.Info("user created",
    zap.String("username", username),
    zap.Uint64("user_id", userID),
)

// 错误日志（带上下文）
logger.Error("failed to create user",
    zap.Error(err),
    zap.String("username", username),
)
```

#### 注释规范

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

## Makefile 命令

```bash
# 开发相关
make build          # 编译所有服务
make run            # 运行服务（开发模式）
make test           # 运行单元测试
make cover          # 测试覆盖率报告

# 代码质量
make lint           # 代码检查
make format         # 代码格式化
make vet            # Go vet 检查

# 代码生成
make swagger        # 生成 Swagger 文档
make protoc         # 编译 Protocol Buffers
make wire           # 依赖注入代码生成

# 部署相关
make image          # 构建 Docker 镜像
make push           # 推送镜像到仓库

# 清理
make clean          # 清理构建产物
make tidy           # 整理依赖
```

## 配置文件详解

### 服务配置

```yaml
# 服务基本配置
server:
  name: bingo                    # 服务名称
  mode: release                  # 运行模式: release/debug/test
  addr: 0.0.0.0:8080            # 监听地址
  timezone: UTC                  # 时区

# gRPC 配置
grpc:
  addr: 0.0.0.0:8081
  network: tcp

# WebSocket 配置
websocket:
  addr: :8082

# 数据库配置
mysql:
  host: 127.0.0.1:3306
  database: bingo
  username: root
  password: your_password
  maxIdleConnections: 100
  maxOpenConnections: 100
  maxConnectionLifeTime: 10s
  logLevel: 4                    # 1:silent 2:error 3:warn 4:info

# Redis 配置
redis:
  host: 127.0.0.1:6379
  password: ""
  database: 1

# JWT 配置
jwt:
  secretKey: your-secret-key     # 建议使用环境变量
  ttl: 1440000                   # Token 有效期（分钟）

# 日志配置
log:
  level: info                    # debug/info/warn/error
  format: console                # console/json
  days: 7                        # 日志保留天数
  path: storage/log/api.log

# 功能开关
feature:
  metrics: true                  # Prometheus 指标
  profiling: true                # pprof 性能分析
  apiDoc: true                   # Swagger 文档
  queueDash: true                # 任务队列监控面板
```

### 环境变量

支持通过环境变量覆盖配置：

```bash
export BINGO_MYSQL_HOST="localhost:3306"
export BINGO_MYSQL_PASSWORD="secret"
export BINGO_REDIS_HOST="localhost:6379"
export BINGO_JWT_SECRET="your-secret-key"
```

## Docker 部署

### 本地开发环境

```bash
# 启动依赖服务（MySQL + Redis）
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis

# 启动所有服务
docker-compose -f deployments/docker/docker-compose.yaml up -d
```

### 生产环境

```bash
# 1. 构建镜像
make image

# 2. 推送到镜像仓库
docker tag bingo-apiserver:latest registry.example.com/bingo-apiserver:v1.0.0
docker push registry.example.com/bingo-apiserver:v1.0.0

# 3. 在生产服务器部署
docker pull registry.example.com/bingo-apiserver:v1.0.0
docker run -d \
  --name bingo-apiserver \
  -p 8080:8080 \
  -v /path/to/config.yaml:/etc/bingo/config.yaml \
  registry.example.com/bingo-apiserver:v1.0.0
```

## 监控与调试

### Prometheus 指标

访问 `http://localhost:8080/metrics` 查看指标。

**主要指标**：
- `http_requests_total`: HTTP 请求总数
- `http_request_duration_seconds`: 请求耗时
- `go_goroutines`: Goroutine 数量
- `go_memstats_alloc_bytes`: 内存使用

### pprof 性能分析

访问 `http://localhost:8080/debug/pprof/`

```bash
# CPU 分析
go tool pprof http://localhost:8080/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine 分析
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### 日志查看

```bash
# 实时查看日志
tail -f storage/log/api.log

# 使用 jq 格式化 JSON 日志
tail -f storage/log/api.log | jq .
```

## 常见问题

### 如何移除不需要的内置功能？

1. 删除对应的 Biz 层代码（`internal/*/biz/`）
2. 删除对应的 Controller 层代码
3. 删除对应的路由注册
4. 删除对应的数据表迁移文件
5. 运行 `make tidy` 清理未使用的依赖

### 如何添加新的数据库？

1. 在配置文件中添加数据库配置
2. 在 Bootstrap 中初始化数据库连接
3. 在 Store 中注入新的数据库连接

### 如何自定义中间件？

```go
// internal/apiserver/middleware/custom.go
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 前置处理

        c.Next()

        // 后置处理
    }
}
```

注册到路由：
```go
g.Use(middleware.CustomMiddleware())
```

### 如何实现服务间通信？

使用 gRPC：

1. 定义 Protocol Buffers
2. 生成代码：`make protoc`
3. 实现 gRPC 服务
4. 在客户端调用

### 如何进行数据库迁移？

```bash
# 创建迁移文件
bingoctl migrate create migration_name

# 执行迁移（应用所有未执行的迁移）
{app}ctl migrate up

# 回滚最后一次迁移
{app}ctl migrate rollback

# 回滚所有迁移
{app}ctl migrate reset
```

**⚠️ Important**: After modifying migration files, you must rebuild the application before running migrations:

```bash
# 1. Modify migration file
vim internal/{app}ctl/database/migration/xxx.go

# 2. Rebuild (DON'T FORGET THIS STEP!)
make build BINS="{app}ctl"

# 3. Run migration
./_output/platforms/{os}/{arch}/{app}ctl migrate up
```

## 最佳实践

### 1. 接口设计

- 使用 RESTful 风格
- 版本化 API（`/v1/`, `/v2/`）
- 统一的响应格式
- 合理的 HTTP 状态码

### 2. 错误处理

- 使用统一的错误码
- 记录详细的错误日志
- 不暴露敏感信息给客户端

### 3. 性能优化

- 合理使用缓存（Redis）
- 数据库查询优化（索引、分页）
- 避免 N+1 查询
- 使用连接池

### 4. 安全性

- 敏感信息不要硬编码，使用环境变量
- 输入验证
- SQL 注入防护（使用参数化查询）
- XSS 防护
- CSRF 防护

### 5. 可测试性

- 依赖注入
- 接口编程
- Mock 外部依赖
- 编写单元测试

## 进阶主题

### 微服务拆分

当业务复杂度增加时，可以按业务领域拆分服务：

```
bingo-user-service      # 用户服务
bingo-order-service     # 订单服务
bingo-payment-service   # 支付服务
bingo-gateway           # API 网关
```

### 服务发现

集成 Consul/Etcd 实现服务注册与发现。

### 链路追踪

集成 Jaeger/Zipkin 实现分布式链路追踪。

### 配置中心

使用 Consul/Nacos 作为配置中心。

## 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发流程

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交修改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

### 代码审查

PR 需要通过：
- 代码规范检查（golangci-lint）
- 单元测试
- 至少一位 Maintainer 的审查

## 许可证

本项目采用 [MIT License](LICENSE) 开源许可证。

## 联系方式

如有问题或建议，请：
- 提交 Issue
- 发送邮件到项目维护者

---

**开始使用 Bingo，专注于你的业务逻辑，让脚手架处理其他一切！**
