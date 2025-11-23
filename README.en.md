# Bingo

A production-ready Go backend scaffold based on microservice architecture, allowing developers to focus solely on business development.

## Project Positioning

Bingo is a **production-grade Go backend scaffold** that provides a complete microservice architecture, core components, and best practices to help teams quickly build scalable backend services.

### Design Philosophy

- **Ready to Use**: Built-in complete tech stack and core components for quick project startup
- **Business Focused**: Scaffold handles technical details, developers focus on business logic
- **Flexible & Extensible**: Modular design, freely combine or remove components as needed
- **Production Ready**: Includes monitoring, logging, distributed tracing, and other production essentials
- **Best Practices**: Follows Go community best practices and design patterns

## Core Features

### Architecture Level
- **Microservice Architecture**: Multiple services with independent deployment, supports horizontal scaling
- **Layered Design**: Clear three-layer architecture: Controller → Biz → Store
- **Dependency Injection**: Interface-based programming, easy to test and extend
- **Service Discovery**: Supports gRPC inter-service communication

### Technical Components
- **Web Framework**: Gin, high-performance HTTP framework
- **ORM**: GORM, supports multiple databases
- **Cache**: Redis integration, supports distributed caching
- **Task Queue**: Asynq, reliable async task processing
- **Access Control**: Casbin, flexible RBAC permission engine
- **Configuration**: Viper, supports multiple configuration formats
- **Logging**: Zap, structured high-performance logging
- **API Documentation**: Swagger, auto-generated API docs

### Engineering Capabilities
- **Hot Reload**: Air support for development hot reload
- **Code Generation**: Auto-generate CRUD code and API documentation
- **Docker Support**: One-command containerized deployment
- **Monitoring Metrics**: Prometheus + pprof performance monitoring
- **Unit Testing**: Complete testing framework and examples

## Built-in Example Features

The scaffold includes some basic features as development references. These features are **optional** and can be kept or removed based on actual needs:

- **User Authentication**: Examples of JWT, OAuth, Web3, and other authentication methods
- **Permission Management**: RBAC-based permission control examples
- **Application Management**: Multi-application and API Key management examples
- **Bot Service**: Discord/Telegram Bot integration examples
- **Scheduled Tasks**: Asynq-based task scheduling examples

These built-in features are primarily used for:
1. Demonstrating scaffold usage and best practices
2. Providing reusable code templates
3. Serving as a starting point for business development

> **Tip**: You can reference these examples to quickly develop your own business features, or directly delete modules you don't need.

## Tech Stack

### Core Framework
- **Go**: 1.23.1+
- **Web Framework**: Gin v1.10.0
- **ORM**: GORM v1.25.10
- **Database**: MySQL 5.7+ / PostgreSQL (optional)
- **Cache**: Redis 6.0+
- **gRPC**: google.golang.org/grpc v1.64.0
- **Task Queue**: Asynq v0.24.1

### Utility Libraries
- **Logging**: Zap v1.27.0
- **Authorization**: Casbin v2.89.0
- **JWT**: golang-jwt/jwt v4.5.0
- **Configuration**: Viper v1.18.2
- **CLI**: Cobra v1.8.0
- **Validation**: validator v10+
- **Utilities**: Lancet v2.3.2

## System Requirements

- Go 1.23.1+
- MySQL 5.7+ or PostgreSQL
- Redis 6.0+
- Docker & Docker Compose (optional)

## Quick Start

### 1. Clone Project

```bash
git clone <repository-url>
cd bingo
```

### 2. Configure Environment

```bash
# Copy configuration file
cp configs/bingo-apiserver.example.yaml bingo-apiserver.yaml

# Modify configuration according to your environment
vim bingo-apiserver.yaml
```

### 3. Start Dependency Services

```bash
# Start MySQL and Redis using Docker Compose
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis
```

### 4. Database Migration

```bash
# Build project (output path: ./_output/platforms/<os>/<arch>/)
make build

# Copy configuration file and modify database settings
cp configs/{app}-admserver.example.yaml {app}-admserver.yaml

# Build your app ctl
make build BINS="{app}ctl"

# Run database migration
./_output/platforms/{os}/{arch}/{app}ctl migrate up
```

