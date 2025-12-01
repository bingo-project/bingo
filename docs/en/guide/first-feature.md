---
title: Develop Your First Feature - Bingo Go Development Tutorial
description: Learn the complete Bingo Go microservices framework development workflow by building an article management feature, including Model, Store, Biz, and Controller layer implementations.
---

# Develop Your First Feature

This guide walks you through developing your first feature in Bingo using the code generation tools and following best practices.

## Example: Building a Blog Post Module

We'll create a complete CRUD feature for blog posts from start to finish.

### Step 1: Generate the Basic Code

Use bingo to generate the complete code skeleton:

```bash
# Generate CRUD code for post module
bingo make crud post
```

This automatically generates:
- `internal/pkg/model/post.go` - Data model
- `internal/apiserver/store/post.go` - Data access layer
- `internal/apiserver/biz/post/post.go` - Business logic layer
- `internal/apiserver/controller/v1/post/post.go` - HTTP handler
- `pkg/api/v1/post.go` - API request/response definitions

### Step 2: Define the Data Model

Edit `internal/pkg/model/post.go`:

```go
package model

import "time"

type Post struct {
    ID        string    `gorm:"primaryKey"`
    Title     string    `gorm:"index"`
    Content   string    `gorm:"type:longtext"`
    Author    string
    Status    string    // draft, published, archived
    ViewCount int64
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Implement TableName for GORM
func (Post) TableName() string {
    return "posts"
}
```

### Step 3: Implement Data Access Layer (Store)

Edit `internal/apiserver/store/post.go`:

```go
package store

import (
    "context"
    "github.com/myorg/myapp/internal/pkg/model"
    "gorm.io/gorm"
)

type PostStore interface {
    CreatePost(ctx context.Context, post *model.Post) error
    GetPostByID(ctx context.Context, id string) (*model.Post, error)
    ListPosts(ctx context.Context, status string, limit int, offset int) ([]*model.Post, error)
    UpdatePost(ctx context.Context, id string, post *model.Post) error
    DeletePost(ctx context.Context, id string) error
}

type postStore struct {
    db *gorm.DB
}

func NewPostStore(db *gorm.DB) PostStore {
    return &postStore{db: db}
}

func (s *postStore) CreatePost(ctx context.Context, post *model.Post) error {
    return s.db.WithContext(ctx).Create(post).Error
}

func (s *postStore) GetPostByID(ctx context.Context, id string) (*model.Post, error) {
    var post model.Post
    err := s.db.WithContext(ctx).First(&post, "id = ?", id).Error
    return &post, err
}

func (s *postStore) ListPosts(ctx context.Context, status string, limit int, offset int) ([]*model.Post, error) {
    var posts []*model.Post
    query := s.db.WithContext(ctx)
    if status != "" {
        query = query.Where("status = ?", status)
    }
    err := query.Limit(limit).Offset(offset).Find(&posts).Error
    return posts, err
}
```

### Step 4: Implement Business Logic Layer (Biz)

Edit `internal/apiserver/biz/post/post.go`:

```go
package post

import (
    "context"
    "fmt"
    "time"
    "github.com/myorg/myapp/internal/pkg/model"
    "github.com/myorg/myapp/internal/apiserver/store"
)

type PostBiz interface {
    CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error)
    GetPost(ctx context.Context, id string) (*Post, error)
    ListPosts(ctx context.Context, status string, limit, offset int) ([]*Post, error)
    PublishPost(ctx context.Context, id string) error
    DeletePost(ctx context.Context, id string) error
}

type postBiz struct {
    store store.PostStore
}

type CreatePostRequest struct {
    Title   string
    Content string
    Author  string
}

type Post struct {
    ID        string    `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Author    string    `json:"author"`
    Status    string    `json:"status"`
    ViewCount int64     `json:"view_count"`
    CreatedAt time.Time `json:"created_at"`
}

func NewPostBiz(s store.PostStore) PostBiz {
    return &postBiz{store: s}
}

func (b *postBiz) CreatePost(ctx context.Context, req *CreatePostRequest) (*Post, error) {
    // Validate input
    if req.Title == "" {
        return nil, fmt.Errorf("title is required")
    }

    // Create model
    post := &model.Post{
        ID:      generateID(),
        Title:   req.Title,
        Content: req.Content,
        Author:  req.Author,
        Status:  "draft",
    }

    // Save to database
    if err := b.store.CreatePost(ctx, post); err != nil {
        return nil, fmt.Errorf("failed to create post: %w", err)
    }

    return b.modelToDTO(post), nil
}

