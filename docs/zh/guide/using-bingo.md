---
title: 使用 Bingo CLI - Bingo CLI 工具指南
description: Bingo CLI 是 Bingo 框架的官方 CLI 工具，用于快速创建项目、生成 CRUD 代码、数据库迁移等。本指南详细介绍 Bingo CLI 的安装和使用方法。
---

# 使用 Bingo CLI

[Bingo CLI](https://github.com/bingo-project/bingoctl) 是 Bingo 框架的官方 CLI 工具,用于快速创建项目和生成代码,极大提升开发效率。

## 安装

```bash
go install github.com/bingo-project/bingoctl/cmd/bingo@latest
```

> 如需安装旧版本（v1.4.x 内置模板版），需从源码编译：`git clone https://github.com/bingo-project/bingoctl && cd bingoctl && git checkout v1.4.7 && go build -o bingo ./cmd/bingoctl`
> 查看 [CHANGELOG](https://github.com/bingo-project/bingoctl/blob/main/docs/en/CHANGELOG.md) 获取版本历史

验证安装:

```bash
bingo version
```

## Shell 命令补全

Bingo 支持多种 shell 的命令行自动补全。

### Zsh

```bash
# 临时生效（当前会话）
source <(bingo completion zsh)

# 永久生效
## Linux
bingo completion zsh > "${fpath[1]}/_bingo"

## macOS (Homebrew)
bingo completion zsh > $(brew --prefix)/share/zsh/site-functions/_bingo
```

> 如果补全不生效，确保 `.zshrc` 中有：`autoload -U compinit; compinit`

### Bash

```bash
# 临时生效（当前会话）
source <(bingo completion bash)

# 永久生效
## Linux
bingo completion bash > /etc/bash_completion.d/bingo

## macOS (Homebrew)
bingo completion bash > $(brew --prefix)/etc/bash_completion.d/bingo
```

> 需要安装 `bash-completion` 包

### Fish

```bash
bingo completion fish > ~/.config/fish/completions/bingo.fish
```

### PowerShell

```powershell
bingo completion powershell > bingo.ps1
# 将生成的脚本添加到 PowerShell profile
```

## 核心功能

### 1. 创建项目

一键创建完整的 Bingo 项目脚手架:

```bash
bingo create github.com/myorg/myapp
```

生成的项目包含:
- 完整的目录结构
- 配置文件模板
- Docker Compose 配置
- Makefile
- 基础示例代码

#### 创建命令选项

**项目名称和模块名**

```bash
# 使用默认模块名（与项目名相同）
bingo create github.com/myorg/myapp

# 自定义模块名（-m 选项）
bingo create myapp -m github.com/mycompany/myapp
```

**模板版本控制**

```bash
# 使用推荐版本（默认）
bingo create myapp

# 使用特定版本
bingo create myapp -r v1.2.3

# 使用分支（开发版本）
bingo create myapp -r main

# 强制重新下载模板
bingo create myapp -r main --no-cache
```

**服务选择**

控制项目中包含哪些服务（apiserver、admserver、scheduler、bot 等）：

```bash
# 只包含 apiserver（默认）
bingo create myapp

# 创建所有可用服务
bingo create myapp --all
# 或使用简写
bingo create myapp -a

# 明确指定服务（多个服务用逗号分隔）
bingo create myapp --services apiserver,admserver,scheduler

# 添加服务到默认的 apiserver
bingo create myapp --add-service scheduler

# 排除特定服务
bingo create myapp --no-service bot

# 仅包含骨架，不包含任何服务
bingo create myapp --services none
```

**Git 初始化**

```bash
# 创建项目并初始化 git 仓库（默认）
bingo create myapp

# 创建项目但不初始化 git
bingo create myapp --init-git=false
```

**构建选项**

```bash
# 创建项目但不构建（默认）
bingo create myapp

# 创建项目并运行 make build
bingo create myapp --build
```

**缓存管理和镜像配置**

```bash
# 缓存位置：~/.bingo/templates/

# 对于 GitHub 访问困难的地区，可以配置镜像
export BINGO_TEMPLATE_MIRROR=https://ghproxy.com/
bingo create myapp

# 或临时设置
BINGO_TEMPLATE_MIRROR=https://ghproxy.com/ bingo create myapp
```

### 2. 代码生成

快速生成各层代码,遵循 Bingo 的最佳实践。

#### 全局选项

```bash
-d, --directory string   指定生成文件的目录
-p, --package string     指定包名
-t, --table string       从数据库表读取字段
-s, --service string     目标服务名称，用于自动推断路径
```

#### 多服务支持

当项目包含多个服务时，可以使用 `--service` 参数自动推断生成路径：

```bash
# 为默认服务（通常是 apiserver）生成代码
bingo make model user

# 为特定服务自动推断路径
bingo make model user --service admserver

# 生成完整 CRUD（为指定服务）
bingo make crud order --service admserver

# 明确指定目录（优先级最高，覆盖 --service）
bingo make model user -d custom/path
```

**路径推断规则：**
1. 扫描 `cmd/` 目录识别已存在的服务
2. 若配置路径包含服务名，则智能替换（如 `internal/apiserver/model` → `internal/admserver/model`）
3. 否则使用默认模式：`internal/{service}/{suffix}`

#### CRUD 完整代码生成

一次性生成所有层的代码:

```bash
bingo make crud user
```

自动生成:
- `internal/pkg/model/user.go` - 数据模型
- `internal/apiserver/store/user.go` - 数据访问层
- `internal/apiserver/biz/user/user.go` - 业务逻辑层
- `internal/apiserver/handler/http/user/user.go` - 处理器层
- `pkg/api/v1/user.go` - 请求/响应定义

并自动注册到:
- Store 接口
- Biz 接口
- 路由

#### 单独生成各层

```bash
# 生成 Model
bingo make model article

# 生成 Store 层
bingo make store article

# 生成 Biz 层
bingo make biz article

# 生成 Handler 层
bingo make handler article

# 生成 Request 验证
bingo make request article
```

### 3. 从数据库生成代码

从现有数据库表自动生成 Model 代码:

```bash
# 生成单个表
bingo gen -t users

# 生成多个表
bingo gen -t users,posts,comments
```

**前提条件**: 需要在 `.bingo.yaml` 中配置数据库连接。

### 4. 数据库迁移

生成数据库迁移文件:

```bash
bingo make migration create_users_table
```

生成的迁移文件位于 `internal/pkg/database/migration/`。

可选: 从数据库表生成迁移

```bash
bingo make migration create_posts_table -t posts
```

**运行迁移**

```bash
bingo migrate <command> [options]

# 选项
-v, --verbose   显示详细编译输出
    --rebuild   强制重新编译迁移程序
-f, --force     在生产环境强制执行

# 子命令
bingo migrate up          # 运行所有待执行的迁移
bingo migrate rollback    # 回滚最后一批迁移
bingo migrate reset       # 回滚所有迁移
bingo migrate refresh     # 回滚所有并重新运行迁移
bingo migrate fresh       # 删除所有表并重新运行迁移
```

**配置迁移表名**（可选，在 `.bingo.yaml` 中）：

```yaml
migrate:
  table: bingo_migration  # 默认值
```

### 5. 生成服务模块

生成完整的服务模块:

```bash
# 生成带 HTTP 服务器的 API 服务
bingo make service api --http --with-store --with-handler

# 生成带 gRPC 服务器的服务
bingo make service rpc --grpc

# 生成同时支持 HTTP 和 gRPC 的服务
bingo make service gateway --http --grpc

# 生成纯业务处理的 worker 服务
bingo make service worker --no-biz
```

服务选项:
- `--http`: 启用 HTTP 服务器
- `--grpc`: 启用 gRPC 服务器
- `--with-biz`: 生成业务层(默认 true)
- `--no-biz`: 不生成业务层
- `--with-store`: 生成存储层
- `--with-handler`: 生成处理器层
- `--with-middleware`: 生成中间件目录
- `--with-router`: 生成路由目录

### 6. 其他生成器

```bash
# 生成中间件
bingo make middleware auth

# 生成定时任务
bingo make job cleanup

# 生成数据填充
bingo make seeder users

# 生成命令行命令
bingo make cmd serve
```

### 7. 运行数据填充

运行用户定义的 seeder 填充数据库：

```bash
bingo db seed [options]

# 选项
-v, --verbose      显示详细编译输出
    --rebuild      强制重新编译 seeder 程序
    --seeder       指定要运行的 seeder 类名

# 示例
bingo db seed                    # 运行所有 seeder
bingo db seed --seeder=User      # 仅运行 UserSeeder
bingo db seed -v                 # 显示详细输出
```

## 配置文件

在项目根目录复制示例文件创建 `.bingo.yaml`:

```bash
cp .bingo.example.yaml .bingo.yaml
```

配置文件内容:

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
  handler: internal/apiserver/handler/http
  middleware: internal/pkg/middleware
  job: internal/watcher/watcher
  migration: internal/pkg/database/migration
  seeder: internal/pkg/database/seeder

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
bingo create github.com/myorg/blog
cd blog

# 2. 配置数据库
vim .bingo.yaml  # 修改 mysql 配置

# 3. 启动依赖服务
docker-compose -f deployments/docker/docker-compose.yaml up -d

# 4. 生成用户模块
bingo make crud user

# 5. 生成文章模块
bingo make crud post

# 6. 生成评论模块
bingo make crud comment

# 7. 运行服务
make build
./_output/platforms/<os>/<arch>/blog-apiserver
```

### 示例 2: 为现有数据库生成代码

如果你有一个现有的数据库:

```bash
# 1. 配置数据库连接
vim .bingo.yaml

# 2. 从数据库生成 Model
bingo gen -t users,posts,comments,tags

# 3. 为每个表生成完整的 CRUD 代码
bingo make crud user
bingo make crud post
bingo make crud comment
bingo make crud tag
```

### 示例 3: 生成定时任务

```bash
# 1. 生成定时任务代码
bingo make job daily_report

# 2. 编辑任务逻辑
vim internal/watcher/watcher/daily_report.go

# 3. 在 scheduler 中注册任务
```

### 示例 4: 生成新的微服务

```bash
# 生成一个独立的通知服务
bingo make service notification \
  --http \
  --with-store \
  --with-handler \
  --with-router

# 生成一个纯 gRPC 服务
bingo make service user-grpc --grpc
```

## 命令参考

### 全局选项

```bash
-c, --config string   配置文件路径(默认 .bingo.yaml)
```

### create 命令

```bash
bingo create <package-name>
```

### make 命令

```bash
bingo make <type> <name> [选项]

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
- `handler` - 处理器层
- `request` - 请求验证
- `middleware` - 中间件
- `job` - 定时任务
- `migration` - 数据库迁移
- `seeder` - 数据填充
- `service` - 服务模块
- `cmd` - 命令行命令

### gen 命令

```bash
bingo gen -t <table1,table2,...>
```

## 开发工作流

推荐的开发工作流:

```
1. 创建项目
   ↓
   bingo create github.com/myorg/app

2. 配置数据库
   ↓
   编辑 .bingo.yaml

3. 生成业务模块
   ↓
   bingo make crud user
   bingo make crud post

4. 自定义业务逻辑
   ↓
   编辑 Biz 层代码

5. 添加中间件/任务
   ↓
   bingo make middleware auth
   bingo make job cleanup

6. 运行和测试
   ↓
   make build && ./_output/platforms/<os>/<arch>/app-apiserver
```

## 最佳实践

### 1. 使用 CRUD 生成器快速开始

对于标准的增删改查功能,直接使用 `make crud`:

```bash
bingo make crud product
```

### 2. 从数据库生成减少手动工作

如果已有数据库设计,使用 `gen` 命令:

```bash
bingo gen -t products,categories,orders
```

### 3. 生成后立即测试

生成代码后立即运行服务验证:

```bash
bingo make crud user
make build
./_output/platforms/<os>/<arch>/app-apiserver
curl http://localhost:8080/v1/users
```

### 4. 自定义生成的代码

生成的代码是起点,根据实际需求自定义:
- Biz 层添加业务规则
- Handler 层添加参数验证
- Store 层优化查询

### 5. 版本控制配置文件

将 `.bingo.yaml` 提交到版本控制:

```bash
git add .bingo.yaml
git commit -m "feat: add bingo config"
```

## 常见问题

### Q: 生成的代码会覆盖我的修改吗?

A: bingo 默认不会覆盖已存在的文件。如果文件已存在,会提示你是否覆盖。

### Q: 如何自定义生成的代码模板?

A: bingo 使用内置模板。如需自定义,可以 fork bingoctl 仓库修改模板。

### Q: 生成的代码能直接用于生产吗?

A: 生成的代码遵循 Bingo 最佳实践,但仍需根据实际业务场景调整:
- 添加业务验证
- 优化查询性能
- 添加错误处理
- 编写单元测试

### Q: Bingo CLI 和项目中的 bingoctl 有什么区别?

A: 它们是两个不同的工具:
- **bingo (CLI 工具)**: 独立的项目脚手架和代码生成工具
- **cmd/bingoctl (项目组件)**: Bingo 项目内置的命令行工具,可扩展自定义命令

## 下一步

- [整体架构](../essentials/architecture.md) - 深入理解微服务架构设计

## 参考资源

- [Bingo CLI GitHub 仓库](https://github.com/bingo-project/bingoctl)
- [Bingo CLI README](https://github.com/bingo-project/bingoctl/blob/main/README.md)