> **Note**: `make build` outputs binaries to `./_output/platforms/<os>/<arch>/` directory (e.g., `./_output/platforms/darwin/arm64/`)

### 5. Start Services

```bash
# Method 1: Direct run
make build
bingo-apiserver -c bingo-apiserver.yaml

# Method 2: Development mode (hot reload)
cp .air.example.toml .air.toml
air
```

### 6. Verify Services

```bash
# Check service status
curl http://localhost:8080/health

# Access Swagger documentation
open http://localhost:8080/swagger/index.html
```

## Project Architecture

### Overall Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Client Layer                          │
│  Web Browser / Mobile App / Third-party Services            │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   API Gateway (Optional)                     │
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
         │   Infrastructure      │
         ├───────────────────────┤
         │  MySQL / PostgreSQL   │
         │  Redis                │
         │  Message Queue        │
         └───────────────────────┘
```

### Service Descriptions

#### 1. API Server (bingo-apiserver)
- **Ports**: 8080 (HTTP), 8081 (gRPC), 8082 (WebSocket)
- **Responsibilities**: Provides RESTful API and WebSocket services
- **Characteristics**: C-side users, high concurrency, horizontally scalable

#### 2. Admin Server (bingo-admserver)
- **Ports**: 18080 (HTTP), 18081 (gRPC)
- **Responsibilities**: Admin backend API service
- **Characteristics**: Internal management, strict permission control

#### 3. Scheduler (bingo-scheduler)
- **Port**: 8080 (Web UI)
- **Responsibilities**: Scheduled task scheduling and async task processing
- **Characteristics**: Based on Asynq, supports task retry and monitoring

#### 4. Bot Service (bingo-bot)
- **Responsibilities**: Third-party platform integration (Discord, Telegram, etc.)
- **Characteristics**: Event-driven, async processing

#### 5. CLI Tool (bingoctl)
- **Responsibilities**: Command-line tool for database migration, code generation, etc.
- **Characteristics**: Improves development efficiency

## Project Structure

```
bingo/
├── cmd/                            # Executable entry points
│   ├── bingo-apiserver/            # API service entry
│   │   └── main.go                 # Main function
│   ├── bingo-admserver/            # Admin service entry
│   ├── bingo-scheduler/            # Scheduler service entry
│   ├── bingo-bot/                  # Bot service entry
│   └── bingoctl/                   # CLI tool entry
│
├── internal/                       # Internal application code (not importable externally)
│   ├── apiserver/                  # API service implementation
│   │   ├── app.go                  # Application initialization
│   │   ├── run.go                  # Service startup logic
│   │   ├── biz/                    # Business logic layer
│   │   │   ├── auth/               # Auth business
│   │   │   ├── user/               # User business
│   │   │   └── ...                 # Other business modules
│   │   ├── controller/             # Controller layer (HTTP Handlers)
│   │   │   └── v1/                 # API v1 version
│   │   │       ├── auth/           # Auth endpoints
│   │   │       ├── user/           # User endpoints
│   │   │       └── ...
│   │   ├── store/                  # Data access layer
│   │   │   ├── user.go             # User data access
│   │   │   └── ...
│   │   ├── router/                 # Route definitions
│   │   │   └── router.go
│   │   ├── middleware/             # Middleware
│   │   │   ├── authn.go            # Authentication middleware
│   │   │   ├── authz.go            # Authorization middleware
│   │   │   └── ...
│   │   └── grpc/                   # gRPC service implementation
│   │
│   ├── admserver/                  # Admin service (same structure as apiserver)
│   ├── scheduler/                  # Scheduler service
│   │   ├── job/                    # Job definitions
│   │   └── scheduler/              # Scheduler
│   ├── bot/                        # Bot service
│   └── pkg/                        # Internal shared packages
│       ├── bootstrap/              # Application bootstrap
│       ├── config/                 # Configuration definitions
│       ├── model/                  # Data models
│       ├── logger/                 # Logging component
│       ├── db/                     # Database component
│       ├── auth/                   # Authentication component
│       ├── util/                   # Utility functions
│       └── ...
│
├── pkg/                            # Public packages importable by external projects
│   ├── api/                        # API definitions
│   ├── proto/                      # Protocol Buffer definitions
│   └── ...
│
├── api/                            # API documentation
│   ├── swagger/                    # Swagger docs
│   └── openapi/                    # OpenAPI spec
│
├── configs/                        # Configuration files
│   ├── bingo-apiserver.example.yaml
│   └── ...
│
├── deployments/                    # Deployment configs
│   └── docker/
│       └── docker-compose.yaml
│
├── build/                          # Build scripts
│   ├── docker/                     # Dockerfiles
│   └── scripts/                    # Build scripts
│
├── scripts/                        # Development scripts
│   └── make-rules/                 # Makefile rules
│
├── storage/                        # Runtime data
│   ├── log/                        # Log files
│   └── public/                     # Static assets
│
├── Makefile                        # Build configuration
├── go.mod                          # Go module definition
└── README.md                       # Project documentation
```

### Directory Descriptions

#### cmd/
Contains entry files for all executable programs, one directory per service. Follows Go standard project layout.

#### internal/
Internal code that won't be imported by external projects. This is Go's package visibility feature, ensuring internal implementations aren't externally depended upon.

**Standard structure for each service**:
- `app.go` / `run.go`: Application initialization and startup logic
- `biz/`: Business logic layer, handles business rules
- `controller/`: HTTP handlers, responsible for request/response
- `store/`: Data access layer, encapsulates database operations
- `router/`: Route configuration
- `middleware/`: Middleware
- `grpc/`: gRPC service implementation

#### pkg/
Public packages that can be imported by external projects. If your project needs to provide an SDK, place it here.

#### internal/pkg/
Internal shared packages used by multiple internal services but not exposed externally.

## Layered Architecture Details

### Three-Layer Architecture Design

```
┌─────────────────────────────────────────┐
│         Controller Layer                │  HTTP/gRPC handling layer
│  - Parameter validation                 │
│  - Request/response conversion          │
│  - Error handling                       │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Business Layer (Biz)           │  Business logic layer
│  - Business rules                       │
│  - Business process orchestration       │
│  - Transaction control                  │
└──────────────┬──────────────────────────┘
               │ Depends on
               ▼