func (b *postBiz) PublishPost(ctx context.Context, id string) error {
    post, err := b.store.GetPostByID(ctx, id)
    if err != nil {
        return err
    }

    post.Status = "published"
    return b.store.UpdatePost(ctx, id, post)
}

func (b *postBiz) modelToDTO(post *model.Post) *Post {
    return &Post{
        ID:        post.ID,
        Title:     post.Title,
        Content:   post.Content,
        Author:    post.Author,
        Status:    post.Status,
        ViewCount: post.ViewCount,
        CreatedAt: post.CreatedAt,
    }
}
```

### Step 5: Implement HTTP Handler (Controller)

Edit `internal/apiserver/controller/v1/post/post.go`:

```go
package post

import (
    "github.com/gin-gonic/gin"
    "github.com/myorg/myapp/internal/apiserver/biz/post"
)

type PostController struct {
    biz post.PostBiz
}

func NewPostController(b post.PostBiz) *PostController {
    return &PostController{biz: b}
}

// Create godoc
// @Summary Create a new post
// @Tags Posts
// @Accept json
// @Produce json
// @Param post body CreatePostRequest true "Post data"
// @Success 201 {object} post.Post
// @Router /v1/posts [post]
func (c *PostController) CreatePost(ctx *gin.Context) {
    var req post.CreatePostRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }

    p, err := c.biz.CreatePost(ctx, &req)
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(201, p)
}

// GetPost godoc
// @Summary Get a post
// @Tags Posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} post.Post
// @Router /v1/posts/{id} [get]
func (c *PostController) GetPost(ctx *gin.Context) {
    id := ctx.Param("id")
    p, err := c.biz.GetPost(ctx, id)
    if err != nil {
        ctx.JSON(404, gin.H{"error": "post not found"})
        return
    }
    ctx.JSON(200, p)
}
```

### Step 6: Register Routes

Update `internal/apiserver/router/api.go`:

```go
package router

import (
    "github.com/gin-gonic/gin"
    postctrl "github.com/myorg/myapp/internal/apiserver/controller/v1/post"
)

func RegisterRoutes(engine *gin.Engine, postController *postctrl.PostController) {
    v1 := engine.Group("/v1")

    // Post routes
    posts := v1.Group("/posts")
    {
        posts.POST("", postController.CreatePost)
        posts.GET("/:id", postController.GetPost)
        posts.PUT("/:id", postController.UpdatePost)
        posts.DELETE("/:id", postController.DeletePost)
        posts.GET("", postController.ListPosts)
    }
}
```

### Step 7: Run and Test

```bash
# Build the project
make build

# Run the service
./myapp-apiserver

# Test creating a post
curl -X POST http://localhost:8080/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello World",
    "content": "This is my first post",
    "author": "John Doe"
  }'

# Test getting the post
curl http://localhost:8080/v1/posts/{post_id}
```

## Best Practices Demonstrated

1. **Separation of Concerns**: Clear layer responsibilities
2. **Error Handling**: Proper error propagation and handling
3. **Input Validation**: Validate data at the boundary
4. **Type Safety**: Use strong typing with structs
5. **Testing**: Write unit tests for each layer
6. **Documentation**: Use Swagger comments for API docs

## Testing Your Feature

Create `internal/apiserver/biz/post/post_test.go`:

```go
package post

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/myorg/myapp/internal/pkg/model"
)

type mockStore struct{}

func (m *mockStore) CreatePost(ctx context.Context, post *model.Post) error {
    return nil
}

func TestCreatePost(t *testing.T) {
    store := &mockStore{}
    biz := NewPostBiz(store)

    req := &CreatePostRequest{
        Title:   "Test Post",
        Content: "Test Content",
        Author:  "Test Author",
    }

    post, err := biz.CreatePost(context.Background(), req)

    assert.NoError(t, err)
    assert.Equal(t, "Test Post", post.Title)
}
```

## Next Step

- [Using bingo CLI](./using-bingo.md) - Boost development efficiency with the code generator
