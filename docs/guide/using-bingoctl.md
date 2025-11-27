# 使用 bingoctl

[bingoctl](https://github.com/bingo-project/bingoctl) 是 Bingo 框架的官方 CLI 工具,用于快速创建项目和生成代码,极大提升开发效率。

## 安装

```bash
go install github.com/bingo-project/bingoctl@latest
```

验证安装:

```bash
bingoctl version
```

## 核心功能

### 1. 创建项目

一键创建完整的 Bingo 项目脚手架:

```bash
bingoctl create github.com/myorg/myapp
```

生成的项目包含:
- 完整的目录结构
- 配置文件模板
- Docker Compose 配置
- Makefile
- 基础示例代码

### 2. 代码生成

快速生成各层代码,遵循 Bingo 的最佳实践。

#### CRUD 完整代码生成

一次性生成所有层的代码:

```bash
bingoctl make crud user
```

自动生成:
- `internal/pkg/model/user.go` - 数据模型
- `internal/apiserver/store/user.go` - 数据访问层
- `internal/apiserver/biz/user/user.go` - 业务逻辑层
- `internal/apiserver/controller/v1/user/user.go` - 控制器层
- `pkg/api/v1/user.go` - 请求/响应定义

并自动注册到:
- Store 接口
- Biz 接口
- 路由

#### 单独生成各层

```bash
# 生成 Model
bingoctl make model article

# 生成 Store 层
bingoctl make store article

# 生成 Biz 层
bingoctl make biz article

# 生成 Controller 层
bingoctl make controller article

# 生成 Request 验证
bingoctl make request article
```

### 3. 从数据库生成代码

从现有数据库表自动生成 Model 代码:

```bash
# 生成单个表
bingoctl gen -t users

# 生成多个表
bingoctl gen -t users,posts,comments
```

**前提条件**: 需要在 `.bingoctl.yaml` 中配置数据库连接。

### 4. 数据库迁移

生成数据库迁移文件:

```bash
bingoctl make migration create_users_table
```

生成的迁移文件位于 `internal/apiserver/database/migration/`。

可选: 从数据库表生成迁移

```bash
bingoctl make migration create_posts_table -t posts
```

### 5. 生成服务模块

生成完整的服务模块:

```bash
# 生成带 HTTP 服务器的 API 服务
bingoctl make service api --http --with-store --with-controller

# 生成带 gRPC 服务器的服务
bingoctl make service rpc --grpc

# 生成同时支持 HTTP 和 gRPC 的服务
bingoctl make service gateway --http --grpc

# 生成纯业务处理的 worker 服务
bingoctl make service worker --no-biz
```

服务选项:
- `--http`: 启用 HTTP 服务器
- `--grpc`: 启用 gRPC 服务器
- `--with-biz`: 生成业务层(默认 true)
- `--no-biz`: 不生成业务层
- `--with-store`: 生成存储层
- `--with-controller`: 生成控制器层
- `--with-middleware`: 生成中间件目录
- `--with-router`: 生成路由目录

### 6. 其他生成器

```bash
# 生成中间件
bingoctl make middleware auth

# 生成定时任务
bingoctl make job cleanup

# 生成数据填充
bingoctl make seeder users

# 生成命令行命令
bingoctl make cmd serve
```

## 配置文件

在项目根目录创建 `.bingoctl.yaml`:

```yaml
version: v1

# 项目包名
rootPackage: github.com/myorg/myapp

# 目录配置
directory:
  cmd: internal/bingoctl/cmd
  model: internal/pkg/model
  store: internal/apiserver/store
  request: pkg/api/v1
  biz: internal/apiserver/biz
  controller: internal/apiserver/controller/v1
  middleware: internal/pkg/middleware
  job: internal/watcher/watcher
  migration: internal/apiserver/database/migration
  seeder: internal/apiserver/database/seeder

# 注册配置
registries:
  router: internal/apiserver/router/api.go
  store:
    filePath: internal/apiserver/store/store.go
    interface: "IStore"
  biz:
    filePath: internal/apiserver/biz/biz.go
    interface: "IBiz"

# 数据库配置(用于 gen 命令)
mysql:
  host: 127.0.0.1:3306
  username: root
  password: your-password
  database: myapp
```

## 实用示例

### 示例 1: 从零开始创建博客系统

```bash
# 1. 创建项目
bingoctl create github.com/myorg/blog
cd blog

# 2. 配置数据库
vim .bingoctl.yaml  # 修改 mysql 配置

# 3. 启动依赖服务
docker-compose -f deployments/docker/docker-compose.yaml up -d

# 4. 生成用户模块
bingoctl make crud user

# 5. 生成文章模块
bingoctl make crud post

# 6. 生成评论模块
bingoctl make crud comment

# 7. 运行服务
make build
./blog-apiserver
```

### 示例 2: 为现有数据库生成代码

如果你有一个现有的数据库:

```bash
# 1. 配置数据库连接
vim .bingoctl.yaml

# 2. 从数据库生成 Model
bingoctl gen -t users,posts,comments,tags

# 3. 为每个表生成完整的 CRUD 代码
bingoctl make crud user
bingoctl make crud post
bingoctl make crud comment
bingoctl make crud tag
```

### 示例 3: 生成定时任务

```bash
# 1. 生成定时任务代码
bingoctl make job daily_report

# 2. 编辑任务逻辑
vim internal/watcher/watcher/daily_report.go

# 3. 在 scheduler 中注册任务
```

### 示例 4: 生成新的微服务

```bash
# 生成一个独立的通知服务
bingoctl make service notification \
  --http \
  --with-store \
  --with-controller \
  --with-router

# 生成一个纯 gRPC 服务
bingoctl make service user-grpc --grpc
```

## 命令参考

### 全局选项

```bash
-c, --config string   配置文件路径(默认 .bingoctl.yaml)
```

### create 命令

```bash
bingoctl create <package-name>
```

### make 命令

```bash
bingoctl make <type> <name> [选项]

选项:
  -d, --directory string   指定生成目录
  -p, --package string     指定包名
  -t, --table string       从数据库表读取字段
```

支持的类型:
- `crud` - 完整 CRUD 代码
- `model` - 数据模型
- `store` - 存储层
- `biz` - 业务逻辑层
- `controller` - 控制器层
- `request` - 请求验证
- `middleware` - 中间件
- `job` - 定时任务
- `migration` - 数据库迁移
- `seeder` - 数据填充
- `service` - 服务模块
- `cmd` - 命令行命令

### gen 命令

```bash
bingoctl gen -t <table1,table2,...>
```

## 开发工作流

推荐的开发工作流:

```
1. 创建项目
   ↓
   bingoctl create github.com/myorg/app

2. 配置数据库
   ↓
   编辑 .bingoctl.yaml

3. 生成业务模块
   ↓
   bingoctl make crud user
   bingoctl make crud post

4. 自定义业务逻辑
   ↓
   编辑 Biz 层代码

5. 添加中间件/任务
   ↓
   bingoctl make middleware auth
   bingoctl make job cleanup

6. 运行和测试
   ↓
   make build && ./app-apiserver
```

## 最佳实践

### 1. 使用 CRUD 生成器快速开始

对于标准的增删改查功能,直接使用 `make crud`:

```bash
bingoctl make crud product
```

### 2. 从数据库生成减少手动工作

如果已有数据库设计,使用 `gen` 命令:

```bash
bingoctl gen -t products,categories,orders
```

### 3. 生成后立即测试

生成代码后立即运行服务验证:

```bash
bingoctl make crud user
make build
./app-apiserver
curl http://localhost:8080/v1/users
```

### 4. 自定义生成的代码

生成的代码是起点,根据实际需求自定义:
- Biz 层添加业务规则
- Controller 层添加参数验证
- Store 层优化查询

### 5. 版本控制配置文件

将 `.bingoctl.yaml` 提交到版本控制:

```bash
git add .bingoctl.yaml
git commit -m "feat: add bingoctl config"
```

## 常见问题

### Q: 生成的代码会覆盖我的修改吗?

A: bingoctl 默认不会覆盖已存在的文件。如果文件已存在,会提示你是否覆盖。

### Q: 如何自定义生成的代码模板?

A: bingoctl 使用内置模板。如需自定义,可以 fork bingoctl 仓库修改模板。

### Q: 生成的代码能直接用于生产吗?

A: 生成的代码遵循 Bingo 最佳实践,但仍需根据实际业务场景调整:
- 添加业务验证
- 优化查询性能
- 添加错误处理
- 编写单元测试

### Q: bingoctl 和项目中的 bingoctl 有什么区别?

A: 它们是两个不同的工具:
- **bingoctl (CLI 工具)**: 独立的项目脚手架和代码生成工具
- **cmd/bingoctl (项目组件)**: Bingo 项目内置的命令行工具,用于数据库迁移等

## 下一步

- [开发第一个功能](./first-feature.md) - 基于生成的代码开发业务逻辑
- [项目结构](./project-structure.md) - 理解生成的代码结构
- [分层架构](../essentials/layered-design.md) - 理解各层职责

## 参考资源

- [bingoctl GitHub 仓库](https://github.com/bingo-project/bingoctl)
- [bingoctl README](https://github.com/bingo-project/bingoctl/blob/main/README.md)