┌─────────────────────────────────────────┐
│          Store Layer                    │  Data access layer
│  - Database operations                  │
│  - Cache operations                     │
│  - Third-party service calls            │
└─────────────────────────────────────────┘
```

### Controller Layer

**Responsibilities**:
- Receive HTTP/gRPC requests
- Parameter validation and binding
- Call Biz layer for business processing
- Return responses

**Example**:
```go
// internal/apiserver/controller/v1/user/user.go
type UserController struct {
    biz biz.IBiz
}

func (ctrl *UserController) Get(c *gin.Context) {
    // 1. Parameter validation
    var req GetUserRequest
    if err := c.ShouldBindUri(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // 2. Call business layer
    user, err := ctrl.biz.Users().Get(c.Context(), req.UserID)
    if err != nil {
        core.WriteResponse(c, err, nil)
        return
    }

    // 3. Return response
    core.WriteResponse(c, nil, user)
}
```

### Biz Layer (Business Logic Layer)

**Responsibilities**:
- Implement core business logic
- Orchestrate multiple Store operations
- Handle transactions
- Business rule validation

**Example**:
```go
// internal/apiserver/biz/user/user.go
type UserBiz struct {
    ds store.IStore
}

func (b *UserBiz) Create(ctx context.Context, req *CreateUserRequest) error {
    // 1. Business rule validation
    if err := b.validateUser(req); err != nil {
        return err
    }

    // 2. Password encryption (business logic)
    req.Password = encryptPassword(req.Password)

    // 3. Data persistence
    user := &model.User{
        Username: req.Username,
        Password: req.Password,
    }

    return b.ds.Users().Create(ctx, user)
}
```

### Store Layer (Data Access Layer)

**Responsibilities**:
- Encapsulate database operations
- Cache operations
- Data transformation

**Example**:
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

### Why Layering?

1. **Separation of Concerns**: Each layer focuses only on its responsibilities
2. **Easy to Test**: Can write unit tests for each layer
3. **Code Reuse**: Biz layer can be reused by multiple Controllers
4. **Easy to Maintain**: Modifying one layer doesn't affect others
5. **Team Collaboration**: Different layers can be developed in parallel

## Core Components Details

### 1. Configuration Management (Bootstrap)

Viper-based configuration management system supporting multiple configuration sources.

**Usage**:
```go
// internal/pkg/bootstrap/bootstrap.go
bootstrap := NewBootstrap()
bootstrap.InitConfig("bingo-apiserver.yaml")
bootstrap.Boot()  // Initialize all components
```

**Configuration file structure**:
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

### 2. Database Layer (Store)

Encapsulates GORM operations, providing a unified data access interface.

**Interface design**:
```go
// internal/apiserver/store/store.go
type IStore interface {
    Users() UserStore
    Apps() AppStore
    // ... more Stores
}
```

**Usage example**:
```go
// Using in Biz layer
user, err := biz.ds.Users().Get(ctx, userID)
```

### 3. Authentication Middleware (Authn)

JWT-based authentication middleware.

**Usage**:
```go
// internal/apiserver/router/router.go
v1 := g.Group("/v1")
{
    // Public endpoints
    v1.POST("/auth/login", authController.Login)

    // Authenticated endpoints
    auth := v1.Group("")
    auth.Use(middleware.Authn())
    {
        auth.GET("/users/:id", userController.Get)
    }
}
```

### 4. Authorization Middleware (Authz)

Casbin-based permission control.

**Permission model**:
```ini
# RBAC model
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

### 5. Logging System (Logger)

Zap-based structured logging.

**Usage**:
```go
import "github.com/bingo-project/bingo/internal/pkg/logger"

// Structured logging
logger.Info("user created",
    zap.String("username", username),
    zap.Uint64("user_id", userID),
)

// Error logging
logger.Error("failed to create user", zap.Error(err))
```

### 6. Async Tasks (Task Queue)

Asynq-based task queue.

**Define task**:
```go
// internal/scheduler/job/email.go
type EmailTask struct {
    To      string
    Subject string
    Body    string
}

func (t *EmailTask) Handle(ctx context.Context) error {
    // Email sending logic
    return sendEmail(t.To, t.Subject, t.Body)
}
```

**Submit task**:
```go
task := &EmailTask{
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "Welcome to our platform!",
}
queue.Enqueue(task)
```

### 7. API Documentation (Swagger)

Auto-generate API documentation.

**Annotation example**:
```go
// @Summary      Get user info
// @Description  Get user details by user ID
// @Tags         User Management
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /v1/users/{id} [get]
func (ctrl *UserController) Get(c *gin.Context) {
    // ...
}
```

Generate documentation:
```bash
make swagger
```

## Business Development Guide

### Standard Flow for New Features

Let's say we want to develop an "Article Management" feature.

#### 1. Define Data Model

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

#### 2. Create Database Migration

```bash
# Use bingoctl to generate migration file
bingoctl make migration create create_articles_table
```

Edit migration file:
```go
// internal/bingoctl/database/migration/xxxx_create_articles_table.go
func up(db *gorm.DB) error {
    return db.AutoMigrate(&model.Article{})
}
```

Run migration:
```bash
{app}ctl migrate up
```

#### 3. Create Store Layer

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

// ... implement other methods
```

#### 4. Create Biz Layer

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
    // 1. Business validation
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // 2. Build model
    article := &model.Article{
        Title:    req.Title,
        Content:  req.Content,
        AuthorID: req.AuthorID,
        Status:   0,
    }

    // 3. Persist
    if err := b.ds.Articles().Create(ctx, article); err != nil {
        return nil, err
    }

    return article, nil
}

// ... implement other methods
```

#### 5. Create Controller Layer

```go
// internal/apiserver/controller/v1/article/article.go
package article

type ArticleController struct {
    biz biz.IBiz
}

func New(biz biz.IBiz) *ArticleController {
    return &ArticleController{biz: biz}
}

// @Summary      Create article
// @Description  Create new article
// @Tags         Article Management
// @Accept       json
// @Produce      json
// @Param        body  body      CreateArticleRequest  true  "Article info"
// @Success      200   {object}  Article
// @Router       /v1/articles [post]
func (ctrl *ArticleController) Create(c *gin.Context) {
    var req CreateArticleRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    // Get current user ID from context
    req.AuthorID = c.GetUint64("user_id")

    article, err := ctrl.biz.Articles().Create(c.Request.Context(), &req)
    if err != nil {
        core.WriteResponse(c, err, nil)
        return
    }

    core.WriteResponse(c, nil, article)
}

// ... implement other methods
```

#### 6. Register Routes

```go
// internal/apiserver/router/router.go
func InstallRouters(g *gin.Engine) {
    // ... other routes

    // Article routes
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

#### 7. Generate API Documentation

```bash
make swagger
```

#### 8. Write Tests

```go
// internal/apiserver/biz/article/article_test.go
func TestArticleBiz_Create(t *testing.T) {
    // Prepare test data
    req := &CreateArticleRequest{
        Title:    "Test Article",
        Content:  "Test Content",
        AuthorID: 1,
    }

    // Execute test
    article, err := articleBiz.Create(context.Background(), req)

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, article)
    assert.Equal(t, "Test Article", article.Title)
}
```

Run tests:
```bash
make test
```

### Development Standards

#### Naming Conventions

- **Package names**: Lowercase, short, meaningful (e.g., `user`, `article`)
- **File names**: Snake case (e.g., `user_controller.go`)
- **Interface names**: `I` prefix (e.g., `IStore`, `IBiz`)
- **Structs**: PascalCase (e.g., `UserController`)
- **Functions/Methods**: PascalCase (exported) or camelCase (private)

#### Error Handling

Use unified error codes:
```go
// internal/pkg/errno/code.go
var (
    ErrUserNotFound = errno.New(10001, "User not found")
    ErrInvalidPassword = errno.New(10002, "Invalid password")
)
```

Return errors:
```go
if user == nil {
    return nil, errno.ErrUserNotFound
}
```

#### Logging Standards

```go
// Business logs
logger.Info("user created",
    zap.String("username", username),
    zap.Uint64("user_id", userID),
)

