---
title: 开发第一个功能 - Bingo Go 开发实战教程
description: 通过开发一个文章管理功能，学习 Bingo Go 微服务框架的完整开发流程，包括 Model、Store、Biz、Controller 各层的实现。
---

# 开发第一个功能

通过开发一个简单的"文章管理"功能,快速掌握 Bingo 的开发流程。

## 快速开始: 使用 bingo CLI (推荐)

如果你使用 [bingo CLI](https://github.com/bingo-project/bingoctl),可以一键生成所有代码:

```bash
# 生成文章模块的完整 CRUD 代码
bingo make crud article
```

这会自动生成 Model、Store、Biz、Controller、Request 的完整代码,并自动注册到相应的接口和路由。

> 想了解更多 bingo CLI 功能? 查看 [使用 bingo CLI](./using-bingo.md)

---

## 手动开发: 理解每一步

如果你想深入理解 Bingo 的分层架构,可以按以下步骤手动开发。

### 开发流程概览

```
1. 定义数据模型 (Model)
    ↓
2. 创建数据库迁移 (Migration)
    ↓
3. 创建 Store 层 (数据访问)
    ↓
4. 创建 Biz 层 (业务逻辑)
    ↓
5. 创建 Controller 层 (HTTP 处理)
    ↓
6. 注册路由 (Router)
    ↓
7. 生成 API 文档 (Swagger)
```

## 1. 定义数据模型

创建文件 `internal/pkg/model/article.go`:

```go
package model

import "time"

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

## 2. 创建数据库迁移

```bash
# 生成迁移文件
bingo make migration create create_articles_table
```

编辑迁移文件 `internal/bingoctl/database/migration/xxxx_create_articles_table.go`:

```go
func up(db *gorm.DB) error {
    return db.AutoMigrate(&model.Article{})
}
```

执行迁移:

```bash
{app}ctl migrate up
```

## 3. 创建 Store 层

创建文件 `internal/apiserver/store/article.go`:

```go
package store

import (
    "context"
    "gorm.io/gorm"
    "github.com/bingo-project/bingo/internal/pkg/model"
)

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

func (s *articleStore) Get(ctx context.Context, id uint64) (*model.Article, error) {
    var article model.Article
    if err := s.db.WithContext(ctx).First(&article, id).Error; err != nil {
        return nil, err
    }
    return &article, nil
}

// ... 实现其他方法
```

## 4. 创建 Biz 层

创建文件 `internal/apiserver/biz/article/article.go`:

```go
package article

import (
    "context"
    "github.com/bingo-project/bingo/internal/apiserver/store"
    "github.com/bingo-project/bingo/internal/pkg/model"
)

type ArticleBiz interface {
    Create(ctx context.Context, req *CreateArticleRequest) (*model.Article, error)
    Get(ctx context.Context, id uint64) (*model.Article, error)
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

func (b *articleBiz) Get(ctx context.Context, id uint64) (*model.Article, error) {
    return b.ds.Articles().Get(ctx, id)
}
```

## 5. 创建 Controller 层

创建文件 `internal/apiserver/controller/v1/article/article.go`:

```go
package article

import (
    "github.com/gin-gonic/gin"
    "github.com/bingo-project/bingo/internal/apiserver/biz"
    "github.com/bingo-project/bingo/pkg/core"
    "github.com/bingo-project/bingo/pkg/errno"
)

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

// @Summary      获取文章
// @Description  根据 ID 获取文章详情
// @Tags         文章管理
// @Param        id   path      int  true  "文章ID"
// @Success      200  {object}  Article
// @Router       /v1/articles/{id} [get]
func (ctrl *ArticleController) Get(c *gin.Context) {
    var req GetArticleRequest
    if err := c.ShouldBindUri(&req); err != nil {
        core.WriteResponse(c, errno.ErrBind, nil)
        return
    }

    article, err := ctrl.biz.Articles().Get(c.Request.Context(), req.ID)
    if err != nil {
        core.WriteResponse(c, err, nil)
        return
    }

    core.WriteResponse(c, nil, article)
}
```

## 6. 注册路由

编辑 `internal/apiserver/router/router.go`:

```go
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
        }
    }
}
```

## 7. 生成 API 文档

```bash
make swagger
```

访问 `http://localhost:8080/swagger/index.html` 查看生成的 API 文档。

## 测试 API

```bash
# 创建文章
curl -X POST http://localhost:8080/v1/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title": "我的第一篇文章",
    "content": "这是文章内容"
  }'

# 获取文章
curl http://localhost:8080/v1/articles/1 \
  -H "Authorization: Bearer <token>"
```

## 总结

你已经完成了第一个功能的开发!这个流程展示了 Bingo 的核心开发模式:

1. **Model**: 定义数据结构
2. **Migration**: 数据库变更
3. **Store**: 数据访问(GORM 操作)
4. **Biz**: 业务逻辑(规则验证、流程编排)
5. **Controller**: HTTP 处理(参数绑定、调用 Biz、返回响应)
6. **Router**: 路由注册

## 下一步

- [使用 bingo CLI](./using-bingo.md) - 使用代码生成器提高开发效率