// Error logs (with context)
logger.Error("failed to create user",
    zap.Error(err),
    zap.String("username", username),
)
```

#### Comment Standards

```go
// CreateUser creates a new user
// Parameters:
//   - ctx: context
//   - req: create user request
// Returns:
//   - *model.User: created user
//   - error: error message
func (b *UserBiz) CreateUser(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
    // ...
}
```

## Makefile Commands

```bash
# Development
make build          # Build all services
make run            # Run services (development mode)
make test           # Run unit tests
make cover          # Test coverage report

# Code quality
make lint           # Code linting
make format         # Code formatting
make vet            # Go vet check

# Code generation
make swagger        # Generate Swagger docs
make protoc         # Compile Protocol Buffers
make wire           # Dependency injection code generation

# Deployment
make image          # Build Docker image
make push           # Push image to registry

# Cleanup
make clean          # Clean build artifacts
make tidy           # Tidy dependencies
```

## Configuration File Details

### Service Configuration

```yaml
# Basic service configuration
server:
  name: bingo                    # Service name
  mode: release                  # Run mode: release/debug/test
  addr: 0.0.0.0:8080            # Listen address
  timezone: UTC                  # Timezone

# gRPC configuration
grpc:
  addr: 0.0.0.0:8081
  network: tcp

# WebSocket configuration
websocket:
  addr: :8082

# Database configuration
mysql:
  host: 127.0.0.1:3306
  database: bingo
  username: root
  password: your_password
  maxIdleConnections: 100
  maxOpenConnections: 100
  maxConnectionLifeTime: 10s
  logLevel: 4                    # 1:silent 2:error 3:warn 4:info

# Redis configuration
redis:
  host: 127.0.0.1:6379
  password: ""
  database: 1

# JWT configuration
jwt:
  secretKey: your-secret-key     # Recommend using environment variable
  ttl: 1440000                   # Token validity (minutes)

# Logging configuration
log:
  level: info                    # debug/info/warn/error
  format: console                # console/json
  days: 7                        # Log retention days
  path: storage/log/api.log

# Feature flags
feature:
  metrics: true                  # Prometheus metrics
  profiling: true                # pprof profiling
  apiDoc: true                   # Swagger docs
  queueDash: true                # Task queue monitoring dashboard
```

### Environment Variables

Override configuration with environment variables:

```bash
export BINGO_MYSQL_HOST="localhost:3306"
export BINGO_MYSQL_PASSWORD="secret"
export BINGO_REDIS_HOST="localhost:6379"
export BINGO_JWT_SECRET="your-secret-key"
```

## Docker Deployment

### Local Development Environment

```bash
# Start dependency services (MySQL + Redis)
docker-compose -f deployments/docker/docker-compose.yaml up -d mysql redis

# Start all services
docker-compose -f deployments/docker/docker-compose.yaml up -d
```

### Production Environment

```bash
# 1. Build image
make image

# 2. Push to image registry
docker tag bingo-apiserver:latest registry.example.com/bingo-apiserver:v1.0.0
docker push registry.example.com/bingo-apiserver:v1.0.0

# 3. Deploy on production server
docker pull registry.example.com/bingo-apiserver:v1.0.0
docker run -d \
  --name bingo-apiserver \
  -p 8080:8080 \
  -v /path/to/config.yaml:/etc/bingo/config.yaml \
  registry.example.com/bingo-apiserver:v1.0.0
```

## Monitoring & Debugging

### Prometheus Metrics

Access `http://localhost:8080/metrics` to view metrics.

**Main metrics**:
- `http_requests_total`: Total HTTP requests
- `http_request_duration_seconds`: Request duration
- `go_goroutines`: Number of goroutines
- `go_memstats_alloc_bytes`: Memory usage

### pprof Performance Analysis

Access `http://localhost:8080/debug/pprof/`

```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

### Log Viewing

```bash
# Real-time log viewing
tail -f storage/log/api.log

# Format JSON logs with jq
tail -f storage/log/api.log | jq .
```

## FAQ

### How to remove unused built-in features?

1. Delete corresponding Biz layer code (`internal/*/biz/`)
2. Delete corresponding Controller layer code
3. Delete corresponding route registration
4. Delete corresponding database migration files
5. Run `make tidy` to clean unused dependencies

### How to add a new database?

1. Add database configuration to config file
2. Initialize database connection in Bootstrap
3. Inject new database connection in Store

### How to customize middleware?

```go
// internal/apiserver/middleware/custom.go
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Pre-processing

        c.Next()

        // Post-processing
    }
}
```

Register to routes:
```go
g.Use(middleware.CustomMiddleware())
```

### How to implement inter-service communication?

Using gRPC:

1. Define Protocol Buffers
2. Generate code: `make protoc`
3. Implement gRPC service
4. Call from client

### How to perform database migrations?

```bash
# Create migration file
bingoctl migrate create migration_name

# Run migration (apply all pending migrations)
{app}ctl migrate up

# Rollback last migration
{app}ctl migrate rollback

# Rollback all migrations
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

## Best Practices

### 1. API Design

- Use RESTful style
- Versioned APIs (`/v1/`, `/v2/`)
- Unified response format
- Appropriate HTTP status codes

### 2. Error Handling

- Use unified error codes
- Log detailed error messages
- Don't expose sensitive information to clients

### 3. Performance Optimization

- Proper use of cache (Redis)
- Database query optimization (indexes, pagination)
- Avoid N+1 queries
- Use connection pools

### 4. Security

- Don't hardcode sensitive information, use environment variables
- Input validation
- SQL injection protection (use parameterized queries)
- XSS protection
- CSRF protection

### 5. Testability

- Dependency injection
- Interface-based programming
- Mock external dependencies
- Write unit tests

## Advanced Topics

### Microservice Splitting

When business complexity increases, split services by business domain:

```
bingo-user-service      # User service
bingo-order-service     # Order service
bingo-payment-service   # Payment service
bingo-gateway           # API gateway
```

### Service Discovery

Integrate Consul/Etcd for service registration and discovery.

### Distributed Tracing

Integrate Jaeger/Zipkin for distributed tracing.

### Configuration Center

Use Consul/Nacos as configuration center.

## Contributing Guidelines

Welcome to submit Issues and Pull Requests!

### Development Flow

1. Fork this repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push branch: `git push origin feature/amazing-feature`
5. Submit Pull Request

### Code Review

PRs must pass:
- Code style check (golangci-lint)
- Unit tests
- At least one Maintainer's review

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

For questions or suggestions, please:
- Submit an Issue
- Email project maintainers

---

**Start using Bingo, focus on your business logic, and let the scaffold handle everything else!**
